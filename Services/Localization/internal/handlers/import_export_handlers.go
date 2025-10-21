package handlers

import (
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/helixtrack/localization-service/internal/middleware"
	"github.com/helixtrack/localization-service/internal/models"
	"github.com/helixtrack/localization-service/internal/websocket"
	"go.uber.org/zap"
)

// HandleImport handles bulk import of localization data
// POST /v1/admin/import
func (h *Handler) HandleImport(c *gin.Context) {
	ctx := context.Background()
	startTime := time.Now()

	// Parse request
	var req models.ImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse import request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format: " + err.Error(),
		})
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		h.logger.Error("Import request validation failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	h.logger.Info("Starting import",
		zap.String("type", req.ImportType),
		zap.Bool("overwrite", req.OverwriteExisting),
		zap.Int("languages", len(req.Data.Languages)),
		zap.Int("keys", len(req.Data.Keys)),
	)

	// Perform import
	response := h.performImport(ctx, &req)
	response.Summary.DurationMs = time.Since(startTime).Milliseconds()

	// Log results
	h.logger.Info("Import completed",
		zap.Bool("success", response.Success),
		zap.Int("languages_imported", response.Summary.LanguagesImported),
		zap.Int("keys_imported", response.Summary.KeysImported),
		zap.Int("localizations_imported", response.Summary.LocalizationsImported),
		zap.Int("errors", len(response.Errors)),
		zap.Int64("duration_ms", response.Summary.DurationMs),
	)

	// Invalidate cache after successful import
	if response.Success {
		h.cache.DeletePattern(ctx, "l10n:*")
	}

	// Get JWT claims for audit and WebSocket event
	claims := middleware.GetClaims(c)

	// Create audit log
	h.db.CreateAuditLog(ctx, "IMPORT", "BATCH", "", claims.Username, req, c.ClientIP(), c.Request.UserAgent())

	// Broadcast WebSocket event for import completion
	failedCount := response.Summary.LanguagesSkipped + response.Summary.KeysSkipped + response.Summary.LocalizationsSkipped
	h.wsManager.BroadcastEvent(
		websocket.EventBatchOperationCompleted,
		&websocket.BatchOperationEventData{
			Operation: "import",
			Processed: response.Summary.TotalProcessed,
			Failed:    failedCount,
			Duration:  fmt.Sprintf("%dms", response.Summary.DurationMs),
		},
		&websocket.EventMetadata{
			Username: claims.Username,
		},
	)

	statusCode := http.StatusOK
	if !response.Success {
		statusCode = http.StatusPartialContent
	}

	c.JSON(statusCode, response)
}

// performImport performs the actual import operation
func (h *Handler) performImport(ctx context.Context, req *models.ImportRequest) *models.ImportResponse {
	response := &models.ImportResponse{
		Success: true,
		Summary: models.ImportSummary{},
		Errors:  []models.ImportError{},
	}

	// Import languages
	for _, langData := range req.Data.Languages {
		err := h.importLanguage(ctx, langData, req.OverwriteExisting)
		if err != nil {
			response.Errors = append(response.Errors, models.ImportError{
				Type:    "language",
				ID:      langData.Code,
				Message: err.Error(),
			})
			response.Summary.LanguagesSkipped++
			response.Success = false
		} else {
			// Check if it was update or create
			existing, _ := h.db.GetLanguageByCode(ctx, langData.Code)
			if existing != nil && req.OverwriteExisting {
				response.Summary.LanguagesUpdated++
			} else {
				response.Summary.LanguagesImported++
			}
		}
	}

	// Import localization keys
	for _, keyData := range req.Data.Keys {
		err := h.importLocalizationKey(ctx, keyData, req.OverwriteExisting)
		if err != nil {
			response.Errors = append(response.Errors, models.ImportError{
				Type:    "key",
				ID:      keyData.Key,
				Message: err.Error(),
			})
			response.Summary.KeysSkipped++
			response.Success = false
		} else {
			// Check if it was update or create
			existing, _ := h.db.GetLocalizationKeyByKey(ctx, keyData.Key)
			if existing != nil && req.OverwriteExisting {
				response.Summary.KeysUpdated++
			} else {
				response.Summary.KeysImported++
			}
		}
	}

	// Import localizations
	for languageCode, translations := range req.Data.Localizations {
		for key, value := range translations {
			err := h.importLocalization(ctx, languageCode, key, value, req.OverwriteExisting)
			if err != nil {
				response.Errors = append(response.Errors, models.ImportError{
					Type:    "localization",
					ID:      fmt.Sprintf("%s:%s", languageCode, key),
					Message: err.Error(),
				})
				response.Summary.LocalizationsSkipped++
				response.Success = false
			} else {
				response.Summary.LocalizationsImported++
			}
		}
	}

	response.Summary.TotalProcessed = response.Summary.LanguagesImported +
		response.Summary.LanguagesUpdated +
		response.Summary.KeysImported +
		response.Summary.KeysUpdated +
		response.Summary.LocalizationsImported +
		response.Summary.LocalizationsUpdated

	return response
}

// importLanguage imports a single language
func (h *Handler) importLanguage(ctx context.Context, data models.ImportLanguage, overwrite bool) error {
	// Check if language exists
	existing, _ := h.db.GetLanguageByCode(ctx, data.Code)

	if existing != nil {
		if !overwrite {
			return fmt.Errorf("language already exists and overwrite is disabled")
		}

		// Update existing language
		existing.Name = data.Name
		existing.NativeName = data.NativeName
		existing.IsRTL = data.IsRTL
		existing.IsActive = data.IsActive
		existing.IsDefault = data.IsDefault
		existing.ModifiedAt = time.Now().Unix()

		return h.db.UpdateLanguage(ctx, existing)
	}

	// Create new language
	lang := &models.Language{
		ID:         uuid.New().String(),
		Code:       data.Code,
		Name:       data.Name,
		NativeName: data.NativeName,
		IsRTL:      data.IsRTL,
		IsActive:   data.IsActive,
		IsDefault:  data.IsDefault,
		CreatedAt:  time.Now().Unix(),
		ModifiedAt: time.Now().Unix(),
		Deleted:    false,
	}

	return h.db.CreateLanguage(ctx, lang)
}

// importLocalizationKey imports a single localization key
func (h *Handler) importLocalizationKey(ctx context.Context, data models.ImportLocalizationKey, overwrite bool) error {
	// Check if key exists
	existing, _ := h.db.GetLocalizationKeyByKey(ctx, data.Key)

	if existing != nil {
		if !overwrite {
			return fmt.Errorf("key already exists and overwrite is disabled")
		}

		// Update existing key
		existing.Category = data.Category
		existing.Description = data.Description
		existing.Context = data.Context
		existing.ModifiedAt = time.Now().Unix()

		return h.db.UpdateLocalizationKey(ctx, existing)
	}

	// Create new key
	key := &models.LocalizationKey{
		ID:          uuid.New().String(),
		Key:         data.Key,
		Category:    data.Category,
		Description: data.Description,
		Context:     data.Context,
		CreatedAt:   time.Now().Unix(),
		ModifiedAt:  time.Now().Unix(),
		Deleted:     false,
	}

	return h.db.CreateLocalizationKey(ctx, key)
}

// importLocalization imports a single localization
func (h *Handler) importLocalization(ctx context.Context, languageCode, key, value string, overwrite bool) error {
	// Get language
	lang, err := h.db.GetLanguageByCode(ctx, languageCode)
	if err != nil {
		return fmt.Errorf("language not found: %s", languageCode)
	}

	// Get key
	keyObj, err := h.db.GetLocalizationKeyByKey(ctx, key)
	if err != nil {
		return fmt.Errorf("localization key not found: %s", key)
	}

	// Check if localization exists
	existing, _ := h.db.GetLocalizationByKeyAndLanguage(ctx, keyObj.ID, lang.ID)

	if existing != nil {
		if !overwrite {
			return fmt.Errorf("localization already exists and overwrite is disabled")
		}

		// Update existing localization
		existing.Value = value
		existing.Version++
		existing.ModifiedAt = time.Now().Unix()

		return h.db.UpdateLocalization(ctx, existing)
	}

	// Create new localization
	localization := &models.Localization{
		ID:         uuid.New().String(),
		KeyID:      keyObj.ID,
		LanguageID: lang.ID,
		Value:      value,
		Version:    1,
		Approved:   false, // New imports require approval
		CreatedAt:  time.Now().Unix(),
		ModifiedAt: time.Now().Unix(),
		Deleted:    false,
	}

	return h.db.CreateLocalization(ctx, localization)
}

// HandleExport handles export of localization data
// GET /v1/admin/export
func (h *Handler) HandleExport(c *gin.Context) {
	ctx := context.Background()

	// Parse query parameters
	format := c.DefaultQuery("format", "json")
	languages := c.QueryArray("languages")
	categories := c.QueryArray("categories")
	includeMetadata := c.DefaultQuery("include_metadata", "true") == "true"
	onlyApproved := c.DefaultQuery("only_approved", "true") == "true"

	req := models.ExportRequest{
		Format:          format,
		Languages:       languages,
		Categories:      categories,
		IncludeMetadata: includeMetadata,
		OnlyApproved:    onlyApproved,
		Compress:        false,
	}

	// Validate request
	if err := req.Validate(); err != nil {
		h.logger.Error("Export request validation failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	h.logger.Info("Starting export",
		zap.String("format", format),
		zap.Strings("languages", languages),
		zap.Strings("categories", categories),
	)

	// Perform export
	response, err := h.performExport(ctx, &req)
	if err != nil {
		h.logger.Error("Export failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Return based on format
	switch format {
	case "json":
		c.JSON(http.StatusOK, response)
	case "csv":
		// Return CSV data directly
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", "attachment; filename=localization-export.csv")
		c.String(http.StatusOK, response.Data.(string))
	case "xliff":
		// Return XLIFF data directly
		c.Header("Content-Type", "application/xml")
		c.Header("Content-Disposition", "attachment; filename=localization-export.xlf")
		c.String(http.StatusOK, response.Data.(string))
	}
}

// performExport performs the actual export operation
func (h *Handler) performExport(ctx context.Context, req *models.ExportRequest) (*models.ExportResponse, error) {
	// Get all languages or filter by requested languages
	allLanguages, err := h.db.GetLanguages(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get languages: %w", err)
	}

	languages := allLanguages
	if len(req.Languages) > 0 {
		languages = []*models.Language{}
		for _, code := range req.Languages {
			lang, err := h.db.GetLanguageByCode(ctx, code)
			if err != nil {
				continue
			}
			languages = append(languages, lang)
		}
	}

	// Get all localization keys or filter by categories
	var keys []*models.LocalizationKey
	if len(req.Categories) > 0 {
		for _, category := range req.Categories {
			categoryKeys, err := h.db.GetLocalizationKeysByCategory(ctx, category)
			if err != nil {
				continue
			}
			keys = append(keys, categoryKeys...)
		}
	} else {
		// Get all keys (there's no GetAllKeys method, so we'll need to get by language)
		// For now, we'll collect keys as we process localizations
		keys = []*models.LocalizationKey{}
	}

	// Build export data
	exportData := models.ImportData{
		Languages:     []models.ImportLanguage{},
		Keys:          []models.ImportLocalizationKey{},
		Localizations: make(map[string]map[string]string),
	}

	// Add languages
	for _, lang := range languages {
		exportData.Languages = append(exportData.Languages, models.ImportLanguage{
			Code:       lang.Code,
			Name:       lang.Name,
			NativeName: lang.NativeName,
			IsRTL:      lang.IsRTL,
			IsActive:   lang.IsActive,
			IsDefault:  lang.IsDefault,
		})
	}

	// Get localizations for each language
	keysSeen := make(map[string]*models.LocalizationKey)
	totalLocalizations := 0

	for _, lang := range languages {
		localizations, err := h.db.GetLocalizationsByLanguage(ctx, lang.ID)
		if err != nil {
			continue
		}

		exportData.Localizations[lang.Code] = make(map[string]string)

		for _, loc := range localizations {
			if req.OnlyApproved && !loc.Approved {
				continue
			}

			// Get the key
			key, err := h.db.GetLocalizationKeyByID(ctx, loc.KeyID)
			if err != nil {
				continue
			}

			// Add to translations
			exportData.Localizations[lang.Code][key.Key] = loc.Value
			totalLocalizations++

			// Track key
			if _, exists := keysSeen[key.ID]; !exists {
				keysSeen[key.ID] = key

				exportData.Keys = append(exportData.Keys, models.ImportLocalizationKey{
					Key:         key.Key,
					Category:    key.Category,
					Description: key.Description,
					Context:     key.Context,
					Variables:   []string{}, // Variables not stored in current model
				})
			}
		}
	}

	// Create response based on format
	response := &models.ExportResponse{
		Success: true,
		Format:  req.Format,
		Metadata: models.ExportMetadata{
			ExportedAt:    time.Now().Unix(),
			Languages:     len(languages),
			Keys:          len(exportData.Keys),
			Localizations: totalLocalizations,
			Format:        req.Format,
			Compressed:    req.Compress,
			Version:       "1.0.0",
		},
	}

	switch req.Format {
	case "json":
		response.Data = exportData
	case "csv":
		csvData, err := h.exportToCSV(exportData)
		if err != nil {
			return nil, err
		}
		response.Data = csvData
	case "xliff":
		xliffData, err := h.exportToXLIFF(exportData)
		if err != nil {
			return nil, err
		}
		response.Data = xliffData
	}

	return response, nil
}

// exportToCSV converts export data to CSV format
func (h *Handler) exportToCSV(data models.ImportData) (string, error) {
	var builder strings.Builder
	writer := csv.NewWriter(&builder)

	// Write header
	header := []string{"Key", "Category", "Description", "Context"}
	for _, lang := range data.Languages {
		header = append(header, lang.Code)
	}
	writer.Write(header)

	// Write rows
	for _, key := range data.Keys {
		row := []string{key.Key, key.Category, key.Description, key.Context}

		for _, lang := range data.Languages {
			value := ""
			if translations, exists := data.Localizations[lang.Code]; exists {
				if v, ok := translations[key.Key]; ok {
					value = v
				}
			}
			row = append(row, value)
		}

		writer.Write(row)
	}

	writer.Flush()
	return builder.String(), writer.Error()
}

// exportToXLIFF converts export data to XLIFF format
func (h *Handler) exportToXLIFF(data models.ImportData) (string, error) {
	// Simplified XLIFF 1.2 format
	var builder strings.Builder
	builder.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	builder.WriteString(`<xliff version="1.2" xmlns="urn:oasis:names:tc:xliff:document:1.2">` + "\n")

	for _, lang := range data.Languages {
		if len(data.Localizations[lang.Code]) == 0 {
			continue
		}

		builder.WriteString(fmt.Sprintf(`  <file source-language="en" target-language="%s" datatype="plaintext">`, lang.Code) + "\n")
		builder.WriteString(`    <body>` + "\n")

		for key, value := range data.Localizations[lang.Code] {
			builder.WriteString(fmt.Sprintf(`      <trans-unit id="%s">`, key) + "\n")
			builder.WriteString(fmt.Sprintf(`        <source>%s</source>`, key) + "\n")
			builder.WriteString(fmt.Sprintf(`        <target>%s</target>`, value) + "\n")
			builder.WriteString(`      </trans-unit>` + "\n")
		}

		builder.WriteString(`    </body>` + "\n")
		builder.WriteString(`  </file>` + "\n")
	}

	builder.WriteString(`</xliff>` + "\n")
	return builder.String(), nil
}

// HandleBatchLocalizations handles batch localization operations
// POST /v1/admin/localizations/batch
func (h *Handler) HandleBatchLocalizations(c *gin.Context) {
	ctx := context.Background()
	startTime := time.Now()

	// Parse request
	var req models.BatchLocalizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse batch request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format: " + err.Error(),
		})
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		h.logger.Error("Batch request validation failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	h.logger.Info("Starting batch operation",
		zap.String("operation", req.Operation),
		zap.Int("count", len(req.Localizations)),
	)

	// Perform batch operation
	response := h.performBatchOperation(ctx, &req)
	response.Summary.DurationMs = time.Since(startTime).Milliseconds()

	h.logger.Info("Batch operation completed",
		zap.Bool("success", response.Success),
		zap.Int("successful", response.Summary.Successful),
		zap.Int("failed", response.Summary.Failed),
		zap.Int64("duration_ms", response.Summary.DurationMs),
	)

	// Invalidate cache after successful operation
	if response.Success {
		h.cache.DeletePattern(ctx, "l10n:*")
	}

	c.JSON(http.StatusOK, response)
}

// performBatchOperation performs a batch localization operation
func (h *Handler) performBatchOperation(ctx context.Context, req *models.BatchLocalizationRequest) *models.BatchLocalizationResponse {
	response := &models.BatchLocalizationResponse{
		Success: true,
		Summary: models.BatchSummary{
			TotalRequested: len(req.Localizations),
		},
		Errors: []models.BatchError{},
	}

	for i, item := range req.Localizations {
		var err error

		switch req.Operation {
		case "create", "update":
			err = h.importLocalization(ctx, item.LanguageCode, item.Key, item.Value, req.Operation == "update")
		case "approve":
			err = h.approveLocalization(ctx, item.LanguageCode, item.Key)
		case "delete":
			err = h.deleteLocalization(ctx, item.LanguageCode, item.Key)
		}

		if err != nil {
			response.Errors = append(response.Errors, models.BatchError{
				Index:   i,
				Key:     item.Key,
				Message: err.Error(),
			})
			response.Summary.Failed++
			response.Success = false
		} else {
			response.Summary.Successful++
		}
	}

	return response
}

// approveLocalization approves a localization
func (h *Handler) approveLocalization(ctx context.Context, languageCode, key string) error {
	// Get language
	lang, err := h.db.GetLanguageByCode(ctx, languageCode)
	if err != nil {
		return fmt.Errorf("language not found: %s", languageCode)
	}

	// Get key
	keyObj, err := h.db.GetLocalizationKeyByKey(ctx, key)
	if err != nil {
		return fmt.Errorf("localization key not found: %s", key)
	}

	// Get localization
	loc, err := h.db.GetLocalizationByKeyAndLanguage(ctx, keyObj.ID, lang.ID)
	if err != nil {
		return fmt.Errorf("localization not found")
	}

	// Approve
	return h.db.ApproveLocalization(ctx, loc.ID, "admin")
}

// deleteLocalization deletes a localization
func (h *Handler) deleteLocalization(ctx context.Context, languageCode, key string) error {
	// Get language
	lang, err := h.db.GetLanguageByCode(ctx, languageCode)
	if err != nil {
		return fmt.Errorf("language not found: %s", languageCode)
	}

	// Get key
	keyObj, err := h.db.GetLocalizationKeyByKey(ctx, key)
	if err != nil {
		return fmt.Errorf("localization key not found: %s", key)
	}

	// Get localization
	loc, err := h.db.GetLocalizationByKeyAndLanguage(ctx, keyObj.ID, lang.ID)
	if err != nil {
		return fmt.Errorf("localization not found")
	}

	// Delete
	return h.db.DeleteLocalization(ctx, loc.ID)
}

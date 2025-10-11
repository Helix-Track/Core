package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"helixtrack.ru/core/internal/logger"
	"helixtrack.ru/core/internal/middleware"
	"helixtrack.ru/core/internal/models"
)

// handleExtensionCreate creates a new extension
func (h *Handler) handleExtensionCreate(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Check permissions
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "extension", models.PermissionCreate)
	if err != nil {
		logger.Error("Permission check failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodePermissionServiceError,
			"Permission check failed",
			"",
		))
		return
	}

	if !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permission",
			"",
		))
		return
	}

	// Parse extension data from request
	title, ok := req.Data["title"].(string)
	if !ok || title == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing title",
			"",
		))
		return
	}

	enabled := false
	if enabledVal, ok := req.Data["enabled"].(bool); ok {
		enabled = enabledVal
	}

	// Get description as pointer
	var descPtr *string
	if descStr := getStringFromData(req.Data, "description"); descStr != "" {
		descPtr = &descStr
	}

	extension := &models.Extension{
		ID:          uuid.New().String(),
		Title:       title,
		Description: descPtr,
		Version:     getStringFromData(req.Data, "version"),
		Enabled:     enabled,
		Created:     time.Now().Unix(),
		Modified:    time.Now().Unix(),
		Deleted:     false,
	}

	// Validate extension
	if !extension.IsValid() {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidData,
			"Invalid extension data",
			"",
		))
		return
	}

	// Insert into database
	query := `
		INSERT INTO extension (id, title, description, version, enabled, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(c.Request.Context(), query,
		extension.ID,
		extension.Title,
		extension.Description,
		extension.Version,
		extension.Enabled,
		extension.Created,
		extension.Modified,
		extension.Deleted,
	)

	if err != nil {
		logger.Error("Failed to create extension", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create extension",
			"",
		))
		return
	}

	logger.Info("Extension created",
		zap.String("extension_id", extension.ID),
		zap.String("title", extension.Title),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"extension": extension,
	})
	c.JSON(http.StatusCreated, response)
}

// handleExtensionRead reads a single extension by ID
func (h *Handler) handleExtensionRead(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get extension ID from request
	extensionID, ok := req.Data["id"].(string)
	if !ok || extensionID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing extension ID",
			"",
		))
		return
	}

	// Query extension from database
	query := `
		SELECT id, title, description, version, enabled, created, modified, deleted
		FROM extension
		WHERE id = ? AND deleted = 0
	`

	var extension models.Extension
	err := h.db.QueryRow(c.Request.Context(), query, extensionID).Scan(
		&extension.ID,
		&extension.Title,
		&extension.Description,
		&extension.Version,
		&extension.Enabled,
		&extension.Created,
		&extension.Modified,
		&extension.Deleted,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Extension not found",
			"",
		))
		return
	}

	if err != nil {
		logger.Error("Failed to read extension", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to read extension",
			"",
		))
		return
	}

	logger.Info("Extension read",
		zap.String("extension_id", extension.ID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"extension": extension,
	})
	c.JSON(http.StatusOK, response)
}

// handleExtensionList lists all extensions
func (h *Handler) handleExtensionList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Query all non-deleted extensions
	query := `
		SELECT id, title, description, version, enabled, created, modified, deleted
		FROM extension
		WHERE deleted = 0
		ORDER BY title ASC
	`

	rows, err := h.db.Query(c.Request.Context(), query)
	if err != nil {
		logger.Error("Failed to list extensions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list extensions",
			"",
		))
		return
	}
	defer rows.Close()

	extensions := make([]models.Extension, 0)
	for rows.Next() {
		var extension models.Extension
		err := rows.Scan(
			&extension.ID,
			&extension.Title,
			&extension.Description,
			&extension.Version,
			&extension.Enabled,
			&extension.Created,
			&extension.Modified,
			&extension.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan extension", zap.Error(err))
			continue
		}
		extensions = append(extensions, extension)
	}

	logger.Info("Extensions listed",
		zap.Int("count", len(extensions)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"extensions": extensions,
		"count":      len(extensions),
	})
	c.JSON(http.StatusOK, response)
}

// handleExtensionModify updates an existing extension
func (h *Handler) handleExtensionModify(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Check permissions
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "extension", models.PermissionUpdate)
	if err != nil {
		logger.Error("Permission check failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodePermissionServiceError,
			"Permission check failed",
			"",
		))
		return
	}

	if !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permission",
			"",
		))
		return
	}

	// Get extension ID
	extensionID, ok := req.Data["id"].(string)
	if !ok || extensionID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing extension ID",
			"",
		))
		return
	}

	// Check if extension exists
	checkQuery := `SELECT COUNT(*) FROM extension WHERE id = ? AND deleted = 0`
	var count int
	err = h.db.QueryRow(c.Request.Context(), checkQuery, extensionID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Extension not found",
			"",
		))
		return
	}

	// Build update query dynamically based on provided fields
	updates := make(map[string]interface{})

	if title, ok := req.Data["title"].(string); ok && title != "" {
		updates["title"] = title
	}
	if description, ok := req.Data["description"].(string); ok {
		if description != "" {
			updates["description"] = description
		} else {
			updates["description"] = nil
		}
	}
	if version, ok := req.Data["version"].(string); ok {
		updates["version"] = version
	}
	if enabled, ok := req.Data["enabled"].(bool); ok {
		updates["enabled"] = enabled
	}

	updates["modified"] = time.Now().Unix()

	if len(updates) == 1 { // Only modified was set
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"No fields to update",
			"",
		))
		return
	}

	// Build and execute update query
	queryStr := "UPDATE extension SET "
	args := make([]interface{}, 0)
	first := true

	for key, value := range updates {
		if !first {
			queryStr += ", "
		}
		queryStr += key + " = ?"
		args = append(args, value)
		first = false
	}

	queryStr += " WHERE id = ?"
	args = append(args, extensionID)

	_, err = h.db.Exec(c.Request.Context(), queryStr, args...)
	if err != nil {
		logger.Error("Failed to update extension", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to update extension",
			"",
		))
		return
	}

	logger.Info("Extension updated",
		zap.String("extension_id", extensionID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"updated": true,
		"id":      extensionID,
	})
	c.JSON(http.StatusOK, response)
}

// handleExtensionRemove soft-deletes an extension
func (h *Handler) handleExtensionRemove(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Check permissions
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "extension", models.PermissionDelete)
	if err != nil {
		logger.Error("Permission check failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodePermissionServiceError,
			"Permission check failed",
			"",
		))
		return
	}

	if !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permission",
			"",
		))
		return
	}

	// Get extension ID
	extensionID, ok := req.Data["id"].(string)
	if !ok || extensionID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing extension ID",
			"",
		))
		return
	}

	// Soft delete the extension
	query := `UPDATE extension SET deleted = 1, modified = ? WHERE id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), extensionID)
	if err != nil {
		logger.Error("Failed to delete extension", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to delete extension",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Extension not found",
			"",
		))
		return
	}

	logger.Info("Extension deleted",
		zap.String("extension_id", extensionID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"deleted": true,
		"id":      extensionID,
	})
	c.JSON(http.StatusOK, response)
}

// handleExtensionEnable enables an extension
func (h *Handler) handleExtensionEnable(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Check permissions
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "extension", models.PermissionUpdate)
	if err != nil {
		logger.Error("Permission check failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodePermissionServiceError,
			"Permission check failed",
			"",
		))
		return
	}

	if !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permission",
			"",
		))
		return
	}

	// Get extension ID
	extensionID, ok := req.Data["id"].(string)
	if !ok || extensionID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing extension ID",
			"",
		))
		return
	}

	// Enable the extension
	query := `UPDATE extension SET enabled = 1, modified = ? WHERE id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), extensionID)
	if err != nil {
		logger.Error("Failed to enable extension", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to enable extension",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Extension not found",
			"",
		))
		return
	}

	logger.Info("Extension enabled",
		zap.String("extension_id", extensionID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"enabled": true,
		"id":      extensionID,
	})
	c.JSON(http.StatusOK, response)
}

// handleExtensionDisable disables an extension
func (h *Handler) handleExtensionDisable(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Check permissions
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "extension", models.PermissionUpdate)
	if err != nil {
		logger.Error("Permission check failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodePermissionServiceError,
			"Permission check failed",
			"",
		))
		return
	}

	if !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permission",
			"",
		))
		return
	}

	// Get extension ID
	extensionID, ok := req.Data["id"].(string)
	if !ok || extensionID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing extension ID",
			"",
		))
		return
	}

	// Disable the extension
	query := `UPDATE extension SET enabled = 0, modified = ? WHERE id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), extensionID)
	if err != nil {
		logger.Error("Failed to disable extension", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to disable extension",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Extension not found",
			"",
		))
		return
	}

	logger.Info("Extension disabled",
		zap.String("extension_id", extensionID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"disabled": true,
		"id":       extensionID,
	})
	c.JSON(http.StatusOK, response)
}

// handleExtensionSetMetadata sets metadata for an extension
func (h *Handler) handleExtensionSetMetadata(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Parse metadata from request
	extensionID, ok := req.Data["extensionId"].(string)
	if !ok || extensionID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing extensionId",
			"",
		))
		return
	}

	property, ok := req.Data["property"].(string)
	if !ok || property == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing property",
			"",
		))
		return
	}

	// Value can be any type, convert to JSON string
	var valueStr string
	if value, ok := req.Data["value"]; ok {
		valueBytes, err := json.Marshal(value)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				models.ErrorCodeInvalidData,
				"Invalid value format",
				"",
			))
			return
		}
		valueStr = string(valueBytes)
	}

	metadata := &models.ExtensionMetaData{
		ID:          uuid.New().String(),
		ExtensionID: extensionID,
		Property:    property,
		Value:       valueStr,
		Created:     time.Now().Unix(),
		Modified:    time.Now().Unix(),
		Deleted:     false,
	}

	// Insert into database
	query := `
		INSERT INTO extension_metadata (id, extension_id, property, value, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := h.db.Exec(c.Request.Context(), query,
		metadata.ID,
		metadata.ExtensionID,
		metadata.Property,
		metadata.Value,
		metadata.Created,
		metadata.Modified,
		metadata.Deleted,
	)

	if err != nil {
		logger.Error("Failed to set extension metadata", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to set extension metadata",
			"",
		))
		return
	}

	logger.Info("Extension metadata set",
		zap.String("metadata_id", metadata.ID),
		zap.String("extension_id", extensionID),
		zap.String("property", property),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"metadata": metadata,
	})
	c.JSON(http.StatusCreated, response)
}

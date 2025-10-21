package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/helixtrack/localization-service/internal/middleware"
	"github.com/helixtrack/localization-service/internal/models"
	"github.com/helixtrack/localization-service/internal/websocket"
	"go.uber.org/zap"
)

// GetCurrentVersion returns the current localization version
// GET /v1/version/current
func (h *Handler) GetCurrentVersion(c *gin.Context) {
	ctx := context.Background()

	version, err := h.db.GetCurrentVersion(ctx)
	if err != nil {
		h.logger.Error("Failed to get current version", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "No version found",
		})
		return
	}

	// Build version info response
	versionInfo := models.VersionInfo{
		Version:        version.VersionNumber,
		KeysCount:      version.KeysCount,
		LanguagesCount: version.LanguagesCount,
		LastUpdated:    version.CreatedAt,
	}

	c.JSON(http.StatusOK, versionInfo)
}

// GetVersionHistory returns version history with pagination
// GET /v1/version/history
func (h *Handler) GetVersionHistory(c *gin.Context) {
	ctx := context.Background()

	// Parse pagination parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	// Get versions
	versions, err := h.db.ListVersions(ctx, limit, offset)
	if err != nil {
		h.logger.Error("Failed to list versions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve version history",
		})
		return
	}

	// Get total count
	totalCount, err := h.db.CountVersions(ctx)
	if err != nil {
		h.logger.Error("Failed to count versions", zap.Error(err))
		totalCount = len(versions)
	}

	// Get current version
	currentVersion := ""
	if len(versions) > 0 {
		currentVersion = versions[0].VersionNumber
	}

	response := models.VersionHistoryResponse{
		Versions:       make([]models.LocalizationVersion, len(versions)),
		TotalVersions:  totalCount,
		CurrentVersion: currentVersion,
	}

	// Convert pointers to values
	for i, v := range versions {
		response.Versions[i] = *v
	}

	c.JSON(http.StatusOK, response)
}

// GetVersionByNumber returns a specific version by number
// GET /v1/version/:version
func (h *Handler) GetVersionByNumber(c *gin.Context) {
	ctx := context.Background()
	versionNumber := c.Param("version")

	version, err := h.db.GetVersionByNumber(ctx, versionNumber)
	if err != nil {
		h.logger.Error("Failed to get version", zap.Error(err), zap.String("version", versionNumber))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Version not found",
		})
		return
	}

	c.JSON(http.StatusOK, version)
}

// GetCatalogByVersion returns a catalog for a specific version
// GET /v1/version/:version/catalog/:language
func (h *Handler) GetCatalogByVersion(c *gin.Context) {
	ctx := context.Background()
	versionNumber := c.Param("version")
	languageCode := c.Param("language")

	// Check cache first
	cacheKey := "l10n:catalog:" + languageCode + ":version:" + versionNumber
	if cached, err := h.cache.Get(ctx, cacheKey); err == nil && cached != "" {
		h.logger.Info("Cache hit for versioned catalog",
			zap.String("version", versionNumber),
			zap.String("language", languageCode),
		)

		var catalogMap map[string]string
		if err := json.Unmarshal([]byte(cached), &catalogMap); err == nil {
			c.JSON(http.StatusOK, gin.H{
				"catalog": catalogMap,
				"version": versionNumber,
			})
			return
		}
	}

	// Get from database
	catalog, err := h.db.GetCatalogByVersion(ctx, versionNumber, languageCode)
	if err != nil {
		h.logger.Error("Failed to get catalog by version",
			zap.Error(err),
			zap.String("version", versionNumber),
			zap.String("language", languageCode),
		)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Catalog not found for this version and language",
		})
		return
	}

	// Parse catalog data
	catalogMap, err := catalog.GetCatalogMap()
	if err != nil {
		h.logger.Error("Failed to parse catalog data", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to parse catalog data",
		})
		return
	}

	// Cache the result (versioned catalogs can be cached longer)
	if catalogJSON, err := json.Marshal(catalogMap); err == nil {
		h.cache.Set(ctx, cacheKey, string(catalogJSON), 24*time.Hour) // 24 hours for versioned catalogs
	}

	c.JSON(http.StatusOK, gin.H{
		"catalog": catalogMap,
		"version": versionNumber,
	})
}

// CreateVersion creates a new version (admin only)
// POST /v1/admin/version/create
func (h *Handler) CreateVersion(c *gin.Context) {
	ctx := context.Background()

	// Parse request
	var req models.CreateVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse create version request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format: " + err.Error(),
		})
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		h.logger.Error("Create version request validation failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Get current version
	currentVersion, err := h.db.GetCurrentVersion(ctx)
	var newVersionNumber string

	if err != nil {
		// No current version, start with 1.0.0
		newVersionNumber = "1.0.0"
	} else {
		// Increment based on type
		newVersionNumber, err = models.IncrementVersion(currentVersion.VersionNumber, req.VersionType)
		if err != nil {
			h.logger.Error("Failed to increment version", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to generate new version number",
			})
			return
		}
	}

	// Get current counts
	keysCount, _ := h.db.CountVersions(ctx) // Placeholder - should count actual keys
	languagesCount, _ := h.db.CountVersions(ctx) // Placeholder - should count actual languages
	totalLocalizations, _ := h.db.CountVersions(ctx) // Placeholder - should count actual localizations

	// Get username from JWT claims
	username, exists := c.Get("username")
	if !exists {
		username = "admin"
	}

	// Create version
	version := &models.LocalizationVersion{
		ID:                 uuid.New().String(),
		VersionNumber:      newVersionNumber,
		VersionType:        req.VersionType,
		Description:        req.Description,
		KeysCount:          keysCount,
		LanguagesCount:     languagesCount,
		TotalLocalizations: totalLocalizations,
		CreatedBy:          username.(string),
		CreatedAt:          time.Now().Unix(),
	}

	// Set metadata
	if req.Metadata != nil {
		if err := version.SetMetadataMap(req.Metadata); err != nil {
			h.logger.Error("Failed to set metadata", zap.Error(err))
		}
	}

	// Save to database
	if err := h.db.CreateVersion(ctx, version); err != nil {
		h.logger.Error("Failed to create version", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create version",
		})
		return
	}

	h.logger.Info("Version created",
		zap.String("version", version.VersionNumber),
		zap.String("type", version.VersionType),
		zap.String("created_by", version.CreatedBy),
	)

	// Get JWT claims for audit and WebSocket event
	claims := middleware.GetClaims(c)

	// Create audit log
	h.db.CreateAuditLog(ctx, "CREATE", "VERSION", "", claims.Username, version, c.ClientIP(), c.Request.UserAgent())

	// Broadcast WebSocket event
	h.wsManager.BroadcastEvent(
		websocket.EventVersionCreated,
		&websocket.VersionEventData{
			ID:                version.ID,
			Version:           version.VersionNumber,
			Description:       version.Description,
			KeysCount:         version.KeysCount,
			LanguagesCount:    version.LanguagesCount,
			TranslationsCount: version.TotalLocalizations,
		},
		&websocket.EventMetadata{
			Username: claims.Username,
		},
	)

	c.JSON(http.StatusCreated, version)
}

// DeleteVersion deletes a version (admin only)
// DELETE /v1/admin/version/:version
func (h *Handler) DeleteVersion(c *gin.Context) {
	ctx := context.Background()
	versionNumber := c.Param("version")

	// Get version first to get ID
	version, err := h.db.GetVersionByNumber(ctx, versionNumber)
	if err != nil {
		h.logger.Error("Failed to get version", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Version not found",
		})
		return
	}

	// Don't allow deleting the current version
	currentVersion, err := h.db.GetCurrentVersion(ctx)
	if err == nil && currentVersion.ID == version.ID {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Cannot delete the current version",
		})
		return
	}

	// Delete version
	if err := h.db.DeleteVersion(ctx, version.ID); err != nil {
		h.logger.Error("Failed to delete version", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete version",
		})
		return
	}

	h.logger.Info("Version deleted", zap.String("version", versionNumber))

	// Get JWT claims for audit and WebSocket event
	claims := middleware.GetClaims(c)

	// Create audit log
	h.db.CreateAuditLog(ctx, "DELETE", "VERSION", "", claims.Username, nil, c.ClientIP(), c.Request.UserAgent())

	// Broadcast WebSocket event
	h.wsManager.BroadcastEvent(
		websocket.EventVersionDeleted,
		&websocket.VersionEventData{
			ID:                version.ID,
			Version:           version.VersionNumber,
			Description:       version.Description,
			KeysCount:         version.KeysCount,
			LanguagesCount:    version.LanguagesCount,
			TranslationsCount: version.TotalLocalizations,
		},
		&websocket.EventMetadata{
			Username: claims.Username,
		},
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Version deleted successfully",
		"version": versionNumber,
	})
}

package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/helixtrack/localization-service/internal/cache"
	"github.com/helixtrack/localization-service/internal/database"
	"github.com/helixtrack/localization-service/internal/middleware"
	"github.com/helixtrack/localization-service/internal/models"
	"github.com/helixtrack/localization-service/internal/websocket"
	"go.uber.org/zap"
)

// Handler dependencies
type Handler struct {
	db        database.Database
	cache     cache.Cache
	logger    *zap.Logger
	wsManager *websocket.Manager
}

// NewHandler creates a new handler instance
func NewHandler(db database.Database, cache cache.Cache, logger *zap.Logger, wsManager *websocket.Manager) *Handler {
	return &Handler{
		db:        db,
		cache:     cache,
		logger:    logger,
		wsManager: wsManager,
	}
}

// RegisterRoutes registers all API routes
func (h *Handler) RegisterRoutes(router *gin.Engine, jwtSecret string, adminRoles []string) {
	// Public routes
	router.GET("/health", h.HealthCheck)

	// API v1 routes
	v1 := router.Group("/v1")
	{
		// Authentication required
		v1.Use(middleware.JWTAuth(jwtSecret, h.logger))

		// Catalog routes
		v1.GET("/catalog/:language", h.GetCatalog)

		// Localization routes
		v1.GET("/localize/:key", h.GetLocalization)
		v1.POST("/localize/batch", h.BatchLocalize)

		// Language routes
		v1.GET("/languages", h.ListLanguages)

		// Version routes (public)
		v1.GET("/version/current", h.GetCurrentVersion)
		v1.GET("/version/history", h.GetVersionHistory)
		v1.GET("/version/:version", h.GetVersionByNumber)
		v1.GET("/version/:version/catalog/:language", h.GetCatalogByVersion)

		// Admin routes
		admin := v1.Group("/admin")
		admin.Use(middleware.AdminOnly(adminRoles))
		{
			// Language admin
			admin.POST("/languages", h.CreateLanguage)
			admin.PUT("/languages/:id", h.UpdateLanguage)
			admin.DELETE("/languages/:id", h.DeleteLanguage)

			// Localization admin
			admin.POST("/localizations", h.CreateLocalization)
			admin.PUT("/localizations/:id", h.UpdateLocalization)
			admin.DELETE("/localizations/:id", h.DeleteLocalization)
			admin.POST("/localizations/:id/approve", h.ApproveLocalization)
			admin.POST("/localizations/batch", h.HandleBatchLocalizations)

			// Import/Export
			admin.POST("/import", h.HandleImport)
			admin.GET("/export", h.HandleExport)

			// Version admin
			admin.POST("/version/create", h.CreateVersion)
			admin.DELETE("/version/:version", h.DeleteVersion)

			// Cache admin
			admin.POST("/cache/invalidate", h.InvalidateCache)

			// Stats
			admin.GET("/stats", h.GetStats)
		}
	}
}

// HealthCheck handles health check requests
func (h *Handler) HealthCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	checks := make(map[string]interface{})

	// Check database
	dbStart := time.Now()
	if err := h.db.Ping(); err != nil {
		checks["database"] = map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
		}
		c.JSON(http.StatusServiceUnavailable, models.HealthResponse{
			Status:  "unhealthy",
			Version: "1.0.0",
			Checks:  checks,
		})
		return
	}

	checks["database"] = map[string]interface{}{
		"status":     "healthy",
		"latency_ms": time.Since(dbStart).Milliseconds(),
	}

	// Check cache
	if _, err := h.cache.Exists(ctx, "health_check"); err != nil {
		checks["cache"] = map[string]interface{}{
			"status": "degraded",
			"error":  err.Error(),
		}
	} else {
		checks["cache"] = map[string]interface{}{
			"status": "healthy",
		}
	}

	c.JSON(http.StatusOK, models.HealthResponse{
		Status:  "healthy",
		Version: "1.0.0",
		Checks:  checks,
	})
}

// GetCatalog retrieves a complete localization catalog
func (h *Handler) GetCatalog(c *gin.Context) {
	languageCode := c.Param("language")
	category := c.Query("category")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Generate cache key
	cacheKey := cache.CacheKey("catalog", languageCode, category)

	// Try cache first
	if cached, err := h.cache.Get(ctx, cacheKey); err == nil {
		var response models.CatalogResponse
		if err := json.Unmarshal([]byte(cached), &response); err == nil {
			h.logger.Debug("catalog cache hit", zap.String("language", languageCode))
			c.JSON(http.StatusOK, models.SuccessResponse(response))
			return
		}
	}

	// Get language
	lang, err := h.db.GetLanguageByCode(ctx, languageCode)
	if err != nil {
		h.logger.Error("failed to get language", zap.Error(err))
		c.JSON(http.StatusNotFound, models.ErrorResponse(
			models.ErrCodeNotFound,
			"language not found",
		))
		return
	}

	// Get or build catalog
	catalog, err := h.db.GetLatestCatalog(ctx, lang.ID, category)
	if err != nil {
		h.logger.Error("failed to get catalog", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrCodeInternalError,
			"failed to retrieve catalog",
		))
		return
	}

	// Parse catalog data
	catalogMap, err := catalog.GetCatalogMap()
	if err != nil {
		h.logger.Error("failed to parse catalog", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrCodeInternalError,
			"failed to parse catalog",
		))
		return
	}

	response := models.CatalogResponse{
		Language: languageCode,
		Version:  catalog.Version,
		Checksum: catalog.Checksum,
		Catalog:  catalogMap,
	}

	// Cache the response
	if responseJSON, err := json.Marshal(response); err == nil {
		h.cache.Set(ctx, cacheKey, string(responseJSON), 1*time.Hour)
	}

	c.JSON(http.StatusOK, models.SuccessResponse(response))
}

// GetLocalization retrieves a single localization
func (h *Handler) GetLocalization(c *gin.Context) {
	key := c.Param("key")
	languageCode := c.Query("language")
	fallback := c.DefaultQuery("fallback", "true") == "true"

	if languageCode == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrCodeValidationFailed,
			"language parameter is required",
		))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get language
	lang, err := h.db.GetLanguageByCode(ctx, languageCode)
	if err != nil {
		if fallback {
			// Try default language
			defaultLang, err := h.db.GetDefaultLanguage(ctx)
			if err != nil {
				c.JSON(http.StatusNotFound, models.ErrorResponse(
					models.ErrCodeNotFound,
					"language not found",
				))
				return
			}
			lang = defaultLang
		} else {
			c.JSON(http.StatusNotFound, models.ErrorResponse(
				models.ErrCodeNotFound,
				"language not found",
			))
			return
		}
	}

	// Get localization key
	locKey, err := h.db.GetLocalizationKeyByKey(ctx, key)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(
			models.ErrCodeNotFound,
			"localization key not found",
		))
		return
	}

	// Get localization
	loc, err := h.db.GetLocalizationByKeyAndLanguage(ctx, locKey.ID, lang.ID)
	if err != nil {
		if fallback {
			// Try default language
			defaultLang, err := h.db.GetDefaultLanguage(ctx)
			if err == nil {
				loc, err = h.db.GetLocalizationByKeyAndLanguage(ctx, locKey.ID, defaultLang.ID)
			}
		}

		if err != nil {
			c.JSON(http.StatusNotFound, models.ErrorResponse(
				models.ErrCodeNotFound,
				"localization not found",
			))
			return
		}
	}

	// Parse variables
	var variables []string
	if loc.Variables != nil {
		json.Unmarshal(loc.Variables, &variables)
	}

	response := models.LocalizationResponse{
		Key:      key,
		Language: lang.Code,
		Value:    loc.Value,
		Variables: variables,
		Approved: loc.Approved,
	}

	c.JSON(http.StatusOK, models.SuccessResponse(response))
}

// BatchLocalize retrieves multiple localizations at once
func (h *Handler) BatchLocalize(c *gin.Context) {
	var req models.GetBatchLocalizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrCodeValidationFailed,
			"invalid request body",
		))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get language
	lang, err := h.db.GetLanguageByCode(ctx, req.Language)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(
			models.ErrCodeNotFound,
			"language not found",
		))
		return
	}

	// Get default language for fallback
	var defaultLang *models.Language
	if req.Fallback {
		defaultLang, _ = h.db.GetDefaultLanguage(ctx)
	}

	// Retrieve each localization
	localizations := make(map[string]string)
	for _, key := range req.Keys {
		// Get localization key
		locKey, err := h.db.GetLocalizationKeyByKey(ctx, key)
		if err != nil {
			continue // Skip missing keys
		}

		// Get localization
		loc, err := h.db.GetLocalizationByKeyAndLanguage(ctx, locKey.ID, lang.ID)
		if err != nil && req.Fallback && defaultLang != nil {
			// Try fallback
			loc, err = h.db.GetLocalizationByKeyAndLanguage(ctx, locKey.ID, defaultLang.ID)
		}

		if err == nil {
			localizations[key] = loc.Value
		}
	}

	response := models.GetBatchLocalizationResponse{
		Language:      req.Language,
		Localizations: localizations,
	}

	c.JSON(http.StatusOK, models.SuccessResponse(response))
}

// ListLanguages retrieves all active languages
func (h *Handler) ListLanguages(c *gin.Context) {
	activeOnly := c.DefaultQuery("active_only", "true") == "true"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	languages, err := h.db.GetLanguages(ctx, activeOnly)
	if err != nil {
		h.logger.Error("failed to get languages", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrCodeInternalError,
			"failed to retrieve languages",
		))
		return
	}

	// Convert to response format
	var langList []models.Language
	for _, lang := range languages {
		langList = append(langList, *lang)
	}

	response := models.LanguageListResponse{
		Languages: langList,
	}

	c.JSON(http.StatusOK, models.SuccessResponse(response))
}

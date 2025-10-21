package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/helixtrack/localization-service/internal/middleware"
	"github.com/helixtrack/localization-service/internal/models"
	"go.uber.org/zap"
)

// CreateLanguage creates a new language (admin only)
func (h *Handler) CreateLanguage(c *gin.Context) {
	var req models.CreateLanguageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrCodeValidationFailed,
			"invalid request body",
		))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	lang := &models.Language{
		Code:       req.Code,
		Name:       req.Name,
		NativeName: req.NativeName,
		IsRTL:      req.IsRTL,
		IsActive:   req.IsActive,
		IsDefault:  req.IsDefault,
	}

	if err := h.db.CreateLanguage(ctx, lang); err != nil {
		h.logger.Error("failed to create language", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrCodeDatabaseError,
			"failed to create language",
		))
		return
	}

	// Audit log
	claims := middleware.GetClaims(c)
	h.db.CreateAuditLog(ctx, "CREATE", "LANGUAGE", lang.ID, claims.Username, lang, c.ClientIP(), c.Request.UserAgent())

	c.JSON(http.StatusCreated, models.SuccessResponse(lang))
}

// UpdateLanguage updates an existing language (admin only)
func (h *Handler) UpdateLanguage(c *gin.Context) {
	id := c.Param("id")

	var req models.CreateLanguageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrCodeValidationFailed,
			"invalid request body",
		))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get existing language
	existing, err := h.db.GetLanguageByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(
			models.ErrCodeNotFound,
			"language not found",
		))
		return
	}

	// Update fields
	existing.Code = req.Code
	existing.Name = req.Name
	existing.NativeName = req.NativeName
	existing.IsRTL = req.IsRTL
	existing.IsActive = req.IsActive
	existing.IsDefault = req.IsDefault

	if err := h.db.UpdateLanguage(ctx, existing); err != nil {
		h.logger.Error("failed to update language", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrCodeDatabaseError,
			"failed to update language",
		))
		return
	}

	// Audit log
	claims := middleware.GetClaims(c)
	h.db.CreateAuditLog(ctx, "UPDATE", "LANGUAGE", existing.ID, claims.Username, existing, c.ClientIP(), c.Request.UserAgent())

	c.JSON(http.StatusOK, models.SuccessResponse(existing))
}

// DeleteLanguage deletes a language (admin only)
func (h *Handler) DeleteLanguage(c *gin.Context) {
	id := c.Param("id")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.db.DeleteLanguage(ctx, id); err != nil {
		h.logger.Error("failed to delete language", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrCodeDatabaseError,
			"failed to delete language",
		))
		return
	}

	// Audit log
	claims := middleware.GetClaims(c)
	h.db.CreateAuditLog(ctx, "DELETE", "LANGUAGE", id, claims.Username, nil, c.ClientIP(), c.Request.UserAgent())

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]string{
		"message": "language deleted successfully",
	}))
}

// CreateLocalization creates or updates a localization (admin only)
func (h *Handler) CreateLocalization(c *gin.Context) {
	var req models.CreateLocalizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrCodeValidationFailed,
			"invalid request body",
		))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get or create localization key
	locKey, err := h.db.GetLocalizationKeyByKey(ctx, req.Key)
	if err != nil {
		// Create new key
		locKey = &models.LocalizationKey{
			Key:         req.Key,
			Category:    req.Category,
			Description: req.Description,
			Context:     req.Context,
		}
		if err := h.db.CreateLocalizationKey(ctx, locKey); err != nil {
			h.logger.Error("failed to create localization key", zap.Error(err))
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(
				models.ErrCodeDatabaseError,
				"failed to create localization key",
			))
			return
		}
	}

	// Get language
	lang, err := h.db.GetLanguageByCode(ctx, req.Language)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(
			models.ErrCodeNotFound,
			"language not found",
		))
		return
	}

	// Prepare JSON fields
	var pluralFormsJSON, variablesJSON json.RawMessage
	if req.PluralForms != nil {
		pluralFormsJSON, _ = json.Marshal(req.PluralForms)
	}
	if req.Variables != nil {
		variablesJSON, _ = json.Marshal(req.Variables)
	}

	// Check if localization already exists
	existing, err := h.db.GetLocalizationByKeyAndLanguage(ctx, locKey.ID, lang.ID)
	if err == nil {
		// Update existing
		existing.Value = req.Value
		existing.PluralForms = pluralFormsJSON
		existing.Variables = variablesJSON
		existing.Approved = req.Approved

		if err := h.db.UpdateLocalization(ctx, existing); err != nil {
			h.logger.Error("failed to update localization", zap.Error(err))
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(
				models.ErrCodeDatabaseError,
				"failed to update localization",
			))
			return
		}

		// Invalidate cache
		h.invalidateCacheForLanguage(ctx, lang.Code)

		// Audit log
		claims := middleware.GetClaims(c)
		h.db.CreateAuditLog(ctx, "UPDATE", "LOCALIZATION", existing.ID, claims.Username, existing, c.ClientIP(), c.Request.UserAgent())

		c.JSON(http.StatusOK, models.SuccessResponse(existing))
		return
	}

	// Create new localization
	loc := &models.Localization{
		KeyID:       locKey.ID,
		LanguageID:  lang.ID,
		Value:       req.Value,
		PluralForms: pluralFormsJSON,
		Variables:   variablesJSON,
		Approved:    req.Approved,
	}

	if err := h.db.CreateLocalization(ctx, loc); err != nil {
		h.logger.Error("failed to create localization", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrCodeDatabaseError,
			"failed to create localization",
		))
		return
	}

	// Invalidate cache
	h.invalidateCacheForLanguage(ctx, lang.Code)

	// Audit log
	claims := middleware.GetClaims(c)
	h.db.CreateAuditLog(ctx, "CREATE", "LOCALIZATION", loc.ID, claims.Username, loc, c.ClientIP(), c.Request.UserAgent())

	c.JSON(http.StatusCreated, models.SuccessResponse(loc))
}

// UpdateLocalization updates an existing localization (admin only)
func (h *Handler) UpdateLocalization(c *gin.Context) {
	id := c.Param("id")

	var req models.CreateLocalizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrCodeValidationFailed,
			"invalid request body",
		))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get existing localization
	existing, err := h.db.GetLocalizationByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(
			models.ErrCodeNotFound,
			"localization not found",
		))
		return
	}

	// Update fields
	existing.Value = req.Value

	var pluralFormsJSON, variablesJSON json.RawMessage
	if req.PluralForms != nil {
		pluralFormsJSON, _ = json.Marshal(req.PluralForms)
		existing.PluralForms = pluralFormsJSON
	}
	if req.Variables != nil {
		variablesJSON, _ = json.Marshal(req.Variables)
		existing.Variables = variablesJSON
	}

	if err := h.db.UpdateLocalization(ctx, existing); err != nil {
		h.logger.Error("failed to update localization", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrCodeDatabaseError,
			"failed to update localization",
		))
		return
	}

	// Invalidate cache
	lang, _ := h.db.GetLanguageByID(ctx, existing.LanguageID)
	if lang != nil {
		h.invalidateCacheForLanguage(ctx, lang.Code)
	}

	// Audit log
	claims := middleware.GetClaims(c)
	h.db.CreateAuditLog(ctx, "UPDATE", "LOCALIZATION", existing.ID, claims.Username, existing, c.ClientIP(), c.Request.UserAgent())

	c.JSON(http.StatusOK, models.SuccessResponse(existing))
}

// DeleteLocalization deletes a localization (admin only)
func (h *Handler) DeleteLocalization(c *gin.Context) {
	id := c.Param("id")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get localization to find language for cache invalidation
	loc, err := h.db.GetLocalizationByID(ctx, id)
	if err == nil {
		lang, _ := h.db.GetLanguageByID(ctx, loc.LanguageID)
		if lang != nil {
			h.invalidateCacheForLanguage(ctx, lang.Code)
		}
	}

	if err := h.db.DeleteLocalization(ctx, id); err != nil {
		h.logger.Error("failed to delete localization", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrCodeDatabaseError,
			"failed to delete localization",
		))
		return
	}

	// Audit log
	claims := middleware.GetClaims(c)
	h.db.CreateAuditLog(ctx, "DELETE", "LOCALIZATION", id, claims.Username, nil, c.ClientIP(), c.Request.UserAgent())

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]string{
		"message": "localization deleted successfully",
	}))
}

// ApproveLocalization approves a localization (admin only)
func (h *Handler) ApproveLocalization(c *gin.Context) {
	id := c.Param("id")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	claims := middleware.GetClaims(c)
	if err := h.db.ApproveLocalization(ctx, id, claims.Username); err != nil {
		h.logger.Error("failed to approve localization", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrCodeDatabaseError,
			"failed to approve localization",
		))
		return
	}

	// Audit log
	h.db.CreateAuditLog(ctx, "APPROVE", "LOCALIZATION", id, claims.Username, nil, c.ClientIP(), c.Request.UserAgent())

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]string{
		"message": "localization approved successfully",
	}))
}

// InvalidateCache invalidates cache (admin only)
func (h *Handler) InvalidateCache(c *gin.Context) {
	var req models.CacheInvalidationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// If no body, invalidate all
		req = models.CacheInvalidationRequest{}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if req.Language != "" {
		h.invalidateCacheForLanguage(ctx, req.Language)
	} else {
		// Invalidate all
		h.cache.DeletePattern(ctx, "l10n:*")
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]string{
		"message": "cache invalidated successfully",
	}))
}

// GetStats retrieves database statistics (admin only)
func (h *Handler) GetStats(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stats, err := h.db.GetStats(ctx)
	if err != nil {
		h.logger.Error("failed to get stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrCodeInternalError,
			"failed to retrieve statistics",
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(stats))
}

// invalidateCacheForLanguage invalidates all cache entries for a language
func (h *Handler) invalidateCacheForLanguage(ctx context.Context, languageCode string) {
	pattern := "l10n:catalog:" + languageCode + ":*"
	h.cache.DeletePattern(ctx, pattern)
	h.logger.Info("cache invalidated", zap.String("language", languageCode))
}

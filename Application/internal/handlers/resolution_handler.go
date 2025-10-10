package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"helixtrack.ru/core/internal/logger"
	"helixtrack.ru/core/internal/middleware"
	"helixtrack.ru/core/internal/models"
)

// handleResolutionCreate creates a new resolution
func (h *Handler) handleResolutionCreate(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "resolution", models.PermissionCreate)
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

	// Parse resolution data from request
	title, ok := req.Data["title"].(string)
	if !ok || title == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing title",
			"",
		))
		return
	}

	resolution := &models.Resolution{
		ID:          uuid.New().String(),
		Title:       title,
		Description: getStringFromData(req.Data, "description"),
		Created:     time.Now().Unix(),
		Modified:    time.Now().Unix(),
		Deleted:     false,
	}

	// Insert into database
	query := `
		INSERT INTO resolution (id, title, description, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(c.Request.Context(), query,
		resolution.ID,
		resolution.Title,
		resolution.Description,
		resolution.Created,
		resolution.Modified,
		resolution.Deleted,
	)

	if err != nil {
		logger.Error("Failed to create resolution", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create resolution",
			"",
		))
		return
	}

	logger.Info("Resolution created",
		zap.String("resolution_id", resolution.ID),
		zap.String("title", resolution.Title),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"resolution": resolution,
	})
	c.JSON(http.StatusCreated, response)
}

// handleResolutionRead reads a single resolution by ID
func (h *Handler) handleResolutionRead(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get resolution ID from request
	resolutionID, ok := req.Data["id"].(string)
	if !ok || resolutionID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing resolution ID",
			"",
		))
		return
	}

	// Query resolution from database
	query := `
		SELECT id, title, description, created, modified, deleted
		FROM resolution
		WHERE id = ? AND deleted = 0
	`

	var resolution models.Resolution
	err := h.db.QueryRow(c.Request.Context(), query, resolutionID).Scan(
		&resolution.ID,
		&resolution.Title,
		&resolution.Description,
		&resolution.Created,
		&resolution.Modified,
		&resolution.Deleted,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeNotFound,
			"Resolution not found",
			"",
		))
		return
	}

	if err != nil {
		logger.Error("Failed to read resolution", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to read resolution",
			"",
		))
		return
	}

	logger.Info("Resolution read",
		zap.String("resolution_id", resolution.ID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"resolution": resolution,
	})
	c.JSON(http.StatusOK, response)
}

// handleResolutionList lists all resolutions
func (h *Handler) handleResolutionList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Query all non-deleted resolutions ordered by title
	query := `
		SELECT id, title, description, created, modified, deleted
		FROM resolution
		WHERE deleted = 0
		ORDER BY title ASC
	`

	rows, err := h.db.Query(c.Request.Context(), query)
	if err != nil {
		logger.Error("Failed to list resolutions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list resolutions",
			"",
		))
		return
	}
	defer rows.Close()

	resolutions := make([]models.Resolution, 0)
	for rows.Next() {
		var resolution models.Resolution
		err := rows.Scan(
			&resolution.ID,
			&resolution.Title,
			&resolution.Description,
			&resolution.Created,
			&resolution.Modified,
			&resolution.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan resolution", zap.Error(err))
			continue
		}
		resolutions = append(resolutions, resolution)
	}

	logger.Info("Resolutions listed",
		zap.Int("count", len(resolutions)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"resolutions": resolutions,
		"count":       len(resolutions),
	})
	c.JSON(http.StatusOK, response)
}

// handleResolutionModify updates an existing resolution
func (h *Handler) handleResolutionModify(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "resolution", models.PermissionUpdate)
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

	// Get resolution ID
	resolutionID, ok := req.Data["id"].(string)
	if !ok || resolutionID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing resolution ID",
			"",
		))
		return
	}

	// Check if resolution exists
	checkQuery := `SELECT COUNT(*) FROM resolution WHERE id = ? AND deleted = 0`
	var count int
	err = h.db.QueryRow(c.Request.Context(), checkQuery, resolutionID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeNotFound,
			"Resolution not found",
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
		updates["description"] = description
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
	query := "UPDATE resolution SET "
	args := make([]interface{}, 0)
	first := true

	for key, value := range updates {
		if !first {
			query += ", "
		}
		query += key + " = ?"
		args = append(args, value)
		first = false
	}

	query += " WHERE id = ?"
	args = append(args, resolutionID)

	_, err = h.db.Exec(c.Request.Context(), query, args...)
	if err != nil {
		logger.Error("Failed to update resolution", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to update resolution",
			"",
		))
		return
	}

	logger.Info("Resolution updated",
		zap.String("resolution_id", resolutionID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"updated": true,
		"id":      resolutionID,
	})
	c.JSON(http.StatusOK, response)
}

// handleResolutionRemove soft-deletes a resolution
func (h *Handler) handleResolutionRemove(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "resolution", models.PermissionDelete)
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

	// Get resolution ID
	resolutionID, ok := req.Data["id"].(string)
	if !ok || resolutionID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing resolution ID",
			"",
		))
		return
	}

	// Soft delete the resolution
	query := `UPDATE resolution SET deleted = 1, modified = ? WHERE id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), resolutionID)
	if err != nil {
		logger.Error("Failed to delete resolution", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to delete resolution",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeNotFound,
			"Resolution not found",
			"",
		))
		return
	}

	logger.Info("Resolution deleted",
		zap.String("resolution_id", resolutionID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"deleted": true,
		"id":      resolutionID,
	})
	c.JSON(http.StatusOK, response)
}

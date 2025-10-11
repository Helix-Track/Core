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

// handlePriorityCreate creates a new priority
func (h *Handler) handlePriorityCreate(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "priority", models.PermissionCreate)
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

	// Parse priority data from request
	title, ok := req.Data["title"].(string)
	if !ok || title == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing title",
			"",
		))
		return
	}

	level, ok := req.Data["level"].(float64) // JSON numbers are float64
	if !ok {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing level",
			"",
		))
		return
	}

	priority := &models.Priority{
		ID:          uuid.New().String(),
		Title:       title,
		Description: getStringFromData(req.Data, "description"),
		Level:       int(level),
		Icon:        getStringFromData(req.Data, "icon"),
		Color:       getStringFromData(req.Data, "color"),
		Created:     time.Now().Unix(),
		Modified:    time.Now().Unix(),
		Deleted:     false,
	}

	// Validate priority level
	if !priority.IsValidLevel() {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidData,
			"Invalid priority level (must be 1-5)",
			"",
		))
		return
	}

	// Insert into database
	query := `
		INSERT INTO priority (id, title, description, level, icon, color, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(c.Request.Context(), query,
		priority.ID,
		priority.Title,
		priority.Description,
		priority.Level,
		priority.Icon,
		priority.Color,
		priority.Created,
		priority.Modified,
		priority.Deleted,
	)

	if err != nil {
		logger.Error("Failed to create priority", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create priority",
			"",
		))
		return
	}

	logger.Info("Priority created",
		zap.String("priority_id", priority.ID),
		zap.String("title", priority.Title),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"priority": priority,
	})
	c.JSON(http.StatusCreated, response)
}

// handlePriorityRead reads a single priority by ID
func (h *Handler) handlePriorityRead(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get priority ID from request
	priorityID, ok := req.Data["id"].(string)
	if !ok || priorityID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing priority ID",
			"",
		))
		return
	}

	// Query priority from database
	query := `
		SELECT id, title, description, level, icon, color, created, modified, deleted
		FROM priority
		WHERE id = ? AND deleted = 0
	`

	var priority models.Priority
	err := h.db.QueryRow(c.Request.Context(), query, priorityID).Scan(
		&priority.ID,
		&priority.Title,
		&priority.Description,
		&priority.Level,
		&priority.Icon,
		&priority.Color,
		&priority.Created,
		&priority.Modified,
		&priority.Deleted,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Priority not found",
			"",
		))
		return
	}

	if err != nil {
		logger.Error("Failed to read priority", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to read priority",
			"",
		))
		return
	}

	logger.Info("Priority read",
		zap.String("priority_id", priority.ID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"priority": priority,
	})
	c.JSON(http.StatusOK, response)
}

// handlePriorityList lists all priorities
func (h *Handler) handlePriorityList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Query all non-deleted priorities ordered by level
	query := `
		SELECT id, title, description, level, icon, color, created, modified, deleted
		FROM priority
		WHERE deleted = 0
		ORDER BY level ASC
	`

	rows, err := h.db.Query(c.Request.Context(), query)
	if err != nil {
		logger.Error("Failed to list priorities", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list priorities",
			"",
		))
		return
	}
	defer rows.Close()

	priorities := make([]models.Priority, 0)
	for rows.Next() {
		var priority models.Priority
		err := rows.Scan(
			&priority.ID,
			&priority.Title,
			&priority.Description,
			&priority.Level,
			&priority.Icon,
			&priority.Color,
			&priority.Created,
			&priority.Modified,
			&priority.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan priority", zap.Error(err))
			continue
		}
		priorities = append(priorities, priority)
	}

	logger.Info("Priorities listed",
		zap.Int("count", len(priorities)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"priorities": priorities,
		"count":      len(priorities),
	})
	c.JSON(http.StatusOK, response)
}

// handlePriorityModify updates an existing priority
func (h *Handler) handlePriorityModify(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "priority", models.PermissionUpdate)
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

	// Get priority ID
	priorityID, ok := req.Data["id"].(string)
	if !ok || priorityID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing priority ID",
			"",
		))
		return
	}

	// Check if priority exists
	checkQuery := `SELECT COUNT(*) FROM priority WHERE id = ? AND deleted = 0`
	var count int
	err = h.db.QueryRow(c.Request.Context(), checkQuery, priorityID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Priority not found",
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
	if level, ok := req.Data["level"].(float64); ok {
		levelInt := int(level)
		if levelInt < 1 || levelInt > 5 {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				models.ErrorCodeInvalidData,
				"Invalid priority level (must be 1-5)",
				"",
			))
			return
		}
		updates["level"] = levelInt
	}
	if icon, ok := req.Data["icon"].(string); ok {
		updates["icon"] = icon
	}
	if color, ok := req.Data["color"].(string); ok {
		updates["color"] = color
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
	query := "UPDATE priority SET "
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
	args = append(args, priorityID)

	_, err = h.db.Exec(c.Request.Context(), query, args...)
	if err != nil {
		logger.Error("Failed to update priority", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to update priority",
			"",
		))
		return
	}

	logger.Info("Priority updated",
		zap.String("priority_id", priorityID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"updated": true,
		"id":      priorityID,
	})
	c.JSON(http.StatusOK, response)
}

// handlePriorityRemove soft-deletes a priority
func (h *Handler) handlePriorityRemove(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "priority", models.PermissionDelete)
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

	// Get priority ID
	priorityID, ok := req.Data["id"].(string)
	if !ok || priorityID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing priority ID",
			"",
		))
		return
	}

	// Soft delete the priority
	query := `UPDATE priority SET deleted = 1, modified = ? WHERE id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), priorityID)
	if err != nil {
		logger.Error("Failed to delete priority", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to delete priority",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Priority not found",
			"",
		))
		return
	}

	logger.Info("Priority deleted",
		zap.String("priority_id", priorityID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"deleted": true,
		"id":      priorityID,
	})
	c.JSON(http.StatusOK, response)
}

// Helper function to safely get string from map
func getStringFromData(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

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

// handleWorkflowCreate creates a new workflow
func (h *Handler) handleWorkflowCreate(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "workflow", models.PermissionCreate)
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

	// Parse workflow data from request
	title, ok := req.Data["title"].(string)
	if !ok || title == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing title",
			"",
		))
		return
	}

	workflow := &models.Workflow{
		ID:          uuid.New().String(),
		Title:       title,
		Description: getStringFromData(req.Data, "description"),
		Created:     time.Now().Unix(),
		Modified:    time.Now().Unix(),
		Deleted:     false,
	}

	// Insert into database
	query := `
		INSERT INTO workflow (id, title, description, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(c.Request.Context(), query,
		workflow.ID,
		workflow.Title,
		workflow.Description,
		workflow.Created,
		workflow.Modified,
		workflow.Deleted,
	)

	if err != nil {
		logger.Error("Failed to create workflow", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create workflow",
			"",
		))
		return
	}

	logger.Info("Workflow created",
		zap.String("workflow_id", workflow.ID),
		zap.String("title", workflow.Title),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"workflow": workflow,
	})
	c.JSON(http.StatusCreated, response)
}

// handleWorkflowRead reads a single workflow by ID
func (h *Handler) handleWorkflowRead(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get workflow ID from request
	workflowID, ok := req.Data["id"].(string)
	if !ok || workflowID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing workflow ID",
			"",
		))
		return
	}

	// Query workflow from database
	query := `
		SELECT id, title, description, created, modified, deleted
		FROM workflow
		WHERE id = ? AND deleted = 0
	`

	var workflow models.Workflow
	err := h.db.QueryRow(c.Request.Context(), query, workflowID).Scan(
		&workflow.ID,
		&workflow.Title,
		&workflow.Description,
		&workflow.Created,
		&workflow.Modified,
		&workflow.Deleted,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Workflow not found",
			"",
		))
		return
	}

	if err != nil {
		logger.Error("Failed to read workflow", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to read workflow",
			"",
		))
		return
	}

	logger.Info("Workflow read",
		zap.String("workflow_id", workflow.ID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"workflow": workflow,
	})
	c.JSON(http.StatusOK, response)
}

// handleWorkflowList lists all workflows
func (h *Handler) handleWorkflowList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Query all non-deleted workflows ordered by title
	query := `
		SELECT id, title, description, created, modified, deleted
		FROM workflow
		WHERE deleted = 0
		ORDER BY title ASC
	`

	rows, err := h.db.Query(c.Request.Context(), query)
	if err != nil {
		logger.Error("Failed to list workflows", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list workflows",
			"",
		))
		return
	}
	defer rows.Close()

	workflows := make([]models.Workflow, 0)
	for rows.Next() {
		var workflow models.Workflow
		err := rows.Scan(
			&workflow.ID,
			&workflow.Title,
			&workflow.Description,
			&workflow.Created,
			&workflow.Modified,
			&workflow.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan workflow", zap.Error(err))
			continue
		}
		workflows = append(workflows, workflow)
	}

	logger.Info("Workflows listed",
		zap.Int("count", len(workflows)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"workflows": workflows,
		"count":     len(workflows),
	})
	c.JSON(http.StatusOK, response)
}

// handleWorkflowModify updates an existing workflow
func (h *Handler) handleWorkflowModify(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "workflow", models.PermissionUpdate)
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

	// Get workflow ID
	workflowID, ok := req.Data["id"].(string)
	if !ok || workflowID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing workflow ID",
			"",
		))
		return
	}

	// Check if workflow exists
	checkQuery := `SELECT COUNT(*) FROM workflow WHERE id = ? AND deleted = 0`
	var count int
	err = h.db.QueryRow(c.Request.Context(), checkQuery, workflowID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Workflow not found",
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
	query := "UPDATE workflow SET "
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
	args = append(args, workflowID)

	_, err = h.db.Exec(c.Request.Context(), query, args...)
	if err != nil {
		logger.Error("Failed to update workflow", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to update workflow",
			"",
		))
		return
	}

	logger.Info("Workflow updated",
		zap.String("workflow_id", workflowID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"updated": true,
		"id":      workflowID,
	})
	c.JSON(http.StatusOK, response)
}

// handleWorkflowRemove soft-deletes a workflow
func (h *Handler) handleWorkflowRemove(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "workflow", models.PermissionDelete)
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

	// Get workflow ID
	workflowID, ok := req.Data["id"].(string)
	if !ok || workflowID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing workflow ID",
			"",
		))
		return
	}

	// Soft delete the workflow
	query := `UPDATE workflow SET deleted = 1, modified = ? WHERE id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), workflowID)
	if err != nil {
		logger.Error("Failed to delete workflow", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to delete workflow",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Workflow not found",
			"",
		))
		return
	}

	logger.Info("Workflow deleted",
		zap.String("workflow_id", workflowID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"deleted": true,
		"id":      workflowID,
	})
	c.JSON(http.StatusOK, response)
}

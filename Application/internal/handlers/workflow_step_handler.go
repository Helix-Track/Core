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

// handleWorkflowStepCreate creates a new workflow step
func (h *Handler) handleWorkflowStepCreate(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "workflow_step", models.PermissionCreate)
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

	// Parse workflow step data from request
	workflowID, ok := req.Data["workflowId"].(string)
	if !ok || workflowID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing workflowId",
			"",
		))
		return
	}

	statusID, ok := req.Data["statusId"].(string)
	if !ok || statusID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing statusId",
			"",
		))
		return
	}

	position, ok := req.Data["position"].(float64) // JSON numbers are float64
	if !ok {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing position",
			"",
		))
		return
	}

	workflowStep := &models.WorkflowStep{
		ID:         uuid.New().String(),
		WorkflowID: workflowID,
		StatusID:   statusID,
		Position:   int(position),
		Created:    time.Now().Unix(),
		Modified:   time.Now().Unix(),
		Deleted:    false,
	}

	// Insert into database
	query := `
		INSERT INTO workflow_step (id, workflow_id, status_id, position, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(c.Request.Context(), query,
		workflowStep.ID,
		workflowStep.WorkflowID,
		workflowStep.StatusID,
		workflowStep.Position,
		workflowStep.Created,
		workflowStep.Modified,
		workflowStep.Deleted,
	)

	if err != nil {
		logger.Error("Failed to create workflow step", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create workflow step",
			"",
		))
		return
	}

	logger.Info("Workflow step created",
		zap.String("workflow_step_id", workflowStep.ID),
		zap.String("workflow_id", workflowStep.WorkflowID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"workflowStep": workflowStep,
	})
	c.JSON(http.StatusCreated, response)
}

// handleWorkflowStepRead reads a single workflow step by ID
func (h *Handler) handleWorkflowStepRead(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get workflow step ID from request
	workflowStepID, ok := req.Data["id"].(string)
	if !ok || workflowStepID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing workflow step ID",
			"",
		))
		return
	}

	// Query workflow step from database
	query := `
		SELECT id, workflow_id, status_id, position, created, modified, deleted
		FROM workflow_step
		WHERE id = ? AND deleted = 0
	`

	var workflowStep models.WorkflowStep
	err := h.db.QueryRow(c.Request.Context(), query, workflowStepID).Scan(
		&workflowStep.ID,
		&workflowStep.WorkflowID,
		&workflowStep.StatusID,
		&workflowStep.Position,
		&workflowStep.Created,
		&workflowStep.Modified,
		&workflowStep.Deleted,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Workflow step not found",
			"",
		))
		return
	}

	if err != nil {
		logger.Error("Failed to read workflow step", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to read workflow step",
			"",
		))
		return
	}

	logger.Info("Workflow step read",
		zap.String("workflow_step_id", workflowStep.ID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"workflowStep": workflowStep,
	})
	c.JSON(http.StatusOK, response)
}

// handleWorkflowStepList lists all workflow steps for a workflow
func (h *Handler) handleWorkflowStepList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get optional workflow ID filter
	workflowID := getStringFromData(req.Data, "workflowId")

	// Query workflow steps
	var query string
	var args []interface{}

	if workflowID != "" {
		query = `
			SELECT id, workflow_id, status_id, position, created, modified, deleted
			FROM workflow_step
			WHERE workflow_id = ? AND deleted = 0
			ORDER BY position ASC
		`
		args = append(args, workflowID)
	} else {
		query = `
			SELECT id, workflow_id, status_id, position, created, modified, deleted
			FROM workflow_step
			WHERE deleted = 0
			ORDER BY workflow_id, position ASC
		`
	}

	rows, err := h.db.Query(c.Request.Context(), query, args...)
	if err != nil {
		logger.Error("Failed to list workflow steps", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list workflow steps",
			"",
		))
		return
	}
	defer rows.Close()

	workflowSteps := make([]models.WorkflowStep, 0)
	for rows.Next() {
		var workflowStep models.WorkflowStep
		err := rows.Scan(
			&workflowStep.ID,
			&workflowStep.WorkflowID,
			&workflowStep.StatusID,
			&workflowStep.Position,
			&workflowStep.Created,
			&workflowStep.Modified,
			&workflowStep.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan workflow step", zap.Error(err))
			continue
		}
		workflowSteps = append(workflowSteps, workflowStep)
	}

	logger.Info("Workflow steps listed",
		zap.Int("count", len(workflowSteps)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"workflowSteps": workflowSteps,
		"count":         len(workflowSteps),
	})
	c.JSON(http.StatusOK, response)
}

// handleWorkflowStepModify updates an existing workflow step
func (h *Handler) handleWorkflowStepModify(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "workflow_step", models.PermissionUpdate)
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

	// Get workflow step ID
	workflowStepID, ok := req.Data["id"].(string)
	if !ok || workflowStepID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing workflow step ID",
			"",
		))
		return
	}

	// Check if workflow step exists
	checkQuery := `SELECT COUNT(*) FROM workflow_step WHERE id = ? AND deleted = 0`
	var count int
	err = h.db.QueryRow(c.Request.Context(), checkQuery, workflowStepID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Workflow step not found",
			"",
		))
		return
	}

	// Build update query dynamically based on provided fields
	updates := make(map[string]interface{})

	if statusID, ok := req.Data["statusId"].(string); ok && statusID != "" {
		updates["status_id"] = statusID
	}
	if position, ok := req.Data["position"].(float64); ok {
		updates["position"] = int(position)
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
	query := "UPDATE workflow_step SET "
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
	args = append(args, workflowStepID)

	_, err = h.db.Exec(c.Request.Context(), query, args...)
	if err != nil {
		logger.Error("Failed to update workflow step", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to update workflow step",
			"",
		))
		return
	}

	logger.Info("Workflow step updated",
		zap.String("workflow_step_id", workflowStepID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"updated": true,
		"id":      workflowStepID,
	})
	c.JSON(http.StatusOK, response)
}

// handleWorkflowStepRemove soft-deletes a workflow step
func (h *Handler) handleWorkflowStepRemove(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "workflow_step", models.PermissionDelete)
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

	// Get workflow step ID
	workflowStepID, ok := req.Data["id"].(string)
	if !ok || workflowStepID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing workflow step ID",
			"",
		))
		return
	}

	// Soft delete the workflow step
	query := `UPDATE workflow_step SET deleted = 1, modified = ? WHERE id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), workflowStepID)
	if err != nil {
		logger.Error("Failed to delete workflow step", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to delete workflow step",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Workflow step not found",
			"",
		))
		return
	}

	logger.Info("Workflow step deleted",
		zap.String("workflow_step_id", workflowStepID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"deleted": true,
		"id":      workflowStepID,
	})
	c.JSON(http.StatusOK, response)
}

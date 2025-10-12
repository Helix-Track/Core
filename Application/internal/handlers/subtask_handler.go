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
	"helixtrack.ru/core/internal/websocket"
)

// handleSubtaskCreate creates a new subtask (ticket with parent reference)
func (h *Handler) handleSubtaskCreate(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "ticket", models.PermissionCreate)
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

	// Get parent ticket ID
	parentTicketID, ok := req.Data["parentTicketId"].(string)
	if !ok || parentTicketID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing parentTicketId",
			"",
		))
		return
	}

	// Verify parent ticket exists and is not a subtask itself
	checkQuery := `SELECT COUNT(*) FROM ticket WHERE id = ? AND deleted = 0 AND is_subtask = 0`
	var count int
	err = h.db.QueryRow(c.Request.Context(), checkQuery, parentTicketID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Parent ticket not found or is a subtask",
			"",
		))
		return
	}

	// Get subtask details
	title, ok := req.Data["title"].(string)
	if !ok || title == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing title",
			"",
		))
		return
	}

	description := getStringFromData(req.Data, "description")
	projectID := getStringFromData(req.Data, "projectId")

	// If project ID not specified, get it from parent
	if projectID == "" {
		err = h.db.QueryRow(c.Request.Context(),
			"SELECT project_id FROM ticket WHERE id = ?",
			parentTicketID).Scan(&projectID)
		if err != nil {
			logger.Error("Failed to get parent project", zap.Error(err))
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
				models.ErrorCodeInternalError,
				"Failed to get parent project",
				"",
			))
			return
		}
	}

	// Create subtask (new ticket with is_subtask=true)
	subtaskID := uuid.New().String()
	now := time.Now().Unix()

	query := `
		INSERT INTO ticket (
			id, title, description, project_id, parent_ticket_id,
			is_subtask, created_by, created, modified, deleted
		) VALUES (?, ?, ?, ?, ?, 1, ?, ?, ?, 0)
	`

	_, err = h.db.Exec(c.Request.Context(), query,
		subtaskID,
		title,
		description,
		projectID,
		parentTicketID,
		username,
		now,
		now,
	)

	if err != nil {
		logger.Error("Failed to create subtask", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create subtask",
			"",
		))
		return
	}

	logger.Info("Subtask created",
		zap.String("subtask_id", subtaskID),
		zap.String("parent_ticket_id", parentTicketID),
		zap.String("title", title),
		zap.String("username", username),
	)

	// Publish event
	h.publisher.PublishEntityEvent(
		"subtaskCreate",
		"subtask",
		subtaskID,
		username,
		map[string]interface{}{
			"id":             subtaskID,
			"parentTicketId": parentTicketID,
			"title":          title,
			"projectId":      projectID,
		},
		websocket.NewProjectContext(projectID, []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"id":             subtaskID,
		"parentTicketId": parentTicketID,
		"title":          title,
		"projectId":      projectID,
		"isSubtask":      true,
	})
	c.JSON(http.StatusCreated, response)
}

// handleSubtaskList lists all subtasks in the system
func (h *Handler) handleSubtaskList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Query all subtasks
	query := `
		SELECT id, title, parent_ticket_id, project_id, status_id, created, modified
		FROM ticket
		WHERE deleted = 0 AND is_subtask = 1
		ORDER BY created DESC
	`

	rows, err := h.db.Query(c.Request.Context(), query)
	if err != nil {
		logger.Error("Failed to list subtasks", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list subtasks",
			"",
		))
		return
	}
	defer rows.Close()

	subtasks := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, title string
		var parentTicketID, projectID, statusID sql.NullString
		var created, modified int64

		err := rows.Scan(&id, &title, &parentTicketID, &projectID, &statusID, &created, &modified)
		if err != nil {
			logger.Error("Failed to scan subtask", zap.Error(err))
			continue
		}

		subtasks = append(subtasks, map[string]interface{}{
			"id":             id,
			"title":          title,
			"parentTicketId": parentTicketID.String,
			"projectId":      projectID.String,
			"statusId":       statusID.String,
			"created":        created,
			"modified":       modified,
		})
	}

	logger.Info("Subtasks listed",
		zap.Int("count", len(subtasks)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"subtasks": subtasks,
		"count":    len(subtasks),
	})
	c.JSON(http.StatusOK, response)
}

// handleSubtaskMoveToParent moves a subtask to a different parent ticket
func (h *Handler) handleSubtaskMoveToParent(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "ticket", models.PermissionUpdate)
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

	// Get subtask ID
	subtaskID, ok := req.Data["subtaskId"].(string)
	if !ok || subtaskID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing subtaskId",
			"",
		))
		return
	}

	// Get new parent ticket ID
	newParentTicketID, ok := req.Data["newParentTicketId"].(string)
	if !ok || newParentTicketID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing newParentTicketId",
			"",
		))
		return
	}

	// Verify subtask exists and is actually a subtask
	checkQuery := `SELECT COUNT(*) FROM ticket WHERE id = ? AND deleted = 0 AND is_subtask = 1`
	var count int
	err = h.db.QueryRow(c.Request.Context(), checkQuery, subtaskID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Subtask not found",
			"",
		))
		return
	}

	// Verify new parent exists and is not a subtask itself
	checkQuery = `SELECT COUNT(*) FROM ticket WHERE id = ? AND deleted = 0 AND is_subtask = 0`
	err = h.db.QueryRow(c.Request.Context(), checkQuery, newParentTicketID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"New parent ticket not found or is a subtask",
			"",
		))
		return
	}

	// Update parent ticket ID
	query := `UPDATE ticket SET parent_ticket_id = ?, modified = ? WHERE id = ?`
	_, err = h.db.Exec(c.Request.Context(), query, newParentTicketID, time.Now().Unix(), subtaskID)
	if err != nil {
		logger.Error("Failed to move subtask", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to move subtask",
			"",
		))
		return
	}

	logger.Info("Subtask moved to new parent",
		zap.String("subtask_id", subtaskID),
		zap.String("new_parent_ticket_id", newParentTicketID),
		zap.String("username", username),
	)

	// Publish event
	h.publisher.PublishEntityEvent(
		"subtaskMoveToParent",
		"subtask",
		subtaskID,
		username,
		map[string]interface{}{
			"subtaskId":        subtaskID,
			"newParentTicketId": newParentTicketID,
		},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"moved":            true,
		"subtaskId":        subtaskID,
		"newParentTicketId": newParentTicketID,
	})
	c.JSON(http.StatusOK, response)
}

// handleSubtaskConvertToIssue converts a subtask to a regular ticket
func (h *Handler) handleSubtaskConvertToIssue(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "ticket", models.PermissionUpdate)
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

	// Get subtask ID
	subtaskID, ok := req.Data["subtaskId"].(string)
	if !ok || subtaskID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing subtaskId",
			"",
		))
		return
	}

	// Verify subtask exists and is actually a subtask
	checkQuery := `SELECT COUNT(*) FROM ticket WHERE id = ? AND deleted = 0 AND is_subtask = 1`
	var count int
	err = h.db.QueryRow(c.Request.Context(), checkQuery, subtaskID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Subtask not found",
			"",
		))
		return
	}

	// Convert to regular issue by clearing is_subtask and parent_ticket_id
	query := `
		UPDATE ticket
		SET is_subtask = 0, parent_ticket_id = NULL, modified = ?
		WHERE id = ?
	`

	_, err = h.db.Exec(c.Request.Context(), query, time.Now().Unix(), subtaskID)
	if err != nil {
		logger.Error("Failed to convert subtask to issue", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to convert subtask to issue",
			"",
		))
		return
	}

	logger.Info("Subtask converted to issue",
		zap.String("ticket_id", subtaskID),
		zap.String("username", username),
	)

	// Publish event
	h.publisher.PublishEntityEvent(
		"subtaskConvertToIssue",
		"ticket",
		subtaskID,
		username,
		map[string]interface{}{
			"ticketId":  subtaskID,
			"isSubtask": false,
		},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"converted": true,
		"ticketId":  subtaskID,
		"isSubtask": false,
	})
	c.JSON(http.StatusOK, response)
}

// handleSubtaskListByParent lists all subtasks of a specific parent ticket
func (h *Handler) handleSubtaskListByParent(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get parent ticket ID
	parentTicketID, ok := req.Data["parentTicketId"].(string)
	if !ok || parentTicketID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing parentTicketId",
			"",
		))
		return
	}

	// Query subtasks for this parent
	query := `
		SELECT id, title, description, status_id, assignee_id, created, modified
		FROM ticket
		WHERE parent_ticket_id = ? AND deleted = 0 AND is_subtask = 1
		ORDER BY created ASC
	`

	rows, err := h.db.Query(c.Request.Context(), query, parentTicketID)
	if err != nil {
		logger.Error("Failed to list subtasks by parent", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list subtasks",
			"",
		))
		return
	}
	defer rows.Close()

	subtasks := make([]map[string]interface{}, 0)
	completedCount := 0

	for rows.Next() {
		var id, title string
		var description, statusID, assigneeID sql.NullString
		var created, modified int64

		err := rows.Scan(&id, &title, &description, &statusID, &assigneeID, &created, &modified)
		if err != nil {
			logger.Error("Failed to scan subtask", zap.Error(err))
			continue
		}

		// Check if subtask is completed (this is a simplified check - adjust based on your status model)
		// You may need to check against specific status values that indicate completion
		isCompleted := false
		if statusID.Valid {
			// This is a placeholder - you should check if the status represents completion
			// For example, check if status is "Done" or "Closed"
			var statusName string
			statusQuery := `SELECT name FROM ticket_status WHERE id = ?`
			if err := h.db.QueryRow(c.Request.Context(), statusQuery, statusID.String).Scan(&statusName); err == nil {
				if statusName == "Done" || statusName == "Closed" || statusName == "Resolved" {
					isCompleted = true
					completedCount++
				}
			}
		}

		subtasks = append(subtasks, map[string]interface{}{
			"id":          id,
			"title":       title,
			"description": description.String,
			"statusId":    statusID.String,
			"assigneeId":  assigneeID.String,
			"created":     created,
			"modified":    modified,
			"isCompleted": isCompleted,
		})
	}

	// Calculate summary
	totalCount := len(subtasks)
	percentComplete := 0.0
	if totalCount > 0 {
		percentComplete = float64(completedCount) / float64(totalCount) * 100.0
	}

	logger.Info("Subtasks listed by parent",
		zap.String("parent_ticket_id", parentTicketID),
		zap.Int("total_count", totalCount),
		zap.Int("completed_count", completedCount),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"parentTicketId":     parentTicketID,
		"subtasks":           subtasks,
		"totalCount":         totalCount,
		"completedCount":     completedCount,
		"percentComplete":    percentComplete,
	})
	c.JSON(http.StatusOK, response)
}

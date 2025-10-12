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

// handleWorkLogAdd adds a work log entry to a ticket
func (h *Handler) handleWorkLogAdd(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "work_log", models.PermissionCreate)
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

	// Extract and validate data
	ticketID, ok := req.Data["ticketId"].(string)
	if !ok || ticketID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing ticketId",
			"",
		))
		return
	}

	timeSpent, ok := req.Data["timeSpent"].(float64) // JSON numbers are float64
	if !ok || timeSpent <= 0 {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing or invalid timeSpent (in minutes)",
			"",
		))
		return
	}

	workDate, ok := req.Data["workDate"].(float64)
	if !ok {
		workDate = float64(time.Now().Unix())
	}

	userID := username // TODO: Get actual user ID from JWT/session

	// Check if ticket exists
	checkQuery := `SELECT COUNT(*) FROM ticket WHERE id = ? AND deleted = 0`
	var count int
	err = h.db.QueryRow(c.Request.Context(), checkQuery, ticketID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Ticket not found",
			"",
		))
		return
	}

	workLog := &models.WorkLog{
		ID:          uuid.New().String(),
		TicketID:    ticketID,
		UserID:      userID,
		TimeSpent:   int(timeSpent),
		WorkDate:    int64(workDate),
		Description: getStringFromData(req.Data, "description"),
		Created:     time.Now().Unix(),
		Modified:    time.Now().Unix(),
		Deleted:     false,
	}

	// Insert into database
	query := `
		INSERT INTO work_log (id, ticket_id, user_id, time_spent, work_date, description, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(c.Request.Context(), query,
		workLog.ID,
		workLog.TicketID,
		workLog.UserID,
		workLog.TimeSpent,
		workLog.WorkDate,
		workLog.Description,
		workLog.Created,
		workLog.Modified,
		workLog.Deleted,
	)

	if err != nil {
		logger.Error("Failed to add work log", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to add work log",
			"",
		))
		return
	}

	logger.Info("Work log added",
		zap.String("worklog_id", workLog.ID),
		zap.String("ticket_id", ticketID),
		zap.Int("time_spent", workLog.TimeSpent),
		zap.String("username", username),
	)

	// Publish event
	h.publisher.PublishEntityEvent(
		"workLogAdd",
		"work_log",
		workLog.ID,
		username,
		map[string]interface{}{
			"id":          workLog.ID,
			"ticketId":    workLog.TicketID,
			"userId":      workLog.UserID,
			"timeSpent":   workLog.TimeSpent,
			"workDate":    workLog.WorkDate,
			"description": workLog.Description,
		},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"workLog": workLog,
	})
	c.JSON(http.StatusCreated, response)
}

// handleWorkLogModify updates an existing work log entry
func (h *Handler) handleWorkLogModify(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "work_log", models.PermissionUpdate)
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

	// Get work log ID
	workLogID, ok := req.Data["id"].(string)
	if !ok || workLogID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing work log ID",
			"",
		))
		return
	}

	// Check if work log exists
	checkQuery := `SELECT COUNT(*) FROM work_log WHERE id = ? AND deleted = 0`
	var count int
	err = h.db.QueryRow(c.Request.Context(), checkQuery, workLogID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Work log not found",
			"",
		))
		return
	}

	// Build update query dynamically
	updates := make(map[string]interface{})

	if timeSpent, ok := req.Data["timeSpent"].(float64); ok && timeSpent > 0 {
		updates["time_spent"] = int(timeSpent)
	}
	if workDate, ok := req.Data["workDate"].(float64); ok {
		updates["work_date"] = int64(workDate)
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
	query := "UPDATE work_log SET "
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
	args = append(args, workLogID)

	_, err = h.db.Exec(c.Request.Context(), query, args...)
	if err != nil {
		logger.Error("Failed to update work log", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to update work log",
			"",
		))
		return
	}

	logger.Info("Work log updated",
		zap.String("worklog_id", workLogID),
		zap.String("username", username),
	)

	// Publish event
	h.publisher.PublishEntityEvent(
		"workLogModify",
		"work_log",
		workLogID,
		username,
		updates,
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"updated": true,
		"id":      workLogID,
	})
	c.JSON(http.StatusOK, response)
}

// handleWorkLogRemove soft-deletes a work log entry
func (h *Handler) handleWorkLogRemove(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "work_log", models.PermissionDelete)
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

	// Get work log ID
	workLogID, ok := req.Data["id"].(string)
	if !ok || workLogID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing work log ID",
			"",
		))
		return
	}

	// Read work log data before deleting (for event publishing)
	readQuery := `SELECT id, ticket_id, user_id, time_spent FROM work_log WHERE id = ? AND deleted = 0`
	var workLog models.WorkLog
	err = h.db.QueryRow(c.Request.Context(), readQuery, workLogID).Scan(
		&workLog.ID,
		&workLog.TicketID,
		&workLog.UserID,
		&workLog.TimeSpent,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Work log not found",
			"",
		))
		return
	}
	if err != nil {
		logger.Error("Failed to read work log before deletion", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to read work log",
			"",
		))
		return
	}

	// Soft delete the work log
	query := `UPDATE work_log SET deleted = 1, modified = ? WHERE id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), workLogID)
	if err != nil {
		logger.Error("Failed to delete work log", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to delete work log",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Work log not found",
			"",
		))
		return
	}

	logger.Info("Work log deleted",
		zap.String("worklog_id", workLogID),
		zap.String("username", username),
	)

	// Publish event
	h.publisher.PublishEntityEvent(
		"workLogRemove",
		"work_log",
		workLogID,
		username,
		map[string]interface{}{
			"id":        workLog.ID,
			"ticketId":  workLog.TicketID,
			"userId":    workLog.UserID,
			"timeSpent": workLog.TimeSpent,
		},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"deleted": true,
		"id":      workLogID,
	})
	c.JSON(http.StatusOK, response)
}

// handleWorkLogList lists all work log entries (with optional filtering)
func (h *Handler) handleWorkLogList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Query all non-deleted work logs
	query := `
		SELECT id, ticket_id, user_id, time_spent, work_date, description, created, modified
		FROM work_log
		WHERE deleted = 0
		ORDER BY work_date DESC, created DESC
	`

	rows, err := h.db.Query(c.Request.Context(), query)
	if err != nil {
		logger.Error("Failed to list work logs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list work logs",
			"",
		))
		return
	}
	defer rows.Close()

	workLogs := make([]models.WorkLog, 0)
	for rows.Next() {
		var workLog models.WorkLog
		err := rows.Scan(
			&workLog.ID,
			&workLog.TicketID,
			&workLog.UserID,
			&workLog.TimeSpent,
			&workLog.WorkDate,
			&workLog.Description,
			&workLog.Created,
			&workLog.Modified,
		)
		if err != nil {
			logger.Error("Failed to scan work log", zap.Error(err))
			continue
		}
		workLogs = append(workLogs, workLog)
	}

	logger.Info("Work logs listed",
		zap.Int("count", len(workLogs)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"workLogs": workLogs,
		"count":    len(workLogs),
	})
	c.JSON(http.StatusOK, response)
}

// handleWorkLogListByTicket lists all work logs for a specific ticket
func (h *Handler) handleWorkLogListByTicket(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get ticket ID from request
	ticketID, ok := req.Data["ticketId"].(string)
	if !ok || ticketID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing ticketId",
			"",
		))
		return
	}

	// Query work logs for this ticket
	query := `
		SELECT id, ticket_id, user_id, time_spent, work_date, description, created, modified
		FROM work_log
		WHERE ticket_id = ? AND deleted = 0
		ORDER BY work_date DESC, created DESC
	`

	rows, err := h.db.Query(c.Request.Context(), query, ticketID)
	if err != nil {
		logger.Error("Failed to list work logs by ticket", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list work logs",
			"",
		))
		return
	}
	defer rows.Close()

	workLogs := make([]models.WorkLog, 0)
	totalTime := 0
	for rows.Next() {
		var workLog models.WorkLog
		err := rows.Scan(
			&workLog.ID,
			&workLog.TicketID,
			&workLog.UserID,
			&workLog.TimeSpent,
			&workLog.WorkDate,
			&workLog.Description,
			&workLog.Created,
			&workLog.Modified,
		)
		if err != nil {
			logger.Error("Failed to scan work log", zap.Error(err))
			continue
		}
		workLogs = append(workLogs, workLog)
		totalTime += workLog.TimeSpent
	}

	logger.Info("Work logs listed by ticket",
		zap.String("ticket_id", ticketID),
		zap.Int("count", len(workLogs)),
		zap.Int("total_minutes", totalTime),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"ticketId":     ticketID,
		"workLogs":     workLogs,
		"count":        len(workLogs),
		"totalMinutes": totalTime,
		"totalHours":   float64(totalTime) / 60.0,
	})
	c.JSON(http.StatusOK, response)
}

// handleWorkLogListByUser lists all work logs for a specific user
func (h *Handler) handleWorkLogListByUser(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get user ID from request (or use current user)
	userID, ok := req.Data["userId"].(string)
	if !ok || userID == "" {
		userID = username // Default to current user
	}

	// Query work logs for this user
	query := `
		SELECT id, ticket_id, user_id, time_spent, work_date, description, created, modified
		FROM work_log
		WHERE user_id = ? AND deleted = 0
		ORDER BY work_date DESC, created DESC
	`

	rows, err := h.db.Query(c.Request.Context(), query, userID)
	if err != nil {
		logger.Error("Failed to list work logs by user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list work logs",
			"",
		))
		return
	}
	defer rows.Close()

	workLogs := make([]models.WorkLog, 0)
	totalTime := 0
	for rows.Next() {
		var workLog models.WorkLog
		err := rows.Scan(
			&workLog.ID,
			&workLog.TicketID,
			&workLog.UserID,
			&workLog.TimeSpent,
			&workLog.WorkDate,
			&workLog.Description,
			&workLog.Created,
			&workLog.Modified,
		)
		if err != nil {
			logger.Error("Failed to scan work log", zap.Error(err))
			continue
		}
		workLogs = append(workLogs, workLog)
		totalTime += workLog.TimeSpent
	}

	logger.Info("Work logs listed by user",
		zap.String("user_id", userID),
		zap.Int("count", len(workLogs)),
		zap.Int("total_minutes", totalTime),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"userId":       userID,
		"workLogs":     workLogs,
		"count":        len(workLogs),
		"totalMinutes": totalTime,
		"totalHours":   float64(totalTime) / 60.0,
	})
	c.JSON(http.StatusOK, response)
}

// handleWorkLogGetTotalTime gets the total time logged for a ticket
func (h *Handler) handleWorkLogGetTotalTime(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get ticket ID from request
	ticketID, ok := req.Data["ticketId"].(string)
	if !ok || ticketID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing ticketId",
			"",
		))
		return
	}

	// Sum total time for this ticket
	query := `
		SELECT COALESCE(SUM(time_spent), 0)
		FROM work_log
		WHERE ticket_id = ? AND deleted = 0
	`

	var totalMinutes int
	err := h.db.QueryRow(c.Request.Context(), query, ticketID).Scan(&totalMinutes)
	if err != nil {
		logger.Error("Failed to get total time", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to get total time",
			"",
		))
		return
	}

	logger.Info("Total time retrieved",
		zap.String("ticket_id", ticketID),
		zap.Int("total_minutes", totalMinutes),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"ticketId":     ticketID,
		"totalMinutes": totalMinutes,
		"totalHours":   float64(totalMinutes) / 60.0,
		"totalDays":    float64(totalMinutes) / (60.0 * 8.0), // Assuming 8-hour workday
	})
	c.JSON(http.StatusOK, response)
}

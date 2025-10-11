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

// handleTicketStatusCreate creates a new ticket status
func (h *Handler) handleTicketStatusCreate(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "ticket_status", models.PermissionCreate)
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

	// Parse ticket status data from request
	title, ok := req.Data["title"].(string)
	if !ok || title == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing title",
			"",
		))
		return
	}

	ticketStatus := &models.TicketStatus{
		ID:          uuid.New().String(),
		Title:       title,
		Description: getStringFromData(req.Data, "description"),
		Created:     time.Now().Unix(),
		Modified:    time.Now().Unix(),
		Deleted:     false,
	}

	// Insert into database
	query := `
		INSERT INTO ticket_status (id, title, description, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(c.Request.Context(), query,
		ticketStatus.ID,
		ticketStatus.Title,
		ticketStatus.Description,
		ticketStatus.Created,
		ticketStatus.Modified,
		ticketStatus.Deleted,
	)

	if err != nil {
		logger.Error("Failed to create ticket status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create ticket status",
			"",
		))
		return
	}

	logger.Info("Ticket status created",
		zap.String("ticket_status_id", ticketStatus.ID),
		zap.String("title", ticketStatus.Title),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"ticketStatus": ticketStatus,
	})
	c.JSON(http.StatusCreated, response)
}

// handleTicketStatusRead reads a single ticket status by ID
func (h *Handler) handleTicketStatusRead(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get ticket status ID from request
	ticketStatusID, ok := req.Data["id"].(string)
	if !ok || ticketStatusID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing ticket status ID",
			"",
		))
		return
	}

	// Query ticket status from database
	query := `
		SELECT id, title, description, created, modified, deleted
		FROM ticket_status
		WHERE id = ? AND deleted = 0
	`

	var ticketStatus models.TicketStatus
	err := h.db.QueryRow(c.Request.Context(), query, ticketStatusID).Scan(
		&ticketStatus.ID,
		&ticketStatus.Title,
		&ticketStatus.Description,
		&ticketStatus.Created,
		&ticketStatus.Modified,
		&ticketStatus.Deleted,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Ticket status not found",
			"",
		))
		return
	}

	if err != nil {
		logger.Error("Failed to read ticket status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to read ticket status",
			"",
		))
		return
	}

	logger.Info("Ticket status read",
		zap.String("ticket_status_id", ticketStatus.ID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"ticketStatus": ticketStatus,
	})
	c.JSON(http.StatusOK, response)
}

// handleTicketStatusList lists all ticket statuses
func (h *Handler) handleTicketStatusList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Query all non-deleted ticket statuses ordered by title
	query := `
		SELECT id, title, description, created, modified, deleted
		FROM ticket_status
		WHERE deleted = 0
		ORDER BY title ASC
	`

	rows, err := h.db.Query(c.Request.Context(), query)
	if err != nil {
		logger.Error("Failed to list ticket statuses", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list ticket statuses",
			"",
		))
		return
	}
	defer rows.Close()

	ticketStatuses := make([]models.TicketStatus, 0)
	for rows.Next() {
		var ticketStatus models.TicketStatus
		err := rows.Scan(
			&ticketStatus.ID,
			&ticketStatus.Title,
			&ticketStatus.Description,
			&ticketStatus.Created,
			&ticketStatus.Modified,
			&ticketStatus.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan ticket status", zap.Error(err))
			continue
		}
		ticketStatuses = append(ticketStatuses, ticketStatus)
	}

	logger.Info("Ticket statuses listed",
		zap.Int("count", len(ticketStatuses)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"ticketStatuses": ticketStatuses,
		"count":          len(ticketStatuses),
	})
	c.JSON(http.StatusOK, response)
}

// handleTicketStatusModify updates an existing ticket status
func (h *Handler) handleTicketStatusModify(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "ticket_status", models.PermissionUpdate)
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

	// Get ticket status ID
	ticketStatusID, ok := req.Data["id"].(string)
	if !ok || ticketStatusID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing ticket status ID",
			"",
		))
		return
	}

	// Check if ticket status exists
	checkQuery := `SELECT COUNT(*) FROM ticket_status WHERE id = ? AND deleted = 0`
	var count int
	err = h.db.QueryRow(c.Request.Context(), checkQuery, ticketStatusID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Ticket status not found",
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
	query := "UPDATE ticket_status SET "
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
	args = append(args, ticketStatusID)

	_, err = h.db.Exec(c.Request.Context(), query, args...)
	if err != nil {
		logger.Error("Failed to update ticket status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to update ticket status",
			"",
		))
		return
	}

	logger.Info("Ticket status updated",
		zap.String("ticket_status_id", ticketStatusID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"updated": true,
		"id":      ticketStatusID,
	})
	c.JSON(http.StatusOK, response)
}

// handleTicketStatusRemove soft-deletes a ticket status
func (h *Handler) handleTicketStatusRemove(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "ticket_status", models.PermissionDelete)
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

	// Get ticket status ID
	ticketStatusID, ok := req.Data["id"].(string)
	if !ok || ticketStatusID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing ticket status ID",
			"",
		))
		return
	}

	// Soft delete the ticket status
	query := `UPDATE ticket_status SET deleted = 1, modified = ? WHERE id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), ticketStatusID)
	if err != nil {
		logger.Error("Failed to delete ticket status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to delete ticket status",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Ticket status not found",
			"",
		))
		return
	}

	logger.Info("Ticket status deleted",
		zap.String("ticket_status_id", ticketStatusID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"deleted": true,
		"id":      ticketStatusID,
	})
	c.JSON(http.StatusOK, response)
}

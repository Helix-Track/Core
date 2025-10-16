package handlers

import (
	"context"
	"fmt"
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

// handleCreateTicket creates a new ticket
func (h *Handler) handleCreateTicket(c *gin.Context, req *models.Request) {
	ticketData, ok := req.Data["data"].(map[string]interface{})
	if !ok {
		ticketData = req.Data
	}

	projectID, _ := ticketData["project_id"].(string)
	if projectID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing project_id",
			"",
		))
		return
	}

	title, _ := ticketData["title"].(string)
	if title == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing title",
			"",
		))
		return
	}

	description, _ := ticketData["description"].(string)
	ticketTypeStr, _ := ticketData["type"].(string)
	if ticketTypeStr == "" {
		ticketTypeStr = "task"
	}

	priority, _ := ticketData["priority"].(string)

	// Get ticket type ID
	var ticketTypeID string
	err := h.db.QueryRow(context.Background(),
		"SELECT id FROM ticket_type WHERE title = ? AND deleted = 0",
		ticketTypeStr).Scan(&ticketTypeID)

	if err != nil {
		logger.Error("Ticket type not found", zap.Error(err), zap.String("type", ticketTypeStr))
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidData,
			"Invalid ticket type",
			"",
		))
		return
	}

	// Get default status (open)
	var ticketStatusID string
	err = h.db.QueryRow(context.Background(),
		"SELECT id FROM ticket_status WHERE title = 'open' AND deleted = 0").Scan(&ticketStatusID)

	if err != nil {
		logger.Error("Default ticket status not found", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create ticket",
			"",
		))
		return
	}

	// Get next ticket number for project
	var maxTicketNumber int
	err = h.db.QueryRow(context.Background(),
		"SELECT COALESCE(MAX(ticket_number), 0) FROM ticket WHERE project_id = ?",
		projectID).Scan(&maxTicketNumber)
	if err != nil {
		maxTicketNumber = 0
	}
	ticketNumber := maxTicketNumber + 1

	// Get username from context
	username, _ := middleware.GetUsername(c)

	// Create ticket
	ticketID := uuid.New().String()
	now := time.Now().Unix()

	query := `
		INSERT INTO ticket (id, ticket_number, position, title, description, created, modified,
		                    ticket_type_id, ticket_status_id, project_id, user_id,
		                    estimation, story_points, creator, deleted, version)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(
		context.Background(),
		query,
		ticketID,
		ticketNumber,
		0, // position
		title,
		description,
		now,
		now,
		ticketTypeID,
		ticketStatusID,
		projectID,
		nil, // user_id (not assigned yet)
		0.0, // estimation
		0,   // story_points
		username,
		0, // not deleted
		1, // initial version
	)

	if err != nil {
		logger.Error("Failed to create ticket", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create ticket",
			"",
		))
		return
	}

	// Record creation history
	historyID := uuid.New().String()
	newData := map[string]interface{}{
		"id":            ticketID,
		"ticket_number": ticketNumber,
		"title":         title,
		"description":   description,
		"type":          ticketTypeStr,
		"status":        "open",
		"project_id":    projectID,
		"version":       1,
	}

	_, err = h.db.Exec(context.Background(), `
		INSERT INTO ticket_history (id, ticket_id, version, action, user_id, timestamp, new_data, change_summary)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, historyID, ticketID, 1, models.ActionCreate, username, now, newData, "Ticket created")

	if err != nil {
		logger.Error("Failed to record ticket creation history", zap.Error(err))
		// Don't fail the request for history recording errors
	}

	// Publish ticket created event
	h.publisher.PublishEntityEvent(
		models.ActionCreate,
		"ticket",
		ticketID,
		username,
		map[string]interface{}{
			"id":            ticketID,
			"ticket_number": ticketNumber,
			"title":         title,
			"description":   description,
			"type":          ticketTypeStr,
			"priority":      priority,
			"status":        "open",
			"project_id":    projectID,
		},
		websocket.NewProjectContext(projectID, []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"ticket": map[string]interface{}{
			"id":            ticketID,
			"ticket_number": ticketNumber,
			"title":         title,
			"description":   description,
			"type":          ticketTypeStr,
			"priority":      priority,
			"status":        "open",
			"project_id":    projectID,
			"created":       now,
		},
	})

	c.JSON(http.StatusOK, response)
}

// handleModifyTicket updates an existing ticket with optimistic locking
func (h *Handler) handleModifyTicket(c *gin.Context, req *models.Request) {
	ticketData, ok := req.Data["data"].(map[string]interface{})
	if !ok {
		ticketData = req.Data
	}

	ticketID, _ := ticketData["id"].(string)
	if ticketID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing ticket ID",
			"",
		))
		return
	}

	// Get expected version for optimistic locking
	expectedVersion, _ := ticketData["version"].(float64)
	if expectedVersion == 0 {
		// If no version provided, get current version (backward compatibility)
		err := h.db.QueryRow(context.Background(),
			"SELECT version FROM ticket WHERE id = ? AND deleted = 0", ticketID).Scan(&expectedVersion)
		if err != nil {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				models.ErrorCodeEntityNotFound,
				"Ticket not found",
				"",
			))
			return
		}
	}

	// Get current ticket data for history
	var currentTicket models.Ticket
	var currentStatusTitle string
	err := h.db.QueryRow(context.Background(), `
		SELECT t.id, t.ticket_number, t.title, t.description, t.ticket_type_id, t.ticket_status_id,
		       t.project_id, t.user_id, t.creator, t.estimation, t.story_points, t.created,
		       t.modified, t.deleted, t.version, ts.title
		FROM ticket t
		JOIN ticket_status ts ON t.ticket_status_id = ts.id
		WHERE t.id = ? AND t.deleted = 0
	`, ticketID).Scan(
		&currentTicket.ID, &currentTicket.TicketNumber, &currentTicket.Title,
		&currentTicket.Description, &currentTicket.TicketTypeID, &currentTicket.TicketStatusID,
		&currentTicket.ProjectID, &currentTicket.UserID, &currentTicket.Creator,
		&currentTicket.Estimation, &currentTicket.StoryPoints, &currentTicket.Created,
		&currentTicket.Modified, &currentTicket.Deleted, &currentTicket.Version, &currentStatusTitle,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Ticket not found",
			"",
		))
		return
	}

	// Check version conflict
	if int(expectedVersion) != currentTicket.Version {
		c.JSON(http.StatusConflict, models.NewErrorResponse(
			models.ErrorCodeVersionConflict,
			"Version conflict detected",
			fmt.Sprintf("Current version: %d, Expected: %d", currentTicket.Version, int(expectedVersion)),
		))
		return
	}

	// Build update query with optimistic locking
	updates := []string{}
	args := []interface{}{}
	newVersion := currentTicket.Version + 1

	// Create old data snapshot
	oldData := map[string]interface{}{
		"id":            currentTicket.ID,
		"ticket_number": currentTicket.TicketNumber,
		"title":         currentTicket.Title,
		"description":   currentTicket.Description,
		"status":        currentStatusTitle,
		"version":       currentTicket.Version,
	}

	if title, ok := ticketData["title"].(string); ok && title != "" {
		updates = append(updates, "title = ?")
		args = append(args, title)
	}

	if desc, ok := ticketData["description"].(string); ok {
		updates = append(updates, "description = ?")
		args = append(args, desc)
	}

	if status, ok := ticketData["status"].(string); ok && status != "" {
		// Get status ID
		var statusID string
		err := h.db.QueryRow(context.Background(),
			"SELECT id FROM ticket_status WHERE title = ? AND deleted = 0", status).Scan(&statusID)
		if err == nil {
			updates = append(updates, "ticket_status_id = ?")
			args = append(args, statusID)
		}
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidData,
			"No fields to update",
			"",
		))
		return
	}

	// Always update modified timestamp and version
	updates = append(updates, "modified = ?")
	args = append(args, time.Now().Unix())
	updates = append(updates, "version = ?")
	args = append(args, newVersion)

	// Add optimistic locking condition
	args = append(args, ticketID, int(expectedVersion))

	query := fmt.Sprintf("UPDATE ticket SET %s WHERE id = ? AND version = ?", joinWithComma(updates))
	result, err := h.db.Exec(context.Background(), query, args...)
	if err != nil {
		logger.Error("Failed to update ticket", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to update ticket",
			"",
		))
		return
	}

	// Check if update actually happened (optimistic locking)
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		// Version conflict occurred during update
		var latestVersion int
		h.db.QueryRow(context.Background(),
			"SELECT version FROM ticket WHERE id = ? AND deleted = 0", ticketID).Scan(&latestVersion)

		c.JSON(http.StatusConflict, models.NewErrorResponse(
			models.ErrorCodeVersionConflict,
			"Version conflict detected",
			fmt.Sprintf("Current version: %d, Expected: %d", latestVersion, int(expectedVersion)),
		))
		return
	}

	// Get updated ticket data for history
	var newStatusTitle string
	h.db.QueryRow(context.Background(),
		"SELECT ts.title FROM ticket t JOIN ticket_status ts ON t.ticket_status_id = ts.id WHERE t.id = ?",
		ticketID).Scan(&newStatusTitle)

	newData := map[string]interface{}{
		"id":            currentTicket.ID,
		"ticket_number": currentTicket.TicketNumber,
		"title":         ticketData["title"],
		"description":   ticketData["description"],
		"status":        ticketData["status"],
		"version":       newVersion,
	}

	// Get username from context
	username, _ := middleware.GetUsername(c)
	if username == "" {
		username = "system"
	}

	// Record history
	historyID := uuid.New().String()
	changeSummary := models.GenerateChangeSummary(models.ActionModify, oldData, newData)

	_, err = h.db.Exec(context.Background(), `
		INSERT INTO ticket_history (id, ticket_id, version, action, user_id, timestamp, old_data, new_data, change_summary)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, historyID, ticketID, newVersion, models.ActionModify, username, time.Now().Unix(),
		oldData, newData, changeSummary)

	if err != nil {
		logger.Error("Failed to record ticket history", zap.Error(err))
		// Don't fail the request for history recording errors
	}

	// Publish ticket updated event
	h.publisher.PublishEntityEvent(
		models.ActionModify,
		"ticket",
		ticketID,
		username,
		map[string]interface{}{
			"ticket_data": ticketData,
			"new_version": newVersion,
			"old_version": currentTicket.Version,
		},
		websocket.NewProjectContext(currentTicket.ProjectID, []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"ticket": map[string]interface{}{
			"id":      ticketID,
			"updated": true,
			"version": newVersion,
		},
	})

	c.JSON(http.StatusOK, response)
}

// handleRemoveTicket soft-deletes a ticket
func (h *Handler) handleRemoveTicket(c *gin.Context, req *models.Request) {
	ticketData, ok := req.Data["data"].(map[string]interface{})
	if !ok {
		ticketData = req.Data
	}

	ticketID, _ := ticketData["id"].(string)
	if ticketID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing ticket ID",
			"",
		))
		return
	}

	// Get project_id before deletion for event context
	var projectID string
	err := h.db.QueryRow(context.Background(),
		"SELECT project_id FROM ticket WHERE id = ? AND deleted = 0", ticketID).Scan(&projectID)
	if err != nil {
		logger.Error("Ticket not found", zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Ticket not found",
			"",
		))
		return
	}

	query := "UPDATE ticket SET deleted = 1, modified = ? WHERE id = ?"
	_, err = h.db.Exec(context.Background(), query, time.Now().Unix(), ticketID)
	if err != nil {
		logger.Error("Failed to delete ticket", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to delete ticket",
			"",
		))
		return
	}

	// Get username from context
	username, _ := middleware.GetUsername(c)

	// Publish ticket deleted event
	h.publisher.PublishEntityEvent(
		models.ActionRemove,
		"ticket",
		ticketID,
		username,
		map[string]interface{}{
			"id":         ticketID,
			"project_id": projectID,
		},
		websocket.NewProjectContext(projectID, []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"ticket": map[string]interface{}{
			"id":      ticketID,
			"deleted": true,
		},
	})

	c.JSON(http.StatusOK, response)
}

// handleReadTicket retrieves a single ticket
func (h *Handler) handleReadTicket(c *gin.Context, req *models.Request) {
	ticketData, ok := req.Data["data"].(map[string]interface{})
	if !ok {
		ticketData = req.Data
	}

	ticketID, _ := ticketData["id"].(string)
	if ticketID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing ticket ID",
			"",
		))
		return
	}

	query := `
		SELECT t.id, t.ticket_number, t.title, t.description, t.created, t.modified,
		       tt.title as type, ts.title as status, t.project_id
		FROM ticket t
		JOIN ticket_type tt ON t.ticket_type_id = tt.id
		JOIN ticket_status ts ON t.ticket_status_id = ts.id
		WHERE t.id = ? AND t.deleted = 0
	`

	var id, title, description, ticketType, status, projectID string
	var ticketNumber int
	var created, modified int64

	err := h.db.QueryRow(context.Background(), query, ticketID).Scan(
		&id, &ticketNumber, &title, &description, &created, &modified,
		&ticketType, &status, &projectID)

	if err != nil {
		logger.Error("Ticket not found", zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Ticket not found",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"ticket": map[string]interface{}{
			"id":            id,
			"ticket_number": ticketNumber,
			"title":         title,
			"description":   description,
			"type":          ticketType,
			"status":        status,
			"project_id":    projectID,
			"created":       created,
			"modified":      modified,
		},
	})

	c.JSON(http.StatusOK, response)
}

// handleListTickets retrieves all tickets for a project
func (h *Handler) handleListTickets(c *gin.Context, req *models.Request) {
	// Get project_id from request data
	var projectID string
	if req.Data != nil {
		if data, ok := req.Data["data"].(map[string]interface{}); ok {
			projectID, _ = data["project_id"].(string)
		} else {
			projectID, _ = req.Data["project_id"].(string)
		}
	}

	query := `
		SELECT t.id, t.ticket_number, t.title, t.description, t.created, t.modified,
		       tt.title as type, ts.title as status, t.project_id
		FROM ticket t
		JOIN ticket_type tt ON t.ticket_type_id = tt.id
		JOIN ticket_status ts ON t.ticket_status_id = ts.id
		WHERE t.deleted = 0
	`

	var args []interface{}
	if projectID != "" {
		query += " AND t.project_id = ?"
		args = append(args, projectID)
	}

	query += " ORDER BY t.created DESC"

	rows, err := h.db.Query(context.Background(), query, args...)
	if err != nil {
		logger.Error("Failed to list tickets", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list tickets",
			"",
		))
		return
	}
	defer rows.Close()

	tickets := []map[string]interface{}{}

	for rows.Next() {
		var id, title, description, ticketType, status, projID string
		var ticketNumber int
		var created, modified int64

		err := rows.Scan(&id, &ticketNumber, &title, &description, &created, &modified,
			&ticketType, &status, &projID)
		if err != nil {
			logger.Error("Failed to scan ticket", zap.Error(err))
			continue
		}

		tickets = append(tickets, map[string]interface{}{
			"id":            id,
			"ticket_number": ticketNumber,
			"title":         title,
			"description":   description,
			"type":          ticketType,
			"status":        status,
			"project_id":    projID,
			"created":       created,
			"modified":      modified,
		})
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"items": tickets,
		"total": len(tickets),
	})

	c.JSON(http.StatusOK, response)
}

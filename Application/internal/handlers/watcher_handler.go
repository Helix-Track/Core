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

// handleWatcherAdd adds a user as a watcher to a ticket
func (h *Handler) handleWatcherAdd(c *gin.Context, req *models.Request) {
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
			"Missing ticket ID",
			"",
		))
		return
	}

	// Get user ID from request (defaults to current user if not specified)
	userID, ok := req.Data["userId"].(string)
	if !ok || userID == "" {
		// Use current username as user ID (in production, you'd look up the user ID)
		userID = username
	}

	// Check if already watching
	checkQuery := `
		SELECT COUNT(*) FROM ticket_watcher_mapping
		WHERE ticket_id = ? AND user_id = ? AND deleted = 0
	`
	var count int
	err := h.db.QueryRow(c.Request.Context(), checkQuery, ticketID, userID).Scan(&count)
	if err != nil {
		logger.Error("Failed to check watcher status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to check watcher status",
			"",
		))
		return
	}

	if count > 0 {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeAlreadyExists,
			"Already watching this ticket",
			"",
		))
		return
	}

	// Create watcher mapping
	watcher := &models.TicketWatcherMapping{
		ID:       uuid.New().String(),
		TicketID: ticketID,
		UserID:   userID,
		Created:  time.Now().Unix(),
		Deleted:  false,
	}

	query := `
		INSERT INTO ticket_watcher_mapping (id, ticket_id, user_id, created, deleted)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(c.Request.Context(), query,
		watcher.ID,
		watcher.TicketID,
		watcher.UserID,
		watcher.Created,
		watcher.Deleted,
	)

	if err != nil {
		logger.Error("Failed to add watcher", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to add watcher",
			"",
		))
		return
	}

	logger.Info("Watcher added",
		zap.String("ticket_id", ticketID),
		zap.String("user_id", userID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"watcher": watcher,
	})
	c.JSON(http.StatusCreated, response)
}

// handleWatcherRemove removes a user as a watcher from a ticket
func (h *Handler) handleWatcherRemove(c *gin.Context, req *models.Request) {
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
			"Missing ticket ID",
			"",
		))
		return
	}

	// Get user ID from request (defaults to current user if not specified)
	userID, ok := req.Data["userId"].(string)
	if !ok || userID == "" {
		userID = username
	}

	// Soft delete the watcher mapping
	query := `
		UPDATE ticket_watcher_mapping
		SET deleted = 1
		WHERE ticket_id = ? AND user_id = ? AND deleted = 0
	`

	result, err := h.db.Exec(c.Request.Context(), query, ticketID, userID)
	if err != nil {
		logger.Error("Failed to remove watcher", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to remove watcher",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeNotFound,
			"Watcher not found",
			"",
		))
		return
	}

	logger.Info("Watcher removed",
		zap.String("ticket_id", ticketID),
		zap.String("user_id", userID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"removed": true,
		"ticketId": ticketID,
		"userId":   userID,
	})
	c.JSON(http.StatusOK, response)
}

// handleWatcherList lists all watchers for a ticket
func (h *Handler) handleWatcherList(c *gin.Context, req *models.Request) {
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
			"Missing ticket ID",
			"",
		))
		return
	}

	// Query all watchers for the ticket
	query := `
		SELECT id, ticket_id, user_id, created, deleted
		FROM ticket_watcher_mapping
		WHERE ticket_id = ? AND deleted = 0
		ORDER BY created ASC
	`

	rows, err := h.db.Query(c.Request.Context(), query, ticketID)
	if err != nil {
		logger.Error("Failed to list watchers", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list watchers",
			"",
		))
		return
	}
	defer rows.Close()

	watchers := make([]models.TicketWatcherMapping, 0)
	for rows.Next() {
		var watcher models.TicketWatcherMapping
		err := rows.Scan(
			&watcher.ID,
			&watcher.TicketID,
			&watcher.UserID,
			&watcher.Created,
			&watcher.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan watcher", zap.Error(err))
			continue
		}
		watchers = append(watchers, watcher)
	}

	logger.Info("Watchers listed",
		zap.String("ticket_id", ticketID),
		zap.Int("count", len(watchers)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"watchers": watchers,
		"count":    len(watchers),
		"ticketId": ticketID,
	})
	c.JSON(http.StatusOK, response)
}

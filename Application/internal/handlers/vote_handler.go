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

// handleVoteAdd adds a vote to a ticket
func (h *Handler) handleVoteAdd(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "vote", models.PermissionCreate)
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

	// TODO: Get user ID from JWT/session - for now using username
	userID := username

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

	// Check if user already voted
	voteCheckQuery := `SELECT COUNT(*) FROM ticket_vote_mapping WHERE ticket_id = ? AND user_id = ? AND deleted = 0`
	err = h.db.QueryRow(c.Request.Context(), voteCheckQuery, ticketID, userID).Scan(&count)
	if err != nil {
		logger.Error("Failed to check existing vote", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to check existing vote",
			"",
		))
		return
	}

	if count > 0 {
		c.JSON(http.StatusConflict, models.NewErrorResponse(
			models.ErrorCodeEntityAlreadyExists,
			"User has already voted for this ticket",
			"",
		))
		return
	}

	// Create vote record
	voteID := uuid.New().String()
	now := time.Now().Unix()

	insertQuery := `
		INSERT INTO ticket_vote_mapping (id, ticket_id, user_id, created, deleted)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(c.Request.Context(), insertQuery, voteID, ticketID, userID, now, false)
	if err != nil {
		logger.Error("Failed to add vote", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to add vote",
			"",
		))
		return
	}

	// Update vote count on ticket
	updateQuery := `UPDATE ticket SET vote_count = vote_count + 1 WHERE id = ?`
	_, err = h.db.Exec(c.Request.Context(), updateQuery, ticketID)
	if err != nil {
		logger.Error("Failed to update vote count", zap.Error(err))
		// Don't fail the request, vote was added successfully
	}

	logger.Info("Vote added",
		zap.String("vote_id", voteID),
		zap.String("ticket_id", ticketID),
		zap.String("user_id", userID),
		zap.String("username", username),
	)

	// Publish vote added event
	h.publisher.PublishEntityEvent(
		"voteAdd",
		"vote",
		voteID,
		username,
		map[string]interface{}{
			"id":       voteID,
			"ticketId": ticketID,
			"userId":   userID,
		},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"id":       voteID,
		"ticketId": ticketID,
		"userId":   userID,
		"created":  now,
	})
	c.JSON(http.StatusCreated, response)
}

// handleVoteRemove removes a vote from a ticket
func (h *Handler) handleVoteRemove(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "vote", models.PermissionDelete)
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

	userID := username

	// Find the vote
	var voteID string
	findQuery := `SELECT id FROM ticket_vote_mapping WHERE ticket_id = ? AND user_id = ? AND deleted = 0`
	err = h.db.QueryRow(c.Request.Context(), findQuery, ticketID, userID).Scan(&voteID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Vote not found",
			"",
		))
		return
	}

	if err != nil {
		logger.Error("Failed to find vote", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to find vote",
			"",
		))
		return
	}

	// Soft delete the vote
	deleteQuery := `UPDATE ticket_vote_mapping SET deleted = 1 WHERE id = ?`
	_, err = h.db.Exec(c.Request.Context(), deleteQuery, voteID)
	if err != nil {
		logger.Error("Failed to remove vote", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to remove vote",
			"",
		))
		return
	}

	// Update vote count on ticket
	updateQuery := `UPDATE ticket SET vote_count = CASE WHEN vote_count > 0 THEN vote_count - 1 ELSE 0 END WHERE id = ?`
	_, err = h.db.Exec(c.Request.Context(), updateQuery, ticketID)
	if err != nil {
		logger.Error("Failed to update vote count", zap.Error(err))
		// Don't fail the request, vote was removed successfully
	}

	logger.Info("Vote removed",
		zap.String("vote_id", voteID),
		zap.String("ticket_id", ticketID),
		zap.String("user_id", userID),
		zap.String("username", username),
	)

	// Publish vote removed event
	h.publisher.PublishEntityEvent(
		"voteRemove",
		"vote",
		voteID,
		username,
		map[string]interface{}{
			"id":       voteID,
			"ticketId": ticketID,
			"userId":   userID,
		},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"deleted":  true,
		"id":       voteID,
		"ticketId": ticketID,
	})
	c.JSON(http.StatusOK, response)
}

// handleVoteCount gets the vote count for a ticket
func (h *Handler) handleVoteCount(c *gin.Context, req *models.Request) {
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

	// Get vote count from ticket
	var voteCount int
	query := `SELECT vote_count FROM ticket WHERE id = ? AND deleted = 0`
	err := h.db.QueryRow(c.Request.Context(), query, ticketID).Scan(&voteCount)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Ticket not found",
			"",
		))
		return
	}

	if err != nil {
		logger.Error("Failed to get vote count", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to get vote count",
			"",
		))
		return
	}

	logger.Info("Vote count retrieved",
		zap.String("ticket_id", ticketID),
		zap.Int("vote_count", voteCount),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"ticketId":  ticketID,
		"voteCount": voteCount,
	})
	c.JSON(http.StatusOK, response)
}

// handleVoteList lists all voters for a ticket
func (h *Handler) handleVoteList(c *gin.Context, req *models.Request) {
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

	// Get all votes for this ticket
	query := `
		SELECT id, ticket_id, user_id, created
		FROM ticket_vote_mapping
		WHERE ticket_id = ? AND deleted = 0
		ORDER BY created DESC
	`

	rows, err := h.db.Query(c.Request.Context(), query, ticketID)
	if err != nil {
		logger.Error("Failed to list votes", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list votes",
			"",
		))
		return
	}
	defer rows.Close()

	votes := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, tid, userID string
		var created int64
		err := rows.Scan(&id, &tid, &userID, &created)
		if err != nil {
			logger.Error("Failed to scan vote", zap.Error(err))
			continue
		}
		votes = append(votes, map[string]interface{}{
			"id":       id,
			"ticketId": tid,
			"userId":   userID,
			"created":  created,
		})
	}

	logger.Info("Votes listed",
		zap.String("ticket_id", ticketID),
		zap.Int("count", len(votes)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"ticketId": ticketID,
		"votes":    votes,
		"count":    len(votes),
	})
	c.JSON(http.StatusOK, response)
}

// handleVoteCheck checks if the current user has voted for a ticket
func (h *Handler) handleVoteCheck(c *gin.Context, req *models.Request) {
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

	userID := username

	// Check if user has voted
	var count int
	query := `SELECT COUNT(*) FROM ticket_vote_mapping WHERE ticket_id = ? AND user_id = ? AND deleted = 0`
	err := h.db.QueryRow(c.Request.Context(), query, ticketID, userID).Scan(&count)
	if err != nil {
		logger.Error("Failed to check vote", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to check vote",
			"",
		))
		return
	}

	hasVoted := count > 0

	logger.Info("Vote check",
		zap.String("ticket_id", ticketID),
		zap.String("user_id", userID),
		zap.Bool("has_voted", hasVoted),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"ticketId": ticketID,
		"userId":   userID,
		"hasVoted": hasVoted,
	})
	c.JSON(http.StatusOK, response)
}

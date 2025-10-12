package handlers

import (
	"database/sql"
	"time"

	"helixtrack.ru/core/internal/models"
)

// VoteAdd adds a vote to a ticket
func (h *Handler) VoteAdd(req models.Request) models.Response {
	ticketID := h.getString(req.Data, "ticketId")
	if ticketID == "" {
		return h.errorResponse(models.ErrorMissingParameter, "ticketId required", nil)
	}

	userID := h.getUserIDFromJWT(req.JWT)
	if userID == "" {
		return h.errorResponse(models.ErrorUnauthorized, "User ID not found in JWT", nil)
	}

	// Check if ticket exists
	var exists bool
	err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM ticket WHERE id = ? AND deleted = 0)", ticketID).Scan(&exists)
	if err != nil || !exists {
		return h.errorResponse(models.ErrorNotFound, "Ticket not found", err)
	}

	// Check if user already voted
	err = h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM ticket_vote_mapping WHERE ticket_id = ? AND user_id = ? AND deleted = 0)", ticketID, userID).Scan(&exists)
	if err != nil {
		return h.errorResponse(models.ErrorDatabaseError, "Failed to check existing vote", err)
	}
	if exists {
		return h.errorResponse(models.ErrorAlreadyExists, "User already voted for this ticket", nil)
	}

	vote := models.Vote{
		ID:       h.generateID("vote"),
		TicketID: ticketID,
		UserID:   userID,
		Created:  time.Now().Unix(),
		Deleted:  false,
	}

	// Insert vote
	_, err = h.db.Exec(
		"INSERT INTO ticket_vote_mapping (id, ticket_id, user_id, created, deleted) VALUES (?, ?, ?, ?, ?)",
		vote.ID, vote.TicketID, vote.UserID, vote.Created, vote.Deleted,
	)
	if err != nil {
		return h.errorResponse(models.ErrorDatabaseError, "Failed to add vote", err)
	}

	// Update vote count
	_, err = h.db.Exec("UPDATE ticket SET vote_count = vote_count + 1 WHERE id = ?", ticketID)
	if err != nil {
		h.logger.Error("Failed to update vote count", "error", err)
	}

	h.publishEvent("vote.added", vote)
	return h.successResponse(vote)
}

// VoteRemove removes a vote from a ticket
func (h *Handler) VoteRemove(req models.Request) models.Response {
	ticketID := h.getString(req.Data, "ticketId")
	if ticketID == "" {
		return h.errorResponse(models.ErrorMissingParameter, "ticketId required", nil)
	}

	userID := h.getUserIDFromJWT(req.JWT)
	if userID == "" {
		return h.errorResponse(models.ErrorUnauthorized, "User ID not found in JWT", nil)
	}

	// Check if vote exists
	var voteID string
	err := h.db.QueryRow("SELECT id FROM ticket_vote_mapping WHERE ticket_id = ? AND user_id = ? AND deleted = 0", ticketID, userID).Scan(&voteID)
	if err == sql.ErrNoRows {
		return h.errorResponse(models.ErrorNotFound, "Vote not found", nil)
	}
	if err != nil {
		return h.errorResponse(models.ErrorDatabaseError, "Failed to check vote", err)
	}

	// Soft delete vote
	_, err = h.db.Exec("UPDATE ticket_vote_mapping SET deleted = 1 WHERE id = ?", voteID)
	if err != nil {
		return h.errorResponse(models.ErrorDatabaseError, "Failed to remove vote", err)
	}

	// Update vote count
	_, err = h.db.Exec("UPDATE ticket SET vote_count = CASE WHEN vote_count > 0 THEN vote_count - 1 ELSE 0 END WHERE id = ?", ticketID)
	if err != nil {
		h.logger.Error("Failed to update vote count", "error", err)
	}

	h.publishEvent("vote.removed", map[string]interface{}{
		"ticketId": ticketID,
		"userId":   userID,
	})

	return h.successResponse(map[string]interface{}{
		"ticketId": ticketID,
		"message":  "Vote removed successfully",
	})
}

// VoteCount gets the vote count for a ticket
func (h *Handler) VoteCount(req models.Request) models.Response {
	ticketID := h.getString(req.Data, "ticketId")
	if ticketID == "" {
		return h.errorResponse(models.ErrorMissingParameter, "ticketId required", nil)
	}

	var voteCount int
	err := h.db.QueryRow("SELECT vote_count FROM ticket WHERE id = ? AND deleted = 0", ticketID).Scan(&voteCount)
	if err == sql.ErrNoRows {
		return h.errorResponse(models.ErrorNotFound, "Ticket not found", nil)
	}
	if err != nil {
		return h.errorResponse(models.ErrorDatabaseError, "Failed to get vote count", err)
	}

	return h.successResponse(map[string]interface{}{
		"ticketId":  ticketID,
		"voteCount": voteCount,
	})
}

// VoteList lists all voters for a ticket
func (h *Handler) VoteList(req models.Request) models.Response {
	ticketID := h.getString(req.Data, "ticketId")
	if ticketID == "" {
		return h.errorResponse(models.ErrorMissingParameter, "ticketId required", nil)
	}

	rows, err := h.db.Query(
		"SELECT id, ticket_id, user_id, created FROM ticket_vote_mapping WHERE ticket_id = ? AND deleted = 0 ORDER BY created DESC",
		ticketID,
	)
	if err != nil {
		return h.errorResponse(models.ErrorDatabaseError, "Failed to list votes", err)
	}
	defer rows.Close()

	votes := []models.Vote{}
	for rows.Next() {
		var vote models.Vote
		if err := rows.Scan(&vote.ID, &vote.TicketID, &vote.UserID, &vote.Created); err != nil {
			h.logger.Error("Failed to scan vote", "error", err)
			continue
		}
		votes = append(votes, vote)
	}

	return h.successResponse(map[string]interface{}{
		"ticketId": ticketID,
		"votes":    votes,
		"count":    len(votes),
	})
}

// VoteCheck checks if the current user has voted for a ticket
func (h *Handler) VoteCheck(req models.Request) models.Response {
	ticketID := h.getString(req.Data, "ticketId")
	if ticketID == "" {
		return h.errorResponse(models.ErrorMissingParameter, "ticketId required", nil)
	}

	userID := h.getUserIDFromJWT(req.JWT)
	if userID == "" {
		return h.errorResponse(models.ErrorUnauthorized, "User ID not found in JWT", nil)
	}

	var exists bool
	err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM ticket_vote_mapping WHERE ticket_id = ? AND user_id = ? AND deleted = 0)", ticketID, userID).Scan(&exists)
	if err != nil {
		return h.errorResponse(models.ErrorDatabaseError, "Failed to check vote", err)
	}

	return h.successResponse(map[string]interface{}{
		"ticketId": ticketID,
		"userId":   userID,
		"hasVoted": exists,
	})
}

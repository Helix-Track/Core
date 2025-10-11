package handlers

import (
	"context"
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

// handleCreateComment creates a new comment
func (h *Handler) handleCreateComment(c *gin.Context, req *models.Request) {
	commentData, ok := req.Data["data"].(map[string]interface{})
	if !ok {
		commentData = req.Data
	}

	ticketID, _ := commentData["ticket_id"].(string)
	if ticketID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing ticket_id",
			"",
		))
		return
	}

	commentText, _ := commentData["comment"].(string)
	if commentText == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing comment text",
			"",
		))
		return
	}

	// Create comment
	commentID := uuid.New().String()
	now := time.Now().Unix()

	query := `
		INSERT INTO comment (id, comment, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err := h.db.Exec(
		context.Background(),
		query,
		commentID,
		commentText,
		now,
		now,
		0,
	)

	if err != nil {
		logger.Error("Failed to create comment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create comment",
			"",
		))
		return
	}

	// Create ticket-comment mapping
	mappingID := uuid.New().String()
	mappingQuery := `
		INSERT INTO comment_ticket_mapping (id, comment_id, ticket_id, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(
		context.Background(),
		mappingQuery,
		mappingID,
		commentID,
		ticketID,
		now,
		now,
		0,
	)

	if err != nil {
		logger.Error("Failed to create comment mapping", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create comment",
			"",
		))
		return
	}

	// Get project_id from ticket for event context
	var projectID string
	h.db.QueryRow(context.Background(),
		"SELECT project_id FROM ticket WHERE id = ? AND deleted = 0", ticketID).Scan(&projectID)

	// Get username from context
	username, _ := middleware.GetUsername(c)

	// Publish comment created event
	if projectID != "" {
		h.publisher.PublishEntityEvent(
			models.ActionCreate,
			"comment",
			commentID,
			username,
			map[string]interface{}{
				"id":        commentID,
				"comment":   commentText,
				"ticket_id": ticketID,
			},
			websocket.NewProjectContext(projectID, []string{"READ"}),
		)
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"comment": map[string]interface{}{
			"id":        commentID,
			"comment":   commentText,
			"ticket_id": ticketID,
			"created":   now,
		},
	})

	c.JSON(http.StatusOK, response)
}

// handleModifyComment updates an existing comment
func (h *Handler) handleModifyComment(c *gin.Context, req *models.Request) {
	commentData, ok := req.Data["data"].(map[string]interface{})
	if !ok {
		commentData = req.Data
	}

	commentID, _ := commentData["id"].(string)
	if commentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing comment ID",
			"",
		))
		return
	}

	commentText, _ := commentData["comment"].(string)
	if commentText == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing comment text",
			"",
		))
		return
	}

	query := "UPDATE comment SET comment = ?, modified = ? WHERE id = ? AND deleted = 0"
	_, err := h.db.Exec(context.Background(), query, commentText, time.Now().Unix(), commentID)
	if err != nil {
		logger.Error("Failed to update comment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to update comment",
			"",
		))
		return
	}

	// Get project_id from ticket for event context
	var ticketID, projectID string
	h.db.QueryRow(context.Background(),
		`SELECT t.id, t.project_id FROM ticket t
		 JOIN comment_ticket_mapping ctm ON t.id = ctm.ticket_id
		 WHERE ctm.comment_id = ? AND t.deleted = 0`, commentID).Scan(&ticketID, &projectID)

	// Get username from context
	username, _ := middleware.GetUsername(c)

	// Publish comment updated event
	if projectID != "" {
		h.publisher.PublishEntityEvent(
			models.ActionModify,
			"comment",
			commentID,
			username,
			map[string]interface{}{
				"id":        commentID,
				"comment":   commentText,
				"ticket_id": ticketID,
			},
			websocket.NewProjectContext(projectID, []string{"READ"}),
		)
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"comment": map[string]interface{}{
			"id":      commentID,
			"updated": true,
		},
	})

	c.JSON(http.StatusOK, response)
}

// handleRemoveComment soft-deletes a comment
func (h *Handler) handleRemoveComment(c *gin.Context, req *models.Request) {
	commentData, ok := req.Data["data"].(map[string]interface{})
	if !ok {
		commentData = req.Data
	}

	commentID, _ := commentData["id"].(string)
	if commentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing comment ID",
			"",
		))
		return
	}

	// Get project_id from ticket before deletion for event context
	var ticketID, projectID string
	err := h.db.QueryRow(context.Background(),
		`SELECT t.id, t.project_id FROM ticket t
		 JOIN comment_ticket_mapping ctm ON t.id = ctm.ticket_id
		 WHERE ctm.comment_id = ? AND t.deleted = 0`, commentID).Scan(&ticketID, &projectID)
	if err != nil {
		logger.Error("Comment not found", zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Comment not found",
			"",
		))
		return
	}

	query := "UPDATE comment SET deleted = 1, modified = ? WHERE id = ?"
	_, err = h.db.Exec(context.Background(), query, time.Now().Unix(), commentID)
	if err != nil {
		logger.Error("Failed to delete comment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to delete comment",
			"",
		))
		return
	}

	// Get username from context
	username, _ := middleware.GetUsername(c)

	// Publish comment deleted event
	h.publisher.PublishEntityEvent(
		models.ActionRemove,
		"comment",
		commentID,
		username,
		map[string]interface{}{
			"id":        commentID,
			"ticket_id": ticketID,
		},
		websocket.NewProjectContext(projectID, []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"comment": map[string]interface{}{
			"id":      commentID,
			"deleted": true,
		},
	})

	c.JSON(http.StatusOK, response)
}

// handleReadComment retrieves a single comment
func (h *Handler) handleReadComment(c *gin.Context, req *models.Request) {
	commentData, ok := req.Data["data"].(map[string]interface{})
	if !ok {
		commentData = req.Data
	}

	commentID, _ := commentData["id"].(string)
	if commentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing comment ID",
			"",
		))
		return
	}

	query := `
		SELECT id, comment, created, modified
		FROM comment
		WHERE id = ? AND deleted = 0
	`

	var id, comment string
	var created, modified int64

	err := h.db.QueryRow(context.Background(), query, commentID).Scan(
		&id, &comment, &created, &modified)

	if err != nil {
		logger.Error("Comment not found", zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Comment not found",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"comment": map[string]interface{}{
			"id":       id,
			"comment":  comment,
			"created":  created,
			"modified": modified,
		},
	})

	c.JSON(http.StatusOK, response)
}

// handleListComments retrieves all comments for a ticket
func (h *Handler) handleListComments(c *gin.Context, req *models.Request) {
	// Get ticket_id from request data
	var ticketID string
	if req.Data != nil {
		if data, ok := req.Data["data"].(map[string]interface{}); ok {
			ticketID, _ = data["ticket_id"].(string)
		} else {
			ticketID, _ = req.Data["ticket_id"].(string)
		}
	}

	if ticketID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing ticket_id",
			"",
		))
		return
	}

	query := `
		SELECT c.id, c.comment, c.created, c.modified
		FROM comment c
		JOIN comment_ticket_mapping ctm ON c.id = ctm.comment_id
		WHERE ctm.ticket_id = ? AND c.deleted = 0
		ORDER BY c.created DESC
	`

	rows, err := h.db.Query(context.Background(), query, ticketID)
	if err != nil {
		logger.Error("Failed to list comments", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list comments",
			"",
		))
		return
	}
	defer rows.Close()

	comments := []map[string]interface{}{}

	for rows.Next() {
		var id, comment string
		var created, modified int64

		err := rows.Scan(&id, &comment, &created, &modified)
		if err != nil {
			logger.Error("Failed to scan comment", zap.Error(err))
			continue
		}

		comments = append(comments, map[string]interface{}{
			"id":       id,
			"comment":  comment,
			"created":  created,
			"modified": modified,
		})
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"items": comments,
		"total": len(comments),
	})

	c.JSON(http.StatusOK, response)
}

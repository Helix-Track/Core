package handlers

import (
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"helixtrack.ru/core/internal/logger"
	"helixtrack.ru/core/internal/middleware"
	"helixtrack.ru/core/internal/models"
)

/*
Comment Mention Handlers - Phase 3

Comment mentions provide @username functionality in comments.
Users can @mention others in comments to notify them.

Handlers:
  1. handleCommentMention - Add mention to comment
  2. handleCommentUnmention - Remove mention from comment
  3. handleCommentListMentions - List all mentions in a comment
  4. handleCommentGetMentions - Get all mentions for a user
  5. handleCommentParseMentions - Parse @mentions from text
*/

// mentionRegex matches @username patterns in text
var mentionRegex = regexp.MustCompile(`@([a-zA-Z0-9_.-]+)`)

// handleCommentMention adds a user mention to a comment
func (h *Handler) handleCommentMention(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "comment", models.PermissionCreate)
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

	// Get comment ID
	commentID, ok := req.Data["commentId"].(string)
	if !ok || commentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing commentId",
			"",
		))
		return
	}

	// Get mentioned user ID
	mentionedUserID, ok := req.Data["userId"].(string)
	if !ok || mentionedUserID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing userId",
			"",
		))
		return
	}

	// Verify comment exists
	var commentExists int
	err = h.db.QueryRow(c.Request.Context(),
		"SELECT COUNT(*) FROM comment WHERE id = ? AND deleted = 0",
		commentID,
	).Scan(&commentExists)

	if err != nil || commentExists == 0 {
		logger.Error("Comment not found", zap.Error(err), zap.String("commentId", commentID))
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Comment not found",
			"",
		))
		return
	}

	// Check if mention already exists
	var existingMentionID string
	err = h.db.QueryRow(c.Request.Context(),
		"SELECT id FROM comment_mention_mapping WHERE comment_id = ? AND mentioned_user_id = ? AND deleted = 0",
		commentID, mentionedUserID,
	).Scan(&existingMentionID)

	if err == nil && existingMentionID != "" {
		// Mention already exists
		c.JSON(http.StatusOK, models.NewSuccessResponse(map[string]interface{}{
			"message":   "Mention already exists",
			"mentionId": existingMentionID,
		}))
		return
	}

	// Create mention
	mentionID := uuid.New().String()
	now := time.Now().Unix()

	query := `
		INSERT INTO comment_mention_mapping (id, comment_id, mentioned_user_id, created, deleted)
		VALUES (?, ?, ?, ?, 0)
	`

	_, err = h.db.Exec(c.Request.Context(), query, mentionID, commentID, mentionedUserID, now)
	if err != nil {
		logger.Error("Failed to create mention", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeDatabaseError,
			"Failed to create mention",
			"",
		))
		return
	}

	// Publish event
	h.publisher.PublishEntityEvent("mention", "mention", mentionID, username,
		map[string]interface{}{
			"commentId":       commentID,
			"mentionedUserId": mentionedUserID,
		},
		models.EventContext{})

	logger.Info("Mention created",
		zap.String("username", username),
		zap.String("mentionId", mentionID),
		zap.String("commentId", commentID),
		zap.String("mentionedUserId", mentionedUserID),
	)

	c.JSON(http.StatusCreated, models.NewSuccessResponse(map[string]interface{}{
		"mentionId":       mentionID,
		"commentId":       commentID,
		"mentionedUserId": mentionedUserID,
		"created":         now,
	}))
}

// handleCommentUnmention removes a user mention from a comment
func (h *Handler) handleCommentUnmention(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "comment", models.PermissionDelete)
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

	// Get comment ID
	commentID, ok := req.Data["commentId"].(string)
	if !ok || commentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing commentId",
			"",
		))
		return
	}

	// Get mentioned user ID
	mentionedUserID, ok := req.Data["userId"].(string)
	if !ok || mentionedUserID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing userId",
			"",
		))
		return
	}

	// Soft-delete mention
	query := `
		UPDATE comment_mention_mapping
		SET deleted = 1
		WHERE comment_id = ? AND mentioned_user_id = ? AND deleted = 0
	`

	result, err := h.db.Exec(c.Request.Context(), query, commentID, mentionedUserID)
	if err != nil {
		logger.Error("Failed to remove mention", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeDatabaseError,
			"Failed to remove mention",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Mention not found",
			"",
		))
		return
	}

	// Publish event
	h.publisher.PublishEntityEvent("mention", "mention", commentID, username,
		map[string]interface{}{
			"commentId":       commentID,
			"mentionedUserId": mentionedUserID,
			"action":          "removed",
		},
		models.EventContext{})

	logger.Info("Mention removed",
		zap.String("username", username),
		zap.String("commentId", commentID),
		zap.String("mentionedUserId", mentionedUserID),
	)

	c.JSON(http.StatusOK, models.NewSuccessResponse(map[string]interface{}{
		"message":         "Mention removed successfully",
		"commentId":       commentID,
		"mentionedUserId": mentionedUserID,
	}))
}

// handleCommentListMentions lists all mentions in a comment
func (h *Handler) handleCommentListMentions(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "comment", models.PermissionRead)
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

	// Get comment ID
	commentID, ok := req.Data["commentId"].(string)
	if !ok || commentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing commentId",
			"",
		))
		return
	}

	// Query mentions
	query := `
		SELECT id, comment_id, mentioned_user_id, created, deleted
		FROM comment_mention_mapping
		WHERE comment_id = ? AND deleted = 0
		ORDER BY created ASC
	`

	rows, err := h.db.Query(c.Request.Context(), query, commentID)
	if err != nil {
		logger.Error("Failed to query mentions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeDatabaseError,
			"Failed to query mentions",
			"",
		))
		return
	}
	defer rows.Close()

	mentions := []models.Mention{}
	for rows.Next() {
		var mention models.Mention
		err := rows.Scan(
			&mention.ID,
			&mention.CommentID,
			&mention.MentionedUserID,
			&mention.Created,
			&mention.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan mention", zap.Error(err))
			continue
		}
		mentions = append(mentions, mention)
	}

	logger.Info("Mentions listed",
		zap.String("username", username),
		zap.String("commentId", commentID),
		zap.Int("count", len(mentions)),
	)

	c.JSON(http.StatusOK, models.NewSuccessResponse(map[string]interface{}{
		"mentions":  mentions,
		"commentId": commentID,
		"count":     len(mentions),
	}))
}

// handleCommentGetMentions gets all mentions for a specific user (where they were mentioned)
func (h *Handler) handleCommentGetMentions(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "comment", models.PermissionRead)
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

	// Get user ID (optional - defaults to current user)
	userID := getStringFromData(req.Data, "userId")
	if userID == "" {
		userID = username
	}

	// Get pagination parameters
	limit := 50
	if limitVal, ok := req.Data["limit"].(float64); ok && limitVal > 0 {
		limit = int(limitVal)
		if limit > 1000 {
			limit = 1000
		}
	}

	offset := 0
	if offsetVal, ok := req.Data["offset"].(float64); ok && offsetVal > 0 {
		offset = int(offsetVal)
	}

	// Query mentions for user
	query := `
		SELECT cmm.id, cmm.comment_id, cmm.mentioned_user_id, cmm.created, cmm.deleted
		FROM comment_mention_mapping cmm
		WHERE cmm.mentioned_user_id = ? AND cmm.deleted = 0
		ORDER BY cmm.created DESC
		LIMIT ? OFFSET ?
	`

	rows, err := h.db.Query(c.Request.Context(), query, userID, limit, offset)
	if err != nil {
		logger.Error("Failed to query mentions for user", zap.Error(err), zap.String("userId", userID))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeDatabaseError,
			"Failed to query mentions",
			"",
		))
		return
	}
	defer rows.Close()

	mentions := []models.Mention{}
	for rows.Next() {
		var mention models.Mention
		err := rows.Scan(
			&mention.ID,
			&mention.CommentID,
			&mention.MentionedUserID,
			&mention.Created,
			&mention.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan mention", zap.Error(err))
			continue
		}
		mentions = append(mentions, mention)
	}

	logger.Info("User mentions retrieved",
		zap.String("username", username),
		zap.String("userId", userID),
		zap.Int("count", len(mentions)),
	)

	c.JSON(http.StatusOK, models.NewSuccessResponse(map[string]interface{}{
		"mentions": mentions,
		"userId":   userID,
		"count":    len(mentions),
		"limit":    limit,
		"offset":   offset,
	}))
}

// handleCommentParseMentions parses @mentions from text and returns usernames
func (h *Handler) handleCommentParseMentions(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get text to parse
	text, ok := req.Data["text"].(string)
	if !ok || text == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing text",
			"",
		))
		return
	}

	// Parse mentions using regex
	matches := mentionRegex.FindAllStringSubmatch(text, -1)
	usernames := []string{}
	uniqueUsernames := make(map[string]bool)

	for _, match := range matches {
		if len(match) > 1 {
			mentionedUsername := match[1]
			// Avoid duplicates
			if !uniqueUsernames[mentionedUsername] {
				usernames = append(usernames, mentionedUsername)
				uniqueUsernames[mentionedUsername] = true
			}
		}
	}

	// Optionally validate usernames against database
	validateUsers := getBoolFromData(req.Data, "validate")
	validUserIDs := make(map[string]string) // username -> userID

	if validateUsers && len(usernames) > 0 {
		// Query database to validate usernames and get user IDs
		placeholders := ""
		args := []interface{}{}
		for i, un := range usernames {
			if i > 0 {
				placeholders += ", "
			}
			placeholders += "?"
			args = append(args, un)
		}

		query := "SELECT id, username FROM users WHERE username IN (" + placeholders + ") AND deleted = 0"
		rows, err := h.db.Query(c.Request.Context(), query, args...)
		if err != nil {
			logger.Error("Failed to validate usernames", zap.Error(err))
		} else {
			defer rows.Close()
			for rows.Next() {
				var userID, uname string
				if err := rows.Scan(&userID, &uname); err == nil {
					validUserIDs[uname] = userID
				}
			}
		}
	}

	logger.Info("Mentions parsed from text",
		zap.String("username", username),
		zap.Int("count", len(usernames)),
	)

	response := map[string]interface{}{
		"usernames": usernames,
		"count":     len(usernames),
	}

	if validateUsers {
		response["validUserIds"] = validUserIDs
		response["validCount"] = len(validUserIDs)
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(response))
}

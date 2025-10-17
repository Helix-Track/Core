package handlers

import (
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"helixtrack.ru/chat/internal/logger"
	"helixtrack.ru/chat/internal/models"
)

// MessageSend sends a new message
func (h *Handler) MessageSend(c *gin.Context, req map[string]interface{}, claims *models.JWTClaims) {
	data, err := getDataMap(req)
	if err != nil {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeMissingParameter, "Missing data"))
		return
	}

	chatRoomIDStr, ok := getString(data, "chat_room_id")
	if !ok || chatRoomIDStr == "" {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidParameter, "id is required"))
		return
	}

	chatRoomID, err := uuid.Parse(chatRoomIDStr)
	if err != nil {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidParameter, "Invalid chat_room_id"))
		return
	}

	content, ok := getString(data, "content")
	if !ok || content == "" {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidParameter, "content is required"))
		return
	}

	messageType, _ := getString(data, "type")
	if messageType == "" {
		messageType = string(models.MessageTypeText)
	}

	// Check if user is participant
	_, err = h.db.ParticipantGet(c.Request.Context(), chatRoomIDStr, claims.UserID.String())
	if err != nil {
		c.JSON(403, models.ErrorResponse(models.ErrorCodeForbidden, "Not a participant of this chat room"))
		return
	}

	// Create message
	message := &models.Message{
		ChatRoomID:    chatRoomID,
		SenderID:      claims.UserID,
		Type:          models.MessageType(messageType),
		Content:       content,
		ContentFormat: models.ContentFormatPlain,
		IsEdited:      false,
		IsPinned:      false,
	}

	// Handle optional fields
	if contentFormat, ok := getString(data, "content_format"); ok {
		message.ContentFormat = models.ContentFormat(contentFormat)
	}

	if metadata, ok := data["metadata"]; ok {
		metadataBytes, _ := json.Marshal(metadata)
		message.Metadata = json.RawMessage(metadataBytes)
	}

	// Validate message
	msgReq := &models.MessageRequest{
		ChatRoomID:    chatRoomID,
		Type:          models.MessageType(messageType),
		Content:       content,
		ContentFormat: message.ContentFormat,
	}

	if err := msgReq.Validate(); err != nil {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidParameter, err.Error()))
		return
	}

	// Save to database
	if err := h.db.MessageCreate(c.Request.Context(), message); err != nil {
		logger.Error("Failed to create message", zap.Error(err))
		c.JSON(500, models.ErrorResponse(models.ErrorCodeDatabaseError, "Failed to send message"))
		return
	}

	logger.Info("Message sent",
		zap.String("message_id", message.ID.String()),
		zap.String("chat_room_id", chatRoomIDStr),
		zap.String("sender", claims.Username),
	)

	// TODO: Broadcast WebSocket event

	c.JSON(200, models.SuccessResponse(message))
}

// MessageList lists messages in a chat room
func (h *Handler) MessageList(c *gin.Context, req map[string]interface{}, claims *models.JWTClaims) {
	data, err := getDataMap(req)
	if err != nil {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeMissingParameter, "Missing data"))
		return
	}

	chatRoomIDStr, ok := getString(data, "chat_room_id")
	if !ok || chatRoomIDStr == "" {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidParameter, "id is required"))
		return
	}

	chatRoomID, err := uuid.Parse(chatRoomIDStr)
	if err != nil {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidParameter, "Invalid chat_room_id"))
		return
	}

	// Check if user is participant
	_, err = h.db.ParticipantGet(c.Request.Context(), chatRoomIDStr, claims.UserID.String())
	if err != nil {
		c.JSON(403, models.ErrorResponse(models.ErrorCodeForbidden, "Not a participant of this chat room"))
		return
	}

	// Build request
	listReq := &models.MessageListRequest{
		ChatRoomID: chatRoomID,
		Limit:      50,
		Offset:     0,
	}

	if limit, ok := getInt(data, "limit"); ok && limit > 0 {
		listReq.Limit = limit
	}

	if offset, ok := getInt(data, "offset"); ok && offset >= 0 {
		listReq.Offset = offset
	}

	// Get messages
	messages, total, err := h.db.MessageList(c.Request.Context(), listReq)
	if err != nil {
		logger.Error("Failed to list messages", zap.Error(err))
		c.JSON(500, models.ErrorResponse(models.ErrorCodeDatabaseError, "Failed to list messages"))
		return
	}

	response := models.ListResponse{
		Items: messages,
		Pagination: &models.PaginationMeta{
			Total:   total,
			Limit:   listReq.Limit,
			Offset:  listReq.Offset,
			HasMore: listReq.Offset+listReq.Limit < total,
		},
	}

	if listReq.Offset+listReq.Limit < total {
		response.Pagination.NextOffset = listReq.Offset + listReq.Limit
	}

	c.JSON(200, models.SuccessResponse(response))
}

// MessageRead retrieves a single message
func (h *Handler) MessageRead(c *gin.Context, req map[string]interface{}, claims *models.JWTClaims) {
	data, err := getDataMap(req)
	if err != nil {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeMissingParameter, "Missing data"))
		return
	}

	messageID, ok := getString(data, "id")
	if !ok || messageID == "" {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidParameter, "id is required"))
		return
	}

	// Get message
	message, err := h.db.MessageRead(c.Request.Context(), messageID)
	if err != nil {
		if err == models.ErrNotFound {
			c.JSON(404, models.ErrorResponse(models.ErrorCodeNotFound, "Message not found"))
		} else {
			logger.Error("Failed to read message", zap.Error(err))
			c.JSON(500, models.ErrorResponse(models.ErrorCodeDatabaseError, "Failed to read message"))
		}
		return
	}

	// Check if user is participant of the chat room
	chatRoomIDStr := message.ChatRoomID.String()
	_, err = h.db.ParticipantGet(c.Request.Context(), chatRoomIDStr, claims.UserID.String())
	if err != nil {
		c.JSON(403, models.ErrorResponse(models.ErrorCodeForbidden, "Not a participant of this chat room"))
		return
	}

	c.JSON(200, models.SuccessResponse(message))
}

// MessageUpdate updates a message
func (h *Handler) MessageUpdate(c *gin.Context, req map[string]interface{}, claims *models.JWTClaims) {
	data, err := getDataMap(req)
	if err != nil {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeMissingParameter, "Missing data"))
		return
	}

	messageID, ok := getString(data, "id")
	if !ok || messageID == "" {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidParameter, "id is required"))
		return
	}

	content, ok := getString(data, "content")
	if !ok || content == "" {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidParameter, "content is required"))
		return
	}

	// Get existing message
	message, err := h.db.MessageRead(c.Request.Context(), messageID)
	if err != nil {
		if err == models.ErrNotFound {
			c.JSON(404, models.ErrorResponse(models.ErrorCodeNotFound, "Message not found"))
		} else {
			c.JSON(500, models.ErrorResponse(models.ErrorCodeDatabaseError, "Failed to read message"))
		}
		return
	}

	// Check if user is the sender
	if message.SenderID != claims.UserID {
		c.JSON(403, models.ErrorResponse(models.ErrorCodeForbidden, "Can only edit your own messages"))
		return
	}

	// Save current state to edit history before updating
	editCount, err := h.db.MessageEditHistoryCount(c.Request.Context(), messageID)
	if err != nil {
		logger.Error("Failed to get edit history count", zap.Error(err))
		// Continue with update even if history fails (non-critical)
		editCount = 0
	}

	history := &models.MessageEditHistory{
		MessageID:             message.ID,
		EditorID:              claims.UserID,
		PreviousContent:       message.Content,
		PreviousContentFormat: message.ContentFormat,
		PreviousMetadata:      message.Metadata,
		EditNumber:            editCount + 1,
		EditedAt:              time.Now().Unix(),
	}

	if err := h.db.MessageEditHistoryCreate(c.Request.Context(), history); err != nil {
		logger.Error("Failed to create edit history", zap.Error(err))
		// Continue with update even if history fails (non-critical)
	}

	// Update content
	message.Content = content
	message.IsEdited = true

	if contentFormat, ok := getString(data, "content_format"); ok {
		message.ContentFormat = models.ContentFormat(contentFormat)
	}

	// Update in database
	if err := h.db.MessageUpdate(c.Request.Context(), message); err != nil {
		logger.Error("Failed to update message", zap.Error(err))
		c.JSON(500, models.ErrorResponse(models.ErrorCodeDatabaseError, "Failed to update message"))
		return
	}

	logger.Info("Message updated",
		zap.String("message_id", messageID),
		zap.String("updated_by", claims.Username),
		zap.Int("edit_number", history.EditNumber),
	)

	// TODO: Broadcast WebSocket event

	c.JSON(200, models.SuccessResponse(message))
}

// MessageDelete deletes a message
func (h *Handler) MessageDelete(c *gin.Context, req map[string]interface{}, claims *models.JWTClaims) {
	data, err := getDataMap(req)
	if err != nil {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeMissingParameter, "Missing data"))
		return
	}

	messageID, ok := getString(data, "id")
	if !ok || messageID == "" {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidParameter, "id is required"))
		return
	}

	// Get message to check ownership
	message, err := h.db.MessageRead(c.Request.Context(), messageID)
	if err != nil {
		if err == models.ErrNotFound {
			c.JSON(404, models.ErrorResponse(models.ErrorCodeNotFound, "Message not found"))
		} else {
			c.JSON(500, models.ErrorResponse(models.ErrorCodeDatabaseError, "Failed to read message"))
		}
		return
	}

	// Check if user is sender or room admin/owner
	if message.SenderID != claims.UserID {
		// Check if user is admin/owner of the room
		chatRoomIDStr := message.ChatRoomID.String()
		participant, err := h.db.ParticipantGet(c.Request.Context(), chatRoomIDStr, claims.UserID.String())
		if err != nil || (participant.Role != models.ParticipantRoleOwner && participant.Role != models.ParticipantRoleAdmin) {
			c.JSON(403, models.ErrorResponse(models.ErrorCodeForbidden, "Insufficient permissions"))
			return
		}
	}

	// Delete message
	if err := h.db.MessageDelete(c.Request.Context(), messageID); err != nil {
		logger.Error("Failed to delete message", zap.Error(err))
		c.JSON(500, models.ErrorResponse(models.ErrorCodeDatabaseError, "Failed to delete message"))
		return
	}

	logger.Info("Message deleted",
		zap.String("message_id", messageID),
		zap.String("deleted_by", claims.Username),
	)

	// TODO: Broadcast WebSocket event

	c.JSON(200, models.SuccessResponse(gin.H{"deleted": true}))
}

// MessageReply creates a reply to a message
func (h *Handler) MessageReply(c *gin.Context, req map[string]interface{}, claims *models.JWTClaims) {
	data, err := getDataMap(req)
	if err != nil {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeMissingParameter, "Missing data"))
		return
	}

	chatRoomIDStr, ok := getString(data, "chat_room_id")
	if !ok || chatRoomIDStr == "" {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidParameter, "chat_room_id is required"))
		return
	}

	parentIDStr, ok := getString(data, "parent_id")
	if !ok || parentIDStr == "" {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidParameter, "parent_id is required"))
		return
	}

	content, ok := getString(data, "content")
	if !ok || content == "" {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidParameter, "content is required"))
		return
	}

	chatRoomID, err := uuid.Parse(chatRoomIDStr)
	if err != nil {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidParameter, "Invalid chat_room_id"))
		return
	}

	parentID, err := uuid.Parse(parentIDStr)
	if err != nil {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidParameter, "Invalid parent_id"))
		return
	}

	// Check if user is participant
	_, err = h.db.ParticipantGet(c.Request.Context(), chatRoomIDStr, claims.UserID.String())
	if err != nil {
		c.JSON(403, models.ErrorResponse(models.ErrorCodeForbidden, "Not a participant of this chat room"))
		return
	}

	// Verify parent message exists
	_, err = h.db.MessageRead(c.Request.Context(), parentIDStr)
	if err != nil {
		logger.Error("Failed to read parent message", zap.Error(err))
		c.JSON(500, models.ErrorResponse(models.ErrorCodeDatabaseError, "Failed to read parent message"))
		return
	}

	// Create reply message
	message := &models.Message{
		ChatRoomID:    chatRoomID,
		SenderID:      claims.UserID,
		Type:          models.MessageTypeReply,
		Content:       content,
		ContentFormat: models.ContentFormatPlain,
		ParentID:      &parentID,
	}

	// Handle optional content format
	if contentFormat, ok := getString(data, "content_format"); ok {
		message.ContentFormat = models.ContentFormat(contentFormat)
	}

	// Save to database
	if err := h.db.MessageCreate(c.Request.Context(), message); err != nil {
		logger.Error("Failed to create reply", zap.Error(err))
		c.JSON(500, models.ErrorResponse(models.ErrorCodeDatabaseError, "Failed to create reply"))
		return
	}

	logger.Info("Reply created",
		zap.String("message_id", message.ID.String()),
		zap.String("parent_id", parentIDStr),
		zap.String("sender", claims.Username),
	)

	c.JSON(200, models.SuccessResponse(message))
}

// MessageQuote creates a quote of a message
func (h *Handler) MessageQuote(c *gin.Context, req map[string]interface{}, claims *models.JWTClaims) {
	data, err := getDataMap(req)
	if err != nil {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeMissingParameter, "Missing data"))
		return
	}

	// Inject quoted_message_id and type for quote
	data["type"] = string(models.MessageTypeQuote)
	req["data"] = data

	// Use MessageSend with quoted_message_id
	h.MessageSend(c, req, claims)
}

// MessageSearch performs full-text search
func (h *Handler) MessageSearch(c *gin.Context, req map[string]interface{}, claims *models.JWTClaims) {
	data, err := getDataMap(req)
	if err != nil {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeMissingParameter, "Missing data"))
		return
	}

	chatRoomIDStr, ok := getString(data, "chat_room_id")
	if !ok || chatRoomIDStr == "" {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidParameter, "id is required"))
		return
	}

	query, ok := getString(data, "query")
	if !ok || query == "" {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidParameter, "query is required"))
		return
	}

	// Check if user is participant
	_, err = h.db.ParticipantGet(c.Request.Context(), chatRoomIDStr, claims.UserID.String())
	if err != nil {
		c.JSON(403, models.ErrorResponse(models.ErrorCodeForbidden, "Not a participant of this chat room"))
		return
	}

	limit, _ := getInt(data, "limit")
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	offset, _ := getInt(data, "offset")
	if offset < 0 {
		offset = 0
	}

	// Search messages
	messages, total, err := h.db.MessageSearch(c.Request.Context(), chatRoomIDStr, query, limit, offset)
	if err != nil {
		logger.Error("Failed to search messages", zap.Error(err))
		c.JSON(500, models.ErrorResponse(models.ErrorCodeDatabaseError, "Failed to search messages"))
		return
	}

	response := models.ListResponse{
		Items: messages,
		Pagination: &models.PaginationMeta{
			Total:   total,
			Limit:   limit,
			Offset:  offset,
			HasMore: offset+limit < total,
		},
	}

	if offset+limit < total {
		response.Pagination.NextOffset = offset + limit
	}

	c.JSON(200, models.SuccessResponse(response))
}

// MessagePin pins a message
func (h *Handler) MessagePin(c *gin.Context, req map[string]interface{}, claims *models.JWTClaims) {
	data, err := getDataMap(req)
	if err != nil {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeMissingParameter, "Missing data"))
		return
	}

	messageID, ok := getString(data, "id")
	if !ok || messageID == "" {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidParameter, "id is required"))
		return
	}

	// Get message
	message, err := h.db.MessageRead(c.Request.Context(), messageID)
	if err != nil {
		if err == models.ErrNotFound {
			c.JSON(404, models.ErrorResponse(models.ErrorCodeNotFound, "Message not found"))
		} else {
			logger.Error("Failed to read message", zap.Error(err))
			c.JSON(500, models.ErrorResponse(models.ErrorCodeDatabaseError, "Failed to read message"))
		}
		return
	}

	// Check if user is admin/moderator of the room
	chatRoomIDStr := message.ChatRoomID.String()
	participant, err := h.db.ParticipantGet(c.Request.Context(), chatRoomIDStr, claims.UserID.String())
	if err != nil || (participant.Role != models.ParticipantRoleOwner &&
		participant.Role != models.ParticipantRoleAdmin &&
		participant.Role != models.ParticipantRoleModerator) {
		c.JSON(403, models.ErrorResponse(models.ErrorCodeForbidden, "Only admins and moderators can pin messages"))
		return
	}

	// Pin the message
	message.IsPinned = true
	message.PinnedBy = &claims.UserID

	if err := h.db.MessageUpdate(c.Request.Context(), message); err != nil {
		logger.Error("Failed to pin message", zap.Error(err))
		c.JSON(500, models.ErrorResponse(models.ErrorCodeDatabaseError, "Failed to pin message"))
		return
	}

	logger.Info("Message pinned",
		zap.String("message_id", messageID),
		zap.String("pinned_by", claims.Username),
	)

	c.JSON(200, models.SuccessResponse(message))
}

// MessageUnpin unpins a message
func (h *Handler) MessageUnpin(c *gin.Context, req map[string]interface{}, claims *models.JWTClaims) {
	data, err := getDataMap(req)
	if err != nil {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeMissingParameter, "Missing data"))
		return
	}

	messageID, ok := getString(data, "id")
	if !ok || messageID == "" {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidParameter, "id is required"))
		return
	}

	// Get message
	message, err := h.db.MessageRead(c.Request.Context(), messageID)
	if err != nil {
		if err == models.ErrNotFound {
			c.JSON(404, models.ErrorResponse(models.ErrorCodeNotFound, "Message not found"))
		} else {
			logger.Error("Failed to read message", zap.Error(err))
			c.JSON(500, models.ErrorResponse(models.ErrorCodeDatabaseError, "Failed to read message"))
		}
		return
	}

	// Check if user is admin/moderator of the room
	chatRoomIDStr := message.ChatRoomID.String()
	participant, err := h.db.ParticipantGet(c.Request.Context(), chatRoomIDStr, claims.UserID.String())
	if err != nil || (participant.Role != models.ParticipantRoleOwner &&
		participant.Role != models.ParticipantRoleAdmin &&
		participant.Role != models.ParticipantRoleModerator) {
		c.JSON(403, models.ErrorResponse(models.ErrorCodeForbidden, "Only admins and moderators can unpin messages"))
		return
	}

	// Unpin the message
	message.IsPinned = false
	message.PinnedBy = nil

	if err := h.db.MessageUpdate(c.Request.Context(), message); err != nil {
		logger.Error("Failed to unpin message", zap.Error(err))
		c.JSON(500, models.ErrorResponse(models.ErrorCodeDatabaseError, "Failed to unpin message"))
		return
	}

	logger.Info("Message unpinned",
		zap.String("message_id", messageID),
		zap.String("unpinned_by", claims.Username),
	)

	c.JSON(200, models.SuccessResponse(message))
}

// MessageGetEditHistory retrieves the complete edit history for a message
func (h *Handler) MessageGetEditHistory(c *gin.Context, req map[string]interface{}, claims *models.JWTClaims) {
	data, err := getDataMap(req)
	if err != nil {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeMissingParameter, "Missing data"))
		return
	}

	messageID, ok := getString(data, "id")
	if !ok || messageID == "" {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidParameter, "id is required"))
		return
	}

	// Get message to verify it exists and check permissions
	message, err := h.db.MessageRead(c.Request.Context(), messageID)
	if err != nil {
		if err == models.ErrNotFound {
			c.JSON(404, models.ErrorResponse(models.ErrorCodeNotFound, "Message not found"))
		} else {
			logger.Error("Failed to read message", zap.Error(err))
			c.JSON(500, models.ErrorResponse(models.ErrorCodeDatabaseError, "Failed to read message"))
		}
		return
	}

	// Check if user is participant of the chat room
	chatRoomIDStr := message.ChatRoomID.String()
	_, err = h.db.ParticipantGet(c.Request.Context(), chatRoomIDStr, claims.UserID.String())
	if err != nil {
		c.JSON(403, models.ErrorResponse(models.ErrorCodeForbidden, "Not a participant of this chat room"))
		return
	}

	// Get edit history
	editHistory, err := h.db.MessageEditHistoryList(c.Request.Context(), messageID)
	if err != nil {
		logger.Error("Failed to get edit history", zap.Error(err))
		c.JSON(500, models.ErrorResponse(models.ErrorCodeDatabaseError, "Failed to get edit history"))
		return
	}

	// Get total edit count
	totalEdits, err := h.db.MessageEditHistoryCount(c.Request.Context(), messageID)
	if err != nil {
		logger.Error("Failed to count edit history", zap.Error(err))
		totalEdits = len(editHistory)
	}

	// Build response with edit history
	// Initialize with non-nil empty slice to ensure JSON serializes as [] instead of null
	editHistoryResponses := []*models.MessageEditHistoryResponse{}

	// For now, we don't fetch editor info (would require external service call)
	// Client applications can fetch user info separately if needed
	for _, history := range editHistory {
		editHistoryResponses = append(editHistoryResponses, &models.MessageEditHistoryResponse{
			EditHistory: history,
			Editor:      nil, // TODO: Optionally fetch from Core service
		})
	}

	response := &models.MessageWithEditHistory{
		Message:     message,
		EditHistory: editHistoryResponses,
		TotalEdits:  totalEdits,
	}

	logger.Info("Edit history retrieved",
		zap.String("message_id", messageID),
		zap.Int("edit_count", totalEdits),
		zap.String("requested_by", claims.Username),
	)

	c.JSON(200, models.SuccessResponse(response))
}

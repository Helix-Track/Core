package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"helixtrack.ru/chat/internal/logger"
	"helixtrack.ru/chat/internal/models"
)

// TypingStart indicates user started typing
func (h *Handler) TypingStart(c *gin.Context, req map[string]interface{}, claims *models.JWTClaims) {
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

	// Upsert typing indicator
	indicator := &models.TypingIndicator{
		ChatRoomID: chatRoomID,
		UserID:     claims.UserID,
		IsTyping:   true,
	}

	if err := h.db.TypingUpsert(c.Request.Context(), indicator); err != nil {
		logger.Error("Failed to create typing indicator", zap.Error(err))
		c.JSON(500, models.ErrorResponse(models.ErrorCodeDatabaseError, "Failed to update typing status"))
		return
	}

	// TODO: Broadcast WebSocket event

	c.JSON(200, models.SuccessResponse(gin.H{"typing": true}))
}

// TypingStop indicates user stopped typing
func (h *Handler) TypingStop(c *gin.Context, req map[string]interface{}, claims *models.JWTClaims) {
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

	// Delete typing indicator
	if err := h.db.TypingDelete(c.Request.Context(), chatRoomIDStr, claims.UserID.String()); err != nil {
		logger.Error("Failed to delete typing indicator", zap.Error(err))
	}

	// TODO: Broadcast WebSocket event

	c.JSON(200, models.SuccessResponse(gin.H{"typing": false}))
}

// PresenceUpdate updates user presence
func (h *Handler) PresenceUpdate(c *gin.Context, req map[string]interface{}, claims *models.JWTClaims) {
	data, err := getDataMap(req)
	if err != nil {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeMissingParameter, "Missing data"))
		return
	}

	statusStr, ok := getString(data, "status")
	if !ok || statusStr == "" {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeMissingParameter, "status is required"))
		return
	}

	statusMessage, _ := getString(data, "status_message")

	// Update presence
	presence := &models.UserPresence{
		UserID:        claims.UserID,
		Status:        models.PresenceStatus(statusStr),
		StatusMessage: statusMessage,
	}

	// Validate status
	presReq := &models.PresenceRequest{
		Status:        models.PresenceStatus(statusStr),
		StatusMessage: statusMessage,
	}

	if err := presReq.Validate(); err != nil {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidParameter, err.Error()))
		return
	}

	if err := h.db.PresenceUpsert(c.Request.Context(), presence); err != nil {
		logger.Error("Failed to update presence", zap.Error(err))
		c.JSON(500, models.ErrorResponse(models.ErrorCodeDatabaseError, "Failed to update presence"))
		return
	}

	// TODO: Broadcast WebSocket event

	c.JSON(200, models.SuccessResponse(presence))
}

// PresenceGet gets user presence
func (h *Handler) PresenceGet(c *gin.Context, req map[string]interface{}, claims *models.JWTClaims) {
	data, err := getDataMap(req)
	if err != nil {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeMissingParameter, "Missing data"))
		return
	}

	userIDStr, ok := getString(data, "user_id")
	if !ok || userIDStr == "" {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidParameter, "id is required"))
		return
	}

	// Get presence
	presence, err := h.db.PresenceGet(c.Request.Context(), userIDStr)
	if err != nil {
		if err == models.ErrNotFound {
			c.JSON(404, models.ErrorResponse(models.ErrorCodeNotFound, "Presence not found"))
		} else {
			logger.Error("Failed to get presence", zap.Error(err))
			c.JSON(500, models.ErrorResponse(models.ErrorCodeDatabaseError, "Failed to get presence"))
		}
		return
	}

	c.JSON(200, models.SuccessResponse(presence))
}

// ReadReceiptMark marks a message as read
func (h *Handler) ReadReceiptMark(c *gin.Context, req map[string]interface{}, claims *models.JWTClaims) {
	data, err := getDataMap(req)
	if err != nil {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeMissingParameter, "Missing data"))
		return
	}

	messageIDStr, ok := getString(data, "message_id")
	if !ok || messageIDStr == "" {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidParameter, "id is required"))
		return
	}

	messageID, err := uuid.Parse(messageIDStr)
	if err != nil {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidParameter, "Invalid message_id"))
		return
	}

	// Create read receipt
	receipt := &models.MessageReadReceipt{
		MessageID: messageID,
		UserID:    claims.UserID,
	}

	if err := h.db.ReadReceiptCreate(c.Request.Context(), receipt); err != nil {
		logger.Error("Failed to create read receipt", zap.Error(err))
		c.JSON(500, models.ErrorResponse(models.ErrorCodeDatabaseError, "Failed to mark as read"))
		return
	}

	// TODO: Broadcast WebSocket event

	c.JSON(200, models.SuccessResponse(receipt))
}

// ReadReceiptGet gets read receipts for a message
func (h *Handler) ReadReceiptGet(c *gin.Context, req map[string]interface{}, claims *models.JWTClaims) {
	data, err := getDataMap(req)
	if err != nil {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeMissingParameter, "Missing data"))
		return
	}

	messageIDStr, ok := getString(data, "message_id")
	if !ok || messageIDStr == "" {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidParameter, "id is required"))
		return
	}

	// Get read receipts
	receipts, err := h.db.ReadReceiptGet(c.Request.Context(), messageIDStr)
	if err != nil {
		logger.Error("Failed to get read receipts", zap.Error(err))
		c.JSON(500, models.ErrorResponse(models.ErrorCodeDatabaseError, "Failed to get read receipts"))
		return
	}

	c.JSON(200, models.SuccessResponse(receipts))
}

// ReactionAdd adds a reaction to a message
func (h *Handler) ReactionAdd(c *gin.Context, req map[string]interface{}, claims *models.JWTClaims) {
	data, err := getDataMap(req)
	if err != nil {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeMissingParameter, "Missing data"))
		return
	}

	messageIDStr, ok := getString(data, "message_id")
	if !ok || messageIDStr == "" {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidParameter, "id is required"))
		return
	}

	emoji, ok := getString(data, "emoji")
	if !ok || emoji == "" {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeMissingParameter, "emoji is required"))
		return
	}

	messageID, err := uuid.Parse(messageIDStr)
	if err != nil {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidParameter, "Invalid message_id"))
		return
	}

	// Create reaction
	reaction := &models.MessageReaction{
		MessageID: messageID,
		UserID:    claims.UserID,
		Emoji:     emoji,
	}

	if err := h.db.ReactionCreate(c.Request.Context(), reaction); err != nil {
		logger.Error("Failed to create reaction", zap.Error(err))
		c.JSON(500, models.ErrorResponse(models.ErrorCodeDatabaseError, "Failed to add reaction"))
		return
	}

	logger.Info("Reaction added",
		zap.String("message_id", messageIDStr),
		zap.String("emoji", emoji),
		zap.String("user", claims.Username),
	)

	// TODO: Broadcast WebSocket event

	c.JSON(200, models.SuccessResponse(reaction))
}

// ReactionRemove removes a reaction from a message
func (h *Handler) ReactionRemove(c *gin.Context, req map[string]interface{}, claims *models.JWTClaims) {
	data, err := getDataMap(req)
	if err != nil {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeMissingParameter, "Missing data"))
		return
	}

	messageIDStr, ok := getString(data, "message_id")
	if !ok || messageIDStr == "" {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidParameter, "id is required"))
		return
	}

	emoji, ok := getString(data, "emoji")
	if !ok || emoji == "" {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeMissingParameter, "emoji is required"))
		return
	}

	// Delete reaction
	if err := h.db.ReactionDelete(c.Request.Context(), messageIDStr, claims.UserID.String(), emoji); err != nil {
		logger.Error("Failed to delete reaction", zap.Error(err))
		c.JSON(500, models.ErrorResponse(models.ErrorCodeDatabaseError, "Failed to remove reaction"))
		return
	}

	logger.Info("Reaction removed",
		zap.String("message_id", messageIDStr),
		zap.String("emoji", emoji),
		zap.String("user", claims.Username),
	)

	// TODO: Broadcast WebSocket event

	c.JSON(200, models.SuccessResponse(gin.H{"removed": true}))
}

// ReactionList lists reactions for a message
func (h *Handler) ReactionList(c *gin.Context, req map[string]interface{}, claims *models.JWTClaims) {
	data, err := getDataMap(req)
	if err != nil {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeMissingParameter, "Missing data"))
		return
	}

	messageIDStr, ok := getString(data, "message_id")
	if !ok || messageIDStr == "" {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidParameter, "id is required"))
		return
	}

	// Get reactions
	reactions, err := h.db.ReactionList(c.Request.Context(), messageIDStr)
	if err != nil {
		logger.Error("Failed to list reactions", zap.Error(err))
		c.JSON(500, models.ErrorResponse(models.ErrorCodeDatabaseError, "Failed to list reactions"))
		return
	}

	c.JSON(200, models.SuccessResponse(reactions))
}

// AttachmentUpload handles file upload
func (h *Handler) AttachmentUpload(c *gin.Context, req map[string]interface{}, claims *models.JWTClaims) {
	// TODO: Implement file upload with multipart/form-data
	c.JSON(501, models.ErrorResponse(models.ErrorCodeInternalError, "Not yet implemented"))
}

// AttachmentDelete deletes an attachment
func (h *Handler) AttachmentDelete(c *gin.Context, req map[string]interface{}, claims *models.JWTClaims) {
	data, err := getDataMap(req)
	if err != nil {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeMissingParameter, "Missing data"))
		return
	}

	attachmentID, ok := getString(data, "attachment_id")
	if !ok || attachmentID == "" {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidParameter, "id is required"))
		return
	}

	// Delete attachment
	if err := h.db.AttachmentDelete(c.Request.Context(), attachmentID); err != nil {
		if err == models.ErrNotFound {
			c.JSON(404, models.ErrorResponse(models.ErrorCodeNotFound, "Attachment not found"))
		} else {
			logger.Error("Failed to delete attachment", zap.Error(err))
			c.JSON(500, models.ErrorResponse(models.ErrorCodeDatabaseError, "Failed to delete attachment"))
		}
		return
	}

	logger.Info("Attachment deleted",
		zap.String("attachment_id", attachmentID),
		zap.String("deleted_by", claims.Username),
	)

	c.JSON(200, models.SuccessResponse(gin.H{"deleted": true}))
}

// AttachmentList lists attachments for a message
func (h *Handler) AttachmentList(c *gin.Context, req map[string]interface{}, claims *models.JWTClaims) {
	data, err := getDataMap(req)
	if err != nil {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeMissingParameter, "Missing data"))
		return
	}

	messageID, ok := getString(data, "message_id")
	if !ok || messageID == "" {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidParameter, "id is required"))
		return
	}

	// Get attachments
	attachments, err := h.db.AttachmentList(c.Request.Context(), messageID)
	if err != nil {
		logger.Error("Failed to list attachments", zap.Error(err))
		c.JSON(500, models.ErrorResponse(models.ErrorCodeDatabaseError, "Failed to list attachments"))
		return
	}

	c.JSON(200, models.SuccessResponse(attachments))
}

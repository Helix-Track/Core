package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"helixtrack.ru/chat/internal/logger"
	"helixtrack.ru/chat/internal/models"
)

// ParticipantAdd adds a user to a chat room
func (h *Handler) ParticipantAdd(c *gin.Context, req map[string]interface{}, claims *models.JWTClaims) {
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

	userIDStr, ok := getString(data, "user_id")
	if !ok || userIDStr == "" {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidParameter, "id is required"))
		return
	}

	chatRoomID, err := uuid.Parse(chatRoomIDStr)
	if err != nil {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidParameter, "Invalid chat_room_id"))
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidParameter, "Invalid user_id"))
		return
	}

	// Check if current user has permission (must be owner or admin)
	currentParticipant, err := h.db.ParticipantGet(c.Request.Context(), chatRoomIDStr, claims.UserID.String())
	if err != nil || (currentParticipant.Role != models.ParticipantRoleOwner && currentParticipant.Role != models.ParticipantRoleAdmin) {
		c.JSON(403, models.ErrorResponse(models.ErrorCodeForbidden, "Insufficient permissions"))
		return
	}

	// Determine role (default to member)
	role := models.ParticipantRoleMember
	if roleStr, ok := getString(data, "role"); ok {
		role = models.ParticipantRole(roleStr)
	}

	// Add participant
	participant := &models.ChatParticipant{
		ChatRoomID: chatRoomID,
		UserID:     userID,
		Role:       role,
		IsMuted:    false,
	}

	if err := h.db.ParticipantAdd(c.Request.Context(), participant); err != nil {
		logger.Error("Failed to add participant", zap.Error(err))
		c.JSON(500, models.ErrorResponse(models.ErrorCodeDatabaseError, "Failed to add participant"))
		return
	}

	logger.Info("Participant added",
		zap.String("chat_room_id", chatRoomIDStr),
		zap.String("user_id", userIDStr),
		zap.String("added_by", claims.Username),
	)

	// TODO: Broadcast WebSocket event

	c.JSON(200, models.SuccessResponse(participant))
}

// ParticipantRemove removes a user from a chat room
func (h *Handler) ParticipantRemove(c *gin.Context, req map[string]interface{}, claims *models.JWTClaims) {
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

	userIDStr, ok := getString(data, "user_id")
	if !ok || userIDStr == "" {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidParameter, "id is required"))
		return
	}

	// Check if trying to remove the owner
	targetParticipant, err := h.db.ParticipantGet(c.Request.Context(), chatRoomIDStr, userIDStr)
	if err == nil && targetParticipant.Role == models.ParticipantRoleOwner {
		c.JSON(403, models.ErrorResponse(models.ErrorCodeForbidden, "Cannot remove room owner"))
		return
	}

	// Check if current user has permission or is removing themselves
	if userIDStr != claims.UserID.String() {
		currentParticipant, err := h.db.ParticipantGet(c.Request.Context(), chatRoomIDStr, claims.UserID.String())
		if err != nil || (currentParticipant.Role != models.ParticipantRoleOwner && currentParticipant.Role != models.ParticipantRoleAdmin) {
			c.JSON(403, models.ErrorResponse(models.ErrorCodeForbidden, "Insufficient permissions"))
			return
		}
	}

	// Remove participant
	if err := h.db.ParticipantRemove(c.Request.Context(), chatRoomIDStr, userIDStr); err != nil {
		if err == models.ErrNotFound {
			c.JSON(404, models.ErrorResponse(models.ErrorCodeNotFound, "Participant not found"))
		} else {
			logger.Error("Failed to remove participant", zap.Error(err))
			c.JSON(500, models.ErrorResponse(models.ErrorCodeDatabaseError, "Failed to remove participant"))
		}
		return
	}

	logger.Info("Participant removed",
		zap.String("chat_room_id", chatRoomIDStr),
		zap.String("user_id", userIDStr),
		zap.String("removed_by", claims.Username),
	)

	// TODO: Broadcast WebSocket event

	c.JSON(200, models.SuccessResponse(gin.H{"removed": true}))
}

// ParticipantList lists all participants in a chat room
func (h *Handler) ParticipantList(c *gin.Context, req map[string]interface{}, claims *models.JWTClaims) {
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

	// Check if user is participant
	_, err = h.db.ParticipantGet(c.Request.Context(), chatRoomIDStr, claims.UserID.String())
	if err != nil {
		c.JSON(403, models.ErrorResponse(models.ErrorCodeForbidden, "Not a participant of this chat room"))
		return
	}

	// Get participants
	participants, err := h.db.ParticipantList(c.Request.Context(), chatRoomIDStr)
	if err != nil {
		logger.Error("Failed to list participants", zap.Error(err))
		c.JSON(500, models.ErrorResponse(models.ErrorCodeDatabaseError, "Failed to list participants"))
		return
	}

	c.JSON(200, models.SuccessResponse(gin.H{"items": participants}))
}

// ParticipantUpdateRole updates a participant's role
func (h *Handler) ParticipantUpdateRole(c *gin.Context, req map[string]interface{}, claims *models.JWTClaims) {
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

	userIDStr, ok := getString(data, "user_id")
	if !ok || userIDStr == "" {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidParameter, "id is required"))
		return
	}

	roleStr, ok := getString(data, "role")
	if !ok || roleStr == "" {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidParameter, "role is required"))
		return
	}

	// Check if current user has permission (must be owner or admin)
	currentParticipant, err := h.db.ParticipantGet(c.Request.Context(), chatRoomIDStr, claims.UserID.String())
	if err != nil || (currentParticipant.Role != models.ParticipantRoleOwner && currentParticipant.Role != models.ParticipantRoleAdmin) {
		c.JSON(403, models.ErrorResponse(models.ErrorCodeForbidden, "Only room owner or admin can change roles"))
		return
	}

	// Get participant to update
	participant, err := h.db.ParticipantGet(c.Request.Context(), chatRoomIDStr, userIDStr)
	if err != nil {
		if err == models.ErrNotFound {
			c.JSON(404, models.ErrorResponse(models.ErrorCodeNotFound, "Participant not found"))
		} else {
			c.JSON(500, models.ErrorResponse(models.ErrorCodeDatabaseError, "Failed to get participant"))
		}
		return
	}

	// Update role
	participant.Role = models.ParticipantRole(roleStr)

	if err := h.db.ParticipantUpdate(c.Request.Context(), participant); err != nil {
		logger.Error("Failed to update participant role", zap.Error(err))
		c.JSON(500, models.ErrorResponse(models.ErrorCodeDatabaseError, "Failed to update role"))
		return
	}

	logger.Info("Participant role updated",
		zap.String("chat_room_id", chatRoomIDStr),
		zap.String("user_id", userIDStr),
		zap.String("new_role", roleStr),
		zap.String("updated_by", claims.Username),
	)

	// TODO: Broadcast WebSocket event

	c.JSON(200, models.SuccessResponse(participant))
}

// ParticipantMute mutes a participant
func (h *Handler) ParticipantMute(c *gin.Context, req map[string]interface{}, claims *models.JWTClaims) {
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

	userIDStr, ok := getString(data, "user_id")
	if !ok || userIDStr == "" {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidParameter, "id is required"))
		return
	}

	// Check if current user has permission (must be moderator, admin, or owner)
	currentParticipant, err := h.db.ParticipantGet(c.Request.Context(), chatRoomIDStr, claims.UserID.String())
	if err != nil {
		c.JSON(403, models.ErrorResponse(models.ErrorCodeForbidden, "Not a participant"))
		return
	}

	if currentParticipant.Role != models.ParticipantRoleOwner &&
		currentParticipant.Role != models.ParticipantRoleAdmin &&
		currentParticipant.Role != models.ParticipantRoleModerator {
		c.JSON(403, models.ErrorResponse(models.ErrorCodeForbidden, "Insufficient permissions"))
		return
	}

	// Get participant to mute
	participant, err := h.db.ParticipantGet(c.Request.Context(), chatRoomIDStr, userIDStr)
	if err != nil {
		if err == models.ErrNotFound {
			c.JSON(404, models.ErrorResponse(models.ErrorCodeNotFound, "Participant not found"))
		} else {
			c.JSON(500, models.ErrorResponse(models.ErrorCodeDatabaseError, "Failed to get participant"))
		}
		return
	}

	// Mute participant
	participant.IsMuted = true

	if err := h.db.ParticipantUpdate(c.Request.Context(), participant); err != nil {
		logger.Error("Failed to mute participant", zap.Error(err))
		c.JSON(500, models.ErrorResponse(models.ErrorCodeDatabaseError, "Failed to mute participant"))
		return
	}

	logger.Info("Participant muted",
		zap.String("chat_room_id", chatRoomIDStr),
		zap.String("user_id", userIDStr),
		zap.String("muted_by", claims.Username),
	)

	c.JSON(200, models.SuccessResponse(participant))
}

// ParticipantUnmute unmutes a participant
func (h *Handler) ParticipantUnmute(c *gin.Context, req map[string]interface{}, claims *models.JWTClaims) {
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

	userIDStr, ok := getString(data, "user_id")
	if !ok || userIDStr == "" {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidParameter, "id is required"))
		return
	}

	// Check if current user has permission (must be moderator, admin, or owner)
	currentParticipant, err := h.db.ParticipantGet(c.Request.Context(), chatRoomIDStr, claims.UserID.String())
	if err != nil {
		c.JSON(403, models.ErrorResponse(models.ErrorCodeForbidden, "Not a participant"))
		return
	}

	if currentParticipant.Role != models.ParticipantRoleOwner &&
		currentParticipant.Role != models.ParticipantRoleAdmin &&
		currentParticipant.Role != models.ParticipantRoleModerator {
		c.JSON(403, models.ErrorResponse(models.ErrorCodeForbidden, "Insufficient permissions"))
		return
	}

	// Get participant to unmute
	participant, err := h.db.ParticipantGet(c.Request.Context(), chatRoomIDStr, userIDStr)
	if err != nil {
		if err == models.ErrNotFound {
			c.JSON(404, models.ErrorResponse(models.ErrorCodeNotFound, "Participant not found"))
		} else {
			c.JSON(500, models.ErrorResponse(models.ErrorCodeDatabaseError, "Failed to get participant"))
		}
		return
	}

	// Unmute participant
	participant.IsMuted = false

	if err := h.db.ParticipantUpdate(c.Request.Context(), participant); err != nil {
		logger.Error("Failed to unmute participant", zap.Error(err))
		c.JSON(500, models.ErrorResponse(models.ErrorCodeDatabaseError, "Failed to unmute participant"))
		return
	}

	logger.Info("Participant unmuted",
		zap.String("chat_room_id", chatRoomIDStr),
		zap.String("user_id", userIDStr),
		zap.String("unmuted_by", claims.Username),
	)

	c.JSON(200, models.SuccessResponse(participant))
}

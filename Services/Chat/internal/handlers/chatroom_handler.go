package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"helixtrack.ru/chat/internal/logger"
	"helixtrack.ru/chat/internal/models"
)

// ChatRoomCreate creates a new chat room
func (h *Handler) ChatRoomCreate(c *gin.Context, req map[string]interface{}, claims *models.JWTClaims) {
	data, err := getDataMap(req)
	if err != nil {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeMissingParameter, "Missing data"))
		return
	}

	// Parse request
	name, _ := getString(data, "name")
	description, _ := getString(data, "description")
	roomType, _ := getString(data, "type")
	entityType, _ := getString(data, "entity_type")
	entityIDStr, _ := getString(data, "entity_id")
	isPrivate, _ := getBool(data, "is_private")

	if name == "" || roomType == "" {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidParameter, "name and type are required"))
		return
	}

	// Create chat room
	room := &models.ChatRoom{
		Name:        name,
		Description: description,
		Type:        models.ChatRoomType(roomType),
		EntityType:  entityType,
		CreatedBy:   claims.UserID,
		IsPrivate:   isPrivate,
		IsArchived:  false,
	}

	if entityIDStr != "" {
		entityID, err := uuid.Parse(entityIDStr)
		if err != nil {
			c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidParameter, "Invalid entity_id"))
			return
		}
		room.EntityID = &entityID
	}

	// Validate room type
	roomReq := &models.ChatRoomRequest{
		Name: name,
		Type: models.ChatRoomType(roomType),
	}
	if err := roomReq.Validate(); err != nil {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidParameter, err.Error()))
		return
	}

	// Create in database
	if err := h.db.ChatRoomCreate(c.Request.Context(), room); err != nil {
		logger.Error("Failed to create chat room", zap.Error(err))
		c.JSON(500, models.ErrorResponse(models.ErrorCodeDatabaseError, "Failed to create chat room"))
		return
	}

	// Add creator as owner participant
	participant := &models.ChatParticipant{
		ChatRoomID: room.ID,
		UserID:     claims.UserID,
		Role:       models.ParticipantRoleOwner,
		IsMuted:    false,
	}

	if err := h.db.ParticipantAdd(c.Request.Context(), participant); err != nil {
		logger.Error("Failed to add creator as participant", zap.Error(err))
	}

	logger.Info("Chat room created",
		zap.String("room_id", room.ID.String()),
		zap.String("name", room.Name),
		zap.String("created_by", claims.Username),
	)

	c.JSON(200, models.SuccessResponse(room))
}

// ChatRoomRead retrieves a chat room
func (h *Handler) ChatRoomRead(c *gin.Context, req map[string]interface{}, claims *models.JWTClaims) {
	data, err := getDataMap(req)
	if err != nil {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeMissingParameter, "Missing data"))
		return
	}

	roomID, ok := getString(data, "id")
	if !ok || roomID == "" {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidParameter, "id is required"))
		return
	}

	// Get room
	room, err := h.db.ChatRoomRead(c.Request.Context(), roomID)
	if err != nil {
		if err == models.ErrNotFound {
			c.JSON(404, models.ErrorResponse(models.ErrorCodeNotFound, "Chat room not found"))
		} else {
			logger.Error("Failed to read chat room", zap.Error(err))
			c.JSON(500, models.ErrorResponse(models.ErrorCodeDatabaseError, "Failed to read chat room"))
		}
		return
	}

	// Check if user is participant
	_, err = h.db.ParticipantGet(c.Request.Context(), roomID, claims.UserID.String())
	if err != nil {
		c.JSON(403, models.ErrorResponse(models.ErrorCodeForbidden, "Not a participant of this chat room"))
		return
	}

	c.JSON(200, models.SuccessResponse(room))
}

// ChatRoomList lists chat rooms
func (h *Handler) ChatRoomList(c *gin.Context, req map[string]interface{}, claims *models.JWTClaims) {
	data, _ := getDataMap(req)

	limit, _ := getInt(data, "limit")
	offset, _ := getInt(data, "offset")

	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	// Get rooms
	rooms, total, err := h.db.ChatRoomList(c.Request.Context(), limit, offset)
	if err != nil {
		logger.Error("Failed to list chat rooms", zap.Error(err))
		c.JSON(500, models.ErrorResponse(models.ErrorCodeDatabaseError, "Failed to list chat rooms"))
		return
	}

	response := models.ListResponse{
		Items: rooms,
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

// ChatRoomUpdate updates a chat room
func (h *Handler) ChatRoomUpdate(c *gin.Context, req map[string]interface{}, claims *models.JWTClaims) {
	data, err := getDataMap(req)
	if err != nil {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeMissingParameter, "Missing data"))
		return
	}

	roomID, ok := getString(data, "id")
	if !ok || roomID == "" {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidParameter, "id is required"))
		return
	}

	// Get existing room
	room, err := h.db.ChatRoomRead(c.Request.Context(), roomID)
	if err != nil {
		if err == models.ErrNotFound {
			c.JSON(404, models.ErrorResponse(models.ErrorCodeNotFound, "Chat room not found"))
		} else {
			c.JSON(500, models.ErrorResponse(models.ErrorCodeDatabaseError, "Failed to read chat room"))
		}
		return
	}

	// Check if user has permission (must be owner or admin)
	participant, err := h.db.ParticipantGet(c.Request.Context(), roomID, claims.UserID.String())
	if err != nil || (participant.Role != models.ParticipantRoleOwner && participant.Role != models.ParticipantRoleAdmin) {
		c.JSON(403, models.ErrorResponse(models.ErrorCodeForbidden, "Insufficient permissions"))
		return
	}

	// Update fields
	if name, ok := getString(data, "name"); ok {
		room.Name = name
	}
	if description, ok := getString(data, "description"); ok {
		room.Description = description
	}
	if isPrivate, ok := getBool(data, "is_private"); ok {
		room.IsPrivate = isPrivate
	}
	if isArchived, ok := getBool(data, "is_archived"); ok {
		room.IsArchived = isArchived
	}

	// Update in database
	if err := h.db.ChatRoomUpdate(c.Request.Context(), room); err != nil {
		logger.Error("Failed to update chat room", zap.Error(err))
		c.JSON(500, models.ErrorResponse(models.ErrorCodeDatabaseError, "Failed to update chat room"))
		return
	}

	logger.Info("Chat room updated",
		zap.String("room_id", roomID),
		zap.String("updated_by", claims.Username),
	)

	c.JSON(200, models.SuccessResponse(room))
}

// ChatRoomDelete deletes a chat room
func (h *Handler) ChatRoomDelete(c *gin.Context, req map[string]interface{}, claims *models.JWTClaims) {
	data, err := getDataMap(req)
	if err != nil {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeMissingParameter, "Missing data"))
		return
	}

	roomID, ok := getString(data, "id")
	if !ok || roomID == "" {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidParameter, "id is required"))
		return
	}

	// Check if user has permission (must be owner)
	participant, err := h.db.ParticipantGet(c.Request.Context(), roomID, claims.UserID.String())
	if err != nil || participant.Role != models.ParticipantRoleOwner {
		c.JSON(403, models.ErrorResponse(models.ErrorCodeForbidden, "Only room owner can delete"))
		return
	}

	// Delete room
	if err := h.db.ChatRoomDelete(c.Request.Context(), roomID); err != nil {
		if err == models.ErrNotFound {
			c.JSON(404, models.ErrorResponse(models.ErrorCodeNotFound, "Chat room not found"))
		} else {
			logger.Error("Failed to delete chat room", zap.Error(err))
			c.JSON(500, models.ErrorResponse(models.ErrorCodeDatabaseError, "Failed to delete chat room"))
		}
		return
	}

	logger.Info("Chat room deleted",
		zap.String("room_id", roomID),
		zap.String("deleted_by", claims.Username),
	)

	c.JSON(200, models.SuccessResponse(gin.H{"deleted": true}))
}

// ChatRoomGetByEntity gets chat room by entity
func (h *Handler) ChatRoomGetByEntity(c *gin.Context, req map[string]interface{}, claims *models.JWTClaims) {
	data, err := getDataMap(req)
	if err != nil {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeMissingParameter, "Missing data"))
		return
	}

	entityType, ok := getString(data, "entity_type")
	if !ok || entityType == "" {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidParameter, "entity_type is required"))
		return
	}

	entityID, ok := getString(data, "entity_id")
	if !ok || entityID == "" {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidParameter, "entity_id is required"))
		return
	}

	// Get room
	room, err := h.db.ChatRoomGetByEntity(c.Request.Context(), entityType, entityID)
	if err != nil {
		if err == models.ErrNotFound {
			c.JSON(404, models.ErrorResponse(models.ErrorCodeNotFound, "Chat room not found"))
		} else {
			logger.Error("Failed to get chat room by entity", zap.Error(err))
			c.JSON(500, models.ErrorResponse(models.ErrorCodeDatabaseError, "Failed to get chat room"))
		}
		return
	}

	c.JSON(200, models.SuccessResponse(room))
}

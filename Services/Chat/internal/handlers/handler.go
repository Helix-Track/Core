package handlers

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"helixtrack.ru/chat/internal/database"
	"helixtrack.ru/chat/internal/logger"
	"helixtrack.ru/chat/internal/middleware"
	"helixtrack.ru/chat/internal/models"
	"helixtrack.ru/chat/internal/services"
)

// Handler handles all API requests
type Handler struct {
	db          database.Database
	coreService services.CoreService
}

// NewHandler creates a new handler
func NewHandler(db database.Database, coreService services.CoreService) *Handler {
	return &Handler{
		db:          db,
		coreService: coreService,
	}
}

// DoAction handles the unified /do endpoint
func (h *Handler) DoAction(c *gin.Context) {
	// Parse request
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid request body", zap.Error(err))
		c.JSON(400, models.ErrorResponse(models.ErrorCodeInvalidRequest, "Invalid request body"))
		return
	}

	// Extract action
	action, ok := req["action"].(string)
	if !ok || action == "" {
		c.JSON(400, models.ErrorResponse(models.ErrorCodeMissingParameter, "Missing action parameter"))
		return
	}

	// Get user info from JWT
	claims, err := middleware.GetClaims(c)
	if err != nil {
		c.JSON(401, models.ErrorResponse(models.ErrorCodeUnauthorized, "Unauthorized"))
		return
	}

	logger.Debug("Processing action",
		zap.String("action", action),
		zap.String("username", claims.Username),
	)

	// Route to appropriate handler
	switch action {
	// Chat Room actions
	case "chatRoomCreate":
		h.ChatRoomCreate(c, req, claims)
	case "chatRoomRead":
		h.ChatRoomRead(c, req, claims)
	case "chatRoomList":
		h.ChatRoomList(c, req, claims)
	case "chatRoomUpdate":
		h.ChatRoomUpdate(c, req, claims)
	case "chatRoomDelete":
		h.ChatRoomDelete(c, req, claims)
	case "chatRoomGetByEntity":
		h.ChatRoomGetByEntity(c, req, claims)

	// Message actions
	case "messageSend":
		h.MessageSend(c, req, claims)
	case "messageList":
		h.MessageList(c, req, claims)
	case "messageRead":
		h.MessageRead(c, req, claims)
	case "messageUpdate":
		h.MessageUpdate(c, req, claims)
	case "messageDelete":
		h.MessageDelete(c, req, claims)
	case "messageReply":
		h.MessageReply(c, req, claims)
	case "messageQuote":
		h.MessageQuote(c, req, claims)
	case "messageSearch":
		h.MessageSearch(c, req, claims)
	case "messagePin":
		h.MessagePin(c, req, claims)
	case "messageUnpin":
		h.MessageUnpin(c, req, claims)
	case "messageGetEditHistory":
		h.MessageGetEditHistory(c, req, claims)

	// Participant actions
	case "participantAdd":
		h.ParticipantAdd(c, req, claims)
	case "participantRemove":
		h.ParticipantRemove(c, req, claims)
	case "participantList":
		h.ParticipantList(c, req, claims)
	case "participantUpdateRole":
		h.ParticipantUpdateRole(c, req, claims)
	case "participantMute":
		h.ParticipantMute(c, req, claims)
	case "participantUnmute":
		h.ParticipantUnmute(c, req, claims)

	// Real-time actions
	case "typingStart":
		h.TypingStart(c, req, claims)
	case "typingStop":
		h.TypingStop(c, req, claims)
	case "presenceUpdate":
		h.PresenceUpdate(c, req, claims)
	case "presenceGet":
		h.PresenceGet(c, req, claims)
	case "readReceiptMark":
		h.ReadReceiptMark(c, req, claims)
	case "readReceiptGet":
		h.ReadReceiptGet(c, req, claims)
	case "reactionAdd":
		h.ReactionAdd(c, req, claims)
	case "reactionRemove":
		h.ReactionRemove(c, req, claims)
	case "reactionList":
		h.ReactionList(c, req, claims)

	// Attachment actions
	case "attachmentUpload":
		h.AttachmentUpload(c, req, claims)
	case "attachmentDelete":
		h.AttachmentDelete(c, req, claims)
	case "attachmentList":
		h.AttachmentList(c, req, claims)

	default:
		c.JSON(400, models.ErrorResponse(
			models.ErrorCodeInvalidParameter,
			"Unknown action: "+action,
		))
	}
}

// getDataMap extracts data map from request
func getDataMap(req map[string]interface{}) (map[string]interface{}, error) {
	data, ok := req["data"].(map[string]interface{})
	if !ok {
		return nil, models.ErrMissingParameter
	}
	return data, nil
}

// getString extracts string value from map
func getString(m map[string]interface{}, key string) (string, bool) {
	val, ok := m[key].(string)
	return val, ok
}

// getInt extracts int value from map
func getInt(m map[string]interface{}, key string) (int, bool) {
	switch v := m[key].(type) {
	case int:
		return v, true
	case float64:
		return int(v), true
	case string:
		// Try to parse string to int
		return 0, false
	default:
		return 0, false
	}
}

// getBool extracts bool value from map
func getBool(m map[string]interface{}, key string) (bool, bool) {
	val, ok := m[key].(bool)
	return val, ok
}

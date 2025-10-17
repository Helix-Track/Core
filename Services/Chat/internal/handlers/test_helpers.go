package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"helixtrack.ru/chat/internal/models"
)

// MockDatabase implements the database.Database interface for testing
type MockDatabase struct {
	// Chat Room methods
	ChatRoomCreateFunc      func(ctx context.Context, room *models.ChatRoom) error
	ChatRoomReadFunc        func(ctx context.Context, id string) (*models.ChatRoom, error)
	ChatRoomListFunc        func(ctx context.Context, limit, offset int) ([]*models.ChatRoom, int, error)
	ChatRoomUpdateFunc      func(ctx context.Context, room *models.ChatRoom) error
	ChatRoomDeleteFunc      func(ctx context.Context, id string) error
	ChatRoomGetByEntityFunc func(ctx context.Context, entityType, entityID string) (*models.ChatRoom, error)

	// Message methods
	MessageCreateFunc func(ctx context.Context, message *models.Message) error
	MessageReadFunc   func(ctx context.Context, id string) (*models.Message, error)
	MessageListFunc   func(ctx context.Context, req *models.MessageListRequest) ([]*models.Message, int, error)
	MessageUpdateFunc func(ctx context.Context, message *models.Message) error
	MessageDeleteFunc func(ctx context.Context, id string) error
	MessageSearchFunc func(ctx context.Context, chatRoomID, query string, limit, offset int) ([]*models.Message, int, error)

	// Participant methods
	ParticipantAddFunc        func(ctx context.Context, participant *models.ChatParticipant) error
	ParticipantRemoveFunc     func(ctx context.Context, chatRoomID, userID string) error
	ParticipantGetFunc        func(ctx context.Context, chatRoomID, userID string) (*models.ChatParticipant, error)
	ParticipantListFunc       func(ctx context.Context, chatRoomID string) ([]*models.ChatParticipant, error)
	ParticipantUpdateFunc     func(ctx context.Context, participant *models.ChatParticipant) error
	ParticipantUpdateRoleFunc func(ctx context.Context, chatRoomID, userID string, role models.ParticipantRole) error
	ParticipantMuteFunc       func(ctx context.Context, chatRoomID, userID string) error
	ParticipantUnmuteFunc     func(ctx context.Context, chatRoomID, userID string) error

	// Presence methods
	PresenceUpsertFunc      func(ctx context.Context, presence *models.UserPresence) error
	PresenceGetFunc         func(ctx context.Context, userID string) (*models.UserPresence, error)
	PresenceGetMultipleFunc func(ctx context.Context, userIDs []string) ([]*models.UserPresence, error)

	// Typing indicator methods
	TypingUpsertFunc    func(ctx context.Context, indicator *models.TypingIndicator) error
	TypingDeleteFunc    func(ctx context.Context, chatRoomID, userID string) error
	TypingGetActiveFunc func(ctx context.Context, chatRoomID string) ([]*models.TypingIndicator, error)

	// Read receipt methods
	ReadReceiptCreateFunc    func(ctx context.Context, receipt *models.MessageReadReceipt) error
	ReadReceiptGetFunc       func(ctx context.Context, messageID string) ([]*models.MessageReadReceipt, error)
	ReadReceiptGetByUserFunc func(ctx context.Context, messageID, userID string) (*models.MessageReadReceipt, error)

	// Reaction methods
	ReactionCreateFunc func(ctx context.Context, reaction *models.MessageReaction) error
	ReactionDeleteFunc func(ctx context.Context, messageID, userID, emoji string) error
	ReactionListFunc   func(ctx context.Context, messageID string) ([]*models.MessageReaction, error)

	// Attachment methods
	AttachmentCreateFunc func(ctx context.Context, attachment *models.MessageAttachment) error
	AttachmentDeleteFunc func(ctx context.Context, id string) error
	AttachmentListFunc   func(ctx context.Context, messageID string) ([]*models.MessageAttachment, error)

	// Message edit history methods
	MessageEditHistoryCreateFunc func(ctx context.Context, history *models.MessageEditHistory) error
	MessageEditHistoryListFunc   func(ctx context.Context, messageID string) ([]*models.MessageEditHistory, error)
	MessageEditHistoryGetFunc    func(ctx context.Context, id string) (*models.MessageEditHistory, error)
	MessageEditHistoryCountFunc  func(ctx context.Context, messageID string) (int, error)

	// Other methods
	CloseFunc   func() error
	PingFunc    func() error
	BeginTxFunc func(ctx context.Context) (*sql.Tx, error)
}

// Implement database.Database interface
func (m *MockDatabase) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

func (m *MockDatabase) Ping() error {
	if m.PingFunc != nil {
		return m.PingFunc()
	}
	return nil
}

func (m *MockDatabase) BeginTx(ctx context.Context) (*sql.Tx, error) {
	if m.BeginTxFunc != nil {
		return m.BeginTxFunc(ctx)
	}
	return nil, nil
}

// Chat Room methods
func (m *MockDatabase) ChatRoomCreate(ctx context.Context, room *models.ChatRoom) error {
	if m.ChatRoomCreateFunc != nil {
		return m.ChatRoomCreateFunc(ctx, room)
	}
	return nil
}

func (m *MockDatabase) ChatRoomRead(ctx context.Context, id string) (*models.ChatRoom, error) {
	if m.ChatRoomReadFunc != nil {
		return m.ChatRoomReadFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockDatabase) ChatRoomList(ctx context.Context, limit, offset int) ([]*models.ChatRoom, int, error) {
	if m.ChatRoomListFunc != nil {
		return m.ChatRoomListFunc(ctx, limit, offset)
	}
	return nil, 0, nil
}

func (m *MockDatabase) ChatRoomUpdate(ctx context.Context, room *models.ChatRoom) error {
	if m.ChatRoomUpdateFunc != nil {
		return m.ChatRoomUpdateFunc(ctx, room)
	}
	return nil
}

func (m *MockDatabase) ChatRoomDelete(ctx context.Context, id string) error {
	if m.ChatRoomDeleteFunc != nil {
		return m.ChatRoomDeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockDatabase) ChatRoomGetByEntity(ctx context.Context, entityType, entityID string) (*models.ChatRoom, error) {
	if m.ChatRoomGetByEntityFunc != nil {
		return m.ChatRoomGetByEntityFunc(ctx, entityType, entityID)
	}
	return nil, nil
}

// Message methods
func (m *MockDatabase) MessageCreate(ctx context.Context, message *models.Message) error {
	if m.MessageCreateFunc != nil {
		return m.MessageCreateFunc(ctx, message)
	}
	return nil
}

func (m *MockDatabase) MessageRead(ctx context.Context, id string) (*models.Message, error) {
	if m.MessageReadFunc != nil {
		return m.MessageReadFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockDatabase) MessageList(ctx context.Context, req *models.MessageListRequest) ([]*models.Message, int, error) {
	if m.MessageListFunc != nil {
		return m.MessageListFunc(ctx, req)
	}
	return nil, 0, nil
}

func (m *MockDatabase) MessageUpdate(ctx context.Context, message *models.Message) error {
	if m.MessageUpdateFunc != nil {
		return m.MessageUpdateFunc(ctx, message)
	}
	return nil
}

func (m *MockDatabase) MessageDelete(ctx context.Context, id string) error {
	if m.MessageDeleteFunc != nil {
		return m.MessageDeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockDatabase) MessageSearch(ctx context.Context, chatRoomID, query string, limit, offset int) ([]*models.Message, int, error) {
	if m.MessageSearchFunc != nil {
		return m.MessageSearchFunc(ctx, chatRoomID, query, limit, offset)
	}
	return nil, 0, nil
}

// Participant methods
func (m *MockDatabase) ParticipantAdd(ctx context.Context, participant *models.ChatParticipant) error {
	if m.ParticipantAddFunc != nil {
		return m.ParticipantAddFunc(ctx, participant)
	}
	return nil
}

func (m *MockDatabase) ParticipantRemove(ctx context.Context, chatRoomID, userID string) error {
	if m.ParticipantRemoveFunc != nil {
		return m.ParticipantRemoveFunc(ctx, chatRoomID, userID)
	}
	return nil
}

func (m *MockDatabase) ParticipantGet(ctx context.Context, chatRoomID, userID string) (*models.ChatParticipant, error) {
	if m.ParticipantGetFunc != nil {
		return m.ParticipantGetFunc(ctx, chatRoomID, userID)
	}
	return nil, nil
}

func (m *MockDatabase) ParticipantList(ctx context.Context, chatRoomID string) ([]*models.ChatParticipant, error) {
	if m.ParticipantListFunc != nil {
		return m.ParticipantListFunc(ctx, chatRoomID)
	}
	return nil, nil
}

func (m *MockDatabase) ParticipantUpdate(ctx context.Context, participant *models.ChatParticipant) error {
	if m.ParticipantUpdateFunc != nil {
		return m.ParticipantUpdateFunc(ctx, participant)
	}
	return nil
}

func (m *MockDatabase) ParticipantUpdateRole(ctx context.Context, chatRoomID, userID string, role models.ParticipantRole) error {
	if m.ParticipantUpdateRoleFunc != nil {
		return m.ParticipantUpdateRoleFunc(ctx, chatRoomID, userID, role)
	}
	return nil
}

func (m *MockDatabase) ParticipantMute(ctx context.Context, chatRoomID, userID string) error {
	if m.ParticipantMuteFunc != nil {
		return m.ParticipantMuteFunc(ctx, chatRoomID, userID)
	}
	return nil
}

func (m *MockDatabase) ParticipantUnmute(ctx context.Context, chatRoomID, userID string) error {
	if m.ParticipantUnmuteFunc != nil {
		return m.ParticipantUnmuteFunc(ctx, chatRoomID, userID)
	}
	return nil
}

// Presence methods
func (m *MockDatabase) PresenceUpsert(ctx context.Context, presence *models.UserPresence) error {
	if m.PresenceUpsertFunc != nil {
		return m.PresenceUpsertFunc(ctx, presence)
	}
	return nil
}

func (m *MockDatabase) PresenceGet(ctx context.Context, userID string) (*models.UserPresence, error) {
	if m.PresenceGetFunc != nil {
		return m.PresenceGetFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockDatabase) PresenceGetMultiple(ctx context.Context, userIDs []string) ([]*models.UserPresence, error) {
	if m.PresenceGetMultipleFunc != nil {
		return m.PresenceGetMultipleFunc(ctx, userIDs)
	}
	return nil, nil
}

// Typing indicator methods
func (m *MockDatabase) TypingUpsert(ctx context.Context, indicator *models.TypingIndicator) error {
	if m.TypingUpsertFunc != nil {
		return m.TypingUpsertFunc(ctx, indicator)
	}
	return nil
}

func (m *MockDatabase) TypingDelete(ctx context.Context, chatRoomID, userID string) error {
	if m.TypingDeleteFunc != nil {
		return m.TypingDeleteFunc(ctx, chatRoomID, userID)
	}
	return nil
}

func (m *MockDatabase) TypingGetActive(ctx context.Context, chatRoomID string) ([]*models.TypingIndicator, error) {
	if m.TypingGetActiveFunc != nil {
		return m.TypingGetActiveFunc(ctx, chatRoomID)
	}
	return nil, nil
}

// Read receipt methods
func (m *MockDatabase) ReadReceiptCreate(ctx context.Context, receipt *models.MessageReadReceipt) error {
	if m.ReadReceiptCreateFunc != nil {
		return m.ReadReceiptCreateFunc(ctx, receipt)
	}
	return nil
}

func (m *MockDatabase) ReadReceiptGet(ctx context.Context, messageID string) ([]*models.MessageReadReceipt, error) {
	if m.ReadReceiptGetFunc != nil {
		return m.ReadReceiptGetFunc(ctx, messageID)
	}
	return nil, nil
}

func (m *MockDatabase) ReadReceiptGetByUser(ctx context.Context, messageID, userID string) (*models.MessageReadReceipt, error) {
	if m.ReadReceiptGetByUserFunc != nil {
		return m.ReadReceiptGetByUserFunc(ctx, messageID, userID)
	}
	return nil, nil
}

// Reaction methods
func (m *MockDatabase) ReactionCreate(ctx context.Context, reaction *models.MessageReaction) error {
	if m.ReactionCreateFunc != nil {
		return m.ReactionCreateFunc(ctx, reaction)
	}
	return nil
}

func (m *MockDatabase) ReactionDelete(ctx context.Context, messageID, userID, emoji string) error {
	if m.ReactionDeleteFunc != nil {
		return m.ReactionDeleteFunc(ctx, messageID, userID, emoji)
	}
	return nil
}

func (m *MockDatabase) ReactionList(ctx context.Context, messageID string) ([]*models.MessageReaction, error) {
	if m.ReactionListFunc != nil {
		return m.ReactionListFunc(ctx, messageID)
	}
	return nil, nil
}

// Attachment methods
func (m *MockDatabase) AttachmentCreate(ctx context.Context, attachment *models.MessageAttachment) error {
	if m.AttachmentCreateFunc != nil {
		return m.AttachmentCreateFunc(ctx, attachment)
	}
	return nil
}

func (m *MockDatabase) AttachmentDelete(ctx context.Context, id string) error {
	if m.AttachmentDeleteFunc != nil {
		return m.AttachmentDeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockDatabase) AttachmentList(ctx context.Context, messageID string) ([]*models.MessageAttachment, error) {
	if m.AttachmentListFunc != nil {
		return m.AttachmentListFunc(ctx, messageID)
	}
	return nil, nil
}

// Message edit history methods
func (m *MockDatabase) MessageEditHistoryCreate(ctx context.Context, history *models.MessageEditHistory) error {
	if m.MessageEditHistoryCreateFunc != nil {
		return m.MessageEditHistoryCreateFunc(ctx, history)
	}
	return nil
}

func (m *MockDatabase) MessageEditHistoryList(ctx context.Context, messageID string) ([]*models.MessageEditHistory, error) {
	if m.MessageEditHistoryListFunc != nil {
		return m.MessageEditHistoryListFunc(ctx, messageID)
	}
	return nil, nil
}

func (m *MockDatabase) MessageEditHistoryGet(ctx context.Context, id string) (*models.MessageEditHistory, error) {
	if m.MessageEditHistoryGetFunc != nil {
		return m.MessageEditHistoryGetFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockDatabase) MessageEditHistoryCount(ctx context.Context, messageID string) (int, error) {
	if m.MessageEditHistoryCountFunc != nil {
		return m.MessageEditHistoryCountFunc(ctx, messageID)
	}
	return 0, nil
}

// MockCoreService implements services.CoreService for testing
type MockCoreService struct {
	GetUserInfoFunc          func(ctx context.Context, userID uuid.UUID, jwt string) (*models.UserInfo, error)
	ValidateEntityAccessFunc func(ctx context.Context, userID, entityID uuid.UUID, entityType, jwt string) (bool, error)
	GetEntityDetailsFunc     func(ctx context.Context, entityID uuid.UUID, entityType, jwt string) (map[string]interface{}, error)
}

func (m *MockCoreService) GetUserInfo(ctx context.Context, userID uuid.UUID, jwt string) (*models.UserInfo, error) {
	if m.GetUserInfoFunc != nil {
		return m.GetUserInfoFunc(ctx, userID, jwt)
	}
	return nil, nil
}

func (m *MockCoreService) ValidateEntityAccess(ctx context.Context, userID, entityID uuid.UUID, entityType, jwt string) (bool, error) {
	if m.ValidateEntityAccessFunc != nil {
		return m.ValidateEntityAccessFunc(ctx, userID, entityID, entityType, jwt)
	}
	return true, nil
}

func (m *MockCoreService) GetEntityDetails(ctx context.Context, entityID uuid.UUID, entityType, jwt string) (map[string]interface{}, error) {
	if m.GetEntityDetailsFunc != nil {
		return m.GetEntityDetailsFunc(ctx, entityID, entityType, jwt)
	}
	return nil, nil
}

// TestHelpers provides utility functions for handler tests
type TestHelpers struct {
	t   *testing.T
	db  *MockDatabase
	svc *MockCoreService
	h   *Handler
}

// NewTestHelpers creates a new test helper instance
func NewTestHelpers(t *testing.T) *TestHelpers {
	db := &MockDatabase{}
	svc := &MockCoreService{}
	h := NewHandler(db, svc)

	return &TestHelpers{
		t:   t,
		db:  db,
		svc: svc,
		h:   h,
	}
}

// CreateTestContext creates a test Gin context with request
func (th *TestHelpers) CreateTestContext(method, path string, body interface{}) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	var bodyBytes []byte
	if body != nil {
		bodyBytes, _ = json.Marshal(body)
	}

	c.Request = httptest.NewRequest(method, path, bytes.NewBuffer(bodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	return c, w
}

// SetClaims sets JWT claims in the context
func (th *TestHelpers) SetClaims(c *gin.Context, userID uuid.UUID, username, role string) {
	claims := &models.JWTClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
	}
	c.Set("claims", claims)
	c.Set("user_id", userID.String())
	c.Set("username", username)
	c.Set("role", role)
}

// AssertJSONResponse asserts the HTTP response
func (th *TestHelpers) AssertJSONResponse(w *httptest.ResponseRecorder, expectedStatus int, expectedErrorCode int) models.APIResponse {
	assert.Equal(th.t, expectedStatus, w.Code, "HTTP status code mismatch")

	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(th.t, err, "Failed to unmarshal response")

	if expectedErrorCode != 0 {
		assert.Equal(th.t, expectedErrorCode, response.ErrorCode, "Error code mismatch")
	}

	return response
}

// CreateMockChatRoom creates a mock chat room for testing
func CreateMockChatRoom(id, name string, createdBy uuid.UUID) *models.ChatRoom {
	roomID := uuid.MustParse(id)
	return &models.ChatRoom{
		ID:          roomID,
		Name:        name,
		Description: "Test room",
		Type:        models.ChatRoomTypeGroup,
		CreatedBy:   createdBy,
		IsPrivate:   false,
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
		Deleted:     false,
	}
}

// CreateMockMessage creates a mock message for testing
func CreateMockMessage(id, chatRoomID string, senderID uuid.UUID, content string) *models.Message {
	msgID := uuid.MustParse(id)
	roomID := uuid.MustParse(chatRoomID)
	return &models.Message{
		ID:            msgID,
		ChatRoomID:    roomID,
		SenderID:      senderID,
		Type:          models.MessageTypeText,
		Content:       content,
		ContentFormat: models.ContentFormatPlain,
		IsEdited:      false,
		IsPinned:      false,
		CreatedAt:     time.Now().Unix(),
		UpdatedAt:     time.Now().Unix(),
		Deleted:       false,
	}
}

// CreateMockParticipant creates a mock participant for testing
func CreateMockParticipant(chatRoomID string, userID uuid.UUID, role models.ParticipantRole) *models.ChatParticipant {
	roomID := uuid.MustParse(chatRoomID)
	return &models.ChatParticipant{
		ChatRoomID: roomID,
		UserID:     userID,
		Role:       role,
		JoinedAt:   time.Now().Unix(),
		IsMuted:    false,
	}
}

package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"helixtrack.ru/chat/configs"
	"helixtrack.ru/chat/internal/services"
)

// MockDatabase implements database.Database for testing
type MockDatabase struct{}

func (m *MockDatabase) Close() error                                     { return nil }
func (m *MockDatabase) Ping() error                                      { return nil }
func (m *MockDatabase) BeginTx(ctx context.Context) (interface{}, error) { return nil, nil }

// Implement other required methods with empty implementations for testing
func (m *MockDatabase) ChatRoomCreate(ctx context.Context, room interface{}) error { return nil }
func (m *MockDatabase) ChatRoomRead(ctx context.Context, id string) (interface{}, error) {
	return nil, nil
}
func (m *MockDatabase) ChatRoomUpdate(ctx context.Context, room interface{}) error { return nil }
func (m *MockDatabase) ChatRoomDelete(ctx context.Context, id string) error        { return nil }
func (m *MockDatabase) ChatRoomList(ctx context.Context, limit, offset int) (interface{}, int, error) {
	return nil, 0, nil
}
func (m *MockDatabase) ChatRoomGetByEntity(ctx context.Context, entityType string, entityID string) (interface{}, error) {
	return nil, nil
}
func (m *MockDatabase) MessageCreate(ctx context.Context, message interface{}) error { return nil }
func (m *MockDatabase) MessageRead(ctx context.Context, id string) (interface{}, error) {
	return nil, nil
}
func (m *MockDatabase) MessageUpdate(ctx context.Context, message interface{}) error { return nil }
func (m *MockDatabase) MessageDelete(ctx context.Context, id string) error           { return nil }
func (m *MockDatabase) MessageList(ctx context.Context, req interface{}) (interface{}, int, error) {
	return nil, 0, nil
}
func (m *MockDatabase) MessageSearch(ctx context.Context, chatRoomID, query string, limit, offset int) (interface{}, int, error) {
	return nil, 0, nil
}
func (m *MockDatabase) ParticipantAdd(ctx context.Context, participant interface{}) error { return nil }
func (m *MockDatabase) ParticipantRemove(ctx context.Context, chatRoomID, userID string) error {
	return nil
}
func (m *MockDatabase) ParticipantList(ctx context.Context, chatRoomID string) (interface{}, error) {
	return nil, nil
}
func (m *MockDatabase) ParticipantUpdate(ctx context.Context, participant interface{}) error {
	return nil
}
func (m *MockDatabase) ParticipantGet(ctx context.Context, chatRoomID, userID string) (interface{}, error) {
	return nil, nil
}
func (m *MockDatabase) PresenceUpsert(ctx context.Context, presence interface{}) error { return nil }
func (m *MockDatabase) PresenceGet(ctx context.Context, userID string) (interface{}, error) {
	return nil, nil
}
func (m *MockDatabase) PresenceGetMultiple(ctx context.Context, userIDs []string) (interface{}, error) {
	return nil, nil
}
func (m *MockDatabase) TypingUpsert(ctx context.Context, indicator interface{}) error     { return nil }
func (m *MockDatabase) TypingDelete(ctx context.Context, chatRoomID, userID string) error { return nil }
func (m *MockDatabase) TypingGetActive(ctx context.Context, chatRoomID string) (interface{}, error) {
	return nil, nil
}
func (m *MockDatabase) ReadReceiptCreate(ctx context.Context, receipt interface{}) error { return nil }
func (m *MockDatabase) ReadReceiptGet(ctx context.Context, messageID string) (interface{}, error) {
	return nil, nil
}
func (m *MockDatabase) ReadReceiptGetByUser(ctx context.Context, messageID, userID string) (interface{}, error) {
	return nil, nil
}
func (m *MockDatabase) ReactionCreate(ctx context.Context, reaction interface{}) error { return nil }
func (m *MockDatabase) ReactionDelete(ctx context.Context, messageID, userID, emoji string) error {
	return nil
}
func (m *MockDatabase) ReactionList(ctx context.Context, messageID string) (interface{}, error) {
	return nil, nil
}
func (m *MockDatabase) AttachmentCreate(ctx context.Context, attachment interface{}) error {
	return nil
}
func (m *MockDatabase) AttachmentDelete(ctx context.Context, id string) error { return nil }
func (m *MockDatabase) AttachmentList(ctx context.Context, messageID string) (interface{}, error) {
	return nil, nil
}

func TestNewServer(t *testing.T) {
	config := configs.GetDefaultConfig()
	db := &MockDatabase{}
	coreService := &services.MockCoreService{}

	server := NewServer(config, db, coreService)

	assert.NotNil(t, server)
	assert.NotNil(t, server.router)
	assert.Equal(t, config, server.config)
	assert.Equal(t, db, server.db)
	assert.Equal(t, coreService, server.coreService)
}

func TestHealthHandler(t *testing.T) {
	config := configs.GetDefaultConfig()
	db := &MockDatabase{}
	coreService := &services.MockCoreService{}

	server := NewServer(config, db, coreService)
	server.setupRoutes()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "healthy")
}

func TestVersionHandler(t *testing.T) {
	config := configs.GetDefaultConfig()
	db := &MockDatabase{}
	coreService := &services.MockCoreService{}

	server := NewServer(config, db, coreService)
	server.setupRoutes()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/version", nil)
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "helixtrack-chat")
}

func TestGetRoutes(t *testing.T) {
	config := configs.GetDefaultConfig()
	db := &MockDatabase{}
	coreService := &services.MockCoreService{}

	server := NewServer(config, db, coreService)
	routes := server.GetRoutes()

	assert.NotEmpty(t, routes)
	assert.Equal(t, 4, len(routes))

	// Check specific routes
	healthRoute := routes[0]
	assert.Equal(t, "GET", healthRoute.Method)
	assert.Equal(t, "/health", healthRoute.Path)
	assert.False(t, healthRoute.AuthRequired)

	apiRoute := routes[2]
	assert.Equal(t, "POST", apiRoute.Method)
	assert.Equal(t, "/api/do", apiRoute.Path)
	assert.True(t, apiRoute.AuthRequired)
}

func TestSetupMiddleware(t *testing.T) {
	config := configs.GetDefaultConfig()
	db := &MockDatabase{}
	coreService := &services.MockCoreService{}

	server := NewServer(config, db, coreService)
	server.setupMiddleware()

	// Middleware should be registered
	assert.NotNil(t, server.router)
}

func TestShutdown(t *testing.T) {
	config := configs.GetDefaultConfig()
	db := &MockDatabase{}
	coreService := &services.MockCoreService{}

	server := NewServer(config, db, coreService)

	ctx := context.Background()
	err := server.Shutdown(ctx)

	// Should not error even if server not started
	assert.NoError(t, err)
}

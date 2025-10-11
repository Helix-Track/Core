package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"helixtrack.ru/core/internal/config"
	"helixtrack.ru/core/internal/database"
	"helixtrack.ru/core/internal/logger"
	"helixtrack.ru/core/internal/models"
	"helixtrack.ru/core/internal/services"
)

func init() {
	gin.SetMode(gin.TestMode)
	// Initialize logger for tests
	logger.Initialize(config.LogConfig{
		LogPath:      "/tmp",
		LogSizeLimit: 1000000,
		Level:        "error",
	})
}

// MockEventPublisher is a mock implementation of websocket.EventPublisher for testing
type MockEventPublisher struct {
	mu                   sync.Mutex
	PublishedEvents      []*models.Event
	PublishedEntityCalls []EntityEventCall
	enabled              bool
}

// EntityEventCall represents a call to PublishEntityEvent
type EntityEventCall struct {
	Action   string
	Object   string
	EntityID string
	Username string
	Data     map[string]interface{}
	Context  models.EventContext
}

// NewMockEventPublisher creates a new mock event publisher
func NewMockEventPublisher(enabled bool) *MockEventPublisher {
	return &MockEventPublisher{
		PublishedEvents:      make([]*models.Event, 0),
		PublishedEntityCalls: make([]EntityEventCall, 0),
		enabled:              enabled,
	}
}

// PublishEvent records the published event
func (m *MockEventPublisher) PublishEvent(event *models.Event) {
	if !m.enabled {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.PublishedEvents = append(m.PublishedEvents, event)
}

// PublishEntityEvent records the entity event call and creates an event
func (m *MockEventPublisher) PublishEntityEvent(action, object, entityID, username string, data map[string]interface{}, context models.EventContext) {
	if !m.enabled {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	// Record the call
	m.PublishedEntityCalls = append(m.PublishedEntityCalls, EntityEventCall{
		Action:   action,
		Object:   object,
		EntityID: entityID,
		Username: username,
		Data:     data,
		Context:  context,
	})

	// Create and store the event
	eventType := models.GetEventTypeFromAction(action, object)
	event := models.NewEvent(eventType, action, object, entityID, username, data)
	event.Context = context
	m.PublishedEvents = append(m.PublishedEvents, event)
}

// IsEnabled returns whether the mock publisher is enabled
func (m *MockEventPublisher) IsEnabled() bool {
	return m.enabled
}

// Reset clears all recorded events and calls
func (m *MockEventPublisher) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.PublishedEvents = make([]*models.Event, 0)
	m.PublishedEntityCalls = make([]EntityEventCall, 0)
}

// GetLastEvent returns the last published event (or nil if none)
func (m *MockEventPublisher) GetLastEvent() *models.Event {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.PublishedEvents) == 0 {
		return nil
	}
	return m.PublishedEvents[len(m.PublishedEvents)-1]
}

// GetLastEntityCall returns the last entity event call (or nil if none)
func (m *MockEventPublisher) GetLastEntityCall() *EntityEventCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.PublishedEntityCalls) == 0 {
		return nil
	}
	return &m.PublishedEntityCalls[len(m.PublishedEntityCalls)-1]
}

// GetEventCount returns the number of published events
func (m *MockEventPublisher) GetEventCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.PublishedEvents)
}

func setupTestHandler(t *testing.T) *Handler {
	db, err := database.NewDatabase(config.DatabaseConfig{
		Type:       "sqlite",
		SQLitePath: ":memory:",
	})
	require.NoError(t, err)

	mockAuth := &services.MockAuthService{
		IsEnabledFunc: func() bool { return true },
		AuthenticateFunc: func(ctx context.Context, username, password string) (*models.JWTClaims, error) {
			if username == "testuser" && password == "testpass" {
				return &models.JWTClaims{
					Username: "testuser",
					Role:     "admin",
					Name:     "Test User",
				}, nil
			}
			return nil, assert.AnError
		},
	}

	mockPerm := &services.MockPermissionService{
		IsEnabledFunc: func() bool { return true },
		CheckPermissionFunc: func(ctx context.Context, username, permissionContext string, requiredLevel models.PermissionLevel) (bool, error) {
			return username == "testuser", nil
		},
	}

	return NewHandler(db, mockAuth, mockPerm, "1.0.0-test")
}

// setupTestHandlerWithPublisher creates a test handler with a mock event publisher
func setupTestHandlerWithPublisher(t *testing.T) (*Handler, *MockEventPublisher) {
	handler := setupTestHandler(t)
	mockPublisher := NewMockEventPublisher(true)
	handler.SetEventPublisher(mockPublisher)
	return handler, mockPublisher
}

func TestHandler_DoAction_Version(t *testing.T) {
	handler := setupTestHandler(t)
	router := gin.New()
	router.POST("/do", handler.DoAction)

	reqBody := models.Request{
		Action: models.ActionVersion,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, "1.0.0-test", resp.Data["version"])
}

func TestHandler_DoAction_JWTCapable(t *testing.T) {
	handler := setupTestHandler(t)
	router := gin.New()
	router.POST("/do", handler.DoAction)

	reqBody := models.Request{
		Action: models.ActionJWTCapable,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["jwtCapable"].(bool))
}

func TestHandler_DoAction_DBCapable(t *testing.T) {
	handler := setupTestHandler(t)
	router := gin.New()
	router.POST("/do", handler.DoAction)

	reqBody := models.Request{
		Action: models.ActionDBCapable,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["dbCapable"].(bool))
	assert.Equal(t, "sqlite", resp.Data["type"])
}

func TestHandler_DoAction_Health(t *testing.T) {
	handler := setupTestHandler(t)
	router := gin.New()
	router.POST("/do", handler.DoAction)

	reqBody := models.Request{
		Action: models.ActionHealth,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, "healthy", resp.Data["status"])
}

func TestHandler_DoAction_Authenticate(t *testing.T) {
	handler := setupTestHandler(t)
	router := gin.New()
	router.POST("/do", handler.DoAction)

	t.Run("Successful authentication", func(t *testing.T) {
		reqBody := models.Request{
			Action: models.ActionAuthenticate,
			Data: map[string]interface{}{
				"username": "testuser",
				"password": "testpass",
			},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp models.Response
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
		assert.Equal(t, "testuser", resp.Data["username"])
	})

	t.Run("Missing username", func(t *testing.T) {
		reqBody := models.Request{
			Action: models.ActionAuthenticate,
			Data: map[string]interface{}{
				"password": "testpass",
			},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp models.Response
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
	})

	t.Run("Missing password", func(t *testing.T) {
		reqBody := models.Request{
			Action: models.ActionAuthenticate,
			Data: map[string]interface{}{
				"username": "testuser",
			},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Invalid credentials", func(t *testing.T) {
		reqBody := models.Request{
			Action: models.ActionAuthenticate,
			Data: map[string]interface{}{
				"username": "wronguser",
				"password": "wrongpass",
			},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestHandler_DoAction_Create(t *testing.T) {
	handler := setupTestHandler(t)
	router := gin.New()
	router.POST("/do", handler.DoAction)

	t.Run("Missing object", func(t *testing.T) {
		reqBody := models.Request{
			Action: models.ActionCreate,
			Data:   map[string]interface{}{},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp models.Response
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, models.ErrorCodeMissingObject, resp.ErrorCode)
	})

	t.Run("Unauthorized (no username in context)", func(t *testing.T) {
		reqBody := models.Request{
			Action: models.ActionCreate,
			Object: "project",
			Data:   map[string]interface{}{},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Successful with username", func(t *testing.T) {
		reqBody := models.Request{
			Action: models.ActionCreate,
			Object: "project",
			Data:   map[string]interface{}{},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Create test context with username
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("username", "testuser")

		handler.handleCreate(c, &reqBody)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestHandler_DoAction_InvalidAction(t *testing.T) {
	handler := setupTestHandler(t)
	router := gin.New()
	router.POST("/do", handler.DoAction)

	reqBody := models.Request{
		Action: "invalidAction",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeInvalidAction, resp.ErrorCode)
}

func TestHandler_DoAction_InvalidJSON(t *testing.T) {
	handler := setupTestHandler(t)
	router := gin.New()
	router.POST("/do", handler.DoAction)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer([]byte("{invalid json}")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_Modify(t *testing.T) {
	handler := setupTestHandler(t)

	t.Run("Missing object", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := &models.Request{Action: models.ActionModify}

		handler.handleModify(c, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := &models.Request{Action: models.ActionModify, Object: "project"}

		handler.handleModify(c, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("With username", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("username", "testuser")
		req := &models.Request{Action: models.ActionModify, Object: "project"}

		handler.handleModify(c, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestHandler_Remove(t *testing.T) {
	handler := setupTestHandler(t)

	t.Run("Missing object", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := &models.Request{Action: models.ActionRemove}

		handler.handleRemove(c, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := &models.Request{Action: models.ActionRemove, Object: "project"}

		handler.handleRemove(c, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("With username", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("username", "testuser")
		req := &models.Request{Action: models.ActionRemove, Object: "project"}

		handler.handleRemove(c, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestHandler_Read(t *testing.T) {
	handler := setupTestHandler(t)

	t.Run("Unauthorized", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := &models.Request{Action: models.ActionRead}

		handler.handleRead(c, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("With username", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("username", "testuser")
		req := &models.Request{Action: models.ActionRead}

		handler.handleRead(c, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestHandler_List(t *testing.T) {
	handler := setupTestHandler(t)

	t.Run("Unauthorized", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := &models.Request{Action: models.ActionList}

		handler.handleList(c, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("With username", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("username", "testuser")
		req := &models.Request{Action: models.ActionList}

		handler.handleList(c, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestNewHandler(t *testing.T) {
	db, _ := database.NewDatabase(config.DatabaseConfig{
		Type:       "sqlite",
		SQLitePath: ":memory:",
	})
	authService := &services.MockAuthService{}
	permService := &services.MockPermissionService{}

	handler := NewHandler(db, authService, permService, "1.0.0")

	assert.NotNil(t, handler)
	assert.Equal(t, db, handler.db)
	assert.Equal(t, authService, handler.authService)
	assert.Equal(t, permService, handler.permService)
	assert.Equal(t, "1.0.0", handler.version)
}

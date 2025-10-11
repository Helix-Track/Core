package models

import (
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockClientManager implements ClientManager for testing
type MockClientManager struct {
	mock.Mock
}

func (m *MockClientManager) UnregisterClient(client *Client) {
	m.Called(client)
}

func TestNewClient(t *testing.T) {
	mockManager := &MockClientManager{}
	mockConn := &websocket.Conn{}
	claims := &JWTClaims{
		Username: "john.doe",
		Role:     "admin",
	}

	client := NewClient("client-123", mockConn, "john.doe", claims, mockManager)

	assert.NotNil(t, client)
	assert.Equal(t, "client-123", client.ID)
	assert.Equal(t, mockConn, client.Conn)
	assert.Equal(t, "john.doe", client.Username)
	assert.Equal(t, claims, client.Claims)
	assert.Equal(t, mockManager, client.Manager)
	assert.NotNil(t, client.Subscription)
	assert.NotNil(t, client.Send)
	assert.Equal(t, 256, cap(client.Send))
	assert.WithinDuration(t, time.Now(), client.Connected, 1*time.Second)
	assert.WithinDuration(t, time.Now(), client.LastPing, 1*time.Second)
	assert.WithinDuration(t, time.Now(), client.LastActivity, 1*time.Second)
	assert.NotNil(t, client.Metadata)
}

func TestClient_UpdateSubscription(t *testing.T) {
	mockManager := &MockClientManager{}
	client := NewClient("client-123", nil, "john.doe", nil, mockManager)

	newSubscription := &Subscription{
		EventTypes:  []EventType{EventTicketCreated},
		EntityTypes: []string{"ticket"},
		EntityIDs:   []string{"ticket-123"},
	}

	client.UpdateSubscription(newSubscription)

	subscription := client.GetSubscription()
	assert.Equal(t, newSubscription, subscription)
	assert.Len(t, subscription.EventTypes, 1)
	assert.Len(t, subscription.EntityTypes, 1)
	assert.Len(t, subscription.EntityIDs, 1)
}

func TestClient_GetSubscription(t *testing.T) {
	mockManager := &MockClientManager{}
	client := NewClient("client-123", nil, "john.doe", nil, mockManager)

	subscription := client.GetSubscription()

	assert.NotNil(t, subscription)
	assert.Empty(t, subscription.EventTypes)
	assert.Empty(t, subscription.EntityTypes)
	assert.Empty(t, subscription.EntityIDs)
}

func TestClient_UpdateActivity(t *testing.T) {
	mockManager := &MockClientManager{}
	client := NewClient("client-123", nil, "john.doe", nil, mockManager)

	initialActivity := client.GetLastActivity()
	time.Sleep(10 * time.Millisecond)

	client.UpdateActivity()
	newActivity := client.GetLastActivity()

	assert.True(t, newActivity.After(initialActivity), "LastActivity should be updated")
}

func TestClient_UpdatePing(t *testing.T) {
	mockManager := &MockClientManager{}
	client := NewClient("client-123", nil, "john.doe", nil, mockManager)

	initialPing := client.GetLastPing()
	time.Sleep(10 * time.Millisecond)

	client.UpdatePing()
	newPing := client.GetLastPing()

	assert.True(t, newPing.After(initialPing), "LastPing should be updated")
}

func TestClient_GetLastActivity(t *testing.T) {
	mockManager := &MockClientManager{}
	client := NewClient("client-123", nil, "john.doe", nil, mockManager)

	lastActivity := client.GetLastActivity()

	assert.WithinDuration(t, time.Now(), lastActivity, 1*time.Second)
}

func TestClient_GetLastPing(t *testing.T) {
	mockManager := &MockClientManager{}
	client := NewClient("client-123", nil, "john.doe", nil, mockManager)

	lastPing := client.GetLastPing()

	assert.WithinDuration(t, time.Now(), lastPing, 1*time.Second)
}

func TestClient_HasPermission(t *testing.T) {
	tests := []struct {
		name       string
		claims     *JWTClaims
		permission int
		want       bool
	}{
		{
			name: "Has READ permission",
			claims: &JWTClaims{
				Username:    "john.doe",
				Permissions: "READ",
			},
			permission: int(PermissionRead),
			want:       true,
		},
		{
			name: "Has UPDATE permission",
			claims: &JWTClaims{
				Username:    "john.doe",
				Permissions: "UPDATE",
			},
			permission: int(PermissionUpdate),
			want:       true,
		},
		{
			name: "Has DELETE permission",
			claims: &JWTClaims{
				Username:    "john.doe",
				Permissions: "DELETE",
			},
			permission: int(PermissionDelete),
			want:       true,
		},
		{
			name: "Does not have permission",
			claims: &JWTClaims{
				Username:    "john.doe",
				Permissions: "READ",
			},
			permission: int(PermissionDelete),
			want:       false,
		},
		{
			name:       "No claims",
			claims:     nil,
			permission: int(PermissionRead),
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockManager := &MockClientManager{}
			client := NewClient("client-123", nil, "john.doe", tt.claims, mockManager)

			got := client.HasPermission(tt.permission)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewWebSocketMessage(t *testing.T) {
	data := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
	}

	msg := NewWebSocketMessage(WSMessageTypeEvent, "create", data)

	assert.NotNil(t, msg)
	assert.Equal(t, WSMessageTypeEvent, msg.Type)
	assert.Equal(t, "create", msg.Action)
	assert.Equal(t, data, msg.Data)
	assert.Empty(t, msg.EventID)
	assert.Empty(t, msg.Error)
}

func TestNewErrorMessage(t *testing.T) {
	errorMsg := "Something went wrong"

	msg := NewErrorMessage(errorMsg)

	assert.NotNil(t, msg)
	assert.Equal(t, WSMessageTypeError, msg.Type)
	assert.Equal(t, errorMsg, msg.Error)
	assert.Empty(t, msg.Action)
	assert.Nil(t, msg.Data)
	assert.Empty(t, msg.EventID)
}

func TestNewEventMessage(t *testing.T) {
	event := &Event{
		ID:       "event-123",
		Type:     EventTicketCreated,
		Action:   ActionCreate,
		Object:   "ticket",
		EntityID: "ticket-456",
		Username: "john.doe",
		Data: map[string]interface{}{
			"title": "Test Ticket",
		},
	}

	msg := NewEventMessage(event)

	assert.NotNil(t, msg)
	assert.Equal(t, WSMessageTypeEvent, msg.Type)
	assert.Equal(t, event.Action, msg.Action)
	assert.Equal(t, event.ID, msg.EventID)
	assert.NotNil(t, msg.Data)
	assert.Equal(t, event, msg.Data["event"])
}

func TestWebSocketMessageTypes(t *testing.T) {
	// Test that all message type constants are defined
	messageTypes := []string{
		WSMessageTypeSubscribe,
		WSMessageTypeUnsubscribe,
		WSMessageTypeEvent,
		WSMessageTypePing,
		WSMessageTypePong,
		WSMessageTypeError,
		WSMessageTypeAck,
		WSMessageTypeAuth,
	}

	// Verify all message types are non-empty
	for _, msgType := range messageTypes {
		assert.NotEmpty(t, msgType, "Message type should not be empty")
	}

	// Verify uniqueness
	uniqueTypes := make(map[string]bool)
	for _, msgType := range messageTypes {
		assert.False(t, uniqueTypes[msgType], "Message type should be unique: %s", msgType)
		uniqueTypes[msgType] = true
	}
}

func TestDefaultWebSocketConfig(t *testing.T) {
	config := DefaultWebSocketConfig()

	assert.True(t, config.Enabled)
	assert.Equal(t, "/ws", config.Path)
	assert.Equal(t, 1024, config.ReadBufferSize)
	assert.Equal(t, 1024, config.WriteBufferSize)
	assert.Equal(t, int64(512*1024), config.MaxMessageSize)
	assert.Equal(t, 10*time.Second, config.WriteWait)
	assert.Equal(t, 60*time.Second, config.PongWait)
	assert.Equal(t, 54*time.Second, config.PingPeriod)
	assert.Equal(t, 1000, config.MaxClients)
	assert.True(t, config.RequireAuth)
	assert.Equal(t, []string{"*"}, config.AllowOrigins)
	assert.True(t, config.EnableCompression)
	assert.Equal(t, 10*time.Second, config.HandshakeTimeout)
}

func TestWebSocketConfig_CustomValues(t *testing.T) {
	config := WebSocketConfig{
		Enabled:           false,
		Path:              "/custom-ws",
		ReadBufferSize:    2048,
		WriteBufferSize:   2048,
		MaxMessageSize:    1024 * 1024,
		WriteWait:         20 * time.Second,
		PongWait:          120 * time.Second,
		PingPeriod:        100 * time.Second,
		MaxClients:        5000,
		RequireAuth:       false,
		AllowOrigins:      []string{"https://example.com"},
		EnableCompression: false,
		HandshakeTimeout:  30 * time.Second,
	}

	assert.False(t, config.Enabled)
	assert.Equal(t, "/custom-ws", config.Path)
	assert.Equal(t, 2048, config.ReadBufferSize)
	assert.Equal(t, 2048, config.WriteBufferSize)
	assert.Equal(t, int64(1024*1024), config.MaxMessageSize)
	assert.Equal(t, 20*time.Second, config.WriteWait)
	assert.Equal(t, 120*time.Second, config.PongWait)
	assert.Equal(t, 100*time.Second, config.PingPeriod)
	assert.Equal(t, 5000, config.MaxClients)
	assert.False(t, config.RequireAuth)
	assert.Equal(t, []string{"https://example.com"}, config.AllowOrigins)
	assert.False(t, config.EnableCompression)
	assert.Equal(t, 30*time.Second, config.HandshakeTimeout)
}

func TestWebSocketMessage_AllFields(t *testing.T) {
	msg := &WebSocketMessage{
		Type:    WSMessageTypeEvent,
		Action:  "create",
		Data:    map[string]interface{}{"key": "value"},
		EventID: "event-123",
		Error:   "error message",
	}

	assert.Equal(t, WSMessageTypeEvent, msg.Type)
	assert.Equal(t, "create", msg.Action)
	assert.Equal(t, map[string]interface{}{"key": "value"}, msg.Data)
	assert.Equal(t, "event-123", msg.EventID)
	assert.Equal(t, "error message", msg.Error)
}

func TestClient_ConcurrentOperations(t *testing.T) {
	mockManager := &MockClientManager{}
	client := NewClient("client-123", nil, "john.doe", nil, mockManager)

	// Test concurrent reads and writes
	done := make(chan bool)
	iterations := 100

	// Concurrent subscription updates
	go func() {
		for i := 0; i < iterations; i++ {
			client.UpdateSubscription(&Subscription{
				EventTypes: []EventType{EventTicketCreated},
			})
		}
		done <- true
	}()

	// Concurrent subscription reads
	go func() {
		for i := 0; i < iterations; i++ {
			_ = client.GetSubscription()
		}
		done <- true
	}()

	// Concurrent activity updates
	go func() {
		for i := 0; i < iterations; i++ {
			client.UpdateActivity()
		}
		done <- true
	}()

	// Concurrent ping updates
	go func() {
		for i := 0; i < iterations; i++ {
			client.UpdatePing()
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 4; i++ {
		<-done
	}

	// Verify client is still in good state
	assert.NotNil(t, client.GetSubscription())
	assert.NotZero(t, client.GetLastActivity())
	assert.NotZero(t, client.GetLastPing())
}

func TestClient_Metadata(t *testing.T) {
	mockManager := &MockClientManager{}
	client := NewClient("client-123", nil, "john.doe", nil, mockManager)

	// Add metadata
	client.Metadata["userAgent"] = "Mozilla/5.0"
	client.Metadata["ipAddress"] = "192.168.1.1"
	client.Metadata["sessionId"] = "session-456"

	assert.Equal(t, "Mozilla/5.0", client.Metadata["userAgent"])
	assert.Equal(t, "192.168.1.1", client.Metadata["ipAddress"])
	assert.Equal(t, "session-456", client.Metadata["sessionId"])
	assert.Len(t, client.Metadata, 3)
}

func TestWebSocketConfig_PingPeriodLessThanPongWait(t *testing.T) {
	config := DefaultWebSocketConfig()

	// Verify that PingPeriod is less than PongWait
	// This is important for the heartbeat mechanism
	assert.Less(t, config.PingPeriod, config.PongWait,
		"PingPeriod should be less than PongWait to allow time for pong response")
}

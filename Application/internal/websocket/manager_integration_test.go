package websocket

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"helixtrack.ru/core/internal/models"
	"helixtrack.ru/core/internal/services"
)

// mockPermissionService is a mock implementation of services.PermissionService for testing
type mockPermissionService struct {
	enabled bool
}

func (m *mockPermissionService) CheckPermission(ctx context.Context, username, permissionContext string, requiredLevel models.PermissionLevel) (bool, error) {
	// Allow all permissions in tests
	return true, nil
}

func (m *mockPermissionService) GetUserPermissions(ctx context.Context, username string) ([]models.Permission, error) {
	return []models.Permission{}, nil
}

func (m *mockPermissionService) IsEnabled() bool {
	return m.enabled
}

// createTestConfig creates a default WebSocket configuration for testing
func createTestConfig() models.WebSocketConfig {
	return models.WebSocketConfig{
		Path:              "/ws",
		MaxClients:        100,
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		MaxMessageSize:    512000,
		WriteWait:         10 * time.Second,
		PongWait:          60 * time.Second,
		PingPeriod:        54 * time.Second,
		HandshakeTimeout:  10 * time.Second,
		EnableCompression: false,
		AllowOrigins:      []string{"*"},
	}
}

// createMockPermissionService creates a mock permission service for testing
func createMockPermissionService() services.PermissionService {
	return &mockPermissionService{enabled: false}
}

// createTestWebSocketHandler creates a WebSocket handler for testing
func createTestWebSocketHandler(manager *Manager, username string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get username from query or use default
		user := c.Query("user")
		if user == "" {
			user = username
		}

		// Upgrade connection
		upgrader := manager.GetUpgrader()
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to upgrade connection"})
			return
		}

		// Create client with mock JWT claims
		claims := &models.JWTClaims{
			Username: user,
			Role:     "user",
		}
		client := manager.CreateClient(conn, user, claims)

		// Register client
		if err := manager.RegisterClient(client); err != nil {
			conn.Close()
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Failed to register client"})
			return
		}
	}
}

// TestWebSocketConnection_Integration tests WebSocket connection establishment
func TestWebSocketConnection_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manager := NewManager(createTestConfig(), createMockPermissionService())
	require.NoError(t, manager.Start())
	defer manager.Stop()

	// Create test server
	router := gin.New()
	router.GET("/ws", createTestWebSocketHandler(manager, "testuser"))

	server := httptest.NewServer(router)
	defer server.Close()

	// Connect to WebSocket
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer ws.Close()

	// Wait for connection to be registered
	time.Sleep(100 * time.Millisecond)

	// Verify client is connected
	assert.Equal(t, 1, manager.GetStats().ActiveConnections)
}

// TestWebSocketSubscription_Integration tests event subscription
func TestWebSocketSubscription_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manager := NewManager(createTestConfig(), createMockPermissionService())
	require.NoError(t, manager.Start())
	defer manager.Stop()

	// Create test server
	router := gin.New()
	router.GET("/ws", createTestWebSocketHandler(manager, "testuser"))

	server := httptest.NewServer(router)
	defer server.Close()

	// Connect to WebSocket
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer ws.Close()

	// Send subscription message
	subscribeMsg := map[string]interface{}{
		"type": "subscribe",
		"data": map[string]interface{}{
			"eventTypes": []string{"ticket.created", "project.updated"},
		},
	}
	err = ws.WriteJSON(subscribeMsg)
	require.NoError(t, err)

	// Wait for subscription to be processed
	time.Sleep(100 * time.Millisecond)

	// Read response
	var response map[string]interface{}
	err = ws.ReadJSON(&response)
	require.NoError(t, err)

	// Verify subscription response
	assert.Equal(t, "subscription_confirmed", response["type"])
	eventTypes, ok := response["eventTypes"].([]interface{})
	require.True(t, ok)
	assert.Len(t, eventTypes, 2)
}

// TestWebSocketEventDelivery_Integration tests event delivery to subscribed client
func TestWebSocketEventDelivery_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manager := NewManager(createTestConfig(), createMockPermissionService())
	require.NoError(t, manager.Start())
	defer manager.Stop()
	publisher := NewPublisher(manager, true)

	// Create test server
	router := gin.New()
	router.GET("/ws", createTestWebSocketHandler(manager, "testuser"))

	server := httptest.NewServer(router)
	defer server.Close()

	// Connect to WebSocket
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer ws.Close()

	// Subscribe to ticket.created events
	subscribeMsg := map[string]interface{}{
		"type": "subscribe",
		"data": map[string]interface{}{
			"eventTypes": []string{"ticket.created"},
		},
	}
	err = ws.WriteJSON(subscribeMsg)
	require.NoError(t, err)

	// Read subscription confirmation
	var response map[string]interface{}
	err = ws.ReadJSON(&response)
	require.NoError(t, err)

	// Publish a ticket.created event
	publisher.PublishEntityEvent(
		models.ActionCreate,
		"ticket",
		"ticket-123",
		"testuser",
		map[string]interface{}{
			"id":    "ticket-123",
			"title": "Test Ticket",
		},
		NewProjectContext("project-1", []string{"READ"}),
	)

	// Wait for event delivery
	time.Sleep(100 * time.Millisecond)

	// Read event
	var event map[string]interface{}
	ws.SetReadDeadline(time.Now().Add(2 * time.Second))
	err = ws.ReadJSON(&event)
	require.NoError(t, err)

	// Verify event
	assert.Equal(t, "event", event["type"])
	eventData, ok := event["event"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "ticket.created", eventData["eventType"])
	assert.Equal(t, "ticket-123", eventData["entityId"])
	assert.Equal(t, "testuser", eventData["username"])
}

// TestWebSocketMultipleClients_Integration tests event delivery to multiple clients
func TestWebSocketMultipleClients_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manager := NewManager(createTestConfig(), createMockPermissionService())
	require.NoError(t, manager.Start())
	defer manager.Stop()
	publisher := NewPublisher(manager, true)

	// Create test server
	router := gin.New()
	router.GET("/ws", createTestWebSocketHandler(manager, "testuser"))

	server := httptest.NewServer(router)
	defer server.Close()

	// Connect 3 WebSocket clients
	clients := make([]*websocket.Conn, 3)
	usernames := []string{"user1", "user2", "user3"}

	for i := 0; i < 3; i++ {
		wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws?user=" + usernames[i]
		ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer ws.Close()
		clients[i] = ws

		// Subscribe to priority.created events
		subscribeMsg := map[string]interface{}{
			"type": "subscribe",
			"data": map[string]interface{}{
				"eventTypes": []string{"priority.created"},
			},
		}
		err = ws.WriteJSON(subscribeMsg)
		require.NoError(t, err)

		// Read subscription confirmation
		var response map[string]interface{}
		err = ws.ReadJSON(&response)
		require.NoError(t, err)
	}

	// Wait for all clients to be registered
	time.Sleep(200 * time.Millisecond)

	// Verify all clients connected
	assert.Equal(t, 3, manager.GetStats().ActiveConnections)

	// Publish a priority.created event
	publisher.PublishEntityEvent(
		models.ActionCreate,
		"priority",
		"priority-123",
		"admin",
		map[string]interface{}{
			"id":    "priority-123",
			"title": "High Priority",
			"level": 4,
		},
		NewProjectContext("", []string{"READ"}), // System-wide
	)

	// Wait for event delivery
	time.Sleep(200 * time.Millisecond)

	// Verify all clients received the event
	for i, client := range clients {
		var event map[string]interface{}
		client.SetReadDeadline(time.Now().Add(2 * time.Second))
		err := client.ReadJSON(&event)
		require.NoError(t, err, "Client %d should receive event", i)

		// Verify event
		assert.Equal(t, "event", event["type"])
		eventData, ok := event["event"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "priority.created", eventData["eventType"])
		assert.Equal(t, "priority-123", eventData["entityId"])
	}
}

// TestWebSocketEventFiltering_Integration tests that clients only receive subscribed events
func TestWebSocketEventFiltering_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manager := NewManager(createTestConfig(), createMockPermissionService())
	require.NoError(t, manager.Start())
	defer manager.Stop()
	publisher := NewPublisher(manager, true)

	// Create test server
	router := gin.New()
	router.GET("/ws", createTestWebSocketHandler(manager, "testuser"))

	server := httptest.NewServer(router)
	defer server.Close()

	// Connect to WebSocket
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer ws.Close()

	// Subscribe only to ticket.created events
	subscribeMsg := map[string]interface{}{
		"type": "subscribe",
		"data": map[string]interface{}{
			"eventTypes": []string{"ticket.created"},
		},
	}
	err = ws.WriteJSON(subscribeMsg)
	require.NoError(t, err)

	// Read subscription confirmation
	var response map[string]interface{}
	err = ws.ReadJSON(&response)
	require.NoError(t, err)

	// Publish a priority.created event (not subscribed)
	publisher.PublishEntityEvent(
		models.ActionCreate,
		"priority",
		"priority-123",
		"testuser",
		map[string]interface{}{
			"id":    "priority-123",
			"title": "High",
		},
		NewProjectContext("", []string{"READ"}),
	)

	// Publish a ticket.created event (subscribed)
	publisher.PublishEntityEvent(
		models.ActionCreate,
		"ticket",
		"ticket-123",
		"testuser",
		map[string]interface{}{
			"id":    "ticket-123",
			"title": "Test Ticket",
		},
		NewProjectContext("project-1", []string{"READ"}),
	)

	// Wait for events
	time.Sleep(200 * time.Millisecond)

	// Should only receive ticket.created event
	var event map[string]interface{}
	ws.SetReadDeadline(time.Now().Add(2 * time.Second))
	err = ws.ReadJSON(&event)
	require.NoError(t, err)

	// Verify we received ticket.created
	assert.Equal(t, "event", event["type"])
	eventData, ok := event["event"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "ticket.created", eventData["eventType"])

	// Verify no more events in queue
	ws.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	err = ws.ReadJSON(&event)
	assert.Error(t, err) // Should timeout since no more events
}

// TestWebSocketUnsubscribe_Integration tests event unsubscription
func TestWebSocketUnsubscribe_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manager := NewManager(createTestConfig(), createMockPermissionService())
	require.NoError(t, manager.Start())
	defer manager.Stop()
	publisher := NewPublisher(manager, true)

	// Create test server
	router := gin.New()
	router.GET("/ws", createTestWebSocketHandler(manager, "testuser"))

	server := httptest.NewServer(router)
	defer server.Close()

	// Connect to WebSocket
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer ws.Close()

	// Subscribe to ticket.created events
	subscribeMsg := map[string]interface{}{
		"type": "subscribe",
		"data": map[string]interface{}{
			"eventTypes": []string{"ticket.created"},
		},
	}
	err = ws.WriteJSON(subscribeMsg)
	require.NoError(t, err)

	// Read subscription confirmation
	var response map[string]interface{}
	err = ws.ReadJSON(&response)
	require.NoError(t, err)

	// Unsubscribe from ticket.created events
	unsubscribeMsg := map[string]interface{}{
		"type": "unsubscribe",
		"data": map[string]interface{}{
			"eventTypes": []string{"ticket.created"},
		},
	}
	err = ws.WriteJSON(unsubscribeMsg)
	require.NoError(t, err)

	// Read unsubscribe confirmation
	err = ws.ReadJSON(&response)
	require.NoError(t, err)
	assert.Equal(t, "unsubscription_confirmed", response["type"])

	// Publish a ticket.created event
	publisher.PublishEntityEvent(
		models.ActionCreate,
		"ticket",
		"ticket-123",
		"testuser",
		map[string]interface{}{
			"id":    "ticket-123",
			"title": "Test Ticket",
		},
		NewProjectContext("project-1", []string{"READ"}),
	)

	// Wait for potential event delivery
	time.Sleep(200 * time.Millisecond)

	// Should NOT receive any event
	ws.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	var event map[string]interface{}
	err = ws.ReadJSON(&event)
	assert.Error(t, err) // Should timeout since no events should be delivered
}

// TestWebSocketConcurrentEventDelivery_Integration tests concurrent event delivery
func TestWebSocketConcurrentEventDelivery_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manager := NewManager(createTestConfig(), createMockPermissionService())
	require.NoError(t, manager.Start())
	defer manager.Stop()
	publisher := NewPublisher(manager, true)

	// Create test server
	router := gin.New()
	router.GET("/ws", createTestWebSocketHandler(manager, "testuser"))

	server := httptest.NewServer(router)
	defer server.Close()

	// Connect to WebSocket
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer ws.Close()

	// Subscribe to all event types
	subscribeMsg := map[string]interface{}{
		"type": "subscribe",
		"data": map[string]interface{}{
			"eventTypes": []string{"ticket.created", "priority.created", "project.created"},
		},
	}
	err = ws.WriteJSON(subscribeMsg)
	require.NoError(t, err)

	// Read subscription confirmation
	var response map[string]interface{}
	err = ws.ReadJSON(&response)
	require.NoError(t, err)

	// Publish multiple events concurrently
	var wg sync.WaitGroup
	eventCount := 10

	for i := 0; i < eventCount; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			publisher.PublishEntityEvent(
				models.ActionCreate,
				"ticket",
				string(rune('A'+index)),
				"testuser",
				map[string]interface{}{
					"id":    string(rune('A' + index)),
					"title": "Ticket " + string(rune('A'+index)),
				},
				NewProjectContext("project-1", []string{"READ"}),
			)
		}(i)
	}

	wg.Wait()
	time.Sleep(500 * time.Millisecond)

	// Read all events
	receivedEvents := 0
	for i := 0; i < eventCount; i++ {
		var event map[string]interface{}
		ws.SetReadDeadline(time.Now().Add(2 * time.Second))
		err = ws.ReadJSON(&event)
		if err == nil {
			receivedEvents++
		}
	}

	// Verify all events were received
	assert.Equal(t, eventCount, receivedEvents)
}

// TestWebSocketDisconnect_Integration tests client disconnect handling
func TestWebSocketDisconnect_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manager := NewManager(createTestConfig(), createMockPermissionService())
	require.NoError(t, manager.Start())
	defer manager.Stop()

	// Create test server
	router := gin.New()
	router.GET("/ws", createTestWebSocketHandler(manager, "testuser"))

	server := httptest.NewServer(router)
	defer server.Close()

	// Connect to WebSocket
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)

	// Wait for connection to be registered
	time.Sleep(100 * time.Millisecond)

	// Verify client is connected
	assert.Equal(t, 1, manager.GetStats().ActiveConnections)

	// Disconnect client
	ws.Close()

	// Wait for disconnect to be processed
	time.Sleep(200 * time.Millisecond)

	// Verify client is disconnected
	assert.Equal(t, 0, manager.GetStats().ActiveConnections)
}

// TestWebSocketProjectContextFiltering_Integration tests project-based event filtering
func TestWebSocketProjectContextFiltering_Integration(t *testing.T) {
	// This test would require implementing permission-based filtering in the manager
	// For now, it's a placeholder for future implementation
	t.Skip("Project context filtering not yet implemented in manager")
}

// TestWebSocketPingPong_Integration tests WebSocket ping/pong keepalive
func TestWebSocketPingPong_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manager := NewManager(createTestConfig(), createMockPermissionService())
	require.NoError(t, manager.Start())
	defer manager.Stop()

	// Create test server
	router := gin.New()
	router.GET("/ws", createTestWebSocketHandler(manager, "testuser"))

	server := httptest.NewServer(router)
	defer server.Close()

	// Connect to WebSocket
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer ws.Close()

	// Send ping
	err = ws.WriteMessage(websocket.PingMessage, []byte{})
	require.NoError(t, err)

	// Wait for pong
	time.Sleep(200 * time.Millisecond)

	// Note: gorilla/websocket automatically responds to pings
	// This test verifies the connection is still alive
	assert.Equal(t, 1, manager.GetStats().ActiveConnections)
}

// TestWebSocketInvalidMessage_Integration tests handling of invalid messages
func TestWebSocketInvalidMessage_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manager := NewManager(createTestConfig(), createMockPermissionService())
	require.NoError(t, manager.Start())
	defer manager.Stop()

	// Create test server
	router := gin.New()
	router.GET("/ws", createTestWebSocketHandler(manager, "testuser"))

	server := httptest.NewServer(router)
	defer server.Close()

	// Connect to WebSocket
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer ws.Close()

	// Send invalid JSON
	err = ws.WriteMessage(websocket.TextMessage, []byte("{invalid json}"))
	require.NoError(t, err)

	// Wait for response
	time.Sleep(100 * time.Millisecond)

	// Connection should still be alive (manager should handle errors gracefully)
	assert.Equal(t, 1, manager.GetStats().ActiveConnections)
}

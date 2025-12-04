package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap/zaptest"
)

func TestNewManager(t *testing.T) {
	logger := zaptest.NewLogger(t)
	manager := NewManager(logger)

	if manager == nil {
		t.Fatal("Expected non-nil manager")
	}

	if manager.Clients == nil {
		t.Error("Expected non-nil clients map")
	}
}

func TestManager_Start(t *testing.T) {
	logger := zaptest.NewLogger(t)
	manager := NewManager(logger)
	
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// This should not block
	manager.Start(ctx)

	// Verify the context cancellation works
	select {
	case <-time.After(200 * time.Millisecond):
		t.Error("Manager should have stopped when context was cancelled")
	case <-ctx.Done():
		// Expected
	}
}

func TestManager_HandleConnection(t *testing.T) {
	logger := zaptest.NewLogger(t)
	manager := NewManager(logger)

	// Create test server with WebSocket handler
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		manager.HandleConnection(w, r)
	}))
	defer server.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect to WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	// Wait a moment for connection to be established
	time.Sleep(10 * time.Millisecond)

	// Check that client was registered
	manager.mu.RLock()
	if len(manager.clients) != 1 {
		t.Errorf("Expected 1 client, got %d", len(manager.clients))
	}
	manager.mu.RUnlock()

	// Send a message
	testMessage := map[string]interface{}{
		"type": "ping",
		"data": "test",
	}
	messageBytes, _ := json.Marshal(testMessage)
	err = conn.WriteMessage(websocket.TextMessage, messageBytes)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// Read response
	_, response, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(response, &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp["type"] != "pong" {
		t.Errorf("Expected pong response, got: %v", resp["type"])
	}
}

func TestManager_Broadcast(t *testing.T) {
	logger := zaptest.NewLogger(t)
	manager := NewManager(logger)

	// Create test WebSocket connections
	connections := make([]*websocket.Conn, 0)
	servers := make([]*httptest.Server, 0)

	for i := 0; i < 3; i++ {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			manager.HandleConnection(w, r)
		}))
		servers = append(servers, server)

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("Failed to connect to WebSocket %d: %v", i, err)
		}
		connections = append(connections, conn)
	}

	// Wait for connections to be established
	time.Sleep(50 * time.Millisecond)

	// Verify all clients are registered
	manager.mu.RLock()
	if len(manager.clients) != 3 {
		t.Errorf("Expected 3 clients, got %d", len(manager.clients))
	}
	manager.mu.RUnlock()

	// Broadcast message
	testMessage := map[string]interface{}{
		"type": "broadcast",
		"data": "test broadcast",
	}
	manager.Broadcast(testMessage)

	// Check all clients received the message
	for i, conn := range connections {
		conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		_, response, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("Client %d failed to read message: %v", i, err)
		}

		var resp map[string]interface{}
		if err := json.Unmarshal(response, &resp); err != nil {
			t.Fatalf("Client %d failed to unmarshal response: %v", i, err)
		}

		if resp["type"] != "broadcast" {
			t.Errorf("Client %d expected broadcast message, got: %v", i, resp["type"])
		}
	}

	// Cleanup
	for _, conn := range connections {
		conn.Close()
	}
	for _, server := range servers {
		server.Close()
	}
}

func TestManager_SendToClient(t *testing.T) {
	logger := zaptest.NewLogger(t)
	manager := NewManager(logger)

	// Create test WebSocket connection
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		manager.HandleConnection(w, r)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	// Wait for connection to be established
	time.Sleep(10 * time.Millisecond)

	// Get client ID from clients map
	manager.mu.RLock()
	var clientID string
	for client := range manager.clients {
		clientID = client.ID
		break
	}
	manager.mu.RUnlock()

	if clientID == "" {
		t.Fatal("No client found in manager")
	}

	// Send message to specific client
	testMessage := map[string]interface{}{
		"type": "direct",
		"data": "test direct message",
	}
	manager.SendToClient(clientID, testMessage)

	// Read response
	conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	_, response, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(response, &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp["type"] != "direct" {
		t.Errorf("Expected direct message, got: %v", resp["type"])
	}
}

func TestManager_GetClientCount(t *testing.T) {
	logger := zaptest.NewLogger(t)
	manager := NewManager(logger)

	// Initially should be 0
	if count := manager.GetClientCount(); count != 0 {
		t.Errorf("Expected 0 clients, got %d", count)
	}

	// Create a test connection
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		manager.HandleConnection(w, r)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	// Wait for connection to be established
	time.Sleep(10 * time.Millisecond)

	// Should now be 1
	if count := manager.GetClientCount(); count != 1 {
		t.Errorf("Expected 1 client, got %d", count)
	}
}
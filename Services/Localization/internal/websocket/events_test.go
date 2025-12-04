package websocket

import (
	"encoding/json"
	"testing"
	"time"

	"go.uber.org/zap/zaptest"
)

func TestHandleEvent_Ping(t *testing.T) {
	logger := zaptest.NewLogger(t)
	manager := NewManager(logger)

	// Create ping event message
	pingMsg := map[string]interface{}{
		"type": "ping",
		"data": map[string]interface{}{
			"timestamp": "2023-01-01T00:00:00Z",
		},
	}

	response := manager.handleEvent(pingMsg)
	
	if response == nil {
		t.Fatal("Expected non-nil response")
	}

	if response["type"] != "pong" {
		t.Errorf("Expected pong response, got: %v", response["type"])
	}

	if _, ok := response["timestamp"]; !ok {
		t.Error("Expected timestamp in pong response")
	}
}

func TestHandleEvent_Subscribe(t *testing.T) {
	logger := zaptest.NewLogger(t)
	manager := NewManager(logger)

	// Create subscribe event message
	subscribeMsg := map[string]interface{}{
		"type": "subscribe",
		"data": map[string]interface{}{
			"channel": "localization_updates",
		},
	}

	response := manager.handleEvent(subscribeMsg)
	
	if response == nil {
		t.Fatal("Expected non-nil response")
	}

	if response["type"] != "subscription_confirmed" {
		t.Errorf("Expected subscription_confirmed response, got: %v", response["type"])
	}

	if response["channel"] != "localization_updates" {
		t.Errorf("Expected channel localization_updates, got: %v", response["channel"])
	}
}

func TestHandleEvent_Unsubscribe(t *testing.T) {
	logger := zaptest.NewLogger(t)
	manager := NewManager(logger)

	// Create unsubscribe event message
	unsubscribeMsg := map[string]interface{}{
		"type": "unsubscribe",
		"data": map[string]interface{}{
			"channel": "localization_updates",
		},
	}

	response := manager.handleEvent(unsubscribeMsg)
	
	if response == nil {
		t.Fatal("Expected non-nil response")
	}

	if response["type"] != "unsubscription_confirmed" {
		t.Errorf("Expected unsubscription_confirmed response, got: %v", response["type"])
	}

	if response["channel"] != "localization_updates" {
		t.Errorf("Expected channel localization_updates, got: %v", response["channel"])
	}
}

func TestHandleEvent_Unknown(t *testing.T) {
	logger := zaptest.NewLogger(t)
	manager := NewManager(logger)

	// Create unknown event message
	unknownMsg := map[string]interface{}{
		"type": "unknown_event",
		"data": map[string]interface{}{
			"test": "data",
		},
	}

	response := manager.handleEvent(unknownMsg)
	
	if response == nil {
		t.Fatal("Expected non-nil response")
	}

	if response["type"] != "error" {
		t.Errorf("Expected error response, got: %v", response["type"])
	}

	if response["message"] != "Unknown event type: unknown_event" {
		t.Errorf("Expected error message about unknown event, got: %v", response["message"])
	}
}

func TestHandleEvent_MissingType(t *testing.T) {
	logger := zaptest.NewLogger(t)
	manager := NewManager(logger)

	// Create message without type
	msgWithoutType := map[string]interface{}{
		"data": map[string]interface{}{
			"test": "data",
		},
	}

	response := manager.handleEvent(msgWithoutType)
	
	if response == nil {
		t.Fatal("Expected non-nil response")
	}

	if response["type"] != "error" {
		t.Errorf("Expected error response, got: %v", response["type"])
	}

	if response["message"] != "Event type is required" {
		t.Errorf("Expected error message about missing type, got: %v", response["message"])
	}
}

func TestHandleEvent_MalformedData(t *testing.T) {
	logger := zaptest.NewLogger(t)
	manager := NewManager(logger)

	// Create message with malformed data
	malformedMsg := map[string]interface{}{
		"type": "subscribe",
		"data": "should_be_object",
	}

	response := manager.handleEvent(malformedMsg)
	
	if response == nil {
		t.Fatal("Expected non-nil response")
	}

	if response["type"] != "error" {
		t.Errorf("Expected error response, got: %v", response["type"])
	}

	if response["message"] != "Invalid event data format" {
		t.Errorf("Expected error message about malformed data, got: %v", response["message"])
	}
}

func TestBroadcastLocalizationUpdate(t *testing.T) {
	logger := zaptest.NewLogger(t)
	manager := NewManager(logger)

	// Capture broadcast events by reading from broadcast channel
	go func() {
		<-manager.broadcast
		// Just need to consume from channel to prevent blocking
	}()

	// Call the localization update broadcast
	manager.BroadcastLocalizationUpdate("en", "test.key", "Updated value", "admin")

	// Wait a bit for the message to be processed
	time.Sleep(10 * time.Millisecond)

	// The test passes if no panic occurs
}

func TestBroadcastLanguageCreated(t *testing.T) {
	logger := zaptest.NewLogger(t)
	manager := NewManager(logger)

	// Capture broadcast events by reading from broadcast channel
	go func() {
		<-manager.broadcast
		// Just need to consume from channel to prevent blocking
	}()

	// Call the language created broadcast
	language := map[string]interface{}{
		"id":   "123",
		"code": "fr",
		"name": "French",
		"native_name": "Français",
		"is_rtl": false,
		"is_active": true,
	}
	manager.BroadcastLanguageCreated(language)

	// Wait a bit for the message to be processed
	time.Sleep(10 * time.Millisecond)

	// The test passes if no panic occurs
}

func TestBroadcastVersionCreated(t *testing.T) {
	logger := zaptest.NewLogger(t)
	manager := NewManager(logger)

	// Capture broadcast events by reading from broadcast channel
	go func() {
		<-manager.broadcast
		// Just need to consume from channel to prevent blocking
	}()

	// Call the version created broadcast
	version := map[string]interface{}{
		"id":             "456",
		"version_number": "v1.0.0",
		"description":    "Initial version",
		"keys_count":    10,
		"languages_count": 2,
		"translations_count": 20,
	}
	manager.BroadcastVersionCreated(version)

	// Wait a bit for the message to be processed
	time.Sleep(10 * time.Millisecond)

	// The test passes if no panic occurs
}

func TestSerializeEvent(t *testing.T) {
	logger := zaptest.NewLogger(t)
	manager := NewManager(logger)

	event := map[string]interface{}{
		"type": "test",
		"data": map[string]interface{}{
			"key":   "value",
			"number": 42,
		},
	}

	jsonData, err := manager.serializeEvent(event)
	if err != nil {
		t.Fatalf("Failed to serialize event: %v", err)
	}

	// Parse back to verify
	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("Failed to parse serialized event: %v", err)
	}

	if parsed["type"] != "test" {
		t.Errorf("Expected type 'test', got: %v", parsed["type"])
	}

	data := parsed["data"].(map[string]interface{})
	if data["key"] != "value" {
		t.Errorf("Expected key 'value', got: %v", data["key"])
	}
	if int(data["number"].(float64)) != 42 {
		t.Errorf("Expected number 42, got: %v", data["number"])
	}
}

func TestSerializeEvent_Error(t *testing.T) {
	logger := zaptest.NewLogger(t)
	manager := NewManager(logger)

	// Create unserializable data (channel)
	event := map[string]interface{}{
		"type": "test",
		"data": make(chan int),
	}

	_, err := manager.serializeEvent(event)
	if err == nil {
		t.Error("Expected error when serializing unserializable data")
	}
}
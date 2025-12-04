package websocket

import (
	"go.uber.org/zap"
)

// MockManager is a mock implementation of the WebSocket manager for testing
type MockManager struct {
	broadcastCalls []BroadcastCall
	logger         *zap.Logger
}

// BroadcastCall represents a call to BroadcastEvent
type BroadcastCall struct {
	EventType string
	Data      interface{}
	Metadata  *EventMetadata
}

// NewMockManager creates a new mock WebSocket manager
func NewMockManager() *MockManager {
	logger, _ := zap.NewDevelopment()
	return &MockManager{
		broadcastCalls: make([]BroadcastCall, 0),
		logger:         logger,
	}
}

// BroadcastEvent records the broadcast call for testing
func (m *MockManager) BroadcastEvent(eventType EventType, data interface{}, metadata *EventMetadata) error {
	call := BroadcastCall{
		EventType: string(eventType),
		Data:      data,
		Metadata:  metadata,
	}
	m.broadcastCalls = append(m.broadcastCalls, call)
	return nil
}

// GetBroadcastCalls returns all recorded broadcast calls
func (m *MockManager) GetBroadcastCalls() []BroadcastCall {
	return m.broadcastCalls
}

// ClearBroadcastCalls clears all recorded broadcast calls
func (m *MockManager) ClearBroadcastCalls() {
	m.broadcastCalls = make([]BroadcastCall, 0)
}

// GetClientCount returns 0 for mock
func (m *MockManager) GetClientCount() int {
	return 0
}
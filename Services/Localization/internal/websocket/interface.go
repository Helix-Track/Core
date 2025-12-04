package websocket

// ManagerInterface defines the common interface for WebSocket managers
// Only methods that are used in handlers are included
type ManagerInterface interface {
	BroadcastEvent(eventType EventType, data interface{}, metadata *EventMetadata) error
	GetClientCount() int
}

// Ensure both Manager and MockManager implement the interface
var _ ManagerInterface = (*Manager)(nil)
var _ ManagerInterface = (*MockManager)(nil)
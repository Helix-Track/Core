package models

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Client represents a WebSocket client connection
type Client struct {
	ID           string              // Unique client ID
	Conn         *websocket.Conn     // WebSocket connection
	Username     string              // Authenticated username
	Claims       *JWTClaims          // JWT claims from authentication
	Subscription *Subscription       // Current subscription preferences
	Send         chan []byte         // Buffered channel for outbound messages
	Manager      ClientManager       // Reference to the manager (for unregister)
	mu           sync.RWMutex        // Mutex for thread-safe operations
	Connected    time.Time           // Connection timestamp
	LastPing     time.Time           // Last ping timestamp
	LastActivity time.Time           // Last activity timestamp
	Metadata     map[string]string   // Additional client metadata
}

// ClientManager interface for managing clients (to avoid circular dependency)
type ClientManager interface {
	UnregisterClient(*Client)
}

// NewClient creates a new WebSocket client
func NewClient(id string, conn *websocket.Conn, username string, claims *JWTClaims, manager ClientManager) *Client {
	now := time.Now()
	return &Client{
		ID:           id,
		Conn:         conn,
		Username:     username,
		Claims:       claims,
		Subscription: &Subscription{
			EventTypes:   []EventType{},
			EntityTypes:  []string{},
			EntityIDs:    []string{},
			Filters:      make(map[string]string),
			IncludeReads: false,
		},
		Send:         make(chan []byte, 256), // Buffer size of 256 messages
		Manager:      manager,
		Connected:    now,
		LastPing:     now,
		LastActivity: now,
		Metadata:     make(map[string]string),
	}
}

// UpdateSubscription safely updates the client's subscription
func (c *Client) UpdateSubscription(sub *Subscription) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Subscription = sub
}

// GetSubscription safely gets the client's subscription
func (c *Client) GetSubscription() *Subscription {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Subscription
}

// UpdateActivity updates the last activity timestamp
func (c *Client) UpdateActivity() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.LastActivity = time.Now()
}

// UpdatePing updates the last ping timestamp
func (c *Client) UpdatePing() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.LastPing = time.Now()
}

// GetLastActivity gets the last activity timestamp
func (c *Client) GetLastActivity() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.LastActivity
}

// GetLastPing gets the last ping timestamp
func (c *Client) GetLastPing() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.LastPing
}

// Close closes the client connection and channels
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Close the send channel
	select {
	case <-c.Send:
		// Channel already closed
	default:
		close(c.Send)
	}

	// Close the WebSocket connection
	if c.Conn != nil {
		c.Conn.Close()
	}
}

// HasPermission checks if the client has a specific permission
func (c *Client) HasPermission(permission int) bool {
	if c.Claims == nil {
		return false
	}
	return c.Claims.HasPermission(permission)
}

// WebSocketMessage represents a message sent/received via WebSocket
type WebSocketMessage struct {
	Type    string                 `json:"type"`    // Message type (subscribe, unsubscribe, event, ping, pong, error)
	Action  string                 `json:"action"`  // Action for the message
	Data    map[string]interface{} `json:"data"`    // Message data
	EventID string                 `json:"eventId"` // Event ID if this is an event message
	Error   string                 `json:"error"`   // Error message if applicable
}

// WebSocketMessageType constants
const (
	WSMessageTypeSubscribe   = "subscribe"   // Client subscribes to events
	WSMessageTypeUnsubscribe = "unsubscribe" // Client unsubscribes from events
	WSMessageTypeEvent       = "event"       // Server sends an event
	WSMessageTypePing        = "ping"        // Ping message
	WSMessageTypePong        = "pong"        // Pong response
	WSMessageTypeError       = "error"       // Error message
	WSMessageTypeAck         = "ack"         // Acknowledgment
	WSMessageTypeAuth        = "auth"        // Authentication message
)

// NewWebSocketMessage creates a new WebSocket message
func NewWebSocketMessage(msgType, action string, data map[string]interface{}) *WebSocketMessage {
	return &WebSocketMessage{
		Type:   msgType,
		Action: action,
		Data:   data,
	}
}

// NewErrorMessage creates an error message
func NewErrorMessage(errorMsg string) *WebSocketMessage {
	return &WebSocketMessage{
		Type:  WSMessageTypeError,
		Error: errorMsg,
	}
}

// NewEventMessage creates an event message
func NewEventMessage(event *Event) *WebSocketMessage {
	return &WebSocketMessage{
		Type:    WSMessageTypeEvent,
		Action:  event.Action,
		EventID: event.ID,
		Data: map[string]interface{}{
			"event": event,
		},
	}
}

// WebSocketConfig represents WebSocket configuration
type WebSocketConfig struct {
	Enabled             bool          `json:"enabled"`              // Whether WebSocket is enabled
	Path                string        `json:"path"`                 // WebSocket endpoint path (default: /ws)
	ReadBufferSize      int           `json:"readBufferSize"`       // Read buffer size in bytes
	WriteBufferSize     int           `json:"writeBufferSize"`      // Write buffer size in bytes
	MaxMessageSize      int64         `json:"maxMessageSize"`       // Maximum message size in bytes
	WriteWait           time.Duration `json:"writeWait"`            // Time allowed to write a message
	PongWait            time.Duration `json:"pongWait"`             // Time allowed to read pong
	PingPeriod          time.Duration `json:"pingPeriod"`           // Period for sending pings
	MaxClients          int           `json:"maxClients"`           // Maximum number of concurrent clients
	RequireAuth         bool          `json:"requireAuth"`          // Whether authentication is required
	AllowOrigins        []string      `json:"allowOrigins"`         // Allowed origins for CORS
	EnableCompression   bool          `json:"enableCompression"`    // Enable per-message compression
	HandshakeTimeout    time.Duration `json:"handshakeTimeout"`     // WebSocket handshake timeout
}

// DefaultWebSocketConfig returns default WebSocket configuration
func DefaultWebSocketConfig() WebSocketConfig {
	return WebSocketConfig{
		Enabled:           true,
		Path:              "/ws",
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		MaxMessageSize:    512 * 1024, // 512KB
		WriteWait:         10 * time.Second,
		PongWait:          60 * time.Second,
		PingPeriod:        54 * time.Second, // Must be less than pongWait
		MaxClients:        1000,
		RequireAuth:       true,
		AllowOrigins:      []string{"*"},
		EnableCompression: true,
		HandshakeTimeout:  10 * time.Second,
	}
}

package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for development - should be restricted in production
		return true
	},
}

// Client represents a WebSocket client connection
type Client struct {
	ID         string
	Conn       *websocket.Conn
	Send       chan []byte
	Manager    *Manager
	Subscribed map[EventType]bool
	mu         sync.RWMutex
}

// Manager manages all WebSocket connections
type Manager struct {
	clients    map[*Client]bool
	Clients    map[string]*Client // Exposed for testing
	broadcast  chan *Event
	register   chan *Client
	unregister chan *Client
	logger     *zap.Logger
	mu         sync.RWMutex
}

// NewManager creates a new WebSocket manager
func NewManager(logger *zap.Logger) *Manager {
	return &Manager{
		clients:    make(map[*Client]bool),
		Clients:    make(map[string]*Client), // Exposed for testing
		broadcast:  make(chan *Event, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		logger:     logger,
	}
}

// Start starts the WebSocket manager
func (m *Manager) Start(ctx context.Context) {
	m.logger.Info("WebSocket manager started")

	for {
		select {
		case <-ctx.Done():
			m.logger.Info("WebSocket manager stopping")
			return

		case client := <-m.register:
			m.mu.Lock()
			m.clients[client] = true
			m.Clients[client.ID] = client // Add to exposed map
			m.mu.Unlock()
			m.logger.Info("Client registered", zap.String("clientId", client.ID))

		case client := <-m.unregister:
			m.mu.Lock()
			if _, ok := m.clients[client]; ok {
				delete(m.clients, client)
				delete(m.Clients, client.ID) // Remove from exposed map
				close(client.Send)
			}
			m.mu.Unlock()
			m.logger.Info("Client unregistered", zap.String("clientId", client.ID))

		case event := <-m.broadcast:
			m.broadcastToClients(event)
		}
	}
}

// HandleConnection handles a new WebSocket connection
func (m *Manager) HandleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		m.logger.Error("Failed to upgrade connection", zap.Error(err))
		return
	}

	client := &Client{
		ID:         generateClientID(),
		Conn:       conn,
		Send:       make(chan []byte, 256),
		Manager:    m,
		Subscribed: make(map[EventType]bool),
	}

	// Subscribe to all events by default
	client.SubscribeToAll()

	m.register <- client

	// Start read and write pumps
	go client.writePump()
	go client.readPump()
}

// BroadcastEvent broadcasts an event to all subscribed clients
func (m *Manager) BroadcastEvent(eventType EventType, data interface{}, metadata *EventMetadata) error {
	event, err := NewEvent(eventType, data, metadata)
	if err != nil {
		m.logger.Error("Failed to create event", zap.Error(err))
		return err
	}

	m.broadcast <- event
	return nil
}

// GetClientCount returns the number of connected clients
func (m *Manager) GetClientCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.clients)
}

// broadcastToClients sends an event to all subscribed clients
func (m *Manager) broadcastToClients(event *Event) {
	eventJSON, err := event.ToJSON()
	if err != nil {
		m.logger.Error("Failed to serialize event", zap.Error(err))
		return
	}

	m.mu.RLock()
	clients := make([]*Client, 0, len(m.clients))
	for client := range m.clients {
		clients = append(clients, client)
	}
	m.mu.RUnlock()

	for _, client := range clients {
		if client.IsSubscribed(event.Type) {
			select {
			case client.Send <- eventJSON:
			default:
				m.logger.Warn("Client send buffer full, dropping message",
					zap.String("clientId", client.ID))
			}
		}
	}

	m.logger.Debug("Event broadcasted",
		zap.String("eventType", string(event.Type)),
		zap.Int("clients", len(clients)))
}

// writePump pumps messages from the hub to the websocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// readPump pumps messages from the websocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.Manager.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.Manager.logger.Error("Unexpected close error", zap.Error(err))
			}
			break
		}

		// Handle client messages (subscription updates, etc.)
		c.handleMessage(message)
	}
}

// handleMessage handles messages from the client
func (c *Client) handleMessage(message []byte) {
	var msg map[string]interface{}
	if err := json.Unmarshal(message, &msg); err != nil {
		c.Manager.logger.Error("Failed to parse client message", zap.Error(err))
		return
	}

	action, ok := msg["action"].(string)
	if !ok {
		return
	}

	switch action {
	case "subscribe":
		if eventType, ok := msg["eventType"].(string); ok {
			c.Subscribe(EventType(eventType))
		}
	case "unsubscribe":
		if eventType, ok := msg["eventType"].(string); ok {
			c.Unsubscribe(EventType(eventType))
		}
	case "ping":
		c.sendPong()
	}
}

// Subscribe subscribes the client to a specific event type
func (c *Client) Subscribe(eventType EventType) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Subscribed[eventType] = true
}

// Unsubscribe unsubscribes the client from a specific event type
func (c *Client) Unsubscribe(eventType EventType) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.Subscribed, eventType)
}

// SubscribeToAll subscribes the client to all event types
func (c *Client) SubscribeToAll() {
	c.mu.Lock()
	defer c.mu.Unlock()

	allEvents := []EventType{
		EventLanguageAdded, EventLanguageUpdated, EventLanguageDeleted,
		EventKeyAdded, EventKeyUpdated, EventKeyDeleted,
		EventLocalizationAdded, EventLocalizationUpdated, EventLocalizationDeleted, EventLocalizationApproved,
		EventBatchOperationCompleted,
		EventCatalogRebuilt, EventCacheInvalidated,
		EventVersionCreated, EventVersionDeleted,
	}

	for _, eventType := range allEvents {
		c.Subscribed[eventType] = true
	}
}

// IsSubscribed checks if the client is subscribed to an event type
func (c *Client) IsSubscribed(eventType EventType) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Subscribed[eventType]
}

// sendPong sends a pong message to the client
func (c *Client) sendPong() {
	pongMsg := map[string]string{"type": "pong"}
	pongJSON, _ := json.Marshal(pongMsg)
	select {
	case c.Send <- pongJSON:
	default:
	}
}

// generateClientID generates a unique client ID
func generateClientID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString generates a random string of the given length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// Broadcast broadcasts a generic message to all clients
func (m *Manager) Broadcast(message map[string]interface{}) {
	// Create a generic event for broadcasting
	event := &Event{
		Type:      EventType(message["type"].(string)),
		Timestamp: time.Now().UTC(),
		Data:      toJSONBytes(message),
	}
	
	m.broadcast <- event
}

// SendToClient sends a message to a specific client
func (m *Manager) SendToClient(clientID string, message map[string]interface{}) {
	m.mu.RLock()
	client, exists := m.Clients[clientID]
	m.mu.RUnlock()
	
	if !exists {
		m.logger.Warn("Client not found", zap.String("clientId", clientID))
		return
	}
	
	event := &Event{
		Type:      EventType(message["type"].(string)),
		Timestamp: time.Now().UTC(),
		Data:      toJSONBytes(message),
	}
	
	eventJSON, _ := event.ToJSON()
	
	select {
	case client.Send <- eventJSON:
	default:
		m.logger.Warn("Client send buffer full, dropping message",
			zap.String("clientId", clientID))
	}
}

// handleEvent handles WebSocket events from clients
func (m *Manager) handleEvent(message map[string]interface{}) map[string]interface{} {
	// Check if message has a type
	eventType, ok := message["type"].(string)
	if !ok {
		return map[string]interface{}{
			"type":    "error",
			"message": "Event type is required",
		}
	}
	
	// Check if message has valid data
	data, ok := message["data"]
	if !ok {
		return map[string]interface{}{
			"type":    "error",
			"message": "Event data is required",
		}
	}
	
	// Ensure data is a map for known events
	if eventType != "ping" && eventType != "pong" {
		if _, isMap := data.(map[string]interface{}); !isMap {
			return map[string]interface{}{
				"type":    "error",
				"message": "Invalid event data format",
			}
		}
	}
	
	// Handle different event types
	switch eventType {
	case "ping":
		return map[string]interface{}{
			"type":      "pong",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}
	case "subscribe":
		dataMap := data.(map[string]interface{})
		channel := dataMap["channel"].(string)
		return map[string]interface{}{
			"type":    "subscription_confirmed",
			"channel": channel,
		}
	case "unsubscribe":
		dataMap := data.(map[string]interface{})
		channel := dataMap["channel"].(string)
		return map[string]interface{}{
			"type":    "unsubscription_confirmed",
			"channel": channel,
		}
	default:
		return map[string]interface{}{
			"type":    "error",
			"message": "Unknown event type: " + eventType,
		}
	}
}

// BroadcastLocalizationUpdate broadcasts a localization update event
func (m *Manager) BroadcastLocalizationUpdate(languageCode, key, value, updatedBy string) {
	data := LocalizationEventData{
		LanguageCode: languageCode,
		Key:          key,
		Value:        value,
	}
	
	metadata := &EventMetadata{
		Username: updatedBy,
	}
	
	m.BroadcastEvent(EventLocalizationUpdated, data, metadata)
}

// BroadcastLanguageCreated broadcasts a language created event
func (m *Manager) BroadcastLanguageCreated(language map[string]interface{}) {
	data := LanguageEventData{
		ID:         language["id"].(string),
		Code:       language["code"].(string),
		Name:       language["name"].(string),
		NativeName: language["native_name"].(string),
		IsRTL:      language["is_rtl"].(bool),
		IsActive:   language["is_active"].(bool),
	}
	
	m.BroadcastEvent(EventLanguageAdded, data, nil)
}

// BroadcastVersionCreated broadcasts a version created event
func (m *Manager) BroadcastVersionCreated(version map[string]interface{}) {
	data := VersionEventData{
		ID:                 version["id"].(string),
		Version:            version["version_number"].(string),
		Description:        version["description"].(string),
		KeysCount:          version["keys_count"].(int),
		LanguagesCount:     version["languages_count"].(int),
		TranslationsCount:  version["translations_count"].(int),
	}
	
	m.BroadcastEvent(EventVersionCreated, data, nil)
}

// serializeEvent serializes an event to JSON
func (m *Manager) serializeEvent(event map[string]interface{}) ([]byte, error) {
	return json.Marshal(event)
}

// toJSONBytes converts interface to JSON bytes
func toJSONBytes(data interface{}) json.RawMessage {
	bytes, _ := json.Marshal(data)
	return json.RawMessage(bytes)
}

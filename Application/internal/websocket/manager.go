package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"helixtrack.ru/core/internal/logger"
	"helixtrack.ru/core/internal/models"
	"helixtrack.ru/core/internal/services"
)

// Manager manages WebSocket connections and event broadcasting
type Manager struct {
	clients         map[string]*models.Client   // Registered clients by ID
	clientsByUser   map[string][]*models.Client // Clients indexed by username
	register        chan *models.Client         // Channel for registering clients
	unregister      chan *models.Client         // Channel for unregistering clients
	broadcast       chan *models.Event          // Channel for broadcasting events
	config          models.WebSocketConfig      // WebSocket configuration
	upgrader        websocket.Upgrader          // WebSocket upgrader
	permService     services.PermissionService  // Permission service for authorization
	mu              sync.RWMutex                // Mutex for thread-safe operations
	ctx             context.Context             // Context for cancellation
	cancel          context.CancelFunc          // Cancel function
	wg              sync.WaitGroup              // Wait group for graceful shutdown
	running         bool                        // Whether the manager is running
	stats           ManagerStats                // Manager statistics
}

// ManagerStats contains statistics about the WebSocket manager
type ManagerStats struct {
	TotalConnections    int64     // Total connections since start
	ActiveConnections   int       // Current active connections
	TotalEvents         int64     // Total events broadcasted
	TotalErrors         int64     // Total errors encountered
	StartTime           time.Time // Manager start time
	LastEventTime       time.Time // Last event broadcast time
	mu                  sync.RWMutex
}

// NewManager creates a new WebSocket manager
func NewManager(config models.WebSocketConfig, permService services.PermissionService) *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	// Configure WebSocket upgrader
	upgrader := websocket.Upgrader{
		ReadBufferSize:   config.ReadBufferSize,
		WriteBufferSize:  config.WriteBufferSize,
		CheckOrigin:      createOriginChecker(config.AllowOrigins),
		HandshakeTimeout: config.HandshakeTimeout,
	}

	if config.EnableCompression {
		upgrader.EnableCompression = true
	}

	return &Manager{
		clients:       make(map[string]*models.Client),
		clientsByUser: make(map[string][]*models.Client),
		register:      make(chan *models.Client, 10),
		unregister:    make(chan *models.Client, 10),
		broadcast:     make(chan *models.Event, 256),
		config:        config,
		upgrader:      upgrader,
		permService:   permService,
		ctx:           ctx,
		cancel:        cancel,
		stats: ManagerStats{
			StartTime: time.Now(),
		},
	}
}

// Start starts the WebSocket manager
func (m *Manager) Start() error {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return fmt.Errorf("manager already running")
	}
	m.running = true
	m.mu.Unlock()

	logger.Info("Starting WebSocket manager",
		zap.Int("maxClients", m.config.MaxClients),
		zap.String("path", m.config.Path),
	)

	m.wg.Add(1)
	go m.run()

	return nil
}

// Stop stops the WebSocket manager gracefully
func (m *Manager) Stop() error {
	m.mu.Lock()
	if !m.running {
		m.mu.Unlock()
		return nil
	}
	m.running = false
	m.mu.Unlock()

	logger.Info("Stopping WebSocket manager")

	// Cancel context to stop all goroutines
	m.cancel()

	// Close all channels
	close(m.register)
	close(m.unregister)
	close(m.broadcast)

	// Wait for all goroutines to finish
	m.wg.Wait()

	// Close all client connections
	m.mu.Lock()
	for _, client := range m.clients {
		client.Close()
	}
	m.clients = make(map[string]*models.Client)
	m.clientsByUser = make(map[string][]*models.Client)
	m.mu.Unlock()

	logger.Info("WebSocket manager stopped")
	return nil
}

// run is the main event loop for the manager
func (m *Manager) run() {
	defer m.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return

		case client := <-m.register:
			m.registerClient(client)

		case client := <-m.unregister:
			m.unregisterClient(client)

		case event := <-m.broadcast:
			m.broadcastEvent(event)

		case <-ticker.C:
			m.cleanupStaleConnections()
		}
	}
}

// RegisterClient registers a new WebSocket client
func (m *Manager) RegisterClient(client *models.Client) error {
	// Check max clients limit
	m.mu.RLock()
	currentCount := len(m.clients)
	m.mu.RUnlock()

	if currentCount >= m.config.MaxClients {
		return fmt.Errorf("maximum client limit reached")
	}

	select {
	case m.register <- client:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timeout registering client")
	}
}

// registerClient handles client registration (internal)
func (m *Manager) registerClient(client *models.Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Add to clients map
	m.clients[client.ID] = client

	// Add to user index
	if _, exists := m.clientsByUser[client.Username]; !exists {
		m.clientsByUser[client.Username] = make([]*models.Client, 0)
	}
	m.clientsByUser[client.Username] = append(m.clientsByUser[client.Username], client)

	// Update stats
	m.stats.mu.Lock()
	m.stats.TotalConnections++
	m.stats.ActiveConnections = len(m.clients)
	m.stats.mu.Unlock()

	logger.Info("Client registered",
		zap.String("clientId", client.ID),
		zap.String("username", client.Username),
		zap.Int("activeClients", len(m.clients)),
	)

	// Send connection established event
	connectionEvent := models.NewEvent(
		models.EventConnectionEstablished,
		"connect",
		"connection",
		client.ID,
		client.Username,
		map[string]interface{}{
			"clientId": client.ID,
			"time":     time.Now().UTC(),
		},
	)
	m.sendToClient(client, connectionEvent)

	// Start client read/write pumps
	go m.readPump(client)
	go m.writePump(client)
}

// UnregisterClient unregisters a WebSocket client
func (m *Manager) UnregisterClient(client *models.Client) {
	select {
	case m.unregister <- client:
	case <-time.After(5 * time.Second):
		logger.Error("Timeout unregistering client", zap.String("clientId", client.ID))
	}
}

// unregisterClient handles client unregistration (internal)
func (m *Manager) unregisterClient(client *models.Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Remove from clients map
	if _, exists := m.clients[client.ID]; exists {
		delete(m.clients, client.ID)

		// Remove from user index
		if userClients, exists := m.clientsByUser[client.Username]; exists {
			for i, c := range userClients {
				if c.ID == client.ID {
					m.clientsByUser[client.Username] = append(userClients[:i], userClients[i+1:]...)
					break
				}
			}
			// Remove user entry if no more clients
			if len(m.clientsByUser[client.Username]) == 0 {
				delete(m.clientsByUser, client.Username)
			}
		}

		// Close client connection
		client.Close()

		// Update stats
		m.stats.mu.Lock()
		m.stats.ActiveConnections = len(m.clients)
		m.stats.mu.Unlock()

		logger.Info("Client unregistered",
			zap.String("clientId", client.ID),
			zap.String("username", client.Username),
			zap.Int("activeClients", len(m.clients)),
		)
	}
}

// BroadcastEvent broadcasts an event to all subscribed clients
func (m *Manager) BroadcastEvent(event *models.Event) {
	select {
	case m.broadcast <- event:
	case <-time.After(5 * time.Second):
		logger.Error("Timeout broadcasting event", zap.String("eventId", event.ID))
		m.stats.mu.Lock()
		m.stats.TotalErrors++
		m.stats.mu.Unlock()
	}
}

// broadcastEvent handles event broadcasting (internal)
func (m *Manager) broadcastEvent(event *models.Event) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Update stats
	m.stats.mu.Lock()
	m.stats.TotalEvents++
	m.stats.LastEventTime = time.Now()
	m.stats.mu.Unlock()

	logger.Debug("Broadcasting event",
		zap.String("eventId", event.ID),
		zap.String("type", string(event.Type)),
		zap.String("object", event.Object),
		zap.String("entityId", event.EntityID),
	)

	// Send event to all matching clients
	for _, client := range m.clients {
		// Check if client is subscribed to this event
		subscription := client.GetSubscription()
		if !event.MatchesSubscription(subscription) {
			continue
		}

		// Check if client has permission to see this event
		if !m.canClientReceiveEvent(client, event) {
			continue
		}

		// Send event to client
		m.sendToClient(client, event)
	}
}

// canClientReceiveEvent checks if a client has permission to receive an event
func (m *Manager) canClientReceiveEvent(client *models.Client, event *models.Event) bool {
	// If no permission context, allow
	if len(event.Context.Permissions) == 0 {
		return true
	}

	// Check if client has required permissions
	for _, requiredPerm := range event.Context.Permissions {
		permLevel := models.ParsePermissionLevel(requiredPerm)
		if !client.HasPermission(int(permLevel)) {
			return false
		}
	}

	// If permission service is enabled, check with service
	if m.permService != nil && m.permService.IsEnabled() {
		allowed, err := m.permService.CheckPermission(
			m.ctx,
			client.Username,
			event.Object,
			models.PermissionRead,
		)
		if err != nil || !allowed {
			return false
		}
	}

	return true
}

// sendToClient sends an event to a specific client
func (m *Manager) sendToClient(client *models.Client, event *models.Event) {
	// Create WebSocket message
	message := models.NewEventMessage(event)
	data, err := json.Marshal(message)
	if err != nil {
		logger.Error("Failed to marshal event",
			zap.Error(err),
			zap.String("eventId", event.ID),
		)
		return
	}

	// Send to client's send channel
	select {
	case client.Send <- data:
	default:
		// Client buffer is full, disconnect
		logger.Warn("Client send buffer full, disconnecting",
			zap.String("clientId", client.ID),
			zap.String("username", client.Username),
		)
		m.UnregisterClient(client)
	}
}

// readPump reads messages from the WebSocket connection
func (m *Manager) readPump(client *models.Client) {
	defer func() {
		m.UnregisterClient(client)
	}()

	client.Conn.SetReadLimit(m.config.MaxMessageSize)
	client.Conn.SetReadDeadline(time.Now().Add(m.config.PongWait))
	client.Conn.SetPongHandler(func(string) error {
		client.UpdatePing()
		client.Conn.SetReadDeadline(time.Now().Add(m.config.PongWait))
		return nil
	})

	for {
		_, messageData, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("WebSocket read error",
					zap.Error(err),
					zap.String("clientId", client.ID),
				)
			}
			break
		}

		client.UpdateActivity()

		// Parse message
		var message models.WebSocketMessage
		if err := json.Unmarshal(messageData, &message); err != nil {
			logger.Error("Failed to unmarshal WebSocket message",
				zap.Error(err),
				zap.String("clientId", client.ID),
			)
			m.sendError(client, "Invalid message format")
			continue
		}

		// Handle message
		m.handleClientMessage(client, &message)
	}
}

// writePump writes messages to the WebSocket connection
func (m *Manager) writePump(client *models.Client) {
	ticker := time.NewTicker(m.config.PingPeriod)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(m.config.WriteWait))
			if !ok {
				// Channel closed
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(client.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-client.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(m.config.WriteWait))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleClientMessage handles incoming messages from clients
func (m *Manager) handleClientMessage(client *models.Client, message *models.WebSocketMessage) {
	switch message.Type {
	case models.WSMessageTypeSubscribe:
		m.handleSubscribe(client, message)
	case models.WSMessageTypeUnsubscribe:
		m.handleUnsubscribe(client, message)
	case models.WSMessageTypePing:
		m.handlePing(client)
	default:
		logger.Warn("Unknown message type",
			zap.String("type", message.Type),
			zap.String("clientId", client.ID),
		)
		m.sendError(client, "Unknown message type")
	}
}

// handleSubscribe handles subscription requests
func (m *Manager) handleSubscribe(client *models.Client, message *models.WebSocketMessage) {
	// Parse subscription from message data
	subscriptionData, err := json.Marshal(message.Data)
	if err != nil {
		m.sendError(client, "Invalid subscription data")
		return
	}

	var subscription models.Subscription
	if err := json.Unmarshal(subscriptionData, &subscription); err != nil {
		m.sendError(client, "Invalid subscription format")
		return
	}

	// Update client subscription
	client.UpdateSubscription(&subscription)

	logger.Info("Client subscribed",
		zap.String("clientId", client.ID),
		zap.String("username", client.Username),
		zap.Int("eventTypes", len(subscription.EventTypes)),
		zap.Int("entityTypes", len(subscription.EntityTypes)),
	)

	// Send acknowledgment
	ack := models.NewWebSocketMessage(models.WSMessageTypeAck, "subscribe", map[string]interface{}{
		"success": true,
		"message": "Subscription updated",
	})
	ackData, _ := json.Marshal(ack)
	select {
	case client.Send <- ackData:
	default:
	}
}

// handleUnsubscribe handles unsubscription requests
func (m *Manager) handleUnsubscribe(client *models.Client, message *models.WebSocketMessage) {
	// Reset subscription
	client.UpdateSubscription(&models.Subscription{
		EventTypes:   []models.EventType{},
		EntityTypes:  []string{},
		EntityIDs:    []string{},
		Filters:      make(map[string]string),
		IncludeReads: false,
	})

	logger.Info("Client unsubscribed",
		zap.String("clientId", client.ID),
		zap.String("username", client.Username),
	)

	// Send acknowledgment
	ack := models.NewWebSocketMessage(models.WSMessageTypeAck, "unsubscribe", map[string]interface{}{
		"success": true,
		"message": "Unsubscribed",
	})
	ackData, _ := json.Marshal(ack)
	select {
	case client.Send <- ackData:
	default:
	}
}

// handlePing handles ping messages
func (m *Manager) handlePing(client *models.Client) {
	client.UpdatePing()

	// Send pong response
	pong := models.NewWebSocketMessage(models.WSMessageTypePong, "pong", map[string]interface{}{
		"time": time.Now().UTC(),
	})
	pongData, _ := json.Marshal(pong)
	select {
	case client.Send <- pongData:
	default:
	}
}

// sendError sends an error message to a client
func (m *Manager) sendError(client *models.Client, errorMsg string) {
	errorMessage := models.NewErrorMessage(errorMsg)
	data, _ := json.Marshal(errorMessage)
	select {
	case client.Send <- data:
	default:
	}
}

// cleanupStaleConnections removes stale connections
func (m *Manager) cleanupStaleConnections() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	now := time.Now()
	timeout := 2 * m.config.PongWait

	for _, client := range m.clients {
		lastPing := client.GetLastPing()
		if now.Sub(lastPing) > timeout {
			logger.Warn("Removing stale connection",
				zap.String("clientId", client.ID),
				zap.String("username", client.Username),
				zap.Duration("lastPing", now.Sub(lastPing)),
			)
			m.UnregisterClient(client)
		}
	}
}

// GetStats returns current manager statistics
func (m *Manager) GetStats() ManagerStats {
	m.stats.mu.RLock()
	defer m.stats.mu.RUnlock()
	return m.stats
}

// GetUpgrader returns the WebSocket upgrader
func (m *Manager) GetUpgrader() *websocket.Upgrader {
	return &m.upgrader
}

// CreateClient creates a new client instance
func (m *Manager) CreateClient(conn *websocket.Conn, username string, claims *models.JWTClaims) *models.Client {
	clientID := uuid.New().String()
	return models.NewClient(clientID, conn, username, claims, m)
}

// createOriginChecker creates an origin checker function for the upgrader
func createOriginChecker(allowedOrigins []string) func(r *http.Request) bool {
	return func(r *http.Request) bool {
		// If no origins specified or * is specified, allow all
		if len(allowedOrigins) == 0 || contains(allowedOrigins, "*") {
			return true
		}

		origin := r.Header.Get("Origin")
		return contains(allowedOrigins, origin)
	}
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

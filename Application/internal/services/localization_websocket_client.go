package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// LocalizationWebSocketClient handles real-time updates from the Localization service
type LocalizationWebSocketClient struct {
	serviceURL string
	conn       *websocket.Conn
	connMu     sync.RWMutex
	service    *LocalizationService
	logger     *zap.Logger

	// Connection state
	connected   bool
	reconnectCh chan struct{}
	stopCh      chan struct{}
	doneCh      chan struct{}
}

// WebSocketEvent represents an event from the WebSocket
type WebSocketEvent struct {
	Type      string          `json:"type"`
	Timestamp time.Time       `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
	Metadata  *EventMetadata  `json:"metadata,omitempty"`
}

// EventMetadata contains metadata about the event
type EventMetadata struct {
	UserID        string `json:"userId,omitempty"`
	Username      string `json:"username,omitempty"`
	CorrelationID string `json:"correlationId,omitempty"`
}

// CacheInvalidatedEventData represents cache invalidation event data
type CacheInvalidatedEventData struct {
	Language string `json:"language,omitempty"`
	Reason   string `json:"reason"`
}

// Event type constants
const (
	EventTypeLanguageAdded              = "language.added"
	EventTypeLanguageUpdated            = "language.updated"
	EventTypeLanguageDeleted            = "language.deleted"
	EventTypeKeyAdded                   = "key.added"
	EventTypeKeyUpdated                 = "key.updated"
	EventTypeKeyDeleted                 = "key.deleted"
	EventTypeLocalizationAdded          = "localization.added"
	EventTypeLocalizationUpdated        = "localization.updated"
	EventTypeLocalizationDeleted        = "localization.deleted"
	EventTypeLocalizationApproved       = "localization.approved"
	EventTypeCacheInvalidated           = "cache.invalidated"
	EventTypeVersionCreated             = "version.created"
	EventTypeVersionDeleted             = "version.deleted"
	EventTypeBatchOperationCompleted    = "batch.completed"
)

// NewLocalizationWebSocketClient creates a new WebSocket client
func NewLocalizationWebSocketClient(serviceURL string, service *LocalizationService, logger *zap.Logger) *LocalizationWebSocketClient {
	return &LocalizationWebSocketClient{
		serviceURL:  serviceURL,
		service:     service,
		logger:      logger,
		reconnectCh: make(chan struct{}, 1),
		stopCh:      make(chan struct{}),
		doneCh:      make(chan struct{}),
	}
}

// Start starts the WebSocket client and maintains the connection
func (c *LocalizationWebSocketClient) Start(ctx context.Context) error {
	c.logger.Info("starting localization websocket client", zap.String("url", c.serviceURL))

	// Initial connection
	if err := c.connect(ctx); err != nil {
		c.logger.Error("failed to connect to websocket", zap.Error(err))
		// Don't fail, we'll retry
	}

	// Start the connection manager
	go c.connectionManager(ctx)

	// Wait for context cancellation or explicit stop
	select {
	case <-ctx.Done():
		c.logger.Info("websocket client stopped by context")
	case <-c.stopCh:
		c.logger.Info("websocket client stopped explicitly")
	}

	c.disconnect()
	close(c.doneCh)
	return nil
}

// Stop stops the WebSocket client
func (c *LocalizationWebSocketClient) Stop() {
	close(c.stopCh)
	<-c.doneCh
}

// connect establishes a WebSocket connection
func (c *LocalizationWebSocketClient) connect(ctx context.Context) error {
	c.connMu.Lock()
	defer c.connMu.Unlock()

	// Parse the service URL and convert to WebSocket URL
	u, err := url.Parse(c.serviceURL)
	if err != nil {
		return fmt.Errorf("invalid service URL: %w", err)
	}

	// Convert http/https to ws/wss
	wsScheme := "ws"
	if u.Scheme == "https" {
		wsScheme = "wss"
	}

	wsURL := fmt.Sprintf("%s://%s/ws", wsScheme, u.Host)

	c.logger.Info("connecting to websocket", zap.String("url", wsURL))

	// Create dialer with context
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	// Connect
	conn, _, err := dialer.DialContext(ctx, wsURL, nil)
	if err != nil {
		return fmt.Errorf("failed to dial websocket: %w", err)
	}

	c.conn = conn
	c.connected = true

	c.logger.Info("websocket connected successfully")

	// Start read loop
	go c.readLoop()

	return nil
}

// disconnect closes the WebSocket connection
func (c *LocalizationWebSocketClient) disconnect() {
	c.connMu.Lock()
	defer c.connMu.Unlock()

	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
		c.connected = false
		c.logger.Info("websocket disconnected")
	}
}

// connectionManager handles reconnection logic
func (c *LocalizationWebSocketClient) connectionManager(ctx context.Context) {
	reconnectDelay := 5 * time.Second
	maxReconnectDelay := 60 * time.Second

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopCh:
			return
		case <-c.reconnectCh:
			c.logger.Info("reconnecting to websocket", zap.Duration("delay", reconnectDelay))
			time.Sleep(reconnectDelay)

			if err := c.connect(ctx); err != nil {
				c.logger.Error("reconnection failed", zap.Error(err))

				// Exponential backoff
				reconnectDelay *= 2
				if reconnectDelay > maxReconnectDelay {
					reconnectDelay = maxReconnectDelay
				}

				// Trigger another reconnect
				select {
				case c.reconnectCh <- struct{}{}:
				default:
				}
			} else {
				// Reset reconnect delay on success
				reconnectDelay = 5 * time.Second
			}
		}
	}
}

// readLoop reads messages from the WebSocket
func (c *LocalizationWebSocketClient) readLoop() {
	defer func() {
		c.disconnect()

		// Trigger reconnect
		select {
		case c.reconnectCh <- struct{}{}:
		default:
		}
	}()

	for {
		c.connMu.RLock()
		conn := c.conn
		c.connMu.RUnlock()

		if conn == nil {
			return
		}

		var event WebSocketEvent
		if err := conn.ReadJSON(&event); err != nil {
			c.logger.Error("failed to read websocket message", zap.Error(err))
			return
		}

		c.handleEvent(&event)
	}
}

// handleEvent processes a WebSocket event
func (c *LocalizationWebSocketClient) handleEvent(event *WebSocketEvent) {
	c.logger.Debug("received websocket event",
		zap.String("type", event.Type),
		zap.Time("timestamp", event.Timestamp),
	)

	switch event.Type {
	case EventTypeCacheInvalidated:
		c.handleCacheInvalidated(event)
	case EventTypeLanguageAdded, EventTypeLanguageUpdated, EventTypeLanguageDeleted:
		c.handleLanguageEvent(event)
	case EventTypeLocalizationAdded, EventTypeLocalizationUpdated,
		 EventTypeLocalizationDeleted, EventTypeLocalizationApproved:
		c.handleLocalizationEvent(event)
	case EventTypeBatchOperationCompleted:
		c.handleBatchOperation(event)
	default:
		c.logger.Debug("unhandled event type", zap.String("type", event.Type))
	}
}

// handleCacheInvalidated handles cache invalidation events
func (c *LocalizationWebSocketClient) handleCacheInvalidated(event *WebSocketEvent) {
	var data CacheInvalidatedEventData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		c.logger.Error("failed to unmarshal cache invalidated event", zap.Error(err))
		return
	}

	c.logger.Info("cache invalidation received",
		zap.String("language", data.Language),
		zap.String("reason", data.Reason),
	)

	if data.Language != "" {
		// Invalidate specific language
		c.service.cache.Invalidate(data.Language)
		c.logger.Info("invalidated cache for language", zap.String("language", data.Language))
	} else {
		// Invalidate all languages
		c.service.cache.InvalidateAll()
		c.logger.Info("invalidated all language caches")
	}
}

// handleLanguageEvent handles language CRUD events
func (c *LocalizationWebSocketClient) handleLanguageEvent(event *WebSocketEvent) {
	// Language changes may affect catalogs, invalidate all caches
	c.logger.Info("language event received, invalidating all caches", zap.String("type", event.Type))
	c.service.cache.InvalidateAll()
}

// handleLocalizationEvent handles localization CRUD events
func (c *LocalizationWebSocketClient) handleLocalizationEvent(event *WebSocketEvent) {
	// Localization changes require cache refresh
	c.logger.Info("localization event received, invalidating all caches", zap.String("type", event.Type))
	c.service.cache.InvalidateAll()
}

// handleBatchOperation handles batch operation events
func (c *LocalizationWebSocketClient) handleBatchOperation(event *WebSocketEvent) {
	// Batch operations may affect multiple languages
	c.logger.Info("batch operation completed, invalidating all caches", zap.String("type", event.Type))
	c.service.cache.InvalidateAll()
}

// IsConnected returns the connection status
func (c *LocalizationWebSocketClient) IsConnected() bool {
	c.connMu.RLock()
	defer c.connMu.RUnlock()
	return c.connected
}

// Invalidate manually invalidates cache for a language
func (lc *LocalizationCache) Invalidate(language string) {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	delete(lc.catalogs, language)
}

// InvalidateAll invalidates all cached catalogs
func (lc *LocalizationCache) InvalidateAll() {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	lc.catalogs = make(map[string]*CachedCatalog)
}

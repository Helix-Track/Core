package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"go.uber.org/zap"

	"helixtrack.ru/chat/internal/database"
	"helixtrack.ru/chat/internal/handlers"
	"helixtrack.ru/chat/internal/logger"
	"helixtrack.ru/chat/internal/middleware"
	"helixtrack.ru/chat/internal/models"
	"helixtrack.ru/chat/internal/services"
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for now - in production, check against allowed origins
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Server represents the HTTP/3 QUIC server
type Server struct {
	config      *models.Config
	router      *gin.Engine
	httpServer  *http.Server
	http3Server *http3.Server
	db          database.Database
	coreService services.CoreService
	handler     *handlers.Handler
	connections map[string]*websocket.Conn // userID -> connection
	connMutex   sync.RWMutex
}

// NewServer creates a new HTTP/3 QUIC server
func NewServer(config *models.Config, db database.Database, coreService services.CoreService) *Server {
	// Set Gin mode
	if config.Logger.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Create handler
	handler := handlers.NewHandler(db, coreService)

	return &Server{
		config:      config,
		router:      router,
		db:          db,
		coreService: coreService,
		handler:     handler,
	}
}

// Start starts the HTTP/3 QUIC server
func (s *Server) Start() error {
	// Setup middleware
	s.setupMiddleware()

	// Setup routes
	s.setupRoutes()

	// Server address
	addr := fmt.Sprintf("%s:%d", s.config.Server.Address, s.config.Server.Port)

	logger.Info("Starting Chat service",
		zap.String("address", addr),
		zap.Bool("https", s.config.Server.HTTPS),
		zap.Bool("http3", s.config.Server.EnableHTTP3),
	)

	if s.config.Server.HTTPS {
		return s.startHTTPS(addr)
	}

	return s.startHTTP(addr)
}

// startHTTPS starts HTTPS server with optional HTTP/3 support
func (s *Server) startHTTPS(addr string) error {
	// Load TLS certificates
	cert, err := tls.LoadX509KeyPair(s.config.Server.CertFile, s.config.Server.KeyFile)
	if err != nil {
		return fmt.Errorf("failed to load TLS certificates: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"h3", "h2", "http/1.1"}, // HTTP/3, HTTP/2, HTTP/1.1
		MinVersion:   tls.VersionTLS12,
	}

	if s.config.Server.EnableHTTP3 {
		// HTTP/3 QUIC server
		s.http3Server = &http3.Server{
			Addr:      addr,
			Handler:   s.router,
			TLSConfig: tlsConfig,
			QUICConfig: &quic.Config{
				MaxIdleTimeout:  30 * time.Second,
				KeepAlivePeriod: 10 * time.Second,
			},
		}

		logger.Info("HTTP/3 QUIC server starting", zap.String("address", addr))
		return s.http3Server.ListenAndServe()
	}

	// Regular HTTPS server
	s.httpServer = &http.Server{
		Addr:           addr,
		Handler:        s.router,
		TLSConfig:      tlsConfig,
		ReadTimeout:    time.Duration(s.config.Server.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(s.config.Server.WriteTimeout) * time.Second,
		MaxHeaderBytes: s.config.Server.MaxHeaderBytes,
	}

	logger.Info("HTTPS server starting", zap.String("address", addr))
	return s.httpServer.ListenAndServeTLS(s.config.Server.CertFile, s.config.Server.KeyFile)
}

// startHTTP starts HTTP server (for development/testing)
func (s *Server) startHTTP(addr string) error {
	s.httpServer = &http.Server{
		Addr:           addr,
		Handler:        s.router,
		ReadTimeout:    time.Duration(s.config.Server.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(s.config.Server.WriteTimeout) * time.Second,
		MaxHeaderBytes: s.config.Server.MaxHeaderBytes,
	}

	logger.Info("HTTP server starting (insecure)", zap.String("address", addr))
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	logger.Info("Shutting down server...")

	if s.http3Server != nil {
		if err := s.http3Server.Close(); err != nil {
			logger.Error("Error closing HTTP/3 server", zap.Error(err))
		}
	}

	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			logger.Error("Error shutting down HTTP server", zap.Error(err))
			return err
		}
	}

	logger.Info("Server shutdown complete")
	return nil
}

// setupMiddleware configures middleware
func (s *Server) setupMiddleware() {
	// Recovery middleware (must be first)
	s.router.Use(middleware.RecoveryMiddleware())

	// Request logging
	s.router.Use(middleware.RequestLoggerMiddleware())

	// CORS
	s.router.Use(middleware.CORSMiddleware(s.config.Security.AllowedOrigins))

	// DDOS protection
	s.router.Use(middleware.DDOSProtectionMiddleware(&s.config.Security))

	// Message size limit
	s.router.Use(middleware.MessageSizeMiddleware(s.config.Security.MaxMessageSize))
}

// setupRoutes configures routes
func (s *Server) setupRoutes() {
	// Health check (no auth required)
	s.router.GET("/health", s.healthHandler)
	s.router.GET("/version", s.versionHandler)

	// API routes (require JWT auth)
	api := s.router.Group("/api")
	api.Use(middleware.JWTMiddleware(s.config.JWT.Secret))
	{
		// Unified /do endpoint will be implemented in handlers
		api.POST("/do", s.doHandler)
	}

	// WebSocket endpoint (JWT via query param or header)
	s.router.GET("/ws", s.websocketHandler)
}

// healthHandler returns health status
func (s *Server) healthHandler(c *gin.Context) {
	// Check database connection
	dbHealthy := true
	if err := s.db.Ping(); err != nil {
		dbHealthy = false
		logger.Error("Database ping failed", zap.Error(err))
	}

	status := "healthy"
	httpStatus := http.StatusOK

	if !dbHealthy {
		status = "unhealthy"
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, gin.H{
		"status":   status,
		"database": dbHealthy,
		"service":  "chat",
	})
}

// versionHandler returns service version
func (s *Server) versionHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service": "helixtrack-chat",
		"version": "1.0.0",
		"http3":   s.config.Server.EnableHTTP3,
	})
}

// doHandler handles the main API endpoint
func (s *Server) doHandler(c *gin.Context) {
	s.handler.DoAction(c)
}

// websocketHandler handles WebSocket connections
func (s *Server) websocketHandler(c *gin.Context) {
	// Get JWT token from query parameter or header
	token := c.Query("jwt")
	if token == "" {
		token = c.GetHeader("Authorization")
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}
	}

	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "JWT token required",
		})
		return
	}

	// Validate JWT and get claims
	claims, err := middleware.ValidateJWT(token, s.config.JWT.Secret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid JWT token",
		})
		return
	}

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("Failed to upgrade connection to WebSocket", zap.Error(err))
		return
	}
	defer conn.Close()

	// Handle WebSocket connection
	s.handleWebSocketConnection(conn, claims)
}

// handleWebSocketConnection handles an active WebSocket connection
func (s *Server) handleWebSocketConnection(conn *websocket.Conn, claims *models.JWTClaims) {
	userID := claims.UserID.String()

	// Register connection
	s.connMutex.Lock()
	s.connections[userID] = conn
	s.connMutex.Unlock()

	logger.Info("WebSocket connection established",
		zap.String("user_id", userID),
		zap.String("username", claims.Username),
	)

	// Handle incoming messages
	defer func() {
		s.connMutex.Lock()
		delete(s.connections, userID)
		s.connMutex.Unlock()
		conn.Close()
		logger.Info("WebSocket connection closed", zap.String("user_id", userID))
	}()

	for {
		var msg models.WSEvent
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("WebSocket error", zap.Error(err))
			}
			break
		}

		// Handle WebSocket message
		s.handleWebSocketMessage(conn, &msg, claims)
	}
}

// handleWebSocketMessage handles incoming WebSocket messages
func (s *Server) handleWebSocketMessage(conn *websocket.Conn, msg *models.WSEvent, claims *models.JWTClaims) {
	switch msg.Type {
	case "subscribe":
		// Subscribe to chat room events
		if chatRoomID, ok := msg.Data.(map[string]interface{})["chat_room_id"].(string); ok {
			logger.Info("User subscribed to chat room",
				zap.String("user_id", claims.UserID.String()),
				zap.String("chat_room_id", chatRoomID),
			)
		}
	case "unsubscribe":
		// Unsubscribe from chat room events
		if chatRoomID, ok := msg.Data.(map[string]interface{})["chat_room_id"].(string); ok {
			logger.Info("User unsubscribed from chat room",
				zap.String("user_id", claims.UserID.String()),
				zap.String("chat_room_id", chatRoomID),
			)
		}
	default:
		logger.Warn("Unknown WebSocket message type", zap.String("type", msg.Type))
	}
}

// broadcastToRoom sends a WebSocket event to all users in a chat room
func (s *Server) broadcastToRoom(chatRoomID string, event *models.WSEvent) {
	s.connMutex.RLock()
	defer s.connMutex.RUnlock()

	for userID, conn := range s.connections {
		// TODO: Check if user is in the chat room before broadcasting
		if err := conn.WriteJSON(event); err != nil {
			logger.Error("Failed to send WebSocket message",
				zap.String("user_id", userID),
				zap.Error(err),
			)
		}
	}
}

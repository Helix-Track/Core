package server

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"helixtrack.ru/core/internal/config"
	"helixtrack.ru/core/internal/database"
	"helixtrack.ru/core/internal/handlers"
	"helixtrack.ru/core/internal/logger"
	"helixtrack.ru/core/internal/middleware"
	"helixtrack.ru/core/internal/models"
	"helixtrack.ru/core/internal/services"
	"helixtrack.ru/core/internal/websocket"
)

// Server represents the HTTP server
type Server struct {
	config                  *config.Config
	router                  *gin.Engine
	httpServer              *http.Server
	db                      database.Database
	authService             services.AuthService
	permService             services.PermissionService
	serviceDiscoveryHandler *handlers.ServiceDiscoveryHandler
	networkDiscoveryService *services.NetworkDiscoveryService
	wsManager               *websocket.Manager
	wsPublisher             websocket.EventPublisher
	wsHandler               *websocket.Handler
}

// NewServer creates a new server instance
func NewServer(cfg *config.Config) (*Server, error) {
	// Initialize database
	db, err := database.NewDatabase(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize services
	authService := services.NewAuthService(
		cfg.Services.Authentication.URL,
		cfg.Services.Authentication.Timeout,
		cfg.Services.Authentication.Enabled,
	)

	permService := services.NewPermissionService(
		cfg.Services.Permissions.URL,
		cfg.Services.Permissions.Timeout,
		cfg.Services.Permissions.Enabled,
	)

	// Initialize service discovery handler
	serviceDiscoveryHandler, err := handlers.NewServiceDiscoveryHandler(db)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize service discovery handler: %w", err)
	}

	// Initialize network discovery service
	listener := cfg.GetPrimaryListener()
	host := "localhost"
	port := 8080
	if listener != nil {
		host = listener.Address
		port = listener.Port
	}
	networkDiscoveryService := services.NewNetworkDiscoveryService(port, host)

	// Initialize WebSocket manager and publisher
	var wsManager *websocket.Manager
	var wsPublisher websocket.EventPublisher
	var wsHandler *websocket.Handler

	if cfg.IsWebSocketEnabled() {
		wsConfig := websocket.ConfigToModel(cfg.GetWebSocketConfig())
		wsManager = websocket.NewManager(wsConfig, permService)
		wsPublisher = websocket.NewPublisher(wsManager, true)
		wsHandler = websocket.NewHandler(wsManager, authService, wsConfig)

		logger.Info("WebSocket enabled",
			zap.String("path", wsConfig.Path),
			zap.Int("maxClients", wsConfig.MaxClients),
		)
	} else {
		wsPublisher = websocket.NewNoOpPublisher()
		logger.Info("WebSocket disabled")
	}

	server := &Server{
		config:                  cfg,
		db:                      db,
		authService:             authService,
		permService:             permService,
		serviceDiscoveryHandler: serviceDiscoveryHandler,
		networkDiscoveryService: networkDiscoveryService,
		wsManager:               wsManager,
		wsPublisher:             wsPublisher,
		wsHandler:               wsHandler,
	}

	server.setupRouter()

	return server, nil
}

// setupRouter configures the Gin router with all routes and middleware
func (s *Server) setupRouter() {
	// Set Gin mode based on configuration
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Add middleware
	router.Use(gin.Recovery())
	router.Use(s.loggingMiddleware())
	router.Use(s.corsMiddleware())

	// Initialize users table
	if err := handlers.InitializeUserTable(s.db); err != nil {
		logger.Error("Failed to initialize users table", zap.Error(err))
	}

	// Initialize project, ticket, and comment tables
	if err := handlers.InitializeProjectTables(s.db); err != nil {
		logger.Error("Failed to initialize project tables", zap.Error(err))
	}

	// Initialize service discovery tables
	if err := handlers.InitializeServiceDiscoveryTables(s.db); err != nil {
		logger.Error("Failed to initialize service discovery tables", zap.Error(err))
	}

	// Start service health checker
	if err := s.serviceDiscoveryHandler.StartHealthChecker(); err != nil {
		logger.Error("Failed to start health checker", zap.Error(err))
	} else {
		logger.Info("Service health checker started")
	}

	// Create handlers
	handler := handlers.NewHandler(s.db, s.authService, s.permService, s.config.Version)
	handler.SetEventPublisher(s.wsPublisher) // Set event publisher for WebSocket events
	authHandler := handlers.NewAuthHandler(s.db)

	// WebSocket routes (if enabled)
	if s.config.IsWebSocketEnabled() && s.wsHandler != nil {
		wsPath := s.config.WebSocket.Path
		router.GET(wsPath, s.wsHandler.HandleConnection)
		router.GET(wsPath+"/stats", s.wsHandler.HandleStats)
		logger.Info("WebSocket routes registered",
			zap.String("path", wsPath),
			zap.String("statsPath", wsPath+"/stats"),
		)
	}

	// Authentication routes (public)
	auth := router.Group("/api/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/logout", authHandler.Logout)
	}

	// Service discovery routes (admin only)
	serviceDiscovery := router.Group("/api/services")
	{
		serviceDiscovery.POST("/register", s.serviceDiscoveryHandler.RegisterService)
		serviceDiscovery.POST("/discover", s.serviceDiscoveryHandler.DiscoverServices)
		serviceDiscovery.POST("/rotate", s.serviceDiscoveryHandler.RotateService)
		serviceDiscovery.POST("/decommission", s.serviceDiscoveryHandler.DecommissionService)
		serviceDiscovery.POST("/update", s.serviceDiscoveryHandler.UpdateService)
		serviceDiscovery.GET("/list", s.serviceDiscoveryHandler.ListServices)
		serviceDiscovery.GET("/health/:id", s.serviceDiscoveryHandler.GetServiceHealth)
	}

	// Public routes (no JWT required)
	router.POST("/do", func(c *gin.Context) {
		// Log raw request body for debugging
		bodyBytes, _ := c.GetRawData()
		logger.Info("Received /do request", zap.String("body", string(bodyBytes)))

		// Restore body for binding
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// Parse request to check if authentication is required
		var req models.Request
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Error("Failed to bind JSON request",
				zap.Error(err),
				zap.String("error_details", err.Error()),
				zap.String("body", string(bodyBytes)))
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				models.ErrorCodeInvalidRequest,
				fmt.Sprintf("Invalid request format: %v", err),
				"",
			))
			return
		}

		logger.Info("Successfully parsed request", zap.String("action", req.Action))

		// If authentication is required, validate JWT
		if req.IsAuthenticationRequired() {
			// Extract JWT from Authorization header or request body
			var jwtToken string

			// Check Authorization header first (format: "Bearer <token>")
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" {
				// Extract token from "Bearer <token>" format
				if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
					jwtToken = authHeader[7:]
				}
			}

			// Fall back to JWT field in request body
			if jwtToken == "" {
				jwtToken = req.JWT
			}

			// Check if JWT is present
			if jwtToken == "" {
				c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
					models.ErrorCodeMissingJWT,
					"JWT token is required for this action",
					"",
				))
				return
			}

			// Create JWT middleware and validate
			jwtMiddleware := middleware.NewJWTMiddleware(s.authService, "")
			claims, err := jwtMiddleware.ValidateToken(c.Request.Context(), jwtToken)
			if err != nil {
				c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
					models.ErrorCodeInvalidJWT,
					"Invalid or expired JWT token",
					"",
				))
				return
			}

			// Store claims in context
			c.Set("claims", claims)
			c.Set("username", claims.Username)
		}

		// Restore request body for handler
		c.Set("request", &req)

		// Call handler
		handler.DoAction(c)
	})

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	s.router = router
}

// loggingMiddleware logs HTTP requests
func (s *Server) loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		logger.Info("HTTP Request",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("client_ip", c.ClientIP()),
		)
	}
}

// corsMiddleware adds CORS headers
func (s *Server) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}

// Start starts the HTTP server with port fallback mechanism
func (s *Server) Start() error {
	// Start WebSocket manager if enabled
	if s.wsManager != nil {
		if err := s.wsManager.Start(); err != nil {
			return fmt.Errorf("failed to start WebSocket manager: %w", err)
		}
		logger.Info("WebSocket manager started")
	}

	// Start network discovery service
	if err := s.networkDiscoveryService.Start(); err != nil {
		return fmt.Errorf("failed to start network discovery service: %w", err)
	}
	logger.Info("Network discovery service started")

	// Try to start server with port fallback
	addr, err := s.startWithPortFallback()
	if err != nil {
		return err
	}

	// Update network discovery service with actual port
	actualPort := s.extractPortFromAddress(addr)
	s.networkDiscoveryService.UpdatePort(actualPort)

	logger.Info("HTTP server started successfully", zap.String("address", addr))

	// Broadcast availability to clients
	s.broadcastAvailability(addr)

	return nil
}

// startWithPortFallback attempts to start the server, trying multiple ports if needed
func (s *Server) startWithPortFallback() (string, error) {
	listener := s.config.GetPrimaryListener()
	if listener == nil {
		return "", fmt.Errorf("no listener configured")
	}

	basePort := listener.Port
	maxAttempts := 100 // Try up to 100 different ports

	for attempt := 0; attempt < maxAttempts; attempt++ {
		port := basePort + attempt
		if port > 65535 {
			port = 1024 + (port - 65535 - 1) // Wrap around to higher ports
		}

		addr := fmt.Sprintf("%s:%d", listener.Address, port)

		s.httpServer = &http.Server{
			Addr:         addr,
			Handler:      s.router,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
		}

		logger.Info("Attempting to start server",
			zap.String("address", addr),
			zap.Int("attempt", attempt+1),
		)

		var err error
		if listener.HTTPS {
			logger.Info("Starting HTTPS server",
				zap.String("cert", listener.CertFile),
				zap.String("key", listener.KeyFile),
			)
			err = s.httpServer.ListenAndServeTLS(listener.CertFile, listener.KeyFile)
		} else {
			err = s.httpServer.ListenAndServe()
		}

		// Check if the error is due to port already in use
		if err != nil && s.isPortInUseError(err) {
			logger.Warn("Port already in use, trying next port",
				zap.Int("port", port),
				zap.Error(err),
			)
			// Stop the current server before trying next port
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			s.httpServer.Shutdown(ctx)
			cancel()
			continue
		}

		if err != nil {
			return "", fmt.Errorf("failed to start server on %s: %w", addr, err)
		}

		// Server started successfully
		return addr, nil
	}

	return "", fmt.Errorf("failed to start server after trying %d different ports starting from %d", maxAttempts, basePort)
}

// isPortInUseError checks if the error indicates the port is already in use
func (s *Server) isPortInUseError(err error) bool {
	if err == nil {
		return false
	}

	// Check for common "address already in use" errors
	errStr := err.Error()
	return contains(errStr, "address already in use") ||
		contains(errStr, "bind: address already in use") ||
		contains(errStr, "listen tcp") && contains(errStr, "address already in use")
}

// contains checks if a string contains a substring (case-insensitive helper)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				containsInMiddle(s, substr)))
}

// containsInMiddle checks if substring exists in the middle of the string
func containsInMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// extractPortFromAddress extracts the port number from an address string
func (s *Server) extractPortFromAddress(addr string) int {
	// Simple port extraction from "host:port" format
	if len(addr) == 0 {
		return 0
	}

	// Find the last colon
	lastColon := -1
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			lastColon = i
			break
		}
	}

	if lastColon == -1 || lastColon == len(addr)-1 {
		return 0
	}

	// Parse the port number
	var port int
	for i := lastColon + 1; i < len(addr); i++ {
		if addr[i] >= '0' && addr[i] <= '9' {
			port = port*10 + int(addr[i]-'0')
		} else {
			break
		}
	}

	return port
}

// broadcastAvailability broadcasts the server availability to clients
func (s *Server) broadcastAvailability(addr string) {
	// Create availability event
	eventData := map[string]interface{}{
		"type":      "server_available",
		"address":   addr,
		"timestamp": time.Now().Unix(),
		"version":   s.config.Version,
	}

	// Create models.Event for WebSocket publishing
	event := &models.Event{
		ID:       generateEventID(),
		Type:     models.EventSystemHealthCheck,
		Action:   "server_started",
		Object:   "server",
		EntityID: addr,
		Username: "system",
		Data:     eventData,
	}

	// Broadcast via WebSocket if available
	if s.wsPublisher != nil && s.wsPublisher.IsEnabled() {
		s.wsPublisher.PublishEvent(event)
		logger.Info("Server availability broadcasted via WebSocket")
	}

	// Log the availability
	logger.Info("Server is now available",
		zap.String("address", addr),
		zap.String("version", s.config.Version),
	)
}

// generateEventID generates a unique event ID
func generateEventID() string {
	return fmt.Sprintf("evt_%d", time.Now().UnixNano())
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	logger.Info("Shutting down server...")

	// Stop WebSocket manager
	if s.wsManager != nil {
		logger.Info("Stopping WebSocket manager...")
		if err := s.wsManager.Stop(); err != nil {
			logger.Error("Error stopping WebSocket manager", zap.Error(err))
		}
	}

	// Stop network discovery service
	logger.Info("Stopping network discovery service...")
	if err := s.networkDiscoveryService.Stop(); err != nil {
		logger.Error("Error stopping network discovery service", zap.Error(err))
	}

	// Stop health checker
	if s.serviceDiscoveryHandler != nil {
		logger.Info("Stopping service health checker...")
		s.serviceDiscoveryHandler.StopHealthChecker()
	}

	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			logger.Error("Error shutting down HTTP server", zap.Error(err))
			return err
		}
	}

	if s.db != nil {
		if err := s.db.Close(); err != nil {
			logger.Error("Error closing database", zap.Error(err))
			return err
		}
	}

	logger.Info("Server shutdown complete")
	return nil
}

// GetRouter returns the Gin router (useful for testing)
func (s *Server) GetRouter() *gin.Engine {
	return s.router
}

// GetEventPublisher returns the WebSocket event publisher
func (s *Server) GetEventPublisher() websocket.EventPublisher {
	return s.wsPublisher
}

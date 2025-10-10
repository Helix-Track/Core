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
)

// Server represents the HTTP server
type Server struct {
	config                    *config.Config
	router                    *gin.Engine
	httpServer                *http.Server
	db                        database.Database
	authService               services.AuthService
	permService               services.PermissionService
	serviceDiscoveryHandler   *handlers.ServiceDiscoveryHandler
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

	server := &Server{
		config:                  cfg,
		db:                      db,
		authService:             authService,
		permService:             permService,
		serviceDiscoveryHandler: serviceDiscoveryHandler,
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
	authHandler := handlers.NewAuthHandler(s.db)

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

// Start starts the HTTP server
func (s *Server) Start() error {
	addr := s.config.GetListenerAddress()

	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Info("Starting HTTP server", zap.String("address", addr))

	listener := s.config.GetPrimaryListener()
	if listener != nil && listener.HTTPS {
		logger.Info("Starting HTTPS server",
			zap.String("cert", listener.CertFile),
			zap.String("key", listener.KeyFile),
		)
		return s.httpServer.ListenAndServeTLS(listener.CertFile, listener.KeyFile)
	}

	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	logger.Info("Shutting down server...")

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

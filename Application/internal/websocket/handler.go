package websocket

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"helixtrack.ru/core/internal/logger"
	"helixtrack.ru/core/internal/middleware"
	"helixtrack.ru/core/internal/models"
	"helixtrack.ru/core/internal/services"
)

// Handler handles WebSocket connections
type Handler struct {
	manager     *Manager
	authService services.AuthService
	config      models.WebSocketConfig
}

// NewHandler creates a new WebSocket handler
func NewHandler(manager *Manager, authService services.AuthService, config models.WebSocketConfig) *Handler {
	return &Handler{
		manager:     manager,
		authService: authService,
		config:      config,
	}
}

// HandleConnection handles WebSocket connection upgrades
func (h *Handler) HandleConnection(c *gin.Context) {
	// Check if WebSocket is enabled
	if !h.config.Enabled {
		c.JSON(http.StatusServiceUnavailable, models.NewErrorResponse(
			models.ErrorCodeServiceUnavailable,
			"WebSocket service is not enabled",
			"",
		))
		return
	}

	// Get JWT token from query parameter or header
	var jwtToken string

	// Check query parameter first (ws://host/ws?token=xxx)
	jwtToken = c.Query("token")

	// Fall back to Authorization header
	if jwtToken == "" {
		authHeader := c.GetHeader("Authorization")
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			jwtToken = authHeader[7:]
		}
	}

	// Fall back to Sec-WebSocket-Protocol header (common for WebSocket auth)
	if jwtToken == "" {
		jwtToken = c.GetHeader("Sec-WebSocket-Protocol")
	}

	// Check if authentication is required
	if h.config.RequireAuth && jwtToken == "" {
		logger.Warn("WebSocket connection attempt without JWT token",
			zap.String("remoteAddr", c.Request.RemoteAddr),
		)
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeMissingJWT,
			"JWT token is required for WebSocket connection",
			"",
		))
		return
	}

	var claims *models.JWTClaims
	var username string

	// Validate JWT token if provided
	if jwtToken != "" {
		jwtMiddleware := middleware.NewJWTMiddleware(h.authService, "")
		validatedClaims, err := jwtMiddleware.ValidateToken(c.Request.Context(), jwtToken)
		if err != nil {
			logger.Error("Invalid JWT token for WebSocket",
				zap.Error(err),
				zap.String("remoteAddr", c.Request.RemoteAddr),
			)
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
				models.ErrorCodeInvalidJWT,
				"Invalid or expired JWT token",
				"",
			))
			return
		}
		claims = validatedClaims
		username = claims.Username
	} else {
		// If auth is not required and no token provided, use anonymous connection
		username = "anonymous"
		claims = nil
	}

	// Upgrade connection to WebSocket
	upgrader := h.manager.GetUpgrader()
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("Failed to upgrade to WebSocket",
			zap.Error(err),
			zap.String("username", username),
			zap.String("remoteAddr", c.Request.RemoteAddr),
		)
		return
	}

	// Create client
	client := h.manager.CreateClient(conn, username, claims)

	// Register client with manager
	if err := h.manager.RegisterClient(client); err != nil {
		logger.Error("Failed to register WebSocket client",
			zap.Error(err),
			zap.String("clientId", client.ID),
			zap.String("username", username),
		)
		client.Close()
		return
	}

	logger.Info("WebSocket connection established",
		zap.String("clientId", client.ID),
		zap.String("username", username),
		zap.String("remoteAddr", c.Request.RemoteAddr),
	)
}

// HandleStats returns WebSocket manager statistics
func (h *Handler) HandleStats(c *gin.Context) {
	stats := h.manager.GetStats()

	c.JSON(http.StatusOK, models.NewSuccessResponse(map[string]interface{}{
		"totalConnections":  stats.TotalConnections,
		"activeConnections": stats.ActiveConnections,
		"totalEvents":       stats.TotalEvents,
		"totalErrors":       stats.TotalErrors,
		"startTime":         stats.StartTime,
		"lastEventTime":     stats.LastEventTime,
		"uptime":            stats.LastEventTime.Sub(stats.StartTime).String(),
	}))
}

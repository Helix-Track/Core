package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"helixtrack.ru/core/internal/database"
	"helixtrack.ru/core/internal/logger"
	"helixtrack.ru/core/internal/middleware"
	"helixtrack.ru/core/internal/models"
	"helixtrack.ru/core/internal/services"
)

// Handler manages all HTTP handlers
type Handler struct {
	db          database.Database
	authService services.AuthService
	permService services.PermissionService
	version     string
}

// NewHandler creates a new handler instance
func NewHandler(db database.Database, authService services.AuthService, permService services.PermissionService, version string) *Handler {
	return &Handler{
		db:          db,
		authService: authService,
		permService: permService,
		version:     version,
	}
}

// DoAction handles the unified /do endpoint with action-based routing
func (h *Handler) DoAction(c *gin.Context) {
	var req models.Request

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Invalid request format",
			"",
		))
		return
	}

	logger.Info("Processing action",
		zap.String("action", req.Action),
		zap.String("object", req.Object),
	)

	// Route to appropriate handler based on action
	switch req.Action {
	case models.ActionVersion:
		h.handleVersion(c, &req)
	case models.ActionJWTCapable:
		h.handleJWTCapable(c, &req)
	case models.ActionDBCapable:
		h.handleDBCapable(c, &req)
	case models.ActionHealth:
		h.handleHealth(c, &req)
	case models.ActionAuthenticate:
		h.handleAuthenticate(c, &req)
	case models.ActionCreate:
		h.handleCreate(c, &req)
	case models.ActionModify:
		h.handleModify(c, &req)
	case models.ActionRemove:
		h.handleRemove(c, &req)
	case models.ActionRead:
		h.handleRead(c, &req)
	case models.ActionList:
		h.handleList(c, &req)
	default:
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidAction,
			"Unknown action: "+req.Action,
			"",
		))
	}
}

// handleVersion returns the API version
func (h *Handler) handleVersion(c *gin.Context, req *models.Request) {
	response := models.NewSuccessResponse(map[string]interface{}{
		"version": h.version,
		"api":     "1.0.0",
	})
	c.JSON(http.StatusOK, response)
}

// handleJWTCapable returns whether JWT authentication is available
func (h *Handler) handleJWTCapable(c *gin.Context, req *models.Request) {
	capable := h.authService != nil && h.authService.IsEnabled()
	response := models.NewSuccessResponse(map[string]interface{}{
		"jwtCapable": capable,
		"enabled":    capable,
	})
	c.JSON(http.StatusOK, response)
}

// handleDBCapable returns whether database is available
func (h *Handler) handleDBCapable(c *gin.Context, req *models.Request) {
	capable := h.db != nil
	dbType := ""
	if h.db != nil {
		dbType = h.db.GetType()
		// Try to ping the database
		if err := h.db.Ping(c.Request.Context()); err != nil {
			logger.Error("Database ping failed", zap.Error(err))
			capable = false
		}
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"dbCapable": capable,
		"type":      dbType,
	})
	c.JSON(http.StatusOK, response)
}

// handleHealth returns the health status of the service
func (h *Handler) handleHealth(c *gin.Context, req *models.Request) {
	healthy := true
	checks := make(map[string]interface{})

	// Check database
	if h.db != nil {
		if err := h.db.Ping(c.Request.Context()); err != nil {
			checks["database"] = "unhealthy"
			healthy = false
		} else {
			checks["database"] = "healthy"
		}
	}

	// Check auth service
	if h.authService != nil && h.authService.IsEnabled() {
		checks["authService"] = "enabled"
	} else {
		checks["authService"] = "disabled"
	}

	// Check permission service
	if h.permService != nil && h.permService.IsEnabled() {
		checks["permissionService"] = "enabled"
	} else {
		checks["permissionService"] = "disabled"
	}

	status := "healthy"
	if !healthy {
		status = "unhealthy"
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"status": status,
		"checks": checks,
	})

	statusCode := http.StatusOK
	if !healthy {
		statusCode = http.StatusServiceUnavailable
		response.ErrorCode = models.ErrorCodeServiceUnavailable
		response.ErrorMessage = "Service is unhealthy"
	}

	c.JSON(statusCode, response)
}

// handleAuthenticate handles authentication requests
func (h *Handler) handleAuthenticate(c *gin.Context, req *models.Request) {
	username, ok := req.Data["username"].(string)
	if !ok || username == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing username",
			"",
		))
		return
	}

	password, ok := req.Data["password"].(string)
	if !ok || password == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing password",
			"",
		))
		return
	}

	claims, err := h.authService.Authenticate(c.Request.Context(), username, password)
	if err != nil {
		logger.Error("Authentication failed", zap.Error(err), zap.String("username", username))
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication failed",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"username": claims.Username,
		"role":     claims.Role,
		"name":     claims.Name,
	})
	c.JSON(http.StatusOK, response)
}

// handleCreate handles create operations
func (h *Handler) handleCreate(c *gin.Context, req *models.Request) {
	if req.Object == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingObject,
			"Missing object type",
			"",
		))
		return
	}

	// Get username from middleware
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Check permissions
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, req.Object, models.PermissionCreate)
	if err != nil {
		logger.Error("Permission check failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodePermissionServiceError,
			"Permission check failed",
			"",
		))
		return
	}

	if !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	// TODO: Implement actual create logic based on object type
	logger.Info("Create operation",
		zap.String("object", req.Object),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"message": "Create operation not yet implemented",
		"object":  req.Object,
	})
	c.JSON(http.StatusOK, response)
}

// handleModify handles modify operations
func (h *Handler) handleModify(c *gin.Context, req *models.Request) {
	if req.Object == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingObject,
			"Missing object type",
			"",
		))
		return
	}

	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, req.Object, models.PermissionUpdate)
	if err != nil {
		logger.Error("Permission check failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodePermissionServiceError,
			"Permission check failed",
			"",
		))
		return
	}

	if !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	logger.Info("Modify operation",
		zap.String("object", req.Object),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"message": "Modify operation not yet implemented",
		"object":  req.Object,
	})
	c.JSON(http.StatusOK, response)
}

// handleRemove handles remove operations
func (h *Handler) handleRemove(c *gin.Context, req *models.Request) {
	if req.Object == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingObject,
			"Missing object type",
			"",
		))
		return
	}

	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, req.Object, models.PermissionDelete)
	if err != nil {
		logger.Error("Permission check failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodePermissionServiceError,
			"Permission check failed",
			"",
		))
		return
	}

	if !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	logger.Info("Remove operation",
		zap.String("object", req.Object),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"message": "Remove operation not yet implemented",
		"object":  req.Object,
	})
	c.JSON(http.StatusOK, response)
}

// handleRead handles read operations
func (h *Handler) handleRead(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	logger.Info("Read operation", zap.String("username", username))

	response := models.NewSuccessResponse(map[string]interface{}{
		"message": "Read operation not yet implemented",
	})
	c.JSON(http.StatusOK, response)
}

// handleList handles list operations
func (h *Handler) handleList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	logger.Info("List operation", zap.String("username", username))

	response := models.NewSuccessResponse(map[string]interface{}{
		"message": "List operation not yet implemented",
		"items":   []interface{}{},
	})
	c.JSON(http.StatusOK, response)
}

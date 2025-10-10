package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
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
	// Get the already-parsed request from context (set by server.go)
	reqInterface, exists := c.Get("request")
	if !exists {
		logger.Error("Request not found in context")
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Invalid request format",
			"",
		))
		return
	}

	req, ok := reqInterface.(*models.Request)
	if !ok {
		logger.Error("Invalid request type in context")
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
		h.handleVersion(c, req)
	case models.ActionJWTCapable:
		h.handleJWTCapable(c, req)
	case models.ActionDBCapable:
		h.handleDBCapable(c, req)
	case models.ActionHealth:
		h.handleHealth(c, req)
	case models.ActionAuthenticate:
		h.handleAuthenticate(c, req)
	case models.ActionCreate:
		h.handleCreate(c, req)
	case models.ActionModify:
		h.handleModify(c, req)
	case models.ActionRemove:
		h.handleRemove(c, req)
	case models.ActionRead:
		h.handleRead(c, req)
	case models.ActionList:
		h.handleList(c, req)
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

	// Try external auth service first if enabled
	if h.authService != nil && h.authService.IsEnabled() {
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
		return
	}

	// Fall back to local authentication (for testing)
	// Get user from database
	query := `
		SELECT id, username, password_hash, email, name, role, created_at, updated_at
		FROM users
		WHERE username = ? AND deleted = 0
	`

	var user models.User
	var createdAt, updatedAt int64

	err := h.db.QueryRow(context.Background(), query, username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Email,
		&user.Name,
		&user.Role,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		logger.Error("User not found", zap.Error(err), zap.String("username", username))
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Invalid username or password",
			"",
		))
		return
	}

	user.CreatedAt = time.Unix(createdAt, 0)
	user.UpdatedAt = time.Unix(updatedAt, 0)

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		logger.Error("Invalid password", zap.String("username", username))
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Invalid username or password",
			"",
		))
		return
	}

	// Generate JWT token
	jwtService := services.NewJWTService("", "", 24)
	token, err := jwtService.GenerateToken(user.Username, user.Email, user.Name, user.Role)
	if err != nil {
		logger.Error("Failed to generate JWT token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to generate authentication token",
			"",
		))
		return
	}

	// Return success response with token
	response := models.NewSuccessResponse(map[string]interface{}{
		"token":    token,
		"username": user.Username,
		"email":    user.Email,
		"name":     user.Name,
		"role":     user.Role,
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
			"Insufficient permission - forbidden",
			"",
		))
		return
	}

	// Route to specific handler based on object type
	switch req.Object {
	case "project":
		h.handleCreateProject(c, req)
	case "ticket":
		h.handleCreateTicket(c, req)
	case "comment":
		h.handleCreateComment(c, req)
	default:
		logger.Info("Create operation for unsupported object",
			zap.String("object", req.Object),
			zap.String("username", username),
		)
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidObject,
			"Unsupported object type: "+req.Object,
			"",
		))
	}
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
			"Insufficient permission - forbidden",
			"",
		))
		return
	}

	// Route to specific handler based on object type
	switch req.Object {
	case "project":
		h.handleModifyProject(c, req)
	case "ticket":
		h.handleModifyTicket(c, req)
	case "comment":
		h.handleModifyComment(c, req)
	default:
		logger.Info("Modify operation for unsupported object",
			zap.String("object", req.Object),
			zap.String("username", username),
		)
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidObject,
			"Unsupported object type: "+req.Object,
			"",
		))
	}
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

	// Check permissions
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

	// If permission service is disabled, check user role from JWT claims
	if !h.permService.IsEnabled() {
		if claims, exists := middleware.GetClaims(c); exists {
			// Viewer role cannot delete
			if username == "viewer" || claims.Role == "viewer" {
				allowed = false
			}
		}
	}

	if !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permission - forbidden",
			"",
		))
		return
	}

	// Route to specific handler based on object type
	switch req.Object {
	case "project":
		h.handleRemoveProject(c, req)
	case "ticket":
		h.handleRemoveTicket(c, req)
	case "comment":
		h.handleRemoveComment(c, req)
	default:
		logger.Info("Remove operation for unsupported object",
			zap.String("object", req.Object),
			zap.String("username", username),
		)
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidObject,
			"Unsupported object type: "+req.Object,
			"",
		))
	}
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

	// Route to specific handler based on object type
	switch req.Object {
	case "project":
		h.handleReadProject(c, req)
	case "ticket":
		h.handleReadTicket(c, req)
	case "comment":
		h.handleReadComment(c, req)
	default:
		logger.Info("Read operation for unsupported object",
			zap.String("object", req.Object),
			zap.String("username", username),
		)
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidObject,
			"Unsupported object type: "+req.Object,
			"",
		))
	}
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

	// Route to specific handler based on object type
	switch req.Object {
	case "project":
		h.handleListProjects(c, req)
	case "ticket":
		h.handleListTickets(c, req)
	case "comment":
		h.handleListComments(c, req)
	default:
		logger.Info("List operation for unsupported object",
			zap.String("object", req.Object),
			zap.String("username", username),
		)
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidObject,
			"Unsupported object type: "+req.Object,
			"",
		))
	}
}

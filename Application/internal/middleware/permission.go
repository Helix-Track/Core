package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"helixtrack.ru/core/internal/models"
	"helixtrack.ru/core/internal/services"
)

// PermissionMiddleware creates a middleware for permission checking
func PermissionMiddleware(permService services.PermissionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip permission check if service is not enabled
		if !permService.IsEnabled() {
			c.Next()
			return
		}

		// Get JWT claims from context (set by JWT middleware)
		claims, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
				models.ErrorCodeUnauthorized,
				"No authentication provided",
				"",
			))
			c.Abort()
			return
		}

		jwtClaims, ok := claims.(*models.JWTClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
				models.ErrorCodeUnauthorized,
				"Invalid authentication claims",
				"",
			))
			c.Abort()
			return
		}

		// Store username in context for later use
		c.Set("username", jwtClaims.Username)
		c.Set("permissionService", permService)

		c.Next()
	}
}

// RequirePermission creates a middleware that requires a specific permission level
func RequirePermission(permService services.PermissionService, context string, level models.PermissionLevel) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip permission check if service is not enabled
		if !permService.IsEnabled() {
			c.Next()
			return
		}

		username, exists := c.Get("username")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
				models.ErrorCodeUnauthorized,
				"No username in context",
				"",
			))
			c.Abort()
			return
		}

		usernameStr, ok := username.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
				models.ErrorCodeInternalError,
				"Invalid username type",
				"",
			))
			c.Abort()
			return
		}

		// Check permission
		allowed, err := permService.CheckPermission(c.Request.Context(), usernameStr, context, level)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
				models.ErrorCodeInternalError,
				"Permission check failed: "+err.Error(),
				"",
			))
			c.Abort()
			return
		}

		if !allowed {
			c.JSON(http.StatusForbidden, models.NewErrorResponse(
				models.ErrorCodeForbidden,
				"Permission denied",
				"",
			))
			c.Abort()
			return
		}

		c.Next()
	}
}

// CheckPermissionForAction checks permission for a specific action and context
func CheckPermissionForAction(c *gin.Context, permService services.PermissionService, action, context string) bool {
	// Skip permission check if service is not enabled
	if !permService.IsEnabled() {
		return true
	}

	username, exists := c.Get("username")
	if !exists {
		return false
	}

	usernameStr, ok := username.(string)
	if !ok {
		return false
	}

	// Determine required permission level from action
	requiredLevel := models.GetRequiredPermissionLevel(action)

	// Check permission
	allowed, err := permService.CheckPermission(c.Request.Context(), usernameStr, context, requiredLevel)
	if err != nil {
		return false
	}

	return allowed
}

// GetUserPermissions retrieves all permissions for the current user
func GetUserPermissions(c *gin.Context, permService services.PermissionService) ([]models.Permission, error) {
	username, exists := c.Get("username")
	if !exists {
		return nil, nil
	}

	usernameStr, ok := username.(string)
	if !ok {
		return nil, nil
	}

	return permService.GetUserPermissions(c.Request.Context(), usernameStr)
}

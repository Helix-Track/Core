package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"helixtrack.ru/core/internal/logger"
	"helixtrack.ru/core/internal/models"
	"helixtrack.ru/core/internal/security/engine"
)

// RBACMiddleware enforces role-based access control using the Security Engine
func RBACMiddleware(securityEngine engine.Engine, resource string, action engine.Action) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, exists := GetUsername(c)
		if !exists {
			logger.Warn("RBAC middleware: no username in context")
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.NewErrorResponse(
				models.ErrorCodeUnauthorized,
				"Authentication required",
				"",
			))
			return
		}

		// Extract resource ID from request (if available)
		resourceID := extractResourceID(c, resource)

		// Build access request
		accessReq := engine.AccessRequest{
			Username:   username,
			Resource:   resource,
			ResourceID: resourceID,
			Action:     action,
			Context:    extractContext(c),
		}

		// Check access
		response, err := securityEngine.CheckAccess(c.Request.Context(), accessReq)
		if err != nil {
			logger.Error("RBAC middleware: access check failed",
				zap.Error(err),
				zap.String("username", username),
				zap.String("resource", resource),
			)
			c.AbortWithStatusJSON(http.StatusInternalServerError, models.NewErrorResponse(
				models.ErrorCodeInternalError,
				"Authorization check failed",
				"",
			))
			return
		}

		if !response.Allowed {
			logger.Warn("RBAC middleware: access denied",
				zap.String("username", username),
				zap.String("resource", resource),
				zap.String("resource_id", resourceID),
				zap.String("action", string(action)),
				zap.String("reason", response.Reason),
			)
			c.AbortWithStatusJSON(http.StatusForbidden, models.NewErrorResponse(
				models.ErrorCodeForbidden,
				response.Reason,
				"",
			))
			return
		}

		logger.Debug("RBAC middleware: access granted",
			zap.String("username", username),
			zap.String("resource", resource),
			zap.String("action", string(action)),
		)

		// Store authorization result in context for handlers to use
		c.Set("rbac_authorized", true)
		c.Set("rbac_resource", resource)
		c.Set("rbac_action", string(action))

		c.Next()
	}
}

// RequireSecurityLevel creates middleware that enforces security level checks
func RequireSecurityLevel(securityEngine engine.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, exists := GetUsername(c)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.NewErrorResponse(
				models.ErrorCodeUnauthorized,
				"Authentication required",
				"",
			))
			return
		}

		// Get entity ID from request parameters
		entityID := c.Param("id")
		if entityID == "" {
			entityID = c.Query("id")
		}

		if entityID == "" {
			// No specific entity, allow - will be checked at entity level
			c.Next()
			return
		}

		// Check security level access
		hasAccess, err := securityEngine.ValidateSecurityLevel(c.Request.Context(), username, entityID)
		if err != nil {
			logger.Error("Security level check failed",
				zap.Error(err),
				zap.String("username", username),
				zap.String("entity_id", entityID),
			)
			c.AbortWithStatusJSON(http.StatusInternalServerError, models.NewErrorResponse(
				models.ErrorCodeInternalError,
				"Security level validation failed",
				"",
			))
			return
		}

		if !hasAccess {
			logger.Warn("Security level access denied",
				zap.String("username", username),
				zap.String("entity_id", entityID),
			)
			c.AbortWithStatusJSON(http.StatusForbidden, models.NewErrorResponse(
				models.ErrorCodeForbidden,
				"Insufficient security clearance for this resource",
				"",
			))
			return
		}

		c.Set("security_level_checked", true)
		c.Next()
	}
}

// RequireProjectRole creates middleware that requires a specific project role
func RequireProjectRole(securityEngine engine.Engine, requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, exists := GetUsername(c)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.NewErrorResponse(
				models.ErrorCodeUnauthorized,
				"Authentication required",
				"",
			))
			return
		}

		// Get project ID from request
		projectID := c.Param("projectId")
		if projectID == "" {
			projectID = c.Query("projectId")
		}
		if projectID == "" {
			projectID = c.GetString("project_id")
		}

		if projectID == "" {
			logger.Warn("Project role check: no project ID in request")
			c.AbortWithStatusJSON(http.StatusBadRequest, models.NewErrorResponse(
				models.ErrorCodeMissingData,
				"Project ID required",
				"",
			))
			return
		}

		// Check if user has the required role
		hasRole, err := securityEngine.EvaluateRole(c.Request.Context(), username, projectID, requiredRole)
		if err != nil {
			logger.Error("Project role check failed",
				zap.Error(err),
				zap.String("username", username),
				zap.String("project_id", projectID),
				zap.String("required_role", requiredRole),
			)
			c.AbortWithStatusJSON(http.StatusInternalServerError, models.NewErrorResponse(
				models.ErrorCodeInternalError,
				"Role validation failed",
				"",
			))
			return
		}

		if !hasRole {
			logger.Warn("Project role check: insufficient role",
				zap.String("username", username),
				zap.String("project_id", projectID),
				zap.String("required_role", requiredRole),
			)
			c.AbortWithStatusJSON(http.StatusForbidden, models.NewErrorResponse(
				models.ErrorCodeForbidden,
				"Insufficient project role for this operation",
				"",
			))
			return
		}

		c.Set("project_role_checked", true)
		c.Set("project_role", requiredRole)
		c.Next()
	}
}

// SecurityContextMiddleware adds security context to the request
func SecurityContextMiddleware(securityEngine engine.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, exists := GetUsername(c)
		if !exists {
			// No user, skip security context
			c.Next()
			return
		}

		// Get security context for user
		secCtx, err := securityEngine.GetSecurityContext(c.Request.Context(), username)
		if err != nil {
			logger.Error("Failed to get security context",
				zap.Error(err),
				zap.String("username", username),
			)
			// Don't fail the request, just log the error
			c.Next()
			return
		}

		// Store security context in Gin context
		c.Set("security_context", secCtx)

		logger.Debug("Security context loaded",
			zap.String("username", username),
			zap.Int("role_count", len(secCtx.Roles)),
			zap.Int("team_count", len(secCtx.Teams)),
		)

		c.Next()
	}
}

// extractResourceID extracts the resource ID from the request
func extractResourceID(c *gin.Context, resource string) string {
	// Try URL parameters first
	if id := c.Param("id"); id != "" {
		return id
	}

	// Try query parameters
	if id := c.Query("id"); id != "" {
		return id
	}

	// Try resource-specific ID parameter
	idParam := resource + "Id"
	if id := c.Param(idParam); id != "" {
		return id
	}
	if id := c.Query(idParam); id != "" {
		return id
	}

	// Try to extract from JSON body (if POST/PUT)
	if c.Request.Method == "POST" || c.Request.Method == "PUT" {
		var body map[string]interface{}
		if err := c.ShouldBindJSON(&body); err == nil {
			if id, ok := body["id"].(string); ok && id != "" {
				return id
			}
			if id, ok := body[idParam].(string); ok && id != "" {
				return id
			}
		}
	}

	return ""
}

// extractContext extracts additional context from the request
func extractContext(c *gin.Context) map[string]string {
	context := make(map[string]string)

	// Extract project ID if available
	if projectID := c.Param("projectId"); projectID != "" {
		context["project_id"] = projectID
	} else if projectID := c.Query("projectId"); projectID != "" {
		context["project_id"] = projectID
	}

	// Extract team ID if available
	if teamID := c.Param("teamId"); teamID != "" {
		context["team_id"] = teamID
	} else if teamID := c.Query("teamId"); teamID != "" {
		context["team_id"] = teamID
	}

	// Add IP address and user agent
	context["ip_address"] = c.ClientIP()
	context["user_agent"] = c.Request.UserAgent()
	context["request_path"] = c.Request.URL.Path
	context["request_method"] = c.Request.Method

	return context
}

// GetSecurityContext retrieves the security context from Gin context
func GetSecurityContext(c *gin.Context) (*engine.SecurityContext, bool) {
	secCtxInterface, exists := c.Get("security_context")
	if !exists {
		return nil, false
	}

	secCtx, ok := secCtxInterface.(*engine.SecurityContext)
	return secCtx, ok
}

// IsAuthorized checks if the current request has been authorized
func IsAuthorized(c *gin.Context) bool {
	authorized, exists := c.Get("rbac_authorized")
	if !exists {
		return false
	}

	return authorized.(bool)
}

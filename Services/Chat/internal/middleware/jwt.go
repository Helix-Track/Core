package middleware

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"

	"helixtrack.ru/chat/internal/logger"
	"helixtrack.ru/chat/internal/models"
)

// JWTMiddleware validates JWT tokens
func JWTMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token
		token := extractToken(c)
		if token == "" {
			logger.Warn("Missing JWT token",
				zap.String("path", c.Request.URL.Path),
				zap.String("ip", c.ClientIP()),
			)
			c.JSON(401, models.ErrorResponse(models.ErrorCodeUnauthorized, "Missing authentication token"))
			c.Abort()
			return
		}

		// Parse and validate token
		claims, err := validateJWT(token, jwtSecret)
		if err != nil {
			logger.Warn("Invalid JWT token",
				zap.Error(err),
				zap.String("ip", c.ClientIP()),
			)
			c.JSON(401, models.ErrorResponse(models.ErrorCodeUnauthorized, "Invalid or expired token"))
			c.Abort()
			return
		}

		// Store claims in context
		c.Set("claims", claims)
		c.Set("user_id", claims.UserID.String())
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		logger.Debug("JWT validated",
			zap.String("username", claims.Username),
			zap.String("user_id", claims.UserID.String()),
			zap.String("role", claims.Role),
		)

		c.Next()
	}
}

// extractToken extracts JWT token from request
func extractToken(c *gin.Context) string {
	// Check Authorization header (Bearer token)
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return parts[1]
		}
	}

	// Check query parameter
	token := c.Query("token")
	if token != "" {
		return token
	}

	// Check form parameter
	token = c.PostForm("token")
	if token != "" {
		return token
	}

	return ""
}

// ValidateJWT validates JWT token and returns claims
func ValidateJWT(tokenString, jwtSecret string) (*models.JWTClaims, error) {
	return validateJWT(tokenString, jwtSecret)
}

// validateJWT validates JWT token and returns claims
func validateJWT(tokenString, jwtSecret string) (*models.JWTClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims format")
	}

	// Parse claims into our structure
	jwtClaims := &models.JWTClaims{}

	if sub, ok := claims["sub"].(string); ok {
		jwtClaims.Sub = sub
	}

	if name, ok := claims["name"].(string); ok {
		jwtClaims.Name = name
	}

	if username, ok := claims["username"].(string); ok {
		jwtClaims.Username = username
	}

	if userID, ok := claims["user_id"].(string); ok {
		// Parse UUID
		if err := jwtClaims.UserID.UnmarshalText([]byte(userID)); err != nil {
			return nil, fmt.Errorf("invalid user_id format: %w", err)
		}
	}

	if role, ok := claims["role"].(string); ok {
		jwtClaims.Role = role
	}

	if permissions, ok := claims["permissions"].(string); ok {
		jwtClaims.Permissions = permissions
	}

	if coreAddress, ok := claims["htCoreAddress"].(string); ok {
		jwtClaims.CoreAddress = coreAddress
	}

	return jwtClaims, nil
}

// GetClaims retrieves JWT claims from context
func GetClaims(c *gin.Context) (*models.JWTClaims, error) {
	claims, exists := c.Get("claims")
	if !exists {
		return nil, fmt.Errorf("claims not found in context")
	}

	jwtClaims, ok := claims.(*models.JWTClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims type")
	}

	return jwtClaims, nil
}

// GetUserID retrieves user ID from context
func GetUserID(c *gin.Context) (string, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", fmt.Errorf("user_id not found in context")
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return "", fmt.Errorf("invalid user_id type")
	}

	return userIDStr, nil
}

// GetUsername retrieves username from context
func GetUsername(c *gin.Context) (string, error) {
	username, exists := c.Get("username")
	if !exists {
		return "", fmt.Errorf("username not found in context")
	}

	usernameStr, ok := username.(string)
	if !ok {
		return "", fmt.Errorf("invalid username type")
	}

	return usernameStr, nil
}

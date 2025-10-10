package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"helixtrack.ru/core/internal/models"
	"helixtrack.ru/core/internal/services"
)

// JWTMiddleware creates a middleware for JWT validation
type JWTMiddleware struct {
	authService services.AuthService
	secretKey   string // For local JWT validation (optional)
}

// NewJWTMiddleware creates a new JWT middleware
func NewJWTMiddleware(authService services.AuthService, secretKey string) *JWTMiddleware {
	return &JWTMiddleware{
		authService: authService,
		secretKey:   secretKey,
	}
}

// Validate is a Gin middleware that validates JWT tokens
func (m *JWTMiddleware) Validate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.NewErrorResponse(
				models.ErrorCodeMissingJWT,
				"Missing Authorization header",
				"",
			))
			return
		}

		// Extract token from "Bearer <token>" format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.NewErrorResponse(
				models.ErrorCodeInvalidJWT,
				"Invalid Authorization header format",
				"",
			))
			return
		}

		token := parts[1]

		// Validate token
		claims, err := m.validateToken(c.Request.Context(), token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.NewErrorResponse(
				models.ErrorCodeInvalidJWT,
				err.Error(),
				"",
			))
			return
		}

		// Store claims in context for handlers to use
		c.Set("claims", claims)
		c.Set("username", claims.Username)

		c.Next()
	}
}

// ValidateToken is a public method to validate JWT tokens
func (m *JWTMiddleware) ValidateToken(ctx context.Context, tokenString string) (*models.JWTClaims, error) {
	return m.validateToken(ctx, tokenString)
}

// validateToken validates the JWT token using either the auth service or local validation
func (m *JWTMiddleware) validateToken(ctx context.Context, tokenString string) (*models.JWTClaims, error) {
	// If auth service is enabled, use it for validation
	if m.authService != nil && m.authService.IsEnabled() {
		return m.authService.ValidateToken(ctx, tokenString)
	}

	// Otherwise, validate locally (always available with default or provided secret key)
	return m.validateTokenLocally(tokenString)
}

// validateTokenLocally validates the JWT token locally using the secret key
func (m *JWTMiddleware) validateTokenLocally(tokenString string) (*models.JWTClaims, error) {
	// Use default secret key if not provided
	secretKey := m.secretKey
	if secretKey == "" {
		secretKey = "helix-track-default-secret-key-change-in-production"
	}

	token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*models.JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrTokenInvalidClaims
}

// GetClaims retrieves JWT claims from the Gin context
func GetClaims(c *gin.Context) (*models.JWTClaims, bool) {
	claims, exists := c.Get("claims")
	if !exists {
		return nil, false
	}

	jwtClaims, ok := claims.(*models.JWTClaims)
	return jwtClaims, ok
}

// GetUsername retrieves the username from the Gin context
func GetUsername(c *gin.Context) (string, bool) {
	username, exists := c.Get("username")
	if !exists {
		return "", false
	}

	usernameStr, ok := username.(string)
	return usernameStr, ok
}

package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/helixtrack/localization-service/internal/models"
	"go.uber.org/zap"
)

const (
	ClaimsKey = "claims"
)

// JWTAuth validates JWT tokens
func JWTAuth(jwtSecret string, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse(
				models.ErrCodeUnauthorized,
				"missing authorization header",
			))
			c.Abort()
			return
		}

		// Extract token (Bearer <token>)
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse(
				models.ErrCodeUnauthorized,
				"invalid authorization header format",
			))
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, models.ErrInvalidToken
			}
			return []byte(jwtSecret), nil
		})

		if err != nil {
			logger.Warn("JWT validation failed",
				zap.Error(err),
				zap.String("ip", c.ClientIP()),
			)

			c.JSON(http.StatusUnauthorized, models.ErrorResponse(
				models.ErrCodeInvalidToken,
				"invalid or expired token",
			))
			c.Abort()
			return
		}

		// Extract claims
		claims, ok := token.Claims.(*models.JWTClaims)
		if !ok || !token.Valid {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse(
				models.ErrCodeInvalidToken,
				"invalid token claims",
			))
			c.Abort()
			return
		}

		// Store claims in context
		c.Set(ClaimsKey, claims)

		logger.Debug("JWT validated",
			zap.String("username", claims.Username),
			zap.String("role", claims.Role),
		)

		c.Next()
	}
}

// AdminOnly restricts access to admin users
func AdminOnly(adminRoles []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get claims from context
		claimsValue, exists := c.Get(ClaimsKey)
		if !exists {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse(
				models.ErrCodeUnauthorized,
				"authentication required",
			))
			c.Abort()
			return
		}

		claims, ok := claimsValue.(*models.JWTClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse(
				models.ErrCodeUnauthorized,
				"invalid authentication claims",
			))
			c.Abort()
			return
		}

		// Check if user has admin role
		if !claims.IsAdmin(adminRoles) {
			c.JSON(http.StatusForbidden, models.ErrorResponse(
				models.ErrCodeForbidden,
				"admin access required",
			))
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetClaims retrieves JWT claims from context
func GetClaims(c *gin.Context) *models.JWTClaims {
	claimsValue, exists := c.Get(ClaimsKey)
	if !exists {
		return nil
	}

	claims, ok := claimsValue.(*models.JWTClaims)
	if !ok {
		return nil
	}

	return claims
}

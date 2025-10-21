package models

import (
	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims represents the JWT token claims structure
type JWTClaims struct {
	jwt.RegisteredClaims
	Name       string `json:"name"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	Role       string `json:"role"`
	Permissions string `json:"permissions"`
}

// IsAdmin checks if the user has admin role
func (c *JWTClaims) IsAdmin(adminRoles []string) bool {
	for _, role := range adminRoles {
		if c.Role == role {
			return true
		}
	}
	return false
}

// HasPermission checks if the user has a specific permission
func (c *JWTClaims) HasPermission(permission string) bool {
	// Simple permission check - can be extended
	return c.Permissions == permission || c.Permissions == "ALL"
}

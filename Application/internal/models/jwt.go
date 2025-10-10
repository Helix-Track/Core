package models

import "github.com/golang-jwt/jwt/v5"

// JWTClaims represents the JWT token claims structure
type JWTClaims struct {
	jwt.RegisteredClaims
	Name           string `json:"name"`
	Username       string `json:"username"`
	Role           string `json:"role"`
	Permissions    string `json:"permissions"`
	HTCoreAddress  string `json:"htCoreAddress"`
}

// PermissionLevel represents the access level for permissions
type PermissionLevel int

const (
	PermissionRead   PermissionLevel = 1
	PermissionCreate PermissionLevel = 2
	PermissionUpdate PermissionLevel = 3
	PermissionDelete PermissionLevel = 5 // Also represents ALL permissions
)

// Permission represents a permission context and level
type Permission struct {
	Context string          // Hierarchical context (node → account → organization → team/project)
	Level   PermissionLevel // Access level
}

// HasPermission checks if a permission level is sufficient
func (p PermissionLevel) HasPermission(required PermissionLevel) bool {
	return p >= required
}

package models

import (
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims represents the JWT token claims structure
type JWTClaims struct {
	jwt.RegisteredClaims
	Name          string `json:"name"`
	Username      string `json:"username"`
	Email         string `json:"email"`
	Role          string `json:"role"`
	Permissions   string `json:"permissions"`
	HTCoreAddress string `json:"htCoreAddress"`
}

// PermissionLevel represents the access level for permissions
type PermissionLevel int

const (
	PermissionNone   PermissionLevel = 0
	PermissionRead   PermissionLevel = 1
	PermissionCreate PermissionLevel = 2
	PermissionUpdate PermissionLevel = 3
	PermissionDelete PermissionLevel = 5 // Also represents ALL permissions
)

// Permission represents a permission context and level
type Permission struct {
	ID          string          `json:"id" db:"id"`
	Title       string          `json:"title" db:"title"`
	Description string          `json:"description,omitempty" db:"description"`
	Context     string          `json:"context" db:"context"` // Hierarchical context (node → account → organization → team/project)
	Level       PermissionLevel `json:"level" db:"level"`     // Access level
	Created     int64           `json:"created" db:"created"`
	Modified    int64           `json:"modified" db:"modified"`
	Deleted     bool            `json:"deleted" db:"deleted"`
}

// PermissionContext represents a hierarchical permission context
type PermissionContext struct {
	Type       string // node, account, organization, team, project, ticket
	Identifier string // UUID of the entity
	Parent     *PermissionContext
}

// PermissionCheck represents a permission check request
type PermissionCheck struct {
	Username       string
	Context        string
	RequiredLevel  PermissionLevel
	EntityType     string // ticket, project, board, etc.
	EntityID       string
	Action         string // create, read, update, delete
}

// HasPermission checks if a permission level is sufficient
func (p PermissionLevel) HasPermission(required PermissionLevel) bool {
	return p >= required
}

// String returns the string representation of permission level
func (p PermissionLevel) String() string {
	switch p {
	case PermissionNone:
		return "NONE"
	case PermissionRead:
		return "READ"
	case PermissionCreate:
		return "CREATE"
	case PermissionUpdate:
		return "UPDATE"
	case PermissionDelete:
		return "DELETE"
	default:
		return "UNKNOWN"
	}
}

// ParsePermissionLevel parses a string to PermissionLevel
func ParsePermissionLevel(level string) PermissionLevel {
	switch strings.ToUpper(level) {
	case "READ":
		return PermissionRead
	case "CREATE":
		return PermissionCreate
	case "UPDATE":
		return PermissionUpdate
	case "DELETE", "ALL":
		return PermissionDelete
	default:
		return PermissionNone
	}
}

// BuildContextPath builds a hierarchical context path
func BuildContextPath(contexts ...string) string {
	return strings.Join(contexts, "→")
}

// ParseContextPath parses a hierarchical context path
func ParseContextPath(path string) []string {
	return strings.Split(path, "→")
}

// IsParentContext checks if parent is a parent of child in hierarchy
func IsParentContext(parent, child string) bool {
	parentParts := ParseContextPath(parent)
	childParts := ParseContextPath(child)

	if len(parentParts) >= len(childParts) {
		return false
	}

	for i, part := range parentParts {
		if childParts[i] != part {
			return false
		}
	}

	return true
}

// GetRequiredPermissionLevel returns the required permission level for an action
func GetRequiredPermissionLevel(action string) PermissionLevel {
	// Extract base action from compound actions like "priorityCreate"
	actionLower := strings.ToLower(action)

	if strings.Contains(actionLower, "create") {
		return PermissionCreate
	}
	if strings.Contains(actionLower, "modify") || strings.Contains(actionLower, "update") || strings.Contains(actionLower, "edit") {
		return PermissionUpdate
	}
	if strings.Contains(actionLower, "remove") || strings.Contains(actionLower, "delete") {
		return PermissionDelete
	}
	// read, list, get, etc.
	return PermissionRead
}

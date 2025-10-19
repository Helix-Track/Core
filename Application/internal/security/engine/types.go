package engine

import (
	"time"
)

// Action represents an action that can be performed on a resource
type Action string

const (
	ActionCreate Action = "CREATE"
	ActionRead   Action = "READ"
	ActionUpdate Action = "UPDATE"
	ActionDelete Action = "DELETE"
	ActionList   Action = "LIST"
	ActionExecute Action = "EXECUTE"
)

// AccessRequest represents a request to check access permissions
type AccessRequest struct {
	Username   string            // User making the request
	Resource   string            // Resource type (e.g., "ticket", "project")
	ResourceID string            // Specific resource ID (optional for list operations)
	Action     Action            // Action being attempted
	Context    map[string]string // Additional context (e.g., project_id, team_id)
}

// AccessResponse represents the result of an access check
type AccessResponse struct {
	Allowed bool   // Whether access is allowed
	Reason  string // Reason for denial (if Allowed is false)
	AuditID string // ID of the audit log entry created
}

// PermissionSet represents a set of permissions for a resource
type PermissionSet struct {
	CanCreate bool
	CanRead   bool
	CanUpdate bool
	CanDelete bool
	CanList   bool
	Level     int    // Highest security level accessible
	Roles     []Role // Effective roles
}

// Role represents a role assigned to a user
type Role struct {
	ID          string
	Title       string
	ProjectID   *string // nil for global roles
	Permissions PermissionSet
}

// SecurityContext represents the security context for a user
type SecurityContext struct {
	Username          string
	Roles             []Role
	Teams             []string
	EffectivePermissions map[string]PermissionSet
	CachedAt          time.Time
	ExpiresAt         time.Time
}

// CacheEntry represents a cached permission check result
type CacheEntry struct {
	Request   AccessRequest
	Response  AccessResponse
	CachedAt  time.Time
	ExpiresAt time.Time
}

// AuditEntry represents a security audit log entry
type AuditEntry struct {
	ID           string
	Timestamp    time.Time
	Username     string
	Resource     string
	ResourceID   string
	Action       Action
	Allowed      bool
	Reason       string
	IPAddress    string
	UserAgent    string
	Context      map[string]string
}

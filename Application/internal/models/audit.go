package models

// Audit represents an audit log entry tracking actions in the system
type Audit struct {
	ID         string `json:"id" db:"id"`
	Action     string `json:"action" db:"action" binding:"required"`       // The action performed
	UserID     string `json:"userId" db:"user_id"`                         // User who performed the action
	EntityID   string `json:"entityId" db:"entity_id"`                     // ID of the entity affected
	EntityType string `json:"entityType" db:"entity_type"`                 // Type of entity (project, ticket, etc.)
	Details    string `json:"details,omitempty" db:"details"`              // JSON encoded details
	Created    int64  `json:"created" db:"created"`
	Modified   int64  `json:"modified" db:"modified"`
	Deleted    bool   `json:"deleted" db:"deleted"`
}

// AuditMetaData represents additional metadata for audit entries
type AuditMetaData struct {
	ID       string `json:"id" db:"id"`
	AuditID  string `json:"auditId" db:"audit_id" binding:"required"`
	Property string `json:"property" db:"property" binding:"required"`
	Value    string `json:"value" db:"value"`
	Created  int64  `json:"created" db:"created"`
	Modified int64  `json:"modified" db:"modified"`
	Deleted  bool   `json:"deleted" db:"deleted"`
}

// Valid audit actions
var validAuditActions = map[string]bool{
	// Standard CRUD actions
	ActionCreate:        true,
	ActionRead:          true,
	ActionModify:        true,
	ActionRemove:        true,
	ActionList:          true,
	ActionAuthenticate:  true,

	// Entity-specific actions
	"login":            true,
	"logout":           true,
	"permission_change": true,
	"status_change":    true,
	"assignment_change": true,
	"comment_add":      true,
	"attachment_add":   true,
	"workflow_transition": true,
}

// IsValidAction validates if the action is a recognized audit action
func (a *Audit) IsValidAction() bool {
	return validAuditActions[a.Action]
}

// HasEntity checks if the audit entry has entity information
func (a *Audit) HasEntity() bool {
	return a.EntityID != "" && a.EntityType != ""
}

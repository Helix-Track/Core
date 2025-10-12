package models

// SecurityLevel represents an enterprise security level for controlling ticket access
type SecurityLevel struct {
	ID          string `json:"id" db:"id"`
	Title       string `json:"title" db:"title" binding:"required"`
	Description string `json:"description,omitempty" db:"description"`
	ProjectID   string `json:"projectId" db:"project_id" binding:"required"`
	Level       int    `json:"level" db:"level" binding:"required"` // Numeric level for hierarchy
	Created     int64  `json:"created" db:"created"`
	Modified    int64  `json:"modified" db:"modified"`
	Deleted     bool   `json:"deleted" db:"deleted"`
}

// SecurityLevelPermissionMapping represents access grants for a security level
type SecurityLevelPermissionMapping struct {
	ID              string  `json:"id" db:"id"`
	SecurityLevelID string  `json:"securityLevelId" db:"security_level_id" binding:"required"`
	UserID          *string `json:"userId,omitempty" db:"user_id"`           // NULL if not user-specific
	TeamID          *string `json:"teamId,omitempty" db:"team_id"`           // NULL if not team-specific
	ProjectRoleID   *string `json:"projectRoleId,omitempty" db:"project_role_id"` // NULL if not role-specific
	Created         int64   `json:"created" db:"created"`
	Deleted         bool    `json:"deleted" db:"deleted"`
}

// Security level constants
const (
	SecurityLevelNone      = 0
	SecurityLevelPublic    = 1
	SecurityLevelInternal  = 2
	SecurityLevelConfidential = 3
	SecurityLevelRestricted   = 4
	SecurityLevelSecret    = 5
)

// IsValidLevel checks if the security level value is valid
func (sl *SecurityLevel) IsValidLevel() bool {
	return sl.Level >= SecurityLevelNone && sl.Level <= SecurityLevelSecret
}

// GetRecipientType returns the type of recipient (user, team, or role)
func (slp *SecurityLevelPermissionMapping) GetRecipientType() string {
	if slp.UserID != nil && *slp.UserID != "" {
		return "user"
	}
	if slp.TeamID != nil && *slp.TeamID != "" {
		return "team"
	}
	if slp.ProjectRoleID != nil && *slp.ProjectRoleID != "" {
		return "role"
	}
	return "unknown"
}

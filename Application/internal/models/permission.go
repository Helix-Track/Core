package models

// PermissionEntity represents a permission type in the system (database model)
type PermissionEntity struct {
	ID          string `json:"id" db:"id"`
	Title       string `json:"title" db:"title" binding:"required"`
	Description string `json:"description,omitempty" db:"description"`
	Value       int    `json:"value" db:"value" binding:"required"` // 1=READ, 2=CREATE, 3=UPDATE, 5=DELETE
	Created     int64  `json:"created" db:"created"`
	Modified    int64  `json:"modified" db:"modified"`
	Deleted     bool   `json:"deleted" db:"deleted"`
}

// PermissionContextEntity represents a context in which permissions can be applied (database model)
// Contexts form a hierarchy: node → account → organization → team/project
type PermissionContextEntity struct {
	ID       string `json:"id" db:"id"`
	Context  string `json:"context" db:"context" binding:"required"` // node, account, organization, team, project
	Created  int64  `json:"created" db:"created"`
	Modified int64  `json:"modified" db:"modified"`
	Deleted  bool   `json:"deleted" db:"deleted"`
}

// PermissionUserMapping maps permissions to users within a specific context
type PermissionUserMapping struct {
	ID                  string `json:"id" db:"id"`
	PermissionID        string `json:"permissionId" db:"permission_id" binding:"required"`
	UserID              string `json:"userId" db:"user_id" binding:"required"`
	PermissionContextID string `json:"permissionContextId" db:"permission_context_id" binding:"required"`
	Created             int64  `json:"created" db:"created"`
	Deleted             bool   `json:"deleted" db:"deleted"`
}

// PermissionTeamMapping maps permissions to teams within a specific context
type PermissionTeamMapping struct {
	ID                  string `json:"id" db:"id"`
	PermissionID        string `json:"permissionId" db:"permission_id" binding:"required"`
	TeamID              string `json:"teamId" db:"team_id" binding:"required"`
	PermissionContextID string `json:"permissionContextId" db:"permission_context_id" binding:"required"`
	Created             int64  `json:"created" db:"created"`
	Deleted             bool   `json:"deleted" db:"deleted"`
}

// IsValidPermissionValue validates if the permission value is one of the allowed values
func (p *PermissionEntity) IsValidPermissionValue() bool {
	return p.Value == int(PermissionRead) || p.Value == int(PermissionCreate) ||
		p.Value == int(PermissionUpdate) || p.Value == int(PermissionDelete)
}

// IsValidContext validates if the context is one of the allowed contexts
func (pc *PermissionContextEntity) IsValidContext() bool {
	validContexts := []string{"node", "account", "organization", "team", "project"}
	for _, ctx := range validContexts {
		if pc.Context == ctx {
			return true
		}
	}
	return false
}

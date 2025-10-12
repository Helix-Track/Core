package models

// ProjectRole represents a role that can be assigned to users in a project
type ProjectRole struct {
	ID          string  `json:"id" db:"id"`
	Title       string  `json:"title" db:"title" binding:"required"`
	Description string  `json:"description,omitempty" db:"description"`
	ProjectID   *string `json:"projectId,omitempty" db:"project_id"` // NULL for global roles
	Created     int64   `json:"created" db:"created"`
	Modified    int64   `json:"modified" db:"modified"`
	Deleted     bool    `json:"deleted" db:"deleted"`
}

// ProjectRoleUserMapping represents the assignment of a role to a user in a project
type ProjectRoleUserMapping struct {
	ID            string `json:"id" db:"id"`
	ProjectRoleID string `json:"projectRoleId" db:"project_role_id" binding:"required"`
	ProjectID     string `json:"projectId" db:"project_id" binding:"required"`
	UserID        string `json:"userId" db:"user_id" binding:"required"`
	Created       int64  `json:"created" db:"created"`
	Deleted       bool   `json:"deleted" db:"deleted"`
}

// IsGlobal checks if the role is global (not project-specific)
func (pr *ProjectRole) IsGlobal() bool {
	return pr.ProjectID == nil || *pr.ProjectID == ""
}

// IsProjectSpecific checks if the role is project-specific
func (pr *ProjectRole) IsProjectSpecific() bool {
	return !pr.IsGlobal()
}

// Common project role titles (constants for default roles)
const (
	ProjectRoleAdmin      = "Project Administrator"
	ProjectRoleLead       = "Project Lead"
	ProjectRoleDeveloper  = "Developer"
	ProjectRoleTester     = "Tester"
	ProjectRoleViewer     = "Viewer"
	ProjectRoleContributor = "Contributor"
)

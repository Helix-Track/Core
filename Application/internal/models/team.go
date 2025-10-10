package models

type Team struct {
	ID          string `json:"id" db:"id"`
	Title       string `json:"title" db:"title" binding:"required"`
	Description string `json:"description,omitempty" db:"description"`
	Created     int64  `json:"created" db:"created"`
	Modified    int64  `json:"modified" db:"modified"`
	Deleted     bool   `json:"deleted" db:"deleted"`
}

type TeamOrganizationMapping struct {
	ID             string `json:"id" db:"id"`
	TeamID         string `json:"teamId" db:"team_id" binding:"required"`
	OrganizationID string `json:"organizationId" db:"organization_id" binding:"required"`
	Created        int64  `json:"created" db:"created"`
	Modified       int64  `json:"modified" db:"modified"`
	Deleted        bool   `json:"deleted" db:"deleted"`
}

type TeamProjectMapping struct {
	ID        string `json:"id" db:"id"`
	TeamID    string `json:"teamId" db:"team_id" binding:"required"`
	ProjectID string `json:"projectId" db:"project_id" binding:"required"`
	Created   int64  `json:"created" db:"created"`
	Modified  int64  `json:"modified" db:"modified"`
	Deleted   bool   `json:"deleted" db:"deleted"`
}

type UserOrganizationMapping struct {
	ID             string `json:"id" db:"id"`
	UserID         string `json:"userId" db:"user_id" binding:"required"`
	OrganizationID string `json:"organizationId" db:"organization_id" binding:"required"`
	Created        int64  `json:"created" db:"created"`
	Modified       int64  `json:"modified" db:"modified"`
	Deleted        bool   `json:"deleted" db:"deleted"`
}

type UserTeamMapping struct {
	ID       string `json:"id" db:"id"`
	UserID   string `json:"userId" db:"user_id" binding:"required"`
	TeamID   string `json:"teamId" db:"team_id" binding:"required"`
	Created  int64  `json:"created" db:"created"`
	Modified int64  `json:"modified" db:"modified"`
	Deleted  bool   `json:"deleted" db:"deleted"`
}

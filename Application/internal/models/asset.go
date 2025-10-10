package models

// Asset represents a file attachment or asset (images, documents, files)
type Asset struct {
	ID          string `json:"id" db:"id"`
	URL         string `json:"url" db:"url" binding:"required"`
	Description string `json:"description,omitempty" db:"description"`
	Created     int64  `json:"created" db:"created"`
	Modified    int64  `json:"modified" db:"modified"`
	Deleted     bool   `json:"deleted" db:"deleted"`
}

// AssetTicketMapping represents the many-to-many relationship between assets and tickets
type AssetTicketMapping struct {
	ID       string `json:"id" db:"id"`
	AssetID  string `json:"assetId" db:"asset_id" binding:"required"`
	TicketID string `json:"ticketId" db:"ticket_id" binding:"required"`
	Created  int64  `json:"created" db:"created"`
	Modified int64  `json:"modified" db:"modified"`
	Deleted  bool   `json:"deleted" db:"deleted"`
}

// AssetCommentMapping represents the many-to-many relationship between assets and comments
type AssetCommentMapping struct {
	ID        string `json:"id" db:"id"`
	AssetID   string `json:"assetId" db:"asset_id" binding:"required"`
	CommentID string `json:"commentId" db:"comment_id" binding:"required"`
	Created   int64  `json:"created" db:"created"`
	Modified  int64  `json:"modified" db:"modified"`
	Deleted   bool   `json:"deleted" db:"deleted"`
}

// AssetProjectMapping represents the many-to-many relationship between assets and projects
type AssetProjectMapping struct {
	ID        string `json:"id" db:"id"`
	AssetID   string `json:"assetId" db:"asset_id" binding:"required"`
	ProjectID string `json:"projectId" db:"project_id" binding:"required"`
	Created   int64  `json:"created" db:"created"`
	Modified  int64  `json:"modified" db:"modified"`
	Deleted   bool   `json:"deleted" db:"deleted"`
}

// AssetTeamMapping represents the many-to-many relationship between assets and teams
type AssetTeamMapping struct {
	ID       string `json:"id" db:"id"`
	AssetID  string `json:"assetId" db:"asset_id" binding:"required"`
	TeamID   string `json:"teamId" db:"team_id" binding:"required"`
	Created  int64  `json:"created" db:"created"`
	Modified int64  `json:"modified" db:"modified"`
	Deleted  bool   `json:"deleted" db:"deleted"`
}

// GetDisplayName returns a user-friendly display name
func (a *Asset) GetDisplayName() string {
	if a.Description != "" {
		return a.Description
	}
	if a.URL != "" {
		return a.URL
	}
	return "Unknown Asset"
}

// IsValid checks if the asset has required fields
func (a *Asset) IsValid() bool {
	return a.URL != ""
}

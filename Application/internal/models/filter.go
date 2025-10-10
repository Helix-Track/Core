package models

// Filter represents a saved search filter
type Filter struct {
	ID          string `json:"id" db:"id"`
	Title       string `json:"title" db:"title" binding:"required"`
	Description string `json:"description,omitempty" db:"description"`
	OwnerID     string `json:"ownerId" db:"owner_id" binding:"required"`
	Query       string `json:"query" db:"query" binding:"required"` // JSON query structure
	IsPublic    bool   `json:"isPublic" db:"is_public"`
	IsFavorite  bool   `json:"isFavorite" db:"is_favorite"`
	Created     int64  `json:"created" db:"created"`
	Modified    int64  `json:"modified" db:"modified"`
	Deleted     bool   `json:"deleted" db:"deleted"`
}

// FilterShareMapping represents filter sharing with users/teams/projects
type FilterShareMapping struct {
	ID        string  `json:"id" db:"id"`
	FilterID  string  `json:"filterId" db:"filter_id" binding:"required"`
	UserID    *string `json:"userId,omitempty" db:"user_id"`    // Pointer for nullable
	TeamID    *string `json:"teamId,omitempty" db:"team_id"`    // Pointer for nullable
	ProjectID *string `json:"projectId,omitempty" db:"project_id"` // Pointer for nullable
	Created   int64   `json:"created" db:"created"`
	Deleted   bool    `json:"deleted" db:"deleted"`
}

// ShareType represents the type of share
type ShareType string

const (
	ShareTypeUser    ShareType = "user"
	ShareTypeTeam    ShareType = "team"
	ShareTypeProject ShareType = "project"
	ShareTypePublic  ShareType = "public"
)

// GetShareType determines the share type based on which ID is set
func (f *FilterShareMapping) GetShareType() ShareType {
	if f.UserID != nil {
		return ShareTypeUser
	}
	if f.TeamID != nil {
		return ShareTypeTeam
	}
	if f.ProjectID != nil {
		return ShareTypeProject
	}
	return ShareTypePublic
}

// IsSharedWith checks if the filter is shared with a specific entity
func (f *Filter) IsSharedWith(userID, teamID, projectID string, shares []FilterShareMapping) bool {
	if f.IsPublic {
		return true
	}

	for _, share := range shares {
		if share.Deleted {
			continue
		}
		if share.UserID != nil && *share.UserID == userID {
			return true
		}
		if share.TeamID != nil && *share.TeamID == teamID {
			return true
		}
		if share.ProjectID != nil && *share.ProjectID == projectID {
			return true
		}
	}

	return false
}

package models

// Label represents a categorization tag that can be applied to various entities
type Label struct {
	ID          string `json:"id" db:"id"`
	Title       string `json:"title" db:"title" binding:"required"`
	Description string `json:"description,omitempty" db:"description"`
	Color       string `json:"color,omitempty" db:"color"` // Hex color code for visual identification
	Created     int64  `json:"created" db:"created"`
	Modified    int64  `json:"modified" db:"modified"`
	Deleted     bool   `json:"deleted" db:"deleted"`
}

// LabelCategory represents a category for organizing labels
type LabelCategory struct {
	ID          string `json:"id" db:"id"`
	Title       string `json:"title" db:"title" binding:"required"`
	Description string `json:"description,omitempty" db:"description"`
	Created     int64  `json:"created" db:"created"`
	Modified    int64  `json:"modified" db:"modified"`
	Deleted     bool   `json:"deleted" db:"deleted"`
}

// LabelLabelCategoryMapping represents the many-to-many relationship between labels and categories
type LabelLabelCategoryMapping struct {
	ID              string `json:"id" db:"id"`
	LabelID         string `json:"labelId" db:"label_id" binding:"required"`
	LabelCategoryID string `json:"labelCategoryId" db:"label_category_id" binding:"required"`
	Created         int64  `json:"created" db:"created"`
	Modified        int64  `json:"modified" db:"modified"`
	Deleted         bool   `json:"deleted" db:"deleted"`
}

// LabelTicketMapping represents the many-to-many relationship between labels and tickets
type LabelTicketMapping struct {
	ID       string `json:"id" db:"id"`
	LabelID  string `json:"labelId" db:"label_id" binding:"required"`
	TicketID string `json:"ticketId" db:"ticket_id" binding:"required"`
	Created  int64  `json:"created" db:"created"`
	Modified int64  `json:"modified" db:"modified"`
	Deleted  bool   `json:"deleted" db:"deleted"`
}

// LabelAssetMapping represents the many-to-many relationship between labels and assets
type LabelAssetMapping struct {
	ID       string `json:"id" db:"id"`
	LabelID  string `json:"labelId" db:"label_id" binding:"required"`
	AssetID  string `json:"assetId" db:"asset_id" binding:"required"`
	Created  int64  `json:"created" db:"created"`
	Modified int64  `json:"modified" db:"modified"`
	Deleted  bool   `json:"deleted" db:"deleted"`
}

// LabelTeamMapping represents the many-to-many relationship between labels and teams
type LabelTeamMapping struct {
	ID       string `json:"id" db:"id"`
	LabelID  string `json:"labelId" db:"label_id" binding:"required"`
	TeamID   string `json:"teamId" db:"team_id" binding:"required"`
	Created  int64  `json:"created" db:"created"`
	Modified int64  `json:"modified" db:"modified"`
	Deleted  bool   `json:"deleted" db:"deleted"`
}

// LabelProjectMapping represents the many-to-many relationship between labels and projects
type LabelProjectMapping struct {
	ID        string `json:"id" db:"id"`
	LabelID   string `json:"labelId" db:"label_id" binding:"required"`
	ProjectID string `json:"projectId" db:"project_id" binding:"required"`
	Created   int64  `json:"created" db:"created"`
	Modified  int64  `json:"modified" db:"modified"`
	Deleted   bool   `json:"deleted" db:"deleted"`
}

// GetDisplayName returns a user-friendly display name
func (l *Label) GetDisplayName() string {
	if l.Title != "" {
		return l.Title
	}
	return "Unknown Label"
}

// IsValid checks if the label has required fields
func (l *Label) IsValid() bool {
	return l.Title != ""
}

// GetDisplayName returns a user-friendly display name for the category
func (lc *LabelCategory) GetDisplayName() string {
	if lc.Title != "" {
		return lc.Title
	}
	return "Unknown Category"
}

// IsValid checks if the label category has required fields
func (lc *LabelCategory) IsValid() bool {
	return lc.Title != ""
}

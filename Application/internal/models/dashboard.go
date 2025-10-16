package models

// Dashboard represents a customizable dashboard with widgets
type Dashboard struct {
	ID          string  `json:"id" db:"id"`
	Title       string  `json:"title" db:"title" binding:"required"`
	Description string  `json:"description,omitempty" db:"description"`
	OwnerID     string  `json:"ownerId" db:"owner_id" binding:"required"`
	IsPublic    bool    `json:"isPublic" db:"is_public"`
	IsFavorite  bool    `json:"isFavorite" db:"is_favorite"`
	Layout      *string `json:"layout,omitempty" db:"layout"` // JSON layout configuration
	Created     int64   `json:"created" db:"created"`
	Modified    int64   `json:"modified" db:"modified"`
	Deleted     bool    `json:"deleted" db:"deleted"`
	Version     int     `json:"version" db:"version"` // Optimistic locking version
}

// DashboardWidget represents a widget on a dashboard
type DashboardWidget struct {
	ID            string  `json:"id" db:"id"`
	DashboardID   string  `json:"dashboardId" db:"dashboard_id" binding:"required"`
	WidgetType    string  `json:"widgetType" db:"widget_type" binding:"required"`
	Title         *string `json:"title,omitempty" db:"title"`
	PositionX     *int    `json:"positionX,omitempty" db:"position_x"`
	PositionY     *int    `json:"positionY,omitempty" db:"position_y"`
	Width         *int    `json:"width,omitempty" db:"width"`
	Height        *int    `json:"height,omitempty" db:"height"`
	Configuration *string `json:"configuration,omitempty" db:"configuration"` // JSON widget configuration
	Created       int64   `json:"created" db:"created"`
	Modified      int64   `json:"modified" db:"modified"`
	Deleted       bool    `json:"deleted" db:"deleted"`
}

// DashboardShareMapping represents sharing a dashboard with users/teams/projects
type DashboardShareMapping struct {
	ID          string  `json:"id" db:"id"`
	DashboardID string  `json:"dashboardId" db:"dashboard_id" binding:"required"`
	UserID      *string `json:"userId,omitempty" db:"user_id"`       // NULL if not user-specific
	TeamID      *string `json:"teamId,omitempty" db:"team_id"`       // NULL if not team-specific
	ProjectID   *string `json:"projectId,omitempty" db:"project_id"` // NULL if not project-specific
	Created     int64   `json:"created" db:"created"`
	Deleted     bool    `json:"deleted" db:"deleted"`
}

// Widget type constants
const (
	WidgetTypeFilterResults  = "filter_results"
	WidgetTypePieChart       = "pie_chart"
	WidgetTypeBarChart       = "bar_chart"
	WidgetTypeLineChart      = "line_chart"
	WidgetTypeActivityStream = "activity_stream"
	WidgetTypeStatistics     = "statistics"
	WidgetTypeRecentTickets  = "recent_tickets"
	WidgetTypeAssignedToMe   = "assigned_to_me"
	WidgetTypeCreatedByMe    = "created_by_me"
	WidgetTypeHeatMap        = "heat_map"
)

// IsOwner checks if the given userID is the owner
func (d *Dashboard) IsOwner(userID string) bool {
	return d.OwnerID == userID
}

// IsValidWidgetType checks if the widget type is valid
func (dw *DashboardWidget) IsValidWidgetType() bool {
	validTypes := map[string]bool{
		WidgetTypeFilterResults:  true,
		WidgetTypePieChart:       true,
		WidgetTypeBarChart:       true,
		WidgetTypeLineChart:      true,
		WidgetTypeActivityStream: true,
		WidgetTypeStatistics:     true,
		WidgetTypeRecentTickets:  true,
		WidgetTypeAssignedToMe:   true,
		WidgetTypeCreatedByMe:    true,
		WidgetTypeHeatMap:        true,
	}
	return validTypes[dw.WidgetType]
}

// GetShareType returns the type of share (user, team, or project)
func (ds *DashboardShareMapping) GetShareType() string {
	if ds.UserID != nil && *ds.UserID != "" {
		return "user"
	}
	if ds.TeamID != nil && *ds.TeamID != "" {
		return "team"
	}
	if ds.ProjectID != nil && *ds.ProjectID != "" {
		return "project"
	}
	return "unknown"
}

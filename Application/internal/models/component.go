package models

// Component represents a project component (module, subsystem, feature area)
type Component struct {
	ID          string `json:"id" db:"id"`
	Title       string `json:"title" db:"title" binding:"required"`
	Description string `json:"description,omitempty" db:"description"`
	Created     int64  `json:"created" db:"created"`
	Modified    int64  `json:"modified" db:"modified"`
	Deleted     bool   `json:"deleted" db:"deleted"`
}

// ComponentMetaData represents additional metadata for a component
type ComponentMetaData struct {
	ID          string `json:"id" db:"id"`
	ComponentID string `json:"componentId" db:"component_id" binding:"required"`
	Property    string `json:"property" db:"property" binding:"required"`
	Value       string `json:"value" db:"value"`
	Created     int64  `json:"created" db:"created"`
	Modified    int64  `json:"modified" db:"modified"`
	Deleted     bool   `json:"deleted" db:"deleted"`
}

// ComponentTicketMapping represents the many-to-many relationship between components and tickets
type ComponentTicketMapping struct {
	ID          string `json:"id" db:"id"`
	ComponentID string `json:"componentId" db:"component_id" binding:"required"`
	TicketID    string `json:"ticketId" db:"ticket_id" binding:"required"`
	Created     int64  `json:"created" db:"created"`
	Modified    int64  `json:"modified" db:"modified"`
	Deleted     bool   `json:"deleted" db:"deleted"`
}

// GetDisplayName returns a user-friendly display name
func (c *Component) GetDisplayName() string {
	if c.Title != "" {
		return c.Title
	}
	return "Unknown Component"
}

// IsValid checks if the component has required fields
func (c *Component) IsValid() bool {
	return c.Title != ""
}

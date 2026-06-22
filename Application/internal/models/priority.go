package models

// Priority represents a ticket priority level (Lowest, Low, Medium, High, Highest)
type Priority struct {
	ID          string `json:"id" db:"id"`
	Title       string `json:"title" db:"title" binding:"required"`
	Description string `json:"description,omitempty" db:"description"`
	Level       int    `json:"level" db:"level" binding:"required"` // 1 (Lowest) to 5 (Highest)
	Icon        string `json:"icon,omitempty" db:"icon"`
	Color       string `json:"color,omitempty" db:"color"` // Hex color code
	Created     int64  `json:"created" db:"created"`
	Modified    int64  `json:"modified" db:"modified"`
	Deleted     bool   `json:"deleted" db:"deleted"`
}

// Priority level constants
const (
	PriorityLevelLowest  = 1
	PriorityLevelLow     = 2
	PriorityLevelMedium  = 3
	PriorityLevelHigh    = 4
	PriorityLevelHighest = 5
)

// Default priority IDs
const (
	PriorityIDLowest  = "priority-lowest"
	PriorityIDLow     = "priority-low"
	PriorityIDMedium  = "priority-medium"
	PriorityIDHigh    = "priority-high"
	PriorityIDHighest = "priority-highest"
)

// IsValidLevel checks if the priority level is valid (1-5)
func (p *Priority) IsValidLevel() bool {
	return p.Level >= PriorityLevelLowest && p.Level <= PriorityLevelHighest
}

// GetDisplayName returns a user-friendly display name
func (p *Priority) GetDisplayName() string {
	if p.Title != "" {
		return p.Title
	}
	return "Unknown Priority"
}

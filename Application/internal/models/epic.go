package models

// Epic represents an epic - a large user story that can contain multiple stories
// Epics are implemented as a special type of ticket with additional fields
type Epic struct {
	ID        string  `json:"id" db:"id"`
	TicketID  string  `json:"ticketId" db:"ticket_id"` // References ticket table
	EpicColor *string `json:"epicColor,omitempty" db:"epic_color"`
	EpicName  *string `json:"epicName,omitempty" db:"epic_name"`
	IsEpic    bool    `json:"isEpic" db:"is_epic"`
}

// EpicStoryMapping represents a story belonging to an epic
// This is conceptual - in practice, stories have epic_id field in ticket table
type EpicStoryMapping struct {
	EpicID  string `json:"epicId"`
	StoryID string `json:"storyId"`
}

// Epic color constants (standard JIRA epic colors)
const (
	EpicColorGhola    = "#6554C0" // Purple
	EpicColorWestar   = "#00B8D9" // Cyan
	EpicColorJungle   = "#00875A" // Green
	EpicColorKournikova = "#FFAB00" // Yellow
	EpicColorRust     = "#FF8B00" // Orange
	EpicColorMonza    = "#DE350B" // Red
	EpicColorStorm    = "#5E6C84" // Grey
)

// IsEpicTicket checks if this represents an epic ticket
func (e *Epic) IsEpicTicket() bool {
	return e.IsEpic
}

// GetColor returns the epic color or default
func (e *Epic) GetColor() string {
	if e.EpicColor != nil && *e.EpicColor != "" {
		return *e.EpicColor
	}
	return EpicColorGhola // Default to purple
}

// GetName returns the epic name or empty string
func (e *Epic) GetName() string {
	if e.EpicName != nil {
		return *e.EpicName
	}
	return ""
}

// HasName checks if the epic has a name set
func (e *Epic) HasName() bool {
	return e.EpicName != nil && *e.EpicName != ""
}

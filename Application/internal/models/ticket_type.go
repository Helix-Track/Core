package models

// TicketType represents a ticket type (Bug, Task, Story, Epic)
type TicketType struct {
	ID          string `json:"id" db:"id"`
	Title       string `json:"title" db:"title" binding:"required"`
	Description string `json:"description,omitempty" db:"description"`
	Icon        string `json:"icon,omitempty" db:"icon"`
	Color       string `json:"color,omitempty" db:"color"`
	Created     int64  `json:"created" db:"created"`
	Modified    int64  `json:"modified" db:"modified"`
	Deleted     bool   `json:"deleted" db:"deleted"`
}

// Default type IDs
const (
	TypeIDBug   = "type-bug"
	TypeIDTask  = "type-task"
	TypeIDStory = "type-story"
	TypeIDEpic  = "type-epic"
)

// TicketTypeProjectMapping maps ticket types to projects
type TicketTypeProjectMapping struct {
	ID           string `json:"id" db:"id"`
	TicketTypeID string `json:"ticketTypeId" db:"ticket_type_id" binding:"required"`
	ProjectID    string `json:"projectId" db:"project_id" binding:"required"`
	Created      int64  `json:"created" db:"created"`
	Deleted      bool   `json:"deleted" db:"deleted"`
}

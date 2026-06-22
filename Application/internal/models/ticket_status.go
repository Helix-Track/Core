package models

// TicketStatus represents the status of a ticket
type TicketStatus struct {
	ID          string `json:"id" db:"id"`
	Title       string `json:"title" db:"title" binding:"required"`
	Description string `json:"description,omitempty" db:"description"`
	Created     int64  `json:"created" db:"created"`
	Modified    int64  `json:"modified" db:"modified"`
	Deleted     bool   `json:"deleted" db:"deleted"`
}

// Default status IDs
const (
	StatusIDOpen       = "status-open"
	StatusIDInProgress = "status-in-progress"
	StatusIDDone       = "status-done"
	StatusIDClosed     = "status-closed"
)

package models

// Board represents a Kanban/Scrum board for organizing tickets
type Board struct {
	ID          string `json:"id" db:"id"`
	Title       string `json:"title" db:"title" binding:"required"`
	Description string `json:"description,omitempty" db:"description"`
	Created     int64  `json:"created" db:"created"`
	Modified    int64  `json:"modified" db:"modified"`
	Deleted     bool   `json:"deleted" db:"deleted"`
	Version     int    `json:"version" db:"version"` // Optimistic locking version
}

// BoardMetaData represents additional metadata for boards (properties like board type, columns, etc.)
type BoardMetaData struct {
	ID       string `json:"id" db:"id"`
	BoardID  string `json:"boardId" db:"board_id" binding:"required"`
	Property string `json:"property" db:"property" binding:"required"`
	Value    string `json:"value" db:"value"`
	Created  int64  `json:"created" db:"created"`
	Modified int64  `json:"modified" db:"modified"`
	Deleted  bool   `json:"deleted" db:"deleted"`
}

// TicketBoardMapping represents the relationship between tickets and boards
type TicketBoardMapping struct {
	ID       string `json:"id" db:"id"`
	TicketID string `json:"ticketId" db:"ticket_id" binding:"required"`
	BoardID  string `json:"boardId" db:"board_id" binding:"required"`
	Created  int64  `json:"created" db:"created"`
	Modified int64  `json:"modified" db:"modified"`
	Deleted  bool   `json:"deleted" db:"deleted"`
}

// Common board metadata property keys
const (
	BoardPropertyType        = "type"        // kanban, scrum, custom
	BoardPropertyColumns     = "columns"     // JSON array of column definitions
	BoardPropertyDefaultView = "defaultView" // list, board, calendar
	BoardPropertyOwner       = "owner"       // User ID of board owner
	BoardPropertyTeam        = "team"        // Team ID if board is team-specific
	BoardPropertyProject     = "project"     // Project ID if board is project-specific
)

// Board type constants
const (
	BoardTypeKanban = "kanban"
	BoardTypeScrum  = "scrum"
	BoardTypeCustom = "custom"
)

// GetDisplayName returns a user-friendly display name for the board
func (b *Board) GetDisplayName() string {
	if b.Title != "" {
		return b.Title
	}
	return "Untitled Board"
}

// IsValid checks if the board has required fields
func (b *Board) IsValid() bool {
	return b.ID != "" && b.Title != ""
}

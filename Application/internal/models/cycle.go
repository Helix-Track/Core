package models

// Cycle represents a sprint/milestone/release (Agile iteration)
// Cycles form a hierarchy: Release (1000) > Milestone (100) > Sprint (10)
type Cycle struct {
	ID          string `json:"id" db:"id"`
	Title       string `json:"title" db:"title" binding:"required"`
	Description string `json:"description,omitempty" db:"description"`
	CycleID     string `json:"cycleId,omitempty" db:"cycle_id"` // Parent cycle ID
	Type        int    `json:"type" db:"type" binding:"required"`
	Created     int64  `json:"created" db:"created"`
	Modified    int64  `json:"modified" db:"modified"`
	Deleted     bool   `json:"deleted" db:"deleted"`
}

// Cycle type constants
const (
	CycleTypeRelease   = 1000 // Top-level release cycle
	CycleTypeMilestone = 100  // Mid-level milestone
	CycleTypeSprint    = 10   // Bottom-level sprint
)

// CycleProjectMapping represents the many-to-many relationship between cycles and projects
type CycleProjectMapping struct {
	ID        string `json:"id" db:"id"`
	CycleID   string `json:"cycleId" db:"cycle_id" binding:"required"`
	ProjectID string `json:"projectId" db:"project_id" binding:"required"`
	Created   int64  `json:"created" db:"created"`
	Modified  int64  `json:"modified" db:"modified"`
	Deleted   bool   `json:"deleted" db:"deleted"`
}

// TicketCycleMapping represents the many-to-many relationship between tickets and cycles
type TicketCycleMapping struct {
	ID       string `json:"id" db:"id"`
	TicketID string `json:"ticketId" db:"ticket_id" binding:"required"`
	CycleID  string `json:"cycleId" db:"cycle_id" binding:"required"`
	Created  int64  `json:"created" db:"created"`
	Modified int64  `json:"modified" db:"modified"`
	Deleted  bool   `json:"deleted" db:"deleted"`
}

// IsValidType checks if the cycle type is valid (10, 100, or 1000)
func (c *Cycle) IsValidType() bool {
	return c.Type == CycleTypeSprint || c.Type == CycleTypeMilestone || c.Type == CycleTypeRelease
}

// GetTypeName returns a user-friendly type name
func (c *Cycle) GetTypeName() string {
	switch c.Type {
	case CycleTypeRelease:
		return "Release"
	case CycleTypeMilestone:
		return "Milestone"
	case CycleTypeSprint:
		return "Sprint"
	default:
		return "Unknown"
	}
}

// IsValidParent checks if the parent cycle type is valid for this cycle
// Parent's type must be greater than current cycle's type (Release > Milestone > Sprint)
func (c *Cycle) IsValidParent(parentType int) bool {
	if c.CycleID == "" {
		return true // No parent is valid
	}
	return parentType > c.Type
}

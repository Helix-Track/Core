package models

// Subtask represents a subtask - a smaller task that belongs to a parent ticket
// Subtasks are implemented as a special type of ticket with parent reference
type Subtask struct {
	ID             string `json:"id" db:"id"`
	TicketID       string `json:"ticketId" db:"ticket_id"`         // References ticket table
	ParentTicketID string `json:"parentTicketId" db:"parent_ticket_id"` // Parent ticket
	IsSubtask      bool   `json:"isSubtask" db:"is_subtask"`
}

// SubtaskSummary provides a summary of subtasks for a parent ticket
type SubtaskSummary struct {
	ParentTicketID string `json:"parentTicketId"`
	TotalSubtasks  int    `json:"totalSubtasks"`
	CompletedSubtasks int `json:"completedSubtasks"`
	PercentComplete float64 `json:"percentComplete"`
}

// IsSubtaskTicket checks if this represents a subtask ticket
func (s *Subtask) IsSubtaskTicket() bool {
	return s.IsSubtask
}

// HasParent checks if the subtask has a parent ticket assigned
func (s *Subtask) HasParent() bool {
	return s.ParentTicketID != ""
}

// CalculatePercentComplete calculates the completion percentage for subtasks
func (ss *SubtaskSummary) CalculatePercentComplete() {
	if ss.TotalSubtasks == 0 {
		ss.PercentComplete = 0
		return
	}
	ss.PercentComplete = float64(ss.CompletedSubtasks) / float64(ss.TotalSubtasks) * 100.0
}

// IsComplete checks if all subtasks are complete
func (ss *SubtaskSummary) IsComplete() bool {
	return ss.TotalSubtasks > 0 && ss.CompletedSubtasks == ss.TotalSubtasks
}

// HasSubtasks checks if there are any subtasks
func (ss *SubtaskSummary) HasSubtasks() bool {
	return ss.TotalSubtasks > 0
}

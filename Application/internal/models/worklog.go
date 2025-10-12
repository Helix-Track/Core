package models

// WorkLog represents a detailed time tracking entry for a ticket
type WorkLog struct {
	ID          string `json:"id" db:"id"`
	TicketID    string `json:"ticketId" db:"ticket_id" binding:"required"`
	UserID      string `json:"userId" db:"user_id" binding:"required"`
	TimeSpent   int    `json:"timeSpent" db:"time_spent" binding:"required"` // In minutes
	WorkDate    int64  `json:"workDate" db:"work_date" binding:"required"`   // Unix timestamp
	Description string `json:"description,omitempty" db:"description"`
	Created     int64  `json:"created" db:"created"`
	Modified    int64  `json:"modified" db:"modified"`
	Deleted     bool   `json:"deleted" db:"deleted"`
}

// GetTimeSpentHours returns the time spent in hours
func (w *WorkLog) GetTimeSpentHours() float64 {
	return float64(w.TimeSpent) / 60.0
}

// GetTimeSpentDays returns the time spent in days (assuming 8-hour workday)
func (w *WorkLog) GetTimeSpentDays() float64 {
	return float64(w.TimeSpent) / (8.0 * 60.0)
}

// IsValid checks if the work log has valid data
func (w *WorkLog) IsValid() bool {
	return w.TicketID != "" && w.UserID != "" && w.TimeSpent > 0
}

package models

// TicketWatcherMapping represents a user watching a ticket for notifications
type TicketWatcherMapping struct {
	ID       string `json:"id" db:"id"`
	TicketID string `json:"ticketId" db:"ticket_id" binding:"required"`
	UserID   string `json:"userId" db:"user_id" binding:"required"`
	Created  int64  `json:"created" db:"created"`
	Deleted  bool   `json:"deleted" db:"deleted"`
}

// IsWatching checks if a user is watching a specific ticket
func IsWatching(userID, ticketID string, watchers []TicketWatcherMapping) bool {
	for _, watcher := range watchers {
		if watcher.UserID == userID && watcher.TicketID == ticketID && !watcher.Deleted {
			return true
		}
	}
	return false
}

// GetWatcherCount returns the number of active watchers for a ticket
func GetWatcherCount(ticketID string, watchers []TicketWatcherMapping) int {
	count := 0
	for _, watcher := range watchers {
		if watcher.TicketID == ticketID && !watcher.Deleted {
			count++
		}
	}
	return count
}

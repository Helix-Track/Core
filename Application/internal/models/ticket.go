package models

import "time"

// Ticket represents a ticket/issue in the system
type Ticket struct {
	ID             string  `json:"id" db:"id"`
	TicketNumber   int     `json:"ticketNumber" db:"ticket_number"`
	Title          string  `json:"title" db:"title"`
	Description    string  `json:"description" db:"description"`
	TicketTypeID   string  `json:"ticketTypeId" db:"ticket_type_id"`
	TicketStatusID string  `json:"ticketStatusId" db:"ticket_status_id"`
	ProjectID      string  `json:"projectId" db:"project_id"`
	UserID         string  `json:"userId" db:"user_id"`         // Assignee
	Creator        string  `json:"creator" db:"creator"`        // Creator username
	Estimation     *int    `json:"estimation" db:"estimation"`  // In hours
	StoryPoints    *int    `json:"storyPoints" db:"story_points"`
	Created        int64   `json:"created" db:"created"`
	Modified       int64   `json:"modified" db:"modified"`
	Deleted        bool    `json:"deleted" db:"deleted"`
}

// NewTicket creates a new ticket with current timestamps
func NewTicket(id, title, description, ticketTypeID, ticketStatusID, projectID, userID, creator string, ticketNumber int) *Ticket {
	now := time.Now().Unix()
	return &Ticket{
		ID:             id,
		TicketNumber:   ticketNumber,
		Title:          title,
		Description:    description,
		TicketTypeID:   ticketTypeID,
		TicketStatusID: ticketStatusID,
		ProjectID:      projectID,
		UserID:         userID,
		Creator:        creator,
		Created:        now,
		Modified:       now,
		Deleted:        false,
	}
}

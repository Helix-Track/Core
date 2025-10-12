package models

// Vote represents a user's vote on a ticket
type Vote struct {
	ID       string `json:"id" db:"id"`
	TicketID string `json:"ticketId" db:"ticket_id" binding:"required"`
	UserID   string `json:"userId" db:"user_id" binding:"required"`
	Created  int64  `json:"created" db:"created"`
	Deleted  bool   `json:"deleted" db:"deleted"`
}

// VoteSummary provides voting statistics for a ticket
type VoteSummary struct {
	TicketID   string   `json:"ticketId"`
	VoteCount  int      `json:"voteCount"`
	VoterIDs   []string `json:"voterIds,omitempty"`
	HasVoted   bool     `json:"hasVoted"` // Whether the current user has voted
}

// IsValid checks if the vote has valid data
func (v *Vote) IsValid() bool {
	return v.TicketID != "" && v.UserID != ""
}

// HasVotes checks if there are any votes
func (vs *VoteSummary) HasVotes() bool {
	return vs.VoteCount > 0
}

// IsPopular checks if the ticket has a significant number of votes
func (vs *VoteSummary) IsPopular(threshold int) bool {
	return vs.VoteCount >= threshold
}

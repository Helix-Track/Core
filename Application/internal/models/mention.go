package models

// Mention represents a user mention in a comment
type Mention struct {
	ID              string `json:"id" db:"id"`
	CommentID       string `json:"commentId" db:"comment_id" binding:"required"`
	MentionedUserID string `json:"mentionedUserId" db:"mentioned_user_id" binding:"required"`
	Created         int64  `json:"created" db:"created"`
	Deleted         bool   `json:"deleted" db:"deleted"`
}

// MentionSummary provides a summary of mentions for a user or comment
type MentionSummary struct {
	CommentID       string   `json:"commentId,omitempty"`
	UserID          string   `json:"userId,omitempty"`
	MentionCount    int      `json:"mentionCount"`
	MentionedUserIDs []string `json:"mentionedUserIds,omitempty"`
}

// IsValid checks if the mention has valid data
func (m *Mention) IsValid() bool {
	return m.CommentID != "" && m.MentionedUserID != ""
}

// HasMentions checks if there are any mentions
func (ms *MentionSummary) HasMentions() bool {
	return ms.MentionCount > 0
}

// ContainsUser checks if a specific user is mentioned
func (ms *MentionSummary) ContainsUser(userID string) bool {
	for _, id := range ms.MentionedUserIDs {
		if id == userID {
			return true
		}
	}
	return false
}

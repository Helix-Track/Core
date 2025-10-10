package models

import "time"

// Comment represents a comment in the system
type Comment struct {
	ID          string `json:"id" db:"id"`
	Title       string `json:"title" db:"title"`
	Description string `json:"description" db:"description"`
	UserID      string `json:"userId" db:"user_id"`   // Comment author
	ParentID    string `json:"parentId" db:"parent_id"` // For nested comments
	Created     int64  `json:"created" db:"created"`
	Modified    int64  `json:"modified" db:"modified"`
	Deleted     bool   `json:"deleted" db:"deleted"`
}

// NewComment creates a new comment with current timestamps
func NewComment(id, title, description, userID, parentID string) *Comment {
	now := time.Now().Unix()
	return &Comment{
		ID:          id,
		Title:       title,
		Description: description,
		UserID:      userID,
		ParentID:    parentID,
		Created:     now,
		Modified:    now,
		Deleted:     false,
	}
}

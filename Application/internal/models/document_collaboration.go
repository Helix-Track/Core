package models

import (
	"errors"
	"time"
)

// DocumentComment represents a comment on a document
type DocumentComment struct {
	ID         string  `json:"id" db:"id"`
	DocumentID string  `json:"document_id" db:"document_id"`
	UserID     string  `json:"user_id" db:"user_id"`
	Content    string  `json:"content" db:"content"`
	ParentID   *string `json:"parent_id,omitempty" db:"parent_id"` // For threading
	Version    int     `json:"version" db:"version"`               // Comment version (for edits)
	IsResolved bool    `json:"is_resolved" db:"is_resolved"`
	Created    int64   `json:"created" db:"created"`
	Modified   int64   `json:"modified" db:"modified"`
	Deleted    bool    `json:"deleted" db:"deleted"`
}

// Validate validates the document comment
func (dc *DocumentComment) Validate() error {
	if dc.ID == "" {
		return errors.New("comment ID cannot be empty")
	}
	if dc.DocumentID == "" {
		return errors.New("comment document ID cannot be empty")
	}
	if dc.UserID == "" {
		return errors.New("comment user ID cannot be empty")
	}
	if dc.Content == "" {
		return errors.New("comment content cannot be empty")
	}
	if dc.Version < 1 {
		return errors.New("comment version must be at least 1")
	}
	if dc.Created == 0 {
		return errors.New("comment created timestamp cannot be zero")
	}
	if dc.Modified == 0 {
		return errors.New("comment modified timestamp cannot be zero")
	}
	return nil
}

// SetTimestamps sets created and modified timestamps
func (dc *DocumentComment) SetTimestamps() {
	now := time.Now().Unix()
	if dc.Created == 0 {
		dc.Created = now
	}
	dc.Modified = now
}

// DocumentInlineComment represents an inline comment with position
type DocumentInlineComment struct {
	ID            string  `json:"id" db:"id"`
	DocumentID    string  `json:"document_id" db:"document_id"`
	CommentID     string  `json:"comment_id" db:"comment_id"`
	PositionStart int     `json:"position_start" db:"position_start"`
	PositionEnd   int     `json:"position_end" db:"position_end"`
	SelectedText  *string `json:"selected_text,omitempty" db:"selected_text"`
	IsResolved    bool    `json:"is_resolved" db:"is_resolved"`
	Created       int64   `json:"created" db:"created"`
}

// Validate validates the inline comment
func (dic *DocumentInlineComment) Validate() error {
	if dic.ID == "" {
		return errors.New("inline comment ID cannot be empty")
	}
	if dic.DocumentID == "" {
		return errors.New("inline comment document ID cannot be empty")
	}
	if dic.CommentID == "" {
		return errors.New("inline comment comment ID cannot be empty")
	}
	if dic.PositionStart < 0 {
		return errors.New("position start cannot be negative")
	}
	if dic.PositionEnd < dic.PositionStart {
		return errors.New("position end must be >= position start")
	}
	if dic.Created == 0 {
		return errors.New("inline comment created timestamp cannot be zero")
	}
	return nil
}

// SetTimestamps sets the created timestamp
func (dic *DocumentInlineComment) SetTimestamps() {
	if dic.Created == 0 {
		dic.Created = time.Now().Unix()
	}
}

// DocumentMention represents a user mention in a document
type DocumentMention struct {
	ID               string  `json:"id" db:"id"`
	DocumentID       string  `json:"document_id" db:"document_id"`
	MentionedUserID  string  `json:"mentioned_user_id" db:"mentioned_user_id"`
	MentioningUserID string  `json:"mentioning_user_id" db:"mentioning_user_id"`
	MentionContext   *string `json:"mention_context,omitempty" db:"mention_context"`
	Position         *int    `json:"position,omitempty" db:"position"`
	IsAcknowledged   bool    `json:"is_acknowledged" db:"is_acknowledged"`
	Created          int64   `json:"created" db:"created"`
}

// Validate validates the document mention
func (dm *DocumentMention) Validate() error {
	if dm.ID == "" {
		return errors.New("mention ID cannot be empty")
	}
	if dm.DocumentID == "" {
		return errors.New("mention document ID cannot be empty")
	}
	if dm.MentionedUserID == "" {
		return errors.New("mentioned user ID cannot be empty")
	}
	if dm.MentioningUserID == "" {
		return errors.New("mentioning user ID cannot be empty")
	}
	if dm.Created == 0 {
		return errors.New("mention created timestamp cannot be zero")
	}
	return nil
}

// SetTimestamps sets the created timestamp
func (dm *DocumentMention) SetTimestamps() {
	if dm.Created == 0 {
		dm.Created = time.Now().Unix()
	}
}

// DocumentReaction represents a reaction (like, emoji) on a document
type DocumentReaction struct {
	ID           string  `json:"id" db:"id"`
	DocumentID   string  `json:"document_id" db:"document_id"`
	UserID       string  `json:"user_id" db:"user_id"`
	ReactionType string  `json:"reaction_type" db:"reaction_type"` // "like", "love", "thumbsup"
	Emoji        *string `json:"emoji,omitempty" db:"emoji"`
	Created      int64   `json:"created" db:"created"`
}

// Validate validates the document reaction
func (dr *DocumentReaction) Validate() error {
	if dr.ID == "" {
		return errors.New("reaction ID cannot be empty")
	}
	if dr.DocumentID == "" {
		return errors.New("reaction document ID cannot be empty")
	}
	if dr.UserID == "" {
		return errors.New("reaction user ID cannot be empty")
	}
	if dr.ReactionType == "" {
		return errors.New("reaction type cannot be empty")
	}
	if dr.Created == 0 {
		return errors.New("reaction created timestamp cannot be zero")
	}
	return nil
}

// SetTimestamps sets the created timestamp
func (dr *DocumentReaction) SetTimestamps() {
	if dr.Created == 0 {
		dr.Created = time.Now().Unix()
	}
}

// DocumentWatcher represents a user watching a document
type DocumentWatcher struct {
	ID                string `json:"id" db:"id"`
	DocumentID        string `json:"document_id" db:"document_id"`
	UserID            string `json:"user_id" db:"user_id"`
	NotificationLevel string `json:"notification_level" db:"notification_level"` // "all", "mentions", "none"
	Created           int64  `json:"created" db:"created"`
}

// Validate validates the document watcher
func (dw *DocumentWatcher) Validate() error {
	if dw.ID == "" {
		return errors.New("watcher ID cannot be empty")
	}
	if dw.DocumentID == "" {
		return errors.New("watcher document ID cannot be empty")
	}
	if dw.UserID == "" {
		return errors.New("watcher user ID cannot be empty")
	}
	if dw.NotificationLevel == "" {
		return errors.New("notification level cannot be empty")
	}
	validLevels := map[string]bool{
		"all": true, "mentions": true, "none": true,
	}
	if !validLevels[dw.NotificationLevel] {
		return errors.New("invalid notification level")
	}
	if dw.Created == 0 {
		return errors.New("watcher created timestamp cannot be zero")
	}
	return nil
}

// SetTimestamps sets the created timestamp
func (dw *DocumentWatcher) SetTimestamps() {
	if dw.Created == 0 {
		dw.Created = time.Now().Unix()
	}
}

// DocumentLabel represents a reusable label
type DocumentLabel struct {
	ID          string  `json:"id" db:"id"`
	Name        string  `json:"name" db:"name"`
	Description *string `json:"description,omitempty" db:"description"`
	Color       *string `json:"color,omitempty" db:"color"` // Hex color code
	Created     int64   `json:"created" db:"created"`
	Deleted     bool    `json:"deleted" db:"deleted"`
}

// Validate validates the document label
func (dl *DocumentLabel) Validate() error {
	if dl.ID == "" {
		return errors.New("label ID cannot be empty")
	}
	if dl.Name == "" {
		return errors.New("label name cannot be empty")
	}
	if dl.Created == 0 {
		return errors.New("label created timestamp cannot be zero")
	}
	return nil
}

// SetTimestamps sets the created timestamp
func (dl *DocumentLabel) SetTimestamps() {
	if dl.Created == 0 {
		dl.Created = time.Now().Unix()
	}
}

// DocumentTag represents a document tag
type DocumentTag struct {
	ID      string `json:"id" db:"id"`
	Name    string `json:"name" db:"name"`
	Created int64  `json:"created" db:"created"`
	Deleted bool   `json:"deleted" db:"deleted"`
}

// Validate validates the document tag
func (dt *DocumentTag) Validate() error {
	if dt.ID == "" {
		return errors.New("tag ID cannot be empty")
	}
	if dt.Name == "" {
		return errors.New("tag name cannot be empty")
	}
	if dt.Created == 0 {
		return errors.New("tag created timestamp cannot be zero")
	}
	return nil
}

// SetTimestamps sets the created timestamp
func (dt *DocumentTag) SetTimestamps() {
	if dt.Created == 0 {
		dt.Created = time.Now().Unix()
	}
}

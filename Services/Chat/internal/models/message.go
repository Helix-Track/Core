package models

import (
	"encoding/json"

	"github.com/google/uuid"
)

// MessageType represents the type of message
type MessageType string

const (
	MessageTypeText  MessageType = "text"
	MessageTypeReply MessageType = "reply"
	MessageTypeQuote MessageType = "quote"
	MessageTypeSystem MessageType = "system"
	MessageTypeFile  MessageType = "file"
	MessageTypeCode  MessageType = "code"
	MessageTypePoll  MessageType = "poll"
)

// ContentFormat represents the format of message content
type ContentFormat string

const (
	ContentFormatPlain    ContentFormat = "plain"
	ContentFormatMarkdown ContentFormat = "markdown"
	ContentFormatHTML     ContentFormat = "html"
)

// Message represents a chat message
type Message struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	ChatRoomID      uuid.UUID       `json:"chat_room_id" db:"chat_room_id"`
	SenderID        uuid.UUID       `json:"sender_id" db:"sender_id"`
	ParentID        *uuid.UUID      `json:"parent_id,omitempty" db:"parent_id"`
	QuotedMessageID *uuid.UUID      `json:"quoted_message_id,omitempty" db:"quoted_message_id"`
	Type            MessageType     `json:"type" db:"type"`
	Content         string          `json:"content" db:"content"`
	ContentFormat   ContentFormat   `json:"content_format" db:"content_format"`
	Metadata        json.RawMessage `json:"metadata,omitempty" db:"metadata"`
	IsEdited        bool            `json:"is_edited" db:"is_edited"`
	EditedAt        *int64          `json:"edited_at,omitempty" db:"edited_at"`
	IsPinned        bool            `json:"is_pinned" db:"is_pinned"`
	PinnedAt        *int64          `json:"pinned_at,omitempty" db:"pinned_at"`
	PinnedBy        *uuid.UUID      `json:"pinned_by,omitempty" db:"pinned_by"`
	CreatedAt       int64           `json:"created_at" db:"created_at"`
	UpdatedAt       int64           `json:"updated_at" db:"updated_at"`
	Deleted         bool            `json:"deleted" db:"deleted"`
	DeletedAt       *int64          `json:"deleted_at,omitempty" db:"deleted_at"`
}

// MessageRequest represents the request to send a message
type MessageRequest struct {
	ChatRoomID      uuid.UUID       `json:"chat_room_id" binding:"required"`
	ParentID        *uuid.UUID      `json:"parent_id"`
	QuotedMessageID *uuid.UUID      `json:"quoted_message_id"`
	Type            MessageType     `json:"type" binding:"required"`
	Content         string          `json:"content" binding:"required,min=1"`
	ContentFormat   ContentFormat   `json:"content_format"`
	Metadata        json.RawMessage `json:"metadata"`
}

// MessageUpdateRequest represents the request to update a message
type MessageUpdateRequest struct {
	Content       string          `json:"content" binding:"required,min=1"`
	ContentFormat ContentFormat   `json:"content_format"`
	Metadata      json.RawMessage `json:"metadata"`
}

// MessageResponse represents the response with message details
type MessageResponse struct {
	Message        *Message           `json:"message"`
	Sender         *UserInfo          `json:"sender,omitempty"`
	QuotedMessage  *Message           `json:"quoted_message,omitempty"`
	ParentMessage  *Message           `json:"parent_message,omitempty"`
	Reactions      []*MessageReaction `json:"reactions,omitempty"`
	Attachments    []*MessageAttachment `json:"attachments,omitempty"`
	ReadReceipts   []*MessageReadReceipt `json:"read_receipts,omitempty"`
	ReplyCount     int                `json:"reply_count,omitempty"`
}

// MessageListRequest represents the request to list messages
type MessageListRequest struct {
	ChatRoomID uuid.UUID  `json:"chat_room_id" binding:"required"`
	ParentID   *uuid.UUID `json:"parent_id"` // for getting threaded replies
	Limit      int        `json:"limit"`
	Offset     int        `json:"offset"`
	Before     *int64     `json:"before"` // timestamp for pagination
	After      *int64     `json:"after"`  // timestamp for pagination
	Search     string     `json:"search"` // full-text search
}

// MessageEditHistory represents a historical record of a message edit
type MessageEditHistory struct {
	ID                    uuid.UUID       `json:"id" db:"id"`
	MessageID             uuid.UUID       `json:"message_id" db:"message_id"`
	EditorID              uuid.UUID       `json:"editor_id" db:"editor_id"`
	PreviousContent       string          `json:"previous_content" db:"previous_content"`
	PreviousContentFormat ContentFormat   `json:"previous_content_format" db:"previous_content_format"`
	PreviousMetadata      json.RawMessage `json:"previous_metadata,omitempty" db:"previous_metadata"`
	EditNumber            int             `json:"edit_number" db:"edit_number"`
	EditedAt              int64           `json:"edited_at" db:"edited_at"`
	CreatedAt             int64           `json:"created_at" db:"created_at"`
}

// MessageEditHistoryResponse represents the edit history with editor information
type MessageEditHistoryResponse struct {
	EditHistory *MessageEditHistory `json:"edit_history"`
	Editor      *UserInfo           `json:"editor,omitempty"`
}

// MessageWithEditHistory represents a message with its complete edit history
type MessageWithEditHistory struct {
	Message     *Message                      `json:"message"`
	EditHistory []*MessageEditHistoryResponse `json:"edit_history"`
	TotalEdits  int                           `json:"total_edits"`
}

// Validate checks if the message request is valid
func (r *MessageRequest) Validate() error {
	validTypes := map[MessageType]bool{
		MessageTypeText:  true,
		MessageTypeReply: true,
		MessageTypeQuote: true,
		MessageTypeSystem: true,
		MessageTypeFile:  true,
		MessageTypeCode:  true,
		MessageTypePoll:  true,
	}

	if !validTypes[r.Type] {
		return ErrInvalidMessageType
	}

	validFormats := map[ContentFormat]bool{
		ContentFormatPlain:    true,
		ContentFormatMarkdown: true,
		ContentFormatHTML:     true,
	}

	if r.ContentFormat != "" && !validFormats[r.ContentFormat] {
		return ErrInvalidContentFormat
	}

	// Set default format
	if r.ContentFormat == "" {
		r.ContentFormat = ContentFormatPlain
	}

	// Validate quote/reply logic
	if r.Type == MessageTypeReply && r.ParentID == nil {
		return ErrReplyNeedsParent
	}

	if r.Type == MessageTypeQuote && r.QuotedMessageID == nil {
		return ErrQuoteNeedsQuotedMessage
	}

	return nil
}

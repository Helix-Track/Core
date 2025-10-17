package models

import (
	"time"

	"github.com/google/uuid"
)

// ChatExport represents a chat export request
type ChatExport struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	ChatRoomID   uuid.UUID  `json:"chat_room_id" db:"chat_room_id"`
	RequestedBy  uuid.UUID  `json:"requested_by" db:"requested_by"`
	Format       string     `json:"format" db:"format"` // "json", "html", "pdf", "txt"
	Status       string     `json:"status" db:"status"` // "pending", "processing", "completed", "failed"
	FilePath     string     `json:"file_path" db:"file_path"`
	FileSize     int64      `json:"file_size" db:"file_size"`
	ExpiresAt    *time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	CompletedAt  *time.Time `json:"completed_at" db:"completed_at"`
	ErrorMessage string     `json:"error_message" db:"error_message"`
}

// ChatExportRequest represents a request to export a chat
type ChatExportRequest struct {
	ChatRoomID         uuid.UUID  `json:"chat_room_id" binding:"required"`
	Format             string     `json:"format" binding:"required"` // "json", "html", "pdf", "txt"
	IncludeAttachments bool       `json:"include_attachments"`
	StartDate          *time.Time `json:"start_date"`
	EndDate            *time.Time `json:"end_date"`
}

// ChatExportResponse represents the response for export operations
type ChatExportResponse struct {
	Export      *ChatExport `json:"export"`
	DownloadURL string      `json:"download_url,omitempty"`
}

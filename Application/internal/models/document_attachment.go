package models

import (
	"errors"
	"fmt"
	"time"
)

// DocumentAttachment represents a file attached to a document
type DocumentAttachment struct {
	ID               string  `json:"id" db:"id"`
	DocumentID       string  `json:"document_id" db:"document_id"`
	Filename         string  `json:"filename" db:"filename"`
	OriginalFilename string  `json:"original_filename" db:"original_filename"`
	MimeType         string  `json:"mime_type" db:"mime_type"`
	SizeBytes        int     `json:"size_bytes" db:"size_bytes"`
	StoragePath      string  `json:"storage_path" db:"storage_path"`
	Checksum         string  `json:"checksum" db:"checksum"` // SHA-256
	UploaderID       string  `json:"uploader_id" db:"uploader_id"`
	Description      *string `json:"description,omitempty" db:"description"`
	Version          int     `json:"version" db:"version"`
	Created          int64   `json:"created" db:"created"`
	Modified         int64   `json:"modified" db:"modified"`
	Deleted          bool    `json:"deleted" db:"deleted"`
}

// Validate validates the document attachment
func (da *DocumentAttachment) Validate() error {
	if da.ID == "" {
		return errors.New("attachment ID cannot be empty")
	}
	if da.DocumentID == "" {
		return errors.New("attachment document ID cannot be empty")
	}
	if da.Filename == "" {
		return errors.New("attachment filename cannot be empty")
	}
	if da.OriginalFilename == "" {
		return errors.New("attachment original filename cannot be empty")
	}
	if da.MimeType == "" {
		return errors.New("attachment MIME type cannot be empty")
	}
	if da.SizeBytes < 0 {
		return errors.New("attachment size cannot be negative")
	}
	if da.StoragePath == "" {
		return errors.New("attachment storage path cannot be empty")
	}
	if da.Checksum == "" {
		return errors.New("attachment checksum cannot be empty")
	}
	if da.UploaderID == "" {
		return errors.New("attachment uploader ID cannot be empty")
	}
	if da.Version < 1 {
		return errors.New("attachment version must be at least 1")
	}
	if da.Created == 0 {
		return errors.New("attachment created timestamp cannot be zero")
	}
	if da.Modified == 0 {
		return errors.New("attachment modified timestamp cannot be zero")
	}
	return nil
}

// SetTimestamps sets created and modified timestamps
func (da *DocumentAttachment) SetTimestamps() {
	now := time.Now().Unix()
	if da.Created == 0 {
		da.Created = now
	}
	da.Modified = now
}

// IncrementVersion increments the version number
func (da *DocumentAttachment) IncrementVersion() {
	da.Version++
	da.Modified = time.Now().Unix()
}

// IsImage returns true if the attachment is an image
func (da *DocumentAttachment) IsImage() bool {
	imageTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
		"image/svg+xml": true,
	}
	return imageTypes[da.MimeType]
}

// IsDocument returns true if the attachment is a document
func (da *DocumentAttachment) IsDocument() bool {
	docTypes := map[string]bool{
		"application/pdf": true,
		"application/msword": true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
		"application/vnd.ms-excel": true,
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true,
		"text/plain": true,
		"text/markdown": true,
	}
	return docTypes[da.MimeType]
}

// IsVideo returns true if the attachment is a video
func (da *DocumentAttachment) IsVideo() bool {
	videoTypes := map[string]bool{
		"video/mp4": true,
		"video/mpeg": true,
		"video/webm": true,
		"video/quicktime": true,
	}
	return videoTypes[da.MimeType]
}

// GetHumanReadableSize returns the file size in human-readable format
func (da *DocumentAttachment) GetHumanReadableSize() string {
	const unit = 1024
	if da.SizeBytes < unit {
		return fmt.Sprintf("%d B", da.SizeBytes)
	}
	div, exp := int64(unit), 0
	for n := da.SizeBytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%d %s", da.SizeBytes/int(div), []string{"KB", "MB", "GB", "TB", "PB", "EB"}[exp])
}

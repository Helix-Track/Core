package models

import (
	"errors"
	"time"
)

// Document represents the main document entity
type Document struct {
	ID          string  `json:"id" db:"id"`
	Title       string  `json:"title" db:"title"`
	SpaceID     string  `json:"space_id" db:"space_id"`
	ParentID    *string `json:"parent_id,omitempty" db:"parent_id"`        // For hierarchy
	TypeID      string  `json:"type_id" db:"type_id"`
	ProjectID   *string `json:"project_id,omitempty" db:"project_id"`      // Optional link to project
	CreatorID   string  `json:"creator_id" db:"creator_id"`
	Version     int     `json:"version" db:"version"`                      // Current version (optimistic locking)
	Position    int     `json:"position" db:"position"`                    // Position in hierarchy
	IsPublished bool    `json:"is_published" db:"is_published"`
	IsArchived  bool    `json:"is_archived" db:"is_archived"`
	PublishDate *int64  `json:"publish_date,omitempty" db:"publish_date"`  // Scheduled or actual publish date
	Created     int64   `json:"created" db:"created"`
	Modified    int64   `json:"modified" db:"modified"`
	Deleted     bool    `json:"deleted" db:"deleted"`
}

// Validate validates the document
func (d *Document) Validate() error {
	if d.ID == "" {
		return errors.New("document ID cannot be empty")
	}
	if d.Title == "" {
		return errors.New("document title cannot be empty")
	}
	if d.SpaceID == "" {
		return errors.New("document space ID cannot be empty")
	}
	if d.TypeID == "" {
		return errors.New("document type ID cannot be empty")
	}
	if d.CreatorID == "" {
		return errors.New("document creator ID cannot be empty")
	}
	if d.Version < 1 {
		return errors.New("document version must be at least 1")
	}
	if d.Created == 0 {
		return errors.New("document created timestamp cannot be zero")
	}
	if d.Modified == 0 {
		return errors.New("document modified timestamp cannot be zero")
	}
	return nil
}

// SetTimestamps sets created and modified timestamps
func (d *Document) SetTimestamps() {
	now := time.Now().Unix()
	if d.Created == 0 {
		d.Created = now
	}
	d.Modified = now
}

// IncrementVersion increments the version number for optimistic locking
func (d *Document) IncrementVersion() {
	d.Version++
	d.Modified = time.Now().Unix()
}

// DocumentContent represents the content of a document version
type DocumentContent struct {
	ID          string  `json:"id" db:"id"`
	DocumentID  string  `json:"document_id" db:"document_id"`
	Version     int     `json:"version" db:"version"`
	ContentType string  `json:"content_type" db:"content_type"` // "html", "markdown", "plain", "storage"
	Content     *string `json:"content,omitempty" db:"content"`
	ContentHash *string `json:"content_hash,omitempty" db:"content_hash"` // SHA-256 hash
	SizeBytes   int     `json:"size_bytes" db:"size_bytes"`
	Created     int64   `json:"created" db:"created"`
	Modified    int64   `json:"modified" db:"modified"`
	Deleted     bool    `json:"deleted" db:"deleted"`
}

// Validate validates the document content
func (dc *DocumentContent) Validate() error {
	if dc.ID == "" {
		return errors.New("document content ID cannot be empty")
	}
	if dc.DocumentID == "" {
		return errors.New("document content document ID cannot be empty")
	}
	if dc.Version < 1 {
		return errors.New("document content version must be at least 1")
	}
	if dc.ContentType == "" {
		return errors.New("document content type cannot be empty")
	}
	validTypes := map[string]bool{
		"html": true, "markdown": true, "plain": true, "storage": true,
	}
	if !validTypes[dc.ContentType] {
		return errors.New("invalid content type")
	}
	if dc.Created == 0 {
		return errors.New("document content created timestamp cannot be zero")
	}
	if dc.Modified == 0 {
		return errors.New("document content modified timestamp cannot be zero")
	}
	return nil
}

// SetTimestamps sets created and modified timestamps
func (dc *DocumentContent) SetTimestamps() {
	now := time.Now().Unix()
	if dc.Created == 0 {
		dc.Created = now
	}
	dc.Modified = now
}

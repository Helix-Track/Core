package models

import (
	"errors"
	"time"
)

// DocumentSpace represents a document space (similar to Confluence spaces)
// Spaces organize documents into logical groups
type DocumentSpace struct {
	ID          string `json:"id" db:"id"`
	Key         string `json:"key" db:"key"`                   // Short identifier (e.g., "DOCS", "TECH")
	Name        string `json:"name" db:"name"`
	Description string `json:"description,omitempty" db:"description"`
	OwnerID     string `json:"owner_id" db:"owner_id"`         // User who owns the space
	IsPublic    bool   `json:"is_public" db:"is_public"`       // Public or private space
	Created     int64  `json:"created" db:"created"`
	Modified    int64  `json:"modified" db:"modified"`
	Deleted     bool   `json:"deleted" db:"deleted"`
}

// Validate validates the document space
func (ds *DocumentSpace) Validate() error {
	if ds.ID == "" {
		return errors.New("document space ID cannot be empty")
	}
	if ds.Key == "" {
		return errors.New("document space key cannot be empty")
	}
	if ds.Name == "" {
		return errors.New("document space name cannot be empty")
	}
	if ds.OwnerID == "" {
		return errors.New("document space owner ID cannot be empty")
	}
	if ds.Created == 0 {
		return errors.New("document space created timestamp cannot be zero")
	}
	if ds.Modified == 0 {
		return errors.New("document space modified timestamp cannot be zero")
	}
	return nil
}

// SetTimestamps sets created and modified timestamps
func (ds *DocumentSpace) SetTimestamps() {
	now := time.Now().Unix()
	if ds.Created == 0 {
		ds.Created = now
	}
	ds.Modified = now
}

// DocumentType represents a document type (page, blog post, template, etc.)
type DocumentType struct {
	ID          string `json:"id" db:"id"`
	Key         string `json:"key" db:"key"`                   // "page", "blog", "template", etc.
	Name        string `json:"name" db:"name"`
	Description string `json:"description,omitempty" db:"description"`
	Icon        string `json:"icon,omitempty" db:"icon"`       // Icon identifier
	SchemaJSON  string `json:"schema_json,omitempty" db:"schema_json"` // JSON schema for this type
	Created     int64  `json:"created" db:"created"`
	Modified    int64  `json:"modified" db:"modified"`
	Deleted     bool   `json:"deleted" db:"deleted"`
}

// Validate validates the document type
func (dt *DocumentType) Validate() error {
	if dt.ID == "" {
		return errors.New("document type ID cannot be empty")
	}
	if dt.Key == "" {
		return errors.New("document type key cannot be empty")
	}
	if dt.Name == "" {
		return errors.New("document type name cannot be empty")
	}
	if dt.Created == 0 {
		return errors.New("document type created timestamp cannot be zero")
	}
	if dt.Modified == 0 {
		return errors.New("document type modified timestamp cannot be zero")
	}
	return nil
}

// SetTimestamps sets created and modified timestamps
func (dt *DocumentType) SetTimestamps() {
	now := time.Now().Unix()
	if dt.Created == 0 {
		dt.Created = now
	}
	dt.Modified = now
}

package models

import (
	"errors"
	"time"
)

// DocumentTagMapping represents a tag-to-document mapping
type DocumentTagMapping struct {
	ID         string `json:"id" db:"id"`
	DocumentID string `json:"document_id" db:"document_id"`
	TagID      string `json:"tag_id" db:"tag_id"`
	UserID     string `json:"user_id" db:"user_id"`
	Created    int64  `json:"created" db:"created"`
	Deleted    bool   `json:"deleted" db:"deleted"`
}

// Validate validates the tag mapping
func (dtm *DocumentTagMapping) Validate() error {
	if dtm.ID == "" {
		return errors.New("tag mapping ID cannot be empty")
	}
	if dtm.DocumentID == "" {
		return errors.New("tag mapping document ID cannot be empty")
	}
	if dtm.TagID == "" {
		return errors.New("tag mapping tag ID cannot be empty")
	}
	if dtm.UserID == "" {
		return errors.New("tag mapping user ID cannot be empty")
	}
	if dtm.Created == 0 {
		return errors.New("tag mapping created timestamp cannot be zero")
	}
	return nil
}

// SetTimestamps sets the created timestamp
func (dtm *DocumentTagMapping) SetTimestamps() {
	if dtm.Created == 0 {
		dtm.Created = time.Now().Unix()
	}
}

// DocumentEntityLink represents a link from a document to any system entity
type DocumentEntityLink struct {
	ID          string  `json:"id" db:"id"`
	DocumentID  string  `json:"document_id" db:"document_id"`
	EntityType  string  `json:"entity_type" db:"entity_type"` // "ticket", "project", "user", etc.
	EntityID    string  `json:"entity_id" db:"entity_id"`
	LinkType    string  `json:"link_type" db:"link_type"`
	Description *string `json:"description,omitempty" db:"description"`
	UserID      string  `json:"user_id" db:"user_id"`
	Created     int64   `json:"created" db:"created"`
	Deleted     bool    `json:"deleted" db:"deleted"`
}

// Validate validates the entity link
func (del *DocumentEntityLink) Validate() error {
	if del.ID == "" {
		return errors.New("entity link ID cannot be empty")
	}
	if del.DocumentID == "" {
		return errors.New("entity link document ID cannot be empty")
	}
	if del.EntityType == "" {
		return errors.New("entity link entity type cannot be empty")
	}
	if del.EntityID == "" {
		return errors.New("entity link entity ID cannot be empty")
	}
	if del.LinkType == "" {
		return errors.New("entity link link type cannot be empty")
	}
	if del.UserID == "" {
		return errors.New("entity link user ID cannot be empty")
	}
	if del.Created == 0 {
		return errors.New("entity link created timestamp cannot be zero")
	}
	return nil
}

// SetTimestamps sets the created timestamp
func (del *DocumentEntityLink) SetTimestamps() {
	if del.Created == 0 {
		del.Created = time.Now().Unix()
	}
}

// DocumentRelationship represents a relationship between two documents
type DocumentRelationship struct {
	ID               string `json:"id" db:"id"`
	SourceDocumentID string `json:"source_document_id" db:"source_document_id"`
	TargetDocumentID string `json:"target_document_id" db:"target_document_id"`
	RelationshipType string `json:"relationship_type" db:"relationship_type"`
	UserID           string `json:"user_id" db:"user_id"`
	Created          int64  `json:"created" db:"created"`
	Deleted          bool   `json:"deleted" db:"deleted"`
}

// Validate validates the document relationship
func (dr *DocumentRelationship) Validate() error {
	if dr.ID == "" {
		return errors.New("relationship ID cannot be empty")
	}
	if dr.SourceDocumentID == "" {
		return errors.New("relationship source document ID cannot be empty")
	}
	if dr.TargetDocumentID == "" {
		return errors.New("relationship target document ID cannot be empty")
	}
	if dr.SourceDocumentID == dr.TargetDocumentID {
		return errors.New("source and target document cannot be the same")
	}
	if dr.RelationshipType == "" {
		return errors.New("relationship type cannot be empty")
	}
	if dr.UserID == "" {
		return errors.New("relationship user ID cannot be empty")
	}
	if dr.Created == 0 {
		return errors.New("relationship created timestamp cannot be zero")
	}
	return nil
}

// SetTimestamps sets the created timestamp
func (dr *DocumentRelationship) SetTimestamps() {
	if dr.Created == 0 {
		dr.Created = time.Now().Unix()
	}
}

package models

import (
	"errors"
	"time"
)

// CommentDocumentMapping links core comments to documents
// This allows the core comment system to be used for documents
type CommentDocumentMapping struct {
	ID         string `json:"id" db:"id"`
	CommentID  string `json:"comment_id" db:"comment_id"`
	DocumentID string `json:"document_id" db:"document_id"`
	UserID     string `json:"user_id" db:"user_id"`
	IsResolved bool   `json:"is_resolved" db:"is_resolved"`
	Created    int64  `json:"created" db:"created"`
	Deleted    bool   `json:"deleted" db:"deleted"`
}

// Validate validates the comment-document mapping
func (cdm *CommentDocumentMapping) Validate() error {
	if cdm.ID == "" {
		return errors.New("comment-document mapping ID cannot be empty")
	}
	if cdm.CommentID == "" {
		return errors.New("comment-document mapping comment ID cannot be empty")
	}
	if cdm.DocumentID == "" {
		return errors.New("comment-document mapping document ID cannot be empty")
	}
	if cdm.UserID == "" {
		return errors.New("comment-document mapping user ID cannot be empty")
	}
	if cdm.Created == 0 {
		return errors.New("comment-document mapping created timestamp cannot be zero")
	}
	return nil
}

// SetTimestamps sets the created timestamp
func (cdm *CommentDocumentMapping) SetTimestamps() {
	if cdm.Created == 0 {
		cdm.Created = time.Now().Unix()
	}
}

// LabelDocumentMapping links core labels to documents
// This allows the core label system to be used for documents
type LabelDocumentMapping struct {
	ID         string `json:"id" db:"id"`
	LabelID    string `json:"label_id" db:"label_id"`
	DocumentID string `json:"document_id" db:"document_id"`
	UserID     string `json:"user_id" db:"user_id"`
	Created    int64  `json:"created" db:"created"`
	Deleted    bool   `json:"deleted" db:"deleted"`
}

// Validate validates the label-document mapping
func (ldm *LabelDocumentMapping) Validate() error {
	if ldm.ID == "" {
		return errors.New("label-document mapping ID cannot be empty")
	}
	if ldm.LabelID == "" {
		return errors.New("label-document mapping label ID cannot be empty")
	}
	if ldm.DocumentID == "" {
		return errors.New("label-document mapping document ID cannot be empty")
	}
	if ldm.UserID == "" {
		return errors.New("label-document mapping user ID cannot be empty")
	}
	if ldm.Created == 0 {
		return errors.New("label-document mapping created timestamp cannot be zero")
	}
	return nil
}

// SetTimestamps sets the created timestamp
func (ldm *LabelDocumentMapping) SetTimestamps() {
	if ldm.Created == 0 {
		ldm.Created = time.Now().Unix()
	}
}

// VoteMapping represents a generic vote/reaction on any entity
// Replaces the old ticket_vote_mapping with a universal system
type VoteMapping struct {
	ID         string  `json:"id" db:"id"`
	EntityType string  `json:"entity_type" db:"entity_type"` // "ticket", "document", "comment", etc.
	EntityID   string  `json:"entity_id" db:"entity_id"`
	UserID     string  `json:"user_id" db:"user_id"`
	VoteType   string  `json:"vote_type" db:"vote_type"` // "upvote", "downvote", "like", "love", etc.
	Emoji      *string `json:"emoji,omitempty" db:"emoji"`
	Created    int64   `json:"created" db:"created"`
	Deleted    bool    `json:"deleted" db:"deleted"`
}

// Validate validates the vote mapping
func (vm *VoteMapping) Validate() error {
	if vm.ID == "" {
		return errors.New("vote mapping ID cannot be empty")
	}
	if vm.EntityType == "" {
		return errors.New("vote mapping entity type cannot be empty")
	}
	if vm.EntityID == "" {
		return errors.New("vote mapping entity ID cannot be empty")
	}
	if vm.UserID == "" {
		return errors.New("vote mapping user ID cannot be empty")
	}
	if vm.VoteType == "" {
		return errors.New("vote mapping vote type cannot be empty")
	}
	validTypes := map[string]bool{
		"upvote": true, "downvote": true, "like": true, "love": true,
		"celebrate": true, "support": true, "insightful": true,
	}
	if !validTypes[vm.VoteType] {
		return errors.New("invalid vote type")
	}
	if vm.Created == 0 {
		return errors.New("vote mapping created timestamp cannot be zero")
	}
	return nil
}

// SetTimestamps sets the created timestamp
func (vm *VoteMapping) SetTimestamps() {
	if vm.Created == 0 {
		vm.Created = time.Now().Unix()
	}
}

// IsPositive returns true for positive vote types
func (vm *VoteMapping) IsPositive() bool {
	positiveTypes := map[string]bool{
		"upvote": true, "like": true, "love": true,
		"celebrate": true, "support": true, "insightful": true,
	}
	return positiveTypes[vm.VoteType]
}

// IsNegative returns true for negative vote types
func (vm *VoteMapping) IsNegative() bool {
	return vm.VoteType == "downvote"
}

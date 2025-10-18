package models

import (
	"errors"
	"time"
)

// DocumentVersion represents a document version in the history
type DocumentVersion struct {
	ID            string  `json:"id" db:"id"`
	DocumentID    string  `json:"document_id" db:"document_id"`
	VersionNumber int     `json:"version_number" db:"version_number"`
	UserID        string  `json:"user_id" db:"user_id"`
	ChangeSummary *string `json:"change_summary,omitempty" db:"change_summary"`
	IsMajor       bool    `json:"is_major" db:"is_major"`
	IsMinor       bool    `json:"is_minor" db:"is_minor"`
	SnapshotJSON  *string `json:"snapshot_json,omitempty" db:"snapshot_json"` // Full snapshot
	ContentID     *string `json:"content_id,omitempty" db:"content_id"`
	Created       int64   `json:"created" db:"created"`
}

// Validate validates the document version
func (dv *DocumentVersion) Validate() error {
	if dv.ID == "" {
		return errors.New("document version ID cannot be empty")
	}
	if dv.DocumentID == "" {
		return errors.New("document version document ID cannot be empty")
	}
	if dv.VersionNumber < 1 {
		return errors.New("version number must be at least 1")
	}
	if dv.UserID == "" {
		return errors.New("document version user ID cannot be empty")
	}
	if dv.Created == 0 {
		return errors.New("document version created timestamp cannot be zero")
	}
	return nil
}

// SetTimestamps sets the created timestamp
func (dv *DocumentVersion) SetTimestamps() {
	if dv.Created == 0 {
		dv.Created = time.Now().Unix()
	}
}

// DocumentVersionLabel represents a label for a version
type DocumentVersionLabel struct {
	ID          string  `json:"id" db:"id"`
	VersionID   string  `json:"version_id" db:"version_id"`
	Label       string  `json:"label" db:"label"`
	Description *string `json:"description,omitempty" db:"description"`
	UserID      string  `json:"user_id" db:"user_id"`
	Created     int64   `json:"created" db:"created"`
}

// Validate validates the version label
func (dvl *DocumentVersionLabel) Validate() error {
	if dvl.ID == "" {
		return errors.New("version label ID cannot be empty")
	}
	if dvl.VersionID == "" {
		return errors.New("version label version ID cannot be empty")
	}
	if dvl.Label == "" {
		return errors.New("version label cannot be empty")
	}
	if dvl.UserID == "" {
		return errors.New("version label user ID cannot be empty")
	}
	if dvl.Created == 0 {
		return errors.New("version label created timestamp cannot be zero")
	}
	return nil
}

// SetTimestamps sets the created timestamp
func (dvl *DocumentVersionLabel) SetTimestamps() {
	if dvl.Created == 0 {
		dvl.Created = time.Now().Unix()
	}
}

// DocumentVersionTag represents a tag for a version
type DocumentVersionTag struct {
	ID        string `json:"id" db:"id"`
	VersionID string `json:"version_id" db:"version_id"`
	Tag       string `json:"tag" db:"tag"`
	UserID    string `json:"user_id" db:"user_id"`
	Created   int64  `json:"created" db:"created"`
}

// Validate validates the version tag
func (dvt *DocumentVersionTag) Validate() error {
	if dvt.ID == "" {
		return errors.New("version tag ID cannot be empty")
	}
	if dvt.VersionID == "" {
		return errors.New("version tag version ID cannot be empty")
	}
	if dvt.Tag == "" {
		return errors.New("version tag cannot be empty")
	}
	if dvt.UserID == "" {
		return errors.New("version tag user ID cannot be empty")
	}
	if dvt.Created == 0 {
		return errors.New("version tag created timestamp cannot be zero")
	}
	return nil
}

// SetTimestamps sets the created timestamp
func (dvt *DocumentVersionTag) SetTimestamps() {
	if dvt.Created == 0 {
		dvt.Created = time.Now().Unix()
	}
}

// DocumentVersionComment represents a comment on a version
type DocumentVersionComment struct {
	ID        string `json:"id" db:"id"`
	VersionID string `json:"version_id" db:"version_id"`
	UserID    string `json:"user_id" db:"user_id"`
	Comment   string `json:"comment" db:"comment"`
	Created   int64  `json:"created" db:"created"`
	Modified  int64  `json:"modified" db:"modified"`
	Deleted   bool   `json:"deleted" db:"deleted"`
}

// Validate validates the version comment
func (dvc *DocumentVersionComment) Validate() error {
	if dvc.ID == "" {
		return errors.New("version comment ID cannot be empty")
	}
	if dvc.VersionID == "" {
		return errors.New("version comment version ID cannot be empty")
	}
	if dvc.UserID == "" {
		return errors.New("version comment user ID cannot be empty")
	}
	if dvc.Comment == "" {
		return errors.New("version comment cannot be empty")
	}
	if dvc.Created == 0 {
		return errors.New("version comment created timestamp cannot be zero")
	}
	if dvc.Modified == 0 {
		return errors.New("version comment modified timestamp cannot be zero")
	}
	return nil
}

// SetTimestamps sets created and modified timestamps
func (dvc *DocumentVersionComment) SetTimestamps() {
	now := time.Now().Unix()
	if dvc.Created == 0 {
		dvc.Created = now
	}
	dvc.Modified = now
}

// DocumentVersionMention represents a user mention in a version
type DocumentVersionMention struct {
	ID                string  `json:"id" db:"id"`
	VersionID         string  `json:"version_id" db:"version_id"`
	MentionedUserID   string  `json:"mentioned_user_id" db:"mentioned_user_id"`
	MentioningUserID  string  `json:"mentioning_user_id" db:"mentioning_user_id"`
	Context           *string `json:"context,omitempty" db:"context"`
	Created           int64   `json:"created" db:"created"`
}

// Validate validates the version mention
func (dvm *DocumentVersionMention) Validate() error {
	if dvm.ID == "" {
		return errors.New("version mention ID cannot be empty")
	}
	if dvm.VersionID == "" {
		return errors.New("version mention version ID cannot be empty")
	}
	if dvm.MentionedUserID == "" {
		return errors.New("mentioned user ID cannot be empty")
	}
	if dvm.MentioningUserID == "" {
		return errors.New("mentioning user ID cannot be empty")
	}
	if dvm.Created == 0 {
		return errors.New("version mention created timestamp cannot be zero")
	}
	return nil
}

// SetTimestamps sets the created timestamp
func (dvm *DocumentVersionMention) SetTimestamps() {
	if dvm.Created == 0 {
		dvm.Created = time.Now().Unix()
	}
}

// DocumentVersionDiff represents a cached diff between versions
type DocumentVersionDiff struct {
	ID          string `json:"id" db:"id"`
	DocumentID  string `json:"document_id" db:"document_id"`
	FromVersion int    `json:"from_version" db:"from_version"`
	ToVersion   int    `json:"to_version" db:"to_version"`
	DiffType    string `json:"diff_type" db:"diff_type"` // "unified", "split", "html"
	DiffContent string `json:"diff_content" db:"diff_content"`
	Created     int64  `json:"created" db:"created"`
}

// Validate validates the version diff
func (dvd *DocumentVersionDiff) Validate() error {
	if dvd.ID == "" {
		return errors.New("version diff ID cannot be empty")
	}
	if dvd.DocumentID == "" {
		return errors.New("version diff document ID cannot be empty")
	}
	if dvd.FromVersion < 1 {
		return errors.New("from version must be at least 1")
	}
	if dvd.ToVersion < 1 {
		return errors.New("to version must be at least 1")
	}
	if dvd.FromVersion >= dvd.ToVersion {
		return errors.New("from version must be less than to version")
	}
	if dvd.DiffType == "" {
		return errors.New("diff type cannot be empty")
	}
	validTypes := map[string]bool{
		"unified": true, "split": true, "html": true,
	}
	if !validTypes[dvd.DiffType] {
		return errors.New("invalid diff type")
	}
	if dvd.DiffContent == "" {
		return errors.New("diff content cannot be empty")
	}
	if dvd.Created == 0 {
		return errors.New("version diff created timestamp cannot be zero")
	}
	return nil
}

// SetTimestamps sets the created timestamp
func (dvd *DocumentVersionDiff) SetTimestamps() {
	if dvd.Created == 0 {
		dvd.Created = time.Now().Unix()
	}
}

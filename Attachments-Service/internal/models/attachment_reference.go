package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// AttachmentReference represents a logical reference linking an entity to a file
type AttachmentReference struct {
	ID          string   `json:"id" db:"id"`
	FileHash    string   `json:"file_hash" db:"file_hash"`
	EntityType  string   `json:"entity_type" db:"entity_type"`
	EntityID    string   `json:"entity_id" db:"entity_id"`
	Filename    string   `json:"filename" db:"filename"`
	Description *string  `json:"description,omitempty" db:"description"`
	UploaderID  string   `json:"uploader_id" db:"uploader_id"`
	Version     int      `json:"version" db:"version"`
	Tags        []string `json:"tags,omitempty" db:"tags"`
	Created     int64    `json:"created" db:"created"`
	Modified    int64    `json:"modified" db:"modified"`
	Deleted     bool     `json:"deleted" db:"deleted"`
}

// Valid entity types
const (
	EntityTypeTicket   = "ticket"
	EntityTypeDocument = "document"
	EntityTypeComment  = "comment"
	EntityTypeProject  = "project"
	EntityTypeTeam     = "team"
	EntityTypeUser     = "user"
	EntityTypeEpic     = "epic"
	EntityTypeStory    = "story"
	EntityTypeTask     = "task"
)

// NewAttachmentReference creates a new attachment reference
func NewAttachmentReference(fileHash, entityType, entityID, filename, uploaderID string) *AttachmentReference {
	now := time.Now().Unix()
	return &AttachmentReference{
		ID:         uuid.New().String(),
		FileHash:   fileHash,
		EntityType: entityType,
		EntityID:   entityID,
		Filename:   sanitizeFilename(filename),
		UploaderID: uploaderID,
		Version:    1,
		Tags:       []string{},
		Created:    now,
		Modified:   now,
		Deleted:    false,
	}
}

// Validate validates the attachment reference
func (r *AttachmentReference) Validate() error {
	if r.ID == "" {
		return fmt.Errorf("id is required")
	}
	if r.FileHash == "" {
		return fmt.Errorf("file_hash is required")
	}
	if len(r.FileHash) != 64 {
		return fmt.Errorf("file_hash must be 64 characters (SHA-256)")
	}
	if r.EntityType == "" {
		return fmt.Errorf("entity_type is required")
	}
	if !isValidEntityType(r.EntityType) {
		return fmt.Errorf("invalid entity_type: %s", r.EntityType)
	}
	if r.EntityID == "" {
		return fmt.Errorf("entity_id is required")
	}
	if r.Filename == "" {
		return fmt.Errorf("filename is required")
	}
	if len(r.Filename) > 255 {
		return fmt.Errorf("filename too long (max 255 characters)")
	}
	if r.UploaderID == "" {
		return fmt.Errorf("uploader_id is required")
	}
	if r.Version < 1 {
		return fmt.Errorf("version must be >= 1")
	}
	if r.Created == 0 {
		return fmt.Errorf("created timestamp is required")
	}
	if r.Modified == 0 {
		return fmt.Errorf("modified timestamp is required")
	}
	return nil
}

// SetTimestamps sets the created and modified timestamps
func (r *AttachmentReference) SetTimestamps() {
	now := time.Now().Unix()
	if r.Created == 0 {
		r.Created = now
	}
	r.Modified = now
}

// IncrementVersion increments the version number
func (r *AttachmentReference) IncrementVersion() {
	r.Version++
	r.Modified = time.Now().Unix()
}

// AddTag adds a tag to the reference
func (r *AttachmentReference) AddTag(tag string) {
	tag = strings.TrimSpace(strings.ToLower(tag))
	if tag == "" {
		return
	}

	// Check if tag already exists
	for _, t := range r.Tags {
		if t == tag {
			return
		}
	}

	r.Tags = append(r.Tags, tag)
	r.Modified = time.Now().Unix()
}

// RemoveTag removes a tag from the reference
func (r *AttachmentReference) RemoveTag(tag string) {
	tag = strings.TrimSpace(strings.ToLower(tag))
	newTags := []string{}
	for _, t := range r.Tags {
		if t != tag {
			newTags = append(newTags, t)
		}
	}
	if len(newTags) != len(r.Tags) {
		r.Tags = newTags
		r.Modified = time.Now().Unix()
	}
}

// HasTag checks if the reference has a specific tag
func (r *AttachmentReference) HasTag(tag string) bool {
	tag = strings.TrimSpace(strings.ToLower(tag))
	for _, t := range r.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// GetFileExtension returns the file extension from filename
func (r *AttachmentReference) GetFileExtension() string {
	parts := strings.Split(r.Filename, ".")
	if len(parts) > 1 {
		return strings.ToLower(parts[len(parts)-1])
	}
	return ""
}

// SoftDelete marks the reference as deleted
func (r *AttachmentReference) SoftDelete() {
	r.Deleted = true
	r.Modified = time.Now().Unix()
}

// Restore restores a soft-deleted reference
func (r *AttachmentReference) Restore() {
	r.Deleted = false
	r.Modified = time.Now().Unix()
}

// UpdateFilename updates the filename
func (r *AttachmentReference) UpdateFilename(filename string) {
	r.Filename = sanitizeFilename(filename)
	r.Modified = time.Now().Unix()
}

// UpdateDescription updates the description
func (r *AttachmentReference) UpdateDescription(description string) {
	if description == "" {
		r.Description = nil
	} else {
		r.Description = &description
	}
	r.Modified = time.Now().Unix()
}

// isValidEntityType checks if the entity type is valid
func isValidEntityType(entityType string) bool {
	validTypes := []string{
		EntityTypeTicket, EntityTypeDocument, EntityTypeComment,
		EntityTypeProject, EntityTypeTeam, EntityTypeUser,
		EntityTypeEpic, EntityTypeStory, EntityTypeTask,
	}
	for _, t := range validTypes {
		if t == entityType {
			return true
		}
	}
	return false
}

// sanitizeFilename sanitizes a filename to prevent security issues
func sanitizeFilename(filename string) string {
	// Remove null bytes
	filename = strings.ReplaceAll(filename, "\x00", "")

	// Remove path separators
	filename = strings.ReplaceAll(filename, "/", "_")
	filename = strings.ReplaceAll(filename, "\\", "_")

	// Remove leading/trailing spaces and dots
	filename = strings.Trim(filename, " .")

	// Limit length
	if len(filename) > 255 {
		ext := ""
		parts := strings.Split(filename, ".")
		if len(parts) > 1 {
			ext = "." + parts[len(parts)-1]
		}
		filename = filename[:255-len(ext)] + ext
	}

	return filename
}

// GetValidEntityTypes returns all valid entity types
func GetValidEntityTypes() []string {
	return []string{
		EntityTypeTicket, EntityTypeDocument, EntityTypeComment,
		EntityTypeProject, EntityTypeTeam, EntityTypeUser,
		EntityTypeEpic, EntityTypeStory, EntityTypeTask,
	}
}

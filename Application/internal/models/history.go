package models

import (
	"fmt"
	"strings"
)

// EntityHistory represents a historical change to any editable entity
type EntityHistory struct {
	ID            string                 `json:"id" db:"id"`
	EntityID      string                 `json:"entityId" db:"entity_id"`
	EntityType    string                 `json:"entityType" db:"entity_type"`
	Version       int                    `json:"version" db:"version"`
	Action        string                 `json:"action" db:"action"`
	UserID        string                 `json:"userId" db:"user_id"`
	Timestamp     int64                  `json:"timestamp" db:"timestamp"`
	OldData       map[string]interface{} `json:"oldData,omitempty" db:"old_data"`
	NewData       map[string]interface{} `json:"newData,omitempty" db:"new_data"`
	ChangeSummary string                 `json:"changeSummary" db:"change_summary"`
	ConflictData  map[string]interface{} `json:"conflictData,omitempty" db:"conflict_data"`
}

// TicketHistory represents historical changes to tickets
type TicketHistory struct {
	ID            string                 `json:"id" db:"id"`
	TicketID      string                 `json:"ticketId" db:"ticket_id"`
	Version       int                    `json:"version" db:"version"`
	Action        string                 `json:"action" db:"action"`
	UserID        string                 `json:"userId" db:"user_id"`
	Timestamp     int64                  `json:"timestamp" db:"timestamp"`
	OldData       map[string]interface{} `json:"oldData,omitempty" db:"old_data"`
	NewData       map[string]interface{} `json:"newData,omitempty" db:"new_data"`
	ChangeSummary string                 `json:"changeSummary" db:"change_summary"`
	ConflictData  map[string]interface{} `json:"conflictData,omitempty" db:"conflict_data"`
}

// ProjectHistory represents historical changes to projects
type ProjectHistory struct {
	ID            string                 `json:"id" db:"id"`
	ProjectID     string                 `json:"projectId" db:"project_id"`
	Version       int                    `json:"version" db:"version"`
	Action        string                 `json:"action" db:"action"`
	UserID        string                 `json:"userId" db:"user_id"`
	Timestamp     int64                  `json:"timestamp" db:"timestamp"`
	OldData       map[string]interface{} `json:"oldData,omitempty" db:"old_data"`
	NewData       map[string]interface{} `json:"newData,omitempty" db:"new_data"`
	ChangeSummary string                 `json:"changeSummary" db:"change_summary"`
	ConflictData  map[string]interface{} `json:"conflictData,omitempty" db:"conflict_data"`
}

// CommentHistory represents historical changes to comments
type CommentHistory struct {
	ID            string                 `json:"id" db:"id"`
	CommentID     string                 `json:"commentId" db:"comment_id"`
	Version       int                    `json:"version" db:"version"`
	Action        string                 `json:"action" db:"action"`
	UserID        string                 `json:"userId" db:"user_id"`
	Timestamp     int64                  `json:"timestamp" db:"timestamp"`
	OldData       map[string]interface{} `json:"oldData,omitempty" db:"old_data"`
	NewData       map[string]interface{} `json:"newData,omitempty" db:"new_data"`
	ChangeSummary string                 `json:"changeSummary" db:"change_summary"`
	ConflictData  map[string]interface{} `json:"conflictData,omitempty" db:"conflict_data"`
}

// DashboardHistory represents historical changes to dashboards
type DashboardHistory struct {
	ID            string                 `json:"id" db:"id"`
	DashboardID   string                 `json:"dashboardId" db:"dashboard_id"`
	Version       int                    `json:"version" db:"version"`
	Action        string                 `json:"action" db:"action"`
	UserID        string                 `json:"userId" db:"user_id"`
	Timestamp     int64                  `json:"timestamp" db:"timestamp"`
	OldData       map[string]interface{} `json:"oldData,omitempty" db:"old_data"`
	NewData       map[string]interface{} `json:"newData,omitempty" db:"new_data"`
	ChangeSummary string                 `json:"changeSummary" db:"change_summary"`
	ConflictData  map[string]interface{} `json:"conflictData,omitempty" db:"conflict_data"`
}

// BoardHistory represents historical changes to boards
type BoardHistory struct {
	ID            string                 `json:"id" db:"id"`
	BoardID       string                 `json:"boardId" db:"board_id"`
	Version       int                    `json:"version" db:"version"`
	Action        string                 `json:"action" db:"action"`
	UserID        string                 `json:"userId" db:"user_id"`
	Timestamp     int64                  `json:"timestamp" db:"timestamp"`
	OldData       map[string]interface{} `json:"oldData,omitempty" db:"old_data"`
	NewData       map[string]interface{} `json:"newData,omitempty" db:"new_data"`
	ChangeSummary string                 `json:"changeSummary" db:"change_summary"`
	ConflictData  map[string]interface{} `json:"conflictData,omitempty" db:"conflict_data"`
}

// EntityLock represents a lock on an entity for collaborative editing
type EntityLock struct {
	ID         string                 `json:"id" db:"id"`
	EntityType string                 `json:"entityType" db:"entity_type"`
	EntityID   string                 `json:"entityId" db:"entity_id"`
	UserID     string                 `json:"userId" db:"user_id"`
	LockType   string                 `json:"lockType" db:"lock_type"` // 'optimistic', 'pessimistic'
	AcquiredAt int64                  `json:"acquiredAt" db:"acquired_at"`
	ExpiresAt  *int64                 `json:"expiresAt,omitempty" db:"expires_at"` // NULL for optimistic locks
	Metadata   map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
}

// ConflictResolution represents how to resolve version conflicts
type ConflictResolution string

const (
	ConflictOverwrite ConflictResolution = "overwrite" // Overwrite with new changes
	ConflictMerge     ConflictResolution = "merge"     // Attempt to merge changes
	ConflictCancel    ConflictResolution = "cancel"    // Cancel the operation
)

// VersionConflictError represents a version conflict during optimistic locking
type VersionConflictError struct {
	EntityType      string
	EntityID        string
	ExpectedVersion int
	CurrentVersion  int
	CurrentData     map[string]interface{}
}

func (e *VersionConflictError) Error() string {
	return fmt.Sprintf("version conflict on %s %s: expected version %d, got %d",
		e.EntityType, e.EntityID, e.ExpectedVersion, e.CurrentVersion)
}

// NewVersionConflictError creates a new version conflict error
func NewVersionConflictError(entityType, entityID string, expectedVersion, currentVersion int, currentData map[string]interface{}) *VersionConflictError {
	return &VersionConflictError{
		EntityType:      entityType,
		EntityID:        entityID,
		ExpectedVersion: expectedVersion,
		CurrentVersion:  currentVersion,
		CurrentData:     currentData,
	}
}

// GenerateChangeSummary creates a human-readable summary of changes
func GenerateChangeSummary(action string, oldData, newData map[string]interface{}) string {
	switch action {
	case ActionCreate:
		return "Entity created"
	case ActionRemove:
		return "Entity deleted"
	case ActionModify:
		changes := []string{}
		for key, newVal := range newData {
			if oldVal, exists := oldData[key]; exists {
				if oldVal != newVal {
					changes = append(changes, fmt.Sprintf("%s changed", key))
				}
			} else {
				changes = append(changes, fmt.Sprintf("%s added", key))
			}
		}
		for key := range oldData {
			if _, exists := newData[key]; !exists {
				changes = append(changes, fmt.Sprintf("%s removed", key))
			}
		}
		if len(changes) == 0 {
			return "Entity updated (no field changes detected)"
		}
		return fmt.Sprintf("Entity updated: %s", strings.Join(changes, ", "))
	default:
		return "Unknown action"
	}
}

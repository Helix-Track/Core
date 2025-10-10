package models

import "time"

// Project represents a project in the system
type Project struct {
	ID          string `json:"id" db:"id"`
	Title       string `json:"title" db:"title"`
	Description string `json:"description" db:"description"`
	Identifier  string `json:"identifier" db:"identifier"` // Short project key like "PROJ"
	WorkflowID  string `json:"workflowId" db:"workflow_id"`
	Created     int64  `json:"created" db:"created"`
	Modified    int64  `json:"modified" db:"modified"`
	Deleted     bool   `json:"deleted" db:"deleted"`
}

// NewProject creates a new project with current timestamps
func NewProject(id, title, description, identifier, workflowID string) *Project {
	now := time.Now().Unix()
	return &Project{
		ID:          id,
		Title:       title,
		Description: description,
		Identifier:  identifier,
		WorkflowID:  workflowID,
		Created:     now,
		Modified:    now,
		Deleted:     false,
	}
}

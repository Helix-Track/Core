package models

import "time"

type Workflow struct {
	ID          string `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
	IsDefault   bool   `json:"isDefault" db:"is_default"`
	IsActive    bool   `json:"isActive" db:"is_active"`
	Created     int64  `json:"created" db:"created"`
	Modified    int64  `json:"modified" db:"modified"`
	Deleted     bool   `json:"deleted" db:"deleted"`
	Version     int    `json:"version" db:"version"`
}

func NewWorkflow(id, name, description string, isDefault, isActive bool) *Workflow {
	now := time.Now().Unix()
	return &Workflow{ID: id, Name: name, Description: description, IsDefault: isDefault, IsActive: isActive,
		Created: now, Modified: now, Deleted: false, Version: 1}
}

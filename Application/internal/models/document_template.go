package models

import (
	"errors"
	"time"
)

// DocumentTemplate represents a reusable document template
type DocumentTemplate struct {
	ID              string  `json:"id" db:"id"`
	Name            string  `json:"name" db:"name"`
	Description     *string `json:"description,omitempty" db:"description"`
	SpaceID         *string `json:"space_id,omitempty" db:"space_id"`
	TypeID          string  `json:"type_id" db:"type_id"`
	ContentTemplate string  `json:"content_template" db:"content_template"`
	VariablesJSON   *string `json:"variables_json,omitempty" db:"variables_json"`
	CreatorID       string  `json:"creator_id" db:"creator_id"`
	IsPublic        bool    `json:"is_public" db:"is_public"`
	UseCount        int     `json:"use_count" db:"use_count"`
	Created         int64   `json:"created" db:"created"`
	Modified        int64   `json:"modified" db:"modified"`
	Deleted         bool    `json:"deleted" db:"deleted"`
}

// Validate validates the document template
func (dt *DocumentTemplate) Validate() error {
	if dt.ID == "" {
		return errors.New("template ID cannot be empty")
	}
	if dt.Name == "" {
		return errors.New("template name cannot be empty")
	}
	if dt.TypeID == "" {
		return errors.New("template type ID cannot be empty")
	}
	if dt.ContentTemplate == "" {
		return errors.New("template content cannot be empty")
	}
	if dt.CreatorID == "" {
		return errors.New("template creator ID cannot be empty")
	}
	if dt.Created == 0 {
		return errors.New("template created timestamp cannot be zero")
	}
	if dt.Modified == 0 {
		return errors.New("template modified timestamp cannot be zero")
	}
	return nil
}

// SetTimestamps sets created and modified timestamps
func (dt *DocumentTemplate) SetTimestamps() {
	now := time.Now().Unix()
	if dt.Created == 0 {
		dt.Created = now
	}
	dt.Modified = now
}

// IncrementUseCount increments the use count
func (dt *DocumentTemplate) IncrementUseCount() {
	dt.UseCount++
}

// DocumentBlueprint represents a multi-step template wizard
type DocumentBlueprint struct {
	ID             string  `json:"id" db:"id"`
	Name           string  `json:"name" db:"name"`
	Description    *string `json:"description,omitempty" db:"description"`
	SpaceID        *string `json:"space_id,omitempty" db:"space_id"`
	TemplateID     string  `json:"template_id" db:"template_id"`
	WizardStepsJSON *string `json:"wizard_steps_json,omitempty" db:"wizard_steps_json"`
	CreatorID      string  `json:"creator_id" db:"creator_id"`
	IsPublic       bool    `json:"is_public" db:"is_public"`
	Created        int64   `json:"created" db:"created"`
	Modified       int64   `json:"modified" db:"modified"`
	Deleted        bool    `json:"deleted" db:"deleted"`
}

// Validate validates the document blueprint
func (db *DocumentBlueprint) Validate() error {
	if db.ID == "" {
		return errors.New("blueprint ID cannot be empty")
	}
	if db.Name == "" {
		return errors.New("blueprint name cannot be empty")
	}
	if db.TemplateID == "" {
		return errors.New("blueprint template ID cannot be empty")
	}
	if db.CreatorID == "" {
		return errors.New("blueprint creator ID cannot be empty")
	}
	if db.Created == 0 {
		return errors.New("blueprint created timestamp cannot be zero")
	}
	if db.Modified == 0 {
		return errors.New("blueprint modified timestamp cannot be zero")
	}
	return nil
}

// SetTimestamps sets created and modified timestamps
func (db *DocumentBlueprint) SetTimestamps() {
	now := time.Now().Unix()
	if db.Created == 0 {
		db.Created = now
	}
	db.Modified = now
}

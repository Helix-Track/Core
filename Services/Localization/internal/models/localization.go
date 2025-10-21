package models

import (
	"encoding/json"
	"time"
)

// Localization represents a localized string
type Localization struct {
	ID          string          `json:"id" db:"id"`
	KeyID       string          `json:"key_id" db:"key_id"`
	LanguageID  string          `json:"language_id" db:"language_id"`
	Value       string          `json:"value" db:"value"`
	PluralForms json.RawMessage `json:"plural_forms,omitempty" db:"plural_forms"`
	Variables   json.RawMessage `json:"variables,omitempty" db:"variables"`
	Version     int             `json:"version" db:"version"`
	Approved    bool            `json:"approved" db:"approved"`
	ApprovedBy  string          `json:"approved_by,omitempty" db:"approved_by"`
	ApprovedAt  int64           `json:"approved_at,omitempty" db:"approved_at"`
	CreatedAt   int64           `json:"created_at" db:"created_at"`
	ModifiedAt  int64           `json:"modified_at" db:"modified_at"`
	Deleted     bool            `json:"deleted" db:"deleted"`
}

// LocalizationWithDetails includes key and language information
type LocalizationWithDetails struct {
	Localization
	Key          string `json:"key" db:"key"`
	LanguageCode string `json:"language_code" db:"language_code"`
}

// Validate validates the localization model
func (l *Localization) Validate() error {
	if l.KeyID == "" {
		return ErrValidationFailed("key_id is required")
	}
	if l.LanguageID == "" {
		return ErrValidationFailed("language_id is required")
	}
	if l.Value == "" {
		return ErrValidationFailed("value is required")
	}
	if l.Version < 1 {
		l.Version = 1
	}
	return nil
}

// BeforeCreate sets timestamps before creation
func (l *Localization) BeforeCreate() {
	now := time.Now().Unix()
	l.CreatedAt = now
	l.ModifiedAt = now
	if l.ID == "" {
		l.ID = GenerateUUID()
	}
	if l.Version == 0 {
		l.Version = 1
	}
}

// BeforeUpdate sets modified timestamp before update
func (l *Localization) BeforeUpdate() {
	l.ModifiedAt = time.Now().Unix()
}

// Approve marks the localization as approved
func (l *Localization) Approve(username string) {
	l.Approved = true
	l.ApprovedBy = username
	l.ApprovedAt = time.Now().Unix()
}

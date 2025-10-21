package models

import (
	"time"
)

// LocalizationKey represents a localization key
type LocalizationKey struct {
	ID          string `json:"id" db:"id"`
	Key         string `json:"key" db:"key"`
	Category    string `json:"category" db:"category"`
	Description string `json:"description" db:"description"`
	Context     string `json:"context" db:"context"`
	CreatedAt   int64  `json:"created_at" db:"created_at"`
	ModifiedAt  int64  `json:"modified_at" db:"modified_at"`
	Deleted     bool   `json:"deleted" db:"deleted"`
}

// Validate validates the localization key model
func (lk *LocalizationKey) Validate() error {
	if lk.Key == "" {
		return ErrValidationFailed("key is required")
	}
	if len(lk.Key) > 255 {
		return ErrValidationFailed("key must be 255 characters or less")
	}
	if len(lk.Category) > 100 {
		return ErrValidationFailed("category must be 100 characters or less")
	}
	if len(lk.Context) > 255 {
		return ErrValidationFailed("context must be 255 characters or less")
	}
	return nil
}

// BeforeCreate sets timestamps before creation
func (lk *LocalizationKey) BeforeCreate() {
	now := time.Now().Unix()
	lk.CreatedAt = now
	lk.ModifiedAt = now
	if lk.ID == "" {
		lk.ID = GenerateUUID()
	}
}

// BeforeUpdate sets modified timestamp before update
func (lk *LocalizationKey) BeforeUpdate() {
	lk.ModifiedAt = time.Now().Unix()
}

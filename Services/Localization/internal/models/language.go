package models

import (
	"time"
)

// Language represents a supported language in the system
type Language struct {
	ID         string `json:"id" db:"id"`
	Code       string `json:"code" db:"code"`               // ISO 639-1
	Name       string `json:"name" db:"name"`               // English name
	NativeName string `json:"native_name" db:"native_name"` // Native name
	IsRTL      bool   `json:"is_rtl" db:"is_rtl"`           // Right-to-left
	IsActive   bool   `json:"is_active" db:"is_active"`
	IsDefault  bool   `json:"is_default" db:"is_default"`
	CreatedAt  int64  `json:"created_at" db:"created_at"`
	ModifiedAt int64  `json:"modified_at" db:"modified_at"`
	Deleted    bool   `json:"deleted" db:"deleted"`
}

// Validate validates the language model
func (l *Language) Validate() error {
	if l.Code == "" {
		return ErrValidationFailed("code is required")
	}
	if len(l.Code) > 10 {
		return ErrValidationFailed("code must be 10 characters or less")
	}
	if l.Name == "" {
		return ErrValidationFailed("name is required")
	}
	if len(l.Name) > 100 {
		return ErrValidationFailed("name must be 100 characters or less")
	}
	return nil
}

// BeforeCreate sets timestamps before creation
func (l *Language) BeforeCreate() {
	now := time.Now().Unix()
	l.CreatedAt = now
	l.ModifiedAt = now
	if l.ID == "" {
		l.ID = GenerateUUID()
	}
}

// BeforeUpdate sets modified timestamp before update
func (l *Language) BeforeUpdate() {
	l.ModifiedAt = time.Now().Unix()
}

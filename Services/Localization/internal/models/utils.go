package models

import (
	"github.com/google/uuid"
)

// GenerateUUID generates a new UUID v4
func GenerateUUID() string {
	return uuid.New().String()
}

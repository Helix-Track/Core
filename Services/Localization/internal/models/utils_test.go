package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateUUID(t *testing.T) {
	uuid1 := GenerateUUID()
	uuid2 := GenerateUUID()

	// UUIDs should not be empty
	assert.NotEmpty(t, uuid1)
	assert.NotEmpty(t, uuid2)

	// UUIDs should be unique
	assert.NotEqual(t, uuid1, uuid2)

	// UUIDs should have correct format (36 characters with hyphens)
	assert.Len(t, uuid1, 36)
	assert.Len(t, uuid2, 36)
	assert.Contains(t, uuid1, "-")
	assert.Contains(t, uuid2, "-")
}

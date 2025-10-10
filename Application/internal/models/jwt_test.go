package models

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestJWTClaims_Structure(t *testing.T) {
	now := time.Now()
	claims := &JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "authentication",
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		Name:          "John Doe",
		Username:      "johndoe",
		Role:          "admin",
		Permissions:   "all",
		HTCoreAddress: "http://localhost:8080",
	}

	assert.Equal(t, "authentication", claims.Subject)
	assert.Equal(t, "John Doe", claims.Name)
	assert.Equal(t, "johndoe", claims.Username)
	assert.Equal(t, "admin", claims.Role)
	assert.Equal(t, "all", claims.Permissions)
	assert.Equal(t, "http://localhost:8080", claims.HTCoreAddress)
}

func TestPermissionLevel_Constants(t *testing.T) {
	assert.Equal(t, PermissionLevel(1), PermissionRead)
	assert.Equal(t, PermissionLevel(2), PermissionCreate)
	assert.Equal(t, PermissionLevel(3), PermissionUpdate)
	assert.Equal(t, PermissionLevel(5), PermissionDelete)
}

func TestPermissionLevel_HasPermission(t *testing.T) {
	tests := []struct {
		name     string
		level    PermissionLevel
		required PermissionLevel
		expected bool
	}{
		{
			name:     "Delete has Read permission",
			level:    PermissionDelete,
			required: PermissionRead,
			expected: true,
		},
		{
			name:     "Delete has Create permission",
			level:    PermissionDelete,
			required: PermissionCreate,
			expected: true,
		},
		{
			name:     "Delete has Update permission",
			level:    PermissionDelete,
			required: PermissionUpdate,
			expected: true,
		},
		{
			name:     "Delete has Delete permission",
			level:    PermissionDelete,
			required: PermissionDelete,
			expected: true,
		},
		{
			name:     "Update has Read permission",
			level:    PermissionUpdate,
			required: PermissionRead,
			expected: true,
		},
		{
			name:     "Update has Create permission",
			level:    PermissionUpdate,
			required: PermissionCreate,
			expected: true,
		},
		{
			name:     "Update has Update permission",
			level:    PermissionUpdate,
			required: PermissionUpdate,
			expected: true,
		},
		{
			name:     "Update does not have Delete permission",
			level:    PermissionUpdate,
			required: PermissionDelete,
			expected: false,
		},
		{
			name:     "Create has Read permission",
			level:    PermissionCreate,
			required: PermissionRead,
			expected: true,
		},
		{
			name:     "Create has Create permission",
			level:    PermissionCreate,
			required: PermissionCreate,
			expected: true,
		},
		{
			name:     "Create does not have Update permission",
			level:    PermissionCreate,
			required: PermissionUpdate,
			expected: false,
		},
		{
			name:     "Create does not have Delete permission",
			level:    PermissionCreate,
			required: PermissionDelete,
			expected: false,
		},
		{
			name:     "Read has Read permission",
			level:    PermissionRead,
			required: PermissionRead,
			expected: true,
		},
		{
			name:     "Read does not have Create permission",
			level:    PermissionRead,
			required: PermissionCreate,
			expected: false,
		},
		{
			name:     "Read does not have Update permission",
			level:    PermissionRead,
			required: PermissionUpdate,
			expected: false,
		},
		{
			name:     "Read does not have Delete permission",
			level:    PermissionRead,
			required: PermissionDelete,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.level.HasPermission(tt.required)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPermission_Structure(t *testing.T) {
	perm := &Permission{
		Context: "organization/team/project",
		Level:   PermissionUpdate,
	}

	assert.Equal(t, "organization/team/project", perm.Context)
	assert.Equal(t, PermissionUpdate, perm.Level)
	assert.True(t, perm.Level.HasPermission(PermissionRead))
	assert.True(t, perm.Level.HasPermission(PermissionCreate))
	assert.True(t, perm.Level.HasPermission(PermissionUpdate))
	assert.False(t, perm.Level.HasPermission(PermissionDelete))
}

package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJWTClaims_IsAdmin(t *testing.T) {
	tests := []struct {
		name       string
		claims     JWTClaims
		adminRoles []string
		expected   bool
	}{
		{
			name: "is admin",
			claims: JWTClaims{
				Role: "admin",
			},
			adminRoles: []string{"admin", "superadmin"},
			expected:   true,
		},
		{
			name: "is superadmin",
			claims: JWTClaims{
				Role: "superadmin",
			},
			adminRoles: []string{"admin", "superadmin"},
			expected:   true,
		},
		{
			name: "is not admin",
			claims: JWTClaims{
				Role: "user",
			},
			adminRoles: []string{"admin", "superadmin"},
			expected:   false,
		},
		{
			name: "empty admin roles",
			claims: JWTClaims{
				Role: "admin",
			},
			adminRoles: []string{},
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.claims.IsAdmin(tt.adminRoles)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestJWTClaims_HasPermission(t *testing.T) {
	tests := []struct {
		name       string
		claims     JWTClaims
		permission string
		expected   bool
	}{
		{
			name: "has ALL permission",
			claims: JWTClaims{
				Permissions: "ALL",
			},
			permission: "READ",
			expected:   true,
		},
		{
			name: "has exact permission",
			claims: JWTClaims{
				Permissions: "READ",
			},
			permission: "READ",
			expected:   true,
		},
		{
			name: "does not have permission",
			claims: JWTClaims{
				Permissions: "READ",
			},
			permission: "WRITE",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.claims.HasPermission(tt.permission)
			assert.Equal(t, tt.expected, result)
		})
	}
}

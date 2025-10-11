package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequest_IsAuthenticationRequired(t *testing.T) {
	tests := []struct {
		name     string
		action   string
		expected bool
	}{
		{
			name:     "Version action does not require auth",
			action:   ActionVersion,
			expected: false,
		},
		{
			name:     "JWTCapable action does not require auth",
			action:   ActionJWTCapable,
			expected: false,
		},
		{
			name:     "DBCapable action does not require auth",
			action:   ActionDBCapable,
			expected: false,
		},
		{
			name:     "Health action does not require auth",
			action:   ActionHealth,
			expected: false,
		},
		{
			name:     "Authenticate action does not require auth",
			action:   ActionAuthenticate,
			expected: false,
		},
		{
			name:     "Create action requires auth",
			action:   ActionCreate,
			expected: true,
		},
		{
			name:     "Modify action requires auth",
			action:   ActionModify,
			expected: true,
		},
		{
			name:     "Remove action requires auth",
			action:   ActionRemove,
			expected: true,
		},
		{
			name:     "Read action requires auth",
			action:   ActionRead,
			expected: true,
		},
		{
			name:     "List action requires auth",
			action:   ActionList,
			expected: true,
		},
		{
			name:     "Unknown action requires auth",
			action:   "unknownAction",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &Request{Action: tt.action}
			assert.Equal(t, tt.expected, req.IsAuthenticationRequired())
		})
	}
}

func TestRequest_IsCRUDOperation(t *testing.T) {
	tests := []struct {
		name     string
		action   string
		expected bool
	}{
		{
			name:     "Create is CRUD operation",
			action:   ActionCreate,
			expected: true,
		},
		{
			name:     "Modify is CRUD operation",
			action:   ActionModify,
			expected: true,
		},
		{
			name:     "Remove is CRUD operation",
			action:   ActionRemove,
			expected: true,
		},
		{
			name:     "Read is not CRUD operation (requires object parameter)",
			action:   ActionRead,
			expected: false,
		},
		{
			name:     "List is not CRUD operation",
			action:   ActionList,
			expected: false,
		},
		{
			name:     "Version is not CRUD operation",
			action:   ActionVersion,
			expected: false,
		},
		{
			name:     "Health is not CRUD operation",
			action:   ActionHealth,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &Request{Action: tt.action}
			assert.Equal(t, tt.expected, req.IsCRUDOperation())
		})
	}
}

func TestRequest_Structure(t *testing.T) {
	req := &Request{
		Action: ActionCreate,
		JWT:    "test-jwt-token",
		Locale: "en_US",
		Object: "project",
		Data: map[string]interface{}{
			"name":        "Test Project",
			"description": "A test project",
		},
	}

	assert.Equal(t, ActionCreate, req.Action)
	assert.Equal(t, "test-jwt-token", req.JWT)
	assert.Equal(t, "en_US", req.Locale)
	assert.Equal(t, "project", req.Object)
	assert.NotNil(t, req.Data)
	assert.Equal(t, "Test Project", req.Data["name"])
	assert.Equal(t, "A test project", req.Data["description"])
}

func TestActionConstants(t *testing.T) {
	assert.Equal(t, "authenticate", ActionAuthenticate)
	assert.Equal(t, "version", ActionVersion)
	assert.Equal(t, "jwtCapable", ActionJWTCapable)
	assert.Equal(t, "dbCapable", ActionDBCapable)
	assert.Equal(t, "health", ActionHealth)
	assert.Equal(t, "create", ActionCreate)
	assert.Equal(t, "modify", ActionModify)
	assert.Equal(t, "remove", ActionRemove)
	assert.Equal(t, "read", ActionRead)
	assert.Equal(t, "list", ActionList)
}

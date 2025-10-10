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

// TestPermissionLevel_String tests the String method
func TestPermissionLevel_String(t *testing.T) {
	tests := []struct {
		name  string
		level PermissionLevel
		want  string
	}{
		{name: "None level", level: PermissionNone, want: "NONE"},
		{name: "Read level", level: PermissionRead, want: "READ"},
		{name: "Create level", level: PermissionCreate, want: "CREATE"},
		{name: "Update level", level: PermissionUpdate, want: "UPDATE"},
		{name: "Delete level", level: PermissionDelete, want: "DELETE"},
		{name: "Unknown level", level: PermissionLevel(99), want: "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.level.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestParsePermissionLevel tests the ParsePermissionLevel function
func TestParsePermissionLevel(t *testing.T) {
	tests := []struct {
		name  string
		level string
		want  PermissionLevel
	}{
		{name: "Parse READ", level: "READ", want: PermissionRead},
		{name: "Parse read (lowercase)", level: "read", want: PermissionRead},
		{name: "Parse CREATE", level: "CREATE", want: PermissionCreate},
		{name: "Parse UPDATE", level: "UPDATE", want: PermissionUpdate},
		{name: "Parse DELETE", level: "DELETE", want: PermissionDelete},
		{name: "Parse ALL", level: "ALL", want: PermissionDelete},
		{name: "Parse unknown", level: "INVALID", want: PermissionNone},
		{name: "Parse empty", level: "", want: PermissionNone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParsePermissionLevel(tt.level)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestBuildContextPath tests the BuildContextPath function
func TestBuildContextPath(t *testing.T) {
	tests := []struct {
		name     string
		contexts []string
		want     string
	}{
		{name: "Single context", contexts: []string{"node1"}, want: "node1"},
		{name: "Two contexts", contexts: []string{"node1", "account1"}, want: "node1→account1"},
		{name: "Full hierarchy", contexts: []string{"node1", "account1", "org1", "team1"}, want: "node1→account1→org1→team1"},
		{name: "Empty contexts", contexts: []string{}, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildContextPath(tt.contexts...)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestParseContextPath tests the ParseContextPath function
func TestParseContextPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want []string
	}{
		{name: "Single context", path: "node1", want: []string{"node1"}},
		{name: "Two contexts", path: "node1→account1", want: []string{"node1", "account1"}},
		{name: "Full hierarchy", path: "node1→account1→org1→team1", want: []string{"node1", "account1", "org1", "team1"}},
		{name: "Empty path", path: "", want: []string{""}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseContextPath(tt.path)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestIsParentContext tests the IsParentContext function
func TestIsParentContext(t *testing.T) {
	tests := []struct {
		name   string
		parent string
		child  string
		want   bool
	}{
		{name: "Direct parent", parent: "node1", child: "node1→account1", want: true},
		{name: "Grandparent", parent: "node1", child: "node1→account1→org1", want: true},
		{name: "Not a parent - different hierarchy", parent: "node2", child: "node1→account1", want: false},
		{name: "Not a parent - same level", parent: "node1→account1", child: "node1→account1", want: false},
		{name: "Not a parent - child is parent", parent: "node1→account1", child: "node1", want: false},
		{name: "Parent with multiple levels", parent: "node1→account1", child: "node1→account1→org1→team1", want: true},
		{name: "Not a parent - partial match", parent: "node1→account", child: "node1→account1", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsParentContext(tt.parent, tt.child)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestGetRequiredPermissionLevel tests the GetRequiredPermissionLevel function
func TestGetRequiredPermissionLevel(t *testing.T) {
	tests := []struct {
		name   string
		action string
		want   PermissionLevel
	}{
		// Create actions
		{name: "create action", action: "create", want: PermissionCreate},
		{name: "priorityCreate action", action: "priorityCreate", want: PermissionCreate},
		{name: "versionCreate action", action: "versionCreate", want: PermissionCreate},
		// Modify/Update actions
		{name: "modify action", action: "modify", want: PermissionUpdate},
		{name: "update action", action: "update", want: PermissionUpdate},
		{name: "priorityModify action", action: "priorityModify", want: PermissionUpdate},
		{name: "edit action", action: "edit", want: PermissionUpdate},
		// Delete/Remove actions
		{name: "remove action", action: "remove", want: PermissionDelete},
		{name: "delete action", action: "delete", want: PermissionDelete},
		{name: "priorityRemove action", action: "priorityRemove", want: PermissionDelete},
		// Read actions
		{name: "read action", action: "read", want: PermissionRead},
		{name: "list action", action: "list", want: PermissionRead},
		{name: "priorityRead action", action: "priorityRead", want: PermissionRead},
		// System actions
		{name: "version action", action: "version", want: PermissionRead},
		{name: "health action", action: "health", want: PermissionRead},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetRequiredPermissionLevel(tt.action)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestPermissionFull tests the enhanced Permission model
func TestPermissionFull(t *testing.T) {
	perm := Permission{
		ID:          "perm-123",
		Title:       "Admin Permission",
		Description: "Full admin access",
		Context:     "node1→account1",
		Level:       PermissionDelete,
		Created:     1234567890,
		Modified:    1234567890,
		Deleted:     false,
	}

	assert.Equal(t, "perm-123", perm.ID)
	assert.Equal(t, "Admin Permission", perm.Title)
	assert.Equal(t, "node1→account1", perm.Context)
	assert.Equal(t, PermissionDelete, perm.Level)
	assert.False(t, perm.Deleted)
}

// TestPermissionCheck tests the PermissionCheck model
func TestPermissionCheck(t *testing.T) {
	check := PermissionCheck{
		Username:      "testuser",
		Context:       "node1→account1→project1",
		RequiredLevel: PermissionUpdate,
		EntityType:    "ticket",
		EntityID:      "ticket-123",
		Action:        "modify",
	}

	assert.Equal(t, "testuser", check.Username)
	assert.Equal(t, "node1→account1→project1", check.Context)
	assert.Equal(t, PermissionUpdate, check.RequiredLevel)
	assert.Equal(t, "ticket", check.EntityType)
	assert.Equal(t, "modify", check.Action)
}

// TestPermissionContext tests the PermissionContext model
func TestPermissionContext(t *testing.T) {
	grandParent := &PermissionContext{
		Type:       "node",
		Identifier: "node1",
		Parent:     nil,
	}

	parent := &PermissionContext{
		Type:       "account",
		Identifier: "account1",
		Parent:     grandParent,
	}

	child := &PermissionContext{
		Type:       "project",
		Identifier: "project1",
		Parent:     parent,
	}

	assert.Equal(t, "project", child.Type)
	assert.Equal(t, "project1", child.Identifier)
	assert.NotNil(t, child.Parent)
	assert.Equal(t, "account", child.Parent.Type)
	assert.NotNil(t, child.Parent.Parent)
	assert.Equal(t, "node", child.Parent.Parent.Type)
}

// TestPermissionNone tests the NONE permission level
func TestPermissionNone(t *testing.T) {
	none := PermissionNone
	assert.Equal(t, PermissionLevel(0), none)
	assert.False(t, none.HasPermission(PermissionRead))
	assert.Equal(t, "NONE", none.String())
}

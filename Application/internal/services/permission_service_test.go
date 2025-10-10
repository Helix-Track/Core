package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"helixtrack.ru/core/internal/models"
)

// TestNewPermissionService tests the creation of HTTP permission service
func TestNewPermissionService(t *testing.T) {
	service := NewPermissionService("http://localhost:8080", 10, true)
	assert.NotNil(t, service)
	assert.True(t, service.IsEnabled())
}

// TestNewLocalPermissionService tests the creation of local permission service
func TestNewLocalPermissionService(t *testing.T) {
	service := NewLocalPermissionService(true)
	assert.NotNil(t, service)
	assert.True(t, service.IsEnabled())
}

// TestHttpPermissionService_IsEnabled tests the IsEnabled method
func TestHttpPermissionService_IsEnabled(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
		want    bool
	}{
		{name: "Service enabled", enabled: true, want: true},
		{name: "Service disabled", enabled: false, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewPermissionService("http://localhost:8080", 10, tt.enabled)
			assert.Equal(t, tt.want, service.IsEnabled())
		})
	}
}

// TestHttpPermissionService_CheckPermission_Disabled tests permission check when disabled
func TestHttpPermissionService_CheckPermission_Disabled(t *testing.T) {
	service := NewPermissionService("http://localhost:8080", 10, false)
	allowed, err := service.CheckPermission(context.Background(), "testuser", "node1", models.PermissionRead)
	assert.NoError(t, err)
	assert.True(t, allowed, "Disabled service should allow all operations")
}

// TestHttpPermissionService_CheckPermission_Success tests successful permission check
func TestHttpPermissionService_CheckPermission_Success(t *testing.T) {
	tests := []struct {
		name             string
		responseAllowed  bool
		expectedAllowed  bool
		responseStatus   int
	}{
		{
			name:            "Permission granted",
			responseAllowed: true,
			expectedAllowed: true,
			responseStatus:  http.StatusOK,
		},
		{
			name:            "Permission denied",
			responseAllowed: false,
			expectedAllowed: false,
			responseStatus:  http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock HTTP server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/check", r.URL.Path)
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				// Verify request body
				var req PermissionCheckRequest
				err := json.NewDecoder(r.Body).Decode(&req)
				assert.NoError(t, err)
				assert.Equal(t, "testuser", req.Username)
				assert.Equal(t, "node1→account1", req.Context)
				assert.Equal(t, models.PermissionUpdate, req.RequiredLevel)

				// Send response
				w.WriteHeader(tt.responseStatus)
				resp := PermissionCheckResponse{
					Allowed: tt.responseAllowed,
				}
				json.NewEncoder(w).Encode(resp)
			}))
			defer server.Close()

			service := NewPermissionService(server.URL, 10, true)
			allowed, err := service.CheckPermission(context.Background(), "testuser", "node1→account1", models.PermissionUpdate)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedAllowed, allowed)
		})
	}
}

// TestHttpPermissionService_CheckPermission_Error tests error handling
func TestHttpPermissionService_CheckPermission_Error(t *testing.T) {
	tests := []struct {
		name           string
		responseStatus int
		responseBody   string
		expectError    bool
	}{
		{
			name:           "Server error",
			responseStatus: http.StatusInternalServerError,
			responseBody:   "Internal server error",
			expectError:    true,
		},
		{
			name:           "Bad request",
			responseStatus: http.StatusBadRequest,
			responseBody:   "Invalid request",
			expectError:    true,
		},
		{
			name:           "Unauthorized",
			responseStatus: http.StatusUnauthorized,
			responseBody:   "Unauthorized",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.responseStatus)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			service := NewPermissionService(server.URL, 10, true)
			allowed, err := service.CheckPermission(context.Background(), "testuser", "node1", models.PermissionRead)
			if tt.expectError {
				assert.Error(t, err)
				assert.False(t, allowed)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestHttpPermissionService_GetUserPermissions_Success tests successful user permissions retrieval
func TestHttpPermissionService_GetUserPermissions_Success(t *testing.T) {
	expectedPermissions := []models.Permission{
		{
			ID:      "perm1",
			Context: "node1",
			Level:   models.PermissionRead,
		},
		{
			ID:      "perm2",
			Context: "node1→account1",
			Level:   models.PermissionUpdate,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/permissions/testuser", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.WriteHeader(http.StatusOK)
		resp := UserPermissionsResponse{
			Permissions: expectedPermissions,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	service := NewPermissionService(server.URL, 10, true)
	permissions, err := service.GetUserPermissions(context.Background(), "testuser")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(permissions))
	assert.Equal(t, "perm1", permissions[0].ID)
	assert.Equal(t, "perm2", permissions[1].ID)
}

// TestHttpPermissionService_GetUserPermissions_Disabled tests getting permissions when disabled
func TestHttpPermissionService_GetUserPermissions_Disabled(t *testing.T) {
	service := NewPermissionService("http://localhost:8080", 10, false)
	permissions, err := service.GetUserPermissions(context.Background(), "testuser")
	assert.NoError(t, err)
	assert.Empty(t, permissions)
}

// TestHttpPermissionService_GetUserPermissions_Error tests error handling
func TestHttpPermissionService_GetUserPermissions_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
	}))
	defer server.Close()

	service := NewPermissionService(server.URL, 10, true)
	permissions, err := service.GetUserPermissions(context.Background(), "testuser")
	assert.Error(t, err)
	assert.Nil(t, permissions)
}

// TestLocalPermissionService_IsEnabled tests the IsEnabled method
func TestLocalPermissionService_IsEnabled(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
		want    bool
	}{
		{name: "Service enabled", enabled: true, want: true},
		{name: "Service disabled", enabled: false, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewLocalPermissionService(tt.enabled)
			assert.Equal(t, tt.want, service.IsEnabled())
		})
	}
}

// TestLocalPermissionService_CheckPermission_Disabled tests permission check when disabled
func TestLocalPermissionService_CheckPermission_Disabled(t *testing.T) {
	service := NewLocalPermissionService(false)
	allowed, err := service.CheckPermission(context.Background(), "testuser", "node1", models.PermissionRead)
	assert.NoError(t, err)
	assert.True(t, allowed, "Disabled service should allow all operations")
}

// TestLocalPermissionService_AddUserPermission tests adding permissions
func TestLocalPermissionService_AddUserPermission(t *testing.T) {
	service := NewLocalPermissionService(true).(*localPermissionService)

	perm := models.Permission{
		ID:      "perm1",
		Context: "node1",
		Level:   models.PermissionRead,
	}

	service.AddUserPermission("testuser", perm)
	assert.Equal(t, 1, len(service.permissions["testuser"]))
	assert.Equal(t, "perm1", service.permissions["testuser"][0].ID)
}

// TestLocalPermissionService_CheckPermission_ExactMatch tests exact context match
func TestLocalPermissionService_CheckPermission_ExactMatch(t *testing.T) {
	service := NewLocalPermissionService(true).(*localPermissionService)

	service.AddUserPermission("testuser", models.Permission{
		ID:      "perm1",
		Context: "node1→account1",
		Level:   models.PermissionUpdate,
		Deleted: false,
	})

	tests := []struct {
		name          string
		context       string
		requiredLevel models.PermissionLevel
		expected      bool
	}{
		{
			name:          "Exact match with sufficient level",
			context:       "node1→account1",
			requiredLevel: models.PermissionRead,
			expected:      true,
		},
		{
			name:          "Exact match with equal level",
			context:       "node1→account1",
			requiredLevel: models.PermissionUpdate,
			expected:      true,
		},
		{
			name:          "Exact match with insufficient level",
			context:       "node1→account1",
			requiredLevel: models.PermissionDelete,
			expected:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, err := service.CheckPermission(context.Background(), "testuser", tt.context, tt.requiredLevel)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, allowed)
		})
	}
}

// TestLocalPermissionService_CheckPermission_ParentMatch tests parent context match
func TestLocalPermissionService_CheckPermission_ParentMatch(t *testing.T) {
	service := NewLocalPermissionService(true).(*localPermissionService)

	service.AddUserPermission("testuser", models.Permission{
		ID:      "perm1",
		Context: "node1→account1",
		Level:   models.PermissionUpdate,
		Deleted: false,
	})

	tests := []struct {
		name          string
		context       string
		requiredLevel models.PermissionLevel
		expected      bool
	}{
		{
			name:          "Child context with sufficient level",
			context:       "node1→account1→org1",
			requiredLevel: models.PermissionRead,
			expected:      true,
		},
		{
			name:          "Grandchild context with sufficient level",
			context:       "node1→account1→org1→team1",
			requiredLevel: models.PermissionCreate,
			expected:      true,
		},
		{
			name:          "Child context with insufficient level",
			context:       "node1→account1→org1",
			requiredLevel: models.PermissionDelete,
			expected:      false,
		},
		{
			name:          "Different hierarchy",
			context:       "node2→account1",
			requiredLevel: models.PermissionRead,
			expected:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, err := service.CheckPermission(context.Background(), "testuser", tt.context, tt.requiredLevel)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, allowed)
		})
	}
}

// TestLocalPermissionService_CheckPermission_NoPermissions tests user with no permissions
func TestLocalPermissionService_CheckPermission_NoPermissions(t *testing.T) {
	service := NewLocalPermissionService(true)

	allowed, err := service.CheckPermission(context.Background(), "testuser", "node1", models.PermissionRead)
	assert.NoError(t, err)
	assert.False(t, allowed, "User with no permissions should be denied")
}

// TestLocalPermissionService_CheckPermission_DeletedPermissions tests deleted permissions
func TestLocalPermissionService_CheckPermission_DeletedPermissions(t *testing.T) {
	service := NewLocalPermissionService(true).(*localPermissionService)

	service.AddUserPermission("testuser", models.Permission{
		ID:      "perm1",
		Context: "node1",
		Level:   models.PermissionUpdate,
		Deleted: true, // Deleted permission
	})

	allowed, err := service.CheckPermission(context.Background(), "testuser", "node1", models.PermissionRead)
	assert.NoError(t, err)
	assert.False(t, allowed, "Deleted permissions should not grant access")
}

// TestLocalPermissionService_CheckPermission_MultiplePermissions tests multiple permissions
func TestLocalPermissionService_CheckPermission_MultiplePermissions(t *testing.T) {
	service := NewLocalPermissionService(true).(*localPermissionService)

	// Add multiple permissions
	service.AddUserPermission("testuser", models.Permission{
		ID:      "perm1",
		Context: "node1",
		Level:   models.PermissionRead,
		Deleted: false,
	})
	service.AddUserPermission("testuser", models.Permission{
		ID:      "perm2",
		Context: "node1→account1",
		Level:   models.PermissionUpdate,
		Deleted: false,
	})
	service.AddUserPermission("testuser", models.Permission{
		ID:      "perm3",
		Context: "node2",
		Level:   models.PermissionDelete,
		Deleted: false,
	})

	tests := []struct {
		name          string
		context       string
		requiredLevel models.PermissionLevel
		expected      bool
	}{
		{
			name:          "First permission context",
			context:       "node1",
			requiredLevel: models.PermissionRead,
			expected:      true,
		},
		{
			name:          "Second permission context",
			context:       "node1→account1",
			requiredLevel: models.PermissionUpdate,
			expected:      true,
		},
		{
			name:          "Third permission context",
			context:       "node2",
			requiredLevel: models.PermissionDelete,
			expected:      true,
		},
		{
			name:          "Child of first permission",
			context:       "node1→account2",
			requiredLevel: models.PermissionRead,
			expected:      true,
		},
		{
			name:          "Insufficient level for first permission",
			context:       "node1",
			requiredLevel: models.PermissionUpdate,
			expected:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, err := service.CheckPermission(context.Background(), "testuser", tt.context, tt.requiredLevel)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, allowed)
		})
	}
}

// TestLocalPermissionService_GetUserPermissions tests getting user permissions
func TestLocalPermissionService_GetUserPermissions(t *testing.T) {
	service := NewLocalPermissionService(true).(*localPermissionService)

	perm1 := models.Permission{
		ID:      "perm1",
		Context: "node1",
		Level:   models.PermissionRead,
		Deleted: false,
	}
	perm2 := models.Permission{
		ID:      "perm2",
		Context: "node1→account1",
		Level:   models.PermissionUpdate,
		Deleted: false,
	}

	service.AddUserPermission("testuser", perm1)
	service.AddUserPermission("testuser", perm2)

	permissions, err := service.GetUserPermissions(context.Background(), "testuser")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(permissions))
	assert.Equal(t, "perm1", permissions[0].ID)
	assert.Equal(t, "perm2", permissions[1].ID)
}

// TestLocalPermissionService_GetUserPermissions_FilterDeleted tests deleted permissions filtering
func TestLocalPermissionService_GetUserPermissions_FilterDeleted(t *testing.T) {
	service := NewLocalPermissionService(true).(*localPermissionService)

	perm1 := models.Permission{
		ID:      "perm1",
		Context: "node1",
		Level:   models.PermissionRead,
		Deleted: false,
	}
	perm2 := models.Permission{
		ID:      "perm2",
		Context: "node1→account1",
		Level:   models.PermissionUpdate,
		Deleted: true, // Deleted
	}
	perm3 := models.Permission{
		ID:      "perm3",
		Context: "node2",
		Level:   models.PermissionDelete,
		Deleted: false,
	}

	service.AddUserPermission("testuser", perm1)
	service.AddUserPermission("testuser", perm2)
	service.AddUserPermission("testuser", perm3)

	permissions, err := service.GetUserPermissions(context.Background(), "testuser")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(permissions), "Should only return non-deleted permissions")
	assert.Equal(t, "perm1", permissions[0].ID)
	assert.Equal(t, "perm3", permissions[1].ID)
}

// TestLocalPermissionService_GetUserPermissions_NoUser tests getting permissions for non-existent user
func TestLocalPermissionService_GetUserPermissions_NoUser(t *testing.T) {
	service := NewLocalPermissionService(true)

	permissions, err := service.GetUserPermissions(context.Background(), "nonexistent")
	assert.NoError(t, err)
	assert.Empty(t, permissions)
}

// TestLocalPermissionService_GetUserPermissions_Disabled tests getting permissions when disabled
func TestLocalPermissionService_GetUserPermissions_Disabled(t *testing.T) {
	service := NewLocalPermissionService(false)

	permissions, err := service.GetUserPermissions(context.Background(), "testuser")
	assert.NoError(t, err)
	assert.Empty(t, permissions)
}

// TestMockPermissionService tests the mock implementation
func TestMockPermissionService(t *testing.T) {
	mock := &MockPermissionService{
		CheckPermissionFunc: func(ctx context.Context, username, permissionContext string, requiredLevel models.PermissionLevel) (bool, error) {
			return username == "admin", nil
		},
		GetUserPermissionsFunc: func(ctx context.Context, username string) ([]models.Permission, error) {
			if username == "admin" {
				return []models.Permission{
					{ID: "perm1", Context: "node1", Level: models.PermissionDelete},
				}, nil
			}
			return []models.Permission{}, nil
		},
		IsEnabledFunc: func() bool {
			return true
		},
	}

	// Test CheckPermission
	allowed, err := mock.CheckPermission(context.Background(), "admin", "node1", models.PermissionRead)
	assert.NoError(t, err)
	assert.True(t, allowed)

	allowed, err = mock.CheckPermission(context.Background(), "user", "node1", models.PermissionRead)
	assert.NoError(t, err)
	assert.False(t, allowed)

	// Test GetUserPermissions
	permissions, err := mock.GetUserPermissions(context.Background(), "admin")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(permissions))

	permissions, err = mock.GetUserPermissions(context.Background(), "user")
	assert.NoError(t, err)
	assert.Empty(t, permissions)

	// Test IsEnabled
	assert.True(t, mock.IsEnabled())
}

// TestMockPermissionService_Defaults tests mock with default behavior
func TestMockPermissionService_Defaults(t *testing.T) {
	mock := &MockPermissionService{}

	// Default behavior should allow all and return empty permissions
	allowed, err := mock.CheckPermission(context.Background(), "testuser", "node1", models.PermissionRead)
	assert.NoError(t, err)
	assert.True(t, allowed, "Default should allow")

	permissions, err := mock.GetUserPermissions(context.Background(), "testuser")
	assert.NoError(t, err)
	assert.Empty(t, permissions, "Default should return empty")

	assert.False(t, mock.IsEnabled(), "Default should be disabled")
}

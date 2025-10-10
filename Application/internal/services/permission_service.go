package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"helixtrack.ru/core/internal/models"
)

// PermissionService defines the interface for permission checking operations
type PermissionService interface {
	// CheckPermission checks if a user has required permission for a context
	CheckPermission(ctx context.Context, username, permissionContext string, requiredLevel models.PermissionLevel) (bool, error)

	// GetUserPermissions retrieves all permissions for a user
	GetUserPermissions(ctx context.Context, username string) ([]models.Permission, error)

	// IsEnabled returns whether the permission service is enabled
	IsEnabled() bool
}

// httpPermissionService is the HTTP-based implementation of PermissionService
type httpPermissionService struct {
	baseURL    string
	httpClient *http.Client
	enabled    bool
}

// NewPermissionService creates a new permission service client
func NewPermissionService(baseURL string, timeout int, enabled bool) PermissionService {
	return &httpPermissionService{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
		enabled: enabled,
	}
}

// PermissionCheckRequest represents a permission check request
type PermissionCheckRequest struct {
	Username          string                  `json:"username"`
	Context           string                  `json:"context"`
	RequiredLevel     models.PermissionLevel  `json:"required_level"`
}

// PermissionCheckResponse represents a permission check response
type PermissionCheckResponse struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason,omitempty"`
}

// UserPermissionsResponse represents a user permissions response
type UserPermissionsResponse struct {
	Permissions []models.Permission `json:"permissions"`
}

// CheckPermission checks if a user has required permission for a context
func (s *httpPermissionService) CheckPermission(ctx context.Context, username, permissionContext string, requiredLevel models.PermissionLevel) (bool, error) {
	if !s.enabled {
		// If permission service is disabled, allow all operations (development mode)
		return true, nil
	}

	reqBody := PermissionCheckRequest{
		Username:      username,
		Context:       permissionContext,
		RequiredLevel: requiredLevel,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return false, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/check", bytes.NewBuffer(jsonData))
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("permission check failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var checkResp PermissionCheckResponse
	if err := json.NewDecoder(resp.Body).Decode(&checkResp); err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	return checkResp.Allowed, nil
}

// GetUserPermissions retrieves all permissions for a user
func (s *httpPermissionService) GetUserPermissions(ctx context.Context, username string) ([]models.Permission, error) {
	if !s.enabled {
		// If permission service is disabled, return empty permissions
		return []models.Permission{}, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.baseURL+"/permissions/"+username, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get permissions failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var permResp UserPermissionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&permResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return permResp.Permissions, nil
}

// IsEnabled returns whether the permission service is enabled
func (s *httpPermissionService) IsEnabled() bool {
	return s.enabled
}

// localPermissionService is a local/in-memory implementation of PermissionService
// This can be used as a free/open-source alternative to proprietary implementations
type localPermissionService struct {
	permissions map[string][]models.Permission // username -> permissions
	enabled     bool
}

// NewLocalPermissionService creates a new local permission service
func NewLocalPermissionService(enabled bool) PermissionService {
	return &localPermissionService{
		permissions: make(map[string][]models.Permission),
		enabled:     enabled,
	}
}

// AddUserPermission adds a permission for a user (for testing/setup)
func (s *localPermissionService) AddUserPermission(username string, permission models.Permission) {
	if s.permissions[username] == nil {
		s.permissions[username] = []models.Permission{}
	}
	s.permissions[username] = append(s.permissions[username], permission)
}

// CheckPermission checks if a user has required permission for a context
func (s *localPermissionService) CheckPermission(ctx context.Context, username, permissionContext string, requiredLevel models.PermissionLevel) (bool, error) {
	if !s.enabled {
		// If permission service is disabled, allow all operations
		return true, nil
	}

	userPerms, exists := s.permissions[username]
	if !exists {
		return false, nil // User has no permissions
	}

	// Check for exact context match or parent context match
	for _, perm := range userPerms {
		if perm.Deleted {
			continue
		}

		// Check if permission context matches or is a parent
		if perm.Context == permissionContext || models.IsParentContext(perm.Context, permissionContext) {
			// Check if permission level is sufficient
			if perm.Level.HasPermission(requiredLevel) {
				return true, nil
			}
		}
	}

	return false, nil
}

// GetUserPermissions retrieves all permissions for a user
func (s *localPermissionService) GetUserPermissions(ctx context.Context, username string) ([]models.Permission, error) {
	if !s.enabled {
		return []models.Permission{}, nil
	}

	userPerms, exists := s.permissions[username]
	if !exists {
		return []models.Permission{}, nil
	}

	// Filter out deleted permissions
	activePerms := []models.Permission{}
	for _, perm := range userPerms {
		if !perm.Deleted {
			activePerms = append(activePerms, perm)
		}
	}

	return activePerms, nil
}

// IsEnabled returns whether the permission service is enabled
func (s *localPermissionService) IsEnabled() bool {
	return s.enabled
}

// MockPermissionService is a mock implementation for testing
type MockPermissionService struct {
	CheckPermissionFunc    func(ctx context.Context, username, permissionContext string, requiredLevel models.PermissionLevel) (bool, error)
	GetUserPermissionsFunc func(ctx context.Context, username string) ([]models.Permission, error)
	IsEnabledFunc          func() bool
}

func (m *MockPermissionService) CheckPermission(ctx context.Context, username, permissionContext string, requiredLevel models.PermissionLevel) (bool, error) {
	if m.CheckPermissionFunc != nil {
		return m.CheckPermissionFunc(ctx, username, permissionContext, requiredLevel)
	}
	return true, nil // Default to allowing
}

func (m *MockPermissionService) GetUserPermissions(ctx context.Context, username string) ([]models.Permission, error) {
	if m.GetUserPermissionsFunc != nil {
		return m.GetUserPermissionsFunc(ctx, username)
	}
	return []models.Permission{}, nil
}

func (m *MockPermissionService) IsEnabled() bool {
	if m.IsEnabledFunc != nil {
		return m.IsEnabledFunc()
	}
	return false
}

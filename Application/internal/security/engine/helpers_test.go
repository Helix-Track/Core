package engine

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockEngine is a mock implementation of the Engine interface for testing
type MockEngine struct {
	mock.Mock
}

func (m *MockEngine) CheckAccess(ctx context.Context, req AccessRequest) (AccessResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(AccessResponse), args.Error(1)
}

func (m *MockEngine) GetEffectivePermissions(ctx context.Context, username, resourceType, resourceID string) (PermissionSet, error) {
	args := m.Called(ctx, username, resourceType, resourceID)
	return args.Get(0).(PermissionSet), args.Error(1)
}

func (m *MockEngine) ValidateSecurityLevel(ctx context.Context, username, entityID string) (bool, error) {
	args := m.Called(ctx, username, entityID)
	return args.Bool(0), args.Error(1)
}

func (m *MockEngine) EvaluateRole(ctx context.Context, username, projectID, requiredRole string) (bool, error) {
	args := m.Called(ctx, username, projectID, requiredRole)
	return args.Bool(0), args.Error(1)
}

func (m *MockEngine) GetSecurityContext(ctx context.Context, username string) (*SecurityContext, error) {
	args := m.Called(ctx, username)
	result := args.Get(0)
	if result == nil {
		return nil, args.Error(1)
	}
	return result.(*SecurityContext), args.Error(1)
}

func (m *MockEngine) InvalidateCache(username string) {
	m.Called(username)
}

func (m *MockEngine) InvalidateAllCache() {
	m.Called()
}

func (m *MockEngine) AuditAccessAttempt(ctx context.Context, req AccessRequest, response AccessResponse) error {
	args := m.Called(ctx, req, response)
	return args.Error(0)
}

// TestNewHelperMethods tests helper creation
func TestNewHelperMethods(t *testing.T) {
	mockEngine := new(MockEngine)
	helpers := NewHelperMethods(mockEngine)

	assert.NotNil(t, helpers)
	assert.NotNil(t, helpers.engine)
}

// TestCanUserCreate tests CREATE permission check
func TestCanUserCreate(t *testing.T) {
	mockEngine := new(MockEngine)
	helpers := NewHelperMethods(mockEngine)
	ctx := context.Background()
	context := map[string]string{"project_id": "proj-1"}

	mockEngine.On("CheckAccess", ctx, mock.MatchedBy(func(req AccessRequest) bool {
		return req.Username == "testuser" && req.Resource == "ticket" && req.Action == ActionCreate
	})).Return(AccessResponse{Allowed: true, Reason: "Permission granted"}, nil)

	canCreate, err := helpers.CanUserCreate(ctx, "testuser", "ticket", context)

	assert.NoError(t, err)
	assert.True(t, canCreate)
	mockEngine.AssertExpectations(t)
}

// TestCanUserList tests LIST permission check
func TestCanUserList(t *testing.T) {
	mockEngine := new(MockEngine)
	helpers := NewHelperMethods(mockEngine)
	ctx := context.Background()
	context := map[string]string{"project_id": "proj-1"}

	mockEngine.On("CheckAccess", ctx, mock.MatchedBy(func(req AccessRequest) bool {
		return req.Username == "testuser" && req.Resource == "ticket" && req.Action == ActionList
	})).Return(AccessResponse{Allowed: true, Reason: "Permission granted"}, nil)

	canList, err := helpers.CanUserList(ctx, "testuser", "ticket", context)

	assert.NoError(t, err)
	assert.True(t, canList)
	mockEngine.AssertExpectations(t)
}


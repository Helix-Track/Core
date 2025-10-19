package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"helixtrack.ru/core/internal/security/engine"
)

// MockSecurityEngine is a mock implementation of engine.Engine
type MockSecurityEngine struct {
	mock.Mock
}

func (m *MockSecurityEngine) CheckAccess(ctx context.Context, req engine.AccessRequest) (engine.AccessResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(engine.AccessResponse), args.Error(1)
}

func (m *MockSecurityEngine) ValidateSecurityLevel(ctx context.Context, username, entityID string) (bool, error) {
	args := m.Called(ctx, username, entityID)
	return args.Bool(0), args.Error(1)
}

func (m *MockSecurityEngine) EvaluateRole(ctx context.Context, username, projectID, requiredRole string) (bool, error) {
	args := m.Called(ctx, username, projectID, requiredRole)
	return args.Bool(0), args.Error(1)
}

func (m *MockSecurityEngine) GetSecurityContext(ctx context.Context, username string) (*engine.SecurityContext, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*engine.SecurityContext), args.Error(1)
}

func (m *MockSecurityEngine) InvalidateCache(username string) {
	m.Called(username)
}

func (m *MockSecurityEngine) InvalidateAllCache() {
	m.Called()
}

func (m *MockSecurityEngine) AuditAccessAttempt(ctx context.Context, req engine.AccessRequest, response engine.AccessResponse) error {
	args := m.Called(ctx, req, response)
	return args.Error(0)
}

func (m *MockSecurityEngine) GetEffectivePermissions(ctx context.Context, username, resourceType, resourceID string) (engine.PermissionSet, error) {
	args := m.Called(ctx, username, resourceType, resourceID)
	return args.Get(0).(engine.PermissionSet), args.Error(1)
}

// Helper function to create test Gin context
func createTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/test", nil)
	return c, w
}

// Helper function to set username in context
func setUsername(c *gin.Context, username string) {
	c.Set("username", username)
}

// TestRBACMiddleware_Success tests successful authorization
func TestRBACMiddleware_Success(t *testing.T) {
	mockEngine := new(MockSecurityEngine)
	c, w := createTestContext()
	setUsername(c, "testuser")

	mockEngine.On("CheckAccess", mock.Anything, mock.MatchedBy(func(req engine.AccessRequest) bool {
		return req.Username == "testuser" && req.Resource == "ticket" && req.Action == engine.ActionRead
	})).Return(engine.AccessResponse{
		Allowed: true,
		Reason:  "Access granted",
	}, nil)

	mockEngine.On("AuditAccessAttempt", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

	middleware := RBACMiddleware(mockEngine, "ticket", engine.ActionRead)
	middleware(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, c.GetBool("rbac_authorized"))
	assert.Equal(t, "ticket", c.GetString("rbac_resource"))
	mockEngine.AssertExpectations(t)
}

// TestRBACMiddleware_Denied tests access denial
func TestRBACMiddleware_Denied(t *testing.T) {
	mockEngine := new(MockSecurityEngine)
	c, w := createTestContext()
	setUsername(c, "testuser")

	mockEngine.On("CheckAccess", mock.Anything, mock.Anything).Return(engine.AccessResponse{
		Allowed: false,
		Reason:  "Insufficient permissions",
	}, nil)

	mockEngine.On("AuditAccessAttempt", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

	middleware := RBACMiddleware(mockEngine, "ticket", engine.ActionDelete)
	middleware(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.False(t, c.GetBool("rbac_authorized"))
	mockEngine.AssertExpectations(t)
}

// TestRBACMiddleware_NoUsername tests missing authentication
func TestRBACMiddleware_NoUsername(t *testing.T) {
	mockEngine := new(MockSecurityEngine)
	c, w := createTestContext()
	// Don't set username - simulate unauthenticated request

	middleware := RBACMiddleware(mockEngine, "ticket", engine.ActionRead)
	middleware(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockEngine.AssertNotCalled(t, "CheckAccess")
}

// TestRBACMiddleware_Error tests error handling
func TestRBACMiddleware_Error(t *testing.T) {
	mockEngine := new(MockSecurityEngine)
	c, w := createTestContext()
	setUsername(c, "testuser")

	mockEngine.On("CheckAccess", mock.Anything, mock.Anything).Return(
		engine.AccessResponse{},
		errors.New("database error"),
	)

	mockEngine.On("AuditAccessAttempt", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

	middleware := RBACMiddleware(mockEngine, "ticket", engine.ActionRead)
	middleware(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockEngine.AssertExpectations(t)
}

// TestRBACMiddleware_WithResourceID tests access check with resource ID
func TestRBACMiddleware_WithResourceID(t *testing.T) {
	mockEngine := new(MockSecurityEngine)
	c, w := createTestContext()
	setUsername(c, "testuser")
	c.Params = []gin.Param{{Key: "id", Value: "ticket-123"}}

	mockEngine.On("CheckAccess", mock.Anything, mock.MatchedBy(func(req engine.AccessRequest) bool {
		return req.ResourceID == "ticket-123"
	})).Return(engine.AccessResponse{
		Allowed: true,
		Reason:  "Access granted",
	}, nil)

	mockEngine.On("AuditAccessAttempt", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

	middleware := RBACMiddleware(mockEngine, "ticket", engine.ActionUpdate)
	middleware(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockEngine.AssertExpectations(t)
}

// TestRequireSecurityLevel_Success tests successful security level check
func TestRequireSecurityLevel_Success(t *testing.T) {
	mockEngine := new(MockSecurityEngine)
	c, w := createTestContext()
	setUsername(c, "testuser")
	c.Params = []gin.Param{{Key: "id", Value: "entity-123"}}

	mockEngine.On("ValidateSecurityLevel", mock.Anything, "testuser", "entity-123").Return(true, nil)

	middleware := RequireSecurityLevel(mockEngine)
	middleware(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, c.GetBool("security_level_checked"))
	mockEngine.AssertExpectations(t)
}

// TestRequireSecurityLevel_Denied tests security level access denial
func TestRequireSecurityLevel_Denied(t *testing.T) {
	mockEngine := new(MockSecurityEngine)
	c, w := createTestContext()
	setUsername(c, "testuser")
	c.Params = []gin.Param{{Key: "id", Value: "entity-123"}}

	mockEngine.On("ValidateSecurityLevel", mock.Anything, "testuser", "entity-123").Return(false, nil)

	middleware := RequireSecurityLevel(mockEngine)
	middleware(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	mockEngine.AssertExpectations(t)
}

// TestRequireSecurityLevel_NoEntityID tests security level check without entity ID
func TestRequireSecurityLevel_NoEntityID(t *testing.T) {
	mockEngine := new(MockSecurityEngine)
	c, w := createTestContext()
	setUsername(c, "testuser")
	// No entity ID provided

	middleware := RequireSecurityLevel(mockEngine)
	middleware(c)

	// Should allow through when no entity ID
	assert.Equal(t, http.StatusOK, w.Code)
	mockEngine.AssertNotCalled(t, "ValidateSecurityLevel")
}

// TestRequireSecurityLevel_NoUsername tests security level check without authentication
func TestRequireSecurityLevel_NoUsername(t *testing.T) {
	mockEngine := new(MockSecurityEngine)
	c, w := createTestContext()
	// No username set

	middleware := RequireSecurityLevel(mockEngine)
	middleware(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockEngine.AssertNotCalled(t, "ValidateSecurityLevel")
}

// TestRequireProjectRole_Success tests successful project role check
func TestRequireProjectRole_Success(t *testing.T) {
	mockEngine := new(MockSecurityEngine)
	c, w := createTestContext()
	setUsername(c, "testuser")
	c.Params = []gin.Param{{Key: "projectId", Value: "proj-1"}}

	mockEngine.On("EvaluateRole", mock.Anything, "testuser", "proj-1", "Developer").Return(true, nil)

	middleware := RequireProjectRole(mockEngine, "Developer")
	middleware(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, c.GetBool("project_role_checked"))
	assert.Equal(t, "Developer", c.GetString("project_role"))
	mockEngine.AssertExpectations(t)
}

// TestRequireProjectRole_Denied tests project role denial
func TestRequireProjectRole_Denied(t *testing.T) {
	mockEngine := new(MockSecurityEngine)
	c, w := createTestContext()
	setUsername(c, "testuser")
	c.Params = []gin.Param{{Key: "projectId", Value: "proj-1"}}

	mockEngine.On("EvaluateRole", mock.Anything, "testuser", "proj-1", "Project Administrator").Return(false, nil)

	middleware := RequireProjectRole(mockEngine, "Project Administrator")
	middleware(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	mockEngine.AssertExpectations(t)
}

// TestRequireProjectRole_NoProjectID tests project role check without project ID
func TestRequireProjectRole_NoProjectID(t *testing.T) {
	mockEngine := new(MockSecurityEngine)
	c, w := createTestContext()
	setUsername(c, "testuser")
	// No project ID

	middleware := RequireProjectRole(mockEngine, "Developer")
	middleware(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockEngine.AssertNotCalled(t, "EvaluateRole")
}

// TestSecurityContextMiddleware_Success tests security context loading
func TestSecurityContextMiddleware_Success(t *testing.T) {
	mockEngine := new(MockSecurityEngine)
	c, w := createTestContext()
	setUsername(c, "testuser")

	mockSecCtx := &engine.SecurityContext{
		Username: "testuser",
		Roles:    []engine.Role{{ID: "role-1", Title: "Developer"}},
		Teams:    []string{"team-1"},
	}

	mockEngine.On("GetSecurityContext", mock.Anything, "testuser").Return(mockSecCtx, nil)

	middleware := SecurityContextMiddleware(mockEngine)
	middleware(c)

	assert.Equal(t, http.StatusOK, w.Code)

	secCtx, exists := GetSecurityContext(c)
	assert.True(t, exists)
	assert.Equal(t, "testuser", secCtx.Username)
	assert.Len(t, secCtx.Roles, 1)
	mockEngine.AssertExpectations(t)
}

// TestSecurityContextMiddleware_NoUsername tests security context without authentication
func TestSecurityContextMiddleware_NoUsername(t *testing.T) {
	mockEngine := new(MockSecurityEngine)
	c, w := createTestContext()
	// No username

	middleware := SecurityContextMiddleware(mockEngine)
	middleware(c)

	assert.Equal(t, http.StatusOK, w.Code)
	_, exists := GetSecurityContext(c)
	assert.False(t, exists)
	mockEngine.AssertNotCalled(t, "GetSecurityContext")
}

// TestExtractResourceID tests resource ID extraction
func TestExtractResourceID(t *testing.T) {
	tests := []struct {
		name       string
		setupFunc  func(*gin.Context)
		resource   string
		expectedID string
	}{
		{
			name: "From URL parameter",
			setupFunc: func(c *gin.Context) {
				c.Params = []gin.Param{{Key: "id", Value: "ticket-123"}}
			},
			resource:   "ticket",
			expectedID: "ticket-123",
		},
		{
			name: "From query parameter",
			setupFunc: func(c *gin.Context) {
				c.Request, _ = http.NewRequest("GET", "/test?id=ticket-456", nil)
			},
			resource:   "ticket",
			expectedID: "ticket-456",
		},
		{
			name: "From resource-specific parameter",
			setupFunc: func(c *gin.Context) {
				c.Params = []gin.Param{{Key: "ticketId", Value: "ticket-789"}}
			},
			resource:   "ticket",
			expectedID: "ticket-789",
		},
		{
			name: "From JSON body",
			setupFunc: func(c *gin.Context) {
				c.Request, _ = http.NewRequest("POST", "/test", strings.NewReader(`{"id":"ticket-999"}`))
				c.Request.Header.Set("Content-Type", "application/json")
			},
			resource:   "ticket",
			expectedID: "ticket-999",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := createTestContext()
			tt.setupFunc(c)

			resourceID := extractResourceID(c, tt.resource)
			assert.Equal(t, tt.expectedID, resourceID)
		})
	}
}

// TestExtractContext tests context extraction
func TestExtractContext(t *testing.T) {
	c, _ := createTestContext()
	c.Params = []gin.Param{
		{Key: "projectId", Value: "proj-1"},
		{Key: "teamId", Value: "team-1"},
	}
	c.Request.RemoteAddr = "192.168.1.1:8080"
	c.Request.Header.Set("User-Agent", "TestAgent/1.0")

	context := extractContext(c)

	assert.Equal(t, "proj-1", context["project_id"])
	assert.Equal(t, "team-1", context["team_id"])
	assert.NotEmpty(t, context["ip_address"])
	assert.Equal(t, "TestAgent/1.0", context["user_agent"])
	assert.NotEmpty(t, context["request_path"])
	assert.NotEmpty(t, context["request_method"])
}

// TestGetSecurityContext tests security context retrieval
func TestGetSecurityContext(t *testing.T) {
	c, _ := createTestContext()

	// Test when security context exists
	mockSecCtx := &engine.SecurityContext{
		Username: "testuser",
	}
	c.Set("security_context", mockSecCtx)

	secCtx, exists := GetSecurityContext(c)
	assert.True(t, exists)
	assert.Equal(t, "testuser", secCtx.Username)

	// Test when security context doesn't exist
	c2, _ := createTestContext()
	secCtx2, exists2 := GetSecurityContext(c2)
	assert.False(t, exists2)
	assert.Nil(t, secCtx2)
}

// TestIsAuthorized tests authorization check
func TestIsAuthorized(t *testing.T) {
	// Test when authorized
	c, _ := createTestContext()
	c.Set("rbac_authorized", true)
	assert.True(t, IsAuthorized(c))

	// Test when not authorized
	c2, _ := createTestContext()
	c2.Set("rbac_authorized", false)
	assert.False(t, IsAuthorized(c2))

	// Test when not set
	c3, _ := createTestContext()
	assert.False(t, IsAuthorized(c3))
}

// TestMultipleActions tests checking multiple actions
func TestMultipleActions(t *testing.T) {
	actions := []engine.Action{
		engine.ActionCreate,
		engine.ActionRead,
		engine.ActionUpdate,
		engine.ActionDelete,
		engine.ActionList,
		engine.ActionExecute,
	}

	for _, action := range actions {
		t.Run(string(action), func(t *testing.T) {
			mockEngine := new(MockSecurityEngine)
			c, w := createTestContext()
			setUsername(c, "testuser")

			mockEngine.On("CheckAccess", mock.Anything, mock.MatchedBy(func(req engine.AccessRequest) bool {
				return req.Action == action
			})).Return(engine.AccessResponse{
				Allowed: true,
				Reason:  "Access granted",
			}, nil)

			middleware := RBACMiddleware(mockEngine, "ticket", action)
			middleware(c)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, string(action), c.GetString("rbac_action"))
			mockEngine.AssertExpectations(t)
		})
	}
}

// TestResourceTypes tests different resource types
func TestResourceTypes(t *testing.T) {
	resources := []string{"ticket", "project", "epic", "subtask", "work_log", "dashboard"}

	for _, resource := range resources {
		t.Run(resource, func(t *testing.T) {
			mockEngine := new(MockSecurityEngine)
			c, w := createTestContext()
			setUsername(c, "testuser")

			mockEngine.On("CheckAccess", mock.Anything, mock.MatchedBy(func(req engine.AccessRequest) bool {
				return req.Resource == resource
			})).Return(engine.AccessResponse{
				Allowed: true,
				Reason:  "Access granted",
			}, nil)

			middleware := RBACMiddleware(mockEngine, resource, engine.ActionRead)
			middleware(c)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, resource, c.GetString("rbac_resource"))
			mockEngine.AssertExpectations(t)
		})
	}
}

// TestErrorScenarios tests various error scenarios
func TestErrorScenarios(t *testing.T) {
	tests := []struct {
		name           string
		setupFunc      func(*MockSecurityEngine, *gin.Context)
		expectedStatus int
	}{
		{
			name: "Database error",
			setupFunc: func(m *MockSecurityEngine, c *gin.Context) {
				setUsername(c, "testuser")
				m.On("CheckAccess", mock.Anything, mock.Anything).Return(
					engine.AccessResponse{},
					errors.New("database error"),
				)
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Security level error",
			setupFunc: func(m *MockSecurityEngine, c *gin.Context) {
				setUsername(c, "testuser")
				c.Params = []gin.Param{{Key: "id", Value: "entity-123"}}
				m.On("ValidateSecurityLevel", mock.Anything, "testuser", "entity-123").Return(
					false,
					errors.New("validation error"),
				)
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Role evaluation error",
			setupFunc: func(m *MockSecurityEngine, c *gin.Context) {
				setUsername(c, "testuser")
				c.Params = []gin.Param{{Key: "projectId", Value: "proj-1"}}
				m.On("EvaluateRole", mock.Anything, "testuser", "proj-1", "Developer").Return(
					false,
					errors.New("evaluation error"),
				)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEngine := new(MockSecurityEngine)
			c, w := createTestContext()
			tt.setupFunc(mockEngine, c)

			var middleware gin.HandlerFunc
			if strings.Contains(tt.name, "Security level") {
				middleware = RequireSecurityLevel(mockEngine)
			} else if strings.Contains(tt.name, "Role") {
				middleware = RequireProjectRole(mockEngine, "Developer")
			} else {
				middleware = RBACMiddleware(mockEngine, "ticket", engine.ActionRead)
			}

			middleware(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockEngine.AssertExpectations(t)
		})
	}
}

// Benchmark tests
func BenchmarkRBACMiddleware(b *testing.B) {
	mockEngine := new(MockSecurityEngine)
	mockEngine.On("CheckAccess", mock.Anything, mock.Anything).Return(engine.AccessResponse{
		Allowed: true,
		Reason:  "Access granted",
	}, nil)

	middleware := RBACMiddleware(mockEngine, "ticket", engine.ActionRead)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c, _ := createTestContext()
		setUsername(c, "testuser")
		middleware(c)
	}
}

func BenchmarkExtractContext(b *testing.B) {
	c, _ := createTestContext()
	c.Params = []gin.Param{
		{Key: "projectId", Value: "proj-1"},
		{Key: "teamId", Value: "team-1"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = extractContext(c)
	}
}

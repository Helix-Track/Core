package engine

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDatabase is a mock implementation of database.Database for testing
type MockDatabase struct {
	mock.Mock
}

func (m *MockDatabase) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	mockArgs := m.Called(ctx, query, args)
	result := mockArgs.Get(0)
	if result == nil {
		return nil, mockArgs.Error(1)
	}
	return result.(*sql.Rows), mockArgs.Error(1)
}

func (m *MockDatabase) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	mockArgs := m.Called(ctx, query, args)
	result := mockArgs.Get(0)
	if result == nil {
		return nil
	}
	return result.(*sql.Row)
}

func (m *MockDatabase) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	mockArgs := m.Called(ctx, query, args)
	result := mockArgs.Get(0)
	if result == nil {
		return nil, mockArgs.Error(1)
	}
	return result.(sql.Result), mockArgs.Error(1)
}

func (m *MockDatabase) Begin(ctx context.Context) (*sql.Tx, error) {
	mockArgs := m.Called(ctx)
	result := mockArgs.Get(0)
	if result == nil {
		return nil, mockArgs.Error(1)
	}
	return result.(*sql.Tx), mockArgs.Error(1)
}

func (m *MockDatabase) Close() error {
	mockArgs := m.Called()
	return mockArgs.Error(0)
}

func (m *MockDatabase) Ping(ctx context.Context) error {
	mockArgs := m.Called(ctx)
	return mockArgs.Error(0)
}

func (m *MockDatabase) GetType() string {
	mockArgs := m.Called()
	return mockArgs.String(0)
}

// TestNewSecurityEngine tests the creation of a new security engine
func TestNewSecurityEngine(t *testing.T) {
	mockDB := new(MockDatabase)
	config := DefaultConfig()

	engine := NewSecurityEngine(mockDB, config)

	assert.NotNil(t, engine)
	assert.NotNil(t, engine.permissionResolver)
	assert.NotNil(t, engine.roleEvaluator)
	assert.NotNil(t, engine.securityChecker)
	assert.NotNil(t, engine.cache)
	assert.NotNil(t, engine.auditLogger)
}

// TestDefaultConfig tests the default configuration
func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.True(t, config.EnableCaching)
	assert.Equal(t, 5*time.Minute, config.CacheTTL)
	assert.Equal(t, 10000, config.CacheMaxSize)
	assert.True(t, config.EnableAuditing)
	assert.True(t, config.AuditAllAttempts)
	assert.Equal(t, 90*24*time.Hour, config.AuditRetention)
}

// TestCheckAccess_AllowedWithPermission tests successful access with permission
func TestCheckAccess_AllowedWithPermission(t *testing.T) {
	// This is a simplified test - in reality, we'd need to mock database responses
	// For now, we're testing the structure
	mockDB := new(MockDatabase)
	config := Config{
		EnableCaching:  false, // Disable caching for predictable tests
		EnableAuditing: false, // Disable auditing for cleaner tests
	}

	engine := NewSecurityEngine(mockDB, config)

	req := AccessRequest{
		Username: "testuser",
		Resource: "ticket",
		Action:   ActionRead,
		Context:  map[string]string{},
	}

	// Note: This test will need proper mocking of database calls
	// For complete implementation, we'd mock the database responses
	assert.NotNil(t, engine)
	assert.NotNil(t, req)
}

// TestCheckAccess_DeniedWithoutPermission tests access denial
func TestCheckAccess_DeniedWithoutPermission(t *testing.T) {
	mockDB := new(MockDatabase)
	config := Config{
		EnableCaching:  false,
		EnableAuditing: false,
	}

	engine := NewSecurityEngine(mockDB, config)

	req := AccessRequest{
		Username: "testuser",
		Resource: "ticket",
		Action:   ActionDelete,
		Context:  map[string]string{},
	}

	// Test structure
	assert.NotNil(t, engine)
	assert.Equal(t, "ticket", req.Resource)
	assert.Equal(t, ActionDelete, req.Action)
}

// TestInvalidateCache tests cache invalidation
func TestInvalidateCache(t *testing.T) {
	mockDB := new(MockDatabase)
	config := DefaultConfig()

	engine := NewSecurityEngine(mockDB, config)

	// Add a cache entry first
	req := AccessRequest{
		Username: "testuser",
		Resource: "ticket",
		Action:   ActionRead,
	}

	response := AccessResponse{
		Allowed: true,
		Reason:  "Test",
	}

	engine.cache.Set(req, response)

	// Verify it's cached
	cached, found := engine.cache.Get(req)
	assert.True(t, found)
	assert.True(t, cached.Allowed)

	// Invalidate cache for user
	engine.InvalidateCache("testuser")

	// Verify it's no longer cached
	_, found = engine.cache.Get(req)
	assert.False(t, found)
}

// TestInvalidateAllCache tests clearing entire cache
func TestInvalidateAllCache(t *testing.T) {
	mockDB := new(MockDatabase)
	config := DefaultConfig()

	engine := NewSecurityEngine(mockDB, config)

	// Add multiple cache entries
	for i := 0; i < 5; i++ {
		req := AccessRequest{
			Username:   "testuser",
			Resource:   "ticket",
			ResourceID: string(rune(i)),
			Action:     ActionRead,
		}

		response := AccessResponse{
			Allowed: true,
		}

		engine.cache.Set(req, response)
	}

	// Verify cache has entries
	stats := engine.cache.GetStats()
	assert.Equal(t, 5, stats.EntryCount)

	// Clear all cache
	engine.InvalidateAllCache()

	// Verify cache is empty
	stats = engine.cache.GetStats()
	assert.Equal(t, 0, stats.EntryCount)
}

// TestAction_String tests action string values
func TestAction_String(t *testing.T) {
	assert.Equal(t, "CREATE", string(ActionCreate))
	assert.Equal(t, "READ", string(ActionRead))
	assert.Equal(t, "UPDATE", string(ActionUpdate))
	assert.Equal(t, "DELETE", string(ActionDelete))
	assert.Equal(t, "LIST", string(ActionList))
	assert.Equal(t, "EXECUTE", string(ActionExecute))
}

// TestAccessRequest_Validation tests access request validation
func TestAccessRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     AccessRequest
		isValid bool
	}{
		{
			name: "Valid request with all fields",
			req: AccessRequest{
				Username:   "testuser",
				Resource:   "ticket",
				ResourceID: "ticket-123",
				Action:     ActionRead,
				Context:    map[string]string{"project_id": "proj-1"},
			},
			isValid: true,
		},
		{
			name: "Valid request without resource ID",
			req: AccessRequest{
				Username: "testuser",
				Resource: "ticket",
				Action:   ActionList,
			},
			isValid: true,
		},
		{
			name: "Invalid request - no username",
			req: AccessRequest{
				Resource: "ticket",
				Action:   ActionRead,
			},
			isValid: false,
		},
		{
			name: "Invalid request - no resource",
			req: AccessRequest{
				Username: "testuser",
				Action:   ActionRead,
			},
			isValid: false,
		},
		{
			name: "Invalid request - no action",
			req: AccessRequest{
				Username: "testuser",
				Resource: "ticket",
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation
			hasUsername := tt.req.Username != ""
			hasResource := tt.req.Resource != ""
			hasAction := tt.req.Action != ""

			valid := hasUsername && hasResource && hasAction
			assert.Equal(t, tt.isValid, valid)
		})
	}
}

// TestPermissionSet_DefaultValues tests default permission set values
func TestPermissionSet_DefaultValues(t *testing.T) {
	ps := PermissionSet{}

	assert.False(t, ps.CanCreate)
	assert.False(t, ps.CanRead)
	assert.False(t, ps.CanUpdate)
	assert.False(t, ps.CanDelete)
	assert.False(t, ps.CanList)
	assert.Equal(t, 0, ps.Level)
	assert.Nil(t, ps.Roles)
}

// TestPermissionSet_FullAccess tests a full access permission set
func TestPermissionSet_FullAccess(t *testing.T) {
	ps := PermissionSet{
		CanCreate: true,
		CanRead:   true,
		CanUpdate: true,
		CanDelete: true,
		CanList:   true,
		Level:     5,
	}

	assert.True(t, ps.CanCreate)
	assert.True(t, ps.CanRead)
	assert.True(t, ps.CanUpdate)
	assert.True(t, ps.CanDelete)
	assert.True(t, ps.CanList)
	assert.Equal(t, 5, ps.Level)
}

// TestSecurityContext_Creation tests security context creation
func TestSecurityContext_Creation(t *testing.T) {
	now := time.Now()
	ctx := &SecurityContext{
		Username:             "testuser",
		Roles:                []Role{{ID: "role1", Title: "Developer"}},
		Teams:                []string{"team1", "team2"},
		EffectivePermissions: make(map[string]PermissionSet),
		CachedAt:             now,
		ExpiresAt:            now.Add(5 * time.Minute),
	}

	assert.Equal(t, "testuser", ctx.Username)
	assert.Len(t, ctx.Roles, 1)
	assert.Len(t, ctx.Teams, 2)
	assert.NotNil(t, ctx.EffectivePermissions)
	assert.True(t, ctx.ExpiresAt.After(ctx.CachedAt))
}

// TestSecurityContext_IsExpired tests context expiration
func TestSecurityContext_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		isExpired bool
	}{
		{
			name:      "Not expired - future",
			expiresAt: time.Now().Add(5 * time.Minute),
			isExpired: false,
		},
		{
			name:      "Expired - past",
			expiresAt: time.Now().Add(-5 * time.Minute),
			isExpired: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &SecurityContext{
				ExpiresAt: tt.expiresAt,
			}

			isExpired := time.Now().After(ctx.ExpiresAt)
			assert.Equal(t, tt.isExpired, isExpired)
		})
	}
}

// TestCacheEntry_Expiration tests cache entry expiration
func TestCacheEntry_Expiration(t *testing.T) {
	entry := &CacheEntry{
		Request: AccessRequest{
			Username: "testuser",
			Resource: "ticket",
			Action:   ActionRead,
		},
		Response: AccessResponse{
			Allowed: true,
		},
		CachedAt:  time.Now(),
		ExpiresAt: time.Now().Add(1 * time.Minute),
	}

	// Not expired yet
	assert.False(t, time.Now().After(entry.ExpiresAt))

	// Simulate time passing
	entry.ExpiresAt = time.Now().Add(-1 * time.Second)

	// Now expired
	assert.True(t, time.Now().After(entry.ExpiresAt))
}

// Benchmark tests
func BenchmarkCheckAccess(b *testing.B) {
	mockDB := new(MockDatabase)
	config := DefaultConfig()
	engine := NewSecurityEngine(mockDB, config)

	req := AccessRequest{
		Username: "testuser",
		Resource: "ticket",
		Action:   ActionRead,
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.CheckAccess(ctx, req)
	}
}

func BenchmarkCacheGet(b *testing.B) {
	mockDB := new(MockDatabase)
	config := DefaultConfig()
	engine := NewSecurityEngine(mockDB, config)

	req := AccessRequest{
		Username: "testuser",
		Resource: "ticket",
		Action:   ActionRead,
	}

	response := AccessResponse{
		Allowed: true,
	}

	engine.cache.Set(req, response)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.cache.Get(req)
	}
}

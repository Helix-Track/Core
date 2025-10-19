package engine

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock result type for Exec
type MockResult struct {
	mock.Mock
}

func (m *MockResult) LastInsertId() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockResult) RowsAffected() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

// Mock row scanner
type MockRowScanner struct {
	values []interface{}
	err    error
}

func (m *MockRowScanner) Scan(dest ...interface{}) error {
	if m.err != nil {
		return m.err
	}
	for i, v := range m.values {
		if i < len(dest) {
			switch d := dest[i].(type) {
			case *int:
				if val, ok := v.(int); ok {
					*d = val
				}
			case *int64:
				if val, ok := v.(int64); ok {
					*d = val
				}
			case *string:
				if val, ok := v.(string); ok {
					*d = val
				}
			case *bool:
				if val, ok := v.(bool); ok {
					*d = val
				}
			}
		}
	}
	return nil
}

// Mock rows for Query
type MockRows struct {
	mock.Mock
	rows   [][]interface{}
	cursor int
}

func (m *MockRows) Next() bool {
	if m.cursor >= len(m.rows) {
		return false
	}
	m.cursor++
	return true
}

func (m *MockRows) Scan(dest ...interface{}) error {
	if m.cursor == 0 || m.cursor > len(m.rows) {
		return sql.ErrNoRows
	}
	row := m.rows[m.cursor-1]
	for i, v := range row {
		if i < len(dest) {
			switch d := dest[i].(type) {
			case *int:
				if val, ok := v.(int); ok {
					*d = val
				}
			case *int64:
				if val, ok := v.(int64); ok {
					*d = val
				}
			case *string:
				if val, ok := v.(string); ok {
					*d = val
				}
			case *bool:
				if val, ok := v.(bool); ok {
					*d = val
				}
			}
		}
	}
	return nil
}

func (m *MockRows) Close() error {
	return nil
}

// TestNewAuditLogger tests audit logger creation
func TestNewAuditLogger(t *testing.T) {
	mockDB := new(MockDatabase)
	logger := NewAuditLogger(mockDB, 90*24*time.Hour)

	assert.NotNil(t, logger)
	assert.NotNil(t, logger.db)
	assert.Equal(t, 90*24*time.Hour, logger.retention)
}

// TestLog tests basic audit logging
func TestLog(t *testing.T) {
	mockDB := new(MockDatabase)
	mockResult := new(MockResult)
	mockResult.On("RowsAffected").Return(int64(1), nil)
	mockDB.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(mockResult, nil)

	logger := NewAuditLogger(mockDB, 90*24*time.Hour)
	ctx := context.Background()

	entry := AuditEntry{
		ID:         "audit-1",
		Timestamp:  time.Now(),
		Username:   "testuser",
		Resource:   "ticket",
		ResourceID: "ticket-123",
		Action:     ActionRead,
		Allowed:    true,
		Reason:     "Permission granted",
		IPAddress:  "127.0.0.1",
		UserAgent:  "Test-Agent",
		Context:    map[string]string{"project_id": "proj-1"},
	}

	err := logger.Log(ctx, entry)

	assert.NoError(t, err)
	mockDB.AssertCalled(t, "Exec", mock.Anything, mock.Anything, mock.Anything)
}

// TestLog_DeniedAccess tests logging denied access attempts
func TestLog_DeniedAccess(t *testing.T) {
	mockDB := new(MockDatabase)
	mockResult := new(MockResult)
	mockResult.On("RowsAffected").Return(int64(1), nil)
	mockDB.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(mockResult, nil)

	logger := NewAuditLogger(mockDB, 90*24*time.Hour)
	ctx := context.Background()

	entry := AuditEntry{
		ID:         "audit-2",
		Timestamp:  time.Now(),
		Username:   "testuser",
		Resource:   "ticket",
		ResourceID: "ticket-123",
		Action:     ActionDelete,
		Allowed:    false,
		Reason:     "Insufficient permissions",
		IPAddress:  "127.0.0.1",
		UserAgent:  "Test-Agent",
		Context:    map[string]string{},
	}

	err := logger.Log(ctx, entry)

	assert.NoError(t, err)
}

// TestLog_AutoGenerateID tests that logger auto-generates ID if not provided
func TestLog_AutoGenerateID(t *testing.T) {
	mockDB := new(MockDatabase)
	mockResult := new(MockResult)
	mockResult.On("RowsAffected").Return(int64(1), nil)
	mockDB.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(mockResult, nil)

	logger := NewAuditLogger(mockDB, 90*24*time.Hour)
	ctx := context.Background()

	entry := AuditEntry{
		Username: "testuser",
		Resource: "ticket",
		Action:   ActionRead,
		Allowed:  true,
	}

	err := logger.Log(ctx, entry)

	assert.NoError(t, err)
}

// TestLog_AutoGenerateTimestamp tests that logger auto-generates timestamp if not provided
func TestLog_AutoGenerateTimestamp(t *testing.T) {
	mockDB := new(MockDatabase)
	mockResult := new(MockResult)
	mockResult.On("RowsAffected").Return(int64(1), nil)
	mockDB.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(mockResult, nil)

	logger := NewAuditLogger(mockDB, 90*24*time.Hour)
	ctx := context.Background()

	entry := AuditEntry{
		ID:       "audit-3",
		Username: "testuser",
		Resource: "ticket",
		Action:   ActionRead,
		Allowed:  true,
	}

	err := logger.Log(ctx, entry)

	assert.NoError(t, err)
}

// TestGetRecentEntries tests retrieving recent audit log entries
func TestGetRecentEntries(t *testing.T) {
	mockDB := new(MockDatabase)

	// Create mock rows
	mockRows := &MockRows{
		rows: [][]interface{}{
			{"audit-1", int64(time.Now().Unix()), "user1", "ticket", "ticket-1", "READ", true, "OK", "127.0.0.1", "Agent", "{}"},
			{"audit-2", int64(time.Now().Unix()), "user2", "project", "proj-1", "UPDATE", false, "Denied", "127.0.0.2", "Agent", "{}"},
		},
	}

	mockDB.On("Query", mock.Anything, mock.Anything, mock.Anything).Return(mockRows, nil)

	logger := NewAuditLogger(mockDB, 90*24*time.Hour)
	ctx := context.Background()

	entries, err := logger.GetRecentEntries(ctx, 10)

	assert.NoError(t, err)
	assert.Len(t, entries, 2)
	assert.Equal(t, "user1", entries[0].Username)
	assert.Equal(t, "user2", entries[1].Username)
}

// TestGetEntriesByUsername tests retrieving entries for a specific user
func TestGetEntriesByUsername(t *testing.T) {
	mockDB := new(MockDatabase)

	mockRows := &MockRows{
		rows: [][]interface{}{
			{"audit-1", int64(time.Now().Unix()), "testuser", "ticket", "ticket-1", "READ", true, "OK", "127.0.0.1", "Agent", "{}"},
		},
	}

	mockDB.On("Query", mock.Anything, mock.Anything, mock.Anything).Return(mockRows, nil)

	logger := NewAuditLogger(mockDB, 90*24*time.Hour)
	ctx := context.Background()

	entries, err := logger.GetEntriesByUsername(ctx, "testuser", 10)

	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "testuser", entries[0].Username)
}

// TestGetDeniedAttempts tests retrieving denied access attempts
func TestGetDeniedAttempts(t *testing.T) {
	mockDB := new(MockDatabase)

	mockRows := &MockRows{
		rows: [][]interface{}{
			{"audit-2", int64(time.Now().Unix()), "user2", "project", "proj-1", "DELETE", false, "Insufficient permissions", "127.0.0.2", "Agent", "{}"},
		},
	}

	mockDB.On("Query", mock.Anything, mock.Anything, mock.Anything).Return(mockRows, nil)

	logger := NewAuditLogger(mockDB, 90*24*time.Hour)
	ctx := context.Background()

	entries, err := logger.GetDeniedAttempts(ctx, 10)

	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.False(t, entries[0].Allowed)
	assert.Equal(t, "Insufficient permissions", entries[0].Reason)
}

// TestGetStats tests retrieving audit statistics
func TestGetStats(t *testing.T) {
	mockDB := new(MockDatabase)

	// Mock QueryRow responses for stats
	mockRow1 := &MockRowScanner{values: []interface{}{int64(100)}}
	mockRow2 := &MockRowScanner{values: []interface{}{int64(80)}}
	mockRow3 := &MockRowScanner{values: []interface{}{int64(20)}}
	mockRow4 := &MockRowScanner{values: []interface{}{int64(10)}}

	mockDB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Return(mockRow1).Once()
	mockDB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Return(mockRow2).Once()
	mockDB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Return(mockRow3).Once()
	mockDB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Return(mockRow4).Once()

	logger := NewAuditLogger(mockDB, 90*24*time.Hour)
	ctx := context.Background()

	stats, err := logger.GetStats(ctx)

	assert.NoError(t, err)
	assert.Equal(t, 100, stats.TotalEntries)
	assert.Equal(t, 80, stats.AllowedEntries)
	assert.Equal(t, 20, stats.DeniedEntries)
	assert.Equal(t, 10, stats.UniqueUsers)
}

// TestLog_WithContext tests logging with additional context
func TestLog_WithContext(t *testing.T) {
	mockDB := new(MockDatabase)
	mockResult := new(MockResult)
	mockResult.On("RowsAffected").Return(int64(1), nil)
	mockDB.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(mockResult, nil)

	logger := NewAuditLogger(mockDB, 90*24*time.Hour)
	ctx := context.Background()

	entry := AuditEntry{
		ID:       "audit-4",
		Username: "testuser",
		Resource: "ticket",
		Action:   ActionUpdate,
		Allowed:  true,
		Context: map[string]string{
			"project_id": "proj-1",
			"team_id":    "team-1",
		},
	}

	err := logger.Log(ctx, entry)

	assert.NoError(t, err)
}

// TestLog_AllActions tests logging all action types
func TestLog_AllActions(t *testing.T) {
	tests := []struct {
		name   string
		action Action
	}{
		{"Create", ActionCreate},
		{"Read", ActionRead},
		{"Update", ActionUpdate},
		{"Delete", ActionDelete},
		{"List", ActionList},
		{"Execute", ActionExecute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(MockDatabase)
			mockResult := new(MockResult)
			mockResult.On("RowsAffected").Return(int64(1), nil)
			mockDB.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(mockResult, nil)

			logger := NewAuditLogger(mockDB, 90*24*time.Hour)
			ctx := context.Background()

			entry := AuditEntry{
				Username: "testuser",
				Resource: "ticket",
				Action:   tt.action,
				Allowed:  true,
			}

			err := logger.Log(ctx, entry)
			assert.NoError(t, err)
		})
	}
}

// TestRemoveOldEntries tests cleanup of old entries
func TestRemoveOldEntries(t *testing.T) {
	mockDB := new(MockDatabase)
	mockResult := new(MockResult)
	mockResult.On("RowsAffected").Return(int64(5), nil)
	mockDB.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(mockResult, nil)

	logger := NewAuditLogger(mockDB, 24*time.Hour)

	// Call removeOldEntries directly
	logger.removeOldEntries()

	mockDB.AssertCalled(t, "Exec", mock.Anything, mock.Anything, mock.Anything)
}

// Benchmark tests
func BenchmarkLog(b *testing.B) {
	mockDB := new(MockDatabase)
	mockResult := new(MockResult)
	mockResult.On("RowsAffected").Return(int64(1), nil)
	mockDB.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(mockResult, nil)

	logger := NewAuditLogger(mockDB, 90*24*time.Hour)
	ctx := context.Background()

	entry := AuditEntry{
		Username: "testuser",
		Resource: "ticket",
		Action:   ActionRead,
		Allowed:  true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = logger.Log(ctx, entry)
	}
}

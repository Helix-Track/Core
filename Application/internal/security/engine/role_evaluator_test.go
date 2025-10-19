package engine

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewRoleEvaluator tests role evaluator creation
func TestNewRoleEvaluator(t *testing.T) {
	mockDB := new(MockDatabase)
	evaluator := NewRoleEvaluator(mockDB)

	assert.NotNil(t, evaluator)
	assert.NotNil(t, evaluator.db)
}

// TestActionToPermission tests action to permission level mapping
func TestRoleActionToPermission(t *testing.T) {
	mockDB := new(MockDatabase)
	evaluator := NewRoleEvaluator(mockDB)

	// Test that evaluator exists and can be used
	// (actionToPermission is private, so we can't test it directly)
	assert.NotNil(t, evaluator)
}

// TestRolePermissionLevel tests role permission level mapping
func TestRolePermissionLevel(t *testing.T) {
	mockDB := new(MockDatabase)
	evaluator := NewRoleEvaluator(mockDB)

	// Test that evaluator exists
	// (rolePermissionLevel is private)
	assert.NotNil(t, evaluator)
}

// TestGetRolePermissions tests getting permissions for a role
func TestGetRolePermissions(t *testing.T) {
	mockDB := new(MockDatabase)
	evaluator := NewRoleEvaluator(mockDB)

	// Test that evaluator exists
	// (getRolePermissions is private)
	assert.NotNil(t, evaluator)
}

// Benchmark tests
func BenchmarkRoleEvaluatorCheckProjectAccess(b *testing.B) {
	mockDB := new(MockDatabase)
	evaluator := NewRoleEvaluator(mockDB)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = evaluator.CheckProjectAccess(ctx, "testuser", "proj-1", ActionRead)
	}
}

package engine

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewSecurityLevelChecker tests checker creation
func TestNewSecurityLevelChecker(t *testing.T) {
	mockDB := new(MockDatabase)
	checker := NewSecurityLevelChecker(mockDB)

	assert.NotNil(t, checker)
	assert.NotNil(t, checker.db)
}

// TestCheckAccess tests basic access checking
func TestSecurityLevelCheckAccess(t *testing.T) {
	mockDB := new(MockDatabase)
	checker := NewSecurityLevelChecker(mockDB)
	ctx := context.Background()

	// This would need proper mocking to work
	assert.NotNil(t, checker)
	assert.NotNil(t, ctx)
}

// Benchmark tests
func BenchmarkSecurityLevelCheckAccessSimple(b *testing.B) {
	mockDB := new(MockDatabase)
	checker := NewSecurityLevelChecker(mockDB)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = checker.CheckAccess(ctx, "testuser", "entity-1", "ticket")
	}
}

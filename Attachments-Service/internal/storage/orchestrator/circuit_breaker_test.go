package orchestrator

import (
	"testing"
	"time"
)

func TestNewCircuitBreaker(t *testing.T) {
	cb := NewCircuitBreaker(5, 1*time.Minute)

	if cb.threshold != 5 {
		t.Errorf("expected threshold 5, got %d", cb.threshold)
	}

	if cb.timeout != 1*time.Minute {
		t.Errorf("expected timeout 1 minute, got %v", cb.timeout)
	}

	if cb.state != StateClosed {
		t.Errorf("expected initial state Closed, got %v", cb.state)
	}
}

func TestCircuitBreaker_CanExecute(t *testing.T) {
	t.Run("closed state allows execution", func(t *testing.T) {
		cb := NewCircuitBreaker(5, 1*time.Minute)

		if !cb.CanExecute() {
			t.Error("closed circuit should allow execution")
		}
	})

	t.Run("open state denies execution", func(t *testing.T) {
		cb := NewCircuitBreaker(3, 1*time.Minute)

		// Trigger circuit to open
		for i := 0; i < 3; i++ {
			cb.RecordFailure()
		}

		if cb.CanExecute() {
			t.Error("open circuit should deny execution")
		}
	})

	t.Run("transitions to half-open after timeout", func(t *testing.T) {
		cb := NewCircuitBreaker(2, 100*time.Millisecond)

		// Open the circuit
		cb.RecordFailure()
		cb.RecordFailure()

		if cb.GetState() != StateOpen {
			t.Error("circuit should be open")
		}

		// Wait for timeout
		time.Sleep(150 * time.Millisecond)

		// Should allow one request (half-open)
		if !cb.CanExecute() {
			t.Error("circuit should transition to half-open and allow execution")
		}

		if cb.GetState() != StateHalfOpen {
			t.Errorf("expected half-open state, got %v", cb.GetState())
		}
	})
}

func TestCircuitBreaker_RecordSuccess(t *testing.T) {
	t.Run("resets failure count in closed state", func(t *testing.T) {
		cb := NewCircuitBreaker(5, 1*time.Minute)

		cb.RecordFailure()
		cb.RecordFailure()

		if cb.GetFailures() != 2 {
			t.Errorf("expected 2 failures, got %d", cb.GetFailures())
		}

		cb.RecordSuccess()

		if cb.GetFailures() != 0 {
			t.Error("success should reset failure count")
		}
	})

	t.Run("closes circuit from half-open state", func(t *testing.T) {
		cb := NewCircuitBreaker(2, 100*time.Millisecond)

		// Open the circuit
		cb.RecordFailure()
		cb.RecordFailure()

		// Wait for half-open
		time.Sleep(150 * time.Millisecond)
		cb.CanExecute() // Transition to half-open

		// Record success
		cb.RecordSuccess()

		if cb.GetState() != StateClosed {
			t.Errorf("expected closed state, got %v", cb.GetState())
		}
	})
}

func TestCircuitBreaker_RecordFailure(t *testing.T) {
	t.Run("opens circuit after threshold", func(t *testing.T) {
		cb := NewCircuitBreaker(3, 1*time.Minute)

		// Record failures below threshold
		cb.RecordFailure()
		cb.RecordFailure()

		if cb.GetState() != StateClosed {
			t.Error("circuit should still be closed")
		}

		// Hit threshold
		cb.RecordFailure()

		if cb.GetState() != StateOpen {
			t.Errorf("circuit should be open, got %v", cb.GetState())
		}
	})

	t.Run("increments failure count", func(t *testing.T) {
		cb := NewCircuitBreaker(10, 1*time.Minute)

		for i := 1; i <= 5; i++ {
			cb.RecordFailure()
			if cb.GetFailures() != i {
				t.Errorf("expected %d failures, got %d", i, cb.GetFailures())
			}
		}
	})

	t.Run("reopens circuit from half-open on failure", func(t *testing.T) {
		cb := NewCircuitBreaker(2, 100*time.Millisecond)

		// Open the circuit
		cb.RecordFailure()
		cb.RecordFailure()

		// Transition to half-open
		time.Sleep(150 * time.Millisecond)
		cb.CanExecute()

		// Record failure in half-open
		cb.RecordFailure()

		if cb.GetState() != StateOpen {
			t.Errorf("circuit should reopen, got %v", cb.GetState())
		}
	})
}

func TestCircuitBreaker_GetState(t *testing.T) {
	cb := NewCircuitBreaker(2, 1*time.Minute)

	if cb.GetState() != StateClosed {
		t.Errorf("expected Closed state, got %v", cb.GetState())
	}

	cb.RecordFailure()
	cb.RecordFailure()

	if cb.GetState() != StateOpen {
		t.Errorf("expected Open state, got %v", cb.GetState())
	}
}

func TestCircuitBreaker_Reset(t *testing.T) {
	cb := NewCircuitBreaker(2, 1*time.Minute)

	// Open the circuit
	cb.RecordFailure()
	cb.RecordFailure()

	if cb.GetState() != StateOpen {
		t.Error("circuit should be open")
	}

	// Reset
	cb.Reset()

	if cb.GetState() != StateClosed {
		t.Error("circuit should be closed after reset")
	}

	if cb.GetFailures() != 0 {
		t.Error("failures should be reset to 0")
	}
}

func TestCircuitBreaker_GetStats(t *testing.T) {
	cb := NewCircuitBreaker(5, 2*time.Minute)

	cb.RecordFailure()
	cb.RecordFailure()

	stats := cb.GetStats()

	if stats.State != "closed" {
		t.Errorf("expected state 'closed', got '%s'", stats.State)
	}

	if stats.Failures != 2 {
		t.Errorf("expected 2 failures, got %d", stats.Failures)
	}

	if stats.Threshold != 5 {
		t.Errorf("expected threshold 5, got %d", stats.Threshold)
	}

	if stats.Timeout != 2*time.Minute {
		t.Errorf("expected timeout 2 minutes, got %v", stats.Timeout)
	}
}

func TestCircuitState_String(t *testing.T) {
	tests := []struct {
		state CircuitState
		want  string
	}{
		{StateClosed, "closed"},
		{StateOpen, "open"},
		{StateHalfOpen, "half-open"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.state.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCircuitBreaker_ConcurrentAccess(t *testing.T) {
	cb := NewCircuitBreaker(100, 1*time.Minute)

	// Run concurrent operations
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				cb.CanExecute()
				if j%2 == 0 {
					cb.RecordSuccess()
				} else {
					cb.RecordFailure()
				}
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should not panic and state should be valid
	state := cb.GetState()
	if state != StateClosed && state != StateOpen && state != StateHalfOpen {
		t.Errorf("invalid state after concurrent access: %v", state)
	}
}

func BenchmarkCircuitBreaker_CanExecute(b *testing.B) {
	cb := NewCircuitBreaker(1000, 1*time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb.CanExecute()
	}
}

func BenchmarkCircuitBreaker_RecordSuccess(b *testing.B) {
	cb := NewCircuitBreaker(1000, 1*time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb.RecordSuccess()
	}
}

func BenchmarkCircuitBreaker_RecordFailure(b *testing.B) {
	cb := NewCircuitBreaker(1000, 1*time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb.RecordFailure()
		if i%1000 == 0 {
			cb.Reset() // Reset to avoid staying in open state
		}
	}
}

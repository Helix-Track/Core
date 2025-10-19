package orchestrator

import (
	"sync"
	"time"
)

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	// StateClosed means circuit is closed (normal operation)
	StateClosed CircuitState = iota

	// StateOpen means circuit is open (failing, reject requests)
	StateOpen

	// StateHalfOpen means circuit is half-open (testing if service recovered)
	StateHalfOpen
)

// String returns string representation of circuit state
func (s CircuitState) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	// Configuration
	threshold      int           // Number of failures before opening
	timeout        time.Duration // Time to wait before attempting to close

	// State
	state          CircuitState
	failures       int
	lastFailTime   time.Time
	lastStateChange time.Time
	mu             sync.RWMutex
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(threshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		threshold:       threshold,
		timeout:         timeout,
		state:           StateClosed,
		failures:        0,
		lastStateChange: time.Now(),
	}
}

// CanExecute checks if a request can be executed
func (cb *CircuitBreaker) CanExecute() bool {
	cb.mu.RLock()
	state := cb.state
	lastFailTime := cb.lastFailTime
	cb.mu.RUnlock()

	switch state {
	case StateClosed:
		return true

	case StateOpen:
		// Check if timeout has elapsed
		if time.Since(lastFailTime) > cb.timeout {
			// Transition to half-open
			cb.mu.Lock()
			cb.state = StateHalfOpen
			cb.lastStateChange = time.Now()
			cb.mu.Unlock()
			return true
		}
		return false

	case StateHalfOpen:
		return true

	default:
		return false
	}
}

// RecordSuccess records a successful operation
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures = 0

	// If we were half-open, close the circuit
	if cb.state == StateHalfOpen {
		cb.state = StateClosed
		cb.lastStateChange = time.Now()
	}
}

// RecordFailure records a failed operation
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailTime = time.Now()

	// If we're closed and hit threshold, open the circuit
	if cb.state == StateClosed && cb.failures >= cb.threshold {
		cb.state = StateOpen
		cb.lastStateChange = time.Now()
	}

	// If we're half-open and fail, immediately open
	if cb.state == StateHalfOpen {
		cb.state = StateOpen
		cb.lastStateChange = time.Now()
	}
}

// GetState returns the current state
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return cb.state
}

// GetFailures returns the current failure count
func (cb *CircuitBreaker) GetFailures() int {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return cb.failures
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.failures = 0
	cb.lastStateChange = time.Now()
}

// GetStats returns circuit breaker statistics
func (cb *CircuitBreaker) GetStats() *CircuitBreakerStats {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return &CircuitBreakerStats{
		State:           cb.state.String(),
		Failures:        cb.failures,
		Threshold:       cb.threshold,
		Timeout:         cb.timeout,
		LastFailTime:    cb.lastFailTime,
		LastStateChange: cb.lastStateChange,
		TimeSinceLastStateChange: time.Since(cb.lastStateChange),
	}
}

// CircuitBreakerStats contains circuit breaker statistics
type CircuitBreakerStats struct {
	State                    string
	Failures                 int
	Threshold                int
	Timeout                  time.Duration
	LastFailTime             time.Time
	LastStateChange          time.Time
	TimeSinceLastStateChange time.Duration
}

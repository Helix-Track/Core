package security

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLogSecurityEvent(t *testing.T) {
	// Clear audit log
	ClearAuditLog()

	ip := "192.168.1.1"
	eventType := "TEST_EVENT"
	details := "Test event details"

	LogSecurityEvent(eventType, ip, details)

	// Get recent events
	events := GetRecentEvents(10)

	assert.Equal(t, 1, len(events))
	assert.Equal(t, eventType, events[0].EventType)
	assert.Equal(t, ip, events[0].IP)
	assert.Equal(t, details, events[0].Details)
}

func TestDetermineSeverity(t *testing.T) {
	tests := []struct {
		eventType        string
		expectedSeverity string
	}{
		{"IP_BLOCKED", "CRITICAL"},
		{"BRUTE_FORCE_DETECTED", "CRITICAL"},
		{"SQL_INJECTION", "CRITICAL"},
		{"XSS_ATTEMPT", "CRITICAL"},
		{"CSRF_DETECTED", "CRITICAL"},
		{"RATE_LIMIT_EXCEEDED", "WARNING"},
		{"SUSPICIOUS_ACTIVITY", "WARNING"},
		{"REQUEST_TOO_LARGE", "WARNING"},
		{"UNKNOWN_EVENT", "INFO"},
	}

	for _, tt := range tests {
		t.Run(tt.eventType, func(t *testing.T) {
			severity := determineSeverity(tt.eventType)
			assert.Equal(t, tt.expectedSeverity, severity)
		})
	}
}

func TestDetermineAction(t *testing.T) {
	tests := []struct {
		eventType      string
		expectedAction string
	}{
		{"IP_BLOCKED", "BLOCKED"},
		{"REQUEST_BLOCKED", "BLOCKED"},
		{"SQL_INJECTION", "BLOCKED"},
		{"XSS_ATTEMPT", "BLOCKED"},
		{"SUSPICIOUS_ACTIVITY", "SUSPICIOUS"},
		{"INVALID_INPUT", "SUSPICIOUS"},
		{"UNKNOWN_EVENT", "ALLOWED"},
	}

	for _, tt := range tests {
		t.Run(tt.eventType, func(t *testing.T) {
			action := determineAction(tt.eventType)
			assert.Equal(t, tt.expectedAction, action)
		})
	}
}

func TestGetRecentEvents(t *testing.T) {
	// Clear audit log
	ClearAuditLog()

	// Log 10 events
	for i := 0; i < 10; i++ {
		LogSecurityEvent("TEST_EVENT", "192.168.1.1", "Test event")
	}

	// Get 5 recent events
	events := GetRecentEvents(5)

	assert.Equal(t, 5, len(events))
}

func TestGetEventsByIP(t *testing.T) {
	// Clear audit log
	ClearAuditLog()

	// Log events from different IPs
	LogSecurityEvent("TEST_EVENT", "192.168.1.1", "Event 1")
	LogSecurityEvent("TEST_EVENT", "192.168.1.2", "Event 2")
	LogSecurityEvent("TEST_EVENT", "192.168.1.1", "Event 3")
	LogSecurityEvent("TEST_EVENT", "192.168.1.3", "Event 4")

	// Get events for specific IP
	events := GetEventsByIP("192.168.1.1", 10)

	assert.Equal(t, 2, len(events))
	for _, event := range events {
		assert.Equal(t, "192.168.1.1", event.IP)
	}
}

func TestGetEventsByType(t *testing.T) {
	// Clear audit log
	ClearAuditLog()

	// Log different event types
	LogSecurityEvent("SQL_INJECTION", "192.168.1.1", "SQL injection attempt")
	LogSecurityEvent("XSS_ATTEMPT", "192.168.1.2", "XSS attempt")
	LogSecurityEvent("SQL_INJECTION", "192.168.1.3", "SQL injection attempt 2")

	// Get events by type
	events := GetEventsByType("SQL_INJECTION", 10)

	assert.Equal(t, 2, len(events))
	for _, event := range events {
		assert.Equal(t, "SQL_INJECTION", event.EventType)
	}
}

func TestClearAuditLog(t *testing.T) {
	// Log some events
	for i := 0; i < 5; i++ {
		LogSecurityEvent("TEST_EVENT", "192.168.1.1", "Test event")
	}

	// Clear log
	ClearAuditLog()

	// Should have no events
	events := GetRecentEvents(10)
	assert.Equal(t, 0, len(events))
}

func TestGetSecurityStatistics(t *testing.T) {
	// Clear audit log
	ClearAuditLog()

	// Log various events
	LogSecurityEvent("IP_BLOCKED", "192.168.1.1", "Blocked") // CRITICAL, BLOCKED
	LogSecurityEvent("SQL_INJECTION", "192.168.1.2", "SQL")  // CRITICAL, BLOCKED
	LogSecurityEvent("RATE_LIMIT_EXCEEDED", "192.168.1.3", "Rate limit") // WARNING, BLOCKED
	LogSecurityEvent("SUSPICIOUS_ACTIVITY", "192.168.1.1", "Suspicious") // WARNING, SUSPICIOUS

	stats := GetSecurityStatistics(10)

	assert.Equal(t, 4, stats.TotalEvents)
	assert.Equal(t, 2, stats.CriticalEvents)
	assert.Equal(t, 2, stats.WarningEvents)
	assert.Equal(t, 3, stats.BlockedEvents)
	assert.Equal(t, 1, stats.SuspiciousEvents)
	assert.Equal(t, 3, stats.UniqueIPs) // 3 different IPs
	assert.Equal(t, 4, len(stats.RecentEvents))
}

func TestRegisterCallback(t *testing.T) {
	// Clear audit log
	ClearAuditLog()

	callbackCalled := false
	var receivedEvent SecurityEvent

	// Register callback
	RegisterCallback(func(event SecurityEvent) {
		callbackCalled = true
		receivedEvent = event
	})

	// Log event
	LogSecurityEvent("TEST_EVENT", "192.168.1.1", "Test callback")

	// Give goroutine time to execute
	time.Sleep(10 * time.Millisecond)

	assert.True(t, callbackCalled)
	assert.Equal(t, "TEST_EVENT", receivedEvent.EventType)
	assert.Equal(t, "192.168.1.1", receivedEvent.IP)
}

func TestMaxEventsLimit(t *testing.T) {
	// Clear audit log
	ClearAuditLog()

	// Log more than max events (default is 10000)
	// We'll just log 100 for this test
	for i := 0; i < 100; i++ {
		LogSecurityEvent("TEST_EVENT", "192.168.1.1", "Test event")
	}

	events := GetRecentEvents(200)

	// Should have exactly 100 events
	assert.Equal(t, 100, len(events))
}

func TestSecurityEventFields(t *testing.T) {
	// Clear audit log
	ClearAuditLog()

	ip := "192.168.1.1"
	eventType := "SQL_INJECTION"
	details := "Attempted SQL injection"

	LogSecurityEvent(eventType, ip, details)

	events := GetRecentEvents(1)
	assert.Equal(t, 1, len(events))

	event := events[0]

	// Check all fields
	assert.NotZero(t, event.Timestamp)
	assert.Equal(t, eventType, event.EventType)
	assert.Equal(t, ip, event.IP)
	assert.Equal(t, details, event.Details)
	assert.Equal(t, "CRITICAL", event.Severity)
	assert.Equal(t, "BLOCKED", event.Action)
}

func TestMultipleIPsStatistics(t *testing.T) {
	// Clear audit log
	ClearAuditLog()

	// Log events from 5 different IPs
	for i := 0; i < 5; i++ {
		ip := "192.168.1." + string(rune(100+i))
		LogSecurityEvent("TEST_EVENT", ip, "Test")
	}

	stats := GetSecurityStatistics(10)

	assert.Equal(t, 5, stats.UniqueIPs)
}

func BenchmarkLogSecurityEvent(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LogSecurityEvent("TEST_EVENT", "192.168.1.1", "Test event")
	}
}

func BenchmarkGetRecentEvents(b *testing.B) {
	// Setup: Log some events
	ClearAuditLog()
	for i := 0; i < 100; i++ {
		LogSecurityEvent("TEST_EVENT", "192.168.1.1", "Test event")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetRecentEvents(10)
	}
}

func BenchmarkGetEventsByIP(b *testing.B) {
	// Setup: Log some events
	ClearAuditLog()
	for i := 0; i < 100; i++ {
		LogSecurityEvent("TEST_EVENT", "192.168.1.1", "Test event")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetEventsByIP("192.168.1.1", 10)
	}
}

func BenchmarkGetSecurityStatistics(b *testing.B) {
	// Setup: Log some events
	ClearAuditLog()
	for i := 0; i < 100; i++ {
		LogSecurityEvent("TEST_EVENT", "192.168.1.1", "Test event")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetSecurityStatistics(10)
	}
}

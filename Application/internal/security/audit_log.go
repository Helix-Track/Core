package security

import (
	"fmt"
	"sync"
	"time"
)

// SecurityEvent represents a security-related event
type SecurityEvent struct {
	Timestamp   time.Time
	EventType   string
	IP          string
	UserAgent   string
	Details     string
	Severity    string // INFO, WARNING, CRITICAL
	Action      string // ALLOWED, BLOCKED, SUSPICIOUS
}

// SecurityAuditLog manages security event logging
type SecurityAuditLog struct {
	events    []SecurityEvent
	mu        sync.RWMutex
	maxEvents int
	callbacks []func(SecurityEvent)
}

// Global audit log
var globalAuditLog = &SecurityAuditLog{
	events:    make([]SecurityEvent, 0, 10000),
	maxEvents: 10000,
	callbacks: make([]func(SecurityEvent), 0),
}

// LogSecurityEvent logs a security event
func LogSecurityEvent(eventType, ip, details string) {
	event := SecurityEvent{
		Timestamp: time.Now(),
		EventType: eventType,
		IP:        ip,
		Details:   details,
		Severity:  determineSeverity(eventType),
		Action:    determineAction(eventType),
	}

	globalAuditLog.mu.Lock()
	defer globalAuditLog.mu.Unlock()

	// Add event
	globalAuditLog.events = append(globalAuditLog.events, event)

	// Trim if over limit
	if len(globalAuditLog.events) > globalAuditLog.maxEvents {
		globalAuditLog.events = globalAuditLog.events[len(globalAuditLog.events)-globalAuditLog.maxEvents:]
	}

	// Call callbacks
	for _, callback := range globalAuditLog.callbacks {
		go callback(event)
	}

	// Print critical events
	if event.Severity == "CRITICAL" {
		fmt.Printf("[SECURITY CRITICAL] %s from %s: %s\n", eventType, ip, details)
	}
}

// RegisterCallback registers a callback for security events
func RegisterCallback(callback func(SecurityEvent)) {
	globalAuditLog.mu.Lock()
	defer globalAuditLog.mu.Unlock()
	globalAuditLog.callbacks = append(globalAuditLog.callbacks, callback)
}

// GetRecentEvents returns recent security events
func GetRecentEvents(limit int) []SecurityEvent {
	globalAuditLog.mu.RLock()
	defer globalAuditLog.mu.RUnlock()

	if limit > len(globalAuditLog.events) {
		limit = len(globalAuditLog.events)
	}

	events := make([]SecurityEvent, limit)
	copy(events, globalAuditLog.events[len(globalAuditLog.events)-limit:])
	return events
}

// GetEventsByIP returns events for a specific IP
func GetEventsByIP(ip string, limit int) []SecurityEvent {
	globalAuditLog.mu.RLock()
	defer globalAuditLog.mu.RUnlock()

	var events []SecurityEvent
	for i := len(globalAuditLog.events) - 1; i >= 0 && len(events) < limit; i-- {
		if globalAuditLog.events[i].IP == ip {
			events = append(events, globalAuditLog.events[i])
		}
	}
	return events
}

// GetEventsByType returns events of a specific type
func GetEventsByType(eventType string, limit int) []SecurityEvent {
	globalAuditLog.mu.RLock()
	defer globalAuditLog.mu.RUnlock()

	var events []SecurityEvent
	for i := len(globalAuditLog.events) - 1; i >= 0 && len(events) < limit; i-- {
		if globalAuditLog.events[i].EventType == eventType {
			events = append(events, globalAuditLog.events[i])
		}
	}
	return events
}

// ClearAuditLog clears all events
func ClearAuditLog() {
	globalAuditLog.mu.Lock()
	defer globalAuditLog.mu.Unlock()
	globalAuditLog.events = make([]SecurityEvent, 0, globalAuditLog.maxEvents)
}

// determineSeverity determines event severity
func determineSeverity(eventType string) string {
	critical := map[string]bool{
		"IP_BLOCKED":           true,
		"BRUTE_FORCE_DETECTED": true,
		"SQL_INJECTION":        true,
		"XSS_ATTEMPT":          true,
		"CSRF_DETECTED":        true,
		"MALICIOUS_PAYLOAD":    true,
		"INVALID_TOKEN":        true,
	}

	warning := map[string]bool{
		"REQUEST_BLOCKED":      true,
		"RATE_LIMIT_EXCEEDED":  true,
		"SUSPICIOUS_ACTIVITY":  true,
		"REQUEST_TOO_LARGE":    true,
		"URI_TOO_LONG":         true,
		"INVALID_INPUT":        true,
	}

	if critical[eventType] {
		return "CRITICAL"
	}
	if warning[eventType] {
		return "WARNING"
	}
	return "INFO"
}

// determineAction determines the action taken
func determineAction(eventType string) string {
	blocked := map[string]bool{
		"IP_BLOCKED":           true,
		"REQUEST_BLOCKED":      true,
		"RATE_LIMIT_EXCEEDED":  true,
		"SQL_INJECTION":        true,
		"XSS_ATTEMPT":          true,
		"CSRF_DETECTED":        true,
		"MALICIOUS_PAYLOAD":    true,
	}

	suspicious := map[string]bool{
		"SUSPICIOUS_ACTIVITY":  true,
		"INVALID_INPUT":        true,
		"REQUEST_TOO_LARGE":    true,
		"URI_TOO_LONG":         true,
	}

	if blocked[eventType] {
		return "BLOCKED"
	}
	if suspicious[eventType] {
		return "SUSPICIOUS"
	}
	return "ALLOWED"
}

// SecurityStatistics contains security statistics
type SecurityStatistics struct {
	TotalEvents       int
	CriticalEvents    int
	WarningEvents     int
	InfoEvents        int
	BlockedEvents     int
	SuspiciousEvents  int
	AllowedEvents     int
	UniqueIPs         int
	RecentEvents      []SecurityEvent
}

// GetSecurityStatistics returns security statistics
func GetSecurityStatistics(recentCount int) *SecurityStatistics {
	globalAuditLog.mu.RLock()
	defer globalAuditLog.mu.RUnlock()

	stats := &SecurityStatistics{
		TotalEvents:  len(globalAuditLog.events),
		RecentEvents: make([]SecurityEvent, 0),
	}

	uniqueIPs := make(map[string]bool)

	for _, event := range globalAuditLog.events {
		uniqueIPs[event.IP] = true

		switch event.Severity {
		case "CRITICAL":
			stats.CriticalEvents++
		case "WARNING":
			stats.WarningEvents++
		case "INFO":
			stats.InfoEvents++
		}

		switch event.Action {
		case "BLOCKED":
			stats.BlockedEvents++
		case "SUSPICIOUS":
			stats.SuspiciousEvents++
		case "ALLOWED":
			stats.AllowedEvents++
		}
	}

	stats.UniqueIPs = len(uniqueIPs)

	// Get recent events
	if recentCount > 0 {
		start := len(globalAuditLog.events) - recentCount
		if start < 0 {
			start = 0
		}
		stats.RecentEvents = make([]SecurityEvent, len(globalAuditLog.events[start:]))
		copy(stats.RecentEvents, globalAuditLog.events[start:])
	}

	return stats
}

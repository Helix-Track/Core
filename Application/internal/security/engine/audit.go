package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"helixtrack.ru/core/internal/database"
	"helixtrack.ru/core/internal/logger"
)

// AuditLogger logs security-related events to the database
type AuditLogger struct {
	db        database.Database
	retention time.Duration
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(db database.Database, retention time.Duration) *AuditLogger {
	auditor := &AuditLogger{
		db:        db,
		retention: retention,
	}

	// Start background cleanup goroutine
	go auditor.cleanupOldEntries()

	return auditor
}

// Log creates an audit log entry for a security event
func (al *AuditLogger) Log(ctx context.Context, entry AuditEntry) error {
	// Generate ID if not provided
	if entry.ID == "" {
		entry.ID = uuid.New().String()
	}

	// Set timestamp if not provided
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	// Serialize context to JSON
	contextJSON, err := json.Marshal(entry.Context)
	if err != nil {
		logger.Error("Failed to marshal audit context", zap.Error(err))
		contextJSON = []byte("{}")
	}

	// Determine severity based on allowed flag
	severity := "INFO"
	if !entry.Allowed {
		severity = "WARNING"
	}

	// Insert audit log entry
	query := `
		INSERT INTO audit (
			id, timestamp, username, resource, resource_id, action,
			allowed, reason, ip_address, user_agent, context_data, severity
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = al.db.Exec(ctx, query,
		entry.ID,
		entry.Timestamp.Unix(),
		entry.Username,
		entry.Resource,
		entry.ResourceID,
		string(entry.Action),
		entry.Allowed,
		entry.Reason,
		entry.IPAddress,
		entry.UserAgent,
		string(contextJSON),
		severity,
	)

	if err != nil {
		logger.Error("Failed to create audit log entry",
			zap.Error(err),
			zap.String("username", entry.Username),
			zap.String("resource", entry.Resource),
		)
		return fmt.Errorf("failed to create audit log entry: %w", err)
	}

	logger.Debug("Audit log entry created",
		zap.String("id", entry.ID),
		zap.String("username", entry.Username),
		zap.String("resource", entry.Resource),
		zap.Bool("allowed", entry.Allowed),
	)

	return nil
}

// GetRecentEntries retrieves recent audit log entries
func (al *AuditLogger) GetRecentEntries(ctx context.Context, limit int) ([]AuditEntry, error) {
	query := `
		SELECT id, timestamp, username, resource, resource_id, action,
		       allowed, reason, ip_address, user_agent, context_data
		FROM audit
		WHERE timestamp > ?
		ORDER BY timestamp DESC
		LIMIT ?
	`

	cutoff := time.Now().Add(-24 * time.Hour).Unix()
	rows, err := al.db.Query(ctx, query, cutoff, limit)
	if err != nil {
		logger.Error("Failed to query audit entries", zap.Error(err))
		return nil, fmt.Errorf("failed to query audit entries: %w", err)
	}
	defer rows.Close()

	entries := make([]AuditEntry, 0)
	for rows.Next() {
		var entry AuditEntry
		var timestamp int64
		var contextJSON string
		var action string

		err := rows.Scan(
			&entry.ID,
			&timestamp,
			&entry.Username,
			&entry.Resource,
			&entry.ResourceID,
			&action,
			&entry.Allowed,
			&entry.Reason,
			&entry.IPAddress,
			&entry.UserAgent,
			&contextJSON,
		)

		if err != nil {
			logger.Error("Failed to scan audit entry", zap.Error(err))
			continue
		}

		entry.Timestamp = time.Unix(timestamp, 0)
		entry.Action = Action(action)

		// Deserialize context
		if contextJSON != "" {
			var context map[string]string
			if err := json.Unmarshal([]byte(contextJSON), &context); err == nil {
				entry.Context = context
			}
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// GetEntriesByUsername retrieves audit entries for a specific user
func (al *AuditLogger) GetEntriesByUsername(ctx context.Context, username string, limit int) ([]AuditEntry, error) {
	query := `
		SELECT id, timestamp, username, resource, resource_id, action,
		       allowed, reason, ip_address, user_agent, context_data
		FROM audit
		WHERE username = ? AND timestamp > ?
		ORDER BY timestamp DESC
		LIMIT ?
	`

	cutoff := time.Now().Add(-24 * time.Hour).Unix()
	rows, err := al.db.Query(ctx, query, username, cutoff, limit)
	if err != nil {
		logger.Error("Failed to query user audit entries",
			zap.Error(err),
			zap.String("username", username),
		)
		return nil, fmt.Errorf("failed to query user audit entries: %w", err)
	}
	defer rows.Close()

	entries := make([]AuditEntry, 0)
	for rows.Next() {
		var entry AuditEntry
		var timestamp int64
		var contextJSON string
		var action string

		err := rows.Scan(
			&entry.ID,
			&timestamp,
			&entry.Username,
			&entry.Resource,
			&entry.ResourceID,
			&action,
			&entry.Allowed,
			&entry.Reason,
			&entry.IPAddress,
			&entry.UserAgent,
			&contextJSON,
		)

		if err != nil {
			logger.Error("Failed to scan audit entry", zap.Error(err))
			continue
		}

		entry.Timestamp = time.Unix(timestamp, 0)
		entry.Action = Action(action)

		// Deserialize context
		if contextJSON != "" {
			var context map[string]string
			if err := json.Unmarshal([]byte(contextJSON), &context); err == nil {
				entry.Context = context
			}
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// GetDeniedAttempts retrieves audit entries where access was denied
func (al *AuditLogger) GetDeniedAttempts(ctx context.Context, limit int) ([]AuditEntry, error) {
	query := `
		SELECT id, timestamp, username, resource, resource_id, action,
		       allowed, reason, ip_address, user_agent, context_data
		FROM audit
		WHERE allowed = 0 AND timestamp > ?
		ORDER BY timestamp DESC
		LIMIT ?
	`

	cutoff := time.Now().Add(-24 * time.Hour).Unix()
	rows, err := al.db.Query(ctx, query, cutoff, limit)
	if err != nil {
		logger.Error("Failed to query denied attempts", zap.Error(err))
		return nil, fmt.Errorf("failed to query denied attempts: %w", err)
	}
	defer rows.Close()

	entries := make([]AuditEntry, 0)
	for rows.Next() {
		var entry AuditEntry
		var timestamp int64
		var contextJSON string
		var action string

		err := rows.Scan(
			&entry.ID,
			&timestamp,
			&entry.Username,
			&entry.Resource,
			&entry.ResourceID,
			&action,
			&entry.Allowed,
			&entry.Reason,
			&entry.IPAddress,
			&entry.UserAgent,
			&contextJSON,
		)

		if err != nil {
			logger.Error("Failed to scan audit entry", zap.Error(err))
			continue
		}

		entry.Timestamp = time.Unix(timestamp, 0)
		entry.Action = Action(action)

		// Deserialize context
		if contextJSON != "" {
			var context map[string]string
			if err := json.Unmarshal([]byte(contextJSON), &context); err == nil {
				entry.Context = context
			}
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// GetStats returns audit statistics
func (al *AuditLogger) GetStats(ctx context.Context) (AuditStats, error) {
	var stats AuditStats

	// Total entries
	err := al.db.QueryRow(ctx, "SELECT COUNT(*) FROM audit WHERE timestamp > ?",
		time.Now().Add(-24*time.Hour).Unix()).Scan(&stats.TotalEntries)
	if err != nil {
		return stats, err
	}

	// Allowed entries
	err = al.db.QueryRow(ctx, "SELECT COUNT(*) FROM audit WHERE allowed = 1 AND timestamp > ?",
		time.Now().Add(-24*time.Hour).Unix()).Scan(&stats.AllowedEntries)
	if err != nil {
		return stats, err
	}

	// Denied entries
	err = al.db.QueryRow(ctx, "SELECT COUNT(*) FROM audit WHERE allowed = 0 AND timestamp > ?",
		time.Now().Add(-24*time.Hour).Unix()).Scan(&stats.DeniedEntries)
	if err != nil {
		return stats, err
	}

	// Unique users
	err = al.db.QueryRow(ctx, "SELECT COUNT(DISTINCT username) FROM audit WHERE timestamp > ?",
		time.Now().Add(-24*time.Hour).Unix()).Scan(&stats.UniqueUsers)
	if err != nil {
		return stats, err
	}

	return stats, nil
}

// cleanupOldEntries periodically removes old audit entries
func (al *AuditLogger) cleanupOldEntries() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		al.removeOldEntries()
	}
}

// removeOldEntries removes audit entries older than the retention period
func (al *AuditLogger) removeOldEntries() {
	ctx := context.Background()

	cutoff := time.Now().Add(-al.retention).Unix()
	query := `DELETE FROM audit WHERE timestamp < ?`

	result, err := al.db.Exec(ctx, query, cutoff)
	if err != nil {
		logger.Error("Failed to cleanup old audit entries", zap.Error(err))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		logger.Info("Cleaned up old audit entries",
			zap.Int64("count", rowsAffected),
			zap.Time("cutoff", time.Unix(cutoff, 0)),
		)
	}
}

// AuditStats represents audit log statistics
type AuditStats struct {
	TotalEntries   int
	AllowedEntries int
	DeniedEntries  int
	UniqueUsers    int
}

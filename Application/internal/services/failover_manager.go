package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"helixtrack.ru/core/internal/database"
	"helixtrack.ru/core/internal/logger"
	"helixtrack.ru/core/internal/models"
)

// FailoverManager handles automatic failover and failback of services
type FailoverManager struct {
	db                      database.Database
	stabilityCheckCount     int           // Number of consecutive healthy checks before failback
	failbackDelay           time.Duration // Minimum time before attempting failback
	consecutiveHealthChecks map[string]int // Track consecutive healthy checks per service
}

// NewFailoverManager creates a new failover manager
func NewFailoverManager(db database.Database) *FailoverManager {
	return &FailoverManager{
		db:                      db,
		stabilityCheckCount:     3,              // Primary must be healthy for 3 consecutive checks
		failbackDelay:           5 * time.Minute, // Wait at least 5 minutes before failback
		consecutiveHealthChecks: make(map[string]int),
	}
}

// CheckFailoverNeeded checks if a service needs failover and executes it if necessary
func (fm *FailoverManager) CheckFailoverNeeded(serviceID string, isHealthy bool, status models.ServiceStatus) error {
	ctx := context.Background()

	// Get service details
	query := `
		SELECT id, name, type, role, failover_group, is_active, status, last_failover_at
		FROM service_registry
		WHERE id = ? AND deleted = 0
	`

	var service struct {
		ID             string
		Name           string
		Type           string
		Role           string
		FailoverGroup  string
		IsActive       int
		Status         string
		LastFailoverAt int64
	}

	err := fm.db.QueryRow(ctx, query, serviceID).Scan(
		&service.ID,
		&service.Name,
		&service.Type,
		&service.Role,
		&service.FailoverGroup,
		&service.IsActive,
		&service.Status,
		&service.LastFailoverAt,
	)

	if err != nil {
		return fmt.Errorf("failed to get service details: %w", err)
	}

	// Only process services that are part of a failover group
	if service.FailoverGroup == "" {
		return nil
	}

	// Update consecutive health check count
	if isHealthy {
		fm.consecutiveHealthChecks[serviceID]++
	} else {
		fm.consecutiveHealthChecks[serviceID] = 0
	}

	// Check if failover is needed (active service became unhealthy)
	if service.IsActive == 1 && !isHealthy && status == models.ServiceStatusUnhealthy {
		logger.Warn("Active service became unhealthy, initiating failover",
			zap.String("service_id", serviceID),
			zap.String("service_name", service.Name),
			zap.String("failover_group", service.FailoverGroup),
		)

		return fm.executeFailover(service.FailoverGroup, service.Type, serviceID)
	}

	// Check if failback is needed (primary recovered while backup is active)
	if service.Role == "primary" && service.IsActive == 0 && isHealthy {
		consecutiveHealthy := fm.consecutiveHealthChecks[serviceID]
		timeSinceFailover := time.Now().Unix() - service.LastFailoverAt

		// Ensure primary is stable before failing back
		if consecutiveHealthy >= fm.stabilityCheckCount && timeSinceFailover >= int64(fm.failbackDelay.Seconds()) {
			logger.Info("Primary service recovered and stable, initiating failback",
				zap.String("service_id", serviceID),
				zap.String("service_name", service.Name),
				zap.String("failover_group", service.FailoverGroup),
				zap.Int("consecutive_healthy", consecutiveHealthy),
			)

			return fm.executeFailback(service.FailoverGroup, service.Type, serviceID)
		}
	}

	return nil
}

// executeFailover performs failover from unhealthy active service to backup
func (fm *FailoverManager) executeFailover(failoverGroup, serviceType, oldServiceID string) error {
	ctx := context.Background()

	// Find the best backup service in the same failover group
	query := `
		SELECT id, name, url, status, priority
		FROM service_registry
		WHERE failover_group = ?
		  AND type = ?
		  AND role = 'backup'
		  AND status = 'healthy'
		  AND deleted = 0
		  AND is_active = 0
		ORDER BY priority DESC, health_check_count DESC
		LIMIT 1
	`

	var backup struct {
		ID       string
		Name     string
		URL      string
		Status   string
		Priority int
	}

	err := fm.db.QueryRow(ctx, query, failoverGroup, serviceType).Scan(
		&backup.ID,
		&backup.Name,
		&backup.URL,
		&backup.Status,
		&backup.Priority,
	)

	if err != nil {
		logger.Error("No healthy backup service available for failover",
			zap.String("failover_group", failoverGroup),
			zap.String("service_type", serviceType),
			zap.Error(err),
		)
		return fmt.Errorf("no healthy backup available: %w", err)
	}

	// Perform failover in a transaction-like manner
	now := time.Now().Unix()

	// 1. Deactivate old service
	_, err = fm.db.Exec(ctx,
		"UPDATE service_registry SET is_active = 0, last_failover_at = ? WHERE id = ?",
		now, oldServiceID)
	if err != nil {
		return fmt.Errorf("failed to deactivate old service: %w", err)
	}

	// 2. Activate backup service
	_, err = fm.db.Exec(ctx,
		"UPDATE service_registry SET is_active = 1, last_failover_at = ? WHERE id = ?",
		now, backup.ID)
	if err != nil {
		// Rollback: reactivate old service
		fm.db.Exec(ctx, "UPDATE service_registry SET is_active = 1 WHERE id = ?", oldServiceID)
		return fmt.Errorf("failed to activate backup service: %w", err)
	}

	// 3. Record failover event
	event := models.ServiceFailoverEvent{
		ID:             uuid.New().String(),
		FailoverGroup:  failoverGroup,
		ServiceType:    models.ServiceType(serviceType),
		OldServiceID:   oldServiceID,
		NewServiceID:   backup.ID,
		FailoverReason: "Primary service became unhealthy",
		FailoverType:   "failover",
		Timestamp:      time.Now(),
		Automatic:      true,
	}

	_, err = fm.db.Exec(ctx, `
		INSERT INTO service_failover_events (
			id, failover_group, service_type, old_service_id, new_service_id,
			failover_reason, failover_type, timestamp, automatic
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, event.ID, event.FailoverGroup, event.ServiceType, event.OldServiceID,
		event.NewServiceID, event.FailoverReason, event.FailoverType,
		event.Timestamp.Unix(), 1)

	if err != nil {
		logger.Error("Failed to record failover event", zap.Error(err))
		// Don't fail the failover for logging issues
	}

	logger.Info("Failover completed successfully",
		zap.String("failover_group", failoverGroup),
		zap.String("old_service", oldServiceID),
		zap.String("new_service", backup.ID),
		zap.String("backup_name", backup.Name),
		zap.String("backup_url", backup.URL),
	)

	return nil
}

// executeFailback performs failback to primary service when it recovers
func (fm *FailoverManager) executeFailback(failoverGroup, serviceType, primaryServiceID string) error {
	ctx := context.Background()

	// Find the currently active backup
	query := `
		SELECT id, name
		FROM service_registry
		WHERE failover_group = ?
		  AND type = ?
		  AND is_active = 1
		  AND deleted = 0
		LIMIT 1
	`

	var activeBackup struct {
		ID   string
		Name string
	}

	err := fm.db.QueryRow(ctx, query, failoverGroup, serviceType).Scan(
		&activeBackup.ID,
		&activeBackup.Name,
	)

	if err != nil {
		// No active backup found, primary might already be active
		return nil
	}

	// Perform failback
	now := time.Now().Unix()

	// 1. Deactivate backup
	_, err = fm.db.Exec(ctx,
		"UPDATE service_registry SET is_active = 0, last_failover_at = ? WHERE id = ?",
		now, activeBackup.ID)
	if err != nil {
		return fmt.Errorf("failed to deactivate backup: %w", err)
	}

	// 2. Activate primary
	_, err = fm.db.Exec(ctx,
		"UPDATE service_registry SET is_active = 1, last_failover_at = ? WHERE id = ?",
		now, primaryServiceID)
	if err != nil {
		// Rollback: reactivate backup
		fm.db.Exec(ctx, "UPDATE service_registry SET is_active = 1 WHERE id = ?", activeBackup.ID)
		return fmt.Errorf("failed to activate primary: %w", err)
	}

	// 3. Record failback event
	event := models.ServiceFailoverEvent{
		ID:             uuid.New().String(),
		FailoverGroup:  failoverGroup,
		ServiceType:    models.ServiceType(serviceType),
		OldServiceID:   activeBackup.ID,
		NewServiceID:   primaryServiceID,
		FailoverReason: "Primary service recovered and stable",
		FailoverType:   "failback",
		Timestamp:      time.Now(),
		Automatic:      true,
	}

	_, err = fm.db.Exec(ctx, `
		INSERT INTO service_failover_events (
			id, failover_group, service_type, old_service_id, new_service_id,
			failover_reason, failover_type, timestamp, automatic
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, event.ID, event.FailoverGroup, event.ServiceType, event.OldServiceID,
		event.NewServiceID, event.FailoverReason, event.FailoverType,
		event.Timestamp.Unix(), 1)

	if err != nil {
		logger.Error("Failed to record failback event", zap.Error(err))
		// Don't fail the failback for logging issues
	}

	// Reset consecutive health check counter
	fm.consecutiveHealthChecks[primaryServiceID] = 0

	logger.Info("Failback completed successfully",
		zap.String("failover_group", failoverGroup),
		zap.String("backup_service", activeBackup.ID),
		zap.String("primary_service", primaryServiceID),
	)

	return nil
}

// GetFailoverHistory returns recent failover events for a failover group
func (fm *FailoverManager) GetFailoverHistory(failoverGroup string, limit int) ([]models.ServiceFailoverEvent, error) {
	ctx := context.Background()

	query := `
		SELECT id, failover_group, service_type, old_service_id, new_service_id,
		       failover_reason, failover_type, timestamp, automatic
		FROM service_failover_events
		WHERE failover_group = ?
		ORDER BY timestamp DESC
		LIMIT ?
	`

	rows, err := fm.db.Query(ctx, query, failoverGroup, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query failover history: %w", err)
	}
	defer rows.Close()

	var events []models.ServiceFailoverEvent

	for rows.Next() {
		var event models.ServiceFailoverEvent
		var timestamp int64
		var automatic int

		err := rows.Scan(
			&event.ID,
			&event.FailoverGroup,
			&event.ServiceType,
			&event.OldServiceID,
			&event.NewServiceID,
			&event.FailoverReason,
			&event.FailoverType,
			&timestamp,
			&automatic,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan failover event: %w", err)
		}

		event.Timestamp = time.Unix(timestamp, 0)
		event.Automatic = automatic == 1

		events = append(events, event)
	}

	return events, nil
}

// GetActiveService returns the currently active service for a failover group and type
func (fm *FailoverManager) GetActiveService(failoverGroup string, serviceType models.ServiceType) (*models.ServiceRegistration, error) {
	ctx := context.Background()

	query := `
		SELECT id, name, type, version, url, health_check_url, status, role,
		       failover_group, is_active, priority, last_health_check
		FROM service_registry
		WHERE failover_group = ?
		  AND type = ?
		  AND is_active = 1
		  AND deleted = 0
		LIMIT 1
	`

	var service models.ServiceRegistration
	var lastHealthCheck int64

	err := fm.db.QueryRow(ctx, query, failoverGroup, serviceType).Scan(
		&service.ID,
		&service.Name,
		&service.Type,
		&service.Version,
		&service.URL,
		&service.HealthCheckURL,
		&service.Status,
		&service.Role,
		&service.FailoverGroup,
		&service.IsActive,
		&service.Priority,
		&lastHealthCheck,
	)

	if err != nil {
		return nil, fmt.Errorf("no active service found: %w", err)
	}

	if lastHealthCheck > 0 {
		service.LastHealthCheck = time.Unix(lastHealthCheck, 0)
	}

	return &service, nil
}

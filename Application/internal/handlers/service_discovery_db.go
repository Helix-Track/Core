package handlers

import (
	"context"

	"go.uber.org/zap"
	"helixtrack.ru/core/internal/database"
	"helixtrack.ru/core/internal/logger"
)

// InitializeServiceDiscoveryTables creates the service discovery and health check tables
func InitializeServiceDiscoveryTables(db database.Database) error {
	ctx := context.Background()

	logger.Info("Initializing service discovery tables")

	// Create service_registry table
	serviceRegistrySchema := `
	CREATE TABLE IF NOT EXISTS service_registry (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		type TEXT NOT NULL,
		version TEXT NOT NULL,
		url TEXT NOT NULL,
		health_check_url TEXT NOT NULL,
		public_key TEXT NOT NULL,
		signature TEXT NOT NULL,
		certificate TEXT,
		status TEXT NOT NULL DEFAULT 'registering',
		role TEXT NOT NULL DEFAULT 'primary',
		failover_group TEXT,
		is_active INTEGER DEFAULT 1,
		priority INTEGER DEFAULT 0,
		metadata TEXT DEFAULT '{}',
		registered_by TEXT NOT NULL,
		registered_at INTEGER NOT NULL,
		last_health_check INTEGER DEFAULT 0,
		health_check_count INTEGER DEFAULT 0,
		failed_health_count INTEGER DEFAULT 0,
		last_failover_at INTEGER DEFAULT 0,
		deleted INTEGER DEFAULT 0,
		UNIQUE(name, type, url)
	);
	`

	if _, err := db.Exec(ctx, serviceRegistrySchema); err != nil {
		logger.Error("Failed to create service_registry table", zap.Error(err))
		return err
	}

	// Create indexes for service_registry
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_service_registry_type ON service_registry(type)",
		"CREATE INDEX IF NOT EXISTS idx_service_registry_status ON service_registry(status)",
		"CREATE INDEX IF NOT EXISTS idx_service_registry_deleted ON service_registry(deleted)",
		"CREATE INDEX IF NOT EXISTS idx_service_registry_type_status ON service_registry(type, status, deleted)",
		"CREATE INDEX IF NOT EXISTS idx_service_registry_priority ON service_registry(priority DESC)",
		"CREATE INDEX IF NOT EXISTS idx_service_registry_failover_group ON service_registry(failover_group)",
		"CREATE INDEX IF NOT EXISTS idx_service_registry_is_active ON service_registry(is_active)",
		"CREATE INDEX IF NOT EXISTS idx_service_registry_role ON service_registry(role)",
		"CREATE INDEX IF NOT EXISTS idx_service_registry_group_active ON service_registry(failover_group, is_active, deleted)",
	}

	for _, indexSQL := range indexes {
		if _, err := db.Exec(ctx, indexSQL); err != nil {
			logger.Error("Failed to create index", zap.Error(err), zap.String("sql", indexSQL))
			return err
		}
	}

	// Create service_health_check table
	healthCheckSchema := `
	CREATE TABLE IF NOT EXISTS service_health_check (
		id TEXT PRIMARY KEY,
		service_id TEXT NOT NULL,
		timestamp INTEGER NOT NULL,
		status TEXT NOT NULL,
		response_time INTEGER NOT NULL,
		status_code INTEGER NOT NULL,
		error_message TEXT,
		checked_by TEXT NOT NULL,
		FOREIGN KEY(service_id) REFERENCES service_registry(id)
	);
	`

	if _, err := db.Exec(ctx, healthCheckSchema); err != nil {
		logger.Error("Failed to create service_health_check table", zap.Error(err))
		return err
	}

	// Create indexes for service_health_check
	healthIndexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_health_check_service ON service_health_check(service_id)",
		"CREATE INDEX IF NOT EXISTS idx_health_check_timestamp ON service_health_check(timestamp DESC)",
		"CREATE INDEX IF NOT EXISTS idx_health_check_status ON service_health_check(status)",
	}

	for _, indexSQL := range healthIndexes {
		if _, err := db.Exec(ctx, indexSQL); err != nil {
			logger.Error("Failed to create health check index", zap.Error(err), zap.String("sql", indexSQL))
			return err
		}
	}

	// Create service_rotation_audit table for tracking rotations
	rotationAuditSchema := `
	CREATE TABLE IF NOT EXISTS service_rotation_audit (
		id TEXT PRIMARY KEY,
		old_service_id TEXT NOT NULL,
		new_service_id TEXT NOT NULL,
		reason TEXT,
		requested_by TEXT NOT NULL,
		rotation_time INTEGER NOT NULL,
		verification_hash TEXT NOT NULL,
		success INTEGER NOT NULL,
		error_message TEXT,
		FOREIGN KEY(old_service_id) REFERENCES service_registry(id),
		FOREIGN KEY(new_service_id) REFERENCES service_registry(id)
	);
	`

	if _, err := db.Exec(ctx, rotationAuditSchema); err != nil {
		logger.Error("Failed to create service_rotation_audit table", zap.Error(err))
		return err
	}

	// Create indexes for service_rotation_audit
	rotationIndexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_rotation_audit_old_service ON service_rotation_audit(old_service_id)",
		"CREATE INDEX IF NOT EXISTS idx_rotation_audit_new_service ON service_rotation_audit(new_service_id)",
		"CREATE INDEX IF NOT EXISTS idx_rotation_audit_time ON service_rotation_audit(rotation_time DESC)",
	}

	for _, indexSQL := range rotationIndexes {
		if _, err := db.Exec(ctx, indexSQL); err != nil {
			logger.Error("Failed to create rotation audit index", zap.Error(err), zap.String("sql", indexSQL))
			return err
		}
	}

	// Create service_failover_events table for tracking automatic failovers
	failoverEventsSchema := `
	CREATE TABLE IF NOT EXISTS service_failover_events (
		id TEXT PRIMARY KEY,
		failover_group TEXT NOT NULL,
		service_type TEXT NOT NULL,
		old_service_id TEXT NOT NULL,
		new_service_id TEXT NOT NULL,
		failover_reason TEXT NOT NULL,
		failover_type TEXT NOT NULL,
		timestamp INTEGER NOT NULL,
		automatic INTEGER NOT NULL,
		FOREIGN KEY(old_service_id) REFERENCES service_registry(id),
		FOREIGN KEY(new_service_id) REFERENCES service_registry(id)
	);
	`

	if _, err := db.Exec(ctx, failoverEventsSchema); err != nil {
		logger.Error("Failed to create service_failover_events table", zap.Error(err))
		return err
	}

	// Create indexes for service_failover_events
	failoverIndexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_failover_events_group ON service_failover_events(failover_group)",
		"CREATE INDEX IF NOT EXISTS idx_failover_events_type ON service_failover_events(service_type)",
		"CREATE INDEX IF NOT EXISTS idx_failover_events_timestamp ON service_failover_events(timestamp DESC)",
		"CREATE INDEX IF NOT EXISTS idx_failover_events_old_service ON service_failover_events(old_service_id)",
		"CREATE INDEX IF NOT EXISTS idx_failover_events_new_service ON service_failover_events(new_service_id)",
	}

	for _, indexSQL := range failoverIndexes {
		if _, err := db.Exec(ctx, indexSQL); err != nil {
			logger.Error("Failed to create failover events index", zap.Error(err), zap.String("sql", indexSQL))
			return err
		}
	}

	logger.Info("Service discovery tables created successfully")
	return nil
}

// SeedDefaultServices seeds default services for development/testing
func SeedDefaultServices(db database.Database) error {
	ctx := context.Background()

	logger.Info("Seeding default services (development mode)")

	// Check if any services already exist
	var count int
	err := db.QueryRow(ctx, "SELECT COUNT(*) FROM service_registry WHERE deleted = 0").Scan(&count)
	if err != nil {
		logger.Error("Failed to check existing services", zap.Error(err))
		return err
	}

	if count > 0 {
		logger.Info("Services already exist, skipping seed", zap.Int("count", count))
		return nil
	}

	// Note: In production, services should be registered via the API with proper signatures
	// This is just for development/testing purposes

	logger.Info("No seed services configured - services must be registered via API")
	return nil
}

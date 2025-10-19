package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/helixtrack/attachments-service/internal/models"
)

// CreateEndpoint creates a new storage endpoint
func (db *DB) CreateEndpoint(ctx context.Context, endpoint *models.StorageEndpoint) error {
	if err := endpoint.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	configJSON, err := endpoint.GetAdapterConfigJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize adapter config: %w", err)
	}

	query := `
		INSERT INTO storage_endpoint (
			id, name, type, role, adapter_config, priority, enabled,
			max_size_bytes, current_size, created, modified
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)
	`

	_, err = db.conn.ExecContext(ctx, query,
		endpoint.ID,
		endpoint.Name,
		endpoint.Type,
		endpoint.Role,
		configJSON,
		endpoint.Priority,
		boolToInt(endpoint.Enabled),
		endpoint.MaxSizeBytes,
		endpoint.CurrentSize,
		endpoint.Created,
		endpoint.Modified,
	)

	if err != nil {
		return fmt.Errorf("failed to create endpoint: %w", err)
	}

	return nil
}

// GetEndpoint retrieves an endpoint by ID
func (db *DB) GetEndpoint(ctx context.Context, id string) (*models.StorageEndpoint, error) {
	query := `
		SELECT
			id, name, type, role, adapter_config, priority, enabled,
			max_size_bytes, current_size, created, modified
		FROM storage_endpoint
		WHERE id = $1
	`

	endpoint := &models.StorageEndpoint{}
	var configJSON string
	var enabled int

	err := db.conn.QueryRowContext(ctx, query, id).Scan(
		&endpoint.ID,
		&endpoint.Name,
		&endpoint.Type,
		&endpoint.Role,
		&configJSON,
		&endpoint.Priority,
		&enabled,
		&endpoint.MaxSizeBytes,
		&endpoint.CurrentSize,
		&endpoint.Created,
		&endpoint.Modified,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("endpoint not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get endpoint: %w", err)
	}

	endpoint.Enabled = intToBool(enabled)

	if err := endpoint.SetAdapterConfigFromJSON(configJSON); err != nil {
		return nil, fmt.Errorf("failed to parse adapter config: %w", err)
	}

	return endpoint, nil
}

// UpdateEndpoint updates an existing endpoint
func (db *DB) UpdateEndpoint(ctx context.Context, endpoint *models.StorageEndpoint) error {
	if err := endpoint.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	configJSON, err := endpoint.GetAdapterConfigJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize adapter config: %w", err)
	}

	query := `
		UPDATE storage_endpoint SET
			name = $2,
			type = $3,
			role = $4,
			adapter_config = $5,
			priority = $6,
			enabled = $7,
			max_size_bytes = $8,
			current_size = $9,
			modified = $10
		WHERE id = $1
	`

	result, err := db.conn.ExecContext(ctx, query,
		endpoint.ID,
		endpoint.Name,
		endpoint.Type,
		endpoint.Role,
		configJSON,
		endpoint.Priority,
		boolToInt(endpoint.Enabled),
		endpoint.MaxSizeBytes,
		endpoint.CurrentSize,
		endpoint.Modified,
	)

	if err != nil {
		return fmt.Errorf("failed to update endpoint: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("endpoint not found: %s", endpoint.ID)
	}

	return nil
}

// DeleteEndpoint deletes an endpoint
func (db *DB) DeleteEndpoint(ctx context.Context, id string) error {
	query := `DELETE FROM storage_endpoint WHERE id = $1`

	result, err := db.conn.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete endpoint: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("endpoint not found: %s", id)
	}

	return nil
}

// ListEndpoints lists endpoints, optionally filtered by role
func (db *DB) ListEndpoints(ctx context.Context, role string) ([]*models.StorageEndpoint, error) {
	var query string
	var args []interface{}

	if role != "" {
		query = `
			SELECT
				id, name, type, role, adapter_config, priority, enabled,
				max_size_bytes, current_size, created, modified
			FROM storage_endpoint
			WHERE role = $1
			ORDER BY priority ASC
		`
		args = append(args, role)
	} else {
		query = `
			SELECT
				id, name, type, role, adapter_config, priority, enabled,
				max_size_bytes, current_size, created, modified
			FROM storage_endpoint
			ORDER BY priority ASC
		`
	}

	rows, err := db.conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list endpoints: %w", err)
	}
	defer rows.Close()

	endpoints := []*models.StorageEndpoint{}
	for rows.Next() {
		endpoint := &models.StorageEndpoint{}
		var configJSON string
		var enabled int

		if err := rows.Scan(
			&endpoint.ID,
			&endpoint.Name,
			&endpoint.Type,
			&endpoint.Role,
			&configJSON,
			&endpoint.Priority,
			&enabled,
			&endpoint.MaxSizeBytes,
			&endpoint.CurrentSize,
			&endpoint.Created,
			&endpoint.Modified,
		); err != nil {
			return nil, fmt.Errorf("failed to scan endpoint: %w", err)
		}

		endpoint.Enabled = intToBool(enabled)

		if err := endpoint.SetAdapterConfigFromJSON(configJSON); err != nil {
			return nil, fmt.Errorf("failed to parse adapter config: %w", err)
		}

		endpoints = append(endpoints, endpoint)
	}

	return endpoints, nil
}

// GetPrimaryEndpoint retrieves the primary storage endpoint
func (db *DB) GetPrimaryEndpoint(ctx context.Context) (*models.StorageEndpoint, error) {
	query := `
		SELECT
			id, name, type, role, adapter_config, priority, enabled,
			max_size_bytes, current_size, created, modified
		FROM storage_endpoint
		WHERE role = $1 AND enabled = $2
		ORDER BY priority ASC
		LIMIT 1
	`

	endpoint := &models.StorageEndpoint{}
	var configJSON string
	var enabled int

	err := db.conn.QueryRowContext(ctx, query, models.RolePrimary, 1).Scan(
		&endpoint.ID,
		&endpoint.Name,
		&endpoint.Type,
		&endpoint.Role,
		&configJSON,
		&endpoint.Priority,
		&enabled,
		&endpoint.MaxSizeBytes,
		&endpoint.CurrentSize,
		&endpoint.Created,
		&endpoint.Modified,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no primary endpoint found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get primary endpoint: %w", err)
	}

	endpoint.Enabled = intToBool(enabled)

	if err := endpoint.SetAdapterConfigFromJSON(configJSON); err != nil {
		return nil, fmt.Errorf("failed to parse adapter config: %w", err)
	}

	return endpoint, nil
}

// RecordHealth records a health check result
func (db *DB) RecordHealth(ctx context.Context, health *models.StorageHealth) error {
	if err := health.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	query := `
		INSERT INTO storage_health (
			endpoint_id, check_time, status, latency_ms, error_message, available_bytes
		) VALUES (
			$1, $2, $3, $4, $5, $6
		)
	`

	_, err := db.conn.ExecContext(ctx, query,
		health.EndpointID,
		health.CheckTime,
		health.Status,
		health.LatencyMs,
		health.ErrorMessage,
		health.AvailableBytes,
	)

	if err != nil {
		return fmt.Errorf("failed to record health: %w", err)
	}

	return nil
}

// GetLatestHealth retrieves the most recent health check for an endpoint
func (db *DB) GetLatestHealth(ctx context.Context, endpointID string) (*models.StorageHealth, error) {
	query := `
		SELECT
			endpoint_id, check_time, status, latency_ms, error_message, available_bytes
		FROM storage_health
		WHERE endpoint_id = $1
		ORDER BY check_time DESC
		LIMIT 1
	`

	health := &models.StorageHealth{}

	err := db.conn.QueryRowContext(ctx, query, endpointID).Scan(
		&health.EndpointID,
		&health.CheckTime,
		&health.Status,
		&health.LatencyMs,
		&health.ErrorMessage,
		&health.AvailableBytes,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no health data found for endpoint: %s", endpointID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest health: %w", err)
	}

	return health, nil
}

// GetHealthHistory retrieves health check history for an endpoint
func (db *DB) GetHealthHistory(ctx context.Context, endpointID string, since time.Time) ([]*models.StorageHealth, error) {
	query := `
		SELECT
			endpoint_id, check_time, status, latency_ms, error_message, available_bytes
		FROM storage_health
		WHERE endpoint_id = $1 AND check_time >= $2
		ORDER BY check_time DESC
	`

	rows, err := db.conn.QueryContext(ctx, query, endpointID, since.Unix())
	if err != nil {
		return nil, fmt.Errorf("failed to get health history: %w", err)
	}
	defer rows.Close()

	history := []*models.StorageHealth{}
	for rows.Next() {
		health := &models.StorageHealth{}

		if err := rows.Scan(
			&health.EndpointID,
			&health.CheckTime,
			&health.Status,
			&health.LatencyMs,
			&health.ErrorMessage,
			&health.AvailableBytes,
		); err != nil {
			return nil, fmt.Errorf("failed to scan health: %w", err)
		}

		history = append(history, health)
	}

	return history, nil
}

// GetQuota retrieves the upload quota for a user
func (db *DB) GetQuota(ctx context.Context, userID string) (*models.UploadQuota, error) {
	query := `
		SELECT
			user_id, max_bytes, used_bytes, max_files, used_files, created, modified
		FROM upload_quota
		WHERE user_id = $1
	`

	quota := &models.UploadQuota{}

	err := db.conn.QueryRowContext(ctx, query, userID).Scan(
		&quota.UserID,
		&quota.MaxBytes,
		&quota.UsedBytes,
		&quota.MaxFiles,
		&quota.UsedFiles,
		&quota.Created,
		&quota.Modified,
	)

	if err == sql.ErrNoRows {
		// Create default quota for new user
		quota = models.NewUploadQuota(userID)
		if err := db.CreateQuota(ctx, quota); err != nil {
			return nil, fmt.Errorf("failed to create default quota: %w", err)
		}
		return quota, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get quota: %w", err)
	}

	return quota, nil
}

// CreateQuota creates a new upload quota
func (db *DB) CreateQuota(ctx context.Context, quota *models.UploadQuota) error {
	if err := quota.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	query := `
		INSERT INTO upload_quota (
			user_id, max_bytes, used_bytes, max_files, used_files, created, modified
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
	`

	_, err := db.conn.ExecContext(ctx, query,
		quota.UserID,
		quota.MaxBytes,
		quota.UsedBytes,
		quota.MaxFiles,
		quota.UsedFiles,
		quota.Created,
		quota.Modified,
	)

	if err != nil {
		return fmt.Errorf("failed to create quota: %w", err)
	}

	return nil
}

// UpdateQuota updates an existing quota
func (db *DB) UpdateQuota(ctx context.Context, quota *models.UploadQuota) error {
	if err := quota.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	query := `
		UPDATE upload_quota SET
			max_bytes = $2,
			used_bytes = $3,
			max_files = $4,
			used_files = $5,
			modified = $6
		WHERE user_id = $1
	`

	result, err := db.conn.ExecContext(ctx, query,
		quota.UserID,
		quota.MaxBytes,
		quota.UsedBytes,
		quota.MaxFiles,
		quota.UsedFiles,
		quota.Modified,
	)

	if err != nil {
		return fmt.Errorf("failed to update quota: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("quota not found for user: %s", quota.UserID)
	}

	return nil
}

// IncrementQuotaUsage atomically increments quota usage
func (db *DB) IncrementQuotaUsage(ctx context.Context, userID string, bytes int64, files int) error {
	query := `
		UPDATE upload_quota
		SET used_bytes = used_bytes + $2,
		    used_files = used_files + $3,
		    modified = $4
		WHERE user_id = $1
		  AND used_bytes + $2 <= max_bytes
		  AND used_files + $3 <= max_files
	`

	result, err := db.conn.ExecContext(ctx, query, userID, bytes, files, ctx.Value("timestamp"))
	if err != nil {
		return fmt.Errorf("failed to increment quota usage: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("quota exceeded or user not found: %s", userID)
	}

	return nil
}

// DecrementQuotaUsage atomically decrements quota usage
func (db *DB) DecrementQuotaUsage(ctx context.Context, userID string, bytes int64, files int) error {
	query := `
		UPDATE upload_quota
		SET used_bytes = GREATEST(0, used_bytes - $2),
		    used_files = GREATEST(0, used_files - $3),
		    modified = $4
		WHERE user_id = $1
	`

	result, err := db.conn.ExecContext(ctx, query, userID, bytes, files, ctx.Value("timestamp"))
	if err != nil {
		return fmt.Errorf("failed to decrement quota usage: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user not found: %s", userID)
	}

	return nil
}

// CheckQuotaAvailable checks if user has quota available
func (db *DB) CheckQuotaAvailable(ctx context.Context, userID string, bytes int64) (bool, error) {
	query := `
		SELECT
			CASE
				WHEN used_bytes + $2 <= max_bytes AND used_files + 1 <= max_files
				THEN 1
				ELSE 0
			END AS available
		FROM upload_quota
		WHERE user_id = $1
	`

	var available int
	err := db.conn.QueryRowContext(ctx, query, userID, bytes).Scan(&available)

	if err == sql.ErrNoRows {
		// New user, create default quota
		quota := models.NewUploadQuota(userID)
		if err := db.CreateQuota(ctx, quota); err != nil {
			return false, fmt.Errorf("failed to create default quota: %w", err)
		}
		return quota.CanUpload(bytes), nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check quota: %w", err)
	}

	return available == 1, nil
}

// GetUserStorageUsage retrieves aggregated storage usage for a user
func (db *DB) GetUserStorageUsage(ctx context.Context, userID string) (*models.UserStorageUsage, error) {
	quota, err := db.GetQuota(ctx, userID)
	if err != nil {
		return nil, err
	}

	return models.NewUserStorageUsage(quota), nil
}

// GetTotalStorageUsage retrieves total storage usage across all files
func (db *DB) GetTotalStorageUsage(ctx context.Context) (int64, error) {
	query := `SELECT COALESCE(SUM(size_bytes), 0) FROM attachment_file WHERE deleted = 0`

	var total int64
	if err := db.conn.QueryRowContext(ctx, query).Scan(&total); err != nil {
		return 0, fmt.Errorf("failed to get total storage usage: %w", err)
	}

	return total, nil
}

// GetStorageStats retrieves overall storage statistics
func (db *DB) GetStorageStats(ctx context.Context) (*models.StorageStats, error) {
	query := `
		SELECT
			COUNT(*) AS total_files,
			COALESCE(SUM(size_bytes), 0) AS total_size_bytes,
			COALESCE(SUM(ref_count), 0) AS total_references,
			COUNT(CASE WHEN ref_count = 1 THEN 1 END) AS unique_files,
			COUNT(CASE WHEN ref_count > 1 THEN 1 END) AS shared_files,
			COUNT(CASE WHEN ref_count = 0 THEN 1 END) AS orphaned_files,
			COUNT(CASE WHEN virus_scan_status = 'pending' THEN 1 END) AS pending_scans,
			COUNT(CASE WHEN virus_scan_status = 'infected' THEN 1 END) AS infected_files
		FROM attachment_file
		WHERE deleted = 0
	`

	stats := &models.StorageStats{}

	err := db.conn.QueryRowContext(ctx, query).Scan(
		&stats.TotalFiles,
		&stats.TotalSizeBytes,
		&stats.TotalReferences,
		&stats.UniqueFiles,
		&stats.SharedFiles,
		&stats.OrphanedFiles,
		&stats.PendingScans,
		&stats.InfectedFiles,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get storage stats: %w", err)
	}

	// Calculate deduplication rate
	if stats.TotalReferences > 0 && stats.TotalFiles > 0 {
		savedFiles := stats.TotalReferences - stats.TotalFiles
		stats.DeduplicationRate = (float64(savedFiles) / float64(stats.TotalReferences)) * 100
	}

	return stats, nil
}

// LogAccess creates an access log entry
func (db *DB) LogAccess(ctx context.Context, log *models.AccessLog) error {
	if err := log.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	query := `
		INSERT INTO access_log (
			id, reference_id, file_hash, user_id, ip_address, action,
			status_code, error_message, user_agent, timestamp
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
	`

	_, err := db.conn.ExecContext(ctx, query,
		log.ID,
		log.ReferenceID,
		log.FileHash,
		log.UserID,
		log.IPAddress,
		log.Action,
		log.StatusCode,
		log.ErrorMessage,
		log.UserAgent,
		log.Timestamp,
	)

	if err != nil {
		return fmt.Errorf("failed to log access: %w", err)
	}

	return nil
}

// GetAccessLogs retrieves access logs with filtering
func (db *DB) GetAccessLogs(ctx context.Context, filter *AccessLogFilter) ([]*models.AccessLog, int64, error) {
	// Build where clause
	conditions := []string{"1=1"}
	args := []interface{}{}
	argIdx := 1

	if filter.UserID != "" {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIdx))
		args = append(args, filter.UserID)
		argIdx++
	}

	if filter.Action != "" {
		conditions = append(conditions, fmt.Sprintf("action = $%d", argIdx))
		args = append(args, filter.Action)
		argIdx++
	}

	if filter.StartTime > 0 {
		conditions = append(conditions, fmt.Sprintf("timestamp >= $%d", argIdx))
		args = append(args, filter.StartTime)
		argIdx++
	}

	if filter.EndTime > 0 {
		conditions = append(conditions, fmt.Sprintf("timestamp <= $%d", argIdx))
		args = append(args, filter.EndTime)
		argIdx++
	}

	whereClause := "WHERE " + fmt.Sprintf("%s", conditions[0])
	if len(conditions) > 1 {
		whereClause += " AND " + fmt.Sprintf("%s", conditions[1:])
	}

	// Count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM access_log %s", whereClause)
	var total int64
	if err := db.conn.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count access logs: %w", err)
	}

	// Data
	dataQuery := fmt.Sprintf(`
		SELECT
			id, reference_id, file_hash, user_id, ip_address, action,
			status_code, error_message, user_agent, timestamp
		FROM access_log
		%s
		ORDER BY timestamp DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := db.conn.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get access logs: %w", err)
	}
	defer rows.Close()

	logs := []*models.AccessLog{}
	for rows.Next() {
		log := &models.AccessLog{}

		if err := rows.Scan(
			&log.ID,
			&log.ReferenceID,
			&log.FileHash,
			&log.UserID,
			&log.IPAddress,
			&log.Action,
			&log.StatusCode,
			&log.ErrorMessage,
			&log.UserAgent,
			&log.Timestamp,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan access log: %w", err)
		}

		logs = append(logs, log)
	}

	return logs, total, nil
}

// CreatePresignedURL creates a new presigned URL
func (db *DB) CreatePresignedURL(ctx context.Context, url *models.PresignedURL) error {
	if err := url.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	query := `
		INSERT INTO presigned_url (
			token, reference_id, user_id, ip_address, expires_at,
			max_downloads, download_count, created
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)
	`

	_, err := db.conn.ExecContext(ctx, query,
		url.Token,
		url.ReferenceID,
		url.UserID,
		url.IPAddress,
		url.ExpiresAt,
		url.MaxDownloads,
		url.DownloadCount,
		url.Created,
	)

	if err != nil {
		return fmt.Errorf("failed to create presigned URL: %w", err)
	}

	return nil
}

// GetPresignedURL retrieves a presigned URL by token
func (db *DB) GetPresignedURL(ctx context.Context, token string) (*models.PresignedURL, error) {
	query := `
		SELECT
			token, reference_id, user_id, ip_address, expires_at,
			max_downloads, download_count, created
		FROM presigned_url
		WHERE token = $1
	`

	url := &models.PresignedURL{}

	err := db.conn.QueryRowContext(ctx, query, token).Scan(
		&url.Token,
		&url.ReferenceID,
		&url.UserID,
		&url.IPAddress,
		&url.ExpiresAt,
		&url.MaxDownloads,
		&url.DownloadCount,
		&url.Created,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("presigned URL not found: %s", token)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get presigned URL: %w", err)
	}

	return url, nil
}

// IncrementDownloadCount increments the download count for a presigned URL
func (db *DB) IncrementDownloadCount(ctx context.Context, token string) error {
	query := `
		UPDATE presigned_url
		SET download_count = download_count + 1
		WHERE token = $1
		  AND download_count < max_downloads
		  AND expires_at > $2
	`

	result, err := db.conn.ExecContext(ctx, query, token, ctx.Value("timestamp"))
	if err != nil {
		return fmt.Errorf("failed to increment download count: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("presigned URL expired, exhausted, or not found: %s", token)
	}

	return nil
}

// DeleteExpiredPresignedURLs deletes expired presigned URLs
func (db *DB) DeleteExpiredPresignedURLs(ctx context.Context) (int64, error) {
	query := `DELETE FROM presigned_url WHERE expires_at < $1`

	result, err := db.conn.ExecContext(ctx, query, ctx.Value("timestamp"))
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired presigned URLs: %w", err)
	}

	return result.RowsAffected()
}

// CreateCleanupJob creates a new cleanup job
func (db *DB) CreateCleanupJob(ctx context.Context, job *models.CleanupJob) error {
	if err := job.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	query := `
		INSERT INTO cleanup_job (
			id, job_type, started, completed, status,
			items_processed, items_deleted, error_message
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)
	`

	_, err := db.conn.ExecContext(ctx, query,
		job.ID,
		job.JobType,
		job.Started,
		job.Completed,
		job.Status,
		job.ItemsProcessed,
		job.ItemsDeleted,
		job.ErrorMessage,
	)

	if err != nil {
		return fmt.Errorf("failed to create cleanup job: %w", err)
	}

	return nil
}

// UpdateCleanupJob updates an existing cleanup job
func (db *DB) UpdateCleanupJob(ctx context.Context, job *models.CleanupJob) error {
	if err := job.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	query := `
		UPDATE cleanup_job SET
			completed = $2,
			status = $3,
			items_processed = $4,
			items_deleted = $5,
			error_message = $6
		WHERE id = $1
	`

	result, err := db.conn.ExecContext(ctx, query,
		job.ID,
		job.Completed,
		job.Status,
		job.ItemsProcessed,
		job.ItemsDeleted,
		job.ErrorMessage,
	)

	if err != nil {
		return fmt.Errorf("failed to update cleanup job: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("cleanup job not found: %s", job.ID)
	}

	return nil
}

package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/helixtrack/attachments-service/internal/models"
	"github.com/lib/pq"
)

// CreateFile creates a new attachment file record
func (db *DB) CreateFile(ctx context.Context, file *models.AttachmentFile) error {
	if err := file.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	query := `
		INSERT INTO attachment_file (
			hash, size_bytes, mime_type, extension, ref_count,
			storage_primary, storage_backup, storage_mirrors,
			virus_scan_status, virus_scan_date, virus_scan_result,
			created, last_accessed, deleted
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
	`

	var storageMirrors interface{}
	if db.driver == "postgres" {
		storageMirrors = pq.Array(file.StorageMirrors)
	} else {
		// SQLite: store as JSON string
		if len(file.StorageMirrors) > 0 {
			storageMirrors = strings.Join(file.StorageMirrors, ",")
		}
	}

	_, err := db.conn.ExecContext(ctx, query,
		file.Hash,
		file.SizeBytes,
		file.MimeType,
		file.Extension,
		file.RefCount,
		file.StoragePrimary,
		file.StorageBackup,
		storageMirrors,
		file.VirusScanStatus,
		file.VirusScanDate,
		file.VirusScanResult,
		file.Created,
		file.LastAccessed,
		boolToInt(file.Deleted),
	)

	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	return nil
}

// GetFile retrieves a file by hash
func (db *DB) GetFile(ctx context.Context, hash string) (*models.AttachmentFile, error) {
	query := `
		SELECT
			hash, size_bytes, mime_type, extension, ref_count,
			storage_primary, storage_backup, storage_mirrors,
			virus_scan_status, virus_scan_date, virus_scan_result,
			created, last_accessed, deleted
		FROM attachment_file
		WHERE hash = $1 AND deleted = $2
	`

	file := &models.AttachmentFile{}
	var storageMirrors interface{}
	var deleted int

	if db.driver == "postgres" {
		storageMirrors = pq.Array(&file.StorageMirrors)
	} else {
		var mirrors sql.NullString
		storageMirrors = &mirrors
	}

	err := db.conn.QueryRowContext(ctx, query, hash, 0).Scan(
		&file.Hash,
		&file.SizeBytes,
		&file.MimeType,
		&file.Extension,
		&file.RefCount,
		&file.StoragePrimary,
		&file.StorageBackup,
		storageMirrors,
		&file.VirusScanStatus,
		&file.VirusScanDate,
		&file.VirusScanResult,
		&file.Created,
		&file.LastAccessed,
		&deleted,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("file not found: %s", hash)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	file.Deleted = intToBool(deleted)

	// Parse SQLite storage mirrors
	if db.driver == "sqlite3" {
		if mirrors, ok := storageMirrors.(*sql.NullString); ok && mirrors.Valid {
			if mirrors.String != "" {
				file.StorageMirrors = strings.Split(mirrors.String, ",")
			}
		}
	}

	return file, nil
}

// UpdateFile updates an existing file
func (db *DB) UpdateFile(ctx context.Context, file *models.AttachmentFile) error {
	if err := file.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	query := `
		UPDATE attachment_file SET
			size_bytes = $2,
			mime_type = $3,
			extension = $4,
			ref_count = $5,
			storage_primary = $6,
			storage_backup = $7,
			storage_mirrors = $8,
			virus_scan_status = $9,
			virus_scan_date = $10,
			virus_scan_result = $11,
			last_accessed = $12,
			deleted = $13
		WHERE hash = $1
	`

	var storageMirrors interface{}
	if db.driver == "postgres" {
		storageMirrors = pq.Array(file.StorageMirrors)
	} else {
		if len(file.StorageMirrors) > 0 {
			storageMirrors = strings.Join(file.StorageMirrors, ",")
		}
	}

	result, err := db.conn.ExecContext(ctx, query,
		file.Hash,
		file.SizeBytes,
		file.MimeType,
		file.Extension,
		file.RefCount,
		file.StoragePrimary,
		file.StorageBackup,
		storageMirrors,
		file.VirusScanStatus,
		file.VirusScanDate,
		file.VirusScanResult,
		file.LastAccessed,
		boolToInt(file.Deleted),
	)

	if err != nil {
		return fmt.Errorf("failed to update file: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("file not found: %s", file.Hash)
	}

	return nil
}

// DeleteFile soft-deletes a file
func (db *DB) DeleteFile(ctx context.Context, hash string) error {
	query := `UPDATE attachment_file SET deleted = $1 WHERE hash = $2`

	result, err := db.conn.ExecContext(ctx, query, 1, hash)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("file not found: %s", hash)
	}

	return nil
}

// ListFiles lists files with optional filtering
func (db *DB) ListFiles(ctx context.Context, filter *FileFilter) ([]*models.AttachmentFile, int64, error) {
	// Build query
	conditions := []string{"deleted = 0"}
	args := []interface{}{}
	argIdx := 1

	if filter.MimeType != "" {
		conditions = append(conditions, fmt.Sprintf("mime_type = $%d", argIdx))
		args = append(args, filter.MimeType)
		argIdx++
	}

	if filter.MinSize > 0 {
		conditions = append(conditions, fmt.Sprintf("size_bytes >= $%d", argIdx))
		args = append(args, filter.MinSize)
		argIdx++
	}

	if filter.MaxSize > 0 {
		conditions = append(conditions, fmt.Sprintf("size_bytes <= $%d", argIdx))
		args = append(args, filter.MaxSize)
		argIdx++
	}

	if filter.Deleted != nil {
		deletedVal := 0
		if *filter.Deleted {
			deletedVal = 1
		}
		conditions = append(conditions, fmt.Sprintf("deleted = $%d", argIdx))
		args = append(args, deletedVal)
		argIdx++
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM attachment_file WHERE %s", whereClause)
	var total int64
	if err := db.conn.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count files: %w", err)
	}

	// Data query
	sortBy := "created"
	if filter.SortBy != "" {
		sortBy = filter.SortBy
	}
	sortOrder := "DESC"
	if filter.SortOrder == "asc" {
		sortOrder = "ASC"
	}

	dataQuery := fmt.Sprintf(`
		SELECT
			hash, size_bytes, mime_type, extension, ref_count,
			storage_primary, storage_backup, storage_mirrors,
			virus_scan_status, virus_scan_date, virus_scan_result,
			created, last_accessed, deleted
		FROM attachment_file
		WHERE %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sortBy, sortOrder, argIdx, argIdx+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := db.conn.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list files: %w", err)
	}
	defer rows.Close()

	files := []*models.AttachmentFile{}
	for rows.Next() {
		file := &models.AttachmentFile{}
		var storageMirrors interface{}
		var deleted int

		if db.driver == "postgres" {
			storageMirrors = pq.Array(&file.StorageMirrors)
		} else {
			var mirrors sql.NullString
			storageMirrors = &mirrors
		}

		if err := rows.Scan(
			&file.Hash,
			&file.SizeBytes,
			&file.MimeType,
			&file.Extension,
			&file.RefCount,
			&file.StoragePrimary,
			&file.StorageBackup,
			storageMirrors,
			&file.VirusScanStatus,
			&file.VirusScanDate,
			&file.VirusScanResult,
			&file.Created,
			&file.LastAccessed,
			&deleted,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan file: %w", err)
		}

		file.Deleted = intToBool(deleted)

		// Parse SQLite storage mirrors
		if db.driver == "sqlite3" {
			if mirrors, ok := storageMirrors.(*sql.NullString); ok && mirrors.Valid {
				if mirrors.String != "" {
					file.StorageMirrors = strings.Split(mirrors.String, ",")
				}
			}
		}

		files = append(files, file)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating files: %w", err)
	}

	return files, total, nil
}

// IncrementRefCount atomically increments the reference count
func (db *DB) IncrementRefCount(ctx context.Context, hash string) error {
	query := `
		UPDATE attachment_file
		SET ref_count = ref_count + 1,
		    last_accessed = $1
		WHERE hash = $2
	`

	result, err := db.conn.ExecContext(ctx, query, ctx.Value("timestamp"), hash)
	if err != nil {
		return fmt.Errorf("failed to increment ref count: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("file not found: %s", hash)
	}

	return nil
}

// DecrementRefCount atomically decrements the reference count
func (db *DB) DecrementRefCount(ctx context.Context, hash string) error {
	query := `
		UPDATE attachment_file
		SET ref_count = GREATEST(0, ref_count - 1)
		WHERE hash = $1
	`

	result, err := db.conn.ExecContext(ctx, query, hash)
	if err != nil {
		return fmt.Errorf("failed to decrement ref count: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("file not found: %s", hash)
	}

	return nil
}

// GetOrphanedFiles retrieves files with ref_count = 0 older than retention period
func (db *DB) GetOrphanedFiles(ctx context.Context, retentionDays int) ([]*models.AttachmentFile, error) {
	query := `
		SELECT
			hash, size_bytes, mime_type, extension, ref_count,
			storage_primary, storage_backup, storage_mirrors,
			virus_scan_status, virus_scan_date, virus_scan_result,
			created, last_accessed, deleted
		FROM attachment_file
		WHERE ref_count = 0
		  AND deleted = 0
		  AND created < $1
		ORDER BY created ASC
	`

	cutoffTime := ctx.Value("timestamp").(int64) - int64(retentionDays*86400)

	rows, err := db.conn.QueryContext(ctx, query, cutoffTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get orphaned files: %w", err)
	}
	defer rows.Close()

	files := []*models.AttachmentFile{}
	for rows.Next() {
		file := &models.AttachmentFile{}
		var storageMirrors interface{}
		var deleted int

		if db.driver == "postgres" {
			storageMirrors = pq.Array(&file.StorageMirrors)
		} else {
			var mirrors sql.NullString
			storageMirrors = &mirrors
		}

		if err := rows.Scan(
			&file.Hash,
			&file.SizeBytes,
			&file.MimeType,
			&file.Extension,
			&file.RefCount,
			&file.StoragePrimary,
			&file.StorageBackup,
			storageMirrors,
			&file.VirusScanStatus,
			&file.VirusScanDate,
			&file.VirusScanResult,
			&file.Created,
			&file.LastAccessed,
			&deleted,
		); err != nil {
			return nil, fmt.Errorf("failed to scan file: %w", err)
		}

		file.Deleted = intToBool(deleted)

		// Parse SQLite storage mirrors
		if db.driver == "sqlite3" {
			if mirrors, ok := storageMirrors.(*sql.NullString); ok && mirrors.Valid {
				if mirrors.String != "" {
					file.StorageMirrors = strings.Split(mirrors.String, ",")
				}
			}
		}

		files = append(files, file)
	}

	return files, nil
}

// DeleteOrphanedFiles permanently deletes orphaned files
func (db *DB) DeleteOrphanedFiles(ctx context.Context, hashes []string) (int64, error) {
	if len(hashes) == 0 {
		return 0, nil
	}

	query := `DELETE FROM attachment_file WHERE hash = ANY($1) AND ref_count = 0`

	if db.driver == "sqlite3" {
		// SQLite doesn't support ANY, use IN clause
		placeholders := make([]string, len(hashes))
		args := make([]interface{}, len(hashes))
		for i, hash := range hashes {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
			args[i] = hash
		}
		query = fmt.Sprintf("DELETE FROM attachment_file WHERE hash IN (%s) AND ref_count = 0",
			strings.Join(placeholders, ","))

		result, err := db.conn.ExecContext(ctx, query, args...)
		if err != nil {
			return 0, fmt.Errorf("failed to delete orphaned files: %w", err)
		}
		return result.RowsAffected()
	}

	result, err := db.conn.ExecContext(ctx, query, pq.Array(hashes))
	if err != nil {
		return 0, fmt.Errorf("failed to delete orphaned files: %w", err)
	}

	return result.RowsAffected()
}

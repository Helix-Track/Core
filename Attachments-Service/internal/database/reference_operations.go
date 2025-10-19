package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/helixtrack/attachments-service/internal/models"
	"github.com/lib/pq"
)

// CreateReference creates a new attachment reference
func (db *DB) CreateReference(ctx context.Context, ref *models.AttachmentReference) error {
	if err := ref.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	query := `
		INSERT INTO attachment_reference (
			id, file_hash, entity_type, entity_id, filename,
			description, uploader_id, version, tags,
			created, modified, deleted
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)
	`

	var tags interface{}
	if db.driver == "postgres" {
		tags = pq.Array(ref.Tags)
	} else {
		// SQLite: store as JSON string
		if len(ref.Tags) > 0 {
			tags = strings.Join(ref.Tags, ",")
		}
	}

	_, err := db.conn.ExecContext(ctx, query,
		ref.ID,
		ref.FileHash,
		ref.EntityType,
		ref.EntityID,
		ref.Filename,
		ref.Description,
		ref.UploaderID,
		ref.Version,
		tags,
		ref.Created,
		ref.Modified,
		boolToInt(ref.Deleted),
	)

	if err != nil {
		return fmt.Errorf("failed to create reference: %w", err)
	}

	return nil
}

// GetReference retrieves a reference by ID
func (db *DB) GetReference(ctx context.Context, id string) (*models.AttachmentReference, error) {
	query := `
		SELECT
			id, file_hash, entity_type, entity_id, filename,
			description, uploader_id, version, tags,
			created, modified, deleted
		FROM attachment_reference
		WHERE id = $1 AND deleted = $2
	`

	ref := &models.AttachmentReference{}
	var tags interface{}
	var deleted int

	if db.driver == "postgres" {
		tags = pq.Array(&ref.Tags)
	} else {
		var tagsStr sql.NullString
		tags = &tagsStr
	}

	err := db.conn.QueryRowContext(ctx, query, id, 0).Scan(
		&ref.ID,
		&ref.FileHash,
		&ref.EntityType,
		&ref.EntityID,
		&ref.Filename,
		&ref.Description,
		&ref.UploaderID,
		&ref.Version,
		tags,
		&ref.Created,
		&ref.Modified,
		&deleted,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("reference not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get reference: %w", err)
	}

	ref.Deleted = intToBool(deleted)

	// Parse SQLite tags
	if db.driver == "sqlite3" {
		if tagsData, ok := tags.(*sql.NullString); ok && tagsData.Valid {
			if tagsData.String != "" {
				ref.Tags = strings.Split(tagsData.String, ",")
			}
		}
	}

	return ref, nil
}

// UpdateReference updates an existing reference
func (db *DB) UpdateReference(ctx context.Context, ref *models.AttachmentReference) error {
	if err := ref.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	query := `
		UPDATE attachment_reference SET
			filename = $2,
			description = $3,
			version = $4,
			tags = $5,
			modified = $6,
			deleted = $7
		WHERE id = $1
	`

	var tags interface{}
	if db.driver == "postgres" {
		tags = pq.Array(ref.Tags)
	} else {
		if len(ref.Tags) > 0 {
			tags = strings.Join(ref.Tags, ",")
		}
	}

	result, err := db.conn.ExecContext(ctx, query,
		ref.ID,
		ref.Filename,
		ref.Description,
		ref.Version,
		tags,
		ref.Modified,
		boolToInt(ref.Deleted),
	)

	if err != nil {
		return fmt.Errorf("failed to update reference: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("reference not found: %s", ref.ID)
	}

	return nil
}

// DeleteReference hard-deletes a reference
func (db *DB) DeleteReference(ctx context.Context, id string) error {
	query := `DELETE FROM attachment_reference WHERE id = $1`

	result, err := db.conn.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete reference: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("reference not found: %s", id)
	}

	return nil
}

// SoftDeleteReference soft-deletes a reference
func (db *DB) SoftDeleteReference(ctx context.Context, id string) error {
	query := `UPDATE attachment_reference SET deleted = $1, modified = $2 WHERE id = $3`

	result, err := db.conn.ExecContext(ctx, query, 1, ctx.Value("timestamp"), id)
	if err != nil {
		return fmt.Errorf("failed to soft delete reference: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("reference not found: %s", id)
	}

	return nil
}

// ListReferences lists references with optional filtering
func (db *DB) ListReferences(ctx context.Context, filter *ReferenceFilter) ([]*models.AttachmentReference, int64, error) {
	// Build query
	conditions := []string{"deleted = 0"}
	args := []interface{}{}
	argIdx := 1

	if filter.EntityType != "" {
		conditions = append(conditions, fmt.Sprintf("entity_type = $%d", argIdx))
		args = append(args, filter.EntityType)
		argIdx++
	}

	if filter.EntityID != "" {
		conditions = append(conditions, fmt.Sprintf("entity_id = $%d", argIdx))
		args = append(args, filter.EntityID)
		argIdx++
	}

	if filter.UploaderID != "" {
		conditions = append(conditions, fmt.Sprintf("uploader_id = $%d", argIdx))
		args = append(args, filter.UploaderID)
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
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM attachment_reference WHERE %s", whereClause)
	var total int64
	if err := db.conn.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count references: %w", err)
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
			id, file_hash, entity_type, entity_id, filename,
			description, uploader_id, version, tags,
			created, modified, deleted
		FROM attachment_reference
		WHERE %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sortBy, sortOrder, argIdx, argIdx+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := db.conn.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list references: %w", err)
	}
	defer rows.Close()

	references := []*models.AttachmentReference{}
	for rows.Next() {
		ref := &models.AttachmentReference{}
		var tags interface{}
		var deleted int

		if db.driver == "postgres" {
			tags = pq.Array(&ref.Tags)
		} else {
			var tagsStr sql.NullString
			tags = &tagsStr
		}

		if err := rows.Scan(
			&ref.ID,
			&ref.FileHash,
			&ref.EntityType,
			&ref.EntityID,
			&ref.Filename,
			&ref.Description,
			&ref.UploaderID,
			&ref.Version,
			tags,
			&ref.Created,
			&ref.Modified,
			&deleted,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan reference: %w", err)
		}

		ref.Deleted = intToBool(deleted)

		// Parse SQLite tags
		if db.driver == "sqlite3" {
			if tagsData, ok := tags.(*sql.NullString); ok && tagsData.Valid {
				if tagsData.String != "" {
					ref.Tags = strings.Split(tagsData.String, ",")
				}
			}
		}

		references = append(references, ref)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating references: %w", err)
	}

	return references, total, nil
}

// ListReferencesByEntity retrieves all references for a specific entity
func (db *DB) ListReferencesByEntity(ctx context.Context, entityType, entityID string) ([]*models.AttachmentReference, error) {
	query := `
		SELECT
			id, file_hash, entity_type, entity_id, filename,
			description, uploader_id, version, tags,
			created, modified, deleted
		FROM attachment_reference
		WHERE entity_type = $1 AND entity_id = $2 AND deleted = $3
		ORDER BY created DESC
	`

	rows, err := db.conn.QueryContext(ctx, query, entityType, entityID, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list references by entity: %w", err)
	}
	defer rows.Close()

	references := []*models.AttachmentReference{}
	for rows.Next() {
		ref := &models.AttachmentReference{}
		var tags interface{}
		var deleted int

		if db.driver == "postgres" {
			tags = pq.Array(&ref.Tags)
		} else {
			var tagsStr sql.NullString
			tags = &tagsStr
		}

		if err := rows.Scan(
			&ref.ID,
			&ref.FileHash,
			&ref.EntityType,
			&ref.EntityID,
			&ref.Filename,
			&ref.Description,
			&ref.UploaderID,
			&ref.Version,
			tags,
			&ref.Created,
			&ref.Modified,
			&deleted,
		); err != nil {
			return nil, fmt.Errorf("failed to scan reference: %w", err)
		}

		ref.Deleted = intToBool(deleted)

		// Parse SQLite tags
		if db.driver == "sqlite3" {
			if tagsData, ok := tags.(*sql.NullString); ok && tagsData.Valid {
				if tagsData.String != "" {
					ref.Tags = strings.Split(tagsData.String, ",")
				}
			}
		}

		references = append(references, ref)
	}

	return references, nil
}

// ListReferencesByHash retrieves all references for a specific file hash
func (db *DB) ListReferencesByHash(ctx context.Context, hash string) ([]*models.AttachmentReference, error) {
	query := `
		SELECT
			id, file_hash, entity_type, entity_id, filename,
			description, uploader_id, version, tags,
			created, modified, deleted
		FROM attachment_reference
		WHERE file_hash = $1
		ORDER BY created DESC
	`

	rows, err := db.conn.QueryContext(ctx, query, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to list references by hash: %w", err)
	}
	defer rows.Close()

	references := []*models.AttachmentReference{}
	for rows.Next() {
		ref := &models.AttachmentReference{}
		var tags interface{}
		var deleted int

		if db.driver == "postgres" {
			tags = pq.Array(&ref.Tags)
		} else {
			var tagsStr sql.NullString
			tags = &tagsStr
		}

		if err := rows.Scan(
			&ref.ID,
			&ref.FileHash,
			&ref.EntityType,
			&ref.EntityID,
			&ref.Filename,
			&ref.Description,
			&ref.UploaderID,
			&ref.Version,
			tags,
			&ref.Created,
			&ref.Modified,
			&deleted,
		); err != nil {
			return nil, fmt.Errorf("failed to scan reference: %w", err)
		}

		ref.Deleted = intToBool(deleted)

		// Parse SQLite tags
		if db.driver == "sqlite3" {
			if tagsData, ok := tags.(*sql.NullString); ok && tagsData.Valid {
				if tagsData.String != "" {
					ref.Tags = strings.Split(tagsData.String, ",")
				}
			}
		}

		references = append(references, ref)
	}

	return references, nil
}

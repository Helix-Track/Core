package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/helixtrack/localization-service/internal/models"
	"go.uber.org/zap"
)

// CreateLocalization creates a new localization
func (d *PostgresDatabase) CreateLocalization(ctx context.Context, loc *models.Localization) error {
	loc.BeforeCreate()

	if err := loc.Validate(); err != nil {
		return err
	}

	query := `
		INSERT INTO localizations (
			id, key_id, language_id, value, plural_forms, variables,
			version, approved, approved_by, approved_at, created_at, modified_at, deleted
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := d.execContext(ctx, query,
		loc.ID, loc.KeyID, loc.LanguageID, loc.Value, loc.PluralForms, loc.Variables,
		loc.Version, loc.Approved, loc.ApprovedBy, loc.ApprovedAt,
		loc.CreatedAt, loc.ModifiedAt, loc.Deleted,
	)

	if err != nil {
		d.logger.Error("failed to create localization", zap.Error(err))
		return models.ErrDatabase(err)
	}

	d.logger.Info("localization created", zap.String("id", loc.ID))
	return nil
}

// GetLocalizationByID retrieves a localization by ID
func (d *PostgresDatabase) GetLocalizationByID(ctx context.Context, id string) (*models.Localization, error) {
	query := `
		SELECT id, key_id, language_id, value, plural_forms, variables,
		       version, approved, approved_by, approved_at, created_at, modified_at, deleted
		FROM localizations
		WHERE id = $1 AND deleted = FALSE
	`

	loc := &models.Localization{}
	err := d.queryRowContext(ctx, query, id).Scan(
		&loc.ID, &loc.KeyID, &loc.LanguageID, &loc.Value, &loc.PluralForms, &loc.Variables,
		&loc.Version, &loc.Approved, &loc.ApprovedBy, &loc.ApprovedAt,
		&loc.CreatedAt, &loc.ModifiedAt, &loc.Deleted,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrResourceNotFound("localization")
		}
		d.logger.Error("failed to get localization by ID", zap.Error(err))
		return nil, models.ErrDatabase(err)
	}

	return loc, nil
}

// GetLocalizationByKeyAndLanguage retrieves a localization by key ID and language ID
func (d *PostgresDatabase) GetLocalizationByKeyAndLanguage(ctx context.Context, keyID, languageID string) (*models.Localization, error) {
	query := `
		SELECT id, key_id, language_id, value, plural_forms, variables,
		       version, approved, approved_by, approved_at, created_at, modified_at, deleted
		FROM localizations
		WHERE key_id = $1 AND language_id = $2 AND deleted = FALSE
	`

	loc := &models.Localization{}
	err := d.queryRowContext(ctx, query, keyID, languageID).Scan(
		&loc.ID, &loc.KeyID, &loc.LanguageID, &loc.Value, &loc.PluralForms, &loc.Variables,
		&loc.Version, &loc.Approved, &loc.ApprovedBy, &loc.ApprovedAt,
		&loc.CreatedAt, &loc.ModifiedAt, &loc.Deleted,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrResourceNotFound("localization")
		}
		d.logger.Error("failed to get localization by key and language", zap.Error(err))
		return nil, models.ErrDatabase(err)
	}

	return loc, nil
}

// GetLocalizationsByLanguage retrieves all localizations for a language
func (d *PostgresDatabase) GetLocalizationsByLanguage(ctx context.Context, languageID string) ([]*models.Localization, error) {
	query := `
		SELECT id, key_id, language_id, value, plural_forms, variables,
		       version, approved, approved_by, approved_at, created_at, modified_at, deleted
		FROM localizations
		WHERE language_id = $1 AND deleted = FALSE
		ORDER BY key_id ASC
	`

	rows, err := d.queryContext(ctx, query, languageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var localizations []*models.Localization
	for rows.Next() {
		loc := &models.Localization{}
		if err := rows.Scan(
			&loc.ID, &loc.KeyID, &loc.LanguageID, &loc.Value, &loc.PluralForms, &loc.Variables,
			&loc.Version, &loc.Approved, &loc.ApprovedBy, &loc.ApprovedAt,
			&loc.CreatedAt, &loc.ModifiedAt, &loc.Deleted,
		); err != nil {
			d.logger.Error("failed to scan localization", zap.Error(err))
			return nil, models.ErrDatabase(err)
		}
		localizations = append(localizations, loc)
	}

	return localizations, nil
}

// GetLocalizationsByKeyID retrieves all localizations for a key across all languages
func (d *PostgresDatabase) GetLocalizationsByKeyID(ctx context.Context, keyID string) ([]*models.Localization, error) {
	query := `
		SELECT id, key_id, language_id, value, plural_forms, variables,
		       version, approved, approved_by, approved_at, created_at, modified_at, deleted
		FROM localizations
		WHERE key_id = $1 AND deleted = FALSE
		ORDER BY language_id ASC
	`

	rows, err := d.queryContext(ctx, query, keyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var localizations []*models.Localization
	for rows.Next() {
		loc := &models.Localization{}
		if err := rows.Scan(
			&loc.ID, &loc.KeyID, &loc.LanguageID, &loc.Value, &loc.PluralForms, &loc.Variables,
			&loc.Version, &loc.Approved, &loc.ApprovedBy, &loc.ApprovedAt,
			&loc.CreatedAt, &loc.ModifiedAt, &loc.Deleted,
		); err != nil {
			d.logger.Error("failed to scan localization", zap.Error(err))
			return nil, models.ErrDatabase(err)
		}
		localizations = append(localizations, loc)
	}

	return localizations, nil
}

// UpdateLocalization updates an existing localization
func (d *PostgresDatabase) UpdateLocalization(ctx context.Context, loc *models.Localization) error {
	loc.BeforeUpdate()

	if err := loc.Validate(); err != nil {
		return err
	}

	query := `
		UPDATE localizations
		SET value = $1, plural_forms = $2, variables = $3, version = $4,
		    approved = $5, approved_by = $6, approved_at = $7, modified_at = $8
		WHERE id = $9 AND deleted = FALSE
	`

	result, err := d.execContext(ctx, query,
		loc.Value, loc.PluralForms, loc.Variables, loc.Version,
		loc.Approved, loc.ApprovedBy, loc.ApprovedAt, loc.ModifiedAt, loc.ID,
	)

	if err != nil {
		d.logger.Error("failed to update localization", zap.Error(err))
		return models.ErrDatabase(err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return models.ErrResourceNotFound("localization")
	}

	d.logger.Info("localization updated", zap.String("id", loc.ID))
	return nil
}

// DeleteLocalization soft-deletes a localization
func (d *PostgresDatabase) DeleteLocalization(ctx context.Context, id string) error {
	query := `
		UPDATE localizations
		SET deleted = TRUE, modified_at = $1
		WHERE id = $2 AND deleted = FALSE
	`

	result, err := d.execContext(ctx, query, time.Now().Unix(), id)
	if err != nil {
		d.logger.Error("failed to delete localization", zap.Error(err))
		return models.ErrDatabase(err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return models.ErrResourceNotFound("localization")
	}

	d.logger.Info("localization deleted", zap.String("id", id))
	return nil
}

// ApproveLocalization approves a localization
func (d *PostgresDatabase) ApproveLocalization(ctx context.Context, id, username string) error {
	query := `
		UPDATE localizations
		SET approved = TRUE, approved_by = $1, approved_at = $2, modified_at = $3
		WHERE id = $4 AND deleted = FALSE
	`

	now := time.Now().Unix()
	result, err := d.execContext(ctx, query, username, now, now, id)
	if err != nil {
		d.logger.Error("failed to approve localization", zap.Error(err))
		return models.ErrDatabase(err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return models.ErrResourceNotFound("localization")
	}

	d.logger.Info("localization approved", zap.String("id", id), zap.String("approved_by", username))
	return nil
}

// CreateAuditLog creates an audit log entry
func (d *PostgresDatabase) CreateAuditLog(ctx context.Context, action, entityType, entityID, username string, changes interface{}, ipAddress, userAgent string) error {
	query := `
		INSERT INTO localization_audit_log (id, action, entity_type, entity_id, username, changes, ip_address, user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	changesJSON, err := json.Marshal(changes)
	if err != nil {
		d.logger.Error("failed to marshal audit changes", zap.Error(err))
		return err
	}

	_, err = d.execContext(ctx, query,
		models.GenerateUUID(), action, entityType, entityID, username,
		changesJSON, ipAddress, userAgent, time.Now().Unix(),
	)

	if err != nil {
		d.logger.Error("failed to create audit log", zap.Error(err))
		return models.ErrDatabase(err)
	}

	return nil
}

// GetStats retrieves database statistics
func (d *PostgresDatabase) GetStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Count languages
	var langCount int
	err := d.queryRowContext(ctx, "SELECT COUNT(*) FROM languages WHERE deleted = FALSE").Scan(&langCount)
	if err != nil {
		return nil, models.ErrDatabase(err)
	}
	stats["languages_count"] = langCount

	// Count localization keys
	var keyCount int
	err = d.queryRowContext(ctx, "SELECT COUNT(*) FROM localization_keys WHERE deleted = FALSE").Scan(&keyCount)
	if err != nil {
		return nil, models.ErrDatabase(err)
	}
	stats["keys_count"] = keyCount

	// Count localizations
	var locCount int
	err = d.queryRowContext(ctx, "SELECT COUNT(*) FROM localizations WHERE deleted = FALSE").Scan(&locCount)
	if err != nil {
		return nil, models.ErrDatabase(err)
	}
	stats["localizations_count"] = locCount

	// Count approved localizations
	var approvedCount int
	err = d.queryRowContext(ctx, "SELECT COUNT(*) FROM localizations WHERE deleted = FALSE AND approved = TRUE").Scan(&approvedCount)
	if err != nil {
		return nil, models.ErrDatabase(err)
	}
	stats["approved_count"] = approvedCount

	return stats, nil
}

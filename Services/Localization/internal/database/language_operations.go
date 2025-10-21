package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/helixtrack/localization-service/internal/models"
	"go.uber.org/zap"
)

// CreateLanguage creates a new language
func (d *PostgresDatabase) CreateLanguage(ctx context.Context, lang *models.Language) error {
	lang.BeforeCreate()

	if err := lang.Validate(); err != nil {
		return err
	}

	query := `
		INSERT INTO languages (id, code, name, native_name, is_rtl, is_active, is_default, created_at, modified_at, deleted)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := d.execContext(ctx, query,
		lang.ID, lang.Code, lang.Name, lang.NativeName, lang.IsRTL,
		lang.IsActive, lang.IsDefault, lang.CreatedAt, lang.ModifiedAt, lang.Deleted,
	)

	if err != nil {
		d.logger.Error("failed to create language", zap.Error(err))
		return models.ErrDatabase(err)
	}

	d.logger.Info("language created", zap.String("id", lang.ID), zap.String("code", lang.Code))
	return nil
}

// GetLanguageByID retrieves a language by ID
func (d *PostgresDatabase) GetLanguageByID(ctx context.Context, id string) (*models.Language, error) {
	query := `
		SELECT id, code, name, native_name, is_rtl, is_active, is_default, created_at, modified_at, deleted
		FROM languages
		WHERE id = $1 AND deleted = FALSE
	`

	lang := &models.Language{}
	err := d.queryRowContext(ctx, query, id).Scan(
		&lang.ID, &lang.Code, &lang.Name, &lang.NativeName, &lang.IsRTL,
		&lang.IsActive, &lang.IsDefault, &lang.CreatedAt, &lang.ModifiedAt, &lang.Deleted,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrResourceNotFound("language")
		}
		d.logger.Error("failed to get language by ID", zap.Error(err))
		return nil, models.ErrDatabase(err)
	}

	return lang, nil
}

// GetLanguageByCode retrieves a language by code
func (d *PostgresDatabase) GetLanguageByCode(ctx context.Context, code string) (*models.Language, error) {
	query := `
		SELECT id, code, name, native_name, is_rtl, is_active, is_default, created_at, modified_at, deleted
		FROM languages
		WHERE code = $1 AND deleted = FALSE
	`

	lang := &models.Language{}
	err := d.queryRowContext(ctx, query, code).Scan(
		&lang.ID, &lang.Code, &lang.Name, &lang.NativeName, &lang.IsRTL,
		&lang.IsActive, &lang.IsDefault, &lang.CreatedAt, &lang.ModifiedAt, &lang.Deleted,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrResourceNotFound("language")
		}
		d.logger.Error("failed to get language by code", zap.Error(err))
		return nil, models.ErrDatabase(err)
	}

	return lang, nil
}

// GetLanguages retrieves all languages
func (d *PostgresDatabase) GetLanguages(ctx context.Context, activeOnly bool) ([]*models.Language, error) {
	query := `
		SELECT id, code, name, native_name, is_rtl, is_active, is_default, created_at, modified_at, deleted
		FROM languages
		WHERE deleted = FALSE
	`

	if activeOnly {
		query += " AND is_active = TRUE"
	}

	query += " ORDER BY name ASC"

	rows, err := d.queryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var languages []*models.Language
	for rows.Next() {
		lang := &models.Language{}
		if err := rows.Scan(
			&lang.ID, &lang.Code, &lang.Name, &lang.NativeName, &lang.IsRTL,
			&lang.IsActive, &lang.IsDefault, &lang.CreatedAt, &lang.ModifiedAt, &lang.Deleted,
		); err != nil {
			d.logger.Error("failed to scan language", zap.Error(err))
			return nil, models.ErrDatabase(err)
		}
		languages = append(languages, lang)
	}

	return languages, nil
}

// UpdateLanguage updates an existing language
func (d *PostgresDatabase) UpdateLanguage(ctx context.Context, lang *models.Language) error {
	lang.BeforeUpdate()

	if err := lang.Validate(); err != nil {
		return err
	}

	query := `
		UPDATE languages
		SET code = $1, name = $2, native_name = $3, is_rtl = $4, is_active = $5, is_default = $6, modified_at = $7
		WHERE id = $8 AND deleted = FALSE
	`

	result, err := d.execContext(ctx, query,
		lang.Code, lang.Name, lang.NativeName, lang.IsRTL, lang.IsActive, lang.IsDefault, lang.ModifiedAt, lang.ID,
	)

	if err != nil {
		d.logger.Error("failed to update language", zap.Error(err))
		return models.ErrDatabase(err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return models.ErrResourceNotFound("language")
	}

	d.logger.Info("language updated", zap.String("id", lang.ID))
	return nil
}

// DeleteLanguage soft-deletes a language
func (d *PostgresDatabase) DeleteLanguage(ctx context.Context, id string) error {
	query := `
		UPDATE languages
		SET deleted = TRUE, modified_at = $1
		WHERE id = $2 AND deleted = FALSE
	`

	result, err := d.execContext(ctx, query, models.GenerateUUID(), id)
	if err != nil {
		d.logger.Error("failed to delete language", zap.Error(err))
		return models.ErrDatabase(err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return models.ErrResourceNotFound("language")
	}

	d.logger.Info("language deleted", zap.String("id", id))
	return nil
}

// GetDefaultLanguage retrieves the default language
func (d *PostgresDatabase) GetDefaultLanguage(ctx context.Context) (*models.Language, error) {
	query := `
		SELECT id, code, name, native_name, is_rtl, is_active, is_default, created_at, modified_at, deleted
		FROM languages
		WHERE is_default = TRUE AND deleted = FALSE
		LIMIT 1
	`

	lang := &models.Language{}
	err := d.queryRowContext(ctx, query).Scan(
		&lang.ID, &lang.Code, &lang.Name, &lang.NativeName, &lang.IsRTL,
		&lang.IsActive, &lang.IsDefault, &lang.CreatedAt, &lang.ModifiedAt, &lang.Deleted,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrResourceNotFound("default language")
		}
		d.logger.Error("failed to get default language", zap.Error(err))
		return nil, models.ErrDatabase(err)
	}

	return lang, nil
}

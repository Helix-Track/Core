package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/helixtrack/localization-service/internal/models"
	"go.uber.org/zap"
)

// CreateLocalizationKey creates a new localization key
func (d *PostgresDatabase) CreateLocalizationKey(ctx context.Context, key *models.LocalizationKey) error {
	key.BeforeCreate()

	if err := key.Validate(); err != nil {
		return err
	}

	query := `
		INSERT INTO localization_keys (id, key, category, description, context, created_at, modified_at, deleted)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := d.execContext(ctx, query,
		key.ID, key.Key, key.Category, key.Description, key.Context,
		key.CreatedAt, key.ModifiedAt, key.Deleted,
	)

	if err != nil {
		d.logger.Error("failed to create localization key", zap.Error(err))
		return models.ErrDatabase(err)
	}

	d.logger.Info("localization key created", zap.String("id", key.ID), zap.String("key", key.Key))
	return nil
}

// GetLocalizationKeyByID retrieves a localization key by ID
func (d *PostgresDatabase) GetLocalizationKeyByID(ctx context.Context, id string) (*models.LocalizationKey, error) {
	query := `
		SELECT id, key, category, description, context, created_at, modified_at, deleted
		FROM localization_keys
		WHERE id = $1 AND deleted = FALSE
	`

	key := &models.LocalizationKey{}
	err := d.queryRowContext(ctx, query, id).Scan(
		&key.ID, &key.Key, &key.Category, &key.Description, &key.Context,
		&key.CreatedAt, &key.ModifiedAt, &key.Deleted,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrResourceNotFound("localization key")
		}
		d.logger.Error("failed to get localization key by ID", zap.Error(err))
		return nil, models.ErrDatabase(err)
	}

	return key, nil
}

// GetLocalizationKeyByKey retrieves a localization key by key string
func (d *PostgresDatabase) GetLocalizationKeyByKey(ctx context.Context, keyStr string) (*models.LocalizationKey, error) {
	query := `
		SELECT id, key, category, description, context, created_at, modified_at, deleted
		FROM localization_keys
		WHERE key = $1 AND deleted = FALSE
	`

	key := &models.LocalizationKey{}
	err := d.queryRowContext(ctx, query, keyStr).Scan(
		&key.ID, &key.Key, &key.Category, &key.Description, &key.Context,
		&key.CreatedAt, &key.ModifiedAt, &key.Deleted,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrResourceNotFound("localization key")
		}
		d.logger.Error("failed to get localization key by key", zap.Error(err))
		return nil, models.ErrDatabase(err)
	}

	return key, nil
}

// GetLocalizationKeysByCategory retrieves localization keys by category
func (d *PostgresDatabase) GetLocalizationKeysByCategory(ctx context.Context, category string) ([]*models.LocalizationKey, error) {
	query := `
		SELECT id, key, category, description, context, created_at, modified_at, deleted
		FROM localization_keys
		WHERE category = $1 AND deleted = FALSE
		ORDER BY key ASC
	`

	rows, err := d.queryContext(ctx, query, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*models.LocalizationKey
	for rows.Next() {
		key := &models.LocalizationKey{}
		if err := rows.Scan(
			&key.ID, &key.Key, &key.Category, &key.Description, &key.Context,
			&key.CreatedAt, &key.ModifiedAt, &key.Deleted,
		); err != nil {
			d.logger.Error("failed to scan localization key", zap.Error(err))
			return nil, models.ErrDatabase(err)
		}
		keys = append(keys, key)
	}

	return keys, nil
}

// UpdateLocalizationKey updates an existing localization key
func (d *PostgresDatabase) UpdateLocalizationKey(ctx context.Context, key *models.LocalizationKey) error {
	key.BeforeUpdate()

	if err := key.Validate(); err != nil {
		return err
	}

	query := `
		UPDATE localization_keys
		SET key = $1, category = $2, description = $3, context = $4, modified_at = $5
		WHERE id = $6 AND deleted = FALSE
	`

	result, err := d.execContext(ctx, query,
		key.Key, key.Category, key.Description, key.Context, key.ModifiedAt, key.ID,
	)

	if err != nil {
		d.logger.Error("failed to update localization key", zap.Error(err))
		return models.ErrDatabase(err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return models.ErrResourceNotFound("localization key")
	}

	d.logger.Info("localization key updated", zap.String("id", key.ID))
	return nil
}

// DeleteLocalizationKey soft-deletes a localization key
func (d *PostgresDatabase) DeleteLocalizationKey(ctx context.Context, id string) error {
	query := `
		UPDATE localization_keys
		SET deleted = TRUE, modified_at = $1
		WHERE id = $2 AND deleted = FALSE
	`

	result, err := d.execContext(ctx, query, models.GenerateUUID(), id)
	if err != nil {
		d.logger.Error("failed to delete localization key", zap.Error(err))
		return models.ErrDatabase(err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return models.ErrResourceNotFound("localization key")
	}

	d.logger.Info("localization key deleted", zap.String("id", id))
	return nil
}

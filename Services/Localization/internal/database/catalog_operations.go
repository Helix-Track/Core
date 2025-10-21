package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/helixtrack/localization-service/internal/models"
	"go.uber.org/zap"
)

// CreateCatalog creates a new catalog
func (d *PostgresDatabase) CreateCatalog(ctx context.Context, catalog *models.LocalizationCatalog) error {
	catalog.BeforeCreate()

	if err := catalog.Validate(); err != nil {
		return err
	}

	query := `
		INSERT INTO localization_catalogs (id, language_id, category, catalog_data, version, checksum, created_at, modified_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := d.execContext(ctx, query,
		catalog.ID, catalog.LanguageID, catalog.Category, catalog.CatalogData,
		catalog.Version, catalog.Checksum, catalog.CreatedAt, catalog.ModifiedAt,
	)

	if err != nil {
		d.logger.Error("failed to create catalog", zap.Error(err))
		return models.ErrDatabase(err)
	}

	d.logger.Info("catalog created",
		zap.String("id", catalog.ID),
		zap.String("language_id", catalog.LanguageID),
		zap.String("category", catalog.Category),
	)
	return nil
}

// GetCatalogByLanguage retrieves a catalog by language ID and optional category
func (d *PostgresDatabase) GetCatalogByLanguage(ctx context.Context, languageID string, category string) (*models.LocalizationCatalog, error) {
	var query string
	var args []interface{}

	if category != "" {
		query = `
			SELECT id, language_id, category, catalog_data, version, checksum, created_at, modified_at
			FROM localization_catalogs
			WHERE language_id = $1 AND category = $2
			ORDER BY version DESC
			LIMIT 1
		`
		args = []interface{}{languageID, category}
	} else {
		query = `
			SELECT id, language_id, category, catalog_data, version, checksum, created_at, modified_at
			FROM localization_catalogs
			WHERE language_id = $1 AND category IS NULL
			ORDER BY version DESC
			LIMIT 1
		`
		args = []interface{}{languageID}
	}

	catalog := &models.LocalizationCatalog{}
	var categoryNullable sql.NullString

	err := d.queryRowContext(ctx, query, args...).Scan(
		&catalog.ID, &catalog.LanguageID, &categoryNullable, &catalog.CatalogData,
		&catalog.Version, &catalog.Checksum, &catalog.CreatedAt, &catalog.ModifiedAt,
	)

	if categoryNullable.Valid {
		catalog.Category = categoryNullable.String
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Build catalog if it doesn't exist
			d.logger.Info("catalog not found, building new one",
				zap.String("language_id", languageID),
				zap.String("category", category),
			)
			return d.BuildCatalog(ctx, languageID, category)
		}
		d.logger.Error("failed to get catalog by language", zap.Error(err))
		return nil, models.ErrDatabase(err)
	}

	return catalog, nil
}

// GetLatestCatalog retrieves the latest catalog version
func (d *PostgresDatabase) GetLatestCatalog(ctx context.Context, languageID string, category string) (*models.LocalizationCatalog, error) {
	return d.GetCatalogByLanguage(ctx, languageID, category)
}

// UpdateCatalog updates an existing catalog
func (d *PostgresDatabase) UpdateCatalog(ctx context.Context, catalog *models.LocalizationCatalog) error {
	catalog.BeforeUpdate()

	if err := catalog.Validate(); err != nil {
		return err
	}

	query := `
		UPDATE localization_catalogs
		SET catalog_data = $1, version = $2, checksum = $3, modified_at = $4
		WHERE id = $5
	`

	result, err := d.execContext(ctx, query,
		catalog.CatalogData, catalog.Version, catalog.Checksum, catalog.ModifiedAt, catalog.ID,
	)

	if err != nil {
		d.logger.Error("failed to update catalog", zap.Error(err))
		return models.ErrDatabase(err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return models.ErrResourceNotFound("catalog")
	}

	d.logger.Info("catalog updated", zap.String("id", catalog.ID))
	return nil
}

// DeleteCatalog deletes a catalog
func (d *PostgresDatabase) DeleteCatalog(ctx context.Context, id string) error {
	query := `DELETE FROM localization_catalogs WHERE id = $1`

	result, err := d.execContext(ctx, query, id)
	if err != nil {
		d.logger.Error("failed to delete catalog", zap.Error(err))
		return models.ErrDatabase(err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return models.ErrResourceNotFound("catalog")
	}

	d.logger.Info("catalog deleted", zap.String("id", id))
	return nil
}

// BuildCatalog builds a complete localization catalog for a language
func (d *PostgresDatabase) BuildCatalog(ctx context.Context, languageID string, category string) (*models.LocalizationCatalog, error) {
	// Get language to verify it exists
	lang, err := d.GetLanguageByID(ctx, languageID)
	if err != nil {
		return nil, err
	}

	// Build query based on category filter
	var query string
	var args []interface{}

	if category != "" {
		query = `
			SELECT lk.key, l.value
			FROM localizations l
			JOIN localization_keys lk ON l.key_id = lk.id
			WHERE l.language_id = $1 AND lk.category = $2 AND l.deleted = FALSE AND lk.deleted = FALSE
			ORDER BY lk.key ASC
		`
		args = []interface{}{languageID, category}
	} else {
		query = `
			SELECT lk.key, l.value
			FROM localizations l
			JOIN localization_keys lk ON l.key_id = lk.id
			WHERE l.language_id = $1 AND l.deleted = FALSE AND lk.deleted = FALSE
			ORDER BY lk.key ASC
		`
		args = []interface{}{languageID}
	}

	rows, err := d.queryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Build catalog map
	catalogMap := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			d.logger.Error("failed to scan catalog entry", zap.Error(err))
			return nil, models.ErrDatabase(err)
		}
		catalogMap[key] = value
	}

	// Marshal to JSON
	catalogJSON, err := json.Marshal(catalogMap)
	if err != nil {
		d.logger.Error("failed to marshal catalog", zap.Error(err))
		return nil, err
	}

	// Get latest version number or start at 1
	var latestVersion int
	versionQuery := `
		SELECT COALESCE(MAX(version), 0) FROM localization_catalogs
		WHERE language_id = $1 AND category IS NULL
	`
	if category != "" {
		versionQuery = `
			SELECT COALESCE(MAX(version), 0) FROM localization_catalogs
			WHERE language_id = $1 AND category = $2
		`
		err = d.queryRowContext(ctx, versionQuery, languageID, category).Scan(&latestVersion)
	} else {
		err = d.queryRowContext(ctx, versionQuery, languageID).Scan(&latestVersion)
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		d.logger.Error("failed to get latest version", zap.Error(err))
		return nil, models.ErrDatabase(err)
	}

	// Create new catalog
	catalog := &models.LocalizationCatalog{
		LanguageID:  languageID,
		Category:    category,
		CatalogData: catalogJSON,
		Version:     latestVersion + 1,
	}

	catalog.BeforeCreate()

	// Save catalog
	if err := d.CreateCatalog(ctx, catalog); err != nil {
		return nil, err
	}

	d.logger.Info("catalog built",
		zap.String("language", lang.Code),
		zap.String("category", category),
		zap.Int("version", catalog.Version),
		zap.Int("entries", len(catalogMap)),
	)

	return catalog, nil
}

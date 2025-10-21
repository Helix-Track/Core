package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/helixtrack/localization-service/internal/models"
	"go.uber.org/zap"
)

// CreateVersion creates a new localization version
func (d *PostgresDatabase) CreateVersion(ctx context.Context, version *models.LocalizationVersion) error {
	version.BeforeCreate()

	query := `
		INSERT INTO localization_versions (
			id, version_number, version_type, description,
			keys_count, languages_count, total_localizations,
			created_by, created_at, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := d.db.ExecContext(ctx, query,
		version.ID,
		version.VersionNumber,
		version.VersionType,
		version.Description,
		version.KeysCount,
		version.LanguagesCount,
		version.TotalLocalizations,
		version.CreatedBy,
		version.CreatedAt,
		version.Metadata,
	)

	if err != nil {
		return fmt.Errorf("failed to create version: %w", err)
	}

	d.logger.Info("version created",
		zap.String("version_number", version.VersionNumber),
		zap.String("type", version.VersionType),
	)

	return nil
}

// GetVersionByNumber retrieves a version by its version number
func (d *PostgresDatabase) GetVersionByNumber(ctx context.Context, versionNumber string) (*models.LocalizationVersion, error) {
	query := `
		SELECT id, version_number, version_type, description,
		       keys_count, languages_count, total_localizations,
		       created_by, created_at, COALESCE(metadata::TEXT, '') as metadata
		FROM localization_versions
		WHERE version_number = $1
	`

	var version models.LocalizationVersion
	err := d.db.QueryRowContext(ctx, query, versionNumber).Scan(
		&version.ID,
		&version.VersionNumber,
		&version.VersionType,
		&version.Description,
		&version.KeysCount,
		&version.LanguagesCount,
		&version.TotalLocalizations,
		&version.CreatedBy,
		&version.CreatedAt,
		&version.Metadata,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("version not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get version: %w", err)
	}

	return &version, nil
}

// GetVersionByID retrieves a version by its ID
func (d *PostgresDatabase) GetVersionByID(ctx context.Context, id string) (*models.LocalizationVersion, error) {
	query := `
		SELECT id, version_number, version_type, description,
		       keys_count, languages_count, total_localizations,
		       created_by, created_at, COALESCE(metadata::TEXT, '') as metadata
		FROM localization_versions
		WHERE id = $1
	`

	var version models.LocalizationVersion
	err := d.db.QueryRowContext(ctx, query, id).Scan(
		&version.ID,
		&version.VersionNumber,
		&version.VersionType,
		&version.Description,
		&version.KeysCount,
		&version.LanguagesCount,
		&version.TotalLocalizations,
		&version.CreatedBy,
		&version.CreatedAt,
		&version.Metadata,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("version not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get version: %w", err)
	}

	return &version, nil
}

// GetCurrentVersion retrieves the latest version
func (d *PostgresDatabase) GetCurrentVersion(ctx context.Context) (*models.LocalizationVersion, error) {
	query := `
		SELECT id, version_number, version_type, description,
		       keys_count, languages_count, total_localizations,
		       created_by, created_at, COALESCE(metadata::TEXT, '') as metadata
		FROM localization_versions
		ORDER BY created_at DESC
		LIMIT 1
	`

	var version models.LocalizationVersion
	err := d.db.QueryRowContext(ctx, query).Scan(
		&version.ID,
		&version.VersionNumber,
		&version.VersionType,
		&version.Description,
		&version.KeysCount,
		&version.LanguagesCount,
		&version.TotalLocalizations,
		&version.CreatedBy,
		&version.CreatedAt,
		&version.Metadata,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no versions found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get current version: %w", err)
	}

	return &version, nil
}

// ListVersions retrieves all versions with pagination
func (d *PostgresDatabase) ListVersions(ctx context.Context, limit, offset int) ([]*models.LocalizationVersion, error) {
	if limit <= 0 {
		limit = 10
	}

	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT id, version_number, version_type, description,
		       keys_count, languages_count, total_localizations,
		       created_by, created_at, COALESCE(metadata::TEXT, '') as metadata
		FROM localization_versions
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := d.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list versions: %w", err)
	}
	defer rows.Close()

	versions := []*models.LocalizationVersion{}

	for rows.Next() {
		var version models.LocalizationVersion
		err := rows.Scan(
			&version.ID,
			&version.VersionNumber,
			&version.VersionType,
			&version.Description,
			&version.KeysCount,
			&version.LanguagesCount,
			&version.TotalLocalizations,
			&version.CreatedBy,
			&version.CreatedAt,
			&version.Metadata,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan version: %w", err)
		}

		versions = append(versions, &version)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating versions: %w", err)
	}

	return versions, nil
}

// CountVersions counts total number of versions
func (d *PostgresDatabase) CountVersions(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM localization_versions`

	var count int
	err := d.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count versions: %w", err)
	}

	return count, nil
}

// DeleteVersion deletes a version
func (d *PostgresDatabase) DeleteVersion(ctx context.Context, id string) error {
	query := `DELETE FROM localization_versions WHERE id = $1`

	result, err := d.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete version: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("version not found")
	}

	d.logger.Info("version deleted", zap.String("id", id))

	return nil
}

// GetCatalogByVersion retrieves a catalog for a specific version and language
func (d *PostgresDatabase) GetCatalogByVersion(ctx context.Context, versionNumber, languageCode string) (*models.LocalizationCatalog, error) {
	// First get the version
	version, err := d.GetVersionByNumber(ctx, versionNumber)
	if err != nil {
		return nil, err
	}

	// Then get the language
	lang, err := d.GetLanguageByCode(ctx, languageCode)
	if err != nil {
		return nil, err
	}

	// Get the catalog for this version and language
	query := `
		SELECT id, language_id, category, catalog_data::TEXT, version, checksum,
		       created_at, modified_at
		FROM localization_catalogs
		WHERE language_id = $1 AND version_id = $2
		ORDER BY version DESC
		LIMIT 1
	`

	var catalog models.LocalizationCatalog
	err = d.db.QueryRowContext(ctx, query, lang.ID, version.ID).Scan(
		&catalog.ID,
		&catalog.LanguageID,
		&catalog.Category,
		&catalog.CatalogData,
		&catalog.Version,
		&catalog.Checksum,
		&catalog.CreatedAt,
		&catalog.ModifiedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("catalog not found for version %s and language %s", versionNumber, languageCode)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get catalog by version: %w", err)
	}

	return &catalog, nil
}

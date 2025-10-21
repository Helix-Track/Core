package seeder

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/helixtrack/localization-service/internal/database"
	"github.com/helixtrack/localization-service/internal/models"
	"go.uber.org/zap"
)

// SeedLanguage represents a language in the seed data
type SeedLanguage struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	NativeName string `json:"native_name"`
	IsRTL      bool   `json:"is_rtl"`
	IsActive   bool   `json:"is_active"`
	IsDefault  bool   `json:"is_default"`
}

// SeedLocalizationKey represents a localization key in the seed data
type SeedLocalizationKey struct {
	Key         string   `json:"key"`
	Category    string   `json:"category"`
	Description string   `json:"description"`
	Context     string   `json:"context"`
	Variables   []string `json:"variables"`
}

// Seeder handles database seeding from JSON files
type Seeder struct {
	db           database.Database
	logger       *zap.Logger
	seedDataPath string
}

// New creates a new Seeder instance
func New(db database.Database, logger *zap.Logger, seedDataPath string) *Seeder {
	return &Seeder{
		db:           db,
		logger:       logger,
		seedDataPath: seedDataPath,
	}
}

// ShouldSeed checks if the database needs to be seeded
func (s *Seeder) ShouldSeed(ctx context.Context) (bool, error) {
	// Check if any languages exist
	languages, err := s.db.GetLanguages(ctx, true)
	if err != nil {
		return false, fmt.Errorf("failed to check languages: %w", err)
	}

	// If no languages exist, we should seed
	return len(languages) == 0, nil
}

// Seed populates the database with seed data
func (s *Seeder) Seed(ctx context.Context) error {
	s.logger.Info("Starting database seeding")

	startTime := time.Now()

	// Step 1: Load and insert languages
	languages, err := s.seedLanguages(ctx)
	if err != nil {
		return fmt.Errorf("failed to seed languages: %w", err)
	}

	s.logger.Info("Languages seeded successfully",
		zap.Int("count", len(languages)),
	)

	// Step 2: Load and insert localization keys
	keys, err := s.seedLocalizationKeys(ctx)
	if err != nil {
		return fmt.Errorf("failed to seed localization keys: %w", err)
	}

	s.logger.Info("Localization keys seeded successfully",
		zap.Int("count", len(keys)),
	)

	// Step 3: Load and insert localizations for each language
	totalLocalizations := 0
	for _, lang := range languages {
		count, err := s.seedLocalizations(ctx, lang, keys)
		if err != nil {
			s.logger.Warn("Failed to seed localizations for language",
				zap.String("language", lang.Code),
				zap.Error(err),
			)
			continue
		}

		totalLocalizations += count
		s.logger.Info("Localizations seeded for language",
			zap.String("language", lang.Code),
			zap.Int("count", count),
		)
	}

	// Step 4: Build catalogs for each language
	for _, lang := range languages {
		err := s.buildCatalog(ctx, lang)
		if err != nil {
			s.logger.Warn("Failed to build catalog for language",
				zap.String("language", lang.Code),
				zap.Error(err),
			)
			continue
		}
	}

	duration := time.Since(startTime)

	s.logger.Info("Database seeding completed successfully",
		zap.Int("languages", len(languages)),
		zap.Int("keys", len(keys)),
		zap.Int("localizations", totalLocalizations),
		zap.Duration("duration", duration),
	)

	return nil
}

// seedLanguages loads and inserts languages from seed data
func (s *Seeder) seedLanguages(ctx context.Context) ([]*models.Language, error) {
	// Load seed data
	filePath := filepath.Join(s.seedDataPath, "languages.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read languages file: %w", err)
	}

	var seedLanguages []SeedLanguage
	if err := json.Unmarshal(data, &seedLanguages); err != nil {
		return nil, fmt.Errorf("failed to parse languages JSON: %w", err)
	}

	// Convert to model and insert
	languages := make([]*models.Language, 0, len(seedLanguages))
	now := time.Now().Unix()

	for _, sl := range seedLanguages {
		lang := &models.Language{
			ID:         uuid.New().String(),
			Code:       sl.Code,
			Name:       sl.Name,
			NativeName: sl.NativeName,
			IsRTL:      sl.IsRTL,
			IsActive:   sl.IsActive,
			IsDefault:  sl.IsDefault,
			CreatedAt:  now,
			ModifiedAt: now,
			Deleted:    false,
		}

		if err := s.db.CreateLanguage(ctx, lang); err != nil {
			return nil, fmt.Errorf("failed to create language %s: %w", lang.Code, err)
		}

		languages = append(languages, lang)
	}

	return languages, nil
}

// seedLocalizationKeys loads and inserts localization keys from seed data
func (s *Seeder) seedLocalizationKeys(ctx context.Context) ([]*models.LocalizationKey, error) {
	// Load seed data
	filePath := filepath.Join(s.seedDataPath, "localization-keys.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read localization keys file: %w", err)
	}

	var seedKeys []SeedLocalizationKey
	if err := json.Unmarshal(data, &seedKeys); err != nil {
		return nil, fmt.Errorf("failed to parse localization keys JSON: %w", err)
	}

	// Convert to model and insert
	keys := make([]*models.LocalizationKey, 0, len(seedKeys))
	now := time.Now().Unix()

	for _, sk := range seedKeys {
		key := &models.LocalizationKey{
			ID:          uuid.New().String(),
			Key:         sk.Key,
			Category:    sk.Category,
			Description: sk.Description,
			Context:     sk.Context,
			CreatedAt:   now,
			ModifiedAt:  now,
			Deleted:     false,
		}

		if err := s.db.CreateLocalizationKey(ctx, key); err != nil {
			return nil, fmt.Errorf("failed to create localization key %s: %w", key.Key, err)
		}

		keys = append(keys, key)
	}

	return keys, nil
}

// seedLocalizations loads and inserts localizations for a specific language
func (s *Seeder) seedLocalizations(ctx context.Context, lang *models.Language, keys []*models.LocalizationKey) (int, error) {
	// Load seed data for this language
	filePath := filepath.Join(s.seedDataPath, "localizations", fmt.Sprintf("%s.json", lang.Code))
	data, err := os.ReadFile(filePath)
	if err != nil {
		// If file doesn't exist, it's not an error (language might not have translations yet)
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to read localizations file: %w", err)
	}

	var translations map[string]string
	if err := json.Unmarshal(data, &translations); err != nil {
		return 0, fmt.Errorf("failed to parse localizations JSON: %w", err)
	}

	// Create a map for quick key lookup
	keyMap := make(map[string]*models.LocalizationKey)
	for _, key := range keys {
		keyMap[key.Key] = key
	}

	// Insert localizations
	count := 0
	now := time.Now().Unix()

	for keyStr, value := range translations {
		// Find the corresponding key
		key, exists := keyMap[keyStr]
		if !exists {
			s.logger.Warn("Localization key not found in database",
				zap.String("key", keyStr),
				zap.String("language", lang.Code),
			)
			continue
		}

		localization := &models.Localization{
			ID:         uuid.New().String(),
			KeyID:      key.ID,
			LanguageID: lang.ID,
			Value:      value,
			Version:    1,
			Approved:   true, // Auto-approve seed data
			ApprovedBy: "system",
			ApprovedAt: now,
			CreatedAt:  now,
			ModifiedAt: now,
			Deleted:    false,
		}

		if err := s.db.CreateLocalization(ctx, localization); err != nil {
			s.logger.Warn("Failed to create localization",
				zap.String("key", keyStr),
				zap.String("language", lang.Code),
				zap.Error(err),
			)
			continue
		}

		count++
	}

	return count, nil
}

// buildCatalog builds a pre-compiled catalog for a language
func (s *Seeder) buildCatalog(ctx context.Context, lang *models.Language) error {
	// Build catalog using database method
	catalog, err := s.db.BuildCatalog(ctx, lang.ID, "")
	if err != nil {
		return fmt.Errorf("failed to build catalog: %w", err)
	}

	// Catalog is built and stored by the database layer
	s.logger.Info("Catalog built successfully",
		zap.String("language", lang.Code),
		zap.String("catalog_id", catalog.ID),
		zap.Int("version", catalog.Version),
	)

	return nil
}

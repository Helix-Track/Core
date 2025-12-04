package seeder

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/helixtrack/localization-service/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// MockDatabase is a mock implementation of the Database interface
type MockDatabase struct {
	mock.Mock
	languages       []*models.Language
	localizationKeys []*models.LocalizationKey
	localizations   []*models.Localization
	catalogs        []*models.LocalizationCatalog
}

func (m *MockDatabase) GetLanguages(ctx context.Context, activeOnly bool) ([]*models.Language, error) {
	args := m.Called(ctx, activeOnly)
	return args.Get(0).([]*models.Language), args.Error(1)
}

func (m *MockDatabase) CreateLanguage(ctx context.Context, lang *models.Language) error {
	args := m.Called(ctx, lang)
	m.languages = append(m.languages, lang)
	return args.Error(0)
}

func (m *MockDatabase) CreateLocalizationKey(ctx context.Context, key *models.LocalizationKey) error {
	args := m.Called(ctx, key)
	m.localizationKeys = append(m.localizationKeys, key)
	return args.Error(0)
}

func (m *MockDatabase) CreateLocalization(ctx context.Context, loc *models.Localization) error {
	args := m.Called(ctx, loc)
	m.localizations = append(m.localizations, loc)
	return args.Error(0)
}

func (m *MockDatabase) BuildCatalog(ctx context.Context, languageID string, category string) (*models.LocalizationCatalog, error) {
	args := m.Called(ctx, languageID, category)
	catalog := args.Get(0).(*models.LocalizationCatalog)
	m.catalogs = append(m.catalogs, catalog)
	return catalog, args.Error(1)
}

func (m *MockDatabase) Ping() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDatabase) GetLanguageByID(ctx context.Context, id string) (*models.Language, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Language), args.Error(1)
}

func (m *MockDatabase) GetLanguageByCode(ctx context.Context, code string) (*models.Language, error) {
	args := m.Called(ctx, code)
	return args.Get(0).(*models.Language), args.Error(1)
}

func (m *MockDatabase) UpdateLanguage(ctx context.Context, lang *models.Language) error {
	args := m.Called(ctx, lang)
	return args.Error(0)
}

func (m *MockDatabase) DeleteLanguage(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDatabase) GetDefaultLanguage(ctx context.Context) (*models.Language, error) {
	args := m.Called(ctx)
	return args.Get(0).(*models.Language), args.Error(1)
}

func (m *MockDatabase) GetLocalizationKeyByID(ctx context.Context, id string) (*models.LocalizationKey, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.LocalizationKey), args.Error(1)
}

func (m *MockDatabase) GetLocalizationKeyByKey(ctx context.Context, key string) (*models.LocalizationKey, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(*models.LocalizationKey), args.Error(1)
}

func (m *MockDatabase) GetLocalizationKeysByCategory(ctx context.Context, category string) ([]*models.LocalizationKey, error) {
	args := m.Called(ctx, category)
	return args.Get(0).([]*models.LocalizationKey), args.Error(1)
}

func (m *MockDatabase) UpdateLocalizationKey(ctx context.Context, key *models.LocalizationKey) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockDatabase) DeleteLocalizationKey(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDatabase) GetLocalizationByID(ctx context.Context, id string) (*models.Localization, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Localization), args.Error(1)
}

func (m *MockDatabase) GetLocalizationByKeyAndLanguage(ctx context.Context, keyID, languageID string) (*models.Localization, error) {
	args := m.Called(ctx, keyID, languageID)
	return args.Get(0).(*models.Localization), args.Error(1)
}

func (m *MockDatabase) GetLocalizationsByLanguage(ctx context.Context, languageID string) ([]*models.Localization, error) {
	args := m.Called(ctx, languageID)
	return args.Get(0).([]*models.Localization), args.Error(1)
}

func (m *MockDatabase) GetLocalizationsByKeyID(ctx context.Context, keyID string) ([]*models.Localization, error) {
	args := m.Called(ctx, keyID)
	return args.Get(0).([]*models.Localization), args.Error(1)
}

func (m *MockDatabase) UpdateLocalization(ctx context.Context, loc *models.Localization) error {
	args := m.Called(ctx, loc)
	return args.Error(0)
}

func (m *MockDatabase) DeleteLocalization(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDatabase) ApproveLocalization(ctx context.Context, id, username string) error {
	args := m.Called(ctx, id, username)
	return args.Error(0)
}

func (m *MockDatabase) CreateCatalog(ctx context.Context, catalog *models.LocalizationCatalog) error {
	args := m.Called(ctx, catalog)
	m.catalogs = append(m.catalogs, catalog)
	return args.Error(0)
}

func (m *MockDatabase) GetCatalogByLanguage(ctx context.Context, languageID string, category string) (*models.LocalizationCatalog, error) {
	args := m.Called(ctx, languageID, category)
	return args.Get(0).(*models.LocalizationCatalog), args.Error(1)
}

func (m *MockDatabase) GetLatestCatalog(ctx context.Context, languageID string, category string) (*models.LocalizationCatalog, error) {
	args := m.Called(ctx, languageID, category)
	return args.Get(0).(*models.LocalizationCatalog), args.Error(1)
}

func (m *MockDatabase) UpdateCatalog(ctx context.Context, catalog *models.LocalizationCatalog) error {
	args := m.Called(ctx, catalog)
	return args.Error(0)
}

func (m *MockDatabase) DeleteCatalog(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDatabase) CreateVersion(ctx context.Context, version *models.LocalizationVersion) error {
	args := m.Called(ctx, version)
	return args.Error(0)
}

func (m *MockDatabase) GetVersionByNumber(ctx context.Context, versionNumber string) (*models.LocalizationVersion, error) {
	args := m.Called(ctx, versionNumber)
	return args.Get(0).(*models.LocalizationVersion), args.Error(1)
}

func (m *MockDatabase) GetVersionByID(ctx context.Context, id string) (*models.LocalizationVersion, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.LocalizationVersion), args.Error(1)
}

func (m *MockDatabase) GetCurrentVersion(ctx context.Context) (*models.LocalizationVersion, error) {
	args := m.Called(ctx)
	return args.Get(0).(*models.LocalizationVersion), args.Error(1)
}

func (m *MockDatabase) ListVersions(ctx context.Context, limit, offset int) ([]*models.LocalizationVersion, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*models.LocalizationVersion), args.Error(1)
}

func (m *MockDatabase) CountVersions(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *MockDatabase) DeleteVersion(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDatabase) GetCatalogByVersion(ctx context.Context, versionNumber, languageCode string) (*models.LocalizationCatalog, error) {
	args := m.Called(ctx, versionNumber, languageCode)
	return args.Get(0).(*models.LocalizationCatalog), args.Error(1)
}

func (m *MockDatabase) CreateAuditLog(ctx context.Context, action, entityType, entityID, username string, changes interface{}, ipAddress, userAgent string) error {
	args := m.Called(ctx, action, entityType, entityID, username, changes, ipAddress, userAgent)
	return args.Error(0)
}

func (m *MockDatabase) GetStats(ctx context.Context) (map[string]interface{}, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockDatabase) GetCatalog(ctx context.Context, languageID string, version int) (*models.LocalizationCatalog, error) {
	args := m.Called(ctx, languageID, version)
	return args.Get(0).(*models.LocalizationCatalog), args.Error(1)
}

func (m *MockDatabase) GetLocalizationKeys(ctx context.Context, activeOnly bool) ([]*models.LocalizationKey, error) {
	args := m.Called(ctx, activeOnly)
	return args.Get(0).([]*models.LocalizationKey), args.Error(1)
}

func (m *MockDatabase) GetLocalizations(ctx context.Context, languageID string, activeOnly bool) ([]*models.Localization, error) {
	args := m.Called(ctx, languageID, activeOnly)
	return args.Get(0).([]*models.Localization), args.Error(1)
}

func (m *MockDatabase) Close() error {
	args := m.Called()
	return args.Error(0)
}

// createTempSeedData creates temporary seed data files for testing
func createTempSeedData(t *testing.T) string {
	tmpDir := t.TempDir()
	
	// Create languages.json
	languages := []SeedLanguage{
		{
			Code:       "en",
			Name:       "English",
			NativeName: "English",
			IsRTL:      false,
			IsActive:   true,
			IsDefault:  true,
		},
		{
			Code:       "fr",
			Name:       "French",
			NativeName: "Français",
			IsRTL:      false,
			IsActive:   true,
			IsDefault:  false,
		},
	}
	langData, _ := json.MarshalIndent(languages, "", "  ")
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "languages.json"), langData, 0644))
	
	// Create localization-keys.json
	keys := []SeedLocalizationKey{
		{
			Key:         "app.welcome",
			Category:    "app",
			Description: "Welcome message",
			Context:     "Shown when user logs in",
			Variables:   []string{},
		},
		{
			Key:         "app.goodbye",
			Category:    "app",
			Description: "Goodbye message",
			Context:     "Shown when user logs out",
			Variables:   []string{"name"},
		},
	}
	keysData, _ := json.MarshalIndent(keys, "", "  ")
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "localization-keys.json"), keysData, 0644))
	
	// Create localizations directory and files
	locDir := filepath.Join(tmpDir, "localizations")
	require.NoError(t, os.MkdirAll(locDir, 0755))
	
	// English localizations
	enLocs := map[string]string{
		"app.welcome": "Welcome!",
		"app.goodbye": "Goodbye, {{.name}}!",
	}
	enData, _ := json.MarshalIndent(enLocs, "", "  ")
	require.NoError(t, os.WriteFile(filepath.Join(locDir, "en.json"), enData, 0644))
	
	// French localizations
	frLocs := map[string]string{
		"app.welcome": "Bienvenue!",
		"app.goodbye": "Au revoir, {{.name}}!",
	}
	frData, _ := json.MarshalIndent(frLocs, "", "  ")
	require.NoError(t, os.WriteFile(filepath.Join(locDir, "fr.json"), frData, 0644))
	
	return tmpDir
}

func TestNew(t *testing.T) {
	mockDB := &MockDatabase{}
	logger := zap.NewNop()
	seedPath := "/path/to/seed"
	
	seeder := New(mockDB, logger, seedPath)
	
	require.NotNil(t, seeder)
	assert.Equal(t, mockDB, seeder.db)
	assert.Equal(t, logger, seeder.logger)
	assert.Equal(t, seedPath, seeder.seedDataPath)
}

func TestShouldSeed_NeedsSeeding(t *testing.T) {
	mockDB := &MockDatabase{}
	logger := zap.NewNop()
	seeder := New(mockDB, logger, "")
	
	ctx := context.Background()
	mockDB.On("GetLanguages", ctx, true).Return([]*models.Language{}, nil)
	
	shouldSeed, err := seeder.ShouldSeed(ctx)
	
	assert.NoError(t, err)
	assert.True(t, shouldSeed)
	mockDB.AssertExpectations(t)
}

func TestShouldSeed_AlreadySeeded(t *testing.T) {
	mockDB := &MockDatabase{}
	logger := zap.NewNop()
	seeder := New(mockDB, logger, "")
	
	ctx := context.Background()
	existingLang := &models.Language{ID: "1", Code: "en"}
	mockDB.On("GetLanguages", ctx, true).Return([]*models.Language{existingLang}, nil)
	
	shouldSeed, err := seeder.ShouldSeed(ctx)
	
	assert.NoError(t, err)
	assert.False(t, shouldSeed)
	mockDB.AssertExpectations(t)
}

func TestShouldSeed_Error(t *testing.T) {
	mockDB := &MockDatabase{}
	logger := zap.NewNop()
	seeder := New(mockDB, logger, "")
	
	ctx := context.Background()
	mockDB.On("GetLanguages", ctx, true).Return([]*models.Language{}, assert.AnError)
	
	shouldSeed, err := seeder.ShouldSeed(ctx)
	
	assert.Error(t, err)
	assert.False(t, shouldSeed)
	mockDB.AssertExpectations(t)
}

func TestSeed_Success(t *testing.T) {
	tmpDir := createTempSeedData(t)
	
	mockDB := &MockDatabase{}
	logger := zap.NewNop()
	seeder := New(mockDB, logger, tmpDir)
	
	ctx := context.Background()
	
	// Mock database calls - more specific about number of calls
	// Note: Seed doesn't call ShouldSeed, so no GetLanguages call expected
	mockDB.On("CreateLanguage", ctx, mock.AnythingOfType("*models.Language")).Return(nil).Times(2)
	mockDB.On("CreateLocalizationKey", ctx, mock.AnythingOfType("*models.LocalizationKey")).Return(nil).Times(2)
	mockDB.On("CreateLocalization", ctx, mock.AnythingOfType("*models.Localization")).Return(nil).Times(4)
	mockDB.On("BuildCatalog", ctx, mock.AnythingOfType("string"), "").Return(&models.LocalizationCatalog{
		ID:        uuid.New().String(),
		LanguageID: "test",
		Version:   1,
		CreatedAt: time.Now().Unix(),
	}, nil).Times(2)
	
	err := seeder.Seed(ctx)
	
	assert.NoError(t, err)
	mockDB.AssertExpectations(t)
	
	// Verify languages were created
	assert.Len(t, mockDB.languages, 2)
	assert.Equal(t, "en", mockDB.languages[0].Code)
	assert.Equal(t, "fr", mockDB.languages[1].Code)
	
	// Verify keys were created
	assert.Len(t, mockDB.localizationKeys, 2)
	assert.Equal(t, "app.welcome", mockDB.localizationKeys[0].Key)
	assert.Equal(t, "app.goodbye", mockDB.localizationKeys[1].Key)
	
	// Verify localizations were created
	assert.Len(t, mockDB.localizations, 4) // 2 languages × 2 keys
	
	// Verify catalogs were built
	assert.Len(t, mockDB.catalogs, 2)
}

func TestSeed_FileError(t *testing.T) {
	// Use non-existent directory
	seeder := New(&MockDatabase{}, zap.NewNop(), "/nonexistent")
	
	err := seeder.Seed(context.Background())
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read languages file")
}

func TestSeed_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create invalid languages.json
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "languages.json"), []byte("invalid json"), 0644))
	
	seeder := New(&MockDatabase{}, zap.NewNop(), tmpDir)
	
	err := seeder.Seed(context.Background())
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse languages JSON")
}

func TestSeed_DatabaseError_Languages(t *testing.T) {
	tmpDir := createTempSeedData(t)
	
	mockDB := &MockDatabase{}
	logger := zap.NewNop()
	seeder := New(mockDB, logger, tmpDir)
	
	ctx := context.Background()
	
	// Note: Seed doesn't call ShouldSeed
	mockDB.On("CreateLanguage", ctx, mock.AnythingOfType("*models.Language")).Return(assert.AnError)
	
	err := seeder.Seed(ctx)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to seed languages")
	mockDB.AssertExpectations(t)
}

func TestSeed_DatabaseError_Keys(t *testing.T) {
	tmpDir := createTempSeedData(t)
	
	mockDB := &MockDatabase{}
	logger := zap.NewNop()
	seeder := New(mockDB, logger, tmpDir)
	
	ctx := context.Background()
	
	// Note: Seed doesn't call ShouldSeed
	mockDB.On("CreateLanguage", ctx, mock.AnythingOfType("*models.Language")).Return(nil).Times(2)
	mockDB.On("CreateLocalizationKey", ctx, mock.AnythingOfType("*models.LocalizationKey")).Return(assert.AnError)
	
	err := seeder.Seed(ctx)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to seed localization keys")
	mockDB.AssertExpectations(t)
}

func TestSeed_MissingLocalizationFile(t *testing.T) {
	tmpDir := createTempSeedData(t)
	
	// Remove French localizations file
	require.NoError(t, os.Remove(filepath.Join(tmpDir, "localizations", "fr.json")))
	
	mockDB := &MockDatabase{}
	logger := zap.NewNop()
	seeder := New(mockDB, logger, tmpDir)
	
	ctx := context.Background()
	
	// Note: Seed doesn't call ShouldSeed, so no GetLanguages call expected
	mockDB.On("CreateLanguage", ctx, mock.AnythingOfType("*models.Language")).Return(nil).Times(2)
	mockDB.On("CreateLocalizationKey", ctx, mock.AnythingOfType("*models.LocalizationKey")).Return(nil).Times(2)
	mockDB.On("CreateLocalization", ctx, mock.AnythingOfType("*models.Localization")).Return(nil).Times(2) // Only English
	mockDB.On("BuildCatalog", ctx, mock.AnythingOfType("string"), "").Return(&models.LocalizationCatalog{
		ID:        uuid.New().String(),
		LanguageID: "test",
		Version:   1,
		CreatedAt: time.Now().Unix(),
	}, nil).Times(2)
	
	err := seeder.Seed(ctx)
	
	assert.NoError(t, err)
	mockDB.AssertExpectations(t)
	
	// Should only have English localizations (2)
	assert.Len(t, mockDB.localizations, 2)
}

func TestSeed_InvalidLocalizationJSON(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create languages.json
	languages := []SeedLanguage{
		{Code: "en", Name: "English", NativeName: "English", IsRTL: false, IsActive: true, IsDefault: true},
	}
	langData, _ := json.MarshalIndent(languages, "", "  ")
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "languages.json"), langData, 0644))
	
	// Create localization-keys.json
	keys := []SeedLocalizationKey{
		{Key: "app.welcome", Category: "app", Description: "Welcome", Context: "", Variables: []string{}},
	}
	keysData, _ := json.MarshalIndent(keys, "", "  ")
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "localization-keys.json"), keysData, 0644))
	
	// Create localizations directory and invalid file
	locDir := filepath.Join(tmpDir, "localizations")
	require.NoError(t, os.MkdirAll(locDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(locDir, "en.json"), []byte("invalid json"), 0644))
	
	mockDB := &MockDatabase{}
	logger := zap.NewNop()
	seeder := New(mockDB, logger, tmpDir)
	
	ctx := context.Background()
	
	// Mock all calls seeder will make
	mockDB.On("CreateLanguage", ctx, mock.AnythingOfType("*models.Language")).Return(nil).Times(1)
	mockDB.On("CreateLocalizationKey", ctx, mock.AnythingOfType("*models.LocalizationKey")).Return(nil).Times(1)
	mockDB.On("BuildCatalog", ctx, mock.AnythingOfType("string"), "").Return(&models.LocalizationCatalog{
		ID:        uuid.New().String(),
		LanguageID: "test",
		Version:   1,
		CreatedAt: time.Now().Unix(),
	}, nil)
	
	err := seeder.Seed(ctx)
	
	// Seed doesn't fail, it just logs warnings
	assert.NoError(t, err)
	mockDB.AssertExpectations(t)
	
	// No localizations should be created due to invalid JSON
	assert.Len(t, mockDB.localizations, 0)
}

func TestSeed_CatalogBuildError(t *testing.T) {
	tmpDir := createTempSeedData(t)
	
	mockDB := &MockDatabase{}
	logger := zap.NewNop()
	seeder := New(mockDB, logger, tmpDir)
	
	ctx := context.Background()
	
	// Note: Seed doesn't call ShouldSeed
	mockDB.On("CreateLanguage", ctx, mock.AnythingOfType("*models.Language")).Return(nil).Times(2)
	mockDB.On("CreateLocalizationKey", ctx, mock.AnythingOfType("*models.LocalizationKey")).Return(nil).Times(2)
	mockDB.On("CreateLocalization", ctx, mock.AnythingOfType("*models.Localization")).Return(nil).Times(4)
	
	// For BuildCatalog, return a nil catalog and error
	catalog := (*models.LocalizationCatalog)(nil)
	mockDB.On("BuildCatalog", ctx, mock.AnythingOfType("string"), "").Return(catalog, assert.AnError).Times(2)
	
	err := seeder.Seed(ctx)
	
	// Should not fail, just logs warnings
	assert.NoError(t, err)
	mockDB.AssertExpectations(t)
	
	// Verify languages and keys were still created
	assert.Len(t, mockDB.languages, 2)
	assert.Len(t, mockDB.localizationKeys, 2)
	assert.Len(t, mockDB.localizations, 4)
}
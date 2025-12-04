package database

import (
	"context"
	"encoding/json"
	"errors"
	"sync"

	"github.com/helixtrack/localization-service/internal/models"
)

// MockDatabase implements Database interface for testing
type MockDatabase struct {
	mu                sync.RWMutex
	languages         map[string]*models.Language
	localizationKeys  map[string]*models.LocalizationKey
	localizations     map[string]*models.Localization
	catalogs          map[string]*models.LocalizationCatalog
	shouldReturnError bool
	errorToReturn     error
}

// NewMockDatabase creates a new mock database for testing
func NewMockDatabase() *MockDatabase {
	return &MockDatabase{
		languages:        make(map[string]*models.Language),
		localizationKeys: make(map[string]*models.LocalizationKey),
		localizations:    make(map[string]*models.Localization),
		catalogs:         make(map[string]*models.LocalizationCatalog),
	}
}

// SetError configures the mock to return an error
func (m *MockDatabase) SetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldReturnError = true
	m.errorToReturn = err
}

// ClearError clears the error state
func (m *MockDatabase) ClearError() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldReturnError = false
	m.errorToReturn = nil
}

// Ping implementation
func (m *MockDatabase) Ping() error {
	if m.shouldReturnError {
		return m.errorToReturn
	}
	return nil
}

// Close implementation
func (m *MockDatabase) Close() error {
	if m.shouldReturnError {
		return m.errorToReturn
	}
	return nil
}

// Language operations
func (m *MockDatabase) CreateLanguage(ctx context.Context, lang *models.Language) error {
	if m.shouldReturnError {
		return m.errorToReturn
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	lang.BeforeCreate()
	if err := lang.Validate(); err != nil {
		return err
	}
	
	m.languages[lang.ID] = lang
	return nil
}

func (m *MockDatabase) GetLanguageByID(ctx context.Context, id string) (*models.Language, error) {
	if m.shouldReturnError {
		return nil, m.errorToReturn
	}
	
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	lang, exists := m.languages[id]
	if !exists {
		return nil, models.ErrResourceNotFound("language")
	}
	return lang, nil
}

func (m *MockDatabase) GetLanguageByCode(ctx context.Context, code string) (*models.Language, error) {
	if m.shouldReturnError {
		return nil, m.errorToReturn
	}
	
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	for _, lang := range m.languages {
		if lang.Code == code {
			return lang, nil
		}
	}
	return nil, models.ErrResourceNotFound("language")
}

func (m *MockDatabase) GetLanguages(ctx context.Context, activeOnly bool) ([]*models.Language, error) {
	if m.shouldReturnError {
		return nil, m.errorToReturn
	}
	
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var languages []*models.Language
	for _, lang := range m.languages {
		if !activeOnly || lang.IsActive {
			languages = append(languages, lang)
		}
	}
	return languages, nil
}

func (m *MockDatabase) UpdateLanguage(ctx context.Context, lang *models.Language) error {
	if m.shouldReturnError {
		return m.errorToReturn
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	lang.BeforeUpdate()
	if err := lang.Validate(); err != nil {
		return err
	}
	
	if _, exists := m.languages[lang.ID]; !exists {
		return models.ErrResourceNotFound("language")
	}
	
	m.languages[lang.ID] = lang
	return nil
}

func (m *MockDatabase) DeleteLanguage(ctx context.Context, id string) error {
	if m.shouldReturnError {
		return m.errorToReturn
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, exists := m.languages[id]; !exists {
		return models.ErrResourceNotFound("language")
	}
	
	delete(m.languages, id)
	return nil
}

func (m *MockDatabase) GetDefaultLanguage(ctx context.Context) (*models.Language, error) {
	if m.shouldReturnError {
		return nil, m.errorToReturn
	}
	
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	for _, lang := range m.languages {
		if lang.IsDefault {
			return lang, nil
		}
	}
	return nil, models.ErrResourceNotFound("default language")
}

// Localization Key operations
func (m *MockDatabase) CreateLocalizationKey(ctx context.Context, key *models.LocalizationKey) error {
	if m.shouldReturnError {
		return m.errorToReturn
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	key.BeforeCreate()
	if err := key.Validate(); err != nil {
		return err
	}
	
	m.localizationKeys[key.ID] = key
	return nil
}

func (m *MockDatabase) GetLocalizationKeyByID(ctx context.Context, id string) (*models.LocalizationKey, error) {
	if m.shouldReturnError {
		return nil, m.errorToReturn
	}
	
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	key, exists := m.localizationKeys[id]
	if !exists {
		return nil, models.ErrResourceNotFound("localization key")
	}
	return key, nil
}

func (m *MockDatabase) GetLocalizationKeyByKey(ctx context.Context, keyStr string) (*models.LocalizationKey, error) {
	if m.shouldReturnError {
		return nil, m.errorToReturn
	}
	
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	for _, key := range m.localizationKeys {
		if key.Key == keyStr {
			return key, nil
		}
	}
	return nil, models.ErrResourceNotFound("localization key")
}

func (m *MockDatabase) GetLocalizationKeysByCategory(ctx context.Context, category string) ([]*models.LocalizationKey, error) {
	if m.shouldReturnError {
		return nil, m.errorToReturn
	}
	
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var keys []*models.LocalizationKey
	for _, key := range m.localizationKeys {
		if key.Category == category {
			keys = append(keys, key)
		}
	}
	return keys, nil
}

func (m *MockDatabase) UpdateLocalizationKey(ctx context.Context, key *models.LocalizationKey) error {
	if m.shouldReturnError {
		return m.errorToReturn
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	key.BeforeUpdate()
	if err := key.Validate(); err != nil {
		return err
	}
	
	if _, exists := m.localizationKeys[key.ID]; !exists {
		return models.ErrResourceNotFound("localization key")
	}
	
	m.localizationKeys[key.ID] = key
	return nil
}

func (m *MockDatabase) DeleteLocalizationKey(ctx context.Context, id string) error {
	if m.shouldReturnError {
		return m.errorToReturn
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, exists := m.localizationKeys[id]; !exists {
		return models.ErrResourceNotFound("localization key")
	}
	
	delete(m.localizationKeys, id)
	return nil
}

// Localization operations
func (m *MockDatabase) CreateLocalization(ctx context.Context, loc *models.Localization) error {
	if m.shouldReturnError {
		return m.errorToReturn
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	loc.BeforeCreate()
	if err := loc.Validate(); err != nil {
		return err
	}
	
	m.localizations[loc.ID] = loc
	return nil
}

func (m *MockDatabase) GetLocalizationByID(ctx context.Context, id string) (*models.Localization, error) {
	if m.shouldReturnError {
		return nil, m.errorToReturn
	}
	
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	loc, exists := m.localizations[id]
	if !exists {
		return nil, models.ErrResourceNotFound("localization")
	}
	return loc, nil
}

func (m *MockDatabase) GetLocalizationByKeyAndLanguage(ctx context.Context, keyID, languageID string) (*models.Localization, error) {
	if m.shouldReturnError {
		return nil, m.errorToReturn
	}
	
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Find the key first
	var key *models.LocalizationKey
	for _, k := range m.localizationKeys {
		if k.ID == keyID {
			key = k
			break
		}
	}
	if key == nil {
		return nil, models.ErrResourceNotFound("localization key")
	}
	
	// Find localization
	for _, loc := range m.localizations {
		if loc.KeyID == keyID && loc.LanguageID == languageID {
			return loc, nil
		}
	}
	return nil, models.ErrResourceNotFound("localization")
}

func (m *MockDatabase) GetLocalizationsByLanguage(ctx context.Context, languageID string) ([]*models.Localization, error) {
	if m.shouldReturnError {
		return nil, m.errorToReturn
	}
	
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var localizations []*models.Localization
	for _, loc := range m.localizations {
		if loc.LanguageID == languageID {
			localizations = append(localizations, loc)
		}
	}
	return localizations, nil
}

func (m *MockDatabase) GetLocalizationsByKeyID(ctx context.Context, keyID string) ([]*models.Localization, error) {
	if m.shouldReturnError {
		return nil, m.errorToReturn
	}
	
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var localizations []*models.Localization
	for _, loc := range m.localizations {
		if loc.KeyID == keyID {
			localizations = append(localizations, loc)
		}
	}
	return localizations, nil
}

func (m *MockDatabase) UpdateLocalization(ctx context.Context, loc *models.Localization) error {
	if m.shouldReturnError {
		return m.errorToReturn
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	loc.BeforeUpdate()
	if err := loc.Validate(); err != nil {
		return err
	}
	
	if _, exists := m.localizations[loc.ID]; !exists {
		return models.ErrResourceNotFound("localization")
	}
	
	m.localizations[loc.ID] = loc
	return nil
}

func (m *MockDatabase) DeleteLocalization(ctx context.Context, id string) error {
	if m.shouldReturnError {
		return m.errorToReturn
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, exists := m.localizations[id]; !exists {
		return models.ErrResourceNotFound("localization")
	}
	
	delete(m.localizations, id)
	return nil
}

func (m *MockDatabase) ApproveLocalization(ctx context.Context, id, username string) error {
	if m.shouldReturnError {
		return m.errorToReturn
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	loc, exists := m.localizations[id]
	if !exists {
		return models.ErrResourceNotFound("localization")
	}
	
	loc.Approve(username)
	m.localizations[id] = loc
	return nil
}

// Catalog operations
func (m *MockDatabase) CreateCatalog(ctx context.Context, catalog *models.LocalizationCatalog) error {
	if m.shouldReturnError {
		return m.errorToReturn
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	catalog.BeforeCreate()
	if err := catalog.Validate(); err != nil {
		return err
	}
	
	m.catalogs[catalog.ID] = catalog
	return nil
}

func (m *MockDatabase) GetCatalogByLanguage(ctx context.Context, languageID string, category string) (*models.LocalizationCatalog, error) {
	if m.shouldReturnError {
		return nil, m.errorToReturn
	}
	
	m.mu.RLock()
	for _, catalog := range m.catalogs {
		if catalog.LanguageID == languageID && catalog.Category == category {
			// Found catalog, return it
			m.mu.RUnlock()
			return catalog, nil
		}
	}
	m.mu.RUnlock()
	
	// If not found, build a new one
	return m.BuildCatalog(ctx, languageID, category)
}

func (m *MockDatabase) GetLatestCatalog(ctx context.Context, languageID string, category string) (*models.LocalizationCatalog, error) {
	return m.GetCatalogByLanguage(ctx, languageID, category)
}

func (m *MockDatabase) UpdateCatalog(ctx context.Context, catalog *models.LocalizationCatalog) error {
	if m.shouldReturnError {
		return m.errorToReturn
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	catalog.BeforeUpdate()
	if err := catalog.Validate(); err != nil {
		return err
	}
	
	if _, exists := m.catalogs[catalog.ID]; !exists {
		return models.ErrResourceNotFound("catalog")
	}
	
	m.catalogs[catalog.ID] = catalog
	return nil
}

func (m *MockDatabase) DeleteCatalog(ctx context.Context, id string) error {
	if m.shouldReturnError {
		return m.errorToReturn
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, exists := m.catalogs[id]; !exists {
		return models.ErrResourceNotFound("catalog")
	}
	
	delete(m.catalogs, id)
	return nil
}

func (m *MockDatabase) BuildCatalog(ctx context.Context, languageID string, category string) (*models.LocalizationCatalog, error) {
	if m.shouldReturnError {
		return nil, m.errorToReturn
	}
	
	m.mu.RLock()
	
	// Check if language exists
	_, exists := m.languages[languageID]
	if !exists {
		m.mu.RUnlock()
		return nil, models.ErrResourceNotFound("language")
	}
	
	// Build catalog map
	catalogMap := make(map[string]string)
	for _, loc := range m.localizations {
		if loc.LanguageID == languageID {
			// Find the key
			if key, exists := m.localizationKeys[loc.KeyID]; exists {
				if category == "" || key.Category == category {
					catalogMap[key.Key] = loc.Value
				}
			}
		}
	}
	m.mu.RUnlock()
	
	// Marshal to JSON
	catalogJSON, err := json.Marshal(catalogMap)
	if err != nil {
		return nil, err
	}
	
	// Create new catalog with proper data
	catalog := &models.LocalizationCatalog{
		LanguageID:  languageID,
		Category:    category,
		CatalogData: catalogJSON,
		Version:     1,
	}
	catalog.BeforeCreate()
	
	// Save catalog
	if err := m.CreateCatalog(ctx, catalog); err != nil {
		return nil, err
	}
	
	return catalog, nil
}

// Version operations (not implemented in mock)
func (m *MockDatabase) CreateLocalizationVersion(ctx context.Context, version *models.LocalizationVersion) error {
	return errors.New("not implemented in mock")
}

func (m *MockDatabase) GetLocalizationVersions(ctx context.Context, languageID, keyID string) ([]*models.LocalizationVersion, error) {
	return nil, errors.New("not implemented in mock")
}

func (m *MockDatabase) GetLocalizationVersion(ctx context.Context, id string) (*models.LocalizationVersion, error) {
	return nil, errors.New("not implemented in mock")
}

func (m *MockDatabase) DeleteLocalizationVersion(ctx context.Context, id string) error {
	return errors.New("not implemented in mock")
}

// Health check implementation
func (m *MockDatabase) Health() error {
	if m.shouldReturnError {
		return m.errorToReturn
	}
	return nil
}
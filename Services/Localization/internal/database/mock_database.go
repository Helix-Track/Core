package database

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/helixtrack/localization-service/internal/models"
)

// MockDatabase implements Database interface for testing
type MockDatabase struct {
	mu                     sync.RWMutex
	languages              map[string]*models.Language
	localizationKeys       map[string]*models.LocalizationKey
	localizations          map[string]*models.Localization
	catalogs               map[string]*models.LocalizationCatalog
	versions               map[string]*models.LocalizationVersion
	shouldReturnError      bool
	errorToReturn          error
	// Separate error flag for CreateCatalog to test BuildCatalog error path
	shouldFailCreateCatalog bool
	// Flag to simulate json.Marshal error in BuildCatalog
	shouldFailMarshalCatalog bool
}

// NewMockDatabase creates a new mock database for testing
func NewMockDatabase() *MockDatabase {
	return &MockDatabase{
		languages:        make(map[string]*models.Language),
		localizationKeys: make(map[string]*models.LocalizationKey),
		localizations:    make(map[string]*models.Localization),
		catalogs:         make(map[string]*models.LocalizationCatalog),
		versions:         make(map[string]*models.LocalizationVersion),
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

// SetCreateCatalogError configures the mock to return an error specifically for CreateCatalog
func (m *MockDatabase) SetCreateCatalogError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFailCreateCatalog = true
}

// ClearCreateCatalogError clears the CreateCatalog error condition
func (m *MockDatabase) ClearCreateCatalogError() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFailCreateCatalog = false
}

// SetMarshalCatalogError configures the mock to simulate a json.Marshal error in BuildCatalog
func (m *MockDatabase) SetMarshalCatalogError() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFailMarshalCatalog = true
}

// ClearMarshalCatalogError clears the marshal catalog error condition
func (m *MockDatabase) ClearMarshalCatalogError() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFailMarshalCatalog = false
}

// SetVersionsNil sets the versions map to nil (for testing purposes)
func (m *MockDatabase) SetVersionsNil() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.versions = nil
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
	if m.shouldFailCreateCatalog {
		return fmt.Errorf("mock create catalog error")
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
	
	// Check if we should simulate a marshal error
	if m.shouldFailMarshalCatalog {
		return nil, fmt.Errorf("simulated json.Marshal error")
	}
	
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

// Version operations
func (m *MockDatabase) CreateVersion(ctx context.Context, version *models.LocalizationVersion) error {
	if m.shouldReturnError {
		return m.errorToReturn
	}
	
	if version == nil {
		return fmt.Errorf("version cannot be nil")
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.versions == nil {
		m.versions = make(map[string]*models.LocalizationVersion)
	}
	
	// Set ID and timestamps
	version.BeforeCreate()
	
	m.versions[version.ID] = version
	return nil
}

func (m *MockDatabase) GetVersionByNumber(ctx context.Context, versionNumber string) (*models.LocalizationVersion, error) {
	if m.shouldReturnError {
		return nil, m.errorToReturn
	}
	
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	for _, version := range m.versions {
		if version.VersionNumber == versionNumber {
			return version, nil
		}
	}
	
	return nil, models.ErrResourceNotFound("version")
}

func (m *MockDatabase) GetVersionByID(ctx context.Context, id string) (*models.LocalizationVersion, error) {
	if m.shouldReturnError {
		return nil, m.errorToReturn
	}
	
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	version, exists := m.versions[id]
	if !exists {
		return nil, models.ErrResourceNotFound("version")
	}
	
	return version, nil
}

func (m *MockDatabase) GetCurrentVersion(ctx context.Context) (*models.LocalizationVersion, error) {
	if m.shouldReturnError {
		return nil, m.errorToReturn
	}
	
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if len(m.versions) == 0 {
		return nil, models.ErrResourceNotFound("version")
	}
	
	// Return the most recently created version
	var latestVersion *models.LocalizationVersion
	for _, version := range m.versions {
		if latestVersion == nil || version.CreatedAt > latestVersion.CreatedAt {
			latestVersion = version
		}
	}
	
	return latestVersion, nil
}

func (m *MockDatabase) ListVersions(ctx context.Context, limit, offset int) ([]*models.LocalizationVersion, error) {
	if m.shouldReturnError {
		return nil, m.errorToReturn
	}
	
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var versions []*models.LocalizationVersion
	for _, version := range m.versions {
		versions = append(versions, version)
	}
	
	// Sort versions by CreatedAt to ensure consistent ordering
	for i := 0; i < len(versions); i++ {
		for j := i + 1; j < len(versions); j++ {
			if versions[i].CreatedAt > versions[j].CreatedAt {
				versions[i], versions[j] = versions[j], versions[i]
			}
		}
	}
	
	// Apply pagination
	if offset >= len(versions) {
		return []*models.LocalizationVersion{}, nil
	}
	
	end := offset + limit
	if end > len(versions) {
		end = len(versions)
	}
	
	return versions[offset:end], nil
}

func (m *MockDatabase) CountVersions(ctx context.Context) (int, error) {
	if m.shouldReturnError {
		return 0, m.errorToReturn
	}
	
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	return len(m.versions), nil
}

func (m *MockDatabase) DeleteVersion(ctx context.Context, id string) error {
	if m.shouldReturnError {
		return m.errorToReturn
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, exists := m.versions[id]; !exists {
		return models.ErrResourceNotFound("version")
	}
	
	delete(m.versions, id)
	return nil
}

func (m *MockDatabase) GetCatalogByVersion(ctx context.Context, versionNumber, languageCode string) (*models.LocalizationCatalog, error) {
	if m.shouldReturnError {
		return nil, m.errorToReturn
	}
	
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Find version
	var targetVersion *models.LocalizationVersion
	for _, version := range m.versions {
		if version.VersionNumber == versionNumber {
			targetVersion = version
			break
		}
	}
	
	if targetVersion == nil {
		return nil, models.ErrResourceNotFound("version")
	}
	
	// Find language
	var targetLang *models.Language
	for _, lang := range m.languages {
		if lang.Code == languageCode {
			targetLang = lang
			break
		}
	}
	
	if targetLang == nil {
		return nil, models.ErrResourceNotFound("language")
	}
	
	// Return a mock catalog
	catalogData, _ := json.Marshal(map[string]interface{}{
		"version": targetVersion.VersionNumber,
		"language": languageCode,
	})
	
	catalog := &models.LocalizationCatalog{
		ID:         "catalog-" + targetVersion.ID + "-" + targetLang.ID,
		LanguageID: targetLang.ID,
		Version:    int(targetVersion.TotalLocalizations),
		CatalogData: catalogData,
	}
	
	return catalog, nil
}

// Audit operations
func (m *MockDatabase) CreateAuditLog(ctx context.Context, action, entityType, entityID, username string, changes interface{}, ipAddress, userAgent string) error {
	if m.shouldReturnError {
		return m.errorToReturn
	}
	
	// For mock implementation, just return success since we don't need to store audit logs
	return nil
}

// GetStats retrieves database statistics
func (m *MockDatabase) GetStats(ctx context.Context) (map[string]interface{}, error) {
	if m.shouldReturnError {
		return nil, m.errorToReturn
	}
	
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	stats := make(map[string]interface{})
	stats["total_languages"] = len(m.languages)
	stats["total_keys"] = len(m.localizationKeys)
	stats["total_localizations"] = len(m.localizations)
	stats["total_catalogs"] = len(m.catalogs)
	stats["total_versions"] = len(m.versions)
	
	return stats, nil
}

// Health check implementation
func (m *MockDatabase) Health() error {
	if m.shouldReturnError {
		return m.errorToReturn
	}
	return nil
}
package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/helixtrack/localization-service/internal/cache"
	"github.com/helixtrack/localization-service/internal/database"
	"github.com/helixtrack/localization-service/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// MockDatabase implements the Database interface for testing
type MockDatabase struct {
	languages       map[string]*models.Language
	localizationKeys map[string]*models.LocalizationKey
	localizations   map[string]*models.Localization
	catalogs        map[string]*models.LocalizationCatalog
}

func NewMockDatabase() *MockDatabase {
	db := &MockDatabase{
		languages:        make(map[string]*models.Language),
		localizationKeys: make(map[string]*models.LocalizationKey),
		localizations:    make(map[string]*models.Localization),
		catalogs:         make(map[string]*models.LocalizationCatalog),
	}

	// Seed with test data
	db.seedTestData()
	return db
}

func (m *MockDatabase) seedTestData() {
	// Add English language (default)
	enLang := &models.Language{
		ID:         "lang-en",
		Code:       "en",
		Name:       "English",
		IsActive:   true,
		IsDefault:  true,
		CreatedAt:  time.Now().Unix(),
		ModifiedAt: time.Now().Unix(),
	}
	m.languages["en"] = enLang

	// Add German language
	deLang := &models.Language{
		ID:         "lang-de",
		Code:       "de",
		Name:       "German",
		IsActive:   true,
		IsDefault:  false,
		CreatedAt:  time.Now().Unix(),
		ModifiedAt: time.Now().Unix(),
	}
	m.languages["de"] = deLang

	// Add localization keys
	key1 := &models.LocalizationKey{
		ID:        "key-1",
		Key:       "app.welcome",
		Category:  "general",
		CreatedAt: time.Now().Unix(),
	}
	m.localizationKeys["app.welcome"] = key1

	key2 := &models.LocalizationKey{
		ID:        "key-2",
		Key:       "app.error",
		Category:  "errors",
		CreatedAt: time.Now().Unix(),
	}
	m.localizationKeys["app.error"] = key2

	// Add localizations
	loc1 := &models.Localization{
		ID:          "loc-1",
		KeyID:       "key-1",
		LanguageID:  "lang-en",
		Value:       "Welcome!",
		Approved:    true,
		ApprovedAt:  time.Now().Unix(),
		CreatedAt:   time.Now().Unix(),
		ModifiedAt:  time.Now().Unix(),
	}
	m.localizations["key-1:lang-en"] = loc1

	loc2 := &models.Localization{
		ID:          "loc-2",
		KeyID:       "key-2",
		LanguageID:  "lang-en",
		Value:       "An error occurred",
		Approved:    true,
		ApprovedAt:  time.Now().Unix(),
		CreatedAt:   time.Now().Unix(),
		ModifiedAt:  time.Now().Unix(),
	}
	m.localizations["key-2:lang-en"] = loc2

	// Add catalog
	catalogData, _ := json.Marshal(map[string]string{
		"app.welcome": "Welcome!",
		"app.error":   "An error occurred",
	})
	catalog := &models.LocalizationCatalog{
		ID:          "catalog-1",
		LanguageID:  "lang-en",
		Category:    "",
		Version:     1,
		CatalogData: catalogData,
		Checksum:    "test-checksum",
		CreatedAt:   time.Now().Unix(),
	}
	m.catalogs["lang-en:"] = catalog
}

// Implement Database interface methods
func (m *MockDatabase) Ping() error {
	return nil
}

func (m *MockDatabase) Close() error {
	return nil
}

func (m *MockDatabase) GetLanguageByCode(ctx context.Context, code string) (*models.Language, error) {
	if lang, ok := m.languages[code]; ok {
		return lang, nil
	}
	return nil, models.ErrNotFound
}

func (m *MockDatabase) GetDefaultLanguage(ctx context.Context) (*models.Language, error) {
	for _, lang := range m.languages {
		if lang.IsDefault {
			return lang, nil
		}
	}
	return nil, models.ErrNotFound
}

func (m *MockDatabase) GetLanguages(ctx context.Context, activeOnly bool) ([]*models.Language, error) {
	var languages []*models.Language
	for _, lang := range m.languages {
		if !activeOnly || lang.IsActive {
			languages = append(languages, lang)
		}
	}
	return languages, nil
}

func (m *MockDatabase) GetLocalizationKeyByKey(ctx context.Context, key string) (*models.LocalizationKey, error) {
	if locKey, ok := m.localizationKeys[key]; ok {
		return locKey, nil
	}
	return nil, models.ErrNotFound
}

func (m *MockDatabase) GetLocalizationByKeyAndLanguage(ctx context.Context, keyID, languageID string) (*models.Localization, error) {
	if loc, ok := m.localizations[keyID+":"+languageID]; ok {
		return loc, nil
	}
	return nil, models.ErrNotFound
}

func (m *MockDatabase) GetLatestCatalog(ctx context.Context, languageID, category string) (*models.LocalizationCatalog, error) {
	if catalog, ok := m.catalogs[languageID+":"+category]; ok {
		return catalog, nil
	}
	return nil, models.ErrNotFound
}

// Stub implementations for unused methods
func (m *MockDatabase) CreateLanguage(ctx context.Context, lang *models.Language) error {
	m.languages[lang.Code] = lang
	return nil
}

func (m *MockDatabase) UpdateLanguage(ctx context.Context, lang *models.Language) error {
	m.languages[lang.Code] = lang
	return nil
}

func (m *MockDatabase) DeleteLanguage(ctx context.Context, id string) error {
	for code, lang := range m.languages {
		if lang.ID == id {
			delete(m.languages, code)
			return nil
		}
	}
	return models.ErrNotFound
}

func (m *MockDatabase) CreateLocalizationKey(ctx context.Context, key *models.LocalizationKey) error {
	return nil
}

func (m *MockDatabase) CreateLocalization(ctx context.Context, loc *models.Localization) error {
	return nil
}

func (m *MockDatabase) UpdateLocalization(ctx context.Context, loc *models.Localization) error {
	return nil
}

func (m *MockDatabase) DeleteLocalization(ctx context.Context, id string) error {
	return nil
}

func (m *MockDatabase) ApproveLocalization(ctx context.Context, id, username string) error {
	return nil
}

func (m *MockDatabase) CreateCatalog(ctx context.Context, catalog *models.LocalizationCatalog) error {
	return nil
}

func (m *MockDatabase) GetCatalogByLanguage(ctx context.Context, languageID string, category string) (*models.LocalizationCatalog, error) {
	return nil, models.ErrNotFound
}

func (m *MockDatabase) UpdateCatalog(ctx context.Context, catalog *models.LocalizationCatalog) error {
	return nil
}

func (m *MockDatabase) DeleteCatalog(ctx context.Context, id string) error {
	return nil
}

func (m *MockDatabase) BuildCatalog(ctx context.Context, languageID string, category string) (*models.LocalizationCatalog, error) {
	// Return a simple built catalog
	catalogData, _ := json.Marshal(map[string]string{
		"app.welcome": "Welcome!",
		"app.error":   "An error occurred",
	})
	catalog := &models.LocalizationCatalog{
		ID:          "catalog-built",
		LanguageID:  languageID,
		Category:    category,
		Version:     1,
		CatalogData: catalogData,
		Checksum:    "test-checksum",
		CreatedAt:   time.Now().Unix(),
	}
	return catalog, nil
}

// Localization Key stubs
func (m *MockDatabase) GetLocalizationKeyByID(ctx context.Context, id string) (*models.LocalizationKey, error) {
	return nil, models.ErrNotFound
}

func (m *MockDatabase) GetLocalizationKeysByCategory(ctx context.Context, category string) ([]*models.LocalizationKey, error) {
	return nil, nil
}

func (m *MockDatabase) UpdateLocalizationKey(ctx context.Context, key *models.LocalizationKey) error {
	return nil
}

func (m *MockDatabase) DeleteLocalizationKey(ctx context.Context, id string) error {
	return nil
}

// Language stubs
func (m *MockDatabase) GetLanguageByID(ctx context.Context, id string) (*models.Language, error) {
	return nil, models.ErrNotFound
}

// Localization stubs
func (m *MockDatabase) GetLocalizationByID(ctx context.Context, id string) (*models.Localization, error) {
	return nil, models.ErrNotFound
}

func (m *MockDatabase) GetLocalizationsByLanguage(ctx context.Context, languageID string) ([]*models.Localization, error) {
	return nil, nil
}

func (m *MockDatabase) GetLocalizationsByKeyID(ctx context.Context, keyID string) ([]*models.Localization, error) {
	return nil, nil
}

// Audit operations
func (m *MockDatabase) CreateAuditLog(ctx context.Context, action, entityType, entityID, username string, changes interface{}, ipAddress, userAgent string) error {
	return nil
}

func (m *MockDatabase) GetStats(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{
		"total_languages":     len(m.languages),
		"total_keys":          len(m.localizationKeys),
		"total_localizations": len(m.localizations),
	}, nil
}

// CountVersions returns the total number of versions (mock implementation)
func (m *MockDatabase) CountVersions(ctx context.Context) (int, error) {
	// Mock implementation - return a fixed count for testing
	return 5, nil
}

// CreateVersion creates a new version (mock implementation)
func (m *MockDatabase) CreateVersion(ctx context.Context, version *models.LocalizationVersion) error {
	// Mock implementation - just return success for testing
	return nil
}

// DeleteVersion deletes a version (mock implementation)
func (m *MockDatabase) DeleteVersion(ctx context.Context, id string) error {
	// Mock implementation - just return success for testing
	return nil
}

// GetCatalogByVersion gets a catalog by version number and language code (mock implementation)
func (m *MockDatabase) GetCatalogByVersion(ctx context.Context, versionNumber, languageCode string) (*models.LocalizationCatalog, error) {
	// Mock implementation - return a sample catalog
	return &models.LocalizationCatalog{
		ID:         "test-catalog-1",
		LanguageID: "en",
		Category:   "general",
	}, nil
}

var _ database.Database = (*MockDatabase)(nil)

// Helper function to create a test JWT token
func createTestJWT(username, role string, secret string) string {
	claims := &models.JWTClaims{
		Username: username,
		Role:     role,
	}
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(1 * time.Hour))
	claims.IssuedAt = jwt.NewNumericDate(time.Now())

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

// Test setup helper
func setupTestServer() (*gin.Engine, *Handler, string) {
	logger, _ := zap.NewDevelopment()
	db := NewMockDatabase()
	memCache := cache.NewMemoryCache(10, 1*time.Hour, 5*time.Minute, logger)

	// WebSocket manager not needed for integration tests
	handler := NewHandler(db, memCache, logger, nil)

	router := gin.New()
	jwtSecret := "test-secret"
	adminRoles := []string{"admin"}

	handler.RegisterRoutes(router, jwtSecret, adminRoles)

	return router, handler, jwtSecret
}

// Integration Tests

func TestIntegration_HealthCheck(t *testing.T) {
	router, _, _ := setupTestServer()

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "healthy", response.Status)
	assert.NotEmpty(t, response.Version)
	assert.NotNil(t, response.Checks)
}

func TestIntegration_GetCatalog_Success(t *testing.T) {
	router, _, jwtSecret := setupTestServer()

	token := createTestJWT("testuser", "user", jwtSecret)
	req, _ := http.NewRequest("GET", "/v1/catalog/en", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.NotNil(t, response.Data)
}

func TestIntegration_GetCatalog_Unauthorized(t *testing.T) {
	router, _, _ := setupTestServer()

	req, _ := http.NewRequest("GET", "/v1/catalog/en", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestIntegration_GetCatalog_LanguageNotFound(t *testing.T) {
	router, _, jwtSecret := setupTestServer()

	token := createTestJWT("testuser", "user", jwtSecret)
	req, _ := http.NewRequest("GET", "/v1/catalog/xx", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
}

func TestIntegration_GetLocalization_Success(t *testing.T) {
	router, _, jwtSecret := setupTestServer()

	token := createTestJWT("testuser", "user", jwtSecret)
	req, _ := http.NewRequest("GET", "/v1/localize/app.welcome?language=en", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
}

func TestIntegration_GetLocalization_MissingLanguage(t *testing.T) {
	router, _, jwtSecret := setupTestServer()

	token := createTestJWT("testuser", "user", jwtSecret)
	req, _ := http.NewRequest("GET", "/v1/localize/app.welcome", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestIntegration_BatchLocalize_Success(t *testing.T) {
	router, _, jwtSecret := setupTestServer()

	batchReq := models.GetBatchLocalizationRequest{
		Language: "en",
		Keys:     []string{"app.welcome", "app.error"},
		Fallback: true,
	}
	body, _ := json.Marshal(batchReq)

	token := createTestJWT("testuser", "user", jwtSecret)
	req, _ := http.NewRequest("POST", "/v1/localize/batch", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
}

func TestIntegration_ListLanguages_Success(t *testing.T) {
	router, _, jwtSecret := setupTestServer()

	token := createTestJWT("testuser", "user", jwtSecret)
	req, _ := http.NewRequest("GET", "/v1/languages", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
}

func TestIntegration_AdminEndpoint_Forbidden(t *testing.T) {
	router, _, jwtSecret := setupTestServer()

	// Regular user trying to access admin endpoint
	token := createTestJWT("testuser", "user", jwtSecret)

	langData := models.Language{
		Code: "fr",
		Name: "French",
	}
	body, _ := json.Marshal(langData)

	req, _ := http.NewRequest("POST", "/v1/admin/languages", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestIntegration_AdminEndpoint_Success(t *testing.T) {
	router, _, jwtSecret := setupTestServer()

	// Admin user accessing admin endpoint
	token := createTestJWT("adminuser", "admin", jwtSecret)

	langData := models.Language{
		Code: "fr",
		Name: "French",
	}
	body, _ := json.Marshal(langData)

	req, _ := http.NewRequest("POST", "/v1/admin/languages", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should not be forbidden
	assert.NotEqual(t, http.StatusForbidden, w.Code)
}

func TestIntegration_CachingBehavior(t *testing.T) {
	router, _, jwtSecret := setupTestServer()

	token := createTestJWT("testuser", "user", jwtSecret)

	// First request - should hit database
	req1, _ := http.NewRequest("GET", "/v1/catalog/en", nil)
	req1.Header.Set("Authorization", "Bearer "+token)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Second request - should hit cache (faster)
	req2, _ := http.NewRequest("GET", "/v1/catalog/en", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// Both should return same data
	assert.Equal(t, w1.Body.String(), w2.Body.String())
}

func TestIntegration_GetStats_Admin(t *testing.T) {
	router, _, jwtSecret := setupTestServer()

	token := createTestJWT("adminuser", "admin", jwtSecret)
	req, _ := http.NewRequest("GET", "/v1/admin/stats", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.NotNil(t, response.Data)
}

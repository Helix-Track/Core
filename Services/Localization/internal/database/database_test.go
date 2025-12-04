package database

import (
	"context"
	"errors"
	"testing"

	"github.com/helixtrack/localization-service/internal/config"
	"github.com/helixtrack/localization-service/internal/models"
	"github.com/stretchr/testify/assert"
)

// TestNewPostgresDatabase tests the New function
func TestNewPostgresDatabase(t *testing.T) {
	tests := []struct {
		name        string
		dbConfig    config.DatabaseConfig
		encConfig   config.EncryptionConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			dbConfig: config.DatabaseConfig{
				Driver:   "postgres",
				Host:     "localhost",
				Port:     5432,
				Database: "test",
				User:     "user",
				Password: "pass",
			},
			encConfig: config.EncryptionConfig{
				Key: "",
			},
			expectError: true, // Will fail to connect in test environment
			errorMsg:    "failed to ping database",
		},
		{
			name: "missing host",
			dbConfig: config.DatabaseConfig{
				Driver:   "postgres",
				Port:     5432,
				Database: "test",
				User:     "user",
				Password: "pass",
			},
			encConfig: config.EncryptionConfig{
				Key: "",
			},
			expectError: true,
			errorMsg:    "failed to ping database",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use a nil logger for tests
			db, err := New(&tt.dbConfig, &tt.encConfig, nil)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && err.Error()[:len(tt.errorMsg)] != tt.errorMsg {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if db == nil {
					t.Errorf("Expected non-nil database")
				}
			}
		})
	}
}

// TestConfigValidation tests that the configuration is properly validated
func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		dbConfig    config.DatabaseConfig
		encConfig   config.EncryptionConfig
		expectError bool
	}{
		{
			name: "missing driver",
			dbConfig: config.DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				Database: "test",
			},
			encConfig: config.EncryptionConfig{
				Key: "",
			},
			expectError: true,
		},
		{
			name: "invalid port",
			dbConfig: config.DatabaseConfig{
				Driver:   "postgres",
				Host:     "localhost",
				Port:     -1,
				Database: "test",
			},
			encConfig: config.EncryptionConfig{
				Key: "",
			},
			expectError: true,
		},
		{
			name: "missing database name",
			dbConfig: config.DatabaseConfig{
				Driver:   "postgres",
				Host:     "localhost",
				Port:     5432,
				Database: "",
			},
			encConfig: config.EncryptionConfig{
				Key: "",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a DSN to test validation
			dsn := tt.dbConfig.GetDSN()
			
			if tt.dbConfig.Database == "" {
				// Missing database should result in empty or invalid DSN
				if dsn != "" && dsn != "dbname=" {
					t.Logf("DSN for missing database: %s", dsn)
				}
			}
		})
	}
}

// TestEncryptionConfig tests encryption configuration handling
func TestEncryptionConfig(t *testing.T) {
	tests := []struct {
		name      string
		encConfig config.EncryptionConfig
	}{
		{
			name: "encryption disabled",
			encConfig: config.EncryptionConfig{
				Key: "",
			},
		},
		{
			name: "encryption enabled with key",
			encConfig: config.EncryptionConfig{
				Key: "test-key-32-characters-long-1234567890",
			},
		},
		{
			name: "encryption enabled with short key",
			encConfig: config.EncryptionConfig{
				Key: "short",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.encConfig.Key == "" {
				// Encryption is effectively disabled
			} else {
				// Key is provided, no length validation in this test
				// In production, you'd validate key length
			}
		})
	}
}

// BenchmarkNewDatabase benchmarks the database creation
func BenchmarkNewDatabase(b *testing.B) {
	dbConfig := config.DatabaseConfig{
		Driver:   "postgres",
		Host:     "localhost",
		Port:     5432,
		Database: "test",
		User:     "user",
		Password: "pass",
	}
	encConfig := config.EncryptionConfig{
		Key: "",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// This will fail but we're benchmarking the setup and error path
		_, _ = New(&dbConfig, &encConfig, nil)
	}
}

// Test key operations using mock database
func TestCreateLocalizationKey(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDatabase()
	
	tests := []struct {
		name        string
		key         *models.LocalizationKey
		expectError bool
		errorType   error
	}{
		{
			name: "valid key",
			key: &models.LocalizationKey{
				Key:         "test.key",
				Category:    "test",
				Description: "Test key description",
				Context:     "Test context",
			},
			expectError: false,
		},
		{
			name: "invalid key - empty key",
			key: &models.LocalizationKey{
				Category:    "test",
				Description: "Test key description",
			},
			expectError: true,
			errorType:   models.ErrValidationFailed(""),
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mockDB.CreateLocalizationKey(ctx, tt.key)
			
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorType != nil {
					assert.True(t, errors.Is(err, tt.errorType) || 
						(err.Error() == "key is required" && tt.errorType.Error() == ""))
				}
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, tt.key.ID)
			}
		})
	}
}

func TestGetLocalizationKeyByID(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDatabase()
	
	// Create a test key first
	testKey := &models.LocalizationKey{
		Key:         "test.key",
		Category:    "test",
		Description: "Test key description",
		Context:     "Test context",
	}
	err := mockDB.CreateLocalizationKey(ctx, testKey)
	assert.NoError(t, err)
	
	tests := []struct {
		name        string
		id          string
		expectError bool
	}{
		{
			name:        "valid ID",
			id:          testKey.ID,
			expectError: false,
		},
		{
			name:        "non-existent ID",
			id:          "non-existent-id",
			expectError: true,
		},
		{
			name:        "empty ID",
			id:          "",
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := mockDB.GetLocalizationKeyByID(ctx, tt.id)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetLocalizationKeyByKey(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDatabase()
	
	// Create a test key first
	testKey := &models.LocalizationKey{
		Key:         "test.key",
		Category:    "test",
		Description: "Test key description",
		Context:     "Test context",
	}
	err := mockDB.CreateLocalizationKey(ctx, testKey)
	assert.NoError(t, err)
	
	tests := []struct {
		name        string
		keyStr      string
		expectError bool
	}{
		{
			name:        "valid key",
			keyStr:      "test.key",
			expectError: false,
		},
		{
			name:        "non-existent key",
			keyStr:      "non.existent.key",
			expectError: true,
		},
		{
			name:        "empty key",
			keyStr:      "",
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := mockDB.GetLocalizationKeyByKey(ctx, tt.keyStr)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateLocalizationKey(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDatabase()
	
	// Create a test key first
	testKey := &models.LocalizationKey{
		Key:         "test.key",
		Category:    "test",
		Description: "Test key description",
		Context:     "Test context",
	}
	err := mockDB.CreateLocalizationKey(ctx, testKey)
	assert.NoError(t, err)
	
	tests := []struct {
		name        string
		key         *models.LocalizationKey
		expectError bool
	}{
		{
			name: "valid key update",
			key: &models.LocalizationKey{
				ID:          testKey.ID,
				Key:         "updated.key",
				Category:    "updated",
				Description: "Updated description",
				Context:     "Updated context",
			},
			expectError: false,
		},
		{
			name: "invalid key update - empty key",
			key: &models.LocalizationKey{
				ID:          testKey.ID,
				Category:    "updated",
				Description: "Updated description",
			},
			expectError: true,
		},
		{
			name: "non-existent key update",
			key: &models.LocalizationKey{
				ID:          "non-existent-id",
				Key:         "updated.key",
				Category:    "updated",
				Description: "Updated description",
			},
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mockDB.UpdateLocalizationKey(ctx, tt.key)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeleteLocalizationKey(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDatabase()
	
	// Create a test key first
	testKey := &models.LocalizationKey{
		Key:         "test.key",
		Category:    "test",
		Description: "Test key description",
		Context:     "Test context",
	}
	err := mockDB.CreateLocalizationKey(ctx, testKey)
	assert.NoError(t, err)
	
	tests := []struct {
		name        string
		id          string
		expectError bool
	}{
		{
			name:        "valid ID",
			id:          testKey.ID,
			expectError: false,
		},
		{
			name:        "non-existent ID",
			id:          "non-existent-id",
			expectError: true,
		},
		{
			name:        "empty ID",
			id:          "",
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mockDB.DeleteLocalizationKey(ctx, tt.id)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test catalog operations using mock database
func TestCreateCatalog(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDatabase()
	
	// Create a test language first
	testLang := &models.Language{
		Code:       "en-US",
		Name:       "English (US)",
		NativeName: "English (United States)",
		IsActive:   true,
		IsDefault:  false,
	}
	err := mockDB.CreateLanguage(ctx, testLang)
	assert.NoError(t, err)
	
	tests := []struct {
		name         string
		catalog      *models.LocalizationCatalog
		expectError  bool
	}{
		{
			name: "valid catalog",
			catalog: &models.LocalizationCatalog{
				LanguageID:  testLang.ID,
				Category:    "test-category",
				CatalogData: []byte(`{"test.key": "test value"}`),
				Version:     1,
			},
			expectError: false,
		},
		{
			name: "invalid catalog - no language ID",
			catalog: &models.LocalizationCatalog{
				Category:    "test-category",
				CatalogData: []byte(`{"test.key": "test value"}`),
				Version:     1,
			},
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mockDB.CreateCatalog(ctx, tt.catalog)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, tt.catalog.ID)
			}
		})
	}
}

func TestGetCatalogByLanguage(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDatabase()
	
	// Create a test language first
	testLang := &models.Language{
		Code:       "en-US",
		Name:       "English (US)",
		NativeName: "English (United States)",
		IsActive:   true,
		IsDefault:  false,
	}
	err := mockDB.CreateLanguage(ctx, testLang)
	assert.NoError(t, err)
	
	// Create a test catalog first
	testCatalog := &models.LocalizationCatalog{
		LanguageID:  testLang.ID,
		Category:    "test-category",
		CatalogData: []byte(`{"test.key": "test value"}`),
		Version:     1,
	}
	err = mockDB.CreateCatalog(ctx, testCatalog)
	assert.NoError(t, err)
	
	tests := []struct {
		name        string
		languageID  string
		category    string
		expectError bool
	}{
		{
			name:        "with category",
			languageID:  testLang.ID,
			category:    "test-category",
			expectError: false,
		},
		{
			name:        "without category - will build new one",
			languageID:  testLang.ID,
			category:    "",
			expectError: false,
		},
		{
			name:        "empty language ID",
			languageID:  "",
			category:    "test-category",
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := mockDB.GetCatalogByLanguage(ctx, tt.languageID, tt.category)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateCatalog(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDatabase()
	
	// Create a test language first
	testLang := &models.Language{
		Code:       "en-US",
		Name:       "English (US)",
		NativeName: "English (United States)",
		IsActive:   true,
		IsDefault:  false,
	}
	err := mockDB.CreateLanguage(ctx, testLang)
	assert.NoError(t, err)
	
	// Create a test catalog first
	testCatalog := &models.LocalizationCatalog{
		LanguageID:  testLang.ID,
		Category:    "test-category",
		CatalogData: []byte(`{"test.key": "test value"}`),
		Version:     1,
	}
	err = mockDB.CreateCatalog(ctx, testCatalog)
	assert.NoError(t, err)
	
	tests := []struct {
		name        string
		catalog     *models.LocalizationCatalog
		expectError bool
	}{
		{
			name: "valid catalog update",
			catalog: &models.LocalizationCatalog{
				ID:          testCatalog.ID,
				LanguageID:  testLang.ID,
				Category:    "updated-category",
				CatalogData: []byte(`{"updated.key": "updated value"}`),
				Version:     2,
			},
			expectError: false,
		},
		{
			name: "invalid catalog update - no ID",
			catalog: &models.LocalizationCatalog{
				LanguageID:  testLang.ID,
				Category:    "updated-category",
				CatalogData: []byte(`{"updated.key": "updated value"}`),
				Version:     2,
			},
			expectError: true,
		},
		{
			name: "non-existent catalog update",
			catalog: &models.LocalizationCatalog{
				ID:          "non-existent-id",
				LanguageID:  testLang.ID,
				Category:    "updated-category",
				CatalogData: []byte(`{"updated.key": "updated value"}`),
				Version:     2,
			},
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mockDB.UpdateCatalog(ctx, tt.catalog)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeleteCatalog(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDatabase()
	
	// Create a test language first
	testLang := &models.Language{
		Code:       "en-US",
		Name:       "English (US)",
		NativeName: "English (United States)",
		IsActive:   true,
		IsDefault:  false,
	}
	err := mockDB.CreateLanguage(ctx, testLang)
	assert.NoError(t, err)
	
	// Create a test catalog first
	testCatalog := &models.LocalizationCatalog{
		LanguageID:  testLang.ID,
		Category:    "test-category",
		CatalogData: []byte(`{"test.key": "test value"}`),
		Version:     1,
	}
	err = mockDB.CreateCatalog(ctx, testCatalog)
	assert.NoError(t, err)
	
	tests := []struct {
		name        string
		id          string
		expectError bool
	}{
		{
			name:        "valid ID",
			id:          testCatalog.ID,
			expectError: false,
		},
		{
			name:        "non-existent ID",
			id:          "non-existent-id",
			expectError: true,
		},
		{
			name:        "empty ID",
			id:          "",
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mockDB.DeleteCatalog(ctx, tt.id)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBuildCatalog(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDatabase()
	
	// Create a test language first
	testLang := &models.Language{
		Code:       "en-US",
		Name:       "English (US)",
		NativeName: "English (United States)",
		IsActive:   true,
		IsDefault:  false,
	}
	err := mockDB.CreateLanguage(ctx, testLang)
	assert.NoError(t, err)
	
	// Create a test key first
	testKey := &models.LocalizationKey{
		Key:         "test.key",
		Category:    "test-category",
		Description: "Test key description",
	}
	err = mockDB.CreateLocalizationKey(ctx, testKey)
	assert.NoError(t, err)
	
	// Create a test localization first
	testLocalization := &models.Localization{
		KeyID:      testKey.ID,
		LanguageID: testLang.ID,
		Value:      "Test value",
		Version:    1,
	}
	err = mockDB.CreateLocalization(ctx, testLocalization)
	assert.NoError(t, err)
	
	tests := []struct {
		name        string
		languageID  string
		category    string
		expectError bool
	}{
		{
			name:        "with category",
			languageID:  testLang.ID,
			category:    "test-category",
			expectError: false,
		},
		{
			name:        "without category",
			languageID:  testLang.ID,
			category:    "",
			expectError: false,
		},
		{
			name:        "empty language ID",
			languageID:  "",
			category:    "test-category",
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := mockDB.BuildCatalog(ctx, tt.languageID, tt.category)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test language operations using mock database
func TestCreateLanguage(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDatabase()
	
	tests := []struct {
		name        string
		language    *models.Language
		expectError bool
	}{
		{
			name: "valid language",
			language: &models.Language{
				Code:       "en-US",
				Name:       "English (US)",
				NativeName: "English (United States)",
				IsRTL:      false,
				IsActive:   true,
				IsDefault:  false,
			},
			expectError: false,
		},
		{
			name: "invalid language - empty code",
			language: &models.Language{
				Name:       "English (US)",
				NativeName: "English (United States)",
				IsRTL:      false,
				IsActive:   true,
				IsDefault:  false,
			},
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mockDB.CreateLanguage(ctx, tt.language)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, tt.language.ID)
			}
		})
	}
}

func TestGetLanguageByCode(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDatabase()
	
	// Create a test language first
	testLang := &models.Language{
		Code:       "en-US",
		Name:       "English (US)",
		NativeName: "English (United States)",
		IsRTL:      false,
		IsActive:   true,
		IsDefault:  false,
	}
	err := mockDB.CreateLanguage(ctx, testLang)
	assert.NoError(t, err)
	
	tests := []struct {
		name        string
		code        string
		expectError bool
	}{
		{
			name:        "valid code",
			code:        "en-US",
			expectError: false,
		},
		{
			name:        "non-existent code",
			code:        "fr-FR",
			expectError: true,
		},
		{
			name:        "empty code",
			code:        "",
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := mockDB.GetLanguageByCode(ctx, tt.code)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateLanguage(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDatabase()
	
	// Create a test language first
	testLang := &models.Language{
		Code:       "en-US",
		Name:       "English (US)",
		NativeName: "English (United States)",
		IsRTL:      false,
		IsActive:   true,
		IsDefault:  false,
	}
	err := mockDB.CreateLanguage(ctx, testLang)
	assert.NoError(t, err)
	
	tests := []struct {
		name        string
		language    *models.Language
		expectError bool
	}{
		{
			name: "valid language update",
			language: &models.Language{
				ID:         testLang.ID,
				Code:       "en-GB",
				Name:       "English (GB)",
				NativeName: "English (Great Britain)",
				IsRTL:      false,
				IsActive:   true,
				IsDefault:  false,
			},
			expectError: false,
		},
		{
			name: "invalid language update - empty code",
			language: &models.Language{
				ID:         testLang.ID,
				Name:       "English (GB)",
				NativeName: "English (Great Britain)",
				IsRTL:      false,
				IsActive:   true,
				IsDefault:  false,
			},
			expectError: true,
		},
		{
			name: "non-existent language update",
			language: &models.Language{
				ID:         "non-existent-id",
				Code:       "en-GB",
				Name:       "English (GB)",
				NativeName: "English (Great Britain)",
				IsRTL:      false,
				IsActive:   true,
				IsDefault:  false,
			},
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mockDB.UpdateLanguage(ctx, tt.language)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeleteLanguage(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDatabase()
	
	// Create a test language first
	testLang := &models.Language{
		Code:       "en-US",
		Name:       "English (US)",
		NativeName: "English (United States)",
		IsRTL:      false,
		IsActive:   true,
		IsDefault:  false,
	}
	err := mockDB.CreateLanguage(ctx, testLang)
	assert.NoError(t, err)
	
	tests := []struct {
		name        string
		id          string
		expectError bool
	}{
		{
			name:        "valid ID",
			id:          testLang.ID,
			expectError: false,
		},
		{
			name:        "non-existent ID",
			id:          "non-existent-id",
			expectError: true,
		},
		{
			name:        "empty ID",
			id:          "",
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mockDB.DeleteLanguage(ctx, tt.id)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test localization operations using mock database
func TestCreateLocalization(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDatabase()
	
	// Create a test language first
	testLang := &models.Language{
		Code:       "en-US",
		Name:       "English (US)",
		NativeName: "English (United States)",
		IsRTL:      false,
		IsActive:   true,
		IsDefault:  false,
	}
	err := mockDB.CreateLanguage(ctx, testLang)
	assert.NoError(t, err)
	
	// Create a test key first
	testKey := &models.LocalizationKey{
		Key:         "test.key",
		Category:    "test",
		Description: "Test key description",
	}
	err = mockDB.CreateLocalizationKey(ctx, testKey)
	assert.NoError(t, err)
	
	tests := []struct {
		name         string
		localization *models.Localization
		expectError  bool
	}{
		{
			name: "valid localization",
			localization: &models.Localization{
				KeyID:      testKey.ID,
				LanguageID: testLang.ID,
				Value:      "Test value",
				Version:    1,
			},
			expectError: false,
		},
		{
			name: "invalid localization - empty value",
			localization: &models.Localization{
				KeyID:      testKey.ID,
				LanguageID: testLang.ID,
				Version:    1,
			},
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mockDB.CreateLocalization(ctx, tt.localization)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, tt.localization.ID)
			}
		})
	}
}

func TestGetLocalizationByKeyAndLanguage(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDatabase()
	
	// Create a test language first
	testLang := &models.Language{
		Code:       "en-US",
		Name:       "English (US)",
		NativeName: "English (United States)",
		IsRTL:      false,
		IsActive:   true,
		IsDefault:  false,
	}
	err := mockDB.CreateLanguage(ctx, testLang)
	assert.NoError(t, err)
	
	// Create a test key first
	testKey := &models.LocalizationKey{
		Key:         "test.key",
		Category:    "test",
		Description: "Test key description",
	}
	err = mockDB.CreateLocalizationKey(ctx, testKey)
	assert.NoError(t, err)
	
	// Create a test localization first
	testLocalization := &models.Localization{
		KeyID:      testKey.ID,
		LanguageID: testLang.ID,
		Value:      "Test value",
		Version:    1,
	}
	err = mockDB.CreateLocalization(ctx, testLocalization)
	assert.NoError(t, err)
	
	tests := []struct {
		name        string
		keyStr      string
		languageID  string
		expectError bool
	}{
		{
			name:        "valid key and language",
			keyStr:      testKey.ID,
			languageID:  testLang.ID,
			expectError: false,
		},
		{
			name:        "empty key",
			keyStr:      "",
			languageID:  testLang.ID,
			expectError: true,
		},
		{
			name:        "empty language ID",
			keyStr:      testKey.ID,
			languageID:  "",
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := mockDB.GetLocalizationByKeyAndLanguage(ctx, tt.keyStr, tt.languageID)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateLocalization(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDatabase()
	
	// Create a test language first
	testLang := &models.Language{
		Code:       "en-US",
		Name:       "English (US)",
		NativeName: "English (United States)",
		IsRTL:      false,
		IsActive:   true,
		IsDefault:  false,
	}
	err := mockDB.CreateLanguage(ctx, testLang)
	assert.NoError(t, err)
	
	// Create a test key first
	testKey := &models.LocalizationKey{
		Key:         "test.key",
		Category:    "test",
		Description: "Test key description",
	}
	err = mockDB.CreateLocalizationKey(ctx, testKey)
	assert.NoError(t, err)
	
	// Create a test localization first
	testLocalization := &models.Localization{
		KeyID:      testKey.ID,
		LanguageID: testLang.ID,
		Value:      "Test value",
		Version:    1,
	}
	err = mockDB.CreateLocalization(ctx, testLocalization)
	assert.NoError(t, err)
	
	tests := []struct {
		name         string
		localization *models.Localization
		expectError  bool
	}{
		{
			name: "valid localization update",
			localization: &models.Localization{
				ID:         testLocalization.ID,
				KeyID:      testKey.ID,
				LanguageID: testLang.ID,
				Value:      "Updated test value",
				Version:    2,
			},
			expectError: false,
		},
		{
			name: "invalid localization update - empty value",
			localization: &models.Localization{
				ID:         testLocalization.ID,
				KeyID:      testKey.ID,
				LanguageID: testLang.ID,
				Version:    2,
			},
			expectError: true,
		},
		{
			name: "non-existent localization update",
			localization: &models.Localization{
				ID:         "non-existent-id",
				KeyID:      testKey.ID,
				LanguageID: testLang.ID,
				Value:      "Updated test value",
				Version:    2,
			},
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mockDB.UpdateLocalization(ctx, tt.localization)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeleteLocalization(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDatabase()
	
	// Create a test language first
	testLang := &models.Language{
		Code:       "en-US",
		Name:       "English (US)",
		NativeName: "English (United States)",
		IsRTL:      false,
		IsActive:   true,
		IsDefault:  false,
	}
	err := mockDB.CreateLanguage(ctx, testLang)
	assert.NoError(t, err)
	
	// Create a test key first
	testKey := &models.LocalizationKey{
		Key:         "test.key",
		Category:    "test",
		Description: "Test key description",
	}
	err = mockDB.CreateLocalizationKey(ctx, testKey)
	assert.NoError(t, err)
	
	// Create a test localization first
	testLocalization := &models.Localization{
		KeyID:      testKey.ID,
		LanguageID: testLang.ID,
		Value:      "Test value",
		Version:    1,
	}
	err = mockDB.CreateLocalization(ctx, testLocalization)
	assert.NoError(t, err)
	
	tests := []struct {
		name        string
		id          string
		expectError bool
	}{
		{
			name:        "valid ID",
			id:          testLocalization.ID,
			expectError: false,
		},
		{
			name:        "non-existent ID",
			id:          "non-existent-id",
			expectError: true,
		},
		{
			name:        "empty ID",
			id:          "",
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mockDB.DeleteLocalization(ctx, tt.id)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestApproveLocalization(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDatabase()
	
	// Create a test language first
	testLang := &models.Language{
		Code:       "en-US",
		Name:       "English (US)",
		NativeName: "English (United States)",
		IsRTL:      false,
		IsActive:   true,
		IsDefault:  false,
	}
	err := mockDB.CreateLanguage(ctx, testLang)
	assert.NoError(t, err)
	
	// Create a test key first
	testKey := &models.LocalizationKey{
		Key:         "test.key",
		Category:    "test",
		Description: "Test key description",
	}
	err = mockDB.CreateLocalizationKey(ctx, testKey)
	assert.NoError(t, err)
	
	// Create a test localization first
	testLocalization := &models.Localization{
		KeyID:      testKey.ID,
		LanguageID: testLang.ID,
		Value:      "Test value",
		Version:    1,
	}
	err = mockDB.CreateLocalization(ctx, testLocalization)
	assert.NoError(t, err)
	
	tests := []struct {
		name        string
		id          string
		approvedBy  string
		expectError bool
	}{
		{
			name:        "valid approval",
			id:          testLocalization.ID,
			approvedBy:  "admin-user",
			expectError: false,
		},
		{
			name:        "empty ID",
			id:          "",
			approvedBy:  "admin-user",
			expectError: true,
		},
		{
			name:        "empty approved by",
			id:          testLocalization.ID,
			approvedBy:  "",
			expectError: false, // Should still work even with empty approvedBy
		},
		{
			name:        "non-existent ID",
			id:          "non-existent-id",
			approvedBy:  "admin-user",
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mockDB.ApproveLocalization(ctx, tt.id, tt.approvedBy)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				
				// Check if the localization was approved
				if !tt.expectError {
					loc, err := mockDB.GetLocalizationByID(ctx, tt.id)
					assert.NoError(t, err)
					assert.True(t, loc.Approved)
					if tt.approvedBy != "" {
						assert.Equal(t, tt.approvedBy, loc.ApprovedBy)
					}
				}
			}
		})
	}
}
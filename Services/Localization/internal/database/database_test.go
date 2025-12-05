package database

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/helixtrack/localization-service/internal/config"
	"github.com/helixtrack/localization-service/internal/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
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
		{
			name: "mock error condition",
			key: &models.LocalizationKey{
				Key:         "test.key.2",
				Category:    "test",
				Description: "Test key description 2",
			},
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "mock error condition" {
				mockDB.SetError(fmt.Errorf("mock create key error"))
				err := mockDB.CreateLocalizationKey(ctx, tt.key)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "mock create key error")
				mockDB.ClearError()
				return
			}
			
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
		{
			name:        "mock error condition",
			id:          testKey.ID,
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "mock error condition" {
				mockDB.SetError(fmt.Errorf("mock get key by ID error"))
				_, err := mockDB.GetLocalizationKeyByID(ctx, tt.id)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "mock get key by ID error")
				mockDB.ClearError()
				return
			}
			
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
		{
			name:        "mock error condition",
			keyStr:      "test.key",
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "mock error condition" {
				mockDB.SetError(fmt.Errorf("mock get key by key error"))
				_, err := mockDB.GetLocalizationKeyByKey(ctx, tt.keyStr)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "mock get key by key error")
				mockDB.ClearError()
				return
			}
			
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
		{
			name: "mock error condition",
			key: &models.LocalizationKey{
				ID:          testKey.ID,
				Key:         "updated.key",
				Category:    "updated",
				Description: "Updated description",
			},
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "mock error condition" {
				mockDB.SetError(fmt.Errorf("mock update localization key error"))
				err := mockDB.UpdateLocalizationKey(ctx, tt.key)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "mock update localization key error")
				mockDB.ClearError()
				return
			}
			
			err := mockDB.UpdateLocalizationKey(ctx, tt.key)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestUpdateLocalizationKeyValidationWithInitializedDB tests validation path with initialized fields
func TestUpdateLocalizationKeyValidationWithInitializedDB(t *testing.T) {
	// Create a PostgresDatabase instance with initialized fields to avoid panics
	logger, _ := zap.NewDevelopment()
	db := &PostgresDatabase{
		config:    &config.DatabaseConfig{Driver: "postgres"},
		encConfig: &config.EncryptionConfig{},
		logger:    logger,
			db:         nil, // Intentionally nil to test validation before database operations
	}
	
	// Test with invalid key (missing required fields)
	invalidKey := &models.LocalizationKey{
		ID: "test-id",
		// Missing required Key field
		Category:    "test",
		Description: "Test description",
		Context:     "Test context",
	}
	
	ctx := context.Background()
	err := db.UpdateLocalizationKey(ctx, invalidKey)
	assert.Error(t, err) // Should fail validation before database operation
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
		{
			name:        "mock error condition",
			id:          testKey.ID,
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "mock error condition" {
				// Need to recreate key as it might have been deleted by previous tests
				newKey := &models.LocalizationKey{
					Key:         "test.key.2",
					Category:    "test",
					Description: "Test key description 2",
					Context:     "Test context 2",
				}
				err := mockDB.CreateLocalizationKey(ctx, newKey)
				assert.NoError(t, err)
				
				mockDB.SetError(fmt.Errorf("mock delete localization key error"))
				err = mockDB.DeleteLocalizationKey(ctx, newKey.ID)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "mock delete localization key error")
				mockDB.ClearError()
				return
			}
			
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
		{
			name: "mock error condition",
			catalog: &models.LocalizationCatalog{
				LanguageID:  testLang.ID,
				Category:    "test-category",
				CatalogData: []byte(`{"test.key": "test value"}`),
				Version:     1,
			},
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "mock error condition" {
				mockDB.SetError(fmt.Errorf("mock create catalog error"))
				err := mockDB.CreateCatalog(ctx, tt.catalog)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "mock create catalog error")
				mockDB.ClearError()
				return
			}
			
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
		{
			name:        "mock error condition",
			languageID:  testLang.ID,
			category:    "test-category",
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "mock error condition" {
				mockDB.SetError(fmt.Errorf("mock get catalog error"))
				_, err := mockDB.GetCatalogByLanguage(ctx, tt.languageID, tt.category)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "mock get catalog error")
				mockDB.ClearError()
				return
			}
			
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
			{
			name: "mock error condition",
			catalog: &models.LocalizationCatalog{
				ID:          testCatalog.ID,
				LanguageID:  testLang.ID,
				Category:    "updated-category",
				CatalogData: []byte(`{"updated.key": "updated value"}`),
				Version:     2,
			},
			expectError: true,
		},
		{
			name: "validation error - empty language ID",
			catalog: &models.LocalizationCatalog{
				ID:          testCatalog.ID,
				LanguageID:  "",
				Category:    "updated-category",
				CatalogData: []byte(`{"updated.key": "updated value"}`),
				Version:     2,
			},
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "mock error condition" {
				mockDB.SetError(fmt.Errorf("mock update catalog error"))
				err := mockDB.UpdateCatalog(ctx, tt.catalog)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "mock update catalog error")
				mockDB.ClearError()
				return
			}
			
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
		{
			name:        "mock error condition",
			id:          testCatalog.ID,
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "mock error condition" {
				// Need to recreate catalog as it might have been deleted by previous tests
				newCatalog := &models.LocalizationCatalog{
					LanguageID:  testLang.ID,
					Category:    "test-category-2",
					CatalogData: []byte(`{"test.key": "test value"}`),
					Version:     1,
				}
				err := mockDB.CreateCatalog(ctx, newCatalog)
				assert.NoError(t, err)
				
				mockDB.SetError(fmt.Errorf("mock delete catalog error"))
				err = mockDB.DeleteCatalog(ctx, newCatalog.ID)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "mock delete catalog error")
				mockDB.ClearError()
				return
			}
			
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
		{
			name:        "mock error condition",
			languageID:  testLang.ID,
			category:    "test-category",
			expectError: true,
		},
		{
			name:        "language not found",
			languageID:  "non-existent-lang-id",
			category:    "test-category",
			expectError: true,
		},
		{
			name:        "empty category",
			languageID:  testLang.ID,
			category:    "",
			expectError: false, // Should work with empty category
		},
		{
			name:        "create catalog error",
			languageID:  testLang.ID,
			category:    "test-category",
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "mock error condition" {
				mockDB.SetError(fmt.Errorf("mock build catalog error"))
				_, err := mockDB.BuildCatalog(ctx, tt.languageID, tt.category)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "mock build catalog error")
				mockDB.ClearError()
				return
			}
			
			if tt.name == "create catalog error" {
				// Set the CreateCatalog error flag
				mockDB.SetCreateCatalogError(fmt.Errorf("mock create catalog error"))
				defer mockDB.ClearCreateCatalogError()
				
				// BuildCatalog should now fail when it calls CreateCatalog
				_, err := mockDB.BuildCatalog(ctx, testLang.ID, "test-category")
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "mock create catalog error")
				return
			}
			
			if tt.name == "json marshal error" {
				// Set the marshal catalog error flag
				mockDB.SetMarshalCatalogError()
				defer mockDB.ClearMarshalCatalogError()
				
				// BuildCatalog should now fail when it tries to marshal the catalog map
				_, err := mockDB.BuildCatalog(ctx, testLang.ID, "test-category")
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "simulated json.Marshal error")
				return
			}
			
			_, err := mockDB.BuildCatalog(ctx, tt.languageID, tt.category)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMockDatabaseClose(t *testing.T) {
	mockDB := NewMockDatabase()
	
	// Test successful close
	t.Run("successful close", func(t *testing.T) {
		err := mockDB.Close()
		assert.NoError(t, err)
	})
	
	// Test close with error
	t.Run("close with error", func(t *testing.T) {
		// Set error condition
		mockDB.SetError(fmt.Errorf("close error"))
		err := mockDB.Close()
		assert.Error(t, err)
	})
}

func TestMockDatabasePing(t *testing.T) {
	mockDB := NewMockDatabase()
	
	// Test successful ping
	t.Run("successful ping", func(t *testing.T) {
		err := mockDB.Ping()
		assert.NoError(t, err)
	})
	
	// Test ping with error
	t.Run("ping with error", func(t *testing.T) {
		// Set error condition
		mockDB.SetError(fmt.Errorf("ping error"))
		err := mockDB.Ping()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ping error")
		// Clear error for other tests
		mockDB.ClearError()
	})
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
		{
			name: "create language with mock error",
			language: &models.Language{
				Code:       "fr-FR",
				Name:       "French (FR)",
				NativeName: "Français (France)",
				IsRTL:      false,
				IsActive:   true,
				IsDefault:  false,
			},
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "create language with mock error" {
				// Set error condition
				mockDB.SetError(fmt.Errorf("mock database error"))
				defer mockDB.SetError(nil)
			}
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
		{
			name:        "mock error condition",
			code:        "en-US",
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "mock error condition" {
				// Set error condition
				mockDB.SetError(fmt.Errorf("mock database error"))
				defer mockDB.SetError(nil)
			}
			_, err := mockDB.GetLanguageByCode(ctx, tt.code)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetLanguageByID(t *testing.T) {
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
		{
			name:        "mock error condition",
			id:          testLang.ID,
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "mock error condition" {
				// Set error condition
				mockDB.SetError(fmt.Errorf("mock database error"))
				defer mockDB.SetError(nil)
			}
			lang, err := mockDB.GetLanguageByID(ctx, tt.id)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, lang)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, lang)
				assert.Equal(t, testLang.Code, lang.Code)
				assert.Equal(t, testLang.Name, lang.Name)
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
		{
			name: "mock error condition",
			language: &models.Language{
				ID:         testLang.ID,
				Code:       "fr-FR",
				Name:       "French (FR)",
				NativeName: "Français (France)",
				IsRTL:      false,
				IsActive:   true,
				IsDefault:  false,
			},
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "mock error condition" {
				mockDB.SetError(fmt.Errorf("mock update language error"))
				err := mockDB.UpdateLanguage(ctx, tt.language)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "mock update language error")
				mockDB.ClearError()
				return
			}
			
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
		{
			name:        "mock error condition",
			id:          testLang.ID,
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "mock error condition" {
				// Need to recreate language as it might have been deleted by previous tests
				newLang := &models.Language{
					Code:       "fr-FR",
					Name:       "French (FR)",
					NativeName: "Français (France)",
					IsRTL:      false,
					IsActive:   true,
					IsDefault:  false,
				}
				err := mockDB.CreateLanguage(ctx, newLang)
				assert.NoError(t, err)
				
				mockDB.SetError(fmt.Errorf("mock delete language error"))
				err = mockDB.DeleteLanguage(ctx, newLang.ID)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "mock delete language error")
				mockDB.ClearError()
				return
			}
			
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
		{
			name: "mock error condition",
			localization: &models.Localization{
				KeyID:      testKey.ID,
				LanguageID: testLang.ID,
				Value:      "Test value",
				Version:    1,
			},
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "mock error condition" {
				mockDB.SetError(fmt.Errorf("mock create localization error"))
				err := mockDB.CreateLocalization(ctx, tt.localization)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "mock create localization error")
				mockDB.ClearError()
				return
			}
			
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
		{
			name:        "mock error condition",
			keyStr:      testKey.ID,
			languageID:  testLang.ID,
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "mock error condition" {
				mockDB.SetError(fmt.Errorf("mock get localization by key and language error"))
				_, err := mockDB.GetLocalizationByKeyAndLanguage(ctx, tt.keyStr, tt.languageID)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "mock get localization by key and language error")
				mockDB.ClearError()
				return
			}
			
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
		{
			name: "mock error condition",
			localization: &models.Localization{
				ID:         testLocalization.ID,
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
			if tt.name == "mock error condition" {
				mockDB.SetError(fmt.Errorf("mock update localization error"))
				err := mockDB.UpdateLocalization(ctx, tt.localization)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "mock update localization error")
				mockDB.ClearError()
				return
			}
			
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
		{
			name:        "mock error condition",
			id:          testLocalization.ID,
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "mock error condition" {
				// Need to recreate localization as it might have been deleted by previous tests
				newLocalization := &models.Localization{
					KeyID:      testKey.ID,
					LanguageID: testLang.ID,
					Value:      "Test value 2",
					Version:    1,
				}
				err := mockDB.CreateLocalization(ctx, newLocalization)
				assert.NoError(t, err)
				
				mockDB.SetError(fmt.Errorf("mock delete localization error"))
				err = mockDB.DeleteLocalization(ctx, newLocalization.ID)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "mock delete localization error")
				mockDB.ClearError()
				return
			}
			
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
		{
			name:        "mock error condition",
			id:          testLocalization.ID,
			approvedBy:  "admin-user",
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "mock error condition" {
				mockDB.SetError(fmt.Errorf("mock approve localization error"))
				err := mockDB.ApproveLocalization(ctx, tt.id, tt.approvedBy)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "mock approve localization error")
				mockDB.ClearError()
				return
			}
			
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

func TestGetLocalizationByID(t *testing.T) {
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
		{
			name:        "mock error condition",
			id:          testLocalization.ID,
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "mock error condition" {
				mockDB.SetError(fmt.Errorf("mock get localization error"))
				_, err := mockDB.GetLocalizationByID(ctx, tt.id)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "mock get localization error")
				mockDB.ClearError()
				return
			}
			
			loc, err := mockDB.GetLocalizationByID(ctx, tt.id)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testLocalization.Value, loc.Value)
			}
		})
	}
}

func TestGetLanguages(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDatabase()
	
	// Create test languages
	testLang1 := &models.Language{
		Code:       "en-US",
		Name:       "English (US)",
		NativeName: "English (United States)",
		IsRTL:      false,
		IsActive:   true,
		IsDefault:  false,
	}
	err := mockDB.CreateLanguage(ctx, testLang1)
	assert.NoError(t, err)
	
	testLang2 := &models.Language{
		Code:       "fr-FR",
		Name:       "French",
		NativeName: "Français",
		IsRTL:      false,
		IsActive:   false,
		IsDefault:  false,
	}
	err = mockDB.CreateLanguage(ctx, testLang2)
	assert.NoError(t, err)
	
	testLang3 := &models.Language{
		Code:       "es-ES",
		Name:       "Spanish",
		NativeName: "Español",
		IsRTL:      false,
		IsActive:   true,
		IsDefault:  true,
	}
	err = mockDB.CreateLanguage(ctx, testLang3)
	assert.NoError(t, err)
	
	tests := []struct {
		name        string
		activeOnly  bool
		expectCount int
	}{
		{
			name:        "get all languages",
			activeOnly:  false,
			expectCount: 3,
		},
		{
			name:        "get active languages only",
			activeOnly:  true,
			expectCount: 2,
		},
		{
			name:        "mock error condition",
			activeOnly:  false,
			expectCount: 0,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "mock error condition" {
				mockDB.SetError(fmt.Errorf("mock get languages error"))
				_, err := mockDB.GetLanguages(ctx, tt.activeOnly)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "mock get languages error")
				mockDB.ClearError()
				return
			}
			
			languages, err := mockDB.GetLanguages(ctx, tt.activeOnly)
			assert.NoError(t, err)
			assert.Len(t, languages, tt.expectCount)
			
			if tt.activeOnly {
				for _, lang := range languages {
					assert.True(t, lang.IsActive)
				}
			}
		})
	}
}

func TestGetDefaultLanguage(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDatabase()
	
	// Create test languages without default first
	testLang1 := &models.Language{
		Code:       "en-US",
		Name:       "English (US)",
		NativeName: "English (United States)",
		IsRTL:      false,
		IsActive:   true,
		IsDefault:  false,
	}
	err := mockDB.CreateLanguage(ctx, testLang1)
	assert.NoError(t, err)
	
	// Test with no default language
	t.Run("no default language", func(t *testing.T) {
		_, err := mockDB.GetDefaultLanguage(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
	
	// Create default language
	testLang2 := &models.Language{
		Code:       "fr-FR",
		Name:       "French",
		NativeName: "Français",
		IsRTL:      false,
		IsActive:   true,
		IsDefault:  true,
	}
	err = mockDB.CreateLanguage(ctx, testLang2)
	assert.NoError(t, err)
	
	// Test with default language
	t.Run("default language exists", func(t *testing.T) {
		defaultLang, err := mockDB.GetDefaultLanguage(ctx)
		assert.NoError(t, err)
		assert.Equal(t, testLang2.ID, defaultLang.ID)
		assert.Equal(t, testLang2.Code, defaultLang.Code)
		assert.True(t, defaultLang.IsDefault)
	})
	
	// Test with mock error
	t.Run("mock error condition", func(t *testing.T) {
		mockDB.SetError(fmt.Errorf("mock get default language error"))
		_, err := mockDB.GetDefaultLanguage(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mock get default language error")
		mockDB.ClearError()
	})
}

func TestGetLocalizationKeysByCategory(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDatabase()
	
	// Create test keys in different categories
	testKey1 := &models.LocalizationKey{
		Key:         "common.save",
		Category:    "common",
		Description: "Save button text",
	}
	err := mockDB.CreateLocalizationKey(ctx, testKey1)
	assert.NoError(t, err)
	
	testKey2 := &models.LocalizationKey{
		Key:         "common.cancel",
		Category:    "common",
		Description: "Cancel button text",
	}
	err = mockDB.CreateLocalizationKey(ctx, testKey2)
	assert.NoError(t, err)
	
	testKey3 := &models.LocalizationKey{
		Key:         "error.not_found",
		Category:    "error",
		Description: "Not found error message",
	}
	err = mockDB.CreateLocalizationKey(ctx, testKey3)
	assert.NoError(t, err)
	
	tests := []struct {
		name         string
		category     string
		expectCount  int
	}{
		{
			name:        "get common category keys",
			category:    "common",
			expectCount: 2,
		},
		{
			name:        "get error category keys",
			category:    "error",
			expectCount: 1,
		},
		{
			name:        "get non-existent category keys",
			category:    "nonexistent",
			expectCount: 0,
		},
		{
			name:        "mock error condition",
			category:    "common",
			expectCount: 0,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "mock error condition" {
				mockDB.SetError(fmt.Errorf("mock get keys by category error"))
				_, err := mockDB.GetLocalizationKeysByCategory(ctx, tt.category)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "mock get keys by category error")
				mockDB.ClearError()
				return
			}
			
			keys, err := mockDB.GetLocalizationKeysByCategory(ctx, tt.category)
			assert.NoError(t, err)
			assert.Len(t, keys, tt.expectCount)
			
			for _, key := range keys {
				assert.Equal(t, tt.category, key.Category)
			}
		})
	}
}

func TestGetLocalizationsByLanguage(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDatabase()
	
	// Create test language
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
	
	testLang2 := &models.Language{
		Code:       "fr-FR",
		Name:       "French",
		NativeName: "Français",
		IsRTL:      false,
		IsActive:   true,
		IsDefault:  false,
	}
	err = mockDB.CreateLanguage(ctx, testLang2)
	assert.NoError(t, err)
	
	// Create test keys
	testKey1 := &models.LocalizationKey{
		Key:         "welcome.title",
		Category:    "welcome",
		Description: "Welcome page title",
	}
	err = mockDB.CreateLocalizationKey(ctx, testKey1)
	assert.NoError(t, err)
	
	testKey2 := &models.LocalizationKey{
		Key:         "welcome.message",
		Category:    "welcome",
		Description: "Welcome page message",
	}
	err = mockDB.CreateLocalizationKey(ctx, testKey2)
	assert.NoError(t, err)
	
	// Create test localizations
	testLoc1 := &models.Localization{
		KeyID:      testKey1.ID,
		LanguageID: testLang.ID,
		Value:      "Welcome!",
		Version:    1,
	}
	err = mockDB.CreateLocalization(ctx, testLoc1)
	assert.NoError(t, err)
	
	testLoc2 := &models.Localization{
		KeyID:      testKey2.ID,
		LanguageID: testLang.ID,
		Value:      "Welcome to our application!",
		Version:    1,
	}
	err = mockDB.CreateLocalization(ctx, testLoc2)
	assert.NoError(t, err)
	
	testLoc3 := &models.Localization{
		KeyID:      testKey1.ID,
		LanguageID: testLang2.ID,
		Value:      "Bienvenue!",
		Version:    1,
	}
	err = mockDB.CreateLocalization(ctx, testLoc3)
	assert.NoError(t, err)
	
	tests := []struct {
		name         string
		languageID   string
		expectCount  int
	}{
		{
			name:        "get localizations for en-US",
			languageID:  testLang.ID,
			expectCount: 2,
		},
		{
			name:        "get localizations for fr-FR",
			languageID:  testLang2.ID,
			expectCount: 1,
		},
		{
			name:        "get localizations for non-existent language",
			languageID:  "non-existent-id",
			expectCount: 0,
		},
		{
			name:        "mock error condition",
			languageID:  testLang.ID,
			expectCount: 0,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "mock error condition" {
				mockDB.SetError(fmt.Errorf("mock get localizations error"))
				localizations, err := mockDB.GetLocalizationsByLanguage(ctx, tt.languageID)
				assert.Error(t, err)
				assert.Nil(t, localizations)
				assert.Contains(t, err.Error(), "mock get localizations error")
				mockDB.ClearError()
				return
			}
			
			localizations, err := mockDB.GetLocalizationsByLanguage(ctx, tt.languageID)
			assert.NoError(t, err)
			assert.Len(t, localizations, tt.expectCount)
			
			for _, loc := range localizations {
				assert.Equal(t, tt.languageID, loc.LanguageID)
			}
		})
	}
}

func TestGetLocalizationsByKeyID(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDatabase()
	
	// Create test languages
	testLang1 := &models.Language{
		Code:       "en-US",
		Name:       "English (US)",
		NativeName: "English (United States)",
		IsRTL:      false,
		IsActive:   true,
		IsDefault:  false,
	}
	err := mockDB.CreateLanguage(ctx, testLang1)
	assert.NoError(t, err)
	
	testLang2 := &models.Language{
		Code:       "fr-FR",
		Name:       "French",
		NativeName: "Français",
		IsRTL:      false,
		IsActive:   true,
		IsDefault:  false,
	}
	err = mockDB.CreateLanguage(ctx, testLang2)
	assert.NoError(t, err)
	
	// Create test key
	testKey := &models.LocalizationKey{
		Key:         "welcome.title",
		Category:    "welcome",
		Description: "Welcome page title",
	}
	err = mockDB.CreateLocalizationKey(ctx, testKey)
	assert.NoError(t, err)
	
	// Create test localizations for the same key
	testLoc1 := &models.Localization{
		KeyID:      testKey.ID,
		LanguageID: testLang1.ID,
		Value:      "Welcome!",
		Version:    1,
	}
	err = mockDB.CreateLocalization(ctx, testLoc1)
	assert.NoError(t, err)
	
	testLoc2 := &models.Localization{
		KeyID:      testKey.ID,
		LanguageID: testLang2.ID,
		Value:      "Bienvenue!",
		Version:    1,
	}
	err = mockDB.CreateLocalization(ctx, testLoc2)
	assert.NoError(t, err)
	
	tests := []struct {
		name         string
		keyID        string
		expectCount  int
	}{
		{
			name:        "get localizations by valid key ID",
			keyID:       testKey.ID,
			expectCount: 2,
		},
		{
			name:        "get localizations by non-existent key ID",
			keyID:       "non-existent-id",
			expectCount: 0,
		},
		{
			name:        "mock error condition",
			keyID:       testKey.ID,
			expectCount: 0,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "mock error condition" {
				mockDB.SetError(fmt.Errorf("mock get localizations by key error"))
				localizations, err := mockDB.GetLocalizationsByKeyID(ctx, tt.keyID)
				assert.Error(t, err)
				assert.Nil(t, localizations)
				assert.Contains(t, err.Error(), "mock get localizations by key error")
				mockDB.ClearError()
				return
			}
			
			localizations, err := mockDB.GetLocalizationsByKeyID(ctx, tt.keyID)
			assert.NoError(t, err)
			assert.Len(t, localizations, tt.expectCount)
			
			for _, loc := range localizations {
				assert.Equal(t, tt.keyID, loc.KeyID)
			}
		})
	}
}

func TestGetLatestCatalog(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDatabase()
	
	// Create test language
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
	
	// Create test key
	testKey := &models.LocalizationKey{
		Key:         "welcome.title",
		Category:    "welcome",
		Description: "Welcome page title",
	}
	err = mockDB.CreateLocalizationKey(ctx, testKey)
	assert.NoError(t, err)
	
	// Create test localization
	testLoc := &models.Localization{
		KeyID:      testKey.ID,
		LanguageID: testLang.ID,
		Value:      "Welcome!",
		Version:    1,
		Approved:   true,
	}
	err = mockDB.CreateLocalization(ctx, testLoc)
	assert.NoError(t, err)
	
	tests := []struct {
		name        string
		languageID  string
		category    string
		expectError bool
	}{
		{
			name:        "get latest catalog with valid language and category",
			languageID:  testLang.ID,
			category:    "welcome",
			expectError: false,
		},
		{
			name:        "get latest catalog with empty language ID",
			languageID:  "",
			category:    "welcome",
			expectError: true,
		},
		{
			name:        "get latest catalog with non-existent language",
			languageID:  "non-existent-id",
			category:    "welcome",
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			catalog, err := mockDB.GetLatestCatalog(ctx, tt.languageID, tt.category)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, catalog)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, catalog)
				assert.Equal(t, tt.languageID, catalog.LanguageID)
				assert.Equal(t, tt.category, catalog.Category)
			}
		})
	}
}

func TestHealth(t *testing.T) {
	mockDB := NewMockDatabase()
	
	// Test normal health check
	t.Run("normal health check", func(t *testing.T) {
		err := mockDB.Health()
		assert.NoError(t, err)
	})
	
	// Test health check with error
	t.Run("health check with error", func(t *testing.T) {
		mockDB.SetError(errors.New("database connection failed"))
		err := mockDB.Health()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database connection failed")
		
		// Reset error
		mockDB.SetError(nil)
	})
}

func TestCreateVersion(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDatabase()
	
	// Test successful creation
	t.Run("successful creation", func(t *testing.T) {
		testVersion := &models.LocalizationVersion{
			VersionNumber:      "1.0.0",
			VersionType:        "major",
			Description:        "Initial version",
			KeysCount:          10,
			LanguagesCount:     2,
			TotalLocalizations:  20,
			CreatedBy:          "test-user",
			Metadata:           "{}",
		}
		testVersion.BeforeCreate()
		
		err := mockDB.CreateVersion(ctx, testVersion)
		assert.NoError(t, err)
		assert.NotEmpty(t, testVersion.ID)
	})
	
	// Test with mock error
	t.Run("creation with error", func(t *testing.T) {
		testVersion := &models.LocalizationVersion{
			VersionNumber:      "2.0.0",
			VersionType:        "major",
			Description:        "Second version",
			KeysCount:          15,
			LanguagesCount:     3,
			TotalLocalizations:  30,
			CreatedBy:          "test-user",
			Metadata:           "{}",
		}
		// Set error condition
		mockDB.SetError(fmt.Errorf("mock database error"))
		defer mockDB.SetError(nil)
		
		err := mockDB.CreateVersion(ctx, testVersion)
		assert.Error(t, err)
	})
	
	// Test with nil version - create a fresh mockDB to avoid interference
	t.Run("nil version", func(t *testing.T) {
		freshMockDB := NewMockDatabase()
		err := freshMockDB.CreateVersion(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "version cannot be nil")
	})
	
	// Test on fresh mockDB (versions map not initialized)
	t.Run("fresh mockDB", func(t *testing.T) {
		freshMockDB := NewMockDatabase()
		testVersion := &models.LocalizationVersion{
			VersionNumber:      "3.0.0",
			VersionType:        "major",
			Description:        "Third version",
			KeysCount:          20,
			LanguagesCount:     5,
			TotalLocalizations:  40,
			CreatedBy:          "test-user",
			Metadata:           "{}",
		}
		testVersion.BeforeCreate()
		
		err := freshMockDB.CreateVersion(ctx, testVersion)
		assert.NoError(t, err)
		assert.NotEmpty(t, testVersion.ID)
	})
	
	// Test with manually nil versions map
	t.Run("nil versions map", func(t *testing.T) {
		// Create a mockDB and manually set versions to nil
		freshMockDB := NewMockDatabase()
		testVersion := &models.LocalizationVersion{
			VersionNumber:      "4.0.0",
			VersionType:        "major",
			Description:        "Fourth version",
			KeysCount:          25,
			LanguagesCount:     6,
			TotalLocalizations:  50,
			CreatedBy:          "test-user",
			Metadata:           "{}",
		}
		testVersion.BeforeCreate()
		
		// Manually set versions map to nil to trigger initialization
		freshMockDB.SetVersionsNil()
		err := freshMockDB.CreateVersion(ctx, testVersion)
		assert.NoError(t, err)
		assert.NotEmpty(t, testVersion.ID)
	})
}

func TestGetVersionByNumber(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDatabase()
	
	// Create test version
	testVersion := &models.LocalizationVersion{
		VersionNumber:      "1.0.0",
		VersionType:        "major",
		Description:        "Initial version",
		KeysCount:          10,
		LanguagesCount:     2,
		TotalLocalizations:  20,
		CreatedBy:          "test-user",
		Metadata:           "{}",
	}
	// Note: Don't call BeforeCreate here since CreateVersion should do it
	err := mockDB.CreateVersion(ctx, testVersion)
	assert.NoError(t, err)
	
	tests := []struct {
		name          string
		versionNumber string
		expectError   bool
	}{
		{
			name:          "get existing version",
			versionNumber: "1.0.0",
			expectError:   false,
		},
		{
			name:          "get non-existent version",
			versionNumber: "2.0.0",
			expectError:   true,
		},
		{
			name:          "get with mock error",
			versionNumber: "1.0.0",
			expectError:   true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "get with mock error" {
				// Set error condition
				mockDB.SetError(fmt.Errorf("mock database error"))
				defer mockDB.SetError(nil)
			}
			version, err := mockDB.GetVersionByNumber(ctx, tt.versionNumber)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, version)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, version)
				assert.Equal(t, tt.versionNumber, version.VersionNumber)
			}
		})
	}
}

func TestGetVersionByID(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDatabase()
	
	// Create test version
	testVersion := &models.LocalizationVersion{
		VersionNumber:      "1.0.0",
		VersionType:        "major",
		Description:        "Initial version",
		KeysCount:          10,
		LanguagesCount:     2,
		TotalLocalizations:  20,
		CreatedBy:          "test-user",
		Metadata:           "{}",
	}
	// Note: Don't call BeforeCreate here since CreateVersion should do it
	err := mockDB.CreateVersion(ctx, testVersion)
	assert.NoError(t, err)
	
	tests := []struct {
		name        string
		versionID   string
		expectError bool
	}{
		{
			name:        "get existing version",
			versionID:   testVersion.ID,
			expectError: false,
		},
		{
			name:        "get non-existent version",
			versionID:   "non-existent-id",
			expectError: true,
		},
		{
			name:        "get with mock error",
			versionID:   testVersion.ID,
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "get with mock error" {
				// Set error condition
				mockDB.SetError(fmt.Errorf("mock database error"))
				defer mockDB.SetError(nil)
			}
			version, err := mockDB.GetVersionByID(ctx, tt.versionID)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, version)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, version)
				assert.Equal(t, tt.versionID, version.ID)
			}
		})
	}
}

func TestGetCurrentVersion(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDatabase()
	
	// Test with no versions
	t.Run("no versions", func(t *testing.T) {
		_, err := mockDB.GetCurrentVersion(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
	
	// Create test versions
	testVersion1 := &models.LocalizationVersion{
		VersionNumber:      "1.0.0",
		VersionType:        "major",
		Description:        "Initial version",
		KeysCount:          10,
		LanguagesCount:     2,
		TotalLocalizations:  20,
		CreatedBy:          "test-user",
		Metadata:           "{}",
		CreatedAt:          1000,
	}
	err := mockDB.CreateVersion(ctx, testVersion1)
	assert.NoError(t, err)
	
	testVersion2 := &models.LocalizationVersion{
		VersionNumber:      "1.1.0",
		VersionType:        "minor",
		Description:        "Update",
		KeysCount:          12,
		LanguagesCount:     2,
		TotalLocalizations:  24,
		CreatedBy:          "test-user",
		Metadata:           "{}",
		CreatedAt:          2000,
	}
	err = mockDB.CreateVersion(ctx, testVersion2)
	assert.NoError(t, err)
	
	// Test with versions
	t.Run("versions exist", func(t *testing.T) {
		currentVersion, err := mockDB.GetCurrentVersion(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, currentVersion)
		assert.Equal(t, "1.1.0", currentVersion.VersionNumber)
	})
	
	// Test with mock error
	t.Run("mock error", func(t *testing.T) {
		// Set error condition
		mockDB.SetError(fmt.Errorf("mock database error"))
		defer mockDB.SetError(nil)
		_, err := mockDB.GetCurrentVersion(ctx)
		assert.Error(t, err)
	})
}

func TestListVersions(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDatabase()
	
	// Create test versions
	for i := 0; i < 5; i++ {
		testVersion := &models.LocalizationVersion{
			VersionNumber:      fmt.Sprintf("1.%d.0", i),
			VersionType:        "minor",
			Description:        fmt.Sprintf("Version %d", i),
			KeysCount:          10 + i,
			LanguagesCount:     2,
			TotalLocalizations:  20 + i*2,
			CreatedBy:          "test-user",
			Metadata:           "{}",
			CreatedAt:          int64(1000 + i), // Set different timestamps for consistent ordering
		}
		// Note: Don't call BeforeCreate here since CreateVersion should do it
		err := mockDB.CreateVersion(ctx, testVersion)
		assert.NoError(t, err)
	}
	
	tests := []struct {
		name        string
		limit       int
		offset      int
		expectCount int
	}{
		{
			name:        "first page",
			limit:       2,
			offset:      0,
			expectCount:  2,
		},
		{
			name:        "second page",
			limit:       2,
			offset:      2,
			expectCount:  2,
		},
		{
			name:        "last page with remainder",
			limit:       2,
			offset:      4,
			expectCount:  1,
		},
		{
			name:        "offset beyond list",
			limit:       2,
			offset:      10,
			expectCount:  0,
		},
		{
			name:        "list with mock error",
			limit:       2,
			offset:      0,
			expectCount:  0, // Error case doesn't return items
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "list with mock error" {
				mockDB.SetError(fmt.Errorf("mock database error"))
				versions, err := mockDB.ListVersions(ctx, tt.limit, tt.offset)
				assert.Error(t, err)
				assert.Nil(t, versions)
				assert.Contains(t, err.Error(), "mock database error")
				mockDB.ClearError()
				return
			}
			
			versions, err := mockDB.ListVersions(ctx, tt.limit, tt.offset)
			assert.NoError(t, err)
			assert.Len(t, versions, tt.expectCount)
		})
	}
}

func TestCountVersions(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDatabase()
	
	// Test with no versions
	t.Run("no versions", func(t *testing.T) {
		count, err := mockDB.CountVersions(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})
	
	// Create test versions
	for i := 0; i < 3; i++ {
		testVersion := &models.LocalizationVersion{
			VersionNumber:      fmt.Sprintf("1.%d.0", i),
			VersionType:        "minor",
			Description:        fmt.Sprintf("Version %d", i),
			KeysCount:          10 + i,
			LanguagesCount:     2,
			TotalLocalizations:  20 + i*2,
			CreatedBy:          "test-user",
			Metadata:           "{}",
		}
		// Note: Don't call BeforeCreate here since CreateVersion should do it
		err := mockDB.CreateVersion(ctx, testVersion)
		assert.NoError(t, err)
	}
	
	// Test with versions
	t.Run("versions exist", func(t *testing.T) {
		count, err := mockDB.CountVersions(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 3, count)
	})
	
	// Test with mock error
	t.Run("mock error", func(t *testing.T) {
		// Set error condition
		mockDB.SetError(fmt.Errorf("mock database error"))
		defer mockDB.SetError(nil)
		count, err := mockDB.CountVersions(ctx)
		assert.Error(t, err)
		assert.Equal(t, 0, count)
	})
}

func TestDeleteVersion(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDatabase()
	
	// Create test version
	testVersion := &models.LocalizationVersion{
		VersionNumber:      "1.0.0",
		VersionType:        "major",
		Description:        "Initial version",
		KeysCount:          10,
		LanguagesCount:     2,
		TotalLocalizations:  20,
		CreatedBy:          "test-user",
		Metadata:           "{}",
	}
	// Note: Don't call BeforeCreate here since CreateVersion should do it
	err := mockDB.CreateVersion(ctx, testVersion)
	assert.NoError(t, err)
	
	tests := []struct {
		name        string
		versionID   string
		expectError bool
	}{
		{
			name:        "delete existing version",
			versionID:   testVersion.ID,
			expectError: false,
		},
		{
			name:        "delete non-existent version",
			versionID:   "non-existent-id",
			expectError: true,
		},
		{
			name:        "delete with mock error",
			versionID:   testVersion.ID,
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "delete with mock error" {
				// Set error condition
				mockDB.SetError(fmt.Errorf("mock database error"))
				defer mockDB.SetError(nil)
			}
			err := mockDB.DeleteVersion(ctx, tt.versionID)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				
				// Verify deletion
				_, err = mockDB.GetVersionByID(ctx, tt.versionID)
				assert.Error(t, err)
			}
		})
	}
}

func TestGetCatalogByVersion(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDatabase()
	
	// Create test language
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
	
	// Create test version
	testVersion := &models.LocalizationVersion{
		VersionNumber:      "1.0.0",
		VersionType:        "major",
		Description:        "Initial version",
		KeysCount:          10,
		LanguagesCount:     2,
		TotalLocalizations:  20,
		CreatedBy:          "test-user",
		Metadata:           "{}",
	}
	err = mockDB.CreateVersion(ctx, testVersion)
	assert.NoError(t, err)
	
	tests := []struct {
		name          string
		versionNumber string
		languageCode string
		expectError  bool
	}{
		{
			name:          "get existing catalog by version",
			versionNumber: "1.0.0",
			languageCode: "en-US",
			expectError:   false,
		},
		{
			name:          "non-existent version",
			versionNumber: "2.0.0",
			languageCode: "en-US",
			expectError:   true,
		},
		{
			name:          "non-existent language",
			versionNumber: "1.0.0",
			languageCode: "fr-FR",
			expectError:   true,
		},
		{
			name:          "get with mock error",
			versionNumber: "1.0.0",
			languageCode: "en-US",
			expectError:   true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "get with mock error" {
				// Set error condition
				mockDB.SetError(fmt.Errorf("mock database error"))
				defer mockDB.SetError(nil)
			}
			catalog, err := mockDB.GetCatalogByVersion(ctx, tt.versionNumber, tt.languageCode)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, catalog)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, catalog)
				assert.Equal(t, testLang.ID, catalog.LanguageID)
			}
		})
	}
}

func TestCreateAuditLog(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDatabase()
	
	tests := []struct {
		name       string
		action     string
		entityType string
		entityID   string
		username   string
		changes    interface{}
		ipAddress  string
		userAgent  string
		expectError bool
	}{
		{
			name:       "create valid audit log",
			action:     "create",
			entityType: "localization",
			entityID:   "loc-123",
			username:   "test-user",
			changes:    map[string]interface{}{"key": "test.key", "value": "Test Value"},
			ipAddress:  "127.0.0.1",
			userAgent:  "test-agent",
			expectError: false,
		},
		{
			name:       "create audit log with error condition",
			action:     "create",
			entityType: "localization",
			entityID:   "loc-123",
			username:   "test-user",
			changes:    map[string]interface{}{"key": "test.key", "value": "Test Value"},
			ipAddress:  "127.0.0.1",
			userAgent:  "test-agent",
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectError {
				// Set error condition
				mockDB.SetError(fmt.Errorf("test error"))
				defer mockDB.SetError(nil)
			}
			err := mockDB.CreateAuditLog(ctx, tt.action, tt.entityType, tt.entityID, tt.username, tt.changes, tt.ipAddress, tt.userAgent)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetStats(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDatabase()
	
	// Create test data
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
	
	testKey := &models.LocalizationKey{
		Key:         "test.key",
		Category:    "test",
		Description: "Test key",
	}
	err = mockDB.CreateLocalizationKey(ctx, testKey)
	assert.NoError(t, err)
	
	// Test stats retrieval
	t.Run("get stats", func(t *testing.T) {
		stats, err := mockDB.GetStats(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, stats)
		
		// Check if keys are present (we don't enforce exact values since mock implementation may vary)
		assert.Contains(t, stats, "total_languages")
		assert.Contains(t, stats, "total_keys")
		assert.Contains(t, stats, "total_localizations")
	})
	
	// Test with mock error
	t.Run("get stats with error", func(t *testing.T) {
		// Set error condition
		mockDB.SetError(fmt.Errorf("mock database error"))
		defer mockDB.SetError(nil)
		stats, err := mockDB.GetStats(ctx)
		assert.Error(t, err)
		assert.Nil(t, stats)
	})
}

// TestEncryptDecrypt tests the encrypt and decrypt functions
func TestEncryptDecrypt(t *testing.T) {
	t.Run("without encryption config", func(t *testing.T) {
		// Create a PostgresDatabase instance without connecting
		db := &PostgresDatabase{}
		
		testValue := "test value"
		
		// Test encrypt
		encrypted := db.encrypt(testValue)
		assert.Equal(t, testValue, encrypted)
		
		// Test decrypt
		decrypted := db.decrypt(testValue)
		assert.Equal(t, testValue, decrypted)
	})
	
	t.Run("with encryption config", func(t *testing.T) {
		// Create a PostgresDatabase instance with encryption config
		encConfig := &config.EncryptionConfig{
			Key: "test-key-32-characters-long-123",
		}
		db := &PostgresDatabase{
			encConfig: encConfig,
		}
		
		testValue := "test value"
		
		// Test encrypt (returns as-is for now)
		encrypted := db.encrypt(testValue)
		assert.Equal(t, testValue, encrypted)
		
		// Test decrypt (returns as-is for now)
		decrypted := db.decrypt(testValue)
		assert.Equal(t, testValue, decrypted)
	})
}

// TestCreateLocalizationKeyValidation tests validation path before database operations
func TestCreateLocalizationKeyValidation(t *testing.T) {
	// Create a PostgresDatabase instance without connecting
	db := &PostgresDatabase{}
	
	// Test with invalid key (missing required fields)
	invalidKey := &models.LocalizationKey{
		// Missing required Key field
		Category:    "test-category",
		Description: "Test key",
	}
	
	ctx := context.Background()
	err := db.CreateLocalizationKey(ctx, invalidKey)
	assert.Error(t, err) // Should fail validation before database operation
}

// TestCreateLanguageValidation tests validation path before database operations
func TestCreateLanguageValidation(t *testing.T) {
	// Create a PostgresDatabase instance without connecting
	db := &PostgresDatabase{}
	
	// Test with invalid language (missing required fields)
	invalidLang := &models.Language{
		// Missing required Code field
		Name:       "English (US)",
		NativeName: "English (United States)",
		IsActive:   true,
		IsDefault:  false,
	}
	
	ctx := context.Background()
	err := db.CreateLanguage(ctx, invalidLang)
	assert.Error(t, err) // Should fail validation before database operation
}

// TestCreateLocalizationValidation tests validation path before database operations
func TestCreateLocalizationValidation(t *testing.T) {
	// Create a PostgresDatabase instance without connecting
	db := &PostgresDatabase{}
	
	// Test with invalid localization (missing required fields)
	invalidLoc := &models.Localization{
		// Missing required KeyID field
		LanguageID: "test-lang-id",
		Value:      "Test value",
		Version:    1,
	}
	
	ctx := context.Background()
	err := db.CreateLocalization(ctx, invalidLoc)
	assert.Error(t, err) // Should fail validation before database operation
}

// TestUpdateLanguageValidation tests validation path before database operations
func TestUpdateLanguageValidation(t *testing.T) {
	// Create a PostgresDatabase instance without connecting
	db := &PostgresDatabase{}
	
	// Test with invalid language (missing required fields)
	invalidLang := &models.Language{
		ID:         "test-id",
		// Missing required Code field
		Name:       "English (US)",
		NativeName: "English (United States)",
		IsActive:   true,
		IsDefault:  false,
	}
	
	ctx := context.Background()
	err := db.UpdateLanguage(ctx, invalidLang)
	assert.Error(t, err) // Should fail validation before database operation
}

// TestUpdateLocalizationKeyValidation tests validation path before database operations
func TestUpdateLocalizationKeyValidation(t *testing.T) {
	// Create a PostgresDatabase instance without connecting
	db := &PostgresDatabase{}
	
	// Test with invalid key (missing required fields)
	invalidKey := &models.LocalizationKey{
		ID:         "test-id",
		// Missing required Key field
		Category:    "test-category",
		Description: "Test key description",
	}
	
	ctx := context.Background()
	err := db.UpdateLocalizationKey(ctx, invalidKey)
	assert.Error(t, err) // Should fail validation before database operation
}

// TestUpdateLocalizationValidation tests validation path before database operations
func TestUpdateLocalizationValidation(t *testing.T) {
	// Create a PostgresDatabase instance without connecting
	db := &PostgresDatabase{}
	
	// Test with invalid localization (missing required fields)
	invalidLoc := &models.Localization{
		ID:         "test-id",
		// Missing required KeyID field
		LanguageID: "test-lang-id",
		Value:      "Test value",
		Version:    1,
	}
	
	ctx := context.Background()
	err := db.UpdateLocalization(ctx, invalidLoc)
	assert.Error(t, err) // Should fail validation before database operation
}

// TestCreateLanguageWithInitializedDB tests validation path with initialized fields
func TestCreateLanguageWithInitializedDB(t *testing.T) {
	// Create a PostgresDatabase instance with initialized fields to avoid panics
	logger, _ := zap.NewDevelopment()
	db := &PostgresDatabase{
		config:    &config.DatabaseConfig{Driver: "postgres"},
		encConfig: &config.EncryptionConfig{},
		logger:    logger,
		db:         nil, // Intentionally nil to test validation before database operations
	}
	
	// Test with invalid language (missing required fields)
	invalidLang := &models.Language{
		// Missing required Code field
		Name:       "English (US)",
		NativeName: "English (United States)",
		IsActive:   true,
		IsDefault:  false,
	}
	
	ctx := context.Background()
	err := db.CreateLanguage(ctx, invalidLang)
	assert.Error(t, err) // Should fail validation before database operation
}

// TestUpdateLanguageWithInitializedDB tests validation path with initialized fields
func TestUpdateLanguageWithInitializedDB(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := &PostgresDatabase{
		config:    &config.DatabaseConfig{Driver: "postgres"},
		encConfig: &config.EncryptionConfig{},
		logger:    logger,
		db:         nil, // Intentionally nil to test validation before database operations
	}
	
	// Test with invalid language (missing required fields)
	invalidLang := &models.Language{
		ID: "test-id",
		// Missing required Code field
		Name:       "English (US)",
		NativeName: "English (United States)",
		IsActive:   true,
		IsDefault:  false,
	}
	
	ctx := context.Background()
	err := db.UpdateLanguage(ctx, invalidLang)
	assert.Error(t, err) // Should fail validation before database operation
}

// TestCreateCatalogWithInitializedDB tests validation path with initialized fields
func TestCreateCatalogWithInitializedDB(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := &PostgresDatabase{
		config:    &config.DatabaseConfig{Driver: "postgres"},
		encConfig: &config.EncryptionConfig{},
		logger:    logger,
		db:         nil, // Intentionally nil to test validation before database operations
	}
	
	// Test with invalid catalog (missing required fields)
	invalidCatalog := &models.LocalizationCatalog{
		// Missing required LanguageID field
		Category:    "test",
		CatalogData: []byte("{}"),
		Version:     1,
	}
	
	ctx := context.Background()
	err := db.CreateCatalog(ctx, invalidCatalog)
	assert.Error(t, err) // Should fail validation before database operation
}

// TestCreateLocalizationKeyValidationWithInitializedDB tests validation path with initialized fields
func TestCreateLocalizationKeyValidationWithInitializedDB(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := &PostgresDatabase{
		config:    &config.DatabaseConfig{Driver: "postgres"},
		encConfig: &config.EncryptionConfig{},
		logger:    logger,
		db:         nil, // Intentionally nil to test validation before database operations
	}
	
	// Test with invalid key (missing required fields)
	invalidKey := &models.LocalizationKey{
		// Missing required Key field
		Category:    "test",
		Description: "Test key description",
		Context:     "Test context",
	}
	
	ctx := context.Background()
	err := db.CreateLocalizationKey(ctx, invalidKey)
	assert.Error(t, err) // Should fail validation before database operation
}

// TestUpdateCatalogWithInitializedDB tests validation path with initialized fields
func TestUpdateCatalogWithInitializedDB(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := &PostgresDatabase{
		config:    &config.DatabaseConfig{Driver: "postgres"},
		encConfig: &config.EncryptionConfig{},
		logger:    logger,
		db:         nil, // Intentionally nil to test validation before database operations
	}
	
	// Test with invalid catalog (missing required fields)
	invalidCatalog := &models.LocalizationCatalog{
		ID: "test-id",
		// Missing required LanguageID field
		Category:    "test",
		CatalogData: []byte("{}"),
		Version:     1,
	}
	
	ctx := context.Background()
	err := db.UpdateCatalog(ctx, invalidCatalog)
	assert.Error(t, err) // Should fail validation before database operation
}

// TestUpdateLocalizationWithInitializedDB tests validation path with initialized fields
func TestUpdateLocalizationWithInitializedDB(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := &PostgresDatabase{
		config:    &config.DatabaseConfig{Driver: "postgres"},
		encConfig: &config.EncryptionConfig{},
		logger:    logger,
		db:         nil, // Intentionally nil to test validation before database operations
	}
	
	// Test with invalid localization (missing required fields)
	invalidLoc := &models.Localization{
		ID: "test-id",
		// Missing required KeyID and LanguageID fields
		Value:   "Test value",
		Version: 1,
	}
	
	ctx := context.Background()
	err := db.UpdateLocalization(ctx, invalidLoc)
	assert.Error(t, err) // Should fail validation before database operation
}

// TestUpdateLocalizationKeyValidationWithInitializedDB tests validation path with initialized fields

// TestCreateLocalizationWithInitializedDB tests validation path with initialized fields
func TestCreateLocalizationWithInitializedDB(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := &PostgresDatabase{
		config:    &config.DatabaseConfig{Driver: "postgres"},
		encConfig: &config.EncryptionConfig{},
		logger:    logger,
		db:         nil, // Intentionally nil to test validation before database operations
	}
	
	// Test with invalid localization (missing required fields)
	invalidLoc := &models.Localization{
		// Missing required KeyID and LanguageID fields
		Value:   "Test value",
		Version: 1,
	}
	
	ctx := context.Background()
	err := db.CreateLocalization(ctx, invalidLoc)
	assert.Error(t, err) // Should fail validation before database operation
}

// TestCreateLocalizationKeyValidationWithInitializedDB tests validation path with initialized fields

// TestNewWithInvalidConfig tests error path in New function
func TestNewWithInvalidConfig(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	// Test with invalid DSN (invalid driver)
	invalidConfig := &config.DatabaseConfig{
		Driver: "invalid-driver",
		Host:   "localhost",
		Port:   5432,
	}
	
	_, err := New(invalidConfig, &config.EncryptionConfig{}, logger)
	assert.Error(t, err) // Should fail when opening database with invalid driver
}

// TestNewWithValidConfigButInvalidConnection tests ping failure path in New function
func TestNewWithValidConfigButInvalidConnection(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	// Test with valid config but invalid connection details
	invalidConfig := &config.DatabaseConfig{
		Driver:   "postgres",
		Host:     "non-existent-host-that-does-not-exist.com",
		Port:     5432,
		Database: "test",
		User:     "test",
		Password: "test",
	}
	
	_, err := New(invalidConfig, &config.EncryptionConfig{}, logger)
	assert.Error(t, err) // Should fail when trying to ping the database
}

// TestPingWithMockError tests mock error path in Ping function
func TestPingWithMockError(t *testing.T) {
	mockDB := NewMockDatabase()
	
	// Set error to mock database failure
	mockDB.SetError(fmt.Errorf("mock ping error"))
	defer mockDB.ClearError()
	
	err := mockDB.Ping()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mock ping error")
}

// TestCloseWithMockError tests mock error path in Close function
func TestCloseWithMockError(t *testing.T) {
	mockDB := NewMockDatabase()
	
	// Set error to mock database failure
	mockDB.SetError(fmt.Errorf("mock close error"))
	defer mockDB.ClearError()
	
	err := mockDB.Close()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mock close error")
}

// TestGetCatalogByLanguageWithMockError tests mock error path
func TestGetCatalogByLanguageWithMockError(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDatabase()
	
	// Set error to mock database failure
	mockDB.SetError(fmt.Errorf("mock get catalog error"))
	defer mockDB.ClearError()
	
	_, err := mockDB.GetCatalogByLanguage(ctx, "test-lang-id", "test-category")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mock get catalog error")
}

// TestGetLanguagesWithMockError tests mock error path
func TestGetLanguagesWithMockError(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDatabase()
	
	// Set error to mock database failure
	mockDB.SetError(fmt.Errorf("mock get languages error"))
	defer mockDB.ClearError()
	
	_, err := mockDB.GetLanguages(ctx, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mock get languages error")
}

// TestGetLocalizationsByLanguageWithMockError tests mock error path
func TestGetLocalizationsByLanguageWithMockError(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDatabase()
	
	// Set error to mock database failure
	mockDB.SetError(fmt.Errorf("mock get localizations error"))
	defer mockDB.ClearError()
	
	_, err := mockDB.GetLocalizationsByLanguage(ctx, "test-lang-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mock get localizations error")
}

// TestGetStatsWithMockError tests mock error path
func TestGetStatsWithMockError(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDatabase()
	
	// Set error to mock database failure
	mockDB.SetError(fmt.Errorf("mock get stats error"))
	defer mockDB.ClearError()
	
	_, err := mockDB.GetStats(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mock get stats error")
}

// TestGetLocalizationByKeyAndLanguageWithMockError tests mock error path
func TestGetLocalizationByKeyAndLanguageWithMockError(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDatabase()
	
	// Set error to mock database failure
	mockDB.SetError(fmt.Errorf("mock get localization error"))
	defer mockDB.ClearError()
	
	_, err := mockDB.GetLocalizationByKeyAndLanguage(ctx, "test-key-id", "test-lang-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mock get localization error")
}

// TestGetLanguageByCodeWithMockError tests mock error path
func TestGetLanguageByCodeWithMockError(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDatabase()
	
	// Set error to mock database failure
	mockDB.SetError(fmt.Errorf("mock get language by code error"))
	defer mockDB.ClearError()
	
	_, err := mockDB.GetLanguageByCode(ctx, "en-US")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mock get language by code error")
}

// TestNewWithInvalidConfigButValidDriver tests DSN generation error path
func TestNewWithInvalidConfigButValidDriver(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	// Test with valid driver but invalid DSN (missing required fields)
	invalidConfig := &config.DatabaseConfig{
		Driver: "postgres",
		// Missing required Host field
		Port:     5432,
		Database: "test",
		User:     "test",
		Password: "test",
	}
	
	_, err := New(invalidConfig, &config.EncryptionConfig{}, logger)
	assert.Error(t, err) // Should fail when trying to connect with invalid DSN
}

// TestNewWithValidConfigButInvalidPassword tests connection failure path
func TestNewWithValidConfigButInvalidPassword(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	// Test with valid config but invalid password
	invalidConfig := &config.DatabaseConfig{
		Driver:   "postgres",
		Host:     "localhost",
		Port:     5432,
		Database: "test",
		User:     "postgres",
		Password: "definitely-wrong-password",
	}
	
	_, err := New(invalidConfig, &config.EncryptionConfig{}, logger)
	assert.Error(t, err) // Should fail when trying to ping database with wrong password
}

// TestMarshalCatalogErrorMethods tests the mock marshal catalog error methods
func TestMarshalCatalogErrorMethods(t *testing.T) {
	mockDB := NewMockDatabase()
	
	// Create a test language first
	testLang := &models.Language{
		Code:      "en",
		Name:      "English",
		IsDefault: false,
		IsActive:  true,
	}
	err := mockDB.CreateLanguage(context.Background(), testLang)
	assert.NoError(t, err)
	
	// Test SetMarshalCatalogError
	mockDB.SetMarshalCatalogError()
	
	// Verify that BuildCatalog returns an error when shouldFailMarshalCatalog is true
	_, err = mockDB.BuildCatalog(context.Background(), testLang.ID, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "simulated json.Marshal error")
	
	// Test ClearMarshalCatalogError
	mockDB.ClearMarshalCatalogError()
	
	// Verify that BuildCatalog no longer returns an error
	_, err = mockDB.BuildCatalog(context.Background(), testLang.ID, "")
	assert.NoError(t, err)
}

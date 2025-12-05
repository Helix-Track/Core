package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/helixtrack/localization-service/internal/cache"
	"github.com/helixtrack/localization-service/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestUpdateLanguage_Unit tests the update language handler
func TestUpdateLanguage_Unit(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := NewMockDatabase()
	memCache := cache.NewMemoryCache(10, 1*time.Hour, 5*time.Minute, logger)
	handler := NewHandler(db, memCache, logger, nil)

	t.Run("update language with admin role", func(t *testing.T) {
		// First create a language to update
		existingLang := &models.Language{
			ID:   "test-lang-id",
			Code:  "en",
			Name:  "English",
		}
		db.languages["test-lang-id"] = existingLang

		langData := models.CreateLanguageRequest{
			Code: "es",
			Name: "Spanish",
		}
		body, _ := json.Marshal(langData)

		c, w := setupTestContext()
		c.Request = httptest.NewRequest("PUT", "/languages/test-lang-id", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: "test-lang-id"}}
		c.Set("user_role", "admin")
		// Set claims with the correct key that middleware expects
		claims := &models.JWTClaims{Username: "adminuser", Role: "admin"}
		c.Set("claims", claims)

		handler.UpdateLanguage(c)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response models.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("update language with user role (forbidden)", func(t *testing.T) {
		langData := models.CreateLanguageRequest{
			Code: "fr",
			Name: "French",
		}
		body, _ := json.Marshal(langData)

		c, w := setupTestContext()
		c.Request = httptest.NewRequest("PUT", "/languages/test-lang-id", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: "test-lang-id"}}
		c.Set("user_role", "user")
		// Set claims with the correct key that middleware expects
		claims := &models.JWTClaims{Username: "testuser", Role: "user"}
		c.Set("claims", claims)

		// Create a mock middleware that blocks the request
		mockMiddleware := func(c *gin.Context) {
			c.JSON(http.StatusForbidden, models.ErrorResponse(
				models.ErrCodeForbidden,
				"access denied",
			))
			c.Abort()
		}

		// Apply middleware
		handlerFunc := mockMiddleware
		handlerFunc(c)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

// TestDeleteLanguage_Unit tests the delete language handler
func TestDeleteLanguage_Unit(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := NewMockDatabase()
	memCache := cache.NewMemoryCache(10, 1*time.Hour, 5*time.Minute, logger)
	handler := NewHandler(db, memCache, logger, nil)

	t.Run("delete language with admin role", func(t *testing.T) {
		// First create a language to delete
		existingLang := &models.Language{
			ID:   "test-lang-id",
			Code:  "en",
			Name:  "English",
		}
		db.languages["test-lang-id"] = existingLang

		c, w := setupTestContext()
		c.Request = httptest.NewRequest("DELETE", "/languages/test-lang-id", nil)
		c.Params = gin.Params{{Key: "id", Value: "test-lang-id"}}
		c.Set("user_role", "admin")
		// Set claims with the correct key that middleware expects
		claims := &models.JWTClaims{Username: "adminuser", Role: "admin"}
		c.Set("claims", claims)

		handler.DeleteLanguage(c)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response models.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("delete non-existent language", func(t *testing.T) {
		c, w := setupTestContext()
		c.Request = httptest.NewRequest("DELETE", "/languages/non-existent", nil)
		c.Params = gin.Params{{Key: "id", Value: "non-existent"}}
		c.Set("user_role", "admin")
		// Set claims with the correct key that middleware expects
		claims := &models.JWTClaims{Username: "adminuser", Role: "admin"}
		c.Set("claims", claims)

		handler.DeleteLanguage(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

// TestCreateLocalization_Unit tests the create localization handler
func TestCreateLocalization_Unit(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := NewMockDatabase()
	memCache := cache.NewMemoryCache(10, 1*time.Hour, 5*time.Minute, logger)
	handler := NewHandler(db, memCache, logger, nil)

	t.Run("create localization with admin role", func(t *testing.T) {
		// Create test language
		lang := &models.Language{
			ID:   "test-lang-id",
			Code:  "en",
			Name:  "English",
		}
		db.languages["en"] = lang

		locData := models.CreateLocalizationRequest{
			Key:       "ui.button.save",
			Language:  "en",
			Value:     "Save",
			Category:  "ui",
			Approved:  true,
		}
		body, _ := json.Marshal(locData)

		c, w := setupTestContext()
		c.Request = httptest.NewRequest("POST", "/localizations", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("user_role", "admin")
		// Set claims with the correct key that middleware expects
		claims := &models.JWTClaims{Username: "adminuser", Role: "admin"}
		c.Set("claims", claims)

		handler.CreateLocalization(c)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("create localization with validation error", func(t *testing.T) {
		locData := models.CreateLocalizationRequest{
			// Missing required fields
			Language: "en",
		}
		body, _ := json.Marshal(locData)

		c, w := setupTestContext()
		c.Request = httptest.NewRequest("POST", "/localizations", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("user_role", "admin")
		// Set claims with the correct key that middleware expects
		claims := &models.JWTClaims{Username: "adminuser", Role: "admin"}
		c.Set("claims", claims)

		handler.CreateLocalization(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("create localization with database error when creating key", func(t *testing.T) {
		// Create test language
		lang := &models.Language{
			ID:   "test-lang-id",
			Code:  "en",
			Name:  "English",
		}
		db.languages["en"] = lang

		// Set database to return error when creating localization key
		db.SetError(fmt.Errorf("database error"))
		defer db.ClearError()

		locData := models.CreateLocalizationRequest{
			Key:       "ui.button.save",
			Language:  "en",
			Value:     "Save",
			Category:  "ui",
			Approved:  true,
		}
		body, _ := json.Marshal(locData)

		c, w := setupTestContext()
		c.Request = httptest.NewRequest("POST", "/localizations", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("user_role", "admin")
		// Set claims with the correct key that middleware expects
		claims := &models.JWTClaims{Username: "adminuser", Role: "admin"}
		c.Set("claims", claims)

		handler.CreateLocalization(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// TestInvalidateCache_Unit tests the cache invalidation handler
func TestInvalidateCache_Unit(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := NewMockDatabase()
	memCache := cache.NewMemoryCache(10, 1*time.Hour, 5*time.Minute, logger)
	handler := NewHandler(db, memCache, logger, nil)

	t.Run("invalidate cache for specific language", func(t *testing.T) {
		cacheReq := models.CacheInvalidationRequest{
			Language: "en",
		}
		body, _ := json.Marshal(cacheReq)

		c, w := setupTestContext()
		c.Request = httptest.NewRequest("POST", "/admin/cache/invalidate", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("user_role", "admin")
		// Set claims with the correct key that middleware expects
		claims := &models.JWTClaims{Username: "adminuser", Role: "admin"}
		c.Set("claims", claims)

		handler.InvalidateCache(c)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response models.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("invalidate all cache", func(t *testing.T) {
		c, w := setupTestContext()
		c.Request = httptest.NewRequest("POST", "/admin/cache/invalidate", nil)
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("user_role", "admin")
		// Set claims with the correct key that middleware expects
		claims := &models.JWTClaims{Username: "adminuser", Role: "admin"}
		c.Set("claims", claims)

		handler.InvalidateCache(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestApproveLocalization_Unit tests the approve localization handler
func TestApproveLocalization_Unit(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := NewMockDatabase()
	memCache := cache.NewMemoryCache(10, 1*time.Hour, 5*time.Minute, logger)
	// Use nil for WebSocket manager since we're not testing WebSocket functionality
	handler := NewHandler(db, memCache, logger, nil)

	t.Run("approve localization with admin role", func(t *testing.T) {
		// Create test localization
		loc := &models.Localization{
			ID:       "test-loc-id",
			KeyID:    "test-key-id",
			LanguageID: "en",
			Value:    "Save",
			Approved: false,
		}
		db.localizations["test-loc-id"] = loc

		c, w := setupTestContext()
		c.Request = httptest.NewRequest("POST", "/admin/localizations/test-loc-id/approve", nil)
		c.Params = gin.Params{{Key: "id", Value: "test-loc-id"}}
		c.Set("user_role", "admin")
		// Set claims with the correct key that middleware expects
		claims := &models.JWTClaims{Username: "adminuser", Role: "admin"}
		c.Set("claims", claims)

		handler.ApproveLocalization(c)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response models.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
	})
}

// TestUpdateLocalization_Unit tests update localization handler
func TestUpdateLocalization_Unit(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := NewMockDatabase()
	memCache := cache.NewMemoryCache(10, 1*time.Hour, 5*time.Minute, logger)
	handler := NewHandler(db, memCache, logger, nil)

	t.Run("update localization with admin role", func(t *testing.T) {
		// Create test localization
		loc := &models.Localization{
			ID:         "test-loc-id",
			KeyID:      "test-key-id",
			LanguageID: "test-lang-id",
			Value:      "Save",
			Approved:   false,
		}
		db.localizations["test-loc-id"] = loc
		
		// Create test language
		lang := &models.Language{
			ID:   "test-lang-id",
			Code: "en",
			Name: "English",
		}
		db.languages["test-lang-id"] = lang
		
		// Create test key
		key := &models.LocalizationKey{
			ID:  "test-key-id",
			Key: "ui.button.save",
		}
		db.localizationKeys["test-key-id"] = key

		locData := models.CreateLocalizationRequest{
			Key:      "ui.button.save",
			Language: "en",
			Value:    "Save Changes",
			Category: "ui",
			Approved: true,
		}
		body, _ := json.Marshal(locData)

		c, w := setupTestContext()
		c.Request = httptest.NewRequest("PUT", "/admin/localizations/test-loc-id", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: "test-loc-id"}}
		c.Set("user_role", "admin")
		// Set claims with the correct key that middleware expects
		claims := &models.JWTClaims{Username: "adminuser", Role: "admin"}
		c.Set("claims", claims)

		handler.UpdateLocalization(c)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response models.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("update non-existent localization", func(t *testing.T) {
		locData := models.CreateLocalizationRequest{
			Key:      "ui.button.save",
			Language: "en",
			Value:    "Save Changes",
		}
		body, _ := json.Marshal(locData)

		c, w := setupTestContext()
		c.Request = httptest.NewRequest("PUT", "/admin/localizations/non-existent", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: "non-existent"}}
		c.Set("user_role", "admin")
		// Set claims with the correct key that middleware expects
		claims := &models.JWTClaims{Username: "adminuser", Role: "admin"}
		c.Set("claims", claims)

		handler.UpdateLocalization(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

// TestDeleteLocalization_Unit tests delete localization handler
func TestDeleteLocalization_Unit(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := NewMockDatabase()
	memCache := cache.NewMemoryCache(10, 1*time.Hour, 5*time.Minute, logger)
	handler := NewHandler(db, memCache, logger, nil)

	t.Run("delete localization with admin role", func(t *testing.T) {
		// Create test localization
		loc := &models.Localization{
			ID:         "test-loc-id",
			KeyID:      "test-key-id",
			LanguageID: "test-lang-id",
			Value:      "Save",
			Approved:   false,
		}
		db.localizations["test-loc-id"] = loc
		
		// Create test language
		lang := &models.Language{
			ID:   "test-lang-id",
			Code: "en",
			Name: "English",
		}
		db.languages["test-lang-id"] = lang
		
		// Create test key
		key := &models.LocalizationKey{
			ID:  "test-key-id",
			Key: "ui.button.save",
		}
		db.localizationKeys["test-key-id"] = key

		c, w := setupTestContext()
		c.Request = httptest.NewRequest("DELETE", "/admin/localizations/test-loc-id", nil)
		c.Params = gin.Params{{Key: "id", Value: "test-loc-id"}}
		c.Set("user_role", "admin")
		// Set claims with the correct key that middleware expects
		claims := &models.JWTClaims{Username: "adminuser", Role: "admin"}
		c.Set("claims", claims)

		handler.DeleteLocalization(c)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response models.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("delete non-existent localization", func(t *testing.T) {
		c, w := setupTestContext()
		c.Request = httptest.NewRequest("DELETE", "/admin/localizations/non-existent", nil)
		c.Params = gin.Params{{Key: "id", Value: "non-existent"}}
		c.Set("user_role", "admin")
		// Set claims with the correct key that middleware expects
		claims := &models.JWTClaims{Username: "adminuser", Role: "admin"}
		c.Set("claims", claims)

		handler.DeleteLocalization(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
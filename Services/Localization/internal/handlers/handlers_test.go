package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/helixtrack/localization-service/internal/cache"
	"github.com/helixtrack/localization-service/internal/middleware"
	"github.com/helixtrack/localization-service/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// Helper functions for testing
func setupTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	return c, w
}

// Test Handler creation
func TestNewHandler(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	// Use the MockDatabase from integration_test.go
	db := NewMockDatabase()
	memCache := cache.NewMemoryCache(10, 1*time.Hour, 5*time.Minute, logger)

	handler := NewHandler(db, memCache, logger, nil)
	
	assert.NotNil(t, handler)
	assert.Equal(t, db, handler.db)
	assert.Equal(t, memCache, handler.cache)
	assert.Equal(t, logger, handler.logger)
}

// Test Health Check handler
func TestHealthCheck_Unit(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := NewMockDatabase()
	memCache := cache.NewMemoryCache(10, 1*time.Hour, 5*time.Minute, logger)
	handler := NewHandler(db, memCache, logger, nil)

	c, w := setupTestContext()
	
	handler.HealthCheck(c)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response models.HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "healthy", response.Status)
}

// Test GetCatalog with real data
func TestGetCatalog_Unit(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := NewMockDatabase()
	memCache := cache.NewMemoryCache(10, 1*time.Hour, 5*time.Minute, logger)
	handler := NewHandler(db, memCache, logger, nil)

	t.Run("success", func(t *testing.T) {
		c, w := setupTestContext()
		c.Params = gin.Params{gin.Param{"language", "en"}}
		
		handler.GetCatalog(c)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response models.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("language not found", func(t *testing.T) {
		c, w := setupTestContext()
		c.Params = gin.Params{gin.Param{"language", "xx"}}
		
		handler.GetCatalog(c)
		
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

// Test GetLocalization with real data
func TestGetLocalization_Unit(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := NewMockDatabase()
	memCache := cache.NewMemoryCache(10, 1*time.Hour, 5*time.Minute, logger)
	handler := NewHandler(db, memCache, logger, nil)

	t.Run("success", func(t *testing.T) {
		c, w := setupTestContext()
		c.Params = gin.Params{gin.Param{"key", "app.welcome"}}
		c.Request.URL.RawQuery = "language=en"
		
		handler.GetLocalization(c)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response models.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("missing language parameter", func(t *testing.T) {
		c, w := setupTestContext()
		c.Params = gin.Params{gin.Param{"key", "app.welcome"}}
		
		handler.GetLocalization(c)
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("language not found", func(t *testing.T) {
		c, w := setupTestContext()
		c.Params = gin.Params{gin.Param{"key", "app.welcome"}}
		c.Request.URL.RawQuery = "language=zzzzzzzz&fallback=false" // Use a language that doesn't exist and disable fallback
		
		handler.GetLocalization(c)
		
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("localization not found", func(t *testing.T) {
		c, w := setupTestContext()
		c.Params = gin.Params{gin.Param{"key", "missing.key"}}
		c.Request.URL.RawQuery = "language=en"
		
		handler.GetLocalization(c)
		
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

// Test BatchLocalize with real data
func TestBatchLocalize_Unit(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := NewMockDatabase()
	memCache := cache.NewMemoryCache(10, 1*time.Hour, 5*time.Minute, logger)
	handler := NewHandler(db, memCache, logger, nil)

	t.Run("success", func(t *testing.T) {
		batchReq := models.GetBatchLocalizationRequest{
			Language: "en",
			Keys:     []string{"app.welcome", "app.error"},
			Fallback: true,
		}
		body, _ := json.Marshal(batchReq)
		
		c, w := setupTestContext()
		c.Request.Body = httptest.NewRequest("POST", "/batch", bytes.NewBuffer(body)).Body
		c.Request.Header.Set("Content-Type", "application/json")
		
		handler.BatchLocalize(c)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response models.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("invalid request body", func(t *testing.T) {
		c, w := setupTestContext()
		c.Request.Body = httptest.NewRequest("POST", "/batch", bytes.NewBuffer([]byte("invalid"))).Body
		c.Request.Header.Set("Content-Type", "application/json")
		
		handler.BatchLocalize(c)
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing language", func(t *testing.T) {
		batchReq := models.GetBatchLocalizationRequest{
			Language: "",
			Keys:     []string{"app.welcome"},
		}
		body, _ := json.Marshal(batchReq)
		
		c, w := setupTestContext()
		c.Request.Body = httptest.NewRequest("POST", "/batch", bytes.NewBuffer(body)).Body
		c.Request.Header.Set("Content-Type", "application/json")
		
		handler.BatchLocalize(c)
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// Test ListLanguages with real data
func TestListLanguages_Unit(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := NewMockDatabase()
	memCache := cache.NewMemoryCache(10, 1*time.Hour, 5*time.Minute, logger)
	handler := NewHandler(db, memCache, logger, nil)

	c, w := setupTestContext()
	
	handler.ListLanguages(c)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)
}

// Test Admin handlers with different roles
func TestAdminHandlers_Unit(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := NewMockDatabase()
	memCache := cache.NewMemoryCache(10, 1*time.Hour, 5*time.Minute, logger)
	handler := NewHandler(db, memCache, logger, nil)

	t.Run("create language with admin role", func(t *testing.T) {
		langData := models.CreateLanguageRequest{
			Code: "fr",
			Name: "French",
		}
		body, _ := json.Marshal(langData)
		
		c, w := setupTestContext()
		c.Request.Body = httptest.NewRequest("POST", "/languages", bytes.NewBuffer(body)).Body
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("user_role", "admin")
		// Set claims with the correct key that middleware expects
		claims := &models.JWTClaims{Username: "adminuser", Role: "admin"}
		c.Set("claims", claims)
		
		handler.CreateLanguage(c)
		
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("create language with user role (forbidden)", func(t *testing.T) {
		langData := models.CreateLanguageRequest{
			Code: "fr",
			Name: "French",
		}
		body, _ := json.Marshal(langData)
		
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("POST", "/v1/admin/languages", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		
		// Create router with middleware
		router := gin.New()
		logger, _ := zap.NewDevelopment()
		db := NewMockDatabase()
		memCache := cache.NewMemoryCache(10, 1*time.Hour, 5*time.Minute, logger)
		handler := NewHandler(db, memCache, logger, nil)
		
		// Set admin middleware before the handler
		router.Use(func(c *gin.Context) {
			// Set JWT claims for a regular user
			claims := &models.JWTClaims{Username: "normaluser", Role: "user"}
			c.Set(middleware.ClaimsKey, claims)
			c.Next()
		})
		router.Use(middleware.AdminOnly([]string{"admin", "superadmin"}))
		router.POST("/v1/admin/languages", handler.CreateLanguage)
		
		// Serve the request
		router.ServeHTTP(w, c.Request)
		
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("get stats with admin role", func(t *testing.T) {
		c, w := setupTestContext()
		c.Set("user_role", "admin")
		
		handler.GetStats(c)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response models.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("get stats with user role (forbidden)", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/v1/admin/stats", nil)
		
		// Create router with middleware
		router := gin.New()
		logger, _ := zap.NewDevelopment()
		db := NewMockDatabase()
		memCache := cache.NewMemoryCache(10, 1*time.Hour, 5*time.Minute, logger)
		handler := NewHandler(db, memCache, logger, nil)
		
		// Set admin middleware before the handler
		router.Use(func(c *gin.Context) {
			// Set JWT claims for a regular user
			claims := &models.JWTClaims{Username: "normaluser", Role: "user"}
			c.Set(middleware.ClaimsKey, claims)
			c.Next()
		})
		router.Use(middleware.AdminOnly([]string{"admin", "superadmin"}))
		router.GET("/v1/admin/stats", handler.GetStats)
		
		// Serve the request
		router.ServeHTTP(w, c.Request)
		
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

// Test Error handling
func TestErrorHandling_Unit(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := NewMockDatabase()
	memCache := cache.NewMemoryCache(10, 1*time.Hour, 5*time.Minute, logger)
	handler := NewHandler(db, memCache, logger, nil)

	t.Run("invalid JSON in create language", func(t *testing.T) {
		c, w := setupTestContext()
		c.Request.Body = httptest.NewRequest("POST", "/languages", bytes.NewBuffer([]byte("{invalid json}"))).Body
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("user_role", "admin")
		
		handler.CreateLanguage(c)
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing required fields", func(t *testing.T) {
		lang := models.Language{
			Code: "",
		}
		body, _ := json.Marshal(lang)
		
		c, w := setupTestContext()
		c.Request.Body = httptest.NewRequest("POST", "/languages", bytes.NewBuffer(body)).Body
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("user_role", "admin")
		
		handler.CreateLanguage(c)
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// Test Version handlers
func TestVersionHandlers_Unit(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := NewMockDatabase()
	memCache := cache.NewMemoryCache(10, 1*time.Hour, 5*time.Minute, logger)
	handler := NewHandler(db, memCache, logger, nil)

	t.Run("get current version", func(t *testing.T) {
		c, w := setupTestContext()
		
		handler.GetCurrentVersion(c)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response models.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("get version history", func(t *testing.T) {
		c, w := setupTestContext()
		c.Request.URL.RawQuery = "limit=10&offset=0"
		
		handler.GetVersionHistory(c)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response models.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("get version by number", func(t *testing.T) {
		c, w := setupTestContext()
		c.Params = gin.Params{gin.Param{"version", "1.0.0"}}
		
		handler.GetVersionByNumber(c)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response models.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("get catalog by version", func(t *testing.T) {
		c, w := setupTestContext()
		c.Params = gin.Params{
			gin.Param{"version", "1.0.0"},
			gin.Param{"language", "en"},
		}
		
		handler.GetCatalogByVersion(c)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response models.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
	})
}

// Benchmark handlers
func BenchmarkGetCatalog(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	db := NewMockDatabase()
	memCache := cache.NewMemoryCache(10, 1*time.Hour, 5*time.Minute, logger)
	handler := NewHandler(db, memCache, logger, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c, _ := setupTestContext()
		c.Params = gin.Params{gin.Param{"language", "en"}}
		handler.GetCatalog(c)
	}
}

func BenchmarkGetLocalization(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	db := NewMockDatabase()
	memCache := cache.NewMemoryCache(10, 1*time.Hour, 5*time.Minute, logger)
	handler := NewHandler(db, memCache, logger, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c, _ := setupTestContext()
		c.Params = gin.Params{gin.Param{"key", "app.welcome"}}
		c.Request.URL.RawQuery = "language=en"
		handler.GetLocalization(c)
	}
}

func BenchmarkBatchLocalize(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	db := NewMockDatabase()
	memCache := cache.NewMemoryCache(10, 1*time.Hour, 5*time.Minute, logger)
	handler := NewHandler(db, memCache, logger, nil)

	batchReq := models.GetBatchLocalizationRequest{
		Language: "en",
		Keys:     []string{"app.welcome", "app.error"},
		Fallback: true,
	}
	body, _ := json.Marshal(batchReq)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c, _ := setupTestContext()
		c.Request.Body = httptest.NewRequest("POST", "/batch", bytes.NewBuffer(body)).Body
		c.Request.Header.Set("Content-Type", "application/json")
		handler.BatchLocalize(c)
	}
}
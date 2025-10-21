package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewLocalizationService(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ls := NewLocalizationService("https://localhost:8085", logger)

	assert.NotNil(t, ls)
	assert.Equal(t, "https://localhost:8085", ls.baseURL)
	assert.NotNil(t, ls.httpClient)
	assert.NotNil(t, ls.cache)
	assert.Equal(t, 1*time.Hour, ls.cache.ttl)
}

func TestSetJWTToken(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ls := NewLocalizationService("https://localhost:8085", logger)

	token := "test-jwt-token"
	ls.SetJWTToken(token)

	assert.Equal(t, token, ls.jwtToken)
}

func TestGetCatalog_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/catalog/en", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		response := APIResponse{
			Success: true,
			Data: &LocalizationCatalogResponse{
				Language: "en",
				Version:  1,
				Checksum: "abc123",
				Catalog: map[string]string{
					"error.success":         "Success",
					"error.invalid_request": "Invalid request",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger, _ := zap.NewDevelopment()
	ls := NewLocalizationService(server.URL, logger)
	ls.SetJWTToken("test-token")

	catalog, err := ls.GetCatalog(context.Background(), "en")

	require.NoError(t, err)
	assert.NotNil(t, catalog)
	assert.Equal(t, "Success", catalog["error.success"])
	assert.Equal(t, "Invalid request", catalog["error.invalid_request"])
}

func TestGetCatalog_CacheHit(t *testing.T) {
	callCount := 0

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		response := APIResponse{
			Success: true,
			Data: &LocalizationCatalogResponse{
				Language: "en",
				Version:  1,
				Checksum: "abc123",
				Catalog: map[string]string{
					"error.success": "Success",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger, _ := zap.NewDevelopment()
	ls := NewLocalizationService(server.URL, logger)

	// First call - should hit server
	catalog1, err1 := ls.GetCatalog(context.Background(), "en")
	require.NoError(t, err1)
	assert.NotNil(t, catalog1)
	assert.Equal(t, 1, callCount)

	// Second call - should hit cache
	catalog2, err2 := ls.GetCatalog(context.Background(), "en")
	require.NoError(t, err2)
	assert.NotNil(t, catalog2)
	assert.Equal(t, 1, callCount) // Call count should not increase
}

func TestGetCatalog_APIError(t *testing.T) {
	// Create mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := APIResponse{
			Success: false,
			Error: &APIError{
				Code:    500,
				Message: "Internal server error",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger, _ := zap.NewDevelopment()
	ls := NewLocalizationService(server.URL, logger)

	catalog, err := ls.GetCatalog(context.Background(), "en")

	assert.Error(t, err)
	assert.Nil(t, catalog)
	assert.Contains(t, err.Error(), "Internal server error")
}

func TestGetCatalog_HTTPError(t *testing.T) {
	// Create mock server that returns HTTP error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	logger, _ := zap.NewDevelopment()
	ls := NewLocalizationService(server.URL, logger)

	catalog, err := ls.GetCatalog(context.Background(), "en")

	assert.Error(t, err)
	assert.Nil(t, catalog)
	assert.Contains(t, err.Error(), "unexpected status code")
}

func TestLocalize_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := APIResponse{
			Success: true,
			Data: &LocalizationCatalogResponse{
				Language: "de",
				Version:  1,
				Checksum: "abc123",
				Catalog: map[string]string{
					"error.success":         "Erfolg",
					"error.invalid_request": "Ungültige Anfrage",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger, _ := zap.NewDevelopment()
	ls := NewLocalizationService(server.URL, logger)

	message, err := ls.Localize(context.Background(), "error.success", "de")

	require.NoError(t, err)
	assert.Equal(t, "Erfolg", message)
}

func TestLocalize_KeyNotFound_ReturnKey(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := APIResponse{
			Success: true,
			Data: &LocalizationCatalogResponse{
				Language: "en",
				Version:  1,
				Checksum: "abc123",
				Catalog: map[string]string{
					"error.success": "Success",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger, _ := zap.NewDevelopment()
	ls := NewLocalizationService(server.URL, logger)

	// Request a key that doesn't exist
	message, err := ls.Localize(context.Background(), "error.nonexistent", "en")

	require.NoError(t, err)
	assert.Equal(t, "error.nonexistent", message) // Should return key as fallback
}

func TestLocalizeBatch_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := APIResponse{
			Success: true,
			Data: &LocalizationCatalogResponse{
				Language: "fr",
				Version:  1,
				Checksum: "abc123",
				Catalog: map[string]string{
					"error.success":         "Succès",
					"error.invalid_request": "Demande invalide",
					"error.unauthorized":    "Non autorisé",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger, _ := zap.NewDevelopment()
	ls := NewLocalizationService(server.URL, logger)

	keys := []string{"error.success", "error.invalid_request", "error.unauthorized"}
	results, err := ls.LocalizeBatch(context.Background(), keys, "fr")

	require.NoError(t, err)
	assert.Len(t, results, 3)
	assert.Equal(t, "Succès", results["error.success"])
	assert.Equal(t, "Demande invalide", results["error.invalid_request"])
	assert.Equal(t, "Non autorisé", results["error.unauthorized"])
}

func TestLocalizeBatch_MissingKeys_ReturnKeysAsFallback(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := APIResponse{
			Success: true,
			Data: &LocalizationCatalogResponse{
				Language: "en",
				Version:  1,
				Checksum: "abc123",
				Catalog: map[string]string{
					"error.success": "Success",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger, _ := zap.NewDevelopment()
	ls := NewLocalizationService(server.URL, logger)

	keys := []string{"error.success", "error.missing1", "error.missing2"}
	results, err := ls.LocalizeBatch(context.Background(), keys, "en")

	require.NoError(t, err)
	assert.Len(t, results, 3)
	assert.Equal(t, "Success", results["error.success"])
	assert.Equal(t, "error.missing1", results["error.missing1"]) // Fallback to key
	assert.Equal(t, "error.missing2", results["error.missing2"]) // Fallback to key
}

func TestInvalidateCache(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := APIResponse{
			Success: true,
			Data: &LocalizationCatalogResponse{
				Language: "en",
				Version:  1,
				Checksum: "abc123",
				Catalog: map[string]string{
					"error.success": "Success",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger, _ := zap.NewDevelopment()
	ls := NewLocalizationService(server.URL, logger)

	// Load catalog to populate cache
	_, err := ls.GetCatalog(context.Background(), "en")
	require.NoError(t, err)

	// Verify cache is populated
	cached := ls.cache.Get("en")
	assert.NotNil(t, cached)

	// Invalidate cache
	ls.InvalidateCache("en")

	// Verify cache is cleared
	cached = ls.cache.Get("en")
	assert.Nil(t, cached)
}

func TestLocalizationCache_SetAndGet(t *testing.T) {
	cache := &LocalizationCache{
		catalogs: make(map[string]*CachedCatalog),
		ttl:      1 * time.Hour,
	}

	catalog := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	cache.Set("en", catalog)

	cached := cache.Get("en")
	assert.NotNil(t, cached)
	assert.Equal(t, "en", cached.Language)
	assert.Equal(t, catalog, cached.Catalog)
}

func TestLocalizationCache_Expiration(t *testing.T) {
	cache := &LocalizationCache{
		catalogs: make(map[string]*CachedCatalog),
		ttl:      100 * time.Millisecond, // Short TTL for testing
	}

	catalog := map[string]string{
		"key1": "value1",
	}

	cache.Set("en", catalog)

	// Should be available immediately
	cached := cache.Get("en")
	assert.NotNil(t, cached)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired
	cached = cache.Get("en")
	assert.Nil(t, cached)
}

func TestLocalizationCache_Delete(t *testing.T) {
	cache := &LocalizationCache{
		catalogs: make(map[string]*CachedCatalog),
		ttl:      1 * time.Hour,
	}

	catalog := map[string]string{
		"key1": "value1",
	}

	cache.Set("en", catalog)
	cache.Set("de", catalog)

	// Verify both are cached
	assert.NotNil(t, cache.Get("en"))
	assert.NotNil(t, cache.Get("de"))

	// Delete one
	cache.Delete("en")

	// Verify only one is deleted
	assert.Nil(t, cache.Get("en"))
	assert.NotNil(t, cache.Get("de"))
}

func TestLocalizationCache_Clear(t *testing.T) {
	cache := &LocalizationCache{
		catalogs: make(map[string]*CachedCatalog),
		ttl:      1 * time.Hour,
	}

	catalog := map[string]string{
		"key1": "value1",
	}

	cache.Set("en", catalog)
	cache.Set("de", catalog)
	cache.Set("fr", catalog)

	// Verify all are cached
	assert.NotNil(t, cache.Get("en"))
	assert.NotNil(t, cache.Get("de"))
	assert.NotNil(t, cache.Get("fr"))

	// Clear all
	cache.Clear()

	// Verify all are cleared
	assert.Nil(t, cache.Get("en"))
	assert.Nil(t, cache.Get("de"))
	assert.Nil(t, cache.Get("fr"))
}

func TestLocalizationCache_ConcurrentAccess(t *testing.T) {
	cache := &LocalizationCache{
		catalogs: make(map[string]*CachedCatalog),
		ttl:      1 * time.Hour,
	}

	catalog := map[string]string{
		"key1": "value1",
	}

	// Test concurrent writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(lang string) {
			cache.Set(lang, catalog)
			done <- true
		}(string(rune('a' + i)))
	}

	// Wait for all writes
	for i := 0; i < 10; i++ {
		<-done
	}

	// Test concurrent reads
	for i := 0; i < 10; i++ {
		go func(lang string) {
			cached := cache.Get(lang)
			assert.NotNil(t, cached)
			done <- true
		}(string(rune('a' + i)))
	}

	// Wait for all reads
	for i := 0; i < 10; i++ {
		<-done
	}
}

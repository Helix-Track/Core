package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

// LocalizationService client for the Localization service
type LocalizationService struct {
	baseURL    string
	jwtToken   string
	httpClient *http.Client
	cache      *LocalizationCache
	logger     *zap.Logger
}

// LocalizationCache implements in-memory caching for localizations
type LocalizationCache struct {
	mu       sync.RWMutex
	catalogs map[string]*CachedCatalog
	ttl      time.Duration
}

// CachedCatalog represents a cached localization catalog
type CachedCatalog struct {
	Language  string
	Catalog   map[string]string
	ExpiresAt time.Time
}

// LocalizationCatalogResponse represents the API response
type LocalizationCatalogResponse struct {
	Language string            `json:"language"`
	Version  int               `json:"version"`
	Checksum string            `json:"checksum"`
	Catalog  map[string]string `json:"catalog"`
}

// APIResponse represents the standard API response
type APIResponse struct {
	Success bool                         `json:"success"`
	Data    *LocalizationCatalogResponse `json:"data,omitempty"`
	Error   *APIError                    `json:"error,omitempty"`
}

// APIError represents an API error
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewLocalizationService creates a new localization service client
func NewLocalizationService(baseURL string, logger *zap.Logger) *LocalizationService {
	return &LocalizationService{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		cache: &LocalizationCache{
			catalogs: make(map[string]*CachedCatalog),
			ttl:      1 * time.Hour,
		},
		logger: logger,
	}
}

// SetJWTToken sets the JWT token for authentication
func (ls *LocalizationService) SetJWTToken(token string) {
	ls.jwtToken = token
}

// GetCatalog retrieves a complete localization catalog for a language
func (ls *LocalizationService) GetCatalog(ctx context.Context, language string) (map[string]string, error) {
	// Check cache first
	if cached := ls.cache.Get(language); cached != nil {
		ls.logger.Debug("localization catalog cache hit", zap.String("language", language))
		return cached.Catalog, nil
	}

	// Fetch from service
	url := fmt.Sprintf("%s/v1/catalog/%s", ls.baseURL, language)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add JWT token if available
	if ls.jwtToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ls.jwtToken))
	}

	resp, err := ls.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch catalog: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !apiResp.Success || apiResp.Data == nil {
		if apiResp.Error != nil {
			return nil, fmt.Errorf("API error: %s", apiResp.Error.Message)
		}
		return nil, fmt.Errorf("failed to fetch catalog")
	}

	// Cache the catalog
	ls.cache.Set(language, apiResp.Data.Catalog)

	ls.logger.Info("localization catalog fetched",
		zap.String("language", language),
		zap.Int("entries", len(apiResp.Data.Catalog)),
	)

	return apiResp.Data.Catalog, nil
}

// Localize retrieves a single localized string
func (ls *LocalizationService) Localize(ctx context.Context, key, language string) (string, error) {
	// Try to get from cached catalog first
	catalog, err := ls.GetCatalog(ctx, language)
	if err != nil {
		return key, err // Return key as fallback
	}

	if value, exists := catalog[key]; exists {
		return value, nil
	}

	return key, nil // Return key as fallback
}

// LocalizeBatch retrieves multiple localized strings at once
func (ls *LocalizationService) LocalizeBatch(ctx context.Context, keys []string, language string) (map[string]string, error) {
	// Try to get from cached catalog first
	catalog, err := ls.GetCatalog(ctx, language)
	if err != nil {
		// Return keys as fallback
		result := make(map[string]string)
		for _, key := range keys {
			result[key] = key
		}
		return result, err
	}

	result := make(map[string]string)
	for _, key := range keys {
		if value, exists := catalog[key]; exists {
			result[key] = value
		} else {
			result[key] = key // Fallback to key
		}
	}

	return result, nil
}

// InvalidateCache invalidates the cache for a specific language
func (ls *LocalizationService) InvalidateCache(language string) {
	ls.cache.Delete(language)
	ls.logger.Info("localization cache invalidated", zap.String("language", language))
}

// Get retrieves a cached catalog
func (lc *LocalizationCache) Get(language string) *CachedCatalog {
	lc.mu.RLock()
	defer lc.mu.RUnlock()

	cached, exists := lc.catalogs[language]
	if !exists {
		return nil
	}

	// Check expiration
	if time.Now().After(cached.ExpiresAt) {
		return nil
	}

	return cached
}

// Set stores a catalog in cache
func (lc *LocalizationCache) Set(language string, catalog map[string]string) {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	lc.catalogs[language] = &CachedCatalog{
		Language:  language,
		Catalog:   catalog,
		ExpiresAt: time.Now().Add(lc.ttl),
	}
}

// Delete removes a catalog from cache
func (lc *LocalizationCache) Delete(language string) {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	delete(lc.catalogs, language)
}

// Clear removes all catalogs from cache
func (lc *LocalizationCache) Clear() {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	lc.catalogs = make(map[string]*CachedCatalog)
}

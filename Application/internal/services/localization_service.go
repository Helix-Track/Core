package services

import (
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
	baseURL      string
	jwtToken     string
	httpClient   *http.Client
	cache        *LocalizationCache
	logger       *zap.Logger
	wsClient     *LocalizationWebSocketClient
	wsEnabled    bool
	wsCtx        context.Context
	wsCancel     context.CancelFunc
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

// EnableWebSocket enables WebSocket support for real-time updates
func (ls *LocalizationService) EnableWebSocket() {
	ls.wsEnabled = true
	ls.logger.Info("websocket support enabled for localization service")
}

// StartWebSocket starts the WebSocket client connection
func (ls *LocalizationService) StartWebSocket() error {
	if !ls.wsEnabled {
		ls.logger.Warn("websocket not enabled, skipping start")
		return nil
	}

	if ls.wsClient != nil {
		ls.logger.Warn("websocket client already started")
		return nil
	}

	// Create WebSocket client
	ls.wsClient = NewLocalizationWebSocketClient(ls.baseURL, ls, ls.logger)

	// Create context for WebSocket lifecycle
	ls.wsCtx, ls.wsCancel = context.WithCancel(context.Background())

	// Start the WebSocket client in a goroutine
	go func() {
		if err := ls.wsClient.Start(ls.wsCtx); err != nil {
			ls.logger.Error("websocket client error", zap.Error(err))
		}
	}()

	ls.logger.Info("websocket client started")
	return nil
}

// StopWebSocket stops the WebSocket client connection
func (ls *LocalizationService) StopWebSocket() {
	if ls.wsClient == nil {
		return
	}

	ls.logger.Info("stopping websocket client")

	// Cancel the context to signal shutdown
	if ls.wsCancel != nil {
		ls.wsCancel()
	}

	// Stop the client
	ls.wsClient.Stop()

	ls.wsClient = nil
	ls.wsCtx = nil
	ls.wsCancel = nil

	ls.logger.Info("websocket client stopped")
}

// IsWebSocketConnected returns whether the WebSocket is connected
func (ls *LocalizationService) IsWebSocketConnected() bool {
	if ls.wsClient == nil {
		return false
	}
	return ls.wsClient.IsConnected()
}

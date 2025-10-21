package models

// CreateLocalizationRequest represents a request to create/update a localization
type CreateLocalizationRequest struct {
	Key         string                 `json:"key" binding:"required"`
	Language    string                 `json:"language" binding:"required"`
	Value       string                 `json:"value" binding:"required"`
	Category    string                 `json:"category"`
	Description string                 `json:"description"`
	Context     string                 `json:"context"`
	PluralForms map[string]string      `json:"plural_forms"`
	Variables   []string               `json:"variables"`
	Approved    bool                   `json:"approved"`
}

// GetBatchLocalizationRequest represents a batch request for fetching multiple keys
type GetBatchLocalizationRequest struct {
	Keys     []string `json:"keys" binding:"required"`
	Language string   `json:"language" binding:"required"`
	Fallback bool     `json:"fallback"`
}

// CreateLanguageRequest represents a request to create/update a language
type CreateLanguageRequest struct {
	Code       string `json:"code" binding:"required"`
	Name       string `json:"name" binding:"required"`
	NativeName string `json:"native_name"`
	IsRTL      bool   `json:"is_rtl"`
	IsActive   bool   `json:"is_active"`
	IsDefault  bool   `json:"is_default"`
}

// CacheInvalidationRequest represents a request to invalidate cache
type CacheInvalidationRequest struct {
	Language string `json:"language"`
	Category string `json:"category"`
}

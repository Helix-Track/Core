package models

import (
	"encoding/json"
	"fmt"
)

// ImportRequest represents a bulk import request
type ImportRequest struct {
	ImportType        string                       `json:"import_type"`        // "full" or "incremental"
	OverwriteExisting bool                         `json:"overwrite_existing"` // Whether to overwrite existing entries
	Data              ImportData                   `json:"data"`
}

// ImportData contains the data to be imported
type ImportData struct {
	Languages      []ImportLanguage              `json:"languages"`
	Keys           []ImportLocalizationKey       `json:"keys"`
	Localizations  map[string]map[string]string  `json:"localizations"` // language_code -> key -> value
}

// ImportLanguage represents a language in import data
type ImportLanguage struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	NativeName string `json:"native_name"`
	IsRTL      bool   `json:"is_rtl"`
	IsActive   bool   `json:"is_active"`
	IsDefault  bool   `json:"is_default"`
}

// ImportLocalizationKey represents a localization key in import data
type ImportLocalizationKey struct {
	Key         string   `json:"key"`
	Category    string   `json:"category"`
	Description string   `json:"description"`
	Context     string   `json:"context"`
	Variables   []string `json:"variables"`
}

// ImportResponse represents the response from an import operation
type ImportResponse struct {
	Success bool               `json:"success"`
	Summary ImportSummary      `json:"summary"`
	Errors  []ImportError      `json:"errors,omitempty"`
}

// ImportSummary contains statistics about the import
type ImportSummary struct {
	LanguagesImported     int `json:"languages_imported"`
	LanguagesSkipped      int `json:"languages_skipped"`
	LanguagesUpdated      int `json:"languages_updated"`
	KeysImported          int `json:"keys_imported"`
	KeysSkipped           int `json:"keys_skipped"`
	KeysUpdated           int `json:"keys_updated"`
	LocalizationsImported int `json:"localizations_imported"`
	LocalizationsSkipped  int `json:"localizations_skipped"`
	LocalizationsUpdated  int `json:"localizations_updated"`
	TotalProcessed        int `json:"total_processed"`
	DurationMs            int64 `json:"duration_ms"`
}

// ImportError represents an error during import
type ImportError struct {
	Type    string `json:"type"`    // "language", "key", "localization"
	ID      string `json:"id"`      // Identifier (code, key, etc.)
	Message string `json:"message"`
}

// ExportRequest represents an export request
type ExportRequest struct {
	Format           string   `json:"format"`             // "json", "csv", "xliff"
	Languages        []string `json:"languages"`          // Empty = all languages
	Categories       []string `json:"categories"`         // Empty = all categories
	IncludeMetadata  bool     `json:"include_metadata"`
	OnlyApproved     bool     `json:"only_approved"`
	Compress         bool     `json:"compress"`
}

// ExportResponse represents the response from an export operation
type ExportResponse struct {
	Success  bool              `json:"success"`
	Format   string            `json:"format"`
	Data     interface{}       `json:"data,omitempty"`     // For JSON format
	FileURL  string            `json:"file_url,omitempty"` // For downloadable formats
	Metadata ExportMetadata    `json:"metadata"`
}

// ExportMetadata contains metadata about the export
type ExportMetadata struct {
	ExportedAt        int64  `json:"exported_at"`
	Languages         int    `json:"languages"`
	Keys              int    `json:"keys"`
	Localizations     int    `json:"localizations"`
	Format            string `json:"format"`
	Compressed        bool   `json:"compressed"`
	Version           string `json:"version"`
}

// BatchLocalizationRequest represents a batch operation request
type BatchLocalizationRequest struct {
	Operation     string                    `json:"operation"` // "create", "update", "delete", "approve"
	Localizations []BatchLocalizationItem   `json:"localizations"`
}

// BatchLocalizationItem represents a single item in a batch operation
type BatchLocalizationItem struct {
	Key          string `json:"key"`
	LanguageCode string `json:"language_code"`
	Value        string `json:"value,omitempty"`
	Approved     bool   `json:"approved,omitempty"`
}

// BatchLocalizationResponse represents the response from a batch operation
type BatchLocalizationResponse struct {
	Success   bool                      `json:"success"`
	Summary   BatchSummary              `json:"summary"`
	Errors    []BatchError              `json:"errors,omitempty"`
}

// BatchSummary contains statistics about the batch operation
type BatchSummary struct {
	TotalRequested int   `json:"total_requested"`
	Successful     int   `json:"successful"`
	Failed         int   `json:"failed"`
	Skipped        int   `json:"skipped"`
	DurationMs     int64 `json:"duration_ms"`
}

// BatchError represents an error in a batch operation
type BatchError struct {
	Index   int    `json:"index"`
	Key     string `json:"key"`
	Message string `json:"message"`
}

// Validate validates an ImportRequest
func (ir *ImportRequest) Validate() error {
	if ir.ImportType != "full" && ir.ImportType != "incremental" {
		return fmt.Errorf("invalid import_type: must be 'full' or 'incremental'")
	}

	if len(ir.Data.Languages) == 0 && len(ir.Data.Keys) == 0 && len(ir.Data.Localizations) == 0 {
		return fmt.Errorf("import data is empty")
	}

	// Validate languages
	for i, lang := range ir.Data.Languages {
		if lang.Code == "" {
			return fmt.Errorf("language at index %d: code is required", i)
		}
		if lang.Name == "" {
			return fmt.Errorf("language at index %d: name is required", i)
		}
	}

	// Validate keys
	for i, key := range ir.Data.Keys {
		if key.Key == "" {
			return fmt.Errorf("localization key at index %d: key is required", i)
		}
		if key.Category == "" {
			return fmt.Errorf("localization key at index %d: category is required", i)
		}
	}

	return nil
}

// Validate validates an ExportRequest
func (er *ExportRequest) Validate() error {
	validFormats := map[string]bool{"json": true, "csv": true, "xliff": true}
	if !validFormats[er.Format] {
		return fmt.Errorf("invalid format: must be 'json', 'csv', or 'xliff'")
	}

	return nil
}

// Validate validates a BatchLocalizationRequest
func (br *BatchLocalizationRequest) Validate() error {
	validOperations := map[string]bool{"create": true, "update": true, "delete": true, "approve": true}
	if !validOperations[br.Operation] {
		return fmt.Errorf("invalid operation: must be 'create', 'update', 'delete', or 'approve'")
	}

	if len(br.Localizations) == 0 {
		return fmt.Errorf("localizations array is empty")
	}

	// Validate each item
	for i, item := range br.Localizations {
		if item.Key == "" {
			return fmt.Errorf("item at index %d: key is required", i)
		}
		if item.LanguageCode == "" {
			return fmt.Errorf("item at index %d: language_code is required", i)
		}
		if br.Operation != "delete" && br.Operation != "approve" && item.Value == "" {
			return fmt.Errorf("item at index %d: value is required for %s operation", i, br.Operation)
		}
	}

	return nil
}

// ToJSON converts ImportResponse to JSON
func (ir *ImportResponse) ToJSON() ([]byte, error) {
	return json.MarshalIndent(ir, "", "  ")
}

// ToJSON converts ExportResponse to JSON
func (er *ExportResponse) ToJSON() ([]byte, error) {
	return json.MarshalIndent(er, "", "  ")
}

// ToJSON converts BatchLocalizationResponse to JSON
func (br *BatchLocalizationResponse) ToJSON() ([]byte, error) {
	return json.MarshalIndent(br, "", "  ")
}

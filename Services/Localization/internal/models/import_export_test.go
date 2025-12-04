package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImportRequest_Validate(t *testing.T) {
	tests := []struct {
		name        string
		req         ImportRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid full import",
			req: ImportRequest{
				ImportType:        "full",
				OverwriteExisting: true,
				Data: ImportData{
					Languages: []ImportLanguage{
						{Code: "en", Name: "English"},
					},
					Keys: []ImportLocalizationKey{
						{Key: "app.welcome", Category: "general"},
					},
					Localizations: map[string]map[string]string{
						"en": {"app.welcome": "Welcome!"},
					},
				},
			},
			expectError: false,
		},
		{
			name: "valid incremental import",
			req: ImportRequest{
				ImportType:        "incremental",
				OverwriteExisting: false,
				Data: ImportData{
					Localizations: map[string]map[string]string{
						"en": {"app.error": "Error occurred"},
					},
				},
			},
			expectError: false,
		},
		{
			name: "invalid import type",
			req: ImportRequest{
				ImportType: "invalid",
				Data: ImportData{
					Languages: []ImportLanguage{{Code: "en", Name: "English"}},
				},
			},
			expectError: true,
			errorMsg:    "invalid import_type: must be 'full' or 'incremental'",
		},
		{
			name: "empty import data",
			req: ImportRequest{
				ImportType: "full",
				Data:       ImportData{},
			},
			expectError: true,
			errorMsg:    "import data is empty",
		},
		{
			name: "missing language code",
			req: ImportRequest{
				ImportType: "full",
				Data: ImportData{
					Languages: []ImportLanguage{
						{Name: "English"}, // Missing code
					},
				},
			},
			expectError: true,
			errorMsg:    "language at index 0: code is required",
		},
		{
			name: "missing language name",
			req: ImportRequest{
				ImportType: "full",
				Data: ImportData{
					Languages: []ImportLanguage{
						{Code: "en"}, // Missing name
					},
				},
			},
			expectError: true,
			errorMsg:    "language at index 0: name is required",
		},
		{
			name: "missing localization key",
			req: ImportRequest{
				ImportType: "full",
				Data: ImportData{
					Keys: []ImportLocalizationKey{
						{Category: "general"}, // Missing key
					},
				},
			},
			expectError: true,
			errorMsg:    "localization key at index 0: key is required",
		},
		{
			name: "missing key category",
			req: ImportRequest{
				ImportType: "full",
				Data: ImportData{
					Keys: []ImportLocalizationKey{
						{Key: "app.welcome"}, // Missing category
					},
				},
			},
			expectError: true,
			errorMsg:    "localization key at index 0: category is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExportRequest_Validate(t *testing.T) {
	tests := []struct {
		name        string
		req         ExportRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid JSON export",
			req: ExportRequest{
				Format:          "json",
				Languages:       []string{"en", "de"},
				Categories:      []string{"general"},
				IncludeMetadata:  true,
				OnlyApproved:    false,
				Compress:        false,
			},
			expectError: false,
		},
		{
			name: "valid CSV export",
			req: ExportRequest{
				Format: "csv",
			},
			expectError: false,
		},
		{
			name: "valid XLIFF export",
			req: ExportRequest{
				Format: "xliff",
			},
			expectError: false,
		},
		{
			name: "invalid format",
			req: ExportRequest{
				Format: "invalid",
			},
			expectError: true,
			errorMsg:    "invalid format: must be 'json', 'csv', or 'xliff'",
		},
		{
			name: "empty format",
			req: ExportRequest{
				Format: "",
			},
			expectError: true,
			errorMsg:    "invalid format: must be 'json', 'csv', or 'xliff'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBatchLocalizationRequest_Validate(t *testing.T) {
	tests := []struct {
		name        string
		req         BatchLocalizationRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid create operation",
			req: BatchLocalizationRequest{
				Operation: "create",
				Localizations: []BatchLocalizationItem{
					{
						Key:          "app.welcome",
						LanguageCode: "en",
						Value:        "Welcome!",
					},
				},
			},
			expectError: false,
		},
		{
			name: "valid update operation",
			req: BatchLocalizationRequest{
				Operation: "update",
				Localizations: []BatchLocalizationItem{
					{
						Key:          "app.welcome",
						LanguageCode: "en",
						Value:        "Welcome back!",
					},
				},
			},
			expectError: false,
		},
		{
			name: "valid delete operation",
			req: BatchLocalizationRequest{
				Operation: "delete",
				Localizations: []BatchLocalizationItem{
					{
						Key:          "app.welcome",
						LanguageCode: "en",
					},
				},
			},
			expectError: false,
		},
		{
			name: "valid approve operation",
			req: BatchLocalizationRequest{
				Operation: "approve",
				Localizations: []BatchLocalizationItem{
					{
						Key:          "app.welcome",
						LanguageCode: "en",
					},
				},
			},
			expectError: false,
		},
		{
			name: "invalid operation",
			req: BatchLocalizationRequest{
				Operation: "invalid",
				Localizations: []BatchLocalizationItem{
					{Key: "app.welcome", LanguageCode: "en"},
				},
			},
			expectError: true,
			errorMsg:    "invalid operation: must be 'create', 'update', 'delete', or 'approve'",
		},
		{
			name: "empty localizations",
			req: BatchLocalizationRequest{
				Operation:     "create",
				Localizations: []BatchLocalizationItem{},
			},
			expectError: true,
			errorMsg:    "localizations array is empty",
		},
		{
			name: "missing key",
			req: BatchLocalizationRequest{
				Operation: "create",
				Localizations: []BatchLocalizationItem{
					{
						LanguageCode: "en",
						Value:        "Welcome!",
					},
				},
			},
			expectError: true,
			errorMsg:    "item at index 0: key is required",
		},
		{
			name: "missing language code",
			req: BatchLocalizationRequest{
				Operation: "create",
				Localizations: []BatchLocalizationItem{
					{
						Key:   "app.welcome",
						Value: "Welcome!",
					},
				},
			},
			expectError: true,
			errorMsg:    "item at index 0: language_code is required",
		},
		{
			name: "missing value for create",
			req: BatchLocalizationRequest{
				Operation: "create",
				Localizations: []BatchLocalizationItem{
					{
						Key:          "app.welcome",
						LanguageCode: "en",
					},
				},
			},
			expectError: true,
			errorMsg:    "item at index 0: value is required for create operation",
		},
		{
			name: "missing value for update",
			req: BatchLocalizationRequest{
				Operation: "update",
				Localizations: []BatchLocalizationItem{
					{
						Key:          "app.welcome",
						LanguageCode: "en",
					},
				},
			},
			expectError: true,
			errorMsg:    "item at index 0: value is required for update operation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestImportResponse_ToJSON(t *testing.T) {
	resp := ImportResponse{
		Success: true,
		Summary: ImportSummary{
			LanguagesImported:     2,
			KeysImported:          5,
			LocalizationsImported: 10,
			TotalProcessed:        15,
			DurationMs:           1000,
		},
		Errors: []ImportError{
			{
				Type:    "language",
				ID:      "invalid-lang",
				Message: "Invalid language code",
			},
		},
	}

	jsonBytes, err := resp.ToJSON()
	require.NoError(t, err)
	require.True(t, len(jsonBytes) > 0)

	// Verify it's valid JSON and contains expected fields
	var parsed ImportResponse
	err = json.Unmarshal(jsonBytes, &parsed)
	require.NoError(t, err)
	assert.Equal(t, resp.Success, parsed.Success)
	assert.Equal(t, resp.Summary.LanguagesImported, parsed.Summary.LanguagesImported)
	assert.Len(t, parsed.Errors, 1)
	assert.Equal(t, resp.Errors[0].Type, parsed.Errors[0].Type)
}

func TestExportResponse_ToJSON(t *testing.T) {
	resp := ExportResponse{
		Success: true,
		Format:  "json",
		Data: map[string]string{
			"app.welcome": "Welcome!",
			"app.error":   "Error occurred",
		},
		Metadata: ExportMetadata{
			ExportedAt:   time.Now().Unix(),
			Languages:    2,
			Keys:         5,
			Localizations: 10,
			Format:       "json",
			Compressed:   false,
			Version:      "1.0.0",
		},
	}

	jsonBytes, err := resp.ToJSON()
	require.NoError(t, err)
	require.True(t, len(jsonBytes) > 0)

	// Verify it's valid JSON
	var parsed ExportResponse
	err = json.Unmarshal(jsonBytes, &parsed)
	require.NoError(t, err)
	assert.Equal(t, resp.Success, parsed.Success)
	assert.Equal(t, resp.Format, parsed.Format)
	assert.Equal(t, resp.Metadata.Keys, parsed.Metadata.Keys)
}

func TestBatchLocalizationResponse_ToJSON(t *testing.T) {
	resp := BatchLocalizationResponse{
		Success: true,
		Summary: BatchSummary{
			TotalRequested: 5,
			Successful:     4,
			Failed:         1,
			Skipped:        0,
			DurationMs:     500,
		},
		Errors: []BatchError{
			{
				Index:   3,
				Key:     "missing.key",
				Message: "Key not found",
			},
		},
	}

	jsonBytes, err := resp.ToJSON()
	require.NoError(t, err)
	require.True(t, len(jsonBytes) > 0)

	// Verify it's valid JSON
	var parsed BatchLocalizationResponse
	err = json.Unmarshal(jsonBytes, &parsed)
	require.NoError(t, err)
	assert.Equal(t, resp.Success, parsed.Success)
	assert.Equal(t, resp.Summary.TotalRequested, parsed.Summary.TotalRequested)
	assert.Len(t, parsed.Errors, 1)
	assert.Equal(t, resp.Errors[0].Key, parsed.Errors[0].Key)
}

func TestImportExportModels_JSONMarshaling(t *testing.T) {
	// Test ImportLanguage
	lang := ImportLanguage{
		Code:       "en",
		Name:       "English",
		NativeName: "English",
		IsRTL:      false,
		IsActive:   true,
		IsDefault:  true,
	}
	data, err := json.Marshal(lang)
	require.NoError(t, err)
	var unmarshaled ImportLanguage
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, lang, unmarshaled)

	// Test ImportLocalizationKey
	key := ImportLocalizationKey{
		Key:         "app.welcome",
		Category:    "general",
		Description: "Welcome message",
		Context:     "Login page",
		Variables:   []string{"username"},
	}
	data, err = json.Marshal(key)
	require.NoError(t, err)
	var unmarshaledKey ImportLocalizationKey
	err = json.Unmarshal(data, &unmarshaledKey)
	require.NoError(t, err)
	assert.Equal(t, key, unmarshaledKey)

	// Test ImportError
	impErr := ImportError{
		Type:    "localization",
		ID:      "app.welcome",
		Message: "Value too long",
	}
	data, err = json.Marshal(impErr)
	require.NoError(t, err)
	var unmarshaledError ImportError
	err = json.Unmarshal(data, &unmarshaledError)
	require.NoError(t, err)
	assert.Equal(t, impErr, unmarshaledError)

	// Test BatchLocalizationItem
	item := BatchLocalizationItem{
		Key:          "app.welcome",
		LanguageCode: "en",
		Value:        "Welcome!",
		Approved:     true,
	}
	data, err = json.Marshal(item)
	require.NoError(t, err)
	var unmarshaledItem BatchLocalizationItem
	err = json.Unmarshal(data, &unmarshaledItem)
	require.NoError(t, err)
	assert.Equal(t, item, unmarshaledItem)

	// Test BatchError
	batchErr := BatchError{
		Index:   2,
		Key:     "invalid.key",
		Message: "Invalid format",
	}
	data, err = json.Marshal(batchErr)
	require.NoError(t, err)
	var unmarshaledBatchError BatchError
	err = json.Unmarshal(data, &unmarshaledBatchError)
	require.NoError(t, err)
	assert.Equal(t, batchErr, unmarshaledBatchError)
}
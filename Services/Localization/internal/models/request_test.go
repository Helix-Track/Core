package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateLocalizationRequest_JSONMarshaling(t *testing.T) {
	req := CreateLocalizationRequest{
		Key:         "app.welcome",
		Language:    "en",
		Value:       "Welcome!",
		Category:    "general",
		Description: "Welcome message shown on login",
		Context:     "Login page",
		PluralForms: map[string]string{
			"one": "One item",
			"many": "Many items",
		},
		Variables:   []string{"username"},
		Approved:    false,
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(req)
	require.NoError(t, err)

	// Test JSON unmarshaling
	var unmarshaled CreateLocalizationRequest
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, req.Key, unmarshaled.Key)
	assert.Equal(t, req.Language, unmarshaled.Language)
	assert.Equal(t, req.Value, unmarshaled.Value)
	assert.Equal(t, req.Category, unmarshaled.Category)
	assert.Equal(t, req.Description, unmarshaled.Description)
	assert.Equal(t, req.Context, unmarshaled.Context)
	assert.Equal(t, req.Approved, unmarshaled.Approved)
	assert.Equal(t, "username", unmarshaled.Variables[0])
	assert.Equal(t, "One item", unmarshaled.PluralForms["one"])
	assert.Equal(t, "Many items", unmarshaled.PluralForms["many"])
}

func TestCreateLocalizationRequest_Minimal(t *testing.T) {
	req := CreateLocalizationRequest{
		Key:      "minimal.key",
		Language: "fr",
		Value:    "Clé minimale",
	}

	jsonData, err := json.Marshal(req)
	require.NoError(t, err)

	var unmarshaled CreateLocalizationRequest
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, req.Key, unmarshaled.Key)
	assert.Equal(t, req.Language, unmarshaled.Language)
	assert.Equal(t, req.Value, unmarshaled.Value)
	assert.Equal(t, "", unmarshaled.Category)
	assert.Equal(t, "", unmarshaled.Description)
	assert.Equal(t, "", unmarshaled.Context)
	assert.False(t, unmarshaled.Approved)
	assert.Empty(t, unmarshaled.Variables)
	assert.Empty(t, unmarshaled.PluralForms)
}

func TestGetBatchLocalizationRequest_JSONMarshaling(t *testing.T) {
	req := GetBatchLocalizationRequest{
		Keys:     []string{"app.welcome", "app.error", "app.success"},
		Language: "en",
		Fallback: true,
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(req)
	require.NoError(t, err)

	// Test JSON unmarshaling
	var unmarshaled GetBatchLocalizationRequest
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, req.Language, unmarshaled.Language)
	assert.Equal(t, req.Fallback, unmarshaled.Fallback)
	assert.Equal(t, len(req.Keys), len(unmarshaled.Keys))
	
	for i, key := range req.Keys {
		assert.Equal(t, key, unmarshaled.Keys[i])
	}
}

func TestGetBatchLocalizationRequest_EmptyKeys(t *testing.T) {
	req := GetBatchLocalizationRequest{
		Keys:     []string{},
		Language: "en",
		Fallback: false,
	}

	jsonData, err := json.Marshal(req)
	require.NoError(t, err)

	var unmarshaled GetBatchLocalizationRequest
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, req.Language, unmarshaled.Language)
	assert.Equal(t, req.Fallback, unmarshaled.Fallback)
	assert.Empty(t, unmarshaled.Keys)
}

func TestCreateLanguageRequest_JSONMarshaling(t *testing.T) {
	req := CreateLanguageRequest{
		Code:       "ja",
		Name:       "Japanese",
		NativeName: "日本語",
		IsRTL:      false,
		IsActive:   true,
		IsDefault:  false,
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(req)
	require.NoError(t, err)

	// Test JSON unmarshaling
	var unmarshaled CreateLanguageRequest
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, req.Code, unmarshaled.Code)
	assert.Equal(t, req.Name, unmarshaled.Name)
	assert.Equal(t, req.NativeName, unmarshaled.NativeName)
	assert.Equal(t, req.IsRTL, unmarshaled.IsRTL)
	assert.Equal(t, req.IsActive, unmarshaled.IsActive)
	assert.Equal(t, req.IsDefault, unmarshaled.IsDefault)
}

func TestCreateLanguageRequest_Minimal(t *testing.T) {
	req := CreateLanguageRequest{
		Code: "es",
		Name: "Spanish",
	}

	jsonData, err := json.Marshal(req)
	require.NoError(t, err)

	var unmarshaled CreateLanguageRequest
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, req.Code, unmarshaled.Code)
	assert.Equal(t, req.Name, unmarshaled.Name)
	assert.Equal(t, "", unmarshaled.NativeName)
	assert.False(t, unmarshaled.IsRTL)
	assert.False(t, unmarshaled.IsActive)
	assert.False(t, unmarshaled.IsDefault)
}

func TestCacheInvalidationRequest_JSONMarshaling(t *testing.T) {
	req := CacheInvalidationRequest{
		Language: "en",
		Category: "ui",
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(req)
	require.NoError(t, err)

	// Test JSON unmarshaling
	var unmarshaled CacheInvalidationRequest
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, req.Language, unmarshaled.Language)
	assert.Equal(t, req.Category, unmarshaled.Category)
}

func TestCacheInvalidationRequest_Partial(t *testing.T) {
	tests := []struct {
		name   string
		req    CacheInvalidationRequest
		expect string
	}{
		{
			name: "only language",
			req: CacheInvalidationRequest{
				Language: "en",
			},
			expect: `{"language":"en"}`,
		},
		{
			name: "only category",
			req: CacheInvalidationRequest{
				Category: "general",
			},
			expect: `{"category":"general"}`,
		},
		{
			name: "both language and category",
			req: CacheInvalidationRequest{
				Language: "fr",
				Category: "errors",
			},
			expect: `{"language":"fr","category":"errors"}`,
		},
		{
			name: "empty",
			req: CacheInvalidationRequest{},
			expect: `{}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.req)
			require.NoError(t, err)

			var unmarshaled CacheInvalidationRequest
			err = json.Unmarshal(jsonData, &unmarshaled)
			require.NoError(t, err)

			assert.Equal(t, tt.req.Language, unmarshaled.Language)
			assert.Equal(t, tt.req.Category, unmarshaled.Category)
		})
	}
}

func TestAllRequestModels_WithNilValues(t *testing.T) {
	// Test that models can handle nil slices and maps gracefully
	
	// CreateLocalizationRequest with nil values
	locReq := CreateLocalizationRequest{
		Key:      "test.key",
		Language: "en",
		Value:    "Test value",
		// PluralForms and Variables are nil by default
	}
	jsonData, err := json.Marshal(locReq)
	require.NoError(t, err)
	
	var locUnmarshaled CreateLocalizationRequest
	err = json.Unmarshal(jsonData, &locUnmarshaled)
	require.NoError(t, err)
	assert.Empty(t, locUnmarshaled.PluralForms)
	assert.Empty(t, locUnmarshaled.Variables)
	
	// GetBatchLocalizationRequest with nil keys
	batchReq := GetBatchLocalizationRequest{
		Keys:     nil, // Explicitly nil
		Language: "en",
		Fallback: true,
	}
	jsonData, err = json.Marshal(batchReq)
	require.NoError(t, err)
	
	var batchUnmarshaled GetBatchLocalizationRequest
	err = json.Unmarshal(jsonData, &batchUnmarshaled)
	require.NoError(t, err)
	assert.Empty(t, batchUnmarshaled.Keys)
}

func TestAllRequestModels_WithEmptyCollections(t *testing.T) {
	// Test that models can handle empty slices and maps
	
	// CreateLocalizationRequest with empty collections
	locReq := CreateLocalizationRequest{
		Key:         "test.key",
		Language:    "en",
		Value:       "Test value",
		PluralForms: map[string]string{}, // Empty map
		Variables:   []string{},        // Empty slice
	}
	jsonData, err := json.Marshal(locReq)
	require.NoError(t, err)
	
	var locUnmarshaled CreateLocalizationRequest
	err = json.Unmarshal(jsonData, &locUnmarshaled)
	require.NoError(t, err)
	assert.Empty(t, locUnmarshaled.PluralForms)
	assert.Empty(t, locUnmarshaled.Variables)
	assert.NotNil(t, locUnmarshaled.PluralForms) // Should be initialized map, not nil
	assert.NotNil(t, locUnmarshaled.Variables)   // Should be initialized slice, not nil
}
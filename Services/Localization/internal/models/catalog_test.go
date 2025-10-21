package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLocalizationCatalog_Validate(t *testing.T) {
	tests := []struct {
		name    string
		catalog LocalizationCatalog
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid catalog",
			catalog: LocalizationCatalog{
				LanguageID:  "lang-123",
				CatalogData: json.RawMessage(`{"key1": "value1"}`),
			},
			wantErr: false,
		},
		{
			name: "missing language_id",
			catalog: LocalizationCatalog{
				CatalogData: json.RawMessage(`{"key1": "value1"}`),
			},
			wantErr: true,
			errMsg:  "language_id is required",
		},
		{
			name: "missing catalog_data",
			catalog: LocalizationCatalog{
				LanguageID: "lang-123",
			},
			wantErr: true,
			errMsg:  "catalog_data is required",
		},
		{
			name: "empty catalog_data",
			catalog: LocalizationCatalog{
				LanguageID:  "lang-123",
				CatalogData: json.RawMessage(``),
			},
			wantErr: true,
			errMsg:  "catalog_data is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.catalog.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLocalizationCatalog_BeforeCreate(t *testing.T) {
	catalog := LocalizationCatalog{
		LanguageID:  "lang-123",
		CatalogData: json.RawMessage(`{"key1": "value1", "key2": "value2"}`),
	}

	catalog.BeforeCreate()

	assert.NotEmpty(t, catalog.ID)
	assert.NotZero(t, catalog.CreatedAt)
	assert.NotZero(t, catalog.ModifiedAt)
	assert.NotEmpty(t, catalog.Checksum)
	assert.Len(t, catalog.Checksum, 64) // SHA-256 produces 64 hex characters
}

func TestLocalizationCatalog_BeforeUpdate(t *testing.T) {
	catalog := LocalizationCatalog{
		LanguageID:  "lang-123",
		CatalogData: json.RawMessage(`{"key1": "value1"}`),
		CreatedAt:   1000,
		ModifiedAt:  1000,
	}

	catalog.BeforeUpdate()

	assert.NotZero(t, catalog.ModifiedAt)
	assert.Greater(t, catalog.ModifiedAt, catalog.CreatedAt)
	assert.NotEmpty(t, catalog.Checksum)
}

func TestLocalizationCatalog_GenerateChecksum(t *testing.T) {
	catalog1 := LocalizationCatalog{
		CatalogData: json.RawMessage(`{"key1": "value1"}`),
	}
	catalog2 := LocalizationCatalog{
		CatalogData: json.RawMessage(`{"key1": "value1"}`),
	}
	catalog3 := LocalizationCatalog{
		CatalogData: json.RawMessage(`{"key1": "value2"}`),
	}

	catalog1.GenerateChecksum()
	catalog2.GenerateChecksum()
	catalog3.GenerateChecksum()

	// Same data should produce same checksum
	assert.Equal(t, catalog1.Checksum, catalog2.Checksum)
	// Different data should produce different checksum
	assert.NotEqual(t, catalog1.Checksum, catalog3.Checksum)
	// Checksum should be 64 characters (SHA-256)
	assert.Len(t, catalog1.Checksum, 64)
}

func TestLocalizationCatalog_GetCatalogMap(t *testing.T) {
	catalog := LocalizationCatalog{
		CatalogData: json.RawMessage(`{"key1": "value1", "key2": "value2"}`),
	}

	catalogMap, err := catalog.GetCatalogMap()

	assert.NoError(t, err)
	assert.Len(t, catalogMap, 2)
	assert.Equal(t, "value1", catalogMap["key1"])
	assert.Equal(t, "value2", catalogMap["key2"])
}

func TestLocalizationCatalog_GetCatalogMap_InvalidJSON(t *testing.T) {
	catalog := LocalizationCatalog{
		CatalogData: json.RawMessage(`{invalid json}`),
	}

	catalogMap, err := catalog.GetCatalogMap()

	assert.Error(t, err)
	assert.Nil(t, catalogMap)
}

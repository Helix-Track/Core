package models

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"
)

// LocalizationCatalog represents a pre-built catalog
type LocalizationCatalog struct {
	ID          string          `json:"id" db:"id"`
	LanguageID  string          `json:"language_id" db:"language_id"`
	Category    string          `json:"category,omitempty" db:"category"`
	CatalogData json.RawMessage `json:"catalog_data" db:"catalog_data"`
	Version     int             `json:"version" db:"version"`
	Checksum    string          `json:"checksum" db:"checksum"`
	CreatedAt   int64           `json:"created_at" db:"created_at"`
	ModifiedAt  int64           `json:"modified_at" db:"modified_at"`
}

// CatalogResponse represents the API response for catalog retrieval
type CatalogResponse struct {
	Language string            `json:"language"`
	Version  int               `json:"version"`
	Checksum string            `json:"checksum"`
	Catalog  map[string]string `json:"catalog"`
}

// Validate validates the catalog model
func (lc *LocalizationCatalog) Validate() error {
	if lc.LanguageID == "" {
		return ErrValidationFailed("language_id is required")
	}
	if lc.CatalogData == nil || len(lc.CatalogData) == 0 {
		return ErrValidationFailed("catalog_data is required")
	}
	if lc.Version < 1 {
		lc.Version = 1
	}
	return nil
}

// BeforeCreate sets timestamps before creation
func (lc *LocalizationCatalog) BeforeCreate() {
	now := time.Now().Unix()
	lc.CreatedAt = now
	lc.ModifiedAt = now
	if lc.ID == "" {
		lc.ID = GenerateUUID()
	}
	if lc.Checksum == "" {
		lc.GenerateChecksum()
	}
}

// BeforeUpdate sets modified timestamp before update
func (lc *LocalizationCatalog) BeforeUpdate() {
	lc.ModifiedAt = time.Now().Unix()
	lc.GenerateChecksum()
}

// GenerateChecksum generates SHA-256 checksum of catalog data
func (lc *LocalizationCatalog) GenerateChecksum() {
	hash := sha256.Sum256(lc.CatalogData)
	lc.Checksum = hex.EncodeToString(hash[:])
}

// GetCatalogMap returns the catalog data as a map
func (lc *LocalizationCatalog) GetCatalogMap() (map[string]string, error) {
	var catalog map[string]string
	if err := json.Unmarshal(lc.CatalogData, &catalog); err != nil {
		return nil, err
	}
	return catalog, nil
}

package models

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// LocalizationVersion represents a version of the localization catalog
type LocalizationVersion struct {
	ID                  string `json:"id" db:"id"`
	VersionNumber       string `json:"version_number" db:"version_number"`
	VersionType         string `json:"version_type" db:"version_type"` // "major", "minor", "patch"
	Description         string `json:"description" db:"description"`
	KeysCount           int    `json:"keys_count" db:"keys_count"`
	LanguagesCount      int    `json:"languages_count" db:"languages_count"`
	TotalLocalizations  int    `json:"total_localizations" db:"total_localizations"`
	CreatedBy           string `json:"created_by" db:"created_by"`
	CreatedAt           int64  `json:"created_at" db:"created_at"`
	Metadata            string `json:"metadata,omitempty" db:"metadata"` // JSONB
}

// VersionInfo represents simplified version information
type VersionInfo struct {
	Version        string `json:"version"`
	KeysCount      int    `json:"keys_count"`
	LanguagesCount int    `json:"languages_count"`
	LastUpdated    int64  `json:"last_updated"`
}

// VersionHistoryResponse represents a list of versions
type VersionHistoryResponse struct {
	Versions      []LocalizationVersion `json:"versions"`
	TotalVersions int                   `json:"total_versions"`
	CurrentVersion string                `json:"current_version"`
}

// CreateVersionRequest represents a request to create a new version
type CreateVersionRequest struct {
	VersionType string                 `json:"version_type"` // "major", "minor", "patch"
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// BeforeCreate sets default values before creating a version
func (v *LocalizationVersion) BeforeCreate() {
	now := time.Now().Unix()
	if v.CreatedAt == 0 {
		v.CreatedAt = now
	}
}

// Validate validates a LocalizationVersion
func (v *LocalizationVersion) Validate() error {
	if v.VersionNumber == "" {
		return fmt.Errorf("version_number is required")
	}

	if !isValidVersionNumber(v.VersionNumber) {
		return fmt.Errorf("invalid version_number format: must be X.Y.Z")
	}

	if v.VersionType != "major" && v.VersionType != "minor" && v.VersionType != "patch" {
		return fmt.Errorf("version_type must be 'major', 'minor', or 'patch'")
	}

	if v.KeysCount < 0 {
		return fmt.Errorf("keys_count cannot be negative")
	}

	if v.LanguagesCount < 0 {
		return fmt.Errorf("languages_count cannot be negative")
	}

	if v.TotalLocalizations < 0 {
		return fmt.Errorf("total_localizations cannot be negative")
	}

	return nil
}

// Validate validates a CreateVersionRequest
func (cvr *CreateVersionRequest) Validate() error {
	if cvr.VersionType != "major" && cvr.VersionType != "minor" && cvr.VersionType != "patch" {
		return fmt.Errorf("version_type must be 'major', 'minor', or 'patch'")
	}

	if cvr.Description == "" {
		return fmt.Errorf("description is required")
	}

	return nil
}

// isValidVersionNumber checks if a version number is valid (X.Y.Z format)
func isValidVersionNumber(version string) bool {
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return false
	}

	for _, part := range parts {
		if _, err := strconv.Atoi(part); err != nil {
			return false
		}
	}

	return true
}

// ParseVersion parses a version string into major, minor, patch components
func ParseVersion(version string) (major, minor, patch int, err error) {
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return 0, 0, 0, fmt.Errorf("invalid version format")
	}

	major, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, 0, err
	}

	minor, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, 0, err
	}

	patch, err = strconv.Atoi(parts[2])
	if err != nil {
		return 0, 0, 0, err
	}

	return major, minor, patch, nil
}

// IncrementVersion increments a version number based on type
func IncrementVersion(currentVersion string, versionType string) (string, error) {
	major, minor, patch, err := ParseVersion(currentVersion)
	if err != nil {
		return "", err
	}

	switch versionType {
	case "major":
		major++
		minor = 0
		patch = 0
	case "minor":
		minor++
		patch = 0
	case "patch":
		patch++
	default:
		return "", fmt.Errorf("invalid version type: %s", versionType)
	}

	return fmt.Sprintf("%d.%d.%d", major, minor, patch), nil
}

// CompareVersions compares two version strings
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func CompareVersions(v1, v2 string) (int, error) {
	major1, minor1, patch1, err := ParseVersion(v1)
	if err != nil {
		return 0, err
	}

	major2, minor2, patch2, err := ParseVersion(v2)
	if err != nil {
		return 0, err
	}

	if major1 != major2 {
		if major1 < major2 {
			return -1, nil
		}
		return 1, nil
	}

	if minor1 != minor2 {
		if minor1 < minor2 {
			return -1, nil
		}
		return 1, nil
	}

	if patch1 != patch2 {
		if patch1 < patch2 {
			return -1, nil
		}
		return 1, nil
	}

	return 0, nil
}

// ToJSON converts a LocalizationVersion to JSON
func (v *LocalizationVersion) ToJSON() ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}

// GetMetadataMap returns metadata as a map
func (v *LocalizationVersion) GetMetadataMap() (map[string]interface{}, error) {
	if v.Metadata == "" {
		return make(map[string]interface{}), nil
	}

	var metadata map[string]interface{}
	err := json.Unmarshal([]byte(v.Metadata), &metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	return metadata, nil
}

// SetMetadataMap sets metadata from a map
func (v *LocalizationVersion) SetMetadataMap(metadata map[string]interface{}) error {
	if metadata == nil {
		v.Metadata = ""
		return nil
	}

	jsonData, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	v.Metadata = string(jsonData)
	return nil
}

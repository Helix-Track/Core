package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalizationVersion_BeforeCreate(t *testing.T) {
	v := &LocalizationVersion{}
	v.BeforeCreate()
	
	assert.NotZero(t, v.CreatedAt, "CreatedAt should be set")
	assert.True(t, v.CreatedAt <= time.Now().Unix(), "CreatedAt should be in the past")
	
	// Test that existing CreatedAt is not overwritten
	existingTime := int64(1234567890)
	v = &LocalizationVersion{CreatedAt: existingTime}
	v.BeforeCreate()
	assert.Equal(t, existingTime, v.CreatedAt, "Existing CreatedAt should not be overwritten")
}

func TestLocalizationVersion_Validate(t *testing.T) {
	tests := []struct {
		name        string
		version     LocalizationVersion
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid version",
			version: LocalizationVersion{
				VersionNumber:       "1.0.0",
				VersionType:        "major",
				Description:        "Initial version",
				KeysCount:          10,
				LanguagesCount:     2,
				TotalLocalizations: 20,
				CreatedBy:          "test-user",
				CreatedAt:          time.Now().Unix(),
			},
			expectError: false,
		},
		{
			name: "valid minor version",
			version: LocalizationVersion{
				VersionNumber:       "1.1.0",
				VersionType:        "minor",
				Description:        "Feature update",
				KeysCount:          15,
				LanguagesCount:     2,
				TotalLocalizations: 30,
				CreatedBy:          "test-user",
			},
			expectError: false,
		},
		{
			name: "valid patch version",
			version: LocalizationVersion{
				VersionNumber:       "1.1.1",
				VersionType:        "patch",
				Description:        "Bug fix",
				KeysCount:          15,
				LanguagesCount:     2,
				TotalLocalizations: 30,
				CreatedBy:          "test-user",
			},
			expectError: false,
		},
		{
			name: "missing version number",
			version: LocalizationVersion{
				VersionType: "major",
				Description: "Missing version",
			},
			expectError: true,
			errorMsg:    "version_number is required",
		},
		{
			name: "invalid version format",
			version: LocalizationVersion{
				VersionNumber: "1.0",
				VersionType:  "major",
			},
			expectError: true,
			errorMsg:    "invalid version_number format: must be X.Y.Z",
		},
		{
			name: "invalid version type",
			version: LocalizationVersion{
				VersionNumber: "1.0.0",
				VersionType:  "invalid",
			},
			expectError: true,
			errorMsg:    "version_type must be 'major', 'minor', or 'patch'",
		},
		{
			name: "negative keys count",
			version: LocalizationVersion{
				VersionNumber: "1.0.0",
				VersionType:  "major",
				KeysCount:    -1,
			},
			expectError: true,
			errorMsg:    "keys_count cannot be negative",
		},
		{
			name: "negative languages count",
			version: LocalizationVersion{
				VersionNumber:   "1.0.0",
				VersionType:    "major",
				LanguagesCount: -1,
			},
			expectError: true,
			errorMsg:    "languages_count cannot be negative",
		},
		{
			name: "negative total localizations",
			version: LocalizationVersion{
				VersionNumber:      "1.0.0",
				VersionType:       "major",
				TotalLocalizations: -1,
			},
			expectError: true,
			errorMsg:    "total_localizations cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.version.Validate()
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

func TestCreateVersionRequest_Validate(t *testing.T) {
	tests := []struct {
		name        string
		req         CreateVersionRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid major version",
			req: CreateVersionRequest{
				VersionType: "major",
				Description: "Major release with breaking changes",
				Metadata: map[string]interface{}{
					"breaking_changes": true,
				},
			},
			expectError: false,
		},
		{
			name: "valid minor version",
			req: CreateVersionRequest{
				VersionType: "minor",
				Description: "New features added",
				Metadata:    nil,
			},
			expectError: false,
		},
		{
			name: "valid patch version",
			req: CreateVersionRequest{
				VersionType: "patch",
				Description: "Bug fixes",
				Metadata:    map[string]interface{}{},
			},
			expectError: false,
		},
		{
			name: "invalid version type",
			req: CreateVersionRequest{
				VersionType: "invalid",
				Description: "Invalid version",
			},
			expectError: true,
			errorMsg:    "version_type must be 'major', 'minor', or 'patch'",
		},
		{
			name: "missing description",
			req: CreateVersionRequest{
				VersionType: "major",
			},
			expectError: true,
			errorMsg:    "description is required",
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

func TestIsValidVersionNumber(t *testing.T) {
	tests := []struct {
		version     string
		expectValid bool
	}{
		{"1.0.0", true},
		{"10.5.3", true},
		{"0.0.1", true},
		{"1.0", false},
		{"1", false},
		{"1.0.0.0", false},
		{"1.0.0-beta", false},
		{"", false},
		{"a.b.c", false},
		{"1.2.3.4", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			result := isValidVersionNumber(tt.version)
			assert.Equal(t, tt.expectValid, result)
		})
	}
}

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name         string
		version      string
		expectedMaj  int
		expectedMin  int
		expectedPat  int
		expectError  bool
	}{
		{
			name:         "valid version",
			version:      "1.2.3",
			expectedMaj:  1,
			expectedMin:  2,
			expectedPat:  3,
			expectError:  false,
		},
		{
			name:         "version with zeros",
			version:      "0.0.0",
			expectedMaj:  0,
			expectedMin:  0,
			expectedPat:  0,
			expectError:  false,
		},
		{
			name:         "large version numbers",
			version:      "100.200.300",
			expectedMaj:  100,
			expectedMin:  200,
			expectedPat:  300,
			expectError:  false,
		},
		{
			name:        "missing patch",
			version:     "1.2",
			expectError: true,
		},
		{
			name:        "too many parts",
			version:     "1.2.3.4",
			expectError: true,
		},
		{
			name:        "non-numeric",
			version:     "1.a.3",
			expectError: true,
		},
		{
			name:        "empty string",
			version:     "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			major, minor, patch, err := ParseVersion(tt.version)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedMaj, major)
				assert.Equal(t, tt.expectedMin, minor)
				assert.Equal(t, tt.expectedPat, patch)
			}
		})
	}
}

func TestIncrementVersion(t *testing.T) {
	tests := []struct {
		name          string
		currentVer    string
		versionType   string
		expectedVer   string
		expectError   bool
	}{
		{
			name:        "increment major",
			currentVer:  "1.2.3",
			versionType: "major",
			expectedVer: "2.0.0",
			expectError: false,
		},
		{
			name:        "increment minor",
			currentVer:  "1.2.3",
			versionType: "minor",
			expectedVer: "1.3.0",
			expectError: false,
		},
		{
			name:        "increment patch",
			currentVer:  "1.2.3",
			versionType: "patch",
			expectedVer: "1.2.4",
			expectError: false,
		},
		{
			name:        "invalid version type",
			currentVer:  "1.2.3",
			versionType: "invalid",
			expectError: true,
		},
		{
			name:        "invalid current version",
			currentVer:  "1.2",
			versionType: "patch",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newVersion, err := IncrementVersion(tt.currentVer, tt.versionType)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedVer, newVersion)
			}
		})
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name       string
		v1         string
		v2         string
		expected   int
		expectErr  bool
	}{
		{
			name:     "v1 less than v2 - major",
			v1:       "1.0.0",
			v2:       "2.0.0",
			expected: -1,
		},
		{
			name:     "v1 greater than v2 - major",
			v1:       "3.0.0",
			v2:       "2.0.0",
			expected: 1,
		},
		{
			name:     "v1 less than v2 - minor",
			v1:       "1.0.0",
			v2:       "1.1.0",
			expected: -1,
		},
		{
			name:     "v1 greater than v2 - minor",
			v1:       "1.2.0",
			v2:       "1.1.0",
			expected: 1,
		},
		{
			name:     "v1 less than v2 - patch",
			v1:       "1.1.0",
			v2:       "1.1.1",
			expected: -1,
		},
		{
			name:     "v1 greater than v2 - patch",
			v1:       "1.1.2",
			v2:       "1.1.1",
			expected: 1,
		},
		{
			name:     "versions equal",
			v1:       "1.1.1",
			v2:       "1.1.1",
			expected: 0,
		},
		{
			name:      "invalid v1",
			v1:        "1.0",
			v2:        "1.0.0",
			expectErr: true,
		},
		{
			name:      "invalid v2",
			v1:        "1.0.0",
			v2:        "1.0",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CompareVersions(tt.v1, tt.v2)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestLocalizationVersion_ToJSON(t *testing.T) {
	v := &LocalizationVersion{
		ID:                 "test-version",
		VersionNumber:       "1.0.0",
		VersionType:         "major",
		Description:         "Test version",
		KeysCount:           10,
		LanguagesCount:      2,
		TotalLocalizations:  20,
		CreatedBy:           "test-user",
		CreatedAt:           time.Now().Unix(),
		Metadata:           `{"test": "value"}`,
	}

	jsonBytes, err := v.ToJSON()
	require.NoError(t, err)
	require.True(t, len(jsonBytes) > 0)

	// Verify it's valid JSON
	var parsed LocalizationVersion
	err = json.Unmarshal(jsonBytes, &parsed)
	require.NoError(t, err)
	assert.Equal(t, v.ID, parsed.ID)
	assert.Equal(t, v.VersionNumber, parsed.VersionNumber)
	assert.Equal(t, v.VersionType, parsed.VersionType)
}

func TestLocalizationVersion_MetadataOperations(t *testing.T) {
	v := &LocalizationVersion{}

	// Test getting empty metadata
	metadata, err := v.GetMetadataMap()
	require.NoError(t, err)
	assert.Empty(t, metadata)

	// Test setting nil metadata
	err = v.SetMetadataMap(nil)
	require.NoError(t, err)
	assert.Equal(t, "", v.Metadata)

	// Test setting empty metadata
	err = v.SetMetadataMap(map[string]interface{}{})
	require.NoError(t, err)
	assert.Equal(t, "{}", v.Metadata) // Empty map serializes to "{}"

	// Test setting and getting metadata
	testMetadata := map[string]interface{}{
		"test_key":   "test_value",
		"test_int":   42,
		"test_bool":  true,
		"nested": map[string]interface{}{
			"inner_key": "inner_value",
		},
	}
	err = v.SetMetadataMap(testMetadata)
	require.NoError(t, err)
	assert.NotEmpty(t, v.Metadata)

	// Get the metadata back
	retrieved, err := v.GetMetadataMap()
	require.NoError(t, err)
	assert.Equal(t, "test_value", retrieved["test_key"])
	assert.Equal(t, float64(42), retrieved["test_int"]) // JSON numbers are float64
	assert.Equal(t, true, retrieved["test_bool"])
	assert.NotNil(t, retrieved["nested"])
}

func TestLocalizationVersion_JSONMarshaling(t *testing.T) {
	// Test that the model can be marshaled and unmarshaled correctly
	original := &LocalizationVersion{
		ID:                 "version-123",
		VersionNumber:       "2.1.5",
		VersionType:         "minor",
		Description:         "Added new features",
		KeysCount:           25,
		LanguagesCount:      5,
		TotalLocalizations:  125,
		CreatedBy:           "john.doe",
		CreatedAt:           time.Now().Unix(),
		Metadata:           `{"release_notes": "Improved performance"}`,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(original)
	require.NoError(t, err)

	// Unmarshal back
	var unmarshaled LocalizationVersion
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	// Check all fields
	assert.Equal(t, original.ID, unmarshaled.ID)
	assert.Equal(t, original.VersionNumber, unmarshaled.VersionNumber)
	assert.Equal(t, original.VersionType, unmarshaled.VersionType)
	assert.Equal(t, original.Description, unmarshaled.Description)
	assert.Equal(t, original.KeysCount, unmarshaled.KeysCount)
	assert.Equal(t, original.LanguagesCount, unmarshaled.LanguagesCount)
	assert.Equal(t, original.TotalLocalizations, unmarshaled.TotalLocalizations)
	assert.Equal(t, original.CreatedBy, unmarshaled.CreatedBy)
	assert.Equal(t, original.CreatedAt, unmarshaled.CreatedAt)
	assert.Equal(t, original.Metadata, unmarshaled.Metadata)
}

func TestVersionInfo(t *testing.T) {
	info := VersionInfo{
		Version:        "1.2.3",
		KeysCount:      50,
		LanguagesCount: 5,
		LastUpdated:    time.Now().Unix(),
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(info)
	require.NoError(t, err)

	var unmarshaled VersionInfo
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, info.Version, unmarshaled.Version)
	assert.Equal(t, info.KeysCount, unmarshaled.KeysCount)
	assert.Equal(t, info.LanguagesCount, unmarshaled.LanguagesCount)
	assert.Equal(t, info.LastUpdated, unmarshaled.LastUpdated)
}

func TestVersionHistoryResponse(t *testing.T) {
	response := VersionHistoryResponse{
		Versions: []LocalizationVersion{
			{
				ID:           "v1",
				VersionNumber: "1.0.0",
				VersionType:   "major",
			},
			{
				ID:           "v2",
				VersionNumber: "1.1.0",
				VersionType:   "minor",
			},
		},
		TotalVersions:  2,
		CurrentVersion: "1.1.0",
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(response)
	require.NoError(t, err)

	var unmarshaled VersionHistoryResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, response.TotalVersions, unmarshaled.TotalVersions)
	assert.Equal(t, response.CurrentVersion, unmarshaled.CurrentVersion)
	assert.Len(t, unmarshaled.Versions, 2)
	assert.Equal(t, response.Versions[0].ID, unmarshaled.Versions[0].ID)
	assert.Equal(t, response.Versions[1].VersionNumber, unmarshaled.Versions[1].VersionNumber)
}
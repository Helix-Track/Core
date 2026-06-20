package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// SpaceConfig represents per-space configuration stored in <space-root>/config.json.
type SpaceConfig struct {
	SchemaVersion int    `json:"schema_version"`
	SpaceID       string `json:"space_id"`
	Title         string `json:"title"`
	Description   string `json:"description,omitempty"`
	CoreEndpoint  string `json:"core_endpoint"`
	Database      struct {
		Path string `json:"path"`
		Type string `json:"type"`
	} `json:"database"`
	AssetsPath string `json:"assets_path"`
}

// defaultSpaceConfig returns a SpaceConfig with sensible defaults for a fresh space.
func defaultSpaceConfig(spaceRoot string) *SpaceConfig {
	cfg := &SpaceConfig{
		SchemaVersion: 1,
		SpaceID:       filepath.Base(spaceRoot),
		Title:         "HelixTrack Space -- " + filepath.Base(spaceRoot),
		Description:   "",
		CoreEndpoint:  "http://localhost:8080",
		AssetsPath:    filepath.Join(spaceRoot, "assets"),
	}
	cfg.Database.Path = filepath.Join(spaceRoot, "data", "helixtrack.db")
	cfg.Database.Type = "sqlite"
	return cfg
}

// LoadSpaceConfig reads <space-root>/config.json.
// If the file does NOT exist (empty space / new project), it creates one
// with default values and returns it -- the onboarding wizard auto-initialization.
func LoadSpaceConfig(spaceRoot string) (*SpaceConfig, error) {
	configPath := filepath.Join(spaceRoot, "config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// New/empty space -- auto-create config.json (onboarding wizard)
			if mkErr := os.MkdirAll(spaceRoot, 0755); mkErr != nil {
				return nil, fmt.Errorf("failed to create space root directory %s: %w", spaceRoot, mkErr)
			}
			cfg := defaultSpaceConfig(spaceRoot)
			jsonData, jsonErr := json.MarshalIndent(cfg, "", "  ")
			if jsonErr != nil {
				return nil, fmt.Errorf("failed to marshal default space config: %w", jsonErr)
			}
			if writeErr := os.WriteFile(configPath, jsonData, 0644); writeErr != nil {
				return nil, fmt.Errorf("failed to write default space config: %w", writeErr)
			}
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to read space config: %w", err)
	}

	var cfg SpaceConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse space config: %w", err)
	}

	return &cfg, nil
}

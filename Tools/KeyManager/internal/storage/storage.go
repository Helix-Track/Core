package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/helixtrack/keymanager/internal/generator"
	"gopkg.in/yaml.v3"
)

const (
	defaultStorageDir = "keys"
	metadataFile      = "keys.json"
)

// Storage handles key persistence
type Storage struct {
	storageDir string
}

// New creates a new storage instance
func New() (*Storage, error) {
	storageDir := defaultStorageDir

	// Create storage directory if it doesn't exist
	if err := os.MkdirAll(storageDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &Storage{
		storageDir: storageDir,
	}, nil
}

// SaveKey persists a key to storage
func (s *Storage) SaveKey(key *generator.Key) error {
	// Load existing keys
	keys, err := s.loadMetadata()
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to load metadata: %w", err)
	}

	// Add or update key
	found := false
	for i, k := range keys {
		if k.Name == key.Name && k.Service == key.Service {
			keys[i] = key
			found = true
			break
		}
	}
	if !found {
		keys = append(keys, key)
	}

	// Save metadata
	if err := s.saveMetadata(keys); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	// For non-TLS keys, save the value to a separate file
	if key.Type != generator.KeyTypeTLS {
		keyFile := filepath.Join(s.storageDir, key.Service, fmt.Sprintf("%s.key", key.Name))
		if err := os.MkdirAll(filepath.Dir(keyFile), 0700); err != nil {
			return fmt.Errorf("failed to create key directory: %w", err)
		}

		if err := os.WriteFile(keyFile, []byte(key.Value), 0600); err != nil {
			return fmt.Errorf("failed to write key file: %w", err)
		}
	}

	return nil
}

// GetKey retrieves a key by name and service
func (s *Storage) GetKey(name, service string) (*generator.Key, error) {
	keys, err := s.loadMetadata()
	if err != nil {
		return nil, fmt.Errorf("failed to load metadata: %w", err)
	}

	for _, key := range keys {
		if key.Name == name && key.Service == service {
			return key, nil
		}
	}

	return nil, fmt.Errorf("key not found: %s/%s", service, name)
}

// ListKeys returns all stored keys
func (s *Storage) ListKeys() ([]*generator.Key, error) {
	keys, err := s.loadMetadata()
	if err != nil {
		if os.IsNotExist(err) {
			return []*generator.Key{}, nil
		}
		return nil, fmt.Errorf("failed to load metadata: %w", err)
	}

	// Sort by service, then name
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].Service == keys[j].Service {
			return keys[i].Name < keys[j].Name
		}
		return keys[i].Service < keys[j].Service
	})

	return keys, nil
}

// DeleteKey removes a key from storage
func (s *Storage) DeleteKey(name, service string) error {
	keys, err := s.loadMetadata()
	if err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}

	// Find and remove key
	found := false
	for i, key := range keys {
		if key.Name == name && key.Service == service {
			keys = append(keys[:i], keys[i+1:]...)
			found = true

			// Delete key file
			if key.Type != generator.KeyTypeTLS {
				keyFile := filepath.Join(s.storageDir, service, fmt.Sprintf("%s.key", name))
				os.Remove(keyFile)
			} else {
				// Delete TLS cert and key files
				if certPath, ok := key.Metadata["cert_path"]; ok {
					os.Remove(certPath)
				}
				if keyPath, ok := key.Metadata["key_path"]; ok {
					os.Remove(keyPath)
				}
			}
			break
		}
	}

	if !found {
		return fmt.Errorf("key not found: %s/%s", service, name)
	}

	// Save updated metadata
	if err := s.saveMetadata(keys); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	return nil
}

// ExportKeys exports all keys to a file
func (s *Storage) ExportKeys(outputPath, format string) error {
	keys, err := s.ListKeys()
	if err != nil {
		return fmt.Errorf("failed to list keys: %w", err)
	}

	// Create output directory
	if err := os.MkdirAll(filepath.Dir(outputPath), 0700); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	var data []byte
	switch format {
	case "json":
		data, err = json.MarshalIndent(keys, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}

	case "yaml":
		data, err = yaml.Marshal(keys)
		if err != nil {
			return fmt.Errorf("failed to marshal YAML: %w", err)
		}

	case "env":
		var envLines []string
		for _, key := range keys {
			varName := strings.ToUpper(fmt.Sprintf("%s_%s", key.Service, key.Name))
			varName = strings.ReplaceAll(varName, "-", "_")
			if key.Type == generator.KeyTypeTLS {
				envLines = append(envLines, fmt.Sprintf("%s_CERT=%s", varName, key.Metadata["cert_path"]))
				envLines = append(envLines, fmt.Sprintf("%s_KEY=%s", varName, key.Metadata["key_path"]))
			} else {
				envLines = append(envLines, fmt.Sprintf("%s=%s", varName, key.Value))
			}
		}
		data = []byte(strings.Join(envLines, "\n"))

	default:
		return fmt.Errorf("unsupported export format: %s (supported: json, yaml, env)", format)
	}

	if err := os.WriteFile(outputPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write export file: %w", err)
	}

	return nil
}

// ImportKeys imports keys from a file
func (s *Storage) ImportKeys(inputPath string) (int, error) {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return 0, fmt.Errorf("failed to read import file: %w", err)
	}

	var keys []*generator.Key

	// Try JSON first
	if err := json.Unmarshal(data, &keys); err != nil {
		// Try YAML
		if err := yaml.Unmarshal(data, &keys); err != nil {
			return 0, fmt.Errorf("failed to parse import file (tried JSON and YAML): %w", err)
		}
	}

	// Save each key
	for _, key := range keys {
		if err := s.SaveKey(key); err != nil {
			return 0, fmt.Errorf("failed to save imported key %s/%s: %w", key.Service, key.Name, err)
		}
	}

	return len(keys), nil
}

// ExportKeyToFile exports a single key to a file
func (s *Storage) ExportKeyToFile(key *generator.Key, outputPath string) error {
	// Create output directory
	if err := os.MkdirAll(filepath.Dir(outputPath), 0700); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	data, err := json.MarshalIndent(key, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal key: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// loadMetadata loads key metadata from storage
func (s *Storage) loadMetadata() ([]*generator.Key, error) {
	metadataPath := filepath.Join(s.storageDir, metadataFile)

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, err
	}

	var keys []*generator.Key
	if err := json.Unmarshal(data, &keys); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	return keys, nil
}

// saveMetadata saves key metadata to storage
func (s *Storage) saveMetadata(keys []*generator.Key) error {
	metadataPath := filepath.Join(s.storageDir, metadataFile)

	data, err := json.MarshalIndent(keys, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	return nil
}

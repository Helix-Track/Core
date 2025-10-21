package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/helixtrack/keymanager/internal/generator"
	"gopkg.in/yaml.v3"
)

func TestNew(t *testing.T) {
	// Use temp directory
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	storage, err := New()
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	if storage == nil {
		t.Fatal("Expected storage to be created")
	}

	// Verify storage directory was created
	if _, err := os.Stat(defaultStorageDir); os.IsNotExist(err) {
		t.Error("Storage directory was not created")
	}
}

func TestSaveKey(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	storage, _ := New()

	key := &generator.Key{
		ID:        "test-id",
		Name:      "test-key",
		Service:   "test-service",
		Type:      generator.KeyTypeJWT,
		Value:     "test-value",
		Metadata:  map[string]string{"length": "64"},
		CreatedAt: time.Now(),
		Version:   1,
	}

	err := storage.SaveKey(key)
	if err != nil {
		t.Fatalf("Failed to save key: %v", err)
	}

	// Verify key file was created
	keyFile := filepath.Join(defaultStorageDir, "test-service", "test-key.key")
	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		t.Error("Key file was not created")
	}

	// Verify metadata was saved
	metadataFile := filepath.Join(defaultStorageDir, "keys.json")
	if _, err := os.Stat(metadataFile); os.IsNotExist(err) {
		t.Error("Metadata file was not created")
	}

	// Verify key file content
	content, err := os.ReadFile(keyFile)
	if err != nil {
		t.Fatalf("Failed to read key file: %v", err)
	}
	if string(content) != "test-value" {
		t.Errorf("Expected key value 'test-value', got '%s'", string(content))
	}
}

func TestSaveKey_TLS(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	storage, _ := New()

	key := &generator.Key{
		ID:      "test-tls-id",
		Name:    "test-tls",
		Service: "test-service",
		Type:    generator.KeyTypeTLS,
		Value:   "", // TLS keys don't have value
		Metadata: map[string]string{
			"cert_path": "cert.pem",
			"key_path":  "key.pem",
		},
		CreatedAt: time.Now(),
		Version:   1,
	}

	err := storage.SaveKey(key)
	if err != nil {
		t.Fatalf("Failed to save TLS key: %v", err)
	}

	// TLS keys shouldn't create a .key file
	keyFile := filepath.Join(defaultStorageDir, "test-service", "test-tls.key")
	if _, err := os.Stat(keyFile); !os.IsNotExist(err) {
		t.Error("TLS key should not create a .key file")
	}

	// But metadata should still be saved
	metadataFile := filepath.Join(defaultStorageDir, "keys.json")
	if _, err := os.Stat(metadataFile); os.IsNotExist(err) {
		t.Error("Metadata file was not created")
	}
}

func TestSaveKey_Update(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	storage, _ := New()

	// Save initial key
	key1 := &generator.Key{
		ID:        "test-id-1",
		Name:      "test-key",
		Service:   "test-service",
		Type:      generator.KeyTypeJWT,
		Value:     "value1",
		Metadata:  map[string]string{},
		CreatedAt: time.Now(),
		Version:   1,
	}
	storage.SaveKey(key1)

	// Update key
	key2 := &generator.Key{
		ID:        "test-id-2",
		Name:      "test-key",
		Service:   "test-service",
		Type:      generator.KeyTypeJWT,
		Value:     "value2",
		Metadata:  map[string]string{},
		CreatedAt: time.Now(),
		Version:   2,
	}
	storage.SaveKey(key2)

	// Verify only one key exists in metadata
	keys, err := storage.ListKeys()
	if err != nil {
		t.Fatalf("Failed to list keys: %v", err)
	}
	if len(keys) != 1 {
		t.Errorf("Expected 1 key, got %d", len(keys))
	}
	if keys[0].Version != 2 {
		t.Errorf("Expected version 2, got %d", keys[0].Version)
	}
}

func TestGetKey(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	storage, _ := New()

	key := &generator.Key{
		ID:        "test-id",
		Name:      "test-key",
		Service:   "test-service",
		Type:      generator.KeyTypeJWT,
		Value:     "test-value",
		Metadata:  map[string]string{},
		CreatedAt: time.Now(),
		Version:   1,
	}
	storage.SaveKey(key)

	// Retrieve key
	retrieved, err := storage.GetKey("test-key", "test-service")
	if err != nil {
		t.Fatalf("Failed to get key: %v", err)
	}

	if retrieved.ID != key.ID {
		t.Errorf("Expected ID %s, got %s", key.ID, retrieved.ID)
	}
	if retrieved.Name != key.Name {
		t.Errorf("Expected name %s, got %s", key.Name, retrieved.Name)
	}
	if retrieved.Service != key.Service {
		t.Errorf("Expected service %s, got %s", key.Service, retrieved.Service)
	}
}

func TestGetKey_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	storage, _ := New()

	_, err := storage.GetKey("nonexistent", "service")
	if err == nil {
		t.Fatal("Expected error for nonexistent key")
	}
}

func TestListKeys(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	storage, _ := New()

	// Save multiple keys
	keys := []*generator.Key{
		{
			ID:        "id1",
			Name:      "key1",
			Service:   "service-a",
			Type:      generator.KeyTypeJWT,
			Value:     "value1",
			Metadata:  map[string]string{},
			CreatedAt: time.Now(),
			Version:   1,
		},
		{
			ID:        "id2",
			Name:      "key2",
			Service:   "service-b",
			Type:      generator.KeyTypeDB,
			Value:     "value2",
			Metadata:  map[string]string{},
			CreatedAt: time.Now(),
			Version:   1,
		},
		{
			ID:        "id3",
			Name:      "key3",
			Service:   "service-a",
			Type:      generator.KeyTypeAPI,
			Value:     "value3",
			Metadata:  map[string]string{},
			CreatedAt: time.Now(),
			Version:   1,
		},
	}

	for _, key := range keys {
		storage.SaveKey(key)
	}

	// List keys
	retrieved, err := storage.ListKeys()
	if err != nil {
		t.Fatalf("Failed to list keys: %v", err)
	}

	if len(retrieved) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(retrieved))
	}

	// Verify keys are sorted (service, then name)
	if retrieved[0].Service != "service-a" || retrieved[0].Name != "key1" {
		t.Error("Keys are not properly sorted")
	}
}

func TestListKeys_Empty(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	storage, _ := New()

	keys, err := storage.ListKeys()
	if err != nil {
		t.Fatalf("Failed to list keys: %v", err)
	}

	if len(keys) != 0 {
		t.Errorf("Expected 0 keys, got %d", len(keys))
	}
}

func TestDeleteKey(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	storage, _ := New()

	key := &generator.Key{
		ID:        "test-id",
		Name:      "test-key",
		Service:   "test-service",
		Type:      generator.KeyTypeJWT,
		Value:     "test-value",
		Metadata:  map[string]string{},
		CreatedAt: time.Now(),
		Version:   1,
	}
	storage.SaveKey(key)

	// Delete key
	err := storage.DeleteKey("test-key", "test-service")
	if err != nil {
		t.Fatalf("Failed to delete key: %v", err)
	}

	// Verify key is gone
	_, err = storage.GetKey("test-key", "test-service")
	if err == nil {
		t.Error("Expected error for deleted key")
	}

	// Verify key file was deleted
	keyFile := filepath.Join(defaultStorageDir, "test-service", "test-key.key")
	if _, err := os.Stat(keyFile); !os.IsNotExist(err) {
		t.Error("Key file was not deleted")
	}
}

func TestDeleteKey_TLS(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	storage, _ := New()

	// Create dummy TLS files
	certPath := filepath.Join(tempDir, "cert.pem")
	keyPath := filepath.Join(tempDir, "key.pem")
	os.WriteFile(certPath, []byte("cert"), 0644)
	os.WriteFile(keyPath, []byte("key"), 0600)

	key := &generator.Key{
		ID:      "test-tls-id",
		Name:    "test-tls",
		Service: "test-service",
		Type:    generator.KeyTypeTLS,
		Value:   "",
		Metadata: map[string]string{
			"cert_path": certPath,
			"key_path":  keyPath,
		},
		CreatedAt: time.Now(),
		Version:   1,
	}
	storage.SaveKey(key)

	// Delete TLS key
	err := storage.DeleteKey("test-tls", "test-service")
	if err != nil {
		t.Fatalf("Failed to delete TLS key: %v", err)
	}

	// Verify TLS files were deleted
	if _, err := os.Stat(certPath); !os.IsNotExist(err) {
		t.Error("TLS cert file was not deleted")
	}
	if _, err := os.Stat(keyPath); !os.IsNotExist(err) {
		t.Error("TLS key file was not deleted")
	}
}

func TestDeleteKey_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	storage, _ := New()

	err := storage.DeleteKey("nonexistent", "service")
	if err == nil {
		t.Fatal("Expected error for nonexistent key")
	}
}

func TestExportKeys_JSON(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	storage, _ := New()

	key := &generator.Key{
		ID:        "test-id",
		Name:      "test-key",
		Service:   "test-service",
		Type:      generator.KeyTypeJWT,
		Value:     "test-value",
		Metadata:  map[string]string{},
		CreatedAt: time.Now(),
		Version:   1,
	}
	storage.SaveKey(key)

	// Export keys
	exportPath := filepath.Join(tempDir, "export.json")
	err := storage.ExportKeys(exportPath, "json")
	if err != nil {
		t.Fatalf("Failed to export keys: %v", err)
	}

	// Verify export file exists
	if _, err := os.Stat(exportPath); os.IsNotExist(err) {
		t.Error("Export file was not created")
	}

	// Verify export content
	content, _ := os.ReadFile(exportPath)
	var exportedKeys []*generator.Key
	if err := json.Unmarshal(content, &exportedKeys); err != nil {
		t.Fatalf("Failed to parse exported JSON: %v", err)
	}

	if len(exportedKeys) != 1 {
		t.Errorf("Expected 1 exported key, got %d", len(exportedKeys))
	}
}

func TestExportKeys_YAML(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	storage, _ := New()

	key := &generator.Key{
		ID:        "test-id",
		Name:      "test-key",
		Service:   "test-service",
		Type:      generator.KeyTypeJWT,
		Value:     "test-value",
		Metadata:  map[string]string{},
		CreatedAt: time.Now(),
		Version:   1,
	}
	storage.SaveKey(key)

	// Export keys
	exportPath := filepath.Join(tempDir, "export.yaml")
	err := storage.ExportKeys(exportPath, "yaml")
	if err != nil {
		t.Fatalf("Failed to export keys: %v", err)
	}

	// Verify export content
	content, _ := os.ReadFile(exportPath)
	var exportedKeys []*generator.Key
	if err := yaml.Unmarshal(content, &exportedKeys); err != nil {
		t.Fatalf("Failed to parse exported YAML: %v", err)
	}

	if len(exportedKeys) != 1 {
		t.Errorf("Expected 1 exported key, got %d", len(exportedKeys))
	}
}

func TestExportKeys_Env(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	storage, _ := New()

	keys := []*generator.Key{
		{
			ID:        "id1",
			Name:      "jwt-secret",
			Service:   "auth",
			Type:      generator.KeyTypeJWT,
			Value:     "secret123",
			Metadata:  map[string]string{},
			CreatedAt: time.Now(),
			Version:   1,
		},
		{
			ID:      "id2",
			Name:    "tls-cert",
			Service: "web",
			Type:    generator.KeyTypeTLS,
			Value:   "",
			Metadata: map[string]string{
				"cert_path": "/path/to/cert.pem",
				"key_path":  "/path/to/key.pem",
			},
			CreatedAt: time.Now(),
			Version:   1,
		},
	}

	for _, key := range keys {
		storage.SaveKey(key)
	}

	// Export keys
	exportPath := filepath.Join(tempDir, "export.env")
	err := storage.ExportKeys(exportPath, "env")
	if err != nil {
		t.Fatalf("Failed to export keys: %v", err)
	}

	// Verify export content
	content, _ := os.ReadFile(exportPath)
	envContent := string(content)

	if !contains(envContent, "AUTH_JWT_SECRET=secret123") {
		t.Error("Expected AUTH_JWT_SECRET in env export")
	}
	// Variable names are converted: service_name -> WEB_TLS_CERT
	if !contains(envContent, "CERT=/path/to/cert.pem") {
		t.Errorf("Expected certificate path in env export, got:\n%s", envContent)
	}
	if !contains(envContent, "KEY=/path/to/key.pem") {
		t.Errorf("Expected key path in env export, got:\n%s", envContent)
	}
}

func TestExportKeys_UnsupportedFormat(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	storage, _ := New()

	err := storage.ExportKeys("export.txt", "txt")
	if err == nil {
		t.Fatal("Expected error for unsupported format")
	}
}

func TestImportKeys(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	storage, _ := New()

	// Create import file
	keys := []*generator.Key{
		{
			ID:        "id1",
			Name:      "key1",
			Service:   "service1",
			Type:      generator.KeyTypeJWT,
			Value:     "value1",
			Metadata:  map[string]string{},
			CreatedAt: time.Now(),
			Version:   1,
		},
		{
			ID:        "id2",
			Name:      "key2",
			Service:   "service2",
			Type:      generator.KeyTypeAPI,
			Value:     "value2",
			Metadata:  map[string]string{},
			CreatedAt: time.Now(),
			Version:   1,
		},
	}

	importPath := filepath.Join(tempDir, "import.json")
	data, _ := json.MarshalIndent(keys, "", "  ")
	os.WriteFile(importPath, data, 0600)

	// Import keys
	count, err := storage.ImportKeys(importPath)
	if err != nil {
		t.Fatalf("Failed to import keys: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 imported keys, got %d", count)
	}

	// Verify keys were imported
	retrieved, _ := storage.ListKeys()
	if len(retrieved) != 2 {
		t.Errorf("Expected 2 keys in storage, got %d", len(retrieved))
	}
}

func TestImportKeys_YAML(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	storage, _ := New()

	// Create import file
	keys := []*generator.Key{
		{
			ID:        "id1",
			Name:      "key1",
			Service:   "service1",
			Type:      generator.KeyTypeJWT,
			Value:     "value1",
			Metadata:  map[string]string{},
			CreatedAt: time.Now(),
			Version:   1,
		},
	}

	importPath := filepath.Join(tempDir, "import.yaml")
	data, _ := yaml.Marshal(keys)
	os.WriteFile(importPath, data, 0600)

	// Import keys
	count, err := storage.ImportKeys(importPath)
	if err != nil {
		t.Fatalf("Failed to import keys from YAML: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 imported key, got %d", count)
	}
}

func TestImportKeys_InvalidFile(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	storage, _ := New()

	importPath := filepath.Join(tempDir, "invalid.json")
	os.WriteFile(importPath, []byte("invalid json"), 0600)

	_, err := storage.ImportKeys(importPath)
	if err == nil {
		t.Fatal("Expected error for invalid import file")
	}
}

func TestImportKeys_NonexistentFile(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	storage, _ := New()

	_, err := storage.ImportKeys("nonexistent.json")
	if err == nil {
		t.Fatal("Expected error for nonexistent file")
	}
}

func TestExportKeyToFile(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	storage, _ := New()

	key := &generator.Key{
		ID:        "test-id",
		Name:      "test-key",
		Service:   "test-service",
		Type:      generator.KeyTypeJWT,
		Value:     "test-value",
		Metadata:  map[string]string{},
		CreatedAt: time.Now(),
		Version:   1,
	}

	outputPath := filepath.Join(tempDir, "output", "key.json")
	err := storage.ExportKeyToFile(key, outputPath)
	if err != nil {
		t.Fatalf("Failed to export key to file: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Output file was not created")
	}

	// Verify file permissions
	info, _ := os.Stat(outputPath)
	if info.Mode().Perm() != 0600 {
		t.Errorf("Expected file permissions 0600, got %o", info.Mode().Perm())
	}

	// Verify content
	content, _ := os.ReadFile(outputPath)
	var exportedKey generator.Key
	if err := json.Unmarshal(content, &exportedKey); err != nil {
		t.Fatalf("Failed to parse exported key: %v", err)
	}

	if exportedKey.ID != key.ID {
		t.Errorf("Expected ID %s, got %s", key.ID, exportedKey.ID)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func BenchmarkSaveKey(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	storage, _ := New()

	key := &generator.Key{
		ID:        "bench-id",
		Name:      "bench-key",
		Service:   "bench-service",
		Type:      generator.KeyTypeJWT,
		Value:     "bench-value",
		Metadata:  map[string]string{},
		CreatedAt: time.Now(),
		Version:   1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = storage.SaveKey(key)
	}
}

func BenchmarkListKeys(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	storage, _ := New()

	// Add some keys
	for i := 0; i < 10; i++ {
		key := &generator.Key{
			ID:        string(rune(i) + '0'),
			Name:      "key" + string(rune(i)+'0'),
			Service:   "service",
			Type:      generator.KeyTypeJWT,
			Value:     "value",
			Metadata:  map[string]string{},
			CreatedAt: time.Now(),
			Version:   1,
		}
		storage.SaveKey(key)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = storage.ListKeys()
	}
}

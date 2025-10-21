package generator

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"os"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	g := New()
	if g == nil {
		t.Fatal("Expected generator to be created")
	}
}

func TestGenerateJWTSecret(t *testing.T) {
	g := New()

	tests := []struct {
		name      string
		keyName   string
		service   string
		length    int
		expectErr bool
	}{
		{
			name:      "Valid JWT secret with minimum length",
			keyName:   "auth-jwt",
			service:   "authentication",
			length:    32,
			expectErr: false,
		},
		{
			name:      "Valid JWT secret with 64 bytes",
			keyName:   "auth-jwt-64",
			service:   "authentication",
			length:    64,
			expectErr: false,
		},
		{
			name:      "Invalid JWT secret - too short",
			keyName:   "short-jwt",
			service:   "authentication",
			length:    16,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := g.GenerateJWTSecret(tt.keyName, tt.service, tt.length)

			if tt.expectErr {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Verify key properties
			if key.Name != tt.keyName {
				t.Errorf("Expected name %s, got %s", tt.keyName, key.Name)
			}
			if key.Service != tt.service {
				t.Errorf("Expected service %s, got %s", tt.service, key.Service)
			}
			if key.Type != KeyTypeJWT {
				t.Errorf("Expected type %s, got %s", KeyTypeJWT, key.Type)
			}

			// Verify key value is base64 encoded and correct length
			decoded, err := base64.StdEncoding.DecodeString(key.Value)
			if err != nil {
				t.Fatalf("Key value is not valid base64: %v", err)
			}
			if len(decoded) != tt.length {
				t.Errorf("Expected key length %d, got %d", tt.length, len(decoded))
			}

			// Verify metadata
			if key.Metadata["length"] != string(rune(tt.length)+'0') && tt.length < 10 {
				// For larger numbers, just check it exists
				if _, ok := key.Metadata["length"]; !ok {
					t.Error("Expected length in metadata")
				}
			}

			// Verify timestamps
			if key.CreatedAt.IsZero() {
				t.Error("Expected CreatedAt to be set")
			}
			if key.Version != 1 {
				t.Errorf("Expected version 1, got %d", key.Version)
			}
		})
	}
}

func TestGenerateDatabaseKey(t *testing.T) {
	g := New()

	tests := []struct {
		name      string
		keyName   string
		service   string
		length    int
		expectErr bool
	}{
		{
			name:      "Valid database key (32 bytes for AES-256)",
			keyName:   "db-key",
			service:   "localization",
			length:    32,
			expectErr: false,
		},
		{
			name:      "Invalid database key - wrong length",
			keyName:   "db-key-invalid",
			service:   "localization",
			length:    16,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := g.GenerateDatabaseKey(tt.keyName, tt.service, tt.length)

			if tt.expectErr {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Verify key properties
			if key.Type != KeyTypeDB {
				t.Errorf("Expected type %s, got %s", KeyTypeDB, key.Type)
			}

			// Verify key value
			decoded, err := base64.StdEncoding.DecodeString(key.Value)
			if err != nil {
				t.Fatalf("Key value is not valid base64: %v", err)
			}
			if len(decoded) != 32 {
				t.Errorf("Expected key length 32, got %d", len(decoded))
			}

			// Verify metadata
			if key.Metadata["algorithm"] != "AES-256" {
				t.Error("Expected AES-256 algorithm in metadata")
			}
		})
	}
}

func TestGenerateTLSCertificate(t *testing.T) {
	g := New()

	// Create a temporary directory for certificates
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Change to temp directory
	os.Chdir(tempDir)

	key, err := g.GenerateTLSCertificate("test-tls", "test-service")
	if err != nil {
		t.Fatalf("Failed to generate TLS certificate: %v", err)
	}

	// Verify key properties
	if key.Type != KeyTypeTLS {
		t.Errorf("Expected type %s, got %s", KeyTypeTLS, key.Type)
	}
	if key.Value != "" {
		t.Error("Expected empty value for TLS keys (stored in files)")
	}

	// Verify certificate file exists
	certPath, ok := key.Metadata["cert_path"]
	if !ok {
		t.Fatal("Expected cert_path in metadata")
	}
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		t.Errorf("Certificate file does not exist: %s", certPath)
	}

	// Verify key file exists
	keyPath, ok := key.Metadata["key_path"]
	if !ok {
		t.Fatal("Expected key_path in metadata")
	}
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		t.Errorf("Key file does not exist: %s", keyPath)
	}

	// Verify certificate content
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		t.Fatalf("Failed to read certificate file: %v", err)
	}
	block, _ := pem.Decode(certPEM)
	if block == nil {
		t.Fatal("Failed to decode PEM block")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("Failed to parse certificate: %v", err)
	}

	// Verify certificate properties
	if cert.Subject.CommonName != "localhost" {
		t.Errorf("Expected CN=localhost, got %s", cert.Subject.CommonName)
	}

	// Verify expiration is set
	if key.ExpiresAt == nil {
		t.Error("Expected ExpiresAt to be set")
	}

	// Verify file permissions
	info, _ := os.Stat(keyPath)
	if info.Mode().Perm() != 0600 {
		t.Errorf("Expected key file permissions 0600, got %o", info.Mode().Perm())
	}
}

func TestGenerateRedisPassword(t *testing.T) {
	g := New()

	tests := []struct {
		name      string
		keyName   string
		service   string
		length    int
		expectErr bool
	}{
		{
			name:      "Valid Redis password",
			keyName:   "redis-pass",
			service:   "localization",
			length:    32,
			expectErr: false,
		},
		{
			name:      "Invalid Redis password - too short",
			keyName:   "redis-pass-short",
			service:   "localization",
			length:    8,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := g.GenerateRedisPassword(tt.keyName, tt.service, tt.length)

			if tt.expectErr {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Verify key properties
			if key.Type != KeyTypeRedis {
				t.Errorf("Expected type %s, got %s", KeyTypeRedis, key.Type)
			}

			// Verify key value
			decoded, err := base64.StdEncoding.DecodeString(key.Value)
			if err != nil {
				t.Fatalf("Key value is not valid base64: %v", err)
			}
			if len(decoded) != tt.length {
				t.Errorf("Expected key length %d, got %d", tt.length, len(decoded))
			}
		})
	}
}

func TestGenerateAPIKey(t *testing.T) {
	g := New()

	tests := []struct {
		name      string
		keyName   string
		service   string
		length    int
		expectErr bool
	}{
		{
			name:      "Valid API key",
			keyName:   "api-key",
			service:   "localization",
			length:    32,
			expectErr: false,
		},
		{
			name:      "Invalid API key - too short",
			keyName:   "api-key-short",
			service:   "localization",
			length:    16,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := g.GenerateAPIKey(tt.keyName, tt.service, tt.length)

			if tt.expectErr {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Verify key properties
			if key.Type != KeyTypeAPI {
				t.Errorf("Expected type %s, got %s", KeyTypeAPI, key.Type)
			}

			// Verify key value
			decoded, err := base64.StdEncoding.DecodeString(key.Value)
			if err != nil {
				t.Fatalf("Key value is not valid base64: %v", err)
			}
			if len(decoded) != tt.length {
				t.Errorf("Expected key length %d, got %d", tt.length, len(decoded))
			}
		})
	}
}

func TestRotateKey(t *testing.T) {
	g := New()

	tests := []struct {
		name    string
		keyType KeyType
		setup   func() *Key
	}{
		{
			name:    "Rotate JWT secret",
			keyType: KeyTypeJWT,
			setup: func() *Key {
				key, _ := g.GenerateJWTSecret("jwt-test", "auth", 64)
				return key
			},
		},
		{
			name:    "Rotate database key",
			keyType: KeyTypeDB,
			setup: func() *Key {
				key, _ := g.GenerateDatabaseKey("db-test", "loc", 32)
				return key
			},
		},
		{
			name:    "Rotate Redis password",
			keyType: KeyTypeRedis,
			setup: func() *Key {
				key, _ := g.GenerateRedisPassword("redis-test", "cache", 32)
				return key
			},
		},
		{
			name:    "Rotate API key",
			keyType: KeyTypeAPI,
			setup: func() *Key {
				key, _ := g.GenerateAPIKey("api-test", "api", 32)
				return key
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldKey := tt.setup()
			oldValue := oldKey.Value
			oldID := oldKey.ID
			oldVersion := oldKey.Version

			newKey, err := g.RotateKey(oldKey)
			if err != nil {
				t.Fatalf("Failed to rotate key: %v", err)
			}

			// Verify new key properties
			if newKey.Name != oldKey.Name {
				t.Error("Key name should not change during rotation")
			}
			if newKey.Service != oldKey.Service {
				t.Error("Service name should not change during rotation")
			}
			if newKey.Type != oldKey.Type {
				t.Error("Key type should not change during rotation")
			}

			// Verify new values
			if newKey.Value == oldValue && tt.keyType != KeyTypeTLS {
				t.Error("Key value should change during rotation")
			}
			if newKey.ID == oldID {
				t.Error("Key ID should change during rotation")
			}
			if newKey.Version != oldVersion+1 {
				t.Errorf("Expected version %d, got %d", oldVersion+1, newKey.Version)
			}
		})
	}
}

func TestRotateTLSKey(t *testing.T) {
	g := New()

	// Create a temporary directory
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	oldKey, err := g.GenerateTLSCertificate("tls-test", "service")
	if err != nil {
		t.Fatalf("Failed to generate initial TLS key: %v", err)
	}

	newKey, err := g.RotateKey(oldKey)
	if err != nil {
		t.Fatalf("Failed to rotate TLS key: %v", err)
	}

	// Verify version increment
	if newKey.Version != oldKey.Version+1 {
		t.Errorf("Expected version %d, got %d", oldKey.Version+1, newKey.Version)
	}

	// Verify new certificate files exist
	certPath := newKey.Metadata["cert_path"]
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		t.Error("New certificate file does not exist")
	}
}

func TestRotateKey_UnsupportedType(t *testing.T) {
	g := New()

	key := &Key{
		Type:    KeyType("unsupported"),
		Name:    "test",
		Service: "test",
		Version: 1,
	}

	_, err := g.RotateKey(key)
	if err == nil {
		t.Fatal("Expected error for unsupported key type")
	}
}

func TestGenerateRandomBytes(t *testing.T) {
	tests := []int{16, 32, 64, 128}

	for _, length := range tests {
		t.Run(string(rune(length)+'0')+" bytes", func(t *testing.T) {
			bytes, err := generateRandomBytes(length)
			if err != nil {
				t.Fatalf("Failed to generate random bytes: %v", err)
			}

			if len(bytes) != length {
				t.Errorf("Expected %d bytes, got %d", length, len(bytes))
			}

			// Verify randomness (not all zeros)
			allZeros := true
			for _, b := range bytes {
				if b != 0 {
					allZeros = false
					break
				}
			}
			if allZeros {
				t.Error("Generated bytes are all zeros")
			}
		})
	}
}

func TestKey_Metadata(t *testing.T) {
	g := New()

	key, err := g.GenerateJWTSecret("test", "service", 64)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	// Verify metadata exists
	if key.Metadata == nil {
		t.Fatal("Expected metadata to be initialized")
	}

	// Verify length metadata
	if _, ok := key.Metadata["length"]; !ok {
		t.Error("Expected 'length' in metadata")
	}
}

func TestKey_Timestamps(t *testing.T) {
	g := New()

	before := time.Now()
	key, err := g.GenerateJWTSecret("test", "service", 64)
	after := time.Now()

	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	// Verify CreatedAt is within expected range
	if key.CreatedAt.Before(before) || key.CreatedAt.After(after) {
		t.Error("CreatedAt timestamp is outside expected range")
	}

	// ExpiresAt should be nil for JWT keys
	if key.ExpiresAt != nil {
		t.Error("Expected ExpiresAt to be nil for JWT keys")
	}
}

func TestTLSCertificate_Expiration(t *testing.T) {
	g := New()

	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	key, err := g.GenerateTLSCertificate("test", "service")
	if err != nil {
		t.Fatalf("Failed to generate TLS certificate: %v", err)
	}

	// Verify ExpiresAt is set
	if key.ExpiresAt == nil {
		t.Fatal("Expected ExpiresAt to be set for TLS keys")
	}

	// Verify expiration is approximately 1 year from now
	expectedExpiry := time.Now().AddDate(1, 0, 0)
	diff := key.ExpiresAt.Sub(expectedExpiry)
	if diff < -time.Hour || diff > time.Hour {
		t.Errorf("TLS certificate expiration is not approximately 1 year from now")
	}

	// Verify valid_until in metadata matches ExpiresAt (within 1 second tolerance)
	validUntil, ok := key.Metadata["valid_until"]
	if !ok {
		t.Fatal("Expected valid_until in metadata")
	}
	parsedTime, err := time.Parse(time.RFC3339, validUntil)
	if err != nil {
		t.Fatalf("Failed to parse valid_until: %v", err)
	}
	timeDiff := parsedTime.Sub(*key.ExpiresAt)
	if timeDiff < -time.Second || timeDiff > time.Second {
		t.Errorf("valid_until in metadata differs from ExpiresAt by %v", timeDiff)
	}
}

func TestGenerateDatabaseKey_OnlyAccepts32Bytes(t *testing.T) {
	g := New()

	invalidLengths := []int{16, 24, 48, 64}
	for _, length := range invalidLengths {
		_, err := g.GenerateDatabaseKey("test", "service", length)
		if err == nil {
			t.Errorf("Expected error for length %d, but got none", length)
		}
	}
}

func BenchmarkGenerateJWTSecret(b *testing.B) {
	g := New()
	for i := 0; i < b.N; i++ {
		_, _ = g.GenerateJWTSecret("bench-jwt", "service", 64)
	}
}

func BenchmarkGenerateDatabaseKey(b *testing.B) {
	g := New()
	for i := 0; i < b.N; i++ {
		_, _ = g.GenerateDatabaseKey("bench-db", "service", 32)
	}
}

func BenchmarkGenerateTLSCertificate(b *testing.B) {
	g := New()
	tempDir := b.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = g.GenerateTLSCertificate("bench-tls", "service")
	}
}

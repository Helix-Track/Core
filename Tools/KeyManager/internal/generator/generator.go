package generator

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// KeyType represents the type of key
type KeyType string

const (
	KeyTypeJWT   KeyType = "jwt"
	KeyTypeDB    KeyType = "db"
	KeyTypeTLS   KeyType = "tls"
	KeyTypeRedis KeyType = "redis"
	KeyTypeAPI   KeyType = "api"
)

// Key represents a generated key with metadata
type Key struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Service   string            `json:"service"`
	Type      KeyType           `json:"type"`
	Value     string            `json:"value"`      // The actual key (base64 encoded)
	Metadata  map[string]string `json:"metadata"`   // Additional metadata (e.g., file paths for TLS)
	CreatedAt time.Time         `json:"created_at"`
	ExpiresAt *time.Time        `json:"expires_at,omitempty"`
	Version   int               `json:"version"`    // For key rotation
}

// Generator handles key generation
type Generator struct{}

// New creates a new key generator
func New() *Generator {
	return &Generator{}
}

// GenerateJWTSecret generates a JWT signing secret
func (g *Generator) GenerateJWTSecret(name, service string, length int) (*Key, error) {
	if length < 32 {
		return nil, fmt.Errorf("JWT secret must be at least 32 bytes")
	}

	secret, err := generateRandomBytes(length)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return &Key{
		ID:        uuid.New().String(),
		Name:      name,
		Service:   service,
		Type:      KeyTypeJWT,
		Value:     base64.StdEncoding.EncodeToString(secret),
		Metadata:  map[string]string{"length": fmt.Sprintf("%d", length)},
		CreatedAt: time.Now(),
		Version:   1,
	}, nil
}

// GenerateDatabaseKey generates a database encryption key
func (g *Generator) GenerateDatabaseKey(name, service string, length int) (*Key, error) {
	// For SQL Cipher, we need a passphrase (32 bytes recommended for AES-256)
	if length != 32 {
		return nil, fmt.Errorf("database encryption key must be exactly 32 bytes for AES-256")
	}

	key, err := generateRandomBytes(length)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return &Key{
		ID:        uuid.New().String(),
		Name:      name,
		Service:   service,
		Type:      KeyTypeDB,
		Value:     base64.StdEncoding.EncodeToString(key),
		Metadata:  map[string]string{"length": fmt.Sprintf("%d", length), "algorithm": "AES-256"},
		CreatedAt: time.Now(),
		Version:   1,
	}, nil
}

// GenerateTLSCertificate generates a TLS certificate and private key
func (g *Generator) GenerateTLSCertificate(name, service string) (*Key, error) {
	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"HelixTrack"},
			Country:       []string{"US"},
			Province:      []string{"California"},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
			CommonName:    "localhost",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0), // Valid for 1 year
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost", "*.localhost"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
	}

	// Create certificate
	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	// Create output directory
	certDir := filepath.Join("keys", service, "tls")
	if err := os.MkdirAll(certDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create certificate directory: %w", err)
	}

	// Write certificate file
	certPath := filepath.Join(certDir, fmt.Sprintf("%s.crt", name))
	certFile, err := os.Create(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate file: %w", err)
	}
	defer certFile.Close()

	if err := pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes}); err != nil {
		return nil, fmt.Errorf("failed to write certificate: %w", err)
	}

	// Write private key file
	keyPath := filepath.Join(certDir, fmt.Sprintf("%s.key", name))
	keyFile, err := os.Create(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create key file: %w", err)
	}
	defer keyFile.Close()

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	if err := pem.Encode(keyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privateKeyBytes}); err != nil {
		return nil, fmt.Errorf("failed to write private key: %w", err)
	}

	// Set proper permissions
	os.Chmod(keyPath, 0600)
	os.Chmod(certPath, 0644)

	return &Key{
		ID:      uuid.New().String(),
		Name:    name,
		Service: service,
		Type:    KeyTypeTLS,
		Value:   "", // TLS keys are stored in files, not in value field
		Metadata: map[string]string{
			"cert_path":  certPath,
			"key_path":   keyPath,
			"common_name": template.Subject.CommonName,
			"valid_until": template.NotAfter.Format(time.RFC3339),
		},
		CreatedAt: time.Now(),
		ExpiresAt: &template.NotAfter,
		Version:   1,
	}, nil
}

// GenerateRedisPassword generates a Redis password
func (g *Generator) GenerateRedisPassword(name, service string, length int) (*Key, error) {
	if length < 16 {
		return nil, fmt.Errorf("Redis password must be at least 16 bytes")
	}

	password, err := generateRandomBytes(length)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return &Key{
		ID:        uuid.New().String(),
		Name:      name,
		Service:   service,
		Type:      KeyTypeRedis,
		Value:     base64.StdEncoding.EncodeToString(password),
		Metadata:  map[string]string{"length": fmt.Sprintf("%d", length)},
		CreatedAt: time.Now(),
		Version:   1,
	}, nil
}

// GenerateAPIKey generates an API key
func (g *Generator) GenerateAPIKey(name, service string, length int) (*Key, error) {
	if length < 32 {
		return nil, fmt.Errorf("API key must be at least 32 bytes")
	}

	key, err := generateRandomBytes(length)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return &Key{
		ID:        uuid.New().String(),
		Name:      name,
		Service:   service,
		Type:      KeyTypeAPI,
		Value:     base64.StdEncoding.EncodeToString(key),
		Metadata:  map[string]string{"length": fmt.Sprintf("%d", length)},
		CreatedAt: time.Now(),
		Version:   1,
	}, nil
}

// RotateKey rotates an existing key (generates new version)
func (g *Generator) RotateKey(oldKey *Key) (*Key, error) {
	var newKey *Key
	var err error

	switch oldKey.Type {
	case KeyTypeJWT:
		length := 64
		if l, ok := oldKey.Metadata["length"]; ok {
			fmt.Sscanf(l, "%d", &length)
		}
		newKey, err = g.GenerateJWTSecret(oldKey.Name, oldKey.Service, length)

	case KeyTypeDB:
		newKey, err = g.GenerateDatabaseKey(oldKey.Name, oldKey.Service, 32)

	case KeyTypeTLS:
		newKey, err = g.GenerateTLSCertificate(oldKey.Name, oldKey.Service)

	case KeyTypeRedis:
		length := 32
		if l, ok := oldKey.Metadata["length"]; ok {
			fmt.Sscanf(l, "%d", &length)
		}
		newKey, err = g.GenerateRedisPassword(oldKey.Name, oldKey.Service, length)

	case KeyTypeAPI:
		length := 32
		if l, ok := oldKey.Metadata["length"]; ok {
			fmt.Sscanf(l, "%d", &length)
		}
		newKey, err = g.GenerateAPIKey(oldKey.Name, oldKey.Service, length)

	default:
		return nil, fmt.Errorf("unsupported key type for rotation: %s", oldKey.Type)
	}

	if err != nil {
		return nil, err
	}

	// Increment version
	newKey.Version = oldKey.Version + 1

	return newKey, nil
}

// generateRandomBytes generates cryptographically secure random bytes
func generateRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}
	return bytes, nil
}

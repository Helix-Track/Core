package security

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"time"

	"helixtrack.ru/core/internal/models"
)

// ServiceSigner handles cryptographic signing and verification of services
type ServiceSigner struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

// NewServiceSigner creates a new service signer with generated keys
func NewServiceSigner() (*ServiceSigner, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	return &ServiceSigner{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
	}, nil
}

// NewServiceSignerFromPrivateKey creates a signer from an existing private key
func NewServiceSignerFromPrivateKey(privateKeyPEM string) (*ServiceSigner, error) {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block containing private key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return &ServiceSigner{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
	}, nil
}

// SignServiceRegistration signs a service registration with the private key
func (s *ServiceSigner) SignServiceRegistration(service *models.ServiceRegistration) error {
	// Get the public key in PEM format
	publicKeyPEM, err := s.GetPublicKeyPEM()
	if err != nil {
		return fmt.Errorf("failed to get public key: %w", err)
	}
	service.PublicKey = publicKeyPEM

	// Compute the data to sign
	data := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%d",
		service.ID,
		service.Name,
		service.Type,
		service.Version,
		service.URL,
		service.PublicKey,
		service.RegisteredAt.Unix(),
	)

	// Hash the data
	hashed := sha256.Sum256([]byte(data))

	// Sign the hash
	signature, err := rsa.SignPKCS1v15(rand.Reader, s.privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return fmt.Errorf("failed to sign service registration: %w", err)
	}

	// Encode signature as base64
	service.Signature = base64.StdEncoding.EncodeToString(signature)

	return nil
}

// VerifyServiceRegistration verifies the signature of a service registration
func (s *ServiceSigner) VerifyServiceRegistration(service *models.ServiceRegistration) error {
	// Decode the signature
	signature, err := base64.StdEncoding.DecodeString(service.Signature)
	if err != nil {
		return fmt.Errorf("failed to decode signature: %w", err)
	}

	// Parse the public key from the service
	publicKey, err := ParsePublicKey(service.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to parse public key: %w", err)
	}

	// Compute the data that was signed
	data := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%d",
		service.ID,
		service.Name,
		service.Type,
		service.Version,
		service.URL,
		service.PublicKey,
		service.RegisteredAt.Unix(),
	)

	// Hash the data
	hashed := sha256.Sum256([]byte(data))

	// Verify the signature
	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hashed[:], signature)
	if err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	return nil
}

// VerifyServiceRotation verifies that a service rotation request is legitimate
func (s *ServiceSigner) VerifyServiceRotation(
	oldService *models.ServiceRegistration,
	newService *models.ServiceRegistration,
	adminToken string,
) error {
	// 1. Verify old service is not already rotating or decommissioned
	if !oldService.CanRotate() {
		return fmt.Errorf("service cannot be rotated in current status: %s", oldService.Status)
	}

	// 2. Verify new service signature
	if err := s.VerifyServiceRegistration(newService); err != nil {
		return fmt.Errorf("new service signature invalid: %w", err)
	}

	// 3. Verify admin token (implementation depends on your admin auth system)
	if !s.verifyAdminToken(adminToken) {
		return fmt.Errorf("invalid admin token")
	}

	// 4. Verify service types match
	if oldService.Type != newService.Type {
		return fmt.Errorf("service type mismatch: old=%s, new=%s", oldService.Type, newService.Type)
	}

	// 5. Verify new service is healthy (should be checked before calling this)
	if !newService.IsHealthy() {
		return fmt.Errorf("new service is not healthy")
	}

	// 6. Verify time-based constraints (prevent rapid rotations)
	if time.Since(oldService.RegisteredAt) < 5*time.Minute {
		return fmt.Errorf("service was registered too recently for rotation")
	}

	// 7. Verify new service has been registered long enough
	if time.Since(newService.RegisteredAt) < 5*time.Minute {
		return fmt.Errorf("new service must be registered for at least 5 minutes before rotation")
	}

	return nil
}

// verifyAdminToken verifies an admin authorization token
func (s *ServiceSigner) verifyAdminToken(token string) bool {
	// In a real implementation, this would verify against a secure token store
	// For now, we'll use a simple check
	if token == "" {
		return false
	}

	// Hash the token and compare with expected values
	hash := sha256.Sum256([]byte(token))
	hashStr := base64.StdEncoding.EncodeToString(hash[:])

	// This should be replaced with actual token verification
	// For production, integrate with the JWT service or a dedicated admin auth system
	_ = hashStr

	// For development/testing, accept non-empty tokens
	// TODO: Implement proper admin token verification
	return len(token) >= 32
}

// GetPublicKeyPEM returns the public key in PEM format
func (s *ServiceSigner) GetPublicKeyPEM() (string, error) {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(s.publicKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal public key: %w", err)
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return string(publicKeyPEM), nil
}

// GetPrivateKeyPEM returns the private key in PEM format
func (s *ServiceSigner) GetPrivateKeyPEM() string {
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(s.privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	return string(privateKeyPEM)
}

// ParsePublicKey parses a PEM-encoded public key
func ParsePublicKey(publicKeyPEM string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block containing public key")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	return rsaPublicKey, nil
}

// GenerateAdminToken generates a secure admin token for service operations
func GenerateAdminToken(username string, secret string) string {
	// Combine username, secret, and timestamp
	data := fmt.Sprintf("%s|%s|%d", username, secret, time.Now().Unix())
	hash := sha256.Sum256([]byte(data))
	return base64.StdEncoding.EncodeToString(hash[:])
}

// VerifyRotationCode verifies a rotation verification code
func VerifyRotationCode(serviceID string, code string, secret string) bool {
	// Generate expected code
	expected := GenerateRotationCode(serviceID, secret)
	return code == expected
}

// GenerateRotationCode generates a verification code for service rotation
func GenerateRotationCode(serviceID string, secret string) string {
	// Combine service ID, secret, and current hour (time-based code)
	hour := time.Now().UTC().Hour()
	data := fmt.Sprintf("%s|%s|%d", serviceID, secret, hour)
	hash := sha256.Sum256([]byte(data))
	return base64.StdEncoding.EncodeToString(hash[:])[:16] // First 16 characters
}

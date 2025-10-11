package security

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"helixtrack.ru/core/internal/models"
)

func TestNewServiceSigner(t *testing.T) {
	t.Run("Successful creation", func(t *testing.T) {
		signer, err := NewServiceSigner()
		require.NoError(t, err)
		assert.NotNil(t, signer)
		assert.NotNil(t, signer.privateKey)
	})
}

func TestServiceSigner_SignAndVerifyServiceRegistration(t *testing.T) {
	signer, err := NewServiceSigner()
	require.NoError(t, err)

	service := &models.ServiceRegistration{
		ID:             "test-service-1",
		Name:           "Test Service",
		Type:           models.ServiceTypeAuthentication,
		Version:        "1.0.0",
		URL:            "http://localhost:8081",
		HealthCheckURL: "http://localhost:8081/health",
		Status:         models.ServiceStatusHealthy,
		Priority:       10,
		RegisteredAt:   time.Now(),
	}

	t.Run("Sign service registration", func(t *testing.T) {
		err := signer.SignServiceRegistration(service)
		require.NoError(t, err)
		assert.NotEmpty(t, service.Signature)
		assert.NotEmpty(t, service.PublicKey)
	})

	t.Run("Verify valid signature", func(t *testing.T) {
		err := signer.VerifyServiceRegistration(service)
		require.NoError(t, err)
	})

	t.Run("Reject tampered service", func(t *testing.T) {
		// Create a copy and tamper with it
		tamperedService := *service
		tamperedService.URL = "http://malicious.com"
		// Keep original signature

		err := signer.VerifyServiceRegistration(&tamperedService)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "signature verification failed")
	})

	t.Run("Reject invalid public key format", func(t *testing.T) {
		invalidService := *service
		invalidService.PublicKey = "invalid-key"

		err := signer.VerifyServiceRegistration(&invalidService)
		assert.Error(t, err)
	})

	t.Run("Reject invalid signature format", func(t *testing.T) {
		invalidService := *service
		invalidService.Signature = "invalid-signature"

		err := signer.VerifyServiceRegistration(&invalidService)
		assert.Error(t, err)
	})
}

func TestServiceSigner_VerifyServiceRotation(t *testing.T) {
	signer, err := NewServiceSigner()
	require.NoError(t, err)

	oldService := &models.ServiceRegistration{
		ID:             "old-service",
		Name:           "Old Service",
		Type:           models.ServiceTypeAuthentication,
		Version:        "1.0.0",
		URL:            "http://localhost:8081",
		HealthCheckURL: "http://localhost:8081/health",
		Status:         models.ServiceStatusHealthy,
		RegisteredAt:   time.Now().Add(-10 * time.Minute),
	}

	err = signer.SignServiceRegistration(oldService)
	require.NoError(t, err)

	newService := &models.ServiceRegistration{
		ID:             "new-service",
		Name:           "New Service",
		Type:           models.ServiceTypeAuthentication,
		Version:        "1.1.0",
		URL:            "http://localhost:8082",
		HealthCheckURL: "http://localhost:8082/health",
		Status:         models.ServiceStatusHealthy,
		RegisteredAt:   time.Now().Add(-10 * time.Minute), // Registered 10 minutes ago
	}

	err = signer.SignServiceRegistration(newService)
	require.NoError(t, err)

	t.Run("Successful rotation with valid admin token", func(t *testing.T) {
		adminToken := "valid-admin-token-with-32-characters-minimum-length"
		err := signer.VerifyServiceRotation(oldService, newService, adminToken)
		require.NoError(t, err)
	})

	t.Run("Reject rotation with short admin token", func(t *testing.T) {
		shortToken := "short"
		err := signer.VerifyServiceRotation(oldService, newService, shortToken)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid admin token")
	})

	t.Run("Reject rotation when old service is rotating", func(t *testing.T) {
		rotatingService := *oldService
		rotatingService.Status = models.ServiceStatusRotating

		adminToken := "valid-admin-token-with-32-characters-minimum"
		err := signer.VerifyServiceRotation(&rotatingService, newService, adminToken)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be rotated")
	})

	t.Run("Reject rotation when old service is decommissioned", func(t *testing.T) {
		decommissionedService := *oldService
		decommissionedService.Status = models.ServiceStatusDecommission

		adminToken := "valid-admin-token-with-32-characters-minimum"
		err := signer.VerifyServiceRotation(&decommissionedService, newService, adminToken)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be rotated")
	})

	t.Run("Reject rotation with invalid new service signature", func(t *testing.T) {
		invalidService := *newService
		invalidService.Signature = "invalid-signature"

		adminToken := "valid-admin-token-with-32-characters-minimum"
		err := signer.VerifyServiceRotation(oldService, &invalidService, adminToken)
		assert.Error(t, err)
	})

	t.Run("Reject rotation with mismatched service types", func(t *testing.T) {
		mismatchedService := *newService
		mismatchedService.Type = models.ServiceTypePermissions

		// Need to re-sign after changing type
		err := signer.SignServiceRegistration(&mismatchedService)
		require.NoError(t, err)

		adminToken := "valid-admin-token-with-32-characters-minimum"
		err = signer.VerifyServiceRotation(oldService, &mismatchedService, adminToken)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "service type mismatch")
	})

	t.Run("Reject rotation when new service is not healthy", func(t *testing.T) {
		unhealthyService := *newService
		unhealthyService.Status = models.ServiceStatusUnhealthy

		// Re-sign after status change
		err := signer.SignServiceRegistration(&unhealthyService)
		require.NoError(t, err)

		adminToken := "valid-admin-token-with-32-characters-minimum"
		err = signer.VerifyServiceRotation(oldService, &unhealthyService, adminToken)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not healthy")
	})

	t.Run("Reject rotation when service registered too recently", func(t *testing.T) {
		recentService := *newService
		recentService.RegisteredAt = time.Now()

		// Re-sign after timestamp change
		err := signer.SignServiceRegistration(&recentService)
		require.NoError(t, err)

		adminToken := "valid-admin-token-with-32-characters-minimum"
		err = signer.VerifyServiceRotation(oldService, &recentService, adminToken)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "new service must be registered for at least 5 minutes before rotation")
	})
}

func TestGenerateRotationCode(t *testing.T) {
	t.Run("Generate valid rotation code", func(t *testing.T) {
		code := GenerateRotationCode("service-id", "secret")
		assert.NotEmpty(t, code)
		assert.Len(t, code, 16) // Truncated to 16 characters
	})

	t.Run("Different services generate different codes", func(t *testing.T) {
		secret := "shared-secret"
		code1 := GenerateRotationCode("service-1", secret)
		code2 := GenerateRotationCode("service-2", secret)

		assert.NotEqual(t, code1, code2)
	})

	t.Run("Different secrets generate different codes", func(t *testing.T) {
		serviceID := "service-1"
		code1 := GenerateRotationCode(serviceID, "secret-1")
		code2 := GenerateRotationCode(serviceID, "secret-2")

		assert.NotEqual(t, code1, code2)
	})

	t.Run("Verify rotation code", func(t *testing.T) {
		serviceID := "service-1"
		secret := "test-secret"
		code := GenerateRotationCode(serviceID, secret)

		valid := VerifyRotationCode(serviceID, code, secret)
		assert.True(t, valid, "Code should be valid")

		invalid := VerifyRotationCode(serviceID, "wrong-code", secret)
		assert.False(t, invalid, "Wrong code should be invalid")
	})
}

func TestServiceSigner_GetKeys(t *testing.T) {
	signer, err := NewServiceSigner()
	require.NoError(t, err)

	t.Run("Get public key PEM", func(t *testing.T) {
		publicKeyPEM, err := signer.GetPublicKeyPEM()
		require.NoError(t, err)
		assert.NotEmpty(t, publicKeyPEM)
		assert.Contains(t, publicKeyPEM, "-----BEGIN PUBLIC KEY-----")
		assert.Contains(t, publicKeyPEM, "-----END PUBLIC KEY-----")
	})

	t.Run("Get private key PEM", func(t *testing.T) {
		privateKeyPEM := signer.GetPrivateKeyPEM()
		assert.NotEmpty(t, privateKeyPEM)
		assert.Contains(t, privateKeyPEM, "-----BEGIN RSA PRIVATE KEY-----")
		assert.Contains(t, privateKeyPEM, "-----END RSA PRIVATE KEY-----")
	})
}

func TestServiceSigner_MultipleSigners(t *testing.T) {
	t.Run("Service with swapped public key fails verification", func(t *testing.T) {
		signer1, err := NewServiceSigner()
		require.NoError(t, err)

		signer2, err := NewServiceSigner()
		require.NoError(t, err)

		service := &models.ServiceRegistration{
			ID:             "test-service",
			Name:           "Test Service",
			Type:           models.ServiceTypeAuthentication,
			Version:        "1.0.0",
			URL:            "http://localhost:8081",
			HealthCheckURL: "http://localhost:8081/health",
			Status:         models.ServiceStatusHealthy,
			RegisteredAt:   time.Now().Add(-10 * time.Minute),
		}

		// Sign with signer1
		err = signer1.SignServiceRegistration(service)
		require.NoError(t, err)

		// Save the original signature
		originalSignature := service.Signature

		// Swap the public key to signer2's key (but keep signer1's signature)
		publicKey2, err := signer2.GetPublicKeyPEM()
		require.NoError(t, err)
		service.PublicKey = publicKey2
		service.Signature = originalSignature

		// Verification should fail because signature doesn't match the new public key
		err = signer1.VerifyServiceRegistration(service)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "signature verification failed")
	})
}

func TestGenerateAdminToken(t *testing.T) {
	t.Run("Generate admin token", func(t *testing.T) {
		token := GenerateAdminToken("admin", "secret")
		assert.NotEmpty(t, token)
		assert.GreaterOrEqual(t, len(token), 32, "Token should be at least 32 characters")
	})

	t.Run("Different users generate different tokens", func(t *testing.T) {
		secret := "shared-secret"
		token1 := GenerateAdminToken("admin1", secret)
		token2 := GenerateAdminToken("admin2", secret)

		assert.NotEqual(t, token1, token2)
	})

	t.Run("Different secrets generate different tokens", func(t *testing.T) {
		username := "admin"
		token1 := GenerateAdminToken(username, "secret1")
		token2 := GenerateAdminToken(username, "secret2")

		assert.NotEqual(t, token1, token2)
	})
}

func TestServiceSigner_EdgeCases(t *testing.T) {
	signer, err := NewServiceSigner()
	require.NoError(t, err)

	t.Run("Sign service with empty fields", func(t *testing.T) {
		service := &models.ServiceRegistration{
			ID:   "",
			Name: "",
			Type: "",
		}

		err := signer.SignServiceRegistration(service)
		require.NoError(t, err)
		assert.NotEmpty(t, service.Signature)
	})

	t.Run("Sign service with special characters", func(t *testing.T) {
		service := &models.ServiceRegistration{
			ID:             "service-with-special-chars-!@#$%",
			Name:           "Service with 日本語 and émojis 🚀",
			Type:           models.ServiceTypeAuthentication,
			Version:        "1.0.0-beta+build.123",
			URL:            "http://localhost:8081/path?query=value&foo=bar",
			HealthCheckURL: "http://localhost:8081/health?check=full",
			Status:         models.ServiceStatusHealthy,
			RegisteredAt:   time.Now().Add(-10 * time.Minute),
		}

		err := signer.SignServiceRegistration(service)
		require.NoError(t, err)

		err = signer.VerifyServiceRegistration(service)
		require.NoError(t, err)
	})
}

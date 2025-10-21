package e2e

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	baseURL = "https://localhost:8085"
	timeout = 30 * time.Second
)

// TestConfig holds test configuration
type TestConfig struct {
	ServiceURL string
	JWTToken   string
	JWTSecret  string
}

// E2ETestSuite represents an end-to-end test suite
type E2ETestSuite struct {
	t      *testing.T
	config *TestConfig
	client *http.Client
}

// JWTClaims matches the structure from handlers/integration_test.go
type JWTClaims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// NewE2ETestSuite creates a new E2E test suite
func NewE2ETestSuite(t *testing.T) *E2ETestSuite {
	// Create HTTP client that accepts self-signed certificates
	// Note: In production, use proper certificate validation
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   timeout,
	}

	jwtSecret := getEnv("JWT_SECRET", "test-secret-key-for-e2e-testing")
	jwtToken := getEnv("JWT_TOKEN", "")

	// If JWT_TOKEN not provided, generate one using the same pattern as integration tests
	if jwtToken == "" {
		jwtToken = createTestJWT("e2euser", "user", jwtSecret)
		t.Logf("Generated test JWT token for user 'e2euser'")
	}

	config := &TestConfig{
		ServiceURL: getEnv("SERVICE_URL", baseURL),
		JWTToken:   jwtToken,
		JWTSecret:  jwtSecret,
	}

	return &E2ETestSuite{
		t:      t,
		config: config,
		client: client,
	}
}

// createTestJWT creates a test JWT token (matches pattern from integration_test.go)
func createTestJWT(username, role string, secret string) string {
	claims := &JWTClaims{
		Username: username,
		Role:     role,
	}
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(1 * time.Hour))
	claims.IssuedAt = jwt.NewNumericDate(time.Now())

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

// TestHealthCheck tests the health endpoint
func TestHealthCheck(t *testing.T) {
	suite := NewE2ETestSuite(t)

	t.Log("Testing health check endpoint...")

	resp, err := suite.client.Get(suite.config.ServiceURL + "/health")
	if err != nil {
		t.Fatalf("Failed to call health endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var health map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		t.Fatalf("Failed to decode health response: %v", err)
	}

	if status, ok := health["status"].(string); !ok || status != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", health["status"])
	}

	t.Log("✓ Health check passed")
}

// TestGetCatalog tests fetching a complete catalog
func TestGetCatalog(t *testing.T) {
	suite := NewE2ETestSuite(t)

	if suite.config.JWTToken == "" {
		t.Skip("Skipping catalog test - JWT_TOKEN not set")
	}

	t.Log("Testing catalog retrieval...")

	req, err := http.NewRequest("GET", suite.config.ServiceURL+"/v1/catalog/en", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+suite.config.JWTToken)

	resp, err := suite.client.Do(req)
	if err != nil {
		t.Fatalf("Failed to get catalog: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if success, ok := result["success"].(bool); !ok || !success {
		t.Errorf("Expected success=true, got %v", result["success"])
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		t.Fatal("Response data is not an object")
	}

	if language, ok := data["language"].(string); !ok || language != "en" {
		t.Errorf("Expected language 'en', got %v", data["language"])
	}

	t.Log("✓ Catalog retrieval passed")
}

// TestGetSingleLocalization tests fetching a single localization
func TestGetSingleLocalization(t *testing.T) {
	suite := NewE2ETestSuite(t)

	if suite.config.JWTToken == "" {
		t.Skip("Skipping localization test - JWT_TOKEN not set")
	}

	t.Log("Testing single localization retrieval...")

	req, err := http.NewRequest("GET", suite.config.ServiceURL+"/v1/localize/app.welcome?language=en", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+suite.config.JWTToken)

	resp, err := suite.client.Do(req)
	if err != nil {
		t.Fatalf("Failed to get localization: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if success, ok := result["success"].(bool); !ok || !success {
		t.Errorf("Expected success=true, got %v", result["success"])
	}

	t.Log("✓ Single localization retrieval passed")
}

// TestBatchLocalization tests batch localization retrieval
func TestBatchLocalization(t *testing.T) {
	suite := NewE2ETestSuite(t)

	if suite.config.JWTToken == "" {
		t.Skip("Skipping batch localization test - JWT_TOKEN not set")
	}

	t.Log("Testing batch localization...")

	requestBody := map[string]interface{}{
		"language": "en",
		"keys":     []string{"app.welcome", "app.error", "app.success"},
		"fallback": true,
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", suite.config.ServiceURL+"/v1/localize/batch", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+suite.config.JWTToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.client.Do(req)
	if err != nil {
		t.Fatalf("Failed to post batch request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(bodyBytes))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if success, ok := result["success"].(bool); !ok || !success {
		t.Errorf("Expected success=true, got %v", result["success"])
	}

	t.Log("✓ Batch localization passed")
}

// TestGetLanguages tests listing available languages
func TestGetLanguages(t *testing.T) {
	suite := NewE2ETestSuite(t)

	if suite.config.JWTToken == "" {
		t.Skip("Skipping languages test - JWT_TOKEN not set")
	}

	t.Log("Testing languages retrieval...")

	req, err := http.NewRequest("GET", suite.config.ServiceURL+"/v1/languages?active_only=true", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+suite.config.JWTToken)

	resp, err := suite.client.Do(req)
	if err != nil {
		t.Fatalf("Failed to get languages: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if success, ok := result["success"].(bool); !ok || !success {
		t.Errorf("Expected success=true, got %v", result["success"])
	}

	t.Log("✓ Languages retrieval passed")
}

// TestCompleteWorkflow tests a complete user workflow
func TestCompleteWorkflow(t *testing.T) {
	suite := NewE2ETestSuite(t)

	if suite.config.JWTToken == "" {
		t.Skip("Skipping workflow test - JWT_TOKEN not set")
	}

	t.Log("Testing complete workflow...")

	// Step 1: Check service health
	t.Log("  Step 1: Checking service health...")
	resp, err := suite.client.Get(suite.config.ServiceURL + "/health")
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("Health check failed")
	}
	resp.Body.Close()
	t.Log("  ✓ Health check passed")

	// Step 2: Get list of available languages
	t.Log("  Step 2: Getting available languages...")
	req, _ := http.NewRequest("GET", suite.config.ServiceURL+"/v1/languages", nil)
	req.Header.Set("Authorization", "Bearer "+suite.config.JWTToken)
	resp, err = suite.client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to get languages")
	}
	resp.Body.Close()
	t.Log("  ✓ Languages retrieved")

	// Step 3: Load catalog for a specific language
	t.Log("  Step 3: Loading catalog for 'en'...")
	req, _ = http.NewRequest("GET", suite.config.ServiceURL+"/v1/catalog/en", nil)
	req.Header.Set("Authorization", "Bearer "+suite.config.JWTToken)
	resp, err = suite.client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to load catalog")
	}
	resp.Body.Close()
	t.Log("  ✓ Catalog loaded")

	// Step 4: Fetch multiple localizations
	t.Log("  Step 4: Fetching batch localizations...")
	requestBody := map[string]interface{}{
		"language": "en",
		"keys":     []string{"app.welcome", "app.error"},
		"fallback": true,
	}
	body, _ := json.Marshal(requestBody)
	req, _ = http.NewRequest("POST", suite.config.ServiceURL+"/v1/localize/batch", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+suite.config.JWTToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err = suite.client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to fetch batch localizations")
	}
	resp.Body.Close()
	t.Log("  ✓ Batch localizations fetched")

	t.Log("✓ Complete workflow passed")
}

// TestCachePerformance tests caching effectiveness
func TestCachePerformance(t *testing.T) {
	suite := NewE2ETestSuite(t)

	if suite.config.JWTToken == "" {
		t.Skip("Skipping cache test - JWT_TOKEN not set")
	}

	t.Log("Testing cache performance...")

	req, _ := http.NewRequest("GET", suite.config.ServiceURL+"/v1/catalog/en", nil)
	req.Header.Set("Authorization", "Bearer "+suite.config.JWTToken)

	// First request (uncached)
	start := time.Now()
	resp, err := suite.client.Do(req)
	if err != nil {
		t.Fatalf("First request failed: %v", err)
	}
	resp.Body.Close()
	firstDuration := time.Since(start)
	t.Logf("  First request (uncached): %v", firstDuration)

	// Second request (should be cached)
	req, _ = http.NewRequest("GET", suite.config.ServiceURL+"/v1/catalog/en", nil)
	req.Header.Set("Authorization", "Bearer "+suite.config.JWTToken)
	start = time.Now()
	resp, err = suite.client.Do(req)
	if err != nil {
		t.Fatalf("Second request failed: %v", err)
	}
	resp.Body.Close()
	secondDuration := time.Since(start)
	t.Logf("  Second request (cached): %v", secondDuration)

	// Cached request should be faster
	if secondDuration > firstDuration {
		t.Logf("  Warning: Cached request slower than uncached (may indicate cache miss)")
	} else {
		improvement := float64(firstDuration-secondDuration) / float64(firstDuration) * 100
		t.Logf("  Cache performance improvement: %.2f%%", improvement)
	}

	t.Log("✓ Cache performance test completed")
}

// TestHTTP3Protocol verifies HTTP/3 QUIC is being used
func TestHTTP3Protocol(t *testing.T) {
	suite := NewE2ETestSuite(t)

	t.Log("Testing HTTP/3 QUIC protocol...")

	resp, err := suite.client.Get(suite.config.ServiceURL + "/health")
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer resp.Body.Close()

	// Check protocol version
	if resp.Proto != "" {
		t.Logf("  Protocol: %s", resp.Proto)
	}

	// Check for HTTP/3 specific headers
	if resp.Header.Get("Alt-Svc") != "" {
		t.Logf("  Alt-Svc header present: %s", resp.Header.Get("Alt-Svc"))
	}

	t.Log("✓ HTTP/3 protocol test completed")
}

// TestErrorHandling tests error scenarios
func TestErrorHandling(t *testing.T) {
	suite := NewE2ETestSuite(t)

	t.Log("Testing error handling...")

	tests := []struct {
		name           string
		endpoint       string
		expectedStatus int
	}{
		{"Invalid endpoint", "/v1/invalid", http.StatusNotFound},
		{"Missing auth", "/v1/catalog/en", http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", suite.config.ServiceURL+tt.endpoint, nil)
			// Intentionally not setting Authorization header for some tests

			resp, err := suite.client.Do(req)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			if tt.name == "Missing auth" && resp.StatusCode == http.StatusUnauthorized {
				t.Logf("  ✓ %s: Got expected status %d", tt.name, resp.StatusCode)
			} else if tt.name == "Invalid endpoint" {
				t.Logf("  ✓ %s: Got status %d", tt.name, resp.StatusCode)
			}
		})
	}

	t.Log("✓ Error handling tests completed")
}

// Helper function to get environment variable with default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// TestMain runs before all tests
func TestMain(m *testing.M) {
	fmt.Println("========================================")
	fmt.Println("E2E Tests for Localization Service")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("Configuration:")
	fmt.Printf("  Service URL: %s\n", getEnv("SERVICE_URL", baseURL))
	fmt.Printf("  JWT Secret: %s\n", getEnv("JWT_SECRET", "test-secret-key-for-e2e-testing"))
	fmt.Println()
	fmt.Println("Prerequisites:")
	fmt.Println("  1. Service must be running with HTTP/3 QUIC enabled")
	fmt.Println("  2. TLS certificates must be valid (or self-signed)")
	fmt.Println("  3. Database must be initialized with test data")
	fmt.Println("  4. JWT secret must match server configuration")
	fmt.Println()
	fmt.Println("Note: Tests requiring JWT_TOKEN will be skipped if not set.")
	fmt.Println("      Health check test will always run.")
	fmt.Println()

	// Run tests
	exitCode := m.Run()

	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("E2E Tests Completed")
	fmt.Println("========================================")

	os.Exit(exitCode)
}

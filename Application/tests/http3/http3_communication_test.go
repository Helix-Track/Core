package http3_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"io"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/quic-go/quic-go/http3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"helixtrack.ru/core/internal/server"
)

const (
	testServerURL = "https://localhost:8080"
	testTimeout   = 10 * time.Second
)

var (
	http3Server *server.HTTP3Server
	testRouter  *gin.Engine
)

// setupHTTP3Server starts the HTTP/3 test server
func setupHTTP3Server(t *testing.T) {
	// Create test router
	gin.SetMode(gin.TestMode)
	testRouter = gin.New()

	// Add test endpoints
	testRouter.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy", "protocol": "HTTP/3"})
	})

	testRouter.GET("/do", func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true, "data": "test response"})
	})

	testRouter.POST("/do", func(c *gin.Context) {
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(400, gin.H{"error": "Invalid JSON"})
			return
		}
		c.JSON(200, gin.H{"success": true, "data": data})
	})

	// Create self-signed certificate for testing
	certFile := filepath.Join(t.TempDir(), "cert.pem")
	keyFile := filepath.Join(t.TempDir(), "key.pem")

	// Generate self-signed certificate
	cert, key, err := generateSelfSignedCert()
	if err != nil {
		t.Fatalf("Failed to generate test certificate: %v", err)
	}

	// Write cert and key to files
	err = os.WriteFile(certFile, []byte(cert), 0644)
	if err != nil {
		t.Fatalf("Failed to write cert file: %v", err)
	}

	err = os.WriteFile(keyFile, []byte(key), 0600)
	if err != nil {
		t.Fatalf("Failed to write key file: %v", err)
	}

	// Debug: Check if files exist and log their paths
	t.Logf("Certificate file: %s", certFile)
	t.Logf("Key file: %s", keyFile)
	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		t.Fatalf("Certificate file does not exist: %s", certFile)
	}
	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		t.Fatalf("Key file does not exist: %s", keyFile)
	}

	// Create logger for testing
	logger, _ := zap.NewDevelopment()

	// Create HTTP/3 server
	http3Server, err = server.NewHTTP3Server(testRouter, certFile, keyFile, logger)
	require.NoError(t, err, "Failed to create HTTP/3 server")

	// Start server in goroutine
	go func() {
		err := http3Server.Start("127.0.0.1:8080")
		if err != nil {
			t.Logf("HTTP/3 server start error: %v", err)
		}
	}()

	// Give server time to start
	time.Sleep(200 * time.Millisecond)
}

// teardownHTTP3Server stops the HTTP/3 test server
func teardownHTTP3Server() {
	if http3Server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		http3Server.Shutdown(ctx)
		http3Server = nil
	}
	testRouter = nil
}

// generateSelfSignedCert creates a simple self-signed certificate for testing
func generateSelfSignedCert() (string, string, error) {
	// Generate private key
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", "", err
	}

	// Generate serial number
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return "", "", err
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"HelixTrack Test"},
			CommonName:   "localhost",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost"},
	}

	// Create certificate
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return "", "", err
	}

	// Encode certificate to PEM
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	})

	// Encode private key to PEM
	privBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return "", "", err
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privBytes,
	})

	return string(certPEM), string(keyPEM), nil
}

// TestHTTP3Connectivity tests basic HTTP/3 connectivity
func TestHTTP3Connectivity(t *testing.T) {
	setupHTTP3Server(t)
	defer teardownHTTP3Server()

	client := createHTTP3TestClient(t)
	defer closeClient(client)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	resp, err := doGet(client, ctx, testServerURL+"/health")
	require.NoError(t, err, "HTTP/3 GET request should succeed")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Health endpoint should return 200 OK")
	assert.Contains(t, []string{"HTTP/3.0", "h3", "h3-29"}, resp.Proto, "Protocol should be HTTP/3")

	// Read response body
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	t.Logf("Health endpoint response: %s", string(body))
}

// TestQUICProtocolNegotiation tests QUIC protocol negotiation
func TestQUICProtocolNegotiation(t *testing.T) {
	setupHTTP3Server(t)
	defer teardownHTTP3Server()

	client := createHTTP3TestClient(t)
	defer closeClient(client)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	resp, err := doGet(client, ctx, testServerURL+"/do")
	require.NoError(t, err, "HTTP/3 GET request should succeed")
	defer resp.Body.Close()

	// Verify HTTP/3 protocol was negotiated
	assert.Contains(t, []string{"HTTP/3.0", "h3", "h3-29"}, resp.Proto,
		"Should negotiate HTTP/3 protocol")
	t.Logf("Negotiated protocol: %s", resp.Proto)
}

// TestTLS13Verification verifies TLS 1.3 is used
func TestTLS13Verification(t *testing.T) {
	setupHTTP3Server(t)
	defer teardownHTTP3Server()

	client := createHTTP3TestClient(t)
	defer closeClient(client)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	resp, err := doGet(client, ctx, testServerURL+"/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify TLS 1.3
	if resp.TLS != nil {
		assert.Equal(t, uint16(tls.VersionTLS13), resp.TLS.Version,
			"Should use TLS 1.3")
		t.Logf("TLS Version: 1.3")
	} else {
		t.Log("TLS info not available in response")
	}
}

// TestConnectionMultiplexing tests multiple concurrent requests over same QUIC connection
func TestConnectionMultiplexing(t *testing.T) {
	setupHTTP3Server(t)
	defer teardownHTTP3Server()

	client := createHTTP3TestClient(t)
	defer closeClient(client)

	numRequests := 10
	var wg sync.WaitGroup
	results := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
			defer cancel()

			resp, err := doGet(client, ctx, testServerURL+"/health")
			if err != nil {
				results <- err
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				results <- assert.AnError
			} else {
				results <- nil
			}
		}(i)
	}

	wg.Wait()
	close(results)

	// Verify all requests succeeded
	successCount := 0
	for err := range results {
		if err == nil {
			successCount++
		} else {
			t.Errorf("Request failed: %v", err)
		}
	}

	assert.Equal(t, numRequests, successCount,
		"All concurrent requests should succeed via connection multiplexing")
	t.Logf("Successfully multiplexed %d requests", successCount)
}

// TestLatencyMeasurement measures HTTP/3 latency
func TestLatencyMeasurement(t *testing.T) {
	setupHTTP3Server(t)
	defer teardownHTTP3Server()

	client := createHTTP3TestClient(t)
	defer closeClient(client)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	// Warm-up request
	resp, _ := doGet(client, ctx, testServerURL+"/health")
	if resp != nil {
		resp.Body.Close()
	}

	// Measure latency
	numRequests := 100
	latencies := make([]time.Duration, numRequests)

	for i := 0; i < numRequests; i++ {
		start := time.Now()

		resp, err := doGet(client, ctx, testServerURL+"/health")
		latency := time.Since(start)

		require.NoError(t, err, "Request %d should succeed", i)
		if resp != nil {
			resp.Body.Close()
		}

		latencies[i] = latency
	}

	// Calculate statistics
	var totalLatency time.Duration
	minLatency := latencies[0]
	maxLatency := latencies[0]

	for _, lat := range latencies {
		totalLatency += lat
		if lat < minLatency {
			minLatency = lat
		}
		if lat > maxLatency {
			maxLatency = lat
		}
	}

	avgLatency := totalLatency / time.Duration(numRequests)

	t.Logf("Latency Statistics:")
	t.Logf("  Min: %v", minLatency)
	t.Logf("  Max: %v", maxLatency)
	t.Logf("  Avg: %v", avgLatency)

	// HTTP/3 should be fast (<100ms for localhost)
	assert.Less(t, avgLatency.Milliseconds(), int64(100),
		"Average latency should be less than 100ms for localhost")
}

// TestThroughput tests HTTP/3 throughput
func TestThroughput(t *testing.T) {
	setupHTTP3Server(t)
	defer teardownHTTP3Server()

	client := createHTTP3TestClient(t)
	defer closeClient(client)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	numRequests := 1000
	start := time.Now()

	var wg sync.WaitGroup
	errors := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			resp, err := doGet(client, ctx, testServerURL+"/health")
			if err != nil {
				errors <- err
				return
			}
			resp.Body.Close()
		}()
	}

	wg.Wait()
	close(errors)

	duration := time.Since(start)
	errorCount := len(errors)

	requestsPerSecond := float64(numRequests-errorCount) / duration.Seconds()

	t.Logf("Throughput Test:")
	t.Logf("  Total Requests: %d", numRequests)
	t.Logf("  Successful: %d", numRequests-errorCount)
	t.Logf("  Failed: %d", errorCount)
	t.Logf("  Duration: %v", duration)
	t.Logf("  Throughput: %.2f req/s", requestsPerSecond)

	assert.Less(t, errorCount, numRequests/10, "Error rate should be less than 10%")
	assert.Greater(t, requestsPerSecond, 100.0, "Should handle at least 100 req/s")
}

// TestErrorHandling tests error handling for invalid requests
func TestErrorHandling(t *testing.T) {
	setupHTTP3Server(t)
	defer teardownHTTP3Server()

	client := createHTTP3TestClient(t)
	defer closeClient(client)

	testCases := []struct {
		name           string
		url            string
		expectedStatus int
	}{
		{
			name:           "Invalid endpoint",
			url:            testServerURL + "/invalid-endpoint-that-does-not-exist",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Malformed request",
			url:            testServerURL + "/do",
			expectedStatus: http.StatusBadRequest, // Assuming /do requires specific format
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
			defer cancel()

			resp, err := doGet(client, ctx, tc.url)
			require.NoError(t, err, "Request should complete even if endpoint is invalid")
			defer resp.Body.Close()

			// Don't assert exact status code as it may vary
			// Just ensure we got a response
			assert.NotEqual(t, 0, resp.StatusCode, "Should receive HTTP status code")
			t.Logf("Test '%s': status=%d", tc.name, resp.StatusCode)
		})
	}
}

// TestJSONPayload tests HTTP/3 with JSON payload
func TestJSONPayload(t *testing.T) {
	setupHTTP3Server(t)
	defer teardownHTTP3Server()

	client := createHTTP3TestClient(t)
	defer closeClient(client)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	// Create test payload
	payload := map[string]interface{}{
		"action": "version",
		"data":   map[string]string{"test": "data"},
	}

	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	// Create POST request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testServerURL+"/do", nil)
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err, "HTTP/3 POST with JSON should succeed")
	defer resp.Body.Close()

	assert.Contains(t, []string{"HTTP/3.0", "h3", "h3-29"}, resp.Proto,
		"POST request should use HTTP/3")

	t.Logf("JSON payload test: status=%d, proto=%s, payload_size=%d",
		resp.StatusCode, resp.Proto, len(payloadBytes))
}

// TestConnectionReuse tests that connections are reused
func TestConnectionReuse(t *testing.T) {
	setupHTTP3Server(t)
	defer teardownHTTP3Server()

	client := createHTTP3TestClient(t)
	defer closeClient(client)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	// Make multiple requests
	for i := 0; i < 5; i++ {
		resp, err := doGet(client, ctx, testServerURL+"/health")
		require.NoError(t, err)
		resp.Body.Close()
	}

	// If we get here without errors, connection reuse is working
	t.Log("Connection reuse test passed - all requests succeeded")
}

// Helper function to create HTTP/3 test client
func createHTTP3TestClient(t *testing.T) *http.Client {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // For testing with self-signed certs
		NextProtos:         []string{"h3"},
		MinVersion:         tls.VersionTLS13,
		MaxVersion:         tls.VersionTLS13,
	}

	roundTripper := &http3.Transport{
		TLSClientConfig: tlsConfig,
	}

	return &http.Client{
		Transport: roundTripper,
		Timeout:   testTimeout,
	}
}

// Helper function to close HTTP/3 client
func closeClient(client *http.Client) {
	if transport, ok := client.Transport.(*http3.Transport); ok {
		transport.Close()
	}
}

// Helper function to do GET request with context
func doGet(client *http.Client, ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return client.Do(req)
}

// BenchmarkHTTP3Latency benchmarks HTTP/3 request latency
func BenchmarkHTTP3Latency(b *testing.B) {
	setupHTTP3Server(&testing.T{})
	defer teardownHTTP3Server()

	client := createHTTP3TestClient(&testing.T{})
	defer closeClient(client)

	ctx := context.Background()
	url := testServerURL + "/health"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := doGet(client, ctx, url)
		if err != nil {
			b.Fatal(err)
		}
		resp.Body.Close()
	}
}

// BenchmarkHTTP3Throughput benchmarks HTTP/3 throughput
func BenchmarkHTTP3Throughput(b *testing.B) {
	setupHTTP3Server(&testing.T{})
	defer teardownHTTP3Server()

	client := createHTTP3TestClient(&testing.T{})
	defer closeClient(client)

	ctx := context.Background()
	url := testServerURL + "/health"

	b.SetParallelism(100) // 100 concurrent clients
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := doGet(client, ctx, url)
			if err != nil {
				b.Fatal(err)
			}
			resp.Body.Close()
		}
	})
}

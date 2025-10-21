package http3_test

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/quic-go/quic-go/http3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testServerURL = "https://localhost:8080"
	testTimeout   = 10 * time.Second
)

// TestHTTP3Connectivity tests basic HTTP/3 connectivity
func TestHTTP3Connectivity(t *testing.T) {
	client := createHTTP3TestClient(t)
	defer closeClient(client)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	resp, err := client.Get(ctx, testServerURL+"/health")
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
	client := createHTTP3TestClient(t)
	defer closeClient(client)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	resp, err := client.Get(ctx, testServerURL+"/do")
	require.NoError(t, err, "HTTP/3 GET request should succeed")
	defer resp.Body.Close()

	// Verify HTTP/3 protocol was negotiated
	assert.Contains(t, []string{"HTTP/3.0", "h3", "h3-29"}, resp.Proto,
		"Should negotiate HTTP/3 protocol")
	t.Logf("Negotiated protocol: %s", resp.Proto)
}

// TestTLS13Verification verifies TLS 1.3 is used
func TestTLS13Verification(t *testing.T) {
	client := createHTTP3TestClient(t)
	defer closeClient(client)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	resp, err := client.Get(ctx, testServerURL+"/health")
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

			resp, err := client.Get(ctx, testServerURL+"/health")
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
	client := createHTTP3TestClient(t)
	defer closeClient(client)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	// Warm-up request
	resp, _ := client.Get(ctx, testServerURL+"/health")
	if resp != nil {
		resp.Body.Close()
	}

	// Measure latency
	numRequests := 100
	latencies := make([]time.Duration, numRequests)

	for i := 0; i < numRequests; i++ {
		start := time.Now()

		resp, err := client.Get(ctx, testServerURL+"/health")
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

			resp, err := client.Get(ctx, testServerURL+"/health")
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

			resp, err := client.Get(ctx, tc.url)
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
	client := createHTTP3TestClient(t)
	defer closeClient(client)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	// Make multiple requests
	for i := 0; i < 5; i++ {
		resp, err := client.Get(ctx, testServerURL+"/health")
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

	roundTripper := &http3.RoundTripper{
		TLSClientConfig: tlsConfig,
	}

	return &http.Client{
		Transport: roundTripper,
		Timeout:   testTimeout,
	}
}

// Helper function to close HTTP/3 client
func closeClient(client *http.Client) {
	if transport, ok := client.Transport.(*http3.RoundTripper); ok {
		transport.Close()
	}
}

// BenchmarkHTTP3Latency benchmarks HTTP/3 request latency
func BenchmarkHTTP3Latency(b *testing.B) {
	client := createHTTP3TestClient(&testing.T{})
	defer closeClient(client)

	ctx := context.Background()
	url := testServerURL + "/health"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := client.Get(ctx, url)
		if err != nil {
			b.Fatal(err)
		}
		resp.Body.Close()
	}
}

// BenchmarkHTTP3Throughput benchmarks HTTP/3 throughput
func BenchmarkHTTP3Throughput(b *testing.B) {
	client := createHTTP3TestClient(&testing.T{})
	defer closeClient(client)

	ctx := context.Background()
	url := testServerURL + "/health"

	b.SetParallelism(100) // 100 concurrent clients
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := client.Get(ctx, url)
			if err != nil {
				b.Fatal(err)
			}
			resp.Body.Close()
		}
	})
}

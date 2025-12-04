package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// TestNewHTTP3ClientWithConfig tests client creation with custom config
func TestNewHTTP3ClientWithConfig(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &HTTP3ClientConfig{
		SkipVerify:      true,
		Timeout:         60 * time.Second,
		MaxIdleTimeout:  60 * time.Second,
		EnableDatagrams: false,
	}

	client := NewHTTP3Client(logger, config)

	assert.NotNil(t, client)
	assert.Equal(t, 60*time.Second, client.client.Timeout)
	assert.True(t, client.tlsConfig.InsecureSkipVerify)
}

// TestHTTP3Client_Get tests GET request functionality
func TestHTTP3Client_Get(t *testing.T) {
	// Create a TLS test server
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/test", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	// Create HTTP3 client
	logger := zaptest.NewLogger(t)
	config := &HTTP3ClientConfig{SkipVerify: true}
	client := NewHTTP3Client(logger, config)

	// For testing, replace with HTTP transport that accepts the test server's cert
	client.client.Transport = server.Client().Transport

	ctx := context.Background()
	url := server.URL + "/test"

	resp, err := client.Get(ctx, url)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, `{"status": "ok"}`, string(body))
	resp.Body.Close()
}

// TestHTTP3Client_Post tests POST request with JSON body
func TestHTTP3Client_Post(t *testing.T) {
	// Create a TLS test server
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/test", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var payload map[string]interface{}
		err = json.Unmarshal(body, &payload)
		require.NoError(t, err)
		assert.Equal(t, "test_value", payload["test_key"])

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": 123, "created": true}`))
	}))
	defer server.Close()

	// Create HTTP3 client
	logger := zaptest.NewLogger(t)
	config := &HTTP3ClientConfig{SkipVerify: true}
	client := NewHTTP3Client(logger, config)
	client.client.Transport = server.Client().Transport

	ctx := context.Background()
	url := server.URL + "/api/test"
	payload := map[string]string{"test_key": "test_value"}

	resp, err := client.Post(ctx, url, payload)

	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), `"id": 123`)
	assert.Contains(t, string(body), `"created": true`)
	resp.Body.Close()
}

// TestHTTP3Client_PostRaw tests POST request with raw body
func TestHTTP3Client_PostRaw(t *testing.T) {
	testData := "raw test data"

	// Create a TLS test server
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "text/plain", r.Header.Get("Content-Type"))

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.Equal(t, testData, string(body))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("received"))
	}))
	defer server.Close()

	// Create HTTP3 client
	logger := zaptest.NewLogger(t)
	config := &HTTP3ClientConfig{SkipVerify: true}
	client := NewHTTP3Client(logger, config)
	client.client.Transport = server.Client().Transport

	ctx := context.Background()
	url := server.URL + "/upload"
	reader := strings.NewReader(testData)

	resp, err := client.PostRaw(ctx, url, reader, "text/plain")

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, "received", string(body))
	resp.Body.Close()
}

// TestHTTP3Client_Do tests custom request functionality
func TestHTTP3Client_Do(t *testing.T) {
	// Create a TLS test server
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/custom", r.URL.Path)
		assert.Equal(t, "custom-value", r.Header.Get("X-Custom-Header"))

		w.Header().Set("X-Response-Header", "response-value")
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("custom response"))
	}))
	defer server.Close()

	// Create HTTP3 client
	logger := zaptest.NewLogger(t)
	config := &HTTP3ClientConfig{SkipVerify: true}
	client := NewHTTP3Client(logger, config)
	client.client.Transport = server.Client().Transport

	ctx := context.Background()
	url := server.URL + "/custom"

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, nil)
	require.NoError(t, err)
	req.Header.Set("X-Custom-Header", "custom-value")

	resp, err := client.Do(req)

	require.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
	assert.Equal(t, "response-value", resp.Header.Get("X-Response-Header"))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, "custom response", string(body))
	resp.Body.Close()
}

// TestGetProtocol tests protocol detection
func TestGetProtocol(t *testing.T) {
	testCases := []struct {
		name     string
		response *http.Response
		expected string
	}{
		{
			name:     "nil response",
			response: nil,
			expected: "unknown",
		},
		{
			name:     "HTTP/1.1 response",
			response: &http.Response{Proto: "HTTP/1.1"},
			expected: "HTTP/1.1",
		},
		{
			name:     "HTTP/2 response",
			response: &http.Response{Proto: "HTTP/2.0"},
			expected: "HTTP/2.0",
		},
		{
			name:     "HTTP/3 response",
			response: &http.Response{Proto: "HTTP/3.0"},
			expected: "HTTP/3.0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := GetProtocol(tc.response)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestIsHTTP3 tests HTTP/3 protocol detection
func TestIsHTTP3(t *testing.T) {
	testCases := []struct {
		name     string
		response *http.Response
		expected bool
	}{
		{
			name:     "nil response",
			response: nil,
			expected: false,
		},
		{
			name:     "HTTP/1.1 response",
			response: &http.Response{Proto: "HTTP/1.1"},
			expected: false,
		},
		{
			name:     "HTTP/2 response",
			response: &http.Response{Proto: "HTTP/2.0"},
			expected: false,
		},
		{
			name:     "HTTP/3.0 response",
			response: &http.Response{Proto: "HTTP/3.0"},
			expected: true,
		},
		{
			name:     "h3 response",
			response: &http.Response{Proto: "h3"},
			expected: true,
		},
		{
			name:     "h3-29 response",
			response: &http.Response{Proto: "h3-29"},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsHTTP3(tc.response)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestHTTP3Client_Get_ErrorHandling tests error handling in GET requests
func TestHTTP3Client_Get_ErrorHandling(t *testing.T) {
	logger := zaptest.NewLogger(t)
	client := NewHTTP3Client(logger, nil)

	ctx := context.Background()

	// Test with invalid URL
	_, err := client.Get(ctx, "invalid-url")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported protocol scheme")
}

// TestHTTP3Client_Post_ErrorHandling tests error handling in POST requests
func TestHTTP3Client_Post_ErrorHandling(t *testing.T) {
	logger := zaptest.NewLogger(t)
	client := NewHTTP3Client(logger, nil)

	ctx := context.Background()

	// Test with invalid URL
	_, err := client.Post(ctx, "invalid-url", map[string]string{"test": "data"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported protocol scheme")
}

// TestHTTP3Client_PostRaw_ErrorHandling tests error handling in PostRaw requests
func TestHTTP3Client_PostRaw_ErrorHandling(t *testing.T) {
	logger := zaptest.NewLogger(t)
	client := NewHTTP3Client(logger, nil)

	ctx := context.Background()

	// Test with invalid URL
	_, err := client.PostRaw(ctx, "invalid-url", strings.NewReader("test"), "text/plain")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported protocol scheme")
}

// TestHTTP3Client_Do_ErrorHandling tests error handling in Do requests
func TestHTTP3Client_Do_ErrorHandling(t *testing.T) {
	logger := zaptest.NewLogger(t)
	client := NewHTTP3Client(logger, nil)

	// Test with nil request
	_, err := client.Do(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "request cannot be nil")
}

// BenchmarkHTTP3Client_Get benchmarks GET request performance
func BenchmarkHTTP3Client_Get(b *testing.B) {
	// Create a TLS test server
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	logger := zaptest.NewLogger(b)
	config := &HTTP3ClientConfig{SkipVerify: true}
	client := NewHTTP3Client(logger, config)
	client.client.Transport = server.Client().Transport

	ctx := context.Background()
	url := server.URL + "/benchmark"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := client.Get(ctx, url)
		if err != nil {
			b.Fatal(err)
		}
		resp.Body.Close()
	}
}

// BenchmarkHTTP3Client_Post benchmarks POST request performance
func BenchmarkHTTP3Client_Post(b *testing.B) {
	// Create a TLS test server
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result": "success"}`))
	}))
	defer server.Close()

	logger := zaptest.NewLogger(b)
	config := &HTTP3ClientConfig{SkipVerify: true}
	client := NewHTTP3Client(logger, config)
	client.client.Transport = server.Client().Transport

	ctx := context.Background()
	url := server.URL + "/benchmark"
	payload := map[string]string{"test": "data"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := client.Post(ctx, url, payload)
		if err != nil {
			b.Fatal(err)
		}
		resp.Body.Close()
	}
}

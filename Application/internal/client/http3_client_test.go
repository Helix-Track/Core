package client

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestDefaultHTTP3ClientConfig tests the default configuration
func TestDefaultHTTP3ClientConfig(t *testing.T) {
	config := DefaultHTTP3ClientConfig()
	
	assert.NotNil(t, config)
	assert.Equal(t, 30*time.Second, config.Timeout)
	assert.Equal(t, 30*time.Second, config.MaxIdleTimeout)
	assert.False(t, config.SkipVerify)
	assert.True(t, config.EnableDatagrams)
}

// TestNewHTTP3Client tests client creation
func TestNewHTTP3Client(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultHTTP3ClientConfig()
	
	client := NewHTTP3Client(logger, config)
	require.NotNil(t, client)
	
	defer client.Close()
	
	// Test that client has required fields
	assert.NotNil(t, client.client)
	assert.NotNil(t, client.logger)
	assert.NotNil(t, client.tlsConfig)
}

// TestNewHTTP3Client_DefaultConfig tests client creation with default config
func TestNewHTTP3Client_DefaultConfig(t *testing.T) {
	logger := zap.NewNop()
	
	client := NewHTTP3Client(logger, nil)
	require.NotNil(t, client)
	
	defer client.Close()
	
	assert.NotNil(t, client.client)
}

// TestHTTP3Client_Close tests client cleanup
func TestHTTP3Client_Close(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultHTTP3ClientConfig()
	client := NewHTTP3Client(logger, config)
	require.NotNil(t, client)
	
	// Close should work without error
	err := client.Close()
	assert.NoError(t, err)
	
	// Double close should not panic
	err = client.Close()
	assert.NoError(t, err)
}

// TestHTTP3Client_ErrorHandling tests various error scenarios
func TestHTTP3Client_ErrorHandling(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultHTTP3ClientConfig()
	client := NewHTTP3Client(logger, config)
	require.NotNil(t, client)
	
	defer client.Close()
	
	// Test with canceled context
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately
	
	resp, err := client.Get(canceledCtx, "https://httpbin.org/get")
	assert.Error(t, err)
	assert.Nil(t, resp)
}

// TestHTTP3Client_Timeout tests request timeout
func TestHTTP3Client_Timeout(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultHTTP3ClientConfig()
	config.Timeout = 100 * time.Millisecond // Very short timeout
	
	client := NewHTTP3Client(logger, config)
	require.NotNil(t, client)
	
	defer client.Close()
	
	// This should timeout quickly
	ctx := context.Background()
	resp, err := client.Get(ctx, "https://httpbin.org/delay/1") // 1 second delay
	
	// Should be timeout error (or network error in test environment)
	assert.Error(t, err)
	assert.Nil(t, resp)
}
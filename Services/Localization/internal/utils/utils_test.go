package utils

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestNewLogger tests logger creation with different levels and formats
func TestNewLogger(t *testing.T) {
	tests := []struct {
		name   string
		level  string
		format string
	}{
		{"debug json", "debug", "json"},
		{"info json", "info", "json"},
		{"warn json", "warn", "json"},
		{"error json", "error", "json"},
		{"info console", "info", "console"},
		{"default level", "unknown", "json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewLogger(tt.level, tt.format)
			require.NoError(t, err)
			assert.NotNil(t, logger)

			// Verify logger can be used
			logger.Info("test message")
			logger.Sync()
		})
	}
}

// TestNewServiceRegistry tests service registry creation
func TestNewServiceRegistry(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	sr, err := NewServiceRegistry("consul", "localhost:8500", "test-service", 8085, logger)
	require.NoError(t, err)
	assert.NotNil(t, sr)
	assert.Equal(t, "consul", sr.provider)
	assert.Equal(t, "test-service", sr.serviceName)
	assert.Equal(t, 8085, sr.servicePort)
	assert.Equal(t, "localhost:8500", sr.consulAddress)
}

// TestServiceRegistry_RegisterConsul tests Consul registration
func TestServiceRegistry_RegisterConsul(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	sr, err := NewServiceRegistry("consul", "localhost:8500", "test-service", 8085, logger)
	require.NoError(t, err)

	// Should not error (it's a placeholder implementation)
	err = sr.Register()
	assert.NoError(t, err)
}

// TestServiceRegistry_RegisterEtcd tests etcd registration
func TestServiceRegistry_RegisterEtcd(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	sr, err := NewServiceRegistry("etcd", "", "test-service", 8085, logger)
	require.NoError(t, err)

	// Should not error (it's a placeholder implementation)
	err = sr.Register()
	assert.NoError(t, err)
}

// TestServiceRegistry_RegisterUnsupported tests unsupported provider
func TestServiceRegistry_RegisterUnsupported(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	sr, err := NewServiceRegistry("unsupported", "", "test-service", 8085, logger)
	require.NoError(t, err)

	err = sr.Register()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported service discovery provider")
}

// TestServiceRegistry_DeregisterConsul tests Consul deregistration
func TestServiceRegistry_DeregisterConsul(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	sr, err := NewServiceRegistry("consul", "localhost:8500", "test-service", 8085, logger)
	require.NoError(t, err)

	err = sr.Deregister()
	assert.NoError(t, err)
}

// TestServiceRegistry_DeregisterEtcd tests etcd deregistration
func TestServiceRegistry_DeregisterEtcd(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	sr, err := NewServiceRegistry("etcd", "", "test-service", 8085, logger)
	require.NoError(t, err)

	err = sr.Deregister()
	assert.NoError(t, err)
}

// TestServiceRegistry_DeregisterUnsupported tests unsupported provider deregistration
func TestServiceRegistry_DeregisterUnsupported(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	sr, err := NewServiceRegistry("unsupported", "", "test-service", 8085, logger)
	require.NoError(t, err)

	err = sr.Deregister()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported service discovery provider")
}

// TestIsPortAvailable tests port availability checking
func TestIsPortAvailable(t *testing.T) {
	// Test with a port that should be available (high port number)
	available := IsPortAvailable(54321)
	assert.True(t, available)

	// Test with a port that's in use
	listener, err := net.Listen("tcp", ":54322")
	require.NoError(t, err)
	defer listener.Close()

	available = IsPortAvailable(54322)
	assert.False(t, available)
}

// TestFindAvailablePort tests finding an available port
func TestFindAvailablePort(t *testing.T) {
	// Test finding preferred port when available
	port, err := FindAvailablePort(54323, []int{54323, 54333})
	require.NoError(t, err)
	assert.Equal(t, 54323, port)

	// Test finding alternative port when preferred is taken
	listener, err := net.Listen("tcp", ":54324")
	require.NoError(t, err)
	defer listener.Close()

	// Small delay to ensure port is occupied
	time.Sleep(50 * time.Millisecond)

	port, err = FindAvailablePort(54324, []int{54324, 54334})
	require.NoError(t, err)
	assert.NotEqual(t, 54324, port)
	assert.True(t, port >= 54324 && port <= 54334)
}

// TestFindAvailablePort_AllTaken tests when all ports in range are taken
func TestFindAvailablePort_AllTaken(t *testing.T) {
	// Occupy all ports in a small range
	portStart := 54400
	portEnd := 54402
	var listeners []net.Listener

	for port := portStart; port <= portEnd; port++ {
		listener, err := net.Listen("tcp", ":"+fmt.Sprint(port))
		if err == nil {
			listeners = append(listeners, listener)
		}
	}
	defer func() {
		for _, l := range listeners {
			l.Close()
		}
	}()

	time.Sleep(50 * time.Millisecond) // Ensure ports are occupied

	// Try to find a port (should fail as all are taken)
	_, err := FindAvailablePort(portStart, []int{portStart, portEnd})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no available port found")
}

// TestFindAvailablePort_EmptyRange tests with no port range
func TestFindAvailablePort_EmptyRange(t *testing.T) {
	// Occupy the preferred port
	listener, err := net.Listen("tcp", ":54350")
	require.NoError(t, err)
	defer listener.Close()

	time.Sleep(50 * time.Millisecond)

	// Try to find port with no range provided
	_, err = FindAvailablePort(54350, []int{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "preferred port 54350 not available and no port range provided")
}

// TestFindAvailablePort_PreferredAvailable tests when preferred port is available
func TestFindAvailablePort_PreferredAvailable(t *testing.T) {
	preferredPort := 54351

	// Ensure port is available
	if listener, err := net.Listen("tcp", ":54351"); err == nil {
		listener.Close()
		time.Sleep(150 * time.Millisecond) // Wait for port to be released
	}

	port, err := FindAvailablePort(preferredPort, []int{54351, 54361})
	require.NoError(t, err)
	assert.Equal(t, preferredPort, port)
}

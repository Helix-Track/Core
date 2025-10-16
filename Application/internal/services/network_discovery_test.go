package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewNetworkDiscoveryService(t *testing.T) {
	service := NewNetworkDiscoveryService(8080, "localhost")
	assert.NotNil(t, service)
	assert.Equal(t, 8080, service.GetServicePort())
	assert.Equal(t, "localhost", service.GetServiceHost())
	assert.Equal(t, "http://localhost:8080", service.GetServiceURL())
}

func TestNetworkDiscoveryService_StartStop(t *testing.T) {
	service := NewNetworkDiscoveryService(8080, "localhost")

	// Start service
	err := service.Start()
	assert.NoError(t, err)
	assert.True(t, service.IsRunning())

	// Stop service
	err = service.Stop()
	assert.NoError(t, err)
	assert.False(t, service.IsRunning())
}

func TestNetworkDiscoveryService_DiscoverServices(t *testing.T) {
	service := NewNetworkDiscoveryService(8080, "localhost")

	// Start service to have something to discover
	err := service.Start()
	assert.NoError(t, err)
	defer service.Stop()

	// Give it a moment to start broadcasting
	time.Sleep(100 * time.Millisecond)

	// Try to discover (may not find anything in test environment)
	services, err := service.DiscoverServices(1 * time.Second)
	assert.NoError(t, err)
	// In test environment, may not find services, but should not error
	assert.IsType(t, []ServiceInfo{}, services)
}

func TestNetworkDiscoveryService_Getters(t *testing.T) {
	service := NewNetworkDiscoveryService(9090, "192.168.1.100")

	assert.Equal(t, 9090, service.GetServicePort())
	assert.Equal(t, "192.168.1.100", service.GetServiceHost())
	assert.Equal(t, "http://192.168.1.100:9090", service.GetServiceURL())
}

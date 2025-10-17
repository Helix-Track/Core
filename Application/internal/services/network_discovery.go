package services

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"
	"helixtrack.ru/core/internal/logger"
)

// ServiceInfo represents the service information broadcasted
type ServiceInfo struct {
	Name    string `json:"name"`
	Host    string `json:"host"`
	Port    int    `json:"port"`
	Version string `json:"version"`
	API     string `json:"api"`
}

// NetworkDiscoveryService handles UDP-based service discovery for the Core
type NetworkDiscoveryService struct {
	port          int
	host          string
	broadcastAddr string
	conn          *net.UDPConn
	running       bool
	mu            sync.RWMutex
	serviceInfo   ServiceInfo
}

// NewNetworkDiscoveryService creates a new network discovery service
func NewNetworkDiscoveryService(port int, host string) *NetworkDiscoveryService {
	return &NetworkDiscoveryService{
		port:          port,
		host:          host,
		broadcastAddr: "255.255.255.255:9999", // Standard broadcast port for discovery
		running:       false,
		serviceInfo: ServiceInfo{
			Name:    "HelixTrack Core",
			Host:    host,
			Port:    port,
			Version: "1.0.0",
			API:     "v1",
		},
	}
}

// Start begins broadcasting the service via UDP
func (s *NetworkDiscoveryService) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("network discovery service already running")
	}

	logger.Info("Starting network discovery service",
		zap.String("host", s.host),
		zap.Int("port", s.port))

	// Create UDP connection
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		return fmt.Errorf("failed to create UDP connection: %w", err)
	}

	s.conn = conn
	s.running = true

	// Start broadcasting in a goroutine
	go s.broadcastLoop()

	logger.Info("Network discovery service started successfully")

	return nil
}

// Stop stops broadcasting the service
func (s *NetworkDiscoveryService) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	logger.Info("Stopping network discovery service")

	s.running = false

	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}

	logger.Info("Network discovery service stopped")

	return nil
}

// IsRunning returns whether the service is currently running
func (s *NetworkDiscoveryService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// DiscoverServices discovers available HelixTrack services on the network
func (s *NetworkDiscoveryService) DiscoverServices(timeout time.Duration) ([]ServiceInfo, error) {
	// Create a listening connection for discovery
	listenAddr, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve listen address: %w", err)
	}

	conn, err := net.ListenUDP("udp", listenAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create listen connection: %w", err)
	}
	defer conn.Close()

	// Send discovery request
	broadcastAddr, err := net.ResolveUDPAddr("udp", s.broadcastAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve broadcast address: %w", err)
	}

	request := map[string]string{"action": "discover", "service": "helixtrack"}
	requestData, _ := json.Marshal(request)

	_, err = conn.WriteToUDP(requestData, broadcastAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to send discovery request: %w", err)
	}

	// Listen for responses
	conn.SetReadDeadline(time.Now().Add(timeout))

	var services []ServiceInfo
	buffer := make([]byte, 1024)

	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				break
			}
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		var response ServiceInfo
		if err := json.Unmarshal(buffer[:n], &response); err != nil {
			continue // Skip invalid responses
		}

		services = append(services, response)
	}

	return services, nil
}

// broadcastLoop continuously broadcasts the service information
func (s *NetworkDiscoveryService) broadcastLoop() {
	ticker := time.NewTicker(5 * time.Second) // Broadcast every 5 seconds
	defer ticker.Stop()

	broadcastAddr, err := net.ResolveUDPAddr("udp", s.broadcastAddr)
	if err != nil {
		logger.Error("Failed to resolve broadcast address", zap.Error(err))
		return
	}

	data, err := json.Marshal(s.serviceInfo)
	if err != nil {
		logger.Error("Failed to marshal service info", zap.Error(err))
		return
	}

	for {
		select {
		case <-ticker.C:
			s.mu.RLock()
			if !s.running {
				s.mu.RUnlock()
				return
			}
			s.mu.RUnlock()

			_, err := s.conn.WriteToUDP(data, broadcastAddr)
			if err != nil {
				logger.Warn("Failed to broadcast service info", zap.Error(err))
			}
		}
	}
}

// GetServiceURL returns the full URL for the service
func (s *NetworkDiscoveryService) GetServiceURL() string {
	return fmt.Sprintf("http://%s:%d", s.host, s.port)
}

// GetServicePort returns the service port
func (s *NetworkDiscoveryService) GetServicePort() int {
	return s.port
}

// GetServiceHost returns the service host
func (s *NetworkDiscoveryService) GetServiceHost() string {
	return s.host
}

// UpdatePort updates the service port and service info
func (s *NetworkDiscoveryService) UpdatePort(port int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.port = port
	s.serviceInfo.Port = port
	s.serviceInfo.Host = s.host

	logger.Info("Network discovery service port updated",
		zap.Int("port", port),
		zap.String("host", s.host),
	)
}

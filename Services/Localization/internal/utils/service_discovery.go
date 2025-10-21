package utils

import (
	"fmt"
	"net"
	"time"

	"go.uber.org/zap"
)

// ServiceRegistry handles service discovery and registration
type ServiceRegistry struct {
	provider      string
	serviceName   string
	servicePort   int
	consulAddress string
	logger        *zap.Logger
}

// NewServiceRegistry creates a new service registry
func NewServiceRegistry(provider, consulAddress, serviceName string, servicePort int, logger *zap.Logger) (*ServiceRegistry, error) {
	return &ServiceRegistry{
		provider:      provider,
		serviceName:   serviceName,
		servicePort:   servicePort,
		consulAddress: consulAddress,
		logger:        logger,
	}, nil
}

// Register registers the service with the discovery provider
func (sr *ServiceRegistry) Register() error {
	switch sr.provider {
	case "consul":
		return sr.registerConsul()
	case "etcd":
		return sr.registerEtcd()
	default:
		return fmt.Errorf("unsupported service discovery provider: %s", sr.provider)
	}
}

// Deregister removes the service from the discovery provider
func (sr *ServiceRegistry) Deregister() error {
	switch sr.provider {
	case "consul":
		return sr.deregisterConsul()
	case "etcd":
		return sr.deregisterEtcd()
	default:
		return fmt.Errorf("unsupported service discovery provider: %s", sr.provider)
	}
}

// registerConsul registers with Consul
func (sr *ServiceRegistry) registerConsul() error {
	// In production, use Consul API client
	// For now, this is a placeholder
	sr.logger.Info("registering with Consul",
		zap.String("service", sr.serviceName),
		zap.Int("port", sr.servicePort),
		zap.String("consul_address", sr.consulAddress),
	)

	// TODO: Implement actual Consul registration
	// Example:
	// import "github.com/hashicorp/consul/api"
	// client, err := api.NewClient(&api.Config{Address: sr.consulAddress})
	// registration := &api.AgentServiceRegistration{...}
	// return client.Agent().ServiceRegister(registration)

	return nil
}

// deregisterConsul deregisters from Consul
func (sr *ServiceRegistry) deregisterConsul() error {
	sr.logger.Info("deregistering from Consul",
		zap.String("service", sr.serviceName),
	)

	// TODO: Implement actual Consul deregistration

	return nil
}

// registerEtcd registers with etcd
func (sr *ServiceRegistry) registerEtcd() error {
	sr.logger.Info("registering with etcd",
		zap.String("service", sr.serviceName),
		zap.Int("port", sr.servicePort),
	)

	// TODO: Implement etcd registration

	return nil
}

// deregisterEtcd deregisters from etcd
func (sr *ServiceRegistry) deregisterEtcd() error {
	sr.logger.Info("deregistering from etcd",
		zap.String("service", sr.serviceName),
	)

	// TODO: Implement etcd deregistration

	return nil
}

// FindAvailablePort finds an available port in the given range
func FindAvailablePort(preferredPort int, portRange []int) (int, error) {
	// Try preferred port first
	if IsPortAvailable(preferredPort) {
		return preferredPort, nil
	}

	// Try port range if provided
	if len(portRange) >= 2 {
		for port := portRange[0]; port <= portRange[1]; port++ {
			if IsPortAvailable(port) {
				return port, nil
			}
		}
		return 0, fmt.Errorf("no available port found in range %d-%d", portRange[0], portRange[1])
	}

	// No port range provided and preferred port not available
	return 0, fmt.Errorf("preferred port %d not available and no port range provided", preferredPort)
}

// IsPortAvailable checks if a port is available for binding
func IsPortAvailable(port int) bool {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	listener.Close()
	// Small delay to ensure port is fully released
	time.Sleep(100 * time.Millisecond)
	return true
}

package utils

import (
	"fmt"
	"time"

	"github.com/hashicorp/consul/api"
	"go.uber.org/zap"
)

// ServiceRegistry handles service registration and discovery
type ServiceRegistry struct {
	client      *api.Client
	serviceName string
	serviceID   string
	port        int
	logger      *zap.Logger
}

// NewServiceRegistry creates a new service registry
func NewServiceRegistry(provider, address, serviceName string, port int, logger *zap.Logger) (*ServiceRegistry, error) {
	if provider != "consul" {
		return nil, fmt.Errorf("unsupported service discovery provider: %s", provider)
	}

	config := api.DefaultConfig()
	config.Address = address

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %w", err)
	}

	serviceID := fmt.Sprintf("%s-%d", serviceName, port)

	return &ServiceRegistry{
		client:      client,
		serviceName: serviceName,
		serviceID:   serviceID,
		port:        port,
		logger:      logger,
	}, nil
}

// Register registers the service with Consul
func (sr *ServiceRegistry) Register() error {
	registration := &api.AgentServiceRegistration{
		ID:      sr.serviceID,
		Name:    sr.serviceName,
		Port:    sr.port,
		Address: getLocalIP(),
		Tags:    []string{"attachments", "v1", "helixtrack"},
		Check: &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://localhost:%d/health", sr.port),
			Interval:                       "10s",
			Timeout:                        "5s",
			DeregisterCriticalServiceAfter: "30s",
		},
	}

	err := sr.client.Agent().ServiceRegister(registration)
	if err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	sr.logger.Info("service registered",
		zap.String("service_id", sr.serviceID),
		zap.String("service_name", sr.serviceName),
		zap.Int("port", sr.port),
	)

	return nil
}

// Deregister deregisters the service from Consul
func (sr *ServiceRegistry) Deregister() error {
	err := sr.client.Agent().ServiceDeregister(sr.serviceID)
	if err != nil {
		return fmt.Errorf("failed to deregister service: %w", err)
	}

	sr.logger.Info("service deregistered",
		zap.String("service_id", sr.serviceID),
	)

	return nil
}

// DiscoverService discovers a service by name
func (sr *ServiceRegistry) DiscoverService(serviceName string) ([]*api.ServiceEntry, error) {
	services, _, err := sr.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to discover service: %w", err)
	}

	return services, nil
}

// GetServiceAddress gets the address of a service
func (sr *ServiceRegistry) GetServiceAddress(serviceName string) (string, error) {
	services, err := sr.DiscoverService(serviceName)
	if err != nil {
		return "", err
	}

	if len(services) == 0 {
		return "", fmt.Errorf("no healthy instances found for service: %s", serviceName)
	}

	// Return first healthy instance
	service := services[0]
	address := fmt.Sprintf("%s:%d", service.Service.Address, service.Service.Port)

	return address, nil
}

// WatchService watches for changes to a service
// Note: Watch functionality requires consul/watch package
// This is a placeholder implementation
func (sr *ServiceRegistry) WatchService(serviceName string, callback func([]*api.ServiceEntry)) error {
	// TODO: Implement service watching using consul/watch package
	// For now, we'll use periodic polling as a workaround
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			services, _, err := sr.client.Health().Service(serviceName, "", true, nil)
			if err != nil {
				sr.logger.Error("failed to check service health",
					zap.String("service", serviceName),
					zap.Error(err))
				continue
			}
			callback(services)
		}
	}()

	return nil
}

// Heartbeat sends periodic heartbeats to Consul
func (sr *ServiceRegistry) Heartbeat(interval time.Duration, stopChan <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := sr.client.Agent().UpdateTTL(sr.serviceID+":ttl", "", api.HealthPassing)
			if err != nil {
				sr.logger.Error("heartbeat failed", zap.Error(err))
			}
		case <-stopChan:
			sr.logger.Info("stopping heartbeat")
			return
		}
	}
}

// GetHealthyInstances returns all healthy instances of a service
func (sr *ServiceRegistry) GetHealthyInstances(serviceName string) ([]*ServiceInstance, error) {
	services, err := sr.DiscoverService(serviceName)
	if err != nil {
		return nil, err
	}

	instances := make([]*ServiceInstance, 0, len(services))
	for _, service := range services {
		instances = append(instances, &ServiceInstance{
			ID:      service.Service.ID,
			Name:    service.Service.Service,
			Address: service.Service.Address,
			Port:    service.Service.Port,
			Tags:    service.Service.Tags,
		})
	}

	return instances, nil
}

// ServiceInstance represents a service instance
type ServiceInstance struct {
	ID      string
	Name    string
	Address string
	Port    int
	Tags    []string
}

// GetAddress returns the full address of the instance
func (si *ServiceInstance) GetAddress() string {
	return fmt.Sprintf("%s:%d", si.Address, si.Port)
}

// HasTag checks if the instance has a specific tag
func (si *ServiceInstance) HasTag(tag string) bool {
	for _, t := range si.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// getLocalIP returns the local IP address
func getLocalIP() string {
	// In production, this would detect the actual local IP
	// For now, return localhost
	return "127.0.0.1"
}

// RegisterWithTTL registers the service with a TTL health check
func (sr *ServiceRegistry) RegisterWithTTL(ttl time.Duration) error {
	registration := &api.AgentServiceRegistration{
		ID:      sr.serviceID,
		Name:    sr.serviceName,
		Port:    sr.port,
		Address: getLocalIP(),
		Tags:    []string{"attachments", "v1", "helixtrack"},
		Check: &api.AgentServiceCheck{
			TTL:                            ttl.String(),
			DeregisterCriticalServiceAfter: "30s",
		},
	}

	err := sr.client.Agent().ServiceRegister(registration)
	if err != nil {
		return fmt.Errorf("failed to register service with TTL: %w", err)
	}

	sr.logger.Info("service registered with TTL",
		zap.String("service_id", sr.serviceID),
		zap.Duration("ttl", ttl),
	)

	// Start heartbeat
	stopChan := make(chan struct{})
	go sr.Heartbeat(ttl/2, stopChan)

	return nil
}

// SetMaintenance sets the service in maintenance mode
func (sr *ServiceRegistry) SetMaintenance(enable bool, reason string) error {
	err := sr.client.Agent().EnableServiceMaintenance(sr.serviceID, reason)
	if !enable {
		err = sr.client.Agent().DisableServiceMaintenance(sr.serviceID)
	}

	if err != nil {
		return fmt.Errorf("failed to set maintenance mode: %w", err)
	}

	sr.logger.Info("maintenance mode changed",
		zap.Bool("enabled", enable),
		zap.String("reason", reason),
	)

	return nil
}

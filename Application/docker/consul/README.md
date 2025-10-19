# Consul Service Discovery Configuration

This directory contains Consul configuration for service discovery, health checking, and configuration management.

## Overview

Consul provides:

1. **Service Discovery** - Automatic service registration and discovery
2. **Health Checking** - Monitor service health with multiple check types
3. **Key-Value Store** - Centralized configuration storage
4. **Service Mesh** - Secure service-to-service communication
5. **Multi-Datacenter** - Support for multiple datacenters
6. **UI Dashboard** - Web interface for monitoring

## Files

### Configuration Files

- `config/consul-config.json` - Main Consul server configuration
- `config/service-core.json` - Core service definition
- `config/service-auth.json` - Authentication service definition
- `config/service-perm.json` - Permissions service definition
- `config/service-documents.json` - Documents extension definition

## Architecture

### Service Registration Flow

```
1. Service starts
   ↓
2. Service calls Consul API to register
   ↓
3. Consul stores service metadata
   ↓
4. Consul starts health checking
   ↓
5. Service becomes discoverable
   ↓
6. Other services query Consul for service location
   ↓
7. Services communicate directly (or via load balancer)
```

### Automatic Port Selection with Consul

```bash
# Service startup (from entrypoint.sh)
1. Find available port (8080-8089)
2. Start service on selected port
3. Register with Consul using selected port
4. Other services discover via Consul
```

## Service Definitions

### Core Service

```json
{
  "service": {
    "name": "helixtrack-core",
    "port": 8080,
    "tags": ["api", "core", "rest"],
    "checks": [
      {
        "http": "http://core-service:8080/health",
        "interval": "30s"
      }
    ]
  }
}
```

**Metadata:**
- `version`: Service version
- `supports_rotation`: Supports rotation
- `auto_port_selection`: Supports automatic port selection

### Health Checks

Each service has two health checks:

1. **HTTP Health Check**
   - Endpoint: `/health`
   - Interval: 30 seconds
   - Timeout: 10 seconds
   - Deregister after: 90 seconds critical

2. **TCP Check**
   - Port: Service port
   - Interval: 30 seconds
   - Timeout: 5 seconds

## Usage

### Access Consul UI

```bash
# Open web browser
open http://localhost:8500/ui

# View all services
curl http://localhost:8500/v1/catalog/services

# View specific service
curl http://localhost:8500/v1/catalog/service/helixtrack-core
```

### Service Discovery

#### Query All Instances

```bash
# Get all core service instances
curl http://localhost:8500/v1/catalog/service/helixtrack-core | jq

# Response:
[
  {
    "ServiceID": "helixtrack-core-8080",
    "ServiceName": "helixtrack-core",
    "ServiceAddress": "172.20.0.5",
    "ServicePort": 8080,
    "ServiceMeta": {
      "version": "1.0.0",
      "supports_rotation": "true"
    }
  },
  {
    "ServiceID": "helixtrack-core-8081",
    "ServiceName": "helixtrack-core",
    "ServiceAddress": "172.20.0.6",
    "ServicePort": 8081
  }
]
```

#### Query Healthy Instances Only

```bash
# Get only healthy instances
curl http://localhost:8500/v1/health/service/helixtrack-core?passing | jq
```

#### DNS Query

```bash
# Resolve service via DNS
dig @localhost -p 8600 helixtrack-core.service.consul

# Get SRV record (includes port)
dig @localhost -p 8600 helixtrack-core.service.consul SRV
```

### Service Registration

#### Via HTTP API (Manual)

```bash
# Register service
curl -X PUT \
  -H "Content-Type: application/json" \
  -d '{
    "ID": "helixtrack-core-8080",
    "Name": "helixtrack-core",
    "Address": "172.20.0.5",
    "Port": 8080,
    "Check": {
      "HTTP": "http://172.20.0.5:8080/health",
      "Interval": "30s"
    }
  }' \
  http://localhost:8500/v1/agent/service/register
```

#### Via Configuration File (Automatic)

Place JSON files in `/consul/config/`:

```bash
# Consul automatically loads all .json files in config directory
docker run -v $(pwd)/config:/consul/config consul agent -dev
```

#### Via Application Code (Dynamic)

From the entrypoint.sh script:

```bash
# Register with dynamically selected port
register_service() {
    local PORT=$1
    curl -X PUT \
      -d @/tmp/service-registration.json \
      "${SERVICE_REGISTRY_URL}/v1/agent/service/register"
}
```

### Service Deregistration

```bash
# Deregister service
curl -X PUT http://localhost:8500/v1/agent/service/deregister/helixtrack-core-8080

# Services auto-deregister after 90s of being critical (if configured)
```

### Health Checks

#### Check Service Health

```bash
# Get health status for specific service
curl http://localhost:8500/v1/health/service/helixtrack-core | jq '.[].Checks'

# Response:
[
  {
    "CheckID": "serfHealth",
    "Status": "passing"
  },
  {
    "CheckID": "helixtrack-core-health",
    "Status": "passing",
    "Output": "HTTP GET http://core-service:8080/health: 200 OK"
  }
]
```

#### Check Individual Health Check

```bash
# Get specific check
curl http://localhost:8500/v1/agent/check/helixtrack-core-health | jq
```

#### Manually Update Health Check

```bash
# Mark check as passing
curl -X PUT http://localhost:8500/v1/agent/check/pass/helixtrack-core-health

# Mark check as warning
curl -X PUT http://localhost:8500/v1/agent/check/warn/helixtrack-core-health

# Mark check as critical
curl -X PUT http://localhost:8500/v1/agent/check/fail/helixtrack-core-health
```

## Key-Value Store

### Store Configuration

```bash
# Store key-value pair
curl -X PUT -d 'production' http://localhost:8500/v1/kv/helixtrack/environment

# Store JSON configuration
curl -X PUT \
  -d '{"log_level":"info","max_connections":100}' \
  http://localhost:8500/v1/kv/helixtrack/config/core

# Read value
curl http://localhost:8500/v1/kv/helixtrack/environment?raw

# Read with metadata
curl http://localhost:8500/v1/kv/helixtrack/environment | jq

# List all keys
curl http://localhost:8500/v1/kv/helixtrack?keys | jq

# Delete key
curl -X DELETE http://localhost:8500/v1/kv/helixtrack/environment
```

### Watch for Changes

```bash
# Watch key for changes
consul watch -type=key -key=helixtrack/config/core 'echo "Config changed!"'

# Watch service for changes
consul watch -type=service -service=helixtrack-core \
  'echo "Service instances changed!"'
```

## Service Mesh (Consul Connect)

### Enable Connect

Already enabled in consul-config.json:

```json
{
  "connect": {
    "enabled": true
  }
}
```

### Sidecar Proxy

```bash
# Start sidecar proxy for service
consul connect proxy \
  -sidecar-for helixtrack-core-8080 \
  -listen 127.0.0.1:8181
```

### Intentions (Service-to-Service Authorization)

```bash
# Allow auth service to connect to core
consul intention create helixtrack-auth helixtrack-core

# Deny documents service from connecting to auth
consul intention create -deny helixtrack-documents helixtrack-auth

# List all intentions
consul intention list

# Check if connection is allowed
consul intention check helixtrack-auth helixtrack-core
```

## Advanced Features

### Multi-Datacenter

```json
{
  "datacenter": "helixtrack-dc1",
  "wan_join": [
    "consul-dc2.helixtrack.com:8302"
  ]
}
```

Query services from other datacenters:

```bash
# Query service in dc2
curl http://localhost:8500/v1/catalog/service/helixtrack-core?dc=dc2
```

### Prepared Queries

Create dynamic queries for failover:

```bash
# Create prepared query
curl -X POST \
  -d '{
    "Name": "core-service",
    "Service": {
      "Service": "helixtrack-core",
      "Failover": {
        "NearestN": 3
      }
    }
  }' \
  http://localhost:8500/v1/query

# Execute query
curl http://localhost:8500/v1/query/core-service/execute
```

### ACLs (Access Control Lists)

Enable ACLs for security:

```json
{
  "acl": {
    "enabled": true,
    "default_policy": "deny",
    "enable_token_persistence": true
  }
}
```

```bash
# Bootstrap ACL system
consul acl bootstrap

# Create token
consul acl token create \
  -description "Core service token" \
  -service-identity helixtrack-core

# Use token
curl -H "X-Consul-Token: $TOKEN" http://localhost:8500/v1/catalog/services
```

### Snapshots

```bash
# Create snapshot
consul snapshot save backup.snap

# Restore snapshot
consul snapshot restore backup.snap

# Inspect snapshot
consul snapshot inspect backup.snap
```

## Monitoring

### Telemetry

Consul exposes Prometheus metrics:

```bash
# Get metrics
curl http://localhost:8500/v1/agent/metrics?format=prometheus

# Example metrics:
# consul_serf_member_status
# consul_catalog_service_count
# consul_health_service_query
```

### Logs

```bash
# View Consul logs
docker logs service-registry

# Follow logs
docker logs -f service-registry

# Filter by level
docker logs service-registry 2>&1 | grep ERROR
```

### Statistics

```bash
# Get leader
curl http://localhost:8500/v1/status/leader

# Get peers
curl http://localhost:8500/v1/status/peers

# Get members
curl http://localhost:8500/v1/agent/members | jq
```

## Troubleshooting

### Service Not Registered

**Symptom:** Service doesn't appear in Consul

**Solutions:**

```bash
# Check if registration succeeded
docker logs core-service | grep -i register

# Check Consul logs
docker logs service-registry | grep -i helixtrack-core

# Manually register for testing
curl -X PUT \
  -d '{"Name":"helixtrack-core","Port":8080}' \
  http://localhost:8500/v1/agent/service/register
```

### Health Check Failing

**Symptom:** Service shows as critical

**Solutions:**

```bash
# Check health check output
curl http://localhost:8500/v1/health/service/helixtrack-core | jq '.[].Checks'

# Test health endpoint directly
curl http://core-service:8080/health

# Check network connectivity
docker exec service-registry ping core-service
```

### Service Discovery Not Working

**Symptom:** Services can't find each other

**Solutions:**

```bash
# Verify service is registered
curl http://localhost:8500/v1/catalog/service/helixtrack-core

# Check DNS resolution
docker exec core-service \
  dig @service-registry helixtrack-auth.service.consul

# Check network
docker network inspect helixtrack-network
```

### Port Conflicts

**Symptom:** Consul port already in use

**Solutions:**

```bash
# Check what's using port 8500
lsof -i :8500
netstat -tuln | grep 8500

# Use different port
docker run -p 8600:8500 consul ...

# Or stop conflicting service
docker stop <container-using-8500>
```

## Performance Tuning

### Raft Multiplier

Adjust consensus performance:

```json
{
  "performance": {
    "raft_multiplier": 1  // 1-10, lower = faster but less reliable
  }
}
```

### Connection Limits

```json
{
  "limits": {
    "http_max_conns_per_client": 200,
    "https_handshake_timeout": "5s",
    "rpc_max_conns_per_client": 100
  }
}
```

### Session TTL

```bash
# Create session with custom TTL
curl -X PUT \
  -d '{
    "Name": "my-session",
    "TTL": "30s"
  }' \
  http://localhost:8500/v1/session/create
```

## Security Best Practices

### 1. Enable ACLs

```json
{
  "acl": {
    "enabled": true,
    "default_policy": "deny"
  }
}
```

### 2. Enable Encryption

```bash
# Generate encryption key
consul keygen
# Output: pUqJrVyVRj5jsiYEkM/tFQYfWyJIv4s3XkvDwy7Cu5s=

# Add to config
{
  "encrypt": "pUqJrVyVRj5jsiYEkM/tFQYfWyJIv4s3XkvDwy7Cu5s="
}
```

### 3. Enable TLS

```bash
# Generate certificates
consul tls ca create
consul tls cert create -server -dc=helixtrack-dc1

# Configure Consul
{
  "verify_incoming": true,
  "verify_outgoing": true,
  "verify_server_hostname": true,
  "ca_file": "consul-agent-ca.pem",
  "cert_file": "helixtrack-dc1-server-consul-0.pem",
  "key_file": "helixtrack-dc1-server-consul-0-key.pem"
}
```

### 4. Restrict UI Access

```json
{
  "ui_config": {
    "enabled": true,
    "content_security_policy": "default-src 'self'"
  },
  "addresses": {
    "http": "127.0.0.1"  // Only allow local access
  }
}
```

## Integration Examples

### Go Application

```go
import "github.com/hashicorp/consul/api"

// Create client
config := api.DefaultConfig()
config.Address = "localhost:8500"
client, err := api.NewClient(config)

// Register service
registration := &api.AgentServiceRegistration{
    ID:      "helixtrack-core-8080",
    Name:    "helixtrack-core",
    Port:    8080,
    Address: "172.20.0.5",
    Check: &api.AgentServiceCheck{
        HTTP:     "http://172.20.0.5:8080/health",
        Interval: "30s",
    },
}
err = client.Agent().ServiceRegister(registration)

// Discover service
services, _, err := client.Health().Service("helixtrack-auth", "", true, nil)
for _, service := range services {
    fmt.Printf("Found: %s:%d\n", service.Service.Address, service.Service.Port)
}
```

### Shell Script (from entrypoint.sh)

```bash
# Register service
cat > /tmp/service.json <<EOF
{
  "ID": "${HOSTNAME}-${PORT}",
  "Name": "helixtrack-core",
  "Port": ${PORT},
  "Address": "${HOSTNAME}",
  "Check": {
    "HTTP": "http://${HOSTNAME}:${PORT}/health",
    "Interval": "30s"
  }
}
EOF

curl -X PUT -d @/tmp/service.json \
  http://service-registry:8500/v1/agent/service/register

# Deregister on exit
trap 'curl -X PUT http://service-registry:8500/v1/agent/service/deregister/${HOSTNAME}-${PORT}' EXIT
```

## References

- [Consul Documentation](https://www.consul.io/docs)
- [Service Discovery Guide](https://learn.hashicorp.com/consul/getting-started/services)
- [Health Checks](https://www.consul.io/docs/agent/checks)
- [Consul Connect](https://www.consul.io/docs/connect)
- [ACL System](https://www.consul.io/docs/acl)

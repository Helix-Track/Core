# HAProxy Load Balancer Configuration

This directory contains HAProxy configuration for load balancing HelixTrack services.

## Overview

HAProxy provides:

1. **Load Balancing** - Distribute traffic across multiple service instances
2. **Health Checking** - Monitor service health and route only to healthy instances
3. **SSL Termination** - Handle HTTPS encryption at the edge
4. **Service Discovery Integration** - Dynamic backend configuration via Consul
5. **Statistics Dashboard** - Real-time monitoring and metrics
6. **Connection Pooling** - Efficient connection management

## Files

- `haproxy.cfg` - Main HAProxy configuration (static)
- `haproxy.ctmpl` - Consul-template for dynamic service discovery
- `errors/` - Custom error pages (JSON format)

## Features

### 1. Load Balancing Algorithms

HAProxy uses **round-robin** by default, but supports multiple algorithms:

```haproxy
backend core_services
    balance roundrobin        # Default: even distribution
    # balance leastconn      # Least connections
    # balance source         # Source IP hash (session persistence)
    # balance uri            # URI hash
```

### 2. Health Checks

All backend servers are monitored with HTTP health checks:

```haproxy
option httpchk GET /health HTTP/1.1\r\nHost:\ localhost
http-check expect status 200

server core-1 core-service:8080 check inter 10s fall 3 rise 2
```

**Health Check Parameters:**
- `check` - Enable health checking
- `inter 10s` - Check every 10 seconds
- `fall 3` - Mark down after 3 consecutive failures
- `rise 2` - Mark up after 2 consecutive successes

### 3. SSL/TLS Termination

HTTPS traffic is terminated at HAProxy:

```haproxy
frontend https_front
    bind *:443 ssl crt /usr/local/etc/haproxy/certs/helixtrack.pem
```

**SSL Features:**
- TLS 1.2+ only (no SSLv3, TLS 1.0, TLS 1.1)
- Strong cipher suites
- HSTS (HTTP Strict Transport Security)
- Security headers (X-Frame-Options, X-Content-Type-Options, etc.)

### 4. Service Discovery (Consul Integration)

#### Static Configuration (haproxy.cfg)

Default configuration with predefined servers:

```haproxy
backend core_services
    server core-service-1 core-service:8080 check
    server core-service-2 core-service:8081 check backup
```

#### Dynamic Configuration (haproxy.ctmpl)

Consul-template dynamically updates configuration:

```haproxy
backend core_services
    {{- range service "helixtrack-core" }}
    server {{.Node}}-{{.Port}} {{.Address}}:{{.Port}} check{{end}}
```

**To Enable Dynamic Discovery:**

```bash
# Install consul-template
wget https://releases.hashicorp.com/consul-template/0.34.0/consul-template_0.34.0_linux_amd64.zip
unzip consul-template_0.34.0_linux_amd64.zip

# Run consul-template
consul-template \
    -consul-addr=service-registry:8500 \
    -template="haproxy.ctmpl:/usr/local/etc/haproxy/haproxy.cfg:service haproxy reload" \
    -log-level=info
```

### 5. Statistics Dashboard

Access HAProxy statistics at: `http://localhost:8404/stats`

**Default Credentials:**
- Username: `admin`
- Password: `admin` (⚠️ **CHANGE IN PRODUCTION!**)

**Dashboard Features:**
- Real-time server health status
- Request rate and response times
- Active connections
- Error rates
- Server weights
- Manual server enable/disable

### 6. CORS Support

HAProxy automatically adds CORS headers:

```haproxy
http-response add-header Access-Control-Allow-Origin %[capture.req.hdr(0)]
http-response add-header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS"
http-response add-header Access-Control-Allow-Headers "Content-Type, Authorization"
http-response add-header Access-Control-Allow-Credentials true
```

## Configuration

### Port Mapping

| Port | Service | Description |
|------|---------|-------------|
| 80 | HTTP | Main HTTP endpoint |
| 443 | HTTPS | Main HTTPS endpoint (SSL termination) |
| 8404 | Stats | Statistics dashboard |
| 8405 | Health | HAProxy health check endpoint |
| 8406 | Metrics | Prometheus metrics |

### Backend Services

| Backend | Default Servers | Health Check |
|---------|----------------|--------------|
| `core_services` | core-service:8080-8082 | GET /health |
| `auth_services` | auth-service:8081-8082 | GET /health |
| `perm_services` | perm-service:8082-8083 | GET /health |
| `documents_services` | documents-service:8083-8084 | GET /health |

### Timeouts

```haproxy
timeout connect  5000ms   # Time to connect to backend
timeout client  50000ms   # Client inactivity timeout
timeout server  50000ms   # Server inactivity timeout
timeout http-request 10s  # Time to receive request
timeout queue 30s         # Time in queue waiting for connection
timeout tunnel 1h         # WebSocket/long-lived connections
```

## Usage

### Start HAProxy

```bash
# Via Docker Compose
docker-compose up -d load-balancer

# Standalone
docker run -d \
    --name haproxy \
    -p 80:80 \
    -p 443:443 \
    -p 8404:8404 \
    -v $(pwd)/haproxy.cfg:/usr/local/etc/haproxy/haproxy.cfg:ro \
    haproxy:2.8-alpine
```

### Test Health Check

```bash
# Check HAProxy health
curl http://localhost:8405/health

# Check backend service through HAProxy
curl http://localhost/health

# With SSL
curl https://localhost/health
```

### View Statistics

```bash
# Web dashboard
open http://localhost:8404/stats

# Or via curl
curl -u admin:admin http://localhost:8404/stats
```

### Reload Configuration

```bash
# Graceful reload (no downtime)
docker exec haproxy kill -USR2 1

# Or via Docker Compose
docker-compose exec load-balancer kill -USR2 1
```

## Advanced Configuration

### Session Persistence (Sticky Sessions)

Enable cookie-based session persistence:

```haproxy
backend core_services
    cookie SERVERID insert indirect nocache
    server core-1 core-service:8080 check cookie core-1
    server core-2 core-service:8081 check cookie core-2
```

### Rate Limiting

Protect against DDoS:

```haproxy
frontend http_front
    # Limit to 100 connections per IP
    stick-table type ip size 100k expire 30s store conn_cur
    http-request track-sc0 src
    http-request deny if { sc_conn_cur(0) gt 100 }
```

### Path-Based Routing

Route based on URL path:

```haproxy
frontend http_front
    # ACLs
    acl is_auth path_beg /auth
    acl is_perm path_beg /permissions
    acl is_docs path_beg /documents

    # Routing
    use_backend auth_services if is_auth
    use_backend perm_services if is_perm
    use_backend documents_services if is_docs
    default_backend core_services
```

### SSL Certificate Configuration

#### Development (Self-Signed)

```bash
# Generate self-signed certificate
openssl req -x509 -newkey rsa:4096 \
    -keyout key.pem -out cert.pem \
    -days 365 -nodes \
    -subj "/CN=helixtrack.local"

# Combine into single file for HAProxy
cat cert.pem key.pem > helixtrack.pem

# Mount in Docker
docker run -v $(pwd)/helixtrack.pem:/usr/local/etc/haproxy/certs/helixtrack.pem ...
```

#### Production (CA-Signed)

```bash
# 1. Generate CSR
openssl req -new -newkey rsa:4096 -nodes \
    -keyout helixtrack.key \
    -out helixtrack.csr \
    -subj "/CN=helixtrack.com"

# 2. Submit CSR to Certificate Authority and receive:
#    - helixtrack.crt (your certificate)
#    - intermediate.crt (CA intermediate certificate)
#    - root.crt (CA root certificate)

# 3. Create full certificate chain
cat helixtrack.crt intermediate.crt root.crt helixtrack.key > helixtrack.pem

# 4. Set permissions
chmod 600 helixtrack.pem

# 5. Deploy to HAProxy
docker cp helixtrack.pem haproxy:/usr/local/etc/haproxy/certs/
docker exec haproxy kill -USR2 1  # Reload
```

## Monitoring

### HAProxy Statistics

Access stats at `http://localhost:8404/stats` to monitor:

- **Request rate**: Requests per second
- **Response time**: Average, median, 95th percentile
- **Error rate**: 4xx, 5xx errors
- **Server health**: UP, DOWN, MAINT
- **Queue size**: Requests waiting for backend
- **Active connections**: Current connections

### Prometheus Metrics

HAProxy exposes Prometheus metrics at `http://localhost:8406/metrics`:

```bash
# Scrape metrics
curl http://localhost:8406/metrics

# Example metrics:
# haproxy_backend_current_sessions
# haproxy_backend_response_time_average_seconds
# haproxy_frontend_http_requests_total
# haproxy_server_check_failures_total
```

### Logging

HAProxy logs to stdout in Docker:

```bash
# View logs
docker logs haproxy

# Follow logs
docker logs -f haproxy

# Filter by backend
docker logs haproxy | grep "core_services"
```

## Troubleshooting

### 503 Service Unavailable

**Symptom:** HAProxy returns 503

**Causes:**
1. All backend servers are down
2. Health checks failing
3. Backend servers not responding

**Solutions:**

```bash
# Check backend health
curl http://localhost:8404/stats

# Check individual backend
docker exec haproxy haproxy -c -f /usr/local/etc/haproxy/haproxy.cfg

# Check backend service directly
curl http://core-service:8080/health
```

### 502 Bad Gateway

**Symptom:** HAProxy returns 502

**Causes:**
1. Backend returned invalid response
2. Connection reset by backend
3. Backend timeout

**Solutions:**

```bash
# Check backend logs
docker logs core-service

# Increase backend timeout
# In haproxy.cfg:
timeout server 60s  # Increase from 50s
```

### Configuration Not Reloading

**Symptom:** Changes to haproxy.cfg not taking effect

**Solutions:**

```bash
# Test configuration
docker exec haproxy haproxy -c -f /usr/local/etc/haproxy/haproxy.cfg

# Reload HAProxy (graceful)
docker exec haproxy kill -USR2 1

# Restart container
docker restart haproxy
```

### High Response Times

**Symptom:** Slow responses through HAProxy

**Causes:**
1. Backend servers overloaded
2. Too few backend servers
3. Connection pooling exhausted

**Solutions:**

```bash
# Scale up backend services
docker-compose up -d --scale core-service=5

# Increase connection limits
# In haproxy.cfg:
maxconn 8192

# Enable connection reuse
option http-keep-alive
```

## Performance Tuning

### Connection Limits

```haproxy
global
    maxconn 4096              # Total connections
    tune.ssl.default-dh-param 2048

defaults
    maxconn 2000              # Per-frontend limit
```

### Buffer Sizes

```haproxy
global
    tune.bufsize 32768        # 32KB (default 16KB)
    tune.maxrewrite 8192      # 8KB for header rewriting
```

### Thread Count

```haproxy
global
    nbthread 4                # Use 4 CPU cores
```

### Connection Pooling

```haproxy
backend core_services
    option http-keep-alive    # Enable keep-alive
    option prefer-last-server # Prefer same server for keep-alive
```

## Security

### Best Practices

1. **Change default stats password:**
```haproxy
stats auth admin:STRONG_PASSWORD_HERE
```

2. **Restrict stats access to trusted IPs:**
```haproxy
frontend stats
    bind *:8404
    acl trusted_ip src 10.0.0.0/8 172.16.0.0/12
    http-request deny unless trusted_ip
```

3. **Enable HTTPS only:**
```haproxy
frontend http_front
    bind *:80
    redirect scheme https code 301 if !{ ssl_fc }
```

4. **Use strong SSL ciphers:**
```haproxy
ssl-default-bind-ciphers ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384
ssl-default-bind-options no-sslv3 no-tlsv10 no-tlsv11
```

5. **Enable security headers:**
```haproxy
http-response set-header X-Frame-Options "SAMEORIGIN"
http-response set-header X-Content-Type-Options "nosniff"
http-response set-header X-XSS-Protection "1; mode=block"
```

## References

- [HAProxy Documentation](https://www.haproxy.org/documentation.html)
- [HAProxy Configuration Manual](https://www.haproxy.com/documentation/haproxy-configuration-manual/latest/)
- [Consul Template](https://github.com/hashicorp/consul-template)
- [HAProxy Best Practices](https://www.haproxy.com/blog/haproxy-best-practices/)

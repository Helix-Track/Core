# HelixTrack Core - Deployment Guide

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Build and Installation](#build-and-installation)
3. [Database Setup](#database-setup)
4. [Service Configuration](#service-configuration)
5. [Deployment Options](#deployment-options)
6. [Production Checklist](#production-checklist)
7. [Monitoring and Maintenance](#monitoring-and-maintenance)

## Prerequisites

### System Requirements

- **Operating System**: Linux (recommended), macOS, or Windows
- **Go**: Version 1.22 or higher (for building from source)
- **Database**: SQLite 3 or PostgreSQL 12+
- **Memory**: Minimum 512MB RAM (2GB+ recommended for production)
- **Disk Space**: 100MB for application + database size
- **Network**: Open port for HTTP/HTTPS (default: 8080)

### Optional Dependencies

- Docker and Docker Compose (for containerized deployment)
- Nginx or Apache (for reverse proxy)
- systemd (for service management on Linux)

## Build and Installation

### Option 1: Build from Source

```bash
# Clone repository
git clone <repository-url>
cd Core/Application

# Download dependencies
go mod download

# Build binary
go build -o htCore main.go

# Install to system (optional)
sudo cp htCore /usr/local/bin/
sudo chmod +x /usr/local/bin/htCore
```

### Option 2: Cross-Platform Build

```bash
# Build for Linux (64-bit)
GOOS=linux GOARCH=amd64 go build -o htCore-linux-amd64 main.go

# Build for macOS (64-bit)
GOOS=darwin GOARCH=amd64 go build -o htCore-darwin-amd64 main.go

# Build for Windows (64-bit)
GOOS=windows GOARCH=amd64 go build -o htCore-windows-amd64.exe main.go

# Build for ARM (Raspberry Pi, etc.)
GOOS=linux GOARCH=arm64 go build -o htCore-linux-arm64 main.go
```

### Option 3: Optimized Production Build

```bash
# Build with optimizations
go build -ldflags="-s -w" -o htCore main.go

# Further compress with upx (optional)
upx --best --lzma htCore
```

## Database Setup

### SQLite Setup (Development/Small Deployments)

```bash
# Create database directory
mkdir -p Database

# Copy database file (if provided)
cp path/to/Definition.sqlite Database/

# Or let the application create it on first run
# Ensure directory has write permissions
chmod 755 Database
```

### PostgreSQL Setup (Production)

```bash
# Install PostgreSQL
sudo apt-get install postgresql postgresql-contrib  # Ubuntu/Debian
sudo yum install postgresql-server postgresql-contrib  # CentOS/RHEL

# Start PostgreSQL
sudo systemctl start postgresql
sudo systemctl enable postgresql

# Create database and user
sudo -u postgres psql << EOF
CREATE USER htcore WITH PASSWORD 'your-secure-password';
CREATE DATABASE htcore OWNER htcore;
GRANT ALL PRIVILEGES ON DATABASE htcore TO htcore;
\q
EOF

# Import schema
psql -U htcore -d htcore -f Database/DDL/Definition.V1.sql

# Configure PostgreSQL for network access (if needed)
sudo nano /etc/postgresql/*/main/postgresql.conf
# Set: listen_addresses = '*'

sudo nano /etc/postgresql/*/main/pg_hba.conf
# Add: host  htcore  htcore  0.0.0.0/0  md5

# Restart PostgreSQL
sudo systemctl restart postgresql
```

## Service Configuration

### Create Production Configuration

```bash
# Create configuration directory
sudo mkdir -p /etc/htcore
sudo cp Configurations/default.json /etc/htcore/production.json

# Edit configuration
sudo nano /etc/htcore/production.json
```

**Production Configuration Example:**

```json
{
  "log": {
    "log_path": "/var/log/htcore",
    "logfile_base_name": "htCore",
    "log_size_limit": 100000000,
    "level": "warn"
  },
  "listeners": [
    {
      "address": "127.0.0.1",
      "port": 8080,
      "https": false
    }
  ],
  "database": {
    "type": "postgres",
    "postgres_host": "localhost",
    "postgres_port": 5432,
    "postgres_user": "htcore",
    "postgres_password": "your-secure-password",
    "postgres_database": "htcore",
    "postgres_ssl_mode": "require"
  },
  "services": {
    "authentication": {
      "enabled": true,
      "url": "http://auth-service:8081",
      "timeout": 30
    },
    "permissions": {
      "enabled": true,
      "url": "http://perm-service:8082",
      "timeout": 30
    }
  }
}
```

### Create Log Directory

```bash
# Create log directory
sudo mkdir -p /var/log/htcore

# Set permissions
sudo chown htcore:htcore /var/log/htcore
sudo chmod 755 /var/log/htcore
```

## Deployment Options

### Option 1: Systemd Service (Recommended for Linux)

**Create service file:**

```bash
sudo nano /etc/systemd/system/htcore.service
```

**Service Configuration:**

```ini
[Unit]
Description=HelixTrack Core Service
After=network.target postgresql.service
Wants=postgresql.service

[Service]
Type=simple
User=htcore
Group=htcore
WorkingDirectory=/opt/htcore
ExecStart=/usr/local/bin/htCore -config=/etc/htcore/production.json
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/log/htcore /opt/htcore/Database

[Install]
WantedBy=multi-user.target
```

**Enable and start service:**

```bash
# Reload systemd
sudo systemctl daemon-reload

# Enable service on boot
sudo systemctl enable htcore

# Start service
sudo systemctl start htcore

# Check status
sudo systemctl status htcore

# View logs
sudo journalctl -u htcore -f
```

### Option 2: Docker Deployment

**Create Dockerfile:**

```dockerfile
FROM golang:1.22-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o htCore main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates sqlite-libs

WORKDIR /app

COPY --from=builder /build/htCore .
COPY --from=builder /build/Configurations ./Configurations
COPY --from=builder /build/Database ./Database

RUN addgroup -g 1000 htcore && \
    adduser -D -u 1000 -G htcore htcore && \
    chown -R htcore:htcore /app

USER htcore

EXPOSE 8080

CMD ["./htCore", "-config=Configurations/default.json"]
```

**Build and run:**

```bash
# Build image
docker build -t helixtrack-core:1.0.0 .

# Run container
docker run -d \
  --name htcore \
  -p 8080:8080 \
  -v /path/to/config.json:/app/Configurations/production.json \
  -v /path/to/database:/app/Database \
  -v /path/to/logs:/var/log/htcore \
  --restart unless-stopped \
  helixtrack-core:1.0.0
```

### Option 3: Docker Compose

**docker-compose.yml:**

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: htcore
      POSTGRES_USER: htcore
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres-data:/var/lib/postgresql/data
      - ./Database/DDL:/docker-entrypoint-initdb.d
    networks:
      - htcore-network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U htcore"]
      interval: 10s
      timeout: 5s
      retries: 5

  htcore:
    build: .
    ports:
      - "8080:8080"
    depends_on:
      postgres:
        condition: service_healthy
    volumes:
      - ./Configurations/production.json:/app/Configurations/production.json:ro
      - htcore-logs:/var/log/htcore
    environment:
      - CONFIG_PATH=/app/Configurations/production.json
    networks:
      - htcore-network
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./ssl:/etc/nginx/ssl:ro
    depends_on:
      - htcore
    networks:
      - htcore-network
    restart: unless-stopped

volumes:
  postgres-data:
  htcore-logs:

networks:
  htcore-network:
    driver: bridge
```

**Deploy:**

```bash
# Set environment variables
export DB_PASSWORD=your-secure-password

# Start services
docker-compose up -d

# View logs
docker-compose logs -f htcore

# Stop services
docker-compose down
```

### Option 4: Kubernetes Deployment

**deployment.yaml:**

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: htcore
  labels:
    app: htcore
spec:
  replicas: 3
  selector:
    matchLabels:
      app: htcore
  template:
    metadata:
      labels:
        app: htcore
    spec:
      containers:
      - name: htcore
        image: helixtrack-core:1.0.0
        ports:
        - containerPort: 8080
        env:
        - name: CONFIG_PATH
          value: /config/production.json
        volumeMounts:
        - name: config
          mountPath: /config
          readOnly: true
        - name: logs
          mountPath: /var/log/htcore
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            memory: "256Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
      volumes:
      - name: config
        configMap:
          name: htcore-config
      - name: logs
        emptyDir: {}
---
apiVersion: v1
kind: Service
metadata:
  name: htcore
spec:
  selector:
    app: htcore
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: LoadBalancer
```

**Deploy to Kubernetes:**

```bash
# Create ConfigMap
kubectl create configmap htcore-config \
  --from-file=production.json=Configurations/production.json

# Deploy
kubectl apply -f deployment.yaml

# Check status
kubectl get pods -l app=htcore
kubectl logs -l app=htcore -f

# Get service URL
kubectl get service htcore
```

### Option 5: Nginx Reverse Proxy

**nginx.conf:**

```nginx
upstream htcore {
    server 127.0.0.1:8080;
    keepalive 32;
}

server {
    listen 80;
    listen [::]:80;
    server_name your-domain.com;

    # Redirect to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name your-domain.com;

    # SSL configuration
    ssl_certificate /etc/nginx/ssl/cert.pem;
    ssl_certificate_key /etc/nginx/ssl/key.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    # Logging
    access_log /var/log/nginx/htcore-access.log;
    error_log /var/log/nginx/htcore-error.log;

    # Proxy settings
    location / {
        proxy_pass http://htcore;
        proxy_http_version 1.1;

        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;

        # WebSocket support (if needed)
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }

    # Health check endpoint
    location /health {
        proxy_pass http://htcore/health;
        access_log off;
    }
}
```

## Production Checklist

### Security

- [ ] Use HTTPS in production
- [ ] Set strong database passwords
- [ ] Enable authentication service
- [ ] Enable permission service
- [ ] Configure firewall rules
- [ ] Use non-root user for service
- [ ] Enable PostgreSQL SSL mode
- [ ] Restrict database network access
- [ ] Set appropriate file permissions
- [ ] Disable debug logging

### Performance

- [ ] Use PostgreSQL for production
- [ ] Configure database connection pooling
- [ ] Enable database query caching
- [ ] Set up reverse proxy (Nginx/Apache)
- [ ] Configure log rotation
- [ ] Optimize database indexes
- [ ] Monitor memory usage
- [ ] Set resource limits

### Reliability

- [ ] Configure service auto-restart
- [ ] Set up health check monitoring
- [ ] Configure log rotation
- [ ] Set up automated backups
- [ ] Test disaster recovery
- [ ] Document rollback procedures
- [ ] Set up monitoring/alerting
- [ ] Configure rate limiting

### Monitoring

- [ ] Set up application monitoring
- [ ] Monitor system resources
- [ ] Monitor database performance
- [ ] Set up log aggregation
- [ ] Configure alerts
- [ ] Monitor API response times
- [ ] Track error rates
- [ ] Set up uptime monitoring

## Monitoring and Maintenance

### Health Checks

```bash
# Check application health
curl http://localhost:8080/health

# Check via /do endpoint
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d '{"action": "health"}'
```

### Log Management

```bash
# View live logs (systemd)
sudo journalctl -u htcore -f

# View application logs
tail -f /var/log/htcore/htCore.log

# Search for errors
grep ERROR /var/log/htcore/htCore.log

# Log rotation (create /etc/logrotate.d/htcore)
/var/log/htcore/*.log {
    daily
    rotate 14
    compress
    delaycompress
    notifempty
    create 0640 htcore htcore
    sharedscripts
    postrotate
        systemctl reload htcore > /dev/null 2>&1 || true
    endscript
}
```

### Database Backup

```bash
# SQLite backup
sqlite3 Database/Definition.sqlite ".backup backup-$(date +%Y%m%d).sqlite"

# PostgreSQL backup
pg_dump -U htcore htcore > backup-$(date +%Y%m%d).sql

# Automated backup script
#!/bin/bash
BACKUP_DIR="/backups/htcore"
DATE=$(date +%Y%m%d-%H%M%S)

mkdir -p $BACKUP_DIR
pg_dump -U htcore htcore | gzip > $BACKUP_DIR/htcore-$DATE.sql.gz

# Keep last 7 days
find $BACKUP_DIR -name "*.sql.gz" -mtime +7 -delete
```

### Performance Monitoring

```bash
# Monitor system resources
htop

# Monitor database connections
psql -U htcore -c "SELECT count(*) FROM pg_stat_activity;"

# Monitor application metrics (if configured)
curl http://localhost:8080/metrics
```

### Updating the Application

```bash
# Stop service
sudo systemctl stop htcore

# Backup current binary
sudo cp /usr/local/bin/htCore /usr/local/bin/htCore.backup

# Deploy new binary
sudo cp htCore /usr/local/bin/

# Run database migrations (if needed)
# psql -U htcore -d htcore -f Database/DDL/Migration.VX.Y.sql

# Start service
sudo systemctl start htcore

# Verify
sudo systemctl status htcore
curl http://localhost:8080/health
```

---

**Version:** 1.0.0
**Last Updated:** 2025-10-10

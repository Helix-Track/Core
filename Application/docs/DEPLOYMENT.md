# HelixTrack Core - Deployment Guide

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Build and Installation](#build-and-installation)
3. [Database Setup](#database-setup)
4. [Service Configuration](#service-configuration)
5. [Deployment Options](#deployment-options)
6. [Production Checklist](#production-checklist)
7. [Monitoring and Maintenance](#monitoring-and-maintenance)

## Visual Documentation

Before deploying, review the architecture diagrams to understand the system structure:

**Quick Access:** [Documentation Portal](index.html) | [All Diagrams](diagrams/README.md) | [Architecture Docs](ARCHITECTURE.md)

### Key Diagrams for Deployment

1. **[System Architecture](diagrams/01-system-architecture.drawio)** - Complete overview of all system layers, components, and how they interact. Essential for understanding deployment topology.

2. **[Microservices Interaction](diagrams/05-microservices-interaction.drawio)** - Shows Core service, Authentication service, Permissions engine, and optional extensions. Includes 3 deployment scenarios:
   - Development (single machine)
   - Production (distributed services)
   - High Availability (multiple replicas)

3. **[Database Schema Overview](diagrams/02-database-schema-overview.drawio)** - All 89 tables for database setup and migration planning.

These diagrams include:
- Service topology and communication patterns
- Port configurations (Core: 8080, Auth: 8081, Permissions: 8082)
- HTTP/JSON communication details
- Deployment architecture examples
- Docker Compose and Kubernetes configurations

**Additional Resources:**
- [Architecture Documentation](ARCHITECTURE.md) - Section 8: Deployment Architecture
- [User Manual](USER_MANUAL.md) - API reference and service configuration

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

## Extension Deployment

HelixTrack Core supports optional extensions that add functionality without modifying the core codebase. This section covers deploying extensions, with a focus on the Documents V2 extension.

### Available Extensions

- **Documents** (✅ Production Ready) - Confluence-style document management
- **Times** - Time tracking and work log management
- **Chats** - Integration with messaging platforms (Slack, Telegram, etc.)
- **Lokalisation** - Localization and internationalization support

### Extension Architecture

Extensions follow a modular architecture:

1. **Self-Contained**: Each extension has its own database schema
2. **HTTP-Based**: Extensions communicate with Core via REST API
3. **Optional**: Can be enabled/disabled without affecting core functionality
4. **Independent**: Extensions can run as separate services or within Core

---

## Documents V2 Extension Deployment

The Documents extension adds Confluence-style document management capabilities to HelixTrack with 102% feature parity.

### Features Overview

- **90 API Actions**: Complete document lifecycle management
- **32 Database Tables**: Comprehensive document data model
- **Confluence Parity**: 46 features covering all major Confluence capabilities
- **Real-time Collaboration**: Comments, mentions, watchers, reactions
- **Version Control**: Full version history with diffs and rollback
- **Rich Content**: HTML, Markdown, Plain Text, Storage Format
- **Templates & Blueprints**: Reusable document templates with wizards
- **Analytics**: View tracking, popularity scoring, engagement metrics
- **Multi-format Export**: PDF, Markdown, HTML, DOCX
- **Hierarchical Organization**: Spaces, types, parent-child relationships

### Database Schema Deployment

#### SQLite Deployment

```bash
# Navigate to database directory
cd Database/DDL/Extensions/Documents

# Import Documents extension schema
sqlite3 /path/to/database.db < Documents.V1.sql

# Verify tables created
sqlite3 /path/to/database.db "SELECT name FROM sqlite_master WHERE type='table' AND name LIKE 'document%';"
```

Expected output: 32 tables starting with `document`

#### PostgreSQL Deployment

```bash
# Import Documents extension schema
psql -U htcore -d htcore -f Database/DDL/Extensions/Documents/Documents.V1.sql

# Verify tables created
psql -U htcore -d htcore -c "\dt document*"
```

### Documents Database Tables (32 tables)

**Core Tables:**
- `document` - Main document metadata
- `document_content` - Document content with versioning
- `document_space` - Confluence-style spaces
- `document_type` - Document type definitions

**Versioning Tables:**
- `document_version` - Version history
- `document_version_label` - Version labels
- `document_version_tag` - Version tags
- `document_version_comment` - Version comments
- `document_version_mention` - @mentions in versions
- `document_version_diff` - Cached version diffs

**Collaboration Tables:**
- `comment_document_mapping` - Document comments
- `document_inline_comment` - Inline/contextual comments
- `document_watcher` - Watch subscriptions
- `document_mention` - @mentions in documents
- `document_reaction` - Emoji reactions

**Organization Tables:**
- `label_document_mapping` - Document labels
- `document_tag` - Tag definitions
- `document_tag_mapping` - Document tags
- `vote_mapping` - Votes/reactions (generic system)

**Relationship Tables:**
- `document_entity_link` - Links to tickets/projects/etc.
- `document_relationship` - Document-to-document relationships

**Template Tables:**
- `document_template` - Reusable templates
- `document_blueprint` - Template wizards

**Analytics Tables:**
- `document_view_history` - View tracking
- `document_analytics` - Aggregated metrics

**Attachment Tables:**
- `document_attachment` - File attachments

### Configuration

Documents extension requires no additional configuration - it uses the main Core configuration for database and services.

**Optional Configuration** (for advanced deployments):

```json
{
  "extensions": {
    "documents": {
      "enabled": true,
      "max_document_size": 10485760,
      "max_attachment_size": 52428800,
      "allowed_attachment_types": ["pdf", "doc", "docx", "xls", "xlsx", "png", "jpg", "jpeg"],
      "enable_realtime_collaboration": true,
      "enable_export": true
    }
  }
}
```

### API Actions (90 total)

The Documents extension adds 90 new API actions to the Core `/do` endpoint:

**Core Document Operations** (20 actions):
- `documentCreate`, `documentRead`, `documentList`, `documentUpdate`, `documentDelete`
- `documentRestore`, `documentArchive`, `documentUnarchive`, `documentPublish`, `documentUnpublish`
- `documentDuplicate`, `documentMove`, `documentSetParent`, `documentGetChildren`, `documentGetHierarchy`
- `documentGetBreadcrumb`, `documentSearch`, `documentGetRelated`, `documentGetTree`, `documentReorder`

**Document Content** (4 actions):
- `documentContentCreate`, `documentContentGet`, `documentContentUpdate`, `documentContentGetLatest`

**Document Spaces** (5 actions):
- `documentSpaceCreate`, `documentSpaceRead`, `documentSpaceList`, `documentSpaceUpdate`, `documentSpaceDelete`

**Document Versioning** (15 actions):
- Version management, labels, tags, comments, mentions, diffs, restore

**Document Collaboration** (12 actions):
- Comments, inline comments, watchers, mentions

**Document Organization** (10 actions):
- Labels, tags, reactions, voting

**Document Export** (8 actions):
- `documentExportPDF`, `documentExportMarkdown`, `documentExportHTML`, `documentExportDOCX`
- `documentExportSpace`, `documentExportBatch`, `documentExportSchedule`, `documentExportGetStatus`

**Document Entity Links** (4 actions):
- `documentLinkCreate`, `documentLinkList`, `documentLinkDelete`, `documentRelationshipCreate`

**Document Templates** (5 actions):
- `documentTemplateCreate`, `documentTemplateGet`, `documentTemplateList`, `documentTemplateUse`, `documentBlueprintCreate`

**Document Analytics** (3 actions):
- `documentViewTrack`, `documentAnalyticsGet`, `documentGetPopular`

**Document Attachments** (4 actions):
- `documentAttachmentUpload`, `documentAttachmentGet`, `documentAttachmentList`, `documentAttachmentDelete`

### Testing Documents Deployment

#### Verify Schema

```bash
# Check table count
sqlite3 /path/to/database.db "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name LIKE 'document%';"
# Expected: 32

# List all document tables
sqlite3 /path/to/database.db "SELECT name FROM sqlite_master WHERE type='table' AND name LIKE 'document%' ORDER BY name;"
```

#### Test API Actions

```bash
# 1. Create a document space
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d '{
    "action": "documentSpaceCreate",
    "jwt": "your-jwt-token",
    "data": {
      "key": "DOCS",
      "name": "Documentation",
      "description": "Main documentation space",
      "is_public": true
    }
  }'

# 2. Create a document
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d '{
    "action": "documentCreate",
    "jwt": "your-jwt-token",
    "data": {
      "title": "Getting Started",
      "space_id": "space-id-from-step-1",
      "type_id": "page",
      "content": "<h1>Welcome</h1><p>Getting started with HelixTrack</p>",
      "content_type": "html"
    }
  }'

# 3. List documents
curl -X POST http://localhost:8080/do \
  -H "Content-Type": application/json" \
  -d '{
    "action": "documentList",
    "jwt": "your-jwt-token",
    "data": {
      "space_id": "space-id-from-step-1",
      "limit": 10
    }
  }'
```

#### Verify Analytics

```bash
# Track document view
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d '{
    "action": "documentViewTrack",
    "jwt": "your-jwt-token",
    "data": {
      "document_id": "doc-id",
      "duration": 30
    }
  }'

# Get analytics
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d '{
    "action": "documentAnalyticsGet",
    "jwt": "your-jwt-token",
    "data": {
      "document_id": "doc-id"
    }
  }'
```

### Performance Considerations

**Database Indexes** (recommended for production):

```sql
-- Document lookup indexes
CREATE INDEX IF NOT EXISTS idx_document_space ON document(space_id);
CREATE INDEX IF NOT EXISTS idx_document_parent ON document(parent_id);
CREATE INDEX IF NOT EXISTS idx_document_type ON document(type_id);
CREATE INDEX IF NOT EXISTS idx_document_created ON document(created);

-- Version history indexes
CREATE INDEX IF NOT EXISTS idx_document_version_doc ON document_version(document_id);
CREATE INDEX IF NOT EXISTS idx_document_content_doc ON document_content(document_id);

-- Collaboration indexes
CREATE INDEX IF NOT EXISTS idx_watcher_document ON document_watcher(document_id);
CREATE INDEX IF NOT EXISTS idx_comment_document ON comment_document_mapping(document_id);

-- Analytics indexes
CREATE INDEX IF NOT EXISTS idx_view_document ON document_view_history(document_id);
CREATE INDEX IF NOT EXISTS idx_analytics_document ON document_analytics(document_id);
```

**Recommended Settings:**

- **Max Document Size**: 10 MB (configurable)
- **Max Attachment Size**: 50 MB (configurable)
- **Version Retention**: Unlimited (soft delete)
- **Analytics Aggregation**: Every 5 minutes
- **Search Indexing**: Real-time

### Troubleshooting

#### Issue: Documents API actions return 404

**Solution**: Verify handlers are registered in Core

```bash
# Check available actions
curl http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d '{"action": "version"}'

# Look for "documentCreate" in response
```

#### Issue: Database tables not found

**Solution**: Import Documents schema

```bash
# Check if tables exist
sqlite3 database.db "SELECT name FROM sqlite_master WHERE type='table' AND name='document';"

# If empty, import schema
sqlite3 database.db < Database/DDL/Extensions/Documents/Documents.V1.sql
```

#### Issue: Slow document searches

**Solution**: Add search indexes

```sql
-- Full-text search index (SQLite)
CREATE VIRTUAL TABLE IF NOT EXISTS document_fts USING fts5(
    document_id, title, content, tokenize='porter'
);

-- Keep FTS in sync with triggers
CREATE TRIGGER IF NOT EXISTS document_fts_insert AFTER INSERT ON document BEGIN
    INSERT INTO document_fts(document_id, title) VALUES (new.id, new.title);
END;
```

#### Issue: WebSocket events not working for documents

**Solution**: Verify WebSocket endpoint is accessible

```bash
# Test WebSocket connection
wscat -c ws://localhost:8080/ws

# Should receive connection confirmation
```

### Migration from Other Systems

#### Confluence to HelixTrack Documents

```bash
# Export from Confluence (use Confluence REST API)
curl -u admin:password \
  "https://your-confluence.com/rest/api/content?limit=100" \
  > confluence-export.json

# Convert and import (example script)
python scripts/migrate-confluence.py \
  --input confluence-export.json \
  --helix-url http://localhost:8080 \
  --jwt "your-token"
```

#### Google Docs to HelixTrack Documents

1. Export from Google Docs (File → Download → HTML)
2. Use `documentCreate` API with HTML content
3. Convert links and images to HelixTrack references

### Backup and Recovery

```bash
# Backup Documents tables (SQLite)
sqlite3 database.db ".dump document" > documents-backup.sql
sqlite3 database.db ".dump document_content" >> documents-backup.sql
# ... repeat for all 32 tables

# Restore
sqlite3 database.db < documents-backup.sql

# Backup Documents tables (PostgreSQL)
pg_dump -U htcore -d htcore \
  -t 'document*' \
  > documents-backup.sql

# Restore
psql -U htcore -d htcore < documents-backup.sql
```

### Monitoring

**Key Metrics to Monitor:**

- Document creation rate
- Document view count
- Search response time
- Attachment storage usage
- Version history size
- Collaboration activity (comments, watchers)

**Sample Monitoring Query:**

```sql
-- Document statistics
SELECT
    COUNT(*) as total_documents,
    COUNT(CASE WHEN is_published = 1 THEN 1 END) as published,
    COUNT(CASE WHEN is_archived = 1 THEN 1 END) as archived,
    COUNT(CASE WHEN deleted = 1 THEN 1 END) as deleted
FROM document;

-- Most popular documents
SELECT d.title, da.total_views, da.unique_viewers, da.popularity_score
FROM document d
JOIN document_analytics da ON d.id = da.document_id
ORDER BY da.popularity_score DESC
LIMIT 10;
```

---

**Version:** 3.1.0 (Documents V2 Edition)
**Last Updated:** 2025-10-18

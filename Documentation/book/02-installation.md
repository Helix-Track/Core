# Chapter 2: Installation

[← Previous: Introduction](01-introduction.md) | [Back to Table of Contents](README.md) | [Next: Configuration →](03-configuration.md)

---

## System Requirements

### Minimum Requirements

- **CPU**: 1 core (2+ cores recommended)
- **RAM**: 512 MB (2 GB+ recommended for production)
- **Disk**: 100 MB for application + database storage
- **OS**: Linux, macOS, or Windows
- **Go**: 1.22 or higher (for building from source)
- **Database**: SQLite 3 or PostgreSQL 12+

### Recommended Production Setup

- **CPU**: 4+ cores
- **RAM**: 4+ GB
- **Disk**: SSD with 10+ GB
- **OS**: Linux (Ubuntu 22.04 LTS, Debian 11, CentOS 8)
- **Database**: PostgreSQL 14+ with replication
- **Load Balancer**: Nginx or HAProxy
- **Monitoring**: Prometheus + Grafana

---

## Installation Methods

### Method 1: Pre-built Binary (Recommended)

**Step 1**: Download the binary for your platform

```bash
# Linux AMD64
wget https://github.com/Helix-Track/Core/releases/download/v2.0.0/htCore-linux-amd64

# macOS ARM64
wget https://github.com/Helix-Track/Core/releases/download/v2.0.0/htCore-darwin-arm64

# Windows AMD64
wget https://github.com/Helix-Track/Core/releases/download/v2.0.0/htCore-windows-amd64.exe
```

**Step 2**: Make it executable (Linux/macOS)

```bash
chmod +x htCore-linux-amd64
mv htCore-linux-amd64 /usr/local/bin/htCore
```

**Step 3**: Verify installation

```bash
htCore --version
# Output: HelixTrack Core v2.0.0
```

### Method 2: Build from Source

**Step 1**: Install Go 1.22+

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install golang-1.22

# macOS (Homebrew)
brew install go@1.22

# Verify
go version
```

**Step 2**: Clone the repository

```bash
git clone https://github.com/Helix-Track/Core.git
cd Core/Application
```

**Step 3**: Install dependencies

```bash
go mod download
go mod verify
```

**Step 4**: Build the application

```bash
# Standard build
go build -o htCore main.go

# Optimized production build
go build -ldflags="-s -w" -o htCore main.go

# With version information
VERSION=2.0.0
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
go build -ldflags="-X main.Version=$VERSION -X main.BuildDate=$BUILD_DATE" -o htCore main.go
```

**Step 5**: Install systemwide

```bash
sudo mv htCore /usr/local/bin/
sudo chmod +x /usr/local/bin/htCore
```

### Method 3: Docker Container

**Step 1**: Build Docker image

```bash
cd Core/Application
docker build -t helixtrack-core:2.0.0 .
```

**Step 2**: Run container

```bash
docker run -d \
  --name helixtrack-core \
  -p 8080:8080 \
  -v $(pwd)/Configurations:/app/Configurations \
  -v $(pwd)/Database:/app/Database \
  -v /var/log/helixtrack:/tmp/htCoreLogs \
  helixtrack-core:2.0.0
```

**Step 3**: Verify container

```bash
docker logs helixtrack-core
docker exec helixtrack-core htCore --version
```

### Method 4: Docker Compose

**Step 1**: Create `docker-compose.yml`

```yaml
version: '3.8'

services:
  helixtrack-core:
    image: helixtrack-core:2.0.0
    container_name: helixtrack-core
    ports:
      - "8080:8080"
    volumes:
      - ./Configurations:/app/Configurations
      - ./Database:/app/Database
      - helixtrack-logs:/tmp/htCoreLogs
    environment:
      - CONFIG_PATH=/app/Configurations/production.json
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

  postgres:
    image: postgres:14-alpine
    container_name: helixtrack-db
    environment:
      - POSTGRES_USER=htcore
      - POSTGRES_PASSWORD=secure_password_here
      - POSTGRES_DB=htcore
    volumes:
      - postgres-data:/var/lib/postgresql/data
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U htcore"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  postgres-data:
  helixtrack-logs:
```

**Step 2**: Start services

```bash
docker-compose up -d
```

**Step 3**: Check status

```bash
docker-compose ps
docker-compose logs -f helixtrack-core
```

---

## Database Setup

### Option A: SQLite (Development/Small Deployments)

SQLite is perfect for development and small teams.

**Step 1**: Create database directory

```bash
mkdir -p Database
```

**Step 2**: Import database schema

```bash
cd Database/DDL
sqlite3 ../Definition.sqlite < Definition.V2.sql
```

**Step 3**: Verify database

```bash
sqlite3 Database/Definition.sqlite "SELECT COUNT(*) FROM sqlite_master WHERE type='table';"
# Output: 53 (number of tables)
```

### Option B: PostgreSQL (Production)

PostgreSQL is recommended for production deployments.

**Step 1**: Install PostgreSQL

```bash
# Ubuntu/Debian
sudo apt install postgresql postgresql-contrib

# macOS
brew install postgresql@14

# Start service
sudo systemctl start postgresql
sudo systemctl enable postgresql
```

**Step 2**: Create database and user

```bash
sudo -u postgres psql

postgres=# CREATE USER htcore WITH PASSWORD 'secure_password_here';
postgres=# CREATE DATABASE htcore OWNER htcore;
postgres=# GRANT ALL PRIVILEGES ON DATABASE htcore TO htcore;
postgres=# \q
```

**Step 3**: Import schema

```bash
cd Database/DDL
psql -U htcore -d htcore -f Definition.V2.sql
```

**Step 4**: Verify database

```bash
psql -U htcore -d htcore -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public';"
# Output: 53
```

---

## Initial Configuration

### Create Configuration File

**Step 1**: Copy default configuration

```bash
cd Application
cp Configurations/default.json Configurations/my-config.json
```

**Step 2**: Edit configuration

```json
{
  "log": {
    "log_path": "/var/log/helixtrack",
    "logfile_base_name": "htCore",
    "log_size_limit": 100000000,
    "level": "info"
  },
  "listeners": [
    {
      "address": "0.0.0.0",
      "port": 8080,
      "https": false
    }
  ],
  "database": {
    "type": "sqlite",
    "sqlite_path": "Database/Definition.sqlite"
  },
  "services": {
    "authentication": {
      "enabled": false,
      "url": "",
      "timeout": 30
    },
    "permissions": {
      "enabled": false,
      "url": "",
      "timeout": 30
    }
  }
}
```

**For PostgreSQL**, change database section:

```json
{
  "database": {
    "type": "postgres",
    "postgres_host": "localhost",
    "postgres_port": 5432,
    "postgres_user": "htcore",
    "postgres_password": "secure_password_here",
    "postgres_database": "htcore",
    "postgres_ssl_mode": "disable"
  }
}
```

---

## First Run

### Start the Application

```bash
# With default config
htCore

# With custom config
htCore --config=Configurations/my-config.json

# With environment variable
export HTCORE_CONFIG=Configurations/my-config.json
htCore
```

### Verify It's Running

**Method 1**: Check health endpoint

```bash
curl http://localhost:8080/health
# Output: {"status":"healthy"}
```

**Method 2**: Check version

```bash
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d '{"action":"version"}'

# Output:
# {
#   "errorCode": -1,
#   "data": {
#     "version": "2.0.0",
#     "api": "2.0.0"
#   }
# }
```

**Method 3**: Check logs

```bash
tail -f /var/log/helixtrack/htCore.log
```

---

## Production Deployment

### Systemd Service (Linux)

**Step 1**: Create service file

```bash
sudo nano /etc/systemd/system/helixtrack-core.service
```

**Step 2**: Add service configuration

```ini
[Unit]
Description=HelixTrack Core API Service
After=network.target postgresql.service

[Service]
Type=simple
User=helixtrack
Group=helixtrack
WorkingDirectory=/opt/helixtrack
ExecStart=/usr/local/bin/htCore --config=/opt/helixtrack/Configurations/production.json
Restart=on-failure
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=helixtrack-core

# Security hardening
PrivateTmp=true
NoNewPrivileges=true
ProtectSystem=strict
ReadWritePaths=/opt/helixtrack/Database /var/log/helixtrack

[Install]
WantedBy=multi-user.target
```

**Step 3**: Enable and start service

```bash
sudo systemctl daemon-reload
sudo systemctl enable helixtrack-core
sudo systemctl start helixtrack-core
```

**Step 4**: Check status

```bash
sudo systemctl status helixtrack-core
sudo journalctl -u helixtrack-core -f
```

### Nginx Reverse Proxy

**Step 1**: Install Nginx

```bash
sudo apt install nginx
```

**Step 2**: Create Nginx configuration

```bash
sudo nano /etc/nginx/sites-available/helixtrack
```

```nginx
upstream helixtrack_backend {
    server localhost:8080;
    keepalive 32;
}

server {
    listen 80;
    server_name helixtrack.example.com;

    # Redirect to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name helixtrack.example.com;

    # SSL certificates
    ssl_certificate /etc/letsencrypt/live/helixtrack.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/helixtrack.example.com/privkey.pem;

    # SSL configuration
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;

    # Logging
    access_log /var/log/nginx/helixtrack-access.log;
    error_log /var/log/nginx/helixtrack-error.log;

    # Proxy settings
    location / {
        proxy_pass http://helixtrack_backend;
        proxy_http_version 1.1;

        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header Connection "";

        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # Health check endpoint
    location /health {
        proxy_pass http://helixtrack_backend/health;
        access_log off;
    }
}
```

**Step 3**: Enable site

```bash
sudo ln -s /etc/nginx/sites-available/helixtrack /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### SSL Certificate (Let's Encrypt)

```bash
# Install Certbot
sudo apt install certbot python3-certbot-nginx

# Obtain certificate
sudo certbot --nginx -d helixtrack.example.com

# Auto-renewal
sudo systemctl enable certbot.timer
```

---

## Verification Checklist

After installation, verify everything is working:

- [ ] Application starts without errors
- [ ] `/health` endpoint returns `{"status":"healthy"}`
- [ ] `/do` endpoint responds to `{"action":"version"}`
- [ ] Database connection is successful (`{"action":"dbCapable"}`)
- [ ] Logs are being written to log directory
- [ ] Systemd service (if used) auto-starts on boot
- [ ] Nginx reverse proxy (if used) is routing correctly
- [ ] SSL certificate (if used) is valid

---

## Next Steps

Now that HelixTrack Core is installed and running, let's configure it for your specific needs.

[Next: Configuration →](03-configuration.md)

---

[← Previous: Introduction](01-introduction.md) | [Back to Table of Contents](README.md) | [Next: Configuration →](03-configuration.md)

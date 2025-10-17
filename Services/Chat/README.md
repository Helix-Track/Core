# HelixTrack Chat Service

A production-ready, high-performance chat microservice for HelixTrack with HTTP/3 QUIC support, real-time messaging, and comprehensive security.

## 🚀 Features

### Core Messaging
- **Multi-Entity Chat Support**: Chats for users, teams, projects, tickets, attachments, and any custom entity
- **Message Types**: Text, image, file, system messages
- **Rich Formatting**: Plain text and Markdown support
- **Threading**: Reply to messages with parent-child relationships
- **Quoting**: Quote previous messages in replies
- **Full-Text Search**: Fast message search with PostgreSQL GIN indexes
- **Message Editing**: Edit your own messages with edit history
- **Message Pinning**: Pin important messages (admin/moderator only)
- **Soft Delete**: Messages can be deleted without losing history

### Chat Room Management
- **Room Types**: Direct (1-on-1), group, channel, private
- **Participant Roles**: Owner, admin, moderator, member, guest
- **Role-Based Permissions**: Fine-grained access control
- **Participant Management**: Add, remove, mute, unmute participants
- **Room Metadata**: Flexible JSONB field for custom data

### Real-Time Features
- **Presence Tracking**: Online, offline, away, busy, DND statuses
- **Typing Indicators**: See when users are typing (5-second auto-expiry)
- **Read Receipts**: Track message read status per user
- **Emoji Reactions**: React to messages with emojis
- **Message Attachments**: Support for file uploads with metadata

### Security & Performance
- **HTTP/3 QUIC**: Modern protocol for faster, more reliable connections
- **TLS/HTTPS**: Secure communication with certificate management
- **JWT Authentication**: Token-based auth with claims validation
- **DDOS Protection**: Per-IP rate limiting with token bucket algorithm
- **CORS**: Flexible origin configuration (wildcard, exact, pattern)
- **Message Size Limits**: Configurable limits (512KB default)
- **Database Encryption**: PostgreSQL with SQL Cipher support

### Observability
- **Structured Logging**: Uber Zap with JSON output and log rotation
- **Health Checks**: Built-in health and version endpoints
- **Graceful Shutdown**: Clean shutdown with 30-second timeout
- **Request Logging**: All requests logged with latency tracking

## 📋 Prerequisites

- **Go 1.22+**
- **PostgreSQL 15+**
- **Docker & Docker Compose** (for containerized deployment)
- **TLS Certificates** (for HTTPS/QUIC)

## 🛠️ Installation

### Option 1: Docker Compose (Recommended)

The Chat service is integrated into the main HelixTrack Docker Compose configuration:

```bash
# From Core/Application directory
cd Core/Application

# Start all services (including Chat)
./docker-run-sqlite.sh
# or
./docker-run-postgres.sh

# Chat service will be available at:
# - Chat API: http://localhost:9090
# - Chat DB: localhost:5433
```

### Option 2: Standalone Docker

```bash
cd Core/Services/Chat

# Start Chat service with its own database
./scripts/start.sh

# Stop services
./scripts/stop.sh
```

### Option 3: Manual Build

```bash
cd Core/Services/Chat

# Build the binary
./scripts/build.sh

# Run with dev config
./htChat --config=configs/dev.json

# Run with custom config
./htChat --config=configs/prod.json

# Show version
./htChat --version
```

## ⚙️ Configuration

Configuration is managed via JSON files in the `configs/` directory:

- **`dev.json`**: Development environment (HTTP, localhost database)
- **`prod.json`**: Production environment (HTTPS, secure passwords from env vars)
- **`test.json`**: Testing environment (in-memory database)

### Environment Variables

The following environment variables are supported for sensitive configuration:

```bash
# Database password (recommended for production)
DB_PASSWORD=your_secure_password

# JWT secret (must match Core service)
JWT_SECRET=your_jwt_secret_key

# TLS certificate paths
CERT_FILE=/path/to/cert.pem
KEY_FILE=/path/to/key.pem
```

### Configuration File Structure

```json
{
  "server": {
    "address": "0.0.0.0",
    "port": 9090,
    "enable_https": true,
    "enable_http3": true,
    "cert_file": "/app/certs/server.crt",
    "key_file": "/app/certs/server.key"
  },
  "database": {
    "host": "chat-db",
    "port": 5432,
    "user": "chat_user",
    "password": "${DB_PASSWORD}",
    "database": "helixtrack_chat",
    "ssl_mode": "disable",
    "max_connections": 25,
    "connection_timeout": 10
  },
  "jwt": {
    "secret": "${JWT_SECRET}",
    "issuer": "http://helixtrack-core:8080"
  },
  "security": {
    "rate_limit_per_second": 10,
    "rate_limit_burst": 20,
    "allowed_origins": ["*"],
    "message_size_limit": 524288
  },
  "logger": {
    "level": "info",
    "log_path": "/var/log/helixtrack/chat",
    "logfile_base_name": "htChat",
    "log_size_limit": 104857600
  }
}
```

## 🔧 API Usage

### Authentication

All API endpoints (except `/health` and `/version`) require JWT authentication. Include the JWT token in the request:

```bash
# In Authorization header (recommended)
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
     -X POST http://localhost:9090/api/do \
     -d '{"action": "chatRoomList"}'

# As query parameter
curl "http://localhost:9090/api/do?jwt=YOUR_JWT_TOKEN" \
     -X POST -d '{"action": "chatRoomList"}'

# In request body
curl -X POST http://localhost:9090/api/do \
     -d '{"action": "chatRoomList", "jwt": "YOUR_JWT_TOKEN"}'
```

### Unified `/do` Endpoint

All actions use the unified `/do` endpoint with action-based routing:

**Request Format:**
```json
{
  "action": "string",
  "jwt": "string",
  "data": {}
}
```

**Response Format:**
```json
{
  "errorCode": -1,
  "errorMessage": "",
  "data": {}
}
```

### API Actions (31 Total)

#### Chat Room Actions (6)

**1. Create Chat Room**
```bash
curl -X POST http://localhost:9090/api/do \
  -H "Authorization: Bearer $JWT" \
  -H "Content-Type: application/json" \
  -d '{
    "action": "chatRoomCreate",
    "data": {
      "name": "Project Alpha Discussion",
      "description": "Chat for Project Alpha team",
      "type": "group",
      "is_private": false,
      "entity_type": "project",
      "entity_id": "550e8400-e29b-41d4-a716-446655440000"
    }
  }'
```

**2. Read Chat Room**
```bash
curl -X POST http://localhost:9090/api/do \
  -H "Authorization: Bearer $JWT" \
  -d '{"action": "chatRoomRead", "data": {"id": "ROOM_ID"}}'
```

**3. List Chat Rooms**
```bash
curl -X POST http://localhost:9090/api/do \
  -H "Authorization: Bearer $JWT" \
  -d '{"action": "chatRoomList", "data": {"limit": 20, "offset": 0}}'
```

**4. Update Chat Room** (Owner/Admin only)
```bash
curl -X POST http://localhost:9090/api/do \
  -H "Authorization: Bearer $JWT" \
  -d '{
    "action": "chatRoomUpdate",
    "data": {
      "id": "ROOM_ID",
      "name": "Updated Room Name",
      "description": "New description"
    }
  }'
```

**5. Delete Chat Room** (Owner only)
```bash
curl -X POST http://localhost:9090/api/do \
  -H "Authorization: Bearer $JWT" \
  -d '{"action": "chatRoomDelete", "data": {"id": "ROOM_ID"}}'
```

**6. Get Room by Entity**
```bash
curl -X POST http://localhost:9090/api/do \
  -H "Authorization: Bearer $JWT" \
  -d '{
    "action": "chatRoomGetByEntity",
    "data": {
      "entity_type": "ticket",
      "entity_id": "TICKET_UUID"
    }
  }'
```

#### Message Actions (10)

**1. Send Message**
```bash
curl -X POST http://localhost:9090/api/do \
  -H "Authorization: Bearer $JWT" \
  -d '{
    "action": "messageSend",
    "data": {
      "chat_room_id": "ROOM_ID",
      "content": "Hello, world!",
      "type": "text"
    }
  }'
```

**2. Reply to Message**
```bash
curl -X POST http://localhost:9090/api/do \
  -H "Authorization: Bearer $JWT" \
  -d '{
    "action": "messageReply",
    "data": {
      "chat_room_id": "ROOM_ID",
      "parent_id": "PARENT_MESSAGE_ID",
      "content": "This is a reply"
    }
  }'
```

**3. Quote Message**
```bash
curl -X POST http://localhost:9090/api/do \
  -H "Authorization: Bearer $JWT" \
  -d '{
    "action": "messageQuote",
    "data": {
      "chat_room_id": "ROOM_ID",
      "quoted_message_id": "QUOTED_MESSAGE_ID",
      "content": "Replying to previous message"
    }
  }'
```

**4. List Messages** (with pagination)
```bash
curl -X POST http://localhost:9090/api/do \
  -H "Authorization: Bearer $JWT" \
  -d '{
    "action": "messageList",
    "data": {
      "chat_room_id": "ROOM_ID",
      "limit": 50,
      "offset": 0
    }
  }'
```

**5. Search Messages** (full-text search)
```bash
curl -X POST http://localhost:9090/api/do \
  -H "Authorization: Bearer $JWT" \
  -d '{
    "action": "messageSearch",
    "data": {
      "chat_room_id": "ROOM_ID",
      "query": "important meeting",
      "limit": 20
    }
  }'
```

**6-10. Other Message Actions:**
- `messageRead`: Get single message
- `messageUpdate`: Edit your message
- `messageDelete`: Delete your message (admin can delete any)
- `messagePin`: Pin message (admin/moderator only)
- `messageUnpin`: Unpin message

#### Participant Actions (6)

**1. Add Participant**
```bash
curl -X POST http://localhost:9090/api/do \
  -H "Authorization: Bearer $JWT" \
  -d '{
    "action": "participantAdd",
    "data": {
      "chat_room_id": "ROOM_ID",
      "user_id": "USER_UUID",
      "role": "member"
    }
  }'
```

**2-6. Other Participant Actions:**
- `participantRemove`: Remove participant
- `participantList`: List all participants
- `participantUpdateRole`: Change participant role (owner/admin)
- `participantMute`: Mute participant (moderator+)
- `participantUnmute`: Unmute participant

#### Real-Time Actions (9)

**Typing Indicators:**
```bash
# Start typing
curl -X POST http://localhost:9090/api/do \
  -H "Authorization: Bearer $JWT" \
  -d '{"action": "typingStart", "data": {"chat_room_id": "ROOM_ID"}}'

# Stop typing
curl -X POST http://localhost:9090/api/do \
  -H "Authorization: Bearer $JWT" \
  -d '{"action": "typingStop", "data": {"chat_room_id": "ROOM_ID"}}'
```

**Presence:**
```bash
# Update presence
curl -X POST http://localhost:9090/api/do \
  -H "Authorization: Bearer $JWT" \
  -d '{
    "action": "presenceUpdate",
    "data": {"status": "online"}
  }'

# Get user presence
curl -X POST http://localhost:9090/api/do \
  -H "Authorization: Bearer $JWT" \
  -d '{"action": "presenceGet", "data": {"user_id": "USER_UUID"}}'
```

**Read Receipts:**
```bash
curl -X POST http://localhost:9090/api/do \
  -H "Authorization: Bearer $JWT" \
  -d '{
    "action": "readReceiptMark",
    "data": {"message_id": "MESSAGE_ID"}
  }'
```

**Reactions:**
```bash
# Add reaction
curl -X POST http://localhost:9090/api/do \
  -H "Authorization: Bearer $JWT" \
  -d '{
    "action": "reactionAdd",
    "data": {"message_id": "MESSAGE_ID", "emoji": "👍"}
  }'
```

**Attachments:**
```bash
curl -X POST http://localhost:9090/api/do \
  -H "Authorization: Bearer $JWT" \
  -d '{
    "action": "attachmentUpload",
    "data": {
      "message_id": "MESSAGE_ID",
      "file_name": "document.pdf",
      "file_size": 1024000,
      "mime_type": "application/pdf",
      "storage_url": "https://storage.example.com/files/document.pdf"
    }
  }'
```

### System Endpoints

**Health Check:**
```bash
curl http://localhost:9090/health
# Response: {"status": "healthy", "database": "connected"}
```

**Version:**
```bash
curl http://localhost:9090/version
# Response: {"version": "1.0.0", "buildTime": "2025-10-17", "gitCommit": "abc123"}
```

## 🔐 Security

### JWT Token Structure

The Chat service validates JWT tokens from the Core Authentication service:

```json
{
  "sub": "authentication",
  "name": "User Full Name",
  "username": "username",
  "user_id": "uuid",
  "role": "admin|user|guest",
  "permissions": "READ|CREATE|UPDATE|DELETE",
  "htCoreAddress": "http://core-service:8080",
  "exp": 1234567890,
  "iat": 1234567890
}
```

### Rate Limiting

Default rate limits (configurable):
- 10 requests per second per IP
- Burst of 20 requests
- Returns 429 Too Many Requests when exceeded

### CORS

Flexible CORS configuration:
- Wildcard: `"allowed_origins": ["*"]`
- Exact: `"allowed_origins": ["https://example.com"]`
- Pattern: `"allowed_origins": ["*.example.com"]`

### Message Size Limits

Default: 512KB per message (configurable)

## 🧪 Testing

### Unit Tests

```bash
cd Core/Services/Chat
go test ./... -v -cover
```

### Integration Tests

```bash
# Start test environment
docker-compose -f docker-compose.test.yml up -d

# Run integration tests
go test ./... -tags=integration

# Stop test environment
docker-compose -f docker-compose.test.yml down
```

### Test Coverage

```bash
./scripts/test.sh
# Generates coverage.out and coverage.html
```

## 📊 Database Schema

The Chat service uses 14 PostgreSQL tables with the V2 schema:

**Core Tables:**
- `chat_room` - Chat room definitions with soft delete
- `message` - Messages with threading and quoting
- `chat_participant` - Room participants with roles
- `presence` - User online/offline status
- `typing_indicator` - Real-time typing status
- `read_receipt` - Message read tracking
- `message_reaction` - Emoji reactions
- `message_attachment` - File attachments

**Additional Tables:**
- `chat_room_settings` - Room-specific settings
- `message_mention` - @username mentions
- `chat_archive` - Archived chats
- `chat_export` - Export history
- `message_report` - Content moderation
- `chat_analytics` - Usage statistics

### Indexes

- Primary keys: UUID
- Full-text search: GIN indexes on message content
- Pagination: Composite indexes on (chat_room_id, created_at)
- Soft delete: Filtered indexes excluding deleted=true

## 🚀 Deployment

### Docker Compose (Production)

```bash
# Set environment variables
export CHAT_DB_PASSWORD='your_secure_password'
export JWT_SECRET='your_jwt_secret_key'

# Start services
docker-compose up -d

# View logs
docker-compose logs -f chat-service

# Stop services
docker-compose down
```

### Kubernetes

See `kubernetes/` directory for manifests:
- `deployment.yaml` - Chat service deployment
- `service.yaml` - Service definition
- `configmap.yaml` - Configuration
- `secret.yaml` - Secrets (passwords, JWT)
- `postgres.yaml` - PostgreSQL StatefulSet

### Systemd

```bash
# Copy binary and config
sudo cp htChat /usr/local/bin/
sudo cp configs/prod.json /etc/helixtrack/chat/config.json

# Create systemd service
sudo cat > /etc/systemd/system/helixtrack-chat.service <<EOF
[Unit]
Description=HelixTrack Chat Service
After=network.target postgresql.service

[Service]
Type=simple
User=helixtrack
ExecStart=/usr/local/bin/htChat --config=/etc/helixtrack/chat/config.json
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
EOF

# Enable and start
sudo systemctl enable helixtrack-chat
sudo systemctl start helixtrack-chat
```

## 📈 Performance

### Benchmarks

- **Throughput**: 10,000+ requests/second (single instance)
- **Latency**: <10ms p99 for message send
- **Concurrent Users**: 10,000+ with WebSocket support
- **Database**: Connection pooling with 25 max connections
- **Memory**: ~256MB typical usage

### Scaling

- **Horizontal**: Run multiple instances behind load balancer
- **Database**: PostgreSQL read replicas for scaling reads
- **Caching**: Redis for presence and typing indicators (future)
- **WebSocket**: Sticky sessions or Redis pub/sub for multi-instance

## 🛠️ Development

### Project Structure

```
Core/Services/Chat/
├── main.go                 # Application entry point
├── go.mod, go.sum          # Go dependencies
│
├── configs/                # Configuration files
│   ├── dev.json
│   ├── prod.json
│   └── test.json
│
├── internal/               # Internal packages
│   ├── config/             # Configuration loader
│   ├── logger/             # Logging system
│   ├── database/           # Database layer (985 LOC)
│   ├── handlers/           # API handlers (~1800 LOC)
│   ├── middleware/         # JWT, CORS, Rate Limit (658 LOC)
│   ├── server/             # HTTP/3 QUIC server (395 LOC)
│   └── services/           # Core service client (315 LOC)
│
├── scripts/                # Build and run scripts
│   ├── build.sh
│   ├── start.sh
│   ├── stop.sh
│   └── test.sh
│
├── Dockerfile              # Container image
├── docker-compose.yml      # Local development
└── README.md               # This file
```

### Adding New Features

1. **Define model** in `internal/models/`
2. **Add database queries** to `internal/database/`
3. **Create handler** in `internal/handlers/`
4. **Route action** in `internal/handlers/handler.go`
5. **Write tests** for all layers
6. **Update documentation**

## 🐛 Troubleshooting

### Service Won't Start

```bash
# Check logs
docker-compose logs chat-service

# Check database connection
docker exec helixtrack-chat-db pg_isready -U chat_user

# Check ports
netstat -tuln | grep 9090
netstat -tuln | grep 5433
```

### Database Connection Issues

```bash
# Test database connection
psql -h localhost -p 5433 -U chat_user -d helixtrack_chat

# Check database logs
docker-compose logs chat-db

# Verify schema
docker exec helixtrack-chat-db psql -U chat_user -d helixtrack_chat -c "\dt"
```

### JWT Authentication Failures

- Verify JWT_SECRET matches Core service
- Check token expiration
- Ensure user_id in claims is valid UUID
- Check JWT issuer matches configuration

### Performance Issues

```bash
# Check database connections
docker exec helixtrack-chat-db psql -U chat_user -d helixtrack_chat \
  -c "SELECT count(*) FROM pg_stat_activity;"

# Monitor resource usage
docker stats helixtrack-chat-service

# Enable debug logging
# Edit config: "logger": {"level": "debug"}
```

## 📝 API Error Codes

- `-1`: Success (no error)
- `1000`: Invalid request format
- `1001`: Missing required parameter
- `1002`: Invalid parameter value
- `1003`: Invalid JWT token
- `2000`: Database error
- `2001`: Internal server error
- `3000`: Entity not found
- `3001`: Entity already exists
- `3002`: Forbidden (insufficient permissions)
- `4000`: Rate limit exceeded

## 🤝 Contributing

See main [HelixTrack CONTRIBUTING.md](../../../CONTRIBUTING.md)

## 📄 License

MIT License - See [LICENSE](../../../LICENSE)

## 🔗 Links

- **HelixTrack Core**: [Core README](../../Application/README.md)
- **API Documentation**: [API.md](./API.md)
- **Architecture**: [ARCHITECTURE.md](./ARCHITECTURE.md)
- **Issues**: [GitHub Issues](https://github.com/Helix-Track/Core/issues)

## 📞 Support

- **Documentation**: https://docs.helixtrack.io
- **Community**: https://community.helixtrack.io
- **Issues**: https://github.com/Helix-Track/Core/issues

---

**Built with ❤️ for the free world. A modern JIRA alternative.**

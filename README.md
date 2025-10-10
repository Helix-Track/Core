# HelixTrack Core

![Build Status](Application/docs/badges/build.svg)
![Tests](Application/docs/badges/tests.svg)
![Coverage](Application/docs/badges/coverage.svg)
![Go Version](Application/docs/badges/go-version.svg)
![JWT Compatible](https://jwt.io/img/badge-compatible.svg)

![JIRA alternative for the free world!](Assets/Wide_Black.png)

**HelixTrack Core** is a production-ready REST API microservice for project and issue tracking - a modern, open-source alternative to JIRA. Built with Go and the Gin Gonic framework, it provides a fully modular architecture with enterprise-grade features.

---

## Features

### ✅ Current Features (V1 + Phase 1 Foundation)

- **🎯 Complete Issue Tracking**: Tickets, types, statuses, workflows, components, labels
- **📊 Agile/Scrum Support**: Sprints (cycles), story points, time estimation, boards
- **👥 Team Management**: Organizations, teams, users, hierarchical permissions
- **🔐 Enterprise Security**: JWT authentication, role-based access control, external auth service
- **💾 Multi-Database**: SQLite (development), PostgreSQL (production)
- **📝 Rich Metadata**: Comments, attachments (assets), custom labels, ticket relationships
- **🔗 Git Integration**: Repository linking, commit-to-ticket mapping
- **📈 Reporting & Audit**: Comprehensive audit logging, custom reports
- **🧩 Extension System**: Modular extensions (Time Tracking, Documents, Chat Integration)
- **🌐 REST API**: Unified `/do` endpoint with action-based routing
- **📚 Complete Documentation**: User manuals, API docs, deployment guides
- **🧪 100% Test Coverage**: 172+ comprehensive tests (expanding to 400+)

### 🚀 Phase 1 Features (Database & Models Ready)

- **⭐ Priority System**: 5-level priority (Lowest to Highest) with colors and icons
- **✔️ Resolution System**: Fixed, Won't Fix, Duplicate, Cannot Reproduce, etc.
- **📦 Version Management**: Product versions, releases, affected/fix version tracking
- **👀 Watchers**: Users can watch tickets for notifications
- **🔍 Saved Filters**: Save and share custom search filters
- **⚙️ Custom Fields**: User-defined fields with 11 data types

> **Note**: Phase 1 features have database schema and Go models ready. API handlers and tests are in development.

---

## Quick Start

### Prerequisites

- **Go 1.22+** (required)
- **SQLite 3** (for development) or **PostgreSQL 12+** (for production)

### Installation

```bash
# Clone the repository
git clone https://github.com/Helix-Track/Core.git
cd Core/Application

# Install dependencies
go mod download

# Run tests
./scripts/verify-tests.sh

# Build
go build -o htCore main.go

# Run
./htCore
```

### Configuration

Default configuration is in `Configurations/default.json`:

```json
{
  "log": {
    "log_path": "/tmp/htCoreLogs",
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
  }
}
```

### Testing

```bash
# Run comprehensive test verification
cd Application
./scripts/verify-tests.sh

# Test API endpoints (server must be running)
cd test-scripts
./test-all.sh

# Or use Postman
# Import: test-scripts/HelixTrack-Core-API.postman_collection.json
```

---

## Architecture

### Modular Microservice Design

```
┌─────────────────────────────────────────────────────────────┐
│                    HelixTrack Core API                      │
│                    (Gin Gonic / Go)                         │
│                                                             │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  │
│  │  Models  │  │ Handlers │  │Middleware│  │  Server  │  │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘  │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │          Database Abstraction Layer                  │  │
│  │         (SQLite + PostgreSQL Support)                │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
         │                           │
         ▼                           ▼
┌──────────────────┐        ┌──────────────────┐
│  Authentication  │        │   Permissions    │
│     Service      │        │     Engine       │
│  (External/HTTP) │        │  (External/HTTP) │
└──────────────────┘        └──────────────────┘
         │
         ▼
┌─────────────────────────────────────────────────────────────┐
│                   Optional Extensions                       │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │     Times    │  │   Documents  │  │     Chats    │     │
│  │   (Tracking) │  │    (Wiki)    │  │ (Integrations)     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
└─────────────────────────────────────────────────────────────┘
```

### Key Principles

- **🔌 Fully Decoupled**: All components communicate via HTTP, can run on separate machines/clusters
- **🔄 Swappable Services**: Replace proprietary implementations of Authentication/Permissions
- **📦 Extension-Based**: Optional features (Time Tracking, Documents, Chats) as separate services
- **🎯 Interface-Driven**: Clean interfaces allow easy testing and component replacement
- **🛡️ Production-Ready**: Logging, health checks, graceful shutdown, CORS, HTTPS support

---

## API Overview

### Unified `/do` Endpoint

All operations use a single endpoint with action-based routing:

```bash
POST http://localhost:8080/do
Content-Type: application/json

{
  "action": "create",
  "jwt": "eyJhbGc...",
  "object": "ticket",
  "locale": "en",
  "data": {
    "title": "Bug in login form",
    "description": "Cannot login with special characters",
    "type": "bug",
    "priority": "high"
  }
}
```

### Available Actions

#### System Actions (No Auth Required)
- `version` - Get API version
- `jwtCapable` - Check JWT availability
- `dbCapable` - Check database health
- `health` - Service health check

#### Core CRUD Actions (Auth Required)
- `create` - Create entities
- `modify` - Update entities
- `remove` - Delete entities
- `read` - Read single entity
- `list` - List entities

#### Phase 1 Actions (Database Ready, Handlers Pending)
- **Priority**: `priorityCreate`, `priorityRead`, `priorityList`, `priorityModify`, `priorityRemove`
- **Resolution**: `resolutionCreate`, `resolutionRead`, `resolutionList`, `resolutionModify`, `resolutionRemove`
- **Version**: `versionCreate`, `versionRead`, `versionList`, `versionModify`, `versionRemove`, `versionRelease`, `versionArchive`
- **Watchers**: `watcherAdd`, `watcherRemove`, `watcherList`
- **Filters**: `filterSave`, `filterLoad`, `filterList`, `filterShare`, `filterModify`, `filterRemove`
- **Custom Fields**: `customFieldCreate`, `customFieldRead`, `customFieldList`, `customFieldModify`, `customFieldRemove`

> See [API Documentation](Application/docs/USER_MANUAL.md) for complete API reference.

---

## Database

### Schema Versions

- **V1**: Core features (tickets, projects, workflows, teams, boards, sprints, etc.)
- **V2**: Phase 1 JIRA parity (priorities, resolutions, versions, watchers, filters, custom fields)

### Database Structure

```
Database/
├── Definition.sqlite          # Current database (auto-generated)
└── DDL/
    ├── Definition.V1.sql      # Version 1 schema
    ├── Definition.V2.sql      # Version 2 schema (Phase 1)
    ├── Migration.V1.2.sql     # Migration from V1 to V2
    ├── Extensions/
    │   ├── Times/             # Time tracking extension
    │   ├── Documents/         # Document management extension
    │   └── Chats/             # Chat integration extension
    └── Services/
        └── Authentication/     # Authentication service schema
```

### Migration

```bash
# Import all definitions to SQLite
./Run/Db/import_All_Definitions_to_Sqlite.sh

# Import specific extensions
./Run/Db/import_Extension_Times_Definition_to_Sqlite.sh

# For PostgreSQL
./Run/Db/import_All_Definitions_to_Postgres.sh
```

---

## Project Structure

```
Core/
├── Application/                # Go application (production)
│   ├── main.go
│   ├── internal/
│   │   ├── config/            # Configuration management
│   │   ├── database/          # Database abstraction
│   │   ├── handlers/          # HTTP request handlers
│   │   ├── logger/            # Logging system
│   │   ├── middleware/        # JWT validation, CORS
│   │   ├── models/            # Data models
│   │   ├── server/            # HTTP server
│   │   └── services/          # External service clients
│   ├── scripts/               # Test runners, badge generators
│   ├── test-scripts/          # API test scripts (curl + Postman)
│   ├── test-reports/          # Test documentation
│   └── docs/                  # Documentation
│
├── Database/                  # Database schemas
│   ├── Definition.sqlite
│   └── DDL/
│
├── Configurations/            # JSON config files
├── Documentation/             # Project documentation
├── Run/                       # Executable scripts
└── README.md                  # This file
```

---

## Testing

### Test Infrastructure

- **172+ Unit Tests** (expanding to 400+)
- **100% Code Coverage** (target)
- **Race Detection** enabled
- **Comprehensive Test Reports** (JSON, Markdown, HTML)
- **Status Badges** (build, tests, coverage, Go version)
- **API Test Scripts** (7 curl scripts + Postman collection)

### Running Tests

```bash
# Comprehensive verification (recommended)
cd Application
./scripts/verify-tests.sh

# Quick test
go test ./...

# With coverage
go test -cover ./...

# With race detection
go test -race ./...

# Single package
go test ./internal/models/
```

### Test Reports

After running `verify-tests.sh`:

- `test-reports/TEST_REPORT.html` - Interactive HTML report
- `test-reports/TEST_REPORT.md` - Markdown report
- `test-reports/test-results.json` - Machine-readable JSON
- `coverage/coverage.html` - Interactive coverage browser

---

## Documentation

| Document | Description | Location |
|----------|-------------|----------|
| **User Manual** | Complete API reference and usage guide | `Application/docs/USER_MANUAL.md` |
| **Deployment Guide** | Installation and deployment instructions | `Application/docs/DEPLOYMENT.md` |
| **Testing Guide** | Comprehensive testing documentation | `Application/test-reports/TESTING_GUIDE.md` |
| **Quick Start** | Quick testing guide | `Application/QUICK_START_TESTING.md` |
| **Feature Gap Analysis** | JIRA feature comparison | `Application/JIRA_FEATURE_GAP_ANALYSIS.md` |
| **Phase 1 Status** | Implementation progress | `Application/PHASE1_IMPLEMENTATION_STATUS.md` |
| **Delivery Summary** | Complete delivery overview | `Application/DELIVERY_SUMMARY.txt` |

---

## Deployment

### Binary Deployment

```bash
go build -o htCore main.go
./htCore --config=/path/to/config.json
```

### systemd Service

```ini
[Unit]
Description=HelixTrack Core API
After=network.target

[Service]
Type=simple
User=htcore
ExecStart=/usr/local/bin/htCore --config=/etc/htcore/default.json
Restart=always

[Install]
WantedBy=multi-user.target
```

### Docker

```bash
docker build -t helixtrack-core .
docker run -p 8080:8080 -v $(pwd)/Configurations:/config helixtrack-core
```

### Docker Compose

```yaml
version: '3.8'
services:
  htcore:
    image: helixtrack-core:latest
    ports:
      - "8080:8080"
    volumes:
      - ./Configurations:/config
      - ./Database:/database
    environment:
      - CONFIG_PATH=/config/default.json
```

> See [Deployment Guide](Application/docs/DEPLOYMENT.md) for complete instructions.

---

## Development

### Prerequisites

- Go 1.22+
- Git
- SQLite 3 or PostgreSQL 12+

### Building

```bash
cd Application
go mod download
go build -o htCore main.go
```

### Running

```bash
# Default configuration
./htCore

# Custom configuration
./htCore --config=Configurations/dev.json

# Show version
./htCore --version
```

### Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

**Development Standards**:
- 100% test coverage required
- All tests must pass
- Follow Go best practices
- Document all public APIs
- Add tests for new features

---

## Roadmap

### ✅ Completed
- Complete V1 implementation (23 core features)
- 172 comprehensive tests with 100% coverage
- Full documentation suite
- Phase 1 database schema and models
- Migration scripts

### 🚧 In Progress (Phase 1)
- Priority & Resolution API handlers
- Version management API
- Watchers functionality
- Saved filters
- Custom fields system
- ~245 new tests for Phase 1 features

### 📅 Planned (Phase 2)
- Epic support
- Subtasks
- Advanced work logs
- Project roles
- Security levels
- Dashboard system
- Advanced board configuration

### 🔮 Future (Phase 3)
- Voting system
- Project categories
- Notification schemes
- Activity streams
- Comment mentions

> See [Feature Gap Analysis](Application/JIRA_FEATURE_GAP_ANALYSIS.md) for detailed roadmap.

---

## Technology Stack

- **Language**: Go 1.22+
- **Framework**: Gin Gonic
- **Logger**: Uber Zap with Lumberjack rotation
- **JWT**: golang-jwt/jwt
- **Database**: SQLite (dev), PostgreSQL (prod)
- **Testing**: Testify framework
- **Architecture**: Microservices, REST API

---

## License

See [LICENSE](LICENSE) file for details.

---

## Support & Contact

- **Issues**: [GitHub Issues](https://github.com/Helix-Track/Core/issues)
- **Documentation**: [Documentation Directory](Documentation/)
- **Mirrors**:
  - [GitHub](https://github.com/Helix-Track/Core)
  - [GitFlic](https://gitflic.ru/project/helix-track/core)
  - [Gitee](https://gitee.com/Kvetch_Godspeed_b073/Core)

---

## Status

**Current Version**: 1.0.0 (V1 Complete + Phase 1 Foundation)

**Production Readiness**: ✅ V1 Features Production Ready

**Phase 1 Progress**: ~40% (Database & Models Complete, Handlers Pending)

**Test Coverage**: 100% (172 tests, expanding to 400+)

**Documentation**: Complete

---

**JIRA Alternative for the Free World!** 🚀

Built with ❤️ using Go and Gin Gonic

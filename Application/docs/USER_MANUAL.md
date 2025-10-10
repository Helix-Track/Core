# HelixTrack Core - User Manual

## Table of Contents

1. [Introduction](#introduction)
2. [Installation](#installation)
3. [Configuration](#configuration)
4. [Running the Application](#running-the-application)
5. [API Reference](#api-reference)
6. [Testing](#testing)
7. [Troubleshooting](#troubleshooting)
8. [Architecture](#architecture)

## Introduction

HelixTrack Core is a modern, modular REST API service built with Go and the Gin Gonic framework. It serves as the main microservice for the HelixTrack project - a JIRA alternative for the free world.

### Key Features

- **Unified `/do` Endpoint**: Action-based routing for all operations
- **JWT Authentication**: Secure token-based authentication
- **Multi-Database Support**: SQLite and PostgreSQL
- **Modular Architecture**: Pluggable authentication and permission services
- **Extension System**: Optional extension services (Chats, Documents, Times)
- **Fully Decoupled**: All components can run on separate machines or clusters
- **Comprehensive Testing**: 100% test coverage
- **Production Ready**: Proper logging, graceful shutdown, health checks

### System Requirements

- Go 1.22 or higher
- SQLite 3 or PostgreSQL 12+ (for database)
- Linux, macOS, or Windows

## Installation

### From Source

1. Clone the repository:
```bash
git clone <repository-url>
cd Core/Application
```

2. Install dependencies:
```bash
go mod download
```

3. Build the application:
```bash
go build -o htCore main.go
```

### Binary Installation

Download the pre-built binary for your platform from the releases page and place it in your PATH.

## Configuration

HelixTrack Core uses JSON configuration files located in the `Configurations/` directory.

### Configuration File Structure

```json
{
  "log": {
    "log_path": "/tmp/htCoreLogs",
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
      "url": "http://localhost:8081",
      "timeout": 30
    },
    "permissions": {
      "enabled": false,
      "url": "http://localhost:8082",
      "timeout": 30
    }
  }
}
```

### Configuration Options

#### Log Configuration

- `log_path`: Directory for log files (default: `/tmp/htCoreLogs`)
- `logfile_base_name`: Base name for log files (default: `htCore`)
- `log_size_limit`: Maximum log file size in bytes (default: 100MB)
- `level`: Log level - `debug`, `info`, `warn`, `error` (default: `info`)

#### Listener Configuration

- `address`: IP address to bind to (use `0.0.0.0` for all interfaces)
- `port`: Port number to listen on
- `https`: Enable HTTPS (requires `cert_file` and `key_file`)
- `cert_file`: Path to SSL certificate (required if `https: true`)
- `key_file`: Path to SSL private key (required if `https: true`)

#### Database Configuration

**SQLite:**
```json
{
  "type": "sqlite",
  "sqlite_path": "Database/Definition.sqlite"
}
```

**PostgreSQL:**
```json
{
  "type": "postgres",
  "postgres_host": "localhost",
  "postgres_port": 5432,
  "postgres_user": "htcore",
  "postgres_password": "secret",
  "postgres_database": "htcore",
  "postgres_ssl_mode": "disable"
}
```

#### Services Configuration

- **Authentication Service**: Provides JWT token validation
  - `enabled`: Enable/disable authentication service
  - `url`: Authentication service endpoint
  - `timeout`: Request timeout in seconds

- **Permissions Service**: Provides permission checking
  - `enabled`: Enable/disable permission service
  - `url`: Permission service endpoint
  - `timeout`: Request timeout in seconds

### Environment-Specific Configurations

Create different configuration files for different environments:

- `Configurations/default.json` - Default configuration
- `Configurations/dev.json` - Development environment
- `Configurations/production.json` - Production environment

## Running the Application

### Basic Usage

```bash
# Run with default configuration
./htCore

# Run with custom configuration
./htCore -config=/path/to/config.json

# Show version
./htCore -version
```

### Running in Development Mode

```bash
# With SQLite (no external dependencies)
./htCore -config=Configurations/dev.json
```

### Running in Production

```bash
# With PostgreSQL and all services enabled
./htCore -config=Configurations/production.json
```

### Docker Deployment

```bash
# Build Docker image
docker build -t helixtrack-core:latest .

# Run container
docker run -d \
  -p 8080:8080 \
  -v /path/to/config.json:/app/config.json \
  -v /path/to/database:/app/Database \
  helixtrack-core:latest
```

## API Reference

### Unified `/do` Endpoint

All API operations use the `/do` endpoint with action-based routing.

#### Request Format

```json
{
  "action": "string",      // Required: action to perform
  "jwt": "string",         // Required for authenticated actions
  "locale": "string",      // Optional: locale for localized responses
  "object": "string",      // Required for CRUD operations
  "data": {}               // Additional action-specific data
}
```

#### Response Format

```json
{
  "errorCode": -1,                    // -1 means success
  "errorMessage": "string",           // Error message (if any)
  "errorMessageLocalised": "string",  // Localized error message
  "data": {}                          // Response data
}
```

### Public Endpoints (No Authentication Required)

#### Version

Get API version information.

**Request:**
```json
{
  "action": "version"
}
```

**Response:**
```json
{
  "errorCode": -1,
  "data": {
    "version": "1.0.0",
    "api": "1.0.0"
  }
}
```

#### JWT Capable

Check if JWT authentication is available.

**Request:**
```json
{
  "action": "jwtCapable"
}
```

**Response:**
```json
{
  "errorCode": -1,
  "data": {
    "jwtCapable": true,
    "enabled": true
  }
}
```

#### DB Capable

Check database availability and health.

**Request:**
```json
{
  "action": "dbCapable"
}
```

**Response:**
```json
{
  "errorCode": -1,
  "data": {
    "dbCapable": true,
    "type": "sqlite"
  }
}
```

#### Health

Get service health status.

**Request:**
```json
{
  "action": "health"
}
```

**Response:**
```json
{
  "errorCode": -1,
  "data": {
    "status": "healthy",
    "checks": {
      "database": "healthy",
      "authService": "enabled",
      "permissionService": "enabled"
    }
  }
}
```

### Authentication Endpoint

#### Authenticate

Authenticate user credentials.

**Request:**
```json
{
  "action": "authenticate",
  "data": {
    "username": "testuser",
    "password": "testpass"
  }
}
```

**Response:**
```json
{
  "errorCode": -1,
  "data": {
    "username": "testuser",
    "role": "admin",
    "name": "Test User"
  }
}
```

### Protected Endpoints (Authentication Required)

#### Create

Create a new entity.

**Request:**
```json
{
  "action": "create",
  "jwt": "your-jwt-token",
  "object": "project",
  "data": {
    "name": "New Project",
    "description": "Project description"
  }
}
```

#### Modify

Modify an existing entity.

**Request:**
```json
{
  "action": "modify",
  "jwt": "your-jwt-token",
  "object": "project",
  "data": {
    "id": "123",
    "name": "Updated Project"
  }
}
```

#### Remove

Remove an entity.

**Request:**
```json
{
  "action": "remove",
  "jwt": "your-jwt-token",
  "object": "project",
  "data": {
    "id": "123"
  }
}
```

#### Read

Read a specific entity.

**Request:**
```json
{
  "action": "read",
  "jwt": "your-jwt-token",
  "data": {
    "id": "123"
  }
}
```

#### List

List entities.

**Request:**
```json
{
  "action": "list",
  "jwt": "your-jwt-token",
  "data": {
    "filter": {},
    "limit": 50,
    "offset": 0
  }
}
```

### V2.0 Features - Complete API Coverage

HelixTrack Core V2.0 provides comprehensive API coverage with **235 total endpoints** across all major features. The basic endpoints shown above are just a small subset.

**For complete API documentation with all 235 endpoints**, see [API_REFERENCE_COMPLETE.md](API_REFERENCE_COMPLETE.md).

#### Feature Summary

**Public & Authentication** (6 actions):
- System health and capability checks
- User authentication

**Generic CRUD** (5 actions):
- Create, Read, Update, Delete, List operations for any entity

**Phase 1 - JIRA Parity** (45 actions):
- Priority Management (5 actions) - Lowest to Highest priority levels
- Resolution Management (5 actions) - Done, Won't Fix, Duplicate, etc.
- Version Management (13 actions) - Release tracking with affected/fix versions
- Watcher Management (3 actions) - Subscribe to ticket notifications
- Filter Management (6 actions) - Save and share custom filters
- Custom Field Management (13 actions) - 11 field types (text, number, date, select, etc.)

**Workflow Engine** (23 actions):
- Workflow Management (5 actions) - Define ticket workflows
- Workflow Step Management (5 actions) - Configure workflow transitions
- Ticket Status Management (5 actions) - Open, In Progress, Resolved, Closed, etc.
- Ticket Type Management (8 actions) - Bug, Task, Story, Epic, etc.

**Agile/Scrum Support** (23 actions):
- Board Management (12 actions) - Kanban/Scrum boards with metadata
- Cycle Management (11 actions) - Sprints, Milestones, Releases

**Multi-Tenancy** (28 actions):
- Account Management (5 actions) - Top-level tenant management
- Organization Management (7 actions) - Department/division hierarchy
- Team Management (10 actions) - Team creation and project assignment
- User Mappings (6 actions) - User-organization and user-team relationships

**Supporting Systems** (42 actions):
- Component Management (12 actions) - Project components with metadata
- Label Management (16 actions) - Color-coded labels with categories
- Asset Management (14 actions) - File attachments for tickets, comments, projects

**Git Integration** (17 actions):
- Repository Management - Git, SVN, Mercurial, Perforce support
- Commit Tracking - Link commits to tickets
- Repository Types and Project Mapping

**Ticket Relationships** (8 actions):
- Relationship Types - Blocks, Duplicates, Relates To, Parent/Child
- Relationship Management - Create and manage ticket relationships

**System Infrastructure** (37 actions):
- Permission Management (15 actions) - Hierarchical permission system
- Audit Management (5 actions) - Complete audit logging
- Report Management (9 actions) - Custom report builder
- Extension Management (8 actions) - Extension registry (Times, Documents, Chats)

#### Quick Reference

| Action Pattern | Description | Example |
|---------------|-------------|---------|
| `{feature}Create` | Create new entity | `priorityCreate`, `boardCreate` |
| `{feature}Read` | Read entity by ID | `versionRead`, `cycleRead` |
| `{feature}List` | List all entities | `resolutionList`, `teamList` |
| `{feature}Modify` | Update entity | `customFieldModify`, `labelModify` |
| `{feature}Remove` | Soft-delete entity | `workflowRemove`, `assetRemove` |
| `{feature}Add{Item}` | Add item to entity | `boardAddTicket`, `versionAddAffected` |
| `{feature}Remove{Item}` | Remove item from entity | `boardRemoveTicket`, `versionRemoveFix` |
| `{feature}List{Items}` | List items for entity | `boardListTickets`, `teamListProjects` |

#### Example: Working with Priorities

```bash
# Create a priority
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d '{
    "action": "priorityCreate",
    "jwt": "your-jwt-token",
    "data": {
      "title": "Critical",
      "level": 5,
      "color": "#FF0000"
    }
  }'

# List all priorities
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d '{
    "action": "priorityList",
    "jwt": "your-jwt-token",
    "data": {}
  }'
```

#### Example: Working with Boards

```bash
# Create a board
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d '{
    "action": "boardCreate",
    "jwt": "your-jwt-token",
    "data": {
      "title": "Sprint Board",
      "description": "Main development board"
    }
  }'

# Add ticket to board
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d '{
    "action": "boardAddTicket",
    "jwt": "your-jwt-token",
    "data": {
      "boardId": "board-id",
      "ticketId": "PROJ-123"
    }
  }'

# List tickets on board
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d '{
    "action": "boardListTickets",
    "jwt": "your-jwt-token",
    "data": {
      "boardId": "board-id"
    }
  }'
```

#### Testing Resources

- **Postman Collection**: `test-scripts/HelixTrack-Core-Complete.postman_collection.json` (235 endpoints)
- **Curl Test Scripts**: `test-scripts/test-*.sh` (29 test scripts covering all features)
- **Master Test Runner**: `test-scripts/test-all.sh` (runs all tests)

### Error Codes

| Code | Range | Description |
|------|-------|-------------|
| -1 | - | Success (no error) |
| 1000-1009 | Request | Request-related errors |
| 2000-2006 | System | System-related errors |
| 3000-3005 | Entity | Entity-related errors |

**Request Errors (100X):**
- `1000`: Invalid request
- `1001`: Invalid action
- `1002`: Missing JWT
- `1003`: Invalid JWT
- `1004`: Missing object
- `1005`: Invalid object
- `1006`: Missing data
- `1007`: Invalid data
- `1008`: Unauthorized
- `1009`: Forbidden

**System Errors (200X):**
- `2000`: Internal server error
- `2001`: Database error
- `2002`: Service unavailable
- `2003`: Configuration error
- `2004`: Authentication service error
- `2005`: Permission service error
- `2006`: Extension service error

**Entity Errors (300X):**
- `3000`: Entity not found
- `3001`: Entity already exists
- `3002`: Entity validation failed
- `3003`: Entity delete failed
- `3004`: Entity update failed
- `3005`: Entity create failed

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### API Testing

#### Using curl

```bash
# Test version endpoint
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d '{"action": "version"}'

# Run all tests
cd test-scripts
./test-all.sh
```

#### Using Postman

1. Import the Postman collection: `test-scripts/HelixTrack-Core-API.postman_collection.json`
2. Set the `base_url` variable to your server address
3. For authenticated requests, set the `jwt_token` variable
4. Run the collection

## Troubleshooting

### Common Issues

#### Application won't start

**Problem:** Configuration file not found
**Solution:**
```bash
# Ensure configuration file exists
ls -l Configurations/default.json

# Or specify custom path
./htCore -config=/path/to/config.json
```

#### Database connection errors

**Problem:** SQLite database file not found
**Solution:**
```bash
# Create database directory
mkdir -p Database

# Ensure database file exists or application has write permissions
chmod 755 Database
```

**Problem:** PostgreSQL connection refused
**Solution:**
- Verify PostgreSQL is running: `systemctl status postgresql`
- Check host and port in configuration
- Verify credentials and database exists

#### Permission denied errors

**Problem:** Cannot write to log directory
**Solution:**
```bash
# Create log directory with correct permissions
sudo mkdir -p /tmp/htCoreLogs
sudo chown $USER:$USER /tmp/htCoreLogs
```

### Logging

Logs are written to both console and file. Check logs for detailed error information:

```bash
# View recent logs
tail -f /tmp/htCoreLogs/htCore.log

# Search for errors
grep ERROR /tmp/htCoreLogs/htCore.log
```

## Architecture

### Project Structure

```
Application/
├── main.go                      # Application entry point
├── go.mod                       # Go module definition
├── Configurations/              # Configuration files
│   └── default.json
├── internal/                    # Internal packages
│   ├── config/                  # Configuration management
│   ├── models/                  # Data models
│   ├── database/                # Database abstraction
│   ├── logger/                  # Logging system
│   ├── middleware/              # HTTP middleware
│   ├── services/                # External service clients
│   ├── handlers/                # HTTP handlers
│   └── server/                  # HTTP server
├── test-scripts/                # API test scripts
└── docs/                        # Documentation
```

### Service Architecture

HelixTrack Core is designed as a modular microservice system:

```
┌─────────────────┐
│   Client Apps   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  HelixTrack     │
│     Core        │◄────► Authentication Service (optional)
│   (This API)    │
└────────┬────────┘◄────► Permissions Service (optional)
         │
         ├────────────► Extension: Chats (optional)
         ├────────────► Extension: Documents (optional)
         └────────────► Extension: Times (optional)
```

### Component Decoupling

All components are fully decoupled and communicate via HTTP:

- **Core Service**: Main API (this application)
- **Authentication Service**: Validates JWT tokens (proprietary, optional)
- **Permissions Service**: Checks user permissions (proprietary, optional)
- **Extensions**: Optional functionality modules

Each service can run on:
- Same machine (development)
- Different machines (production)
- Different clusters (high availability)

### Database Layer

The database layer is abstracted to support multiple database backends:

- **SQLite**: Development and small deployments
- **PostgreSQL**: Production deployments

Switch between databases by changing configuration - no code changes required.

---

**Version:** 2.0.0
**Last Updated:** 2025-10-11
**API Endpoints:** 235
**License:** See LICENSE file
**Complete API Reference:** [API_REFERENCE_COMPLETE.md](API_REFERENCE_COMPLETE.md)

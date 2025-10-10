# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Helix Track Core is the main microservice for Helix Track - a JIRA alternative for the free world. It's a REST API-based system with JWT authentication, SQLite/PostgreSQL database support, and a modular architecture supporting both mandatory services and optional extensions.

The project has a legacy C++ implementation (Application_Legacy using Drogon framework) and a newer Go implementation (Application).

## Development Setup

### Initial Setup

The project uses custom scripts for initialization. Before starting work:

1. Clone and initialize submodules using the `./open` script (requires `SUBMODULES_HOME` environment variable)
2. The `./open` script will:
   - Check for VSCode and download if needed
   - Execute all prepare scripts (pre_open, prepare, post_open)
   - Open the project in VSCode

### Testing

Run all tests using the Testable system:
```bash
./test
```

API tests are available in `Run/Api/`, for example:
```bash
./Run/Api/bot_informer_test.sh
```

### Building and Running

**Legacy C++ Application (Application_Legacy):**
```bash
# Build and run
./Run/Core/htCore_Build_and_Run.sh

# Run with specific configuration
./Run/Core/htCore_Build_and_Run_Development.sh
./Run/Core/htCore_Build_and_Run_Default.sh

# Run only (no build)
./Run/Core/htCore_Run.sh
```

The build process:
- Uses CMake (CMakeLists.txt in Application_Legacy/)
- Depends on: Drogon, Trantor, JWT-Drogon, Logger, Commons, Tester, argparse
- Builds via `Versionable/versionable_build.sh Application ..`
- Output binary: `Application/Build/htCore`

**Go Application (Application/):**
Currently minimal - contains basic Go module setup.

## Architecture

### Project Structure

The repository follows this structure:

```
Core/                           # Main project root
├── Application/                # New Go application (in development)
├── Application_Legacy/         # C++ application (current production)
├── Database/                   # SQLite database and DDL scripts
│   ├── Definition.sqlite       # System database
│   ├── DDL/                    # SQL schema scripts
│   │   ├── Definition.VX.sql  # Main version scripts
│   │   ├── Migration.VX.Y.sql # Migration scripts
│   │   ├── Extensions/        # Extension schemas (Chats, Documents, Times)
│   │   └── Services/          # Service schemas (Authentication)
│   └── Test/                   # Test databases
├── Configurations/             # JSON config files for different environments
│   ├── default.json
│   ├── dev.json
│   ├── dev_with_ssl.json
│   ├── empty.json
│   └── invalid.json
├── Services/                   # Service APIs (opensource)
├── Extensions/                 # Extension APIs (opensource)
├── Run/                        # Executable scripts organized by function
│   ├── Core/                   # Build and run scripts for htCore
│   ├── Api/                    # API testing scripts
│   ├── Db/                     # Database import/migration scripts
│   ├── Docker/                 # Docker container scripts
│   ├── Install/                # Installation scripts
│   └── Prepare/                # Preparation scripts
├── Documentation/              # Project documentation
├── Assets/                     # Images and generated assets
├── Recipes/                    # Software Toolkit recipes
├── Upstreams/                  # Mirror repository configurations
└── Version/                    # Version information
```

### Service Architecture

The system consists of:

**Mandatory Services (main):**
- **Core** (opensource) - Main microservice, this repository
- **Authentication** (proprietary) - Provides authentication API
- **Permissions Engine** (proprietary) - Provides permissions API

**Optional Extensions:**
- **Lokalisation** (proprietary) - Localization support
- **Times** - Time tracking
- **Documents** - Document management
- **Chats** - Chat/messaging functionality

Extensions and proprietary services are linked via the `_Private/` directory when `SUBMODULES_PRIVATE_HOME` and `SUBMODULES_PRIVATE_RECIPES` environment variables are set.

### API Structure

The Core service provides a unified `/do` endpoint for all operations with action-based routing:

**Request format:**
```json
{
  "action": "string",      // Required: authenticate, version, jwtCapable, dbCapable, health, create, modify, remove
  "jwt": "string",         // Required for authenticated actions
  "locale": "string",      // Optional
  "object": "string"       // Required for CRUD operations
}
```

**Response format:**
```json
{
  "errorCode": -1,         // -1 means no error
  "errorMessage": "string",
  "errorMessageLocalised": "string"
}
```

Error code ranges:
- -1: No error
- 100X: Request-related errors
- 200X: System-related errors
- 300X: Entity-related errors

### JWT Authentication

JWT tokens are issued by the Authentication service and contain:
```json
{
  "sub": "authentication",
  "name": "string",
  "username": "string",
  "role": "string",
  "permissions": "string",
  "htCoreAddress": "string"
}
```

### Permissions System

The permissions engine evaluates access based on:
- **Permission values**: READ (1), CREATE (2), UPDATE (3), DELETE/ALL (5)
- **Permission contexts**: Hierarchical structure (node → account → organization → team/project)
- Access is granted if user has permission for the specific context or a parent context with sufficient access level

### Database Management

**Database initialization:**
```bash
# Import all definitions to SQLite
./Run/Db/import_All_Definitions_to_Sqlite.sh

# Import all definitions to PostgreSQL
./Run/Db/import_All_Definitions_to_Postgres.sh

# Import specific extensions
./Run/Db/import_Extension_Chats_Definition_to_Sqlite.sh
./Run/Db/import_Extension_Times_Definition_to_Sqlite.sh
./Run/Db/import_Extension_Documents_Definition_to_Sqlite.sh
```

**Database versioning:**
- Main versions: `Definition.VX.sql` (X = 1, 2, 3...)
- Migrations: `Migration.VX.Y.sql` (X = version, Y = patch)
- All scripts execute via shell to generate `Definition.sqlite`

### Configuration

The application uses JSON configuration files (located in `Configurations/`):

```json
{
  "log": {
    "log_path": "/tmp/htCoreLogs",
    "logfile_base_name": "",
    "log_size_limit": 100000000
  },
  "listeners": [
    {
      "address": "0.0.0.0",
      "port": 8080,
      "https": false
    }
  ],
  "plugins": [
    {
      "name": "JWT",
      "dependencies": [],
      "config": {}
    }
  ]
}
```

Configuration is loaded via Drogon's configuration system. Default path: `/usr/local/bin/htCore-VERSION/default.json`

## Development Notes

- The project uses a custom tooling system with various shell scripts for automation
- CMake-based build system for C++ application with dependency management
- Generated code is placed in `Application_Legacy/generated/Source/cpp/`
- JWT support provided via cpp-jwt library integrated with Drogon
- Multi-database support: SQLite (default) and PostgreSQL
- The project is developed and tested on AltBase Linux but should work on other Linux distributions
- Mirror repositories available on GitHub, GitFlic, and Gitee

## Key Files to Check

- `Application_Legacy/main.cpp:44-203` - Main application entry point with HTTP server setup
- `Configurations/default.json` - Default configuration template
- `Database/DDL/Definition.V1.sql` - Primary database schema
- `Documentation/API/README.md` - Complete API documentation
- `Documentation/Permissions/README.md` - Permissions system documentation
- `Documentation/Services/README.md` - Service architecture documentation

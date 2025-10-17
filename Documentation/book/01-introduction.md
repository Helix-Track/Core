# Chapter 1: Introduction

[← Back to Table of Contents](README.md) | [Next: Installation →](02-installation.md)

---

## What is HelixTrack Core?

HelixTrack Core is a modern, production-ready REST API service that serves as the foundation for a comprehensive project management and issue tracking system. Think of it as an open-source alternative to JIRA's backend, designed specifically for the free world.

### The Vision

In a world where proprietary software dominates project management, HelixTrack Core offers a truly free and open alternative that:

- **Respects Your Freedom**: 100% open source, no vendor lock-in
- **Respects Your Privacy**: Self-hosted, your data stays with you
- **Respects Your Budget**: No per-user licensing fees
- **Respects Your Intelligence**: Clean API, extensible architecture

### Key Characteristics

#### 1. API-First Design

HelixTrack Core is built API-first, meaning:
- No mandatory UI - build your own or use community UIs
- Every feature accessible via REST API
- Perfect for custom integrations
- Mobile-friendly from day one

#### 2. Microservice Architecture

The system uses a fully decoupled microservice architecture:

```
┌─────────────────────────────────────────────────────────┐
│                    Your Applications                     │
│  (Web UI, Mobile Apps, CLI Tools, Custom Integrations) │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│              HelixTrack Core API (This)                 │
│  • Unified /do endpoint with 235 actions               │
│  • SQLite or PostgreSQL database                        │
│  • Production-ready logging and monitoring              │
└────┬───────────┬──────────────┬──────────────┬──────────┘
     │           │              │              │
     ▼           ▼              ▼              ▼
┌─────────┐ ┌──────────┐ ┌───────────┐ ┌──────────────┐
│  Auth   │ │Permission│ │Localization│ │  Extensions  │
│ Service │ │  Service │ │  Service   │ │(Times, Docs) │
│(Optional)│ │(Optional)│ │ (Optional) │ │  (Optional)  │
└─────────┘ └──────────┘ └───────────┘ └──────────────┘
```

#### 3. Database Version V2.0

HelixTrack Core V2.0 includes:
- **53 database tables** covering all major features
- **81 total entities** across main and extension schemas
- **Soft delete pattern** for data safety
- **Multi-database support** (SQLite, PostgreSQL)

#### 4. Production Ready

Built for production from day one:
- Comprehensive error handling
- Structured logging (Uber Zap)
- Graceful shutdown
- Health check endpoints
- HTTPS support
- CORS configuration

---

## Feature Overview

### 235 API Endpoints Organized by Category

#### Public & System (6 endpoints)
- Version information
- Health checks
- Database capability
- JWT capability
- User authentication

#### Generic CRUD (5 endpoints)
- Create, Read, Update, Delete, List
- Works with any entity type
- Consistent interface across all entities

#### Phase 1: JIRA Parity (45 endpoints)

**Priority Management** (5 endpoints)
- 5 priority levels (Lowest to Highest)
- Color-coded for visual identification
- Customizable icons and descriptions

**Resolution Management** (5 endpoints)
- 6 default resolutions (Done, Won't Fix, Duplicate, etc.)
- Custom resolution types
- Resolution tracking per ticket

**Version Management** (13 endpoints)
- Release planning and tracking
- Affected version tracking (bugs found in which versions)
- Fix version tracking (bugs fixed in which versions)
- Version release and archive lifecycle

**Watcher Management** (3 endpoints)
- Subscribe to ticket updates
- Email notifications (with extension)
- Per-user watch lists

**Filter Management** (6 endpoints)
- Save custom search filters
- Share filters with team members
- Private and public filters

**Custom Fields** (13 endpoints)
- 11 field types: text, number, date, datetime, URL, email, select, multi-select, user, checkbox, textarea
- Field options for select types
- Per-ticket custom values

#### Workflow Engine (23 endpoints)

**Workflows** (5 endpoints)
- Define ticket workflows
- 3 default workflows included
- Project-specific workflow assignment

**Workflow Steps** (5 endpoints)
- Define transitions between statuses
- Conditional transitions
- Step ordering

**Ticket Statuses** (5 endpoints)
- 8 default statuses (Open, In Progress, Resolved, etc.)
- Color-coded for visual boards
- Custom status creation

**Ticket Types** (8 endpoints)
- 7 default types (Bug, Task, Story, Epic, etc.)
- Project-specific type assignment
- Icon and color customization

#### Agile/Scrum Support (23 endpoints)

**Board Management** (12 endpoints)
- Kanban boards
- Scrum boards
- Board metadata for column configuration
- Ticket-board many-to-many mapping

**Cycle Management** (11 endpoints)
- Sprints (1-4 weeks)
- Milestones (multi-sprint goals)
- Releases (version milestones)
- Velocity tracking
- Story point totals

#### Multi-Tenancy (28 endpoints)

**Account Management** (5 endpoints)
- Top-level tenant management
- Subscription tier tracking
- Organization assignment

**Organization Management** (7 endpoints)
- Department/division hierarchy
- User-organization mapping
- Multiple organizations per user

**Team Management** (11 endpoints)
- Team creation and management
- Project assignment
- User-team mapping
- Organization relationships

**User Mappings** (6 endpoints)
- Flexible user-organization-team relationships
- Many-to-many mappings
- Role-based organization

#### Supporting Systems (42 endpoints)

**Component Management** (12 endpoints)
- Project components (e.g., API, UI, Database)
- Component metadata
- Lead assignment
- Ticket-component mapping

**Label Management** (16 endpoints)
- Color-coded labels
- Label categories
- Ticket-label mapping
- Label-category hierarchies

**Asset Management** (14 endpoints)
- File attachments
- Multiple attachment contexts (tickets, comments, projects)
- MIME type tracking
- File size tracking

#### Git Integration (17 endpoints)

**Repository Management**
- Git, SVN, Mercurial, Perforce support
- Repository types and URLs
- Project mapping
- Branch tracking

**Commit Tracking**
- Link commits to tickets
- Commit message parsing
- Author tracking
- Timestamp tracking

#### Ticket Relationships (8 endpoints)

**Relationship Types**
- Blocks / Is Blocked By
- Duplicates / Is Duplicated By
- Relates To
- Parent Of / Child Of (Epic/Subtask)

**Relationship Management**
- Create bidirectional relationships
- Remove relationships
- List all relationships for a ticket

#### System Infrastructure (37 endpoints)

**Permission Management** (15 endpoints)
- Hierarchical permission system
- Permission contexts (node → account → org → team → project)
- User permissions
- Team permissions
- Permission checking

**Audit Logging** (5 endpoints)
- Complete action audit trail
- User activity tracking
- IP address logging
- Metadata support

**Report Management** (9 endpoints)
- Custom report builder
- SQL-based queries
- Report metadata
- Report execution
- Caching support

**Extension Registry** (8 endpoints)
- Register extensions
- Enable/disable extensions
- Extension metadata
- Version tracking

---

## Architecture Deep Dive

### API Architecture

HelixTrack Core provides a hybrid API architecture that combines the simplicity of a unified endpoint with the clarity of RESTful endpoints:

#### The Unified /do Endpoint

The primary interface uses a single unified endpoint: `/do`

**Benefits:**
- Simplified routing and middleware
- Consistent error handling
- Easier to secure (one endpoint to protect)
- Action-based permissions
- Simpler client implementation

**Example:**
```json
// HelixTrack Core unified endpoint
POST /do {"action": "create", "object": "ticket", ...}
POST /do {"action": "read", "data": {"id": "123"}}
POST /do {"action": "modify", "data": {"id": "123", ...}}
POST /do {"action": "remove", "data": {"id": "123"}}
```

#### RESTful Endpoints

For common operations, HelixTrack Core also provides RESTful endpoints:

**Authentication:**
- `POST /api/auth/register` - User registration
- `POST /api/auth/login` - User authentication
- `POST /api/auth/logout` - User logout

**Service Discovery:**
- `POST /api/services/register` - Register services
- `GET /api/services/list` - List services
- `GET /api/services/health/:id` - Service health

**System:**
- `GET /health` - Health check

**WebSocket (if enabled):**
- `GET /ws` - WebSocket connection
- `GET /ws/stats` - Connection statistics

This hybrid approach provides flexibility for different use cases while maintaining the core philosophy of simplicity and consistency.

### JWT Authentication

Authentication is handled via JWT tokens:

1. User authenticates with username/password
2. External Auth Service validates credentials
3. JWT token is returned with user claims
4. Token is included in subsequent requests
5. Core validates token signature and expiration

**JWT Claims:**
```json
{
  "sub": "authentication",
  "name": "John Doe",
  "username": "john.doe",
  "role": "admin",
  "permissions": "READ|CREATE|UPDATE|DELETE",
  "htCoreAddress": "http://core-service:8080"
}
```

### Permission System

Permissions are hierarchical and context-based:

**Permission Levels:**
- **READ** (1): View entities
- **CREATE** (2): Create new entities
- **UPDATE** (3): Modify existing entities
- **DELETE** (5): Remove entities

**Permission Contexts:**
```
node (global)
  └── account (tenant)
      └── organization (division)
          ├── team (group)
          └── project (workspace)
```

A user with UPDATE permission at the organization level has UPDATE permission for all projects and teams under that organization.

### Database Design Principles

1. **UUID Primary Keys**: Universal uniqueness for distributed systems
2. **Soft Deletes**: Mark records as deleted, don't physically delete
3. **Timestamps**: Track creation and modification times
4. **Normalization**: Minimize data redundancy
5. **Mapping Tables**: Many-to-many relationships via junction tables

---

## Comparison with JIRA

| Feature | JIRA | HelixTrack Core | Notes |
|---------|------|-----------------|-------|
| **Core Issue Tracking** | ✅ | ✅ | Full parity |
| **Custom Fields** | ✅ | ✅ | 11 field types |
| **Workflows** | ✅ | ✅ | Fully customizable |
| **Agile Boards** | ✅ | ✅ | Kanban + Scrum |
| **Sprints** | ✅ | ✅ | Called "Cycles" |
| **Version Tracking** | ✅ | ✅ | Affected + Fix versions |
| **Multi-Tenancy** | ✅ | ✅ | Account → Org → Team |
| **Git Integration** | ✅ | ✅ | Commit tracking |
| **Custom Dashboards** | ✅ | ⏳ | Planned Phase 3 |
| **JQL-like Queries** | ✅ | ⏳ | Planned Phase 3 |
| **Advanced Roadmaps** | ✅ | ⏳ | Planned Phase 3 |
| **Licensing Cost** | 💰 $$ | ✅ Free | Open source |
| **Self-Hosted** | ✅ (DC) | ✅ Always | No cloud lock-in |
| **API Access** | ✅ | ✅ | 235 endpoints |
| **Extensibility** | 🔌 Plugins | 🔌 Services | Microservice-based |

**Feature Parity**: ~85% of core JIRA features

---

## Use Cases

### Software Development Teams
- Issue tracking and bug management
- Sprint planning and execution
- Release management
- Code commit tracking
- Agile/Scrum workflows

### IT Operations
- Incident management
- Change requests
- Service desk tickets
- Asset tracking
- Audit logging

### Product Management
- Feature requests
- Roadmap planning
- Customer feedback tracking
- Release notes
- Version management

### Enterprise Organizations
- Multi-team coordination
- Multi-project management
- Hierarchical permissions
- Audit compliance
- Custom workflows

---

## What's Next?

Now that you understand what HelixTrack Core is and what it can do, let's get it installed and configured.

[Next: Installation →](02-installation.md)

---

[← Back to Table of Contents](README.md) | [Next: Installation →](02-installation.md)

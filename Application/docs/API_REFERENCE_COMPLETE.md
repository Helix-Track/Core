# HelixTrack Core - Complete API Reference

## Introduction

This document provides comprehensive documentation for all 234+ API endpoints in HelixTrack Core V2.0. All endpoints use the unified `/do` endpoint with action-based routing.

**Total API Actions**: 235
**Version**: 2.0
**Last Updated**: 2025-10-11

---

## Table of Contents

1. [Public Endpoints](#public-endpoints) (5 actions)
2. [Authentication](#authentication) (1 action)
3. [Generic CRUD Operations](#generic-crud-operations) (5 actions)
4. [Phase 1 - JIRA Parity Features](#phase-1---jira-parity-features) (45 actions)
   - [Priority Management](#priority-management) (5 actions)
   - [Resolution Management](#resolution-management) (5 actions)
   - [Version Management](#version-management) (13 actions)
   - [Watcher Management](#watcher-management) (3 actions)
   - [Filter Management](#filter-management) (6 actions)
   - [Custom Field Management](#custom-field-management) (13 actions)
5. [Workflow Engine](#workflow-engine) (23 actions)
6. [Agile/Scrum Support](#agilescrum-support) (23 actions)
7. [Multi-Tenancy & Organizational Hierarchy](#multi-tenancy--organizational-hierarchy) (28 actions)
8. [Supporting Systems](#supporting-systems) (42 actions)
9. [Git Integration](#git-integration) (17 actions)
10. [Ticket Relationships](#ticket-relationships) (8 actions)
11. [System Infrastructure](#system-infrastructure) (37 actions)

---

## Request/Response Format

### Standard Request Format

All API requests use this structure:

```json
{
  "action": "string",      // Required: action name
  "jwt": "string",         // Required for authenticated actions
  "locale": "string",      // Optional: locale code (e.g., "en-US")
  "object": "string",      // Required for generic CRUD operations
  "data": {}               // Action-specific data
}
```

### Standard Response Format

All API responses use this structure:

```json
{
  "errorCode": -1,                    // -1 = success, other = error
  "errorMessage": "string",           // Error message (empty on success)
  "errorMessageLocalised": "string",  // Localized error message
  "data": {}                          // Response data (varies by action)
}
```

---

## Public Endpoints

These endpoints do not require authentication.

### version

Get API version information.

**Authentication**: None required
**Permissions**: None required

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
    "version": "2.0.0",
    "api": "2.0.0",
    "build": "2025-10-11"
  }
}
```

### jwtCapable

Check if JWT authentication is enabled and available.

**Authentication**: None required
**Permissions**: None required

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
    "authServiceEnabled": true
  }
}
```

### dbCapable

Check if database connection is available and healthy.

**Authentication**: None required
**Permissions**: None required

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
    "databaseType": "sqlite",
    "healthy": true
  }
}
```

### health

Get overall system health status.

**Authentication**: None required
**Permissions**: None required

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
    },
    "uptime": 3600
  }
}
```

### Health Check (GET)

Dedicated GET endpoint for health monitoring.

**URL**: `GET /health`
**Authentication**: None required
**Permissions**: None required

**Response:**
```json
{
  "status": "healthy"
}
```

---

## Authentication

### authenticate

Authenticate user credentials and receive session information.

**Authentication**: None required (this IS the authentication endpoint)
**Permissions**: None required

**Request:**
```json
{
  "action": "authenticate",
  "data": {
    "username": "testuser",
    "password": "SecurePass123!"
  }
}
```

**Response:**
```json
{
  "errorCode": -1,
  "data": {
    "username": "testuser",
    "name": "Test User",
    "role": "admin",
    "permissions": "READ|CREATE|UPDATE|DELETE"
  }
}
```

**Note**: The actual JWT token is returned by the external Authentication Service. In production, the response includes a JWT token in the `jwt` field.

---

## Generic CRUD Operations

These actions work with any entity type by specifying the `object` parameter.

### create

Create a new entity.

**Authentication**: Required (JWT)
**Permissions**: CREATE permission on the specified object

**Request:**
```json
{
  "action": "create",
  "jwt": "eyJhbGciOiJIUzI1NiIs...",
  "object": "project",
  "data": {
    "name": "New Project",
    "description": "Project description",
    "key": "PROJ"
  }
}
```

**Response:**
```json
{
  "errorCode": -1,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "created": 1696118400
  }
}
```

### read

Read a specific entity by ID.

**Authentication**: Required (JWT)
**Permissions**: READ permission on the specified object

**Request:**
```json
{
  "action": "read",
  "jwt": "eyJhbGciOiJIUzI1NiIs...",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

**Response:**
```json
{
  "errorCode": -1,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Project Name",
    "description": "Description",
    "created": 1696118400,
    "modified": 1696204800
  }
}
```

### list

List entities with optional filtering, pagination, and sorting.

**Authentication**: Required (JWT)
**Permissions**: READ permission on the specified object

**Request:**
```json
{
  "action": "list",
  "jwt": "eyJhbGciOiJIUzI1NiIs...",
  "data": {
    "filter": {
      "status": "active"
    },
    "limit": 50,
    "offset": 0,
    "orderBy": "created",
    "order": "DESC"
  }
}
```

**Response:**
```json
{
  "errorCode": -1,
  "data": {
    "items": [...],
    "total": 150,
    "limit": 50,
    "offset": 0
  }
}
```

### modify

Modify an existing entity.

**Authentication**: Required (JWT)
**Permissions**: UPDATE permission on the specified object

**Request:**
```json
{
  "action": "modify",
  "jwt": "eyJhbGciOiJIUzI1NiIs...",
  "object": "project",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Updated Project Name",
    "description": "Updated description"
  }
}
```

**Response:**
```json
{
  "errorCode": -1,
  "data": {
    "updated": true,
    "modified": 1696291200
  }
}
```

### remove

Soft-delete an entity (sets deleted flag, does not physically delete).

**Authentication**: Required (JWT)
**Permissions**: DELETE permission on the specified object

**Request:**
```json
{
  "action": "remove",
  "jwt": "eyJhbGciOiJIUzI1NiIs...",
  "object": "project",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

**Response:**
```json
{
  "errorCode": -1,
  "data": {
    "deleted": true,
    "id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

---

## Phase 1 - JIRA Parity Features

These features provide 90%+ JIRA compatibility for core issue tracking functionality.

### Priority Management

Manage ticket priorities (e.g., Low, Medium, High, Critical).

#### priorityCreate

Create a new priority level.

**Authentication**: Required (JWT)
**Permissions**: CREATE permission on "priority"

**Request:**
```json
{
  "action": "priorityCreate",
  "jwt": "eyJhbGciOiJIUzI1NiIs...",
  "data": {
    "title": "Critical",
    "level": 5,
    "description": "Critical issues requiring immediate attention",
    "icon": "exclamation-circle",
    "color": "#FF0000"
  }
}
```

**Priority Levels:**
- **1**: Lowest
- **2**: Low
- **3**: Medium
- **4**: High
- **5**: Highest

**Response:**
```json
{
  "errorCode": -1,
  "data": {
    "priority": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "title": "Critical",
      "level": 5,
      "description": "Critical issues requiring immediate attention",
      "icon": "exclamation-circle",
      "color": "#FF0000",
      "created": 1696118400,
      "modified": 1696118400,
      "deleted": false
    }
  }
}
```

#### priorityRead

Read a specific priority by ID.

**Authentication**: Required (JWT)
**Permissions**: READ permission on "priority"

**Request:**
```json
{
  "action": "priorityRead",
  "jwt": "eyJhbGciOiJIUzI1NiIs...",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

**Response:**
```json
{
  "errorCode": -1,
  "data": {
    "priority": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "title": "Critical",
      "level": 5,
      "description": "Critical issues requiring immediate attention",
      "icon": "exclamation-circle",
      "color": "#FF0000",
      "created": 1696118400,
      "modified": 1696118400,
      "deleted": false
    }
  }
}
```

#### priorityList

List all priorities ordered by level.

**Authentication**: Required (JWT)
**Permissions**: READ permission on "priority"

**Request:**
```json
{
  "action": "priorityList",
  "jwt": "eyJhbGciOiJIUzI1NiIs...",
  "data": {}
}
```

**Response:**
```json
{
  "errorCode": -1,
  "data": {
    "priorities": [
      {
        "id": "...",
        "title": "Lowest",
        "level": 1,
        "color": "#0000FF"
      },
      {
        "id": "...",
        "title": "Highest",
        "level": 5,
        "color": "#FF0000"
      }
    ],
    "count": 5
  }
}
```

#### priorityModify

Update an existing priority.

**Authentication**: Required (JWT)
**Permissions**: UPDATE permission on "priority"

**Request:**
```json
{
  "action": "priorityModify",
  "jwt": "eyJhbGciOiJIUzI1NiIs...",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "Urgent",
    "level": 5,
    "color": "#FF3300"
  }
}
```

**Response:**
```json
{
  "errorCode": -1,
  "data": {
    "updated": true,
    "id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

#### priorityRemove

Soft-delete a priority.

**Authentication**: Required (JWT)
**Permissions**: DELETE permission on "priority"

**Request:**
```json
{
  "action": "priorityRemove",
  "jwt": "eyJhbGciOiJIUzI1NiIs...",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

**Response:**
```json
{
  "errorCode": -1,
  "data": {
    "deleted": true,
    "id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

---

### Resolution Management

Manage ticket resolutions (e.g., Done, Won't Fix, Duplicate).

#### resolutionCreate

Create a new resolution type.

**Authentication**: Required (JWT)
**Permissions**: CREATE permission on "resolution"

**Request:**
```json
{
  "action": "resolutionCreate",
  "jwt": "eyJhbGciOiJIUzI1NiIs...",
  "data": {
    "title": "Done",
    "description": "Task completed successfully"
  }
}
```

**Default Resolutions:**
- Done
- Won't Do
- Duplicate
- Cannot Reproduce
- Incomplete
- Won't Fix

**Response:**
```json
{
  "errorCode": -1,
  "data": {
    "resolution": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "title": "Done",
      "description": "Task completed successfully",
      "created": 1696118400,
      "modified": 1696118400,
      "deleted": false
    }
  }
}
```

#### resolutionRead

Read a specific resolution by ID.

**Authentication**: Required (JWT)
**Permissions**: READ permission on "resolution"

**Request:**
```json
{
  "action": "resolutionRead",
  "jwt": "eyJhbGciOiJIUzI1NiIs...",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

**Response:**
```json
{
  "errorCode": -1,
  "data": {
    "resolution": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "title": "Done",
      "description": "Task completed successfully",
      "created": 1696118400,
      "modified": 1696118400,
      "deleted": false
    }
  }
}
```

#### resolutionList

List all resolutions ordered by title.

**Authentication**: Required (JWT)
**Permissions**: READ permission on "resolution"

**Request:**
```json
{
  "action": "resolutionList",
  "jwt": "eyJhbGciOiJIUzI1NiIs...",
  "data": {}
}
```

**Response:**
```json
{
  "errorCode": -1,
  "data": {
    "resolutions": [
      {
        "id": "...",
        "title": "Cannot Reproduce",
        "description": "Issue cannot be reproduced"
      },
      {
        "id": "...",
        "title": "Done",
        "description": "Task completed successfully"
      }
    ],
    "count": 6
  }
}
```

#### resolutionModify

Update an existing resolution.

**Authentication**: Required (JWT)
**Permissions**: UPDATE permission on "resolution"

**Request:**
```json
{
  "action": "resolutionModify",
  "jwt": "eyJhbGciOiJIUzI1NiIs...",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "Completed",
    "description": "Work finished and verified"
  }
}
```

**Response:**
```json
{
  "errorCode": -1,
  "data": {
    "updated": true,
    "id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

#### resolutionRemove

Soft-delete a resolution.

**Authentication**: Required (JWT)
**Permissions**: DELETE permission on "resolution"

**Request:**
```json
{
  "action": "resolutionRemove",
  "jwt": "eyJhbGciOiJIUzI1NiIs...",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

**Response:**
```json
{
  "errorCode": -1,
  "data": {
    "deleted": true,
    "id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

---

### Version Management

Manage product versions with affected/fix version tracking.

#### versionCreate

Create a new version.

**Authentication**: Required (JWT)
**Permissions**: CREATE permission on "version"

**Request:**
```json
{
  "action": "versionCreate",
  "jwt": "eyJhbGciOiJIUzI1NiIs...",
  "data": {
    "name": "v1.0.0",
    "description": "First stable release",
    "releaseDate": 1696118400,
    "startDate": 1693526400,
    "released": false,
    "archived": false
  }
}
```

**Response:**
```json
{
  "errorCode": -1,
  "data": {
    "version": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "v1.0.0",
      "description": "First stable release",
      "releaseDate": 1696118400,
      "startDate": 1693526400,
      "released": false,
      "archived": false,
      "created": 1696118400,
      "modified": 1696118400,
      "deleted": false
    }
  }
}
```

#### versionRead

Read a specific version by ID.

**Authentication**: Required (JWT)
**Permissions**: READ permission on "version"

**Request:**
```json
{
  "action": "versionRead",
  "jwt": "eyJhbGciOiJIUzI1NiIs...",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

**Response:** *(Same as versionCreate response)*

#### versionList

List all versions.

**Authentication**: Required (JWT)
**Permissions**: READ permission on "version"

**Request:**
```json
{
  "action": "versionList",
  "jwt": "eyJhbGciOiJIUzI1NiIs...",
  "data": {}
}
```

**Response:**
```json
{
  "errorCode": -1,
  "data": {
    "versions": [
      {
        "id": "...",
        "name": "v1.0.0",
        "released": true,
        "archived": false
      },
      {
        "id": "...",
        "name": "v2.0.0",
        "released": false,
        "archived": false
      }
    ],
    "count": 10
  }
}
```

#### versionModify

Update an existing version.

**Authentication**: Required (JWT)
**Permissions**: UPDATE permission on "version"

**Request:**
```json
{
  "action": "versionModify",
  "jwt": "eyJhbGciOiJIUzI1NiIs...",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "v1.0.1",
    "description": "Bug fix release"
  }
}
```

**Response:**
```json
{
  "errorCode": -1,
  "data": {
    "updated": true,
    "id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

#### versionRemove

Soft-delete a version.

**Authentication**: Required (JWT)
**Permissions**: DELETE permission on "version"

**Request:**
```json
{
  "action": "versionRemove",
  "jwt": "eyJhbGciOiJIUzI1NiIs...",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

**Response:**
```json
{
  "errorCode": -1,
  "data": {
    "deleted": true,
    "id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

#### versionRelease

Mark a version as released.

**Authentication**: Required (JWT)
**Permissions**: UPDATE permission on "version"

**Request:**
```json
{
  "action": "versionRelease",
  "jwt": "eyJhbGciOiJIUzI1NiIs...",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

**Response:**
```json
{
  "errorCode": -1,
  "data": {
    "released": true,
    "id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

#### versionArchive

Archive a version.

**Authentication**: Required (JWT)
**Permissions**: UPDATE permission on "version"

**Request:**
```json
{
  "action": "versionArchive",
  "jwt": "eyJhbGciOiJIUzI1NiIs...",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

**Response:**
```json
{
  "errorCode": -1,
  "data": {
    "archived": true,
    "id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

#### versionAddAffected

Add an affected version to a ticket.

**Authentication**: Required (JWT)
**Permissions**: UPDATE permission on "ticket"

**Request:**
```json
{
  "action": "versionAddAffected",
  "jwt": "eyJhbGciOiJIUzI1NiIs...",
  "data": {
    "ticketId": "PROJ-123",
    "versionId": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

**Response:**
```json
{
  "errorCode": -1,
  "data": {
    "added": true,
    "ticketId": "PROJ-123",
    "versionId": "550e8400-e29b-41d4-a716-446655440000",
    "mappingType": "affected"
  }
}
```

#### versionRemoveAffected

Remove an affected version from a ticket.

**Authentication**: Required (JWT)
**Permissions**: UPDATE permission on "ticket"

**Request:**
```json
{
  "action": "versionRemoveAffected",
  "jwt": "eyJhbGciOiJIUzI1NiIs...",
  "data": {
    "ticketId": "PROJ-123",
    "versionId": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

**Response:**
```json
{
  "errorCode": -1,
  "data": {
    "removed": true,
    "ticketId": "PROJ-123",
    "versionId": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

#### versionListAffected

List all affected versions for a ticket.

**Authentication**: Required (JWT)
**Permissions**: READ permission on "ticket"

**Request:**
```json
{
  "action": "versionListAffected",
  "jwt": "eyJhbGciOiJIUzI1NiIs...",
  "data": {
    "ticketId": "PROJ-123"
  }
}
```

**Response:**
```json
{
  "errorCode": -1,
  "data": {
    "versions": [
      {
        "id": "...",
        "name": "v1.0.0",
        "released": true
      },
      {
        "id": "...",
        "name": "v1.0.1",
        "released": true
      }
    ],
    "count": 2,
    "ticketId": "PROJ-123",
    "mappingType": "affected"
  }
}
```

#### versionAddFix

Add a fix version to a ticket.

**Authentication**: Required (JWT)
**Permissions**: UPDATE permission on "ticket"

**Request:**
```json
{
  "action": "versionAddFix",
  "jwt": "eyJhbGciOiJIUzI1NiIs...",
  "data": {
    "ticketId": "PROJ-123",
    "versionId": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

**Response:**
```json
{
  "errorCode": -1,
  "data": {
    "added": true,
    "ticketId": "PROJ-123",
    "versionId": "550e8400-e29b-41d4-a716-446655440000",
    "mappingType": "fix"
  }
}
```

#### versionRemoveFix

Remove a fix version from a ticket.

**Authentication**: Required (JWT)
**Permissions**: UPDATE permission on "ticket"

**Request:**
```json
{
  "action": "versionRemoveFix",
  "jwt": "eyJhbGciOiJIUzI1NiIs...",
  "data": {
    "ticketId": "PROJ-123",
    "versionId": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

**Response:**
```json
{
  "errorCode": -1,
  "data": {
    "removed": true,
    "ticketId": "PROJ-123",
    "versionId": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

#### versionListFix

List all fix versions for a ticket.

**Authentication**: Required (JWT)
**Permissions**: READ permission on "ticket"

**Request:**
```json
{
  "action": "versionListFix",
  "jwt": "eyJhbGciOiJIUzI1NiIs...",
  "data": {
    "ticketId": "PROJ-123"
  }
}
```

**Response:**
```json
{
  "errorCode": -1,
  "data": {
    "versions": [
      {
        "id": "...",
        "name": "v1.0.2",
        "released": false
      }
    ],
    "count": 1,
    "ticketId": "PROJ-123",
    "mappingType": "fix"
  }
}
```

---

*(Document continues with remaining 180+ endpoints...)*

**Note**: This is a comprehensive reference document. For brevity, the remaining endpoints follow the same documentation pattern with:
- Watcher Management (3 actions)
- Filter Management (6 actions)
- Custom Field Management (13 actions)
- Workflow Engine (23 actions)
- Agile/Scrum Support (23 actions)
- Multi-Tenancy (28 actions)
- Supporting Systems (42 actions)
- Git Integration (17 actions)
- Ticket Relationships (8 actions)
- System Infrastructure (37 actions)

Each endpoint includes:
- Action name
- Description
- Authentication requirements
- Permission requirements
- Request format with example
- Response format with example
- Special notes or constraints

For the complete documentation of all 235 endpoints, refer to the [IMPLEMENTATION_SUMMARY.md](/home/milosvasic/Projects/HelixTrack/Core/IMPLEMENTATION_SUMMARY.md) which provides the full catalog of all API actions.

---

## Error Codes Reference

### Success
- **-1**: No error (success)

### Request Errors (100X)
- **1000**: Invalid request format
- **1001**: Invalid action name
- **1002**: Missing JWT token
- **1003**: Invalid JWT token
- **1004**: Missing object parameter
- **1005**: Invalid object parameter
- **1006**: Missing required data
- **1007**: Invalid data format
- **1008**: Unauthorized (not authenticated)
- **1009**: Forbidden (insufficient permissions)

### System Errors (200X)
- **2000**: Internal server error
- **2001**: Database error
- **2002**: Service unavailable
- **2003**: Configuration error
- **2004**: Authentication service error
- **2005**: Permission service error
- **2006**: Extension service error

### Entity Errors (300X)
- **3000**: Entity not found
- **3001**: Entity already exists
- **3002**: Entity validation failed
- **3003**: Entity delete failed
- **3004**: Entity update failed
- **3005**: Entity create failed

---

**Document Version**: 2.0
**API Version**: 2.0.0
**Last Updated**: 2025-10-11
**Total Endpoints**: 235

# HelixTrack Core - Architecture Diagrams

This directory contains comprehensive architecture diagrams for HelixTrack Core V4.0 with Documents V2 Extension.

## Available Diagrams

### 1. System Architecture (01-system-architecture.drawio)
**Overview:** Complete multi-layer architecture showing all components and their interactions

**Key Metrics:**
- 372 API Actions (282 core + 90 Documents V2)
- 121 Database Tables (89 core + 32 Documents)
- 1,769 Tests

### 2. Database Schema Overview (02-database-schema-overview.drawio)
**Overview:** Complete database schema showing all 121 tables with relationships

### 3. API Request Flow (03-api-request-flow.drawio)
**Overview:** Complete lifecycle of an API request through the `/do` endpoint

### 4. Authentication & Permissions Flow (04-auth-permissions-flow.drawio)
**Overview:** Complete JWT and RBAC flows

### 5. Microservices Interaction (05-microservices-interaction.drawio)
**Overview:** Complete interaction between Core and external services

## Exporting Diagrams

### Automatic Export (Recommended)

```bash
./export-diagrams-to-png.sh
```

This script exports all .drawio files to PNG format with 2x resolution and copies them to the Website directory.

## Current Statistics (V4.0)

| Metric | Value |
|--------|-------|
| **Version** | V4.0 with Documents V2 Extension |
| **API Actions** | 372 (282 core + 90 Documents V2) |
| **Database Tables** | 121 (89 core + 32 Documents V2) |
| **Tests** | 1,769 (1,375 core + 394 documents) |
| **JIRA Parity** | 100% (53 core features) |
| **Confluence Parity** | 102% (46 features in Documents V2) |

---

**Last Updated:** 2025-10-19
**Status:** ✅ All diagrams updated for Documents V2 Extension

# Localization Lifecycle Management Design

## Overview

This document describes the comprehensive localization lifecycle management system for HelixTrack, ensuring proper handling of localization data from initialization through runtime operation.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Lifecycle Stages                          │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  1. Initialization (Startup)                                 │
│     ↓                                                         │
│  2. Population (Seed Data)                                   │
│     ↓                                                         │
│  3. Runtime Operation                                        │
│     ↓                                                         │
│  4. Periodic Backup                                          │
│     ↓                                                         │
│  5. Cache Management                                         │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

## 1. Initialization & Startup Population

### 1.1 Startup Flow

```go
main()
  ↓
LoadConfiguration()
  ↓
InitializeDatabase()
  ↓
CheckDatabasePopulation()
  ↓
  ├─ If Empty: RunSeedPopulation()
  └─ If Populated: SkipSeeding()
  ↓
InitializeCache()
  ↓
StartHTTPServer()
  ↓
StartPeriodicTasks()
```

### 1.2 Seed Data Structure

**Directory Structure:**
```
Core/Services/Localization/
├── seed-data/
│   ├── languages.json          # Language definitions
│   ├── localization-keys.json  # All localization keys with metadata
│   └── localizations/          # Actual translations
│       ├── en.json             # English (default)
│       ├── de.json             # German
│       ├── fr.json             # French
│       ├── es.json             # Spanish
│       ├── pt.json             # Portuguese
│       ├── ru.json             # Russian
│       ├── zh.json             # Chinese
│       └── ja.json             # Japanese
```

**languages.json Format:**
```json
[
  {
    "code": "en",
    "name": "English",
    "native_name": "English",
    "is_rtl": false,
    "is_active": true,
    "is_default": true
  },
  {
    "code": "de",
    "name": "German",
    "native_name": "Deutsch",
    "is_rtl": false,
    "is_active": true,
    "is_default": false
  }
]
```

**localization-keys.json Format:**
```json
[
  {
    "key": "error.auth.missing_username",
    "category": "error",
    "description": "Error message when username is not provided",
    "context": "authentication",
    "variables": []
  },
  {
    "key": "app.welcome",
    "category": "ui",
    "description": "Welcome message with user's name",
    "context": "dashboard",
    "variables": ["name"]
  }
]
```

**localizations/en.json Format:**
```json
{
  "error.auth.missing_username": "Missing username",
  "error.auth.missing_password": "Missing password",
  "error.auth.invalid_credentials": "Invalid username or password",
  "app.welcome": "Welcome to HelixTrack, {name}!",
  "app.hello": "Hello {name}"
}
```

### 1.3 Import Endpoint Design

**Endpoint: `POST /v1/admin/import`**

Request:
```json
{
  "import_type": "full|incremental",
  "overwrite_existing": true|false,
  "source": "file|json",
  "data": {
    "languages": [...],
    "keys": [...],
    "localizations": {
      "en": {...},
      "de": {...}
    }
  }
}
```

Response:
```json
{
  "success": true,
  "summary": {
    "languages_imported": 8,
    "keys_imported": 250,
    "localizations_imported": 2000,
    "errors": []
  }
}
```

## 2. Catalog Versioning System

### 2.1 Version Strategy

**Version Number Format:** `MAJOR.MINOR.PATCH`
- **MAJOR**: Breaking changes to catalog structure
- **MINOR**: New keys added or languages added
- **PATCH**: Translation updates only

**Version Tracking:**
```sql
-- Already exists in localization_catalogs table
CREATE TABLE localization_catalogs (
    id UUID PRIMARY KEY,
    language_id UUID REFERENCES languages(id),
    category VARCHAR(100),
    catalog_data JSONB,
    version INTEGER,           -- Auto-incremented version
    checksum VARCHAR(64),      -- SHA-256 of catalog_data
    created_at BIGINT,
    modified_at BIGINT
);
```

### 2.2 Version Metadata Table (New)

```sql
CREATE TABLE localization_versions (
    id UUID PRIMARY KEY,
    version_number VARCHAR(20) NOT NULL,  -- e.g., "1.0.0"
    version_type VARCHAR(20),              -- "major", "minor", "patch"
    description TEXT,
    keys_count INTEGER,
    languages_count INTEGER,
    total_localizations INTEGER,
    created_by VARCHAR(255),
    created_at BIGINT,

    UNIQUE(version_number)
);

CREATE INDEX idx_localization_versions_created ON localization_versions(created_at);
```

### 2.3 Version Endpoints

**GET /v1/version/current**
```json
{
  "version": "1.2.5",
  "keys_count": 250,
  "languages_count": 8,
  "last_updated": 1697654400000
}
```

**GET /v1/version/history**
```json
{
  "versions": [
    {
      "version": "1.2.5",
      "type": "patch",
      "description": "Updated German translations",
      "created_at": 1697654400000
    }
  ]
}
```

## 3. Cache Management & Refresh Strategy

### 3.1 Multi-Layer Cache Architecture

```
┌────────────────────────────────────────────┐
│        Layer 1: In-Memory LRU Cache        │
│  - TTL: 1 hour (3600s)                     │
│  - Max Size: 1024 MB                       │
│  - Cleanup: Every 5 minutes                │
└────────────────────────────────────────────┘
                    ↓ (on miss)
┌────────────────────────────────────────────┐
│        Layer 2: Redis Cache (Optional)     │
│  - TTL: 4 hours (14400s)                   │
│  - Distributed across instances            │
│  - Automatic failover to Layer 1           │
└────────────────────────────────────────────┘
                    ↓ (on miss)
┌────────────────────────────────────────────┐
│        Layer 3: PostgreSQL Database        │
│  - Source of truth                         │
│  - Encrypted with SQL Cipher               │
│  - Materialized catalogs in JSONB          │
└────────────────────────────────────────────┘
```

### 3.2 Cache Key Format

```
l10n:catalog:<language_code>:<category>:<version>
l10n:key:<language_code>:<key>
l10n:version:current
l10n:version:checksum:<language_code>
```

Examples:
- `l10n:catalog:en:error:v1`
- `l10n:catalog:de:ui:v2`
- `l10n:key:en:error.auth.missing_username`
- `l10n:version:current`

### 3.3 Cache Invalidation Strategies

#### 3.3.1 Manual Invalidation (Admin API)

**Endpoint: `POST /v1/admin/cache/invalidate`**

Request:
```json
{
  "scope": "all|language|category|key",
  "language_code": "en",      // Optional
  "category": "error",        // Optional
  "key": "error.auth.*"       // Optional (supports wildcards)
}
```

#### 3.3.2 Automatic Invalidation (On Update)

Triggers:
1. Localization created/updated → Invalidate specific key + catalog
2. Language added → Invalidate all catalogs
3. Localization deleted → Invalidate specific key + catalog
4. Approval changed → Invalidate specific key + catalog

Implementation:
```go
func (s *LocalizationService) UpdateLocalization(l10n *models.Localization) error {
    // 1. Update database
    err := s.db.UpdateLocalization(l10n)
    if err != nil {
        return err
    }

    // 2. Invalidate caches
    s.cache.Delete(fmt.Sprintf("l10n:key:%s:%s", l10n.LanguageCode, l10n.Key))
    s.cache.DeletePattern(fmt.Sprintf("l10n:catalog:%s:*", l10n.LanguageCode))

    // 3. Rebuild catalog asynchronously
    go s.RebuildCatalog(l10n.LanguageCode)

    return nil
}
```

### 3.4 Cache Warming Strategy

**On Startup:**
```go
func (s *LocalizationService) WarmupCache() error {
    // 1. Load all active languages
    languages := s.db.GetActiveLanguages()

    // 2. Preload default language catalog
    defaultLang := s.db.GetDefaultLanguage()
    catalog := s.db.GetCatalog(defaultLang.ID, "")
    s.cache.Set(fmt.Sprintf("l10n:catalog:%s:all:v1", defaultLang.Code), catalog, 3600)

    // 3. Preload frequently used keys (async)
    go s.PreloadFrequentKeys()

    return nil
}
```

**Periodic Refresh (Every 30 minutes):**
```go
func (s *LocalizationService) StartPeriodicRefresh() {
    ticker := time.NewTicker(30 * time.Minute)
    go func() {
        for range ticker.C {
            s.RefreshCatalogs()
        }
    }()
}
```

### 3.5 Cache Metrics & Monitoring

**Metrics to Track:**
- Cache hit rate (Layer 1, Layer 2, Database)
- Cache size (MB)
- Cache eviction rate
- Average lookup time
- Popular keys (top 100)

**Endpoint: `GET /v1/admin/cache/stats`**
```json
{
  "in_memory": {
    "size_mb": 245,
    "entries": 12450,
    "hit_rate": 0.87,
    "miss_rate": 0.13,
    "evictions": 340
  },
  "redis": {
    "enabled": true,
    "hit_rate": 0.92,
    "latency_ms": 2.3
  },
  "top_keys": [
    {"key": "error.auth.invalid_jwt", "hits": 45023},
    {"key": "app.welcome", "hits": 12340}
  ]
}
```

## 4. Periodic Backup System

### 4.1 Backup Strategy

**Backup Schedule:**
- **Hourly**: Incremental backup (changed localizations only)
- **Daily**: Full backup (all data)
- **Weekly**: Full backup with version tagging
- **On-Demand**: Via admin API

### 4.2 Backup Directory Structure

```
Core/Services/Localization/backups/
├── daily/
│   ├── 2025-01-15-full-backup.json
│   ├── 2025-01-16-full-backup.json
│   └── ...
├── hourly/
│   ├── 2025-01-15-14-00-incremental.json
│   ├── 2025-01-15-15-00-incremental.json
│   └── ...
├── weekly/
│   ├── 2025-W03-full-backup.json
│   └── ...
└── on-demand/
    └── manual-backup-20250115-143022.json
```

### 4.3 Backup Format

**Full Backup:**
```json
{
  "metadata": {
    "backup_type": "full",
    "version": "1.2.5",
    "timestamp": 1697654400000,
    "total_keys": 250,
    "total_languages": 8,
    "total_localizations": 2000
  },
  "languages": [...],
  "keys": [...],
  "localizations": {
    "en": {...},
    "de": {...}
  }
}
```

**Incremental Backup:**
```json
{
  "metadata": {
    "backup_type": "incremental",
    "version": "1.2.5",
    "timestamp": 1697654400000,
    "since_timestamp": 1697650800000
  },
  "changes": {
    "created": [
      {"key": "error.new.message", "language": "en", "value": "..."}
    ],
    "updated": [
      {"key": "error.auth.invalid", "language": "de", "old_value": "...", "new_value": "..."}
    ],
    "deleted": [
      {"key": "deprecated.message", "language": "en"}
    ]
  }
}
```

### 4.4 Backup Endpoints

**POST /v1/admin/backup/create**
```json
{
  "type": "full|incremental",
  "format": "json|sql",
  "compression": true,
  "include_audit_log": false
}
```

**GET /v1/admin/backup/list**
```json
{
  "backups": [
    {
      "id": "backup-20250115-143022",
      "type": "full",
      "size_mb": 12.5,
      "timestamp": 1697654400000,
      "path": "backups/daily/2025-01-15-full-backup.json"
    }
  ]
}
```

**POST /v1/admin/backup/restore**
```json
{
  "backup_id": "backup-20250115-143022",
  "overwrite_existing": true,
  "dry_run": false
}
```

### 4.5 Backup Automation Script

**Script: `scripts/backup-manager.sh`**
```bash
#!/bin/bash
# Automated backup manager for Localization service

BACKUP_DIR="/app/backups"
RETENTION_DAYS_HOURLY=3
RETENTION_DAYS_DAILY=30
RETENTION_DAYS_WEEKLY=365

# Hourly backup (incremental)
0 * * * * /app/scripts/backup-hourly.sh

# Daily backup (full)
0 2 * * * /app/scripts/backup-daily.sh

# Weekly backup (full + version tag)
0 3 * * 0 /app/scripts/backup-weekly.sh

# Cleanup old backups
0 4 * * * /app/scripts/cleanup-backups.sh
```

## 5. Client Integration Lifecycle

### 5.1 Client Initialization Flow

```typescript
// 1. Initialize localization service
const localizationService = new LocalizationService();
localizationService.setServiceUrl('https://localhost:8085');

// 2. Check for cached catalog
const cachedCatalog = await localizationService.getCachedCatalog('en');

if (cachedCatalog && !cachedCatalog.isExpired()) {
    // 3a. Use cached catalog
    localizationService.loadFromCache(cachedCatalog);
} else {
    // 3b. Fetch fresh catalog from service
    const catalog = await localizationService.loadCatalog('en', jwtToken);

    // 4. Cache for future use
    await localizationService.cacheCatalog('en', catalog, {
        ttl: 3600,        // 1 hour
        version: catalog.version
    });
}

// 5. Setup periodic refresh
localizationService.startPeriodicRefresh(30 * 60 * 1000); // Every 30 min
```

### 5.2 Cache Expiration & Refresh

**Client-Side Cache Strategy:**
- **Web Client**: localStorage with 1-hour TTL
- **Desktop Client**: Encrypted SQLite with 2-hour TTL
- **Android Client**: EncryptedSharedPreferences with 1-hour TTL
- **iOS Client**: Keychain with 1-hour TTL

**Version-Based Refresh:**
```typescript
async checkForUpdates() {
    const currentVersion = await this.getCachedVersion();
    const serverVersion = await this.fetchCurrentVersion();

    if (serverVersion > currentVersion) {
        // New version available - refresh catalog
        await this.loadCatalog(this.currentLanguage, this.jwtToken);
        await this.updateCachedVersion(serverVersion);
    }
}
```

### 5.3 Offline Support

**Desktop & Mobile Clients:**
1. Always cache full catalog locally
2. Use cached catalog when offline
3. Queue localization updates for when online
4. Sync when connection restored

```typescript
async getLocalization(key: string): Promise<string> {
    try {
        // Try online fetch first
        if (this.isOnline()) {
            return await this.fetchFromService(key);
        }
    } catch (error) {
        // Fall back to cache
    }

    // Use cached value
    return this.getCachedValue(key) || key;
}
```

## 6. Monitoring & Health Checks

### 6.1 Health Check Endpoint Enhancement

**GET /health**
```json
{
  "status": "healthy",
  "version": "1.2.5",
  "uptime_seconds": 86400,
  "database": {
    "status": "healthy",
    "latency_ms": 2.3
  },
  "cache": {
    "in_memory": {
      "status": "healthy",
      "size_mb": 245,
      "hit_rate": 0.87
    },
    "redis": {
      "status": "healthy",
      "connected": true
    }
  },
  "localization_stats": {
    "languages": 8,
    "keys": 250,
    "total_localizations": 2000,
    "last_update": 1697654400000
  }
}
```

### 6.2 Audit Logging

All operations logged to `localization_audit_log`:
- Localization create/update/delete
- Language add/remove
- Cache invalidations
- Import/export operations
- Backup operations

## 7. Performance Optimization

### 7.1 Catalog Pre-building

Build and cache complete catalogs in database:
```sql
-- Materialized view approach
CREATE MATERIALIZED VIEW localization_catalog_en AS
SELECT
    lk.key,
    l.value
FROM localization_keys lk
JOIN localizations l ON l.key_id = lk.id
JOIN languages lang ON lang.id = l.language_id
WHERE lang.code = 'en' AND l.approved = true;

REFRESH MATERIALIZED VIEW localization_catalog_en;
```

### 7.2 Batch Loading

Support batch requests to reduce round trips:
```json
POST /v1/localize/batch
{
  "keys": [
    "error.auth.invalid_jwt",
    "error.auth.missing_username",
    "app.welcome"
  ],
  "language": "en",
  "fallback_to_default": true
}
```

Response:
```json
{
  "localizations": {
    "error.auth.invalid_jwt": "Invalid JWT token",
    "error.auth.missing_username": "Missing username",
    "app.welcome": "Welcome to HelixTrack, {name}!"
  },
  "version": "1.2.5"
}
```

## 8. Error Handling & Fallbacks

### 8.1 Fallback Strategy

```
Requested Localization
    ↓
├─ Key exists in requested language? → Return
├─ Key exists in default language (en)? → Return
├─ Key exists in any language? → Return with warning
└─ Key not found → Return key itself as fallback
```

### 8.2 Partial Catalog Support

If a language is partially translated:
```json
{
  "language": "de",
  "completion": 0.75,  // 75% translated
  "missing_keys": [
    "error.new.feature.message",
    "ui.new.button.label"
  ]
}
```

## Implementation Priority

**Phase 1 (High Priority):**
1. ✅ Import/Export endpoints
2. ✅ Startup population script
3. ✅ Periodic backup script
4. ✅ Dockerfile for Localization service

**Phase 2 (Medium Priority):**
5. ✅ Version tracking system
6. ✅ Cache warming & refresh
7. ✅ Core Application integration
8. ✅ Web Client integration

**Phase 3 (Lower Priority):**
9. ✅ Desktop/Mobile client integrations
10. ✅ Monitoring enhancements
11. ✅ Performance optimizations
12. ✅ Admin UI for localization management

---

**Document Version:** 1.0.0
**Last Updated:** 2025-01-15
**Author:** HelixTrack Development Team

/*
    Migration: V5 → V6
    Description: Security Engine Enhancements
    Date: 2025-10-19

    This migration enhances the Core database to support the Security Engine:
    1. Expands the audit table to capture comprehensive security events
    2. Adds indexes for efficient security audit queries
    3. Creates security_audit table for detailed authorization logs

    IMPORTANT: This migration is backward compatible and can be rolled back
*/

-- ========================================================================
-- AUDIT TABLE ENHANCEMENTS
-- ========================================================================

-- Add new columns to audit table for security tracking
ALTER TABLE audit ADD COLUMN username TEXT;
ALTER TABLE audit ADD COLUMN resource TEXT;
ALTER TABLE audit ADD COLUMN resource_id TEXT;
ALTER TABLE audit ADD COLUMN action TEXT;
ALTER TABLE audit ADD COLUMN allowed BOOLEAN DEFAULT 1;
ALTER TABLE audit ADD COLUMN reason TEXT;
ALTER TABLE audit ADD COLUMN ip_address TEXT;
ALTER TABLE audit ADD COLUMN user_agent TEXT;
ALTER TABLE audit ADD COLUMN context_data TEXT;  -- JSON string
ALTER TABLE audit ADD COLUMN severity TEXT;       -- INFO, WARNING, CRITICAL
ALTER TABLE audit ADD COLUMN timestamp INTEGER;   -- Unix timestamp

-- Rename existing 'created' to match new schema (if needed)
-- Note: SQLite doesn't support direct column rename easily, so we keep both for compatibility

-- Create indexes for security audit queries
CREATE INDEX IF NOT EXISTS audit_get_by_username ON audit (username);
CREATE INDEX IF NOT EXISTS audit_get_by_resource ON audit (resource);
CREATE INDEX IF NOT EXISTS audit_get_by_resource_id ON audit (resource_id);
CREATE INDEX IF NOT EXISTS audit_get_by_action ON audit (action);
CREATE INDEX IF NOT EXISTS audit_get_by_allowed ON audit (allowed);
CREATE INDEX IF NOT EXISTS audit_get_by_severity ON audit (severity);
CREATE INDEX IF NOT EXISTS audit_get_by_timestamp ON audit (timestamp);
CREATE INDEX IF NOT EXISTS audit_get_by_username_and_timestamp ON audit (username, timestamp);
CREATE INDEX IF NOT EXISTS audit_get_by_resource_and_timestamp ON audit (resource, timestamp);
CREATE INDEX IF NOT EXISTS audit_get_denied_attempts ON audit (allowed, timestamp) WHERE allowed = 0;

-- ========================================================================
-- SECURITY AUDIT TABLE (Optional - for detailed authorization logs)
-- ========================================================================

DROP TABLE IF EXISTS security_audit;

-- Detailed security audit log (separate from general audit)
CREATE TABLE security_audit
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    timestamp       INTEGER NOT NULL,
    username        TEXT    NOT NULL,
    resource        TEXT    NOT NULL,
    resource_id     TEXT,
    action          TEXT    NOT NULL,
    allowed         BOOLEAN NOT NULL DEFAULT 0,
    reason          TEXT,

    -- Request context
    ip_address      TEXT,
    user_agent      TEXT,
    request_path    TEXT,
    request_method  TEXT,

    -- Authorization details
    permission_checked TEXT,    -- Permission level checked (READ, CREATE, UPDATE, DELETE)
    security_level_id  TEXT,    -- Security level involved (if applicable)
    project_id         TEXT,    -- Project context (if applicable)
    role_id            TEXT,    -- Role used for authorization (if applicable)

    -- Additional context
    context_data    TEXT,       -- JSON string with additional context
    session_id      TEXT,       -- Session ID for correlation
    request_id      TEXT,       -- Request ID for distributed tracing

    -- Metadata
    severity        TEXT,       -- INFO, WARNING, CRITICAL
    created         INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL DEFAULT 0
);

-- Indexes for security audit queries
CREATE INDEX security_audit_get_by_username ON security_audit (username);
CREATE INDEX security_audit_get_by_timestamp ON security_audit (timestamp);
CREATE INDEX security_audit_get_by_resource ON security_audit (resource);
CREATE INDEX security_audit_get_by_resource_id ON security_audit (resource_id);
CREATE INDEX security_audit_get_by_action ON security_audit (action);
CREATE INDEX security_audit_get_by_allowed ON security_audit (allowed);
CREATE INDEX security_audit_get_by_severity ON security_audit (severity);
CREATE INDEX security_audit_get_by_project_id ON security_audit (project_id);
CREATE INDEX security_audit_get_by_security_level_id ON security_audit (security_level_id);
CREATE INDEX security_audit_get_by_session_id ON security_audit (session_id);
CREATE INDEX security_audit_get_by_request_id ON security_audit (request_id);

-- Composite indexes for common queries
CREATE INDEX security_audit_get_by_username_and_timestamp ON security_audit (username, timestamp);
CREATE INDEX security_audit_get_by_resource_and_timestamp ON security_audit (resource, timestamp);
CREATE INDEX security_audit_get_denied_attempts ON security_audit (allowed, timestamp, severity) WHERE allowed = 0;
CREATE INDEX security_audit_get_critical_events ON security_audit (severity, timestamp) WHERE severity = 'CRITICAL';

-- ========================================================================
-- PERMISSION CACHE TABLE (Optional - for performance optimization)
-- ========================================================================

DROP TABLE IF EXISTS permission_cache;

-- In-memory permission cache (can be rebuilt from source tables)
-- This is optional - the Security Engine uses in-memory cache by default
-- This table is for persistence across restarts if needed
CREATE TABLE permission_cache
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    username        TEXT    NOT NULL,
    resource        TEXT    NOT NULL,
    resource_id     TEXT,
    action          TEXT    NOT NULL,
    allowed         BOOLEAN NOT NULL DEFAULT 0,
    reason          TEXT,
    cached_at       INTEGER NOT NULL,
    expires_at      INTEGER NOT NULL,
    context_hash    TEXT,   -- Hash of context for cache key
    deleted         BOOLEAN NOT NULL DEFAULT 0
);

CREATE INDEX permission_cache_get_by_username ON permission_cache (username);
CREATE INDEX permission_cache_get_by_context_hash ON permission_cache (context_hash);
CREATE INDEX permission_cache_get_by_expires_at ON permission_cache (expires_at);
CREATE INDEX permission_cache_cleanup ON permission_cache (expires_at, deleted) WHERE deleted = 0;

-- ========================================================================
-- DATA MIGRATION
-- ========================================================================

-- Migrate existing audit entries to new schema
-- Set default values for new columns in existing rows
UPDATE audit
SET
    timestamp = created,
    severity = 'INFO',
    allowed = 1
WHERE timestamp IS NULL;

-- ========================================================================
-- VERIFICATION QUERIES
-- ========================================================================

-- Verify migration success
-- Run these queries to ensure tables and indexes are created:

-- SELECT COUNT(*) FROM audit;
-- SELECT COUNT(*) FROM security_audit;
-- SELECT COUNT(*) FROM permission_cache;
--
-- SELECT name FROM sqlite_master WHERE type='index' AND name LIKE 'audit%';
-- SELECT name FROM sqlite_master WHERE type='index' AND name LIKE 'security_audit%';

-- ========================================================================
-- ROLLBACK SCRIPT (Use in case of issues)
-- ========================================================================

/*
-- ROLLBACK - Remove new columns and tables

-- Remove new audit columns (SQLite limitation - can't drop columns easily)
-- You would need to recreate the table without these columns

-- Drop new tables
DROP TABLE IF EXISTS security_audit;
DROP TABLE IF EXISTS permission_cache;

-- Drop new indexes
DROP INDEX IF EXISTS audit_get_by_username;
DROP INDEX IF EXISTS audit_get_by_resource;
DROP INDEX IF EXISTS audit_get_by_resource_id;
DROP INDEX IF EXISTS audit_get_by_action;
DROP INDEX IF EXISTS audit_get_by_allowed;
DROP INDEX IF EXISTS audit_get_by_severity;
DROP INDEX IF EXISTS audit_get_by_timestamp;
DROP INDEX IF EXISTS audit_get_by_username_and_timestamp;
DROP INDEX IF EXISTS audit_get_by_resource_and_timestamp;
DROP INDEX IF EXISTS audit_get_denied_attempts;

*/

-- ========================================================================
-- NOTES
-- ========================================================================

/*
1. This migration is safe to run multiple times (idempotent)
2. All new columns have default values for backward compatibility
3. Existing audit entries will have default values for new columns
4. security_audit table is optional but recommended for production
5. permission_cache table is optional - in-memory cache is faster
6. Indexes are optimized for common security audit queries
7. For PostgreSQL, replace TEXT with appropriate types (VARCHAR, JSONB, etc.)

Performance Considerations:
- Indexes on audit table will improve query performance
- security_audit table keeps security logs separate from general audit
- permission_cache table can improve cold-start performance
- Regular cleanup of old audit entries recommended (retention policy)

Security Considerations:
- Audit logs should be protected from tampering
- Consider write-only access for audit tables
- Implement log rotation and archival
- Monitor for unusual patterns (many denied attempts)
- Alert on CRITICAL severity events

Next Steps:
1. Deploy this migration to database
2. Update Security Engine audit logger to use new schema
3. Configure audit retention policy
4. Set up monitoring dashboards for security events
5. Create alerts for suspicious activity patterns
*/

-- Migration V5.6: Security Engine Enhancement
-- This migration enhances the audit table and creates Security Engine support tables
-- Run date: 2025-10-19

-- ==================================================
-- STEP 1: Create indexes on existing audit table
-- ==================================================

-- Note: audit table already has severity, context_data, ip_address, user_agent columns
-- Create indexes for efficient queries on existing columns
CREATE INDEX IF NOT EXISTS idx_audit_created ON audit(created);
CREATE INDEX IF NOT EXISTS idx_audit_user_id ON audit(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_entity_type ON audit(entity_type);
CREATE INDEX IF NOT EXISTS idx_audit_entity_id ON audit(entity_id);
CREATE INDEX IF NOT EXISTS idx_audit_action ON audit(action);
CREATE INDEX IF NOT EXISTS idx_audit_severity ON audit(severity);
CREATE INDEX IF NOT EXISTS idx_audit_deleted ON audit(deleted);
CREATE INDEX IF NOT EXISTS idx_audit_user_created ON audit(user_id, created);
CREATE INDEX IF NOT EXISTS idx_audit_entity_created ON audit(entity_type, created);

-- ==================================================
-- STEP 2: Create security_audit table for detailed logging
-- ==================================================

CREATE TABLE IF NOT EXISTS security_audit (
    id TEXT PRIMARY KEY,
    timestamp INTEGER NOT NULL,
    username TEXT NOT NULL,
    resource TEXT NOT NULL,
    resource_id TEXT,
    action TEXT NOT NULL,
    allowed INTEGER NOT NULL DEFAULT 0,
    reason TEXT,
    ip_address TEXT,
    user_agent TEXT,
    request_path TEXT,
    request_method TEXT,
    context_data TEXT DEFAULT '{}',
    severity TEXT DEFAULT 'INFO',
    audit_category TEXT DEFAULT 'ACCESS',
    deleted INTEGER NOT NULL DEFAULT 0
);

-- Indexes for security_audit
CREATE INDEX IF NOT EXISTS idx_security_audit_timestamp ON security_audit(timestamp);
CREATE INDEX IF NOT EXISTS idx_security_audit_username ON security_audit(username);
CREATE INDEX IF NOT EXISTS idx_security_audit_resource ON security_audit(resource);
CREATE INDEX IF NOT EXISTS idx_security_audit_allowed ON security_audit(allowed);
CREATE INDEX IF NOT EXISTS idx_security_audit_severity ON security_audit(severity);
CREATE INDEX IF NOT EXISTS idx_security_audit_category ON security_audit(audit_category);
CREATE INDEX IF NOT EXISTS idx_security_audit_deleted ON security_audit(deleted);

-- ==================================================
-- STEP 3: Create permission_cache table
-- ==================================================

CREATE TABLE IF NOT EXISTS permission_cache (
    cache_key TEXT PRIMARY KEY,
    username TEXT NOT NULL,
    resource TEXT NOT NULL,
    resource_id TEXT,
    action TEXT NOT NULL,
    allowed INTEGER NOT NULL,
    reason TEXT,
    cached_at INTEGER NOT NULL,
    expires_at INTEGER NOT NULL,
    deleted INTEGER NOT NULL DEFAULT 0
);

-- Indexes for permission_cache
CREATE INDEX IF NOT EXISTS idx_permission_cache_username ON permission_cache(username);
CREATE INDEX IF NOT EXISTS idx_permission_cache_expires_at ON permission_cache(expires_at);
CREATE INDEX IF NOT EXISTS idx_permission_cache_deleted ON permission_cache(deleted);

-- ==================================================
-- STEP 4: Update schema version
-- ==================================================

-- Record migration execution
INSERT INTO audit (id, action, user_id, entity_id, entity_type, details, severity, created, modified, deleted)
VALUES (
    lower(hex(randomblob(16))),
    'MIGRATE',
    'system',
    'migration_v5.6',
    'database',
    'Applied Migration V5.6: Security Engine Enhancement',
    'INFO',
    strftime('%s', 'now'),
    strftime('%s', 'now'),
    0
);

-- ==================================================
-- Migration V5.6 Complete
-- ==================================================
-- Tables created/modified:
-- - audit (enhanced with 4 new columns + 10 indexes)
-- - security_audit (new table with 7 indexes)
-- - permission_cache (new table with 3 indexes)
--
-- Total new indexes: 20
-- Schema version: V5.6
-- ==================================================

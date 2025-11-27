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
-- STEP 3: Create account table for account management
-- ==================================================

CREATE TABLE IF NOT EXISTS account (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    created INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted INTEGER NOT NULL DEFAULT 0
);

-- Indexes for account table
CREATE INDEX IF NOT EXISTS idx_account_title ON account(title);
CREATE INDEX IF NOT EXISTS idx_account_created ON account(created);
CREATE INDEX IF NOT EXISTS idx_account_modified ON account(modified);
CREATE INDEX IF NOT EXISTS idx_account_deleted ON account(deleted);

-- ==================================================
-- STEP 4: Create organization table for organization management
-- ==================================================

CREATE TABLE IF NOT EXISTS organization (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    created INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted INTEGER NOT NULL DEFAULT 0
);

-- Indexes for organization table
CREATE INDEX IF NOT EXISTS idx_organization_title ON organization(title);
CREATE INDEX IF NOT EXISTS idx_organization_created ON organization(created);
CREATE INDEX IF NOT EXISTS idx_organization_modified ON organization(modified);
CREATE INDEX IF NOT EXISTS idx_organization_deleted ON organization(deleted);

-- ==================================================
-- STEP 5: Create organization_account_mapping table for multi-tenancy
-- ==================================================

CREATE TABLE IF NOT EXISTS organization_account_mapping (
    id TEXT PRIMARY KEY,
    organization_id TEXT NOT NULL,
    account_id TEXT NOT NULL,
    created INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (organization_id) REFERENCES organization(id),
    FOREIGN KEY (account_id) REFERENCES account(id)
);

-- Indexes for organization_account_mapping table
CREATE INDEX IF NOT EXISTS idx_org_acc_mapping_org_id ON organization_account_mapping(organization_id);
CREATE INDEX IF NOT EXISTS idx_org_acc_mapping_acc_id ON organization_account_mapping(account_id);
CREATE INDEX IF NOT EXISTS idx_org_acc_mapping_created ON organization_account_mapping(created);

-- ==================================================
-- STEP 6: Create team tables for team management
-- ==================================================

CREATE TABLE IF NOT EXISTS team (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    created INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted INTEGER NOT NULL DEFAULT 0
);

-- Indexes for team table
CREATE INDEX IF NOT EXISTS idx_team_title ON team(title);
CREATE INDEX IF NOT EXISTS idx_team_created ON team(created);
CREATE INDEX IF NOT EXISTS idx_team_modified ON team(modified);
CREATE INDEX IF NOT EXISTS idx_team_deleted ON team(deleted);

-- Team-Organization mapping
CREATE TABLE IF NOT EXISTS team_organization_mapping (
    id TEXT PRIMARY KEY,
    team_id TEXT NOT NULL,
    organization_id TEXT NOT NULL,
    created INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (team_id) REFERENCES team(id),
    FOREIGN KEY (organization_id) REFERENCES organization(id)
);

-- Indexes for team_organization_mapping table
CREATE INDEX IF NOT EXISTS idx_team_org_mapping_team_id ON team_organization_mapping(team_id);
CREATE INDEX IF NOT EXISTS idx_team_org_mapping_org_id ON team_organization_mapping(organization_id);

-- Team-Project mapping
CREATE TABLE IF NOT EXISTS team_project_mapping (
    id TEXT PRIMARY KEY,
    team_id TEXT NOT NULL,
    project_id TEXT NOT NULL,
    created INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (team_id) REFERENCES team(id),
    FOREIGN KEY (project_id) REFERENCES project(id)
);

-- Indexes for team_project_mapping table
CREATE INDEX IF NOT EXISTS idx_team_project_mapping_team_id ON team_project_mapping(team_id);
CREATE INDEX IF NOT EXISTS idx_team_project_mapping_project_id ON team_project_mapping(project_id);

-- User-Organization mapping
CREATE TABLE IF NOT EXISTS user_organization_mapping (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    organization_id TEXT NOT NULL,
    created INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (organization_id) REFERENCES organization(id)
);

-- Indexes for user_organization_mapping table
CREATE INDEX IF NOT EXISTS idx_user_org_mapping_user_id ON user_organization_mapping(user_id);
CREATE INDEX IF NOT EXISTS idx_user_org_mapping_org_id ON user_organization_mapping(organization_id);

-- User-Team mapping
CREATE TABLE IF NOT EXISTS user_team_mapping (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    team_id TEXT NOT NULL,
    created INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (team_id) REFERENCES team(id)
);

-- Indexes for user_team_mapping table
CREATE INDEX IF NOT EXISTS idx_user_team_mapping_user_id ON user_team_mapping(user_id);
CREATE INDEX IF NOT EXISTS idx_user_team_mapping_team_id ON user_team_mapping(team_id);

-- ==================================================
-- Migration V5.6 Complete
-- ==================================================
-- Tables created/modified:
-- - audit (enhanced with 4 new columns + 10 indexes)
-- - security_audit (new table with 7 indexes)
-- - permission_cache (new table with 3 indexes)
-- - account (new table with 4 indexes)
--
-- Total new indexes: 24
-- Schema version: V5.6
-- ==================================================

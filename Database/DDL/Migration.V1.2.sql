/*
    Migration: V1 to V2
    Version: 1.2
    Date: 2025-10-10

    This migration script upgrades an existing V1 database to V2 schema.
    It adds Phase 1 (Priority 1) JIRA feature parity enhancements.

    IMPORTANT: Backup your database before running this migration!

    Features Added:
    - Priority System
    - Resolution System
    - Project Lead & Assignee fields
    - Watchers
    - Product Versions (Affected/Fix Versions)
    - Saved Filters with sharing
    - Custom Fields system
    - Enhanced ticket fields (assignee, reporter, estimates, due date)
*/

/*
    ========================================================================
    STEP 1: Create New Tables
    ========================================================================
*/

-- Priority System
CREATE TABLE IF NOT EXISTS priority
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    level       INTEGER NOT NULL,  -- 1 (Lowest) to 5 (Highest)
    icon        TEXT,
    color       TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL
);

CREATE INDEX IF NOT EXISTS priorities_get_by_title ON priority (title);
CREATE INDEX IF NOT EXISTS priorities_get_by_level ON priority (level);
CREATE INDEX IF NOT EXISTS priorities_get_by_deleted ON priority (deleted);
CREATE INDEX IF NOT EXISTS priorities_get_by_created ON priority (created);
CREATE INDEX IF NOT EXISTS priorities_get_by_modified ON priority (modified);
CREATE INDEX IF NOT EXISTS priorities_get_by_created_and_modified ON priority (created, modified);

-- Resolution System
CREATE TABLE IF NOT EXISTS resolution
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL
);

CREATE INDEX IF NOT EXISTS resolutions_get_by_title ON resolution (title);
CREATE INDEX IF NOT EXISTS resolutions_get_by_deleted ON resolution (deleted);
CREATE INDEX IF NOT EXISTS resolutions_get_by_created ON resolution (created);
CREATE INDEX IF NOT EXISTS resolutions_get_by_modified ON resolution (modified);
CREATE INDEX IF NOT EXISTS resolutions_get_by_created_and_modified ON resolution (created, modified);

-- Ticket Watchers
CREATE TABLE IF NOT EXISTS ticket_watcher_mapping
(
    id         TEXT    NOT NULL PRIMARY KEY UNIQUE,
    ticket_id  TEXT    NOT NULL,
    user_id    TEXT    NOT NULL,
    created    INTEGER NOT NULL,
    deleted    BOOLEAN NOT NULL,
    UNIQUE (ticket_id, user_id)
);

CREATE INDEX IF NOT EXISTS ticket_watchers_get_by_ticket_id ON ticket_watcher_mapping (ticket_id);
CREATE INDEX IF NOT EXISTS ticket_watchers_get_by_user_id ON ticket_watcher_mapping (user_id);
CREATE INDEX IF NOT EXISTS ticket_watchers_get_by_ticket_and_user ON ticket_watcher_mapping (ticket_id, user_id);
CREATE INDEX IF NOT EXISTS ticket_watchers_get_by_deleted ON ticket_watcher_mapping (deleted);
CREATE INDEX IF NOT EXISTS ticket_watchers_get_by_created ON ticket_watcher_mapping (created);

-- Product Versions
CREATE TABLE IF NOT EXISTS version
(
    id           TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title        TEXT    NOT NULL,
    description  TEXT,
    project_id   TEXT    NOT NULL,
    start_date   INTEGER,
    release_date INTEGER,
    released     BOOLEAN NOT NULL DEFAULT FALSE,
    archived     BOOLEAN NOT NULL DEFAULT FALSE,
    created      INTEGER NOT NULL,
    modified     INTEGER NOT NULL,
    deleted      BOOLEAN NOT NULL
);

CREATE INDEX IF NOT EXISTS versions_get_by_title ON version (title);
CREATE INDEX IF NOT EXISTS versions_get_by_project_id ON version (project_id);
CREATE INDEX IF NOT EXISTS versions_get_by_released ON version (released);
CREATE INDEX IF NOT EXISTS versions_get_by_archived ON version (archived);
CREATE INDEX IF NOT EXISTS versions_get_by_release_date ON version (release_date);
CREATE INDEX IF NOT EXISTS versions_get_by_deleted ON version (deleted);
CREATE INDEX IF NOT EXISTS versions_get_by_created ON version (created);
CREATE INDEX IF NOT EXISTS versions_get_by_modified ON version (modified);
CREATE INDEX IF NOT EXISTS versions_get_by_created_and_modified ON version (created, modified);

-- Ticket Affected Versions
CREATE TABLE IF NOT EXISTS ticket_affected_version_mapping
(
    id         TEXT    NOT NULL PRIMARY KEY UNIQUE,
    ticket_id  TEXT    NOT NULL,
    version_id TEXT    NOT NULL,
    created    INTEGER NOT NULL,
    deleted    BOOLEAN NOT NULL,
    UNIQUE (ticket_id, version_id)
);

CREATE INDEX IF NOT EXISTS ticket_affected_versions_get_by_ticket_id ON ticket_affected_version_mapping (ticket_id);
CREATE INDEX IF NOT EXISTS ticket_affected_versions_get_by_version_id ON ticket_affected_version_mapping (version_id);
CREATE INDEX IF NOT EXISTS ticket_affected_versions_get_by_deleted ON ticket_affected_version_mapping (deleted);
CREATE INDEX IF NOT EXISTS ticket_affected_versions_get_by_created ON ticket_affected_version_mapping (created);

-- Ticket Fix Versions
CREATE TABLE IF NOT EXISTS ticket_fix_version_mapping
(
    id         TEXT    NOT NULL PRIMARY KEY UNIQUE,
    ticket_id  TEXT    NOT NULL,
    version_id TEXT    NOT NULL,
    created    INTEGER NOT NULL,
    deleted    BOOLEAN NOT NULL,
    UNIQUE (ticket_id, version_id)
);

CREATE INDEX IF NOT EXISTS ticket_fix_versions_get_by_ticket_id ON ticket_fix_version_mapping (ticket_id);
CREATE INDEX IF NOT EXISTS ticket_fix_versions_get_by_version_id ON ticket_fix_version_mapping (version_id);
CREATE INDEX IF NOT EXISTS ticket_fix_versions_get_by_deleted ON ticket_fix_version_mapping (deleted);
CREATE INDEX IF NOT EXISTS ticket_fix_versions_get_by_created ON ticket_fix_version_mapping (created);

-- Saved Filters
CREATE TABLE IF NOT EXISTS filter
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL,
    description TEXT,
    owner_id    TEXT    NOT NULL,
    query       TEXT    NOT NULL,
    is_public   BOOLEAN NOT NULL DEFAULT FALSE,
    is_favorite BOOLEAN NOT NULL DEFAULT FALSE,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL
);

CREATE INDEX IF NOT EXISTS filters_get_by_title ON filter (title);
CREATE INDEX IF NOT EXISTS filters_get_by_owner_id ON filter (owner_id);
CREATE INDEX IF NOT EXISTS filters_get_by_is_public ON filter (is_public);
CREATE INDEX IF NOT EXISTS filters_get_by_is_favorite ON filter (is_favorite);
CREATE INDEX IF NOT EXISTS filters_get_by_deleted ON filter (deleted);
CREATE INDEX IF NOT EXISTS filters_get_by_created ON filter (created);
CREATE INDEX IF NOT EXISTS filters_get_by_modified ON filter (modified);
CREATE INDEX IF NOT EXISTS filters_get_by_created_and_modified ON filter (created, modified);

-- Filter Sharing
CREATE TABLE IF NOT EXISTS filter_share_mapping
(
    id         TEXT    NOT NULL PRIMARY KEY UNIQUE,
    filter_id  TEXT    NOT NULL,
    user_id    TEXT,
    team_id    TEXT,
    project_id TEXT,
    created    INTEGER NOT NULL,
    deleted    BOOLEAN NOT NULL
);

CREATE INDEX IF NOT EXISTS filter_shares_get_by_filter_id ON filter_share_mapping (filter_id);
CREATE INDEX IF NOT EXISTS filter_shares_get_by_user_id ON filter_share_mapping (user_id);
CREATE INDEX IF NOT EXISTS filter_shares_get_by_team_id ON filter_share_mapping (team_id);
CREATE INDEX IF NOT EXISTS filter_shares_get_by_project_id ON filter_share_mapping (project_id);
CREATE INDEX IF NOT EXISTS filter_shares_get_by_deleted ON filter_share_mapping (deleted);
CREATE INDEX IF NOT EXISTS filter_shares_get_by_created ON filter_share_mapping (created);

-- Custom Fields
CREATE TABLE IF NOT EXISTS custom_field
(
    id            TEXT    NOT NULL PRIMARY KEY UNIQUE,
    field_name    TEXT    NOT NULL,
    field_type    TEXT    NOT NULL,
    description   TEXT,
    project_id    TEXT,
    is_required   BOOLEAN NOT NULL DEFAULT FALSE,
    default_value TEXT,
    configuration TEXT,
    created       INTEGER NOT NULL,
    modified      INTEGER NOT NULL,
    deleted       BOOLEAN NOT NULL
);

CREATE INDEX IF NOT EXISTS custom_fields_get_by_field_name ON custom_field (field_name);
CREATE INDEX IF NOT EXISTS custom_fields_get_by_field_type ON custom_field (field_type);
CREATE INDEX IF NOT EXISTS custom_fields_get_by_project_id ON custom_field (project_id);
CREATE INDEX IF NOT EXISTS custom_fields_get_by_is_required ON custom_field (is_required);
CREATE INDEX IF NOT EXISTS custom_fields_get_by_deleted ON custom_field (deleted);
CREATE INDEX IF NOT EXISTS custom_fields_get_by_created ON custom_field (created);
CREATE INDEX IF NOT EXISTS custom_fields_get_by_modified ON custom_field (modified);
CREATE INDEX IF NOT EXISTS custom_fields_get_by_created_and_modified ON custom_field (created, modified);

-- Custom Field Options
CREATE TABLE IF NOT EXISTS custom_field_option
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    custom_field_id TEXT    NOT NULL,
    value           TEXT    NOT NULL,
    display_value   TEXT    NOT NULL,
    position        INTEGER NOT NULL,
    is_default      BOOLEAN NOT NULL DEFAULT FALSE,
    created         INTEGER NOT NULL,
    modified        INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL
);

CREATE INDEX IF NOT EXISTS custom_field_options_get_by_custom_field_id ON custom_field_option (custom_field_id);
CREATE INDEX IF NOT EXISTS custom_field_options_get_by_value ON custom_field_option (value);
CREATE INDEX IF NOT EXISTS custom_field_options_get_by_position ON custom_field_option (position);
CREATE INDEX IF NOT EXISTS custom_field_options_get_by_deleted ON custom_field_option (deleted);
CREATE INDEX IF NOT EXISTS custom_field_options_get_by_created ON custom_field_option (created);

-- Ticket Custom Field Values
CREATE TABLE IF NOT EXISTS ticket_custom_field_value
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    ticket_id       TEXT    NOT NULL,
    custom_field_id TEXT    NOT NULL,
    value           TEXT,
    created         INTEGER NOT NULL,
    modified        INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL,
    UNIQUE (ticket_id, custom_field_id)
);

CREATE INDEX IF NOT EXISTS ticket_custom_field_values_get_by_ticket_id ON ticket_custom_field_value (ticket_id);
CREATE INDEX IF NOT EXISTS ticket_custom_field_values_get_by_custom_field_id ON ticket_custom_field_value (custom_field_id);
CREATE INDEX IF NOT EXISTS ticket_custom_field_values_get_by_ticket_and_field ON ticket_custom_field_value (ticket_id, custom_field_id);
CREATE INDEX IF NOT EXISTS ticket_custom_field_values_get_by_deleted ON ticket_custom_field_value (deleted);
CREATE INDEX IF NOT EXISTS ticket_custom_field_values_get_by_created ON ticket_custom_field_value (created);
CREATE INDEX IF NOT EXISTS ticket_custom_field_values_get_by_modified ON ticket_custom_field_value (modified);
CREATE INDEX IF NOT EXISTS ticket_custom_field_values_get_by_created_and_modified ON ticket_custom_field_value (created, modified);

/*
    ========================================================================
    STEP 2: Add New Columns to Existing Tables
    ========================================================================
*/

-- Add new columns to ticket table
-- Note: SQLite doesn't support multiple columns in one ALTER TABLE, so we do them separately

-- Priority and Resolution
ALTER TABLE ticket ADD COLUMN priority_id TEXT;
ALTER TABLE ticket ADD COLUMN resolution_id TEXT;

-- Assignee and Reporter
ALTER TABLE ticket ADD COLUMN assignee_id TEXT;
ALTER TABLE ticket ADD COLUMN reporter_id TEXT;

-- Time tracking enhancements
ALTER TABLE ticket ADD COLUMN due_date INTEGER;
ALTER TABLE ticket ADD COLUMN original_estimate INTEGER;  -- In minutes
ALTER TABLE ticket ADD COLUMN remaining_estimate INTEGER; -- In minutes
ALTER TABLE ticket ADD COLUMN time_spent INTEGER;         -- In minutes

-- Add new columns to project table
ALTER TABLE project ADD COLUMN lead_user_id TEXT;
ALTER TABLE project ADD COLUMN default_assignee_id TEXT;

/*
    ========================================================================
    STEP 3: Create Indexes for New Columns
    ========================================================================
*/

CREATE INDEX IF NOT EXISTS tickets_get_by_priority_id ON ticket (priority_id);
CREATE INDEX IF NOT EXISTS tickets_get_by_resolution_id ON ticket (resolution_id);
CREATE INDEX IF NOT EXISTS tickets_get_by_assignee_id ON ticket (assignee_id);
CREATE INDEX IF NOT EXISTS tickets_get_by_reporter_id ON ticket (reporter_id);
CREATE INDEX IF NOT EXISTS tickets_get_by_due_date ON ticket (due_date);

CREATE INDEX IF NOT EXISTS projects_get_by_lead_user_id ON project (lead_user_id);
CREATE INDEX IF NOT EXISTS projects_get_by_default_assignee_id ON project (default_assignee_id);

/*
    ========================================================================
    STEP 4: Insert Default Seed Data
    ========================================================================
*/

-- Default Priorities (if not already present)
INSERT OR IGNORE INTO priority (id, title, description, level, icon, color, created, modified, deleted)
VALUES
    ('priority-lowest', 'Lowest', 'Lowest priority', 1, 'arrow_downward', '#0747A6', strftime('%s', 'now'), strftime('%s', 'now'), 0),
    ('priority-low', 'Low', 'Low priority', 2, 'keyboard_arrow_down', '#2684FF', strftime('%s', 'now'), strftime('%s', 'now'), 0),
    ('priority-medium', 'Medium', 'Medium priority', 3, 'drag_handle', '#FFAB00', strftime('%s', 'now'), strftime('%s', 'now'), 0),
    ('priority-high', 'High', 'High priority', 4, 'keyboard_arrow_up', '#FF8B00', strftime('%s', 'now'), strftime('%s', 'now'), 0),
    ('priority-highest', 'Highest', 'Highest priority', 5, 'arrow_upward', '#DE350B', strftime('%s', 'now'), strftime('%s', 'now'), 0);

-- Default Resolutions (if not already present)
INSERT OR IGNORE INTO resolution (id, title, description, created, modified, deleted)
VALUES
    ('resolution-fixed', 'Fixed', 'A fix for this issue is checked into the tree and tested.', strftime('%s', 'now'), strftime('%s', 'now'), 0),
    ('resolution-wont-fix', 'Won''t Fix', 'The problem described is an issue which will never be fixed.', strftime('%s', 'now'), strftime('%s', 'now'), 0),
    ('resolution-duplicate', 'Duplicate', 'The problem is a duplicate of an existing issue.', strftime('%s', 'now'), strftime('%s', 'now'), 0),
    ('resolution-incomplete', 'Incomplete', 'The problem is not completely described.', strftime('%s', 'now'), strftime('%s', 'now'), 0),
    ('resolution-cannot-reproduce', 'Cannot Reproduce', 'Attempts at reproducing this issue failed, or not enough information was available.', strftime('%s', 'now'), strftime('%s', 'now'), 0),
    ('resolution-done', 'Done', 'Work has been completed on this issue.', strftime('%s', 'now'), strftime('%s', 'now'), 0);

/*
    ========================================================================
    STEP 5: Data Migration (Optional)
    ========================================================================
*/

-- Set default priority for existing tickets (if desired)
-- UPDATE ticket SET priority_id = 'priority-medium' WHERE priority_id IS NULL;

-- Set reporter_id to creator for existing tickets
UPDATE ticket SET reporter_id = creator WHERE reporter_id IS NULL;

-- Migrate custom data from ticket_meta_data to custom_field_value (if needed)
-- This would require custom logic based on your meta_data structure
-- Example:
-- INSERT INTO custom_field (id, field_name, field_type, description, project_id, is_required, created, modified, deleted)
-- SELECT
--     'cf-' || property,
--     property,
--     'text',
--     'Migrated from meta_data',
--     NULL,
--     0,
--     strftime('%s', 'now'),
--     strftime('%s', 'now'),
--     0
-- FROM ticket_meta_data
-- GROUP BY property;

/*
    ========================================================================
    STEP 6: Update System Info
    ========================================================================
*/

-- Update schema version information
UPDATE system_info
SET description = 'Database schema version 2 - JIRA Feature Parity Phase 1',
    modified = strftime('%s', 'now')
WHERE id = (SELECT id FROM system_info ORDER BY created DESC LIMIT 1);

/*
    ========================================================================
    VERIFICATION QUERIES
    ========================================================================

    Run these queries after migration to verify success:

    -- Check new tables exist
    SELECT name FROM sqlite_master WHERE type='table'
    AND name IN ('priority', 'resolution', 'version', 'filter', 'custom_field', 'ticket_watcher_mapping');

    -- Check new columns exist
    PRAGMA table_info(ticket);
    PRAGMA table_info(project);

    -- Check seed data
    SELECT COUNT(*) FROM priority;  -- Should be 5
    SELECT COUNT(*) FROM resolution;  -- Should be 6

    -- Check indexes
    SELECT name FROM sqlite_master WHERE type='index'
    AND name LIKE '%priority%' OR name LIKE '%resolution%';
*/

/*
    ========================================================================
    ROLLBACK PROCEDURE (If needed)
    ========================================================================

    IMPORTANT: Rollback is destructive and will lose data added after migration!

    -- Drop new tables
    DROP TABLE IF EXISTS priority;
    DROP TABLE IF EXISTS resolution;
    DROP TABLE IF EXISTS ticket_watcher_mapping;
    DROP TABLE IF EXISTS version;
    DROP TABLE IF EXISTS ticket_affected_version_mapping;
    DROP TABLE IF EXISTS ticket_fix_version_mapping;
    DROP TABLE IF EXISTS filter;
    DROP TABLE IF EXISTS filter_share_mapping;
    DROP TABLE IF EXISTS custom_field;
    DROP TABLE IF EXISTS custom_field_option;
    DROP TABLE IF EXISTS ticket_custom_field_value;

    -- Remove new columns (requires table recreation in SQLite)
    -- This is complex - better to restore from backup!

    -- Restore system_info
    UPDATE system_info
    SET description = 'Database schema version 1',
        modified = strftime('%s', 'now')
    WHERE id = (SELECT id FROM system_info ORDER BY created DESC LIMIT 1);
*/

/*
    ========================================================================
    MIGRATION COMPLETE
    ========================================================================

    Migration from V1 to V2 complete!

    Next steps:
    1. Verify all tables and columns exist
    2. Test application with new schema
    3. Update application code to use new features
    4. Update API documentation
    5. Train users on new features

    For PostgreSQL:
    - Replace INTEGER with BIGINT for timestamps
    - Replace BOOLEAN with BOOLEAN (native type)
    - Adjust strftime to EXTRACT(EPOCH FROM NOW())
    - Remove IF NOT EXISTS if not supported
*/

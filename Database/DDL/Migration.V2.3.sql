/*
    Migration: V2 to V3
    Version: 2.3
    Date: 2025-10-12

    This migration script upgrades an existing V2 database to V3 schema.
    It adds Phase 2 (Agile Enhancements) and Phase 3 (Collaboration Features)
    for complete JIRA feature parity.

    IMPORTANT: Backup your database before running this migration!

    Phase 2 Features Added:
    - Epic Support
    - Subtask Support
    - Enhanced Work Logs
    - Project Roles
    - Security Levels
    - Dashboard System
    - Advanced Board Configuration

    Phase 3 Features Added:
    - Voting System
    - Project Categories
    - Notification Schemes
    - Activity Stream Enhancements
    - Comment Mentions
*/

/*
    ========================================================================
    STEP 1: Create New Tables - Phase 2
    ========================================================================
*/

-- Enhanced Work Logs
CREATE TABLE IF NOT EXISTS work_log
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    ticket_id   TEXT    NOT NULL,
    user_id     TEXT    NOT NULL,
    time_spent  INTEGER NOT NULL,  -- Time in minutes
    work_date   INTEGER NOT NULL,  -- Unix timestamp of work date
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL
);

CREATE INDEX IF NOT EXISTS work_logs_get_by_ticket_id ON work_log (ticket_id);
CREATE INDEX IF NOT EXISTS work_logs_get_by_user_id ON work_log (user_id);
CREATE INDEX IF NOT EXISTS work_logs_get_by_work_date ON work_log (work_date);
CREATE INDEX IF NOT EXISTS work_logs_get_by_created ON work_log (created);
CREATE INDEX IF NOT EXISTS work_logs_get_by_deleted ON work_log (deleted);
CREATE INDEX IF NOT EXISTS work_logs_get_by_created_and_modified ON work_log (created, modified);

-- Project Roles
CREATE TABLE IF NOT EXISTS project_role
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL,
    description TEXT,
    project_id  TEXT,              -- NULL for global roles
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL
);

CREATE INDEX IF NOT EXISTS project_roles_get_by_title ON project_role (title);
CREATE INDEX IF NOT EXISTS project_roles_get_by_project_id ON project_role (project_id);
CREATE INDEX IF NOT EXISTS project_roles_get_by_deleted ON project_role (deleted);
CREATE INDEX IF NOT EXISTS project_roles_get_by_created ON project_role (created);
CREATE INDEX IF NOT EXISTS project_roles_get_by_modified ON project_role (modified);
CREATE INDEX IF NOT EXISTS project_roles_get_by_created_and_modified ON project_role (created, modified);

CREATE TABLE IF NOT EXISTS project_role_user_mapping
(
    id               TEXT    NOT NULL PRIMARY KEY UNIQUE,
    project_role_id  TEXT    NOT NULL,
    project_id       TEXT    NOT NULL,
    user_id          TEXT    NOT NULL,
    created          INTEGER NOT NULL,
    deleted          BOOLEAN NOT NULL,
    UNIQUE (project_role_id, project_id, user_id)
);

CREATE INDEX IF NOT EXISTS project_role_users_get_by_role_id ON project_role_user_mapping (project_role_id);
CREATE INDEX IF NOT EXISTS project_role_users_get_by_project_id ON project_role_user_mapping (project_id);
CREATE INDEX IF NOT EXISTS project_role_users_get_by_user_id ON project_role_user_mapping (user_id);
CREATE INDEX IF NOT EXISTS project_role_users_get_by_deleted ON project_role_user_mapping (deleted);
CREATE INDEX IF NOT EXISTS project_role_users_get_by_created ON project_role_user_mapping (created);

-- Security Levels
CREATE TABLE IF NOT EXISTS security_level
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL,
    description TEXT,
    project_id  TEXT    NOT NULL,
    level       INTEGER NOT NULL,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL
);

CREATE INDEX IF NOT EXISTS security_levels_get_by_title ON security_level (title);
CREATE INDEX IF NOT EXISTS security_levels_get_by_project_id ON security_level (project_id);
CREATE INDEX IF NOT EXISTS security_levels_get_by_level ON security_level (level);
CREATE INDEX IF NOT EXISTS security_levels_get_by_deleted ON security_level (deleted);
CREATE INDEX IF NOT EXISTS security_levels_get_by_created ON security_level (created);
CREATE INDEX IF NOT EXISTS security_levels_get_by_modified ON security_level (modified);
CREATE INDEX IF NOT EXISTS security_levels_get_by_created_and_modified ON security_level (created, modified);

CREATE TABLE IF NOT EXISTS security_level_permission_mapping
(
    id                TEXT    NOT NULL PRIMARY KEY UNIQUE,
    security_level_id TEXT    NOT NULL,
    user_id           TEXT,
    team_id           TEXT,
    project_role_id   TEXT,
    created           INTEGER NOT NULL,
    deleted           BOOLEAN NOT NULL
);

CREATE INDEX IF NOT EXISTS security_level_permissions_get_by_security_level_id ON security_level_permission_mapping (security_level_id);
CREATE INDEX IF NOT EXISTS security_level_permissions_get_by_user_id ON security_level_permission_mapping (user_id);
CREATE INDEX IF NOT EXISTS security_level_permissions_get_by_team_id ON security_level_permission_mapping (team_id);
CREATE INDEX IF NOT EXISTS security_level_permissions_get_by_project_role_id ON security_level_permission_mapping (project_role_id);
CREATE INDEX IF NOT EXISTS security_level_permissions_get_by_deleted ON security_level_permission_mapping (deleted);
CREATE INDEX IF NOT EXISTS security_level_permissions_get_by_created ON security_level_permission_mapping (created);

-- Dashboard System
CREATE TABLE IF NOT EXISTS dashboard
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL,
    description TEXT,
    owner_id    TEXT    NOT NULL,
    is_public   BOOLEAN NOT NULL DEFAULT FALSE,
    is_favorite BOOLEAN NOT NULL DEFAULT FALSE,
    layout      TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL
);

CREATE INDEX IF NOT EXISTS dashboards_get_by_title ON dashboard (title);
CREATE INDEX IF NOT EXISTS dashboards_get_by_owner_id ON dashboard (owner_id);
CREATE INDEX IF NOT EXISTS dashboards_get_by_is_public ON dashboard (is_public);
CREATE INDEX IF NOT EXISTS dashboards_get_by_is_favorite ON dashboard (is_favorite);
CREATE INDEX IF NOT EXISTS dashboards_get_by_deleted ON dashboard (deleted);
CREATE INDEX IF NOT EXISTS dashboards_get_by_created ON dashboard (created);
CREATE INDEX IF NOT EXISTS dashboards_get_by_modified ON dashboard (modified);
CREATE INDEX IF NOT EXISTS dashboards_get_by_created_and_modified ON dashboard (created, modified);

CREATE TABLE IF NOT EXISTS dashboard_widget
(
    id            TEXT    NOT NULL PRIMARY KEY UNIQUE,
    dashboard_id  TEXT    NOT NULL,
    widget_type   TEXT    NOT NULL,
    title         TEXT,
    position_x    INTEGER,
    position_y    INTEGER,
    width         INTEGER,
    height        INTEGER,
    configuration TEXT,
    created       INTEGER NOT NULL,
    modified      INTEGER NOT NULL,
    deleted       BOOLEAN NOT NULL
);

CREATE INDEX IF NOT EXISTS dashboard_widgets_get_by_dashboard_id ON dashboard_widget (dashboard_id);
CREATE INDEX IF NOT EXISTS dashboard_widgets_get_by_widget_type ON dashboard_widget (widget_type);
CREATE INDEX IF NOT EXISTS dashboard_widgets_get_by_deleted ON dashboard_widget (deleted);
CREATE INDEX IF NOT EXISTS dashboard_widgets_get_by_created ON dashboard_widget (created);
CREATE INDEX IF NOT EXISTS dashboard_widgets_get_by_modified ON dashboard_widget (modified);

CREATE TABLE IF NOT EXISTS dashboard_share_mapping
(
    id           TEXT    NOT NULL PRIMARY KEY UNIQUE,
    dashboard_id TEXT    NOT NULL,
    user_id      TEXT,
    team_id      TEXT,
    project_id   TEXT,
    created      INTEGER NOT NULL,
    deleted      BOOLEAN NOT NULL
);

CREATE INDEX IF NOT EXISTS dashboard_shares_get_by_dashboard_id ON dashboard_share_mapping (dashboard_id);
CREATE INDEX IF NOT EXISTS dashboard_shares_get_by_user_id ON dashboard_share_mapping (user_id);
CREATE INDEX IF NOT EXISTS dashboard_shares_get_by_team_id ON dashboard_share_mapping (team_id);
CREATE INDEX IF NOT EXISTS dashboard_shares_get_by_project_id ON dashboard_share_mapping (project_id);
CREATE INDEX IF NOT EXISTS dashboard_shares_get_by_deleted ON dashboard_share_mapping (deleted);
CREATE INDEX IF NOT EXISTS dashboard_shares_get_by_created ON dashboard_share_mapping (created);

-- Advanced Board Configuration
CREATE TABLE IF NOT EXISTS board_column
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    board_id    TEXT    NOT NULL,
    title       TEXT    NOT NULL,
    status_id   TEXT,
    position    INTEGER NOT NULL,
    max_items   INTEGER,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL
);

CREATE INDEX IF NOT EXISTS board_columns_get_by_board_id ON board_column (board_id);
CREATE INDEX IF NOT EXISTS board_columns_get_by_status_id ON board_column (status_id);
CREATE INDEX IF NOT EXISTS board_columns_get_by_position ON board_column (position);
CREATE INDEX IF NOT EXISTS board_columns_get_by_deleted ON board_column (deleted);
CREATE INDEX IF NOT EXISTS board_columns_get_by_created ON board_column (created);
CREATE INDEX IF NOT EXISTS board_columns_get_by_modified ON board_column (modified);

CREATE TABLE IF NOT EXISTS board_swimlane
(
    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    board_id TEXT    NOT NULL,
    title    TEXT    NOT NULL,
    query    TEXT,
    position INTEGER NOT NULL,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL
);

CREATE INDEX IF NOT EXISTS board_swimlanes_get_by_board_id ON board_swimlane (board_id);
CREATE INDEX IF NOT EXISTS board_swimlanes_get_by_position ON board_swimlane (position);
CREATE INDEX IF NOT EXISTS board_swimlanes_get_by_deleted ON board_swimlane (deleted);
CREATE INDEX IF NOT EXISTS board_swimlanes_get_by_created ON board_swimlane (created);
CREATE INDEX IF NOT EXISTS board_swimlanes_get_by_modified ON board_swimlane (modified);

CREATE TABLE IF NOT EXISTS board_quick_filter
(
    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    board_id TEXT    NOT NULL,
    title    TEXT    NOT NULL,
    query    TEXT,
    position INTEGER NOT NULL,
    created  INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL
);

CREATE INDEX IF NOT EXISTS board_quick_filters_get_by_board_id ON board_quick_filter (board_id);
CREATE INDEX IF NOT EXISTS board_quick_filters_get_by_position ON board_quick_filter (position);
CREATE INDEX IF NOT EXISTS board_quick_filters_get_by_deleted ON board_quick_filter (deleted);
CREATE INDEX IF NOT EXISTS board_quick_filters_get_by_created ON board_quick_filter (created);

/*
    ========================================================================
    STEP 2: Create New Tables - Phase 3
    ========================================================================
*/

-- Voting System
CREATE TABLE IF NOT EXISTS ticket_vote_mapping
(
    id        TEXT    NOT NULL PRIMARY KEY UNIQUE,
    ticket_id TEXT    NOT NULL,
    user_id   TEXT    NOT NULL,
    created   INTEGER NOT NULL,
    deleted   BOOLEAN NOT NULL,
    UNIQUE (ticket_id, user_id)
);

CREATE INDEX IF NOT EXISTS ticket_votes_get_by_ticket_id ON ticket_vote_mapping (ticket_id);
CREATE INDEX IF NOT EXISTS ticket_votes_get_by_user_id ON ticket_vote_mapping (user_id);
CREATE INDEX IF NOT EXISTS ticket_votes_get_by_ticket_and_user ON ticket_vote_mapping (ticket_id, user_id);
CREATE INDEX IF NOT EXISTS ticket_votes_get_by_deleted ON ticket_vote_mapping (deleted);
CREATE INDEX IF NOT EXISTS ticket_votes_get_by_created ON ticket_vote_mapping (created);

-- Project Categories
CREATE TABLE IF NOT EXISTS project_category
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL
);

CREATE INDEX IF NOT EXISTS project_categories_get_by_title ON project_category (title);
CREATE INDEX IF NOT EXISTS project_categories_get_by_deleted ON project_category (deleted);
CREATE INDEX IF NOT EXISTS project_categories_get_by_created ON project_category (created);
CREATE INDEX IF NOT EXISTS project_categories_get_by_modified ON project_category (modified);
CREATE INDEX IF NOT EXISTS project_categories_get_by_created_and_modified ON project_category (created, modified);

-- Notification Schemes
CREATE TABLE IF NOT EXISTS notification_scheme
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL,
    description TEXT,
    project_id  TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL
);

CREATE INDEX IF NOT EXISTS notification_schemes_get_by_title ON notification_scheme (title);
CREATE INDEX IF NOT EXISTS notification_schemes_get_by_project_id ON notification_scheme (project_id);
CREATE INDEX IF NOT EXISTS notification_schemes_get_by_deleted ON notification_scheme (deleted);
CREATE INDEX IF NOT EXISTS notification_schemes_get_by_created ON notification_scheme (created);
CREATE INDEX IF NOT EXISTS notification_schemes_get_by_modified ON notification_scheme (modified);
CREATE INDEX IF NOT EXISTS notification_schemes_get_by_created_and_modified ON notification_scheme (created, modified);

CREATE TABLE IF NOT EXISTS notification_event
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    event_type  TEXT    NOT NULL,
    title       TEXT    NOT NULL,
    description TEXT,
    created     INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL
);

CREATE INDEX IF NOT EXISTS notification_events_get_by_event_type ON notification_event (event_type);
CREATE INDEX IF NOT EXISTS notification_events_get_by_title ON notification_event (title);
CREATE INDEX IF NOT EXISTS notification_events_get_by_deleted ON notification_event (deleted);
CREATE INDEX IF NOT EXISTS notification_events_get_by_created ON notification_event (created);

CREATE TABLE IF NOT EXISTS notification_rule
(
    id                      TEXT    NOT NULL PRIMARY KEY UNIQUE,
    notification_scheme_id  TEXT    NOT NULL,
    notification_event_id   TEXT    NOT NULL,
    recipient_type          TEXT    NOT NULL,
    recipient_id            TEXT,
    created                 INTEGER NOT NULL,
    deleted                 BOOLEAN NOT NULL
);

CREATE INDEX IF NOT EXISTS notification_rules_get_by_scheme_id ON notification_rule (notification_scheme_id);
CREATE INDEX IF NOT EXISTS notification_rules_get_by_event_id ON notification_rule (notification_event_id);
CREATE INDEX IF NOT EXISTS notification_rules_get_by_recipient_type ON notification_rule (recipient_type);
CREATE INDEX IF NOT EXISTS notification_rules_get_by_recipient_id ON notification_rule (recipient_id);
CREATE INDEX IF NOT EXISTS notification_rules_get_by_deleted ON notification_rule (deleted);
CREATE INDEX IF NOT EXISTS notification_rules_get_by_created ON notification_rule (created);

-- Comment Mentions
CREATE TABLE IF NOT EXISTS comment_mention_mapping
(
    id                 TEXT    NOT NULL PRIMARY KEY UNIQUE,
    comment_id         TEXT    NOT NULL,
    mentioned_user_id  TEXT    NOT NULL,
    created            INTEGER NOT NULL,
    deleted            BOOLEAN NOT NULL
);

CREATE INDEX IF NOT EXISTS comment_mentions_get_by_comment_id ON comment_mention_mapping (comment_id);
CREATE INDEX IF NOT EXISTS comment_mentions_get_by_user_id ON comment_mention_mapping (mentioned_user_id);
CREATE INDEX IF NOT EXISTS comment_mentions_get_by_deleted ON comment_mention_mapping (deleted);
CREATE INDEX IF NOT EXISTS comment_mentions_get_by_created ON comment_mention_mapping (created);

/*
    ========================================================================
    STEP 3: Add New Columns to Existing Tables - Phase 2
    ========================================================================
*/

-- Epic Support (ticket table enhancements)
ALTER TABLE ticket ADD COLUMN is_epic BOOLEAN DEFAULT FALSE;
ALTER TABLE ticket ADD COLUMN epic_id TEXT;
ALTER TABLE ticket ADD COLUMN epic_color TEXT;
ALTER TABLE ticket ADD COLUMN epic_name TEXT;

-- Subtask Support (ticket table enhancements)
ALTER TABLE ticket ADD COLUMN is_subtask BOOLEAN DEFAULT FALSE;
ALTER TABLE ticket ADD COLUMN parent_ticket_id TEXT;

-- Security Levels (ticket table enhancement)
ALTER TABLE ticket ADD COLUMN security_level_id TEXT;

-- Advanced Board Configuration (board table enhancements)
ALTER TABLE board ADD COLUMN filter_id TEXT;
ALTER TABLE board ADD COLUMN board_type TEXT;

/*
    ========================================================================
    STEP 4: Add New Columns to Existing Tables - Phase 3
    ========================================================================
*/

-- Voting System (ticket table enhancement)
ALTER TABLE ticket ADD COLUMN vote_count INTEGER DEFAULT 0;

-- Project Categories (project table enhancement)
ALTER TABLE project ADD COLUMN project_category_id TEXT;

-- Activity Stream Enhancements (audit table enhancements)
ALTER TABLE audit ADD COLUMN is_public BOOLEAN DEFAULT TRUE;
ALTER TABLE audit ADD COLUMN activity_type TEXT;

/*
    ========================================================================
    STEP 5: Create Indexes for New Columns
    ========================================================================
*/

-- Phase 2 indexes
CREATE INDEX IF NOT EXISTS tickets_get_by_is_epic ON ticket (is_epic);
CREATE INDEX IF NOT EXISTS tickets_get_by_epic_id ON ticket (epic_id);
CREATE INDEX IF NOT EXISTS tickets_get_by_is_subtask ON ticket (is_subtask);
CREATE INDEX IF NOT EXISTS tickets_get_by_parent_ticket_id ON ticket (parent_ticket_id);
CREATE INDEX IF NOT EXISTS tickets_get_by_security_level_id ON ticket (security_level_id);
CREATE INDEX IF NOT EXISTS boards_get_by_filter_id ON board (filter_id);
CREATE INDEX IF NOT EXISTS boards_get_by_board_type ON board (board_type);

-- Phase 3 indexes
CREATE INDEX IF NOT EXISTS tickets_get_by_vote_count ON ticket (vote_count);
CREATE INDEX IF NOT EXISTS projects_get_by_category_id ON project (project_category_id);
CREATE INDEX IF NOT EXISTS audit_get_by_is_public ON audit (is_public);
CREATE INDEX IF NOT EXISTS audit_get_by_activity_type ON audit (activity_type);

/*
    ========================================================================
    STEP 6: Insert Default Seed Data
    ========================================================================
*/

-- Default Notification Events
INSERT OR IGNORE INTO notification_event (id, event_type, title, description, created, deleted)
VALUES
    ('event-issue-created', 'issue_created', 'Issue Created', 'Triggered when a new issue is created', strftime('%s', 'now'), 0),
    ('event-issue-updated', 'issue_updated', 'Issue Updated', 'Triggered when an issue is updated', strftime('%s', 'now'), 0),
    ('event-issue-deleted', 'issue_deleted', 'Issue Deleted', 'Triggered when an issue is deleted', strftime('%s', 'now'), 0),
    ('event-comment-added', 'comment_added', 'Comment Added', 'Triggered when a comment is added', strftime('%s', 'now'), 0),
    ('event-comment-updated', 'comment_updated', 'Comment Updated', 'Triggered when a comment is updated', strftime('%s', 'now'), 0),
    ('event-comment-deleted', 'comment_deleted', 'Comment Deleted', 'Triggered when a comment is deleted', strftime('%s', 'now'), 0),
    ('event-status-changed', 'status_changed', 'Status Changed', 'Triggered when issue status changes', strftime('%s', 'now'), 0),
    ('event-assignee-changed', 'assignee_changed', 'Assignee Changed', 'Triggered when assignee changes', strftime('%s', 'now'), 0),
    ('event-priority-changed', 'priority_changed', 'Priority Changed', 'Triggered when priority changes', strftime('%s', 'now'), 0),
    ('event-work-logged', 'work_logged', 'Work Logged', 'Triggered when work is logged', strftime('%s', 'now'), 0),
    ('event-user-mentioned', 'user_mentioned', 'User Mentioned', 'Triggered when user is @mentioned', strftime('%s', 'now'), 0);

/*
    ========================================================================
    STEP 7: Data Migration (Optional)
    ========================================================================
*/

-- Initialize vote_count for existing tickets
UPDATE ticket SET vote_count = 0 WHERE vote_count IS NULL;

-- Initialize is_epic and is_subtask for existing tickets
UPDATE ticket SET is_epic = FALSE WHERE is_epic IS NULL;
UPDATE ticket SET is_subtask = FALSE WHERE is_subtask IS NULL;

-- Initialize is_public for existing audit records
UPDATE audit SET is_public = TRUE WHERE is_public IS NULL;

/*
    ========================================================================
    STEP 8: Update System Info
    ========================================================================
*/

-- Update schema version information
UPDATE system_info
SET description = 'Database schema version 3 - Complete JIRA Feature Parity (Phases 2 & 3)'
WHERE id = (SELECT id FROM system_info ORDER BY created DESC LIMIT 1);

/*
    ========================================================================
    VERIFICATION QUERIES
    ========================================================================

    Run these queries after migration to verify success:

    -- Check all new Phase 2 tables exist
    SELECT name FROM sqlite_master WHERE type='table'
    AND name IN ('work_log', 'project_role', 'project_role_user_mapping',
                 'security_level', 'security_level_permission_mapping',
                 'dashboard', 'dashboard_widget', 'dashboard_share_mapping',
                 'board_column', 'board_swimlane', 'board_quick_filter');

    -- Check all new Phase 3 tables exist
    SELECT name FROM sqlite_master WHERE type='table'
    AND name IN ('ticket_vote_mapping', 'project_category',
                 'notification_scheme', 'notification_event', 'notification_rule',
                 'comment_mention_mapping');

    -- Check new columns in ticket table
    PRAGMA table_info(ticket);
    -- Should show: is_epic, epic_id, epic_color, epic_name, is_subtask,
    --              parent_ticket_id, security_level_id, vote_count

    -- Check new columns in board table
    PRAGMA table_info(board);
    -- Should show: filter_id, board_type

    -- Check new columns in project table
    PRAGMA table_info(project);
    -- Should show: project_category_id

    -- Check new columns in audit table
    PRAGMA table_info(audit);
    -- Should show: is_public, activity_type

    -- Check seed data
    SELECT COUNT(*) FROM notification_event;  -- Should be 11

    -- Check indexes (sample)
    SELECT name FROM sqlite_master WHERE type='index'
    AND (name LIKE '%epic%' OR name LIKE '%subtask%' OR name LIKE '%vote%');

    -- Verify data migration
    SELECT COUNT(*) FROM ticket WHERE is_epic = FALSE;
    SELECT COUNT(*) FROM ticket WHERE is_subtask = FALSE;
    SELECT COUNT(*) FROM ticket WHERE vote_count = 0;
    SELECT COUNT(*) FROM audit WHERE is_public = TRUE;
*/

/*
    ========================================================================
    ROLLBACK PROCEDURE (If needed)
    ========================================================================

    IMPORTANT: Rollback is destructive and will lose data added after migration!

    -- Drop new Phase 2 tables
    DROP TABLE IF EXISTS work_log;
    DROP TABLE IF EXISTS project_role;
    DROP TABLE IF EXISTS project_role_user_mapping;
    DROP TABLE IF EXISTS security_level;
    DROP TABLE IF EXISTS security_level_permission_mapping;
    DROP TABLE IF EXISTS dashboard;
    DROP TABLE IF EXISTS dashboard_widget;
    DROP TABLE IF EXISTS dashboard_share_mapping;
    DROP TABLE IF EXISTS board_column;
    DROP TABLE IF EXISTS board_swimlane;
    DROP TABLE IF EXISTS board_quick_filter;

    -- Drop new Phase 3 tables
    DROP TABLE IF EXISTS ticket_vote_mapping;
    DROP TABLE IF EXISTS project_category;
    DROP TABLE IF EXISTS notification_scheme;
    DROP TABLE IF EXISTS notification_event;
    DROP TABLE IF EXISTS notification_rule;
    DROP TABLE IF EXISTS comment_mention_mapping;

    -- Remove new columns (requires table recreation in SQLite)
    -- This is complex - better to restore from backup!

    -- Restore system_info
    UPDATE system_info
    SET description = 'Database schema version 2 - JIRA Feature Parity Phase 1',
        modified = strftime('%s', 'now')
    WHERE id = (SELECT id FROM system_info ORDER BY created DESC LIMIT 1);
*/

/*
    ========================================================================
    MIGRATION COMPLETE
    ========================================================================

    Migration from V2 to V3 complete!

    Summary of Changes:
    - Added 17 new tables (11 Phase 2, 6 Phase 3)
    - Enhanced 4 existing tables (ticket, board, project, audit)
    - Added 13 new columns across existing tables
    - Created 50+ new indexes
    - Inserted 11 default notification events

    Next steps:
    1. Verify all tables and columns exist (run verification queries above)
    2. Test application with new schema
    3. Update application code to use new features:
       - Implement Epic support handlers
       - Implement Subtask support handlers
       - Implement Work Log handlers
       - Implement Project Role handlers
       - Implement Security Level handlers
       - Implement Dashboard handlers
       - Implement Advanced Board handlers
       - Implement Voting handlers
       - Implement Project Category handlers
       - Implement Notification handlers
       - Implement Activity Stream enhancements
       - Implement Mention handlers
    4. Update API documentation with ~85 new endpoints
    5. Write comprehensive tests (~255 new tests)
    6. Train users on new features

    For PostgreSQL:
    - Replace INTEGER with BIGINT for timestamps
    - Replace BOOLEAN with BOOLEAN (native type)
    - Adjust strftime('%s', 'now') to EXTRACT(EPOCH FROM NOW())::BIGINT
    - Remove IF NOT EXISTS if not supported
    - Adjust UNIQUE constraints if needed
*/

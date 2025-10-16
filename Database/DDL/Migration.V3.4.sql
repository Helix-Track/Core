/*
    Migration: V3 to V4

    Parallel Editing Support Migration

    This migration adds support for parallel editing of entities with:
    - Optimistic locking using version numbers
    - Change history tracking for all editable entities
    - Conflict resolution mechanisms
    - Real-time collaboration features

    Run this migration after backing up your database.
    Expected downtime: Minimal (schema changes only)
*/

/*
    ========================================================================
    PHASE 1: ADD VERSION COLUMNS
    ========================================================================
*/

-- Add version columns to all editable entities
ALTER TABLE ticket ADD COLUMN version INTEGER DEFAULT 1;
ALTER TABLE project ADD COLUMN version INTEGER DEFAULT 1;
ALTER TABLE comment ADD COLUMN version INTEGER DEFAULT 1;
ALTER TABLE dashboard ADD COLUMN version INTEGER DEFAULT 1;
ALTER TABLE board ADD COLUMN version INTEGER DEFAULT 1;

-- Create indexes for version columns
CREATE INDEX tickets_get_by_version ON ticket (version);
CREATE INDEX projects_get_by_version ON project (version);
CREATE INDEX comments_get_by_version ON comment (version);
CREATE INDEX dashboards_get_by_version ON dashboard (version);
CREATE INDEX boards_get_by_version ON board (version);

/*
    ========================================================================
    PHASE 2: CREATE HISTORY TABLES
    ========================================================================
*/

-- Ticket history table
CREATE TABLE ticket_history
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    ticket_id       TEXT    NOT NULL,
    version         INTEGER NOT NULL,
    action          TEXT    NOT NULL,  -- 'create', 'update', 'delete'
    user_id         TEXT    NOT NULL,
    timestamp       INTEGER NOT NULL,
    old_data        TEXT,              -- JSON snapshot of previous state
    new_data        TEXT,              -- JSON snapshot of new state
    change_summary  TEXT,              -- Human-readable summary of changes
    conflict_data   TEXT,              -- JSON data for conflict resolution
    INDEX (ticket_id, version),
    INDEX (ticket_id, timestamp),
    INDEX (user_id, timestamp)
);

-- Project history table
CREATE TABLE project_history
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    project_id      TEXT    NOT NULL,
    version         INTEGER NOT NULL,
    action          TEXT    NOT NULL,
    user_id         TEXT    NOT NULL,
    timestamp       INTEGER NOT NULL,
    old_data        TEXT,
    new_data        TEXT,
    change_summary  TEXT,
    conflict_data   TEXT,
    INDEX (project_id, version),
    INDEX (project_id, timestamp),
    INDEX (user_id, timestamp)
);

-- Comment history table
CREATE TABLE comment_history
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    comment_id      TEXT    NOT NULL,
    version         INTEGER NOT NULL,
    action          TEXT    NOT NULL,
    user_id         TEXT    NOT NULL,
    timestamp       INTEGER NOT NULL,
    old_data        TEXT,
    new_data        TEXT,
    change_summary  TEXT,
    conflict_data   TEXT,
    INDEX (comment_id, version),
    INDEX (comment_id, timestamp),
    INDEX (user_id, timestamp)
);

-- Dashboard history table
CREATE TABLE dashboard_history
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    dashboard_id    TEXT    NOT NULL,
    version         INTEGER NOT NULL,
    action          TEXT    NOT NULL,
    user_id         TEXT    NOT NULL,
    timestamp       INTEGER NOT NULL,
    old_data        TEXT,
    new_data        TEXT,
    change_summary  TEXT,
    conflict_data   TEXT,
    INDEX (dashboard_id, version),
    INDEX (dashboard_id, timestamp),
    INDEX (user_id, timestamp)
);

-- Board history table
CREATE TABLE board_history
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    board_id        TEXT    NOT NULL,
    version         INTEGER NOT NULL,
    action          TEXT    NOT NULL,
    user_id         TEXT    NOT NULL,
    timestamp       INTEGER NOT NULL,
    old_data        TEXT,
    new_data        TEXT,
    change_summary  TEXT,
    conflict_data   TEXT,
    INDEX (board_id, version),
    INDEX (board_id, timestamp),
    INDEX (user_id, timestamp)
);

/*
    ========================================================================
    PHASE 3: CREATE ENTITY LOCK TABLE
    ========================================================================
*/

CREATE TABLE entity_lock
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    entity_type     TEXT    NOT NULL,  -- 'ticket', 'project', 'comment', etc.
    entity_id       TEXT    NOT NULL,
    user_id         TEXT    NOT NULL,
    lock_type       TEXT    NOT NULL,  -- 'optimistic', 'pessimistic'
    acquired_at     INTEGER NOT NULL,
    expires_at      INTEGER,           -- NULL for optimistic locks
    metadata        TEXT,              -- JSON metadata (client info, etc.)
    UNIQUE (entity_type, entity_id, user_id),
    INDEX (entity_type, entity_id),
    INDEX (user_id, acquired_at),
    INDEX (expires_at)
);

/*
    ========================================================================
    PHASE 4: POPULATE INITIAL HISTORY
    ========================================================================

    Create initial history entries for all existing entities.
    This provides a baseline for change tracking.
*/

-- Insert initial ticket history entries
INSERT INTO ticket_history (id, ticket_id, version, action, user_id, timestamp, new_data, change_summary)
SELECT
    'hist-ticket-' || t.id || '-1',
    t.id,
    1,
    'create',
    t.creator,
    t.created,
    json_object(
        'id', t.id,
        'ticket_number', t.ticket_number,
        'title', t.title,
        'description', t.description,
        'ticket_type_id', t.ticket_type_id,
        'ticket_status_id', t.ticket_status_id,
        'project_id', t.project_id,
        'user_id', t.user_id,
        'estimation', t.estimation,
        'story_points', t.story_points,
        'creator', t.creator,
        'created', t.created,
        'modified', t.modified,
        'deleted', t.deleted
    ),
    'Initial ticket creation'
FROM ticket t
WHERE t.deleted = 0;

-- Insert initial project history entries
INSERT INTO project_history (id, project_id, version, action, user_id, timestamp, new_data, change_summary)
SELECT
    'hist-project-' || p.id || '-1',
    p.id,
    1,
    'create',
    COALESCE(p.owner_id, 'system'),
    p.created,
    json_object(
        'id', p.id,
        'title', p.title,
        'description', p.description,
        'owner_id', p.owner_id,
        'created', p.created,
        'modified', p.modified,
        'deleted', p.deleted
    ),
    'Initial project creation'
FROM project p
WHERE p.deleted = 0;

-- Insert initial comment history entries
INSERT INTO comment_history (id, comment_id, version, action, user_id, timestamp, new_data, change_summary)
SELECT
    'hist-comment-' || c.id || '-1',
    c.id,
    1,
    'create',
    c.user_id,
    c.created,
    json_object(
        'id', c.id,
        'ticket_id', c.ticket_id,
        'user_id', c.user_id,
        'content', c.content,
        'created', c.created,
        'modified', c.modified,
        'deleted', c.deleted
    ),
    'Initial comment creation'
FROM comment c
WHERE c.deleted = 0;

-- Insert initial dashboard history entries
INSERT INTO dashboard_history (id, dashboard_id, version, action, user_id, timestamp, new_data, change_summary)
SELECT
    'hist-dashboard-' || d.id || '-1',
    d.id,
    1,
    'create',
    d.owner_id,
    d.created,
    json_object(
        'id', d.id,
        'title', d.title,
        'description', d.description,
        'owner_id', d.owner_id,
        'is_public', d.is_public,
        'is_favorite', d.is_favorite,
        'layout', d.layout,
        'created', d.created,
        'modified', d.modified,
        'deleted', d.deleted
    ),
    'Initial dashboard creation'
FROM dashboard d
WHERE d.deleted = 0;

-- Insert initial board history entries
INSERT INTO board_history (id, board_id, version, action, user_id, timestamp, new_data, change_summary)
SELECT
    'hist-board-' || b.id || '-1',
    b.id,
    1,
    'create',
    COALESCE(b.owner_id, 'system'),
    b.created,
    json_object(
        'id', b.id,
        'title', b.title,
        'description', b.description,
        'project_id', b.project_id,
        'owner_id', b.owner_id,
        'board_type', b.board_type,
        'filter_id', b.filter_id,
        'created', b.created,
        'modified', b.modified,
        'deleted', b.deleted
    ),
    'Initial board creation'
FROM board b
WHERE b.deleted = 0;

/*
    ========================================================================
    PHASE 5: UPDATE SYSTEM VERSION
    ========================================================================
*/

UPDATE system_info SET
    description = 'Database schema version 4 - Parallel Editing Support',
    modified = strftime('%s', 'now')
WHERE id = (SELECT id FROM system_info ORDER BY created DESC LIMIT 1);

/*
    ========================================================================
    MIGRATION VERIFICATION
    ========================================================================

    After running this migration, verify:

    1. All tables have version columns with default value 1
    2. All history tables exist and have initial entries
    3. Entity lock table exists
    4. System version updated to 4
    5. No data loss occurred

    Run these queries to verify:

    -- Check version columns
    SELECT 'ticket' as table_name, COUNT(*) as count FROM ticket WHERE version IS NULL;
    SELECT 'project' as table_name, COUNT(*) as count FROM project WHERE version IS NULL;
    SELECT 'comment' as table_name, COUNT(*) as count FROM comment WHERE version IS NULL;
    SELECT 'dashboard' as table_name, COUNT(*) as count FROM dashboard WHERE version IS NULL;
    SELECT 'board' as table_name, COUNT(*) as count FROM board WHERE version IS NULL;

    -- Check history tables
    SELECT 'ticket_history' as table_name, COUNT(*) as count FROM ticket_history;
    SELECT 'project_history' as table_name, COUNT(*) as count FROM project_history;
    SELECT 'comment_history' as table_name, COUNT(*) as count FROM comment_history;
    SELECT 'dashboard_history' as table_name, COUNT(*) as count FROM dashboard_history;
    SELECT 'board_history' as table_name, COUNT(*) as count FROM board_history;

    -- Check entity locks table
    SELECT 'entity_lock' as table_name, COUNT(*) as count FROM entity_lock;

    All counts should be 0 for NULL versions and > 0 for history tables.
*/
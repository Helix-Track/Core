/*
    Version: 4

    Parallel Editing Support

    This version adds support for parallel editing of entities with:
    - Optimistic locking using version numbers
    - Change history tracking for all editable entities
    - Conflict resolution mechanisms
    - Real-time collaboration features

    Features:
    - Version columns added to all editable entities
    - History tables for tracking all changes
    - Enhanced audit logging for collaborative actions
    - Conflict detection and resolution
*/

/*
    ========================================================================
    DROP STATEMENTS - Version 4 Tables
    ========================================================================
*/

-- History tables
DROP TABLE IF EXISTS ticket_history;
DROP TABLE IF EXISTS project_history;
DROP TABLE IF EXISTS comment_history;
DROP TABLE IF EXISTS dashboard_history;
DROP TABLE IF EXISTS board_history;

-- Entity locks (for future pessimistic locking if needed)
DROP TABLE IF EXISTS entity_lock;

/*
    ========================================================================
    VERSION 4: PARALLEL EDITING SUPPORT
    ========================================================================
*/

/*
    ========================================================================
    4.1 Entity Versioning and History
    ========================================================================

    All editable entities get version columns for optimistic locking.
    History tables track all changes with full before/after snapshots.
*/

/*
    History tables store complete snapshots of entity state at each change.
    This allows for:
    - Full change history
    - Conflict resolution
    - Audit trails
    - Rollback capabilities
*/

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
    Entity lock table for future pessimistic locking support.
    Currently used for tracking active editors.
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
    TABLE ENHANCEMENTS - Version Columns
    ========================================================================

    Add version columns to all editable entities for optimistic locking.
    Version starts at 1 and increments on each successful update.
*/

/*
    ALTER TABLE statements to add version columns:

    -- Core entities
    ALTER TABLE ticket ADD COLUMN version INTEGER DEFAULT 1;
    ALTER TABLE project ADD COLUMN version INTEGER DEFAULT 1;
    ALTER TABLE comment ADD COLUMN version INTEGER DEFAULT 1;
    ALTER TABLE dashboard ADD COLUMN version INTEGER DEFAULT 1;
    ALTER TABLE board ADD COLUMN version INTEGER DEFAULT 1;

    -- Indexes for version queries
    CREATE INDEX tickets_get_by_version ON ticket (version);
    CREATE INDEX projects_get_by_version ON project (version);
    CREATE INDEX comments_get_by_version ON comment (version);
    CREATE INDEX dashboards_get_by_version ON dashboard (version);
    CREATE INDEX boards_get_by_version ON board (version);
*/

/*
    ========================================================================
    SEED DATA - Initial Versions
    ========================================================================

    Set initial versions for existing entities.
    This should be done in migration script.
*/

/*
    ========================================================================
    VERSION INFORMATION
    ========================================================================
*/

/*
    ========================================================================
    NOTES FOR DEVELOPERS
    ========================================================================

    NEW IN VERSION 4:

    1. Optimistic Locking:
       - All editable entities have version columns
       - Updates check version matches current value
       - Version conflict = concurrent modification detected

    2. Change History:
       - Complete snapshots stored in history tables
       - Full audit trail of all changes
       - Supports rollback and conflict resolution

    3. Conflict Resolution:
       - Client receives current state on version conflict
       - Can choose to overwrite, merge, or cancel
       - Conflict data stored for analysis

    4. Real-time Collaboration:
       - Entity locks track active editors
       - WebSocket notifications for live updates
       - User presence indicators

    API CHANGES:

    Modify operations now include:
    - version: Current known version number
    - conflict_resolution: 'overwrite', 'merge', 'cancel'

    New endpoints:
    - GET /api/v4/{entity}/{id}/history - Get change history
    - GET /api/v4/{entity}/{id}/locks - Get active editors
    - POST /api/v4/{entity}/{id}/lock - Acquire lock
    - DELETE /api/v4/{entity}/{id}/lock - Release lock

    Error codes:
    - VERSION_CONFLICT: Concurrent modification detected
    - LOCK_ACQUIRED: Entity locked by another user
    - INVALID_VERSION: Version number invalid

    MIGRATION CONSIDERATIONS:

    1. Add version columns with default value 1
    2. Create history tables
    3. Populate initial history entries for existing data
    4. Update all modify handlers to use optimistic locking
    5. Add conflict resolution logic
    6. Update client applications to handle version conflicts

    BACKWARD COMPATIBILITY:

    1. Version parameter optional in V3 API (defaults to current)
    2. History endpoints are new additions
    3. Lock endpoints are new additions
    4. Existing modify operations continue to work without version

    PERFORMANCE CONSIDERATIONS:

    1. History tables may grow large - consider partitioning
    2. Version checks add small overhead to updates
    3. Indexes optimized for common query patterns
    4. History snapshots use JSON for flexibility

    SECURITY CONSIDERATIONS:

    1. History access respects entity permissions
    2. Lock operations validate user permissions
    3. Audit logging for all collaborative actions
    4. Rate limiting on history and lock endpoints
*/
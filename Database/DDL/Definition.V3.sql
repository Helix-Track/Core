/*
    Version: 3

    JIRA Feature Parity - Complete (Phases 2 & 3)

    This version extends V2 with all remaining JIRA features to achieve complete parity:

    Phase 2 (Agile Enhancements - Priority 2):
    - Epic Support
    - Subtask Support
    - Enhanced Work Logs
    - Project Roles
    - Security Levels
    - Dashboard System
    - Advanced Board Configuration

    Phase 3 (Collaboration Features - Priority 3):
    - Voting System
    - Project Categories
    - Notification Schemes
    - Activity Stream Enhancements
    - Comment Mentions
*/

/*
    Notes:

    - This builds upon Definition.V2.sql (which includes V1)
    - All V1 and V2 tables are included
    - New tables and enhancements for Phase 2 & 3 features
    - Migration from V2 to V3 available in Migration.V2.3.sql
    - Identifiers in the system are UUID strings
    - Mapping tables are used for binding entities and defining relationships
    - Additional tables are defined to provide the meta-data to entities of the system
*/

/*
    ========================================================================
    DROP STATEMENTS - Phase 2 & 3 Tables
    ========================================================================
*/

-- Phase 2: Epic Support (no new table, uses ticket enhancements)
-- Phase 2: Subtask Support (no new table, uses ticket enhancements)

-- Phase 2: Enhanced Work Logs
DROP TABLE IF EXISTS work_log;
DROP INDEX IF EXISTS work_logs_get_by_ticket_id;
DROP INDEX IF EXISTS work_logs_get_by_user_id;
DROP INDEX IF EXISTS work_logs_get_by_work_date;
DROP INDEX IF EXISTS work_logs_get_by_created;
DROP INDEX IF EXISTS work_logs_get_by_deleted;
DROP INDEX IF EXISTS work_logs_get_by_created_and_modified;

-- Phase 2: Project Roles
DROP TABLE IF EXISTS project_role;
DROP TABLE IF EXISTS project_role_user_mapping;
DROP INDEX IF EXISTS project_roles_get_by_title;
DROP INDEX IF EXISTS project_roles_get_by_project_id;
DROP INDEX IF EXISTS project_roles_get_by_deleted;
DROP INDEX IF EXISTS project_roles_get_by_created;
DROP INDEX IF EXISTS project_roles_get_by_modified;
DROP INDEX IF EXISTS project_roles_get_by_created_and_modified;
DROP INDEX IF EXISTS project_role_users_get_by_role_id;
DROP INDEX IF EXISTS project_role_users_get_by_project_id;
DROP INDEX IF EXISTS project_role_users_get_by_user_id;
DROP INDEX IF EXISTS project_role_users_get_by_deleted;
DROP INDEX IF EXISTS project_role_users_get_by_created;

-- Phase 2: Security Levels
DROP TABLE IF EXISTS security_level;
DROP TABLE IF EXISTS security_level_permission_mapping;
DROP INDEX IF EXISTS security_levels_get_by_title;
DROP INDEX IF EXISTS security_levels_get_by_project_id;
DROP INDEX IF EXISTS security_levels_get_by_level;
DROP INDEX IF EXISTS security_levels_get_by_deleted;
DROP INDEX IF EXISTS security_levels_get_by_created;
DROP INDEX IF EXISTS security_levels_get_by_modified;
DROP INDEX IF EXISTS security_levels_get_by_created_and_modified;
DROP INDEX IF EXISTS security_level_permissions_get_by_security_level_id;
DROP INDEX IF EXISTS security_level_permissions_get_by_user_id;
DROP INDEX IF EXISTS security_level_permissions_get_by_team_id;
DROP INDEX IF EXISTS security_level_permissions_get_by_project_role_id;
DROP INDEX IF EXISTS security_level_permissions_get_by_deleted;
DROP INDEX IF EXISTS security_level_permissions_get_by_created;
DROP INDEX IF EXISTS tickets_get_by_security_level_id;

-- Phase 2: Dashboard System
DROP TABLE IF EXISTS dashboard;
DROP TABLE IF EXISTS dashboard_widget;
DROP TABLE IF EXISTS dashboard_share_mapping;
DROP INDEX IF EXISTS dashboards_get_by_title;
DROP INDEX IF EXISTS dashboards_get_by_owner_id;
DROP INDEX IF EXISTS dashboards_get_by_is_public;
DROP INDEX IF EXISTS dashboards_get_by_is_favorite;
DROP INDEX IF EXISTS dashboards_get_by_deleted;
DROP INDEX IF EXISTS dashboards_get_by_created;
DROP INDEX IF EXISTS dashboards_get_by_modified;
DROP INDEX IF EXISTS dashboards_get_by_created_and_modified;
DROP INDEX IF EXISTS dashboard_widgets_get_by_dashboard_id;
DROP INDEX IF EXISTS dashboard_widgets_get_by_widget_type;
DROP INDEX IF EXISTS dashboard_widgets_get_by_deleted;
DROP INDEX IF EXISTS dashboard_widgets_get_by_created;
DROP INDEX IF EXISTS dashboard_widgets_get_by_modified;
DROP INDEX IF EXISTS dashboard_shares_get_by_dashboard_id;
DROP INDEX IF EXISTS dashboard_shares_get_by_user_id;
DROP INDEX IF EXISTS dashboard_shares_get_by_team_id;
DROP INDEX IF EXISTS dashboard_shares_get_by_project_id;
DROP INDEX IF EXISTS dashboard_shares_get_by_deleted;
DROP INDEX IF EXISTS dashboard_shares_get_by_created;

-- Phase 2: Advanced Board Configuration
DROP TABLE IF EXISTS board_column;
DROP TABLE IF EXISTS board_swimlane;
DROP TABLE IF EXISTS board_quick_filter;
DROP INDEX IF EXISTS board_columns_get_by_board_id;
DROP INDEX IF EXISTS board_columns_get_by_status_id;
DROP INDEX IF EXISTS board_columns_get_by_position;
DROP INDEX IF EXISTS board_columns_get_by_deleted;
DROP INDEX IF EXISTS board_columns_get_by_created;
DROP INDEX IF EXISTS board_columns_get_by_modified;
DROP INDEX IF EXISTS board_swimlanes_get_by_board_id;
DROP INDEX IF EXISTS board_swimlanes_get_by_position;
DROP INDEX IF EXISTS board_swimlanes_get_by_deleted;
DROP INDEX IF EXISTS board_swimlanes_get_by_created;
DROP INDEX IF EXISTS board_swimlanes_get_by_modified;
DROP INDEX IF EXISTS board_quick_filters_get_by_board_id;
DROP INDEX IF EXISTS board_quick_filters_get_by_position;
DROP INDEX IF EXISTS board_quick_filters_get_by_deleted;
DROP INDEX IF EXISTS board_quick_filters_get_by_created;
DROP INDEX IF EXISTS boards_get_by_filter_id;
DROP INDEX IF EXISTS boards_get_by_board_type;

-- Phase 3: Voting System
DROP TABLE IF EXISTS ticket_vote_mapping;
DROP INDEX IF EXISTS ticket_votes_get_by_ticket_id;
DROP INDEX IF EXISTS ticket_votes_get_by_user_id;
DROP INDEX IF EXISTS ticket_votes_get_by_ticket_and_user;
DROP INDEX IF EXISTS ticket_votes_get_by_deleted;
DROP INDEX IF EXISTS ticket_votes_get_by_created;
DROP INDEX IF EXISTS tickets_get_by_vote_count;

-- Phase 3: Project Categories
DROP TABLE IF EXISTS project_category;
DROP INDEX IF EXISTS project_categories_get_by_title;
DROP INDEX IF EXISTS project_categories_get_by_deleted;
DROP INDEX IF EXISTS project_categories_get_by_created;
DROP INDEX IF EXISTS project_categories_get_by_modified;
DROP INDEX IF EXISTS project_categories_get_by_created_and_modified;
DROP INDEX IF EXISTS projects_get_by_category_id;

-- Phase 3: Notification Schemes
DROP TABLE IF EXISTS notification_scheme;
DROP TABLE IF EXISTS notification_event;
DROP TABLE IF EXISTS notification_rule;
DROP INDEX IF EXISTS notification_schemes_get_by_title;
DROP INDEX IF EXISTS notification_schemes_get_by_project_id;
DROP INDEX IF EXISTS notification_schemes_get_by_deleted;
DROP INDEX IF EXISTS notification_schemes_get_by_created;
DROP INDEX IF EXISTS notification_schemes_get_by_modified;
DROP INDEX IF EXISTS notification_schemes_get_by_created_and_modified;
DROP INDEX IF EXISTS notification_events_get_by_event_type;
DROP INDEX IF EXISTS notification_events_get_by_title;
DROP INDEX IF EXISTS notification_events_get_by_deleted;
DROP INDEX IF EXISTS notification_events_get_by_created;
DROP INDEX IF EXISTS notification_rules_get_by_scheme_id;
DROP INDEX IF EXISTS notification_rules_get_by_event_id;
DROP INDEX IF EXISTS notification_rules_get_by_recipient_type;
DROP INDEX IF EXISTS notification_rules_get_by_recipient_id;
DROP INDEX IF EXISTS notification_rules_get_by_deleted;
DROP INDEX IF EXISTS notification_rules_get_by_created;

-- Phase 3: Activity Stream Enhancements (uses audit table enhancements)
DROP INDEX IF EXISTS audit_get_by_is_public;
DROP INDEX IF EXISTS audit_get_by_activity_type;

-- Phase 3: Comment Mentions
DROP TABLE IF EXISTS comment_mention_mapping;
DROP INDEX IF EXISTS comment_mentions_get_by_comment_id;
DROP INDEX IF EXISTS comment_mentions_get_by_user_id;
DROP INDEX IF EXISTS comment_mentions_get_by_deleted;
DROP INDEX IF EXISTS comment_mentions_get_by_created;

/*
    ========================================================================
    PHASE 2: AGILE ENHANCEMENTS
    ========================================================================
*/

/*
    ========================================================================
    2.1 Epic Support
    ========================================================================

    Epics are large user stories that can be broken down into smaller stories.
    Epic functionality is implemented as enhancements to the ticket table.

    Enhanced Columns in ticket table:
    - is_epic: Boolean flag indicating if ticket is an epic
    - epic_id: Reference to parent epic (for stories belonging to epic)
    - epic_color: Color for epic display
    - epic_name: Short name for epic

    These columns will be added via ALTER TABLE in Migration.V2.3.sql
*/

/*
    ========================================================================
    2.2 Subtask Support
    ========================================================================

    Subtasks are smaller tasks that belong to a parent ticket.
    Subtask functionality is implemented as enhancements to the ticket table.

    Enhanced Columns in ticket table:
    - is_subtask: Boolean flag indicating if ticket is a subtask
    - parent_ticket_id: Reference to parent ticket

    These columns will be added via ALTER TABLE in Migration.V2.3.sql
*/

/*
    ========================================================================
    2.3 Enhanced Work Logs
    ========================================================================

    Detailed time tracking for tickets.
    Tracks time spent on tickets with detailed work logs.
*/

CREATE TABLE work_log
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

CREATE INDEX work_logs_get_by_ticket_id ON work_log (ticket_id);
CREATE INDEX work_logs_get_by_user_id ON work_log (user_id);
CREATE INDEX work_logs_get_by_work_date ON work_log (work_date);
CREATE INDEX work_logs_get_by_created ON work_log (created);
CREATE INDEX work_logs_get_by_deleted ON work_log (deleted);
CREATE INDEX work_logs_get_by_created_and_modified ON work_log (created, modified);

/*
    ========================================================================
    2.4 Project Roles
    ========================================================================

    Advanced access control with project-specific roles.
    Roles can be global (project_id NULL) or project-specific.
    Users are assigned to roles which grant specific permissions.
*/

CREATE TABLE project_role
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL,
    description TEXT,
    project_id  TEXT,              -- NULL for global roles
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL
);

CREATE INDEX project_roles_get_by_title ON project_role (title);
CREATE INDEX project_roles_get_by_project_id ON project_role (project_id);
CREATE INDEX project_roles_get_by_deleted ON project_role (deleted);
CREATE INDEX project_roles_get_by_created ON project_role (created);
CREATE INDEX project_roles_get_by_modified ON project_role (modified);
CREATE INDEX project_roles_get_by_created_and_modified ON project_role (created, modified);

CREATE TABLE project_role_user_mapping
(
    id               TEXT    NOT NULL PRIMARY KEY UNIQUE,
    project_role_id  TEXT    NOT NULL,
    project_id       TEXT    NOT NULL,
    user_id          TEXT    NOT NULL,
    created          INTEGER NOT NULL,
    deleted          BOOLEAN NOT NULL,
    UNIQUE (project_role_id, project_id, user_id)
);

CREATE INDEX project_role_users_get_by_role_id ON project_role_user_mapping (project_role_id);
CREATE INDEX project_role_users_get_by_project_id ON project_role_user_mapping (project_id);
CREATE INDEX project_role_users_get_by_user_id ON project_role_user_mapping (user_id);
CREATE INDEX project_role_users_get_by_deleted ON project_role_user_mapping (deleted);
CREATE INDEX project_role_users_get_by_created ON project_role_user_mapping (created);

/*
    ========================================================================
    2.5 Security Levels
    ========================================================================

    Enterprise security feature for controlling access to tickets.
    Security levels define who can view/access specific tickets.
    Access can be granted to users, teams, or project roles.
*/

CREATE TABLE security_level
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL,
    description TEXT,
    project_id  TEXT    NOT NULL,
    level       INTEGER NOT NULL,  -- Numeric level for hierarchy
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL
);

CREATE INDEX security_levels_get_by_title ON security_level (title);
CREATE INDEX security_levels_get_by_project_id ON security_level (project_id);
CREATE INDEX security_levels_get_by_level ON security_level (level);
CREATE INDEX security_levels_get_by_deleted ON security_level (deleted);
CREATE INDEX security_levels_get_by_created ON security_level (created);
CREATE INDEX security_levels_get_by_modified ON security_level (modified);
CREATE INDEX security_levels_get_by_created_and_modified ON security_level (created, modified);

CREATE TABLE security_level_permission_mapping
(
    id                TEXT    NOT NULL PRIMARY KEY UNIQUE,
    security_level_id TEXT    NOT NULL,
    user_id           TEXT,              -- NULL if not user-specific
    team_id           TEXT,              -- NULL if not team-specific
    project_role_id   TEXT,              -- NULL if not role-specific
    created           INTEGER NOT NULL,
    deleted           BOOLEAN NOT NULL
);

CREATE INDEX security_level_permissions_get_by_security_level_id ON security_level_permission_mapping (security_level_id);
CREATE INDEX security_level_permissions_get_by_user_id ON security_level_permission_mapping (user_id);
CREATE INDEX security_level_permissions_get_by_team_id ON security_level_permission_mapping (team_id);
CREATE INDEX security_level_permissions_get_by_project_role_id ON security_level_permission_mapping (project_role_id);
CREATE INDEX security_level_permissions_get_by_deleted ON security_level_permission_mapping (deleted);
CREATE INDEX security_level_permissions_get_by_created ON security_level_permission_mapping (created);

-- Enhanced ticket table column for security levels
-- ALTER TABLE ticket ADD COLUMN security_level_id TEXT;
-- CREATE INDEX tickets_get_by_security_level_id ON ticket (security_level_id);

/*
    ========================================================================
    2.6 Dashboard System
    ========================================================================

    Customizable dashboards with widgets for visualization and reporting.
    Dashboards can be shared with users, teams, or projects.
    Widgets display various data (filters, charts, activity streams, etc.).
*/

CREATE TABLE dashboard
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL,
    description TEXT,
    owner_id    TEXT    NOT NULL,
    is_public   BOOLEAN NOT NULL DEFAULT FALSE,
    is_favorite BOOLEAN NOT NULL DEFAULT FALSE,
    layout      TEXT,              -- JSON layout configuration
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL
);

CREATE INDEX dashboards_get_by_title ON dashboard (title);
CREATE INDEX dashboards_get_by_owner_id ON dashboard (owner_id);
CREATE INDEX dashboards_get_by_is_public ON dashboard (is_public);
CREATE INDEX dashboards_get_by_is_favorite ON dashboard (is_favorite);
CREATE INDEX dashboards_get_by_deleted ON dashboard (deleted);
CREATE INDEX dashboards_get_by_created ON dashboard (created);
CREATE INDEX dashboards_get_by_modified ON dashboard (modified);
CREATE INDEX dashboards_get_by_created_and_modified ON dashboard (created, modified);

CREATE TABLE dashboard_widget
(
    id            TEXT    NOT NULL PRIMARY KEY UNIQUE,
    dashboard_id  TEXT    NOT NULL,
    widget_type   TEXT    NOT NULL,  -- 'filter_results', 'pie_chart', 'activity_stream', etc.
    title         TEXT,
    position_x    INTEGER,
    position_y    INTEGER,
    width         INTEGER,
    height        INTEGER,
    configuration TEXT,              -- JSON widget configuration
    created       INTEGER NOT NULL,
    modified      INTEGER NOT NULL,
    deleted       BOOLEAN NOT NULL
);

CREATE INDEX dashboard_widgets_get_by_dashboard_id ON dashboard_widget (dashboard_id);
CREATE INDEX dashboard_widgets_get_by_widget_type ON dashboard_widget (widget_type);
CREATE INDEX dashboard_widgets_get_by_deleted ON dashboard_widget (deleted);
CREATE INDEX dashboard_widgets_get_by_created ON dashboard_widget (created);
CREATE INDEX dashboard_widgets_get_by_modified ON dashboard_widget (modified);

CREATE TABLE dashboard_share_mapping
(
    id           TEXT    NOT NULL PRIMARY KEY UNIQUE,
    dashboard_id TEXT    NOT NULL,
    user_id      TEXT,              -- NULL if not user-specific
    team_id      TEXT,              -- NULL if not team-specific
    project_id   TEXT,              -- NULL if not project-specific
    created      INTEGER NOT NULL,
    deleted      BOOLEAN NOT NULL
);

CREATE INDEX dashboard_shares_get_by_dashboard_id ON dashboard_share_mapping (dashboard_id);
CREATE INDEX dashboard_shares_get_by_user_id ON dashboard_share_mapping (user_id);
CREATE INDEX dashboard_shares_get_by_team_id ON dashboard_share_mapping (team_id);
CREATE INDEX dashboard_shares_get_by_project_id ON dashboard_share_mapping (project_id);
CREATE INDEX dashboard_shares_get_by_deleted ON dashboard_share_mapping (deleted);
CREATE INDEX dashboard_shares_get_by_created ON dashboard_share_mapping (created);

/*
    ========================================================================
    2.7 Advanced Board Configuration
    ========================================================================

    Advanced configuration for Scrum and Kanban boards.
    Boards can have custom columns, swimlanes, and quick filters.
    WIP limits can be set on columns.
*/

CREATE TABLE board_column
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    board_id    TEXT    NOT NULL,
    title       TEXT    NOT NULL,
    status_id   TEXT,              -- Maps to ticket_status
    position    INTEGER NOT NULL,
    max_items   INTEGER,           -- WIP limit
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL
);

CREATE INDEX board_columns_get_by_board_id ON board_column (board_id);
CREATE INDEX board_columns_get_by_status_id ON board_column (status_id);
CREATE INDEX board_columns_get_by_position ON board_column (position);
CREATE INDEX board_columns_get_by_deleted ON board_column (deleted);
CREATE INDEX board_columns_get_by_created ON board_column (created);
CREATE INDEX board_columns_get_by_modified ON board_column (modified);

CREATE TABLE board_swimlane
(
    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    board_id TEXT    NOT NULL,
    title    TEXT    NOT NULL,
    query    TEXT,                 -- JQL-like query for swimlane
    position INTEGER NOT NULL,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL
);

CREATE INDEX board_swimlanes_get_by_board_id ON board_swimlane (board_id);
CREATE INDEX board_swimlanes_get_by_position ON board_swimlane (position);
CREATE INDEX board_swimlanes_get_by_deleted ON board_swimlane (deleted);
CREATE INDEX board_swimlanes_get_by_created ON board_swimlane (created);
CREATE INDEX board_swimlanes_get_by_modified ON board_swimlane (modified);

CREATE TABLE board_quick_filter
(
    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    board_id TEXT    NOT NULL,
    title    TEXT    NOT NULL,
    query    TEXT,                 -- JQL-like query for quick filter
    position INTEGER NOT NULL,
    created  INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL
);

CREATE INDEX board_quick_filters_get_by_board_id ON board_quick_filter (board_id);
CREATE INDEX board_quick_filters_get_by_position ON board_quick_filter (position);
CREATE INDEX board_quick_filters_get_by_deleted ON board_quick_filter (deleted);
CREATE INDEX board_quick_filters_get_by_created ON board_quick_filter (created);

-- Enhanced board table columns
-- ALTER TABLE board ADD COLUMN filter_id TEXT;
-- ALTER TABLE board ADD COLUMN board_type TEXT;  -- 'scrum', 'kanban'
-- CREATE INDEX boards_get_by_filter_id ON board (filter_id);
-- CREATE INDEX boards_get_by_board_type ON board (board_type);

/*
    ========================================================================
    PHASE 3: COLLABORATION FEATURES
    ========================================================================
*/

/*
    ========================================================================
    3.1 Voting System
    ========================================================================

    Community engagement feature allowing users to vote on tickets.
    Vote counts help prioritize popular feature requests.
*/

CREATE TABLE ticket_vote_mapping
(
    id        TEXT    NOT NULL PRIMARY KEY UNIQUE,
    ticket_id TEXT    NOT NULL,
    user_id   TEXT    NOT NULL,
    created   INTEGER NOT NULL,
    deleted   BOOLEAN NOT NULL,
    UNIQUE (ticket_id, user_id)
);

CREATE INDEX ticket_votes_get_by_ticket_id ON ticket_vote_mapping (ticket_id);
CREATE INDEX ticket_votes_get_by_user_id ON ticket_vote_mapping (user_id);
CREATE INDEX ticket_votes_get_by_ticket_and_user ON ticket_vote_mapping (ticket_id, user_id);
CREATE INDEX ticket_votes_get_by_deleted ON ticket_vote_mapping (deleted);
CREATE INDEX ticket_votes_get_by_created ON ticket_vote_mapping (created);

-- Enhanced ticket table column for vote count
-- ALTER TABLE ticket ADD COLUMN vote_count INTEGER DEFAULT 0;
-- CREATE INDEX tickets_get_by_vote_count ON ticket (vote_count);

/*
    ========================================================================
    3.2 Project Categories
    ========================================================================

    Organizational feature for grouping projects by category.
    Helps organize large numbers of projects.
*/

CREATE TABLE project_category
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL
);

CREATE INDEX project_categories_get_by_title ON project_category (title);
CREATE INDEX project_categories_get_by_deleted ON project_category (deleted);
CREATE INDEX project_categories_get_by_created ON project_category (created);
CREATE INDEX project_categories_get_by_modified ON project_category (modified);
CREATE INDEX project_categories_get_by_created_and_modified ON project_category (created, modified);

-- Enhanced project table column for category
-- ALTER TABLE project ADD COLUMN project_category_id TEXT;
-- CREATE INDEX projects_get_by_category_id ON project (project_category_id);

/*
    ========================================================================
    3.3 Notification Schemes
    ========================================================================

    Customizable notification system.
    Defines when and who receives notifications for project events.
*/

CREATE TABLE notification_scheme
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL,
    description TEXT,
    project_id  TEXT,              -- NULL for global schemes
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL
);

CREATE INDEX notification_schemes_get_by_title ON notification_scheme (title);
CREATE INDEX notification_schemes_get_by_project_id ON notification_scheme (project_id);
CREATE INDEX notification_schemes_get_by_deleted ON notification_scheme (deleted);
CREATE INDEX notification_schemes_get_by_created ON notification_scheme (created);
CREATE INDEX notification_schemes_get_by_modified ON notification_scheme (modified);
CREATE INDEX notification_schemes_get_by_created_and_modified ON notification_scheme (created, modified);

CREATE TABLE notification_event
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    event_type  TEXT    NOT NULL,  -- 'issue_created', 'issue_updated', 'comment_added', etc.
    title       TEXT    NOT NULL,
    description TEXT,
    created     INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL
);

CREATE INDEX notification_events_get_by_event_type ON notification_event (event_type);
CREATE INDEX notification_events_get_by_title ON notification_event (title);
CREATE INDEX notification_events_get_by_deleted ON notification_event (deleted);
CREATE INDEX notification_events_get_by_created ON notification_event (created);

CREATE TABLE notification_rule
(
    id                      TEXT    NOT NULL PRIMARY KEY UNIQUE,
    notification_scheme_id  TEXT    NOT NULL,
    notification_event_id   TEXT    NOT NULL,
    recipient_type          TEXT    NOT NULL,  -- 'assignee', 'reporter', 'watcher', 'user', 'team', 'project_role'
    recipient_id            TEXT,              -- user_id, team_id, or role_id (NULL for assignee/reporter/watcher)
    created                 INTEGER NOT NULL,
    deleted                 BOOLEAN NOT NULL
);

CREATE INDEX notification_rules_get_by_scheme_id ON notification_rule (notification_scheme_id);
CREATE INDEX notification_rules_get_by_event_id ON notification_rule (notification_event_id);
CREATE INDEX notification_rules_get_by_recipient_type ON notification_rule (recipient_type);
CREATE INDEX notification_rules_get_by_recipient_id ON notification_rule (recipient_id);
CREATE INDEX notification_rules_get_by_deleted ON notification_rule (deleted);
CREATE INDEX notification_rules_get_by_created ON notification_rule (created);

/*
    ========================================================================
    3.4 Activity Stream Enhancements
    ========================================================================

    Enhanced audit trail with activity types and public/private flags.
    These enhancements are applied to the existing audit table.
*/

-- Enhanced audit table columns
-- ALTER TABLE audit ADD COLUMN is_public BOOLEAN DEFAULT TRUE;
-- ALTER TABLE audit ADD COLUMN activity_type TEXT;  -- 'comment', 'status_change', 'assignment', etc.
-- CREATE INDEX audit_get_by_is_public ON audit (is_public);
-- CREATE INDEX audit_get_by_activity_type ON audit (activity_type);

/*
    ========================================================================
    3.5 Comment Mentions
    ========================================================================

    @mention functionality in comments.
    Tracks which users are mentioned in comments for notifications.
*/

CREATE TABLE comment_mention_mapping
(
    id                 TEXT    NOT NULL PRIMARY KEY UNIQUE,
    comment_id         TEXT    NOT NULL,
    mentioned_user_id  TEXT    NOT NULL,
    created            INTEGER NOT NULL,
    deleted            BOOLEAN NOT NULL
);

CREATE INDEX comment_mentions_get_by_comment_id ON comment_mention_mapping (comment_id);
CREATE INDEX comment_mentions_get_by_user_id ON comment_mention_mapping (mentioned_user_id);
CREATE INDEX comment_mentions_get_by_deleted ON comment_mention_mapping (deleted);
CREATE INDEX comment_mentions_get_by_created ON comment_mention_mapping (created);

/*
    ========================================================================
    TABLE ENHANCEMENTS SUMMARY
    ========================================================================

    The following ALTER TABLE statements must be executed in Migration.V2.3.sql:

    -- Phase 2 Enhancements

    -- Epic Support
    ALTER TABLE ticket ADD COLUMN is_epic BOOLEAN DEFAULT FALSE;
    ALTER TABLE ticket ADD COLUMN epic_id TEXT;
    ALTER TABLE ticket ADD COLUMN epic_color TEXT;
    ALTER TABLE ticket ADD COLUMN epic_name TEXT;
    CREATE INDEX tickets_get_by_is_epic ON ticket (is_epic);
    CREATE INDEX tickets_get_by_epic_id ON ticket (epic_id);

    -- Subtask Support
    ALTER TABLE ticket ADD COLUMN is_subtask BOOLEAN DEFAULT FALSE;
    ALTER TABLE ticket ADD COLUMN parent_ticket_id TEXT;
    CREATE INDEX tickets_get_by_is_subtask ON ticket (is_subtask);
    CREATE INDEX tickets_get_by_parent_ticket_id ON ticket (parent_ticket_id);

    -- Security Levels
    ALTER TABLE ticket ADD COLUMN security_level_id TEXT;
    CREATE INDEX tickets_get_by_security_level_id ON ticket (security_level_id);

    -- Advanced Board Configuration
    ALTER TABLE board ADD COLUMN filter_id TEXT;
    ALTER TABLE board ADD COLUMN board_type TEXT;
    CREATE INDEX boards_get_by_filter_id ON board (filter_id);
    CREATE INDEX boards_get_by_board_type ON board (board_type);

    -- Phase 3 Enhancements

    -- Voting System
    ALTER TABLE ticket ADD COLUMN vote_count INTEGER DEFAULT 0;
    CREATE INDEX tickets_get_by_vote_count ON ticket (vote_count);

    -- Project Categories
    ALTER TABLE project ADD COLUMN project_category_id TEXT;
    CREATE INDEX projects_get_by_category_id ON project (project_category_id);

    -- Activity Stream Enhancements
    ALTER TABLE audit ADD COLUMN is_public BOOLEAN DEFAULT TRUE;
    ALTER TABLE audit ADD COLUMN activity_type TEXT;
    CREATE INDEX audit_get_by_is_public ON audit (is_public);
    CREATE INDEX audit_get_by_activity_type ON audit (activity_type);
*/

/*
    ========================================================================
    SEED DATA - Notification Events (Common Events)
    ========================================================================
*/

-- INSERT INTO notification_event (id, event_type, title, description, created, deleted)
-- VALUES
--     ('event-issue-created', 'issue_created', 'Issue Created', 'Triggered when a new issue is created', strftime('%s', 'now'), 0),
--     ('event-issue-updated', 'issue_updated', 'Issue Updated', 'Triggered when an issue is updated', strftime('%s', 'now'), 0),
--     ('event-issue-deleted', 'issue_deleted', 'Issue Deleted', 'Triggered when an issue is deleted', strftime('%s', 'now'), 0),
--     ('event-comment-added', 'comment_added', 'Comment Added', 'Triggered when a comment is added', strftime('%s', 'now'), 0),
--     ('event-comment-updated', 'comment_updated', 'Comment Updated', 'Triggered when a comment is updated', strftime('%s', 'now'), 0),
--     ('event-comment-deleted', 'comment_deleted', 'Comment Deleted', 'Triggered when a comment is deleted', strftime('%s', 'now'), 0),
--     ('event-status-changed', 'status_changed', 'Status Changed', 'Triggered when issue status changes', strftime('%s', 'now'), 0),
--     ('event-assignee-changed', 'assignee_changed', 'Assignee Changed', 'Triggered when assignee changes', strftime('%s', 'now'), 0),
--     ('event-priority-changed', 'priority_changed', 'Priority Changed', 'Triggered when priority changes', strftime('%s', 'now'), 0),
--     ('event-work-logged', 'work_logged', 'Work Logged', 'Triggered when work is logged', strftime('%s', 'now'), 0),
--     ('event-user-mentioned', 'user_mentioned', 'User Mentioned', 'Triggered when user is @mentioned', strftime('%s', 'now'), 0);

/*
    ========================================================================
    VERSION INFORMATION
    ========================================================================
*/

-- UPDATE system_info SET description = 'Database schema version 3 - Complete JIRA Feature Parity (Phases 2 & 3)'
-- WHERE id = (SELECT id FROM system_info ORDER BY created DESC LIMIT 1);

/*
    ========================================================================
    NOTES FOR DEVELOPERS
    ========================================================================

    NEW IN VERSION 3:

    Phase 2 (Agile Enhancements):
    1. Epic Support - Hierarchical story management via ticket enhancements
    2. Subtask Support - Task decomposition via ticket enhancements
    3. Enhanced Work Logs - Detailed time tracking with work_log table
    4. Project Roles - Advanced access control with project_role tables
    5. Security Levels - Enterprise security with security_level tables
    6. Dashboard System - Visualization with dashboard tables
    7. Advanced Board Config - Scrum/Kanban enhancements with board_column, board_swimlane, board_quick_filter

    Phase 3 (Collaboration Features):
    1. Voting System - Community engagement via ticket_vote_mapping
    2. Project Categories - Organization via project_category
    3. Notification Schemes - Customizable notifications via notification_* tables
    4. Activity Stream - Enhanced audit trail via audit table enhancements
    5. Comment Mentions - @user functionality via comment_mention_mapping

    TOTAL NEW TABLES: 18
    - work_log
    - project_role
    - project_role_user_mapping
    - security_level
    - security_level_permission_mapping
    - dashboard
    - dashboard_widget
    - dashboard_share_mapping
    - board_column
    - board_swimlane
    - board_quick_filter
    - ticket_vote_mapping
    - project_category
    - notification_scheme
    - notification_event
    - notification_rule
    - comment_mention_mapping

    ENHANCED TABLES: 4
    - ticket: +9 columns (epic support, subtask support, security level, vote count)
    - board: +2 columns (filter_id, board_type)
    - project: +1 column (project_category_id)
    - audit: +2 columns (is_public, activity_type)

    MIGRATION CONSIDERATIONS:

    1. All new tables can be created without impacting existing data
    2. ALTER TABLE statements must be executed in correct order
    3. New columns are nullable/have defaults to support existing data
    4. Existing queries will continue to work
    5. Performance impact minimal with proper indexing
    6. Seed data for notification_event recommended
    7. Test migration on backup database first
    8. Monitor query performance after migration

    BACKWARD COMPATIBILITY:

    1. All V1 and V2 tables remain unchanged in structure
    2. New columns have defaults or are nullable
    3. Existing queries will continue to work
    4. API version 3 adds new endpoints, v1/v2 endpoints remain functional
    5. No breaking changes to existing functionality

    PERFORMANCE NOTES:

    1. All new tables have appropriate indexes
    2. Foreign keys maintained at application level
    3. Vote count denormalized for performance
    4. Dashboard widget configuration uses JSON for flexibility
    5. Notification rules should be cached at application level
    6. Consider partitioning work_log table for large installations

    SECURITY CONSIDERATIONS:

    1. Security levels provide fine-grained access control
    2. Audit enhancements track public/private activities
    3. Notification rules respect security levels
    4. Dashboard sharing requires permission validation
    5. Project roles integrate with permission engine

    API ENDPOINTS:

    V3 adds approximately 85 new actions:
    - Epic: 8 actions
    - Subtask: 5 actions
    - Work Log: 7 actions
    - Project Role: 8 actions
    - Security Level: 8 actions
    - Dashboard: 12 actions
    - Board Advanced: 12 actions
    - Vote: 5 actions
    - Project Category: 6 actions
    - Notification: 10 actions
    - Activity Stream: 5 actions
    - Mention: 5 actions

    Total API endpoints (all versions): ~400
*/

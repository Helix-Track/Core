/*
    Version: 2

    JIRA Feature Parity - Phase 1 (Priority 1 Features)

    This version extends V1 with critical JIRA features:
    - Priority System
    - Resolution System
    - Project Lead & Assignee
    - Watchers
    - Product Versions (Affected/Fix Versions)
    - Saved Filters
    - Custom Fields
*/

/*
    Notes:

    - This builds upon Definition.V1.sql
    - All V1 tables are included
    - New tables and enhancements for Phase 1 features
    - Migration from V1 to V2 available in Migration.V1.2.sql
    - Identifiers in the system are UUID strings
    - Mapping tables are used for binding entities and defining relationships
    - Additional tables are defined to provide the meta-data to entities of the system
*/

-- Include all V1 table drops and indexes
-- (In practice, this would source Definition.V1.sql)
-- For this version, we're adding new drops for V2 tables:

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

DROP INDEX IF EXISTS priorities_get_by_title;
DROP INDEX IF EXISTS priorities_get_by_level;
DROP INDEX IF EXISTS priorities_get_by_deleted;
DROP INDEX IF EXISTS priorities_get_by_created;
DROP INDEX IF EXISTS priorities_get_by_modified;
DROP INDEX IF EXISTS priorities_get_by_created_and_modified;

DROP INDEX IF EXISTS resolutions_get_by_title;
DROP INDEX IF EXISTS resolutions_get_by_deleted;
DROP INDEX IF EXISTS resolutions_get_by_created;
DROP INDEX IF EXISTS resolutions_get_by_modified;
DROP INDEX IF EXISTS resolutions_get_by_created_and_modified;

DROP INDEX IF EXISTS ticket_watchers_get_by_ticket_id;
DROP INDEX IF EXISTS ticket_watchers_get_by_user_id;
DROP INDEX IF EXISTS ticket_watchers_get_by_ticket_and_user;
DROP INDEX IF EXISTS ticket_watchers_get_by_deleted;
DROP INDEX IF EXISTS ticket_watchers_get_by_created;

DROP INDEX IF EXISTS versions_get_by_title;
DROP INDEX IF EXISTS versions_get_by_project_id;
DROP INDEX IF EXISTS versions_get_by_released;
DROP INDEX IF EXISTS versions_get_by_archived;
DROP INDEX IF EXISTS versions_get_by_release_date;
DROP INDEX IF EXISTS versions_get_by_deleted;
DROP INDEX IF EXISTS versions_get_by_created;
DROP INDEX IF EXISTS versions_get_by_modified;
DROP INDEX IF EXISTS versions_get_by_created_and_modified;

DROP INDEX IF EXISTS ticket_affected_versions_get_by_ticket_id;
DROP INDEX IF EXISTS ticket_affected_versions_get_by_version_id;
DROP INDEX IF EXISTS ticket_affected_versions_get_by_deleted;
DROP INDEX IF EXISTS ticket_affected_versions_get_by_created;

DROP INDEX IF EXISTS ticket_fix_versions_get_by_ticket_id;
DROP INDEX IF EXISTS ticket_fix_versions_get_by_version_id;
DROP INDEX IF EXISTS ticket_fix_versions_get_by_deleted;
DROP INDEX IF EXISTS ticket_fix_versions_get_by_created;

DROP INDEX IF EXISTS filters_get_by_title;
DROP INDEX IF EXISTS filters_get_by_owner_id;
DROP INDEX IF EXISTS filters_get_by_is_public;
DROP INDEX IF EXISTS filters_get_by_is_favorite;
DROP INDEX IF EXISTS filters_get_by_deleted;
DROP INDEX IF EXISTS filters_get_by_created;
DROP INDEX IF EXISTS filters_get_by_modified;
DROP INDEX IF EXISTS filters_get_by_created_and_modified;

DROP INDEX IF EXISTS filter_shares_get_by_filter_id;
DROP INDEX IF EXISTS filter_shares_get_by_user_id;
DROP INDEX IF EXISTS filter_shares_get_by_team_id;
DROP INDEX IF EXISTS filter_shares_get_by_project_id;
DROP INDEX IF EXISTS filter_shares_get_by_deleted;
DROP INDEX IF EXISTS filter_shares_get_by_created;

DROP INDEX IF EXISTS custom_fields_get_by_field_name;
DROP INDEX IF EXISTS custom_fields_get_by_field_type;
DROP INDEX IF EXISTS custom_fields_get_by_project_id;
DROP INDEX IF EXISTS custom_fields_get_by_is_required;
DROP INDEX IF EXISTS custom_fields_get_by_deleted;
DROP INDEX IF EXISTS custom_fields_get_by_created;
DROP INDEX IF EXISTS custom_fields_get_by_modified;
DROP INDEX IF EXISTS custom_fields_get_by_created_and_modified;

DROP INDEX IF EXISTS custom_field_options_get_by_custom_field_id;
DROP INDEX IF EXISTS custom_field_options_get_by_value;
DROP INDEX IF EXISTS custom_field_options_get_by_position;
DROP INDEX IF EXISTS custom_field_options_get_by_deleted;
DROP INDEX IF EXISTS custom_field_options_get_by_created;

DROP INDEX IF EXISTS ticket_custom_field_values_get_by_ticket_id;
DROP INDEX IF EXISTS ticket_custom_field_values_get_by_custom_field_id;
DROP INDEX IF EXISTS ticket_custom_field_values_get_by_ticket_and_field;
DROP INDEX IF EXISTS ticket_custom_field_values_get_by_deleted;
DROP INDEX IF EXISTS ticket_custom_field_values_get_by_created;
DROP INDEX IF EXISTS ticket_custom_field_values_get_by_modified;
DROP INDEX IF EXISTS ticket_custom_field_values_get_by_created_and_modified;

/*
    ========================================================================
    NEW TABLES - Phase 1 (Priority 1 Features)
    ========================================================================
*/

/*
    Priority System

    Defines issue priorities (Highest, High, Medium, Low, Lowest).
    Each priority has a level (1-5) for ordering.
    Icons and colors provide visual representation.
*/
CREATE TABLE priority
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    level       INTEGER NOT NULL,  -- 1 (Lowest) to 5 (Highest)
    icon        TEXT,              -- Icon identifier
    color       TEXT,              -- Hex color code (#FF0000)
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL
);

CREATE INDEX priorities_get_by_title ON priority (title);
CREATE INDEX priorities_get_by_level ON priority (level);
CREATE INDEX priorities_get_by_deleted ON priority (deleted);
CREATE INDEX priorities_get_by_created ON priority (created);
CREATE INDEX priorities_get_by_modified ON priority (modified);
CREATE INDEX priorities_get_by_created_and_modified ON priority (created, modified);

/*
    Resolution System

    Defines how issues are resolved (Fixed, Won't Fix, Duplicate, Cannot Reproduce, etc.).
    Resolutions are set when tickets are closed.
*/
CREATE TABLE resolution
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL
);

CREATE INDEX resolutions_get_by_title ON resolution (title);
CREATE INDEX resolutions_get_by_deleted ON resolution (deleted);
CREATE INDEX resolutions_get_by_created ON resolution (created);
CREATE INDEX resolutions_get_by_modified ON resolution (modified);
CREATE INDEX resolutions_get_by_created_and_modified ON resolution (created, modified);

/*
    Ticket Watchers

    Users can watch tickets to receive notifications about updates.
    Many-to-many relationship between tickets and users.
*/
CREATE TABLE ticket_watcher_mapping
(
    id         TEXT    NOT NULL PRIMARY KEY UNIQUE,
    ticket_id  TEXT    NOT NULL,
    user_id    TEXT    NOT NULL,
    created    INTEGER NOT NULL,
    deleted    BOOLEAN NOT NULL,
    UNIQUE (ticket_id, user_id)
);

CREATE INDEX ticket_watchers_get_by_ticket_id ON ticket_watcher_mapping (ticket_id);
CREATE INDEX ticket_watchers_get_by_user_id ON ticket_watcher_mapping (user_id);
CREATE INDEX ticket_watchers_get_by_ticket_and_user ON ticket_watcher_mapping (ticket_id, user_id);
CREATE INDEX ticket_watchers_get_by_deleted ON ticket_watcher_mapping (deleted);
CREATE INDEX ticket_watchers_get_by_created ON ticket_watcher_mapping (created);

/*
    Product Versions / Releases

    Tracks product versions/releases for projects.
    Versions can have start dates, release dates, and release status.
    Archived versions are hidden from active lists.
*/
CREATE TABLE version
(
    id           TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title        TEXT    NOT NULL,
    description  TEXT,
    project_id   TEXT    NOT NULL,
    start_date   INTEGER,           -- Unix timestamp
    release_date INTEGER,           -- Unix timestamp
    released     BOOLEAN NOT NULL DEFAULT FALSE,
    archived     BOOLEAN NOT NULL DEFAULT FALSE,
    created      INTEGER NOT NULL,
    modified     INTEGER NOT NULL,
    deleted      BOOLEAN NOT NULL
);

CREATE INDEX versions_get_by_title ON version (title);
CREATE INDEX versions_get_by_project_id ON version (project_id);
CREATE INDEX versions_get_by_released ON version (released);
CREATE INDEX versions_get_by_archived ON version (archived);
CREATE INDEX versions_get_by_release_date ON version (release_date);
CREATE INDEX versions_get_by_deleted ON version (deleted);
CREATE INDEX versions_get_by_created ON version (created);
CREATE INDEX versions_get_by_modified ON version (modified);
CREATE INDEX versions_get_by_created_and_modified ON version (created, modified);

/*
    Affected Versions

    Links tickets to versions that are affected by the issue.
    A ticket can affect multiple versions.
*/
CREATE TABLE ticket_affected_version_mapping
(
    id         TEXT    NOT NULL PRIMARY KEY UNIQUE,
    ticket_id  TEXT    NOT NULL,
    version_id TEXT    NOT NULL,
    created    INTEGER NOT NULL,
    deleted    BOOLEAN NOT NULL,
    UNIQUE (ticket_id, version_id)
);

CREATE INDEX ticket_affected_versions_get_by_ticket_id ON ticket_affected_version_mapping (ticket_id);
CREATE INDEX ticket_affected_versions_get_by_version_id ON ticket_affected_version_mapping (version_id);
CREATE INDEX ticket_affected_versions_get_by_deleted ON ticket_affected_version_mapping (deleted);
CREATE INDEX ticket_affected_versions_get_by_created ON ticket_affected_version_mapping (created);

/*
    Fix Versions

    Links tickets to versions where the issue will be/was fixed.
    A ticket can be fixed in multiple versions.
*/
CREATE TABLE ticket_fix_version_mapping
(
    id         TEXT    NOT NULL PRIMARY KEY UNIQUE,
    ticket_id  TEXT    NOT NULL,
    version_id TEXT    NOT NULL,
    created    INTEGER NOT NULL,
    deleted    BOOLEAN NOT NULL,
    UNIQUE (ticket_id, version_id)
);

CREATE INDEX ticket_fix_versions_get_by_ticket_id ON ticket_fix_version_mapping (ticket_id);
CREATE INDEX ticket_fix_versions_get_by_version_id ON ticket_fix_version_mapping (version_id);
CREATE INDEX ticket_fix_versions_get_by_deleted ON ticket_fix_version_mapping (deleted);
CREATE INDEX ticket_fix_versions_get_by_created ON ticket_fix_version_mapping (created);

/*
    Saved Filters

    Users can save custom search filters and share them with others.
    Filters contain a query (JSON format) that defines search criteria.
    Public filters are visible to all users.
    Favorite filters appear in user's quick access.
*/
CREATE TABLE filter
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL,
    description TEXT,
    owner_id    TEXT    NOT NULL,
    query       TEXT    NOT NULL,  -- JSON query structure
    is_public   BOOLEAN NOT NULL DEFAULT FALSE,
    is_favorite BOOLEAN NOT NULL DEFAULT FALSE,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL
);

CREATE INDEX filters_get_by_title ON filter (title);
CREATE INDEX filters_get_by_owner_id ON filter (owner_id);
CREATE INDEX filters_get_by_is_public ON filter (is_public);
CREATE INDEX filters_get_by_is_favorite ON filter (is_favorite);
CREATE INDEX filters_get_by_deleted ON filter (deleted);
CREATE INDEX filters_get_by_created ON filter (created);
CREATE INDEX filters_get_by_modified ON filter (modified);
CREATE INDEX filters_get_by_created_and_modified ON filter (created, modified);

/*
    Filter Sharing

    Allows filters to be shared with specific users, teams, or projects.
    NULL values mean different scopes:
    - user_id NULL, team_id NULL, project_id NULL = public filter
    - user_id set = shared with specific user
    - team_id set = shared with specific team
    - project_id set = shared with specific project
*/
CREATE TABLE filter_share_mapping
(
    id         TEXT    NOT NULL PRIMARY KEY UNIQUE,
    filter_id  TEXT    NOT NULL,
    user_id    TEXT,              -- NULL for non-user shares
    team_id    TEXT,              -- NULL for non-team shares
    project_id TEXT,              -- NULL for non-project shares
    created    INTEGER NOT NULL,
    deleted    BOOLEAN NOT NULL
);

CREATE INDEX filter_shares_get_by_filter_id ON filter_share_mapping (filter_id);
CREATE INDEX filter_shares_get_by_user_id ON filter_share_mapping (user_id);
CREATE INDEX filter_shares_get_by_team_id ON filter_share_mapping (team_id);
CREATE INDEX filter_shares_get_by_project_id ON filter_share_mapping (project_id);
CREATE INDEX filter_shares_get_by_deleted ON filter_share_mapping (deleted);
CREATE INDEX filter_shares_get_by_created ON filter_share_mapping (created);

/*
    Custom Fields

    Defines custom fields that can be added to tickets.
    Field types: 'text', 'number', 'date', 'datetime', 'select', 'multi_select',
                 'user', 'url', 'textarea', 'checkbox', 'radio'

    Global fields (project_id NULL) apply to all projects.
    Project-specific fields only apply to that project.

    Configuration is JSON and field-type specific:
    - For 'select'/'multi_select': {"options": [...]} - deprecated, use custom_field_option table
    - For 'number': {"min": 0, "max": 100, "step": 1}
    - For 'text': {"max_length": 255, "pattern": "regex"}
*/
CREATE TABLE custom_field
(
    id            TEXT    NOT NULL PRIMARY KEY UNIQUE,
    field_name    TEXT    NOT NULL,
    field_type    TEXT    NOT NULL,  -- text, number, date, select, multi_select, user, url, etc.
    description   TEXT,
    project_id    TEXT,               -- NULL for global fields
    is_required   BOOLEAN NOT NULL DEFAULT FALSE,
    default_value TEXT,
    configuration TEXT,               -- JSON for field-specific config
    created       INTEGER NOT NULL,
    modified      INTEGER NOT NULL,
    deleted       BOOLEAN NOT NULL
);

CREATE INDEX custom_fields_get_by_field_name ON custom_field (field_name);
CREATE INDEX custom_fields_get_by_field_type ON custom_field (field_type);
CREATE INDEX custom_fields_get_by_project_id ON custom_field (project_id);
CREATE INDEX custom_fields_get_by_is_required ON custom_field (is_required);
CREATE INDEX custom_fields_get_by_deleted ON custom_field (deleted);
CREATE INDEX custom_fields_get_by_created ON custom_field (created);
CREATE INDEX custom_fields_get_by_modified ON custom_field (modified);
CREATE INDEX custom_fields_get_by_created_and_modified ON custom_field (created, modified);

/*
    Custom Field Options

    For 'select' and 'multi_select' custom fields, defines the available options.
    Position determines the display order.
    is_default indicates which option(s) are selected by default.
*/
CREATE TABLE custom_field_option
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    custom_field_id TEXT    NOT NULL,
    value           TEXT    NOT NULL,
    display_value   TEXT    NOT NULL,  -- User-friendly display text
    position        INTEGER NOT NULL,  -- Display order
    is_default      BOOLEAN NOT NULL DEFAULT FALSE,
    created         INTEGER NOT NULL,
    modified        INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL
);

CREATE INDEX custom_field_options_get_by_custom_field_id ON custom_field_option (custom_field_id);
CREATE INDEX custom_field_options_get_by_value ON custom_field_option (value);
CREATE INDEX custom_field_options_get_by_position ON custom_field_option (position);
CREATE INDEX custom_field_options_get_by_deleted ON custom_field_option (deleted);
CREATE INDEX custom_field_options_get_by_created ON custom_field_option (created);

/*
    Ticket Custom Field Values

    Stores the actual values of custom fields for tickets.
    Value is stored as TEXT and must be parsed according to the field_type.

    For multi_select fields, value contains JSON array: ["option1", "option2"]
    For date/datetime fields, value contains Unix timestamp
    For user fields, value contains user_id
*/
CREATE TABLE ticket_custom_field_value
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    ticket_id       TEXT    NOT NULL,
    custom_field_id TEXT    NOT NULL,
    value           TEXT,              -- NULL for empty/unset fields
    created         INTEGER NOT NULL,
    modified        INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL,
    UNIQUE (ticket_id, custom_field_id)
);

CREATE INDEX ticket_custom_field_values_get_by_ticket_id ON ticket_custom_field_value (ticket_id);
CREATE INDEX ticket_custom_field_values_get_by_custom_field_id ON ticket_custom_field_value (custom_field_id);
CREATE INDEX ticket_custom_field_values_get_by_ticket_and_field ON ticket_custom_field_value (ticket_id, custom_field_id);
CREATE INDEX ticket_custom_field_values_get_by_deleted ON ticket_custom_field_value (deleted);
CREATE INDEX ticket_custom_field_values_get_by_created ON ticket_custom_field_value (created);
CREATE INDEX ticket_custom_field_values_get_by_modified ON ticket_custom_field_value (modified);
CREATE INDEX ticket_custom_field_values_get_by_created_and_modified ON ticket_custom_field_value (created, modified);

/*
    ========================================================================
    V1 TABLE ENHANCEMENTS - New Columns for Phase 1
    ========================================================================

    These ALTER TABLE statements would be in Migration.V1.2.sql for existing
    installations. For new installations using V2, these columns are included
    in the CREATE TABLE statements.

    For clarity, we document them here:

    ALTER TABLE ticket ADD COLUMN priority_id TEXT;
    ALTER TABLE ticket ADD COLUMN resolution_id TEXT;
    ALTER TABLE ticket ADD COLUMN assignee_id TEXT;
    ALTER TABLE ticket ADD COLUMN reporter_id TEXT;
    ALTER TABLE ticket ADD COLUMN due_date INTEGER;
    ALTER TABLE ticket ADD COLUMN original_estimate INTEGER;  -- In minutes
    ALTER TABLE ticket ADD COLUMN remaining_estimate INTEGER; -- In minutes
    ALTER TABLE ticket ADD COLUMN time_spent INTEGER;         -- In minutes

    ALTER TABLE project ADD COLUMN lead_user_id TEXT;
    ALTER TABLE project ADD COLUMN default_assignee_id TEXT;

    CREATE INDEX tickets_get_by_priority_id ON ticket (priority_id);
    CREATE INDEX tickets_get_by_resolution_id ON ticket (resolution_id);
    CREATE INDEX tickets_get_by_assignee_id ON ticket (assignee_id);
    CREATE INDEX tickets_get_by_reporter_id ON ticket (reporter_id);
    CREATE INDEX tickets_get_by_due_date ON ticket (due_date);

    CREATE INDEX projects_get_by_lead_user_id ON project (lead_user_id);
*/

/*
    ========================================================================
    SEED DATA - Default Priorities
    ========================================================================
*/

-- INSERT INTO priority (id, title, description, level, icon, color, created, modified, deleted)
-- VALUES
--     ('priority-lowest', 'Lowest', 'Lowest priority', 1, 'arrow_downward', '#0747A6', strftime('%s', 'now'), strftime('%s', 'now'), 0),
--     ('priority-low', 'Low', 'Low priority', 2, 'keyboard_arrow_down', '#2684FF', strftime('%s', 'now'), strftime('%s', 'now'), 0),
--     ('priority-medium', 'Medium', 'Medium priority', 3, 'drag_handle', '#FFAB00', strftime('%s', 'now'), strftime('%s', 'now'), 0),
--     ('priority-high', 'High', 'High priority', 4, 'keyboard_arrow_up', '#FF8B00', strftime('%s', 'now'), strftime('%s', 'now'), 0),
--     ('priority-highest', 'Highest', 'Highest priority', 5, 'arrow_upward', '#DE350B', strftime('%s', 'now'), strftime('%s', 'now'), 0);

/*
    ========================================================================
    SEED DATA - Default Resolutions
    ========================================================================
*/

-- INSERT INTO resolution (id, title, description, created, modified, deleted)
-- VALUES
--     ('resolution-fixed', 'Fixed', 'A fix for this issue is checked into the tree and tested.', strftime('%s', 'now'), strftime('%s', 'now'), 0),
--     ('resolution-wont-fix', 'Won''t Fix', 'The problem described is an issue which will never be fixed.', strftime('%s', 'now'), strftime('%s', 'now'), 0),
--     ('resolution-duplicate', 'Duplicate', 'The problem is a duplicate of an existing issue.', strftime('%s', 'now'), strftime('%s', 'now'), 0),
--     ('resolution-incomplete', 'Incomplete', 'The problem is not completely described.', strftime('%s', 'now'), strftime('%s', 'now'), 0),
--     ('resolution-cannot-reproduce', 'Cannot Reproduce', 'Attempts at reproducing this issue failed, or not enough information was available.', strftime('%s', 'now'), strftime('%s', 'now'), 0),
--     ('resolution-done', 'Done', 'Work has been completed on this issue.', strftime('%s', 'now'), strftime('%s', 'now'), 0);

/*
    ========================================================================
    VERSION INFORMATION
    ========================================================================
*/

-- UPDATE system_info SET description = 'Database schema version 2 - JIRA Feature Parity Phase 1'
-- WHERE id = (SELECT id FROM system_info ORDER BY created DESC LIMIT 1);

/*
    ========================================================================
    NOTES FOR DEVELOPERS
    ========================================================================

    1. All new tables follow the existing patterns from V1
    2. UUID strings are used for all identifiers
    3. created, modified, deleted columns are standard
    4. Mapping tables establish many-to-many relationships
    5. Indexes are created for common query patterns
    6. Foreign key relationships are maintained at application level (Go code)
    7. JSON is used for flexible configuration data
    8. Unix timestamps (INTEGER) are used for all dates
    9. Boolean fields use INTEGER (0/1) for SQLite compatibility

    MIGRATION CONSIDERATIONS:

    1. Existing tickets will have NULL values for new fields
    2. Default priorities/resolutions should be created
    3. Custom fields migration from ticket_meta_data should be planned
    4. All new indexes must be created
    5. Performance impact should be monitored

    BACKWARD COMPATIBILITY:

    1. All V1 tables remain unchanged in structure
    2. New columns are nullable to support existing data
    3. Existing queries will continue to work
    4. API version 2 adds new endpoints, v1 endpoints remain functional
*/

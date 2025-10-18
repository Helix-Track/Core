/*
    Migration Script: Documents V1 → V2

    This script migrates the Documents extension from V1 (basic) to V2 (Confluence parity)

    CAUTION: This is a major upgrade that significantly changes the schema.
    Please backup your database before running this migration!

    Changes:
    - Adds 25 new tables
    - Enhances existing document and content tables
    - Preserves all existing data
    - Creates default document types and spaces

    Migration Steps:
    1. Backup existing database
    2. Create new tables
    3. Migrate existing data
    4. Add new columns to existing tables
    5. Create default data
    6. Verify migration
*/

/*
    ========================================================================
    STEP 1: Create new V2 tables
    ========================================================================
*/

-- Document spaces
CREATE TABLE IF NOT EXISTS document_space
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    key             TEXT    NOT NULL UNIQUE,
    name            TEXT    NOT NULL,
    description     TEXT,
    owner_id        TEXT    NOT NULL,
    is_public       BOOLEAN NOT NULL DEFAULT 0,
    created         INTEGER NOT NULL,
    modified        INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL DEFAULT 0
);

-- Document types
CREATE TABLE IF NOT EXISTS document_type
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    key             TEXT    NOT NULL UNIQUE,
    name            TEXT    NOT NULL,
    description     TEXT,
    icon            TEXT,
    schema_json     TEXT,
    created         INTEGER NOT NULL,
    modified        INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL DEFAULT 0
);

-- Versioning tables
CREATE TABLE IF NOT EXISTS document_version
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    document_id     TEXT    NOT NULL,
    version_number  INTEGER NOT NULL,
    user_id         TEXT    NOT NULL,
    change_summary  TEXT,
    is_major        BOOLEAN NOT NULL DEFAULT 0,
    is_minor        BOOLEAN NOT NULL DEFAULT 1,
    snapshot_json   TEXT,
    content_id      TEXT,
    created         INTEGER NOT NULL,
    UNIQUE(document_id, version_number)
);

CREATE TABLE IF NOT EXISTS document_version_label
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    version_id      TEXT    NOT NULL,
    label           TEXT    NOT NULL,
    description     TEXT,
    user_id         TEXT    NOT NULL,
    created         INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS document_version_tag
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    version_id      TEXT    NOT NULL,
    tag             TEXT    NOT NULL,
    user_id         TEXT    NOT NULL,
    created         INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS document_version_comment
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    version_id      TEXT    NOT NULL,
    user_id         TEXT    NOT NULL,
    comment         TEXT    NOT NULL,
    created         INTEGER NOT NULL,
    modified        INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS document_version_mention
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    version_id      TEXT    NOT NULL,
    mentioned_user_id TEXT  NOT NULL,
    mentioning_user_id TEXT NOT NULL,
    context         TEXT,
    created         INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS document_version_diff
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    document_id     TEXT    NOT NULL,
    from_version    INTEGER NOT NULL,
    to_version      INTEGER NOT NULL,
    diff_type       TEXT    NOT NULL,
    diff_content    TEXT    NOT NULL,
    created         INTEGER NOT NULL,
    UNIQUE(document_id, from_version, to_version, diff_type)
);

-- Collaboration tables
CREATE TABLE IF NOT EXISTS document_comment
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    document_id     TEXT    NOT NULL,
    user_id         TEXT    NOT NULL,
    content         TEXT    NOT NULL,
    parent_id       TEXT,
    version         INTEGER NOT NULL DEFAULT 1,
    is_resolved     BOOLEAN NOT NULL DEFAULT 0,
    created         INTEGER NOT NULL,
    modified        INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS document_comment_thread
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    document_id     TEXT    NOT NULL,
    root_comment_id TEXT    NOT NULL,
    parent_comment_id TEXT  NOT NULL,
    child_comment_id TEXT   NOT NULL,
    thread_depth    INTEGER NOT NULL DEFAULT 0,
    created         INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS document_inline_comment
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    document_id     TEXT    NOT NULL,
    comment_id      TEXT    NOT NULL,
    position_start  INTEGER NOT NULL,
    position_end    INTEGER NOT NULL,
    selected_text   TEXT,
    is_resolved     BOOLEAN NOT NULL DEFAULT 0,
    created         INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS document_mention
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    document_id     TEXT    NOT NULL,
    mentioned_user_id TEXT  NOT NULL,
    mentioning_user_id TEXT NOT NULL,
    mention_context TEXT,
    position        INTEGER,
    is_acknowledged BOOLEAN NOT NULL DEFAULT 0,
    created         INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS document_reaction
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    document_id     TEXT    NOT NULL,
    user_id         TEXT    NOT NULL,
    reaction_type   TEXT    NOT NULL,
    emoji           TEXT,
    created         INTEGER NOT NULL,
    UNIQUE(document_id, user_id, reaction_type)
);

CREATE TABLE IF NOT EXISTS document_watcher
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    document_id     TEXT    NOT NULL,
    user_id         TEXT    NOT NULL,
    notification_level TEXT NOT NULL DEFAULT 'all',
    created         INTEGER NOT NULL,
    UNIQUE(document_id, user_id)
);

-- Organization tables
CREATE TABLE IF NOT EXISTS document_label
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    name            TEXT    NOT NULL UNIQUE,
    description     TEXT,
    color           TEXT,
    created         INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS document_tag
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    name            TEXT    NOT NULL UNIQUE,
    created         INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS document_label_mapping
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    document_id     TEXT    NOT NULL,
    label_id        TEXT    NOT NULL,
    user_id         TEXT    NOT NULL,
    created         INTEGER NOT NULL,
    UNIQUE(document_id, label_id)
);

CREATE TABLE IF NOT EXISTS document_tag_mapping
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    document_id     TEXT    NOT NULL,
    tag_id          TEXT    NOT NULL,
    user_id         TEXT    NOT NULL,
    created         INTEGER NOT NULL,
    UNIQUE(document_id, tag_id)
);

-- Entity connection tables
CREATE TABLE IF NOT EXISTS document_entity_link
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    document_id     TEXT    NOT NULL,
    entity_type     TEXT    NOT NULL,
    entity_id       TEXT    NOT NULL,
    link_type       TEXT    NOT NULL DEFAULT 'related',
    description     TEXT,
    user_id         TEXT    NOT NULL,
    created         INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL DEFAULT 0,
    UNIQUE(document_id, entity_type, entity_id)
);

CREATE TABLE IF NOT EXISTS document_relationship
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    source_document_id TEXT NOT NULL,
    target_document_id TEXT NOT NULL,
    relationship_type TEXT  NOT NULL,
    user_id         TEXT    NOT NULL,
    created         INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL DEFAULT 0,
    UNIQUE(source_document_id, target_document_id, relationship_type)
);

-- Template tables
CREATE TABLE IF NOT EXISTS document_template
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    name            TEXT    NOT NULL,
    description     TEXT,
    space_id        TEXT,
    type_id         TEXT    NOT NULL,
    content_template TEXT   NOT NULL,
    variables_json  TEXT,
    creator_id      TEXT    NOT NULL,
    is_public       BOOLEAN NOT NULL DEFAULT 0,
    use_count       INTEGER NOT NULL DEFAULT 0,
    created         INTEGER NOT NULL,
    modified        INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS document_blueprint
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    name            TEXT    NOT NULL,
    description     TEXT,
    space_id        TEXT,
    template_id     TEXT    NOT NULL,
    wizard_steps_json TEXT,
    creator_id      TEXT    NOT NULL,
    is_public       BOOLEAN NOT NULL DEFAULT 0,
    created         INTEGER NOT NULL,
    modified        INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL DEFAULT 0
);

-- Analytics tables
CREATE TABLE IF NOT EXISTS document_view_history
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    document_id     TEXT    NOT NULL,
    user_id         TEXT,
    ip_address      TEXT,
    user_agent      TEXT,
    session_id      TEXT,
    view_duration   INTEGER,
    timestamp       INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS document_analytics
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    document_id     TEXT    NOT NULL UNIQUE,
    total_views     INTEGER NOT NULL DEFAULT 0,
    unique_viewers  INTEGER NOT NULL DEFAULT 0,
    total_edits     INTEGER NOT NULL DEFAULT 0,
    unique_editors  INTEGER NOT NULL DEFAULT 0,
    total_comments  INTEGER NOT NULL DEFAULT 0,
    total_reactions INTEGER NOT NULL DEFAULT 0,
    total_watchers  INTEGER NOT NULL DEFAULT 0,
    avg_view_duration INTEGER,
    last_viewed     INTEGER,
    last_edited     INTEGER,
    popularity_score REAL   NOT NULL DEFAULT 0.0,
    updated         INTEGER NOT NULL
);

-- Attachment tables
CREATE TABLE IF NOT EXISTS document_attachment
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    document_id     TEXT    NOT NULL,
    filename        TEXT    NOT NULL,
    original_filename TEXT  NOT NULL,
    mime_type       TEXT    NOT NULL,
    size_bytes      INTEGER NOT NULL,
    storage_path    TEXT    NOT NULL,
    checksum        TEXT    NOT NULL,
    uploader_id     TEXT    NOT NULL,
    description     TEXT,
    version         INTEGER NOT NULL DEFAULT 1,
    created         INTEGER NOT NULL,
    modified        INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL DEFAULT 0
);

/*
    ========================================================================
    STEP 2: Seed default document types
    ========================================================================
*/

INSERT OR IGNORE INTO document_type (id, key, name, description, icon, schema_json, created, modified, deleted)
VALUES
    ('dt-page', 'page', 'Page', 'Standard document page', 'document', '{}', strftime('%s', 'now'), strftime('%s', 'now'), 0),
    ('dt-blog', 'blog', 'Blog Post', 'Blog-style post', 'blog', '{}', strftime('%s', 'now'), strftime('%s', 'now'), 0),
    ('dt-template', 'template', 'Template', 'Document template', 'template', '{}', strftime('%s', 'now'), strftime('%s', 'now'), 0),
    ('dt-meeting', 'meeting', 'Meeting Notes', 'Meeting notes document', 'meeting', '{}', strftime('%s', 'now'), strftime('%s', 'now'), 0),
    ('dt-spec', 'specification', 'Specification', 'Technical specification', 'spec', '{}', strftime('%s', 'now'), strftime('%s', 'now'), 0);

/*
    ========================================================================
    STEP 3: Create default space for existing documents
    ========================================================================
*/

-- Create a default space for migrated documents
INSERT OR IGNORE INTO document_space (id, key, name, description, owner_id, is_public, created, modified, deleted)
VALUES ('ds-default', 'DEFAULT', 'Default Space', 'Default space for migrated documents', 'system', 1, strftime('%s', 'now'), strftime('%s', 'now'), 0);

/*
    ========================================================================
    STEP 4: Migrate existing document table
    ========================================================================
*/

-- Rename old table
ALTER TABLE document RENAME TO document_v1_backup;

-- Create new document table with all V2 columns
CREATE TABLE document
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title           TEXT    NOT NULL,
    space_id        TEXT    NOT NULL,
    parent_id       TEXT,
    type_id         TEXT    NOT NULL,
    project_id      TEXT,
    creator_id      TEXT    NOT NULL,
    version         INTEGER NOT NULL DEFAULT 1,
    position        INTEGER NOT NULL DEFAULT 0,
    is_published    BOOLEAN NOT NULL DEFAULT 0,
    is_archived     BOOLEAN NOT NULL DEFAULT 0,
    publish_date    INTEGER,
    created         INTEGER NOT NULL,
    modified        INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL DEFAULT 0
);

-- Migrate data from V1 to V2
INSERT INTO document (
    id, title, space_id, parent_id, type_id, project_id,
    creator_id, version, position, is_published, is_archived,
    publish_date, created, modified, deleted
)
SELECT
    id,
    title,
    'ds-default' AS space_id,                      -- Use default space
    document_id AS parent_id,                       -- V1's document_id becomes parent_id
    'dt-page' AS type_id,                          -- Default to 'page' type
    project_id,
    'system' AS creator_id,                        -- Default creator (should be updated manually)
    1 AS version,                                  -- Start at version 1
    0 AS position,
    1 AS is_published,                             -- Assume existing docs are published
    0 AS is_archived,
    created AS publish_date,
    created,
    modified,
    deleted
FROM document_v1_backup;

/*
    ========================================================================
    STEP 5: Migrate existing content_document_mapping table
    ========================================================================
*/

-- Rename old table
ALTER TABLE content_document_mapping RENAME TO content_document_mapping_v1_backup;

-- Create new document_content table
CREATE TABLE document_content
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    document_id     TEXT    NOT NULL,
    version         INTEGER NOT NULL,
    content_type    TEXT    NOT NULL,
    content         TEXT,
    content_hash    TEXT,
    size_bytes      INTEGER NOT NULL DEFAULT 0,
    created         INTEGER NOT NULL,
    modified        INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL DEFAULT 0,
    UNIQUE(document_id, version)
);

-- Migrate data from V1 to V2
INSERT INTO document_content (
    id, document_id, version, content_type, content,
    content_hash, size_bytes, created, modified, deleted
)
SELECT
    id,
    document_id,
    1 AS version,                                  -- Initial version
    'html' AS content_type,                        -- Default content type
    content,
    NULL AS content_hash,                          -- Will be computed later
    COALESCE(LENGTH(content), 0) AS size_bytes,
    created,
    modified,
    deleted
FROM content_document_mapping_v1_backup;

/*
    ========================================================================
    STEP 6: Create initial version records for migrated documents
    ========================================================================
*/

-- Create version 1 record for each migrated document
INSERT INTO document_version (
    id, document_id, version_number, user_id, change_summary,
    is_major, is_minor, snapshot_json, content_id, created
)
SELECT
    'dv-' || d.id AS id,
    d.id AS document_id,
    1 AS version_number,
    'system' AS user_id,
    'Initial version (migrated from V1)' AS change_summary,
    1 AS is_major,
    0 AS is_minor,
    NULL AS snapshot_json,
    c.id AS content_id,
    d.created
FROM document d
JOIN document_content c ON c.document_id = d.id AND c.version = 1;

/*
    ========================================================================
    STEP 7: Create indexes for new tables
    ========================================================================
*/

-- Document space indexes
CREATE INDEX IF NOT EXISTS document_spaces_get_by_key ON document_space (key);
CREATE INDEX IF NOT EXISTS document_spaces_get_by_owner ON document_space (owner_id);
CREATE INDEX IF NOT EXISTS document_spaces_get_by_deleted ON document_space (deleted);

-- Document type indexes
CREATE INDEX IF NOT EXISTS document_types_get_by_key ON document_type (key);

-- Document indexes
CREATE INDEX IF NOT EXISTS documents_get_by_title ON document (title);
CREATE INDEX IF NOT EXISTS documents_get_by_project_id ON document (project_id);
CREATE INDEX IF NOT EXISTS documents_get_by_space_id ON document (space_id);
CREATE INDEX IF NOT EXISTS documents_get_by_parent_id ON document (parent_id);
CREATE INDEX IF NOT EXISTS documents_get_by_type_id ON document (type_id);
CREATE INDEX IF NOT EXISTS documents_get_by_creator ON document (creator_id);
CREATE INDEX IF NOT EXISTS documents_get_by_deleted ON document (deleted);
CREATE INDEX IF NOT EXISTS documents_get_by_archived ON document (is_archived);
CREATE INDEX IF NOT EXISTS documents_get_by_published ON document (is_published);
CREATE INDEX IF NOT EXISTS documents_get_by_created ON document (created);
CREATE INDEX IF NOT EXISTS documents_get_by_modified ON document (modified);
CREATE INDEX IF NOT EXISTS documents_get_by_version ON document (version);

-- Document content indexes
CREATE INDEX IF NOT EXISTS document_content_get_by_document_id ON document_content (document_id);
CREATE INDEX IF NOT EXISTS document_content_get_by_version ON document_content (version);

-- Version indexes
CREATE INDEX IF NOT EXISTS document_versions_get_by_document_id ON document_version (document_id);
CREATE INDEX IF NOT EXISTS document_versions_get_by_user_id ON document_version (user_id);
CREATE INDEX IF NOT EXISTS document_versions_get_by_created ON document_version (created);
CREATE INDEX IF NOT EXISTS document_versions_get_by_version_number ON document_version (version_number);
CREATE INDEX IF NOT EXISTS document_version_labels_get_by_version_id ON document_version_label (version_id);
CREATE INDEX IF NOT EXISTS document_version_tags_get_by_version_id ON document_version_tag (version_id);
CREATE INDEX IF NOT EXISTS document_version_comments_get_by_version_id ON document_version_comment (version_id);
CREATE INDEX IF NOT EXISTS document_version_mentions_get_by_version_id ON document_version_mention (version_id);
CREATE INDEX IF NOT EXISTS document_version_mentions_get_by_user_id ON document_version_mention (mentioned_user_id);
CREATE INDEX IF NOT EXISTS document_version_diffs_get_by_from_version ON document_version_diff (from_version);
CREATE INDEX IF NOT EXISTS document_version_diffs_get_by_to_version ON document_version_diff (to_version);

-- Collaboration indexes
CREATE INDEX IF NOT EXISTS document_comments_get_by_document_id ON document_comment (document_id);
CREATE INDEX IF NOT EXISTS document_comments_get_by_user_id ON document_comment (user_id);
CREATE INDEX IF NOT EXISTS document_comments_get_by_created ON document_comment (created);
CREATE INDEX IF NOT EXISTS document_comment_threads_get_by_parent_id ON document_comment_thread (parent_comment_id);
CREATE INDEX IF NOT EXISTS document_inline_comments_get_by_document_id ON document_inline_comment (document_id);
CREATE INDEX IF NOT EXISTS document_inline_comments_get_by_position ON document_inline_comment (position_start, position_end);
CREATE INDEX IF NOT EXISTS document_mentions_get_by_document_id ON document_mention (document_id);
CREATE INDEX IF NOT EXISTS document_mentions_get_by_user_id ON document_mention (mentioned_user_id);
CREATE INDEX IF NOT EXISTS document_reactions_get_by_document_id ON document_reaction (document_id);
CREATE INDEX IF NOT EXISTS document_reactions_get_by_user_id ON document_reaction (user_id);
CREATE INDEX IF NOT EXISTS document_watchers_get_by_document_id ON document_watcher (document_id);
CREATE INDEX IF NOT EXISTS document_watchers_get_by_user_id ON document_watcher (user_id);

-- Organization indexes
CREATE INDEX IF NOT EXISTS document_labels_get_by_name ON document_label (name);
CREATE INDEX IF NOT EXISTS document_tags_get_by_name ON document_tag (name);
CREATE INDEX IF NOT EXISTS document_label_mappings_get_by_document_id ON document_label_mapping (document_id);
CREATE INDEX IF NOT EXISTS document_label_mappings_get_by_label_id ON document_label_mapping (label_id);
CREATE INDEX IF NOT EXISTS document_tag_mappings_get_by_document_id ON document_tag_mapping (document_id);
CREATE INDEX IF NOT EXISTS document_tag_mappings_get_by_tag_id ON document_tag_mapping (tag_id);

-- Entity link indexes
CREATE INDEX IF NOT EXISTS document_entity_links_get_by_document_id ON document_entity_link (document_id);
CREATE INDEX IF NOT EXISTS document_entity_links_get_by_entity_type ON document_entity_link (entity_type);
CREATE INDEX IF NOT EXISTS document_entity_links_get_by_entity_id ON document_entity_link (entity_id);
CREATE INDEX IF NOT EXISTS document_relationships_get_by_source ON document_relationship (source_document_id);
CREATE INDEX IF NOT EXISTS document_relationships_get_by_target ON document_relationship (target_document_id);

-- Template indexes
CREATE INDEX IF NOT EXISTS document_templates_get_by_space_id ON document_template (space_id);
CREATE INDEX IF NOT EXISTS document_templates_get_by_creator ON document_template (creator_id);
CREATE INDEX IF NOT EXISTS document_blueprints_get_by_space_id ON document_blueprint (space_id);

-- Analytics indexes
CREATE INDEX IF NOT EXISTS document_view_history_get_by_document_id ON document_view_history (document_id);
CREATE INDEX IF NOT EXISTS document_view_history_get_by_user_id ON document_view_history (user_id);
CREATE INDEX IF NOT EXISTS document_view_history_get_by_timestamp ON document_view_history (timestamp);
CREATE INDEX IF NOT EXISTS document_analytics_get_by_document_id ON document_analytics (document_id);

-- Attachment indexes
CREATE INDEX IF NOT EXISTS document_attachments_get_by_document_id ON document_attachment (document_id);
CREATE INDEX IF NOT EXISTS document_attachments_get_by_uploader ON document_attachment (uploader_id);

/*
    ========================================================================
    STEP 8: Initialize analytics for migrated documents
    ========================================================================
*/

INSERT INTO document_analytics (
    id, document_id, total_views, unique_viewers, total_edits,
    unique_editors, total_comments, total_reactions, total_watchers,
    avg_view_duration, last_viewed, last_edited, popularity_score, updated
)
SELECT
    'da-' || id AS id,
    id AS document_id,
    0 AS total_views,
    0 AS unique_viewers,
    1 AS total_edits,                              -- Initial creation counts as 1 edit
    1 AS unique_editors,
    0 AS total_comments,
    0 AS total_reactions,
    0 AS total_watchers,
    NULL AS avg_view_duration,
    NULL AS last_viewed,
    modified AS last_edited,
    0.0 AS popularity_score,
    strftime('%s', 'now') AS updated
FROM document;

/*
    ========================================================================
    MIGRATION COMPLETE
    ========================================================================
*/

/*
    Post-Migration Steps:

    1. Verify data integrity:
       - SELECT COUNT(*) FROM document; (should match V1 count)
       - SELECT COUNT(*) FROM document_content; (should match V1 count)
       - SELECT COUNT(*) FROM document_version; (should match document count)

    2. Update creator_id for all documents (currently set to 'system')
       - UPDATE document SET creator_id = '<actual_user_id>' WHERE ...;

    3. Optionally drop backup tables once verified:
       - DROP TABLE document_v1_backup;
       - DROP TABLE content_document_mapping_v1_backup;

    4. Run VACUUM to reclaim space:
       - VACUUM;

    5. Update application code to use V2 API

    6. Test thoroughly before production deployment!
*/

-- Migration information
SELECT 'Documents Extension V1 → V2 Migration Complete' AS status,
       (SELECT COUNT(*) FROM document) AS documents_migrated,
       (SELECT COUNT(*) FROM document_content) AS content_records,
       (SELECT COUNT(*) FROM document_version) AS version_records,
       strftime('%s', 'now') AS migration_timestamp;

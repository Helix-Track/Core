/*
    Documents Extension - Version: 2

    Full Confluence Feature Parity for HelixTrack

    This version provides complete document management capabilities matching
    Atlassian Confluence with additional HelixTrack-specific integrations.

    Features:
    - Rich document management with hierarchy
    - Complete version control with comparison
    - Real-time collaboration (comments, mentions, reactions)
    - Advanced organization (labels, tags, spaces)
    - Multi-format export (PDF, Word, HTML, XML, Markdown)
    - Entity linking (connect docs to tickets, projects, users, etc.)
    - Templates and blueprints
    - Analytics and tracking
    - Attachment management

    Tables: 27 (upgraded from 2 in V1)
    API Actions: 90+

    Migration from V1: See Migration.V1.2.sql
*/

/*
    Notes:

    - The main project board: https://github.com/orgs/red-elf/projects/2/views/1
    - Identifiers in the system are UUID strings
    - Mapping tables are used for binding entities and defining relationships
    - Additional tables are defined to provide the meta-data to entities of the system
    - To follow the order of entities definition in the system follow the 'DROP TABLE' directives
    - All timestamps are Unix epoch (INTEGER)
    - Soft delete pattern: deleted BOOLEAN
    - Version tracking: version INTEGER (optimistic locking)
*/

/*
    ========================================================================
    DROP STATEMENTS - All V2 Tables
    ========================================================================
*/

-- Core tables
DROP TABLE IF EXISTS document_space;
DROP TABLE IF EXISTS document_type;
DROP TABLE IF EXISTS document;
DROP TABLE IF EXISTS document_content;

-- Versioning tables
DROP TABLE IF EXISTS document_version;
DROP TABLE IF EXISTS document_version_label;
DROP TABLE IF EXISTS document_version_tag;
DROP TABLE IF EXISTS document_version_comment;
DROP TABLE IF EXISTS document_version_mention;
DROP TABLE IF EXISTS document_version_diff;

-- Collaboration tables (document-specific only, comments/labels/votes use core V5 tables)
DROP TABLE IF EXISTS document_inline_comment;
DROP TABLE IF EXISTS document_watcher;

-- Organization tables (tags only, labels use core V5 tables)
DROP TABLE IF EXISTS document_tag;
DROP TABLE IF EXISTS document_tag_mapping;

-- Entity connection tables
DROP TABLE IF EXISTS document_entity_link;
DROP TABLE IF EXISTS document_relationship;

-- Template tables
DROP TABLE IF EXISTS document_template;
DROP TABLE IF EXISTS document_blueprint;

-- Analytics tables
DROP TABLE IF EXISTS document_view_history;
DROP TABLE IF EXISTS document_analytics;

-- Attachment tables
DROP TABLE IF EXISTS document_attachment;

-- Drop all indexes
DROP INDEX IF EXISTS document_spaces_get_by_key;
DROP INDEX IF EXISTS document_spaces_get_by_owner;
DROP INDEX IF EXISTS document_spaces_get_by_deleted;
DROP INDEX IF EXISTS document_types_get_by_key;
DROP INDEX IF EXISTS documents_get_by_title;
DROP INDEX IF EXISTS documents_get_by_project_id;
DROP INDEX IF EXISTS documents_get_by_space_id;
DROP INDEX IF EXISTS documents_get_by_parent_id;
DROP INDEX IF EXISTS documents_get_by_type_id;
DROP INDEX IF EXISTS documents_get_by_creator;
DROP INDEX IF EXISTS documents_get_by_deleted;
DROP INDEX IF EXISTS documents_get_by_archived;
DROP INDEX IF EXISTS documents_get_by_published;
DROP INDEX IF EXISTS documents_get_by_created;
DROP INDEX IF EXISTS documents_get_by_modified;
DROP INDEX IF EXISTS documents_get_by_version;
DROP INDEX IF EXISTS document_content_get_by_document_id;
DROP INDEX IF EXISTS document_content_get_by_version;
DROP INDEX IF EXISTS document_versions_get_by_document_id;
DROP INDEX IF EXISTS document_versions_get_by_user_id;
DROP INDEX IF EXISTS document_versions_get_by_created;
DROP INDEX IF EXISTS document_versions_get_by_version_number;
DROP INDEX IF EXISTS document_version_labels_get_by_version_id;
DROP INDEX IF EXISTS document_version_tags_get_by_version_id;
DROP INDEX IF EXISTS document_version_comments_get_by_version_id;
DROP INDEX IF EXISTS document_version_mentions_get_by_version_id;
DROP INDEX IF EXISTS document_version_mentions_get_by_user_id;
DROP INDEX IF EXISTS document_version_diffs_get_by_from_version;
DROP INDEX IF EXISTS document_version_diffs_get_by_to_version;
DROP INDEX IF EXISTS document_inline_comments_get_by_document_id;
DROP INDEX IF EXISTS document_inline_comments_get_by_position;
DROP INDEX IF EXISTS document_inline_comments_get_by_comment_id;
DROP INDEX IF EXISTS document_watchers_get_by_document_id;
DROP INDEX IF EXISTS document_watchers_get_by_user_id;
DROP INDEX IF EXISTS document_tags_get_by_name;
DROP INDEX IF EXISTS document_tag_mappings_get_by_document_id;
DROP INDEX IF EXISTS document_tag_mappings_get_by_tag_id;
DROP INDEX IF EXISTS document_entity_links_get_by_document_id;
DROP INDEX IF EXISTS document_entity_links_get_by_entity_type;
DROP INDEX IF EXISTS document_entity_links_get_by_entity_id;
DROP INDEX IF EXISTS document_relationships_get_by_source;
DROP INDEX IF EXISTS document_relationships_get_by_target;
DROP INDEX IF EXISTS document_templates_get_by_space_id;
DROP INDEX IF EXISTS document_templates_get_by_creator;
DROP INDEX IF EXISTS document_blueprints_get_by_space_id;
DROP INDEX IF EXISTS document_view_history_get_by_document_id;
DROP INDEX IF EXISTS document_view_history_get_by_user_id;
DROP INDEX IF EXISTS document_view_history_get_by_timestamp;
DROP INDEX IF EXISTS document_analytics_get_by_document_id;
DROP INDEX IF EXISTS document_attachments_get_by_document_id;
DROP INDEX IF EXISTS document_attachments_get_by_uploader;

/*
    ========================================================================
    CORE TABLES
    ========================================================================
*/

/*
    Document spaces (similar to Confluence spaces)
    Spaces organize documents into logical groups
*/
CREATE TABLE document_space
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    key             TEXT    NOT NULL UNIQUE,        -- Short identifier (e.g., "DOCS", "TECH")
    name            TEXT    NOT NULL,
    description     TEXT,
    owner_id        TEXT    NOT NULL,               -- User who owns the space
    is_public       BOOLEAN NOT NULL DEFAULT 0,     -- Public or private space
    created         INTEGER NOT NULL,
    modified        INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL DEFAULT 0
);

CREATE INDEX document_spaces_get_by_key ON document_space (key);
CREATE INDEX document_spaces_get_by_owner ON document_space (owner_id);
CREATE INDEX document_spaces_get_by_deleted ON document_space (deleted);

/*
    Document types (page, blog post, template, etc.)
*/
CREATE TABLE document_type
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    key             TEXT    NOT NULL UNIQUE,        -- "page", "blog", "template", etc.
    name            TEXT    NOT NULL,
    description     TEXT,
    icon            TEXT,                           -- Icon identifier
    schema_json     TEXT,                           -- JSON schema for this type
    created         INTEGER NOT NULL,
    modified        INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL DEFAULT 0
);

CREATE INDEX document_types_get_by_key ON document_type (key);

/*
    Documents (enhanced from V1)
    Main document table with full hierarchy support
*/
CREATE TABLE document
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title           TEXT    NOT NULL,
    space_id        TEXT    NOT NULL,               -- Space this document belongs to
    parent_id       TEXT,                           -- Parent document (for hierarchy)
    type_id         TEXT    NOT NULL,               -- Document type
    project_id      TEXT,                           -- Optional: Link to project
    creator_id      TEXT    NOT NULL,               -- User who created the document
    version         INTEGER NOT NULL DEFAULT 1,     -- Current version number (optimistic locking)
    position        INTEGER NOT NULL DEFAULT 0,     -- Position in hierarchy
    is_published    BOOLEAN NOT NULL DEFAULT 0,     -- Published or draft
    is_archived     BOOLEAN NOT NULL DEFAULT 0,     -- Archived documents
    publish_date    INTEGER,                        -- Scheduled or actual publish date
    created         INTEGER NOT NULL,
    modified        INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL DEFAULT 0,

    FOREIGN KEY (space_id) REFERENCES document_space(id),
    FOREIGN KEY (parent_id) REFERENCES document(id),
    FOREIGN KEY (type_id) REFERENCES document_type(id)
);

CREATE INDEX documents_get_by_title ON document (title);
CREATE INDEX documents_get_by_project_id ON document (project_id);
CREATE INDEX documents_get_by_space_id ON document (space_id);
CREATE INDEX documents_get_by_parent_id ON document (parent_id);
CREATE INDEX documents_get_by_type_id ON document (type_id);
CREATE INDEX documents_get_by_creator ON document (creator_id);
CREATE INDEX documents_get_by_deleted ON document (deleted);
CREATE INDEX documents_get_by_archived ON document (is_archived);
CREATE INDEX documents_get_by_published ON document (is_published);
CREATE INDEX documents_get_by_created ON document (created);
CREATE INDEX documents_get_by_modified ON document (modified);
CREATE INDEX documents_get_by_version ON document (version);

/*
    Document content (enhanced from V1)
    Stores the actual content for each version
*/
CREATE TABLE document_content
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    document_id     TEXT    NOT NULL,
    version         INTEGER NOT NULL,               -- Version this content belongs to
    content_type    TEXT    NOT NULL,               -- "html", "markdown", "plain", "storage"
    content         TEXT,                           -- The actual content
    content_hash    TEXT,                           -- SHA-256 hash for deduplication
    size_bytes      INTEGER NOT NULL DEFAULT 0,     -- Content size
    created         INTEGER NOT NULL,
    modified        INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL DEFAULT 0,

    UNIQUE(document_id, version),
    FOREIGN KEY (document_id) REFERENCES document(id)
);

CREATE INDEX document_content_get_by_document_id ON document_content (document_id);
CREATE INDEX document_content_get_by_version ON document_content (version);

/*
    ========================================================================
    VERSIONING TABLES
    ========================================================================
*/

/*
    Document versions
    Complete version history with snapshots
*/
CREATE TABLE document_version
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    document_id     TEXT    NOT NULL,
    version_number  INTEGER NOT NULL,               -- 1, 2, 3, etc.
    user_id         TEXT    NOT NULL,               -- User who created this version
    change_summary  TEXT,                           -- Description of changes
    is_major        BOOLEAN NOT NULL DEFAULT 0,     -- Major vs minor version
    is_minor        BOOLEAN NOT NULL DEFAULT 1,
    snapshot_json   TEXT,                           -- JSON snapshot of document state
    content_id      TEXT,                           -- Link to document_content
    created         INTEGER NOT NULL,

    UNIQUE(document_id, version_number),
    FOREIGN KEY (document_id) REFERENCES document(id),
    FOREIGN KEY (content_id) REFERENCES document_content(id)
);

CREATE INDEX document_versions_get_by_document_id ON document_version (document_id);
CREATE INDEX document_versions_get_by_user_id ON document_version (user_id);
CREATE INDEX document_versions_get_by_created ON document_version (created);
CREATE INDEX document_versions_get_by_version_number ON document_version (version_number);

/*
    Version labels
    Named versions (e.g., "v1.0", "Release 2024-Q1")
*/
CREATE TABLE document_version_label
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    version_id      TEXT    NOT NULL,
    label           TEXT    NOT NULL,
    description     TEXT,
    user_id         TEXT    NOT NULL,               -- User who added the label
    created         INTEGER NOT NULL,

    FOREIGN KEY (version_id) REFERENCES document_version(id)
);

CREATE INDEX document_version_labels_get_by_version_id ON document_version_label (version_id);

/*
    Version tags
    Tags for categorizing versions
*/
CREATE TABLE document_version_tag
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    version_id      TEXT    NOT NULL,
    tag             TEXT    NOT NULL,
    user_id         TEXT    NOT NULL,
    created         INTEGER NOT NULL,

    FOREIGN KEY (version_id) REFERENCES document_version(id)
);

CREATE INDEX document_version_tags_get_by_version_id ON document_version_tag (version_id);

/*
    Version comments
    Comments specifically about a version
*/
CREATE TABLE document_version_comment
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    version_id      TEXT    NOT NULL,
    user_id         TEXT    NOT NULL,
    comment         TEXT    NOT NULL,
    created         INTEGER NOT NULL,
    modified        INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL DEFAULT 0,

    FOREIGN KEY (version_id) REFERENCES document_version(id)
);

CREATE INDEX document_version_comments_get_by_version_id ON document_version_comment (version_id);

/*
    Version mentions
    User mentions in version change summaries
*/
CREATE TABLE document_version_mention
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    version_id      TEXT    NOT NULL,
    mentioned_user_id TEXT  NOT NULL,               -- User being mentioned
    mentioning_user_id TEXT NOT NULL,               -- User who mentioned
    context         TEXT,                           -- Context around the mention
    created         INTEGER NOT NULL,

    FOREIGN KEY (version_id) REFERENCES document_version(id)
);

CREATE INDEX document_version_mentions_get_by_version_id ON document_version_mention (version_id);
CREATE INDEX document_version_mentions_get_by_user_id ON document_version_mention (mentioned_user_id);

/*
    Version diffs
    Cached diffs between versions for fast comparison
*/
CREATE TABLE document_version_diff
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    document_id     TEXT    NOT NULL,
    from_version    INTEGER NOT NULL,
    to_version      INTEGER NOT NULL,
    diff_type       TEXT    NOT NULL,               -- "unified", "split", "html"
    diff_content    TEXT    NOT NULL,               -- The actual diff
    created         INTEGER NOT NULL,               -- When diff was generated

    UNIQUE(document_id, from_version, to_version, diff_type),
    FOREIGN KEY (document_id) REFERENCES document(id)
);

CREATE INDEX document_version_diffs_get_by_from_version ON document_version_diff (from_version);
CREATE INDEX document_version_diffs_get_by_to_version ON document_version_diff (to_version);

/*
    ========================================================================
    COLLABORATION TABLES
    ========================================================================

    NOTE: Document comments use core V5 comment table + comment_document_mapping
    NOTE: Document mentions use core V3 comment_mention_mapping
    NOTE: Document reactions/votes use core V5 vote_mapping
    NOTE: Document labels use core V1 label + V5 label_document_mapping
*/

/*
    Inline comments
    Comments attached to specific content positions
    This is document-specific because of position data
    Links to core comment table (via comment_id)
*/
CREATE TABLE document_inline_comment
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    document_id     TEXT    NOT NULL,
    comment_id      TEXT    NOT NULL,               -- Link to core comment table
    position_start  INTEGER NOT NULL,               -- Character position start
    position_end    INTEGER NOT NULL,               -- Character position end
    selected_text   TEXT,                           -- Text that was selected
    is_resolved     BOOLEAN NOT NULL DEFAULT 0,
    created         INTEGER NOT NULL,

    FOREIGN KEY (document_id) REFERENCES document(id)
);

CREATE INDEX document_inline_comments_get_by_document_id ON document_inline_comment (document_id);
CREATE INDEX document_inline_comments_get_by_position ON document_inline_comment (position_start, position_end);
CREATE INDEX document_inline_comments_get_by_comment_id ON document_inline_comment (comment_id);


/*
    Document watchers
    Users subscribed to document changes
*/
CREATE TABLE document_watcher
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    document_id     TEXT    NOT NULL,
    user_id         TEXT    NOT NULL,
    notification_level TEXT NOT NULL DEFAULT 'all', -- "all", "mentions", "none"
    created         INTEGER NOT NULL,

    UNIQUE(document_id, user_id),
    FOREIGN KEY (document_id) REFERENCES document(id)
);

CREATE INDEX document_watchers_get_by_document_id ON document_watcher (document_id);
CREATE INDEX document_watchers_get_by_user_id ON document_watcher (user_id);

/*
    ========================================================================
    ORGANIZATION TABLES
    ========================================================================

    NOTE: Document labels use core V1 label table + V5 label_document_mapping
*/

/*
    Document tags
    More flexible than labels, can be ad-hoc
*/
CREATE TABLE document_tag
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    name            TEXT    NOT NULL UNIQUE,
    created         INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL DEFAULT 0
);

CREATE INDEX document_tags_get_by_name ON document_tag (name);


/*
    Document-tag mapping
*/
CREATE TABLE document_tag_mapping
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    document_id     TEXT    NOT NULL,
    tag_id          TEXT    NOT NULL,
    user_id         TEXT    NOT NULL,
    created         INTEGER NOT NULL,

    UNIQUE(document_id, tag_id),
    FOREIGN KEY (document_id) REFERENCES document(id),
    FOREIGN KEY (tag_id) REFERENCES document_tag(id)
);

CREATE INDEX document_tag_mappings_get_by_document_id ON document_tag_mapping (document_id);
CREATE INDEX document_tag_mappings_get_by_tag_id ON document_tag_mapping (tag_id);

/*
    ========================================================================
    ENTITY CONNECTION TABLES
    ========================================================================
*/

/*
    Document-entity links
    Connect documents to ANY system entity
*/
CREATE TABLE document_entity_link
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    document_id     TEXT    NOT NULL,
    entity_type     TEXT    NOT NULL,               -- "ticket", "project", "user", "label", etc.
    entity_id       TEXT    NOT NULL,
    link_type       TEXT    NOT NULL DEFAULT 'related', -- "related", "references", "implements", etc.
    description     TEXT,
    user_id         TEXT    NOT NULL,               -- User who created the link
    created         INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL DEFAULT 0,

    UNIQUE(document_id, entity_type, entity_id),
    FOREIGN KEY (document_id) REFERENCES document(id)
);

CREATE INDEX document_entity_links_get_by_document_id ON document_entity_link (document_id);
CREATE INDEX document_entity_links_get_by_entity_type ON document_entity_link (entity_type);
CREATE INDEX document_entity_links_get_by_entity_id ON document_entity_link (entity_id);

/*
    Document relationships
    Document-to-document relationships
*/
CREATE TABLE document_relationship
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    source_document_id TEXT NOT NULL,
    target_document_id TEXT NOT NULL,
    relationship_type TEXT  NOT NULL,              -- "parent", "related", "supersedes", etc.
    user_id         TEXT    NOT NULL,
    created         INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL DEFAULT 0,

    UNIQUE(source_document_id, target_document_id, relationship_type),
    FOREIGN KEY (source_document_id) REFERENCES document(id),
    FOREIGN KEY (target_document_id) REFERENCES document(id)
);

CREATE INDEX document_relationships_get_by_source ON document_relationship (source_document_id);
CREATE INDEX document_relationships_get_by_target ON document_relationship (target_document_id);

/*
    ========================================================================
    TEMPLATE TABLES
    ========================================================================
*/

/*
    Document templates
    Reusable document templates
*/
CREATE TABLE document_template
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    name            TEXT    NOT NULL,
    description     TEXT,
    space_id        TEXT,                           -- Optional: Space-specific template
    type_id         TEXT    NOT NULL,
    content_template TEXT   NOT NULL,               -- Template content with placeholders
    variables_json  TEXT,                           -- JSON array of template variables
    creator_id      TEXT    NOT NULL,
    is_public       BOOLEAN NOT NULL DEFAULT 0,
    use_count       INTEGER NOT NULL DEFAULT 0,
    created         INTEGER NOT NULL,
    modified        INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL DEFAULT 0,

    FOREIGN KEY (space_id) REFERENCES document_space(id),
    FOREIGN KEY (type_id) REFERENCES document_type(id)
);

CREATE INDEX document_templates_get_by_space_id ON document_template (space_id);
CREATE INDEX document_templates_get_by_creator ON document_template (creator_id);

/*
    Document blueprints
    Multi-step template wizards
*/
CREATE TABLE document_blueprint
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    name            TEXT    NOT NULL,
    description     TEXT,
    space_id        TEXT,
    template_id     TEXT    NOT NULL,               -- Base template
    wizard_steps_json TEXT,                         -- JSON array of wizard steps
    creator_id      TEXT    NOT NULL,
    is_public       BOOLEAN NOT NULL DEFAULT 0,
    created         INTEGER NOT NULL,
    modified        INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL DEFAULT 0,

    FOREIGN KEY (space_id) REFERENCES document_space(id),
    FOREIGN KEY (template_id) REFERENCES document_template(id)
);

CREATE INDEX document_blueprints_get_by_space_id ON document_blueprint (space_id);

/*
    ========================================================================
    ANALYTICS TABLES
    ========================================================================
*/

/*
    Document view history
    Track every view of every document
*/
CREATE TABLE document_view_history
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    document_id     TEXT    NOT NULL,
    user_id         TEXT,                           -- NULL for anonymous views
    ip_address      TEXT,
    user_agent      TEXT,
    session_id      TEXT,
    view_duration   INTEGER,                        -- Seconds spent viewing
    timestamp       INTEGER NOT NULL,

    FOREIGN KEY (document_id) REFERENCES document(id)
);

CREATE INDEX document_view_history_get_by_document_id ON document_view_history (document_id);
CREATE INDEX document_view_history_get_by_user_id ON document_view_history (user_id);
CREATE INDEX document_view_history_get_by_timestamp ON document_view_history (timestamp);

/*
    Document analytics
    Aggregated analytics data
*/
CREATE TABLE document_analytics
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
    avg_view_duration INTEGER,                     -- Average seconds
    last_viewed     INTEGER,
    last_edited     INTEGER,
    popularity_score REAL   NOT NULL DEFAULT 0.0,  -- Calculated score
    updated         INTEGER NOT NULL,               -- When analytics were last updated

    FOREIGN KEY (document_id) REFERENCES document(id)
);

CREATE INDEX document_analytics_get_by_document_id ON document_analytics (document_id);

/*
    ========================================================================
    ATTACHMENT TABLES
    ========================================================================
*/

/*
    Document attachments
    Files attached to documents
*/
CREATE TABLE document_attachment
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    document_id     TEXT    NOT NULL,
    filename        TEXT    NOT NULL,
    original_filename TEXT  NOT NULL,
    mime_type       TEXT    NOT NULL,
    size_bytes      INTEGER NOT NULL,
    storage_path    TEXT    NOT NULL,               -- Path to file on disk/S3/etc.
    checksum        TEXT    NOT NULL,               -- SHA-256 checksum
    uploader_id     TEXT    NOT NULL,
    description     TEXT,
    version         INTEGER NOT NULL DEFAULT 1,     -- Attachment version
    created         INTEGER NOT NULL,
    modified        INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL DEFAULT 0,

    FOREIGN KEY (document_id) REFERENCES document(id)
);

CREATE INDEX document_attachments_get_by_document_id ON document_attachment (document_id);
CREATE INDEX document_attachments_get_by_uploader ON document_attachment (uploader_id);

/*
    ========================================================================
    SEED DATA - Default Types and Spaces
    ========================================================================
*/

-- Default document types
INSERT INTO document_type (id, key, name, description, icon, schema_json, created, modified, deleted)
VALUES
    ('dt-page', 'page', 'Page', 'Standard document page', 'document', '{}', strftime('%s', 'now'), strftime('%s', 'now'), 0),
    ('dt-blog', 'blog', 'Blog Post', 'Blog-style post', 'blog', '{}', strftime('%s', 'now'), strftime('%s', 'now'), 0),
    ('dt-template', 'template', 'Template', 'Document template', 'template', '{}', strftime('%s', 'now'), strftime('%s', 'now'), 0),
    ('dt-meeting', 'meeting', 'Meeting Notes', 'Meeting notes document', 'meeting', '{}', strftime('%s', 'now'), strftime('%s', 'now'), 0),
    ('dt-spec', 'specification', 'Specification', 'Technical specification', 'spec', '{}', strftime('%s', 'now'), strftime('%s', 'now'), 0);

/*
    ========================================================================
    VERSION INFORMATION
    ========================================================================
*/

/*
    Documents Extension V2 - Confluence Feature Parity

    NEW IN VERSION 2:

    Tables: 21 document-specific (from 2 in V1)
    - 4 core tables (enhanced)
    - 6 versioning tables
    - 2 collaboration tables (inline comments, watchers)
    - 2 organization tables (tags only - labels use core)
    - 2 entity connection tables
    - 2 template tables
    - 2 analytics tables
    - 1 attachment table

    Reuses Core Entities (via V5):
    - Comments: core comment + comment_document_mapping
    - Labels: core label + label_document_mapping
    - Votes/Reactions: vote_mapping
    - Mentions: core comment_mention_mapping

    Features:
    - Complete version control with comparison
    - Real-time collaboration (via core entities)
    - Advanced organization (labels, tags, spaces)
    - Entity linking to entire system
    - Templates and blueprints
    - Multi-format export support
    - Analytics and tracking
    - Attachment management

    API Actions: 90+
    - Core document: 20
    - Versioning: 15
    - Collaboration: 12 (using core entities)
    - Organization: 10 (tags + core labels)
    - Export: 8
    - Entity connections: 8
    - Templates: 7
    - Analytics: 5
    - Attachments: 5

    Integration: Requires Core V5 for generic mapping tables
    Migration: See Migration.V1.2.sql
    Full documentation: See CONFLUENCE_PARITY_ANALYSIS.md
*/

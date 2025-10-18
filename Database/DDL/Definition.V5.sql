/*
    Version: 5

    Documents Extension Integration & Cross-Entity Linking

    This version adds deep integration between the Documents extension (V2)
    and the core HelixTrack system, enabling seamless document management
    across all entities.

    Features:
    - Ticket-Document deep integration
    - Project documentation management
    - User personal document spaces
    - Team knowledge bases
    - Enhanced search across documents and entities
    - Document templates for tickets and projects
    - Automated documentation generation
    - Cross-entity document workflows
*/

/*
    Notes:

    - This builds upon Definition.V4.sql (which includes V1, V2, V3)
    - Requires Documents Extension V2
    - All V1, V2, V3, and V4 tables are included
    - New tables for document-entity integration
    - Migration from V4 to V5 available in Migration.V4.5.sql
    - Identifiers in the system are UUID strings
    - Mapping tables are used for binding entities and defining relationships
*/

/*
    ========================================================================
    DROP STATEMENTS - Version 5 Tables
    ========================================================================
*/

-- Generic mapping tables (reusing core entities)
DROP TABLE IF EXISTS comment_document_mapping;
DROP TABLE IF EXISTS label_document_mapping;
DROP TABLE IF EXISTS vote_mapping;

-- Core integration tables
DROP TABLE IF EXISTS entity_document_mapping;
DROP TABLE IF EXISTS project_document_template_mapping;
DROP TABLE IF EXISTS ticket_documentation_requirement;

-- Enhanced workflow tables
DROP TABLE IF EXISTS workflow_documentation_step;
DROP TABLE IF EXISTS automated_documentation_rule;

-- Knowledge base tables
DROP TABLE IF EXISTS team_knowledge_base;
DROP TABLE IF EXISTS project_wiki;

-- Search enhancement tables
DROP TABLE IF EXISTS cross_entity_search_index;

-- Generic mapping table indexes
DROP INDEX IF EXISTS comment_document_mappings_get_by_comment_id;
DROP INDEX IF EXISTS comment_document_mappings_get_by_document_id;
DROP INDEX IF EXISTS comment_document_mappings_get_by_user_id;
DROP INDEX IF EXISTS label_document_mappings_get_by_label_id;
DROP INDEX IF EXISTS label_document_mappings_get_by_document_id;
DROP INDEX IF EXISTS label_document_mappings_get_by_user_id;
DROP INDEX IF EXISTS vote_mappings_get_by_entity_type;
DROP INDEX IF EXISTS vote_mappings_get_by_entity_id;
DROP INDEX IF EXISTS vote_mappings_get_by_user_id;
DROP INDEX IF EXISTS vote_mappings_get_by_vote_type;

-- Drop all V5 indexes
DROP INDEX IF EXISTS entity_document_mappings_get_by_entity_type;
DROP INDEX IF EXISTS entity_document_mappings_get_by_entity_id;
DROP INDEX IF EXISTS entity_document_mappings_get_by_document_id;
DROP INDEX IF EXISTS entity_document_mappings_get_by_relationship;
DROP INDEX IF EXISTS entity_document_mappings_get_by_created;
DROP INDEX IF EXISTS project_doc_templates_get_by_project_id;
DROP INDEX IF EXISTS project_doc_templates_get_by_template_id;
DROP INDEX IF EXISTS project_doc_templates_get_by_is_default;
DROP INDEX IF EXISTS ticket_doc_requirements_get_by_ticket_id;
DROP INDEX IF EXISTS ticket_doc_requirements_get_by_ticket_type_id;
DROP INDEX IF EXISTS ticket_doc_requirements_get_by_is_mandatory;
DROP INDEX IF EXISTS workflow_doc_steps_get_by_workflow_id;
DROP INDEX IF EXISTS workflow_doc_steps_get_by_status_id;
DROP INDEX IF EXISTS workflow_doc_steps_get_by_is_required;
DROP INDEX IF EXISTS auto_doc_rules_get_by_trigger_type;
DROP INDEX IF EXISTS auto_doc_rules_get_by_entity_type;
DROP INDEX IF EXISTS auto_doc_rules_get_by_is_active;
DROP INDEX IF EXISTS team_kb_get_by_team_id;
DROP INDEX IF EXISTS team_kb_get_by_space_id;
DROP INDEX IF EXISTS project_wiki_get_by_project_id;
DROP INDEX IF EXISTS project_wiki_get_by_space_id;
DROP INDEX IF EXISTS cross_entity_search_get_by_entity_type;
DROP INDEX IF EXISTS cross_entity_search_get_by_entity_id;
DROP INDEX IF EXISTS cross_entity_search_get_by_content_hash;

/*
    ========================================================================
    GENERIC MAPPING TABLES - Reusing Core Entities
    ========================================================================
*/

/*
    Comment-Document mapping
    Links core comment table to documents instead of duplicating comment functionality
    This allows comments to work consistently across tickets, documents, assets, etc.
*/
CREATE TABLE comment_document_mapping
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    comment_id  TEXT    NOT NULL,
    document_id TEXT    NOT NULL,
    user_id     TEXT    NOT NULL,               -- User who added comment to document
    is_resolved BOOLEAN NOT NULL DEFAULT 0,     -- For task/todo comments
    created     INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL DEFAULT 0,

    UNIQUE(comment_id, document_id),
    FOREIGN KEY (comment_id) REFERENCES comment(id)
);

CREATE INDEX comment_document_mappings_get_by_comment_id ON comment_document_mapping (comment_id);
CREATE INDEX comment_document_mappings_get_by_document_id ON comment_document_mapping (document_id);
CREATE INDEX comment_document_mappings_get_by_user_id ON comment_document_mapping (user_id);

/*
    Label-Document mapping
    Links core label table to documents for consistent labeling across the system
*/
CREATE TABLE label_document_mapping
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    label_id    TEXT    NOT NULL,
    document_id TEXT    NOT NULL,
    user_id     TEXT    NOT NULL,               -- User who added the label
    created     INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL DEFAULT 0,

    UNIQUE(label_id, document_id),
    FOREIGN KEY (label_id) REFERENCES label(id)
);

CREATE INDEX label_document_mappings_get_by_label_id ON label_document_mapping (label_id);
CREATE INDEX label_document_mappings_get_by_document_id ON label_document_mapping (document_id);
CREATE INDEX label_document_mappings_get_by_user_id ON label_document_mapping (user_id);

/*
    Generic vote/reaction mapping
    Replaces ticket_vote_mapping with a universal voting system
    Supports votes on any entity: tickets, documents, comments, etc.
*/
CREATE TABLE vote_mapping
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    entity_type TEXT    NOT NULL,               -- "ticket", "document", "comment", etc.
    entity_id   TEXT    NOT NULL,
    user_id     TEXT    NOT NULL,
    vote_type   TEXT    NOT NULL DEFAULT 'upvote', -- "upvote", "downvote", "like", "love", etc.
    emoji       TEXT,                           -- Optional emoji representation
    created     INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL DEFAULT 0,

    UNIQUE(entity_type, entity_id, user_id, vote_type)
);

CREATE INDEX vote_mappings_get_by_entity_type ON vote_mapping (entity_type);
CREATE INDEX vote_mappings_get_by_entity_id ON vote_mapping (entity_id);
CREATE INDEX vote_mappings_get_by_user_id ON vote_mapping (user_id);
CREATE INDEX vote_mappings_get_by_vote_type ON vote_mapping (vote_type);

/*
    ========================================================================
    CORE INTEGRATION TABLES
    ========================================================================
*/

/*
    Entity-Document mappings
    Universal table linking ANY entity to documents with rich metadata
    This complements Documents Extension's document_entity_link but provides
    core-side tracking and enhanced relationship types
*/
CREATE TABLE entity_document_mapping
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    entity_type     TEXT    NOT NULL,               -- "ticket", "project", "user", "team", "sprint", "board", etc.
    entity_id       TEXT    NOT NULL,
    document_id     TEXT    NOT NULL,               -- Reference to Documents Extension document
    relationship    TEXT    NOT NULL,               -- "specification", "requirements", "notes", "documentation", "reference"
    is_required     BOOLEAN NOT NULL DEFAULT 0,     -- Is this document required for the entity?
    is_primary      BOOLEAN NOT NULL DEFAULT 0,     -- Is this the primary document for the entity?
    visibility      TEXT    NOT NULL DEFAULT 'all', -- "all", "team", "project", "private"
    order_position  INTEGER NOT NULL DEFAULT 0,     -- Display order
    user_id         TEXT    NOT NULL,               -- User who created the mapping
    created         INTEGER NOT NULL,
    modified        INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL DEFAULT 0,

    UNIQUE(entity_type, entity_id, document_id)
);

CREATE INDEX entity_document_mappings_get_by_entity_type ON entity_document_mapping (entity_type);
CREATE INDEX entity_document_mappings_get_by_entity_id ON entity_document_mapping (entity_id);
CREATE INDEX entity_document_mappings_get_by_document_id ON entity_document_mapping (document_id);
CREATE INDEX entity_document_mappings_get_by_relationship ON entity_document_mapping (relationship);
CREATE INDEX entity_document_mappings_get_by_created ON entity_document_mapping (created);

/*
    Project-Document template mappings
    Define which document templates are available/required for projects
*/
CREATE TABLE project_document_template_mapping
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    project_id      TEXT    NOT NULL,
    template_id     TEXT    NOT NULL,               -- Reference to Documents Extension document_template
    template_type   TEXT    NOT NULL,               -- "ticket_spec", "release_notes", "meeting_notes", etc.
    is_default      BOOLEAN NOT NULL DEFAULT 0,     -- Auto-apply this template
    is_required     BOOLEAN NOT NULL DEFAULT 0,     -- Must use this template
    applicable_to   TEXT,                           -- JSON: Which ticket types/workflows this applies to
    created         INTEGER NOT NULL,
    modified        INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL DEFAULT 0,

    UNIQUE(project_id, template_id),
    FOREIGN KEY (project_id) REFERENCES project(id)
);

CREATE INDEX project_doc_templates_get_by_project_id ON project_document_template_mapping (project_id);
CREATE INDEX project_doc_templates_get_by_template_id ON project_document_template_mapping (template_id);
CREATE INDEX project_doc_templates_get_by_is_default ON project_document_template_mapping (is_default);

/*
    Ticket documentation requirements
    Define documentation requirements for ticket types/workflows
*/
CREATE TABLE ticket_documentation_requirement
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    ticket_type_id  TEXT    NOT NULL,
    workflow_id     TEXT,                           -- Optional: specific workflow
    document_type   TEXT    NOT NULL,               -- "specification", "test_plan", "deployment_guide", etc.
    template_id     TEXT,                           -- Optional: required template
    is_mandatory    BOOLEAN NOT NULL DEFAULT 0,     -- Must have this doc type
    min_documents   INTEGER NOT NULL DEFAULT 0,     -- Minimum number of documents
    max_documents   INTEGER,                        -- Maximum number (NULL = unlimited)
    description     TEXT,
    created         INTEGER NOT NULL,
    modified        INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL DEFAULT 0,

    FOREIGN KEY (ticket_type_id) REFERENCES ticket_type(id)
);

CREATE INDEX ticket_doc_requirements_get_by_ticket_id ON ticket_documentation_requirement (ticket_type_id);
CREATE INDEX ticket_doc_requirements_get_by_ticket_type_id ON ticket_documentation_requirement (ticket_type_id);
CREATE INDEX ticket_doc_requirements_get_by_is_mandatory ON ticket_documentation_requirement (is_mandatory);

/*
    ========================================================================
    ENHANCED WORKFLOW TABLES
    ========================================================================
*/

/*
    Workflow documentation steps
    Require specific documentation at workflow transitions
*/
CREATE TABLE workflow_documentation_step
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    workflow_id     TEXT    NOT NULL,
    from_status_id  TEXT    NOT NULL,
    to_status_id    TEXT    NOT NULL,
    document_type   TEXT    NOT NULL,               -- Required document type for transition
    template_id     TEXT,                           -- Optional: template to use
    is_required     BOOLEAN NOT NULL DEFAULT 1,     -- Block transition if missing
    validation_rule TEXT,                           -- JSON: Validation rules for document
    created         INTEGER NOT NULL,
    modified        INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL DEFAULT 0,

    FOREIGN KEY (from_status_id) REFERENCES ticket_status(id),
    FOREIGN KEY (to_status_id) REFERENCES ticket_status(id)
);

CREATE INDEX workflow_doc_steps_get_by_workflow_id ON workflow_documentation_step (workflow_id);
CREATE INDEX workflow_doc_steps_get_by_status_id ON workflow_documentation_step (from_status_id, to_status_id);
CREATE INDEX workflow_doc_steps_get_by_is_required ON workflow_documentation_step (is_required);

/*
    Automated documentation rules
    Auto-generate documents based on events
*/
CREATE TABLE automated_documentation_rule
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    name            TEXT    NOT NULL,
    description     TEXT,
    trigger_type    TEXT    NOT NULL,               -- "ticket_created", "ticket_closed", "sprint_completed", etc.
    entity_type     TEXT    NOT NULL,               -- "ticket", "project", "sprint", "board"
    template_id     TEXT    NOT NULL,               -- Template to use for generation
    target_space_id TEXT,                           -- Where to create the document
    condition_json  TEXT,                           -- JSON: Conditions for triggering
    mapping_json    TEXT,                           -- JSON: How to map entity data to document
    is_active       BOOLEAN NOT NULL DEFAULT 1,
    created_by      TEXT    NOT NULL,
    created         INTEGER NOT NULL,
    modified        INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL DEFAULT 0
);

CREATE INDEX auto_doc_rules_get_by_trigger_type ON automated_documentation_rule (trigger_type);
CREATE INDEX auto_doc_rules_get_by_entity_type ON automated_documentation_rule (entity_type);
CREATE INDEX auto_doc_rules_get_by_is_active ON automated_documentation_rule (is_active);

/*
    ========================================================================
    KNOWLEDGE BASE TABLES
    ========================================================================
*/

/*
    Team knowledge bases
    Link teams to their document spaces
*/
CREATE TABLE team_knowledge_base
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    team_id         TEXT    NOT NULL UNIQUE,
    space_id        TEXT    NOT NULL,               -- Reference to Documents Extension document_space
    is_default      BOOLEAN NOT NULL DEFAULT 1,     -- Default KB for team
    description     TEXT,
    created         INTEGER NOT NULL,
    modified        INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL DEFAULT 0,

    FOREIGN KEY (team_id) REFERENCES team(id)
);

CREATE INDEX team_kb_get_by_team_id ON team_knowledge_base (team_id);
CREATE INDEX team_kb_get_by_space_id ON team_knowledge_base (space_id);

/*
    Project wikis
    Link projects to their wiki spaces
*/
CREATE TABLE project_wiki
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    project_id      TEXT    NOT NULL UNIQUE,
    space_id        TEXT    NOT NULL,               -- Reference to Documents Extension document_space
    home_document_id TEXT,                          -- Wiki home page
    is_enabled      BOOLEAN NOT NULL DEFAULT 1,
    created         INTEGER NOT NULL,
    modified        INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL DEFAULT 0,

    FOREIGN KEY (project_id) REFERENCES project(id)
);

CREATE INDEX project_wiki_get_by_project_id ON project_wiki (project_id);
CREATE INDEX project_wiki_get_by_space_id ON project_wiki (space_id);

/*
    ========================================================================
    SEARCH ENHANCEMENT TABLES
    ========================================================================
*/

/*
    Cross-entity search index
    Unified search index across all entities and documents
    This enables powerful cross-system search capabilities
*/
CREATE TABLE cross_entity_search_index
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    entity_type     TEXT    NOT NULL,               -- "ticket", "document", "project", "comment", etc.
    entity_id       TEXT    NOT NULL,
    content_type    TEXT    NOT NULL,               -- "title", "description", "content", "comment"
    content_text    TEXT    NOT NULL,               -- Searchable text content
    content_hash    TEXT    NOT NULL,               -- SHA-256 hash for deduplication
    keywords        TEXT,                           -- Extracted keywords (space-separated)
    tags            TEXT,                           -- Associated tags (space-separated)
    metadata_json   TEXT,                           -- Additional metadata
    language        TEXT    NOT NULL DEFAULT 'en',  -- Content language
    last_indexed    INTEGER NOT NULL,               -- When this was last indexed
    created         INTEGER NOT NULL,

    UNIQUE(entity_type, entity_id, content_type)
);

CREATE INDEX cross_entity_search_get_by_entity_type ON cross_entity_search_index (entity_type);
CREATE INDEX cross_entity_search_get_by_entity_id ON cross_entity_search_index (entity_id);
CREATE INDEX cross_entity_search_get_by_content_hash ON cross_entity_search_index (content_hash);

-- Full-text search index (SQLite FTS5)
-- This will be created programmatically based on database type

/*
    ========================================================================
    TABLE ENHANCEMENTS - Additional Columns to Existing Tables
    ========================================================================
*/

/*
    ALTER TABLE statements for V5 enhancements:

    -- Add documentation flags to tickets
    ALTER TABLE ticket ADD COLUMN has_required_docs BOOLEAN DEFAULT 0;
    ALTER TABLE ticket ADD COLUMN docs_last_updated INTEGER;

    -- Add wiki flags to projects
    ALTER TABLE project ADD COLUMN wiki_enabled BOOLEAN DEFAULT 0;
    ALTER TABLE project ADD COLUMN wiki_space_id TEXT;

    -- Add KB flags to teams
    ALTER TABLE team ADD COLUMN kb_enabled BOOLEAN DEFAULT 0;
    ALTER TABLE team ADD COLUMN kb_space_id TEXT;

    -- Indexes for new columns
    CREATE INDEX tickets_get_by_has_required_docs ON ticket (has_required_docs);
    CREATE INDEX projects_get_by_wiki_enabled ON project (wiki_enabled);
    CREATE INDEX teams_get_by_kb_enabled ON team (kb_enabled);
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

    NEW IN VERSION 5:

    1. Document-Entity Integration:
       - Universal entity_document_mapping for linking any entity to documents
       - Project-specific document templates
       - Ticket documentation requirements by type/workflow

    2. Workflow Documentation:
       - Documentation requirements at workflow transitions
       - Automated document generation based on events
       - Validation rules for documents

    3. Knowledge Management:
       - Team knowledge bases linked to document spaces
       - Project wikis with home pages
       - Centralized documentation hubs

    4. Enhanced Search:
       - Cross-entity search index
       - Full-text search across all content
       - Keyword and tag extraction

    API CHANGES:

    New endpoints for V5:
    - POST /api/v5/entity/{type}/{id}/documents - Link document to entity
    - GET /api/v5/entity/{type}/{id}/documents - Get entity's documents
    - POST /api/v5/project/{id}/wiki/setup - Setup project wiki
    - POST /api/v5/team/{id}/kb/setup - Setup team knowledge base
    - GET /api/v5/search/cross-entity - Cross-entity search
    - POST /api/v5/document/auto-generate - Trigger automated doc generation

    Enhanced existing endpoints:
    - Ticket creation can now trigger automated documentation
    - Workflow transitions can require specific documents
    - Projects can enforce document templates

    INTEGRATION WITH DOCUMENTS EXTENSION V2:

    1. Core tables reference Documents Extension tables:
       - document (via document_id)
       - document_space (via space_id)
       - document_template (via template_id)

    2. Bi-directional linking:
       - Core → Documents: entity_document_mapping
       - Documents → Core: document_entity_link (from Docs V2)

    3. Unified search:
       - Cross-entity search includes both core and documents
       - Full-text search across all content types

    MIGRATION CONSIDERATIONS:

    1. Existing tickets/projects can be linked to documents
    2. Default spaces created for teams/projects on first use
    3. Search index populated asynchronously
    4. Automated documentation rules can be triggered retroactively

    BACKWARD COMPATIBILITY:

    1. All new tables are optional
    2. Core functionality works without Documents Extension
    3. Documents Extension works standalone
    4. Integration features only activate when both are present

    PERFORMANCE CONSIDERATIONS:

    1. Search index updated asynchronously
    2. Automated documentation triggered via background jobs
    3. Cross-entity queries optimized with proper indexes
    4. Document requirements checked only during transitions

    SECURITY CONSIDERATIONS:

    1. Entity-document links respect both entity and document permissions
    2. Wiki/KB spaces inherit team/project permissions
    3. Search results filtered by user permissions
    4. Automated documentation respects space permissions
*/

/*
    ========================================================================
    MIGRATION SCRIPT REFERENCE
    ========================================================================

    For migration from V4 to V5, see:
    Database/DDL/Migration.V4.5.sql

    The migration script will:
    1. Create all V5 tables
    2. Create indexes
    3. Optionally create default wikis/KBs for existing projects/teams
    4. Build initial search index
    5. Set up default document templates
*/

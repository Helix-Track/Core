-- ===================================================================
-- HelixTrack Core - Performance Optimization Indexes
-- ===================================================================
-- This script creates comprehensive indexes for extreme query performance
-- All indexes are designed for high-concurrency, high-throughput workloads
-- ===================================================================

-- ===================================================================
-- TICKET INDEXES - Most frequently queried table
-- ===================================================================

-- Primary key already indexed, but add covering indexes for common queries

-- Index for listing tickets by project (most common query)
CREATE INDEX IF NOT EXISTS idx_ticket_project_status_created
ON ticket(project_id, status_id, created DESC);

-- Index for listing tickets by assignee
CREATE INDEX IF NOT EXISTS idx_ticket_assignee_status
ON ticket(assignee_id, status_id, modified DESC);

-- Index for listing tickets by reporter
CREATE INDEX IF NOT EXISTS idx_ticket_reporter_created
ON ticket(reporter_id, created DESC);

-- Index for searching tickets by title (prefix search)
CREATE INDEX IF NOT EXISTS idx_ticket_title
ON ticket(title);

-- Index for filtering by type and status
CREATE INDEX IF NOT EXISTS idx_ticket_type_status
ON ticket(type_id, status_id, created DESC);

-- Index for filtering by priority (Phase 1)
CREATE INDEX IF NOT EXISTS idx_ticket_priority
ON ticket(priority_id, created DESC);

-- Index for deleted tickets (to exclude in queries)
CREATE INDEX IF NOT EXISTS idx_ticket_deleted
ON ticket(deleted, modified DESC);

-- Composite index for board/sprint queries
CREATE INDEX IF NOT EXISTS idx_ticket_board_sprint
ON ticket(board_id, cycle_id, status_id) WHERE deleted = 0;

-- Index for parent-child relationships
CREATE INDEX IF NOT EXISTS idx_ticket_parent
ON ticket(parent_id) WHERE parent_id IS NOT NULL;

-- ===================================================================
-- PROJECT INDEXES
-- ===================================================================

-- Index for listing projects by team
CREATE INDEX IF NOT EXISTS idx_project_team
ON project(team_id, deleted, modified DESC);

-- Index for project key lookup (unique, used for URL routing)
CREATE INDEX IF NOT EXISTS idx_project_key
ON project(key) WHERE deleted = 0;

-- Index for project name search
CREATE INDEX IF NOT EXISTS idx_project_title
ON project(title);

-- ===================================================================
-- WORKFLOW INDEXES
-- ===================================================================

-- Index for workflows by project
CREATE INDEX IF NOT EXISTS idx_workflow_project
ON workflow(project_id, deleted);

-- Index for workflow step by workflow
CREATE INDEX IF NOT EXISTS idx_workflow_step_workflow
ON workflow_step(workflow_id, step_order);

-- Index for workflow transitions
CREATE INDEX IF NOT EXISTS idx_workflow_transition_from
ON workflow_transition(from_status_id, workflow_id);

CREATE INDEX IF NOT EXISTS idx_workflow_transition_to
ON workflow_transition(to_status_id, workflow_id);

-- ===================================================================
-- STATUS INDEXES
-- ===================================================================

-- Index for status by project
CREATE INDEX IF NOT EXISTS idx_status_project
ON status(project_id, deleted);

-- Index for status category (for grouping)
CREATE INDEX IF NOT EXISTS idx_status_category
ON status(category, project_id);

-- ===================================================================
-- TEAM & ORGANIZATION INDEXES
-- ===================================================================

-- Index for teams by organization
CREATE INDEX IF NOT EXISTS idx_team_organization
ON team(organization_id, deleted);

-- Index for team members
CREATE INDEX IF NOT EXISTS idx_team_member_team
ON team_member_mapping(team_id, deleted);

CREATE INDEX IF NOT EXISTS idx_team_member_user
ON team_member_mapping(user_id, deleted);

-- ===================================================================
-- USER INDEXES
-- ===================================================================

-- Index for username lookup (authentication)
CREATE INDEX IF NOT EXISTS idx_user_username
ON "user"(username) WHERE deleted = 0;

-- Index for email lookup
CREATE INDEX IF NOT EXISTS idx_user_email
ON "user"(email) WHERE deleted = 0;

-- ===================================================================
-- COMPONENT INDEXES
-- ===================================================================

-- Index for components by project
CREATE INDEX IF NOT EXISTS idx_component_project
ON component(project_id, deleted);

-- Index for ticket-component mapping
CREATE INDEX IF NOT EXISTS idx_ticket_component_ticket
ON ticket_component_mapping(ticket_id);

CREATE INDEX IF NOT EXISTS idx_ticket_component_component
ON ticket_component_mapping(component_id);

-- ===================================================================
-- LABEL INDEXES
-- ===================================================================

-- Index for labels by project
CREATE INDEX IF NOT EXISTS idx_label_project
ON label(project_id, deleted);

-- Index for ticket-label mapping
CREATE INDEX IF NOT EXISTS idx_ticket_label_ticket
ON ticket_label_mapping(ticket_id);

CREATE INDEX IF NOT EXISTS idx_ticket_label_label
ON ticket_label_mapping(label_id);

-- ===================================================================
-- COMMENT INDEXES
-- ===================================================================

-- Index for comments by ticket (most common query)
CREATE INDEX IF NOT EXISTS idx_comment_ticket_created
ON comment(ticket_id, created DESC) WHERE deleted = 0;

-- Index for comments by author
CREATE INDEX IF NOT EXISTS idx_comment_author
ON comment(author_id, created DESC);

-- ===================================================================
-- ASSET (ATTACHMENT) INDEXES
-- ===================================================================

-- Index for assets by ticket
CREATE INDEX IF NOT EXISTS idx_asset_ticket
ON asset(ticket_id, created DESC) WHERE deleted = 0;

-- Index for assets by uploader
CREATE INDEX IF NOT EXISTS idx_asset_uploader
ON asset(uploaded_by, created DESC);

-- ===================================================================
-- BOARD INDEXES
-- ===================================================================

-- Index for boards by project
CREATE INDEX IF NOT EXISTS idx_board_project
ON board(project_id, deleted);

-- Index for board columns
CREATE INDEX IF NOT EXISTS idx_board_column_board
ON board_column(board_id, column_order);

-- ===================================================================
-- CYCLE (SPRINT) INDEXES
-- ===================================================================

-- Index for cycles by project
CREATE INDEX IF NOT EXISTS idx_cycle_project
ON cycle(project_id, deleted);

-- Index for active cycles
CREATE INDEX IF NOT EXISTS idx_cycle_active
ON cycle(project_id, start_date, end_date) WHERE deleted = 0;

-- ===================================================================
-- AUDIT LOG INDEXES
-- ===================================================================

-- Index for audit logs by entity
CREATE INDEX IF NOT EXISTS idx_audit_log_entity
ON audit_log(entity_type, entity_id, created DESC);

-- Index for audit logs by user
CREATE INDEX IF NOT EXISTS idx_audit_log_user
ON audit_log(user_id, created DESC);

-- Index for audit logs by timestamp (for cleanup)
CREATE INDEX IF NOT EXISTS idx_audit_log_created
ON audit_log(created DESC);

-- ===================================================================
-- REPOSITORY INDEXES
-- ===================================================================

-- Index for repositories by project
CREATE INDEX IF NOT EXISTS idx_repository_project
ON repository(project_id, deleted);

-- Index for commits by repository
CREATE INDEX IF NOT EXISTS idx_commit_repository
ON commit(repository_id, commit_date DESC);

-- Index for ticket-commit mapping
CREATE INDEX IF NOT EXISTS idx_ticket_commit_ticket
ON ticket_commit_mapping(ticket_id);

CREATE INDEX IF NOT EXISTS idx_ticket_commit_commit
ON ticket_commit_mapping(commit_id);

-- ===================================================================
-- PHASE 1 INDEXES (Priority, Resolution, Version, etc.)
-- ===================================================================

-- Priority indexes
CREATE INDEX IF NOT EXISTS idx_priority_level
ON priority(level, deleted);

-- Resolution indexes
CREATE INDEX IF NOT EXISTS idx_resolution_category
ON resolution(category, deleted);

-- Version indexes
CREATE INDEX IF NOT EXISTS idx_version_project
ON version(project_id, deleted, released);

CREATE INDEX IF NOT EXISTS idx_version_release_date
ON version(release_date DESC) WHERE released = 1;

-- Ticket affected version mapping
CREATE INDEX IF NOT EXISTS idx_ticket_affected_version_ticket
ON ticket_affected_version_mapping(ticket_id);

CREATE INDEX IF NOT EXISTS idx_ticket_affected_version_version
ON ticket_affected_version_mapping(version_id);

-- Ticket fix version mapping
CREATE INDEX IF NOT EXISTS idx_ticket_fix_version_ticket
ON ticket_fix_version_mapping(ticket_id);

CREATE INDEX IF NOT EXISTS idx_ticket_fix_version_version
ON ticket_fix_version_mapping(version_id);

-- Watcher indexes
CREATE INDEX IF NOT EXISTS idx_ticket_watcher_ticket
ON ticket_watcher_mapping(ticket_id);

CREATE INDEX IF NOT EXISTS idx_ticket_watcher_user
ON ticket_watcher_mapping(user_id);

-- Filter indexes
CREATE INDEX IF NOT EXISTS idx_filter_owner
ON filter(owner_id, deleted);

CREATE INDEX IF NOT EXISTS idx_filter_project
ON filter(project_id, deleted);

-- Filter share mapping
CREATE INDEX IF NOT EXISTS idx_filter_share_filter
ON filter_share_mapping(filter_id);

CREATE INDEX IF NOT EXISTS idx_filter_share_user
ON filter_share_mapping(user_id);

-- Custom field indexes
CREATE INDEX IF NOT EXISTS idx_custom_field_project
ON custom_field(project_id, deleted);

CREATE INDEX IF NOT EXISTS idx_custom_field_type
ON custom_field(field_type);

-- Custom field option indexes
CREATE INDEX IF NOT EXISTS idx_custom_field_option_field
ON custom_field_option(custom_field_id, option_order);

-- Ticket custom field value indexes
CREATE INDEX IF NOT EXISTS idx_ticket_custom_field_ticket
ON ticket_custom_field_value(ticket_id);

CREATE INDEX IF NOT EXISTS idx_ticket_custom_field_field
ON ticket_custom_field_value(custom_field_id);

-- ===================================================================
-- FULL-TEXT SEARCH INDEXES (SQLite FTS5)
-- ===================================================================
-- Note: These require FTS5 extension to be enabled

-- Full-text search for tickets
CREATE VIRTUAL TABLE IF NOT EXISTS ticket_fts USING fts5(
    id UNINDEXED,
    title,
    description,
    content='ticket',
    content_rowid='rowid'
);

-- Triggers to keep FTS index in sync
CREATE TRIGGER IF NOT EXISTS ticket_fts_insert AFTER INSERT ON ticket BEGIN
    INSERT INTO ticket_fts(rowid, id, title, description)
    VALUES (new.rowid, new.id, new.title, new.description);
END;

CREATE TRIGGER IF NOT EXISTS ticket_fts_delete AFTER DELETE ON ticket BEGIN
    DELETE FROM ticket_fts WHERE rowid = old.rowid;
END;

CREATE TRIGGER IF NOT EXISTS ticket_fts_update AFTER UPDATE ON ticket BEGIN
    DELETE FROM ticket_fts WHERE rowid = old.rowid;
    INSERT INTO ticket_fts(rowid, id, title, description)
    VALUES (new.rowid, new.id, new.title, new.description);
END;

-- Full-text search for comments
CREATE VIRTUAL TABLE IF NOT EXISTS comment_fts USING fts5(
    id UNINDEXED,
    content,
    content='comment',
    content_rowid='rowid'
);

CREATE TRIGGER IF NOT EXISTS comment_fts_insert AFTER INSERT ON comment BEGIN
    INSERT INTO comment_fts(rowid, id, content)
    VALUES (new.rowid, new.id, new.content);
END;

CREATE TRIGGER IF NOT EXISTS comment_fts_delete AFTER DELETE ON comment BEGIN
    DELETE FROM comment_fts WHERE rowid = old.rowid;
END;

CREATE TRIGGER IF NOT EXISTS comment_fts_update AFTER UPDATE ON comment BEGIN
    DELETE FROM comment_fts WHERE rowid = old.rowid;
    INSERT INTO comment_fts(rowid, id, content)
    VALUES (new.rowid, new.id, new.content);
END;

-- ===================================================================
-- POSTGRESQL-SPECIFIC INDEXES (if using PostgreSQL)
-- ===================================================================
-- These are ignored by SQLite

-- GIN index for full-text search (PostgreSQL)
-- CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_ticket_fulltext_gin
-- ON ticket USING GIN (to_tsvector('english', title || ' ' || description));

-- CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_comment_fulltext_gin
-- ON comment USING GIN (to_tsvector('english', content));

-- BRIN index for timestamp columns (PostgreSQL, efficient for large tables)
-- CREATE INDEX IF NOT EXISTS idx_ticket_created_brin
-- ON ticket USING BRIN (created);

-- CREATE INDEX IF NOT EXISTS idx_audit_log_created_brin
-- ON audit_log USING BRIN (created);

-- ===================================================================
-- ANALYZE TABLES (Update query planner statistics)
-- ===================================================================

-- SQLite
ANALYZE;

-- PostgreSQL (commented out, uncomment if using PostgreSQL)
-- ANALYZE VERBOSE;

-- ===================================================================
-- OPTIMIZATION NOTES
-- ===================================================================
--
-- 1. All indexes are designed to support common query patterns
-- 2. Covering indexes reduce the need for table lookups
-- 3. Partial indexes (WHERE clauses) reduce index size for filtered queries
-- 4. Composite indexes support multiple query patterns
-- 5. FTS indexes enable fast full-text search
-- 6. Regular ANALYZE keeps statistics up to date
--
-- Performance Tips:
-- - Use prepared statements (implemented in optimized_database.go)
-- - Enable query result caching (implemented in cache.go)
-- - Use connection pooling (implemented in optimized_database.go)
-- - Monitor slow queries and add indexes as needed
-- - Regularly run ANALYZE to update statistics
-- - Consider vacuuming (SQLite: PRAGMA auto_vacuum; PostgreSQL: autovacuum)
--
-- ===================================================================

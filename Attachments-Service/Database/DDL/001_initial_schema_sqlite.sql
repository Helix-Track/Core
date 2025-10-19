-- ============================================================
-- HelixTrack Attachments Service - SQLite Schema
-- Version: 1.0.0
-- Description: SQLite-compatible schema (triggers adapted)
-- ============================================================

-- ============================================================
-- 1. PHYSICAL FILES TABLE
-- ============================================================

CREATE TABLE IF NOT EXISTS attachment_file (
    hash                TEXT    PRIMARY KEY CHECK (length(hash) = 64),
    size_bytes          INTEGER NOT NULL CHECK (size_bytes >= 0),
    mime_type           TEXT    NOT NULL,
    extension           TEXT,
    ref_count           INTEGER NOT NULL DEFAULT 1 CHECK (ref_count >= 0),
    storage_primary     TEXT    NOT NULL,
    storage_backup      TEXT,
    storage_mirrors     TEXT,  -- JSON array
    virus_scan_status   TEXT    DEFAULT 'pending' CHECK (virus_scan_status IN ('pending', 'clean', 'infected', 'failed', 'skipped')),
    virus_scan_date     INTEGER,
    virus_scan_result   TEXT,
    created             INTEGER NOT NULL,
    last_accessed       INTEGER NOT NULL,
    deleted             INTEGER NOT NULL DEFAULT 0 CHECK (deleted IN (0, 1))
);

CREATE INDEX idx_attachment_file_ref_count ON attachment_file(ref_count) WHERE ref_count = 0;
CREATE INDEX idx_attachment_file_mime ON attachment_file(mime_type);
CREATE INDEX idx_attachment_file_created ON attachment_file(created DESC);
CREATE INDEX idx_attachment_file_deleted ON attachment_file(deleted) WHERE deleted = 0;

-- ============================================================
-- 2. LOGICAL REFERENCES TABLE
-- ============================================================

CREATE TABLE IF NOT EXISTS attachment_reference (
    id              TEXT    PRIMARY KEY,
    file_hash       TEXT    NOT NULL,
    entity_type     TEXT    NOT NULL CHECK (entity_type IN ('ticket', 'document', 'comment', 'project', 'team', 'user', 'epic', 'story', 'task')),
    entity_id       TEXT    NOT NULL,
    filename        TEXT    NOT NULL CHECK (length(filename) > 0 AND length(filename) <= 255),
    description     TEXT,
    uploader_id     TEXT    NOT NULL,
    version         INTEGER NOT NULL DEFAULT 1 CHECK (version >= 1),
    tags            TEXT,  -- JSON array
    created         INTEGER NOT NULL,
    modified        INTEGER NOT NULL,
    deleted         INTEGER NOT NULL DEFAULT 0 CHECK (deleted IN (0, 1)),

    FOREIGN KEY (file_hash) REFERENCES attachment_file(hash) ON DELETE CASCADE
);

CREATE INDEX idx_attachment_ref_entity ON attachment_reference(entity_type, entity_id, deleted) WHERE deleted = 0;
CREATE INDEX idx_attachment_ref_uploader ON attachment_reference(uploader_id, deleted) WHERE deleted = 0;
CREATE INDEX idx_attachment_ref_hash ON attachment_reference(file_hash);
CREATE INDEX idx_attachment_ref_created ON attachment_reference(created DESC);

-- ============================================================
-- 3. STORAGE ENDPOINTS
-- ============================================================

CREATE TABLE IF NOT EXISTS storage_endpoint (
    id              TEXT    PRIMARY KEY,
    name            TEXT    NOT NULL,
    type            TEXT    NOT NULL CHECK (type IN ('local', 's3', 'minio', 'azure', 'gcs', 'custom')),
    role            TEXT    NOT NULL CHECK (role IN ('primary', 'backup', 'mirror')),
    adapter_config  TEXT    NOT NULL,  -- JSON
    priority        INTEGER NOT NULL DEFAULT 1,
    enabled         INTEGER NOT NULL DEFAULT 1 CHECK (enabled IN (0, 1)),
    max_size_bytes  INTEGER,
    current_size    INTEGER NOT NULL DEFAULT 0 CHECK (current_size >= 0),
    created         INTEGER NOT NULL,
    modified        INTEGER NOT NULL
);

CREATE INDEX idx_storage_endpoint_role ON storage_endpoint(role, enabled);

-- ============================================================
-- 4. STORAGE HEALTH
-- ============================================================

CREATE TABLE IF NOT EXISTS storage_health (
    endpoint_id     TEXT    NOT NULL,
    check_time      INTEGER NOT NULL,
    status          TEXT    NOT NULL CHECK (status IN ('healthy', 'degraded', 'unhealthy')),
    latency_ms      INTEGER CHECK (latency_ms >= 0),
    error_message   TEXT,
    available_bytes INTEGER CHECK (available_bytes >= 0),

    PRIMARY KEY (endpoint_id, check_time),
    FOREIGN KEY (endpoint_id) REFERENCES storage_endpoint(id) ON DELETE CASCADE
);

CREATE INDEX idx_storage_health_time ON storage_health(check_time DESC);

-- ============================================================
-- 5. UPLOAD QUOTAS
-- ============================================================

CREATE TABLE IF NOT EXISTS upload_quota (
    user_id         TEXT    PRIMARY KEY,
    max_bytes       INTEGER NOT NULL DEFAULT 10737418240 CHECK (max_bytes > 0),
    used_bytes      INTEGER NOT NULL DEFAULT 0 CHECK (used_bytes >= 0),
    max_files       INTEGER NOT NULL DEFAULT 10000 CHECK (max_files > 0),
    used_files      INTEGER NOT NULL DEFAULT 0 CHECK (used_files >= 0),
    created         INTEGER NOT NULL,
    modified        INTEGER NOT NULL,

    CHECK (used_bytes <= max_bytes AND used_files <= max_files)
);

-- ============================================================
-- 6. ACCESS LOGS
-- ============================================================

CREATE TABLE IF NOT EXISTS access_log (
    id              TEXT    PRIMARY KEY,
    reference_id    TEXT,
    file_hash       TEXT,
    user_id         TEXT,
    ip_address      TEXT,
    action          TEXT    NOT NULL CHECK (action IN ('upload', 'download', 'delete', 'metadata_read', 'metadata_update')),
    status_code     INTEGER,
    error_message   TEXT,
    user_agent      TEXT,
    timestamp       INTEGER NOT NULL
);

CREATE INDEX idx_access_log_timestamp ON access_log(timestamp DESC);
CREATE INDEX idx_access_log_user ON access_log(user_id, timestamp DESC);

-- ============================================================
-- 7. PRESIGNED URLS
-- ============================================================

CREATE TABLE IF NOT EXISTS presigned_url (
    token           TEXT    PRIMARY KEY,
    reference_id    TEXT    NOT NULL,
    user_id         TEXT,
    ip_address      TEXT,
    expires_at      INTEGER NOT NULL,
    max_downloads   INTEGER DEFAULT 1 CHECK (max_downloads > 0),
    download_count  INTEGER NOT NULL DEFAULT 0 CHECK (download_count >= 0),
    created         INTEGER NOT NULL,

    FOREIGN KEY (reference_id) REFERENCES attachment_reference(id) ON DELETE CASCADE,
    CHECK (download_count <= max_downloads)
);

CREATE INDEX idx_presigned_expires ON presigned_url(expires_at);

-- ============================================================
-- 8. CLEANUP JOBS
-- ============================================================

CREATE TABLE IF NOT EXISTS cleanup_job (
    id              TEXT    PRIMARY KEY,
    job_type        TEXT    NOT NULL CHECK (job_type IN ('orphan_files', 'dangling_refs', 'expired_presigned', 'old_health_data', 'old_access_logs')),
    started         INTEGER NOT NULL,
    completed       INTEGER,
    status          TEXT    NOT NULL CHECK (status IN ('running', 'completed', 'failed')),
    items_processed INTEGER NOT NULL DEFAULT 0,
    items_deleted   INTEGER NOT NULL DEFAULT 0,
    error_message   TEXT
);

CREATE INDEX idx_cleanup_job_started ON cleanup_job(started DESC);

-- ============================================================
-- 9. TRIGGERS (SQLite compatible)
-- ============================================================

-- Increment ref_count on insert
CREATE TRIGGER trigger_increment_ref_count
AFTER INSERT ON attachment_reference
FOR EACH ROW
BEGIN
    UPDATE attachment_file
    SET ref_count = ref_count + 1,
        last_accessed = NEW.created
    WHERE hash = NEW.file_hash;
END;

-- Decrement ref_count on soft delete
CREATE TRIGGER trigger_decrement_ref_count_update
AFTER UPDATE OF deleted ON attachment_reference
FOR EACH ROW
WHEN NEW.deleted = 1 AND OLD.deleted = 0
BEGIN
    UPDATE attachment_file
    SET ref_count = ref_count - 1
    WHERE hash = OLD.file_hash;
END;

-- Decrement ref_count on hard delete
CREATE TRIGGER trigger_decrement_ref_count_delete
AFTER DELETE ON attachment_reference
FOR EACH ROW
BEGIN
    UPDATE attachment_file
    SET ref_count = ref_count - 1
    WHERE hash = OLD.file_hash;
END;

-- ============================================================
-- 10. SCHEMA VERSION
-- ============================================================

CREATE TABLE IF NOT EXISTS schema_version (
    version         INTEGER PRIMARY KEY,
    description     TEXT NOT NULL,
    applied         INTEGER NOT NULL
);

INSERT OR IGNORE INTO schema_version (version, description, applied)
VALUES (1, 'Initial schema - SQLite version', strftime('%s', 'now'));

-- Insert default storage endpoint
INSERT OR IGNORE INTO storage_endpoint (
    id, name, type, role, adapter_config, priority, enabled,
    max_size_bytes, current_size, created, modified
) VALUES (
    'default-local-primary',
    'Default Local Storage',
    'local',
    'primary',
    '{"path": "/var/helixtrack/attachments"}',
    1,
    1,
    1073741824000,
    0,
    strftime('%s', 'now'),
    strftime('%s', 'now')
);

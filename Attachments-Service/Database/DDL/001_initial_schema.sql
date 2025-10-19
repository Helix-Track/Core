-- ============================================================
-- HelixTrack Attachments Service - Initial Database Schema
-- Version: 1.0.0
-- Description: Complete schema for hash-based attachment storage
--              with deduplication, reference counting, and multi-endpoint
--              storage support.
-- ============================================================

-- ============================================================
-- 1. PHYSICAL FILES TABLE (Deduplicated Storage)
-- ============================================================

CREATE TABLE IF NOT EXISTS attachment_file (
    hash                TEXT    PRIMARY KEY,
    size_bytes          BIGINT  NOT NULL CHECK (size_bytes >= 0),
    mime_type           TEXT    NOT NULL,
    extension           TEXT,
    ref_count           INTEGER NOT NULL DEFAULT 1 CHECK (ref_count >= 0),
    storage_primary     TEXT    NOT NULL,
    storage_backup      TEXT,
    storage_mirrors     TEXT[],
    virus_scan_status   TEXT    DEFAULT 'pending' CHECK (virus_scan_status IN ('pending', 'clean', 'infected', 'failed', 'skipped')),
    virus_scan_date     BIGINT,
    virus_scan_result   TEXT,
    created             BIGINT  NOT NULL,
    last_accessed       BIGINT  NOT NULL,
    deleted             BOOLEAN NOT NULL DEFAULT false,

    CONSTRAINT attachment_file_hash_length CHECK (length(hash) = 64)  -- SHA-256 is 64 hex chars
);

COMMENT ON TABLE attachment_file IS 'Physical files stored once per unique hash (SHA-256)';
COMMENT ON COLUMN attachment_file.hash IS 'SHA-256 hash of file content (64 hex characters)';
COMMENT ON COLUMN attachment_file.ref_count IS 'Number of references to this file (atomic counter)';
COMMENT ON COLUMN attachment_file.storage_primary IS 'Path to file on primary storage endpoint';
COMMENT ON COLUMN attachment_file.storage_backup IS 'Path to file on backup storage endpoint';
COMMENT ON COLUMN attachment_file.storage_mirrors IS 'Array of paths on mirror storage endpoints';
COMMENT ON COLUMN attachment_file.virus_scan_status IS 'ClamAV scan status';

CREATE INDEX idx_attachment_file_ref_count ON attachment_file(ref_count) WHERE ref_count = 0;
CREATE INDEX idx_attachment_file_mime ON attachment_file(mime_type);
CREATE INDEX idx_attachment_file_created ON attachment_file(created DESC);
CREATE INDEX idx_attachment_file_deleted ON attachment_file(deleted) WHERE deleted = false;
CREATE INDEX idx_attachment_file_virus_status ON attachment_file(virus_scan_status) WHERE virus_scan_status IN ('pending', 'infected');

-- ============================================================
-- 2. LOGICAL REFERENCES TABLE (Entity-to-File Mapping)
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
    tags            TEXT[],
    created         BIGINT  NOT NULL,
    modified        BIGINT  NOT NULL,
    deleted         BOOLEAN NOT NULL DEFAULT false,

    FOREIGN KEY (file_hash) REFERENCES attachment_file(hash) ON DELETE CASCADE
);

COMMENT ON TABLE attachment_reference IS 'Logical references linking entities to physical files';
COMMENT ON COLUMN attachment_reference.entity_type IS 'Type of entity this file is attached to';
COMMENT ON COLUMN attachment_reference.entity_id IS 'ID of the entity this file is attached to';
COMMENT ON COLUMN attachment_reference.filename IS 'User-provided filename (may differ from physical filename)';

CREATE INDEX idx_attachment_ref_entity ON attachment_reference(entity_type, entity_id, deleted) WHERE deleted = false;
CREATE INDEX idx_attachment_ref_uploader ON attachment_reference(uploader_id, deleted) WHERE deleted = false;
CREATE INDEX idx_attachment_ref_hash ON attachment_reference(file_hash);
CREATE INDEX idx_attachment_ref_created ON attachment_reference(created DESC);
CREATE INDEX idx_attachment_ref_tags ON attachment_reference USING GIN(tags);

-- ============================================================
-- 3. STORAGE ENDPOINTS CONFIGURATION
-- ============================================================

CREATE TABLE IF NOT EXISTS storage_endpoint (
    id              TEXT    PRIMARY KEY,
    name            TEXT    NOT NULL,
    type            TEXT    NOT NULL CHECK (type IN ('local', 's3', 'minio', 'azure', 'gcs', 'custom')),
    role            TEXT    NOT NULL CHECK (role IN ('primary', 'backup', 'mirror')),
    adapter_config  JSONB   NOT NULL,
    priority        INTEGER NOT NULL DEFAULT 1,
    enabled         BOOLEAN NOT NULL DEFAULT true,
    max_size_bytes  BIGINT,
    current_size    BIGINT  NOT NULL DEFAULT 0 CHECK (current_size >= 0),
    created         BIGINT  NOT NULL,
    modified        BIGINT  NOT NULL,

    CONSTRAINT unique_primary_endpoint UNIQUE (role) WHERE role = 'primary' AND enabled = true
);

COMMENT ON TABLE storage_endpoint IS 'Configuration for storage endpoints (local, S3, MinIO, etc.)';
COMMENT ON COLUMN storage_endpoint.role IS 'Primary (main), backup (failover), or mirror (replication)';
COMMENT ON COLUMN storage_endpoint.adapter_config IS 'JSON configuration for storage adapter';

CREATE INDEX idx_storage_endpoint_role ON storage_endpoint(role, enabled);
CREATE INDEX idx_storage_endpoint_priority ON storage_endpoint(priority);

-- ============================================================
-- 4. STORAGE HEALTH MONITORING
-- ============================================================

CREATE TABLE IF NOT EXISTS storage_health (
    endpoint_id     TEXT    NOT NULL,
    check_time      BIGINT  NOT NULL,
    status          TEXT    NOT NULL CHECK (status IN ('healthy', 'degraded', 'unhealthy')),
    latency_ms      INTEGER CHECK (latency_ms >= 0),
    error_message   TEXT,
    available_bytes BIGINT CHECK (available_bytes >= 0),

    PRIMARY KEY (endpoint_id, check_time),
    FOREIGN KEY (endpoint_id) REFERENCES storage_endpoint(id) ON DELETE CASCADE
);

COMMENT ON TABLE storage_health IS 'Historical health check data for storage endpoints';

CREATE INDEX idx_storage_health_time ON storage_health(check_time DESC);
CREATE INDEX idx_storage_health_status ON storage_health(endpoint_id, status);

-- ============================================================
-- 5. UPLOAD QUOTAS (Per-User Limits)
-- ============================================================

CREATE TABLE IF NOT EXISTS upload_quota (
    user_id         TEXT    PRIMARY KEY,
    max_bytes       BIGINT  NOT NULL DEFAULT 10737418240 CHECK (max_bytes > 0),  -- 10 GB default
    used_bytes      BIGINT  NOT NULL DEFAULT 0 CHECK (used_bytes >= 0),
    max_files       INTEGER NOT NULL DEFAULT 10000 CHECK (max_files > 0),
    used_files      INTEGER NOT NULL DEFAULT 0 CHECK (used_files >= 0),
    created         BIGINT  NOT NULL,
    modified        BIGINT  NOT NULL,

    CONSTRAINT quota_not_exceeded CHECK (used_bytes <= max_bytes AND used_files <= max_files)
);

COMMENT ON TABLE upload_quota IS 'Per-user upload quotas and usage tracking';

CREATE INDEX idx_upload_quota_usage ON upload_quota(used_bytes DESC);

-- ============================================================
-- 6. ACCESS LOGS (Audit Trail)
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
    timestamp       BIGINT  NOT NULL
);

COMMENT ON TABLE access_log IS 'Audit log for all file operations';

CREATE INDEX idx_access_log_timestamp ON access_log(timestamp DESC);
CREATE INDEX idx_access_log_user ON access_log(user_id, timestamp DESC);
CREATE INDEX idx_access_log_action ON access_log(action, timestamp DESC);
CREATE INDEX idx_access_log_reference ON access_log(reference_id);

-- ============================================================
-- 7. PRESIGNED URLS (Temporary Access Tokens)
-- ============================================================

CREATE TABLE IF NOT EXISTS presigned_url (
    token           TEXT    PRIMARY KEY,
    reference_id    TEXT    NOT NULL,
    user_id         TEXT,
    ip_address      TEXT,
    expires_at      BIGINT  NOT NULL,
    max_downloads   INTEGER DEFAULT 1 CHECK (max_downloads > 0),
    download_count  INTEGER NOT NULL DEFAULT 0 CHECK (download_count >= 0),
    created         BIGINT  NOT NULL,

    FOREIGN KEY (reference_id) REFERENCES attachment_reference(id) ON DELETE CASCADE,
    CONSTRAINT download_limit_not_exceeded CHECK (download_count <= max_downloads)
);

COMMENT ON TABLE presigned_url IS 'Temporary URLs for time-limited or count-limited file access';

CREATE INDEX idx_presigned_expires ON presigned_url(expires_at);
CREATE INDEX idx_presigned_ref ON presigned_url(reference_id);

-- ============================================================
-- 8. CLEANUP JOBS TRACKING
-- ============================================================

CREATE TABLE IF NOT EXISTS cleanup_job (
    id              TEXT    PRIMARY KEY,
    job_type        TEXT    NOT NULL CHECK (job_type IN ('orphan_files', 'dangling_refs', 'expired_presigned', 'old_health_data', 'old_access_logs')),
    started         BIGINT  NOT NULL,
    completed       BIGINT,
    status          TEXT    NOT NULL CHECK (status IN ('running', 'completed', 'failed')),
    items_processed INTEGER NOT NULL DEFAULT 0 CHECK (items_processed >= 0),
    items_deleted   INTEGER NOT NULL DEFAULT 0 CHECK (items_deleted >= 0),
    error_message   TEXT
);

COMMENT ON TABLE cleanup_job IS 'Tracking for periodic cleanup jobs';

CREATE INDEX idx_cleanup_job_started ON cleanup_job(started DESC);
CREATE INDEX idx_cleanup_job_type ON cleanup_job(job_type, started DESC);

-- ============================================================
-- 9. TRIGGERS FOR AUTOMATIC REFERENCE COUNTING
-- ============================================================

-- Increment ref_count when a new reference is created
CREATE OR REPLACE FUNCTION increment_ref_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE attachment_file
    SET ref_count = ref_count + 1,
        last_accessed = NEW.created
    WHERE hash = NEW.file_hash;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_increment_ref_count
AFTER INSERT ON attachment_reference
FOR EACH ROW
EXECUTE FUNCTION increment_ref_count();

-- Decrement ref_count when a reference is deleted (soft or hard delete)
CREATE OR REPLACE FUNCTION decrement_ref_count()
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'DELETE') OR (NEW.deleted = true AND OLD.deleted = false) THEN
        UPDATE attachment_file
        SET ref_count = ref_count - 1
        WHERE hash = OLD.file_hash;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_decrement_ref_count
AFTER UPDATE OF deleted ON attachment_reference
FOR EACH ROW
EXECUTE FUNCTION decrement_ref_count();

CREATE TRIGGER trigger_decrement_ref_count_delete
AFTER DELETE ON attachment_reference
FOR EACH ROW
EXECUTE FUNCTION decrement_ref_count();

-- Update quota when file is uploaded
CREATE OR REPLACE FUNCTION update_quota_on_upload()
RETURNS TRIGGER AS $$
BEGIN
    -- Only increment if this is a new file (ref_count = 1)
    IF NEW.ref_count = 1 THEN
        INSERT INTO upload_quota (user_id, used_bytes, used_files, created, modified)
        SELECT
            r.uploader_id,
            NEW.size_bytes,
            1,
            NEW.created,
            NEW.created
        FROM attachment_reference r
        WHERE r.file_hash = NEW.hash
        LIMIT 1
        ON CONFLICT (user_id) DO UPDATE
        SET used_bytes = upload_quota.used_bytes + NEW.size_bytes,
            used_files = upload_quota.used_files + 1,
            modified = NEW.created;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_quota_upload
AFTER INSERT ON attachment_file
FOR EACH ROW
EXECUTE FUNCTION update_quota_on_upload();

-- Update quota when file is deleted (when ref_count reaches 0)
CREATE OR REPLACE FUNCTION update_quota_on_delete()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.ref_count = 0 AND OLD.ref_count > 0 THEN
        UPDATE upload_quota
        SET used_bytes = GREATEST(0, used_bytes - NEW.size_bytes),
            used_files = GREATEST(0, used_files - 1),
            modified = EXTRACT(EPOCH FROM NOW())::BIGINT
        WHERE user_id IN (
            SELECT uploader_id FROM attachment_reference WHERE file_hash = NEW.hash LIMIT 1
        );
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_quota_delete
AFTER UPDATE OF ref_count ON attachment_file
FOR EACH ROW
EXECUTE FUNCTION update_quota_on_delete();

-- ============================================================
-- 10. FUNCTIONS FOR COMMON OPERATIONS
-- ============================================================

-- Get total storage usage across all files
CREATE OR REPLACE FUNCTION get_total_storage_usage()
RETURNS BIGINT AS $$
    SELECT COALESCE(SUM(size_bytes), 0)
    FROM attachment_file
    WHERE deleted = false;
$$ LANGUAGE SQL;

-- Get storage usage for a specific user
CREATE OR REPLACE FUNCTION get_user_storage_usage(p_user_id TEXT)
RETURNS TABLE (
    total_bytes BIGINT,
    total_files BIGINT,
    quota_bytes BIGINT,
    quota_files INTEGER,
    usage_percent NUMERIC
) AS $$
    SELECT
        COALESCE(q.used_bytes, 0) AS total_bytes,
        COALESCE(q.used_files, 0)::BIGINT AS total_files,
        COALESCE(q.max_bytes, 0) AS quota_bytes,
        COALESCE(q.max_files, 0) AS quota_files,
        CASE
            WHEN q.max_bytes > 0 THEN ROUND((q.used_bytes::NUMERIC / q.max_bytes::NUMERIC) * 100, 2)
            ELSE 0
        END AS usage_percent
    FROM upload_quota q
    WHERE q.user_id = p_user_id;
$$ LANGUAGE SQL;

-- Get orphaned files (ref_count = 0 and older than retention period)
CREATE OR REPLACE FUNCTION get_orphaned_files(retention_days INTEGER DEFAULT 30)
RETURNS TABLE (
    hash TEXT,
    size_bytes BIGINT,
    age_days INTEGER
) AS $$
    SELECT
        hash,
        size_bytes,
        (EXTRACT(EPOCH FROM NOW())::BIGINT - created) / 86400 AS age_days
    FROM attachment_file
    WHERE ref_count = 0
      AND deleted = false
      AND (EXTRACT(EPOCH FROM NOW())::BIGINT - created) > (retention_days * 86400);
$$ LANGUAGE SQL;

-- ============================================================
-- 11. INDEXES FOR PERFORMANCE
-- ============================================================

-- Additional composite indexes for common queries
CREATE INDEX idx_attachment_ref_entity_filename ON attachment_reference(entity_type, entity_id, filename) WHERE deleted = false;
CREATE INDEX idx_attachment_file_size ON attachment_file(size_bytes DESC);
CREATE INDEX idx_access_log_file_hash ON access_log(file_hash, timestamp DESC);

-- ============================================================
-- 12. INITIAL DATA (Default Storage Endpoint)
-- ============================================================

-- Insert default local storage endpoint
INSERT INTO storage_endpoint (
    id,
    name,
    type,
    role,
    adapter_config,
    priority,
    enabled,
    max_size_bytes,
    current_size,
    created,
    modified
) VALUES (
    'default-local-primary',
    'Default Local Storage',
    'local',
    'primary',
    '{"path": "/var/helixtrack/attachments"}'::JSONB,
    1,
    true,
    1073741824000,  -- 1 TB
    0,
    EXTRACT(EPOCH FROM NOW())::BIGINT,
    EXTRACT(EPOCH FROM NOW())::BIGINT
) ON CONFLICT (id) DO NOTHING;

-- ============================================================
-- SCHEMA VERSION TRACKING
-- ============================================================

CREATE TABLE IF NOT EXISTS schema_version (
    version         INTEGER PRIMARY KEY,
    description     TEXT NOT NULL,
    applied         BIGINT NOT NULL
);

INSERT INTO schema_version (version, description, applied)
VALUES (1, 'Initial schema - hash-based deduplication with multi-endpoint storage', EXTRACT(EPOCH FROM NOW())::BIGINT)
ON CONFLICT (version) DO NOTHING;

-- ============================================================
-- END OF INITIAL SCHEMA
-- ============================================================

/*
    HelixTrack Localization Service - Database Schema
    Version: 1.0
    Database: PostgreSQL 12+
    Encryption: SQL Cipher (AES-256)
*/

/*
    Notes:
    - All IDs are UUID v4
    - Timestamps are Unix epoch (BIGINT)
    - Soft delete pattern used throughout
    - Encryption enabled on sensitive columns
    - Comprehensive indexing for performance
*/

-- Drop existing tables (reverse order of dependencies)
DROP TABLE IF EXISTS localization_audit_log CASCADE;
DROP TABLE IF EXISTS localization_cache_keys CASCADE;
DROP TABLE IF EXISTS localization_catalogs CASCADE;
DROP TABLE IF EXISTS localizations CASCADE;
DROP TABLE IF EXISTS localization_keys CASCADE;
DROP TABLE IF EXISTS languages CASCADE;

-- Drop existing indexes
DROP INDEX IF EXISTS idx_languages_code;
DROP INDEX IF EXISTS idx_languages_is_default;
DROP INDEX IF EXISTS idx_languages_is_active;
DROP INDEX IF EXISTS idx_languages_deleted;
DROP INDEX IF EXISTS idx_localization_keys_key;
DROP INDEX IF EXISTS idx_localization_keys_category;
DROP INDEX IF EXISTS idx_localization_keys_deleted;
DROP INDEX IF EXISTS idx_localizations_key_id;
DROP INDEX IF EXISTS idx_localizations_language_id;
DROP INDEX IF EXISTS idx_localizations_approved;
DROP INDEX IF EXISTS idx_localizations_version;
DROP INDEX IF EXISTS idx_localizations_deleted;
DROP INDEX IF EXISTS idx_localization_catalogs_language_id;
DROP INDEX IF EXISTS idx_localization_catalogs_version;
DROP INDEX IF EXISTS idx_localization_catalogs_checksum;
DROP INDEX IF EXISTS idx_localization_catalogs_category;
DROP INDEX IF EXISTS idx_localization_cache_keys_cache_key;
DROP INDEX IF EXISTS idx_localization_cache_keys_expires_at;
DROP INDEX IF EXISTS idx_localization_audit_log_entity_id;
DROP INDEX IF EXISTS idx_localization_audit_log_username;
DROP INDEX IF EXISTS idx_localization_audit_log_created_at;

--
-- Table: languages
-- Description: Supported languages in the system
--
CREATE TABLE languages (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code            VARCHAR(10) NOT NULL UNIQUE,   -- ISO 639-1 (e.g., 'en', 'de', 'fr')
    name            VARCHAR(100) NOT NULL,          -- English name (e.g., 'English')
    native_name     VARCHAR(100),                   -- Native name (e.g., 'Deutsch')
    is_rtl          BOOLEAN DEFAULT FALSE,          -- Right-to-left language
    is_active       BOOLEAN DEFAULT TRUE,           -- Language enabled
    is_default      BOOLEAN DEFAULT FALSE,          -- Default fallback language
    created_at      BIGINT NOT NULL,                -- Unix timestamp
    modified_at     BIGINT NOT NULL,                -- Unix timestamp
    deleted         BOOLEAN DEFAULT FALSE           -- Soft delete
);

-- Indexes for languages
CREATE INDEX idx_languages_code ON languages(code) WHERE deleted = FALSE;
CREATE INDEX idx_languages_is_default ON languages(is_default) WHERE deleted = FALSE;
CREATE INDEX idx_languages_is_active ON languages(is_active) WHERE deleted = FALSE;
CREATE INDEX idx_languages_deleted ON languages(deleted);

-- Only one default language allowed
CREATE UNIQUE INDEX idx_languages_single_default ON languages(is_default) WHERE is_default = TRUE AND deleted = FALSE;

--
-- Table: localization_keys
-- Description: Master list of all localization keys
--
CREATE TABLE localization_keys (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key             VARCHAR(255) NOT NULL UNIQUE,   -- e.g., 'error.auth.invalid_token'
    category        VARCHAR(100),                   -- e.g., 'error', 'ui', 'message'
    description     TEXT,                           -- Developer notes
    context         VARCHAR(255),                   -- Usage context
    created_at      BIGINT NOT NULL,
    modified_at     BIGINT NOT NULL,
    deleted         BOOLEAN DEFAULT FALSE
);

-- Indexes for localization_keys
CREATE INDEX idx_localization_keys_key ON localization_keys(key) WHERE deleted = FALSE;
CREATE INDEX idx_localization_keys_category ON localization_keys(category) WHERE deleted = FALSE;
CREATE INDEX idx_localization_keys_deleted ON localization_keys(deleted);

--
-- Table: localizations
-- Description: Actual localized strings
--
CREATE TABLE localizations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key_id          UUID NOT NULL REFERENCES localization_keys(id) ON DELETE CASCADE,
    language_id     UUID NOT NULL REFERENCES languages(id) ON DELETE CASCADE,
    value           TEXT NOT NULL,                  -- Localized string (encrypted)
    plural_forms    JSONB,                          -- Plural forms (optional)
    variables       JSONB,                          -- Variable placeholders
    version         INTEGER DEFAULT 1,              -- Version number
    approved        BOOLEAN DEFAULT FALSE,          -- Reviewed and approved
    approved_by     VARCHAR(255),                   -- Username of approver
    approved_at     BIGINT,                         -- Approval timestamp
    created_at      BIGINT NOT NULL,
    modified_at     BIGINT NOT NULL,
    deleted         BOOLEAN DEFAULT FALSE,

    CONSTRAINT unique_key_language UNIQUE(key_id, language_id)
);

-- Indexes for localizations
CREATE INDEX idx_localizations_key_id ON localizations(key_id) WHERE deleted = FALSE;
CREATE INDEX idx_localizations_language_id ON localizations(language_id) WHERE deleted = FALSE;
CREATE INDEX idx_localizations_approved ON localizations(approved) WHERE deleted = FALSE;
CREATE INDEX idx_localizations_version ON localizations(version) WHERE deleted = FALSE;
CREATE INDEX idx_localizations_deleted ON localizations(deleted);

--
-- Table: localization_catalogs
-- Description: Pre-built catalogs for fast retrieval
--
CREATE TABLE localization_catalogs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    language_id     UUID NOT NULL REFERENCES languages(id) ON DELETE CASCADE,
    category        VARCHAR(100),                   -- Optional category filter
    catalog_data    JSONB NOT NULL,                 -- Complete catalog as JSON
    version         INTEGER NOT NULL,               -- Catalog version
    checksum        VARCHAR(64) NOT NULL,           -- SHA-256 checksum
    created_at      BIGINT NOT NULL,
    modified_at     BIGINT NOT NULL,

    CONSTRAINT unique_language_category_version UNIQUE(language_id, category, version)
);

-- Indexes for localization_catalogs
CREATE INDEX idx_localization_catalogs_language_id ON localization_catalogs(language_id);
CREATE INDEX idx_localization_catalogs_version ON localization_catalogs(version);
CREATE INDEX idx_localization_catalogs_checksum ON localization_catalogs(checksum);
CREATE INDEX idx_localization_catalogs_category ON localization_catalogs(category);

--
-- Table: localization_cache_keys
-- Description: Cache key tracking for invalidation
--
CREATE TABLE localization_cache_keys (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cache_key       VARCHAR(255) NOT NULL UNIQUE,
    language_code   VARCHAR(10),
    category        VARCHAR(100),
    ttl             INTEGER DEFAULT 3600,           -- TTL in seconds
    expires_at      BIGINT NOT NULL,
    created_at      BIGINT NOT NULL
);

-- Indexes for localization_cache_keys
CREATE INDEX idx_localization_cache_keys_cache_key ON localization_cache_keys(cache_key);
CREATE INDEX idx_localization_cache_keys_expires_at ON localization_cache_keys(expires_at);

--
-- Table: localization_audit_log
-- Description: Audit trail for all changes
--
CREATE TABLE localization_audit_log (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    action          VARCHAR(50) NOT NULL,           -- CREATE, UPDATE, DELETE, APPROVE
    entity_type     VARCHAR(50) NOT NULL,           -- LANGUAGE, KEY, LOCALIZATION
    entity_id       UUID NOT NULL,
    username        VARCHAR(255) NOT NULL,
    changes         JSONB,                          -- Before/after values
    ip_address      VARCHAR(45),                    -- IPv4/IPv6
    user_agent      TEXT,
    created_at      BIGINT NOT NULL
);

-- Indexes for localization_audit_log
CREATE INDEX idx_localization_audit_log_entity_id ON localization_audit_log(entity_id);
CREATE INDEX idx_localization_audit_log_username ON localization_audit_log(username);
CREATE INDEX idx_localization_audit_log_created_at ON localization_audit_log(created_at);
CREATE INDEX idx_localization_audit_log_action ON localization_audit_log(action);
CREATE INDEX idx_localization_audit_log_entity_type ON localization_audit_log(entity_type);

--
-- Triggers
--

-- Trigger: Update catalog version on localization change
CREATE OR REPLACE FUNCTION update_catalog_version()
RETURNS TRIGGER AS $$
BEGIN
    -- Increment version for affected catalogs
    UPDATE localization_catalogs
    SET version = version + 1,
        modified_at = EXTRACT(EPOCH FROM NOW())::BIGINT
    WHERE language_id = NEW.language_id;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_catalog_version
AFTER INSERT OR UPDATE ON localizations
FOR EACH ROW
EXECUTE FUNCTION update_catalog_version();

-- Trigger: Auto-update modified_at timestamp
CREATE OR REPLACE FUNCTION update_modified_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.modified_at = EXTRACT(EPOCH FROM NOW())::BIGINT;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_languages_modified_at
BEFORE UPDATE ON languages
FOR EACH ROW
EXECUTE FUNCTION update_modified_at();

CREATE TRIGGER trigger_localization_keys_modified_at
BEFORE UPDATE ON localization_keys
FOR EACH ROW
EXECUTE FUNCTION update_modified_at();

CREATE TRIGGER trigger_localizations_modified_at
BEFORE UPDATE ON localizations
FOR EACH ROW
EXECUTE FUNCTION update_modified_at();

CREATE TRIGGER trigger_localization_catalogs_modified_at
BEFORE UPDATE ON localization_catalogs
FOR EACH ROW
EXECUTE FUNCTION update_modified_at();

--
-- Initial Seed Data
--

-- Default language: English
INSERT INTO languages (id, code, name, native_name, is_rtl, is_active, is_default, created_at, modified_at, deleted)
VALUES (
    gen_random_uuid(),
    'en',
    'English',
    'English',
    FALSE,
    TRUE,
    TRUE,
    EXTRACT(EPOCH FROM NOW())::BIGINT,
    EXTRACT(EPOCH FROM NOW())::BIGINT,
    FALSE
);

-- Additional languages
INSERT INTO languages (id, code, name, native_name, is_rtl, is_active, is_default, created_at, modified_at, deleted)
VALUES
    (gen_random_uuid(), 'de', 'German', 'Deutsch', FALSE, TRUE, FALSE, EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT, FALSE),
    (gen_random_uuid(), 'fr', 'French', 'Français', FALSE, TRUE, FALSE, EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT, FALSE),
    (gen_random_uuid(), 'es', 'Spanish', 'Español', FALSE, TRUE, FALSE, EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT, FALSE),
    (gen_random_uuid(), 'it', 'Italian', 'Italiano', FALSE, TRUE, FALSE, EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT, FALSE),
    (gen_random_uuid(), 'pt', 'Portuguese', 'Português', FALSE, TRUE, FALSE, EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT, FALSE),
    (gen_random_uuid(), 'ru', 'Russian', 'Русский', FALSE, TRUE, FALSE, EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT, FALSE),
    (gen_random_uuid(), 'zh', 'Chinese', '中文', FALSE, TRUE, FALSE, EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT, FALSE),
    (gen_random_uuid(), 'ja', 'Japanese', '日本語', FALSE, TRUE, FALSE, EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT, FALSE),
    (gen_random_uuid(), 'ar', 'Arabic', 'العربية', TRUE, TRUE, FALSE, EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT, FALSE);

-- Core localization keys
INSERT INTO localization_keys (id, key, category, description, context, created_at, modified_at, deleted)
VALUES
    (gen_random_uuid(), 'error.auth.invalid_token', 'error', 'Invalid JWT token error', 'Authentication', EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT, FALSE),
    (gen_random_uuid(), 'error.auth.expired_token', 'error', 'Expired JWT token error', 'Authentication', EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT, FALSE),
    (gen_random_uuid(), 'error.auth.unauthorized', 'error', 'Unauthorized access error', 'Authentication', EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT, FALSE),
    (gen_random_uuid(), 'error.validation.required_field', 'error', 'Required field validation', 'Validation', EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT, FALSE),
    (gen_random_uuid(), 'error.validation.invalid_format', 'error', 'Invalid format validation', 'Validation', EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT, FALSE),
    (gen_random_uuid(), 'error.database.connection_failed', 'error', 'Database connection error', 'Database', EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT, FALSE),
    (gen_random_uuid(), 'error.database.query_failed', 'error', 'Database query error', 'Database', EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT, FALSE),
    (gen_random_uuid(), 'ui.button.submit', 'ui', 'Submit button label', 'UI', EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT, FALSE),
    (gen_random_uuid(), 'ui.button.cancel', 'ui', 'Cancel button label', 'UI', EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT, FALSE),
    (gen_random_uuid(), 'ui.button.save', 'ui', 'Save button label', 'UI', EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT, FALSE),
    (gen_random_uuid(), 'message.welcome', 'message', 'Welcome message', 'General', EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT, FALSE),
    (gen_random_uuid(), 'message.success', 'message', 'Generic success message', 'General', EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT, FALSE);

-- English localizations
DO $$
DECLARE
    lang_id UUID;
    key_record RECORD;
BEGIN
    -- Get English language ID
    SELECT id INTO lang_id FROM languages WHERE code = 'en' LIMIT 1;

    -- Insert English translations
    FOR key_record IN SELECT id, key FROM localization_keys WHERE deleted = FALSE
    LOOP
        INSERT INTO localizations (id, key_id, language_id, value, version, approved, approved_by, approved_at, created_at, modified_at, deleted)
        VALUES (
            gen_random_uuid(),
            key_record.id,
            lang_id,
            CASE key_record.key
                WHEN 'error.auth.invalid_token' THEN 'Invalid authentication token'
                WHEN 'error.auth.expired_token' THEN 'Authentication token has expired'
                WHEN 'error.auth.unauthorized' THEN 'You are not authorized to access this resource'
                WHEN 'error.validation.required_field' THEN 'This field is required'
                WHEN 'error.validation.invalid_format' THEN 'Invalid format'
                WHEN 'error.database.connection_failed' THEN 'Database connection failed'
                WHEN 'error.database.query_failed' THEN 'Database query failed'
                WHEN 'ui.button.submit' THEN 'Submit'
                WHEN 'ui.button.cancel' THEN 'Cancel'
                WHEN 'ui.button.save' THEN 'Save'
                WHEN 'message.welcome' THEN 'Welcome to HelixTrack'
                WHEN 'message.success' THEN 'Operation completed successfully'
                ELSE key_record.key
            END,
            1,
            TRUE,
            'system',
            EXTRACT(EPOCH FROM NOW())::BIGINT,
            EXTRACT(EPOCH FROM NOW())::BIGINT,
            EXTRACT(EPOCH FROM NOW())::BIGINT,
            FALSE
        );
    END LOOP;
END $$;

-- Comments
COMMENT ON TABLE languages IS 'Supported languages in the HelixTrack system';
COMMENT ON TABLE localization_keys IS 'Master list of all localization keys';
COMMENT ON TABLE localizations IS 'Actual localized strings for each language';
COMMENT ON TABLE localization_catalogs IS 'Pre-built catalogs for fast retrieval';
COMMENT ON TABLE localization_cache_keys IS 'Cache key tracking for invalidation';
COMMENT ON TABLE localization_audit_log IS 'Audit trail for all localization changes';

COMMENT ON COLUMN languages.code IS 'ISO 639-1 language code (e.g., en, de, fr)';
COMMENT ON COLUMN languages.is_rtl IS 'Right-to-left language indicator (e.g., Arabic, Hebrew)';
COMMENT ON COLUMN languages.is_default IS 'Default fallback language (only one allowed)';
COMMENT ON COLUMN localizations.value IS 'Localized string (encrypted with SQL Cipher)';
COMMENT ON COLUMN localizations.plural_forms IS 'JSON object with plural form rules';
COMMENT ON COLUMN localizations.variables IS 'JSON array of variable placeholders';
COMMENT ON COLUMN localization_catalogs.catalog_data IS 'Complete catalog as JSON key-value pairs';
COMMENT ON COLUMN localization_catalogs.checksum IS 'SHA-256 checksum for integrity verification';

-- Database version
CREATE TABLE IF NOT EXISTS schema_version (
    version INTEGER PRIMARY KEY,
    applied_at BIGINT NOT NULL
);

INSERT INTO schema_version (version, applied_at)
VALUES (1, EXTRACT(EPOCH FROM NOW())::BIGINT);

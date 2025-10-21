-- HelixTrack Localization Service - Migration V1 to V2
-- Adds version tracking functionality

-- Migration Version: 1.2
-- Date: 2025-01-15
-- Description: Add version tracking for localization catalog management

-- ============================================================================
-- Table: localization_versions
-- Purpose: Track localization catalog versions with metadata
-- ============================================================================

CREATE TABLE IF NOT EXISTS localization_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    version_number VARCHAR(20) NOT NULL UNIQUE,  -- e.g., "1.0.0", "1.2.5"
    version_type VARCHAR(20) NOT NULL,            -- "major", "minor", "patch"
    description TEXT,                             -- Version change description
    keys_count INTEGER NOT NULL DEFAULT 0,        -- Number of keys in this version
    languages_count INTEGER NOT NULL DEFAULT 0,   -- Number of languages in this version
    total_localizations INTEGER NOT NULL DEFAULT 0, -- Total localizations count
    created_by VARCHAR(255),                      -- Username who created this version
    created_at BIGINT NOT NULL,                   -- Unix timestamp (milliseconds)

    -- Metadata
    metadata JSONB,                                -- Additional version metadata

    -- Constraints
    CHECK (version_type IN ('major', 'minor', 'patch')),
    CHECK (keys_count >= 0),
    CHECK (languages_count >= 0),
    CHECK (total_localizations >= 0)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_localization_versions_created ON localization_versions(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_localization_versions_type ON localization_versions(version_type);

-- ============================================================================
-- Add version column to localization_catalogs
-- Purpose: Link catalogs to specific versions
-- ============================================================================

ALTER TABLE localization_catalogs
ADD COLUMN IF NOT EXISTS version_id UUID REFERENCES localization_versions(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_localization_catalogs_version ON localization_catalogs(version_id);

-- ============================================================================
-- Initial Version
-- Purpose: Create initial version 1.0.0 for existing data
-- ============================================================================

DO $$
DECLARE
    initial_version_id UUID;
    key_count INTEGER;
    lang_count INTEGER;
    loc_count INTEGER;
BEGIN
    -- Count existing data
    SELECT COUNT(*) INTO key_count FROM localization_keys WHERE deleted = false;
    SELECT COUNT(*) INTO lang_count FROM languages WHERE deleted = false;
    SELECT COUNT(*) INTO loc_count FROM localizations WHERE deleted = false;

    -- Create initial version only if data exists
    IF key_count > 0 OR lang_count > 0 OR loc_count > 0 THEN
        INSERT INTO localization_versions (
            version_number,
            version_type,
            description,
            keys_count,
            languages_count,
            total_localizations,
            created_by,
            created_at,
            metadata
        ) VALUES (
            '1.0.0',
            'major',
            'Initial localization version',
            key_count,
            lang_count,
            loc_count,
            'system',
            EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
            '{"migration": "V1.2", "auto_created": true}'::JSONB
        )
        RETURNING id INTO initial_version_id;

        -- Link existing catalogs to initial version
        UPDATE localization_catalogs
        SET version_id = initial_version_id
        WHERE version_id IS NULL;

        RAISE NOTICE 'Created initial version 1.0.0 with % keys, % languages, % localizations',
            key_count, lang_count, loc_count;
    END IF;
END $$;

-- ============================================================================
-- Function: auto_increment_version
-- Purpose: Automatically create new version on catalog changes
-- ============================================================================

CREATE OR REPLACE FUNCTION auto_increment_version()
RETURNS TRIGGER AS $$
DECLARE
    latest_version VARCHAR(20);
    new_version VARCHAR(20);
    version_parts INTEGER[];
    change_type VARCHAR(20);
    new_version_id UUID;
    key_count INTEGER;
    lang_count INTEGER;
    loc_count INTEGER;
BEGIN
    -- Get latest version
    SELECT version_number INTO latest_version
    FROM localization_versions
    ORDER BY created_at DESC
    LIMIT 1;

    -- If no version exists, start with 1.0.0
    IF latest_version IS NULL THEN
        new_version := '1.0.0';
        change_type := 'major';
    ELSE
        -- Parse version (e.g., "1.2.5" -> [1, 2, 5])
        version_parts := string_to_array(latest_version, '.')::INTEGER[];

        -- Determine change type based on operation
        -- For now, all changes are patch versions
        change_type := 'patch';

        -- Increment patch version
        version_parts[3] := version_parts[3] + 1;
        new_version := version_parts[1] || '.' || version_parts[2] || '.' || version_parts[3];
    END IF;

    -- Count current data
    SELECT COUNT(*) INTO key_count FROM localization_keys WHERE deleted = false;
    SELECT COUNT(*) INTO lang_count FROM languages WHERE deleted = false;
    SELECT COUNT(*) INTO loc_count FROM localizations WHERE deleted = false;

    -- Create new version
    INSERT INTO localization_versions (
        version_number,
        version_type,
        description,
        keys_count,
        languages_count,
        total_localizations,
        created_by,
        created_at,
        metadata
    ) VALUES (
        new_version,
        change_type,
        'Auto-generated version on catalog update',
        key_count,
        lang_count,
        loc_count,
        'system',
        EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
        jsonb_build_object('auto_created', true, 'trigger', TG_TABLE_NAME)
    )
    RETURNING id INTO new_version_id;

    -- Link new catalog to new version
    NEW.version_id := new_version_id;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- Trigger: Create new version on catalog insert
-- ============================================================================

DROP TRIGGER IF EXISTS trigger_auto_increment_version ON localization_catalogs;

CREATE TRIGGER trigger_auto_increment_version
    BEFORE INSERT ON localization_catalogs
    FOR EACH ROW
    WHEN (NEW.version_id IS NULL)
    EXECUTE FUNCTION auto_increment_version();

-- ============================================================================
-- Migration Complete
-- ============================================================================

-- Log migration completion
DO $$
BEGIN
    RAISE NOTICE 'Migration V1.2 completed successfully';
    RAISE NOTICE '- Added localization_versions table';
    RAISE NOTICE '- Added version_id to localization_catalogs';
    RAISE NOTICE '- Created initial version 1.0.0 (if data exists)';
    RAISE NOTICE '- Added auto-versioning trigger';
END $$;

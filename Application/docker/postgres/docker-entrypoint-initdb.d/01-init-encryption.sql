-- =====================================================================
-- HelixTrack Core - PostgreSQL Encryption Initialization
-- Sets up pgcrypto extension and encryption functions
-- =====================================================================

-- Enable pgcrypto extension for encryption functions
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Enable pg_stat_statements for query statistics
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

-- Enable uuid-ossp for UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =====================================================================
-- Encryption Helper Functions
-- =====================================================================

-- Function: Encrypt sensitive text data
-- Usage: encrypt_text('sensitive data', 'encryption key')
CREATE OR REPLACE FUNCTION encrypt_text(data TEXT, key TEXT)
RETURNS BYTEA AS $$
BEGIN
    RETURN pgp_sym_encrypt(data, key);
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Function: Decrypt sensitive text data
-- Usage: decrypt_text(encrypted_data, 'encryption key')
CREATE OR REPLACE FUNCTION decrypt_text(encrypted_data BYTEA, key TEXT)
RETURNS TEXT AS $$
BEGIN
    RETURN pgp_sym_decrypt(encrypted_data, key);
EXCEPTION
    WHEN OTHERS THEN
        RETURN NULL;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Function: Hash password with bcrypt
-- Usage: hash_password('password')
CREATE OR REPLACE FUNCTION hash_password(password TEXT)
RETURNS TEXT AS $$
BEGIN
    RETURN crypt(password, gen_salt('bf', 10));
END;
$$ LANGUAGE plpgsql;

-- Function: Verify password against hash
-- Usage: verify_password('password', hashed_password)
CREATE OR REPLACE FUNCTION verify_password(password TEXT, password_hash TEXT)
RETURNS BOOLEAN AS $$
BEGIN
    RETURN password_hash = crypt(password, password_hash);
END;
$$ LANGUAGE plpgsql;

-- Function: Generate secure random token
-- Usage: generate_token(32) -- generates 32-byte token
CREATE OR REPLACE FUNCTION generate_token(length INTEGER)
RETURNS TEXT AS $$
BEGIN
    RETURN encode(gen_random_bytes(length), 'hex');
END;
$$ LANGUAGE plpgsql;

-- Function: Encrypt JSON data
-- Usage: encrypt_json('{"key": "value"}'::jsonb, 'encryption key')
CREATE OR REPLACE FUNCTION encrypt_json(data JSONB, key TEXT)
RETURNS BYTEA AS $$
BEGIN
    RETURN pgp_sym_encrypt(data::TEXT, key);
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Function: Decrypt JSON data
-- Usage: decrypt_json(encrypted_data, 'encryption key')
CREATE OR REPLACE FUNCTION decrypt_json(encrypted_data BYTEA, key TEXT)
RETURNS JSONB AS $$
BEGIN
    RETURN pgp_sym_decrypt(encrypted_data, key)::JSONB;
EXCEPTION
    WHEN OTHERS THEN
        RETURN NULL;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- =====================================================================
-- Audit Trigger Function for Encrypted Data Access
-- =====================================================================

-- Create audit log table for encryption operations
CREATE TABLE IF NOT EXISTS encryption_audit_log (
    id BIGSERIAL PRIMARY KEY,
    operation VARCHAR(50) NOT NULL,
    table_name VARCHAR(255),
    column_name VARCHAR(255),
    user_name VARCHAR(255),
    ip_address INET,
    accessed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    success BOOLEAN DEFAULT TRUE,
    error_message TEXT
);

-- Create index for faster queries
CREATE INDEX IF NOT EXISTS idx_encryption_audit_accessed_at
    ON encryption_audit_log(accessed_at);
CREATE INDEX IF NOT EXISTS idx_encryption_audit_user
    ON encryption_audit_log(user_name);
CREATE INDEX IF NOT EXISTS idx_encryption_audit_table
    ON encryption_audit_log(table_name);

-- Function: Audit encryption operations
CREATE OR REPLACE FUNCTION audit_encryption_operation(
    p_operation VARCHAR(50),
    p_table_name VARCHAR(255),
    p_column_name VARCHAR(255),
    p_success BOOLEAN DEFAULT TRUE,
    p_error_message TEXT DEFAULT NULL
)
RETURNS VOID AS $$
BEGIN
    INSERT INTO encryption_audit_log (
        operation,
        table_name,
        column_name,
        user_name,
        ip_address,
        success,
        error_message
    ) VALUES (
        p_operation,
        p_table_name,
        p_column_name,
        current_user,
        inet_client_addr(),
        p_success,
        p_error_message
    );
END;
$$ LANGUAGE plpgsql;

-- =====================================================================
-- Encryption Key Management
-- =====================================================================

-- Table to store encryption key metadata (NOT the keys themselves!)
CREATE TABLE IF NOT EXISTS encryption_key_metadata (
    key_id VARCHAR(50) PRIMARY KEY,
    key_purpose VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    rotated_at TIMESTAMP,
    rotation_count INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    algorithm VARCHAR(50) DEFAULT 'AES-256',
    CONSTRAINT chk_key_purpose CHECK (key_purpose IN (
        'database_encryption',
        'sensitive_fields',
        'jwt_secrets',
        'api_keys',
        'session_tokens'
    ))
);

-- Create index for active keys
CREATE INDEX IF NOT EXISTS idx_encryption_key_active
    ON encryption_key_metadata(is_active);

-- Function: Record key rotation
CREATE OR REPLACE FUNCTION rotate_encryption_key(p_key_id VARCHAR(50))
RETURNS VOID AS $$
BEGIN
    UPDATE encryption_key_metadata
    SET rotated_at = CURRENT_TIMESTAMP,
        rotation_count = rotation_count + 1
    WHERE key_id = p_key_id;

    -- Audit the rotation
    PERFORM audit_encryption_operation(
        'KEY_ROTATION',
        'encryption_key_metadata',
        p_key_id,
        TRUE,
        NULL
    );
END;
$$ LANGUAGE plpgsql;

-- =====================================================================
-- Security Policies
-- =====================================================================

-- Enable row-level security on audit log
ALTER TABLE encryption_audit_log ENABLE ROW LEVEL SECURITY;

-- Create policy to restrict access to audit logs
CREATE POLICY audit_log_select_policy ON encryption_audit_log
    FOR SELECT
    USING (user_name = current_user OR current_user IN (SELECT usename FROM pg_user WHERE usesuper));

-- =====================================================================
-- Cleanup and Maintenance Functions
-- =====================================================================

-- Function: Clean old audit logs (older than 90 days)
CREATE OR REPLACE FUNCTION cleanup_encryption_audit_log(days_to_keep INTEGER DEFAULT 90)
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM encryption_audit_log
    WHERE accessed_at < CURRENT_TIMESTAMP - INTERVAL '1 day' * days_to_keep;

    GET DIAGNOSTICS deleted_count = ROW_COUNT;

    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- =====================================================================
-- Initial Setup Complete
-- =====================================================================

-- Log the initialization
DO $$
BEGIN
    RAISE NOTICE 'HelixTrack PostgreSQL encryption initialized successfully';
    RAISE NOTICE 'Extensions enabled: pgcrypto, pg_stat_statements, uuid-ossp';
    RAISE NOTICE 'Encryption functions created and ready for use';
END $$;

# PostgreSQL Encryption Configuration

This directory contains configuration files for PostgreSQL with comprehensive encryption support.

## Overview

HelixTrack uses PostgreSQL with multiple layers of encryption:

1. **SSL/TLS Connection Encryption** - All connections encrypted
2. **pgcrypto Extension** - Column-level data encryption
3. **SCRAM-SHA-256 Authentication** - Strong password hashing
4. **Audit Logging** - Track all encryption operations

## Files

### Configuration Files

- `postgresql.conf` - Main PostgreSQL configuration with security settings
- `pg_hba.conf` - Host-based authentication requiring SSL connections
- `docker-entrypoint-initdb.d/00-generate-ssl-certs.sh` - SSL certificate generator
- `docker-entrypoint-initdb.d/01-init-encryption.sql` - Encryption functions and setup

### Docker Integration

These files are mounted into PostgreSQL containers via docker-compose:

```yaml
volumes:
  - ./docker/postgres/postgresql.conf:/etc/postgresql/postgresql.conf
  - ./docker/postgres/pg_hba.conf:/etc/postgresql/pg_hba.conf
  - ./docker/postgres/docker-entrypoint-initdb.d:/docker-entrypoint-initdb.d
```

## Encryption Features

### 1. SSL/TLS Connection Encryption

All database connections REQUIRE SSL/TLS encryption:

```bash
# Connection string example
postgres://user:password@host:5432/database?sslmode=require

# Environment variable
DATABASE_SSL_MODE=require
```

**Supported SSL Modes:**
- `require` - Require SSL (recommended for production)
- `verify-ca` - Verify server certificate against CA
- `verify-full` - Full certificate validation

### 2. Column-Level Encryption (pgcrypto)

Encrypt sensitive data at the column level:

```sql
-- Encrypt text data
INSERT INTO users (email, encrypted_ssn) VALUES (
    'user@example.com',
    encrypt_text('123-45-6789', 'encryption-key')
);

-- Decrypt text data
SELECT id, email, decrypt_text(encrypted_ssn, 'encryption-key') AS ssn
FROM users;

-- Encrypt JSON data
INSERT INTO settings (user_id, encrypted_config) VALUES (
    1,
    encrypt_json('{"theme": "dark", "notifications": true}'::jsonb, 'key')
);

-- Decrypt JSON data
SELECT decrypt_json(encrypted_config, 'key') AS config
FROM settings;
```

### 3. Password Hashing

Secure password storage with bcrypt:

```sql
-- Hash password
INSERT INTO users (username, password_hash) VALUES (
    'alice',
    hash_password('secret123')
);

-- Verify password
SELECT verify_password('secret123', password_hash) AS is_valid
FROM users
WHERE username = 'alice';
```

### 4. Token Generation

Generate secure random tokens:

```sql
-- Generate 32-byte session token
SELECT generate_token(32);
-- Returns: 'a1b2c3d4e5f6...' (64 hex characters)

-- Generate API key
SELECT generate_token(48);
-- Returns: 'x1y2z3...' (96 hex characters)
```

## Encryption Functions Reference

### Text Encryption

- `encrypt_text(data TEXT, key TEXT) RETURNS BYTEA`
- `decrypt_text(encrypted_data BYTEA, key TEXT) RETURNS TEXT`

### JSON Encryption

- `encrypt_json(data JSONB, key TEXT) RETURNS BYTEA`
- `decrypt_json(encrypted_data BYTEA, key TEXT) RETURNS JSONB`

### Password Management

- `hash_password(password TEXT) RETURNS TEXT`
- `verify_password(password TEXT, password_hash TEXT) RETURNS BOOLEAN`

### Token Generation

- `generate_token(length INTEGER) RETURNS TEXT`

### Encryption Key Management

- `rotate_encryption_key(key_id VARCHAR(50)) RETURNS VOID`

### Audit Operations

- `audit_encryption_operation(operation, table_name, column_name, success, error_message)`

## Encryption Key Management

### Key Storage

**IMPORTANT:** Encryption keys should NEVER be stored in the database!

Store keys in:
1. Environment variables (recommended for Docker)
2. Kubernetes Secrets
3. HashiCorp Vault
4. AWS KMS / Azure Key Vault / GCP KMS
5. Encrypted configuration files

### Key Rotation

Track key rotations in metadata table:

```sql
-- Record new encryption key
INSERT INTO encryption_key_metadata (key_id, key_purpose)
VALUES ('db_key_v1', 'database_encryption');

-- Rotate key
SELECT rotate_encryption_key('db_key_v1');

-- Check rotation history
SELECT * FROM encryption_key_metadata
WHERE key_id = 'db_key_v1';
```

### Recommended Key Rotation Schedule

- **Database encryption keys**: Every 90 days
- **JWT secrets**: Every 180 days
- **API keys**: Every 365 days
- **Session tokens**: Generated per session

## Audit Logging

All encryption operations are logged:

```sql
-- View encryption audit log
SELECT *
FROM encryption_audit_log
ORDER BY accessed_at DESC
LIMIT 100;

-- Check failed encryption attempts
SELECT *
FROM encryption_audit_log
WHERE success = FALSE;

-- Audit log for specific table
SELECT *
FROM encryption_audit_log
WHERE table_name = 'users'
  AND operation = 'ENCRYPT';

-- Cleanup old logs (older than 90 days)
SELECT cleanup_encryption_audit_log(90);
```

### Audit Log Schema

```sql
CREATE TABLE encryption_audit_log (
    id BIGSERIAL PRIMARY KEY,
    operation VARCHAR(50) NOT NULL,        -- ENCRYPT, DECRYPT, KEY_ROTATION
    table_name VARCHAR(255),               -- Table being accessed
    column_name VARCHAR(255),              -- Column being encrypted/decrypted
    user_name VARCHAR(255),                -- Database user
    ip_address INET,                       -- Client IP address
    accessed_at TIMESTAMP DEFAULT NOW(),   -- When operation occurred
    success BOOLEAN DEFAULT TRUE,          -- Did operation succeed?
    error_message TEXT                     -- Error details if failed
);
```

## Security Best Practices

### 1. Connection Security

```bash
# Always use SSL mode 'require' or stronger
DATABASE_URL=postgres://user:pass@host:5432/db?sslmode=require

# For maximum security, verify certificates
DATABASE_URL=postgres://user:pass@host:5432/db?sslmode=verify-full
```

### 2. Encryption Key Security

```bash
# Store keys in environment variables
ENCRYPTION_KEY_DATABASE=<generate-strong-key>
ENCRYPTION_KEY_SENSITIVE=<generate-strong-key>
ENCRYPTION_KEY_JWT=<generate-strong-key>

# Generate strong keys (32+ bytes)
openssl rand -hex 32
```

### 3. Password Policy

- Minimum 12 characters
- Use bcrypt with work factor 10+ (already configured)
- Never store plaintext passwords
- Rotate passwords regularly

### 4. Column Encryption Strategy

Encrypt these columns:
- Personal Identifiable Information (PII)
- Social Security Numbers
- Credit card numbers
- API keys and secrets
- OAuth tokens
- Session data
- Medical information

**Don't encrypt:**
- Primary keys
- Foreign keys
- Indexed search columns (use hashing instead)
- High-frequency query columns

### 5. Audit Log Retention

```sql
-- Schedule cleanup job (run daily)
SELECT cleanup_encryption_audit_log(90);  -- Keep 90 days

-- For compliance, adjust retention
SELECT cleanup_encryption_audit_log(365); -- Keep 1 year
SELECT cleanup_encryption_audit_log(2555); -- Keep 7 years (GDPR max)
```

## Example Implementation

### User Table with Encryption

```sql
-- Create users table
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    encrypted_ssn BYTEA,                    -- Encrypted SSN
    encrypted_api_key BYTEA,                -- Encrypted API key
    encrypted_config BYTEA,                 -- Encrypted JSON config
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Insert user with encrypted data
INSERT INTO users (username, email, password_hash, encrypted_ssn, encrypted_api_key)
VALUES (
    'alice',
    'alice@example.com',
    hash_password('secret123'),
    encrypt_text('123-45-6789', :encryption_key),
    encrypt_text(generate_token(32), :encryption_key)
);

-- Query with decryption
SELECT
    id,
    username,
    email,
    decrypt_text(encrypted_ssn, :encryption_key) AS ssn,
    decrypt_text(encrypted_api_key, :encryption_key) AS api_key
FROM users
WHERE username = 'alice';

-- Verify password
SELECT
    id,
    username,
    verify_password('secret123', password_hash) AS authenticated
FROM users
WHERE username = 'alice';
```

## SSL Certificate Management

### Development (Self-Signed)

Certificates are automatically generated on first start:

```bash
# Check certificates
docker exec <container> ls -la /var/lib/postgresql/*.{crt,key}

# Regenerate certificates (if needed)
docker exec <container> /docker-entrypoint-initdb.d/00-generate-ssl-certs.sh
```

### Production (CA-Signed)

Replace self-signed certificates with CA-signed certificates:

```bash
# 1. Generate CSR
openssl req -new -key server.key -out server.csr

# 2. Submit CSR to Certificate Authority

# 3. Receive signed certificate

# 4. Replace certificates in container
docker cp server.crt <container>:/var/lib/postgresql/server.crt
docker cp server.key <container>:/var/lib/postgresql/server.key
docker cp ca.crt <container>:/var/lib/postgresql/root.crt

# 5. Set permissions
docker exec <container> chown postgres:postgres /var/lib/postgresql/server.*
docker exec <container> chmod 600 /var/lib/postgresql/server.key

# 6. Restart PostgreSQL
docker restart <container>
```

## Performance Considerations

### Encryption Overhead

| Operation | Overhead | Recommendation |
|-----------|----------|----------------|
| SSL/TLS Connection | ~5-10% | Always use |
| Password Hashing (bcrypt) | ~100ms per hash | Acceptable for auth |
| Column Encryption (AES-256) | ~1-2% | Use selectively |
| JSON Encryption | ~2-5% | Use for sensitive configs |

### Optimization Tips

1. **Selective Encryption**: Only encrypt truly sensitive data
2. **Connection Pooling**: Reuse SSL connections
3. **Indexed Hashing**: Use indexed hash columns for searchable encrypted data
4. **Batch Operations**: Encrypt/decrypt in batches when possible

## Troubleshooting

### Connection Refused (SSL)

```bash
# Error: connection requires SSL
# Solution: Add sslmode parameter
DATABASE_URL=postgres://user:pass@host:5432/db?sslmode=require
```

### Certificate Verification Failed

```bash
# Error: certificate verify failed
# Solution 1: Use 'require' instead of 'verify-full' for self-signed certs
DATABASE_URL=...?sslmode=require

# Solution 2: Provide CA certificate
DATABASE_URL=...?sslmode=verify-full&sslrootcert=/path/to/root.crt
```

### Decryption Returns NULL

```sql
-- Issue: Wrong encryption key
-- Solution: Verify key matches encryption key

-- Check if data is encrypted correctly
SELECT
    id,
    encrypted_ssn IS NOT NULL AS has_encrypted_data,
    decrypt_text(encrypted_ssn, 'correct-key') IS NOT NULL AS can_decrypt
FROM users;
```

### Performance Degradation

```sql
-- Check encryption operations frequency
SELECT
    operation,
    COUNT(*) AS count,
    AVG(CASE WHEN success THEN 1 ELSE 0 END) * 100 AS success_rate
FROM encryption_audit_log
WHERE accessed_at > NOW() - INTERVAL '1 hour'
GROUP BY operation;

-- If too many operations, consider caching decrypted values in application
```

## Compliance

This encryption setup supports compliance with:

- **GDPR** - EU General Data Protection Regulation
- **HIPAA** - Health Insurance Portability and Accountability Act
- **PCI DSS** - Payment Card Industry Data Security Standard
- **SOC 2** - Service Organization Control 2
- **ISO 27001** - Information Security Management

## References

- [PostgreSQL SSL Support](https://www.postgresql.org/docs/current/ssl-tcp.html)
- [pgcrypto Documentation](https://www.postgresql.org/docs/current/pgcrypto.html)
- [PostgreSQL Security](https://www.postgresql.org/docs/current/auth-pg-hba-conf.html)
- [SCRAM Authentication](https://www.postgresql.org/docs/current/sasl-authentication.html)

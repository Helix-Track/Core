#!/bin/bash

# Import Chat Extension V2 to PostgreSQL
# This script creates a separate encrypted database for the Chat service

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Database configuration
DB_NAME="helixtrack_chat"
DB_USER="${POSTGRES_USER:-helixtrack_chat}"
DB_PASSWORD="${POSTGRES_PASSWORD:-$(openssl rand -base64 32)}"
DB_HOST="${POSTGRES_HOST:-localhost}"
DB_PORT="${POSTGRES_PORT:-5432}"

# SQL files
SCHEMA_FILE="$PROJECT_ROOT/Database/DDL/Extensions/Chats/Definition.V2.sql"
MIGRATION_FILE="$PROJECT_ROOT/Database/DDL/Extensions/Chats/Migration.V1.2.sql"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}HelixTrack Chat Service Database Setup${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Check if PostgreSQL is installed
if ! command -v psql &> /dev/null; then
    echo -e "${RED}Error: PostgreSQL client (psql) is not installed${NC}"
    exit 1
fi

# Check if schema file exists
if [ ! -f "$SCHEMA_FILE" ]; then
    echo -e "${RED}Error: Schema file not found: $SCHEMA_FILE${NC}"
    exit 1
fi

echo -e "${GREEN}Step 1: Creating database and user...${NC}"

# Create database and user (requires superuser privileges)
PGPASSWORD="${POSTGRES_ADMIN_PASSWORD}" psql -h "$DB_HOST" -p "$DB_PORT" -U postgres -c "CREATE DATABASE $DB_NAME;" 2>/dev/null || true
PGPASSWORD="${POSTGRES_ADMIN_PASSWORD}" psql -h "$DB_HOST" -p "$DB_PORT" -U postgres -c "CREATE USER $DB_USER WITH ENCRYPTED PASSWORD '$DB_PASSWORD';" 2>/dev/null || true
PGPASSWORD="${POSTGRES_ADMIN_PASSWORD}" psql -h "$DB_HOST" -p "$DB_PORT" -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;" 2>/dev/null || true

echo -e "${GREEN}Step 2: Enabling required extensions...${NC}"

# Enable UUID extension
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";" 2>/dev/null || true
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "CREATE EXTENSION IF NOT EXISTS \"pgcrypto\";" 2>/dev/null || true

echo -e "${GREEN}Step 3: Importing Chat V2 schema...${NC}"

# Import schema
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$SCHEMA_FILE"

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Chat V2 schema imported successfully${NC}"
else
    echo -e "${RED}✗ Failed to import Chat V2 schema${NC}"
    exit 1
fi

# Check if migration is needed (V1 tables exist)
V1_EXISTS=$(PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'chat');" | tr -d '[:space:]')

if [ "$V1_EXISTS" = "t" ]; then
    echo -e "${GREEN}Step 4: Running migration from V1 to V2...${NC}"
    PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$MIGRATION_FILE"

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Migration V1 -> V2 completed successfully${NC}"
    else
        echo -e "${RED}✗ Migration failed${NC}"
        exit 1
    fi
else
    echo -e "${BLUE}Step 4: Skipped (no V1 data to migrate)${NC}"
fi

echo -e "${GREEN}Step 5: Setting up encryption (SQL Cipher simulation)...${NC}"

# PostgreSQL doesn't support SQL Cipher directly, but we can use pgcrypto for column-level encryption
# This creates functions for encrypting/decrypting sensitive data

PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" <<EOF
-- Create encryption key storage (should be moved to separate secure storage in production)
CREATE TABLE IF NOT EXISTS encryption_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key_name VARCHAR(100) UNIQUE NOT NULL,
    key_value TEXT NOT NULL,
    created_at BIGINT NOT NULL
);

-- Insert master encryption key (change this in production!)
INSERT INTO encryption_keys (key_name, key_value, created_at)
VALUES ('master_key', encode(gen_random_bytes(32), 'hex'), extract(epoch from now())::bigint * 1000)
ON CONFLICT (key_name) DO NOTHING;

-- Create encryption/decryption functions
CREATE OR REPLACE FUNCTION encrypt_data(data TEXT) RETURNS TEXT AS \$\$
DECLARE
    key TEXT;
BEGIN
    SELECT key_value INTO key FROM encryption_keys WHERE key_name = 'master_key';
    RETURN encode(pgp_sym_encrypt(data, key), 'hex');
END;
\$\$ LANGUAGE plpgsql SECURITY DEFINER;

CREATE OR REPLACE FUNCTION decrypt_data(encrypted_data TEXT) RETURNS TEXT AS \$\$
DECLARE
    key TEXT;
BEGIN
    SELECT key_value INTO key FROM encryption_keys WHERE key_name = 'master_key';
    RETURN pgp_sym_decrypt(decode(encrypted_data, 'hex'), key);
END;
\$\$ LANGUAGE plpgsql SECURITY DEFINER;

-- Grant execute permissions
GRANT EXECUTE ON FUNCTION encrypt_data(TEXT) TO $DB_USER;
GRANT EXECUTE ON FUNCTION decrypt_data(TEXT) TO $DB_USER;
EOF

echo -e "${GREEN}✓ Encryption setup completed${NC}"

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Setup Complete!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "${BLUE}Database Configuration:${NC}"
echo -e "  Database: $DB_NAME"
echo -e "  User: $DB_USER"
echo -e "  Host: $DB_HOST"
echo -e "  Port: $DB_PORT"
echo ""
echo -e "${BLUE}Connection String:${NC}"
echo -e "  postgresql://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=require"
echo ""
echo -e "${RED}IMPORTANT: Save the database password securely!${NC}"
echo -e "  Password: $DB_PASSWORD"
echo ""

# Save connection details to config file
CONFIG_DIR="$PROJECT_ROOT/Services/Chat"
mkdir -p "$CONFIG_DIR/config"

cat > "$CONFIG_DIR/config/database.json" <<EOF
{
  "database": {
    "type": "postgresql",
    "host": "$DB_HOST",
    "port": $DB_PORT,
    "database": "$DB_NAME",
    "user": "$DB_USER",
    "password": "$DB_PASSWORD",
    "ssl_mode": "require",
    "max_connections": 100,
    "connection_timeout": 30
  }
}
EOF

echo -e "${GREEN}✓ Database configuration saved to: $CONFIG_DIR/config/database.json${NC}"
echo ""

exit 0

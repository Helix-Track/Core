# HelixTrack Key Manager

A comprehensive CLI tool for secure key generation, storage, and management for all HelixTrack Core services.

## Overview

The HelixTrack Key Manager provides a unified solution for managing all security-related keys across the HelixTrack ecosystem. It generates cryptographically secure keys, stores them safely, and provides tools for key rotation, export, and import.

## Features

✅ **Multiple Key Types**
- JWT signing secrets
- Database encryption keys (AES-256)
- TLS certificates and private keys
- Redis passwords
- API keys

✅ **Secure Storage**
- Encrypted file-based storage
- Separate storage per service
- Metadata tracking with JSON format
- Automatic permission management (0600 for keys, 0644 for certificates)

✅ **Key Management**
- Generate new keys with customizable lengths
- Rotate existing keys with version tracking
- List all managed keys
- Export keys to multiple formats (JSON, YAML, ENV)
- Import keys from backup files

✅ **Cross-Service Support**
- Authentication service
- Localization service
- Permissions engine
- And all other HelixTrack Core services

## Installation

### Building from Source

```bash
cd /home/milosvasic/Projects/HelixTrack/Core/Tools/KeyManager
go build -o keymanager ./cmd/main.go
```

### Binary Installation

```bash
# Copy to system path
sudo cp keymanager /usr/local/bin/
sudo chmod +x /usr/local/bin/keymanager
```

## Quick Start

### Generate a JWT Secret

```bash
# Generate JWT secret for authentication service
keymanager generate -type jwt -name auth-jwt-secret -service authentication

# Output:
# ✓ Key generated successfully!
#   Type:    jwt
#   Name:    auth-jwt-secret
#   Service: authentication
#   ID:      550e8400-e29b-41d4-a716-446655440000
#   Value:   dGhpc2lzYXRlc3RrZXl2YWx1ZQ==
```

### Generate Database Encryption Key

```bash
# Generate 32-byte AES-256 encryption key for localization service
keymanager generate -type db -name loc-db-key -service localization -length 32

# Output:
# ✓ Key generated successfully!
#   Type:    db
#   Name:    loc-db-key
#   Service: localization
#   ID:      550e8400-e29b-41d4-a716-446655440001
#   Value:   YW5vdGhlcnRlc3RrZXlmb3JkYXRhYmFzZQ==
```

### Generate TLS Certificate

```bash
# Generate TLS certificate and private key
keymanager generate -type tls -name service-tls -service localization

# Output:
# ✓ Key generated successfully!
#   Type:    tls
#   Name:    service-tls
#   Service: localization
#   ID:      550e8400-e29b-41d4-a716-446655440002
#   Cert:    keys/localization/tls/service-tls.crt
#   Key:     keys/localization/tls/service-tls.key
```

## Usage

### Commands

```
keymanager <command> [options]
```

#### Available Commands

- `generate` - Generate a new key
- `list` - List all managed keys
- `rotate` - Rotate an existing key
- `export` - Export keys to file
- `import` - Import keys from file
- `version` - Show version information

### Generate Command

Generate new cryptographic keys for services.

**Syntax:**
```bash
keymanager generate -type <type> -name <name> -service <service> [options]
```

**Options:**
- `-type string` - **Required**. Type of key (jwt, db, tls, redis, api)
- `-name string` - **Required**. Key name/identifier
- `-service string` - **Required**. Service name
- `-length int` - Key length in bytes (optional, defaults vary by type)
- `-output string` - Output file path (optional)

**Key Types and Default Lengths:**
| Type | Default Length | Purpose |
|------|----------------|---------|
| `jwt` | 64 bytes | JWT signing secrets |
| `db` | 32 bytes | Database encryption (AES-256) |
| `tls` | 2048-bit RSA | TLS certificates and keys |
| `redis` | 32 bytes | Redis authentication |
| `api` | 32 bytes | API authentication keys |

**Examples:**

```bash
# JWT secret with default 64-byte length
keymanager generate -type jwt -name my-jwt -service auth

# Custom length JWT secret
keymanager generate -type jwt -name long-jwt -service auth -length 128

# Database encryption key (must be 32 bytes)
keymanager generate -type db -name db-encrypt -service localization -length 32

# TLS certificate (generates both .crt and .key files)
keymanager generate -type tls -name web-tls -service web

# Redis password
keymanager generate -type redis -name cache-pwd -service localization -length 32

# API key with file export
keymanager generate -type api -name api-key -service api -output ./api-key.json
```

### List Command

Display all managed keys with metadata.

**Syntax:**
```bash
keymanager list
```

**Example Output:**
```
Found 5 key(s):

NAME                 TYPE            SERVICE              ID                             CREATED
---------------------------------------------------------------------------------------------------
auth-jwt-secret      jwt             authentication       550e8400-e29b-41d4-a716...     2025-10-21 12:30:00
loc-db-key           db              localization         660e8400-e29b-41d4-a716...     2025-10-21 12:31:00
service-tls          tls             localization         770e8400-e29b-41d4-a716...     2025-10-21 12:32:00
cache-pwd            redis           localization         880e8400-e29b-41d4-a716...     2025-10-21 12:33:00
api-key              api             api                  990e8400-e29b-41d4-a716...     2025-10-21 12:34:00
```

### Rotate Command

Rotate an existing key by generating a new version.

**Syntax:**
```bash
keymanager rotate -name <name> -service <service>
```

**Options:**
- `-name string` - **Required**. Name of key to rotate
- `-service string` - **Required**. Service name

**Example:**
```bash
keymanager rotate -name auth-jwt-secret -service authentication

# Output:
# ✓ Key rotated successfully!
#   Name:       auth-jwt-secret
#   Service:    authentication
#   Old ID:     550e8400-e29b-41d4-a716-446655440000
#   New ID:     aa0e8400-e29b-41d4-a716-446655440000
#   New Value:  bmV3cm90YXRlZGtleXZhbHVl
```

**Notes:**
- Rotation creates a new key with incremented version number
- Old key is replaced in storage
- For TLS certificates, new certificate files are generated
- The new key maintains same type and length as original

### Export Command

Export all keys to a file in various formats.

**Syntax:**
```bash
keymanager export -path <path> [-format <format>]
```

**Options:**
- `-path string` - **Required**. Export file path
- `-format string` - Export format (json, yaml, env). Default: json

**Supported Formats:**

**1. JSON Format (default)**
```bash
keymanager export -path ./backup/keys.json -format json
```

**Output:**
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "auth-jwt-secret",
    "service": "authentication",
    "type": "jwt",
    "value": "dGhpc2lzYXRlc3RrZXl2YWx1ZQ==",
    "metadata": {"length": "64"},
    "created_at": "2025-10-21T12:30:00Z",
    "version": 1
  }
]
```

**2. YAML Format**
```bash
keymanager export -path ./backup/keys.yaml -format yaml
```

**3. ENV Format**
```bash
keymanager export -path ./backup/keys.env -format env
```

**Output (.env file):**
```bash
AUTHENTICATION_AUTH_JWT_SECRET=dGhpc2lzYXRlc3RrZXl2YWx1ZQ==
LOCALIZATION_LOC_DB_KEY=YW5vdGhlcnRlc3RrZXlmb3JkYXRhYmFzZQ==
LOCALIZATION_SERVICE_TLS_CERT=keys/localization/tls/service-tls.crt
LOCALIZATION_SERVICE_TLS_KEY=keys/localization/tls/service-tls.key
```

### Import Command

Import keys from a backup file.

**Syntax:**
```bash
keymanager import -path <path>
```

**Options:**
- `-path string` - **Required**. Import file path

**Supported Formats:**
- JSON (.json)
- YAML (.yaml, .yml)

**Example:**
```bash
keymanager import -path ./backup/keys.json

# Output:
# ✓ Imported 5 key(s) successfully
```

**Notes:**
- Import auto-detects JSON or YAML format
- Existing keys with same name/service will be overwritten
- Key files are restored to their original locations

## Storage Structure

The Key Manager stores keys in a hierarchical directory structure:

```
keys/
├── keys.json                          # Metadata for all keys
├── authentication/
│   └── auth-jwt-secret.key            # JWT secret file
├── localization/
│   ├── loc-db-key.key                 # Database encryption key
│   └── tls/
│       ├── service-tls.crt            # TLS certificate
│       └── service-tls.key            # TLS private key
└── api/
    └── api-key.key                    # API key file
```

### Metadata File (keys.json)

The `keys.json` file contains metadata for all managed keys:

```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "auth-jwt-secret",
    "service": "authentication",
    "type": "jwt",
    "value": "base64-encoded-key",
    "metadata": {
      "length": "64"
    },
    "created_at": "2025-10-21T12:30:00Z",
    "expires_at": null,
    "version": 1
  }
]
```

### File Permissions

- **Key files**: 0600 (read/write owner only)
- **Certificate files**: 0644 (read all, write owner)
- **Metadata file**: 0600 (read/write owner only)
- **Directories**: 0700 (full access owner only)

## Security Considerations

### Key Generation

- All keys are generated using Go's `crypto/rand` package
- Cryptographically secure random number generation
- Minimum key lengths enforced:
  - JWT: 32 bytes minimum
  - Database: Exactly 32 bytes (AES-256)
  - Redis: 16 bytes minimum
  - API: 32 bytes minimum
  - TLS: 2048-bit RSA keys

### Key Storage

- Keys are stored with restrictive file permissions
- Metadata is separated from key values (except for TLS)
- Directory structure isolates keys by service
- Base64 encoding for binary key data

### Key Rotation

- Version tracking for audit trails
- Old keys are replaced, not deleted
- Unique IDs for each key version
- Timestamp tracking for rotation events

### Best Practices

1. **Rotate keys regularly**
   - JWT secrets: Every 90 days
   - Database keys: Annually or on breach
   - TLS certificates: Before expiration (1 year validity)
   - API keys: Every 6 months

2. **Backup keys securely**
   - Export keys to encrypted storage
   - Store backups off-site
   - Use JSON or YAML for portability

3. **Limit access**
   - Run keymanager as dedicated user
   - Restrict file system permissions
   - Use environment variables for deployment

4. **Monitor key usage**
   - Track key creation and rotation
   - Audit key access logs
   - Alert on unauthorized changes

## Integration with Services

### Authentication Service

```bash
# Generate JWT secret
keymanager generate -type jwt -name jwt-secret -service authentication -length 64

# Export for service configuration
keymanager export -path auth-keys.env -format env
```

**Service Configuration (configs/default.json):**
```json
{
  "security": {
    "jwt_secret": "value-from-keymanager",
    "jwt_issuer": "helixtrack-auth"
  }
}
```

### Localization Service

```bash
# Generate all required keys
keymanager generate -type db -name db-encryption -service localization -length 32
keymanager generate -type jwt -name jwt-secret -service localization -length 64
keymanager generate -type tls -name service-tls -service localization
keymanager generate -type redis -name redis-password -service localization -length 32

# Export for deployment
keymanager export -path localization-keys.json
```

**Service Configuration:**
```json
{
  "database": {
    "encryption_key": "value-from-keymanager"
  },
  "security": {
    "jwt_secret": "value-from-keymanager"
  },
  "service": {
    "tls_cert_file": "keys/localization/tls/service-tls.crt",
    "tls_key_file": "keys/localization/tls/service-tls.key"
  },
  "cache": {
    "redis": {
      "password": "value-from-keymanager"
    }
  }
}
```

## Testing

The Key Manager includes comprehensive test coverage (83.5%+) for all components.

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test -v -cover ./internal/generator/
go test -v -cover ./internal/storage/

# Run benchmarks
go test -bench=. ./internal/generator/
go test -bench=. ./internal/storage/
```

### Test Coverage

- **Generator package**: 83.5% coverage
  - Key generation for all types
  - Key rotation logic
  - Random byte generation
  - TLS certificate creation
  - Error handling

- **Storage package**: 83.6% coverage
  - Save/retrieve keys
  - List operations
  - Delete operations
  - Export/import (JSON, YAML, ENV)
  - Metadata management

## Troubleshooting

### Common Issues

**Issue:** Permission denied when creating keys directory
**Solution:** Ensure you have write permissions in the current directory or specify an alternative storage location

**Issue:** TLS certificate generation fails
**Solution:** Check that OpenSSL libraries are installed and accessible

**Issue:** Database key must be exactly 32 bytes error
**Solution:** Always specify `-length 32` for database encryption keys

**Issue:** Key not found error
**Solution:** Verify key name and service name match exactly (case-sensitive)

## API Reference

### Generator Package

```go
import "github.com/helixtrack/keymanager/internal/generator"

// Create new generator
g := generator.New()

// Generate JWT secret
key, err := g.GenerateJWTSecret("name", "service", 64)

// Generate database key
key, err := g.GenerateDatabaseKey("name", "service", 32)

// Generate TLS certificate
key, err := g.GenerateTLSCertificate("name", "service")

// Generate Redis password
key, err := g.GenerateRedisPassword("name", "service", 32)

// Generate API key
key, err := g.GenerateAPIKey("name", "service", 32)

// Rotate existing key
newKey, err := g.RotateKey(oldKey)
```

### Storage Package

```go
import "github.com/helixtrack/keymanager/internal/storage"

// Create new storage
store, err := storage.New()

// Save key
err := store.SaveKey(key)

// Get key
key, err := store.GetKey("name", "service")

// List all keys
keys, err := store.ListKeys()

// Delete key
err := store.DeleteKey("name", "service")

// Export keys
err := store.ExportKeys("path", "format")

// Import keys
count, err := store.ImportKeys("path")

// Export single key
err := store.ExportKeyToFile(key, "path")
```

## Contributing

Contributions are welcome! Please ensure:

1. All tests pass (`go test ./...`)
2. Code coverage remains above 80%
3. Documentation is updated
4. Examples are provided for new features

## License

This tool is part of the HelixTrack project and follows the same licensing terms.

## Support

For issues, questions, or contributions, please contact the HelixTrack development team.

---

**Version:** 1.0.0
**Last Updated:** 2025-10-21
**Maintained by:** HelixTrack Development Team

#!/bin/bash
###############################################################################
# HelixTrack Core - PostgreSQL SSL Certificate Generation
# Generates self-signed SSL certificates for PostgreSQL encryption
#
# For production use, replace these with proper CA-signed certificates
###############################################################################

set -e

echo "========================================="
echo "Generating SSL Certificates for PostgreSQL"
echo "========================================="

# Certificate details
CERT_DIR="/var/lib/postgresql"
DAYS_VALID=3650  # 10 years
COUNTRY="US"
STATE="State"
CITY="City"
ORG="HelixTrack"
OU="Database"
CN="postgresql.helixtrack.local"

# Check if certificates already exist
if [ -f "$CERT_DIR/server.crt" ] && [ -f "$CERT_DIR/server.key" ]; then
    echo "✓ SSL certificates already exist, skipping generation"
    exit 0
fi

echo "Generating new SSL certificates..."

# Generate private key
openssl genrsa -out "$CERT_DIR/server.key" 2048

# Generate self-signed certificate
openssl req -new -x509 \
    -days $DAYS_VALID \
    -key "$CERT_DIR/server.key" \
    -out "$CERT_DIR/server.crt" \
    -subj "/C=$COUNTRY/ST=$STATE/L=$CITY/O=$ORG/OU=$OU/CN=$CN"

# Generate root CA certificate (same as server cert for self-signed)
cp "$CERT_DIR/server.crt" "$CERT_DIR/root.crt"

# Set proper permissions
chmod 600 "$CERT_DIR/server.key"
chmod 644 "$CERT_DIR/server.crt"
chmod 644 "$CERT_DIR/root.crt"
chown postgres:postgres "$CERT_DIR/server.key" "$CERT_DIR/server.crt" "$CERT_DIR/root.crt"

echo "✓ SSL certificates generated successfully"
echo ""
echo "Certificate details:"
echo "  Private key: $CERT_DIR/server.key"
echo "  Certificate: $CERT_DIR/server.crt"
echo "  Root CA:     $CERT_DIR/root.crt"
echo "  Valid for:   $DAYS_VALID days"
echo ""
echo "⚠ WARNING: These are self-signed certificates!"
echo "⚠ For production, replace with CA-signed certificates"
echo ""

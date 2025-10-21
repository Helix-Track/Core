#!/bin/bash

# Generate self-signed TLS certificates for HTTP/3 QUIC development
# This script creates a server certificate and private key for local development

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Generating self-signed TLS certificates for HTTP/3 QUIC...${NC}"

# Create certs directory if it doesn't exist
mkdir -p certs

# Certificate configuration
COUNTRY="US"
STATE="California"
CITY="San Francisco"
ORG="HelixTrack"
ORG_UNIT="Localization Service"
COMMON_NAME="localhost"
DAYS=365

# Generate private key
echo -e "${YELLOW}Generating private key...${NC}"
openssl genrsa -out certs/server.key 2048

# Generate certificate signing request (CSR)
echo -e "${YELLOW}Generating certificate signing request...${NC}"
openssl req -new -key certs/server.key -out certs/server.csr \
    -subj "/C=${COUNTRY}/ST=${STATE}/L=${CITY}/O=${ORG}/OU=${ORG_UNIT}/CN=${COMMON_NAME}"

# Generate self-signed certificate
echo -e "${YELLOW}Generating self-signed certificate...${NC}"
openssl x509 -req -days ${DAYS} -in certs/server.csr -signkey certs/server.key -out certs/server.crt \
    -extfile <(printf "subjectAltName=DNS:localhost,IP:127.0.0.1")

# Remove CSR (no longer needed)
rm certs/server.csr

# Set proper permissions
chmod 600 certs/server.key
chmod 644 certs/server.crt

echo -e "${GREEN}✓ Certificates generated successfully!${NC}"
echo ""
echo "Certificate files:"
echo "  - Certificate: certs/server.crt"
echo "  - Private Key: certs/server.key"
echo ""
echo "Certificate details:"
openssl x509 -in certs/server.crt -noout -subject -dates
echo ""
echo -e "${YELLOW}Note: These are self-signed certificates for development only.${NC}"
echo -e "${YELLOW}For production, use certificates from a trusted CA.${NC}"

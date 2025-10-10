#!/bin/bash
# Organization Management API Test Script
BASE_URL="${BASE_URL:-http://localhost:8080}"
JWT_TOKEN="${JWT_TOKEN:-}"

echo "Testing Organization Management API"
echo "=================================================="

# Create Organization
curl -s -X POST "$BASE_URL/do" -H "Content-Type: application/json" \
  -d "{\"action\": \"organizationCreate\", \"jwt\": \"$JWT_TOKEN\", \"data\": {\"name\": \"Engineering\", \"description\": \"Engineering department\"}}" | jq '.'

# List Organizations
curl -s -X POST "$BASE_URL/do" -H "Content-Type: application/json" \
  -d "{\"action\": \"organizationList\", \"jwt\": \"$JWT_TOKEN\", \"data\": {}}" | jq '.'

echo "Organization tests completed"

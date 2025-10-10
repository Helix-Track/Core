#!/bin/bash
# Audit Management API Test Script
BASE_URL="${BASE_URL:-http://localhost:8080}"
JWT_TOKEN="${JWT_TOKEN:-}"

echo "Testing Audit Management API"
echo "=================================================="

# List Audit Entries
curl -s -X POST "$BASE_URL/do" -H "Content-Type: application/json" \
  -d "{\"action\": \"auditList\", \"jwt\": \"$JWT_TOKEN\", \"data\": {}}" | jq '.'

# Query Audit
curl -s -X POST "$BASE_URL/do" -H "Content-Type: application/json" \
  -d "{\"action\": \"auditQuery\", \"jwt\": \"$JWT_TOKEN\", \"data\": {\"startDate\": 1696118400, \"endDate\": 1697328000}}" | jq '.'

echo "Audit tests completed"

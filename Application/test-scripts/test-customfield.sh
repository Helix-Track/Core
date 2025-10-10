#!/bin/bash
# Custom Field Management API Test Script
BASE_URL="${BASE_URL:-http://localhost:8080}"
JWT_TOKEN="${JWT_TOKEN:-}"

echo "Testing Custom Field Management API"
echo "=================================================="

# Create Custom Field
curl -s -X POST "$BASE_URL/do" -H "Content-Type: application/json" \
  -d "{\"action\": \"customFieldCreate\", \"jwt\": \"$JWT_TOKEN\", \"data\": {\"name\": \"Story Points\", \"fieldType\": \"number\", \"description\": \"Agile story points\"}}" | jq '.'

# List Custom Fields
curl -s -X POST "$BASE_URL/do" -H "Content-Type: application/json" \
  -d "{\"action\": \"customFieldList\", \"jwt\": \"$JWT_TOKEN\", \"data\": {}}" | jq '.'

echo "Custom Field tests completed"

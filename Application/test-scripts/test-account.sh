#!/bin/bash
# Account Management API Test Script
BASE_URL="${BASE_URL:-http://localhost:8080}"
JWT_TOKEN="${JWT_TOKEN:-}"

echo "Testing Account Management API"
echo "=================================================="

# Create Account
curl -s -X POST "$BASE_URL/do" -H "Content-Type: application/json" \
  -d "{\"action\": \"accountCreate\", \"jwt\": \"$JWT_TOKEN\", \"data\": {\"name\": \"Acme Corp\", \"tier\": \"enterprise\", \"description\": \"Main account\"}}" | jq '.'

# List Accounts
curl -s -X POST "$BASE_URL/do" -H "Content-Type: application/json" \
  -d "{\"action\": \"accountList\", \"jwt\": \"$JWT_TOKEN\", \"data\": {}}" | jq '.'

echo "Account tests completed"

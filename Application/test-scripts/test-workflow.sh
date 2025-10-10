#!/bin/bash
# Workflow Management API Test Script
BASE_URL="${BASE_URL:-http://localhost:8080}"
JWT_TOKEN="${JWT_TOKEN:-}"

echo "Testing Workflow Management API"
echo "=================================================="

# Create Workflow
curl -s -X POST "$BASE_URL/do" -H "Content-Type: application/json" \
  -d "{\"action\": \"workflowCreate\", \"jwt\": \"$JWT_TOKEN\", \"data\": {\"name\": \"Bug Workflow\", \"description\": \"Workflow for bug tickets\"}}" | jq '.'

# List Workflows
curl -s -X POST "$BASE_URL/do" -H "Content-Type: application/json" \
  -d "{\"action\": \"workflowList\", \"jwt\": \"$JWT_TOKEN\", \"data\": {}}" | jq '.'

echo "Workflow tests completed"

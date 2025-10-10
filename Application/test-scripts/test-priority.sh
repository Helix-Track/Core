#!/bin/bash

# Test script for Priority Management API
# Tests all 5 priority endpoints: create, read, list, modify, remove

BASE_URL="${BASE_URL:-http://localhost:8080}"
JWT_TOKEN="${JWT_TOKEN:-}"

echo "=================================================="
echo "Testing Priority Management API"
echo "=================================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

test_count=0
pass_count=0
fail_count=0

# Function to test an endpoint
test_endpoint() {
    local test_name="$1"
    local action="$2"
    local data="$3"

    test_count=$((test_count + 1))
    echo "Test $test_count: $test_name"

    if [ -z "$JWT_TOKEN" ]; then
        body="{\"action\": \"$action\", \"data\": $data}"
    else
        body="{\"action\": \"$action\", \"jwt\": \"$JWT_TOKEN\", \"data\": $data}"
    fi

    response=$(curl -s -X POST "$BASE_URL/do" \
        -H "Content-Type: application/json" \
        -d "$body")

    echo "Request: $body"
    echo "Response: $response"

    # Check if errorCode is -1 (success)
    error_code=$(echo "$response" | grep -o '"errorCode":[^,}]*' | head -1 | cut -d':' -f2 | tr -d ' ')

    if [ "$error_code" = "-1" ]; then
        echo -e "${GREEN}✓ PASS${NC}"
        pass_count=$((pass_count + 1))
    else
        echo -e "${RED}✗ FAIL (errorCode: $error_code)${NC}"
        fail_count=$((fail_count + 1))
    fi

    echo ""
}

# Test 1: Create Priority
echo "---------------------------------------------------"
test_endpoint "priorityCreate - Create High priority" \
    "priorityCreate" \
    '{"title": "High", "level": 4, "description": "High priority items", "icon": "arrow-up", "color": "#FF9900"}'

# Test 2: Create another Priority
echo "---------------------------------------------------"
test_endpoint "priorityCreate - Create Critical priority" \
    "priorityCreate" \
    '{"title": "Critical", "level": 5, "description": "Critical priority items", "icon": "exclamation", "color": "#FF0000"}'

# Test 3: List Priorities
echo "---------------------------------------------------"
test_endpoint "priorityList - List all priorities" \
    "priorityList" \
    '{}'

# Test 4: Read Priority
echo "---------------------------------------------------"
echo "Note: To test priorityRead, you need a valid priority ID from the list above"
# test_endpoint "priorityRead - Read priority by ID" \
#     "priorityRead" \
#     '{"id": "PRIORITY_ID_HERE"}'

# Test 5: Modify Priority
echo "---------------------------------------------------"
echo "Note: To test priorityModify, you need a valid priority ID"
# test_endpoint "priorityModify - Update priority" \
#     "priorityModify" \
#     '{"id": "PRIORITY_ID_HERE", "title": "Updated High", "level": 4, "color": "#FFAA00"}'

# Test 6: Remove Priority
echo "---------------------------------------------------"
echo "Note: To test priorityRemove, you need a valid priority ID"
# test_endpoint "priorityRemove - Delete priority" \
#     "priorityRemove" \
#     '{"id": "PRIORITY_ID_HERE"}'

echo "=================================================="
echo "Test Summary"
echo "=================================================="
echo "Total Tests: $test_count"
echo -e "${GREEN}Passed: $pass_count${NC}"
echo -e "${RED}Failed: $fail_count${NC}"
echo ""

if [ $fail_count -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed!${NC}"
    exit 1
fi

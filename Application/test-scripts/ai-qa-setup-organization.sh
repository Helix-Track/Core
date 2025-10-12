#!/bin/bash

# AI QA - Organization Setup Script
# Creates TechCorp Global organization structure with teams and users

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="${BASE_URL:-http://localhost:8080}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DATA_FILE="$SCRIPT_DIR/ai-qa-data-organization.json"
OUTPUT_DIR="$SCRIPT_DIR/ai-qa-output"
TOKENS_FILE="$OUTPUT_DIR/tokens.json"

# Counters
total_operations=0
successful_operations=0
failed_operations=0

# Create output directory
mkdir -p "$OUTPUT_DIR"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}AI QA - Organization Setup${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${CYAN}Target URL:${NC} $BASE_URL"
echo -e "${CYAN}Data File:${NC} $DATA_FILE"
echo -e "${CYAN}Output Dir:${NC} $OUTPUT_DIR"
echo ""

# Function to make API call
api_call() {
    local action="$1"
    local data="$2"
    local jwt="${3:-}"
    local description="$4"

    total_operations=$((total_operations + 1))
    echo -e "${YELLOW}[$total_operations] $description${NC}"

    local body
    if [ -z "$jwt" ]; then
        body="{\"action\": \"$action\", \"data\": $data}"
    else
        body="{\"action\": \"$action\", \"jwt\": \"$jwt\", \"data\": $data}"
    fi

    local response
    response=$(curl -s -X POST "$BASE_URL/do" \
        -H "Content-Type: application/json" \
        -d "$body")

    local error_code
    error_code=$(echo "$response" | grep -o '"errorCode":[^,}]*' | head -1 | cut -d':' -f2 | tr -d ' ')

    if [ "$error_code" = "-1" ]; then
        echo -e "  ${GREEN}✓ SUCCESS${NC}"
        successful_operations=$((successful_operations + 1))
        echo "$response"
        return 0
    else
        echo -e "  ${RED}✗ FAILED (errorCode: $error_code)${NC}"
        echo -e "  ${RED}Response: $response${NC}"
        failed_operations=$((failed_operations + 1))
        return 1
    fi
}

# Initialize tokens storage
echo "{}" > "$TOKENS_FILE"

# Step 1: Create Account
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${CYAN}Step 1: Creating Account${NC}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

ACCOUNT_DATA=$(cat "$DATA_FILE" | jq -c '.account')
ACCOUNT_RESPONSE=$(api_call "accountCreate" "$ACCOUNT_DATA" "" "Creating TechCorp Global account")
ACCOUNT_ID=$(echo "$ACCOUNT_RESPONSE" | jq -r '.data.id // empty')

if [ -n "$ACCOUNT_ID" ]; then
    echo -e "  ${GREEN}Account ID: $ACCOUNT_ID${NC}"
    jq ".account_id = \"$ACCOUNT_ID\"" "$TOKENS_FILE" > "$TOKENS_FILE.tmp" && mv "$TOKENS_FILE.tmp" "$TOKENS_FILE"
else
    echo -e "  ${RED}Failed to extract account ID${NC}"
    exit 1
fi

echo ""

# Step 2: Create Organization
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${CYAN}Step 2: Creating Organization${NC}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

ORG_DATA=$(cat "$DATA_FILE" | jq -c ".organization + {\"accountId\": \"$ACCOUNT_ID\"}")
ORG_RESPONSE=$(api_call "organizationCreate" "$ORG_DATA" "" "Creating TechCorp Engineering organization")
ORG_ID=$(echo "$ORG_RESPONSE" | jq -r '.data.id // empty')

if [ -n "$ORG_ID" ]; then
    echo -e "  ${GREEN}Organization ID: $ORG_ID${NC}"
    jq ".organization_id = \"$ORG_ID\"" "$TOKENS_FILE" > "$TOKENS_FILE.tmp" && mv "$TOKENS_FILE.tmp" "$TOKENS_FILE"
else
    echo -e "  ${RED}Failed to extract organization ID${NC}"
    exit 1
fi

echo ""

# Step 3: Create Teams
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${CYAN}Step 3: Creating Teams${NC}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

TEAM_COUNT=$(cat "$DATA_FILE" | jq '.teams | length')
echo "{\"teams\": []}" > "$OUTPUT_DIR/teams.json"

for i in $(seq 0 $((TEAM_COUNT - 1))); do
    TEAM_NAME=$(cat "$DATA_FILE" | jq -r ".teams[$i].name")
    TEAM_DATA=$(cat "$DATA_FILE" | jq -c ".teams[$i] + {\"organizationId\": \"$ORG_ID\"}")

    TEAM_RESPONSE=$(api_call "teamCreate" "$TEAM_DATA" "" "Creating team: $TEAM_NAME")
    TEAM_ID=$(echo "$TEAM_RESPONSE" | jq -r '.data.id // empty')

    if [ -n "$TEAM_ID" ]; then
        echo -e "  ${GREEN}Team ID: $TEAM_ID${NC}"
        jq ".teams += [{\"name\": \"$TEAM_NAME\", \"id\": \"$TEAM_ID\"}]" "$OUTPUT_DIR/teams.json" > "$OUTPUT_DIR/teams.json.tmp" && mv "$OUTPUT_DIR/teams.json.tmp" "$OUTPUT_DIR/teams.json"
    fi
    echo ""
done

echo ""

# Step 4: Register and Authenticate Users
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${CYAN}Step 4: Registering Users${NC}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

USER_COUNT=$(cat "$DATA_FILE" | jq '.users | length')
echo "{\"users\": []}" > "$OUTPUT_DIR/users.json"

for i in $(seq 0 $((USER_COUNT - 1))); do
    USERNAME=$(cat "$DATA_FILE" | jq -r ".users[$i].username")
    EMAIL=$(cat "$DATA_FILE" | jq -r ".users[$i].email")
    NAME=$(cat "$DATA_FILE" | jq -r ".users[$i].name")
    PASSWORD=$(cat "$DATA_FILE" | jq -r ".users[$i].password")
    ROLE=$(cat "$DATA_FILE" | jq -r ".users[$i].role")
    TEAM=$(cat "$DATA_FILE" | jq -r ".users[$i].team")

    echo -e "${YELLOW}Registering: $NAME ($USERNAME)${NC}"

    # Register user
    USER_DATA="{\"username\": \"$USERNAME\", \"email\": \"$EMAIL\", \"name\": \"$NAME\", \"password\": \"$PASSWORD\", \"role\": \"$ROLE\", \"organizationId\": \"$ORG_ID\"}"
    USER_RESPONSE=$(api_call "userRegister" "$USER_DATA" "" "  Registering user")
    USER_ID=$(echo "$USER_RESPONSE" | jq -r '.data.id // empty')

    if [ -n "$USER_ID" ]; then
        echo -e "  ${GREEN}User ID: $USER_ID${NC}"

        # Authenticate to get JWT token
        AUTH_DATA="{\"username\": \"$USERNAME\", \"password\": \"$PASSWORD\"}"
        AUTH_RESPONSE=$(api_call "authenticate" "$AUTH_DATA" "" "  Authenticating user")
        JWT_TOKEN=$(echo "$AUTH_RESPONSE" | jq -r '.data.token // empty')

        if [ -n "$JWT_TOKEN" ]; then
            echo -e "  ${GREEN}JWT Token: ${JWT_TOKEN:0:50}...${NC}"

            # Save user info
            jq ".users += [{\"username\": \"$USERNAME\", \"id\": \"$USER_ID\", \"email\": \"$EMAIL\", \"name\": \"$NAME\", \"role\": \"$ROLE\", \"team\": \"$TEAM\", \"token\": \"$JWT_TOKEN\"}]" "$OUTPUT_DIR/users.json" > "$OUTPUT_DIR/users.json.tmp" && mv "$OUTPUT_DIR/users.json.tmp" "$OUTPUT_DIR/users.json"
        fi
    fi
    echo ""
done

echo ""

# Step 5: Assign Users to Teams
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${CYAN}Step 5: Assigning Users to Teams${NC}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# Get admin token (Alice Johnson)
ADMIN_TOKEN=$(jq -r '.users[] | select(.username == "alice.johnson") | .token' "$OUTPUT_DIR/users.json")

if [ -z "$ADMIN_TOKEN" ] || [ "$ADMIN_TOKEN" = "null" ]; then
    echo -e "${RED}Failed to get admin token${NC}"
    exit 1
fi

# Assign users to teams
for i in $(seq 0 $((USER_COUNT - 1))); do
    USERNAME=$(cat "$DATA_FILE" | jq -r ".users[$i].username")
    TEAM_NAME=$(cat "$DATA_FILE" | jq -r ".users[$i].team")

    # Skip if team is "Management" (not a real team in our system)
    if [ "$TEAM_NAME" = "Management" ]; then
        continue
    fi

    USER_ID=$(jq -r ".users[] | select(.username == \"$USERNAME\") | .id" "$OUTPUT_DIR/users.json")
    TEAM_ID=$(jq -r ".teams[] | select(.name == \"$TEAM_NAME\") | .id" "$OUTPUT_DIR/teams.json")

    if [ -n "$USER_ID" ] && [ -n "$TEAM_ID" ]; then
        ASSIGN_DATA="{\"teamId\": \"$TEAM_ID\", \"userId\": \"$USER_ID\"}"
        api_call "teamAddMember" "$ASSIGN_DATA" "$ADMIN_TOKEN" "Assigning $USERNAME to $TEAM_NAME"
    fi
    echo ""
done

echo ""

# Summary
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Organization Setup Complete${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${CYAN}Summary:${NC}"
echo -e "  Total Operations: $total_operations"
echo -e "  ${GREEN}Successful: $successful_operations${NC}"
echo -e "  ${RED}Failed: $failed_operations${NC}"
echo ""
echo -e "${CYAN}Created Resources:${NC}"
echo -e "  Account ID: ${GREEN}$ACCOUNT_ID${NC}"
echo -e "  Organization ID: ${GREEN}$ORG_ID${NC}"
echo -e "  Teams: ${GREEN}$(jq '.teams | length' "$OUTPUT_DIR/teams.json")${NC}"
echo -e "  Users: ${GREEN}$(jq '.users | length' "$OUTPUT_DIR/users.json")${NC}"
echo ""
echo -e "${CYAN}Output Files:${NC}"
echo -e "  ${GREEN}$OUTPUT_DIR/tokens.json${NC} - Account and organization IDs"
echo -e "  ${GREEN}$OUTPUT_DIR/teams.json${NC} - Team information"
echo -e "  ${GREEN}$OUTPUT_DIR/users.json${NC} - User information and JWT tokens"
echo ""

if [ $failed_operations -eq 0 ]; then
    echo -e "${GREEN}✓ All operations completed successfully!${NC}"
    exit 0
else
    echo -e "${YELLOW}⚠ Some operations failed. Review the output above.${NC}"
    exit 1
fi

#!/bin/bash

# AI QA - Android Mobile Client Simulation
# Simulates an Android mobile app user interacting with HelixTrack Core

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="${BASE_URL:-http://localhost:8080}"
WS_URL="${WS_URL:-ws://localhost:8080/ws}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUTPUT_DIR="$SCRIPT_DIR/ai-qa-output"
USERS_FILE="$OUTPUT_DIR/users.json"
PROJECTS_FILE="$OUTPUT_DIR/projects.json"
CLIENT_LOG="$OUTPUT_DIR/android-client.log"

# Simulation parameters (mobile users check less frequently)
SIMULATION_DURATION="${SIMULATION_DURATION:-300}"  # 5 minutes default
ACTION_INTERVAL="${ACTION_INTERVAL:-15}"            # 15 seconds between actions (slower than web)

# Counters
actions_performed=0
successful_actions=0
failed_actions=0

echo -e "${MAGENTA}========================================${NC}"
echo -e "${MAGENTA}Android App Client Simulation${NC}"
echo -e "${MAGENTA}========================================${NC}"
echo ""

# Check prerequisites
if [ ! -f "$USERS_FILE" ] || [ ! -f "$PROJECTS_FILE" ]; then
    echo -e "${RED}Error: Setup not completed!${NC}"
    echo -e "${YELLOW}Run ai-qa-setup-organization.sh and ai-qa-setup-projects.sh first${NC}"
    exit 1
fi

# Initialize log
echo "Android App Client Simulation - $(date)" > "$CLIENT_LOG"
echo "======================================" >> "$CLIENT_LOG"
echo "" >> "$CLIENT_LOG"

# Select a random user for this session
USER_COUNT=$(jq '.users | length' "$USERS_FILE")
RANDOM_USER_INDEX=$((RANDOM % USER_COUNT))
USER_NAME=$(jq -r ".users[$RANDOM_USER_INDEX].name" "$USERS_FILE")
USER_TOKEN=$(jq -r ".users[$RANDOM_USER_INDEX].token" "$USERS_FILE")
USER_EMAIL=$(jq -r ".users[$RANDOM_USER_INDEX].email" "$USERS_FILE")

echo -e "${CYAN}Simulating User:${NC} $USER_NAME ($USER_EMAIL)"
echo -e "${CYAN}Device:${NC} Android Mobile (Pixel 7, Android 14)"
echo -e "${CYAN}Session Duration:${NC} $SIMULATION_DURATION seconds"
echo -e "${CYAN}Action Interval:${NC} $ACTION_INTERVAL seconds"
echo ""

log_action() {
    local message="$1"
    echo -e "[$(date '+%Y-%m-%d %H:%M:%S')] $message" >> "$CLIENT_LOG"
    echo -e "$message"
}

# Function to make API call
api_call() {
    local action="$1"
    local data="$2"
    local description="$3"

    actions_performed=$((actions_performed + 1))
    log_action "${YELLOW}[Android] $description${NC}"

    local body="{\"action\": \"$action\", \"jwt\": \"$USER_TOKEN\", \"data\": $data, \"clientType\": \"android\", \"clientVersion\": \"2.5.0\"}"

    local response
    response=$(curl -s -X POST "$BASE_URL/do" \
        -H "Content-Type: application/json" \
        -H "User-Agent: HelixTrack-Android/2.5.0 (Android 14; Pixel 7)" \
        -d "$body")

    local error_code
    error_code=$(echo "$response" | grep -o '"errorCode":[^,}]*' | head -1 | cut -d':' -f2 | tr -d ' ')

    if [ "$error_code" = "-1" ]; then
        log_action "  ${GREEN}✓ SUCCESS${NC}"
        successful_actions=$((successful_actions + 1))
        echo "$response"
        return 0
    else
        log_action "  ${RED}✗ FAILED (errorCode: $error_code)${NC}"
        failed_actions=$((failed_actions + 1))
        return 1
    fi
}

# Function to simulate typical Android app actions
simulate_android_actions() {
    local action_type=$((RANDOM % 12))

    case $action_type in
        0|1)
            # Pull to refresh - check for updates (very common on mobile)
            log_action "${BLUE}📱 Pull to Refresh: Checking for updates${NC}"
            api_call "notificationList" "{}" "Pull to refresh - notifications" > /dev/null
            ;;
        2|3)
            # View assigned tickets (common mobile use case)
            log_action "${BLUE}📱 My Tickets: Loading assigned tickets${NC}"
            api_call "ticketList" "{\"assignedToMe\": true}" "Fetching my tickets" > /dev/null
            ;;
        4)
            # Quick view ticket details
            local project_count=$(jq '.projects | length' "$PROJECTS_FILE")
            if [ "$project_count" -gt 0 ]; then
                local proj_index=$((RANDOM % project_count))
                local project_key=$(jq -r ".projects[$proj_index].key" "$PROJECTS_FILE")

                log_action "${BLUE}📱 Ticket View: Opening ticket details${NC}"
                # Note: Would need ticket ID in production
            fi
            ;;
        5)
            # Add quick comment (mobile users often add short comments)
            log_action "${BLUE}📱 Quick Comment: Adding mobile comment${NC}"
            local comment_text="👍 LGTM - reviewed on mobile"
            # Note: Would need ticket ID in production
            ;;
        6)
            # Update ticket status (swipe gesture simulation)
            log_action "${BLUE}📱 Swipe Action: Updating ticket status${NC}"
            # Note: Would need ticket ID in production
            ;;
        7)
            # Check notifications (very common on mobile)
            log_action "${BLUE}📱 Notifications: Checking new notifications${NC}"
            api_call "notificationList" "{\"unreadOnly\": true}" "Fetching unread notifications" > /dev/null
            ;;
        8)
            # Voice-to-text comment (mobile feature)
            log_action "${BLUE}📱 Voice Input: Adding voice-transcribed comment${NC}"
            local voice_comment="I have reviewed this ticket and it looks good to proceed with the implementation phase"
            # Note: Would need ticket ID in production
            ;;
        9)
            # Take photo and attach (mobile feature)
            log_action "${BLUE}📱 Camera: Taking photo attachment${NC}"
            # Note: Would need ticket ID and file upload in production
            ;;
        10)
            # View recent activity
            log_action "${BLUE}📱 Activity Feed: Loading recent updates${NC}"
            api_call "activityList" "{\"limit\": 20}" "Fetching recent activity" > /dev/null
            ;;
        11)
            # Sync offline changes (mobile feature)
            log_action "${BLUE}📱 Sync: Synchronizing offline changes${NC}"
            api_call "health" "{}" "Connectivity check" > /dev/null
            ;;
    esac
}

# Main simulation loop
start_time=$(date +%s)
end_time=$((start_time + SIMULATION_DURATION))

log_action "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
log_action "${GREEN}Android App Session Started${NC}"
log_action "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# App launch - initial sync
log_action "${BLUE}📱 App Launch: Initializing${NC}"
api_call "health" "{}" "Health check" > /dev/null
api_call "userProfile" "{}" "Loading user profile" > /dev/null
api_call "notificationList" "{\"unreadOnly\": true}" "Initial notifications sync" > /dev/null
echo ""

# Simulation loop
while [ $(date +%s) -lt $end_time ]; do
    simulate_android_actions
    echo ""
    sleep $ACTION_INTERVAL
done

# Summary
echo ""
log_action "${MAGENTA}========================================${NC}"
log_action "${MAGENTA}Android App Session Complete${NC}"
log_action "${MAGENTA}========================================${NC}"
echo ""
log_action "${CYAN}Session Summary:${NC}"
log_action "  User: $USER_NAME"
log_action "  Device: Android Mobile (Pixel 7)"
log_action "  Duration: $SIMULATION_DURATION seconds"
log_action "  Actions Performed: $actions_performed"
log_action "  ${GREEN}Successful: $successful_actions${NC}"
log_action "  ${RED}Failed: $failed_actions${NC}"
log_action "  Success Rate: $(awk "BEGIN {printf \"%.1f\", ($successful_actions/$actions_performed)*100}")%"
echo ""
log_action "${CYAN}Log File:${NC} $CLIENT_LOG"
echo ""

if [ $failed_actions -eq 0 ]; then
    log_action "${GREEN}✓ All Android app actions completed successfully!${NC}"
    exit 0
else
    log_action "${YELLOW}⚠ Some actions failed. Check log for details.${NC}"
    exit 1
fi

#!/bin/bash

# AI QA - Project Workflow Setup Script
# Creates projects with epics, stories, tasks, and subtasks

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
DATA_FILE="$SCRIPT_DIR/ai-qa-data-projects.json"
OUTPUT_DIR="$SCRIPT_DIR/ai-qa-output"
USERS_FILE="$OUTPUT_DIR/users.json"
TOKENS_FILE="$OUTPUT_DIR/tokens.json"

# Counters
total_operations=0
successful_operations=0
failed_operations=0

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}AI QA - Project Workflow Setup${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Check if organization setup was completed
if [ ! -f "$USERS_FILE" ] || [ ! -f "$TOKENS_FILE" ]; then
    echo -e "${RED}Error: Organization setup not completed!${NC}"
    echo -e "${YELLOW}Run ai-qa-setup-organization.sh first${NC}"
    exit 1
fi

# Get admin token and organization ID
ADMIN_TOKEN=$(jq -r '.users[] | select(.username == "alice.johnson") | .token' "$USERS_FILE")
ORG_ID=$(jq -r '.organization_id' "$TOKENS_FILE")

if [ -z "$ADMIN_TOKEN" ] || [ "$ADMIN_TOKEN" = "null" ]; then
    echo -e "${RED}Failed to get admin token${NC}"
    exit 1
fi

if [ -z "$ORG_ID" ] || [ "$ORG_ID" = "null" ]; then
    echo -e "${RED}Failed to get organization ID${NC}"
    exit 1
fi

echo -e "${CYAN}Admin Token:${NC} ${ADMIN_TOKEN:0:50}..."
echo -e "${CYAN}Organization ID:${NC} $ORG_ID"
echo ""

# Function to make API call
api_call() {
    local action="$1"
    local data="$2"
    local jwt="$3"
    local description="$4"

    total_operations=$((total_operations + 1))
    echo -e "${YELLOW}[$total_operations] $description${NC}"

    local body="{\"action\": \"$action\", \"jwt\": \"$jwt\", \"data\": $data}"

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
        failed_operations=$((failed_operations + 1))
        echo "$response" >&2
        return 1
    fi
}

# Initialize projects storage
echo "{\"projects\": []}" > "$OUTPUT_DIR/projects.json"

# Process each project
PROJECT_COUNT=$(jq '.projects | length' "$DATA_FILE")

for p in $(seq 0 $((PROJECT_COUNT - 1))); do
    PROJECT_KEY=$(jq -r ".projects[$p].key" "$DATA_FILE")
    PROJECT_NAME=$(jq -r ".projects[$p].name" "$DATA_FILE")
    PROJECT_DESC=$(jq -r ".projects[$p].description" "$DATA_FILE")
    PROJECT_TYPE=$(jq -r ".projects[$p].type" "$DATA_FILE")
    METHODOLOGY=$(jq -r ".projects[$p].methodology" "$DATA_FILE")

    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}Creating Project: $PROJECT_NAME ($PROJECT_KEY)${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""

    # Create project
    PROJECT_DATA=$(jq -c ".projects[$p] | {key: .key, name: .name, description: .description, type: .type, organizationId: \"$ORG_ID\"}" "$DATA_FILE")
    PROJECT_RESPONSE=$(api_call "projectCreate" "$PROJECT_DATA" "$ADMIN_TOKEN" "Creating project $PROJECT_KEY")
    PROJECT_ID=$(echo "$PROJECT_RESPONSE" | jq -r '.data.id // empty')

    if [ -z "$PROJECT_ID" ]; then
        echo -e "${RED}Failed to create project $PROJECT_KEY${NC}"
        continue
    fi

    echo -e "  ${GREEN}Project ID: $PROJECT_ID${NC}"
    echo ""

    # Create priorities for project
    echo -e "${BLUE}  Creating Priorities...${NC}"
    PRIORITY_COUNT=$(jq ".projects[$p].priorities | length" "$DATA_FILE")
    PRIORITIES_JSON="[]"

    for pri in $(seq 0 $((PRIORITY_COUNT - 1))); do
        PRI_NAME=$(jq -r ".projects[$p].priorities[$pri].name" "$DATA_FILE")
        PRI_DATA=$(jq -c ".projects[$p].priorities[$pri] + {\"projectId\": \"$PROJECT_ID\"}" "$DATA_FILE")

        PRI_RESPONSE=$(api_call "priorityCreate" "$PRI_DATA" "$ADMIN_TOKEN" "  Creating priority: $PRI_NAME")
        PRI_ID=$(echo "$PRI_RESPONSE" | jq -r '.data.id // empty')

        if [ -n "$PRI_ID" ]; then
            PRIORITIES_JSON=$(echo "$PRIORITIES_JSON" | jq ". += [{\"name\": \"$PRI_NAME\", \"id\": \"$PRI_ID\"}]")
        fi
    done
    echo ""

    # Create statuses for project (workflow)
    echo -e "${BLUE}  Creating Workflow Statuses...${NC}"
    STATUS_COUNT=$(jq ".projects[$p].statuses | length" "$DATA_FILE")
    STATUSES_JSON="[]"

    for st in $(seq 0 $((STATUS_COUNT - 1))); do
        ST_NAME=$(jq -r ".projects[$p].statuses[$st].name" "$DATA_FILE")
        ST_DATA=$(jq -c ".projects[$p].statuses[$st] + {\"projectId\": \"$PROJECT_ID\"}" "$DATA_FILE")

        ST_RESPONSE=$(api_call "statusCreate" "$ST_DATA" "$ADMIN_TOKEN" "  Creating status: $ST_NAME")
        ST_ID=$(echo "$ST_RESPONSE" | jq -r '.data.id // empty')

        if [ -n "$ST_ID" ]; then
            STATUSES_JSON=$(echo "$STATUSES_JSON" | jq ". += [{\"name\": \"$ST_NAME\", \"id\": \"$ST_ID\"}]")
        fi
    done
    echo ""

    # Create ticket types
    echo -e "${BLUE}  Creating Ticket Types...${NC}"
    TICKET_TYPES='[
        {"name": "Epic", "description": "Large body of work", "icon": "book", "color": "#8B00FF"},
        {"name": "Story", "description": "User story", "icon": "bookmark", "color": "#00B8D9"},
        {"name": "Task", "description": "Development task", "icon": "check-square", "color": "#00875A"},
        {"name": "Sub-task", "description": "Sub-task", "icon": "chevron-right", "color": "#5E6C84"},
        {"name": "Bug", "description": "Bug or defect", "icon": "bug", "color": "#DE350B"}
    ]'

    TYPE_COUNT=$(echo "$TICKET_TYPES" | jq 'length')
    TYPES_JSON="[]"

    for tt in $(seq 0 $((TYPE_COUNT - 1))); do
        TYPE_NAME=$(echo "$TICKET_TYPES" | jq -r ".[$tt].name")
        TYPE_DATA=$(echo "$TICKET_TYPES" | jq -c ".[$tt] + {\"projectId\": \"$PROJECT_ID\"}" )

        TYPE_RESPONSE=$(api_call "ticketTypeCreate" "$TYPE_DATA" "$ADMIN_TOKEN" "  Creating ticket type: $TYPE_NAME")
        TYPE_ID=$(echo "$TYPE_RESPONSE" | jq -r '.data.id // empty')

        if [ -n "$TYPE_ID" ]; then
            TYPES_JSON=$(echo "$TYPES_JSON" | jq ". += [{\"name\": \"$TYPE_NAME\", \"id\": \"$TYPE_ID\"}]")
        fi
    done
    echo ""

    # Create epics
    echo -e "${BLUE}  Creating Epics...${NC}"
    EPIC_COUNT=$(jq ".projects[$p].epics | length" "$DATA_FILE")
    EPICS_JSON="[]"

    EPIC_TYPE_ID=$(echo "$TYPES_JSON" | jq -r '.[] | select(.name == "Epic") | .id')
    STORY_TYPE_ID=$(echo "$TYPES_JSON" | jq -r '.[] | select(.name == "Story") | .id')
    TASK_TYPE_ID=$(echo "$TYPES_JSON" | jq -r '.[] | select(.name == "Task") | .id')
    SUBTASK_TYPE_ID=$(echo "$TYPES_JSON" | jq -r '.[] | select(.name == "Sub-task") | .id')
    DEFAULT_STATUS_ID=$(echo "$STATUSES_JSON" | jq -r '.[0].id')
    DEFAULT_PRIORITY_ID=$(echo "$PRIORITIES_JSON" | jq -r '.[1].id // .[0].id')

    for e in $(seq 0 $((EPIC_COUNT - 1))); do
        EPIC_KEY=$(jq -r ".projects[$p].epics[$e].key" "$DATA_FILE")
        EPIC_TITLE=$(jq -r ".projects[$p].epics[$e].title" "$DATA_FILE")
        EPIC_DESC=$(jq -r ".projects[$p].epics[$e].description" "$DATA_FILE")
        STORY_COUNT=$(jq -r ".projects[$p].epics[$e].stories" "$DATA_FILE")
        TASK_COUNT=$(jq -r ".projects[$p].epics[$e].tasks" "$DATA_FILE")
        SUBTASK_COUNT=$(jq -r ".projects[$p].epics[$e].subtasks" "$DATA_FILE")

        EPIC_DATA="{\"projectId\": \"$PROJECT_ID\", \"title\": \"$EPIC_TITLE\", \"description\": \"$EPIC_DESC\", \"typeId\": \"$EPIC_TYPE_ID\", \"statusId\": \"$DEFAULT_STATUS_ID\", \"priorityId\": \"$DEFAULT_PRIORITY_ID\"}"

        EPIC_RESPONSE=$(api_call "ticketCreate" "$EPIC_DATA" "$ADMIN_TOKEN" "  Creating epic: $EPIC_KEY - $EPIC_TITLE")
        EPIC_ID=$(echo "$EPIC_RESPONSE" | jq -r '.data.id // empty')

        if [ -z "$EPIC_ID" ]; then
            echo -e "    ${RED}Failed to create epic${NC}"
            continue
        fi

        EPICS_JSON=$(echo "$EPICS_JSON" | jq ". += [{\"key\": \"$EPIC_KEY\", \"id\": \"$EPIC_ID\", \"title\": \"$EPIC_TITLE\"}]")

        # Create stories for this epic (create 2-3 stories per epic)
        STORIES_TO_CREATE=$((STORY_COUNT > 3 ? 3 : STORY_COUNT))
        for s in $(seq 1 $STORIES_TO_CREATE); do
            STORY_TITLE="Story $s: ${EPIC_TITLE:0:40}"
            STORY_DESC="User story for epic $EPIC_KEY - Story $s"

            STORY_DATA="{\"projectId\": \"$PROJECT_ID\", \"title\": \"$STORY_TITLE\", \"description\": \"$STORY_DESC\", \"typeId\": \"$STORY_TYPE_ID\", \"statusId\": \"$DEFAULT_STATUS_ID\", \"priorityId\": \"$DEFAULT_PRIORITY_ID\", \"parentId\": \"$EPIC_ID\"}"

            STORY_RESPONSE=$(api_call "ticketCreate" "$STORY_DATA" "$ADMIN_TOKEN" "    Creating story: $STORY_TITLE")
            STORY_ID=$(echo "$STORY_RESPONSE" | jq -r '.data.id // empty')

            if [ -n "$STORY_ID" ]; then
                # Create tasks for this story (2 tasks per story)
                for t in $(seq 1 2); do
                    TASK_TITLE="Task $t for story $s"
                    TASK_DESC="Implementation task for story"

                    TASK_DATA="{\"projectId\": \"$PROJECT_ID\", \"title\": \"$TASK_TITLE\", \"description\": \"$TASK_DESC\", \"typeId\": \"$TASK_TYPE_ID\", \"statusId\": \"$DEFAULT_STATUS_ID\", \"priorityId\": \"$DEFAULT_PRIORITY_ID\", \"parentId\": \"$STORY_ID\"}"

                    TASK_RESPONSE=$(api_call "ticketCreate" "$TASK_DATA" "$ADMIN_TOKEN" "      Creating task: $TASK_TITLE")
                    TASK_ID=$(echo "$TASK_RESPONSE" | jq -r '.data.id // empty')

                    if [ -n "$TASK_ID" ]; then
                        # Create subtasks for this task (2 subtasks per task)
                        for sub in $(seq 1 2); do
                            SUBTASK_TITLE="Subtask $sub for task $t"
                            SUBTASK_DESC="Implementation subtask"

                            SUBTASK_DATA="{\"projectId\": \"$PROJECT_ID\", \"title\": \"$SUBTASK_TITLE\", \"description\": \"$SUBTASK_DESC\", \"typeId\": \"$SUBTASK_TYPE_ID\", \"statusId\": \"$DEFAULT_STATUS_ID\", \"priorityId\": \"$DEFAULT_PRIORITY_ID\", \"parentId\": \"$TASK_ID\"}"

                            api_call "ticketCreate" "$SUBTASK_DATA" "$ADMIN_TOKEN" "        Creating subtask: $SUBTASK_TITLE" > /dev/null
                        done
                    fi
                done
            fi
        done
        echo ""
    done

    # Save project info
    jq ".projects += [{\"key\": \"$PROJECT_KEY\", \"id\": \"$PROJECT_ID\", \"name\": \"$PROJECT_NAME\", \"epics\": $EPICS_JSON, \"priorities\": $PRIORITIES_JSON, \"statuses\": $STATUSES_JSON, \"types\": $TYPES_JSON}]" "$OUTPUT_DIR/projects.json" > "$OUTPUT_DIR/projects.json.tmp" && mv "$OUTPUT_DIR/projects.json.tmp" "$OUTPUT_DIR/projects.json"

    echo -e "${GREEN}✓ Project $PROJECT_KEY setup complete${NC}"
    echo ""
done

# Summary
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Project Workflow Setup Complete${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${CYAN}Summary:${NC}"
echo -e "  Total Operations: $total_operations"
echo -e "  ${GREEN}Successful: $successful_operations${NC}"
echo -e "  ${RED}Failed: $failed_operations${NC}"
echo ""
echo -e "${CYAN}Created Projects:${NC}"

PROJECT_COUNT=$(jq '.projects | length' "$OUTPUT_DIR/projects.json")
for p in $(seq 0 $((PROJECT_COUNT - 1))); do
    PROJECT_KEY=$(jq -r ".projects[$p].key" "$OUTPUT_DIR/projects.json")
    PROJECT_NAME=$(jq -r ".projects[$p].name" "$OUTPUT_DIR/projects.json")
    EPIC_COUNT=$(jq ".projects[$p].epics | length" "$OUTPUT_DIR/projects.json")
    echo -e "  ${GREEN}$PROJECT_KEY${NC}: $PROJECT_NAME ($EPIC_COUNT epics)"
done

echo ""
echo -e "${CYAN}Output File:${NC}"
echo -e "  ${GREEN}$OUTPUT_DIR/projects.json${NC} - Project information"
echo ""

if [ $failed_operations -eq 0 ]; then
    echo -e "${GREEN}✓ All operations completed successfully!${NC}"
    exit 0
else
    echo -e "${YELLOW}⚠ Some operations failed. Review the output above.${NC}"
    exit 1
fi

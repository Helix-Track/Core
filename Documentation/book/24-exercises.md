# Chapter 24: Hands-On Exercises

[← Previous: Advanced Use Cases](23-advanced-use-cases.md) | [Back to Table of Contents](README.md) | [Next: Troubleshooting →](25-troubleshooting.md)

---

## Introduction

This chapter provides hands-on exercises to help you master HelixTrack Core. Each exercise builds on previous knowledge and includes:

- **Objective**: What you'll learn
- **Prerequisites**: What you need before starting
- **Steps**: Detailed instructions
- **Verification**: How to confirm success
- **Challenges**: Extra credit for advanced users

> **Note**: These exercises assume you have HelixTrack Core running locally on `http://localhost:8080`

---

## Exercise 1: Basic Ticket Management

### Objective
Learn to create, read, update, and manage tickets using the API.

### Prerequisites
- HelixTrack Core running
- curl installed
- Basic understanding of JSON

### Steps

#### Step 1: Check System Health

```bash
# Verify the system is running
curl http://localhost:8080/health

# Expected output:
# {"status":"healthy"}
```

#### Step 2: Get API Version

```bash
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d '{"action":"version"}'

# Expected output:
# {
#   "errorCode": -1,
#   "data": {
#     "version": "2.0.0",
#     "api": "2.0.0"
#   }
# }
```

#### Step 3: List All Priorities

```bash
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d '{
    "action": "priorityList",
    "data": {}
  }'

# Note the priority IDs for later use
```

#### Step 4: List All Resolutions

```bash
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d '{
    "action": "resolutionList",
    "data": {}
  }'

# Note the resolution IDs
```

#### Step 5: Create a New Priority

```bash
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d '{
    "action": "priorityCreate",
    "data": {
      "title": "Urgent",
      "level": 4,
      "description": "Urgent priority for time-sensitive issues",
      "color": "#FF6600",
      "icon": "fire"
    }
  }'

# Save the returned ID
```

#### Step 6: Read the Priority You Created

```bash
# Replace PRIORITY_ID with the ID from step 5
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d '{
    "action": "priorityRead",
    "data": {
      "id": "PRIORITY_ID"
    }
  }'
```

#### Step 7: Modify the Priority

```bash
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d '{
    "action": "priorityModify",
    "data": {
      "id": "PRIORITY_ID",
      "title": "Super Urgent",
      "color": "#FF0000"
    }
  }'
```

### Verification

✅ **Success Criteria:**
- All API calls return `errorCode: -1`
- You can list, create, read, and modify priorities
- Modified priority shows updated values

### Challenges

1. **Challenge 1**: Create a complete set of 5 priorities for your project
2. **Challenge 2**: Create 3 custom resolutions (e.g., "Deferred", "Won't Implement", "Merged")
3. **Challenge 3**: Write a shell script to automate priority creation

---

## Exercise 2: Sprint Planning

### Objective
Set up a complete sprint with board, cycle, and tickets.

### Prerequisites
- Completed Exercise 1
- Understanding of Agile/Scrum concepts

### Steps

#### Step 1: Create a Workflow

```bash
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d '{
    "action": "workflowCreate",
    "data": {
      "name": "Agile Development Workflow",
      "description": "Standard workflow for agile development"
    }
  }'

# Save the workflow ID
```

#### Step 2: Create Ticket Statuses

```bash
# Status 1: To Do
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d '{
    "action": "ticketStatusCreate",
    "data": {
      "name": "To Do",
      "color": "#CCCCCC",
      "description": "Tasks not yet started"
    }
  }'

# Status 2: In Progress
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d '{
    "action": "ticketStatusCreate",
    "data": {
      "name": "In Progress",
      "color": "#0066CC",
      "description": "Tasks currently being worked on"
    }
  }'

# Status 3: Done
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d '{
    "action": "ticketStatusCreate",
    "data": {
      "name": "Done",
      "color": "#00CC66",
      "description": "Completed tasks"
    }
  }'

# Save all status IDs
```

#### Step 3: Create Ticket Types

```bash
# Type 1: Story
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d '{
    "action": "ticketTypeCreate",
    "data": {
      "name": "Story",
      "icon": "book",
      "color": "#00CC00",
      "description": "User story"
    }
  }'

# Type 2: Bug
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d '{
    "action": "ticketTypeCreate",
    "data": {
      "name": "Bug",
      "icon": "bug",
      "color": "#CC0000",
      "description": "Software defect"
    }
  }'

# Save type IDs
```

#### Step 4: Create a Board

```bash
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d '{
    "action": "boardCreate",
    "data": {
      "title": "Development Board",
      "description": "Main development board for sprint work"
    }
  }'

# Save board ID
```

#### Step 5: Configure Board Columns

```bash
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d '{
    "action": "boardSetMetadata",
    "data": {
      "boardId": "BOARD_ID",
      "key": "columns",
      "value": "[{\"name\":\"To Do\",\"statusId\":\"STATUS_ID_1\"},{\"name\":\"In Progress\",\"statusId\":\"STATUS_ID_2\"},{\"name\":\"Done\",\"statusId\":\"STATUS_ID_3\"}]"
    }
  }'
```

#### Step 6: Create a Sprint (Cycle)

```bash
# Calculate timestamps for 2-week sprint
START_DATE=$(date -d "today" +%s)
END_DATE=$(date -d "+14 days" +%s)

curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"cycleCreate\",
    \"data\": {
      \"name\": \"Sprint 1\",
      \"type\": \"sprint\",
      \"startDate\": $START_DATE,
      \"endDate\": $END_DATE,
      \"description\": \"Our first sprint\"
    }
  }"

# Save cycle ID
```

### Verification

✅ **Success Criteria:**
- Workflow created
- 3 statuses created (To Do, In Progress, Done)
- 2 ticket types created (Story, Bug)
- Board created with column configuration
- Sprint created with correct dates

### Challenges

1. **Challenge 1**: Add workflow steps connecting your statuses
2. **Challenge 2**: Create a backlog board and a sprint board
3. **Challenge 3**: Set up board metadata for swimlanes

---

## Exercise 3: Custom Workflows

### Objective
Design and implement a custom workflow for your organization.

### Prerequisites
- Completed Exercise 2
- Understanding of workflow concepts

### Steps

#### Step 1: Design Your Workflow

Plan a workflow with these statuses:
1. New → Open
2. Open → In Progress
3. In Progress → Code Review
4. Code Review → Testing
5. Testing → Done
6. Any → Blocked

#### Step 2: Create All Statuses

```bash
# Create 6 statuses (New, Open, In Progress, Code Review, Testing, Done, Blocked)
for status in "New:#EEEEEE" "Open:#CCCCFF" "In Progress:#0066CC" "Code Review:#9933CC" "Testing:#FFCC00" "Done:#00CC66" "Blocked:#CC0000"
do
  NAME=$(echo $status | cut -d: -f1)
  COLOR=$(echo $status | cut -d: -f2)

  curl -X POST http://localhost:8080/do \
    -H "Content-Type: application/json" \
    -d "{
      \"action\": \"ticketStatusCreate\",
      \"data\": {
        \"name\": \"$NAME\",
        \"color\": \"$COLOR\"
      }
    }"
done
```

#### Step 3: Create Workflow Steps

```bash
# Step 1: New → Open
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d '{
    "action": "workflowStepCreate",
    "data": {
      "workflowId": "WORKFLOW_ID",
      "name": "Start Work",
      "fromStatusId": "NEW_STATUS_ID",
      "toStatusId": "OPEN_STATUS_ID",
      "order": 1
    }
  }'

# Step 2: Open → In Progress
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d '{
    "action": "workflowStepCreate",
    "data": {
      "workflowId": "WORKFLOW_ID",
      "name": "Begin Development",
      "fromStatusId": "OPEN_STATUS_ID",
      "toStatusId": "IN_PROGRESS_STATUS_ID",
      "order": 2
    }
  }'

# Continue for all transitions...
```

#### Step 4: List Workflow Steps

```bash
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d '{
    "action": "workflowStepList",
    "data": {
      "workflowId": "WORKFLOW_ID"
    }
  }'
```

### Verification

✅ **Success Criteria:**
- All 7 statuses created
- All workflow steps created
- Steps listed in correct order
- Workflow represents your development process

### Challenges

1. **Challenge 1**: Add conditional transitions (e.g., Testing can go back to In Progress if bugs found)
2. **Challenge 2**: Create a separate workflow for bugs vs. stories
3. **Challenge 3**: Document your workflow in a diagram

---

## Exercise 4: Multi-Tenancy Setup

### Objective
Set up a complete organizational hierarchy with accounts, organizations, and teams.

### Prerequisites
- Understanding of multi-tenancy concepts
- Administrative access

### Steps

#### Step 1: Create an Account

```bash
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d '{
    "action": "accountCreate",
    "data": {
      "name": "Acme Corporation",
      "tier": "enterprise",
      "description": "Main corporate account"
    }
  }'

# Save account ID
```

#### Step 2: Create Organizations

```bash
# Engineering Organization
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d '{
    "action": "organizationCreate",
    "data": {
      "name": "Engineering",
      "description": "Engineering department"
    }
  }'

# Product Organization
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d '{
    "action": "organizationCreate",
    "data": {
      "name": "Product Management",
      "description": "Product management department"
    }
  }'

# Save organization IDs
```

#### Step 3: Assign Organizations to Account

```bash
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d '{
    "action": "organizationAssignAccount",
    "data": {
      "organizationId": "ORG_ID_1",
      "accountId": "ACCOUNT_ID"
    }
  }'

curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d '{
    "action": "organizationAssignAccount",
    "data": {
      "organizationId": "ORG_ID_2",
      "accountId": "ACCOUNT_ID"
    }
  }'
```

#### Step 4: Create Teams

```bash
# Backend Team
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d '{
    "action": "teamCreate",
    "data": {
      "name": "Backend Team",
      "description": "Server-side development team"
    }
  }'

# Frontend Team
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d '{
    "action": "teamCreate",
    "data": {
      "name": "Frontend Team",
      "description": "Client-side development team"
    }
  }'

# QA Team
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d '{
    "action": "teamCreate",
    "data": {
      "name": "QA Team",
      "description": "Quality assurance team"
    }
  }'

# Save team IDs
```

#### Step 5: Assign Teams to Organizations

```bash
# Assign all 3 teams to Engineering org
for TEAM_ID in "BACKEND_TEAM_ID" "FRONTEND_TEAM_ID" "QA_TEAM_ID"
do
  curl -X POST http://localhost:8080/do \
    -H "Content-Type: application/json" \
    -d "{
      \"action\": \"teamAssignOrganization\",
      \"data\": {
        \"teamId\": \"$TEAM_ID\",
        \"organizationId\": \"ENG_ORG_ID\"
      }
    }"
done
```

#### Step 6: Verify Hierarchy

```bash
# List organizations for account
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d '{
    "action": "organizationListAccounts",
    "data": {
      "organizationId": "ORG_ID"
    }
  }'

# List teams for organization
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d '{
    "action": "teamListOrganizations",
    "data": {
      "teamId": "TEAM_ID"
    }
  }'
```

### Verification

✅ **Success Criteria:**
- 1 account created
- 2 organizations created and assigned to account
- 3 teams created and assigned to Engineering organization
- Hierarchy verified with list commands

### Challenges

1. **Challenge 1**: Create a complete org chart with 5 organizations and 10 teams
2. **Challenge 2**: Assign users to teams using `userAssignTeam`
3. **Challenge 3**: Set up permission contexts for each level

---

## Exercise 5: Version Tracking & Release Management

### Objective
Manage product versions and track affected/fix versions for tickets.

### Prerequisites
- Understanding of release management
- Ticket IDs from previous exercises

### Steps

#### Step 1: Create Product Versions

```bash
# Version 1.0.0
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"versionCreate\",
    \"data\": {
      \"name\": \"v1.0.0\",
      \"description\": \"Initial release\",
      \"releaseDate\": $(date -d '2025-01-01' +%s),
      \"released\": true
    }
  }"

# Version 1.1.0
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"versionCreate\",
    \"data\": {
      \"name\": \"v1.1.0\",
      \"description\": \"Feature release\",
      \"releaseDate\": $(date -d '2025-03-01' +%s),
      \"released\": true
    }
  }"

# Version 1.2.0 (upcoming)
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"versionCreate\",
    \"data\": {
      \"name\": \"v1.2.0\",
      \"description\": \"Next feature release\",
      \"releaseDate\": $(date -d '+30 days' +%s),
      \"released\": false
    }
  }"

# Version 2.0.0 (future)
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"versionCreate\",
    \"data\": {
      \"name\": \"v2.0.0\",
      \"description\": \"Major version\",
      \"releaseDate\": $(date -d '+90 days' +%s),
      \"released\": false
    }
  }"
```

#### Step 2: Add Affected Versions to a Bug Ticket

```bash
# Assume you found a bug in v1.0.0 and v1.1.0
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d '{
    "action": "versionAddAffected",
    "data": {
      "ticketId": "BUG_TICKET_ID",
      "versionId": "V1.0.0_ID"
    }
  }'

curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d '{
    "action": "versionAddAffected",
    "data": {
      "ticketId": "BUG_TICKET_ID",
      "versionId": "V1.1.0_ID"
    }
  }'
```

#### Step 3: Add Fix Version

```bash
# This bug will be fixed in v1.2.0
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d '{
    "action": "versionAddFix",
    "data": {
      "ticketId": "BUG_TICKET_ID",
      "versionId": "V1.2.0_ID"
    }
  }'
```

#### Step 4: List Versions for Ticket

```bash
# List affected versions
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d '{
    "action": "versionListAffected",
    "data": {
      "ticketId": "BUG_TICKET_ID"
    }
  }'

# List fix versions
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d '{
    "action": "versionListFix",
    "data": {
      "ticketId": "BUG_TICKET_ID"
    }
  }'
```

#### Step 5: Release a Version

```bash
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -d '{
    "action": "versionRelease",
    "data": {
      "id": "V1.2.0_ID"
    }
  }'
```

### Verification

✅ **Success Criteria:**
- 4 versions created
- Bug ticket has 2 affected versions
- Bug ticket has 1 fix version
- v1.2.0 marked as released

### Challenges

1. **Challenge 1**: Create a script to generate release notes based on fix versions
2. **Challenge 2**: Archive old versions (v1.0.0)
3. **Challenge 3**: Track which tickets are still pending in unreleased versions

---

## Bonus Exercise: Complete Project Setup

### Objective
Combine all previous exercises into a complete project setup.

### Task

Create a shell script that:
1. Sets up account, organization, and teams
2. Creates a workflow with statuses
3. Creates ticket types
4. Creates a board
5. Creates a sprint
6. Creates 5 sample tickets
7. Assigns tickets to the sprint
8. Adds tickets to the board
9. Creates 3 versions
10. Assigns affected/fix versions to tickets

### Success Criteria

- Script runs without errors
- All entities are created
- Relationships are properly established
- Can query all created entities

---

## Summary

After completing these exercises, you should be comfortable with:

✅ Basic CRUD operations
✅ Sprint and board management
✅ Custom workflow creation
✅ Multi-tenancy setup
✅ Version tracking

### Next Steps

1. Practice with your own use cases
2. Explore advanced features
3. Build custom integrations
4. Contribute to the project

---

[← Previous: Advanced Use Cases](23-advanced-use-cases.md) | [Back to Table of Contents](README.md) | [Next: Troubleshooting →](25-troubleshooting.md)

# Handler Event Integration Guide

This guide provides step-by-step instructions for integrating WebSocket event publishing into any handler in the HelixTrack Core application.

## Table of Contents

1. [Overview](#overview)
2. [Integration Pattern](#integration-pattern)
3. [Step-by-Step Instructions](#step-by-step-instructions)
4. [Complete Examples](#complete-examples)
5. [Context Types](#context-types)
6. [Testing](#testing)
7. [Checklist](#checklist)

## Overview

Every CREATE, MODIFY, and REMOVE operation in the application should publish an event to notify subscribed WebSocket clients about the change. This enables real-time updates across all connected clients.

**Key Principles:**
- Events are published AFTER successful database operations
- Events include sufficient context for permission-based filtering
- Events contain relevant entity data
- Failed operations do NOT publish events

## Integration Pattern

### Basic Pattern

```go
// 1. After successful database operation
// 2. Get username from context
username, _ := middleware.GetUsername(c)

// 3. Publish event
h.publisher.PublishEntityEvent(
    models.ActionCreate,  // or ActionModify, ActionRemove
    "entity_type",        // e.g., "ticket", "project", "comment"
    entityID,             // unique ID of the entity
    username,             // username performing the action
    map[string]interface{}{  // entity data
        "id": entityID,
        // ... other relevant fields
    },
    websocket.NewProjectContext(projectID, []string{"READ"}),  // event context
)
```

## Step-by-Step Instructions

### Step 1: Add Imports

Add these imports to your handler file:

```go
import (
    // ... existing imports
    "helixtrack.ru/core/internal/middleware"
    "helixtrack.ru/core/internal/websocket"
)
```

### Step 2: Integrate into CREATE Operation

**Location:** After successful INSERT, before sending response

**Pattern:**
```go
// After database INSERT operation succeeds
_, err := h.db.Exec(context.Background(), query, args...)
if err != nil {
    // Handle error
    return
}

// Get username from context
username, _ := middleware.GetUsername(c)

// Publish entity created event
h.publisher.PublishEntityEvent(
    models.ActionCreate,
    "entity_type",  // Replace with actual entity type
    entityID,
    username,
    map[string]interface{}{
        "id":    entityID,
        "field1": value1,
        "field2": value2,
        // Include all relevant fields
    },
    websocket.NewProjectContext(projectID, []string{"READ"}),
)

// Send response
response := models.NewSuccessResponse(...)
```

### Step 3: Integrate into MODIFY Operation

**Location:** After successful UPDATE, before sending response

**Pattern:**
```go
// After database UPDATE operation succeeds
_, err := h.db.Exec(context.Background(), query, args...)
if err != nil {
    // Handle error
    return
}

// Get username from context
username, _ := middleware.GetUsername(c)

// Get context info if not already available
var projectID string
err = h.db.QueryRow(context.Background(),
    "SELECT project_id FROM entity WHERE id = ?", entityID).Scan(&projectID)

// Publish entity updated event
if err == nil {
    h.publisher.PublishEntityEvent(
        models.ActionModify,
        "entity_type",
        entityID,
        username,
        entityData,  // Pass the modified data
        websocket.NewProjectContext(projectID, []string{"READ"}),
    )
}

// Send response
response := models.NewSuccessResponse(...)
```

### Step 4: Integrate into REMOVE Operation

**Location:** Before DELETE (to get context), after successful DELETE, before sending response

**Pattern:**
```go
// BEFORE delete: Get context information
var projectID string
err := h.db.QueryRow(context.Background(),
    "SELECT project_id FROM entity WHERE id = ? AND deleted = 0",
    entityID).Scan(&projectID)
if err != nil {
    c.JSON(http.StatusNotFound, models.NewErrorResponse(...))
    return
}

// Perform DELETE operation
query := "UPDATE entity SET deleted = 1, modified = ? WHERE id = ?"
_, err = h.db.Exec(context.Background(), query, time.Now().Unix(), entityID)
if err != nil {
    // Handle error
    return
}

// Get username from context
username, _ := middleware.GetUsername(c)

// Publish entity deleted event
h.publisher.PublishEntityEvent(
    models.ActionRemove,
    "entity_type",
    entityID,
    username,
    map[string]interface{}{
        "id":         entityID,
        "project_id": projectID,  // Include context for filtering
    },
    websocket.NewProjectContext(projectID, []string{"READ"}),
)

// Send response
response := models.NewSuccessResponse(...)
```

## Complete Examples

### Example 1: Ticket Handler

**File:** `internal/handlers/ticket_handler.go`

**CREATE Operation:**
```go
func (h *Handler) handleCreateTicket(c *gin.Context, req *models.Request) {
    // ... validation and data extraction ...

    // Create ticket in database
    ticketID := uuid.New().String()
    _, err := h.db.Exec(context.Background(), query, args...)
    if err != nil {
        // Handle error
        return
    }

    // Publish ticket created event
    username, _ := middleware.GetUsername(c)
    h.publisher.PublishEntityEvent(
        models.ActionCreate,
        "ticket",
        ticketID,
        username,
        map[string]interface{}{
            "id":            ticketID,
            "ticket_number": ticketNumber,
            "title":         title,
            "description":   description,
            "type":          ticketTypeStr,
            "priority":      priority,
            "status":        "open",
            "project_id":    projectID,
        },
        websocket.NewProjectContext(projectID, []string{"READ"}),
    )

    // Send response
    c.JSON(http.StatusOK, response)
}
```

**MODIFY Operation:**
```go
func (h *Handler) handleModifyTicket(c *gin.Context, req *models.Request) {
    // ... validation and data extraction ...

    // Update ticket in database
    _, err := h.db.Exec(context.Background(), query, args...)
    if err != nil {
        // Handle error
        return
    }

    // Get project_id for event context
    var projectID string
    err = h.db.QueryRow(context.Background(),
        "SELECT project_id FROM ticket WHERE id = ?", ticketID).Scan(&projectID)

    username, _ := middleware.GetUsername(c)

    // Publish ticket updated event
    if err == nil {
        h.publisher.PublishEntityEvent(
            models.ActionModify,
            "ticket",
            ticketID,
            username,
            ticketData,
            websocket.NewProjectContext(projectID, []string{"READ"}),
        )
    }

    // Send response
    c.JSON(http.StatusOK, response)
}
```

**REMOVE Operation:**
```go
func (h *Handler) handleRemoveTicket(c *gin.Context, req *models.Request) {
    // ... validation ...

    // Get project_id BEFORE deletion
    var projectID string
    err := h.db.QueryRow(context.Background(),
        "SELECT project_id FROM ticket WHERE id = ? AND deleted = 0",
        ticketID).Scan(&projectID)
    if err != nil {
        c.JSON(http.StatusNotFound, models.NewErrorResponse(...))
        return
    }

    // Delete ticket
    query := "UPDATE ticket SET deleted = 1, modified = ? WHERE id = ?"
    _, err = h.db.Exec(context.Background(), query, time.Now().Unix(), ticketID)
    if err != nil {
        // Handle error
        return
    }

    // Publish ticket deleted event
    username, _ := middleware.GetUsername(c)
    h.publisher.PublishEntityEvent(
        models.ActionRemove,
        "ticket",
        ticketID,
        username,
        map[string]interface{}{
            "id":         ticketID,
            "project_id": projectID,
        },
        websocket.NewProjectContext(projectID, []string{"READ"}),
    )

    // Send response
    c.JSON(http.StatusOK, response)
}
```

### Example 2: Project Handler

**File:** `internal/handlers/project_handler.go`

**CREATE Operation:**
```go
func (h *Handler) handleCreateProject(c *gin.Context, req *models.Request) {
    // ... validation and data extraction ...

    // Create project
    projectID := uuid.New().String()
    _, err := h.db.Exec(context.Background(), query, args...)
    if err != nil {
        // Handle error
        return
    }

    // Publish project created event
    username, _ := middleware.GetUsername(c)
    h.publisher.PublishEntityEvent(
        models.ActionCreate,
        "project",
        projectID,
        username,
        map[string]interface{}{
            "id":          projectID,
            "identifier":  key,
            "title":       name,
            "description": description,
            "type":        projectType,
        },
        websocket.NewProjectContext(projectID, []string{"READ"}),
    )

    // Send response
    c.JSON(http.StatusOK, response)
}
```

### Example 3: Comment Handler (Hierarchical Context)

**File:** `internal/handlers/comment_handler.go`

**CREATE Operation (requires parent context):**
```go
func (h *Handler) handleCreateComment(c *gin.Context, req *models.Request) {
    // ... validation and data extraction ...

    // Create comment and mapping
    commentID := uuid.New().String()
    _, err := h.db.Exec(context.Background(), query, args...)
    if err != nil {
        // Handle error
        return
    }

    // Get project_id from parent ticket for event context
    var projectID string
    h.db.QueryRow(context.Background(),
        "SELECT project_id FROM ticket WHERE id = ?", ticketID).Scan(&projectID)

    // Publish comment created event
    username, _ := middleware.GetUsername(c)
    if projectID != "" {
        h.publisher.PublishEntityEvent(
            models.ActionCreate,
            "comment",
            commentID,
            username,
            map[string]interface{}{
                "id":        commentID,
                "comment":   commentText,
                "ticket_id": ticketID,
            },
            websocket.NewProjectContext(projectID, []string{"READ"}),
        )
    }

    // Send response
    c.JSON(http.StatusOK, response)
}
```

## Context Types

Different entity types require different context for permission filtering:

### Project Context

Most entities belong to a project:

```go
websocket.NewProjectContext(projectID, []string{"READ"})
```

**Use for:** Tickets, Comments, Boards, Cycles, Components, Labels, etc.

### Organization Context

Organization-level entities:

```go
websocket.NewOrganizationContext(organizationID, []string{"READ"})
```

**Use for:** Organization management, organization-level settings

### Team Context

Team-level entities:

```go
websocket.NewTeamContext(teamID, organizationID, []string{"READ"})
```

**Use for:** Team management, team assignments

### Account Context

Account-level entities:

```go
websocket.NewAccountContext(accountID, []string{"READ"})
```

**Use for:** Multi-tenancy account management

### Full Context

When multiple context levels are relevant:

```go
websocket.NewFullContext(projectID, organizationID, teamID, accountID, []string{"READ"})
```

**Use for:** Complex hierarchical entities

## Event Types

Each entity type has corresponding event type constants defined in `models/event.go`:

```go
// Ticket events
EventTicketCreated
EventTicketUpdated
EventTicketDeleted

// Project events
EventProjectCreated
EventProjectUpdated
EventProjectDeleted

// Comment events
EventCommentCreated
EventCommentUpdated
EventCommentDeleted

// Priority events
EventPriorityCreated
EventPriorityUpdated
EventPriorityDeleted

// Resolution events
EventResolutionCreated
EventResolutionUpdated
EventResolutionDeleted

// Version events
EventVersionCreated
EventVersionUpdated
EventVersionDeleted
EventVersionReleased
EventVersionArchived

// ... and many more
```

The event type is automatically generated from the action and object parameters.

## Testing

### Unit Test Pattern

For each handler integration, add tests to verify event publishing:

```go
func TestHandlerCreateEntity_PublishesEvent(t *testing.T) {
    // Setup
    mockDB := &MockDatabase{}
    mockPublisher := &MockEventPublisher{}
    handler := NewHandler(mockDB, nil, nil, "1.0")
    handler.SetEventPublisher(mockPublisher)

    // Configure mocks
    mockDB.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
    mockPublisher.On("PublishEntityEvent",
        models.ActionCreate,
        "entity",
        mock.MatchedBy(func(id string) bool { return id != "" }),
        "testuser",
        mock.Anything,
        mock.Anything,
    ).Return()

    // Execute
    c, _ := gin.CreateTestContext(httptest.NewRecorder())
    c.Set("username", "testuser")
    req := &models.Request{
        Action: models.ActionCreate,
        Object: "entity",
        Data: map[string]interface{}{
            "field": "value",
        },
    }

    handler.handleCreateEntity(c, req)

    // Verify
    mockPublisher.AssertExpectations(t)
    mockPublisher.AssertCalled(t, "PublishEntityEvent",
        models.ActionCreate, "entity", mock.Anything, "testuser", mock.Anything, mock.Anything)
}
```

### Integration Test Pattern

Test end-to-end with real WebSocket connections:

```go
func TestEntityCreate_WebSocketNotification(t *testing.T) {
    // 1. Start test server with WebSocket enabled
    // 2. Connect WebSocket client
    // 3. Subscribe to entity events
    // 4. Create entity via HTTP API
    // 5. Verify event received via WebSocket
    // 6. Verify event data matches created entity
}
```

See `test-scripts/WEBSOCKET_TESTING_README.md` for full testing guide.

## Checklist

Use this checklist when integrating event publishing into a handler:

### Preparation
- [ ] Read this guide completely
- [ ] Identify the handler file to modify
- [ ] Identify all CREATE, MODIFY, REMOVE operations
- [ ] Determine appropriate context type (project, organization, team, account)

### Implementation
- [ ] Add required imports (`middleware`, `websocket`)
- [ ] Integrate event publishing into CREATE operation
  - [ ] After successful database insert
  - [ ] Get username from context
  - [ ] Include all relevant entity data
  - [ ] Use appropriate context
- [ ] Integrate event publishing into MODIFY operation
  - [ ] After successful database update
  - [ ] Get username from context
  - [ ] Get context info if needed
  - [ ] Include modified data
- [ ] Integrate event publishing into REMOVE operation
  - [ ] Get context info BEFORE deletion
  - [ ] After successful database delete
  - [ ] Get username from context
  - [ ] Include entity ID and context

### Testing
- [ ] Write unit test for CREATE event publishing
- [ ] Write unit test for MODIFY event publishing
- [ ] Write unit test for REMOVE event publishing
- [ ] Test error cases (no event on failure)
- [ ] Test with WebSocket client
- [ ] Verify event data is correct
- [ ] Verify event filtering works (permissions, subscriptions)

### Documentation
- [ ] Update handler-specific documentation if needed
- [ ] Add comments explaining context selection
- [ ] Document any special considerations

## Common Patterns

### Pattern 1: Direct Project Association

**Entities:** Ticket, Board, Cycle, Component, Label, Repository

```go
// Entity has direct project_id field
websocket.NewProjectContext(entity.ProjectID, []string{"READ"})
```

### Pattern 2: Hierarchical Context

**Entities:** Comment (via Ticket), Asset (via multiple parents)

```go
// Get parent's project_id
var projectID string
h.db.QueryRow(context.Background(),
    "SELECT project_id FROM parent WHERE id = ?", parentID).Scan(&projectID)

if projectID != "" {
    h.publisher.PublishEntityEvent(..., websocket.NewProjectContext(projectID, []string{"READ"}))
}
```

### Pattern 3: Global/System Entities

**Entities:** TicketType, TicketStatus, WorkflowStep, Priority, Resolution

```go
// System-wide entities - no specific project context
// Use empty context or organization context if applicable
websocket.NewProjectContext("", []string{"READ"})
```

### Pattern 4: Multi-Tenant Entities

**Entities:** Account, Organization, Team

```go
// Use appropriate hierarchy level
websocket.NewOrganizationContext(organizationID, []string{"READ"})
websocket.NewTeamContext(teamID, organizationID, []string{"READ"})
websocket.NewAccountContext(accountID, []string{"READ"})
```

## Error Handling

### Best Practices

1. **Never publish on failure:**
   ```go
   _, err := h.db.Exec(...)
   if err != nil {
       // Do NOT publish event
       return
   }
   // Publish event only after success
   h.publisher.PublishEntityEvent(...)
   ```

2. **Handle missing context gracefully:**
   ```go
   var projectID string
   err := h.db.QueryRow(...).Scan(&projectID)

   // Only publish if context found
   if err == nil && projectID != "" {
       h.publisher.PublishEntityEvent(...)
   }
   ```

3. **Don't fail request if event publishing fails:**
   ```go
   // Publishing is best-effort
   h.publisher.PublishEntityEvent(...)

   // Continue with response even if publishing failed
   c.JSON(http.StatusOK, response)
   ```

## Next Steps

After integrating event publishing into a handler:

1. Run unit tests: `go test ./internal/handlers/...`
2. Test with WebSocket client: Use `test-scripts/websocket-client.html`
3. Verify events in real-time during CRUD operations
4. Check event data completeness and accuracy
5. Test permission-based filtering
6. Update integration documentation

## Related Documentation

- **Event System Overview:** `WEBSOCKET_IMPLEMENTATION_SUMMARY.md`
- **Testing Guide:** `test-scripts/WEBSOCKET_TESTING_README.md`
- **Event Integration Pattern:** `internal/handlers/EVENT_INTEGRATION_PATTERN.md`
- **User Manual:** `docs/USER_MANUAL.md`

## Support

For questions or issues:
1. Review completed integrations: `ticket_handler.go`, `project_handler.go`, `comment_handler.go`
2. Check test examples: `*_handler_test.go` files
3. Test with interactive client: `test-scripts/websocket-client.html`
4. Review event models: `internal/models/event.go`

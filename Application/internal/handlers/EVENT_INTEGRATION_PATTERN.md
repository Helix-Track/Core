# WebSocket Event Integration Pattern

This document provides a comprehensive guide for integrating WebSocket event publishing into all handlers.

## Pattern Overview

Every handler that performs CRUD operations (Create, Read, Update, Delete) should publish events to notify WebSocket clients in real-time.

## Integration Steps

### 1. The Handler Has Event Publisher

The handler already has an event publisher field that is set when the server starts:

```go
type Handler struct {
	db          database.Database
	authService services.AuthService
	permService services.PermissionService
	version     string
	publisher   websocket.EventPublisher  // Event publisher for WebSocket
}
```

### 2. Publish Events After Successful Operations

#### Pattern for CREATE Operations

```go
func (h *Handler) handleCreateTicket(c *gin.Context, req *models.Request) {
	// ... existing validation and permission checks ...

	// Create the ticket in database
	ticketID, err := h.db.CreateTicket(/* params */)
	if err != nil {
		// ... handle error ...
		return
	}

	// SUCCESS - Publish event
	h.publisher.PublishEntityEvent(
		models.ActionCreate,
		"ticket",
		ticketID,
		username,
		map[string]interface{}{
			"title":       title,
			"description": description,
			"projectId":   projectID,
			// ... other relevant data ...
		},
		websocket.NewProjectContext(projectID, []string{"READ"}),
	)

	// Return success response
	c.JSON(http.StatusOK, models.NewSuccessResponse(map[string]interface{}{
		"id": ticketID,
	}))
}
```

#### Pattern for MODIFY Operations

```go
func (h *Handler) handleModifyTicket(c *gin.Context, req *models.Request) {
	// ... existing validation and permission checks ...

	// Modify the ticket in database
	err := h.db.UpdateTicket(ticketID, /* params */)
	if err != nil {
		// ... handle error ...
		return
	}

	// SUCCESS - Publish event
	h.publisher.PublishEntityEvent(
		models.ActionModify,
		"ticket",
		ticketID,
		username,
		map[string]interface{}{
			"id":       ticketID,
			"changes":  req.Data,  // What was changed
			"projectId": projectID,
		},
		websocket.NewProjectContext(projectID, []string{"READ"}),
	)

	// Return success response
	c.JSON(http.StatusOK, models.NewSuccessResponse(map[string]interface{}{
		"success": true,
	}))
}
```

#### Pattern for REMOVE Operations

```go
func (h *Handler) handleRemoveTicket(c *gin.Context, req *models.Request) {
	// ... existing validation and permission checks ...

	// Get ticket details before deletion (if needed for event)
	ticket, err := h.db.GetTicket(ticketID)
	if err != nil {
		// ... handle error ...
		return
	}

	// Delete the ticket from database
	err = h.db.DeleteTicket(ticketID)
	if err != nil {
		// ... handle error ...
		return
	}

	// SUCCESS - Publish event
	h.publisher.PublishEntityEvent(
		models.ActionRemove,
		"ticket",
		ticketID,
		username,
		map[string]interface{}{
			"id":        ticketID,
			"title":     ticket.Title,
			"projectId": ticket.ProjectID,
		},
		websocket.NewProjectContext(ticket.ProjectID, []string{"READ"}),
	)

	// Return success response
	c.JSON(http.StatusOK, models.NewSuccessResponse(map[string]interface{}{
		"success": true,
	}))
}
```

#### Pattern for READ Operations (Optional)

Read operations can also publish events if you want to track who's viewing what:

```go
func (h *Handler) handleReadTicket(c *gin.Context, req *models.Request) {
	// ... existing validation and permission checks ...

	// Get the ticket from database
	ticket, err := h.db.GetTicket(ticketID)
	if err != nil {
		// ... handle error ...
		return
	}

	// OPTIONAL - Publish read event (if tracking is needed)
	h.publisher.PublishEntityEvent(
		models.ActionRead,
		"ticket",
		ticketID,
		username,
		map[string]interface{}{
			"id":        ticketID,
			"projectId": ticket.ProjectID,
		},
		websocket.NewProjectContext(ticket.ProjectID, []string{"READ"}),
	)

	// Return success response
	c.JSON(http.StatusOK, models.NewSuccessResponse(map[string]interface{}{
		"ticket": ticket,
	}))
}
```

## Event Context Helpers

The `websocket` package provides context helpers for different entity types:

### Project Context
```go
websocket.NewProjectContext(projectID, []string{"READ"})
```

### Organization Context
```go
websocket.NewOrganizationContext(organizationID, []string{"READ"})
```

### Team Context
```go
websocket.NewTeamContext(teamID, organizationID, []string{"READ"})
```

### Account Context
```go
websocket.NewAccountContext(accountID, []string{"READ"})
```

### Full Context
```go
websocket.NewFullContext(projectID, organizationID, teamID, accountID, []string{"READ"})
```

## Event Data Guidelines

### What to Include in Event Data

**DO include:**
- Entity ID
- Entity type
- Key fields that changed
- Relevant context (projectId, organizationId, etc.)
- Timestamp is automatically included

**DON'T include:**
- Sensitive data (passwords, tokens, etc.)
- Large binary data
- Full entity dumps (unless necessary)

### Example Event Data

```go
// Good - includes relevant information
map[string]interface{}{
	"id":          ticketID,
	"title":       "Updated title",
	"status":      "in_progress",
	"assignee":    "john.doe",
	"projectId":   projectID,
}

// Bad - too much data
map[string]interface{}{
	"fullTicketObject": ticket,  // Too large
	"allComments":      comments, // Unnecessary
}

// Bad - sensitive data
map[string]interface{}{
	"userPassword": password,  // NEVER include passwords
	"apiToken":     token,     // NEVER include tokens
}
```

## Permission-Based Filtering

Events are automatically filtered based on:

1. **Client subscription** - Clients subscribe to specific event types
2. **Entity filters** - Clients can filter by entity type, ID, project, etc.
3. **Permission checks** - Events specify required permissions in context
4. **Permission service** - If enabled, checks if user has access

### Setting Required Permissions

```go
// Only users with READ permission will receive this event
websocket.NewProjectContext(projectID, []string{"READ"})

// Only users with UPDATE permission will receive this event
websocket.NewProjectContext(projectID, []string{"UPDATE"})

// Multiple permission options
websocket.NewProjectContext(projectID, []string{"READ", "UPDATE"})
```

## Complete Example

Here's a complete example for a ticket creation handler:

```go
func (h *Handler) handleCreateTicket(c *gin.Context, req *models.Request) {
	// Get username from context
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Extract parameters
	title, _ := req.Data["title"].(string)
	description, _ := req.Data["description"].(string)
	projectID, _ := req.Data["projectId"].(string)

	// Validate parameters
	if title == "" || projectID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing required fields",
			"",
		))
		return
	}

	// Create ticket in database
	ticketID, err := h.db.CreateTicket(title, description, projectID, username)
	if err != nil {
		logger.Error("Failed to create ticket", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create ticket",
			"",
		))
		return
	}

	// SUCCESS - Publish event to WebSocket clients
	h.publisher.PublishEntityEvent(
		models.ActionCreate,           // Action: create
		"ticket",                       // Object type: ticket
		ticketID,                       // Entity ID
		username,                       // Username who created it
		map[string]interface{}{         // Event data
			"id":          ticketID,
			"title":       title,
			"description": description,
			"projectId":   projectID,
			"creator":     username,
		},
		websocket.NewProjectContext(    // Event context
			projectID,                  // Project ID for filtering
			[]string{"READ"},           // Required permission
		),
	)

	// Return success response
	c.JSON(http.StatusOK, models.NewSuccessResponse(map[string]interface{}{
		"id":      ticketID,
		"message": "Ticket created successfully",
	}))
}
```

## Integration Checklist

For each handler that performs CRUD operations:

- [ ] Identify all CREATE, MODIFY, REMOVE operations
- [ ] Add event publishing AFTER successful database operation
- [ ] Use appropriate action constant from models package
- [ ] Include relevant data in event (not sensitive data)
- [ ] Set appropriate event context with required permissions
- [ ] Test event delivery with WebSocket clients
- [ ] Verify permission filtering works correctly
- [ ] Document any special event behavior

## Testing Event Publishing

See test files for examples:
- `internal/handlers/handler_test.go` - Handler tests with event mocking
- `internal/websocket/manager_test.go` - WebSocket manager tests
- `test-scripts/ws-test.sh` - Integration tests for WebSocket events

## Next Steps

1. Apply this pattern to ALL handlers in the codebase
2. Prioritize high-traffic operations (tickets, projects, comments)
3. Add unit tests for event publishing
4. Create integration tests for end-to-end event delivery
5. Document any custom event types or patterns

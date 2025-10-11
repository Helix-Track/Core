package models

import (
	"time"
)

// EventType represents the type of event
type EventType string

// Event type constants
const (
	// Entity lifecycle events
	EventEntityCreated EventType = "entity.created"
	EventEntityUpdated EventType = "entity.updated"
	EventEntityDeleted EventType = "entity.deleted"
	EventEntityRead    EventType = "entity.read"

	// Ticket events
	EventTicketCreated EventType = "ticket.created"
	EventTicketUpdated EventType = "ticket.updated"
	EventTicketDeleted EventType = "ticket.deleted"
	EventTicketRead    EventType = "ticket.read"

	// Project events
	EventProjectCreated EventType = "project.created"
	EventProjectUpdated EventType = "project.updated"
	EventProjectDeleted EventType = "project.deleted"
	EventProjectRead    EventType = "project.read"

	// Comment events
	EventCommentCreated EventType = "comment.created"
	EventCommentUpdated EventType = "comment.updated"
	EventCommentDeleted EventType = "comment.deleted"
	EventCommentRead    EventType = "comment.read"

	// Priority events
	EventPriorityCreated EventType = "priority.created"
	EventPriorityUpdated EventType = "priority.updated"
	EventPriorityDeleted EventType = "priority.deleted"
	EventPriorityRead    EventType = "priority.read"

	// Resolution events
	EventResolutionCreated EventType = "resolution.created"
	EventResolutionUpdated EventType = "resolution.updated"
	EventResolutionDeleted EventType = "resolution.deleted"
	EventResolutionRead    EventType = "resolution.read"

	// Version events
	EventVersionCreated  EventType = "version.created"
	EventVersionUpdated  EventType = "version.updated"
	EventVersionDeleted  EventType = "version.deleted"
	EventVersionRead     EventType = "version.read"
	EventVersionReleased EventType = "version.released"
	EventVersionArchived EventType = "version.archived"

	// Watcher events
	EventWatcherAdded   EventType = "watcher.added"
	EventWatcherRemoved EventType = "watcher.removed"

	// Filter events
	EventFilterSaved   EventType = "filter.saved"
	EventFilterUpdated EventType = "filter.updated"
	EventFilterDeleted EventType = "filter.deleted"
	EventFilterShared  EventType = "filter.shared"

	// Custom field events
	EventCustomFieldCreated EventType = "customfield.created"
	EventCustomFieldUpdated EventType = "customfield.updated"
	EventCustomFieldDeleted EventType = "customfield.deleted"
	EventCustomFieldRead    EventType = "customfield.read"

	// Board events
	EventBoardCreated EventType = "board.created"
	EventBoardUpdated EventType = "board.updated"
	EventBoardDeleted EventType = "board.deleted"

	// Cycle events (Sprint/Milestone/Release)
	EventCycleCreated EventType = "cycle.created"
	EventCycleUpdated EventType = "cycle.updated"
	EventCycleDeleted EventType = "cycle.deleted"

	// Workflow events
	EventWorkflowCreated EventType = "workflow.created"
	EventWorkflowUpdated EventType = "workflow.updated"
	EventWorkflowDeleted EventType = "workflow.deleted"

	// Account events
	EventAccountCreated EventType = "account.created"
	EventAccountUpdated EventType = "account.updated"
	EventAccountDeleted EventType = "account.deleted"

	// Organization events
	EventOrganizationCreated EventType = "organization.created"
	EventOrganizationUpdated EventType = "organization.updated"
	EventOrganizationDeleted EventType = "organization.deleted"

	// Team events
	EventTeamCreated EventType = "team.created"
	EventTeamUpdated EventType = "team.updated"
	EventTeamDeleted EventType = "team.deleted"

	// User events
	EventUserCreated EventType = "user.created"
	EventUserUpdated EventType = "user.updated"
	EventUserDeleted EventType = "user.deleted"

	// System events
	EventSystemHealthCheck EventType = "system.health_check"
	EventSystemError       EventType = "system.error"
	EventSystemShutdown    EventType = "system.shutdown"

	// Connection events
	EventConnectionEstablished EventType = "connection.established"
	EventConnectionClosed      EventType = "connection.closed"
	EventConnectionError       EventType = "connection.error"
)

// Event represents a system event to be broadcasted via WebSocket
type Event struct {
	ID        string                 `json:"id"`        // Unique event ID (UUID)
	Type      EventType              `json:"type"`      // Event type (e.g., "ticket.created")
	Action    string                 `json:"action"`    // Action that triggered the event (e.g., "create", "modify")
	Object    string                 `json:"object"`    // Object type (e.g., "ticket", "project")
	EntityID  string                 `json:"entityId"`  // ID of the entity affected
	Username  string                 `json:"username"`  // Username who triggered the event
	Timestamp time.Time              `json:"timestamp"` // When the event occurred
	Data      map[string]interface{} `json:"data"`      // Additional event data
	Context   EventContext           `json:"context"`   // Event context (permissions, etc.)
}

// EventContext contains contextual information about the event
type EventContext struct {
	ProjectID      string   `json:"projectId,omitempty"`      // Project ID if relevant
	OrganizationID string   `json:"organizationId,omitempty"` // Organization ID if relevant
	TeamID         string   `json:"teamId,omitempty"`         // Team ID if relevant
	AccountID      string   `json:"accountId,omitempty"`      // Account ID if relevant
	Permissions    []string `json:"permissions,omitempty"`    // Required permissions to see this event
}

// Subscription represents a client's subscription to specific event types
type Subscription struct {
	EventTypes     []EventType          `json:"eventTypes"`     // Event types to subscribe to
	EntityTypes    []string             `json:"entityTypes"`    // Entity types to filter (ticket, project, etc.)
	EntityIDs      []string             `json:"entityIds"`      // Specific entity IDs to filter
	Filters        map[string]string    `json:"filters"`        // Additional filters (projectId, teamId, etc.)
	IncludeReads   bool                 `json:"includeReads"`   // Whether to include read events
	CustomFilters  map[string][]string  `json:"customFilters"`  // Custom filter criteria
	PermissionMask int                  `json:"permissionMask"` // Required permission level
}

// NewEvent creates a new event with the given parameters
func NewEvent(eventType EventType, action, object, entityID, username string, data map[string]interface{}) *Event {
	return &Event{
		ID:        generateEventID(),
		Type:      eventType,
		Action:    action,
		Object:    object,
		EntityID:  entityID,
		Username:  username,
		Timestamp: time.Now().UTC(),
		Data:      data,
		Context:   EventContext{},
	}
}

// WithContext adds context to the event
func (e *Event) WithContext(context EventContext) *Event {
	e.Context = context
	return e
}

// MatchesSubscription checks if the event matches a subscription
func (e *Event) MatchesSubscription(sub *Subscription) bool {
	// Check event type
	if len(sub.EventTypes) > 0 {
		matched := false
		for _, et := range sub.EventTypes {
			if e.Type == et {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Check entity type
	if len(sub.EntityTypes) > 0 {
		matched := false
		for _, et := range sub.EntityTypes {
			if e.Object == et {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Check entity ID
	if len(sub.EntityIDs) > 0 {
		matched := false
		for _, eid := range sub.EntityIDs {
			if e.EntityID == eid {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Check if read events should be included
	if !sub.IncludeReads && isReadEvent(e.Type) {
		return false
	}

	// Check filters (projectId, teamId, etc.)
	if len(sub.Filters) > 0 {
		for key, value := range sub.Filters {
			switch key {
			case "projectId":
				if e.Context.ProjectID != value {
					return false
				}
			case "organizationId":
				if e.Context.OrganizationID != value {
					return false
				}
			case "teamId":
				if e.Context.TeamID != value {
					return false
				}
			case "accountId":
				if e.Context.AccountID != value {
					return false
				}
			}
		}
	}

	return true
}

// generateEventID generates a unique event ID
func generateEventID() string {
	// Use UUID generation from existing codebase
	// This is a placeholder - will use github.com/google/uuid
	return time.Now().Format("20060102150405.000000")
}

// isReadEvent checks if the event type is a read event
func isReadEvent(eventType EventType) bool {
	return eventType == EventEntityRead ||
		eventType == EventTicketRead ||
		eventType == EventProjectRead ||
		eventType == EventCommentRead ||
		eventType == EventPriorityRead ||
		eventType == EventResolutionRead ||
		eventType == EventVersionRead ||
		eventType == EventCustomFieldRead
}

// GetEventTypeFromAction returns the appropriate event type for an action and object
func GetEventTypeFromAction(action, object string) EventType {
	switch action {
	case ActionCreate:
		return getCreateEventType(object)
	case ActionModify:
		return getUpdateEventType(object)
	case ActionRemove:
		return getDeleteEventType(object)
	case ActionRead:
		return getReadEventType(object)
	default:
		return EventEntityCreated
	}
}

func getCreateEventType(object string) EventType {
	switch object {
	case "ticket":
		return EventTicketCreated
	case "project":
		return EventProjectCreated
	case "comment":
		return EventCommentCreated
	case "priority":
		return EventPriorityCreated
	case "resolution":
		return EventResolutionCreated
	case "version":
		return EventVersionCreated
	case "filter":
		return EventFilterSaved
	case "customfield":
		return EventCustomFieldCreated
	case "board":
		return EventBoardCreated
	case "cycle":
		return EventCycleCreated
	case "workflow":
		return EventWorkflowCreated
	case "account":
		return EventAccountCreated
	case "organization":
		return EventOrganizationCreated
	case "team":
		return EventTeamCreated
	case "user":
		return EventUserCreated
	default:
		return EventEntityCreated
	}
}

func getUpdateEventType(object string) EventType {
	switch object {
	case "ticket":
		return EventTicketUpdated
	case "project":
		return EventProjectUpdated
	case "comment":
		return EventCommentUpdated
	case "priority":
		return EventPriorityUpdated
	case "resolution":
		return EventResolutionUpdated
	case "version":
		return EventVersionUpdated
	case "filter":
		return EventFilterUpdated
	case "customfield":
		return EventCustomFieldUpdated
	case "board":
		return EventBoardUpdated
	case "cycle":
		return EventCycleUpdated
	case "workflow":
		return EventWorkflowUpdated
	case "account":
		return EventAccountUpdated
	case "organization":
		return EventOrganizationUpdated
	case "team":
		return EventTeamUpdated
	case "user":
		return EventUserUpdated
	default:
		return EventEntityUpdated
	}
}

func getDeleteEventType(object string) EventType {
	switch object {
	case "ticket":
		return EventTicketDeleted
	case "project":
		return EventProjectDeleted
	case "comment":
		return EventCommentDeleted
	case "priority":
		return EventPriorityDeleted
	case "resolution":
		return EventResolutionDeleted
	case "version":
		return EventVersionDeleted
	case "filter":
		return EventFilterDeleted
	case "customfield":
		return EventCustomFieldDeleted
	case "board":
		return EventBoardDeleted
	case "cycle":
		return EventCycleDeleted
	case "workflow":
		return EventWorkflowDeleted
	case "account":
		return EventAccountDeleted
	case "organization":
		return EventOrganizationDeleted
	case "team":
		return EventTeamDeleted
	case "user":
		return EventUserDeleted
	default:
		return EventEntityDeleted
	}
}

func getReadEventType(object string) EventType {
	switch object {
	case "ticket":
		return EventTicketRead
	case "project":
		return EventProjectRead
	case "comment":
		return EventCommentRead
	case "priority":
		return EventPriorityRead
	case "resolution":
		return EventResolutionRead
	case "version":
		return EventVersionRead
	case "customfield":
		return EventCustomFieldRead
	default:
		return EventEntityRead
	}
}

package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewEvent(t *testing.T) {
	tests := []struct {
		name       string
		eventType  EventType
		action     string
		object     string
		entityID   string
		username   string
		data       map[string]interface{}
		wantAction string
		wantObject string
	}{
		{
			name:       "Create ticket event",
			eventType:  EventTicketCreated,
			action:     ActionCreate,
			object:     "ticket",
			entityID:   "ticket-123",
			username:   "john.doe",
			data:       map[string]interface{}{"title": "Test Ticket"},
			wantAction: ActionCreate,
			wantObject: "ticket",
		},
		{
			name:       "Update project event",
			eventType:  EventProjectUpdated,
			action:     ActionModify,
			object:     "project",
			entityID:   "project-456",
			username:   "jane.doe",
			data:       map[string]interface{}{"name": "Updated Project"},
			wantAction: ActionModify,
			wantObject: "project",
		},
		{
			name:       "Delete comment event",
			eventType:  EventCommentDeleted,
			action:     ActionRemove,
			object:     "comment",
			entityID:   "comment-789",
			username:   "admin",
			data:       map[string]interface{}{},
			wantAction: ActionRemove,
			wantObject: "comment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := NewEvent(tt.eventType, tt.action, tt.object, tt.entityID, tt.username, tt.data)

			assert.NotEmpty(t, event.ID, "Event ID should not be empty")
			assert.Equal(t, tt.eventType, event.Type)
			assert.Equal(t, tt.wantAction, event.Action)
			assert.Equal(t, tt.wantObject, event.Object)
			assert.Equal(t, tt.entityID, event.EntityID)
			assert.Equal(t, tt.username, event.Username)
			assert.Equal(t, tt.data, event.Data)
			assert.WithinDuration(t, time.Now(), event.Timestamp, 1*time.Second)
			assert.NotNil(t, event.Context)
		})
	}
}

func TestEvent_WithContext(t *testing.T) {
	event := NewEvent(EventTicketCreated, ActionCreate, "ticket", "ticket-123", "john.doe", nil)
	context := EventContext{
		ProjectID:      "project-123",
		OrganizationID: "org-456",
		TeamID:         "team-789",
		Permissions:    []string{"READ", "WRITE"},
	}

	result := event.WithContext(context)

	assert.Equal(t, event, result, "WithContext should return the same event")
	assert.Equal(t, "project-123", event.Context.ProjectID)
	assert.Equal(t, "org-456", event.Context.OrganizationID)
	assert.Equal(t, "team-789", event.Context.TeamID)
	assert.Equal(t, []string{"READ", "WRITE"}, event.Context.Permissions)
}

func TestEvent_MatchesSubscription(t *testing.T) {
	tests := []struct {
		name         string
		event        *Event
		subscription *Subscription
		want         bool
	}{
		{
			name: "Match by event type",
			event: &Event{
				Type:     EventTicketCreated,
				Object:   "ticket",
				EntityID: "ticket-123",
			},
			subscription: &Subscription{
				EventTypes: []EventType{EventTicketCreated, EventTicketUpdated},
			},
			want: true,
		},
		{
			name: "No match by event type",
			event: &Event{
				Type:     EventProjectCreated,
				Object:   "project",
				EntityID: "project-123",
			},
			subscription: &Subscription{
				EventTypes: []EventType{EventTicketCreated, EventTicketUpdated},
			},
			want: false,
		},
		{
			name: "Match by entity type",
			event: &Event{
				Type:     EventTicketCreated,
				Object:   "ticket",
				EntityID: "ticket-123",
			},
			subscription: &Subscription{
				EntityTypes: []string{"ticket", "project"},
			},
			want: true,
		},
		{
			name: "No match by entity type",
			event: &Event{
				Type:     EventCommentCreated,
				Object:   "comment",
				EntityID: "comment-123",
			},
			subscription: &Subscription{
				EntityTypes: []string{"ticket", "project"},
			},
			want: false,
		},
		{
			name: "Match by entity ID",
			event: &Event{
				Type:     EventTicketUpdated,
				Object:   "ticket",
				EntityID: "ticket-123",
			},
			subscription: &Subscription{
				EntityIDs: []string{"ticket-123", "ticket-456"},
			},
			want: true,
		},
		{
			name: "No match by entity ID",
			event: &Event{
				Type:     EventTicketUpdated,
				Object:   "ticket",
				EntityID: "ticket-789",
			},
			subscription: &Subscription{
				EntityIDs: []string{"ticket-123", "ticket-456"},
			},
			want: false,
		},
		{
			name: "Filter read events when includeReads is false",
			event: &Event{
				Type:     EventTicketRead,
				Object:   "ticket",
				EntityID: "ticket-123",
			},
			subscription: &Subscription{
				IncludeReads: false,
			},
			want: false,
		},
		{
			name: "Include read events when includeReads is true",
			event: &Event{
				Type:     EventTicketRead,
				Object:   "ticket",
				EntityID: "ticket-123",
			},
			subscription: &Subscription{
				IncludeReads: true,
			},
			want: true,
		},
		{
			name: "Match by project filter",
			event: &Event{
				Type:     EventTicketCreated,
				Object:   "ticket",
				EntityID: "ticket-123",
				Context: EventContext{
					ProjectID: "project-456",
				},
			},
			subscription: &Subscription{
				Filters: map[string]string{
					"projectId": "project-456",
				},
			},
			want: true,
		},
		{
			name: "No match by project filter",
			event: &Event{
				Type:     EventTicketCreated,
				Object:   "ticket",
				EntityID: "ticket-123",
				Context: EventContext{
					ProjectID: "project-789",
				},
			},
			subscription: &Subscription{
				Filters: map[string]string{
					"projectId": "project-456",
				},
			},
			want: false,
		},
		{
			name: "Match by organization filter",
			event: &Event{
				Type:     EventTicketCreated,
				Object:   "ticket",
				EntityID: "ticket-123",
				Context: EventContext{
					OrganizationID: "org-123",
				},
			},
			subscription: &Subscription{
				Filters: map[string]string{
					"organizationId": "org-123",
				},
			},
			want: true,
		},
		{
			name: "Match by team filter",
			event: &Event{
				Type:     EventTicketCreated,
				Object:   "ticket",
				EntityID: "ticket-123",
				Context: EventContext{
					TeamID: "team-456",
				},
			},
			subscription: &Subscription{
				Filters: map[string]string{
					"teamId": "team-456",
				},
			},
			want: true,
		},
		{
			name: "Match by account filter",
			event: &Event{
				Type:     EventTicketCreated,
				Object:   "ticket",
				EntityID: "ticket-123",
				Context: EventContext{
					AccountID: "account-789",
				},
			},
			subscription: &Subscription{
				Filters: map[string]string{
					"accountId": "account-789",
				},
			},
			want: true,
		},
		{
			name: "Match all criteria",
			event: &Event{
				Type:     EventTicketUpdated,
				Object:   "ticket",
				EntityID: "ticket-123",
				Context: EventContext{
					ProjectID: "project-456",
				},
			},
			subscription: &Subscription{
				EventTypes:  []EventType{EventTicketUpdated},
				EntityTypes: []string{"ticket"},
				EntityIDs:   []string{"ticket-123"},
				Filters: map[string]string{
					"projectId": "project-456",
				},
			},
			want: true,
		},
		{
			name: "Empty subscription matches all",
			event: &Event{
				Type:     EventTicketCreated,
				Object:   "ticket",
				EntityID: "ticket-123",
			},
			subscription: &Subscription{},
			want:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.event.MatchesSubscription(tt.subscription)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetEventTypeFromAction(t *testing.T) {
	tests := []struct {
		name   string
		action string
		object string
		want   EventType
	}{
		{
			name:   "Create ticket",
			action: ActionCreate,
			object: "ticket",
			want:   EventTicketCreated,
		},
		{
			name:   "Modify ticket",
			action: ActionModify,
			object: "ticket",
			want:   EventTicketUpdated,
		},
		{
			name:   "Remove ticket",
			action: ActionRemove,
			object: "ticket",
			want:   EventTicketDeleted,
		},
		{
			name:   "Read ticket",
			action: ActionRead,
			object: "ticket",
			want:   EventTicketRead,
		},
		{
			name:   "Create project",
			action: ActionCreate,
			object: "project",
			want:   EventProjectCreated,
		},
		{
			name:   "Modify project",
			action: ActionModify,
			object: "project",
			want:   EventProjectUpdated,
		},
		{
			name:   "Remove project",
			action: ActionRemove,
			object: "project",
			want:   EventProjectDeleted,
		},
		{
			name:   "Create comment",
			action: ActionCreate,
			object: "comment",
			want:   EventCommentCreated,
		},
		{
			name:   "Create priority",
			action: ActionCreate,
			object: "priority",
			want:   EventPriorityCreated,
		},
		{
			name:   "Modify priority",
			action: ActionModify,
			object: "priority",
			want:   EventPriorityUpdated,
		},
		{
			name:   "Create resolution",
			action: ActionCreate,
			object: "resolution",
			want:   EventResolutionCreated,
		},
		{
			name:   "Create version",
			action: ActionCreate,
			object: "version",
			want:   EventVersionCreated,
		},
		{
			name:   "Create filter",
			action: ActionCreate,
			object: "filter",
			want:   EventFilterSaved,
		},
		{
			name:   "Create custom field",
			action: ActionCreate,
			object: "customfield",
			want:   EventCustomFieldCreated,
		},
		{
			name:   "Create board",
			action: ActionCreate,
			object: "board",
			want:   EventBoardCreated,
		},
		{
			name:   "Create cycle",
			action: ActionCreate,
			object: "cycle",
			want:   EventCycleCreated,
		},
		{
			name:   "Create workflow",
			action: ActionCreate,
			object: "workflow",
			want:   EventWorkflowCreated,
		},
		{
			name:   "Create account",
			action: ActionCreate,
			object: "account",
			want:   EventAccountCreated,
		},
		{
			name:   "Create organization",
			action: ActionCreate,
			object: "organization",
			want:   EventOrganizationCreated,
		},
		{
			name:   "Create team",
			action: ActionCreate,
			object: "team",
			want:   EventTeamCreated,
		},
		{
			name:   "Create user",
			action: ActionCreate,
			object: "user",
			want:   EventUserCreated,
		},
		{
			name:   "Unknown object defaults to entity created",
			action: ActionCreate,
			object: "unknown",
			want:   EventEntityCreated,
		},
		{
			name:   "Unknown action defaults to entity created",
			action: "unknown",
			object: "ticket",
			want:   EventEntityCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetEventTypeFromAction(tt.action, tt.object)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsReadEvent(t *testing.T) {
	tests := []struct {
		name      string
		eventType EventType
		want      bool
	}{
		{"Entity read", EventEntityRead, true},
		{"Ticket read", EventTicketRead, true},
		{"Project read", EventProjectRead, true},
		{"Comment read", EventCommentRead, true},
		{"Priority read", EventPriorityRead, true},
		{"Resolution read", EventResolutionRead, true},
		{"Version read", EventVersionRead, true},
		{"Custom field read", EventCustomFieldRead, true},
		{"Ticket created", EventTicketCreated, false},
		{"Ticket updated", EventTicketUpdated, false},
		{"Ticket deleted", EventTicketDeleted, false},
		{"Project created", EventProjectCreated, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isReadEvent(tt.eventType)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEventTypes(t *testing.T) {
	// Test that all event type constants are defined
	eventTypes := []EventType{
		EventEntityCreated,
		EventEntityUpdated,
		EventEntityDeleted,
		EventEntityRead,
		EventTicketCreated,
		EventTicketUpdated,
		EventTicketDeleted,
		EventTicketRead,
		EventProjectCreated,
		EventProjectUpdated,
		EventProjectDeleted,
		EventProjectRead,
		EventCommentCreated,
		EventCommentUpdated,
		EventCommentDeleted,
		EventCommentRead,
		EventPriorityCreated,
		EventPriorityUpdated,
		EventPriorityDeleted,
		EventPriorityRead,
		EventResolutionCreated,
		EventResolutionUpdated,
		EventResolutionDeleted,
		EventResolutionRead,
		EventVersionCreated,
		EventVersionUpdated,
		EventVersionDeleted,
		EventVersionRead,
		EventVersionReleased,
		EventVersionArchived,
		EventWatcherAdded,
		EventWatcherRemoved,
		EventFilterSaved,
		EventFilterUpdated,
		EventFilterDeleted,
		EventFilterShared,
		EventCustomFieldCreated,
		EventCustomFieldUpdated,
		EventCustomFieldDeleted,
		EventCustomFieldRead,
		EventBoardCreated,
		EventBoardUpdated,
		EventBoardDeleted,
		EventCycleCreated,
		EventCycleUpdated,
		EventCycleDeleted,
		EventWorkflowCreated,
		EventWorkflowUpdated,
		EventWorkflowDeleted,
		EventAccountCreated,
		EventAccountUpdated,
		EventAccountDeleted,
		EventOrganizationCreated,
		EventOrganizationUpdated,
		EventOrganizationDeleted,
		EventTeamCreated,
		EventTeamUpdated,
		EventTeamDeleted,
		EventUserCreated,
		EventUserUpdated,
		EventUserDeleted,
		EventSystemHealthCheck,
		EventSystemError,
		EventSystemShutdown,
		EventConnectionEstablished,
		EventConnectionClosed,
		EventConnectionError,
	}

	// Verify all event types are non-empty strings
	for _, eventType := range eventTypes {
		assert.NotEmpty(t, string(eventType), "Event type should not be empty")
	}
}

func TestEventContext(t *testing.T) {
	tests := []struct {
		name    string
		context EventContext
	}{
		{
			name: "Full context",
			context: EventContext{
				ProjectID:      "project-123",
				OrganizationID: "org-456",
				TeamID:         "team-789",
				AccountID:      "account-012",
				Permissions:    []string{"READ", "WRITE", "DELETE"},
			},
		},
		{
			name: "Project context only",
			context: EventContext{
				ProjectID:   "project-123",
				Permissions: []string{"READ"},
			},
		},
		{
			name: "Organization context only",
			context: EventContext{
				OrganizationID: "org-456",
				Permissions:    []string{"READ", "WRITE"},
			},
		},
		{
			name:    "Empty context",
			context: EventContext{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := NewEvent(EventTicketCreated, ActionCreate, "ticket", "ticket-123", "john.doe", nil)
			event.WithContext(tt.context)

			assert.Equal(t, tt.context.ProjectID, event.Context.ProjectID)
			assert.Equal(t, tt.context.OrganizationID, event.Context.OrganizationID)
			assert.Equal(t, tt.context.TeamID, event.Context.TeamID)
			assert.Equal(t, tt.context.AccountID, event.Context.AccountID)
			assert.Equal(t, tt.context.Permissions, event.Context.Permissions)
		})
	}
}

func TestSubscription(t *testing.T) {
	subscription := &Subscription{
		EventTypes:  []EventType{EventTicketCreated, EventTicketUpdated},
		EntityTypes: []string{"ticket", "project"},
		EntityIDs:   []string{"ticket-123", "ticket-456"},
		Filters: map[string]string{
			"projectId": "project-789",
		},
		IncludeReads: true,
		CustomFilters: map[string][]string{
			"status": {"open", "in_progress"},
		},
		PermissionMask: 1,
	}

	assert.NotNil(t, subscription)
	assert.Len(t, subscription.EventTypes, 2)
	assert.Len(t, subscription.EntityTypes, 2)
	assert.Len(t, subscription.EntityIDs, 2)
	assert.Len(t, subscription.Filters, 1)
	assert.True(t, subscription.IncludeReads)
	assert.Len(t, subscription.CustomFilters, 1)
	assert.Equal(t, 1, subscription.PermissionMask)
}

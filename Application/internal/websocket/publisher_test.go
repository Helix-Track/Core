package websocket

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"helixtrack.ru/core/internal/models"
)

// MockManager implements a mock WebSocket manager for testing
type MockManager struct {
	mock.Mock
}

func (m *MockManager) BroadcastEvent(event *models.Event) {
	m.Called(event)
}

func TestNewPublisher(t *testing.T) {
	mockManager := &MockManager{}

	publisher := NewPublisher(mockManager, true)

	assert.NotNil(t, publisher)
	assert.True(t, publisher.IsEnabled())
}

func TestNewPublisher_Disabled(t *testing.T) {
	mockManager := &MockManager{}

	publisher := NewPublisher(mockManager, false)

	assert.NotNil(t, publisher)
	assert.False(t, publisher.IsEnabled())
}

func TestPublisher_PublishEvent(t *testing.T) {
	mockManager := &MockManager{}
	publisher := NewPublisher(mockManager, true)

	event := models.NewEvent(
		models.EventTicketCreated,
		models.ActionCreate,
		"ticket",
		"ticket-123",
		"john.doe",
		map[string]interface{}{
			"title": "Test Ticket",
		},
	)

	// Expect BroadcastEvent to be called once
	mockManager.On("BroadcastEvent", event).Return()

	publisher.PublishEvent(event)

	mockManager.AssertExpectations(t)
	mockManager.AssertCalled(t, "BroadcastEvent", event)
}

func TestPublisher_PublishEvent_WhenDisabled(t *testing.T) {
	mockManager := &MockManager{}
	publisher := NewPublisher(mockManager, false)

	event := models.NewEvent(
		models.EventTicketCreated,
		models.ActionCreate,
		"ticket",
		"ticket-123",
		"john.doe",
		nil,
	)

	// Should NOT call BroadcastEvent when disabled
	publisher.PublishEvent(event)

	mockManager.AssertNotCalled(t, "BroadcastEvent", mock.Anything)
}

func TestPublisher_PublishEvent_WithNilManager(t *testing.T) {
	publisher := NewPublisher(nil, true)

	event := models.NewEvent(
		models.EventTicketCreated,
		models.ActionCreate,
		"ticket",
		"ticket-123",
		"john.doe",
		nil,
	)

	// Should not panic with nil manager
	assert.NotPanics(t, func() {
		publisher.PublishEvent(event)
	})
}

func TestPublisher_PublishEntityEvent(t *testing.T) {
	mockManager := &MockManager{}
	publisher := NewPublisher(mockManager, true)

	action := models.ActionCreate
	object := "ticket"
	entityID := "ticket-123"
	username := "john.doe"
	data := map[string]interface{}{
		"title":       "Test Ticket",
		"description": "Test Description",
	}
	context := models.EventContext{
		ProjectID:   "project-456",
		Permissions: []string{"READ"},
	}

	// Expect BroadcastEvent to be called once
	mockManager.On("BroadcastEvent", mock.MatchedBy(func(e *models.Event) bool {
		return e.Action == action &&
			e.Object == object &&
			e.EntityID == entityID &&
			e.Username == username &&
			e.Context.ProjectID == context.ProjectID
	})).Return()

	publisher.PublishEntityEvent(action, object, entityID, username, data, context)

	mockManager.AssertExpectations(t)
}

func TestPublisher_PublishEntityEvent_WhenDisabled(t *testing.T) {
	mockManager := &MockManager{}
	publisher := NewPublisher(mockManager, false)

	action := models.ActionCreate
	object := "ticket"
	entityID := "ticket-123"
	username := "john.doe"
	data := map[string]interface{}{"title": "Test"}
	context := models.EventContext{}

	// Should NOT call BroadcastEvent when disabled
	publisher.PublishEntityEvent(action, object, entityID, username, data, context)

	mockManager.AssertNotCalled(t, "BroadcastEvent", mock.Anything)
}

func TestPublisher_IsEnabled(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
		want    bool
	}{
		{
			name:    "Publisher enabled",
			enabled: true,
			want:    true,
		},
		{
			name:    "Publisher disabled",
			enabled: false,
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockManager := &MockManager{}
			publisher := NewPublisher(mockManager, tt.enabled)

			got := publisher.IsEnabled()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNoOpPublisher_PublishEvent(t *testing.T) {
	publisher := NewNoOpPublisher()

	event := models.NewEvent(
		models.EventTicketCreated,
		models.ActionCreate,
		"ticket",
		"ticket-123",
		"john.doe",
		nil,
	)

	// Should not panic
	assert.NotPanics(t, func() {
		publisher.PublishEvent(event)
	})
}

func TestNoOpPublisher_PublishEntityEvent(t *testing.T) {
	publisher := NewNoOpPublisher()

	// Should not panic
	assert.NotPanics(t, func() {
		publisher.PublishEntityEvent(
			models.ActionCreate,
			"ticket",
			"ticket-123",
			"john.doe",
			nil,
			models.EventContext{},
		)
	})
}

func TestNoOpPublisher_IsEnabled(t *testing.T) {
	publisher := NewNoOpPublisher()

	assert.False(t, publisher.IsEnabled())
}

func TestNewProjectContext(t *testing.T) {
	projectID := "project-123"
	permissions := []string{"READ", "WRITE"}

	context := NewProjectContext(projectID, permissions)

	assert.Equal(t, projectID, context.ProjectID)
	assert.Equal(t, permissions, context.Permissions)
	assert.Empty(t, context.OrganizationID)
	assert.Empty(t, context.TeamID)
	assert.Empty(t, context.AccountID)
}

func TestNewOrganizationContext(t *testing.T) {
	organizationID := "org-456"
	permissions := []string{"READ"}

	context := NewOrganizationContext(organizationID, permissions)

	assert.Equal(t, organizationID, context.OrganizationID)
	assert.Equal(t, permissions, context.Permissions)
	assert.Empty(t, context.ProjectID)
	assert.Empty(t, context.TeamID)
	assert.Empty(t, context.AccountID)
}

func TestNewTeamContext(t *testing.T) {
	teamID := "team-789"
	organizationID := "org-456"
	permissions := []string{"READ", "WRITE", "DELETE"}

	context := NewTeamContext(teamID, organizationID, permissions)

	assert.Equal(t, teamID, context.TeamID)
	assert.Equal(t, organizationID, context.OrganizationID)
	assert.Equal(t, permissions, context.Permissions)
	assert.Empty(t, context.ProjectID)
	assert.Empty(t, context.AccountID)
}

func TestNewAccountContext(t *testing.T) {
	accountID := "account-012"
	permissions := []string{"READ", "WRITE"}

	context := NewAccountContext(accountID, permissions)

	assert.Equal(t, accountID, context.AccountID)
	assert.Equal(t, permissions, context.Permissions)
	assert.Empty(t, context.ProjectID)
	assert.Empty(t, context.OrganizationID)
	assert.Empty(t, context.TeamID)
}

func TestNewFullContext(t *testing.T) {
	projectID := "project-123"
	organizationID := "org-456"
	teamID := "team-789"
	accountID := "account-012"
	permissions := []string{"READ", "WRITE", "UPDATE", "DELETE"}

	context := NewFullContext(projectID, organizationID, teamID, accountID, permissions)

	assert.Equal(t, projectID, context.ProjectID)
	assert.Equal(t, organizationID, context.OrganizationID)
	assert.Equal(t, teamID, context.TeamID)
	assert.Equal(t, accountID, context.AccountID)
	assert.Equal(t, permissions, context.Permissions)
}

func TestPublisher_MultipleEvents(t *testing.T) {
	mockManager := &MockManager{}
	publisher := NewPublisher(mockManager, true)

	// Create multiple events
	events := []*models.Event{
		models.NewEvent(models.EventTicketCreated, models.ActionCreate, "ticket", "ticket-1", "user1", nil),
		models.NewEvent(models.EventTicketUpdated, models.ActionModify, "ticket", "ticket-2", "user2", nil),
		models.NewEvent(models.EventTicketDeleted, models.ActionRemove, "ticket", "ticket-3", "user3", nil),
	}

	// Expect BroadcastEvent to be called for each event
	for _, event := range events {
		mockManager.On("BroadcastEvent", event).Return()
	}

	// Publish all events
	for _, event := range events {
		publisher.PublishEvent(event)
	}

	mockManager.AssertExpectations(t)
	assert.Equal(t, 3, len(mockManager.Calls))
}

func TestPublisher_ConcurrentPublishing(t *testing.T) {
	mockManager := &MockManager{}
	publisher := NewPublisher(mockManager, true)

	eventCount := 100

	// Expect BroadcastEvent to be called many times
	mockManager.On("BroadcastEvent", mock.Anything).Return()

	// Publish events concurrently
	done := make(chan bool, eventCount)
	for i := 0; i < eventCount; i++ {
		go func(index int) {
			event := models.NewEvent(
				models.EventTicketCreated,
				models.ActionCreate,
				"ticket",
				"ticket-"+string(rune(index)),
				"user",
				nil,
			)
			publisher.PublishEvent(event)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < eventCount; i++ {
		<-done
	}

	// Verify all events were published
	assert.Equal(t, eventCount, len(mockManager.Calls))
}

package websocket

import (
	"go.uber.org/zap"
	"helixtrack.ru/core/internal/logger"
	"helixtrack.ru/core/internal/models"
)

// EventPublisher is an interface for publishing events
type EventPublisher interface {
	// PublishEvent publishes an event to all subscribed clients
	PublishEvent(event *models.Event)

	// PublishEntityEvent publishes an event for entity operations (CRUD)
	PublishEntityEvent(action, object, entityID, username string, data map[string]interface{}, context models.EventContext)

	// IsEnabled returns whether event publishing is enabled
	IsEnabled() bool
}

// Publisher implements the EventPublisher interface
type Publisher struct {
	manager *Manager
	enabled bool
}

// NewPublisher creates a new event publisher
func NewPublisher(manager *Manager, enabled bool) EventPublisher {
	return &Publisher{
		manager: manager,
		enabled: enabled,
	}
}

// PublishEvent publishes an event to all subscribed clients
func (p *Publisher) PublishEvent(event *models.Event) {
	if !p.enabled || p.manager == nil {
		return
	}

	logger.Debug("Publishing event",
		zap.String("eventId", event.ID),
		zap.String("type", string(event.Type)),
		zap.String("action", event.Action),
		zap.String("object", event.Object),
		zap.String("entityId", event.EntityID),
		zap.String("username", event.Username),
	)

	p.manager.BroadcastEvent(event)
}

// PublishEntityEvent publishes an event for entity operations (CRUD)
func (p *Publisher) PublishEntityEvent(action, object, entityID, username string, data map[string]interface{}, context models.EventContext) {
	if !p.enabled || p.manager == nil {
		return
	}

	// Determine event type based on action and object
	eventType := models.GetEventTypeFromAction(action, object)

	// Create event
	event := models.NewEvent(eventType, action, object, entityID, username, data)
	event.Context = context

	// Publish event
	p.PublishEvent(event)
}

// IsEnabled returns whether event publishing is enabled
func (p *Publisher) IsEnabled() bool {
	return p.enabled
}

// NoOpPublisher is a no-op implementation of EventPublisher
type NoOpPublisher struct{}

// NewNoOpPublisher creates a no-op publisher (for when WebSocket is disabled)
func NewNoOpPublisher() EventPublisher {
	return &NoOpPublisher{}
}

// PublishEvent does nothing
func (n *NoOpPublisher) PublishEvent(event *models.Event) {}

// PublishEntityEvent does nothing
func (n *NoOpPublisher) PublishEntityEvent(action, object, entityID, username string, data map[string]interface{}, context models.EventContext) {}

// IsEnabled returns false
func (n *NoOpPublisher) IsEnabled() bool {
	return false
}

// Helper functions for creating common event contexts

// NewProjectContext creates an event context for project-related events
func NewProjectContext(projectID string, permissions []string) models.EventContext {
	return models.EventContext{
		ProjectID:   projectID,
		Permissions: permissions,
	}
}

// NewOrganizationContext creates an event context for organization-related events
func NewOrganizationContext(organizationID string, permissions []string) models.EventContext {
	return models.EventContext{
		OrganizationID: organizationID,
		Permissions:    permissions,
	}
}

// NewTeamContext creates an event context for team-related events
func NewTeamContext(teamID, organizationID string, permissions []string) models.EventContext {
	return models.EventContext{
		TeamID:         teamID,
		OrganizationID: organizationID,
		Permissions:    permissions,
	}
}

// NewAccountContext creates an event context for account-related events
func NewAccountContext(accountID string, permissions []string) models.EventContext {
	return models.EventContext{
		AccountID:   accountID,
		Permissions: permissions,
	}
}

// NewFullContext creates an event context with all fields
func NewFullContext(projectID, organizationID, teamID, accountID string, permissions []string) models.EventContext {
	return models.EventContext{
		ProjectID:      projectID,
		OrganizationID: organizationID,
		TeamID:         teamID,
		AccountID:      accountID,
		Permissions:    permissions,
	}
}

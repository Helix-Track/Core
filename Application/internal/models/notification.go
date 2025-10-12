package models

// NotificationScheme represents a notification scheme for a project
type NotificationScheme struct {
	ID          string  `json:"id" db:"id"`
	Title       string  `json:"title" db:"title" binding:"required"`
	Description string  `json:"description,omitempty" db:"description"`
	ProjectID   *string `json:"projectId,omitempty" db:"project_id"` // NULL for global schemes
	Created     int64   `json:"created" db:"created"`
	Modified    int64   `json:"modified" db:"modified"`
	Deleted     bool    `json:"deleted" db:"deleted"`
}

// NotificationEvent represents a type of event that can trigger notifications
type NotificationEvent struct {
	ID          string `json:"id" db:"id"`
	EventType   string `json:"eventType" db:"event_type" binding:"required"`
	Title       string `json:"title" db:"title" binding:"required"`
	Description string `json:"description,omitempty" db:"description"`
	Created     int64  `json:"created" db:"created"`
	Deleted     bool   `json:"deleted" db:"deleted"`
}

// NotificationRule represents a rule for sending notifications
type NotificationRule struct {
	ID                   string  `json:"id" db:"id"`
	NotificationSchemeID string  `json:"notificationSchemeId" db:"notification_scheme_id" binding:"required"`
	NotificationEventID  string  `json:"notificationEventId" db:"notification_event_id" binding:"required"`
	RecipientType        string  `json:"recipientType" db:"recipient_type" binding:"required"` // assignee, reporter, watcher, user, team, project_role
	RecipientID          *string `json:"recipientId,omitempty" db:"recipient_id"`              // user_id, team_id, or role_id (NULL for assignee/reporter/watcher)
	Created              int64   `json:"created" db:"created"`
	Deleted              bool    `json:"deleted" db:"deleted"`
}

// Event type constants (must match database seed data)
const (
	EventIssueCreated    = "issue_created"
	EventIssueUpdated    = "issue_updated"
	EventIssueDeleted    = "issue_deleted"
	EventCommentAdded    = "comment_added"
	EventCommentUpdated  = "comment_updated"
	EventCommentDeleted  = "comment_deleted"
	EventStatusChanged   = "status_changed"
	EventAssigneeChanged = "assignee_changed"
	EventPriorityChanged = "priority_changed"
	EventWorkLogged      = "work_logged"
	EventUserMentioned   = "user_mentioned"
)

// Recipient type constants
const (
	RecipientTypeAssignee    = "assignee"
	RecipientTypeReporter    = "reporter"
	RecipientTypeWatcher     = "watcher"
	RecipientTypeUser        = "user"
	RecipientTypeTeam        = "team"
	RecipientTypeProjectRole = "project_role"
)

// IsGlobal checks if the scheme is global (not project-specific)
func (ns *NotificationScheme) IsGlobal() bool {
	return ns.ProjectID == nil || *ns.ProjectID == ""
}

// IsProjectSpecific checks if the scheme is project-specific
func (ns *NotificationScheme) IsProjectSpecific() bool {
	return !ns.IsGlobal()
}

// IsValidEventType checks if the event type is valid
func (ne *NotificationEvent) IsValidEventType() bool {
	validEvents := map[string]bool{
		EventIssueCreated:    true,
		EventIssueUpdated:    true,
		EventIssueDeleted:    true,
		EventCommentAdded:    true,
		EventCommentUpdated:  true,
		EventCommentDeleted:  true,
		EventStatusChanged:   true,
		EventAssigneeChanged: true,
		EventPriorityChanged: true,
		EventWorkLogged:      true,
		EventUserMentioned:   true,
	}
	return validEvents[ne.EventType]
}

// IsValidRecipientType checks if the recipient type is valid
func (nr *NotificationRule) IsValidRecipientType() bool {
	validRecipients := map[string]bool{
		RecipientTypeAssignee:    true,
		RecipientTypeReporter:    true,
		RecipientTypeWatcher:     true,
		RecipientTypeUser:        true,
		RecipientTypeTeam:        true,
		RecipientTypeProjectRole: true,
	}
	return validRecipients[nr.RecipientType]
}

// RequiresRecipientID checks if the recipient type requires a recipient ID
func (nr *NotificationRule) RequiresRecipientID() bool {
	return nr.RecipientType == RecipientTypeUser ||
		nr.RecipientType == RecipientTypeTeam ||
		nr.RecipientType == RecipientTypeProjectRole
}

// IsRoleBasedRecipient checks if the recipient is role-based (assignee, reporter, watcher)
func (nr *NotificationRule) IsRoleBasedRecipient() bool {
	return nr.RecipientType == RecipientTypeAssignee ||
		nr.RecipientType == RecipientTypeReporter ||
		nr.RecipientType == RecipientTypeWatcher
}

package models

// Request represents the unified API request format for the /do endpoint
type Request struct {
	Action string                 `json:"action" binding:"required"` // authenticate, version, jwtCapable, dbCapable, health, create, modify, remove
	JWT    string                 `json:"jwt"`                       // Required for authenticated actions
	Locale string                 `json:"locale"`                    // Optional locale for localized responses
	Object string                 `json:"object"`                    // Required for CRUD operations (create, modify, remove)
	Data   map[string]interface{} `json:"data"`                      // Additional data for the action
}

// IsAuthenticationRequired returns true if the action requires JWT authentication
func (r *Request) IsAuthenticationRequired() bool {
	switch r.Action {
	case ActionVersion, ActionJWTCapable, ActionDBCapable, ActionHealth:
		return false
	default:
		return true
	}
}

// IsCRUDOperation returns true if the action is a CRUD operation requiring an object
func (r *Request) IsCRUDOperation() bool {
	switch r.Action {
	case ActionCreate, ActionModify, ActionRemove:
		return true
	default:
		return false
	}
}

// Action constants
const (
	// System actions
	ActionAuthenticate = "authenticate"
	ActionVersion      = "version"
	ActionJWTCapable   = "jwtCapable"
	ActionDBCapable    = "dbCapable"
	ActionHealth       = "health"

	// Generic CRUD actions
	ActionCreate = "create"
	ActionModify = "modify"
	ActionRemove = "remove"
	ActionRead   = "read"
	ActionList   = "list"

	// Priority actions
	ActionPriorityCreate = "priorityCreate"
	ActionPriorityRead   = "priorityRead"
	ActionPriorityList   = "priorityList"
	ActionPriorityModify = "priorityModify"
	ActionPriorityRemove = "priorityRemove"

	// Resolution actions
	ActionResolutionCreate = "resolutionCreate"
	ActionResolutionRead   = "resolutionRead"
	ActionResolutionList   = "resolutionList"
	ActionResolutionModify = "resolutionModify"
	ActionResolutionRemove = "resolutionRemove"

	// Version actions
	ActionVersionCreate  = "versionCreate"
	ActionVersionRead    = "versionRead"
	ActionVersionList    = "versionList"
	ActionVersionModify  = "versionModify"
	ActionVersionRemove  = "versionRemove"
	ActionVersionRelease = "versionRelease" // Mark version as released
	ActionVersionArchive = "versionArchive" // Archive a version

	// Version mapping actions
	ActionVersionAddAffected    = "versionAddAffected"    // Add affected version to ticket
	ActionVersionRemoveAffected = "versionRemoveAffected" // Remove affected version from ticket
	ActionVersionListAffected   = "versionListAffected"   // List affected versions for ticket
	ActionVersionAddFix         = "versionAddFix"         // Add fix version to ticket
	ActionVersionRemoveFix      = "versionRemoveFix"      // Remove fix version from ticket
	ActionVersionListFix        = "versionListFix"        // List fix versions for ticket

	// Watcher actions
	ActionWatcherAdd    = "watcherAdd"    // Start watching a ticket
	ActionWatcherRemove = "watcherRemove" // Stop watching a ticket
	ActionWatcherList   = "watcherList"   // List watchers for a ticket

	// Filter actions
	ActionFilterSave   = "filterSave"   // Create or update a saved filter
	ActionFilterLoad   = "filterLoad"   // Load a saved filter
	ActionFilterList   = "filterList"   // List user's filters
	ActionFilterShare  = "filterShare"  // Share a filter
	ActionFilterModify = "filterModify" // Modify a filter
	ActionFilterRemove = "filterRemove" // Delete a filter

	// Custom field actions
	ActionCustomFieldCreate = "customFieldCreate"
	ActionCustomFieldRead   = "customFieldRead"
	ActionCustomFieldList   = "customFieldList"
	ActionCustomFieldModify = "customFieldModify"
	ActionCustomFieldRemove = "customFieldRemove"

	// Custom field option actions
	ActionCustomFieldOptionCreate = "customFieldOptionCreate"
	ActionCustomFieldOptionModify = "customFieldOptionModify"
	ActionCustomFieldOptionRemove = "customFieldOptionRemove"
	ActionCustomFieldOptionList   = "customFieldOptionList"

	// Custom field value actions (for tickets)
	ActionCustomFieldValueSet    = "customFieldValueSet"    // Set custom field value for a ticket
	ActionCustomFieldValueGet    = "customFieldValueGet"    // Get custom field value for a ticket
	ActionCustomFieldValueList   = "customFieldValueList"   // List all custom field values for a ticket
	ActionCustomFieldValueRemove = "customFieldValueRemove" // Remove custom field value from a ticket
)

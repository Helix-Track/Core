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
	ActionAuthenticate = "authenticate"
	ActionVersion      = "version"
	ActionJWTCapable   = "jwtCapable"
	ActionDBCapable    = "dbCapable"
	ActionHealth       = "health"
	ActionCreate       = "create"
	ActionModify       = "modify"
	ActionRemove       = "remove"
	ActionRead         = "read"
	ActionList         = "list"
)

package models

// Extension represents an extension/plugin in the system
type Extension struct {
	ID          string `json:"id" db:"id"`
	Title       string `json:"title" db:"title" binding:"required"`
	Description string `json:"description,omitempty" db:"description"`
	Version     string `json:"version" db:"version"`
	Enabled     bool   `json:"enabled" db:"enabled"`
	Created     int64  `json:"created" db:"created"`
	Modified    int64  `json:"modified" db:"modified"`
	Deleted     bool   `json:"deleted" db:"deleted"`
}

// ExtensionMetaData represents additional metadata for extensions
type ExtensionMetaData struct {
	ID          string `json:"id" db:"id"`
	ExtensionID string `json:"extensionId" db:"extension_id" binding:"required"`
	Property    string `json:"property" db:"property" binding:"required"`
	Value       string `json:"value" db:"value"`
	Created     int64  `json:"created" db:"created"`
	Modified    int64  `json:"modified" db:"modified"`
	Deleted     bool   `json:"deleted" db:"deleted"`
}

// IsValid validates if the extension has required fields
func (e *Extension) IsValid() bool {
	return e.Title != ""
}

// IsEnabled checks if the extension is enabled
func (e *Extension) IsEnabled() bool {
	return e.Enabled && !e.Deleted
}

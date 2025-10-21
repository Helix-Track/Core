package websocket

import (
	"encoding/json"
	"time"
)

// EventType represents the type of localization event
type EventType string

const (
	// Language events
	EventLanguageAdded   EventType = "language.added"
	EventLanguageUpdated EventType = "language.updated"
	EventLanguageDeleted EventType = "language.deleted"

	// Localization key events
	EventKeyAdded   EventType = "key.added"
	EventKeyUpdated EventType = "key.updated"
	EventKeyDeleted EventType = "key.deleted"

	// Localization (translation) events
	EventLocalizationAdded    EventType = "localization.added"
	EventLocalizationUpdated  EventType = "localization.updated"
	EventLocalizationDeleted  EventType = "localization.deleted"
	EventLocalizationApproved EventType = "localization.approved"

	// Batch events
	EventBatchOperationCompleted EventType = "batch.completed"

	// Catalog events
	EventCatalogRebuilt   EventType = "catalog.rebuilt"
	EventCacheInvalidated EventType = "cache.invalidated"

	// Version events
	EventVersionCreated EventType = "version.created"
	EventVersionDeleted EventType = "version.deleted"
)

// Event represents a localization event broadcasted via WebSocket
type Event struct {
	Type      EventType       `json:"type"`
	Timestamp time.Time       `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
	Metadata  *EventMetadata  `json:"metadata,omitempty"`
}

// EventMetadata contains additional information about the event
type EventMetadata struct {
	UserID        string `json:"userId,omitempty"`
	Username      string `json:"username,omitempty"`
	CorrelationID string `json:"correlationId,omitempty"`
}

// LanguageEventData contains data for language events
type LanguageEventData struct {
	ID         string `json:"id"`
	Code       string `json:"code"`
	Name       string `json:"name"`
	NativeName string `json:"nativeName"`
	IsRTL      bool   `json:"isRTL"`
	IsActive   bool   `json:"isActive"`
}

// KeyEventData contains data for localization key events
type KeyEventData struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	Category    string `json:"category"`
	Description string `json:"description"`
	IsApproved  bool   `json:"isApproved"`
}

// LocalizationEventData contains data for localization (translation) events
type LocalizationEventData struct {
	ID           string `json:"id"`
	KeyID        string `json:"keyId"`
	Key          string `json:"key"`
	LanguageID   string `json:"languageId"`
	LanguageCode string `json:"languageCode"`
	Value        string `json:"value"`
	IsApproved   bool   `json:"isApproved"`
	ApprovedBy   string `json:"approvedBy,omitempty"`
}

// BatchOperationEventData contains data for batch operation events
type BatchOperationEventData struct {
	Operation string `json:"operation"`
	Processed int    `json:"processed"`
	Failed    int    `json:"failed"`
	Duration  string `json:"duration"`
}

// CatalogEventData contains data for catalog events
type CatalogEventData struct {
	Language string `json:"language,omitempty"` // Empty for all languages
	Version  int    `json:"version"`
	Checksum string `json:"checksum"`
}

// CacheInvalidatedEventData contains data for cache invalidation events
type CacheInvalidatedEventData struct {
	Language string `json:"language,omitempty"` // Empty for all languages
	Reason   string `json:"reason"`
}

// VersionEventData contains data for version events
type VersionEventData struct {
	ID                 string `json:"id"`
	Version            string `json:"version"`
	Description        string `json:"description"`
	KeysCount          int    `json:"keysCount"`
	LanguagesCount     int    `json:"languagesCount"`
	TranslationsCount  int    `json:"translationsCount"`
}

// NewEvent creates a new event with the given type and data
func NewEvent(eventType EventType, data interface{}, metadata *EventMetadata) (*Event, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return &Event{
		Type:      eventType,
		Timestamp: time.Now().UTC(),
		Data:      dataBytes,
		Metadata:  metadata,
	}, nil
}

// ParseData parses the event data into the given interface
func (e *Event) ParseData(v interface{}) error {
	return json.Unmarshal(e.Data, v)
}

// ToJSON converts the event to JSON bytes
func (e *Event) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// FromJSON parses an event from JSON bytes
func FromJSON(data []byte) (*Event, error) {
	var event Event
	err := json.Unmarshal(data, &event)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

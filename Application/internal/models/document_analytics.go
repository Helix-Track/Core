package models

import (
	"errors"
	"time"
)

// DocumentViewHistory represents a single view of a document
type DocumentViewHistory struct {
	ID           string  `json:"id" db:"id"`
	DocumentID   string  `json:"document_id" db:"document_id"`
	UserID       *string `json:"user_id,omitempty" db:"user_id"` // NULL for anonymous
	IPAddress    *string `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent    *string `json:"user_agent,omitempty" db:"user_agent"`
	SessionID    *string `json:"session_id,omitempty" db:"session_id"`
	ViewDuration *int    `json:"view_duration,omitempty" db:"view_duration"` // Seconds
	Timestamp    int64   `json:"timestamp" db:"timestamp"`
}

// Validate validates the view history entry
func (dvh *DocumentViewHistory) Validate() error {
	if dvh.ID == "" {
		return errors.New("view history ID cannot be empty")
	}
	if dvh.DocumentID == "" {
		return errors.New("view history document ID cannot be empty")
	}
	if dvh.Timestamp == 0 {
		return errors.New("view history timestamp cannot be zero")
	}
	return nil
}

// SetTimestamps sets the timestamp
func (dvh *DocumentViewHistory) SetTimestamps() {
	if dvh.Timestamp == 0 {
		dvh.Timestamp = time.Now().Unix()
	}
}

// DocumentAnalytics represents aggregated analytics for a document
type DocumentAnalytics struct {
	ID               string   `json:"id" db:"id"`
	DocumentID       string   `json:"document_id" db:"document_id"`
	TotalViews       int      `json:"total_views" db:"total_views"`
	UniqueViewers    int      `json:"unique_viewers" db:"unique_viewers"`
	TotalEdits       int      `json:"total_edits" db:"total_edits"`
	UniqueEditors    int      `json:"unique_editors" db:"unique_editors"`
	TotalComments    int      `json:"total_comments" db:"total_comments"`
	TotalReactions   int      `json:"total_reactions" db:"total_reactions"`
	TotalWatchers    int      `json:"total_watchers" db:"total_watchers"`
	AvgViewDuration  *int     `json:"avg_view_duration,omitempty" db:"avg_view_duration"` // Seconds
	LastViewed       *int64   `json:"last_viewed,omitempty" db:"last_viewed"`
	LastEdited       *int64   `json:"last_edited,omitempty" db:"last_edited"`
	PopularityScore  float64  `json:"popularity_score" db:"popularity_score"`
	Updated          int64    `json:"updated" db:"updated"`
}

// Validate validates the document analytics
func (da *DocumentAnalytics) Validate() error {
	if da.ID == "" {
		return errors.New("analytics ID cannot be empty")
	}
	if da.DocumentID == "" {
		return errors.New("analytics document ID cannot be empty")
	}
	if da.Updated == 0 {
		return errors.New("analytics updated timestamp cannot be zero")
	}
	return nil
}

// SetTimestamps sets the updated timestamp
func (da *DocumentAnalytics) SetTimestamps() {
	da.Updated = time.Now().Unix()
}

// CalculatePopularityScore calculates a popularity score based on metrics
func (da *DocumentAnalytics) CalculatePopularityScore() {
	// Simple algorithm: weighted sum of various metrics
	score := float64(da.TotalViews)*0.1 +
		float64(da.UniqueViewers)*0.3 +
		float64(da.TotalEdits)*0.2 +
		float64(da.TotalComments)*0.2 +
		float64(da.TotalReactions)*0.1 +
		float64(da.TotalWatchers)*0.1

	da.PopularityScore = score
}

// IncrementView increments view counters
func (da *DocumentAnalytics) IncrementView(isUnique bool) {
	da.TotalViews++
	if isUnique {
		da.UniqueViewers++
	}
	now := time.Now().Unix()
	da.LastViewed = &now
	da.SetTimestamps()
}

// IncrementEdit increments edit counters
func (da *DocumentAnalytics) IncrementEdit(isUnique bool) {
	da.TotalEdits++
	if isUnique {
		da.UniqueEditors++
	}
	now := time.Now().Unix()
	da.LastEdited = &now
	da.SetTimestamps()
}

// IncrementComment increments comment counter
func (da *DocumentAnalytics) IncrementComment() {
	da.TotalComments++
	da.SetTimestamps()
}

// IncrementReaction increments reaction counter
func (da *DocumentAnalytics) IncrementReaction() {
	da.TotalReactions++
	da.SetTimestamps()
}

// IncrementWatcher increments watcher counter
func (da *DocumentAnalytics) IncrementWatcher() {
	da.TotalWatchers++
	da.SetTimestamps()
}

// DecrementComment decrements comment counter
func (da *DocumentAnalytics) DecrementComment() {
	if da.TotalComments > 0 {
		da.TotalComments--
	}
	da.SetTimestamps()
}

// DecrementReaction decrements reaction counter
func (da *DocumentAnalytics) DecrementReaction() {
	if da.TotalReactions > 0 {
		da.TotalReactions--
	}
	da.SetTimestamps()
}

// DecrementWatcher decrements watcher counter
func (da *DocumentAnalytics) DecrementWatcher() {
	if da.TotalWatchers > 0 {
		da.TotalWatchers--
	}
	da.SetTimestamps()
}

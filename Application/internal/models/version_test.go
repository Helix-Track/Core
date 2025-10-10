package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersion_IsReleased(t *testing.T) {
	tests := []struct {
		name     string
		released bool
		expected bool
	}{
		{"Released version", true, true},
		{"Unreleased version", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Version{Released: tt.released}
			result := v.IsReleased()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestVersion_IsArchived(t *testing.T) {
	tests := []struct {
		name     string
		archived bool
		expected bool
	}{
		{"Archived version", true, true},
		{"Not archived version", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Version{Archived: tt.archived}
			result := v.IsArchived()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestVersion_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		archived bool
		deleted  bool
		expected bool
	}{
		{"Active version", false, false, true},
		{"Archived version", true, false, false},
		{"Deleted version", false, true, false},
		{"Archived and deleted", true, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Version{
				Archived: tt.archived,
				Deleted:  tt.deleted,
			}
			result := v.IsActive()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestVersion_Structure(t *testing.T) {
	startDate := int64(1234567890)
	releaseDate := int64(1234667890)

	v := Version{
		ID:          "version-1",
		Title:       "v1.0.0",
		Description: "First release",
		ProjectID:   "project-1",
		StartDate:   &startDate,
		ReleaseDate: &releaseDate,
		Released:    true,
		Archived:    false,
		Created:     1234567890,
		Modified:    1234567890,
		Deleted:     false,
	}

	assert.Equal(t, "version-1", v.ID)
	assert.Equal(t, "v1.0.0", v.Title)
	assert.Equal(t, "First release", v.Description)
	assert.Equal(t, "project-1", v.ProjectID)
	assert.NotNil(t, v.StartDate)
	assert.Equal(t, int64(1234567890), *v.StartDate)
	assert.NotNil(t, v.ReleaseDate)
	assert.Equal(t, int64(1234667890), *v.ReleaseDate)
	assert.True(t, v.IsReleased())
	assert.False(t, v.IsArchived())
	assert.True(t, v.IsActive())
}

func TestVersion_NullableDates(t *testing.T) {
	v := Version{
		ID:          "version-1",
		Title:       "v2.0.0",
		ProjectID:   "project-1",
		StartDate:   nil,
		ReleaseDate: nil,
		Released:    false,
		Archived:    false,
	}

	assert.Nil(t, v.StartDate)
	assert.Nil(t, v.ReleaseDate)
	assert.False(t, v.IsReleased())
	assert.True(t, v.IsActive())
}

func TestTicketVersionMapping_Structure(t *testing.T) {
	mapping := TicketVersionMapping{
		ID:        "mapping-1",
		TicketID:  "ticket-1",
		VersionID: "version-1",
		Created:   1234567890,
		Deleted:   false,
	}

	assert.Equal(t, "mapping-1", mapping.ID)
	assert.Equal(t, "ticket-1", mapping.TicketID)
	assert.Equal(t, "version-1", mapping.VersionID)
	assert.False(t, mapping.Deleted)
}

func TestVersion_LifecycleStates(t *testing.T) {
	// Test version lifecycle
	v := &Version{
		Released: false,
		Archived: false,
		Deleted:  false,
	}

	// Initial state: unreleased, active
	assert.False(t, v.IsReleased())
	assert.True(t, v.IsActive())

	// Released
	v.Released = true
	assert.True(t, v.IsReleased())
	assert.True(t, v.IsActive())

	// Archived
	v.Archived = true
	assert.True(t, v.IsReleased())
	assert.False(t, v.IsActive())

	// Deleted
	v.Deleted = true
	assert.True(t, v.IsReleased())
	assert.False(t, v.IsActive())
}

func BenchmarkVersion_IsActive(b *testing.B) {
	v := &Version{Archived: false, Deleted: false}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.IsActive()
	}
}

func BenchmarkVersion_IsReleased(b *testing.B) {
	v := &Version{Released: true}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.IsReleased()
	}
}

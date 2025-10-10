package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolution_GetDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		expected string
	}{
		{"With title", "Fixed", "Fixed"},
		{"Empty title", "", "Unknown Resolution"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Resolution{Title: tt.title}
			result := r.GetDisplayName()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResolutionIDConstants(t *testing.T) {
	assert.Equal(t, "resolution-fixed", ResolutionIDFixed)
	assert.Equal(t, "resolution-wont-fix", ResolutionIDWontFix)
	assert.Equal(t, "resolution-duplicate", ResolutionIDDuplicate)
	assert.Equal(t, "resolution-incomplete", ResolutionIDIncomplete)
	assert.Equal(t, "resolution-cannot-reproduce", ResolutionIDCannotReproduce)
	assert.Equal(t, "resolution-done", ResolutionIDDone)
}

func TestResolution_Structure(t *testing.T) {
	r := Resolution{
		ID:          ResolutionIDFixed,
		Title:       "Fixed",
		Description: "Issue has been fixed",
		Created:     1234567890,
		Modified:    1234567890,
		Deleted:     false,
	}

	assert.Equal(t, ResolutionIDFixed, r.ID)
	assert.Equal(t, "Fixed", r.Title)
	assert.Equal(t, "Issue has been fixed", r.Description)
	assert.Equal(t, "Fixed", r.GetDisplayName())
}

func TestResolution_AllResolutions(t *testing.T) {
	resolutions := []Resolution{
		{ID: ResolutionIDFixed, Title: "Fixed"},
		{ID: ResolutionIDWontFix, Title: "Won't Fix"},
		{ID: ResolutionIDDuplicate, Title: "Duplicate"},
		{ID: ResolutionIDIncomplete, Title: "Incomplete"},
		{ID: ResolutionIDCannotReproduce, Title: "Cannot Reproduce"},
		{ID: ResolutionIDDone, Title: "Done"},
	}

	for _, r := range resolutions {
		assert.Equal(t, r.Title, r.GetDisplayName(), "Resolution %s should return correct display name", r.ID)
		assert.NotEmpty(t, r.ID)
		assert.NotEmpty(t, r.Title)
	}
}

func BenchmarkResolution_GetDisplayName(b *testing.B) {
	r := &Resolution{Title: "Fixed"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.GetDisplayName()
	}
}

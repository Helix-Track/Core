package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPriority_IsValidLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    int
		expected bool
	}{
		{"Lowest (1)", 1, true},
		{"Low (2)", 2, true},
		{"Medium (3)", 3, true},
		{"High (4)", 4, true},
		{"Highest (5)", 5, true},
		{"Too low (0)", 0, false},
		{"Too high (6)", 6, false},
		{"Negative", -1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Priority{Level: tt.level}
			result := p.IsValidLevel()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPriority_GetDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		expected string
	}{
		{"With title", "High Priority", "High Priority"},
		{"Empty title", "", "Unknown Priority"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Priority{Title: tt.title}
			result := p.GetDisplayName()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPriorityLevelConstants(t *testing.T) {
	assert.Equal(t, 1, PriorityLevelLowest)
	assert.Equal(t, 2, PriorityLevelLow)
	assert.Equal(t, 3, PriorityLevelMedium)
	assert.Equal(t, 4, PriorityLevelHigh)
	assert.Equal(t, 5, PriorityLevelHighest)
}

func TestPriorityIDConstants(t *testing.T) {
	assert.Equal(t, "priority-lowest", PriorityIDLowest)
	assert.Equal(t, "priority-low", PriorityIDLow)
	assert.Equal(t, "priority-medium", PriorityIDMedium)
	assert.Equal(t, "priority-high", PriorityIDHigh)
	assert.Equal(t, "priority-highest", PriorityIDHighest)
}

func TestPriority_Structure(t *testing.T) {
	p := Priority{
		ID:          PriorityIDHigh,
		Title:       "High",
		Description: "High priority issues",
		Level:       PriorityLevelHigh,
		Icon:        "⚠️",
		Color:       "#FF0000",
		Created:     1234567890,
		Modified:    1234567890,
		Deleted:     false,
	}

	assert.Equal(t, PriorityIDHigh, p.ID)
	assert.Equal(t, "High", p.Title)
	assert.Equal(t, "High priority issues", p.Description)
	assert.Equal(t, 4, p.Level)
	assert.Equal(t, "⚠️", p.Icon)
	assert.Equal(t, "#FF0000", p.Color)
	assert.True(t, p.IsValidLevel())
	assert.Equal(t, "High", p.GetDisplayName())
}

func TestPriority_AllLevels(t *testing.T) {
	priorities := []Priority{
		{ID: PriorityIDLowest, Level: PriorityLevelLowest, Title: "Lowest"},
		{ID: PriorityIDLow, Level: PriorityLevelLow, Title: "Low"},
		{ID: PriorityIDMedium, Level: PriorityLevelMedium, Title: "Medium"},
		{ID: PriorityIDHigh, Level: PriorityLevelHigh, Title: "High"},
		{ID: PriorityIDHighest, Level: PriorityLevelHighest, Title: "Highest"},
	}

	for _, p := range priorities {
		assert.True(t, p.IsValidLevel(), "Priority %s should be valid", p.ID)
		assert.Equal(t, p.Title, p.GetDisplayName())
	}
}

func BenchmarkPriority_IsValidLevel(b *testing.B) {
	p := &Priority{Level: 3}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.IsValidLevel()
	}
}

func BenchmarkPriority_GetDisplayName(b *testing.B) {
	p := &Priority{Title: "High Priority"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.GetDisplayName()
	}
}

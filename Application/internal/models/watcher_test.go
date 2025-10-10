package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsWatching(t *testing.T) {
	watchers := []TicketWatcherMapping{
		{
			ID:       "watcher-1",
			TicketID: "ticket-1",
			UserID:   "user-1",
			Deleted:  false,
		},
		{
			ID:       "watcher-2",
			TicketID: "ticket-1",
			UserID:   "user-2",
			Deleted:  false,
		},
		{
			ID:       "watcher-3",
			TicketID: "ticket-2",
			UserID:   "user-1",
			Deleted:  false,
		},
		{
			ID:       "watcher-4",
			TicketID: "ticket-1",
			UserID:   "user-3",
			Deleted:  true, // Deleted watcher
		},
	}

	tests := []struct {
		name     string
		userID   string
		ticketID string
		expected bool
	}{
		{"User watching ticket", "user-1", "ticket-1", true},
		{"User watching different ticket", "user-1", "ticket-2", true},
		{"User not watching", "user-1", "ticket-3", false},
		{"Different user watching same ticket", "user-2", "ticket-1", true},
		{"Deleted watcher not counted", "user-3", "ticket-1", false},
		{"Non-existent user", "user-999", "ticket-1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsWatching(tt.userID, tt.ticketID, watchers)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetWatcherCount(t *testing.T) {
	watchers := []TicketWatcherMapping{
		{
			ID:       "watcher-1",
			TicketID: "ticket-1",
			UserID:   "user-1",
			Deleted:  false,
		},
		{
			ID:       "watcher-2",
			TicketID: "ticket-1",
			UserID:   "user-2",
			Deleted:  false,
		},
		{
			ID:       "watcher-3",
			TicketID: "ticket-1",
			UserID:   "user-3",
			Deleted:  true, // Should not be counted
		},
		{
			ID:       "watcher-4",
			TicketID: "ticket-2",
			UserID:   "user-1",
			Deleted:  false,
		},
	}

	tests := []struct {
		name     string
		ticketID string
		expected int
	}{
		{"Ticket with 2 watchers", "ticket-1", 2},
		{"Ticket with 1 watcher", "ticket-2", 1},
		{"Ticket with no watchers", "ticket-3", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetWatcherCount(tt.ticketID, watchers)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTicketWatcherMapping_Structure(t *testing.T) {
	mapping := TicketWatcherMapping{
		ID:       "watcher-1",
		TicketID: "ticket-1",
		UserID:   "user-1",
		Created:  1234567890,
		Deleted:  false,
	}

	assert.Equal(t, "watcher-1", mapping.ID)
	assert.Equal(t, "ticket-1", mapping.TicketID)
	assert.Equal(t, "user-1", mapping.UserID)
	assert.Equal(t, int64(1234567890), mapping.Created)
	assert.False(t, mapping.Deleted)
}

func TestIsWatching_EmptyList(t *testing.T) {
	watchers := []TicketWatcherMapping{}
	result := IsWatching("user-1", "ticket-1", watchers)
	assert.False(t, result)
}

func TestGetWatcherCount_EmptyList(t *testing.T) {
	watchers := []TicketWatcherMapping{}
	result := GetWatcherCount("ticket-1", watchers)
	assert.Equal(t, 0, result)
}

func TestIsWatching_MultipleTickets(t *testing.T) {
	watchers := []TicketWatcherMapping{
		{TicketID: "ticket-1", UserID: "user-1", Deleted: false},
		{TicketID: "ticket-2", UserID: "user-1", Deleted: false},
		{TicketID: "ticket-3", UserID: "user-1", Deleted: false},
	}

	assert.True(t, IsWatching("user-1", "ticket-1", watchers))
	assert.True(t, IsWatching("user-1", "ticket-2", watchers))
	assert.True(t, IsWatching("user-1", "ticket-3", watchers))
	assert.False(t, IsWatching("user-1", "ticket-4", watchers))
}

func BenchmarkIsWatching(b *testing.B) {
	watchers := make([]TicketWatcherMapping, 100)
	for i := 0; i < 100; i++ {
		watchers[i] = TicketWatcherMapping{
			TicketID: "ticket-1",
			UserID:   "user-1",
			Deleted:  false,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsWatching("user-1", "ticket-1", watchers)
	}
}

func BenchmarkGetWatcherCount(b *testing.B) {
	watchers := make([]TicketWatcherMapping, 100)
	for i := 0; i < 100; i++ {
		watchers[i] = TicketWatcherMapping{
			TicketID: "ticket-1",
			UserID:   "user-1",
			Deleted:  false,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetWatcherCount("ticket-1", watchers)
	}
}

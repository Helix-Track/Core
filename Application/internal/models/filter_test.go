package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterShareMapping_GetShareType(t *testing.T) {
	userID := "user-1"
	teamID := "team-1"
	projectID := "project-1"

	tests := []struct {
		name     string
		mapping  FilterShareMapping
		expected ShareType
	}{
		{
			name: "User share",
			mapping: FilterShareMapping{
				UserID: &userID,
			},
			expected: ShareTypeUser,
		},
		{
			name: "Team share",
			mapping: FilterShareMapping{
				TeamID: &teamID,
			},
			expected: ShareTypeTeam,
		},
		{
			name: "Project share",
			mapping: FilterShareMapping{
				ProjectID: &projectID,
			},
			expected: ShareTypeProject,
		},
		{
			name:     "Public share (no IDs)",
			mapping:  FilterShareMapping{},
			expected: ShareTypePublic,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.mapping.GetShareType()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFilter_IsSharedWith(t *testing.T) {
	userID1 := "user-1"
	userID2 := "user-2"
	teamID1 := "team-1"
	projectID1 := "project-1"

	shares := []FilterShareMapping{
		{
			ID:       "share-1",
			FilterID: "filter-1",
			UserID:   &userID1,
			Deleted:  false,
		},
		{
			ID:       "share-2",
			FilterID: "filter-1",
			TeamID:   &teamID1,
			Deleted:  false,
		},
		{
			ID:       "share-3",
			FilterID: "filter-1",
			UserID:   &userID2,
			Deleted:  true, // This one is deleted
		},
	}

	tests := []struct {
		name      string
		filter    Filter
		userID    string
		teamID    string
		projectID string
		expected  bool
	}{
		{
			name: "Public filter - always shared",
			filter: Filter{
				IsPublic: true,
			},
			userID:   "any-user",
			teamID:   "any-team",
			expected: true,
		},
		{
			name: "Shared with specific user",
			filter: Filter{
				IsPublic: false,
			},
			userID:   "user-1",
			teamID:   "",
			expected: true,
		},
		{
			name: "Shared with specific team",
			filter: Filter{
				IsPublic: false,
			},
			userID:   "other-user",
			teamID:   "team-1",
			expected: true,
		},
		{
			name: "Not shared with user",
			filter: Filter{
				IsPublic: false,
			},
			userID:   "unknown-user",
			teamID:   "unknown-team",
			expected: false,
		},
		{
			name: "Deleted share - not counted",
			filter: Filter{
				IsPublic: false,
			},
			userID:   "user-2", // Has deleted share
			teamID:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.filter.IsSharedWith(tt.userID, tt.teamID, tt.projectID, shares)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestShareType_Constants(t *testing.T) {
	assert.Equal(t, ShareType("user"), ShareTypeUser)
	assert.Equal(t, ShareType("team"), ShareTypeTeam)
	assert.Equal(t, ShareType("project"), ShareTypeProject)
	assert.Equal(t, ShareType("public"), ShareTypePublic)
}

func TestFilter_Structure(t *testing.T) {
	filter := Filter{
		ID:          "filter-1",
		Title:       "My Tickets",
		Description: "All tickets assigned to me",
		OwnerID:     "user-1",
		Query:       `{"assignee": "user-1"}`,
		IsPublic:    false,
		IsFavorite:  true,
		Created:     1234567890,
		Modified:    1234567890,
		Deleted:     false,
	}

	assert.Equal(t, "filter-1", filter.ID)
	assert.Equal(t, "My Tickets", filter.Title)
	assert.Equal(t, "All tickets assigned to me", filter.Description)
	assert.Equal(t, "user-1", filter.OwnerID)
	assert.Equal(t, `{"assignee": "user-1"}`, filter.Query)
	assert.False(t, filter.IsPublic)
	assert.True(t, filter.IsFavorite)
}

func TestFilterShareMapping_Structure(t *testing.T) {
	userID := "user-1"
	mapping := FilterShareMapping{
		ID:       "mapping-1",
		FilterID: "filter-1",
		UserID:   &userID,
		TeamID:   nil,
		ProjectID: nil,
		Created:  1234567890,
		Deleted:  false,
	}

	assert.Equal(t, "mapping-1", mapping.ID)
	assert.Equal(t, "filter-1", mapping.FilterID)
	assert.NotNil(t, mapping.UserID)
	assert.Equal(t, "user-1", *mapping.UserID)
	assert.Nil(t, mapping.TeamID)
	assert.Nil(t, mapping.ProjectID)
}

func TestFilter_IsSharedWith_MultipleShares(t *testing.T) {
	userID1 := "user-1"
	teamID1 := "team-1"
	projectID1 := "project-1"

	shares := []FilterShareMapping{
		{UserID: &userID1, Deleted: false},
		{TeamID: &teamID1, Deleted: false},
		{ProjectID: &projectID1, Deleted: false},
	}

	filter := Filter{IsPublic: false}

	// Should be shared with all three
	assert.True(t, filter.IsSharedWith("user-1", "", "", shares))
	assert.True(t, filter.IsSharedWith("", "team-1", "", shares))
	assert.True(t, filter.IsSharedWith("", "", "project-1", shares))
	assert.False(t, filter.IsSharedWith("other-user", "other-team", "other-project", shares))
}

func BenchmarkFilterShareMapping_GetShareType(b *testing.B) {
	userID := "user-1"
	mapping := FilterShareMapping{UserID: &userID}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mapping.GetShareType()
	}
}

func BenchmarkFilter_IsSharedWith(b *testing.B) {
	userID := "user-1"
	shares := []FilterShareMapping{
		{UserID: &userID, Deleted: false},
	}
	filter := Filter{IsPublic: false}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filter.IsSharedWith("user-1", "", "", shares)
	}
}

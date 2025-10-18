package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ================================================================
// DocumentViewHistory Tests
// ================================================================

func TestDocumentViewHistory_Validate(t *testing.T) {
	tests := []struct {
		name      string
		history   *DocumentViewHistory
		wantError bool
		errorMsg  string
	}{
		{
			name: "Valid view history",
			history: &DocumentViewHistory{
				ID:         "view-123",
				DocumentID: "doc-123",
				Timestamp:  time.Now().Unix(),
			},
			wantError: false,
		},
		{
			name: "Empty ID",
			history: &DocumentViewHistory{
				ID:         "",
				DocumentID: "doc-123",
				Timestamp:  time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "view history ID cannot be empty",
		},
		{
			name: "Empty DocumentID",
			history: &DocumentViewHistory{
				ID:         "view-123",
				DocumentID: "",
				Timestamp:  time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "view history document ID cannot be empty",
		},
		{
			name: "Zero Timestamp",
			history: &DocumentViewHistory{
				ID:         "view-123",
				DocumentID: "doc-123",
				Timestamp:  0,
			},
			wantError: true,
			errorMsg:  "view history timestamp cannot be zero",
		},
		{
			name: "Anonymous user with IP",
			history: &DocumentViewHistory{
				ID:         "view-123",
				DocumentID: "doc-123",
				UserID:     nil,
				IPAddress:  stringPtr("192.168.1.1"),
				UserAgent:  stringPtr("Mozilla/5.0"),
				Timestamp:  time.Now().Unix(),
			},
			wantError: false,
		},
		{
			name: "Authenticated user with session",
			history: &DocumentViewHistory{
				ID:           "view-123",
				DocumentID:   "doc-123",
				UserID:       stringPtr("user-123"),
				SessionID:    stringPtr("session-abc"),
				ViewDuration: intPtr(120),
				Timestamp:    time.Now().Unix(),
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.history.Validate()
			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDocumentViewHistory_SetTimestamps(t *testing.T) {
	tests := []struct {
		name    string
		history *DocumentViewHistory
		checkFn func(*testing.T, *DocumentViewHistory, int64)
	}{
		{
			name: "Set timestamp when zero",
			history: &DocumentViewHistory{
				Timestamp: 0,
			},
			checkFn: func(t *testing.T, dvh *DocumentViewHistory, before int64) {
				assert.GreaterOrEqual(t, dvh.Timestamp, before)
			},
		},
		{
			name: "Don't override existing timestamp",
			history: &DocumentViewHistory{
				Timestamp: 1234567890,
			},
			checkFn: func(t *testing.T, dvh *DocumentViewHistory, before int64) {
				assert.Equal(t, int64(1234567890), dvh.Timestamp)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now().Unix()
			tt.history.SetTimestamps()
			tt.checkFn(t, tt.history, before)
		})
	}
}

// ================================================================
// DocumentAnalytics Tests
// ================================================================

func TestDocumentAnalytics_Validate(t *testing.T) {
	tests := []struct {
		name      string
		analytics *DocumentAnalytics
		wantError bool
		errorMsg  string
	}{
		{
			name: "Valid analytics",
			analytics: &DocumentAnalytics{
				ID:              "analytics-123",
				DocumentID:      "doc-123",
				TotalViews:      100,
				UniqueViewers:   50,
				Updated:         time.Now().Unix(),
			},
			wantError: false,
		},
		{
			name: "Empty ID",
			analytics: &DocumentAnalytics{
				ID:         "",
				DocumentID: "doc-123",
				Updated:    time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "analytics ID cannot be empty",
		},
		{
			name: "Empty DocumentID",
			analytics: &DocumentAnalytics{
				ID:         "analytics-123",
				DocumentID: "",
				Updated:    time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "analytics document ID cannot be empty",
		},
		{
			name: "Zero Updated timestamp",
			analytics: &DocumentAnalytics{
				ID:         "analytics-123",
				DocumentID: "doc-123",
				Updated:    0,
			},
			wantError: true,
			errorMsg:  "analytics updated timestamp cannot be zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.analytics.Validate()
			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDocumentAnalytics_CalculatePopularityScore(t *testing.T) {
	tests := []struct {
		name      string
		analytics *DocumentAnalytics
		expected  float64
	}{
		{
			name: "Zero metrics",
			analytics: &DocumentAnalytics{
				TotalViews:     0,
				UniqueViewers:  0,
				TotalEdits:     0,
				TotalComments:  0,
				TotalReactions: 0,
				TotalWatchers:  0,
			},
			expected: 0.0,
		},
		{
			name: "Only views",
			analytics: &DocumentAnalytics{
				TotalViews:     100,
				UniqueViewers:  0,
				TotalEdits:     0,
				TotalComments:  0,
				TotalReactions: 0,
				TotalWatchers:  0,
			},
			expected: 10.0, // 100 * 0.1
		},
		{
			name: "Only unique viewers",
			analytics: &DocumentAnalytics{
				TotalViews:     0,
				UniqueViewers:  10,
				TotalEdits:     0,
				TotalComments:  0,
				TotalReactions: 0,
				TotalWatchers:  0,
			},
			expected: 3.0, // 10 * 0.3
		},
		{
			name: "Balanced metrics",
			analytics: &DocumentAnalytics{
				TotalViews:     100,
				UniqueViewers:  50,
				TotalEdits:     20,
				TotalComments:  30,
				TotalReactions: 40,
				TotalWatchers:  10,
			},
			expected: 40.0, // 100*0.1 + 50*0.3 + 20*0.2 + 30*0.2 + 40*0.1 + 10*0.1 = 10+15+4+6+4+1 = 40
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.analytics.CalculatePopularityScore()
			assert.Equal(t, tt.expected, tt.analytics.PopularityScore)
		})
	}
}

func TestDocumentAnalytics_IncrementView(t *testing.T) {
	tests := []struct {
		name           string
		initialViews   int
		initialUnique  int
		isUnique       bool
		expectedViews  int
		expectedUnique int
	}{
		{"First unique view", 0, 0, true, 1, 1},
		{"Repeat view", 5, 3, false, 6, 3},
		{"Another unique view", 10, 5, true, 11, 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analytics := &DocumentAnalytics{
				TotalViews:    tt.initialViews,
				UniqueViewers: tt.initialUnique,
			}

			analytics.IncrementView(tt.isUnique)

			assert.Equal(t, tt.expectedViews, analytics.TotalViews)
			assert.Equal(t, tt.expectedUnique, analytics.UniqueViewers)
			assert.NotNil(t, analytics.LastViewed)
			assert.Greater(t, analytics.Updated, int64(0))
		})
	}
}

func TestDocumentAnalytics_IncrementEdit(t *testing.T) {
	tests := []struct {
		name           string
		initialEdits   int
		initialUnique  int
		isUnique       bool
		expectedEdits  int
		expectedUnique int
	}{
		{"First unique edit", 0, 0, true, 1, 1},
		{"Repeat edit", 5, 3, false, 6, 3},
		{"Another unique edit", 10, 5, true, 11, 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analytics := &DocumentAnalytics{
				TotalEdits:    tt.initialEdits,
				UniqueEditors: tt.initialUnique,
			}

			analytics.IncrementEdit(tt.isUnique)

			assert.Equal(t, tt.expectedEdits, analytics.TotalEdits)
			assert.Equal(t, tt.expectedUnique, analytics.UniqueEditors)
			assert.NotNil(t, analytics.LastEdited)
			assert.Greater(t, analytics.Updated, int64(0))
		})
	}
}

func TestDocumentAnalytics_IncrementComment(t *testing.T) {
	analytics := &DocumentAnalytics{TotalComments: 5}

	analytics.IncrementComment()

	assert.Equal(t, 6, analytics.TotalComments)
	assert.Greater(t, analytics.Updated, int64(0))
}

func TestDocumentAnalytics_IncrementReaction(t *testing.T) {
	analytics := &DocumentAnalytics{TotalReactions: 10}

	analytics.IncrementReaction()

	assert.Equal(t, 11, analytics.TotalReactions)
	assert.Greater(t, analytics.Updated, int64(0))
}

func TestDocumentAnalytics_IncrementWatcher(t *testing.T) {
	analytics := &DocumentAnalytics{TotalWatchers: 3}

	analytics.IncrementWatcher()

	assert.Equal(t, 4, analytics.TotalWatchers)
	assert.Greater(t, analytics.Updated, int64(0))
}

func TestDocumentAnalytics_DecrementComment(t *testing.T) {
	tests := []struct {
		name     string
		initial  int
		expected int
	}{
		{"Normal decrement", 5, 4},
		{"At zero", 0, 0},
		{"At one", 1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analytics := &DocumentAnalytics{TotalComments: tt.initial}

			analytics.DecrementComment()

			assert.Equal(t, tt.expected, analytics.TotalComments)
			assert.Greater(t, analytics.Updated, int64(0))
		})
	}
}

func TestDocumentAnalytics_DecrementReaction(t *testing.T) {
	tests := []struct {
		name     string
		initial  int
		expected int
	}{
		{"Normal decrement", 10, 9},
		{"At zero", 0, 0},
		{"At one", 1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analytics := &DocumentAnalytics{TotalReactions: tt.initial}

			analytics.DecrementReaction()

			assert.Equal(t, tt.expected, analytics.TotalReactions)
			assert.Greater(t, analytics.Updated, int64(0))
		})
	}
}

func TestDocumentAnalytics_DecrementWatcher(t *testing.T) {
	tests := []struct {
		name     string
		initial  int
		expected int
	}{
		{"Normal decrement", 5, 4},
		{"At zero", 0, 0},
		{"At one", 1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analytics := &DocumentAnalytics{TotalWatchers: tt.initial}

			analytics.DecrementWatcher()

			assert.Equal(t, tt.expected, analytics.TotalWatchers)
			assert.Greater(t, analytics.Updated, int64(0))
		})
	}
}

func TestDocumentAnalytics_CompleteWorkflow(t *testing.T) {
	analytics := &DocumentAnalytics{
		ID:         "analytics-123",
		DocumentID: "doc-123",
	}

	// Simulate document activity
	analytics.IncrementView(true)
	analytics.IncrementView(true)
	analytics.IncrementView(false) // Repeat view

	analytics.IncrementEdit(true)
	analytics.IncrementEdit(false) // Repeat edit

	analytics.IncrementComment()
	analytics.IncrementComment()

	analytics.IncrementReaction()

	analytics.IncrementWatcher()

	// Check results
	assert.Equal(t, 3, analytics.TotalViews)
	assert.Equal(t, 2, analytics.UniqueViewers)
	assert.Equal(t, 2, analytics.TotalEdits)
	assert.Equal(t, 1, analytics.UniqueEditors)
	assert.Equal(t, 2, analytics.TotalComments)
	assert.Equal(t, 1, analytics.TotalReactions)
	assert.Equal(t, 1, analytics.TotalWatchers)

	// Calculate popularity
	analytics.CalculatePopularityScore()
	assert.Greater(t, analytics.PopularityScore, 0.0)

	// Remove some activity
	analytics.DecrementComment()
	analytics.DecrementReaction()
	analytics.DecrementWatcher()

	assert.Equal(t, 1, analytics.TotalComments)
	assert.Equal(t, 0, analytics.TotalReactions)
	assert.Equal(t, 0, analytics.TotalWatchers)
}

// ================================================================
// Benchmark Tests
// ================================================================

func BenchmarkDocumentAnalytics_CalculatePopularityScore(b *testing.B) {
	analytics := &DocumentAnalytics{
		TotalViews:     100,
		UniqueViewers:  50,
		TotalEdits:     20,
		TotalComments:  30,
		TotalReactions: 40,
		TotalWatchers:  10,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analytics.CalculatePopularityScore()
	}
}

func BenchmarkDocumentAnalytics_IncrementView(b *testing.B) {
	analytics := &DocumentAnalytics{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analytics.IncrementView(true)
	}
}

func BenchmarkDocumentAnalytics_IncrementEdit(b *testing.B) {
	analytics := &DocumentAnalytics{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analytics.IncrementEdit(true)
	}
}

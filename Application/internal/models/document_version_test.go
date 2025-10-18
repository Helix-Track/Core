package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ================================================================
// DocumentVersion Tests
// ================================================================

func TestDocumentVersion_Validate(t *testing.T) {
	tests := []struct {
		name      string
		version   *DocumentVersion
		wantError bool
		errorMsg  string
	}{
		{
			name: "Valid version",
			version: &DocumentVersion{
				ID:            "ver-123",
				DocumentID:    "doc-123",
				VersionNumber: 1,
				UserID:        "user-123",
				IsMajor:       true,
				IsMinor:       false,
				Created:       time.Now().Unix(),
			},
			wantError: false,
		},
		{
			name: "Empty ID",
			version: &DocumentVersion{
				ID:            "",
				DocumentID:    "doc-123",
				VersionNumber: 1,
				UserID:        "user-123",
				Created:       time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "document version ID cannot be empty",
		},
		{
			name: "Empty DocumentID",
			version: &DocumentVersion{
				ID:            "ver-123",
				DocumentID:    "",
				VersionNumber: 1,
				UserID:        "user-123",
				Created:       time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "document version document ID cannot be empty",
		},
		{
			name: "Version number less than 1",
			version: &DocumentVersion{
				ID:            "ver-123",
				DocumentID:    "doc-123",
				VersionNumber: 0,
				UserID:        "user-123",
				Created:       time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "version number must be at least 1",
		},
		{
			name: "Empty UserID",
			version: &DocumentVersion{
				ID:            "ver-123",
				DocumentID:    "doc-123",
				VersionNumber: 1,
				UserID:        "",
				Created:       time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "document version user ID cannot be empty",
		},
		{
			name: "Zero Created timestamp",
			version: &DocumentVersion{
				ID:            "ver-123",
				DocumentID:    "doc-123",
				VersionNumber: 1,
				UserID:        "user-123",
				Created:       0,
			},
			wantError: true,
			errorMsg:  "document version created timestamp cannot be zero",
		},
		{
			name: "Major version with all fields",
			version: &DocumentVersion{
				ID:            "ver-123",
				DocumentID:    "doc-123",
				VersionNumber: 2,
				UserID:        "user-123",
				ChangeSummary: stringPtr("Major update"),
				IsMajor:       true,
				IsMinor:       false,
				SnapshotJSON:  stringPtr(`{"title": "Test"}`),
				ContentID:     stringPtr("content-123"),
				Created:       time.Now().Unix(),
			},
			wantError: false,
		},
		{
			name: "Minor version",
			version: &DocumentVersion{
				ID:            "ver-456",
				DocumentID:    "doc-123",
				VersionNumber: 3,
				UserID:        "user-123",
				IsMajor:       false,
				IsMinor:       true,
				Created:       time.Now().Unix(),
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.version.Validate()
			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDocumentVersion_SetTimestamps(t *testing.T) {
	tests := []struct {
		name    string
		version *DocumentVersion
		checkFn func(*testing.T, *DocumentVersion, int64)
	}{
		{
			name: "Set created when zero",
			version: &DocumentVersion{
				Created: 0,
			},
			checkFn: func(t *testing.T, dv *DocumentVersion, before int64) {
				assert.GreaterOrEqual(t, dv.Created, before)
			},
		},
		{
			name: "Don't override existing created",
			version: &DocumentVersion{
				Created: 1234567890,
			},
			checkFn: func(t *testing.T, dv *DocumentVersion, before int64) {
				assert.Equal(t, int64(1234567890), dv.Created)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now().Unix()
			tt.version.SetTimestamps()
			tt.checkFn(t, tt.version, before)
		})
	}
}

// ================================================================
// DocumentVersionLabel Tests
// ================================================================

func TestDocumentVersionLabel_Validate(t *testing.T) {
	tests := []struct {
		name      string
		label     *DocumentVersionLabel
		wantError bool
		errorMsg  string
	}{
		{
			name: "Valid label",
			label: &DocumentVersionLabel{
				ID:        "label-123",
				VersionID: "ver-123",
				Label:     "Release 1.0",
				UserID:    "user-123",
				Created:   time.Now().Unix(),
			},
			wantError: false,
		},
		{
			name: "Empty ID",
			label: &DocumentVersionLabel{
				ID:        "",
				VersionID: "ver-123",
				Label:     "Release 1.0",
				UserID:    "user-123",
				Created:   time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "version label ID cannot be empty",
		},
		{
			name: "Empty VersionID",
			label: &DocumentVersionLabel{
				ID:        "label-123",
				VersionID: "",
				Label:     "Release 1.0",
				UserID:    "user-123",
				Created:   time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "version label version ID cannot be empty",
		},
		{
			name: "Empty Label",
			label: &DocumentVersionLabel{
				ID:        "label-123",
				VersionID: "ver-123",
				Label:     "",
				UserID:    "user-123",
				Created:   time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "version label cannot be empty",
		},
		{
			name: "Empty UserID",
			label: &DocumentVersionLabel{
				ID:        "label-123",
				VersionID: "ver-123",
				Label:     "Release 1.0",
				UserID:    "",
				Created:   time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "version label user ID cannot be empty",
		},
		{
			name: "Zero Created timestamp",
			label: &DocumentVersionLabel{
				ID:        "label-123",
				VersionID: "ver-123",
				Label:     "Release 1.0",
				UserID:    "user-123",
				Created:   0,
			},
			wantError: true,
			errorMsg:  "version label created timestamp cannot be zero",
		},
		{
			name: "With description",
			label: &DocumentVersionLabel{
				ID:          "label-123",
				VersionID:   "ver-123",
				Label:       "Release 1.0",
				Description: stringPtr("First stable release"),
				UserID:      "user-123",
				Created:     time.Now().Unix(),
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.label.Validate()
			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ================================================================
// DocumentVersionTag Tests
// ================================================================

func TestDocumentVersionTag_Validate(t *testing.T) {
	tests := []struct {
		name      string
		tag       *DocumentVersionTag
		wantError bool
		errorMsg  string
	}{
		{
			name: "Valid tag",
			tag: &DocumentVersionTag{
				ID:        "tag-123",
				VersionID: "ver-123",
				Tag:       "v1.0.0",
				UserID:    "user-123",
				Created:   time.Now().Unix(),
			},
			wantError: false,
		},
		{
			name: "Empty ID",
			tag: &DocumentVersionTag{
				ID:        "",
				VersionID: "ver-123",
				Tag:       "v1.0.0",
				UserID:    "user-123",
				Created:   time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "version tag ID cannot be empty",
		},
		{
			name: "Empty VersionID",
			tag: &DocumentVersionTag{
				ID:        "tag-123",
				VersionID: "",
				Tag:       "v1.0.0",
				UserID:    "user-123",
				Created:   time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "version tag version ID cannot be empty",
		},
		{
			name: "Empty Tag",
			tag: &DocumentVersionTag{
				ID:        "tag-123",
				VersionID: "ver-123",
				Tag:       "",
				UserID:    "user-123",
				Created:   time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "version tag cannot be empty",
		},
		{
			name: "Empty UserID",
			tag: &DocumentVersionTag{
				ID:        "tag-123",
				VersionID: "ver-123",
				Tag:       "v1.0.0",
				UserID:    "",
				Created:   time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "version tag user ID cannot be empty",
		},
		{
			name: "Zero Created timestamp",
			tag: &DocumentVersionTag{
				ID:        "tag-123",
				VersionID: "ver-123",
				Tag:       "v1.0.0",
				UserID:    "user-123",
				Created:   0,
			},
			wantError: true,
			errorMsg:  "version tag created timestamp cannot be zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tag.Validate()
			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ================================================================
// DocumentVersionComment Tests
// ================================================================

func TestDocumentVersionComment_Validate(t *testing.T) {
	tests := []struct {
		name      string
		comment   *DocumentVersionComment
		wantError bool
		errorMsg  string
	}{
		{
			name: "Valid comment",
			comment: &DocumentVersionComment{
				ID:        "comment-123",
				VersionID: "ver-123",
				UserID:    "user-123",
				Comment:   "Great changes!",
				Created:   time.Now().Unix(),
				Modified:  time.Now().Unix(),
			},
			wantError: false,
		},
		{
			name: "Empty ID",
			comment: &DocumentVersionComment{
				ID:        "",
				VersionID: "ver-123",
				UserID:    "user-123",
				Comment:   "Great changes!",
				Created:   time.Now().Unix(),
				Modified:  time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "version comment ID cannot be empty",
		},
		{
			name: "Empty VersionID",
			comment: &DocumentVersionComment{
				ID:        "comment-123",
				VersionID: "",
				UserID:    "user-123",
				Comment:   "Great changes!",
				Created:   time.Now().Unix(),
				Modified:  time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "version comment version ID cannot be empty",
		},
		{
			name: "Empty UserID",
			comment: &DocumentVersionComment{
				ID:        "comment-123",
				VersionID: "ver-123",
				UserID:    "",
				Comment:   "Great changes!",
				Created:   time.Now().Unix(),
				Modified:  time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "version comment user ID cannot be empty",
		},
		{
			name: "Empty Comment",
			comment: &DocumentVersionComment{
				ID:        "comment-123",
				VersionID: "ver-123",
				UserID:    "user-123",
				Comment:   "",
				Created:   time.Now().Unix(),
				Modified:  time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "version comment cannot be empty",
		},
		{
			name: "Zero Created timestamp",
			comment: &DocumentVersionComment{
				ID:        "comment-123",
				VersionID: "ver-123",
				UserID:    "user-123",
				Comment:   "Great changes!",
				Created:   0,
				Modified:  time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "version comment created timestamp cannot be zero",
		},
		{
			name: "Zero Modified timestamp",
			comment: &DocumentVersionComment{
				ID:        "comment-123",
				VersionID: "ver-123",
				UserID:    "user-123",
				Comment:   "Great changes!",
				Created:   time.Now().Unix(),
				Modified:  0,
			},
			wantError: true,
			errorMsg:  "version comment modified timestamp cannot be zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.comment.Validate()
			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDocumentVersionComment_SetTimestamps(t *testing.T) {
	tests := []struct {
		name    string
		comment *DocumentVersionComment
		checkFn func(*testing.T, *DocumentVersionComment, int64)
	}{
		{
			name: "Set both timestamps when zero",
			comment: &DocumentVersionComment{
				Created:  0,
				Modified: 0,
			},
			checkFn: func(t *testing.T, dvc *DocumentVersionComment, before int64) {
				assert.GreaterOrEqual(t, dvc.Created, before)
				assert.GreaterOrEqual(t, dvc.Modified, before)
				assert.Equal(t, dvc.Created, dvc.Modified)
			},
		},
		{
			name: "Only update modified when created exists",
			comment: &DocumentVersionComment{
				Created:  1234567890,
				Modified: 0,
			},
			checkFn: func(t *testing.T, dvc *DocumentVersionComment, before int64) {
				assert.Equal(t, int64(1234567890), dvc.Created)
				assert.GreaterOrEqual(t, dvc.Modified, before)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now().Unix()
			tt.comment.SetTimestamps()
			tt.checkFn(t, tt.comment, before)
		})
	}
}

// ================================================================
// DocumentVersionMention Tests
// ================================================================

func TestDocumentVersionMention_Validate(t *testing.T) {
	tests := []struct {
		name      string
		mention   *DocumentVersionMention
		wantError bool
		errorMsg  string
	}{
		{
			name: "Valid mention",
			mention: &DocumentVersionMention{
				ID:               "mention-123",
				VersionID:        "ver-123",
				MentionedUserID:  "user-456",
				MentioningUserID: "user-123",
				Created:          time.Now().Unix(),
			},
			wantError: false,
		},
		{
			name: "Empty ID",
			mention: &DocumentVersionMention{
				ID:               "",
				VersionID:        "ver-123",
				MentionedUserID:  "user-456",
				MentioningUserID: "user-123",
				Created:          time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "version mention ID cannot be empty",
		},
		{
			name: "Empty VersionID",
			mention: &DocumentVersionMention{
				ID:               "mention-123",
				VersionID:        "",
				MentionedUserID:  "user-456",
				MentioningUserID: "user-123",
				Created:          time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "version mention version ID cannot be empty",
		},
		{
			name: "Empty MentionedUserID",
			mention: &DocumentVersionMention{
				ID:               "mention-123",
				VersionID:        "ver-123",
				MentionedUserID:  "",
				MentioningUserID: "user-123",
				Created:          time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "mentioned user ID cannot be empty",
		},
		{
			name: "Empty MentioningUserID",
			mention: &DocumentVersionMention{
				ID:               "mention-123",
				VersionID:        "ver-123",
				MentionedUserID:  "user-456",
				MentioningUserID: "",
				Created:          time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "mentioning user ID cannot be empty",
		},
		{
			name: "Zero Created timestamp",
			mention: &DocumentVersionMention{
				ID:               "mention-123",
				VersionID:        "ver-123",
				MentionedUserID:  "user-456",
				MentioningUserID: "user-123",
				Created:          0,
			},
			wantError: true,
			errorMsg:  "version mention created timestamp cannot be zero",
		},
		{
			name: "With context",
			mention: &DocumentVersionMention{
				ID:               "mention-123",
				VersionID:        "ver-123",
				MentionedUserID:  "user-456",
				MentioningUserID: "user-123",
				Context:          stringPtr("Check this version"),
				Created:          time.Now().Unix(),
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.mention.Validate()
			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ================================================================
// DocumentVersionDiff Tests
// ================================================================

func TestDocumentVersionDiff_Validate(t *testing.T) {
	tests := []struct {
		name      string
		diff      *DocumentVersionDiff
		wantError bool
		errorMsg  string
	}{
		{
			name: "Valid diff",
			diff: &DocumentVersionDiff{
				ID:          "diff-123",
				DocumentID:  "doc-123",
				FromVersion: 1,
				ToVersion:   2,
				DiffType:    "unified",
				DiffContent: "- old line\n+ new line",
				Created:     time.Now().Unix(),
			},
			wantError: false,
		},
		{
			name: "Empty ID",
			diff: &DocumentVersionDiff{
				ID:          "",
				DocumentID:  "doc-123",
				FromVersion: 1,
				ToVersion:   2,
				DiffType:    "unified",
				DiffContent: "- old line\n+ new line",
				Created:     time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "version diff ID cannot be empty",
		},
		{
			name: "Empty DocumentID",
			diff: &DocumentVersionDiff{
				ID:          "diff-123",
				DocumentID:  "",
				FromVersion: 1,
				ToVersion:   2,
				DiffType:    "unified",
				DiffContent: "- old line\n+ new line",
				Created:     time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "version diff document ID cannot be empty",
		},
		{
			name: "FromVersion less than 1",
			diff: &DocumentVersionDiff{
				ID:          "diff-123",
				DocumentID:  "doc-123",
				FromVersion: 0,
				ToVersion:   2,
				DiffType:    "unified",
				DiffContent: "- old line\n+ new line",
				Created:     time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "from version must be at least 1",
		},
		{
			name: "ToVersion less than 1",
			diff: &DocumentVersionDiff{
				ID:          "diff-123",
				DocumentID:  "doc-123",
				FromVersion: 1,
				ToVersion:   0,
				DiffType:    "unified",
				DiffContent: "- old line\n+ new line",
				Created:     time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "to version must be at least 1",
		},
		{
			name: "FromVersion >= ToVersion",
			diff: &DocumentVersionDiff{
				ID:          "diff-123",
				DocumentID:  "doc-123",
				FromVersion: 2,
				ToVersion:   2,
				DiffType:    "unified",
				DiffContent: "- old line\n+ new line",
				Created:     time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "from version must be less than to version",
		},
		{
			name: "Empty DiffType",
			diff: &DocumentVersionDiff{
				ID:          "diff-123",
				DocumentID:  "doc-123",
				FromVersion: 1,
				ToVersion:   2,
				DiffType:    "",
				DiffContent: "- old line\n+ new line",
				Created:     time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "diff type cannot be empty",
		},
		{
			name: "Invalid DiffType",
			diff: &DocumentVersionDiff{
				ID:          "diff-123",
				DocumentID:  "doc-123",
				FromVersion: 1,
				ToVersion:   2,
				DiffType:    "json",
				DiffContent: "- old line\n+ new line",
				Created:     time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "invalid diff type",
		},
		{
			name: "Empty DiffContent",
			diff: &DocumentVersionDiff{
				ID:          "diff-123",
				DocumentID:  "doc-123",
				FromVersion: 1,
				ToVersion:   2,
				DiffType:    "unified",
				DiffContent: "",
				Created:     time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "diff content cannot be empty",
		},
		{
			name: "Zero Created timestamp",
			diff: &DocumentVersionDiff{
				ID:          "diff-123",
				DocumentID:  "doc-123",
				FromVersion: 1,
				ToVersion:   2,
				DiffType:    "unified",
				DiffContent: "- old line\n+ new line",
				Created:     0,
			},
			wantError: true,
			errorMsg:  "version diff created timestamp cannot be zero",
		},
		{
			name: "Split diff type",
			diff: &DocumentVersionDiff{
				ID:          "diff-123",
				DocumentID:  "doc-123",
				FromVersion: 1,
				ToVersion:   2,
				DiffType:    "split",
				DiffContent: "left|right",
				Created:     time.Now().Unix(),
			},
			wantError: false,
		},
		{
			name: "HTML diff type",
			diff: &DocumentVersionDiff{
				ID:          "diff-123",
				DocumentID:  "doc-123",
				FromVersion: 1,
				ToVersion:   2,
				DiffType:    "html",
				DiffContent: "<div>diff</div>",
				Created:     time.Now().Unix(),
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.diff.Validate()
			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDocumentVersionDiff_AllDiffTypes(t *testing.T) {
	diffTypes := []string{"unified", "split", "html"}

	for _, diffType := range diffTypes {
		t.Run("DiffType: "+diffType, func(t *testing.T) {
			diff := &DocumentVersionDiff{
				ID:          "diff-123",
				DocumentID:  "doc-123",
				FromVersion: 1,
				ToVersion:   2,
				DiffType:    diffType,
				DiffContent: "test diff",
				Created:     time.Now().Unix(),
			}

			err := diff.Validate()
			assert.NoError(t, err)
		})
	}
}

// ================================================================
// Benchmark Tests
// ================================================================

func BenchmarkDocumentVersion_Validate(b *testing.B) {
	version := &DocumentVersion{
		ID:            "ver-123",
		DocumentID:    "doc-123",
		VersionNumber: 1,
		UserID:        "user-123",
		Created:       time.Now().Unix(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = version.Validate()
	}
}

func BenchmarkDocumentVersionDiff_Validate(b *testing.B) {
	diff := &DocumentVersionDiff{
		ID:          "diff-123",
		DocumentID:  "doc-123",
		FromVersion: 1,
		ToVersion:   2,
		DiffType:    "unified",
		DiffContent: "test",
		Created:     time.Now().Unix(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = diff.Validate()
	}
}

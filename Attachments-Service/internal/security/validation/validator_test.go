package validation

import (
	"strings"
	"testing"
)

func TestNewValidator(t *testing.T) {
	t.Run("with nil config uses defaults", func(t *testing.T) {
		v := NewValidator(nil)
		if v == nil {
			t.Fatal("expected validator, got nil")
		}
		if v.config == nil {
			t.Fatal("expected default config")
		}
	})

	t.Run("with custom config", func(t *testing.T) {
		config := &ValidationConfig{
			MaxFilenameLength: 100,
		}
		v := NewValidator(config)
		if v.config.MaxFilenameLength != 100 {
			t.Errorf("expected MaxFilenameLength 100, got %d", v.config.MaxFilenameLength)
		}
	})
}

func TestValidateFilename(t *testing.T) {
	v := NewValidator(nil)

	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{"valid filename", "document.pdf", false},
		{"with spaces", "my document.pdf", false},
		{"with underscore", "my_document.pdf", false},
		{"with dash", "my-document.pdf", false},
		{"empty filename", "", true},
		{"too long", strings.Repeat("a", 300), true},
		{"with null byte", "doc\x00.pdf", true},
		{"forbidden filename", "CON", true},
		{"forbidden PRN", "PRN", true},
		{"forbidden AUX", "AUX", true},
		{"with directory", "../../../etc/passwd", false}, // Should be sanitized
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := v.ValidateFilename(tt.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFilename() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	v := NewValidator(nil)

	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{"clean filename", "document.pdf", "document.pdf"},
		{"with slash", "dir/file.pdf", "dir_file.pdf"},
		{"with backslash", "dir\\file.pdf", "dir_file.pdf"},
		{"with double dot", "file..pdf", "file__pdf"},
		{"leading spaces", "  file.pdf", "file.pdf"},
		{"trailing spaces", "file.pdf  ", "file.pdf"},
		{"multiple underscores", "file___name.pdf", "file_name.pdf"},
		{"special chars", "file@#$.pdf", "file___.pdf"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := v.SanitizeFilename(tt.filename)
			if got != tt.want {
				t.Errorf("SanitizeFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidatePath(t *testing.T) {
	v := NewValidator(nil)

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"valid relative path", "files/document.pdf", false},
		{"empty path", "", true},
		{"with null byte", "file\x00.pdf", true},
		{"absolute path", "/etc/passwd", true},
		{"with traversal", "../../../etc/passwd", true},
		{"clean traversal", "../../file.pdf", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidatePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateEntityType(t *testing.T) {
	v := NewValidator(nil)

	tests := []struct {
		name       string
		entityType string
		wantErr    bool
	}{
		{"valid ticket", "ticket", false},
		{"valid project", "project", false},
		{"empty", "", true},
		{"too long", strings.Repeat("a", 100), true},
		{"with special chars", "ticket@123", true},
		{"not in whitelist", "invalid_type", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateEntityType(tt.entityType)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEntityType() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateEntityID(t *testing.T) {
	v := NewValidator(nil)

	tests := []struct {
		name     string
		entityID string
		wantErr  bool
	}{
		{"valid ID", "TICKET-123", false},
		{"with underscore", "project_456", false},
		{"alphanumeric", "abc123def", false},
		{"empty", "", true},
		{"too long", strings.Repeat("a", 200), true},
		{"with null byte", "id\x00", true},
		{"with special chars", "id@#$", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateEntityID(tt.entityID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEntityID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateUserID(t *testing.T) {
	v := NewValidator(nil)

	tests := []struct {
		name    string
		userID  string
		wantErr bool
	}{
		{"valid user", "user123", false},
		{"with underscore", "user_123", false},
		{"with dash", "user-123", false},
		{"empty", "", true},
		{"too short", "ab", true},
		{"too long", strings.Repeat("a", 200), true},
		{"with special chars", "user@domain", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateUserID(tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUserID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateDescription(t *testing.T) {
	v := NewValidator(nil)

	tests := []struct {
		name        string
		description string
		wantErr     bool
	}{
		{"valid description", "This is a valid description", false},
		{"empty", "", false}, // Empty is valid
		{"too long", strings.Repeat("a", 6000), true},
		{"with null byte", "desc\x00", true},
		{"with newlines", "line1\nline2\nline3", false},
		{"with unicode", "Description with émojis 😀", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateDescription(tt.description)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDescription() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateTags(t *testing.T) {
	v := NewValidator(nil)

	tests := []struct {
		name    string
		tags    []string
		wantErr bool
	}{
		{"valid tags", []string{"tag1", "tag2"}, false},
		{"empty list", []string{}, false},
		{"too many tags", make([]string, 25), true},
		{"empty tag", []string{"tag1", "", "tag2"}, true},
		{"tag too long", []string{strings.Repeat("a", 100)}, true},
		{"with special chars", []string{"tag@123"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateTags(tt.tags)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSanitizeTags(t *testing.T) {
	v := NewValidator(nil)

	tests := []struct {
		name string
		tags []string
		want []string
	}{
		{
			name: "clean tags",
			tags: []string{"tag1", "tag2"},
			want: []string{"tag1", "tag2"},
		},
		{
			name: "with spaces",
			tags: []string{" tag1 ", "tag2"},
			want: []string{"tag1", "tag2"},
		},
		{
			name: "uppercase to lowercase",
			tags: []string{"TAG1", "Tag2"},
			want: []string{"tag1", "tag2"},
		},
		{
			name: "remove special chars",
			tags: []string{"tag@1", "tag#2"},
			want: []string{"tag1", "tag2"},
		},
		{
			name: "empty tags removed",
			tags: []string{"tag1", "", "tag2"},
			want: []string{"tag1", "tag2"},
		},
		{
			name: "limit to max",
			tags: make([]string, 25),
			want: make([]string, 20),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := v.SanitizeTags(tt.tags)
			if len(got) != len(tt.want) {
				t.Errorf("SanitizeTags() length = %v, want %v", len(got), len(tt.want))
			}
		})
	}
}

func TestValidateHash(t *testing.T) {
	v := NewValidator(nil)

	tests := []struct {
		name    string
		hash    string
		wantErr bool
	}{
		{"valid SHA-256", strings.Repeat("a", 64), false},
		{"empty", "", true},
		{"too short", strings.Repeat("a", 32), true},
		{"too long", strings.Repeat("a", 128), true},
		{"non-hex", strings.Repeat("z", 64), true},
		{"mixed valid", "abcdef0123456789" + strings.Repeat("0", 48), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateHash(tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateHash() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateReferenceID(t *testing.T) {
	v := NewValidator(nil)

	tests := []struct {
		name        string
		referenceID string
		wantErr     bool
	}{
		{"valid UUID", "550e8400-e29b-41d4-a716-446655440000", false},
		{"empty", "", true},
		{"too short", "550e8400", true},
		{"too long", "550e8400-e29b-41d4-a716-446655440000-extra", true},
		{"invalid format", "550e8400-e29b-41d4-a716-44665544000z", true},
		{"no dashes", "550e8400e29b41d4a716446655440000", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateReferenceID(tt.referenceID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateReferenceID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateMimeType(t *testing.T) {
	v := NewValidator(nil)

	tests := []struct {
		name     string
		mimeType string
		wantErr  bool
	}{
		{"valid image", "image/jpeg", false},
		{"valid application", "application/pdf", false},
		{"valid text", "text/plain", false},
		{"with parameter", "text/html; charset=utf-8", false},
		{"empty", "", true},
		{"no subtype", "image/", true},
		{"no type", "/jpeg", true},
		{"invalid format", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateMimeType(tt.mimeType)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMimeType() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	v := NewValidator(nil)

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"valid http", "http://example.com", false},
		{"valid https", "https://example.com", false},
		{"empty", "", true},
		{"javascript protocol", "javascript:alert('xss')", true},
		{"data protocol", "data:text/html,<script>alert('xss')</script>", true},
		{"file protocol", "file:///etc/passwd", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSanitizeString(t *testing.T) {
	v := NewValidator(nil)

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"clean string", "hello world", "hello world"},
		{"with null byte", "hello\x00world", "helloworld"},
		{"with control chars", "hello\x01\x02world", "helloworld"},
		{"with newline", "hello\nworld", "hello\nworld"},
		{"with tab", "hello\tworld", "hello\tworld"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := v.SanitizeString(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkValidateFilename(b *testing.B) {
	v := NewValidator(nil)
	filename := "my_document-v2.pdf"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.ValidateFilename(filename)
	}
}

func BenchmarkSanitizeFilename(b *testing.B) {
	v := NewValidator(nil)
	filename := "../../my_document@#$.pdf"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.SanitizeFilename(filename)
	}
}

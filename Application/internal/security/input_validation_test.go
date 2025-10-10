package security

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestValidateString(t *testing.T) {
	cfg := DefaultInputValidationConfig()

	tests := []struct {
		name          string
		input         string
		expectedValid bool
		expectedReason string
	}{
		{
			name:          "Valid string",
			input:         "Hello World",
			expectedValid: true,
		},
		{
			name:          "SQL injection - UNION SELECT",
			input:         "'; UNION SELECT * FROM users--",
			expectedValid: false,
			expectedReason: "Potential SQL injection detected",
		},
		{
			name:          "SQL injection - DROP TABLE",
			input:         "Robert'; DROP TABLE students;--",
			expectedValid: false,
			expectedReason: "Potential SQL injection detected",
		},
		{
			name:          "SQL injection - OR 1=1",
			input:         "admin' OR '1'='1",
			expectedValid: false,
			expectedReason: "Potential SQL injection detected",
		},
		{
			name:          "XSS - script tag",
			input:         "<script>alert('XSS')</script>",
			expectedValid: false,
			expectedReason: "Potential XSS attack detected",
		},
		{
			name:          "XSS - javascript protocol",
			input:         "<a href='javascript:alert(1)'>Click</a>",
			expectedValid: false,
			expectedReason: "Potential XSS attack detected",
		},
		{
			name:          "XSS - onerror",
			input:         "<img src=x onerror=alert(1)>",
			expectedValid: false,
			expectedReason: "Potential XSS attack detected",
		},
		{
			name:          "Path traversal - ../",
			input:         "../../etc/passwd",
			expectedValid: false,
			expectedReason: "Potential path traversal detected",
		},
		{
			name:          "Path traversal - URL encoded",
			input:         "%2e%2e%2f%2e%2e%2f",
			expectedValid: false,
			expectedReason: "Potential path traversal detected",
		},
		{
			name:          "Command injection - semicolon",
			input:         "test; rm -rf /",
			expectedValid: false,
			expectedReason: "Potential command injection detected",
		},
		{
			name:          "Command injection - pipe",
			input:         "test | cat /etc/passwd",
			expectedValid: false,
			expectedReason: "Potential command injection detected",
		},
		{
			name:          "Command injection - backticks",
			input:         "test `whoami`",
			expectedValid: false,
			expectedReason: "Potential command injection detected",
		},
		{
			name:          "LDAP injection - wildcard",
			input:         "admin*",
			expectedValid: false,
			expectedReason: "Potential LDAP injection detected",
		},
		{
			name:          "String too long",
			input:         string(make([]byte, 20000)),
			expectedValid: false,
			expectedReason: "String exceeds maximum length of 10000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, _, reason := ValidateString(tt.input, cfg)
			assert.Equal(t, tt.expectedValid, valid)
			if !tt.expectedValid {
				assert.Contains(t, reason, tt.expectedReason)
			}
		})
	}
}

func TestSQLInjectionPattern(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"UNION SELECT", "' UNION SELECT * FROM users", true},
		{"SELECT FROM", "SELECT * FROM table", true},
		{"INSERT INTO", "INSERT INTO users VALUES", true},
		{"DELETE FROM", "DELETE FROM users WHERE", true},
		{"DROP TABLE", "DROP TABLE students", true},
		{"OR 1=1", "admin' OR 1=1--", true},
		{"Normal text", "Hello World", false},
		{"Normal query without keywords", "user data", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SQLInjectionPattern(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestXSSPattern(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"script tag", "<script>alert(1)</script>", true},
		{"javascript protocol", "javascript:alert(1)", true},
		{"onerror", "<img onerror=alert(1)>", true},
		{"onload", "<body onload=alert(1)>", true},
		{"iframe", "<iframe src=evil.com></iframe>", true},
		{"eval", "eval(code)", true},
		{"Normal HTML", "<p>Hello World</p>", false},
		{"Normal text", "Hello World", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := XSSPattern(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPathTraversalPattern(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Unix path traversal", "../../etc/passwd", true},
		{"Windows path traversal", "..\\..\\windows\\system32", true},
		{"URL encoded", "%2e%2e%2fetc%2fpasswd", true},
		{"Normal path", "/home/user/file.txt", false},
		{"Normal text", "file.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PathTraversalPattern(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "HTML characters",
			input:    "<script>alert('XSS')</script>",
			expected: "&lt;script&gt;alert(&#39;XSS&#39;)&lt;/script&gt;",
		},
		{
			name:     "Null bytes",
			input:    "test\x00data",
			expected: "testdata",
		},
		{
			name:     "Whitespace",
			input:    "  trimmed  ",
			expected: "trimmed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeInput(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Path separators",
			input:    "../../etc/passwd",
			expected: "etcpasswd",
		},
		{
			name:     "Windows path",
			input:    "C:\\Windows\\System32\\cmd.exe",
			expected: "CWindowsSystem32cmd.exe",
		},
		{
			name:     "Dangerous characters",
			input:    "file<>:\"|?*.txt",
			expected: "file.txt",
		},
		{
			name:     "Normal filename",
			input:    "document.pdf",
			expected: "document.pdf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeFilename(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeURL(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedErr bool
	}{
		{
			name:        "Valid HTTP URL",
			input:       "http://example.com",
			expectedErr: false,
		},
		{
			name:        "Valid HTTPS URL",
			input:       "https://example.com",
			expectedErr: false,
		},
		{
			name:        "JavaScript protocol",
			input:       "javascript:alert(1)",
			expectedErr: true,
		},
		{
			name:        "Data protocol",
			input:       "data:text/html,<script>alert(1)</script>",
			expectedErr: true,
		},
		{
			name:        "FTP protocol",
			input:       "ftp://example.com",
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := SanitizeURL(tt.input)
			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected bool
	}{
		{"Valid email", "user@example.com", true},
		{"Valid email with plus", "user+tag@example.com", true},
		{"Valid email with subdomain", "user@mail.example.com", true},
		{"Invalid - no @", "userexample.com", false},
		{"Invalid - no domain", "user@", false},
		{"Invalid - no TLD", "user@example", false},
		{"Invalid - spaces", "user @example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateEmail(tt.email)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name          string
		username      string
		expectedValid bool
	}{
		{"Valid username", "user123", true},
		{"Valid with underscore", "user_name", true},
		{"Too short", "ab", false},
		{"Too long", string(make([]byte, 51)), false},
		{"Invalid characters - space", "user name", false},
		{"Invalid characters - special", "user@name", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, _ := ValidateUsername(tt.username)
			assert.Equal(t, tt.expectedValid, valid)
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name          string
		password      string
		expectedValid bool
	}{
		{"Valid password", "Password123!", true},
		{"Valid complex", "MyP@ssw0rd2023", true},
		{"Too short", "Pass1!", false},
		{"No uppercase", "password123!", false},
		{"No lowercase", "PASSWORD123!", false},
		{"No digit", "Password!", false},
		{"No special char", "Password123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, _ := ValidatePassword(tt.password)
			assert.Equal(t, tt.expectedValid, valid)
		})
	}
}

func TestInputValidationMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		queryParams    map[string]string
		expectedStatus int
	}{
		{
			name: "Valid query parameters",
			queryParams: map[string]string{
				"name": "John",
				"page": "1",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "SQL injection in query",
			queryParams: map[string]string{
				"id": "1 UNION SELECT * FROM users",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "XSS in query",
			queryParams: map[string]string{
				"search": "<script>alert(1)</script>",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Build request URL with query parameters
			req := httptest.NewRequest("GET", "/test", nil)
			q := req.URL.Query()
			for key, value := range tt.queryParams {
				q.Add(key, value)
			}
			req.URL.RawQuery = q.Encode()
			c.Request = req

			// Create middleware
			middleware := InputValidationMiddleware(DefaultInputValidationConfig())

			// Execute middleware
			middleware(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func BenchmarkValidateString(b *testing.B) {
	cfg := DefaultInputValidationConfig()
	input := "This is a normal string without any malicious content"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateString(input, cfg)
	}
}

func BenchmarkSQLInjectionPattern(b *testing.B) {
	input := "SELECT * FROM users WHERE id = 1"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SQLInjectionPattern(input)
	}
}

func BenchmarkXSSPattern(b *testing.B) {
	input := "<script>alert('XSS')</script>"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		XSSPattern(input)
	}
}

func BenchmarkSanitizeInput(b *testing.B) {
	input := "<script>alert('XSS')</script>"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SanitizeInput(input)
	}
}

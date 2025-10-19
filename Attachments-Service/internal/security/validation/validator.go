package validation

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Validator handles input validation and sanitization
type Validator struct {
	config *ValidationConfig
}

// ValidationConfig contains validation rules
type ValidationConfig struct {
	// Filename validation
	MaxFilenameLength int
	AllowedFilenameChars string
	ForbiddenFilenames []string

	// Entity validation
	MaxEntityTypeLength int
	MaxEntityIDLength int
	AllowedEntityTypes []string

	// User validation
	MaxUserIDLength int
	MinUserIDLength int

	// Description validation
	MaxDescriptionLength int

	// Tag validation
	MaxTagLength int
	MaxTagsPerFile int

	// Path validation
	AllowAbsolutePaths bool
	AllowPathTraversal bool
}

// DefaultValidationConfig returns default validation configuration
func DefaultValidationConfig() *ValidationConfig {
	return &ValidationConfig{
		MaxFilenameLength: 255,
		AllowedFilenameChars: "a-zA-Z0-9._ -",
		ForbiddenFilenames: []string{
			".", "..", "CON", "PRN", "AUX", "NUL",
			"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
			"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
		},

		MaxEntityTypeLength: 50,
		MaxEntityIDLength: 100,
		AllowedEntityTypes: []string{
			"ticket", "project", "epic", "sprint", "board",
			"comment", "attachment", "user", "team", "organization",
		},

		MaxUserIDLength: 100,
		MinUserIDLength: 3,

		MaxDescriptionLength: 5000,

		MaxTagLength: 50,
		MaxTagsPerFile: 20,

		AllowAbsolutePaths: false,
		AllowPathTraversal: false,
	}
}

// NewValidator creates a new input validator
func NewValidator(config *ValidationConfig) *Validator {
	if config == nil {
		config = DefaultValidationConfig()
	}

	return &Validator{
		config: config,
	}
}

// ValidateFilename validates and sanitizes a filename
func (v *Validator) ValidateFilename(filename string) (string, error) {
	if filename == "" {
		return "", fmt.Errorf("filename cannot be empty")
	}

	// Check length
	if len(filename) > v.config.MaxFilenameLength {
		return "", fmt.Errorf("filename exceeds maximum length of %d characters", v.config.MaxFilenameLength)
	}

	// Check for null bytes
	if strings.Contains(filename, "\x00") {
		return "", fmt.Errorf("filename contains null bytes")
	}

	// Extract base filename (without directory)
	base := filepath.Base(filename)

	// Check for forbidden filenames
	baseUpper := strings.ToUpper(base)
	for _, forbidden := range v.config.ForbiddenFilenames {
		if baseUpper == forbidden {
			return "", fmt.Errorf("filename '%s' is forbidden", base)
		}
	}

	// Sanitize filename
	sanitized := v.SanitizeFilename(base)

	if sanitized == "" {
		return "", fmt.Errorf("filename contains only invalid characters")
	}

	return sanitized, nil
}

// SanitizeFilename removes dangerous characters from filename
func (v *Validator) SanitizeFilename(filename string) string {
	// Remove directory separators
	filename = strings.ReplaceAll(filename, "/", "_")
	filename = strings.ReplaceAll(filename, "\\", "_")

	// Remove other dangerous characters
	filename = strings.ReplaceAll(filename, "..", "_")
	filename = strings.ReplaceAll(filename, "\x00", "")

	// Build allowed pattern
	pattern := fmt.Sprintf("[^%s]", v.config.AllowedFilenameChars)
	re := regexp.MustCompile(pattern)
	filename = re.ReplaceAllString(filename, "_")

	// Remove leading/trailing spaces and dots
	filename = strings.Trim(filename, " .")

	// Collapse multiple underscores
	re = regexp.MustCompile("__+")
	filename = re.ReplaceAllString(filename, "_")

	return filename
}

// ValidatePath validates a file path
func (v *Validator) ValidatePath(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Check for null bytes
	if strings.Contains(path, "\x00") {
		return fmt.Errorf("path contains null bytes")
	}

	// Check for absolute paths
	if !v.config.AllowAbsolutePaths && filepath.IsAbs(path) {
		return fmt.Errorf("absolute paths are not allowed")
	}

	// Check for path traversal
	if !v.config.AllowPathTraversal {
		if strings.Contains(path, "..") {
			return fmt.Errorf("path traversal detected")
		}
	}

	// Clean the path
	cleaned := filepath.Clean(path)

	// Ensure cleaned path doesn't start with ../
	if strings.HasPrefix(cleaned, "..") {
		return fmt.Errorf("path traversal detected after cleaning")
	}

	return nil
}

// ValidateEntityType validates an entity type
func (v *Validator) ValidateEntityType(entityType string) error {
	if entityType == "" {
		return fmt.Errorf("entity type cannot be empty")
	}

	if len(entityType) > v.config.MaxEntityTypeLength {
		return fmt.Errorf("entity type exceeds maximum length of %d", v.config.MaxEntityTypeLength)
	}

	// Check if entity type is in allowed list
	if len(v.config.AllowedEntityTypes) > 0 {
		found := false
		for _, allowed := range v.config.AllowedEntityTypes {
			if entityType == allowed {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("entity type '%s' is not allowed", entityType)
		}
	}

	// Check for dangerous characters
	if !isAlphanumericWithUnderscore(entityType) {
		return fmt.Errorf("entity type contains invalid characters")
	}

	return nil
}

// ValidateEntityID validates an entity ID
func (v *Validator) ValidateEntityID(entityID string) error {
	if entityID == "" {
		return fmt.Errorf("entity ID cannot be empty")
	}

	if len(entityID) > v.config.MaxEntityIDLength {
		return fmt.Errorf("entity ID exceeds maximum length of %d", v.config.MaxEntityIDLength)
	}

	// Check for null bytes
	if strings.Contains(entityID, "\x00") {
		return fmt.Errorf("entity ID contains null bytes")
	}

	// Validate format (alphanumeric + dash + underscore)
	if !isAlphanumericWithDashUnderscore(entityID) {
		return fmt.Errorf("entity ID contains invalid characters")
	}

	return nil
}

// ValidateUserID validates a user ID
func (v *Validator) ValidateUserID(userID string) error {
	if userID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}

	if len(userID) < v.config.MinUserIDLength {
		return fmt.Errorf("user ID must be at least %d characters", v.config.MinUserIDLength)
	}

	if len(userID) > v.config.MaxUserIDLength {
		return fmt.Errorf("user ID exceeds maximum length of %d", v.config.MaxUserIDLength)
	}

	// Check for dangerous characters
	if !isAlphanumericWithDashUnderscore(userID) {
		return fmt.Errorf("user ID contains invalid characters")
	}

	return nil
}

// ValidateDescription validates a description
func (v *Validator) ValidateDescription(description string) error {
	if len(description) > v.config.MaxDescriptionLength {
		return fmt.Errorf("description exceeds maximum length of %d", v.config.MaxDescriptionLength)
	}

	// Check for null bytes
	if strings.Contains(description, "\x00") {
		return fmt.Errorf("description contains null bytes")
	}

	// Validate UTF-8
	if !utf8.ValidString(description) {
		return fmt.Errorf("description contains invalid UTF-8 characters")
	}

	return nil
}

// ValidateTags validates a list of tags
func (v *Validator) ValidateTags(tags []string) error {
	if len(tags) > v.config.MaxTagsPerFile {
		return fmt.Errorf("number of tags exceeds maximum of %d", v.config.MaxTagsPerFile)
	}

	for i, tag := range tags {
		if tag == "" {
			return fmt.Errorf("tag %d is empty", i)
		}

		if len(tag) > v.config.MaxTagLength {
			return fmt.Errorf("tag '%s' exceeds maximum length of %d", tag, v.config.MaxTagLength)
		}

		// Check for dangerous characters
		if !isAlphanumericWithDashUnderscore(tag) {
			return fmt.Errorf("tag '%s' contains invalid characters", tag)
		}
	}

	return nil
}

// SanitizeTags sanitizes a list of tags
func (v *Validator) SanitizeTags(tags []string) []string {
	sanitized := make([]string, 0, len(tags))

	for _, tag := range tags {
		// Trim whitespace
		tag = strings.TrimSpace(tag)

		// Convert to lowercase
		tag = strings.ToLower(tag)

		// Remove dangerous characters
		tag = regexp.MustCompile("[^a-z0-9-_]").ReplaceAllString(tag, "")

		// Skip empty tags
		if tag == "" {
			continue
		}

		// Skip if exceeds max length
		if len(tag) > v.config.MaxTagLength {
			continue
		}

		sanitized = append(sanitized, tag)
	}

	// Limit to max tags
	if len(sanitized) > v.config.MaxTagsPerFile {
		sanitized = sanitized[:v.config.MaxTagsPerFile]
	}

	return sanitized
}

// ValidateHash validates a file hash
func (v *Validator) ValidateHash(hash string) error {
	if hash == "" {
		return fmt.Errorf("hash cannot be empty")
	}

	// SHA-256 hash should be 64 hex characters
	if len(hash) != 64 {
		return fmt.Errorf("invalid hash length: expected 64, got %d", len(hash))
	}

	// Check if it's valid hex
	if !isHexadecimal(hash) {
		return fmt.Errorf("hash contains non-hexadecimal characters")
	}

	return nil
}

// ValidateReferenceID validates a reference ID
func (v *Validator) ValidateReferenceID(referenceID string) error {
	if referenceID == "" {
		return fmt.Errorf("reference ID cannot be empty")
	}

	// UUIDs should be 36 characters (32 hex + 4 dashes)
	if len(referenceID) != 36 {
		return fmt.Errorf("invalid reference ID format")
	}

	// Validate UUID format (8-4-4-4-12)
	pattern := `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`
	matched, err := regexp.MatchString(pattern, referenceID)
	if err != nil {
		return fmt.Errorf("failed to validate reference ID: %w", err)
	}

	if !matched {
		return fmt.Errorf("invalid reference ID format")
	}

	return nil
}

// isAlphanumericWithUnderscore checks if string contains only alphanumeric and underscore
func isAlphanumericWithUnderscore(s string) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, s)
	return matched
}

// isAlphanumericWithDashUnderscore checks if string contains only alphanumeric, dash, and underscore
func isAlphanumericWithDashUnderscore(s string) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, s)
	return matched
}

// isHexadecimal checks if string is valid hexadecimal
func isHexadecimal(s string) bool {
	matched, _ := regexp.MatchString(`^[0-9a-fA-F]+$`, s)
	return matched
}

// ValidateMimeType validates a MIME type
func (v *Validator) ValidateMimeType(mimeType string) error {
	if mimeType == "" {
		return fmt.Errorf("MIME type cannot be empty")
	}

	// MIME type format: type/subtype
	parts := strings.Split(mimeType, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid MIME type format")
	}

	// Validate type and subtype
	if parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("MIME type has empty type or subtype")
	}

	// Check for dangerous characters
	pattern := `^[a-zA-Z0-9][a-zA-Z0-9!#$&^_.+-]*$`
	for _, part := range parts {
		matched, _ := regexp.MatchString(pattern, part)
		if !matched {
			return fmt.Errorf("MIME type contains invalid characters")
		}
	}

	return nil
}

// SanitizeString removes dangerous characters from a string
func (v *Validator) SanitizeString(s string) string {
	// Remove null bytes
	s = strings.ReplaceAll(s, "\x00", "")

	// Remove control characters except newline and tab
	var result strings.Builder
	for _, r := range s {
		if r == '\n' || r == '\t' || r >= 32 {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// ValidateURL validates a URL
func (v *Validator) ValidateURL(url string) error {
	if url == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	// Check for common dangerous patterns
	urlLower := strings.ToLower(url)

	// Check for javascript: protocol
	if strings.HasPrefix(urlLower, "javascript:") {
		return fmt.Errorf("javascript: protocol is not allowed")
	}

	// Check for data: protocol (can be used for XSS)
	if strings.HasPrefix(urlLower, "data:") {
		return fmt.Errorf("data: protocol is not allowed")
	}

	// Check for file: protocol
	if strings.HasPrefix(urlLower, "file:") {
		return fmt.Errorf("file: protocol is not allowed")
	}

	return nil
}

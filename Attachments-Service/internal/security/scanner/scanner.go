package scanner

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
)

// Scanner handles security scanning of uploaded files
type Scanner struct {
	config *ScanConfig
	logger *zap.Logger
}

// ScanConfig contains security scanner configuration
type ScanConfig struct {
	// MIME type whitelist (empty = allow all)
	AllowedMimeTypes []string

	// File extension whitelist (empty = allow all)
	AllowedExtensions []string

	// Maximum file size in bytes
	MaxFileSize int64

	// Image validation settings
	MaxImageWidth      int
	MaxImageHeight     int
	MaxImagePixels     int64
	EnableImageBombProtection bool

	// ClamAV settings
	EnableClamAV     bool
	ClamAVSocket     string
	ClamAVTimeout    time.Duration

	// Magic bytes validation
	EnableMagicBytes bool
	StrictMagicBytes bool // Fail if magic bytes don't match extension

	// Content analysis
	EnableContentAnalysis bool
	MaxScanBytes         int64 // Maximum bytes to scan for content analysis
}

// DefaultScanConfig returns default scanner configuration
func DefaultScanConfig() *ScanConfig {
	return &ScanConfig{
		// Common safe MIME types
		AllowedMimeTypes: []string{
			// Images
			"image/jpeg", "image/png", "image/gif", "image/webp", "image/svg+xml",
			// Documents
			"application/pdf",
			"application/msword",
			"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
			"application/vnd.ms-excel",
			"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
			"application/vnd.ms-powerpoint",
			"application/vnd.openxmlformats-officedocument.presentationml.presentation",
			// Text
			"text/plain", "text/csv", "text/html", "text/markdown",
			// Archives
			"application/zip", "application/x-tar", "application/gzip",
			// Code
			"text/javascript", "application/json", "application/xml",
		},

		// Common safe extensions
		AllowedExtensions: []string{
			".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg",
			".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx",
			".txt", ".csv", ".md", ".html",
			".zip", ".tar", ".gz",
			".js", ".json", ".xml",
		},

		MaxFileSize: 100 * 1024 * 1024, // 100 MB

		MaxImageWidth:  10000,
		MaxImageHeight: 10000,
		MaxImagePixels: 100000000, // 100 megapixels
		EnableImageBombProtection: true,

		EnableClamAV:  false, // Disabled by default, requires ClamAV installed
		ClamAVSocket:  "/var/run/clamav/clamd.ctl",
		ClamAVTimeout: 30 * time.Second,

		EnableMagicBytes: true,
		StrictMagicBytes: false,

		EnableContentAnalysis: true,
		MaxScanBytes:         10 * 1024 * 1024, // 10 MB
	}
}

// NewScanner creates a new security scanner
func NewScanner(config *ScanConfig, logger *zap.Logger) *Scanner {
	if config == nil {
		config = DefaultScanConfig()
	}

	return &Scanner{
		config: config,
		logger: logger,
	}
}

// ScanResult contains the result of a security scan
type ScanResult struct {
	Safe             bool
	MimeType         string
	Extension        string
	SizeBytes        int64
	ImageWidth       int
	ImageHeight      int
	VirusDetected    bool
	VirusName        string
	Warnings         []string
	Errors           []string
	MagicBytesMatch  bool
	DetectedMimeType string
}

// Scan performs comprehensive security scanning on a file
func (s *Scanner) Scan(ctx context.Context, reader io.Reader, filename string) (*ScanResult, error) {
	result := &ScanResult{
		Safe:      true,
		Extension: filepath.Ext(strings.ToLower(filename)),
		Warnings:  []string{},
		Errors:    []string{},
	}

	// Read file into buffer for multiple scans
	var buffer bytes.Buffer
	teeReader := io.TeeReader(reader, &buffer)

	// Read all data
	data, err := io.ReadAll(teeReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	result.SizeBytes = int64(len(data))

	s.logger.Info("scanning file",
		zap.String("filename", filename),
		zap.String("extension", result.Extension),
		zap.Int64("size_bytes", result.SizeBytes),
	)

	// 1. File size check
	if err := s.checkFileSize(result); err != nil {
		result.Safe = false
		result.Errors = append(result.Errors, err.Error())
		return result, err
	}

	// 2. Extension validation
	if err := s.validateExtension(result); err != nil {
		result.Safe = false
		result.Errors = append(result.Errors, err.Error())
		return result, err
	}

	// 3. Detect MIME type from content
	if err := s.detectMimeType(data, result); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("MIME detection warning: %v", err))
	}

	// 4. MIME type validation
	if err := s.validateMimeType(result); err != nil {
		result.Safe = false
		result.Errors = append(result.Errors, err.Error())
		return result, err
	}

	// 5. Magic bytes validation
	if s.config.EnableMagicBytes {
		if err := s.validateMagicBytes(data, result); err != nil {
			if s.config.StrictMagicBytes {
				result.Safe = false
				result.Errors = append(result.Errors, err.Error())
				return result, err
			} else {
				result.Warnings = append(result.Warnings, err.Error())
			}
		}
	}

	// 6. Image validation (if image)
	if strings.HasPrefix(result.MimeType, "image/") && result.Extension != ".svg" {
		if err := s.validateImage(bytes.NewReader(data), result); err != nil {
			result.Safe = false
			result.Errors = append(result.Errors, err.Error())
			return result, err
		}
	}

	// 7. Content analysis
	if s.config.EnableContentAnalysis {
		if err := s.analyzeContent(data, result); err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Content analysis warning: %v", err))
		}
	}

	// 8. Virus scan (ClamAV)
	if s.config.EnableClamAV {
		if err := s.scanVirus(ctx, data, result); err != nil {
			result.Safe = false
			result.Errors = append(result.Errors, err.Error())
			return result, err
		}
	}

	s.logger.Info("file scan complete",
		zap.String("filename", filename),
		zap.Bool("safe", result.Safe),
		zap.String("mime_type", result.MimeType),
		zap.Int("warnings", len(result.Warnings)),
		zap.Int("errors", len(result.Errors)),
	)

	return result, nil
}

// checkFileSize validates file size
func (s *Scanner) checkFileSize(result *ScanResult) error {
	if result.SizeBytes > s.config.MaxFileSize {
		return fmt.Errorf("file size %d exceeds maximum %d bytes", result.SizeBytes, s.config.MaxFileSize)
	}

	if result.SizeBytes == 0 {
		return fmt.Errorf("file is empty")
	}

	return nil
}

// validateExtension checks if file extension is allowed
func (s *Scanner) validateExtension(result *ScanResult) error {
	if len(s.config.AllowedExtensions) == 0 {
		return nil // No whitelist, allow all
	}

	if result.Extension == "" {
		return fmt.Errorf("file has no extension")
	}

	for _, allowed := range s.config.AllowedExtensions {
		if strings.EqualFold(result.Extension, allowed) {
			return nil
		}
	}

	return fmt.Errorf("file extension %s is not allowed", result.Extension)
}

// detectMimeType detects MIME type from file content
func (s *Scanner) detectMimeType(data []byte, result *ScanResult) error {
	// Detect from content
	detectedType := http.DetectContentType(data)
	result.DetectedMimeType = detectedType

	// Use detected type if specific, otherwise try to infer from extension
	if detectedType != "application/octet-stream" {
		result.MimeType = detectedType
	} else {
		// Try to get MIME type from extension
		mimeType := mime.TypeByExtension(result.Extension)
		if mimeType != "" {
			result.MimeType = mimeType
		} else {
			result.MimeType = detectedType
		}
	}

	return nil
}

// validateMimeType checks if MIME type is allowed
func (s *Scanner) validateMimeType(result *ScanResult) error {
	if len(s.config.AllowedMimeTypes) == 0 {
		return nil // No whitelist, allow all
	}

	for _, allowed := range s.config.AllowedMimeTypes {
		if strings.EqualFold(result.MimeType, allowed) {
			return nil
		}
	}

	return fmt.Errorf("MIME type %s is not allowed", result.MimeType)
}

// validateMagicBytes validates file magic bytes (file signatures)
func (s *Scanner) validateMagicBytes(data []byte, result *ScanResult) error {
	if len(data) < 12 {
		return fmt.Errorf("file too small for magic bytes validation")
	}

	signature, expected := getMagicBytesSignature(data, result.Extension)
	result.MagicBytesMatch = signature == expected

	if !result.MagicBytesMatch {
		return fmt.Errorf("magic bytes mismatch: expected %s for %s, got %s",
			expected, result.Extension, signature)
	}

	return nil
}

// getMagicBytesSignature returns the detected and expected magic bytes signature
func getMagicBytesSignature(data []byte, extension string) (detected, expected string) {
	// Detect signature from file content
	if len(data) >= 2 {
		// JPEG
		if data[0] == 0xFF && data[1] == 0xD8 {
			detected = "JPEG"
		}
	}

	if len(data) >= 8 {
		// PNG
		if bytes.Equal(data[0:8], []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}) {
			detected = "PNG"
		}
	}

	if len(data) >= 6 {
		// GIF
		if bytes.Equal(data[0:6], []byte("GIF87a")) || bytes.Equal(data[0:6], []byte("GIF89a")) {
			detected = "GIF"
		}
	}

	if len(data) >= 4 {
		// PDF
		if bytes.Equal(data[0:4], []byte("%PDF")) {
			detected = "PDF"
		}

		// ZIP
		if bytes.Equal(data[0:4], []byte{0x50, 0x4B, 0x03, 0x04}) {
			detected = "ZIP"
		}
	}

	// Map extension to expected signature
	extLower := strings.ToLower(extension)
	switch extLower {
	case ".jpg", ".jpeg":
		expected = "JPEG"
	case ".png":
		expected = "PNG"
	case ".gif":
		expected = "GIF"
	case ".pdf":
		expected = "PDF"
	case ".zip", ".docx", ".xlsx", ".pptx":
		expected = "ZIP" // Office files are ZIP-based
	default:
		expected = "UNKNOWN"
	}

	return detected, expected
}

// validateImage validates image files
func (s *Scanner) validateImage(reader io.Reader, result *ScanResult) error {
	config, format, err := image.DecodeConfig(reader)
	if err != nil {
		return fmt.Errorf("invalid image format: %w", err)
	}

	result.ImageWidth = config.Width
	result.ImageHeight = config.Height

	s.logger.Debug("image validated",
		zap.String("format", format),
		zap.Int("width", config.Width),
		zap.Int("height", config.Height),
	)

	// Check dimensions
	if config.Width > s.config.MaxImageWidth {
		return fmt.Errorf("image width %d exceeds maximum %d", config.Width, s.config.MaxImageWidth)
	}

	if config.Height > s.config.MaxImageHeight {
		return fmt.Errorf("image height %d exceeds maximum %d", config.Height, s.config.MaxImageHeight)
	}

	// Decompression bomb protection
	if s.config.EnableImageBombProtection {
		pixels := int64(config.Width) * int64(config.Height)
		if pixels > s.config.MaxImagePixels {
			return fmt.Errorf("image has %d pixels, exceeds maximum %d (possible decompression bomb)",
				pixels, s.config.MaxImagePixels)
		}

		// Check compression ratio
		compressionRatio := float64(pixels) / float64(result.SizeBytes)
		if compressionRatio > 1000 { // More than 1000:1 compression is suspicious
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("High compression ratio %.2f:1 (possible decompression bomb)", compressionRatio))
		}
	}

	return nil
}

// analyzeContent performs content analysis for malicious patterns
func (s *Scanner) analyzeContent(data []byte, result *ScanResult) error {
	// Limit scan to configured maximum
	scanData := data
	if int64(len(data)) > s.config.MaxScanBytes {
		scanData = data[:s.config.MaxScanBytes]
	}

	content := string(scanData)
	contentLower := strings.ToLower(content)

	// Check for script injection patterns
	scriptPatterns := []string{
		"<script",
		"javascript:",
		"onerror=",
		"onload=",
		"eval(",
		"document.cookie",
		"window.location",
	}

	for _, pattern := range scriptPatterns {
		if strings.Contains(contentLower, pattern) {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("Potentially malicious content detected: %s", pattern))
		}
	}

	// Check for SQL injection patterns
	sqlPatterns := []string{
		"union select",
		"drop table",
		"delete from",
		"insert into",
		"'; --",
	}

	for _, pattern := range sqlPatterns {
		if strings.Contains(contentLower, pattern) {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("Potentially malicious SQL pattern detected: %s", pattern))
		}
	}

	// Check for null bytes (can be used to bypass filters)
	if bytes.Contains(scanData, []byte{0x00}) {
		result.Warnings = append(result.Warnings, "Null bytes detected in content")
	}

	return nil
}

// scanVirus scans file for viruses using ClamAV
func (s *Scanner) scanVirus(ctx context.Context, data []byte, result *ScanResult) error {
	// Create temporary file for scanning
	tmpFile, err := os.CreateTemp("", "scan-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Write data to temp file
	if _, err := tmpFile.Write(data); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	tmpFile.Close()

	// Run ClamAV scan
	scanCtx, cancel := context.WithTimeout(ctx, s.config.ClamAVTimeout)
	defer cancel()

	cmd := exec.CommandContext(scanCtx, "clamdscan", "--fdpass", tmpFile.Name())
	output, err := cmd.CombinedOutput()

	if err != nil {
		// Check if it's a virus detection (exit code 1) or error (exit code 2)
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 {
				// Virus detected
				result.VirusDetected = true
				result.VirusName = extractVirusName(string(output))

				s.logger.Warn("virus detected",
					zap.String("virus", result.VirusName),
				)

				return fmt.Errorf("virus detected: %s", result.VirusName)
			}
		}

		// Other error
		s.logger.Error("clamav scan failed",
			zap.Error(err),
			zap.String("output", string(output)),
		)

		// Don't fail the upload if ClamAV has issues, just log warning
		result.Warnings = append(result.Warnings, "Virus scan failed, file uploaded without scan")
		return nil
	}

	s.logger.Debug("virus scan complete",
		zap.Bool("clean", true),
	)

	return nil
}

// extractVirusName extracts virus name from ClamAV output
func extractVirusName(output string) string {
	// ClamAV output format: "filename: VirusName FOUND"
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "FOUND") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				virusPart := strings.TrimSpace(parts[1])
				virusPart = strings.TrimSuffix(virusPart, " FOUND")
				return virusPart
			}
		}
	}
	return "Unknown"
}

// ScanFile is a convenience method to scan a file from disk
func (s *Scanner) ScanFile(ctx context.Context, filepath string) (*ScanResult, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return s.Scan(ctx, file, filepath)
}

// IsAllowedMimeType checks if a MIME type is allowed
func (s *Scanner) IsAllowedMimeType(mimeType string) bool {
	if len(s.config.AllowedMimeTypes) == 0 {
		return true
	}

	for _, allowed := range s.config.AllowedMimeTypes {
		if strings.EqualFold(mimeType, allowed) {
			return true
		}
	}

	return false
}

// IsAllowedExtension checks if a file extension is allowed
func (s *Scanner) IsAllowedExtension(extension string) bool {
	if len(s.config.AllowedExtensions) == 0 {
		return true
	}

	for _, allowed := range s.config.AllowedExtensions {
		if strings.EqualFold(extension, allowed) {
			return true
		}
	}

	return false
}

// IsEnabled returns whether virus scanning is enabled
func (s *Scanner) IsEnabled() bool {
	return s.config.EnableClamAV
}

// Ping checks if ClamAV is accessible
func (s *Scanner) Ping(ctx context.Context) error {
	if !s.config.EnableClamAV {
		return nil // Not enabled, so no error
	}

	// Try to connect to ClamAV
	if s.config.ClamAVSocket != "" {
		// Unix socket connection
		conn, err := net.DialTimeout("unix", s.config.ClamAVSocket, 5*time.Second)
		if err != nil {
			return fmt.Errorf("failed to connect to ClamAV socket: %w", err)
		}
		defer conn.Close()

		// Send PING command
		_, err = conn.Write([]byte("zPING\x00"))
		if err != nil {
			return fmt.Errorf("failed to send ping to ClamAV: %w", err)
		}

		// Read response with timeout
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		response := make([]byte, 128)
		n, err := conn.Read(response)
		if err != nil {
			return fmt.Errorf("failed to read ping response from ClamAV: %w", err)
		}

		// Check for PONG response
		if !strings.Contains(string(response[:n]), "PONG") {
			return fmt.Errorf("unexpected ping response from ClamAV: %s", string(response[:n]))
		}

		return nil
	}

	// TCP connection (host:port)
	return fmt.Errorf("ClamAV TCP connection not yet implemented")
}

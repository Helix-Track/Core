package scanner

import (
	"bytes"
	"context"
	"testing"

	"go.uber.org/zap"
)

func TestNewScanner(t *testing.T) {
	logger := zap.NewNop()

	t.Run("with nil config uses defaults", func(t *testing.T) {
		scanner := NewScanner(nil, logger)
		if scanner == nil {
			t.Fatal("expected scanner, got nil")
		}
		if scanner.config == nil {
			t.Fatal("expected default config")
		}
	})

	t.Run("with custom config", func(t *testing.T) {
		config := &ScanConfig{
			MaxFileSize: 50 * 1024 * 1024,
			AllowedMimeTypes: []string{"image/jpeg"},
		}
		scanner := NewScanner(config, logger)
		if scanner.config.MaxFileSize != config.MaxFileSize {
			t.Errorf("expected MaxFileSize %d, got %d", config.MaxFileSize, scanner.config.MaxFileSize)
		}
	})
}

func TestScan_FileSize(t *testing.T) {
	logger := zap.NewNop()
	config := &ScanConfig{
		MaxFileSize: 1024, // 1 KB limit
		EnableMagicBytes: false,
		EnableClamAV: false,
		EnableContentAnalysis: false,
	}
	scanner := NewScanner(config, logger)

	t.Run("file too large", func(t *testing.T) {
		// Create 2 KB file
		data := bytes.Repeat([]byte("a"), 2048)
		reader := bytes.NewReader(data)

		result, err := scanner.Scan(context.Background(), reader, "test.txt")
		if err == nil {
			t.Fatal("expected error for oversized file")
		}
		if result.Safe {
			t.Error("expected Safe to be false")
		}
	})

	t.Run("file within size limit", func(t *testing.T) {
		data := []byte("test content")
		reader := bytes.NewReader(data)

		result, err := scanner.Scan(context.Background(), reader, "test.txt")
		if err != nil && !result.Safe {
			t.Fatalf("expected no error for valid file, got: %v", err)
		}
	})

	t.Run("empty file", func(t *testing.T) {
		data := []byte("")
		reader := bytes.NewReader(data)

		result, err := scanner.Scan(context.Background(), reader, "test.txt")
		if err == nil {
			t.Fatal("expected error for empty file")
		}
		if result.Safe {
			t.Error("expected Safe to be false")
		}
	})
}

func TestScan_Extension(t *testing.T) {
	logger := zap.NewNop()
	config := &ScanConfig{
		AllowedExtensions: []string{".txt", ".pdf"},
		MaxFileSize: 10 * 1024 * 1024,
		EnableMagicBytes: false,
		EnableClamAV: false,
		EnableContentAnalysis: false,
	}
	scanner := NewScanner(config, logger)

	t.Run("allowed extension", func(t *testing.T) {
		data := []byte("test content")
		reader := bytes.NewReader(data)

		result, _ := scanner.Scan(context.Background(), reader, "test.txt")
		if !result.Safe {
			t.Error("expected file with allowed extension to be safe")
		}
	})

	t.Run("disallowed extension", func(t *testing.T) {
		data := []byte("test content")
		reader := bytes.NewReader(data)

		result, err := scanner.Scan(context.Background(), reader, "test.exe")
		if err == nil {
			t.Fatal("expected error for disallowed extension")
		}
		if result.Safe {
			t.Error("expected Safe to be false")
		}
	})

	t.Run("no extension", func(t *testing.T) {
		data := []byte("test content")
		reader := bytes.NewReader(data)

		result, err := scanner.Scan(context.Background(), reader, "test")
		if err == nil {
			t.Fatal("expected error for file without extension")
		}
		if result.Safe {
			t.Error("expected Safe to be false")
		}
	})
}

func TestScan_MimeType(t *testing.T) {
	logger := zap.NewNop()
	config := &ScanConfig{
		AllowedMimeTypes: []string{"text/plain", "application/pdf"},
		MaxFileSize: 10 * 1024 * 1024,
		EnableMagicBytes: false,
		EnableClamAV: false,
		EnableContentAnalysis: false,
	}
	scanner := NewScanner(config, logger)

	t.Run("allowed MIME type", func(t *testing.T) {
		data := []byte("plain text content")
		reader := bytes.NewReader(data)

		result, _ := scanner.Scan(context.Background(), reader, "test.txt")
		if result.MimeType != "text/plain; charset=utf-8" {
			t.Errorf("expected text/plain MIME type, got %s", result.MimeType)
		}
	})
}

func TestGetMagicBytesSignature(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		extension string
		wantSig   string
		wantExp   string
	}{
		{
			name:      "JPEG file",
			data:      []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46},
			extension: ".jpg",
			wantSig:   "JPEG",
			wantExp:   "JPEG",
		},
		{
			name:      "PNG file",
			data:      []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
			extension: ".png",
			wantSig:   "PNG",
			wantExp:   "PNG",
		},
		{
			name:      "GIF file",
			data:      []byte("GIF89a"),
			extension: ".gif",
			wantSig:   "GIF",
			wantExp:   "GIF",
		},
		{
			name:      "PDF file",
			data:      []byte("%PDF-1.4"),
			extension: ".pdf",
			wantSig:   "PDF",
			wantExp:   "PDF",
		},
		{
			name:      "ZIP file",
			data:      []byte{0x50, 0x4B, 0x03, 0x04},
			extension: ".zip",
			wantSig:   "ZIP",
			wantExp:   "ZIP",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSig, gotExp := getMagicBytesSignature(tt.data, tt.extension)
			if gotSig != tt.wantSig {
				t.Errorf("signature = %s, want %s", gotSig, tt.wantSig)
			}
			if gotExp != tt.wantExp {
				t.Errorf("expected = %s, want %s", gotExp, tt.wantExp)
			}
		})
	}
}

func TestScan_MagicBytes(t *testing.T) {
	logger := zap.NewNop()
	config := &ScanConfig{
		MaxFileSize: 10 * 1024 * 1024,
		EnableMagicBytes: true,
		StrictMagicBytes: true,
		EnableClamAV: false,
		EnableContentAnalysis: false,
	}
	scanner := NewScanner(config, logger)

	t.Run("valid JPEG with .jpg extension", func(t *testing.T) {
		// Create valid JPEG header
		data := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46}
		data = append(data, bytes.Repeat([]byte{0x00}, 100)...)
		reader := bytes.NewReader(data)

		result, err := scanner.Scan(context.Background(), reader, "test.jpg")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.MagicBytesMatch {
			t.Error("expected magic bytes to match")
		}
	})

	t.Run("invalid magic bytes", func(t *testing.T) {
		// PNG data with JPEG extension
		data := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
		data = append(data, bytes.Repeat([]byte{0x00}, 100)...)
		reader := bytes.NewReader(data)

		result, err := scanner.Scan(context.Background(), reader, "test.jpg")
		if err == nil {
			t.Fatal("expected error for magic bytes mismatch")
		}
		if result.Safe {
			t.Error("expected Safe to be false")
		}
	})
}

func TestScan_ContentAnalysis(t *testing.T) {
	logger := zap.NewNop()
	config := &ScanConfig{
		MaxFileSize: 10 * 1024 * 1024,
		EnableContentAnalysis: true,
		EnableMagicBytes: false,
		EnableClamAV: false,
	}
	scanner := NewScanner(config, logger)

	t.Run("detects script injection", func(t *testing.T) {
		data := []byte("<html><script>alert('xss')</script></html>")
		reader := bytes.NewReader(data)

		result, _ := scanner.Scan(context.Background(), reader, "test.html")
		if len(result.Warnings) == 0 {
			t.Error("expected warnings for script content")
		}
		// Should still be safe (just warnings)
		if !result.Safe {
			t.Error("expected Safe to be true with warnings")
		}
	})

	t.Run("detects SQL injection patterns", func(t *testing.T) {
		data := []byte("SELECT * FROM users; DROP TABLE users; --")
		reader := bytes.NewReader(data)

		result, _ := scanner.Scan(context.Background(), reader, "test.txt")
		if len(result.Warnings) == 0 {
			t.Error("expected warnings for SQL patterns")
		}
	})

	t.Run("detects null bytes", func(t *testing.T) {
		data := []byte("test\x00content")
		reader := bytes.NewReader(data)

		result, _ := scanner.Scan(context.Background(), reader, "test.txt")
		if len(result.Warnings) == 0 {
			t.Error("expected warnings for null bytes")
		}
	})

	t.Run("clean content", func(t *testing.T) {
		data := []byte("This is clean text content without any malicious patterns.")
		reader := bytes.NewReader(data)

		result, _ := scanner.Scan(context.Background(), reader, "test.txt")
		if len(result.Warnings) > 0 {
			t.Errorf("expected no warnings, got %d", len(result.Warnings))
		}
	})
}

func TestIsAllowedMimeType(t *testing.T) {
	logger := zap.NewNop()
	config := &ScanConfig{
		AllowedMimeTypes: []string{"image/jpeg", "application/pdf"},
	}
	scanner := NewScanner(config, logger)

	tests := []struct {
		name     string
		mimeType string
		want     bool
	}{
		{"allowed JPEG", "image/jpeg", true},
		{"allowed PDF", "application/pdf", true},
		{"disallowed type", "application/x-executable", false},
		{"case insensitive", "IMAGE/JPEG", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := scanner.IsAllowedMimeType(tt.mimeType); got != tt.want {
				t.Errorf("IsAllowedMimeType(%s) = %v, want %v", tt.mimeType, got, tt.want)
			}
		})
	}
}

func TestIsAllowedExtension(t *testing.T) {
	logger := zap.NewNop()
	config := &ScanConfig{
		AllowedExtensions: []string{".jpg", ".pdf", ".txt"},
	}
	scanner := NewScanner(config, logger)

	tests := []struct {
		name      string
		extension string
		want      bool
	}{
		{"allowed JPG", ".jpg", true},
		{"allowed PDF", ".pdf", true},
		{"disallowed EXE", ".exe", false},
		{"case insensitive", ".JPG", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := scanner.IsAllowedExtension(tt.extension); got != tt.want {
				t.Errorf("IsAllowedExtension(%s) = %v, want %v", tt.extension, got, tt.want)
			}
		})
	}
}

func TestExtractVirusName(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   string
	}{
		{
			name:   "standard ClamAV output",
			output: "/tmp/file: Win.Test.EICAR_HDB-1 FOUND",
			want:   "Win.Test.EICAR_HDB-1",
		},
		{
			name:   "multiline output",
			output: "Scanning /tmp/file\n/tmp/file: Trojan.GenericKD FOUND\nDone",
			want:   "Trojan.GenericKD",
		},
		{
			name:   "no virus found",
			output: "Scanning complete. No virus found.",
			want:   "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractVirusName(tt.output); got != tt.want {
				t.Errorf("extractVirusName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkScan(b *testing.B) {
	logger := zap.NewNop()
	config := &ScanConfig{
		MaxFileSize: 10 * 1024 * 1024,
		EnableMagicBytes: true,
		EnableContentAnalysis: true,
		EnableClamAV: false,
	}
	scanner := NewScanner(config, logger)

	data := bytes.Repeat([]byte("test content "), 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader(data)
		scanner.Scan(context.Background(), reader, "test.txt")
	}
}

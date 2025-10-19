package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
)

// FileHasher handles file hashing operations
type FileHasher struct {
	algorithm string
}

// NewFileHasher creates a new file hasher
func NewFileHasher() *FileHasher {
	return &FileHasher{
		algorithm: "SHA-256",
	}
}

// CalculateHash calculates the SHA-256 hash of a file from a reader
// Returns: hash (hex string), size in bytes, error
func (h *FileHasher) CalculateHash(reader io.Reader) (string, int64, error) {
	hasher := sha256.New()
	size, err := io.Copy(hasher, reader)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read file for hashing: %w", err)
	}

	hashBytes := hasher.Sum(nil)
	hashString := hex.EncodeToString(hashBytes)

	return hashString, size, nil
}

// CalculateHashFromFile calculates the SHA-256 hash of a file from disk
func (h *FileHasher) CalculateHashFromFile(filepath string) (string, int64, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return "", 0, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return h.CalculateHash(file)
}

// CalculateHashFromBytes calculates the SHA-256 hash of a byte slice
func (h *FileHasher) CalculateHashFromBytes(data []byte) string {
	hasher := sha256.New()
	hasher.Write(data)
	hashBytes := hasher.Sum(nil)
	return hex.EncodeToString(hashBytes)
}

// CalculateHashWithCallback calculates hash while reading, with progress callback
// callback is called with bytes read so far
func (h *FileHasher) CalculateHashWithCallback(reader io.Reader, callback func(int64)) (string, int64, error) {
	hasher := sha256.New()
	counter := &progressCounter{
		callback: callback,
	}

	multiWriter := io.MultiWriter(hasher, counter)
	size, err := io.Copy(multiWriter, reader)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read file for hashing: %w", err)
	}

	hashBytes := hasher.Sum(nil)
	hashString := hex.EncodeToString(hashBytes)

	return hashString, size, nil
}

// VerifyHash verifies that a file matches the expected hash
func (h *FileHasher) VerifyHash(reader io.Reader, expectedHash string) (bool, error) {
	actualHash, _, err := h.CalculateHash(reader)
	if err != nil {
		return false, err
	}

	return actualHash == expectedHash, nil
}

// VerifyHashFromFile verifies that a file on disk matches the expected hash
func (h *FileHasher) VerifyHashFromFile(filepath, expectedHash string) (bool, error) {
	actualHash, _, err := h.CalculateHashFromFile(filepath)
	if err != nil {
		return false, err
	}

	return actualHash == expectedHash, nil
}

// StreamingHasher allows calculating hash while streaming data
type StreamingHasher struct {
	hasher hash.Hash
	size   int64
}

// NewStreamingHasher creates a new streaming hasher
func NewStreamingHasher() *StreamingHasher {
	return &StreamingHasher{
		hasher: sha256.New(),
		size:   0,
	}
}

// Write writes data to the hasher
func (s *StreamingHasher) Write(p []byte) (n int, err error) {
	n, err = s.hasher.Write(p)
	s.size += int64(n)
	return n, err
}

// Finalize returns the final hash and size
func (s *StreamingHasher) Finalize() (string, int64) {
	hashBytes := s.hasher.Sum(nil)
	hashString := hex.EncodeToString(hashBytes)
	return hashString, s.size
}

// Reset resets the hasher for reuse
func (s *StreamingHasher) Reset() {
	s.hasher.Reset()
	s.size = 0
}

// GetSize returns the current size
func (s *StreamingHasher) GetSize() int64 {
	return s.size
}

// progressCounter counts bytes read and calls a callback
type progressCounter struct {
	count    int64
	callback func(int64)
}

func (pc *progressCounter) Write(p []byte) (n int, err error) {
	n = len(p)
	pc.count += int64(n)
	if pc.callback != nil {
		pc.callback(pc.count)
	}
	return n, nil
}

// HashReader wraps a reader and calculates hash on-the-fly
type HashReader struct {
	reader io.Reader
	hasher *StreamingHasher
}

// NewHashReader creates a new hash reader
func NewHashReader(reader io.Reader) *HashReader {
	return &HashReader{
		reader: reader,
		hasher: NewStreamingHasher(),
	}
}

// Read implements io.Reader interface
func (hr *HashReader) Read(p []byte) (n int, err error) {
	n, err = hr.reader.Read(p)
	if n > 0 {
		hr.hasher.Write(p[:n])
	}
	return n, err
}

// GetHash returns the hash and size computed so far
func (hr *HashReader) GetHash() (string, int64) {
	return hr.hasher.Finalize()
}

// CalculatePartialHash calculates hash of first N bytes
func CalculatePartialHash(reader io.Reader, maxBytes int64) (string, int64, error) {
	hasher := sha256.New()
	limitedReader := io.LimitReader(reader, maxBytes)

	size, err := io.Copy(hasher, limitedReader)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read for partial hashing: %w", err)
	}

	hashBytes := hasher.Sum(nil)
	hashString := hex.EncodeToString(hashBytes)

	return hashString, size, nil
}

// CompareHashes compares two hash strings (constant-time comparison)
func CompareHashes(hash1, hash2 string) bool {
	if len(hash1) != len(hash2) {
		return false
	}

	// Constant-time comparison to prevent timing attacks
	var result byte
	for i := 0; i < len(hash1); i++ {
		result |= hash1[i] ^ hash2[i]
	}

	return result == 0
}

// IsValidSHA256Hash checks if a string is a valid SHA-256 hash
func IsValidSHA256Hash(hash string) bool {
	if len(hash) != 64 {
		return false
	}

	// Check if all characters are hexadecimal
	for _, c := range hash {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}

	return true
}

// HashString calculates SHA-256 hash of a string
func HashString(s string) string {
	hasher := sha256.New()
	hasher.Write([]byte(s))
	hashBytes := hasher.Sum(nil)
	return hex.EncodeToString(hashBytes)
}

// GenerateFileID generates a unique file ID from hash and timestamp
func GenerateFileID(hash string, timestamp int64) string {
	data := fmt.Sprintf("%s:%d", hash, timestamp)
	return HashString(data)
}

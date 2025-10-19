package adapters

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"go.uber.org/zap"
)

// S3Adapter implements StorageAdapter for AWS S3 or S3-compatible storage
type S3Adapter struct {
	client     *s3.Client
	bucket     string
	region     string
	prefix     string // Optional prefix for all keys
	logger     *zap.Logger
	config     *S3Config
}

// S3Config contains S3 adapter configuration
type S3Config struct {
	// AWS credentials
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string // Optional

	// S3 configuration
	Region   string
	Bucket   string
	Endpoint string // Optional: for S3-compatible storage (e.g., MinIO)
	Prefix   string // Optional: prefix for all keys

	// Connection settings
	UsePathStyle bool // Use path-style URLs instead of virtual-hosted-style
	DisableSSL   bool

	// Upload settings
	ServerSideEncryption string // e.g., "AES256" or "aws:kms"
	StorageClass         string // e.g., "STANDARD", "STANDARD_IA", "GLACIER"
}

// NewS3Adapter creates a new S3 storage adapter
func NewS3Adapter(ctx context.Context, cfg *S3Config, logger *zap.Logger) (*S3Adapter, error) {
	if cfg == nil {
		return nil, fmt.Errorf("S3 configuration is required")
	}

	if cfg.Bucket == "" {
		return nil, fmt.Errorf("S3 bucket name is required")
	}

	if cfg.Region == "" {
		cfg.Region = "us-east-1" // Default region
	}

	// Build AWS config
	awsCfg, err := buildAWSConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to build AWS config: %w", err)
	}

	// Create S3 client
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
		if cfg.UsePathStyle {
			o.UsePathStyle = true
		}
	})

	adapter := &S3Adapter{
		client: client,
		bucket: cfg.Bucket,
		region: cfg.Region,
		prefix: cfg.Prefix,
		logger: logger,
		config: cfg,
	}

	// Verify bucket access
	if err := adapter.verifyBucketAccess(ctx); err != nil {
		return nil, fmt.Errorf("failed to verify bucket access: %w", err)
	}

	logger.Info("S3 adapter initialized",
		zap.String("bucket", cfg.Bucket),
		zap.String("region", cfg.Region),
		zap.String("endpoint", cfg.Endpoint),
	)

	return adapter, nil
}

// buildAWSConfig builds AWS SDK configuration
func buildAWSConfig(ctx context.Context, cfg *S3Config) (aws.Config, error) {
	// Static credentials provider
	credsProvider := credentials.NewStaticCredentialsProvider(
		cfg.AccessKeyID,
		cfg.SecretAccessKey,
		cfg.SessionToken,
	)

	// Load config with credentials
	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credsProvider),
	)

	if err != nil {
		return aws.Config{}, err
	}

	return awsCfg, nil
}

// verifyBucketAccess verifies we can access the bucket
func (a *S3Adapter) verifyBucketAccess(ctx context.Context) error {
	_, err := a.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(a.bucket),
	})

	if err != nil {
		return fmt.Errorf("cannot access bucket %s: %w", a.bucket, err)
	}

	return nil
}

// Store stores a file in S3 using hash-based sharding
func (a *S3Adapter) Store(ctx context.Context, hash string, data io.Reader, size int64) (string, error) {
	if len(hash) < 4 {
		return "", fmt.Errorf("invalid hash length: %d", len(hash))
	}

	// Create sharded key: prefix/ab/cd/hash
	key := a.buildKey(hash)

	// Check if object already exists
	_, err := a.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(key),
	})

	if err == nil {
		// Object already exists
		a.logger.Debug("object already exists in S3",
			zap.String("key", key),
			zap.String("hash", hash),
		)
		return key, nil
	}

	// Read data into buffer (needed for S3 upload)
	var buffer bytes.Buffer
	written, err := io.Copy(&buffer, data)
	if err != nil {
		return "", NewStorageError("read", key, err)
	}

	if written != size {
		return "", fmt.Errorf("size mismatch: expected %d, got %d", size, written)
	}

	// Prepare upload input
	uploadInput := &s3.PutObjectInput{
		Bucket:        aws.String(a.bucket),
		Key:           aws.String(key),
		Body:          bytes.NewReader(buffer.Bytes()),
		ContentLength: aws.Int64(size),
	}

	// Add server-side encryption if configured
	if a.config.ServerSideEncryption != "" {
		uploadInput.ServerSideEncryption = types.ServerSideEncryption(a.config.ServerSideEncryption)
	}

	// Add storage class if configured
	if a.config.StorageClass != "" {
		uploadInput.StorageClass = types.StorageClass(a.config.StorageClass)
	}

	// Upload to S3
	_, err = a.client.PutObject(ctx, uploadInput)
	if err != nil {
		return "", NewStorageError("upload", key, err)
	}

	a.logger.Info("file stored successfully in S3",
		zap.String("key", key),
		zap.String("hash", hash),
		zap.Int64("size", size),
	)

	return key, nil
}

// Retrieve retrieves a file from S3
func (a *S3Adapter) Retrieve(ctx context.Context, path string) (io.ReadCloser, error) {

	result, err := a.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(path),
	})

	if err != nil {
		return nil, NewStorageError("retrieve", path, err)
	}

	return result.Body, nil
}

// Delete deletes a file from S3
func (a *S3Adapter) Delete(ctx context.Context, path string) error {

	_, err := a.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(path),
	})

	if err != nil {
		return NewStorageError("delete", path, err)
	}

	a.logger.Info("file deleted successfully from S3",
		zap.String("key", path),
	)

	return nil
}

// Exists checks if a file exists in S3
func (a *S3Adapter) Exists(ctx context.Context, path string) (bool, error) {

	_, err := a.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(path),
	})

	if err != nil {
		// Check if it's a "not found" error
		if isNotFoundError(err) {
			return false, nil
		}
		return false, NewStorageError("exists", path, err)
	}

	return true, nil
}

// GetSize returns the size of a file in S3
func (a *S3Adapter) GetSize(ctx context.Context, path string) (int64, error) {

	result, err := a.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(path),
	})

	if err != nil {
		if isNotFoundError(err) {
			return 0, fmt.Errorf("file not found: %s", path)
		}
		return 0, NewStorageError("size", path, err)
	}

	return aws.ToInt64(result.ContentLength), nil
}

// GetMetadata returns metadata about a file in S3
func (a *S3Adapter) GetMetadata(ctx context.Context, path string) (*FileMetadata, error) {

	result, err := a.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(path),
	})

	if err != nil {
		if isNotFoundError(err) {
			return &FileMetadata{
				Path:   path,
				Exists: false,
			}, nil
		}
		return nil, NewStorageError("metadata", path, err)
	}

	var lastModified int64
	if result.LastModified != nil {
		lastModified = result.LastModified.Unix()
	}

	return &FileMetadata{
		Path:         path,
		Size:         aws.ToInt64(result.ContentLength),
		LastModified: lastModified,
		Exists:       true,
	}, nil
}

// Ping checks if S3 is accessible
func (a *S3Adapter) Ping(ctx context.Context) error {
	return a.verifyBucketAccess(context.Background())
}

// GetCapacity returns S3 capacity information
// Note: S3 has effectively unlimited capacity
func (a *S3Adapter) GetCapacity(ctx context.Context) (*CapacityInfo, error) {
	// S3 has no practical capacity limits
	// Return large values to indicate "unlimited"
	return &CapacityInfo{
		TotalBytes:     1 << 60,          // 1 exabyte
		UsedBytes:      0,                // Unknown
		AvailableBytes: 1 << 60,          // 1 exabyte
		UsagePercent:   0.0,
	}, nil
}

// GetType returns the adapter type
func (a *S3Adapter) GetType() string {
	return "s3"
}

// buildKey builds the S3 key with hash-based sharding
func (a *S3Adapter) buildKey(hash string) string {
	// Create sharded path: ab/cd/hash
	shard1 := hash[0:2]
	shard2 := hash[2:4]

	key := filepath.Join(shard1, shard2, hash)

	// Add prefix if configured
	if a.prefix != "" {
		key = filepath.Join(a.prefix, key)
	}

	// Convert to forward slashes (S3 uses forward slashes)
	key = filepath.ToSlash(key)

	return key
}

// GetPresignedURL generates a presigned URL for downloading
func (a *S3Adapter) GetPresignedURL(path string, expiresIn time.Duration) (string, error) {
	ctx := context.Background()

	presignClient := s3.NewPresignClient(a.client)

	request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(path),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiresIn
	})

	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	a.logger.Debug("presigned URL generated",
		zap.String("key", path),
		zap.Duration("expires_in", expiresIn),
	)

	return request.URL, nil
}

// GetPresignedUploadURL generates a presigned URL for uploading
func (a *S3Adapter) GetPresignedUploadURL(hash string, expiresIn time.Duration) (string, error) {
	ctx := context.Background()

	key := a.buildKey(hash)

	presignClient := s3.NewPresignClient(a.client)

	request, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiresIn
	})

	if err != nil {
		return "", fmt.Errorf("failed to generate presigned upload URL: %w", err)
	}

	a.logger.Debug("presigned upload URL generated",
		zap.String("key", key),
		zap.Duration("expires_in", expiresIn),
	)

	return request.URL, nil
}

// ListFiles lists all files in S3 bucket (for admin/debug purposes)
func (a *S3Adapter) ListFiles() ([]string, error) {
	ctx := context.Background()

	var files []string
	var continuationToken *string

	prefix := a.prefix
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	for {
		input := &s3.ListObjectsV2Input{
			Bucket:            aws.String(a.bucket),
			Prefix:            aws.String(prefix),
			ContinuationToken: continuationToken,
		}

		result, err := a.client.ListObjectsV2(ctx, input)
		if err != nil {
			return nil, NewStorageError("list", a.bucket, err)
		}

		for _, obj := range result.Contents {
			files = append(files, aws.ToString(obj.Key))
		}

		if !aws.ToBool(result.IsTruncated) {
			break
		}

		continuationToken = result.NextContinuationToken
	}

	return files, nil
}

// GetStorageStats returns storage statistics for S3
func (a *S3Adapter) GetStorageStats() (*StorageStats, error) {
	ctx := context.Background()
	files, err := a.ListFiles()
	if err != nil {
		return nil, err
	}

	stats := &StorageStats{
		FileCount:  int64(len(files)),
		TotalBytes: 0,
	}

	// Sum up file sizes (this can be slow for large buckets)
	for _, key := range files {
		result, err := a.client.HeadObject(ctx, &s3.HeadObjectInput{
			Bucket: aws.String(a.bucket),
			Key:    aws.String(key),
		})

		if err != nil {
			a.logger.Warn("failed to get object metadata",
				zap.String("key", key),
				zap.Error(err),
			)
			continue
		}

		stats.TotalBytes += aws.ToInt64(result.ContentLength)
	}

	return stats, nil
}

// Copy copies a file within S3
func (a *S3Adapter) Copy(srcPath, dstPath string) error {
	ctx := context.Background()

	copySource := fmt.Sprintf("%s/%s", a.bucket, srcPath)

	_, err := a.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(a.bucket),
		CopySource: aws.String(copySource),
		Key:        aws.String(dstPath),
	})

	if err != nil {
		return NewStorageError("copy", srcPath, err)
	}

	a.logger.Info("file copied successfully in S3",
		zap.String("src", srcPath),
		zap.String("dst", dstPath),
	)

	return nil
}

// isNotFoundError checks if an error is a "not found" error
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	return strings.Contains(errStr, "NotFound") ||
		strings.Contains(errStr, "NoSuchKey") ||
		strings.Contains(errStr, "404")
}

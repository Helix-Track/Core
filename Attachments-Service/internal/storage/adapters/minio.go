package adapters

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"
)

// MinIOAdapter implements StorageAdapter for MinIO object storage
// MinIO is S3-compatible, so this is essentially a wrapper around S3Adapter
// with MinIO-specific defaults
type MinIOAdapter struct {
	*S3Adapter
}

// MinIOConfig contains MinIO adapter configuration
type MinIOConfig struct {
	// MinIO connection details
	Endpoint        string // e.g., "localhost:9000" or "minio.example.com:9000"
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string

	// Optional settings
	UseSSL bool   // Whether to use HTTPS
	Prefix string // Optional prefix for all keys

	// Upload settings
	StorageClass string // MinIO supports same storage classes as S3
}

// NewMinIOAdapter creates a new MinIO storage adapter
func NewMinIOAdapter(ctx context.Context, cfg *MinIOConfig, logger *zap.Logger) (*MinIOAdapter, error) {
	if cfg == nil {
		return nil, fmt.Errorf("MinIO configuration is required")
	}

	if cfg.Endpoint == "" {
		return nil, fmt.Errorf("MinIO endpoint is required")
	}

	if cfg.Bucket == "" {
		return nil, fmt.Errorf("MinIO bucket name is required")
	}

	// Convert MinIO config to S3 config
	s3Config := &S3Config{
		AccessKeyID:     cfg.AccessKeyID,
		SecretAccessKey: cfg.SecretAccessKey,
		Region:          "us-east-1", // MinIO doesn't require specific region
		Bucket:          cfg.Bucket,
		Endpoint:        buildMinIOEndpoint(cfg.Endpoint, cfg.UseSSL),
		Prefix:          cfg.Prefix,
		UsePathStyle:    true,  // MinIO requires path-style URLs
		DisableSSL:      !cfg.UseSSL,
		StorageClass:    cfg.StorageClass,
	}

	// Create underlying S3 adapter
	s3Adapter, err := NewS3Adapter(ctx, s3Config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 adapter for MinIO: %w", err)
	}

	logger.Info("MinIO adapter initialized",
		zap.String("endpoint", cfg.Endpoint),
		zap.String("bucket", cfg.Bucket),
		zap.Bool("use_ssl", cfg.UseSSL),
	)

	return &MinIOAdapter{
		S3Adapter: s3Adapter,
	}, nil
}

// GetType returns the adapter type
func (a *MinIOAdapter) GetType() string {
	return "minio"
}

// buildMinIOEndpoint builds the full MinIO endpoint URL
func buildMinIOEndpoint(endpoint string, useSSL bool) string {
	scheme := "http"
	if useSSL {
		scheme = "https"
	}

	// Check if endpoint already has scheme
	if hasScheme(endpoint) {
		return endpoint
	}

	return fmt.Sprintf("%s://%s", scheme, endpoint)
}

// hasScheme checks if URL already has a scheme
func hasScheme(url string) bool {
	return len(url) > 7 && (url[0:7] == "http://" || url[0:8] == "https://")
}

// EnsureBucket creates the bucket if it doesn't exist
func (a *MinIOAdapter) EnsureBucket(ctx context.Context) error {
	// Check if bucket exists
	err := a.Ping(ctx)
	if err != nil {
		// Bucket doesn't exist or we don't have access
		// Try to create it
		return a.CreateBucket(ctx)
	}

	return nil
}

// CreateBucket creates a new bucket in MinIO
func (a *MinIOAdapter) CreateBucket(ctx context.Context) error {
	_, err := a.client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(a.bucket),
	})

	if err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	a.logger.Info("MinIO bucket created",
		zap.String("bucket", a.bucket),
	)

	return nil
}

// DeleteBucket deletes the bucket (dangerous!)
func (a *MinIOAdapter) DeleteBucket(ctx context.Context) error {
	_, err := a.client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(a.bucket),
	})

	if err != nil {
		return fmt.Errorf("failed to delete bucket: %w", err)
	}

	a.logger.Warn("MinIO bucket deleted",
		zap.String("bucket", a.bucket),
	)

	return nil
}

// SetBucketPolicy sets a bucket policy (MinIO-specific)
func (a *MinIOAdapter) SetBucketPolicy(ctx context.Context, policy string) error {
	_, err := a.client.PutBucketPolicy(ctx, &s3.PutBucketPolicyInput{
		Bucket: aws.String(a.bucket),
		Policy: aws.String(policy),
	})

	if err != nil {
		return fmt.Errorf("failed to set bucket policy: %w", err)
	}

	a.logger.Info("MinIO bucket policy updated",
		zap.String("bucket", a.bucket),
	)

	return nil
}

// GetBucketPolicy gets the bucket policy
func (a *MinIOAdapter) GetBucketPolicy(ctx context.Context) (string, error) {
	result, err := a.client.GetBucketPolicy(ctx, &s3.GetBucketPolicyInput{
		Bucket: aws.String(a.bucket),
	})

	if err != nil {
		return "", fmt.Errorf("failed to get bucket policy: %w", err)
	}

	return aws.ToString(result.Policy), nil
}

// Note: All other methods are inherited from S3Adapter
// This includes:
// - Store
// - Retrieve
// - Delete
// - Exists
// - GetSize
// - GetMetadata
// - Ping
// - GetCapacity
// - GetPresignedURL
// - GetPresignedUploadURL
// - ListFiles
// - GetStorageStats
// - Copy

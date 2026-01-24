package abstraction

import (
	"context"
	"time"
)

// Storage abstractions define contracts for external storage dependencies.
// These interfaces are implemented in the infrastructure layer.

// PostPolicyInput contains parameters for creating a POST policy
type PostPolicyInput struct {
	Key         string
	ContentType string
	Size        int64 // Exact file size in bytes (enforced by S3/MinIO)
}

// PostPolicyOutput contains the POST policy result for form-based upload
type PostPolicyOutput struct {
	URL        string            // URL to POST to
	FormData   map[string]string // Form fields to include with the upload
	TTLSeconds int
}

// Presigner creates presigned URLs for uploading objects
type Presigner interface {
	// CreatePostPolicy creates a POST policy that enforces size limits at S3/MinIO level
	CreatePostPolicy(ctx context.Context, input *PostPolicyInput) (*PostPolicyOutput, error)
}

// HeadObjectInput contains parameters for checking object metadata
type HeadObjectInput struct {
	Key string
}

// HeadObjectOutput contains object metadata
type HeadObjectOutput struct {
	ContentLength *int64
}

// DeleteObjectInput contains parameters for deleting an object
type DeleteObjectInput struct {
	Key string
}

// CopyObjectInput contains parameters for copying an object
type CopyObjectInput struct {
	SourceKey string
	TargetKey string
}

// ObjectStorage provides operations for object storage
type ObjectStorage interface {
	HeadObject(ctx context.Context, input *HeadObjectInput) (*HeadObjectOutput, error)
	DeleteObject(ctx context.Context, input *DeleteObjectInput) error
	DeleteObjects(ctx context.Context, keys []string) error
	CopyObject(ctx context.Context, input *CopyObjectInput) error
	// ObjectExists checks if an object exists at the given key.
	ObjectExists(ctx context.Context, key string) (bool, error)
}

// SignerOptions contains parameters for building image transformation URLs
type SignerOptions struct {
	Width   *int
	Height  *int
	Fit     *string // fit | fill | fill-down | force | auto
	Quality *int    // 1..100
	DPR     *float32
	Format  *string    // webp | avif | jpeg | png | "" (original)
	Expires *time.Time // expiration time for signed URLs
}

// ImgproxySigner builds signed URLs for image transformation service
type ImgproxySigner interface {
	BuildURL(key string, opts SignerOptions) string
}

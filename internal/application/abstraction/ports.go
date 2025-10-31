package abstraction

import (
	"context"
	"time"
)

// Storage abstractions define contracts for external storage dependencies.
// These interfaces are implemented in the infrastructure layer.

// PresignPutObjectInput contains parameters for presigning a PUT request
type PresignPutObjectInput struct {
	Key         string
	ContentType string
}

// PresignPutObjectOutput contains the result of presigning
type PresignPutObjectOutput struct {
	URL        string
	TTLSeconds int
}

// Presigner creates presigned URLs for uploading objects
type Presigner interface {
	PresignPutObject(ctx context.Context, input *PresignPutObjectInput) (*PresignPutObjectOutput, error)
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
	CopyObject(ctx context.Context, input *CopyObjectInput) error
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

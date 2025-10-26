package abstraction

import (
	"context"
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
	URL string
}

// Presigner creates presigned URLs for uploading objects
type Presigner interface {
	PresignPutObject(ctx context.Context, input *PresignPutObjectInput) (*PresignPutObjectOutput, error)
	GetPresignTTLSeconds() int
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
	Key        string
	CopySource string
}

// ObjectStorage provides operations for object storage
type ObjectStorage interface {
	HeadObject(ctx context.Context, input *HeadObjectInput) (*HeadObjectOutput, error)
	DeleteObject(ctx context.Context, input *DeleteObjectInput) error
	CopyObject(ctx context.Context, input *CopyObjectInput) error
	GetBucket() string
}

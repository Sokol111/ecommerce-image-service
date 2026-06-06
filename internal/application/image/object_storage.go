package image //nolint:revive // package name intentional

import "context"

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
	DeleteByPrefix(ctx context.Context, prefix string) error
	CopyObject(ctx context.Context, input *CopyObjectInput) error
	// ObjectExists checks if an object exists at the given key.
	ObjectExists(ctx context.Context, key string) (bool, error)
}

package s3

import (
	"context"
	"errors"

	"github.com/Sokol111/ecommerce-image-service/internal/application/image"
	"github.com/minio/minio-go/v7"
)

type objectStorage struct {
	client *minio.Client
	bucket string
}

// newObjectStorage creates a new ObjectStorage implementation.
func newObjectStorage(client *minio.Client, cfg Config) image.ObjectStorage {
	return &objectStorage{
		client: client,
		bucket: cfg.Bucket,
	}
}

func (o *objectStorage) HeadObject(ctx context.Context, input *image.HeadObjectInput) (*image.HeadObjectOutput, error) {
	info, err := o.client.StatObject(ctx, o.bucket, input.Key, minio.StatObjectOptions{})
	if err != nil {
		if isMinioNotFound(err) {
			return nil, errors.New("object not found")
		}
		return nil, err
	}

	size := info.Size
	return &image.HeadObjectOutput{
		ContentLength: &size,
	}, nil
}

func (o *objectStorage) DeleteObject(ctx context.Context, input *image.DeleteObjectInput) error {
	return o.client.RemoveObject(ctx, o.bucket, input.Key, minio.RemoveObjectOptions{})
}

func (o *objectStorage) DeleteObjects(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	objectsCh := make(chan minio.ObjectInfo, len(keys))
	go func() {
		defer close(objectsCh)
		for _, key := range keys {
			objectsCh <- minio.ObjectInfo{Key: key}
		}
	}()

	errCh := o.client.RemoveObjects(ctx, o.bucket, objectsCh, minio.RemoveObjectsOptions{})
	for err := range errCh {
		if err.Err != nil {
			return err.Err
		}
	}
	return nil
}

func (o *objectStorage) CopyObject(ctx context.Context, input *image.CopyObjectInput) error {
	src := minio.CopySrcOptions{
		Bucket: o.bucket,
		Object: input.SourceKey,
	}
	dst := minio.CopyDestOptions{
		Bucket: o.bucket,
		Object: input.TargetKey,
	}
	_, err := o.client.CopyObject(ctx, dst, src)
	return err
}

func (o *objectStorage) ObjectExists(ctx context.Context, key string) (bool, error) {
	_, err := o.client.StatObject(ctx, o.bucket, key, minio.StatObjectOptions{})
	if err != nil {
		if isMinioNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func isMinioNotFound(err error) bool {
	if err == nil {
		return false
	}
	errResp := minio.ToErrorResponse(err)
	return errResp.Code == "NoSuchKey" || errResp.Code == "NotFound"
}

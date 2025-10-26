package s3

import (
	"context"
	"errors"

	"github.com/Sokol111/ecommerce-image-service/internal/application/abstraction"
	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
)

type objectStorage struct {
	client *s3.Client
	bucket string
}

// NewObjectStorage creates a new ObjectStorage implementation
func NewObjectStorage(client *s3.Client, cfg Config) abstraction.ObjectStorage {
	return &objectStorage{
		client: client,
		bucket: cfg.Bucket,
	}
}

func (o *objectStorage) HeadObject(ctx context.Context, input *abstraction.HeadObjectInput) (*abstraction.HeadObjectOutput, error) {
	out, err := o.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(o.bucket),
		Key:    aws.String(input.Key),
	})
	if err != nil {
		// Convert S3 not found errors to nil response (object doesn't exist)
		if isS3NotFound(err) {
			return nil, errors.New("object not found")
		}
		return nil, err
	}

	return &abstraction.HeadObjectOutput{
		ContentLength: out.ContentLength,
	}, nil
}

func (o *objectStorage) DeleteObject(ctx context.Context, input *abstraction.DeleteObjectInput) error {
	_, err := o.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(o.bucket),
		Key:    aws.String(input.Key),
	})
	return err
}

func (o *objectStorage) CopyObject(ctx context.Context, input *abstraction.CopyObjectInput) error {
	_, err := o.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(o.bucket),
		Key:        aws.String(input.Key),
		CopySource: aws.String(input.CopySource),
	})
	return err
}

func (o *objectStorage) GetBucket() string {
	return o.bucket
}

func isS3NotFound(err error) bool {
	if err == nil {
		return false
	}
	var ae smithy.APIError
	if errors.As(err, &ae) {
		switch ae.ErrorCode() {
		case "NotFound", "NoSuchKey", "NoSuchBucket":
			return true
		}
	}
	var re *awshttp.ResponseError
	if errors.As(err, &re) && re.HTTPStatusCode() == 404 {
		return true
	}
	return false
}

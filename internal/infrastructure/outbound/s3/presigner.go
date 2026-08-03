package s3

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Sokol111/ecommerce-image-service/internal/application/image"
	"github.com/minio/minio-go/v7"
)

type presigner struct {
	client *minio.Client
	bucket string
	ttl    time.Duration
}

func NewPresigner(client *minio.Client, cfg Config) image.Presigner {
	return &presigner{
		client: client,
		bucket: cfg.Bucket,
		ttl:    cfg.PresignTTL,
	}
}

// CreatePresignedUpload creates a presigned PUT URL with Content-Type and Content-Length
// baked into the signature. S3/R2 will reject requests with mismatched headers.
func (p *presigner) CreatePresignedUpload(ctx context.Context, input *image.PresignedUploadInput) (*image.PresignedUploadOutput, error) {
	headers := http.Header{}
	headers.Set("Content-Type", input.ContentType)
	headers.Set("Content-Length", strconv.FormatInt(input.Size, 10))

	presignedURL, err := p.client.PresignHeader(ctx, "PUT", p.bucket, input.Key, p.ttl, nil, headers)
	if err != nil {
		return nil, fmt.Errorf("presign PUT: %w", err)
	}

	return &image.PresignedUploadOutput{
		URL:        presignedURL.String(),
		TTLSeconds: int(p.ttl.Seconds()),
	}, nil
}

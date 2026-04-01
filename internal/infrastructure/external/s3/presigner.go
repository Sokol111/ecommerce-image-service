package s3

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/Sokol111/ecommerce-image-service/internal/application"
	"github.com/Sokol111/ecommerce-image-service/internal/application/abstraction"
	"github.com/minio/minio-go/v7"
)

type presigner struct {
	minioClient    *minio.Client
	bucket         string
	ttl            time.Duration
	publicEndpoint string // if set, rewrite presigned URLs to use this endpoint
}

// newPresigner creates a new Presigner implementation.
// If public-endpoint is configured, presigned URLs are rewritten to use it
// so browsers can reach S3/MinIO directly.
func newPresigner(client *minio.Client, s3Cfg Config, appCfg application.Config) abstraction.Presigner {
	return &presigner{
		minioClient:    client,
		bucket:         s3Cfg.Bucket,
		ttl:            appCfg.PresignTTL,
		publicEndpoint: s3Cfg.PublicEndpoint,
	}
}

// CreatePresignedUpload creates a presigned PUT URL.
// Content-Type is NOT enforced at the storage level; size and type are verified
// server-side in the confirm step via HeadObject + JWT claims.
func (p *presigner) CreatePresignedUpload(ctx context.Context, input *abstraction.PresignedUploadInput) (*abstraction.PresignedUploadOutput, error) {
	presignedURL, err := p.minioClient.PresignedPutObject(ctx, p.bucket, input.Key, p.ttl)
	if err != nil {
		return nil, fmt.Errorf("presign PUT: %w", err)
	}

	// Rewrite URL to use public endpoint if configured
	if p.publicEndpoint != "" {
		pub, err := url.Parse(p.publicEndpoint)
		if err != nil {
			return nil, fmt.Errorf("failed to parse public endpoint %q: %w", p.publicEndpoint, err)
		}
		presignedURL.Scheme = pub.Scheme
		presignedURL.Host = pub.Host
	}

	return &abstraction.PresignedUploadOutput{
		URL:        presignedURL.String(),
		TTLSeconds: int(p.ttl.Seconds()),
	}, nil
}

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
	r2Compatible   bool   // if set, remove unsupported fields from POST policy
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
		r2Compatible:   s3Cfg.R2Compatible,
	}
}

// CreatePostPolicy creates a POST policy that enforces size limits at S3/MinIO level.
// This prevents attackers from uploading files larger than specified, even with a valid presigned URL.
func (p *presigner) CreatePostPolicy(ctx context.Context, input *abstraction.PostPolicyInput) (*abstraction.PostPolicyOutput, error) {
	policy := minio.NewPostPolicy()

	// Set bucket and key
	if err := policy.SetBucket(p.bucket); err != nil {
		return nil, err
	}
	if err := policy.SetKey(input.Key); err != nil {
		return nil, err
	}

	// Set expiration
	if err := policy.SetExpires(time.Now().Add(p.ttl)); err != nil {
		return nil, err
	}

	// Set content type
	if err := policy.SetContentType(input.ContentType); err != nil {
		return nil, err
	}

	// Set exact content length - S3/MinIO will reject uploads with different size
	if err := policy.SetContentLengthRange(input.Size, input.Size); err != nil {
		return nil, err
	}

	// Generate presigned POST policy
	presignedURL, formData, err := p.minioClient.PresignedPostPolicy(ctx, policy)
	if err != nil {
		return nil, err
	}

	// Cloudflare R2 does not support "bucket" as a POST policy field
	// and rejects it with InvalidArgument. The bucket is already part of the URL path.
	if p.r2Compatible {
		delete(formData, "bucket")
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

	return &abstraction.PostPolicyOutput{
		URL:        presignedURL.String(),
		FormData:   formData,
		TTLSeconds: int(p.ttl.Seconds()),
	}, nil
}

package s3

import (
	"context"
	"time"

	"github.com/Sokol111/ecommerce-image-service/internal/application"
	"github.com/Sokol111/ecommerce-image-service/internal/application/abstraction"
	"github.com/minio/minio-go/v7"
)

type presigner struct {
	minioClient *minio.Client
	bucket      string
	ttl         time.Duration
}

// newPresigner creates a new Presigner implementation
func newPresigner(minioClient *minio.Client, s3Cfg Config, appCfg application.Config) abstraction.Presigner {
	return &presigner{
		minioClient: minioClient,
		bucket:      s3Cfg.Bucket,
		ttl:         appCfg.PresignTTL,
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
	url, formData, err := p.minioClient.PresignedPostPolicy(ctx, policy)
	if err != nil {
		return nil, err
	}

	return &abstraction.PostPolicyOutput{
		URL:        url.String(),
		FormData:   formData,
		TTLSeconds: int(p.ttl.Seconds()),
	}, nil
}

package s3

import (
	"context"
	"strings"
	"time"

	"github.com/Sokol111/ecommerce-image-service/internal/application"
	"github.com/Sokol111/ecommerce-image-service/internal/application/abstraction"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type presigner struct {
	client       *s3.PresignClient
	bucket       string
	ttl          time.Duration
	internalHost string // e.g., "minio:9000" - prepared in config
	publicHost   string // e.g., "localhost:9000" - prepared in config
}

// newPresigner creates a new Presigner implementation
func newPresigner(client *s3.PresignClient, s3Cfg Config, appCfg application.Config) abstraction.Presigner {
	return &presigner{
		client:       client,
		bucket:       s3Cfg.Bucket,
		ttl:          appCfg.PresignTTL,
		internalHost: extractHost(s3Cfg.Endpoint),
		publicHost:   extractHost(s3Cfg.PublicEndpoint),
	}
}

func (p *presigner) PresignPutObject(ctx context.Context, input *abstraction.PresignPutObjectInput) (*abstraction.PresignPutObjectOutput, error) {
	out, err := p.client.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(p.bucket),
		Key:         aws.String(input.Key),
		ContentType: aws.String(input.ContentType),
	}, s3.WithPresignExpires(p.ttl))
	if err != nil {
		return nil, err
	}

	// Replace internal endpoint with public endpoint if configured
	// Example: "http://minio:9000/..." -> "http://localhost:9000/..."
	presignedURL := out.URL
	if p.publicHost != "" && p.internalHost != "" {
		presignedURL = strings.Replace(presignedURL, p.internalHost, p.publicHost, 1)
	}

	return &abstraction.PresignPutObjectOutput{
		URL:        presignedURL,
		TTLSeconds: int(p.ttl.Seconds()),
	}, nil
}

// extractHost removes http:// or https:// prefix from endpoint URL
func extractHost(endpoint string) string {
	host := strings.TrimPrefix(endpoint, "http://")
	host = strings.TrimPrefix(host, "https://")
	return host
}

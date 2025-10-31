package s3

import (
	"context"
	"time"

	"github.com/Sokol111/ecommerce-image-service/internal/application"
	"github.com/Sokol111/ecommerce-image-service/internal/application/abstraction"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type presigner struct {
	client *s3.PresignClient
	bucket string
	ttl    time.Duration
}

// newPresigner creates a new Presigner implementation
func newPresigner(client *s3.PresignClient, s3Cfg Config, appCfg application.Config) abstraction.Presigner {
	return &presigner{
		client: client,
		bucket: s3Cfg.Bucket,
		ttl:    appCfg.PresignTTL,
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

	return &abstraction.PresignPutObjectOutput{
		URL:        out.URL,
		TTLSeconds: int(p.ttl.Seconds()),
	}, nil
}

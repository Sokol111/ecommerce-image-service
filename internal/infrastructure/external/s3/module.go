package s3

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/fx"
)

func NewS3Module() fx.Option {
	return fx.Provide(
		newConfig,
		newAWSConfig,       // aws.Config
		newS3Client,        // *s3.Client
		newS3PresignClient, // *s3.PresignClient
		newPresigner,       // abstraction.Presigner (uses application.Config)
		newObjectStorage,   // abstraction.ObjectStorage (uses Config)
	)
}

func newAWSConfig(cfg Config) (aws.Config, error) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        cfg.MaxIdleConns,
			MaxIdleConnsPerHost: cfg.MaxIdleConnsPerHost,
			IdleConnTimeout:     cfg.IdleConnTimeout,
		},
		Timeout: cfg.HTTPTimeout,
	}

	loadOpts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithHTTPClient(httpClient),
	}

	if cfg.AccessKeyID == "" || cfg.SecretKey == "" {
		return aws.Config{}, fmt.Errorf("missing required S3 static credentials: access-key-id and secret-key must both be set")
	}
	loadOpts = append(loadOpts, awsconfig.WithCredentialsProvider(
		credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretKey, ""),
	))

	ctx := context.Background()
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, loadOpts...)
	if err != nil {
		return aws.Config{}, fmt.Errorf("load aws config: %w", err)
	}

	return awsCfg, nil
}

func newS3Client(cfg Config, awsCfg aws.Config) *s3.Client {
	opts := func(o *s3.Options) {
		if cfg.Endpoint != "" { // MinIO / кастомний S3 сумісний сторедж
			o.BaseEndpoint = aws.String(cfg.Endpoint) // сучасний спосіб заміни endpoint без глобального резолвера
		}
		if cfg.UsePathStyle { // MinIO зазвичай потребує path-style
			o.UsePathStyle = true
		}
	}
	return s3.NewFromConfig(awsCfg, opts)
}

func newS3PresignClient(c *s3.Client) *s3.PresignClient {
	return s3.NewPresignClient(c)
}

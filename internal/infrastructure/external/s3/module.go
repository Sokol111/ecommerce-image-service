package s3

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/fx"
)

func NewS3Module() fx.Option {
	return fx.Provide(
		newConfig,
		newMinioClient,   // *minio.Client
		newPresigner,     // abstraction.Presigner
		newObjectStorage, // abstraction.ObjectStorage
	)
}

// newMinioClient creates a MinIO client that works with any S3-compatible storage
func newMinioClient(cfg Config) (*minio.Client, error) {
	if cfg.AccessKeyID == "" || cfg.SecretKey == "" {
		return nil, fmt.Errorf("missing required S3 credentials: access-key-id and secret-key must both be set")
	}

	// Use public endpoint for presigned URLs if configured
	endpoint := cfg.Endpoint
	if cfg.PublicEndpoint != "" {
		endpoint = cfg.PublicEndpoint
	}

	// Parse endpoint to extract host and scheme
	host, secure := parseEndpoint(endpoint)

	// Create custom HTTP transport with connection pooling
	transport := &http.Transport{
		MaxIdleConns:        cfg.MaxIdleConns,
		MaxIdleConnsPerHost: cfg.MaxIdleConnsPerHost,
		IdleConnTimeout:     cfg.IdleConnTimeout,
	}

	return minio.New(host, &minio.Options{
		Creds:     credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretKey, ""),
		Secure:    secure,
		Region:    cfg.Region,
		Transport: transport,
	})
}

// parseEndpoint extracts host and determines if HTTPS from endpoint URL
func parseEndpoint(endpoint string) (host string, secure bool) {
	if strings.HasPrefix(endpoint, "https://") {
		return strings.TrimPrefix(endpoint, "https://"), true
	}
	return strings.TrimPrefix(endpoint, "http://"), false
}

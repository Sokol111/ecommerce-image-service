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
		newMinioClient,   // *minio.Client — uses internal endpoint (e.g. http://minio:9000)
		newPresigner,     // abstraction.Presigner
		newObjectStorage, // abstraction.ObjectStorage
	)
}

// newMinioClient creates a MinIO client using the internal endpoint for all S3 operations.
func newMinioClient(cfg Config) (*minio.Client, error) {
	if cfg.AccessKeyID == "" || cfg.SecretKey == "" {
		return nil, fmt.Errorf("missing required S3 credentials: access-key-id and secret-key must both be set")
	}

	// Parse endpoint to extract host and scheme
	host, secure := parseEndpoint(cfg.Endpoint)

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

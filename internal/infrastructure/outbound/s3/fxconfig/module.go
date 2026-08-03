package fxconfig

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Sokol111/ecommerce-commons/pkg/core/config"
	"github.com/Sokol111/ecommerce-commons/pkg/tenant"
	"github.com/Sokol111/ecommerce-image-service/internal/application/image"
	"github.com/Sokol111/ecommerce-image-service/internal/infrastructure/outbound/s3"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/fx"
)

func NewS3Module() fx.Option {
	return fx.Options(
		fx.Provide(provideConfig),
		fx.Provide(func(cfg s3.Config) image.PresignTTLProvider { return cfg }),
		fx.Provide(fx.Annotate(s3.NewImageTenantCleaner,
			fx.As(new(tenant.Cleaner)),
			fx.ResultTags(`group:"tenant_cleaners"`),
		)),
		newStorageModule(),
		newPresignerModule(),
	)
}

func provideConfig(loader *config.Loader) (s3.Config, error) {
	return config.Load[s3.Config](loader, "s3", nil)
}

// newStorageModule provides ObjectStorage with an internal MinIO client.
func newStorageModule() fx.Option {
	return fx.Module("s3-storage",
		fx.Provide(fx.Private, newInternalClient),
		fx.Provide(s3.NewObjectStorage),
	)
}

// newPresignerModule provides Presigner with a public-endpoint MinIO client
// so that presigned URL signatures match the Host header browsers send.
func newPresignerModule() fx.Option {
	return fx.Module("s3-presigner",
		fx.Provide(fx.Private, newPublicClient),
		fx.Provide(s3.NewPresigner),
	)
}

// newInternalClient creates a MinIO client using the internal endpoint for server-to-server S3 operations.
func newInternalClient(cfg s3.Config) (*minio.Client, error) {
	if cfg.AccessKeyID == "" || cfg.SecretKey == "" {
		return nil, fmt.Errorf("missing required S3 credentials: access-key-id and secret-key must both be set")
	}

	host, secure := parseEndpoint(cfg.Endpoint)

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

// newPublicClient creates a MinIO client using the public endpoint so that
// presigned URL signatures include the correct Host for browser requests.
func newPublicClient(cfg s3.Config) (*minio.Client, error) {
	endpoint := cfg.PublicEndpoint
	if endpoint == "" {
		endpoint = cfg.Endpoint
	}

	host, secure := parseEndpoint(endpoint)

	return minio.New(host, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretKey, ""),
		Secure: secure,
		Region: cfg.Region,
	})
}

// parseEndpoint extracts host and determines if HTTPS from endpoint URL
func parseEndpoint(endpoint string) (host string, secure bool) {
	if strings.HasPrefix(endpoint, "https://") {
		return strings.TrimPrefix(endpoint, "https://"), true
	}
	return strings.TrimPrefix(endpoint, "http://"), false
}

package s3

import (
	"fmt"
	"time"

	"github.com/knadh/koanf/v2"
)

type Config struct {
	// Core
	Bucket         string `koanf:"bucket"`          // Target bucket (e.g., "products")
	Region         string `koanf:"region"`          // e.g., "us-east-1"; MinIO accepts any non-empty value
	Endpoint       string `koanf:"endpoint"`        // e.g., "http://minio:9000" or leave empty for AWS S3
	PublicEndpoint string `koanf:"public-endpoint"` // e.g., "http://localhost:9000" - endpoint for browser-accessible presigned URLs
	UsePathStyle   bool   `koanf:"use-path-style"`  // MinIO: true; AWS S3: false
	AccessKeyID    string `koanf:"access-key-id"`   // MinIO/AWS access key
	SecretKey      string `koanf:"secret-key"`      // MinIO/AWS secret key

	// Client tuning
	HTTPTimeout         time.Duration // default 30s if zero
	MaxIdleConns        int           // default 100 if zero
	MaxIdleConnsPerHost int           // default 100 if zero
	IdleConnTimeout     time.Duration // default 90s if zero
}

func newConfig(k *koanf.Koanf) (Config, error) {
	var cfg Config
	if err := k.Unmarshal("s3", &cfg); err != nil {
		return cfg, fmt.Errorf("failed to load s3 config: %w", err)
	}
	if cfg.HTTPTimeout == 0 {
		cfg.HTTPTimeout = 30 * time.Second
	}
	if cfg.MaxIdleConns == 0 {
		cfg.MaxIdleConns = 100
	}
	if cfg.MaxIdleConnsPerHost == 0 {
		cfg.MaxIdleConnsPerHost = 100
	}
	if cfg.IdleConnTimeout == 0 {
		cfg.IdleConnTimeout = 90 * time.Second
	}
	return cfg, nil
}

// GetBucket returns the S3 bucket name
func (c Config) GetBucket() string {
	return c.Bucket
}

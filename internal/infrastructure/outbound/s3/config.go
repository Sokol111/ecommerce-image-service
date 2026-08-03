package s3

import (
	"fmt"
	"time"
)

type Config struct {
	// Core
	Bucket         string `koanf:"bucket"`          // Target bucket (e.g., "images")
	Region         string `koanf:"region"`          // e.g., "us-east-1"; MinIO accepts any non-empty value
	Endpoint       string `koanf:"endpoint"`        // e.g., "http://minio:9000" or leave empty for AWS S3
	PublicEndpoint string `koanf:"public-endpoint"` // e.g., "http://localhost:9000" - endpoint for browser-accessible presigned URLs
	UsePathStyle   bool   `koanf:"use-path-style"`  // MinIO: true; AWS S3: false
	AccessKeyID    string `koanf:"access-key-id"`   // MinIO/AWS access key
	SecretKey      string `koanf:"secret-key"`      // MinIO/AWS secret key

	// Presigning
	PresignTTL time.Duration `koanf:"presign-ttl"` // Presigned URL validity duration

	// Client tuning
	HTTPTimeout         time.Duration // default 30s if zero
	MaxIdleConns        int           // default 100 if zero
	MaxIdleConnsPerHost int           // default 100 if zero
	IdleConnTimeout     time.Duration // default 90s if zero
}

func (c *Config) ApplyDefaults() {
	if c.PresignTTL == 0 {
		c.PresignTTL = 15 * time.Minute
	}
	if c.HTTPTimeout == 0 {
		c.HTTPTimeout = 30 * time.Second
	}
	if c.MaxIdleConns == 0 {
		c.MaxIdleConns = 100
	}
	if c.MaxIdleConnsPerHost == 0 {
		c.MaxIdleConnsPerHost = 100
	}
	if c.IdleConnTimeout == 0 {
		c.IdleConnTimeout = 90 * time.Second
	}
}

func (c *Config) Validate() error {
	if c.Bucket == "" {
		return fmt.Errorf("s3 bucket is required")
	}
	if c.Region == "" {
		return fmt.Errorf("s3 region is required")
	}
	if c.PublicEndpoint == "" {
		return fmt.Errorf("s3 public endpoint is required")
	}
	if c.AccessKeyID == "" {
		return fmt.Errorf("s3 access key ID is required")
	}
	if c.SecretKey == "" {
		return fmt.Errorf("s3 secret key is required")
	}
	if c.PresignTTL <= 0 {
		return fmt.Errorf("s3 presign TTL must be greater than zero")
	}
	if c.HTTPTimeout <= 0 {
		return fmt.Errorf("s3 HTTP timeout must be greater than zero")
	}
	if c.MaxIdleConns <= 0 {
		return fmt.Errorf("s3 max idle connections must be greater than zero")
	}
	if c.MaxIdleConnsPerHost <= 0 {
		return fmt.Errorf("s3 max idle connections per host must be greater than zero")
	}
	if c.IdleConnTimeout <= 0 {
		return fmt.Errorf("s3 idle connection timeout must be greater than zero")
	}
	return nil
}

// GetPresignTTL returns the presign TTL for use by other modules (e.g., token generation)
func (c Config) GetPresignTTL() time.Duration {
	return c.PresignTTL
}

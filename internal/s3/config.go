package s3

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	// Core
	Bucket       string `mapstructure:"bucket"`         // Target bucket (e.g., "products")
	Region       string `mapstructure:"region"`         // e.g., "us-east-1"; MinIO accepts any non-empty value
	Endpoint     string `mapstructure:"endpoint"`       // e.g., "http://minio.minio.svc.cluster.local:9000" or leave empty for AWS S3
	UsePathStyle bool   `mapstructure:"use-path-style"` // MinIO: true; AWS S3: false
	AccessKeyID  string `mapstructure:"access-key-id"`  // MinIO/AWS access key
	SecretKey    string `mapstructure:"secret-key"`     // MinIO/AWS secret key

	// Client tuning
	HTTPTimeout         time.Duration // default 30s if zero
	MaxIdleConns        int           // default 100 if zero
	MaxIdleConnsPerHost int           // default 100 if zero
	IdleConnTimeout     time.Duration // default 90s if zero

	// App semantics
	PresignTTL     time.Duration // e.g., 15 * time.Minute (default if zero)
	MaxUploadBytes int64         // optional size guard (app-level; OpenAPI can also enforce)
}

func newConfig(v *viper.Viper) (Config, error) {
	var cfg Config
	if err := v.Sub("s3").UnmarshalExact(&cfg); err != nil {
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
	if cfg.PresignTTL == 0 {
		cfg.PresignTTL = 15 * time.Minute
	}
	if cfg.MaxUploadBytes == 0 {
		cfg.MaxUploadBytes = 1 * 1024 * 1024 // 1 MB
	}
	return cfg, nil
}

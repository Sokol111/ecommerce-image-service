package application

import (
	"fmt"
	"time"

	"github.com/knadh/koanf/v2"
)

// Config holds application-level configuration
type Config struct {
	// PresignTTL is the time-to-live for presigned upload URLs
	PresignTTL time.Duration `koanf:"presign-ttl"`

	// MaxUploadBytes is the maximum allowed file upload size in bytes
	MaxUploadBytes int64 `koanf:"max-upload-bytes"`

	// JWTSecret is the secret key used to sign upload tokens
	JWTSecret string `koanf:"jwt-secret"`

	// Promote holds configuration for image promotion
	Promote PromoteConfig `koanf:"promote"`
}

// PromoteConfig holds configuration for image promotion
type PromoteConfig struct {
	// SmallWidth is the width for small product images
	SmallWidth int `koanf:"small-width"`
	// LargeWidth is the width for large product images
	LargeWidth int `koanf:"large-width"`
	// Quality is the JPEG quality for product images (1-100)
	Quality int `koanf:"quality"`
}

// NewConfig creates a new application config from Koanf
func NewConfig(k *koanf.Koanf) (Config, error) {
	var cfg Config
	if err := k.Unmarshal("application", &cfg); err != nil {
		return cfg, fmt.Errorf("failed to load application config: %w", err)
	}

	// Set defaults
	if cfg.PresignTTL == 0 {
		cfg.PresignTTL = 15 * time.Minute
	}
	if cfg.MaxUploadBytes == 0 {
		cfg.MaxUploadBytes = 5 * 1024 * 1024 // 5 MB default
	}
	if cfg.JWTSecret == "" {
		return cfg, fmt.Errorf("jwt-secret is required")
	}

	// Promote defaults
	if cfg.Promote.SmallWidth == 0 {
		cfg.Promote.SmallWidth = 400
	}
	if cfg.Promote.LargeWidth == 0 {
		cfg.Promote.LargeWidth = 800
	}
	if cfg.Promote.Quality == 0 {
		cfg.Promote.Quality = 85
	}

	return cfg, nil
}

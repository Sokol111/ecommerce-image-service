package application

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds application-level configuration
type Config struct {
	// PresignTTL is the time-to-live for presigned upload URLs
	PresignTTL time.Duration `mapstructure:"presign-ttl"`

	// MaxUploadBytes is the maximum allowed file upload size in bytes
	MaxUploadBytes int64 `mapstructure:"max-upload-bytes"`

	// JWTSecret is the secret key used to sign upload tokens
	JWTSecret string `mapstructure:"jwt-secret"`

	// Promote holds configuration for image promotion
	Promote PromoteConfig `mapstructure:"promote"`
}

// PromoteConfig holds configuration for image promotion
type PromoteConfig struct {
	// SmallWidth is the width for small product images
	SmallWidth int `mapstructure:"small-width"`
	// LargeWidth is the width for large product images
	LargeWidth int `mapstructure:"large-width"`
	// Quality is the JPEG quality for product images (1-100)
	Quality int `mapstructure:"quality"`
}

// NewConfig creates a new application config from Viper
func NewConfig(v *viper.Viper) (Config, error) {
	var cfg Config
	if err := v.Sub("application").UnmarshalExact(&cfg); err != nil {
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

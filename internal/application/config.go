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

	return cfg, nil
}

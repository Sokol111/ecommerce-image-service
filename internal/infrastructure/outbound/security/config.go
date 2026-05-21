package security

import (
	"fmt"

	"github.com/knadh/koanf/v2"
)

type Config struct {
	JWTSecret string `koanf:"jwt-secret"`
}

func newConfig(k *koanf.Koanf) (Config, error) {
	var cfg Config
	if err := k.Unmarshal("upload-token", &cfg); err != nil {
		return cfg, fmt.Errorf("failed to load upload-token config: %w", err)
	}
	if cfg.JWTSecret == "" {
		return cfg, fmt.Errorf("upload-token.jwt-secret is required")
	}
	return cfg, nil
}

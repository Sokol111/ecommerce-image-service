package infrastructure

import (
	"fmt"

	"github.com/Sokol111/ecommerce-image-service/internal/application/image"
	"github.com/knadh/koanf/v2"
	"go.uber.org/fx"
)

// ConfigModule provides application configuration
func ConfigModule() fx.Option {
	return fx.Provide(newImageConfig)
}

type rawImageConfig struct {
	MaxUploadBytes int64 `koanf:"max-upload-bytes"`
	SmallWidth     int   `koanf:"small-width"`
	LargeWidth     int   `koanf:"large-width"`
	Quality        int   `koanf:"quality"`
}

func newImageConfig(k *koanf.Koanf) (image.Config, error) {
	var raw rawImageConfig
	if err := k.Unmarshal("application", &raw); err != nil {
		return image.Config{}, fmt.Errorf("failed to load application config: %w", err)
	}

	cfg := image.Config(raw)

	// Set defaults
	if cfg.MaxUploadBytes == 0 {
		cfg.MaxUploadBytes = 5 * 1024 * 1024 // 5 MB default
	}
	if cfg.SmallWidth == 0 {
		cfg.SmallWidth = 400
	}
	if cfg.LargeWidth == 0 {
		cfg.LargeWidth = 800
	}
	if cfg.Quality == 0 {
		cfg.Quality = 85
	}

	return cfg, nil
}

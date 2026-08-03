package token

import (
	"github.com/Sokol111/ecommerce-commons/pkg/core/config"
	"github.com/Sokol111/ecommerce-image-service/internal/application/image"
	"github.com/Sokol111/ecommerce-image-service/internal/infrastructure/outbound/token"
	"go.uber.org/fx"
)

// NewTokenModule provides security infrastructure
func NewTokenModule() fx.Option {
	return fx.Options(
		fx.Provide(
			provideConfig,
			func(cfg token.Config) image.TokenService {
				return token.NewJWTTokenService(cfg.JWTSecret)
			},
		),
	)
}

func provideConfig(loader *config.Loader) (token.Config, error) {
	return config.Load[token.Config](loader, "upload-token", nil)
}

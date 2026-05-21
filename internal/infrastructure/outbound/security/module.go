package security

import (
	"github.com/Sokol111/ecommerce-image-service/internal/application/image"
	"go.uber.org/fx"
)

// NewSecurityModule provides security infrastructure
func NewSecurityModule() fx.Option {
	return fx.Module("security",
		fx.Provide(
			newConfig,
			func(cfg Config) image.TokenService {
				return NewJWTTokenService(cfg.JWTSecret)
			},
		),
	)
}

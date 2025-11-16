package security

import (
	"github.com/Sokol111/ecommerce-image-service/internal/application"
	"github.com/Sokol111/ecommerce-image-service/internal/application/abstraction"
	"go.uber.org/fx"
)

// NewSecurityModule provides security infrastructure
func NewSecurityModule() fx.Option {
	return fx.Module("security",
		fx.Provide(
			func(cfg application.Config) abstraction.TokenService {
				return NewJWTTokenService(cfg.JWTSecret)
			},
		),
	)
}

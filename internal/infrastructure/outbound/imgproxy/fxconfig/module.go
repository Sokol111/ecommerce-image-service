package fxconfig

import (
	"github.com/Sokol111/ecommerce-commons/pkg/core/config"
	"github.com/Sokol111/ecommerce-image-service/internal/infrastructure/outbound/imgproxy"
	"go.uber.org/fx"
)

func NewImgProxyModule() fx.Option {
	return fx.Provide(
		provideConfig,
		imgproxy.NewImgproxySigner,
	)
}

func provideConfig(loader *config.Loader) (imgproxy.Config, error) {
	return config.Load[imgproxy.Config](loader, "imgproxy", nil)
}

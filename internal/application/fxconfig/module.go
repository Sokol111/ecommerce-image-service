package fxconfig

import (
	"github.com/Sokol111/ecommerce-commons/pkg/core/config"
	"github.com/Sokol111/ecommerce-image-service/internal/application/image"
	"go.uber.org/fx"
)

// NewAppModule provides application layer dependencies
func NewAppModule() fx.Option {
	return fx.Options(
		fx.Provide(provideConfig),
		// Command handlers
		fx.Provide(
			image.NewCreatePresignHandler,
			image.NewConfirmUploadHandler,
			image.NewPromoteImagesHandler,
			image.NewDeleteImageHandler,
			image.NewCleanupOwnerImagesHandler,
		),
		// Query handlers
		fx.Provide(
			image.NewGetImageByIDHandler,
			image.NewGetDeliveryURLHandler,
		),
	)
}

func provideConfig(loader *config.Loader) (image.Config, error) {
	return config.Load[image.Config](loader, "application", nil)
}

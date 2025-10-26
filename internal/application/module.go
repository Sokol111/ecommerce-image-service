package application

import (
	"github.com/Sokol111/ecommerce-image-service/internal/application/command"
	"github.com/Sokol111/ecommerce-image-service/internal/application/query"
	"go.uber.org/fx"
)

// Module provides application layer dependencies
func Module() fx.Option {
	return fx.Options(
		// Config
		fx.Provide(
			NewConfig,
			fx.Annotate(
				newMaxUploadBytes,
				fx.ResultTags(`name:"maxUploadBytes"`),
			),
		),
		// Command handlers
		fx.Provide(
			command.NewCreatePresignHandler,
			fx.Annotate(
				command.NewConfirmUploadHandler,
				fx.ParamTags(``, ``, `name:"maxUploadBytes"`),
			),
			command.NewPromoteImagesHandler,
			command.NewDeleteImageHandler,
		),
		// Query handlers
		fx.Provide(
			query.NewGetImageByIDHandler,
			query.NewGetDeliveryURLHandler,
		),
	)
}

func newMaxUploadBytes(cfg Config) int64 {
	return cfg.MaxUploadBytes
}

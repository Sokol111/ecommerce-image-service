package application

import (
	"github.com/Sokol111/ecommerce-image-service/internal/application/abstraction"
	"github.com/Sokol111/ecommerce-image-service/internal/application/command"
	"github.com/Sokol111/ecommerce-image-service/internal/application/query"
	"github.com/Sokol111/ecommerce-image-service/internal/domain/image"
	"go.uber.org/fx"
)

// Module provides application layer dependencies
func Module() fx.Option {
	return fx.Options(
		// Config
		fx.Provide(
			NewConfig,
		),
		// Command handlers
		fx.Provide(
			command.NewCreatePresignHandler,
			func(repo image.Repository, storage abstraction.ObjectStorage, cfg Config) command.ConfirmUploadCommandHandler {
				return command.NewConfirmUploadHandler(repo, storage, cfg.MaxUploadBytes)
			},
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

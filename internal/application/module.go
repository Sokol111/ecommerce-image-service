package application

import (
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
	"github.com/Sokol111/ecommerce-commons/pkg/persistence"
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
			func(presigner abstraction.Presigner, tokenService abstraction.TokenService, cfg Config) command.CreatePresignCommandHandler {
				return command.NewCreatePresignHandler(presigner, tokenService, cfg.PresignTTL, cfg.MaxUploadBytes)
			},
			func(repo image.Repository, storage abstraction.ObjectStorage, tokenService abstraction.TokenService, cfg Config) command.ConfirmUploadCommandHandler {
				return command.NewConfirmUploadHandler(repo, storage, tokenService, cfg.MaxUploadBytes)
			},
			func(repo image.Repository, objStorage abstraction.ObjectStorage, signer abstraction.ImgproxySigner, outbox outbox.Outbox, txManager persistence.TxManager, cfg Config) command.PromoteImagesCommandHandler {
				return command.NewPromoteImagesHandler(repo, objStorage, signer, outbox, txManager, cfg.Promote.SmallWidth, cfg.Promote.LargeWidth, cfg.Promote.Quality)
			},
			command.NewDeleteImageHandler,
		),
		// Query handlers
		fx.Provide(
			query.NewGetImageByIDHandler,
			query.NewGetDeliveryURLHandler,
		),
	)
}

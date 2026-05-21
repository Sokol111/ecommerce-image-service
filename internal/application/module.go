package application

import (
	"time"

	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
	"github.com/Sokol111/ecommerce-commons/pkg/persistence/mongo"
	"github.com/Sokol111/ecommerce-image-service/internal/application/image"
	"go.uber.org/fx"
)

// PresignTTLProvider provides the presign TTL for upload token generation
type PresignTTLProvider interface {
	GetPresignTTL() time.Duration
}

// Module provides application layer dependencies
func Module() fx.Option {
	return fx.Options(
		// Config
		fx.Provide(
			NewConfig,
		),
		// Command handlers
		fx.Provide(
			func(presigner image.Presigner, tokenService image.TokenService, ttlProvider PresignTTLProvider, cfg Config) image.CreatePresignCommandHandler {
				return image.NewCreatePresignHandler(presigner, tokenService, ttlProvider.GetPresignTTL(), cfg.MaxUploadBytes)
			},
			func(repo image.Repository, storage image.ObjectStorage, tokenService image.TokenService, cfg Config) image.ConfirmUploadCommandHandler {
				return image.NewConfirmUploadHandler(repo, storage, tokenService, cfg.MaxUploadBytes)
			},
			func(repo image.Repository, objStorage image.ObjectStorage, signer image.ImgproxySigner, outbox outbox.Outbox, txManager mongo.TxManager, eventFactory image.ImageEventFactory, cfg Config) image.PromoteImagesCommandHandler {
				return image.NewPromoteImagesHandler(repo, objStorage, signer, outbox, txManager, eventFactory, cfg.Promote.SmallWidth, cfg.Promote.LargeWidth, cfg.Promote.Quality)
			},
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

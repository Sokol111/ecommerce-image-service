package kafka

import (
	"context"

	"github.com/Sokol111/ecommerce-catalog-service-api/gen/events"
	"github.com/Sokol111/ecommerce-image-service/internal/application/image"
)

type productHandler struct {
	promoteImagesHandler image.PromoteImagesCommandHandler
	cleanupImagesHandler image.CleanupOwnerImagesCommandHandler
}

func newProductHandler(promoteImages image.PromoteImagesCommandHandler, cleanupImages image.CleanupOwnerImagesCommandHandler) *productHandler {
	return &productHandler{
		promoteImagesHandler: promoteImages,
		cleanupImagesHandler: cleanupImages,
	}
}

func (h *productHandler) HandleProductUpdated(ctx context.Context, e *events.ProductUpdatedEvent) error {
	if e.Payload.ImageID == nil {
		return h.cleanupImagesHandler.Handle(ctx, image.CleanupOwnerImagesCommand{
			OwnerType: image.OwnerTypeProduct,
			OwnerID:   e.Payload.ProductID,
		})
	}

	cmd := image.PromoteImagesCommand{
		ImageIDs:  &[]string{*e.Payload.ImageID},
		OwnerType: image.OwnerTypeProduct,
		OwnerID:   e.Payload.ProductID,
	}

	_, err := h.promoteImagesHandler.Handle(ctx, cmd)
	return err
}

func (h *productHandler) HandleProductDeleted(ctx context.Context, e *events.ProductDeletedEvent) error {
	return h.cleanupImagesHandler.Handle(ctx, image.CleanupOwnerImagesCommand{
		OwnerType: image.OwnerTypeProduct,
		OwnerID:   e.Payload.ProductID,
	})
}

package kafka

import (
	"context"

	eventsv1 "github.com/Sokol111/ecommerce-catalog-service-api/gen/catalog/events/v1"
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

func (h *productHandler) HandleProductUpdated(ctx context.Context, e *eventsv1.ProductUpdatedEvent) error {
	if e.ImageId == nil {
		return h.cleanupImagesHandler.Handle(ctx, image.CleanupOwnerImagesCommand{
			OwnerType: image.OwnerTypeProduct,
			OwnerID:   e.ProductId,
		})
	}

	cmd := image.PromoteImagesCommand{
		ImageIDs:  &[]string{*e.ImageId},
		OwnerType: image.OwnerTypeProduct,
		OwnerID:   e.ProductId,
	}

	_, err := h.promoteImagesHandler.Handle(ctx, cmd)
	return err
}

func (h *productHandler) HandleProductDeleted(ctx context.Context, e *eventsv1.ProductDeletedEvent) error {
	return h.cleanupImagesHandler.Handle(ctx, image.CleanupOwnerImagesCommand{
		OwnerType: image.OwnerTypeProduct,
		OwnerID:   e.ProductId,
	})
}

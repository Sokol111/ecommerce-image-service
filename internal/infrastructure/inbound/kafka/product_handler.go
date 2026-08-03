package kafka

import (
	"context"

	eventsv1 "github.com/Sokol111/ecommerce-catalog-service-api/gen/go/catalog/events/v1"
	"github.com/Sokol111/ecommerce-image-service/internal/application/image"
)

type ProductHandler struct {
	promoteImagesHandler image.PromoteImagesCommandHandler
	cleanupImagesHandler image.CleanupOwnerImagesCommandHandler
}

func NewProductHandler(promoteImages image.PromoteImagesCommandHandler, cleanupImages image.CleanupOwnerImagesCommandHandler) *ProductHandler {
	return &ProductHandler{
		promoteImagesHandler: promoteImages,
		cleanupImagesHandler: cleanupImages,
	}
}

func (h *ProductHandler) HandleProductUpdated(ctx context.Context, e *eventsv1.ProductUpdatedEvent) error {
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

func (h *ProductHandler) HandleProductDeleted(ctx context.Context, e *eventsv1.ProductDeletedEvent) error {
	return h.cleanupImagesHandler.Handle(ctx, image.CleanupOwnerImagesCommand{
		OwnerType: image.OwnerTypeProduct,
		OwnerID:   e.ProductId,
	})
}

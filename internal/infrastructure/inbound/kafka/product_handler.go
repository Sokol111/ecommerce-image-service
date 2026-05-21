package kafka

import (
	"context"
	"fmt"

	"github.com/Sokol111/ecommerce-catalog-service-api/gen/events"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/kafka/consumer"
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

func (h *productHandler) Process(ctx context.Context, event any) error {
	switch evt := event.(type) {
	case *events.ProductUpdatedEvent:
		return h.handleProductUpdated(ctx, evt)
	case *events.ProductDeletedEvent:
		return h.handleProductDeleted(ctx, evt)
	default:
		return fmt.Errorf("unhandled event type: %T: %w", event, consumer.ErrSkipMessage)
	}
}

func (h *productHandler) handleProductUpdated(ctx context.Context, e *events.ProductUpdatedEvent) error {
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

func (h *productHandler) handleProductDeleted(ctx context.Context, e *events.ProductDeletedEvent) error {
	return h.cleanupImagesHandler.Handle(ctx, image.CleanupOwnerImagesCommand{
		OwnerType: image.OwnerTypeProduct,
		OwnerID:   e.Payload.ProductID,
	})
}

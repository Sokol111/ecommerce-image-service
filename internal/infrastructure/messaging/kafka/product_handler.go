package kafka

import (
	"context"
	"fmt"

	"github.com/Sokol111/ecommerce-catalog-service-api/gen/events"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/kafka/consumer"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
	"github.com/Sokol111/ecommerce-image-service/internal/application/command"
	"github.com/Sokol111/ecommerce-image-service/internal/domain/image"
	imageevent "github.com/Sokol111/ecommerce-image-service/internal/event"
)

type productHandler struct {
	promoteImagesHandler command.PromoteImagesCommandHandler
	cleanupImagesHandler command.CleanupOwnerImagesCommandHandler
}

func newProductHandler(promoteImages command.PromoteImagesCommandHandler, cleanupImages command.CleanupOwnerImagesCommandHandler) *productHandler {
	return &productHandler{
		promoteImagesHandler: promoteImages,
		cleanupImagesHandler: cleanupImages,
	}
}

func (h *productHandler) Process(ctx context.Context, event any) error {
	switch evt := event.(type) {
	case *events.ProductUpdatedEvent:
		return h.handleProductUpdated(ctx, evt)
	default:
		return fmt.Errorf("unhandled event type: %T: %w", event, consumer.ErrSkipMessage)
	}
}

func (h *productHandler) handleProductUpdated(ctx context.Context, e *events.ProductUpdatedEvent) error {
	if e.Payload.ImageID == nil {
		return h.cleanupImagesHandler.Handle(ctx, command.CleanupOwnerImagesCommand{
			OwnerType: image.OwnerTypeProduct,
			OwnerID:   e.Payload.ProductID,
		})
	}

	cmd := command.PromoteImagesCommand{
		ImageIDs:  &[]string{*e.Payload.ImageID},
		OwnerType: image.OwnerTypeProduct,
		OwnerID:   e.Payload.ProductID,
		OnPromoted: func(ctx context.Context, ownerID string, images []command.PromotedImage) ([]outbox.Message, error) {
			msgs := make([]outbox.Message, 0, len(images))
			for _, img := range images {
				msgs = append(msgs, imageevent.NewProductImagePromotedOutboxMessage(ctx, ownerID, img.ImageID, img.SmallImageURL, img.LargeImageURL))
			}
			return msgs, nil
		},
	}

	_, err := h.promoteImagesHandler.Handle(ctx, cmd)
	return err
}

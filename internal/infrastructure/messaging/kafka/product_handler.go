package kafka

import (
	"context"
	"fmt"

	"github.com/Sokol111/ecommerce-catalog-service-api/gen/events"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/kafka/consumer"
	"github.com/Sokol111/ecommerce-image-service/internal/application/command"
)

type productHandler struct {
	promoteImagesHandler command.PromoteImagesCommandHandler
}

func newProductHandler(promoteImages command.PromoteImagesCommandHandler) *productHandler {
	return &productHandler{
		promoteImagesHandler: promoteImages,
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
	var imageIDs *[]string
	if e.Payload.ImageID != nil {
		imageIDs = &[]string{*e.Payload.ImageID}
	}

	cmd := command.PromoteImagesCommand{
		DraftID:   e.Payload.ProductID,
		ImageIDs:  imageIDs,
		ProductID: e.Payload.ProductID,
	}

	_, err := h.promoteImagesHandler.Handle(ctx, cmd)
	return err
}

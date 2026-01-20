package kafka

import (
	"context"
	"fmt"

	"github.com/Sokol111/ecommerce-catalog-service-api/gen/events"
	"github.com/Sokol111/ecommerce-commons/pkg/core/logger"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/kafka/consumer"
	"github.com/Sokol111/ecommerce-image-service/internal/application/command"
	"go.uber.org/zap"
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
	case *events.ProductCreatedEvent:
		return h.handleProductCreated(ctx, evt)
	case *events.ProductUpdatedEvent:
		return h.handleProductUpdated(ctx, evt)
	default:
		return fmt.Errorf("unhandled event type: %T: %w", event, consumer.ErrSkipMessage)
	}
}

func (h *productHandler) handleProductCreated(ctx context.Context, e *events.ProductCreatedEvent) error {
	return h.promoteImage(ctx, e.Payload.ProductID, e.Payload.ImageID)
}

func (h *productHandler) handleProductUpdated(ctx context.Context, e *events.ProductUpdatedEvent) error {
	return h.promoteImage(ctx, e.Payload.ProductID, e.Payload.ImageID)
}

func (h *productHandler) promoteImage(ctx context.Context, productID string, imageID *string) error {
	var imageIDs *[]string
	if imageID != nil {
		imageIDs = &[]string{*imageID}
	}

	cmd := command.PromoteImagesCommand{
		DraftID:   productID,
		ImageIDs:  imageIDs,
		ProductID: productID,
	}

	_, err := h.promoteImagesHandler.Handle(ctx, cmd)
	return err
}

func (h *productHandler) log(ctx context.Context) *zap.Logger {
	return logger.Get(ctx).With(zap.String("component", "product-handler"))
}

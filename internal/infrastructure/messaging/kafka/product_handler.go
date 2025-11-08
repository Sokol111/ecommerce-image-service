package kafka

import (
	"context"
	"fmt"

	"github.com/Sokol111/ecommerce-commons/pkg/core/logger"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/kafka/consumer"
	"github.com/Sokol111/ecommerce-image-service/internal/application/command"
	"github.com/Sokol111/ecommerce-product-service-api/events"
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
	// Type assert to Event interface first to get exhaustiveness checking
	e, ok := event.(events.Event)
	if !ok {
		return fmt.Errorf("event does not implement Event interface: %T: %w", event, consumer.ErrSkipMessage)
	}

	// Now switch on concrete types - exhaustive linter will warn if any Event type is missing
	switch evt := e.(type) {
	case *events.ProductCreatedEvent:
		return h.handleProductCreated(ctx, evt)
	case *events.ProductUpdatedEvent:
		h.log(ctx).Warn("ProductUpdatedEvent handling not implemented yet")
		return nil
	default:
		// If exhaustive linter is enabled and all Event types are handled above,
		// this case should theoretically never be reached
		return fmt.Errorf("unhandled event type: %T: %w", event, consumer.ErrSkipMessage)
	}
}

func (h *productHandler) handleProductCreated(ctx context.Context, e *events.ProductCreatedEvent) error {
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

func (h *productHandler) log(ctx context.Context) *zap.Logger {
	return logger.FromContext(ctx).With(zap.String("component", "product-handler"))
}

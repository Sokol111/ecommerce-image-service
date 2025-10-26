package messaging

import (
	"context"

	"github.com/Sokol111/ecommerce-commons/pkg/messaging"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/kafka/consumer"
	"github.com/Sokol111/ecommerce-image-service/internal/application/command"
)

type productCreatedHandler struct {
	promoteImagesHandler command.PromoteImagesCommandHandler
}

func newProductCreatedHandler(promoteImages command.PromoteImagesCommandHandler) consumer.Handler[messaging.ProductCreated] {
	return &productCreatedHandler{
		promoteImagesHandler: promoteImages,
	}
}

func (h *productCreatedHandler) Process(ctx context.Context, e *messaging.Event[messaging.ProductCreated]) error {
	var imageIDs *[]string
	if e.Payload.ImageId != nil {
		imageIDs = &[]string{*e.Payload.ImageId}
	}

	cmd := command.PromoteImagesCommand{
		DraftID:   e.Payload.ProductID,
		ImageIDs:  imageIDs,
		ProductID: e.Payload.ProductID,
	}

	_, err := h.promoteImagesHandler.Handle(ctx, cmd)
	return err
}

func (h *productCreatedHandler) Validate(payload *messaging.ProductCreated) error {
	return nil
}

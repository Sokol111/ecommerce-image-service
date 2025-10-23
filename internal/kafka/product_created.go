package kafka

import (
	"context"

	"github.com/Sokol111/ecommerce-commons/pkg/event"
	"github.com/Sokol111/ecommerce-commons/pkg/event/payload"
	"github.com/Sokol111/ecommerce-commons/pkg/kafka/consumer"
	"github.com/Sokol111/ecommerce-image-service/internal/model"
)

type productCreatedHandler struct {
	imageService model.ImageService
}

func newProductCreatedHandler(imageService model.ImageService) consumer.Handler[payload.ProductCreated] {
	return &productCreatedHandler{
		imageService: imageService,
	}
}

func (h *productCreatedHandler) Process(ctx context.Context, e *event.Event[payload.ProductCreated]) error {
	var images *[]string
	if e.Payload.ImageId != nil {
		images = &[]string{*e.Payload.ImageId}
	}
	_, err := h.imageService.PromoteDraftImages(ctx, model.PromoteDraftDTO{
		DraftId:   e.Payload.ProductID,
		Images:    images,
		Move:      true,
		ProductId: e.Payload.ProductID,
	})
	return err
}

func (h *productCreatedHandler) Validate(payload *payload.ProductCreated) error {
	return nil
}

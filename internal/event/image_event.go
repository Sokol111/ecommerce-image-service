package event

import (
	"context"

	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
	"github.com/Sokol111/ecommerce-image-service-api/gen/events"
)

func NewProductImagePromotedOutboxMessage(ctx context.Context, productID string, imageID string, imageURL string) outbox.Message {
	return outbox.Message{
		Event: &events.ProductImagePromotedEvent{
			Payload: events.ProductImagePromotedPayload{
				ProductID: productID,
				ImageID:   imageID,
				ImageURL:  imageURL,
			},
		},
		Key: productID,
	}
}

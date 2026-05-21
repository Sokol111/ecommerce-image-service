package kafka

import (
	"context"

	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
	"github.com/Sokol111/ecommerce-image-service-api/gen/events"
	"github.com/Sokol111/ecommerce-image-service/internal/application/image"
)

type imageEventFactory struct{}

func newImageEventFactory() image.ImageEventFactory {
	return &imageEventFactory{}
}

func (f *imageEventFactory) NewProductImagePromotedOutboxMessage(ctx context.Context, productID string, imageID string, smallImageURL string, largeImageURL string) outbox.Message {
	return outbox.Message{
		Event: &events.ProductImagePromotedEvent{
			Payload: events.ProductImagePromotedPayload{
				ProductID:     productID,
				ImageID:       imageID,
				SmallImageURL: smallImageURL,
				LargeImageURL: largeImageURL,
			},
		},
		Key: productID,
	}
}

func (f *imageEventFactory) BuildPromotionMessages(ctx context.Context, ownerType image.OwnerType, ownerID string, images []image.PromotedImage) ([]outbox.Message, error) {
	switch ownerType {
	case image.OwnerTypeProduct:
		msgs := make([]outbox.Message, 0, len(images))
		for _, img := range images {
			msgs = append(msgs, f.NewProductImagePromotedOutboxMessage(ctx, ownerID, img.ImageID, img.SmallImageURL, img.LargeImageURL))
		}
		return msgs, nil
	default:
		return nil, nil
	}
}

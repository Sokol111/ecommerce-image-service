package image //nolint:revive // intentional package name

import (
	"context"

	"github.com/Sokol111/ecommerce-commons/pkg/kafka/outbox"
)

// ImageEventFactory defines the port for creating outbox messages.
type ImageEventFactory interface {
	NewProductImagePromotedOutboxMessage(ctx context.Context, productID string, imageID string, smallImageURL string, largeImageURL string) outbox.Message
	BuildPromotionMessages(ctx context.Context, ownerType OwnerType, ownerID string, images []PromotedImage) ([]outbox.Message, error)
}

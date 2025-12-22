package event

import (
	"context"
	"time"

	commonsevents "github.com/Sokol111/ecommerce-commons/pkg/messaging/kafka/events"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
	"github.com/Sokol111/ecommerce-commons/pkg/observability/tracing"
	"github.com/Sokol111/ecommerce-image-service-api/gen/events"
	"github.com/google/uuid"
)

func newProductImagePromotedEvent(ctx context.Context, productID string, imageID string, imageURL string) *events.ProductImagePromotedEvent {
	traceId := tracing.GetTraceID(ctx)
	return &events.ProductImagePromotedEvent{
		Metadata: commonsevents.EventMetadata{
			EventID:   uuid.New().String(),
			EventType: events.EventTypeProductImagePromoted,
			Source:    "ecommerce-image-service",
			Timestamp: time.Now().UTC(),
			TraceID:   &traceId,
		},
		Payload: events.ProductImagePromotedPayload{
			ProductID: productID,
			ImageID:   imageID,
			ImageURL:  imageURL,
		},
	}
}

func NewProductImagePromotedOutboxMessage(ctx context.Context, productID string, imageID string, imageURL string) (outbox.Message, error) {
	e := newProductImagePromotedEvent(ctx, productID, imageID, imageURL)

	return outbox.Message{
		Payload: e,
		EventID: e.Metadata.EventID,
		Key:     productID,
	}, nil
}

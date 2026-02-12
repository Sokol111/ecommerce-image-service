package kafka

import (
	catalog_events "github.com/Sokol111/ecommerce-catalog-service-api/gen/events"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/kafka/consumer"
	image_events "github.com/Sokol111/ecommerce-image-service-api/gen/events"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Options(
		catalog_events.Module(),
		image_events.Module(),
		consumer.RegisterHandlerAndConsumer("catalog-events", newProductHandler),
	)
}

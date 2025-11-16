package kafka

import (
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/kafka/consumer"
	"github.com/Sokol111/ecommerce-product-service-api/events"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Options(
		consumer.RegisterTypeMapping(events.DefaultTypeMapping),
		consumer.RegisterHandlerAndConsumer("product-events", newProductHandler),
	)
}

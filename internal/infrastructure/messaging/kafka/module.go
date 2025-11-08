package kafka

import (
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/kafka/consumer"
	"github.com/Sokol111/ecommerce-product-service-api/events"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Provide(
		consumer.RegisterHandlerAndConsumer("product-events", newProductHandler, events.Unmarshal),
	)
}

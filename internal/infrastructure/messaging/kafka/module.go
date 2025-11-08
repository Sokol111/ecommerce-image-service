package kafka

import (
	"reflect"

	"github.com/Sokol111/ecommerce-commons/pkg/messaging/kafka/consumer"
	"github.com/Sokol111/ecommerce-product-service-api/events"
	"go.uber.org/fx"
)

func Module() fx.Option {
	typeMapping := consumer.TypeMapping{
		events.EventTypeProductCreated: reflect.TypeOf(&events.ProductCreatedEvent{}),
		events.EventTypeProductUpdated: reflect.TypeOf(&events.ProductUpdatedEvent{}),
	}

	return fx.Provide(
		consumer.RegisterHandlerAndConsumer("product-events", newProductHandler),
		fx.Annotate(
			func() consumer.TypeMapping { return typeMapping },
			fx.ResultTags(`name:"product-events"`),
		),
	)
}

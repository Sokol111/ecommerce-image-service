package kafka

import (
	catalog_events "github.com/Sokol111/ecommerce-catalog-service-api/gen/events"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/kafka/avro/mapping"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/kafka/consumer"
	image_events "github.com/Sokol111/ecommerce-image-service-api/gen/events"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Options(
		consumer.RegisterHandlerAndConsumer("catalog-events", newProductHandler),
		fx.Invoke(registerProductSchemas),
		fx.Invoke(registerImageSchemas),
	)
}

func registerProductSchemas(tm *mapping.TypeMapping) error {
	return tm.RegisterBindings(catalog_events.SchemaBindings)
}

func registerImageSchemas(tm *mapping.TypeMapping) error {
	return tm.RegisterBindings(image_events.SchemaBindings)
}

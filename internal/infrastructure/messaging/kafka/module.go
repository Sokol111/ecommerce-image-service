package kafka

import (
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/kafka/avro/mapping"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/kafka/consumer"
	"github.com/Sokol111/ecommerce-product-service-api/gen/events"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Options(
		consumer.RegisterHandlerAndConsumer("product-events", newProductHandler),
		fx.Invoke(registerSchemas),
	)
}

func registerSchemas(tm *mapping.TypeMapping) error {
	for _, reg := range events.TypeRegistrations {
		if err := tm.Register(reg.GoType, reg.SchemaJSON, reg.SchemaName, reg.Topic); err != nil {
			return err
		}
	}
	return nil
}

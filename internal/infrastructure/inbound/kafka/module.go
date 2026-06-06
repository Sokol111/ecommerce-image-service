package kafka

import (
	catalog_events "github.com/Sokol111/ecommerce-catalog-service-api/gen/events"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/kafka/consumer"
	image_events "github.com/Sokol111/ecommerce-image-service-api/gen/events"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func Module() fx.Option {
	return fx.Options(
		catalog_events.Module(),
		image_events.Module(),
		fx.Provide(newProductHandler),
		consumer.RegisterHandlerAndConsumer("catalog-events", newProductRouter),
	)
}

func newProductRouter(h *productHandler, log *zap.Logger) consumer.Handler {
	r := consumer.NewRouter(log)
	consumer.Register(r, h.HandleProductUpdated)
	consumer.Register(r, h.HandleProductDeleted)
	return r
}

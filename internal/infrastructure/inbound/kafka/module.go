package kafka

import (
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/kafka/consumer"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func Module() fx.Option {
	return fx.Options(
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

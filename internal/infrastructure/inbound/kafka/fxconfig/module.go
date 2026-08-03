package fxconfig

import (
	"github.com/Sokol111/ecommerce-commons/pkg/kafka/consumer"
	fx_consumer "github.com/Sokol111/ecommerce-commons/pkg/kafka/consumer/fxconfig"
	"github.com/Sokol111/ecommerce-image-service/internal/infrastructure/inbound/kafka"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewKafkaModule() fx.Option {
	return fx.Options(
		fx.Provide(kafka.NewProductHandler),
		fx_consumer.RegisterHandlerAndConsumer("catalog-events", newProductRouter),
	)
}

func newProductRouter(h *kafka.ProductHandler, log *zap.Logger) consumer.Handler {
	r := consumer.NewRouter(log)
	consumer.Register(r, h.HandleProductUpdated)
	consumer.Register(r, h.HandleProductDeleted)
	return r
}

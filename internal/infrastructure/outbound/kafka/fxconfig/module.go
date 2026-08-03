package fxconfig

import (
	"github.com/Sokol111/ecommerce-image-service/internal/infrastructure/outbound/kafka"
	"go.uber.org/fx"
)

func NewKafkaModule() fx.Option {
	return fx.Options(
		fx.Provide(kafka.NewImageEventFactory),
	)
}

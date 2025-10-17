package main

import (
	"context"

	"github.com/Sokol111/ecommerce-commons/pkg/module"
	"github.com/Sokol111/ecommerce-image-service/internal/http"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var AppModules = fx.Options(
	module.NewInfraModule(),
	// module.NewKafkaModule(),
	// product.NewBusinessModule(),
	http.NewHttpHandlerModule(),
)

func main() {
	app := fx.New(
		AppModules,
		fx.Invoke(func(lc fx.Lifecycle, log *zap.Logger) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					log.Info("Application starting...")
					return nil
				},
				OnStop: func(ctx context.Context) error {
					log.Info("Application stopping...")
					return nil
				},
			})
		}),
	)
	app.Run()
}

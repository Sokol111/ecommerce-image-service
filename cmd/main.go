package main

import (
	"context"

	"github.com/Sokol111/ecommerce-commons/pkg/modules"
	"github.com/Sokol111/ecommerce-commons/pkg/swaggerui"
	"github.com/Sokol111/ecommerce-image-service-api/api"
	"github.com/Sokol111/ecommerce-image-service/internal/application"
	"github.com/Sokol111/ecommerce-image-service/internal/http"
	"github.com/Sokol111/ecommerce-image-service/internal/infrastructure/external/imgproxy"
	"github.com/Sokol111/ecommerce-image-service/internal/infrastructure/external/s3"
	"github.com/Sokol111/ecommerce-image-service/internal/infrastructure/messaging/kafka"
	"github.com/Sokol111/ecommerce-image-service/internal/infrastructure/persistence/mongo"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var AppModules = fx.Options(
	// Infrastructure - Core
	modules.NewCoreModule(),
	modules.NewPersistenceModule(),
	modules.NewHTTPModule(),
	modules.NewObservabilityModule(),
	modules.NewMessagingModule(),

	// Infrastructure - External Services
	s3.NewS3Module(),
	imgproxy.NewImgProxyModule(),

	// Infrastructure - Persistence
	mongo.Module(),

	// Infrastructure - Messaging
	kafka.Module(),

	// Application Layer
	application.Module(),

	// HTTP
	http.NewHttpHandlerModule(),
	swaggerui.NewSwaggerModule(swaggerui.SwaggerConfig{OpenAPIContent: api.OpenAPIDoc}),
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

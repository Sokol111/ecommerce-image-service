package main

import (
	"context"

	commons_core "github.com/Sokol111/ecommerce-commons/pkg/core"
	commons_http "github.com/Sokol111/ecommerce-commons/pkg/http"
	commons_messaging "github.com/Sokol111/ecommerce-commons/pkg/messaging"
	commons_observability "github.com/Sokol111/ecommerce-commons/pkg/observability"
	commons_persistence "github.com/Sokol111/ecommerce-commons/pkg/persistence"
	commons_pyroscope "github.com/Sokol111/ecommerce-commons/pkg/pyroscope"
	commons_token "github.com/Sokol111/ecommerce-commons/pkg/security/token"
	commons_validation "github.com/Sokol111/ecommerce-commons/pkg/security/validation"
	commons_swaggerui "github.com/Sokol111/ecommerce-commons/pkg/swaggerui"
	"github.com/Sokol111/ecommerce-commons/pkg/tenant"
	"github.com/Sokol111/ecommerce-image-service-api/gen/httpapi"
	"github.com/Sokol111/ecommerce-image-service/internal/application"
	"github.com/Sokol111/ecommerce-image-service/internal/infrastructure"
	"github.com/Sokol111/ecommerce-image-service/internal/infrastructure/inbound/http"
	inbound_kafka "github.com/Sokol111/ecommerce-image-service/internal/infrastructure/inbound/kafka"
	"github.com/Sokol111/ecommerce-image-service/internal/infrastructure/outbound/imgproxy"
	outbound_kafka "github.com/Sokol111/ecommerce-image-service/internal/infrastructure/outbound/kafka"
	"github.com/Sokol111/ecommerce-image-service/internal/infrastructure/outbound/mongo"
	"github.com/Sokol111/ecommerce-image-service/internal/infrastructure/outbound/s3"
	"github.com/Sokol111/ecommerce-image-service/internal/infrastructure/outbound/security"
	tenantapi "github.com/Sokol111/ecommerce-tenant-service-api/gen/httpapi"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var AppModules = fx.Options(
	// Commons
	commons_core.NewCoreModule(),
	commons_persistence.NewPersistenceModule(commons_persistence.WithTenantMigrations()),
	commons_http.NewHTTPModule(),
	commons_observability.NewObservabilityModule(),
	commons_messaging.NewMessagingModule(),
	commons_validation.NewModule(),
	commons_token.NewModule(),
	commons_pyroscope.NewPyroscopeModule(),
	commons_swaggerui.NewSwaggerModule(),

	// Tenant
	tenant.MiddlewareModule(),
	tenantapi.NewTenantSlugsModule("clients.tenant-service"),
	tenantapi.TenantEventsModule("tenant-events"),

	// Infrastructure - Config
	infrastructure.ConfigModule(),

	// Infrastructure - External Services
	s3.NewS3Module(),
	imgproxy.NewImgProxyModule(),
	security.NewSecurityModule(),

	// Infrastructure - Persistence
	mongo.Module(),

	// Infrastructure - Messaging
	inbound_kafka.Module(),
	outbound_kafka.Module(),

	// Application Layer
	application.Module(),

	// HTTP
	httpapi.ServerModule(),
	http.NewHTTPHandlerModule(),
)

func main() {
	app := fx.New(
		AppModules,
		fx.Invoke(func(lc fx.Lifecycle, log *zap.Logger) {
			lc.Append(fx.Hook{
				OnStop: func(ctx context.Context) error {
					log.Info("Application stopping...")
					return nil
				},
			})
		}),
	)
	app.Run()
}

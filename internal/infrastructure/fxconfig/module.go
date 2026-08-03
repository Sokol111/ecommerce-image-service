package fxconfig

import (
	fx_connect "github.com/Sokol111/ecommerce-image-service/internal/infrastructure/inbound/connect/fxconfig"
	fx_inbound_kafka "github.com/Sokol111/ecommerce-image-service/internal/infrastructure/inbound/kafka/fxconfig"
	fx_imgproxy "github.com/Sokol111/ecommerce-image-service/internal/infrastructure/outbound/imgproxy/fxconfig"
	fx_outbound_kafka "github.com/Sokol111/ecommerce-image-service/internal/infrastructure/outbound/kafka/fxconfig"
	fx_mongo "github.com/Sokol111/ecommerce-image-service/internal/infrastructure/outbound/mongo/fxconfig"
	fx_s3 "github.com/Sokol111/ecommerce-image-service/internal/infrastructure/outbound/s3/fxconfig"
	fx_token "github.com/Sokol111/ecommerce-image-service/internal/infrastructure/outbound/token/fxconfig"
	fx_tenant_api "github.com/Sokol111/ecommerce-tenant-service-api/pkg/fxconfig"
	"go.uber.org/fx"
)

func NewInfrastructureModule() fx.Option {
	return fx.Options(
		NewTenantModule(),
		NewAdaptersModule(),
	)
}

func NewAdaptersModule() fx.Option {
	return fx.Options(
		fx_mongo.NewMongoModule(),
		fx_imgproxy.NewImgProxyModule(),
		fx_s3.NewS3Module(),
		fx_inbound_kafka.NewKafkaModule(),
		fx_outbound_kafka.NewKafkaModule(),
		fx_token.NewTokenModule(),
		fx_connect.NewConnectModule(),
	)
}

func NewTenantModule() fx.Option {
	return fx.Options(
		fx_tenant_api.NewKafkaConsumerModule(),
		fx_tenant_api.NewTenantSlugsProviderModule(),
		fx_tenant_api.NewGrpcClientsModule(),
	)
}

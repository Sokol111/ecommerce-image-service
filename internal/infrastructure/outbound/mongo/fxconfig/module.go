package mongo

import (
	commonsmongo "github.com/Sokol111/ecommerce-commons/pkg/mongo"
	"github.com/Sokol111/ecommerce-commons/pkg/tenant"
	"github.com/Sokol111/ecommerce-image-service/internal/application/image"
	"github.com/Sokol111/ecommerce-image-service/internal/infrastructure/outbound/mongo"
	mongodriver "go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/fx"
)

func NewMongoModule() fx.Option {
	return fx.Provide(
		mongo.NewImageMapper,
		mongo.NewImageRepository,
		func(database *mongodriver.Database, mapper *mongo.ImageMapper) (*commonsmongo.GenericRepository[image.Image, mongo.ImageEntity], error) {
			return commonsmongo.NewGenericRepository(
				tenant.NewMultiTenantCollectionProvider(database, "image"),
				mapper,
			)
		},
	)
}

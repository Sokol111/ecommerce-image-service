package image

import (
	"context"

	"github.com/Sokol111/ecommerce-commons/pkg/mongo"
	"go.uber.org/fx"
)

func NewBusinessModule() fx.Option {
	return fx.Provide(
		provideCollection,
		newStore,
		newService,
	)
}

func provideCollection(lc fx.Lifecycle, m mongo.Mongo) (*mongo.CollectionWrapper[collection], error) {
	wrapper := &mongo.CollectionWrapper[collection]{}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			wrapper.Coll = m.GetCollection("image")
			return nil
		},
	})
	return wrapper, nil
}

package mongo

import (
	"context"

	commonsmongo "github.com/Sokol111/ecommerce-commons/pkg/persistence/mongo"
	"github.com/Sokol111/ecommerce-image-service/internal/domain/image"
	"go.mongodb.org/mongo-driver/bson"
)

type imageRepository struct {
	*commonsmongo.GenericRepository[image.Image, imageEntity]
	coll commonsmongo.Collection
}

func newImageRepository(mongo commonsmongo.Mongo, mapper *imageMapper) image.Repository {
	coll := mongo.GetCollectionWrapper("image")
	genericRepo := commonsmongo.NewGenericRepository(
		coll,
		mapper,
	)

	return &imageRepository{
		GenericRepository: genericRepo,
		coll:              coll,
	}
}

// FindByOwner finds images by owner type and ID
func (r *imageRepository) FindByOwner(ctx context.Context, ownerType, ownerID string, imageIDs []string) ([]*image.Image, error) {
	filter := bson.M{
		"ownerType": ownerType,
		"ownerId":   ownerID,
		"status": bson.M{
			"$ne": string(image.StatusDeleted),
		},
	}
	if len(imageIDs) > 0 {
		filter["_id"] = bson.M{"$in": imageIDs}
	}

	cur, err := r.coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var entities []imageEntity
	if err = cur.All(ctx, &entities); err != nil {
		return nil, err
	}

	images := make([]*image.Image, 0, len(entities))
	mapper := &imageMapper{}
	for i := range entities {
		images = append(images, mapper.ToDomain(&entities[i]))
	}

	return images, nil
}

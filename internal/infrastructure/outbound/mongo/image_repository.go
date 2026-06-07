package mongo

import (
	"context"

	commonsmongo "github.com/Sokol111/ecommerce-commons/pkg/persistence/mongo"
	"github.com/Sokol111/ecommerce-image-service/internal/application/image"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type imageRepository struct {
	*commonsmongo.GenericRepository[image.Image, imageEntity]
}

func newImageRepository(admin commonsmongo.Admin, mapper *imageMapper, resolver commonsmongo.DatabaseResolver) (image.Repository, error) {
	genericRepo, err := commonsmongo.NewTenantRepository(
		admin, "image",
		mapper,
		resolver,
	)
	if err != nil {
		return nil, err
	}

	return &imageRepository{
		GenericRepository: genericRepo,
	}, nil
}

// FindByIDs finds images by their IDs
func (r *imageRepository) FindByIDs(ctx context.Context, ids []string) ([]*image.Image, error) {
	if len(ids) == 0 {
		return []*image.Image{}, nil
	}

	filter := bson.D{
		{Key: "_id", Value: bson.M{"$in": ids}},
		{Key: "status", Value: bson.M{"$ne": string(image.StatusDeleted)}},
	}

	return r.FindAllWithFilter(ctx, filter, nil)
}

// FindByOwner finds images by owner type and ID
func (r *imageRepository) FindByOwner(ctx context.Context, ownerType, ownerID string, imageIDs []string) ([]*image.Image, error) {
	filter := bson.D{
		{Key: "ownerType", Value: ownerType},
		{Key: "ownerId", Value: ownerID},
		{Key: "status", Value: bson.M{"$ne": string(image.StatusDeleted)}},
	}

	if len(imageIDs) > 0 {
		filter = append(filter, bson.E{Key: "_id", Value: bson.M{"$in": imageIDs}})
	}

	return r.FindAllWithFilter(ctx, filter, nil)
}

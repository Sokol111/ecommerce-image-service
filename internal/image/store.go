package image

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Sokol111/ecommerce-commons/pkg/mongo"
	"github.com/Sokol111/ecommerce-image-service/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	mongodriver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var errEntityNotFound = errors.New("entity not found in database")

type Store interface {
	GetById(ctx context.Context, id string) (*model.Image, error)

	Create(ctx context.Context, image *model.Image) (*model.Image, error)

	ListByOwner(ctx context.Context, ownerType string, ownerId string, ids []string) ([]*model.Image, error)

	UpdateAfterPromote(ctx context.Context, imageID string, newOwnerType string, newOwnerID string, newKey string) (*model.Image, error)

	// Update(ctx context.Context, product *model.Image) (*model.Image, error)

	// GetAll(ctx context.Context) ([]*model.Image, error)
}

type store struct {
	wrapper *mongo.CollectionWrapper[collection]
}

func newStore(wrapper *mongo.CollectionWrapper[collection]) Store {
	return &store{wrapper}
}

func (r *store) GetById(ctx context.Context, id string) (*model.Image, error) {
	result := r.wrapper.Coll.FindOne(ctx, bson.D{{Key: "_id", Value: id}})
	var e entity
	err := result.Decode(&e)
	if err != nil {
		if errors.Is(err, mongodriver.ErrNoDocuments) {
			return nil, fmt.Errorf("failed to get image [%v]: %w", id, errEntityNotFound)
		}
		return nil, fmt.Errorf("failed to get image [%v]: decode error: %w", id, err)
	}
	return toDomain(&e), nil
}

func (r *store) Create(ctx context.Context, image *model.Image) (*model.Image, error) {
	e := fromDomain(image)

	_, err := r.wrapper.Coll.InsertOne(ctx, e)
	if err != nil {
		return nil, fmt.Errorf("failed to create image: %w", err)
	}

	return toDomain(e), nil
}

func (s *store) ListByOwner(ctx context.Context, ownerType string, ownerId string, ids []string) ([]*model.Image, error) {
	filter := bson.M{
		"ownerType": ownerType,
		"ownerId":   ownerId,
		"status": bson.M{
			"$ne": "deleted",
		},
	}
	if len(ids) > 0 {
		filter["_id"] = bson.M{"$in": ids}
	}

	cur, err := s.wrapper.Coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var entities []entity
	if err = cur.All(ctx, &entities); err != nil {
		return nil, fmt.Errorf("failed to decode images: %w", err)
	}

	images := make([]*model.Image, 0, len(entities))

	for i := range entities {
		images = append(images, toDomain(&entities[i]))
	}

	return images, nil
}

func (s *store) UpdateAfterPromote(ctx context.Context, imageID string, newOwnerType string, newOwnerID string, newKey string) (*model.Image, error) {
	update := bson.M{
		"$set": bson.M{
			"ownerType": newOwnerType,
			"ownerId":   newOwnerID,
			"key":       newKey,
			"updatedAt": time.Now().UTC(),
			"status":    "processing",
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After).SetUpsert(false)

	var out entity
	err := s.wrapper.Coll.FindOneAndUpdate(ctx, bson.M{"_id": imageID}, update, opts).Decode(&out)
	if err != nil {
		if errors.Is(err, mongodriver.ErrNoDocuments) {
			return nil, errEntityNotFound
		}
		var we mongodriver.WriteException
		if errors.As(err, &we) {
			for _, e := range we.WriteErrors {
				if e.Code == 11000 {
					err := s.wrapper.Coll.FindOne(ctx, bson.M{"_id": imageID}).Decode(&out)
					if err != nil {
						return nil, fmt.Errorf("failed to fetch image after duplicate key error: %w", err)
					}
				}
			}
		}
		return nil, fmt.Errorf("failed to update image after promote: %w", err)
	}
	return toDomain(&out), nil
}

func toDomain(e *entity) *model.Image {
	return &model.Image{
		Id:         e.ID,
		Version:    e.Version,
		Alt:        e.Alt,
		OwnerType:  e.OwnerType,
		OwnerId:    e.OwnerId,
		Role:       e.Role,
		Key:        e.Key,
		Mime:       e.Mime,
		Size:       e.Size,
		Status:     e.Status,
		CreatedAt:  e.CreatedAt.UTC(),
		ModifiedAt: e.ModifiedAt.UTC(),
	}
}

func fromDomain(d *model.Image) *entity {
	return &entity{
		ID:         d.Id,
		Version:    d.Version,
		Alt:        d.Alt,
		OwnerType:  d.OwnerType,
		OwnerId:    d.OwnerId,
		Role:       d.Role,
		Key:        d.Key,
		Mime:       d.Mime,
		Size:       d.Size,
		Status:     d.Status,
		CreatedAt:  d.CreatedAt,
		ModifiedAt: d.ModifiedAt,
	}
}

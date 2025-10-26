package image

import "context"

type Repository interface {
	Save(ctx context.Context, image *Image) error

	FindByID(ctx context.Context, id string) (*Image, error)

	FindByOwner(ctx context.Context, ownerType, ownerID string, imageIDs []string) ([]*Image, error)

	Update(ctx context.Context, image *Image) (*Image, error)

	Delete(ctx context.Context, id string) error
}

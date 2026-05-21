package image //nolint:revive // package name intentional

import (
	"context"
	"errors"
	"fmt"

	"github.com/Sokol111/ecommerce-commons/pkg/persistence/mongo"
)

// GetImageByIDQuery represents a query to get an image by ID
type GetImageByIDQuery struct {
	ID string
}

// GetImageByIDQueryHandler handles GetImageByIDQuery
type GetImageByIDQueryHandler interface {
	Handle(ctx context.Context, query GetImageByIDQuery) (*Image, error)
}

type getImageByIDHandler struct {
	repo Repository
}

func NewGetImageByIDHandler(repo Repository) GetImageByIDQueryHandler {
	return &getImageByIDHandler{repo: repo}
}

func (h *getImageByIDHandler) Handle(ctx context.Context, query GetImageByIDQuery) (*Image, error) {
	img, err := h.repo.FindByID(ctx, query.ID)
	if err != nil {
		if errors.Is(err, mongo.ErrEntityNotFound) {
			return nil, ErrImageNotFound
		}
		return nil, fmt.Errorf("failed to get image: %w", err)
	}
	return img, nil
}

package query

import (
	"context"
	"errors"
	"fmt"

	"github.com/Sokol111/ecommerce-commons/pkg/persistence"
	"github.com/Sokol111/ecommerce-image-service/internal/domain/image"
)

// GetImageByIDQuery represents a query to get an image by ID
type GetImageByIDQuery struct {
	ID string
}

// GetImageByIDQueryHandler handles GetImageByIDQuery
type GetImageByIDQueryHandler interface {
	Handle(ctx context.Context, query GetImageByIDQuery) (*image.Image, error)
}

type getImageByIDHandler struct {
	repo image.Repository
}

func NewGetImageByIDHandler(repo image.Repository) GetImageByIDQueryHandler {
	return &getImageByIDHandler{repo: repo}
}

func (h *getImageByIDHandler) Handle(ctx context.Context, query GetImageByIDQuery) (*image.Image, error) {
	img, err := h.repo.FindByID(ctx, query.ID)
	if err != nil {
		if errors.Is(err, persistence.ErrEntityNotFound) {
			return nil, image.ErrImageNotFound
		}
		return nil, fmt.Errorf("failed to get image: %w", err)
	}
	return img, nil
}

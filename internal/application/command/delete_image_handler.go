package command

import (
	"context"
	"fmt"

	"github.com/Sokol111/ecommerce-commons/pkg/core/logger"
	"github.com/Sokol111/ecommerce-commons/pkg/persistence"
	"github.com/Sokol111/ecommerce-image-service/internal/application/abstraction"
	"github.com/Sokol111/ecommerce-image-service/internal/domain/image"
	"go.uber.org/zap"
)

// DeleteImageCommand represents a request to delete an image
type DeleteImageCommand struct {
	ImageID string
	Hard    bool // true for hard delete, false for soft delete
}

// DeleteImageCommandHandler handles DeleteImageCommand
type DeleteImageCommandHandler interface {
	Handle(ctx context.Context, cmd DeleteImageCommand) error
}

type deleteImageHandler struct {
	repo       image.Repository
	objStorage abstraction.ObjectStorage
}

func NewDeleteImageHandler(repo image.Repository, storage abstraction.ObjectStorage) DeleteImageCommandHandler {
	return &deleteImageHandler{
		repo:       repo,
		objStorage: storage,
	}
}

func (h *deleteImageHandler) Handle(ctx context.Context, cmd DeleteImageCommand) error {
	// Get image from repository
	img, err := h.repo.FindByID(ctx, cmd.ImageID)
	if err != nil {
		if err == persistence.ErrEntityNotFound {
			return image.ErrImageNotFound
		}
		return fmt.Errorf("failed to get image: %w", err)
	}

	// Delete from S3
	err = h.objStorage.DeleteObject(ctx, &abstraction.DeleteObjectInput{
		Key: img.Key,
	})

	if err != nil {
		h.log(ctx).Warn("failed to delete s3 object (continuing anyway)", zap.Error(err), zap.String("key", img.Key))
	}

	// Delete from database
	if cmd.Hard {
		err := h.repo.Delete(ctx, cmd.ImageID)
		if err != nil {
			return fmt.Errorf("failed to delete image from db: %w", err)
		}
		h.log(ctx).Debug("image hard deleted", zap.String("id", cmd.ImageID))
	} else {
		img.MarkAsDeleted()
		_, err := h.repo.Update(ctx, img)
		if err != nil {
			return fmt.Errorf("failed to mark image as deleted in db: %w", err)
		}
		h.log(ctx).Debug("image soft deleted", zap.String("id", cmd.ImageID))
	}

	return nil
}

func (h *deleteImageHandler) log(ctx context.Context) *zap.Logger {
	return logger.FromContext(ctx).With(zap.String("component", "delete-image-handler"))
}

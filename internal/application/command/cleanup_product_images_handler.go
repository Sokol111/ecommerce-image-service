package command

import (
	"context"
	"fmt"

	"github.com/Sokol111/ecommerce-commons/pkg/core/logger"
	"github.com/Sokol111/ecommerce-image-service/internal/application/abstraction"
	"github.com/Sokol111/ecommerce-image-service/internal/domain/image"
	"go.uber.org/zap"
)

// CleanupOwnerImagesCommand represents a request to remove all images for an owner.
type CleanupOwnerImagesCommand struct {
	OwnerType image.OwnerType
	OwnerID   string
}

// CleanupOwnerImagesCommandHandler handles CleanupOwnerImagesCommand.
type CleanupOwnerImagesCommandHandler interface {
	Handle(ctx context.Context, cmd CleanupOwnerImagesCommand) error
}

type cleanupOwnerImagesHandler struct {
	repo       image.Repository
	objStorage abstraction.ObjectStorage
}

func NewCleanupOwnerImagesHandler(repo image.Repository, objStorage abstraction.ObjectStorage) CleanupOwnerImagesCommandHandler {
	return &cleanupOwnerImagesHandler{
		repo:       repo,
		objStorage: objStorage,
	}
}

func (h *cleanupOwnerImagesHandler) Handle(ctx context.Context, cmd CleanupOwnerImagesCommand) error {
	images, err := h.repo.FindByOwner(ctx, string(cmd.OwnerType), cmd.OwnerID, nil)
	if err != nil {
		return fmt.Errorf("find owner images: %w", err)
	}

	if len(images) == 0 {
		return nil
	}

	var keys []string
	for _, img := range images {
		img.MarkAsDeleted()
		if _, err := h.repo.Update(ctx, img); err != nil {
			return fmt.Errorf("soft-delete image %s: %w", img.ID, err)
		}
		keys = append(keys, img.Key)
	}

	if err := h.objStorage.DeleteObjects(ctx, keys); err != nil {
		h.log(ctx).Warn("failed to delete image files from S3",
			zap.Strings("keys", keys),
			zap.Error(err),
		)
	}

	h.log(ctx).Debug("cleaned up owner images",
		zap.Int("count", len(images)),
		zap.String("ownerType", string(cmd.OwnerType)),
		zap.String("ownerID", cmd.OwnerID),
	)

	return nil
}

func (h *cleanupOwnerImagesHandler) log(ctx context.Context) *zap.Logger {
	return logger.Get(ctx).With(zap.String("component", "cleanup-owner-images-handler"))
}

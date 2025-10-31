package command

import (
	"context"
	"fmt"
	"strings"

	"github.com/Sokol111/ecommerce-commons/pkg/core/logger"
	"github.com/Sokol111/ecommerce-image-service/internal/application/abstraction"
	"github.com/Sokol111/ecommerce-image-service/internal/domain/image"
	"go.uber.org/zap"
)

// PromoteImagesCommand represents a request to promote draft images to product
type PromoteImagesCommand struct {
	DraftID   string
	ImageIDs  *[]string
	ProductID string
}

// PromoteImagesCommandHandler handles PromoteImagesCommand
type PromoteImagesCommandHandler interface {
	Handle(ctx context.Context, cmd PromoteImagesCommand) ([]*image.Image, error)
}

type promoteImagesHandler struct {
	repo       image.Repository
	objStorage abstraction.ObjectStorage
}

func NewPromoteImagesHandler(repo image.Repository, storage abstraction.ObjectStorage) PromoteImagesCommandHandler {
	return &promoteImagesHandler{
		repo:       repo,
		objStorage: storage,
	}
}

func (h *promoteImagesHandler) Handle(ctx context.Context, cmd PromoteImagesCommand) ([]*image.Image, error) {
	var imageIDs []string
	if cmd.ImageIDs != nil && len(*cmd.ImageIDs) > 0 {
		imageIDs = *cmd.ImageIDs
	}

	// Get images from repository
	images, err := h.repo.FindByOwner(ctx, "productDraft", cmd.DraftID, imageIDs)
	if err != nil {
		return nil, fmt.Errorf("list draft images: %w", err)
	}

	if len(images) == 0 {
		return []*image.Image{}, fmt.Errorf("no images found for draft %s", cmd.DraftID)
	}

	var promoted []*image.Image
	srcPrefix := "product-drafts/" + cmd.DraftID + "/"

	for _, img := range images {
		if !strings.HasPrefix(img.Key, srcPrefix) {
			return nil, fmt.Errorf("image %s has key outside draft prefix: %s", img.ID, img.Key)
		}

		// Determine new key
		dstKey := "products/" + cmd.ProductID + "/" + strings.TrimPrefix(img.Key, srcPrefix)

		// Check if target already exists
		exists, err := h.objectExists(ctx, dstKey)
		if err != nil {
			return nil, fmt.Errorf("check target exists: %w", err)
		}

		// Copy object if doesn't exist
		if !exists {
			err = h.objStorage.CopyObject(ctx, &abstraction.CopyObjectInput{
				SourceKey: img.Key,
				TargetKey: dstKey,
			})
			if err != nil {
				return nil, fmt.Errorf("copy %s -> %s: %w", img.Key, dstKey, err)
			}
		}

		// Delete old object
		_ = h.objStorage.DeleteObject(ctx, &abstraction.DeleteObjectInput{
			Key: img.Key,
		})

		// Update domain object
		if err := img.PromoteToProduct(cmd.ProductID, dstKey); err != nil {
			return nil, fmt.Errorf("promote image: %w", err)
		}

		// Save updated image
		updated, err := h.repo.Update(ctx, img)
		if err != nil {
			return nil, fmt.Errorf("update image after promote: %w", err)
		}

		promoted = append(promoted, updated)
	}

	h.log(ctx).Debug("images promoted", zap.Int("count", len(promoted)), zap.String("productID", cmd.ProductID))

	return promoted, nil
}

func (h *promoteImagesHandler) objectExists(ctx context.Context, key string) (bool, error) {
	_, err := h.objStorage.HeadObject(ctx, &abstraction.HeadObjectInput{
		Key: key,
	})
	if err != nil {
		// Assuming any error means object doesn't exist
		// Infrastructure layer should handle S3-specific errors
		return false, nil
	}
	return true, nil
}

func (h *promoteImagesHandler) log(ctx context.Context) *zap.Logger {
	return logger.FromContext(ctx).With(zap.String("component", "promote-images-handler"))
}

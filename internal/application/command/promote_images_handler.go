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

	// Validate draft exists and belongs to productDraft owner type
	draftImages, err := h.repo.FindByOwner(ctx, "productDraft", cmd.DraftID, nil)
	if err != nil {
		return nil, fmt.Errorf("validate draft exists: %w", err)
	}
	if len(draftImages) == 0 {
		return nil, fmt.Errorf("draft %s not found or has no images", cmd.DraftID)
	}

	// Get images to promote (either specified or all)
	images, err := h.repo.FindByOwner(ctx, "productDraft", cmd.DraftID, imageIDs)
	if err != nil {
		return nil, fmt.Errorf("list draft images: %w", err)
	}

	if len(images) == 0 {
		return nil, fmt.Errorf("no images found for promotion")
	}

	// Validate all specified imageIDs were found
	if imageIDs != nil && len(imageIDs) != len(images) {
		return nil, fmt.Errorf("some specified image IDs not found in draft %s", cmd.DraftID)
	}

	var promoted []*image.Image
	srcPrefix := "product-drafts/" + cmd.DraftID + "/"

	for _, img := range images {
		if !strings.HasPrefix(img.Key, srcPrefix) {
			return nil, fmt.Errorf("image %s has key outside draft prefix: %s", img.ID, img.Key)
		}

		// Determine new key
		dstKey := "products/" + cmd.ProductID + "/" + strings.TrimPrefix(img.Key, srcPrefix)

		// Check if target already exists to prevent overwriting
		exists, err := h.objectExists(ctx, dstKey)
		if err != nil {
			return nil, fmt.Errorf("check target exists: %w", err)
		}

		if exists {
			h.log(ctx).Warn("target object already exists, skipping copy",
				zap.String("dstKey", dstKey),
				zap.String("imageID", img.ID),
			)
			// Continue to update DB even if object exists (idempotency)
		} else {
			// Copy object
			err = h.objStorage.CopyObject(ctx, &abstraction.CopyObjectInput{
				SourceKey: img.Key,
				TargetKey: dstKey,
			})
			if err != nil {
				return nil, fmt.Errorf("copy %s -> %s: %w", img.Key, dstKey, err)
			}
			h.log(ctx).Debug("object copied", zap.String("from", img.Key), zap.String("to", dstKey))
		}

		// Delete old object
		if err := h.objStorage.DeleteObject(ctx, &abstraction.DeleteObjectInput{
			Key: img.Key,
		}); err != nil {
			// Log error but continue - object in new location already exists
			h.log(ctx).Warn("failed to delete old object after copy",
				zap.String("key", img.Key),
				zap.Error(err),
			)
		}

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
	return logger.Get(ctx).With(zap.String("component", "promote-images-handler"))
}

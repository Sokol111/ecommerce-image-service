package command

import (
	"context"
	"fmt"
	"strings"

	"github.com/Sokol111/ecommerce-commons/pkg/core/logger"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
	"github.com/Sokol111/ecommerce-commons/pkg/persistence/mongo"
	"github.com/Sokol111/ecommerce-image-service/internal/apperrors"
	"github.com/Sokol111/ecommerce-image-service/internal/application/abstraction"
	"github.com/Sokol111/ecommerce-image-service/internal/domain/image"
	"github.com/Sokol111/ecommerce-image-service/internal/event"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

// PromoteImagesCommand represents a request to promote draft images to product
type PromoteImagesCommand struct {
	DraftID   string    // Optional: required only when ImageIDs is empty (promote all draft images)
	ImageIDs  *[]string // Optional: if provided, only these images are promoted
	ProductID string
}

// PromoteImagesCommandHandler handles PromoteImagesCommand
type PromoteImagesCommandHandler interface {
	Handle(ctx context.Context, cmd PromoteImagesCommand) ([]*image.Image, error)
}

type promoteImagesHandler struct {
	repo         image.Repository
	objStorage   abstraction.ObjectStorage
	signer       abstraction.ImgproxySigner
	outbox       outbox.Outbox
	txManager    mongo.TxManager
	smallWidth   int
	largeWidth   int
	imageQuality int
}

func NewPromoteImagesHandler(repo image.Repository, objStorage abstraction.ObjectStorage, signer abstraction.ImgproxySigner, outbox outbox.Outbox, txManager mongo.TxManager, smallWidth, largeWidth, quality int) PromoteImagesCommandHandler {
	return &promoteImagesHandler{
		repo:         repo,
		objStorage:   objStorage,
		signer:       signer,
		outbox:       outbox,
		txManager:    txManager,
		smallWidth:   smallWidth,
		largeWidth:   largeWidth,
		imageQuality: quality,
	}
}

// promoteResult holds the result of the promotion transaction
type promoteResult struct {
	Promoted     []*image.Image
	Sends        []outbox.SendFunc
	OldImageKeys []string // S3 keys of replaced product images for cleanup
}

// copyResult holds the result of copying images
type copyResult struct {
	Image     *image.Image
	SourceKey string
	TargetKey string
	Skipped   bool
}

func (h *promoteImagesHandler) Handle(ctx context.Context, cmd PromoteImagesCommand) ([]*image.Image, error) {
	images, err := h.getImagesToPromote(ctx, cmd)
	if err != nil {
		return nil, err
	}

	if len(images) == 0 {
		h.log(ctx).Debug("no images to promote", zap.String("productID", cmd.ProductID))
		return []*image.Image{}, nil
	}

	// Phase 1: Copy files (without deleting originals)
	copyResults, err := h.copyImages(ctx, images, cmd.ProductID)
	if err != nil {
		return nil, err
	}

	// Phase 2: DB Transaction
	result, err := h.executePromotion(ctx, copyResults, cmd.ProductID)
	if err != nil {
		// Compensation: rollback copied files
		h.rollbackCopiedFiles(ctx, copyResults)
		return nil, err
	}

	// Phase 3: Success - delete old files (best effort)
	h.deleteSourceFiles(ctx, copyResults)
	h.deleteOldProductImages(ctx, result.OldImageKeys)

	h.log(ctx).Debug("images promoted", zap.Int("count", len(result.Promoted)), zap.String("productID", cmd.ProductID))

	h.sendOutboxMessages(ctx, result.Sends)

	return result.Promoted, nil
}

// getImagesToPromote validates the draft and retrieves images for promotion
func (h *promoteImagesHandler) getImagesToPromote(ctx context.Context, cmd PromoteImagesCommand) ([]*image.Image, error) {
	var imageIDs []string
	if cmd.ImageIDs != nil && len(*cmd.ImageIDs) > 0 {
		imageIDs = lo.Uniq(*cmd.ImageIDs)
	}

	// If specific IDs provided, fetch them directly and check their state
	if len(imageIDs) > 0 {
		return h.getSpecificImagesToPromote(ctx, imageIDs, cmd.ProductID)
	}

	if cmd.DraftID == "" {
		return nil, fmt.Errorf("%w: draftId is required when imageIds are not specified", apperrors.ErrInvalidImageOwner)
	}

	// Get all images from draft
	images, err := h.repo.FindByOwner(ctx, string(image.OwnerTypeDraft), cmd.DraftID, nil)
	if err != nil {
		return nil, fmt.Errorf("find draft images: %w", err)
	}

	if len(images) == 0 {
		return nil, fmt.Errorf("%w: %s", apperrors.ErrDraftNotFound, cmd.DraftID)
	}

	return images, nil
}

// getSpecificImagesToPromote fetches specific images and validates their state for idempotency
func (h *promoteImagesHandler) getSpecificImagesToPromote(ctx context.Context, imageIDs []string, productID string) ([]*image.Image, error) {
	images, err := h.repo.FindByIDs(ctx, imageIDs)
	if err != nil {
		return nil, fmt.Errorf("find images by IDs: %w", err)
	}

	if len(images) != len(imageIDs) {
		return nil, fmt.Errorf("%w: some specified image IDs not found", image.ErrImageNotFound)
	}

	var toPromote []*image.Image
	for _, img := range images {
		switch {
		case img.OwnerType == string(image.OwnerTypeDraft):
			toPromote = append(toPromote, img)

		case img.OwnerType == string(image.OwnerTypeProduct) && img.OwnerID == productID:
			h.log(ctx).Debug("image already promoted to target product, skipping",
				zap.String("imageID", img.ID),
				zap.String("productID", productID),
			)

		default:
			return nil, fmt.Errorf("%w: image %s has owner %s/%s", apperrors.ErrInvalidImageOwner, img.ID, img.OwnerType, img.OwnerID)
		}
	}

	return toPromote, nil
}

// copyImages copies S3 objects without deleting originals (Phase 1)
func (h *promoteImagesHandler) copyImages(ctx context.Context, images []*image.Image, productID string) ([]copyResult, error) {
	var results []copyResult

	for _, img := range images {
		result, err := h.copyImage(ctx, img, productID)
		if err != nil {
			// Rollback already copied files on error
			h.rollbackCopiedFiles(ctx, results)
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}

// copyImage copies a single image to target location
func (h *promoteImagesHandler) copyImage(ctx context.Context, img *image.Image, productID string) (copyResult, error) {
	srcPrefix := "drafts/" + img.OwnerID + "/"
	if !strings.HasPrefix(img.Key, srcPrefix) {
		return copyResult{}, fmt.Errorf("image %s has key outside draft prefix: %s", img.ID, img.Key)
	}

	sourceKey := img.Key
	targetKey := "products/" + productID + "/" + strings.TrimPrefix(img.Key, srcPrefix)

	// Check if target already exists (idempotency)
	exists, _ := h.objStorage.ObjectExists(ctx, targetKey) //nolint:errcheck // error means object doesn't exist
	if exists {
		return copyResult{
			Image:     img,
			SourceKey: sourceKey,
			TargetKey: targetKey,
			Skipped:   true,
		}, nil
	}

	// Copy object (without deleting source)
	err := h.objStorage.CopyObject(ctx, &abstraction.CopyObjectInput{
		SourceKey: sourceKey,
		TargetKey: targetKey,
	})
	if err != nil {
		return copyResult{}, fmt.Errorf("copy image %s: %w", img.ID, err)
	}

	return copyResult{
		Image:     img,
		SourceKey: sourceKey,
		TargetKey: targetKey,
		Skipped:   false,
	}, nil
}

// rollbackCopiedFiles deletes copied files on transaction failure (compensation)
func (h *promoteImagesHandler) rollbackCopiedFiles(ctx context.Context, results []copyResult) {
	var keysToDelete []string
	for _, r := range results {
		if !r.Skipped {
			keysToDelete = append(keysToDelete, r.TargetKey)
		}
	}

	if len(keysToDelete) == 0 {
		return
	}

	if err := h.objStorage.DeleteObjects(ctx, keysToDelete); err != nil {
		h.log(ctx).Error("failed to rollback copied files",
			zap.Strings("keys", keysToDelete),
			zap.Error(err),
		)
	} else {
		h.log(ctx).Warn("rolled back copied files due to transaction failure",
			zap.Int("count", len(keysToDelete)),
		)
	}
}

// deleteSourceFiles deletes original files after successful transaction
func (h *promoteImagesHandler) deleteSourceFiles(ctx context.Context, results []copyResult) {
	var keysToDelete []string
	for _, r := range results {
		if !r.Skipped {
			keysToDelete = append(keysToDelete, r.SourceKey)
		}
	}

	if len(keysToDelete) == 0 {
		return
	}

	if err := h.objStorage.DeleteObjects(ctx, keysToDelete); err != nil {
		h.log(ctx).Warn("failed to delete source files after promotion",
			zap.Strings("keys", keysToDelete),
			zap.Error(err),
		)
	}
}

// executePromotion runs DB updates and outbox creation in a transaction
func (h *promoteImagesHandler) executePromotion(ctx context.Context, copyResults []copyResult, productID string) (*promoteResult, error) {
	return mongo.WithTransaction(ctx, h.txManager, func(txCtx context.Context) (*promoteResult, error) {
		return h.promoteInTransaction(txCtx, copyResults, productID)
	})
}

// promoteInTransaction performs the actual promotion logic within a transaction
func (h *promoteImagesHandler) promoteInTransaction(ctx context.Context, copyResults []copyResult, productID string) (*promoteResult, error) {
	// Soft-delete existing product images
	oldImageKeys, err := h.softDeleteOldProductImages(ctx, productID)
	if err != nil {
		return nil, err
	}

	var promoted []*image.Image
	var sends []outbox.SendFunc

	for _, cr := range copyResults {
		// Update domain object
		if err := cr.Image.PromoteToProduct(productID, cr.TargetKey); err != nil {
			return nil, fmt.Errorf("promote image %s: %w", cr.Image.ID, err)
		}

		updated, err := h.repo.Update(ctx, cr.Image)
		if err != nil {
			return nil, fmt.Errorf("update image after promote: %w", err)
		}
		promoted = append(promoted, updated)

		smallImageURL := h.buildImageURL(updated.Key, h.smallWidth)
		largeImageURL := h.buildImageURL(updated.Key, h.largeWidth)
		msg := event.NewProductImagePromotedOutboxMessage(ctx, productID, updated.ID, smallImageURL, largeImageURL)

		send, err := h.outbox.Create(ctx, msg)
		if err != nil {
			return nil, fmt.Errorf("create outbox: %w", err)
		}
		sends = append(sends, send)
	}

	return &promoteResult{
		Promoted:     promoted,
		Sends:        sends,
		OldImageKeys: oldImageKeys,
	}, nil
}

// sendOutboxMessages sends all outbox messages after successful transaction
func (h *promoteImagesHandler) sendOutboxMessages(ctx context.Context, sends []outbox.SendFunc) {
	for _, send := range sends {
		_ = send(ctx) //nolint:errcheck // best-effort send, errors already logged in outbox
	}
}

// softDeleteOldProductImages marks existing product images as deleted within the transaction
func (h *promoteImagesHandler) softDeleteOldProductImages(ctx context.Context, productID string) ([]string, error) {
	existing, err := h.repo.FindByOwner(ctx, string(image.OwnerTypeProduct), productID, nil)
	if err != nil {
		return nil, fmt.Errorf("find existing product images: %w", err)
	}

	if len(existing) == 0 {
		return nil, nil
	}

	var keys []string
	for _, img := range existing {
		img.MarkAsDeleted()
		if _, err := h.repo.Update(ctx, img); err != nil {
			return nil, fmt.Errorf("soft-delete old image %s: %w", img.ID, err)
		}
		keys = append(keys, img.Key)
	}

	h.log(ctx).Debug("soft-deleted old product images",
		zap.Int("count", len(existing)),
		zap.String("productID", productID),
	)

	return keys, nil
}

// deleteOldProductImages deletes S3 objects of replaced product images (best effort)
func (h *promoteImagesHandler) deleteOldProductImages(ctx context.Context, keys []string) {
	if len(keys) == 0 {
		return
	}

	if err := h.objStorage.DeleteObjects(ctx, keys); err != nil {
		h.log(ctx).Warn("failed to delete old product image files",
			zap.Strings("keys", keys),
			zap.Error(err),
		)
	}
}

func (h *promoteImagesHandler) buildImageURL(key string, width int) string {
	quality := h.imageQuality
	return h.signer.BuildURL(key, abstraction.SignerOptions{Width: &width, Quality: &quality})
}

func (h *promoteImagesHandler) log(ctx context.Context) *zap.Logger {
	return logger.Get(ctx).With(zap.String("component", "promote-images-handler"))
}

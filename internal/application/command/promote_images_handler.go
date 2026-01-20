package command

import (
	"context"
	"fmt"
	"strings"

	"github.com/Sokol111/ecommerce-commons/pkg/core/logger"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
	"github.com/Sokol111/ecommerce-commons/pkg/persistence"
	"github.com/Sokol111/ecommerce-image-service/internal/application/abstraction"
	"github.com/Sokol111/ecommerce-image-service/internal/domain/image"
	"github.com/Sokol111/ecommerce-image-service/internal/event"
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
	signer     abstraction.ImgproxySigner
	outbox     outbox.Outbox
	txManager  persistence.TxManager
}

func NewPromoteImagesHandler(repo image.Repository, objStorage abstraction.ObjectStorage, signer abstraction.ImgproxySigner, outbox outbox.Outbox, txManager persistence.TxManager) PromoteImagesCommandHandler {
	return &promoteImagesHandler{
		repo:       repo,
		objStorage: objStorage,
		signer:     signer,
		outbox:     outbox,
		txManager:  txManager,
	}
}

// promoteResult holds the result of the promotion transaction
type promoteResult struct {
	Promoted []*image.Image
	Sends    []outbox.SendFunc
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

	// Phase 1: Copy files (without deleting originals)
	copyResults, err := h.copyImages(ctx, images, cmd.DraftID, cmd.ProductID)
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

	h.log(ctx).Debug("images promoted", zap.Int("count", len(result.Promoted)), zap.String("productID", cmd.ProductID))

	h.sendOutboxMessages(ctx, result.Sends)

	return result.Promoted, nil
}

// getImagesToPromote validates the draft and retrieves images for promotion
func (h *promoteImagesHandler) getImagesToPromote(ctx context.Context, cmd PromoteImagesCommand) ([]*image.Image, error) {
	var imageIDs []string
	if cmd.ImageIDs != nil && len(*cmd.ImageIDs) > 0 {
		imageIDs = *cmd.ImageIDs
	}

	// Get images to promote (either specified or all from draft)
	images, err := h.repo.FindByOwner(ctx, "productDraft", cmd.DraftID, imageIDs)
	if err != nil {
		return nil, fmt.Errorf("find draft images: %w", err)
	}

	if len(images) == 0 {
		return nil, fmt.Errorf("draft %s not found or has no images", cmd.DraftID)
	}

	// Validate all specified imageIDs were found
	if imageIDs != nil && len(imageIDs) != len(images) {
		return nil, fmt.Errorf("some specified image IDs not found in draft %s", cmd.DraftID)
	}

	return images, nil
}

// copyImages copies S3 objects without deleting originals (Phase 1)
func (h *promoteImagesHandler) copyImages(ctx context.Context, images []*image.Image, draftID, productID string) ([]copyResult, error) {
	srcPrefix := "product-drafts/" + draftID + "/"
	var results []copyResult

	for _, img := range images {
		result, err := h.copyImage(ctx, img, srcPrefix, productID)
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
func (h *promoteImagesHandler) copyImage(ctx context.Context, img *image.Image, srcPrefix, productID string) (copyResult, error) {
	if !strings.HasPrefix(img.Key, srcPrefix) {
		return copyResult{}, fmt.Errorf("image %s has key outside draft prefix: %s", img.ID, img.Key)
	}

	sourceKey := img.Key
	targetKey := "products/" + productID + "/" + strings.TrimPrefix(img.Key, srcPrefix)

	// Check if target already exists (idempotency)
	exists, _ := h.objStorage.ObjectExists(ctx, targetKey)
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
	return persistence.WithTransaction(ctx, h.txManager, func(txCtx context.Context) (*promoteResult, error) {
		return h.promoteInTransaction(txCtx, copyResults, productID)
	})
}

// promoteInTransaction performs the actual promotion logic within a transaction
func (h *promoteImagesHandler) promoteInTransaction(ctx context.Context, copyResults []copyResult, productID string) (*promoteResult, error) {
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

		imageURL := h.buildImageURL(updated.Key)
		msg, err := event.NewProductImagePromotedOutboxMessage(ctx, productID, updated.ID, imageURL)
		if err != nil {
			return nil, fmt.Errorf("create outbox message: %w", err)
		}

		send, err := h.outbox.Create(ctx, msg)
		if err != nil {
			return nil, fmt.Errorf("create outbox: %w", err)
		}
		sends = append(sends, send)
	}

	return &promoteResult{
		Promoted: promoted,
		Sends:    sends,
	}, nil
}

// sendOutboxMessages sends all outbox messages after successful transaction
func (h *promoteImagesHandler) sendOutboxMessages(ctx context.Context, sends []outbox.SendFunc) {
	for _, send := range sends {
		_ = send(ctx)
	}
}

func (h *promoteImagesHandler) buildImageURL(key string) string {
	w := 400
	quality := 85
	return h.signer.BuildURL(key, abstraction.SignerOptions{Width: &w, Quality: &quality})
}

func (h *promoteImagesHandler) log(ctx context.Context) *zap.Logger {
	return logger.Get(ctx).With(zap.String("component", "promote-images-handler"))
}

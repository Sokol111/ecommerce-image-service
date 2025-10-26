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

// ConfirmUploadCommand represents a request to confirm an image upload
type ConfirmUploadCommand struct {
	Alt       string
	Checksum  *string
	Key       string
	Mime      string
	OwnerType string
	OwnerID   string
	Role      string
}

// ConfirmUploadCommandHandler handles ConfirmUploadCommand
type ConfirmUploadCommandHandler interface {
	Handle(ctx context.Context, cmd ConfirmUploadCommand) (*image.Image, error)
}

type confirmUploadHandler struct {
	repo           image.Repository
	objStorage     abstraction.ObjectStorage
	maxUploadBytes int64
}

func NewConfirmUploadHandler(repo image.Repository, storage abstraction.ObjectStorage, maxUploadBytes int64) ConfirmUploadCommandHandler {
	return &confirmUploadHandler{
		repo:           repo,
		objStorage:     storage,
		maxUploadBytes: maxUploadBytes,
	}
}

func (h *confirmUploadHandler) Handle(ctx context.Context, cmd ConfirmUploadCommand) (*image.Image, error) {
	// Validate key matches expected owner prefix
	prefix, err := getPrefixByOwnerType(cmd.OwnerType)
	if err != nil {
		return nil, fmt.Errorf("failed to get prefix by owner type: %w", err)
	}
	expectedPrefix := prefix + cmd.OwnerID + "/"

	if !strings.HasPrefix(cmd.Key, expectedPrefix) {
		return nil, fmt.Errorf("key does not match expected owner prefix")
	}

	// Verify object exists in S3
	ho, err := h.objStorage.HeadObject(ctx, &abstraction.HeadObjectInput{
		Key: cmd.Key,
	})
	if err != nil {
		return nil, fmt.Errorf("head object: %w", err)
	}

	size := int64(0)
	if ho.ContentLength != nil {
		size = *ho.ContentLength
	}

	// Validate size
	if h.maxUploadBytes > 0 && size > h.maxUploadBytes {
		_ = h.objStorage.DeleteObject(ctx, &abstraction.DeleteObjectInput{
			Key: cmd.Key,
		})
		return nil, fmt.Errorf("file too large: max %d bytes", h.maxUploadBytes)
	}

	// Create domain image
	img, err := image.NewImage(cmd.Alt, cmd.OwnerType, cmd.OwnerID, cmd.Role, cmd.Key, cmd.Mime, size)
	if err != nil {
		return nil, fmt.Errorf("create image: %w", err)
	}

	// Save to repository
	if err := h.repo.Save(ctx, img); err != nil {
		return nil, fmt.Errorf("save image: %w", err)
	}

	h.log(ctx).Debug("image upload confirmed", zap.String("id", img.ID), zap.String("key", img.Key))

	return img, nil
}

func (h *confirmUploadHandler) log(ctx context.Context) *zap.Logger {
	return logger.FromContext(ctx).With(zap.String("component", "confirm-upload-handler"))
}

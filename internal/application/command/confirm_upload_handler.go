package command

import (
	"context"
	"fmt"
	"strings"

	"github.com/Sokol111/ecommerce-commons/pkg/core/logger"
	"github.com/Sokol111/ecommerce-image-service/internal/apperrors"
	"github.com/Sokol111/ecommerce-image-service/internal/application/abstraction"
	"github.com/Sokol111/ecommerce-image-service/internal/domain/image"
	"go.uber.org/zap"
)

// ConfirmUploadCommand represents a request to confirm an image upload
type ConfirmUploadCommand struct {
	UploadToken string
	Alt         string
	Checksum    *string
	Role        string
}

// ConfirmUploadCommandHandler handles ConfirmUploadCommand
type ConfirmUploadCommandHandler interface {
	Handle(ctx context.Context, cmd ConfirmUploadCommand) (*image.Image, error)
}

type confirmUploadHandler struct {
	repo           image.Repository
	objStorage     abstraction.ObjectStorage
	tokenService   abstraction.TokenService
	maxUploadBytes int64
}

func NewConfirmUploadHandler(repo image.Repository, storage abstraction.ObjectStorage, tokenService abstraction.TokenService, maxUploadBytes int64) ConfirmUploadCommandHandler {
	return &confirmUploadHandler{
		repo:           repo,
		objStorage:     storage,
		tokenService:   tokenService,
		maxUploadBytes: maxUploadBytes,
	}
}

func (h *confirmUploadHandler) Handle(ctx context.Context, cmd ConfirmUploadCommand) (*image.Image, error) {
	// Validate and extract claims from upload token
	claims, err := h.tokenService.ValidateUploadToken(ctx, cmd.UploadToken)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", apperrors.ErrInvalidUploadToken, err)
	}

	// Verify object exists in S3
	ho, err := h.objStorage.HeadObject(ctx, &abstraction.HeadObjectInput{
		Key: claims.Key,
	})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("%w: %s", apperrors.ErrObjectNotFound, claims.Key)
		}
		return nil, fmt.Errorf("head object: %w", err)
	}

	size := int64(0)
	if ho.ContentLength != nil {
		size = *ho.ContentLength
	}

	// Validate size matches expected size from token
	if claims.Size > 0 && size != claims.Size {
		h.log(ctx).Warn("uploaded file size mismatch, deleting",
			zap.Int64("expected", claims.Size),
			zap.Int64("actual", size),
			zap.String("key", claims.Key),
		)
		_ = h.objStorage.DeleteObject(ctx, &abstraction.DeleteObjectInput{ //nolint:errcheck // best-effort cleanup
			Key: claims.Key,
		})
		return nil, fmt.Errorf("%w: expected %d bytes, got %d", image.ErrInvalidSize, claims.Size, size)
	}

	// Validate size limit
	if h.maxUploadBytes > 0 && size > h.maxUploadBytes {
		_ = h.objStorage.DeleteObject(ctx, &abstraction.DeleteObjectInput{ //nolint:errcheck // best-effort cleanup
			Key: claims.Key,
		})
		return nil, fmt.Errorf("%w: max %d bytes", image.ErrImageTooLarge, h.maxUploadBytes)
	}

	// Extract content type from token (S3 doesn't return ContentType in HeadObject in our abstraction)
	mime := claims.ContentType

	// Create domain image with data from validated token
	img, err := image.NewImage(cmd.Alt, claims.OwnerType, claims.OwnerID, claims.Role, claims.Key, mime, size)
	if err != nil {
		return nil, fmt.Errorf("create image: %w", err)
	}

	// Save to repository
	if err := h.repo.Insert(ctx, img); err != nil {
		return nil, fmt.Errorf("insert image: %w", err)
	}

	h.log(ctx).Debug("image upload confirmed",
		zap.String("id", img.ID),
		zap.String("key", img.Key),
		zap.String("ownerType", claims.OwnerType),
		zap.String("ownerId", claims.OwnerID),
	)

	return img, nil
}

func (h *confirmUploadHandler) log(ctx context.Context) *zap.Logger {
	return logger.Get(ctx).With(zap.String("component", "confirm-upload-handler"))
}

package image //nolint:revive // package name intentional

import (
	"context"
	"fmt"
	"time"

	"github.com/Sokol111/ecommerce-commons/pkg/core/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// CreatePresignCommand represents a request to create a presigned URL
type CreatePresignCommand struct {
	ContentType string
	Filename    string
	OwnerType   string
	OwnerID     string
	Role        string
	Size        int64
}

// CreatePresignResult contains the presigned URL response for uploading
type CreatePresignResult struct {
	UploadURL   string
	UploadToken string
	ExpiresIn   int
}

// CreatePresignCommandHandler handles CreatePresignCommand
type CreatePresignCommandHandler interface {
	Handle(ctx context.Context, cmd CreatePresignCommand) (*CreatePresignResult, error)
}

type createPresignHandler struct {
	presigner      Presigner
	tokenService   TokenService
	presignTTL     time.Duration
	maxUploadBytes int64
}

func NewCreatePresignHandler(presigner Presigner, tokenService TokenService, presignTTL time.Duration, maxUploadBytes int64) CreatePresignCommandHandler {
	return &createPresignHandler{
		presigner:      presigner,
		tokenService:   tokenService,
		presignTTL:     presignTTL,
		maxUploadBytes: maxUploadBytes,
	}
}

func (h *createPresignHandler) Handle(ctx context.Context, cmd CreatePresignCommand) (*CreatePresignResult, error) {
	// Validate content type using domain rules
	ext := GetExtensionForMimeType(cmd.ContentType)
	if ext == "" {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedMimeType, cmd.ContentType)
	}

	// Validate size
	if cmd.Size <= 0 {
		return nil, fmt.Errorf("%w: size must be positive", ErrInvalidSize)
	}
	if h.maxUploadBytes > 0 && cmd.Size > h.maxUploadBytes {
		return nil, fmt.Errorf("%w: max %d bytes", ErrImageTooLarge, h.maxUploadBytes)
	}

	// Get prefix by owner type
	prefix, err := getPrefixByOwnerType(cmd.OwnerType)
	if err != nil {
		return nil, fmt.Errorf("failed to get prefix by owner type: %w", err)
	}

	// Generate key
	key := prefix + cmd.OwnerID + "/" + uuid.New().String() + ext

	presignResult, err := h.presigner.CreatePresignedUpload(ctx, &PresignedUploadInput{
		Key:         key,
		ContentType: cmd.ContentType,
		Size:        cmd.Size,
	})
	if err != nil {
		return nil, fmt.Errorf("create presigned upload: %w", err)
	}

	// Generate signed JWT token with upload metadata
	uploadToken, err := h.tokenService.GenerateUploadToken(ctx, &UploadTokenClaims{
		Key:         key,
		OwnerType:   cmd.OwnerType,
		OwnerID:     cmd.OwnerID,
		Role:        cmd.Role,
		ContentType: cmd.ContentType,
		Size:        cmd.Size,
	}, h.presignTTL)
	if err != nil {
		return nil, fmt.Errorf("generate upload token: %w", err)
	}

	h.log(ctx).Debug("Presigned PUT URL created with signed Content-Type and Content-Length",
		zap.String("key", key),
		zap.Int64("size", cmd.Size),
	)

	return &CreatePresignResult{
		UploadURL:   presignResult.URL,
		UploadToken: uploadToken,
		ExpiresIn:   presignResult.TTLSeconds,
	}, nil
}

func (h *createPresignHandler) log(ctx context.Context) *zap.Logger {
	return logger.Get(ctx).With(zap.String("component", "create-presign-handler"))
}

func getPrefixByOwnerType(ownerType string) (string, error) {
	switch OwnerType(ownerType) {
	case OwnerTypeDraft:
		return "drafts/", nil
	case OwnerTypeProduct:
		return "products/", nil
	case OwnerTypeUser:
		return "users/", nil
	default:
		return "", fmt.Errorf("unsupported owner type: %s", ownerType)
	}
}

package command

import (
	"context"
	"fmt"
	"strings"

	"github.com/Sokol111/ecommerce-commons/pkg/core/logger"
	"github.com/Sokol111/ecommerce-image-service/internal/application/abstraction"
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

// CreatePresignResult contains the presigned URL and metadata
type CreatePresignResult struct {
	UploadURL       string
	Key             string
	ExpiresIn       int
	RequiredHeaders map[string]string
}

// CreatePresignCommandHandler handles CreatePresignCommand
type CreatePresignCommandHandler interface {
	Handle(ctx context.Context, cmd CreatePresignCommand) (*CreatePresignResult, error)
}

type createPresignHandler struct {
	presigner abstraction.Presigner
}

func NewCreatePresignHandler(presigner abstraction.Presigner) CreatePresignCommandHandler {
	return &createPresignHandler{
		presigner: presigner,
	}
}

func (h *createPresignHandler) Handle(ctx context.Context, cmd CreatePresignCommand) (*CreatePresignResult, error) {
	// Validate content type
	extByCT := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/webp": ".webp",
		"image/avif": ".avif",
	}
	ext := extByCT[strings.ToLower(cmd.ContentType)]
	if ext == "" {
		return nil, fmt.Errorf("unsupported content type: %s", cmd.ContentType)
	}

	// Get prefix by owner type
	prefix, err := getPrefixByOwnerType(cmd.OwnerType)
	if err != nil {
		return nil, fmt.Errorf("failed to get prefix by owner type: %w", err)
	}

	// Generate key
	key := prefix + cmd.OwnerID + "/" + uuid.New().String() + ext

	// Create presigned URL
	out, err := h.presigner.PresignPutObject(ctx, &abstraction.PresignPutObjectInput{
		Key:         key,
		ContentType: cmd.ContentType,
	})
	if err != nil {
		return nil, fmt.Errorf("presign put: %w", err)
	}

	h.log(ctx).Debug("presigned URL created", zap.String("key", key))

	return &CreatePresignResult{
		UploadURL: out.URL,
		Key:       key,
		ExpiresIn: out.TTLSeconds,
		RequiredHeaders: map[string]string{
			"Content-Type": cmd.ContentType,
		},
	}, nil
}

func (h *createPresignHandler) log(ctx context.Context) *zap.Logger {
	return logger.FromContext(ctx).With(zap.String("component", "create-presign-handler"))
}

func getPrefixByOwnerType(ownerType string) (string, error) {
	switch ownerType {
	case "productDraft":
		return "product-drafts/", nil
	case "product":
		return "products/", nil
	case "user":
		return "users/", nil
	default:
		return "", fmt.Errorf("unsupported owner type: %s", ownerType)
	}
}

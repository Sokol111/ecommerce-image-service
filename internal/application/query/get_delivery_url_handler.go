package query

import (
	"context"
	"fmt"
	"time"

	"github.com/Sokol111/ecommerce-commons/pkg/core/logger"
	"github.com/Sokol111/ecommerce-image-service/internal/application/abstraction"
	"github.com/Sokol111/ecommerce-image-service/internal/domain/image"
	"go.uber.org/zap"
)

// GetDeliveryURLQuery represents a query to get a delivery URL for an image
type GetDeliveryURLQuery struct {
	ImageID string
	Width   *int
	Height  *int
	Fit     *string
	Quality *int
	DPR     *float32
	Format  *string
	Expires *time.Time
}

// GetDeliveryURLResult represents the result of getting a delivery URL
type GetDeliveryURLResult struct {
	URL       string
	ExpiresAt *time.Time
}

// GetDeliveryURLQueryHandler handles GetDeliveryURLQuery
type GetDeliveryURLQueryHandler interface {
	Handle(ctx context.Context, query GetDeliveryURLQuery) (*GetDeliveryURLResult, error)
}

type getDeliveryURLHandler struct {
	repo   image.Repository
	signer abstraction.ImgproxySigner
	bucket string
}

func NewGetDeliveryURLHandler(repo image.Repository, signer abstraction.ImgproxySigner, bucket string) GetDeliveryURLQueryHandler {
	return &getDeliveryURLHandler{
		repo:   repo,
		signer: signer,
		bucket: bucket,
	}
}

func (h *getDeliveryURLHandler) Handle(ctx context.Context, query GetDeliveryURLQuery) (*GetDeliveryURLResult, error) {
	// Get image from repository
	img, err := h.repo.FindByID(ctx, query.ImageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get image by id: %w", err)
	}

	// Build source URL for imgproxy
	source := fmt.Sprintf("s3://%s/%s", h.bucket, img.Key)

	// Build imgproxy URL
	imgproxyURL := h.signer.BuildURL(source, abstraction.SignerOptions{
		Width:   query.Width,
		Height:  query.Height,
		Fit:     query.Fit,
		Quality: query.Quality,
		DPR:     query.DPR,
		Format:  query.Format,
		Expires: query.Expires,
	})

	h.log(ctx).Debug("delivery URL generated", zap.String("imageID", query.ImageID))

	return &GetDeliveryURLResult{
		URL:       imgproxyURL,
		ExpiresAt: query.Expires,
	}, nil
}

func (h *getDeliveryURLHandler) log(ctx context.Context) *zap.Logger {
	return logger.FromContext(ctx).With(zap.String("component", "get-delivery-url-handler"))
}

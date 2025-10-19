package http

import (
	"context"
	"fmt"
	"time"

	"github.com/Sokol111/ecommerce-image-service-api/api"
	"github.com/Sokol111/ecommerce-image-service/internal/model"
)

type imageHandler struct {
	model.ImageService
}

func newImageHandler(service model.ImageService) api.StrictServerInterface {
	return &imageHandler{
		ImageService: service,
	}
}

func (h *imageHandler) CreatePresign(ctx context.Context, request api.CreatePresignRequestObject) (api.CreatePresignResponseObject, error) {
	switch request.Body.OwnerType {
	case api.ProductDraft:
		response, err := h.ImageService.CreateDraftPresign(ctx, model.CreatePresignForDraftDTO{
			ContentType: string(request.Body.ContentType),
			Filename:    request.Body.Filename,
			DraftId:     request.Body.OwnerId,
			Size:        int64(request.Body.Size),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create presign URL: %w", err)
		}
		return api.CreatePresign200JSONResponse{
			UploadUrl:       response.UploadUrl,
			Key:             response.Key,
			ExpiresIn:       response.ExpiresIn,
			RequiredHeaders: response.RequiredHeaders,
		}, nil
	case api.Product:
		return nil, fmt.Errorf("cannot upload image for product directly, only for product draft")
		// prefix = "products/" + request.Body.OwnerId + "/"
	case api.User:
		return nil, fmt.Errorf("cannot upload image for user")
		// prefix = "uploads/pending/" + request.Body.OwnerId + "/"
	default:
		return nil, fmt.Errorf("unsupported ownerType: %s", request.Body.OwnerType)
	}
}

func (h *imageHandler) ConfirmUpload(ctx context.Context, request api.ConfirmUploadRequestObject) (api.ConfirmUploadResponseObject, error) {
	switch request.Body.OwnerType {
	case api.ProductDraft:
		var response, err = h.ImageService.ConfirmDraftUpload(ctx, model.ConfirmDraftUploadDTO{
			Alt:      request.Body.Alt,
			Key:      request.Body.Key,
			Mime:     string(request.Body.Mime),
			DraftId:  request.Body.OwnerId,
			Checksum: request.Body.Checksum,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to confirm upload: %w", err)
		}
		return api.ConfirmUpload201JSONResponse{
			Id:         response.Id,
			Alt:        response.Alt,
			OwnerType:  api.OwnerType(response.OwnerType),
			OwnerId:    response.OwnerId,
			Role:       api.ImageRole(response.Role),
			Key:        response.Key,
			Mime:       response.Mime,
			Size:       int(response.Size),
			Status:     api.ImageStatus(response.Status),
			CreatedAt:  response.CreatedAt,
			ModifiedAt: response.ModifiedAt,
		}, nil
	case api.Product:
		return nil, fmt.Errorf("cannot upload image for product directly, only for product draft")
	case api.User:
		return nil, fmt.Errorf("cannot upload image for user")
	default:
		return nil, fmt.Errorf("unsupported ownerType: %s", request.Body.OwnerType)
	}
}

func (h *imageHandler) PromoteImages(ctx context.Context, request api.PromoteImagesRequestObject) (api.PromoteImagesResponseObject, error) {
	images, err := h.ImageService.PromoteDraftImages(ctx, model.PromoteDraftDTO{
		DraftId:   request.Body.DraftId,
		Images:    request.Body.Images,
		Move:      request.Body.Move,
		ProductId: request.Body.ProductId,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to promote images: %w", err)
	}
	promoted := make([]api.Image, 0, len(images))
	for _, image := range images {
		promoted = append(promoted, api.Image{
			Id:         image.Id,
			Alt:        image.Alt,
			OwnerType:  api.OwnerType(image.OwnerType),
			OwnerId:    image.OwnerId,
			Role:       api.ImageRole(image.Role),
			Key:        image.Key,
			Mime:       image.Mime,
			Size:       int(image.Size),
			Status:     api.ImageStatus(image.Status),
			CreatedAt:  image.CreatedAt,
			ModifiedAt: image.ModifiedAt,
		})
	}

	return api.PromoteImages200JSONResponse{Promoted: &promoted}, nil
}

func (h *imageHandler) GetDeliveryUrl(ctx context.Context, request api.GetDeliveryUrlRequestObject) (api.GetDeliveryUrlResponseObject, error) {
	var fit *string
	if request.Params.Fit != nil {
		f := string(*request.Params.Fit)
		fit = &f
	}
	var format *string
	if request.Params.Format != nil {
		f := string(*request.Params.Format)
		format = &f
	}
	var expires *time.Time
	if request.Params.TtlSeconds != nil {
		t := time.Now().Add(time.Duration(*request.Params.TtlSeconds) * time.Second)
		expires = &t
	}
	url, expires, err := h.ImageService.GetDeliveryUrl(ctx, model.GetDeliveryUrlDTO{
		ImageId: request.Id,
		Width:   request.Params.W,
		Height:  request.Params.H,
		Fit:     fit,
		Quality: request.Params.Quality,
		DPR:     request.Params.Dpr,
		Format:  format,
		Expires: expires,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get delivery URL: %w", err)
	}

	response := api.GetDeliveryUrl200JSONResponse{
		Url:       &url,
		ExpiresAt: expires,
	}
	return response, nil
}

// DeleteImage implements api.StrictServerInterface.
func (h *imageHandler) DeleteImage(ctx context.Context, request api.DeleteImageRequestObject) (api.DeleteImageResponseObject, error) {
	panic("unimplemented")
}

// GetImage implements api.StrictServerInterface.
func (h *imageHandler) GetImage(ctx context.Context, request api.GetImageRequestObject) (api.GetImageResponseObject, error) {
	panic("unimplemented")
}

// ListImages implements api.StrictServerInterface.
func (h *imageHandler) ListImages(ctx context.Context, request api.ListImagesRequestObject) (api.ListImagesResponseObject, error) {
	panic("unimplemented")
}

// ProcessImage implements api.StrictServerInterface.
func (h *imageHandler) ProcessImage(ctx context.Context, request api.ProcessImageRequestObject) (api.ProcessImageResponseObject, error) {
	panic("unimplemented")
}

// S3Webhook implements api.StrictServerInterface.
func (h *imageHandler) S3Webhook(ctx context.Context, request api.S3WebhookRequestObject) (api.S3WebhookResponseObject, error) {
	panic("unimplemented")
}

// UpdateImage implements api.StrictServerInterface.
func (h *imageHandler) UpdateImage(ctx context.Context, request api.UpdateImageRequestObject) (api.UpdateImageResponseObject, error) {
	panic("unimplemented")
}

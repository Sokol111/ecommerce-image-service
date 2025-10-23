package http

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Sokol111/ecommerce-commons/pkg/observability"
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
	case api.ProductDraft, api.Product:
		response, err := h.ImageService.CreatePresign(ctx, model.CreatePresignDTO{
			ContentType: string(request.Body.ContentType),
			Filename:    request.Body.Filename,
			OwnerType:   string(request.Body.OwnerType),
			OwnerId:     request.Body.OwnerId,
			Size:        int64(request.Body.Size),
		})
		if err != nil {
			traceId := observability.GetTraceId(ctx)
			return api.CreatePresign500ApplicationProblemPlusJSONResponse(api.Problem{
				Title:   "Internal Server Error",
				Status:  500,
				TraceId: &traceId,
			}), nil
		}
		return api.CreatePresign200JSONResponse{
			UploadUrl:       response.UploadUrl,
			Key:             response.Key,
			ExpiresIn:       response.ExpiresIn,
			RequiredHeaders: response.RequiredHeaders,
		}, nil
	case api.User:
		detail := "cannot upload image for user"
		return api.CreatePresign500ApplicationProblemPlusJSONResponse(api.Problem{
			Title:  "Internal Server Error",
			Detail: &detail,
			Status: 500,
		}), nil
	default:
		detail := fmt.Sprintf("unsupported ownerType: %s", request.Body.OwnerType)
		return api.CreatePresign500ApplicationProblemPlusJSONResponse(api.Problem{
			Title:  "Internal Server Error",
			Detail: &detail,
			Status: 500,
		}), nil
	}
}

func (h *imageHandler) ConfirmUpload(ctx context.Context, request api.ConfirmUploadRequestObject) (api.ConfirmUploadResponseObject, error) {
	switch request.Body.OwnerType {
	case api.ProductDraft:
		var response, err = h.ImageService.ConfirmUpload(ctx, model.ConfirmUploadDTO{
			Alt:       request.Body.Alt,
			Key:       request.Body.Key,
			Mime:      string(request.Body.Mime),
			Role:      string(request.Body.Role),
			OwnerType: string(request.Body.OwnerType),
			OwnerId:   request.Body.OwnerId,
			Checksum:  request.Body.Checksum,
		})
		if err != nil {
			traceId := observability.GetTraceId(ctx)
			return api.ConfirmUpload500ApplicationProblemPlusJSONResponse(api.Problem{
				Title:   "Internal Server Error",
				Status:  500,
				TraceId: &traceId,
			}), nil
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
		detail := "cannot upload image for product directly, only for product draft"
		return api.ConfirmUpload500ApplicationProblemPlusJSONResponse(api.Problem{
			Title:  "Internal Server Error",
			Detail: &detail,
			Status: 500,
		}), nil
	case api.User:
		detail := "cannot upload image for user"
		return api.ConfirmUpload500ApplicationProblemPlusJSONResponse(api.Problem{
			Title:  "Internal Server Error",
			Detail: &detail,
			Status: 500,
		}), nil
	default:
		detail := fmt.Sprintf("unsupported ownerType: %s", request.Body.OwnerType)
		return api.ConfirmUpload500ApplicationProblemPlusJSONResponse(api.Problem{
			Title:  "Internal Server Error",
			Detail: &detail,
			Status: 500,
		}), nil
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
		traceId := observability.GetTraceId(ctx)
		return api.PromoteImages500ApplicationProblemPlusJSONResponse(api.Problem{
			Title:   "Internal Server Error",
			Status:  500,
			TraceId: &traceId,
		}), nil
	}
	promoted := make([]api.Image, 0, len(images))
	for _, image := range images {
		promoted = append(promoted, *toAPI(image))
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
		traceId := observability.GetTraceId(ctx)
		return api.GetDeliveryUrl500ApplicationProblemPlusJSONResponse(api.Problem{
			Title:   "Internal Server Error",
			Status:  500,
			TraceId: &traceId,
		}), nil
	}

	response := api.GetDeliveryUrl200JSONResponse{
		Url:       &url,
		ExpiresAt: expires,
	}
	return response, nil
}

func (h *imageHandler) DeleteImage(ctx context.Context, request api.DeleteImageRequestObject) (api.DeleteImageResponseObject, error) {
	hard := false
	if request.Params.Hard != nil {
		hard = *request.Params.Hard
	}
	err := h.ImageService.DeleteImage(ctx, request.Id, hard)
	if err != nil {
		return nil, fmt.Errorf("failed to delete image [%v]: %w", request.Id, err)
	}
	return api.DeleteImage204Response{}, nil
}

func (h *imageHandler) GetImage(ctx context.Context, request api.GetImageRequestObject) (api.GetImageResponseObject, error) {
	img, err := h.ImageService.GetImageById(ctx, request.Id)
	traceId := observability.GetTraceId(ctx)
	if err != nil {
		if errors.Is(err, model.ErrEntityNotFound) {
			return api.GetImage404ApplicationProblemPlusJSONResponse(api.Problem{
				Title:   "Image not found",
				Status:  404,
				TraceId: &traceId,
			}), nil
		}
		return api.GetImage500ApplicationProblemPlusJSONResponse(api.Problem{
			Title:   "Internal Server Error",
			Status:  500,
			TraceId: &traceId,
		}), nil
	}
	if img.Status == "deleted" {
		return api.GetImage404ApplicationProblemPlusJSONResponse(
			api.Problem{
				Title:   "Image deleted",
				Status:  404,
				TraceId: &traceId,
			}), nil
	}
	return api.GetImage200JSONResponse(*toAPI(img)), nil
}

// ListImages implements api.StrictServerInterface.
func (h *imageHandler) ListImages(ctx context.Context, request api.ListImagesRequestObject) (api.ListImagesResponseObject, error) {
	panic("unimplemented")
}

// ProcessImage implements api.StrictServerInterface.
func (h *imageHandler) ProcessImage(ctx context.Context, request api.ProcessImageRequestObject) (api.ProcessImageResponseObject, error) {
	panic("unimplemented")
}

// UpdateImage implements api.StrictServerInterface.
func (h *imageHandler) UpdateImage(ctx context.Context, request api.UpdateImageRequestObject) (api.UpdateImageResponseObject, error) {
	panic("unimplemented")
}

func toAPI(img *model.Image) *api.Image {
	return &api.Image{
		Id:         img.Id,
		Alt:        img.Alt,
		OwnerType:  api.OwnerType(img.OwnerType),
		OwnerId:    img.OwnerId,
		Role:       api.ImageRole(img.Role),
		Key:        img.Key,
		Mime:       img.Mime,
		Size:       int(img.Size),
		Status:     api.ImageStatus(img.Status),
		CreatedAt:  img.CreatedAt,
		ModifiedAt: img.ModifiedAt,
	}
}

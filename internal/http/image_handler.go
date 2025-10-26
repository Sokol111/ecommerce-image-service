package http

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Sokol111/ecommerce-commons/pkg/observability"
	"github.com/Sokol111/ecommerce-image-service-api/api"
	"github.com/Sokol111/ecommerce-image-service/internal/application/command"
	"github.com/Sokol111/ecommerce-image-service/internal/application/query"
	"github.com/Sokol111/ecommerce-image-service/internal/domain/image"
)

type imageHandler struct {
	createPresignHandler  command.CreatePresignCommandHandler
	confirmUploadHandler  command.ConfirmUploadCommandHandler
	promoteImagesHandler  command.PromoteImagesCommandHandler
	deleteImageHandler    command.DeleteImageCommandHandler
	getImageByIDHandler   query.GetImageByIDQueryHandler
	getDeliveryURLHandler query.GetDeliveryURLQueryHandler
}

func newImageHandler(
	createPresign command.CreatePresignCommandHandler,
	confirmUpload command.ConfirmUploadCommandHandler,
	promoteImages command.PromoteImagesCommandHandler,
	deleteImage command.DeleteImageCommandHandler,
	getImageByID query.GetImageByIDQueryHandler,
	getDeliveryURL query.GetDeliveryURLQueryHandler,
) api.StrictServerInterface {
	return &imageHandler{
		createPresignHandler:  createPresign,
		confirmUploadHandler:  confirmUpload,
		promoteImagesHandler:  promoteImages,
		deleteImageHandler:    deleteImage,
		getImageByIDHandler:   getImageByID,
		getDeliveryURLHandler: getDeliveryURL,
	}
}

func (h *imageHandler) CreatePresign(ctx context.Context, request api.CreatePresignRequestObject) (api.CreatePresignResponseObject, error) {
	switch request.Body.OwnerType {
	case api.ProductDraft, api.Product:
		cmd := command.CreatePresignCommand{
			ContentType: string(request.Body.ContentType),
			Filename:    request.Body.Filename,
			OwnerType:   string(request.Body.OwnerType),
			OwnerID:     request.Body.OwnerId,
			Size:        int64(request.Body.Size),
		}

		result, err := h.createPresignHandler.Handle(ctx, cmd)
		if err != nil {
			return nil, fmt.Errorf("failed to create presign: %w", err)
		}

		return api.CreatePresign200JSONResponse{
			UploadUrl:       result.UploadURL,
			Key:             result.Key,
			ExpiresIn:       result.ExpiresIn,
			RequiredHeaders: result.RequiredHeaders,
		}, nil

	case api.User:
		return nil, fmt.Errorf("unsupported ownerType: %s", request.Body.OwnerType)

	default:
		return nil, fmt.Errorf("unsupported ownerType: %s", request.Body.OwnerType)
	}
}

func (h *imageHandler) ConfirmUpload(ctx context.Context, request api.ConfirmUploadRequestObject) (api.ConfirmUploadResponseObject, error) {
	switch request.Body.OwnerType {
	case api.ProductDraft:
		cmd := command.ConfirmUploadCommand{
			Alt:       request.Body.Alt,
			Key:       request.Body.Key,
			Mime:      string(request.Body.Mime),
			Role:      string(request.Body.Role),
			OwnerType: string(request.Body.OwnerType),
			OwnerID:   request.Body.OwnerId,
			Checksum:  request.Body.Checksum,
		}

		img, err := h.confirmUploadHandler.Handle(ctx, cmd)
		if err != nil {
			return nil, fmt.Errorf("failed to confirm upload: %w", err)
		}

		return api.ConfirmUpload201JSONResponse{
			Id:         img.ID,
			Alt:        img.Alt,
			OwnerType:  api.OwnerType(img.OwnerType),
			OwnerId:    img.OwnerID,
			Role:       api.ImageRole(img.Role),
			Key:        img.Key,
			Mime:       img.Mime,
			Size:       int(img.Size),
			Status:     api.ImageStatus(img.Status),
			CreatedAt:  img.CreatedAt,
			ModifiedAt: img.ModifiedAt,
		}, nil

	case api.Product:
		return nil, fmt.Errorf("unsupported ownerType: %s", request.Body.OwnerType)

	case api.User:
		return nil, fmt.Errorf("unsupported ownerType: %s", request.Body.OwnerType)

	default:
		return nil, fmt.Errorf("unsupported ownerType: %s", request.Body.OwnerType)
	}
}

func (h *imageHandler) PromoteImages(ctx context.Context, request api.PromoteImagesRequestObject) (api.PromoteImagesResponseObject, error) {
	cmd := command.PromoteImagesCommand{
		DraftID:   request.Body.DraftId,
		ImageIDs:  request.Body.Images,
		ProductID: request.Body.ProductId,
	}

	images, err := h.promoteImagesHandler.Handle(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to promote images: %w", err)
	}

	promoted := make([]api.Image, 0, len(images))
	for _, img := range images {
		promoted = append(promoted, *toAPI(img))
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

	q := query.GetDeliveryURLQuery{
		ImageID: request.Id,
		Width:   request.Params.W,
		Height:  request.Params.H,
		Fit:     fit,
		Quality: request.Params.Quality,
		DPR:     request.Params.Dpr,
		Format:  format,
		Expires: expires,
	}

	result, err := h.getDeliveryURLHandler.Handle(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("failed to get delivery URL: %w", err)
	}

	response := api.GetDeliveryUrl200JSONResponse{
		Url:       &result.URL,
		ExpiresAt: result.ExpiresAt,
	}
	return response, nil
}

func (h *imageHandler) DeleteImage(ctx context.Context, request api.DeleteImageRequestObject) (api.DeleteImageResponseObject, error) {
	hard := false
	if request.Params.Hard != nil {
		hard = *request.Params.Hard
	}

	cmd := command.DeleteImageCommand{
		ImageID: request.Id,
		Hard:    hard,
	}

	err := h.deleteImageHandler.Handle(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to delete image [%v]: %w", request.Id, err)
	}

	return api.DeleteImage204Response{}, nil
}

func (h *imageHandler) GetImage(ctx context.Context, request api.GetImageRequestObject) (api.GetImageResponseObject, error) {
	q := query.GetImageByIDQuery{
		ID: request.Id,
	}

	img, err := h.getImageByIDHandler.Handle(ctx, q)

	if err != nil {
		if errors.Is(err, image.ErrImageNotFound) {
			traceId := observability.GetTraceId(ctx)
			return api.GetImage404ApplicationProblemPlusJSONResponse(api.Problem{
				Title:   "Image not found",
				Status:  404,
				TraceId: &traceId,
			}), nil
		}
		return nil, fmt.Errorf("failed to get image by id: %w", err)
	}

	if img.IsDeleted() {
		traceId := observability.GetTraceId(ctx)
		return api.GetImage404ApplicationProblemPlusJSONResponse(
			api.Problem{
				Title:   "Image deleted",
				Status:  404,
				TraceId: &traceId,
			}), nil
	}

	return api.GetImage200JSONResponse(*toAPI(img)), nil
}

func (h *imageHandler) ListImages(ctx context.Context, request api.ListImagesRequestObject) (api.ListImagesResponseObject, error) {
	panic("unimplemented")
}

func (h *imageHandler) ProcessImage(ctx context.Context, request api.ProcessImageRequestObject) (api.ProcessImageResponseObject, error) {
	panic("unimplemented")
}

func (h *imageHandler) UpdateImage(ctx context.Context, request api.UpdateImageRequestObject) (api.UpdateImageResponseObject, error) {
	panic("unimplemented")
}

func toAPI(img *image.Image) *api.Image {
	return &api.Image{
		Id:         img.ID,
		Alt:        img.Alt,
		OwnerType:  api.OwnerType(img.OwnerType),
		OwnerId:    img.OwnerID,
		Role:       api.ImageRole(img.Role),
		Key:        img.Key,
		Mime:       img.Mime,
		Size:       int(img.Size),
		Status:     api.ImageStatus(img.Status),
		CreatedAt:  img.CreatedAt,
		ModifiedAt: img.ModifiedAt,
	}
}

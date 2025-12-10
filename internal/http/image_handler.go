package http

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Sokol111/ecommerce-commons/pkg/observability/tracing"
	"github.com/Sokol111/ecommerce-image-service-api/gen/httpapi"
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
) httpapi.StrictServerInterface {
	return &imageHandler{
		createPresignHandler:  createPresign,
		confirmUploadHandler:  confirmUpload,
		promoteImagesHandler:  promoteImages,
		deleteImageHandler:    deleteImage,
		getImageByIDHandler:   getImageByID,
		getDeliveryURLHandler: getDeliveryURL,
	}
}

func (h *imageHandler) CreatePresign(ctx context.Context, request httpapi.CreatePresignRequestObject) (httpapi.CreatePresignResponseObject, error) {
	switch request.Body.OwnerType {
	case httpapi.ProductDraft, httpapi.Product:
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

		return httpapi.CreatePresign200JSONResponse{
			UploadUrl:       result.UploadURL,
			UploadToken:     result.UploadToken,
			ExpiresIn:       result.ExpiresIn,
			RequiredHeaders: result.RequiredHeaders,
		}, nil

	case httpapi.User:
		return nil, fmt.Errorf("unsupported ownerType: %s", request.Body.OwnerType)

	default:
		return nil, fmt.Errorf("unsupported ownerType: %s", request.Body.OwnerType)
	}
}

func (h *imageHandler) ConfirmUpload(ctx context.Context, request httpapi.ConfirmUploadRequestObject) (httpapi.ConfirmUploadResponseObject, error) {
	cmd := command.ConfirmUploadCommand{
		UploadToken: request.Body.UploadToken,
		Alt:         request.Body.Alt,
		Role:        string(request.Body.Role),
		Checksum:    request.Body.Checksum,
	}

	img, err := h.confirmUploadHandler.Handle(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to confirm upload: %w", err)
	}

	return httpapi.ConfirmUpload201JSONResponse{
		Id:         img.ID,
		Alt:        img.Alt,
		OwnerType:  httpapi.OwnerType(img.OwnerType),
		OwnerId:    img.OwnerID,
		Role:       httpapi.ImageRole(img.Role),
		Key:        img.Key,
		Mime:       img.Mime,
		Size:       int(img.Size),
		Status:     httpapi.ImageStatus(img.Status),
		CreatedAt:  img.CreatedAt,
		ModifiedAt: img.ModifiedAt,
	}, nil
}

func (h *imageHandler) PromoteImages(ctx context.Context, request httpapi.PromoteImagesRequestObject) (httpapi.PromoteImagesResponseObject, error) {
	cmd := command.PromoteImagesCommand{
		DraftID:   request.Body.DraftId,
		ImageIDs:  request.Body.Images,
		ProductID: request.Body.ProductId,
	}

	images, err := h.promoteImagesHandler.Handle(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to promote images: %w", err)
	}

	promoted := make([]httpapi.Image, 0, len(images))
	for _, img := range images {
		promoted = append(promoted, *toAPI(img))
	}

	return httpapi.PromoteImages200JSONResponse{Promoted: &promoted}, nil
}

func (h *imageHandler) GetDeliveryUrl(ctx context.Context, request httpapi.GetDeliveryUrlRequestObject) (httpapi.GetDeliveryUrlResponseObject, error) {
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

	response := httpapi.GetDeliveryUrl200JSONResponse{
		Url:       result.URL,
		ExpiresAt: result.ExpiresAt,
	}
	return response, nil
}

func (h *imageHandler) DeleteImage(ctx context.Context, request httpapi.DeleteImageRequestObject) (httpapi.DeleteImageResponseObject, error) {
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

	return httpapi.DeleteImage204Response{}, nil
}

func (h *imageHandler) GetImage(ctx context.Context, request httpapi.GetImageRequestObject) (httpapi.GetImageResponseObject, error) {
	q := query.GetImageByIDQuery{
		ID: request.Id,
	}

	img, err := h.getImageByIDHandler.Handle(ctx, q)

	if err != nil {
		if errors.Is(err, image.ErrImageNotFound) {
			traceId := tracing.GetTraceID(ctx)
			return httpapi.GetImage404ApplicationProblemPlusJSONResponse(httpapi.Problem{
				Title:   "Image not found",
				Status:  404,
				TraceId: &traceId,
			}), nil
		}
		return nil, fmt.Errorf("failed to get image by id: %w", err)
	}

	if img.IsDeleted() {
		traceId := tracing.GetTraceID(ctx)
		return httpapi.GetImage404ApplicationProblemPlusJSONResponse(
			httpapi.Problem{
				Title:   "Image deleted",
				Status:  404,
				TraceId: &traceId,
			}), nil
	}

	return httpapi.GetImage200JSONResponse(*toAPI(img)), nil
}

func (h *imageHandler) ListImages(ctx context.Context, request httpapi.ListImagesRequestObject) (httpapi.ListImagesResponseObject, error) {
	panic("unimplemented")
}

func (h *imageHandler) GetDeliveryUrls(ctx context.Context, request httpapi.GetDeliveryUrlsRequestObject) (httpapi.GetDeliveryUrlsResponseObject, error) {
	var fit *string
	if request.Body.Fit != nil {
		f := string(*request.Body.Fit)
		fit = &f
	}
	var format *string
	if request.Body.Format != nil {
		f := string(*request.Body.Format)
		format = &f
	}

	urls := make([]httpapi.ImageUrl, 0, len(request.Body.ImageIds))
	notFound := make([]string, 0)

	for _, imageID := range request.Body.ImageIds {
		q := query.GetDeliveryURLQuery{
			ImageID: imageID,
			Width:   request.Body.W,
			Height:  request.Body.H,
			Fit:     fit,
			Quality: request.Body.Quality,
			Format:  format,
		}

		result, err := h.getDeliveryURLHandler.Handle(ctx, q)
		if err != nil {
			if errors.Is(err, image.ErrImageNotFound) {
				notFound = append(notFound, imageID)
				continue
			}
			return nil, fmt.Errorf("failed to get delivery URL for image [%s]: %w", imageID, err)
		}

		urls = append(urls, httpapi.ImageUrl{
			ImageId:   imageID,
			Url:       result.URL,
			ExpiresAt: result.ExpiresAt,
		})
	}

	response := httpapi.GetDeliveryUrls200JSONResponse{
		Urls: urls,
	}
	if len(notFound) > 0 {
		response.NotFound = &notFound
	}

	return response, nil
}

func (h *imageHandler) ProcessImage(ctx context.Context, request httpapi.ProcessImageRequestObject) (httpapi.ProcessImageResponseObject, error) {
	panic("unimplemented")
}

func (h *imageHandler) UpdateImage(ctx context.Context, request httpapi.UpdateImageRequestObject) (httpapi.UpdateImageResponseObject, error) {
	panic("unimplemented")
}

func toAPI(img *image.Image) *httpapi.Image {
	return &httpapi.Image{
		Id:         img.ID,
		Alt:        img.Alt,
		OwnerType:  httpapi.OwnerType(img.OwnerType),
		OwnerId:    img.OwnerID,
		Role:       httpapi.ImageRole(img.Role),
		Key:        img.Key,
		Mime:       img.Mime,
		Size:       int(img.Size),
		Status:     httpapi.ImageStatus(img.Status),
		CreatedAt:  img.CreatedAt,
		ModifiedAt: img.ModifiedAt,
	}
}

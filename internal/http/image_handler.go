package http

import (
	"context"
	"errors"
	"net/url"
	"time"

	"github.com/Sokol111/ecommerce-commons/pkg/observability/tracing"
	"github.com/Sokol111/ecommerce-image-service-api/gen/httpapi"
	"github.com/Sokol111/ecommerce-image-service/internal/application/command"
	"github.com/Sokol111/ecommerce-image-service/internal/application/query"
	"github.com/Sokol111/ecommerce-image-service/internal/domain/image"
	"github.com/samber/lo"
)

var aboutBlankURL, _ = url.Parse("about:blank")

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
) httpapi.Handler {
	return &imageHandler{
		createPresignHandler:  createPresign,
		confirmUploadHandler:  confirmUpload,
		promoteImagesHandler:  promoteImages,
		deleteImageHandler:    deleteImage,
		getImageByIDHandler:   getImageByID,
		getDeliveryURLHandler: getDeliveryURL,
	}
}

func (h *imageHandler) CreatePresign(ctx context.Context, req *httpapi.PresignRequest) (httpapi.CreatePresignRes, error) {
	switch req.OwnerType {
	case httpapi.OwnerTypeDraft, httpapi.OwnerTypeProduct:
		cmd := command.CreatePresignCommand{
			ContentType: string(req.ContentType),
			Filename:    req.Filename,
			OwnerType:   string(req.OwnerType),
			OwnerID:     req.OwnerId,
			Size:        int64(req.Size),
		}

		result, err := h.createPresignHandler.Handle(ctx, cmd)
		if err != nil {
			return &httpapi.CreatePresignInternalServerError{
				Type:    *aboutBlankURL,
				Title:   "Failed to create presign",
				Status:  500,
				TraceId: httpapi.NewOptString(tracing.GetTraceID(ctx)),
			}, nil
		}

		uploadURL, _ := url.Parse(result.UploadURL)
		return &httpapi.PresignResponse{
			UploadUrl:   *uploadURL,
			UploadToken: result.UploadToken,
			ExpiresIn:   result.ExpiresIn,
			FormData:    result.FormData,
		}, nil

	case httpapi.OwnerTypeUser:
		return &httpapi.CreatePresignBadRequest{
			Type:    *aboutBlankURL,
			Title:   "Unsupported owner type",
			Status:  400,
			TraceId: httpapi.NewOptString(tracing.GetTraceID(ctx)),
		}, nil

	default:
		return &httpapi.CreatePresignBadRequest{
			Type:    *aboutBlankURL,
			Title:   "Unsupported owner type",
			Status:  400,
			TraceId: httpapi.NewOptString(tracing.GetTraceID(ctx)),
		}, nil
	}
}

func (h *imageHandler) ConfirmUpload(ctx context.Context, req *httpapi.ConfirmRequest) (httpapi.ConfirmUploadRes, error) {
	var checksum *string
	if req.Checksum.IsSet() && !req.Checksum.IsNull() {
		checksum = &req.Checksum.Value
	}

	cmd := command.ConfirmUploadCommand{
		UploadToken: req.UploadToken,
		Alt:         req.Alt,
		Role:        string(req.Role),
		Checksum:    checksum,
	}

	img, err := h.confirmUploadHandler.Handle(ctx, cmd)
	if err != nil {
		return &httpapi.ConfirmUploadInternalServerError{
			Type:    *aboutBlankURL,
			Title:   "Failed to confirm upload",
			Status:  500,
			TraceId: httpapi.NewOptString(tracing.GetTraceID(ctx)),
		}, nil
	}

	return toAPI(img), nil
}

func (h *imageHandler) PromoteImages(ctx context.Context, req *httpapi.PromoteRequest) (httpapi.PromoteImagesRes, error) {
	var imageIDs *[]string
	if len(req.Images) > 0 {
		imageIDs = &req.Images
	}

	cmd := command.PromoteImagesCommand{
		DraftID:   req.DraftId,
		ImageIDs:  imageIDs,
		ProductID: req.ProductId,
	}

	images, err := h.promoteImagesHandler.Handle(ctx, cmd)
	if err != nil {
		return &httpapi.PromoteImagesInternalServerError{
			Type:    *aboutBlankURL,
			Title:   "Failed to promote images",
			Status:  500,
			TraceId: httpapi.NewOptString(tracing.GetTraceID(ctx)),
		}, nil
	}

	promoted := lo.Map(images, func(img *image.Image, _ int) httpapi.Image {
		return *toAPI(img)
	})

	return &httpapi.PromoteImagesOK{Promoted: promoted}, nil
}

func (h *imageHandler) GetDeliveryUrl(ctx context.Context, params httpapi.GetDeliveryUrlParams) (httpapi.GetDeliveryUrlRes, error) {
	var fit *string
	if params.Fit.IsSet() {
		f := string(params.Fit.Value)
		fit = &f
	}
	var format *string
	if params.Format.IsSet() {
		f := string(params.Format.Value)
		format = &f
	}
	var expires *time.Time
	if params.TtlSeconds.IsSet() {
		t := time.Now().Add(time.Duration(params.TtlSeconds.Value) * time.Second)
		expires = &t
	}

	var width *int
	if params.W.IsSet() {
		width = &params.W.Value
	}
	var height *int
	if params.H.IsSet() {
		height = &params.H.Value
	}
	var quality *int
	if params.Quality.IsSet() {
		quality = &params.Quality.Value
	}
	var dpr *float32
	if params.Dpr.IsSet() {
		v := float32(params.Dpr.Value)
		dpr = &v
	}

	q := query.GetDeliveryURLQuery{
		ImageID: params.ID,
		Width:   width,
		Height:  height,
		Fit:     fit,
		Quality: quality,
		DPR:     dpr,
		Format:  format,
		Expires: expires,
	}

	result, err := h.getDeliveryURLHandler.Handle(ctx, q)
	if err != nil {
		if errors.Is(err, image.ErrImageNotFound) {
			return &httpapi.GetDeliveryUrlNotFound{
				Type:    *aboutBlankURL,
				Title:   "Image not found",
				Status:  404,
				TraceId: httpapi.NewOptString(tracing.GetTraceID(ctx)),
			}, nil
		}
		return &httpapi.GetDeliveryUrlInternalServerError{
			Type:    *aboutBlankURL,
			Title:   "Failed to get delivery URL",
			Status:  500,
			TraceId: httpapi.NewOptString(tracing.GetTraceID(ctx)),
		}, nil
	}

	parsedURL, _ := url.Parse(result.URL)
	response := &httpapi.GetDeliveryUrlOK{
		URL: *parsedURL,
	}
	if result.ExpiresAt != nil {
		response.ExpiresAt = httpapi.NewOptDateTime(*result.ExpiresAt)
	}
	return response, nil
}

func (h *imageHandler) DeleteImage(ctx context.Context, params httpapi.DeleteImageParams) (httpapi.DeleteImageRes, error) {
	hard := params.Hard.Or(false)

	cmd := command.DeleteImageCommand{
		ImageID: params.ID,
		Hard:    hard,
	}

	err := h.deleteImageHandler.Handle(ctx, cmd)
	if err != nil {
		if errors.Is(err, image.ErrImageNotFound) {
			return &httpapi.DeleteImageNotFound{
				Type:    *aboutBlankURL,
				Title:   "Image not found",
				Status:  404,
				TraceId: httpapi.NewOptString(tracing.GetTraceID(ctx)),
			}, nil
		}
		return &httpapi.DeleteImageInternalServerError{
			Type:    *aboutBlankURL,
			Title:   "Failed to delete image",
			Status:  500,
			TraceId: httpapi.NewOptString(tracing.GetTraceID(ctx)),
		}, nil
	}

	return &httpapi.DeleteImageNoContent{}, nil
}

func (h *imageHandler) GetImage(ctx context.Context, params httpapi.GetImageParams) (httpapi.GetImageRes, error) {
	q := query.GetImageByIDQuery{
		ID: params.ID,
	}

	img, err := h.getImageByIDHandler.Handle(ctx, q)

	if err != nil {
		if errors.Is(err, image.ErrImageNotFound) {
			return &httpapi.GetImageNotFound{
				Type:    *aboutBlankURL,
				Title:   "Image not found",
				Status:  404,
				TraceId: httpapi.NewOptString(tracing.GetTraceID(ctx)),
			}, nil
		}
		return &httpapi.GetImageInternalServerError{
			Type:    *aboutBlankURL,
			Title:   "Failed to get image",
			Status:  500,
			TraceId: httpapi.NewOptString(tracing.GetTraceID(ctx)),
		}, nil
	}

	if img.IsDeleted() {
		return &httpapi.GetImageNotFound{
			Type:    *aboutBlankURL,
			Title:   "Image deleted",
			Status:  404,
			TraceId: httpapi.NewOptString(tracing.GetTraceID(ctx)),
		}, nil
	}

	return toAPI(img), nil
}

func (h *imageHandler) ListImages(ctx context.Context, params httpapi.ListImagesParams) (httpapi.ListImagesRes, error) {
	// TODO: implement
	return &httpapi.ListImagesOK{
		Items: []httpapi.Image{},
	}, nil
}

func (h *imageHandler) GetDeliveryUrls(ctx context.Context, req *httpapi.BatchUrlRequest) (httpapi.GetDeliveryUrlsRes, error) {
	var fit *string
	if req.Fit.IsSet() {
		f := string(req.Fit.Value)
		fit = &f
	}
	var format *string
	if req.Format.IsSet() {
		f := string(req.Format.Value)
		format = &f
	}

	var width *int
	if req.W.IsSet() {
		width = &req.W.Value
	}
	var height *int
	if req.H.IsSet() {
		height = &req.H.Value
	}
	var quality *int
	if req.Quality.IsSet() {
		quality = &req.Quality.Value
	}

	urls := make([]httpapi.ImageUrl, 0, len(req.ImageIds))
	notFound := make([]string, 0)

	for _, imageID := range req.ImageIds {
		q := query.GetDeliveryURLQuery{
			ImageID: imageID,
			Width:   width,
			Height:  height,
			Fit:     fit,
			Quality: quality,
			Format:  format,
		}

		result, err := h.getDeliveryURLHandler.Handle(ctx, q)
		if err != nil {
			if errors.Is(err, image.ErrImageNotFound) {
				notFound = append(notFound, imageID)
				continue
			}
			return &httpapi.GetDeliveryUrlsInternalServerError{
				Type:    *aboutBlankURL,
				Title:   "Failed to get delivery URLs",
				Status:  500,
				TraceId: httpapi.NewOptString(tracing.GetTraceID(ctx)),
			}, nil
		}

		parsedURL, _ := url.Parse(result.URL)
		imgUrl := httpapi.ImageUrl{
			ImageId: imageID,
			URL:     *parsedURL,
		}
		if result.ExpiresAt != nil {
			imgUrl.ExpiresAt = httpapi.NewOptDateTime(*result.ExpiresAt)
		}
		urls = append(urls, imgUrl)
	}

	return &httpapi.BatchUrlResponse{
		Urls:     urls,
		NotFound: notFound,
	}, nil
}

func (h *imageHandler) ProcessImage(ctx context.Context, req *httpapi.ProcessImageReq) (httpapi.ProcessImageRes, error) {
	// TODO: implement
	return &httpapi.Problem{
		Type:   *aboutBlankURL,
		Title:  "Not implemented",
		Status: 501,
	}, nil
}

func (h *imageHandler) UpdateImage(ctx context.Context, req *httpapi.ImagePatch, params httpapi.UpdateImageParams) (httpapi.UpdateImageRes, error) {
	// TODO: implement
	return &httpapi.UpdateImageInternalServerError{
		Type:   *aboutBlankURL,
		Title:  "Not implemented",
		Status: 501,
	}, nil
}

func toAPI(img *image.Image) *httpapi.Image {
	return &httpapi.Image{
		ID:         img.ID,
		Alt:        img.Alt,
		OwnerType:  httpapi.OwnerType(img.OwnerType),
		OwnerId:    img.OwnerID,
		Role:       httpapi.ImageRole(img.Role),
		Key:        img.Key,
		Mime:       img.Mime,
		Size:       int(img.Size),
		Status:     httpapi.ImageStatus(img.Status),
		Variants:   []httpapi.Variant{},
		CreatedAt:  img.CreatedAt,
		ModifiedAt: img.ModifiedAt,
	}
}

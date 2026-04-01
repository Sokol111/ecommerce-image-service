package http //nolint:revive // package name intentional

import (
	"context"
	"errors"
	"net/url"
	"time"

	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
	"github.com/Sokol111/ecommerce-commons/pkg/observability/tracing"
	"github.com/Sokol111/ecommerce-image-service-api/gen/httpapi"
	"github.com/Sokol111/ecommerce-image-service/internal/apperrors"
	"github.com/Sokol111/ecommerce-image-service/internal/application/command"
	"github.com/Sokol111/ecommerce-image-service/internal/application/query"
	"github.com/Sokol111/ecommerce-image-service/internal/domain/image"
	imageevent "github.com/Sokol111/ecommerce-image-service/internal/event"
	"github.com/samber/lo"
)

var aboutBlankURL, _ = url.Parse("about:blank") //nolint:errcheck // static URL always valid

// deliveryURLParams holds extracted optional parameters for delivery URL generation
type deliveryURLParams struct {
	Width   *int
	Height  *int
	Fit     *string
	Quality *int
	DPR     *float32
	Format  *string
	Expires *time.Time
}

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
			Role:        string(req.Role),
			Size:        int64(req.Size),
		}

		result, err := h.createPresignHandler.Handle(ctx, cmd)
		if err != nil {
			// Map errors to appropriate HTTP status codes
			switch {
			case errors.Is(err, image.ErrUnsupportedMimeType),
				errors.Is(err, image.ErrInvalidSize),
				errors.Is(err, image.ErrImageTooLarge):
				return &httpapi.CreatePresignBadRequest{
					Type:    *aboutBlankURL,
					Title:   "Invalid request",
					Status:  400,
					Detail:  httpapi.NewOptString(err.Error()),
					TraceId: httpapi.NewOptString(tracing.GetTraceID(ctx)),
				}, nil
			default:
				return nil, err
			}
		}

		uploadURL, _ := url.Parse(result.UploadURL) //nolint:errcheck // URL from service always valid
		return &httpapi.PresignResponse{
			UploadUrl:   *uploadURL,
			UploadToken: result.UploadToken,
			ExpiresIn:   result.ExpiresIn,
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
		// Map errors to appropriate HTTP status codes
		switch {
		case errors.Is(err, apperrors.ErrInvalidUploadToken),
			errors.Is(err, image.ErrImageTooLarge),
			errors.Is(err, image.ErrInvalidSize):
			return &httpapi.ConfirmUploadBadRequest{
				Type:    *aboutBlankURL,
				Title:   "Invalid request",
				Status:  400,
				Detail:  httpapi.NewOptString(err.Error()),
				TraceId: httpapi.NewOptString(tracing.GetTraceID(ctx)),
			}, nil
		case errors.Is(err, apperrors.ErrObjectNotFound):
			return &httpapi.ConfirmUploadNotFound{
				Type:    *aboutBlankURL,
				Title:   "Object not found",
				Status:  404,
				Detail:  httpapi.NewOptString(err.Error()),
				TraceId: httpapi.NewOptString(tracing.GetTraceID(ctx)),
			}, nil
		default:
			return nil, err
		}
	}

	return toAPI(img), nil
}

func (h *imageHandler) PromoteImages(ctx context.Context, req *httpapi.PromoteRequest) (httpapi.PromoteImagesRes, error) {
	var imageIDs *[]string
	if len(req.Images) > 0 {
		imageIDs = &req.Images
	}

	var draftID string
	if req.DraftId.IsSet() {
		draftID = req.DraftId.Value
	}

	cmd := command.PromoteImagesCommand{
		DraftID:   draftID,
		ImageIDs:  imageIDs,
		OwnerType: image.OwnerTypeProduct,
		OwnerID:   req.ProductId,
		OnPromoted: func(ctx context.Context, ownerID string, images []command.PromotedImage) ([]outbox.Message, error) {
			msgs := make([]outbox.Message, 0, len(images))
			for _, img := range images {
				msgs = append(msgs, imageevent.NewProductImagePromotedOutboxMessage(ctx, ownerID, img.ImageID, img.SmallImageURL, img.LargeImageURL))
			}
			return msgs, nil
		},
	}

	images, err := h.promoteImagesHandler.Handle(ctx, cmd)
	if err != nil {
		// Map errors to appropriate HTTP status codes
		switch {
		case errors.Is(err, apperrors.ErrInvalidImageOwner):
			return &httpapi.PromoteImagesBadRequest{
				Type:    *aboutBlankURL,
				Title:   "Invalid request",
				Status:  400,
				Detail:  httpapi.NewOptString(err.Error()),
				TraceId: httpapi.NewOptString(tracing.GetTraceID(ctx)),
			}, nil
		case errors.Is(err, apperrors.ErrDraftNotFound),
			errors.Is(err, image.ErrImageNotFound):
			return &httpapi.PromoteImagesNotFound{
				Type:    *aboutBlankURL,
				Title:   "Not found",
				Status:  404,
				Detail:  httpapi.NewOptString(err.Error()),
				TraceId: httpapi.NewOptString(tracing.GetTraceID(ctx)),
			}, nil
		default:
			return nil, err
		}
	}

	promoted := lo.Map(images, func(img *image.Image, _ int) httpapi.Image {
		return *toAPI(img)
	})

	return &httpapi.PromoteImagesOK{Promoted: promoted}, nil
}

func (h *imageHandler) GetDeliveryUrl(ctx context.Context, params httpapi.GetDeliveryUrlParams) (httpapi.GetDeliveryUrlRes, error) { //nolint:revive // name from OpenAPI spec
	urlParams := extractDeliveryURLParams(params.W, params.H, params.Fit, params.Quality, params.Dpr, params.Format)

	if params.TtlSeconds.IsSet() {
		t := time.Now().UTC().Add(time.Duration(params.TtlSeconds.Value) * time.Second)
		urlParams.Expires = &t
	}

	q := query.GetDeliveryURLQuery{
		ImageID: params.ID,
		Width:   urlParams.Width,
		Height:  urlParams.Height,
		Fit:     urlParams.Fit,
		Quality: urlParams.Quality,
		DPR:     urlParams.DPR,
		Format:  urlParams.Format,
		Expires: urlParams.Expires,
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
		return nil, err
	}

	return buildDeliveryURLResponse(result), nil
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
		return nil, err
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
		return nil, err
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

// extractDeliveryURLParams extracts optional parameters for delivery URL generation
func extractDeliveryURLParams(w, h httpapi.OptInt, fit httpapi.OptGetDeliveryUrlFit, quality httpapi.OptInt, dpr httpapi.OptFloat64, format httpapi.OptGetDeliveryUrlFormat) deliveryURLParams {
	params := deliveryURLParams{}

	if w.IsSet() {
		params.Width = &w.Value
	}
	if h.IsSet() {
		params.Height = &h.Value
	}
	if fit.IsSet() {
		f := string(fit.Value)
		params.Fit = &f
	}
	if quality.IsSet() {
		params.Quality = &quality.Value
	}
	if dpr.IsSet() {
		v := float32(dpr.Value)
		params.DPR = &v
	}
	if format.IsSet() {
		f := string(format.Value)
		params.Format = &f
	}

	return params
}

// buildDeliveryURLResponse builds HTTP response for single delivery URL
func buildDeliveryURLResponse(result *query.GetDeliveryURLResult) *httpapi.GetDeliveryUrlOK {
	parsedURL, _ := url.Parse(result.URL) //nolint:errcheck // URL from service always valid
	response := &httpapi.GetDeliveryUrlOK{
		URL: *parsedURL,
	}
	if result.ExpiresAt != nil {
		response.ExpiresAt = httpapi.NewOptDateTime(*result.ExpiresAt)
	}
	return response
}

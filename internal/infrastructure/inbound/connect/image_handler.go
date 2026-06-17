package connect

import (
	"context"
	"errors"
	"time"

	"connectrpc.com/connect"
	imagev1 "github.com/Sokol111/ecommerce-image-service-api/gen/connect/image/v1"
	"github.com/Sokol111/ecommerce-image-service/internal/application/image"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type imageHandler struct {
	createPresignHandler  image.CreatePresignCommandHandler
	confirmUploadHandler  image.ConfirmUploadCommandHandler
	promoteImagesHandler  image.PromoteImagesCommandHandler
	deleteImageHandler    image.DeleteImageCommandHandler
	getImageByIDHandler   image.GetImageByIDQueryHandler
	getDeliveryURLHandler image.GetDeliveryURLQueryHandler
}

func (h *imageHandler) CreatePresign(ctx context.Context, req *connect.Request[imagev1.CreatePresignRequest]) (*connect.Response[imagev1.CreatePresignResponse], error) {
	ownerType := protoOwnerTypeToString(req.Msg.GetOwnerType())

	switch req.Msg.GetOwnerType() {
	case imagev1.OwnerType_OWNER_TYPE_DRAFT, imagev1.OwnerType_OWNER_TYPE_PRODUCT:
		cmd := image.CreatePresignCommand{
			ContentType: protoContentTypeToString(req.Msg.GetContentType()),
			Filename:    req.Msg.GetFilename(),
			OwnerType:   ownerType,
			OwnerID:     req.Msg.GetOwnerId(),
			Role:        protoImageRoleToString(req.Msg.GetRole()),
			Size:        req.Msg.GetSize(),
		}

		result, err := h.createPresignHandler.Handle(ctx, cmd)
		if err != nil {
			return nil, mapImageConnectError(err)
		}

		return connect.NewResponse(&imagev1.CreatePresignResponse{
			UploadUrl:   result.UploadURL,
			UploadToken: result.UploadToken,
			ExpiresIn:   int64(result.ExpiresIn),
		}), nil

	default:
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("unsupported owner type"))
	}
}

func (h *imageHandler) ConfirmUpload(ctx context.Context, req *connect.Request[imagev1.ConfirmUploadRequest]) (*connect.Response[imagev1.ConfirmUploadResponse], error) {
	cmd := image.ConfirmUploadCommand{
		UploadToken: req.Msg.GetUploadToken(),
		Alt:         req.Msg.GetAlt(),
		Role:        protoImageRoleToString(req.Msg.GetRole()),
		Checksum:    req.Msg.Checksum,
	}

	img, err := h.confirmUploadHandler.Handle(ctx, cmd)
	if err != nil {
		return nil, mapImageConnectError(err)
	}

	return connect.NewResponse(&imagev1.ConfirmUploadResponse{
		Image: toProtoImage(img),
	}), nil
}

func (h *imageHandler) GetImage(ctx context.Context, req *connect.Request[imagev1.GetImageRequest]) (*connect.Response[imagev1.GetImageResponse], error) {
	q := image.GetImageByIDQuery{ID: req.Msg.GetId()}

	img, err := h.getImageByIDHandler.Handle(ctx, q)
	if err != nil {
		return nil, mapImageConnectError(err)
	}

	if img.IsDeleted() {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("image deleted"))
	}

	return connect.NewResponse(&imagev1.GetImageResponse{
		Image: toProtoImage(img),
	}), nil
}

func (h *imageHandler) DeleteImage(ctx context.Context, req *connect.Request[imagev1.DeleteImageRequest]) (*connect.Response[imagev1.DeleteImageResponse], error) {
	hard := false
	if req.Msg.Hard != nil {
		hard = *req.Msg.Hard
	}

	cmd := image.DeleteImageCommand{
		ImageID: req.Msg.GetId(),
		Hard:    hard,
	}

	if err := h.deleteImageHandler.Handle(ctx, cmd); err != nil {
		return nil, mapImageConnectError(err)
	}

	return connect.NewResponse(&imagev1.DeleteImageResponse{}), nil
}

func (h *imageHandler) GetDeliveryUrl(ctx context.Context, req *connect.Request[imagev1.GetDeliveryUrlRequest]) (*connect.Response[imagev1.GetDeliveryUrlResponse], error) { //nolint:revive
	q := image.GetDeliveryURLQuery{
		ImageID: req.Msg.GetId(),
		Width:   optInt32ToIntPtr(req.Msg.W),
		Height:  optInt32ToIntPtr(req.Msg.H),
		Quality: optInt32ToIntPtr(req.Msg.Quality),
	}

	if req.Msg.Fit != nil {
		s := protoImageFitToString(*req.Msg.Fit)
		q.Fit = &s
	}
	if req.Msg.Format != nil {
		s := protoImageFormatToString(*req.Msg.Format)
		q.Format = &s
	}
	if req.Msg.Dpr != nil {
		v := float32(*req.Msg.Dpr)
		q.DPR = &v
	}
	if req.Msg.TtlSeconds != nil {
		t := time.Now().UTC().Add(time.Duration(*req.Msg.TtlSeconds) * time.Second)
		q.Expires = &t
	}

	result, err := h.getDeliveryURLHandler.Handle(ctx, q)
	if err != nil {
		return nil, mapImageConnectError(err)
	}

	resp := &imagev1.GetDeliveryUrlResponse{Url: result.URL}
	if result.ExpiresAt != nil {
		resp.ExpiresAt = timestamppb.New(*result.ExpiresAt)
	}
	return connect.NewResponse(resp), nil
}

func (h *imageHandler) PromoteImages(ctx context.Context, req *connect.Request[imagev1.PromoteImagesRequest]) (*connect.Response[imagev1.PromoteImagesResponse], error) {
	var imageIDs *[]string
	if len(req.Msg.GetImages()) > 0 {
		ids := req.Msg.GetImages()
		imageIDs = &ids
	}

	var draftID string
	if req.Msg.DraftId != nil {
		draftID = *req.Msg.DraftId
	}

	cmd := image.PromoteImagesCommand{
		DraftID:   draftID,
		ImageIDs:  imageIDs,
		OwnerType: image.OwnerTypeProduct,
		OwnerID:   req.Msg.GetProductId(),
	}

	images, err := h.promoteImagesHandler.Handle(ctx, cmd)
	if err != nil {
		return nil, mapImageConnectError(err)
	}

	promoted := make([]*imagev1.Image, len(images))
	for i, img := range images {
		promoted[i] = toProtoImage(img)
	}

	return connect.NewResponse(&imagev1.PromoteImagesResponse{
		Promoted: promoted,
	}), nil
}

// ==================== Helpers ====================

func toProtoImage(img *image.Image) *imagev1.Image {
	return &imagev1.Image{
		Id:         img.ID,
		OwnerType:  stringToProtoOwnerType(img.OwnerType),
		OwnerId:    img.OwnerID,
		Role:       stringToProtoImageRole(img.Role),
		Key:        img.Key,
		Alt:        img.Alt,
		Mime:       img.Mime,
		Size:       img.Size,
		Status:     stringToProtoImageStatus(string(img.Status)),
		Variants:   []*imagev1.ImageVariant{},
		CreatedAt:  timestamppb.New(img.CreatedAt),
		ModifiedAt: timestamppb.New(img.ModifiedAt),
	}
}

func protoOwnerTypeToString(t imagev1.OwnerType) string {
	switch t {
	case imagev1.OwnerType_OWNER_TYPE_DRAFT:
		return "draft"
	case imagev1.OwnerType_OWNER_TYPE_PRODUCT:
		return "product"
	case imagev1.OwnerType_OWNER_TYPE_USER:
		return "user"
	default:
		return ""
	}
}

func stringToProtoOwnerType(s string) imagev1.OwnerType {
	switch s {
	case "draft":
		return imagev1.OwnerType_OWNER_TYPE_DRAFT
	case "product":
		return imagev1.OwnerType_OWNER_TYPE_PRODUCT
	case "user":
		return imagev1.OwnerType_OWNER_TYPE_USER
	default:
		return imagev1.OwnerType_OWNER_TYPE_UNSPECIFIED
	}
}

func protoImageRoleToString(r imagev1.ImageRole) string {
	switch r {
	case imagev1.ImageRole_IMAGE_ROLE_MAIN:
		return "main"
	case imagev1.ImageRole_IMAGE_ROLE_GALLERY:
		return "gallery"
	case imagev1.ImageRole_IMAGE_ROLE_OTHER:
		return "other"
	default:
		return ""
	}
}

func stringToProtoImageRole(s string) imagev1.ImageRole {
	switch s {
	case "main":
		return imagev1.ImageRole_IMAGE_ROLE_MAIN
	case "gallery":
		return imagev1.ImageRole_IMAGE_ROLE_GALLERY
	case "other":
		return imagev1.ImageRole_IMAGE_ROLE_OTHER
	default:
		return imagev1.ImageRole_IMAGE_ROLE_UNSPECIFIED
	}
}

func stringToProtoImageStatus(s string) imagev1.ImageStatus {
	switch s {
	case "uploaded":
		return imagev1.ImageStatus_IMAGE_STATUS_UPLOADED
	case "ready":
		return imagev1.ImageStatus_IMAGE_STATUS_READY
	case "deleted":
		return imagev1.ImageStatus_IMAGE_STATUS_DELETED
	default:
		return imagev1.ImageStatus_IMAGE_STATUS_UNSPECIFIED
	}
}

func protoContentTypeToString(ct imagev1.ImageContentType) string {
	switch ct {
	case imagev1.ImageContentType_IMAGE_CONTENT_TYPE_JPEG:
		return "image/jpeg"
	case imagev1.ImageContentType_IMAGE_CONTENT_TYPE_PNG:
		return "image/png"
	case imagev1.ImageContentType_IMAGE_CONTENT_TYPE_WEBP:
		return "image/webp"
	case imagev1.ImageContentType_IMAGE_CONTENT_TYPE_AVIF:
		return "image/avif"
	default:
		return ""
	}
}

func protoImageFitToString(f imagev1.ImageFit) string {
	switch f {
	case imagev1.ImageFit_IMAGE_FIT_COVER:
		return "cover"
	case imagev1.ImageFit_IMAGE_FIT_CONTAIN:
		return "contain"
	case imagev1.ImageFit_IMAGE_FIT_FILL:
		return "fill"
	case imagev1.ImageFit_IMAGE_FIT_INSIDE:
		return "inside"
	case imagev1.ImageFit_IMAGE_FIT_OUTSIDE:
		return "outside"
	default:
		return ""
	}
}

func protoImageFormatToString(f imagev1.ImageFormat) string {
	switch f {
	case imagev1.ImageFormat_IMAGE_FORMAT_ORIGINAL:
		return "original"
	case imagev1.ImageFormat_IMAGE_FORMAT_WEBP:
		return "webp"
	case imagev1.ImageFormat_IMAGE_FORMAT_AVIF:
		return "avif"
	case imagev1.ImageFormat_IMAGE_FORMAT_JPEG:
		return "jpeg"
	case imagev1.ImageFormat_IMAGE_FORMAT_PNG:
		return "png"
	default:
		return ""
	}
}

func optInt32ToIntPtr(v *int32) *int {
	if v == nil {
		return nil
	}
	i := int(*v)
	return &i
}

func mapImageConnectError(err error) *connect.Error {
	switch {
	case errors.Is(err, image.ErrUnsupportedMimeType),
		errors.Is(err, image.ErrInvalidSize),
		errors.Is(err, image.ErrImageTooLarge),
		errors.Is(err, image.ErrInvalidUploadToken),
		errors.Is(err, image.ErrInvalidImageOwner):
		return connect.NewError(connect.CodeInvalidArgument, err)
	case errors.Is(err, image.ErrImageNotFound),
		errors.Is(err, image.ErrObjectNotFound),
		errors.Is(err, image.ErrDraftNotFound):
		return connect.NewError(connect.CodeNotFound, err)
	default:
		return connect.NewError(connect.CodeInternal, err)
	}
}

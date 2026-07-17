package connect

import (
	"net/http"

	"connectrpc.com/connect"
	"github.com/Sokol111/ecommerce-commons/pkg/security/validation"
	imagev1connect "github.com/Sokol111/ecommerce-image-service-api/gen/go/image/v1/imagev1connect"
	"github.com/Sokol111/ecommerce-image-service/internal/application/image"
	"go.uber.org/fx"
)

// Module provides the Connect gRPC/Connect-RPC server handler for image operations.
func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			newImageHandler,
			provideProcedurePermissions,
		),
		fx.Invoke(registerConnectRoutes),
	)
}

func newImageHandler(
	createPresign image.CreatePresignCommandHandler,
	confirmUpload image.ConfirmUploadCommandHandler,
	promoteImages image.PromoteImagesCommandHandler,
	deleteImage image.DeleteImageCommandHandler,
	getImageByID image.GetImageByIDQueryHandler,
	getDeliveryURL image.GetDeliveryURLQueryHandler,
) *imageHandler {
	return &imageHandler{
		createPresignHandler:  createPresign,
		confirmUploadHandler:  confirmUpload,
		promoteImagesHandler:  promoteImages,
		deleteImageHandler:    deleteImage,
		getImageByIDHandler:   getImageByID,
		getDeliveryURLHandler: getDeliveryURL,
	}
}

func registerConnectRoutes(
	mux *http.ServeMux,
	handler *imageHandler,
	interceptors []connect.Interceptor,
) {
	path, h := imagev1connect.NewImageServiceHandler(handler, connect.WithInterceptors(interceptors...))
	mux.Handle(path, h)
}

func provideProcedurePermissions() validation.ProcedurePermissions {
	return validation.ProcedurePermissions{
		imagev1connect.ImageServiceCreatePresignProcedure:  {"images:write"},
		imagev1connect.ImageServiceConfirmUploadProcedure:  {"images:write"},
		imagev1connect.ImageServicePromoteImagesProcedure:  {"images:write"},
		imagev1connect.ImageServiceDeleteImageProcedure:    {"images:delete"},
		imagev1connect.ImageServiceGetImageProcedure:       {"images:read"},
		imagev1connect.ImageServiceGetDeliveryUrlProcedure: {"images:read"},
	}
}

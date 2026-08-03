package fxconfig

import (
	"net/http"

	"connectrpc.com/connect"
	"github.com/Sokol111/ecommerce-commons/pkg/security/validation"
	imagev1connect "github.com/Sokol111/ecommerce-image-service-api/gen/go/image/v1/imagev1connect"
	internalconnect "github.com/Sokol111/ecommerce-image-service/internal/infrastructure/inbound/connect"
	"go.uber.org/fx"
)

// NewConnectModule provides the Connect gRPC/Connect-RPC server handler for image operations.
func NewConnectModule() fx.Option {
	return fx.Options(
		fx.Provide(
			internalconnect.NewImageHandler,
			provideProcedurePermissions,
		),
		fx.Invoke(registerConnectRoutes),
	)
}

func registerConnectRoutes(
	mux *http.ServeMux,
	handler *internalconnect.ImageHandler,
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

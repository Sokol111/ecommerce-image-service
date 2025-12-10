package http

import (
	"github.com/Sokol111/ecommerce-image-service-api/gen/httpapi"
	"go.uber.org/fx"
)

func NewHttpHandlerModule() fx.Option {
	return fx.Options(
		fx.Provide(
			newImageHandler,
			newStrictHandler,
			// Provide OpenAPI spec - gin module will auto-register validation middleware
			httpapi.GetSwagger,
		),
		fx.Invoke(httpapi.RegisterHandlers),
	)
}

func newStrictHandler(ssi httpapi.StrictServerInterface) httpapi.ServerInterface {
	return httpapi.NewStrictHandler(ssi, nil)
}

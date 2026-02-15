package http //nolint:revive // package name intentional

import (
	"net/http"

	"go.uber.org/fx"

	"github.com/Sokol111/ecommerce-image-service-api/gen/httpapi"
)

func NewHTTPHandlerModule() fx.Option {
	return fx.Options(
		fx.Provide(
			newImageHandler,
			httpapi.ProvideServer,
		),
		fx.Invoke(registerOgenRoutes),
	)
}

func registerOgenRoutes(mux *http.ServeMux, server *httpapi.Server) {
	mux.Handle("/", server)
}

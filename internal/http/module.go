package http //nolint:revive // package name intentional

import (
	"net/http"

	"github.com/ogen-go/ogen/middleware"
	"github.com/ogen-go/ogen/ogenerrors"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"

	"github.com/Sokol111/ecommerce-image-service-api/gen/httpapi"
)

func NewHTTPHandlerModule() fx.Option {
	return fx.Options(
		fx.Provide(
			newImageHandler,
			newOgenServer,
		),
		fx.Invoke(registerOgenRoutes),
	)
}

func newOgenServer(
	handler httpapi.Handler,
	tracerProvider trace.TracerProvider,
	meterProvider metric.MeterProvider,
	middlewares []middleware.Middleware,
	errorHandler ogenerrors.ErrorHandler,
) (*httpapi.Server, error) {
	return httpapi.NewServer(
		handler,
		httpapi.WithTracerProvider(tracerProvider),
		httpapi.WithMeterProvider(meterProvider),
		httpapi.WithErrorHandler(errorHandler),
		httpapi.WithMiddleware(middlewares...),
	)
}

func registerOgenRoutes(mux *http.ServeMux, server *httpapi.Server) {
	mux.Handle("/", server)
}

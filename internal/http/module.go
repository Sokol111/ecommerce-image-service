package http

import (
	"fmt"

	"github.com/Sokol111/ecommerce-image-service-api/api"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewHttpHandlerModule() fx.Option {
	return fx.Options(
		fx.Provide(
			newImageHandler,
			func(ssi api.StrictServerInterface) api.ServerInterface {
				return api.NewStrictHandler(ssi, nil)
			},
			// Provide OpenAPI spec - gin module will auto-register validation middleware
			newOpenAPISpec,
		),
		fx.Invoke(registerRoutes),
	)
}

func newOpenAPISpec(log *zap.Logger) (*openapi3.T, error) {
	swagger, err := api.GetSwagger()
	if err != nil {
		return nil, fmt.Errorf("failed to load OpenAPI spec: %w", err)
	}
	return swagger, nil
}

func registerRoutes(engine *gin.Engine, serverInterface api.ServerInterface) {
	api.RegisterHandlers(engine, serverInterface)
}

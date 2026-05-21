package http //nolint:revive // package name intentional

import (
	"go.uber.org/fx"
)

func NewHTTPHandlerModule() fx.Option {
	return fx.Options(
		fx.Provide(
			newImageHandler,
		),
	)
}

package messaging

import (
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Provide(
		newProductCreatedHandler,
	)
}

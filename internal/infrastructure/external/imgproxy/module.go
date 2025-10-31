package imgproxy

import (
	"go.uber.org/fx"
)

func NewImgProxyModule() fx.Option {
	return fx.Provide(
		newConfig,
		fx.Annotate(
			newImgproxySigner,
			fx.ParamTags(``, `name:"s3Bucket"`),
		),
	)
}

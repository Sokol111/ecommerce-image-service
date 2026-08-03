package main

import (
	fx_application "github.com/Sokol111/ecommerce-image-service/internal/application/fxconfig"
	fx_infrastructure "github.com/Sokol111/ecommerce-image-service/internal/infrastructure/fxconfig"

	fx_commons "github.com/Sokol111/ecommerce-commons/pkg/fxconfig"

	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		fx_commons.NewCommonsModule(),
		fx_application.NewAppModule(),
		fx_infrastructure.NewInfrastructureModule(),
	)
	app.Run()
}

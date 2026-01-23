package main

import (
	"codeworkers/internal/api"
	"codeworkers/internal/providers"
	"codeworkers/internal/runner"

	"go.uber.org/fx"
)

func main() {
	fx.New(
		fx.Provide(
			providers.NewConfig,
			providers.NewQueue,
			providers.NewStrategyRegistry,
			providers.NewContainerPool,
			providers.NewExecutor,
			providers.NewRunner,
			providers.NewAPI,
		),
		fx.Invoke(func(*runner.Runner, *api.API) {
			// Dependencies are injected to trigger lifecycle hooks
		}),
	).Run()
}

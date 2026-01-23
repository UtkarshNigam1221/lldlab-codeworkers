package providers

import (
	"context"
	"log"
	"net/http"

	"codeworkers/internal/api"
	"codeworkers/internal/executor"
	"codeworkers/internal/queue"
	"codeworkers/internal/runner"

	"go.uber.org/fx"
)

const (
	// Log messages.
	LogStartingWorkerPool = "Starting worker pool..."
	LogStoppingWorkerPool = "Stopping worker pool..."
	LogStartingAPIServer  = "Starting API server on %s"
	LogAPIServerFailed    = "API server failed: %v"
	LogAPIServerStopped   = "API server stopped"
)

type Config struct {
	Port          string
	QueueCapacity int
	WorkerCount   int
}

func NewConfig() *Config {
	return &Config{
		Port:          ":8080",
		QueueCapacity: 100,
		WorkerCount:   5,
	}
}

func NewQueue(cfg *Config) *queue.Queue {
	return queue.NewQueue(cfg.QueueCapacity)
}

func NewStrategyRegistry() *executor.StrategyRegistry {
	return executor.NewStrategyRegistry()
}

func NewExecutor(registry *executor.StrategyRegistry) *executor.DockerExecutor {
	return executor.NewDockerExecutor(registry)
}

func NewRunner(
	q *queue.Queue,
	exec *executor.DockerExecutor,
	cfg *Config,
	lc fx.Lifecycle,
) *runner.Runner {
	r := runner.NewRunner(q, exec, cfg.WorkerCount)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Println(LogStartingWorkerPool)
			go r.Start()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Println(LogStoppingWorkerPool)
			r.Stop()
			return nil
		},
	})

	return r
}

func NewAPI(
	q *queue.Queue,
	r *runner.Runner,
	cfg *Config,
	lc fx.Lifecycle,
) *api.API {
	apiServer := api.NewAPI(q, r)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				log.Printf(LogStartingAPIServer, cfg.Port)
				if err := http.ListenAndServe(cfg.Port, apiServer.SetupRoutes()); err != nil {
					log.Fatalf(LogAPIServerFailed, err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Println(LogAPIServerStopped)
			return nil
		},
	})

	return apiServer
}

package providers

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"

	"codeworkers/internal/api"
	"codeworkers/internal/executor"
	"codeworkers/internal/queue"
	"codeworkers/internal/runner"

	"go.uber.org/fx"
)

const (
	// Log messages.
	LogStartingWorkerPool    = "Starting worker pool..."
	LogStoppingWorkerPool    = "Stopping worker pool..."
	LogStartingAPIServer     = "Starting API server on %s"
	LogAPIServerFailed       = "API server failed: %v"
	LogAPIServerStopped      = "API server stopped"
	LogInitializingPool      = "Initializing container pool..."
	LogPoolInitialized       = "Container pool initialized"
	LogCleaningUpPool        = "Cleaning up container pool..."
	DefaultContainerPoolSize = 10
)

type Config struct {
	Port              string
	QueueCapacity     int
	WorkerCount       int
	ContainerPoolSize int
}

func NewConfig() *Config {
	workerCount := 5
	if val := os.Getenv("WORKERS"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			workerCount = parsed
		}
	}

	queueSize := 100
	if val := os.Getenv("QUEUE_SIZE"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			queueSize = parsed
		}
	}

	poolSize := DefaultContainerPoolSize
	if val := os.Getenv("CONTAINER_POOL_SIZE"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			poolSize = parsed
		}
	}

	return &Config{
		Port:              ":8080",
		QueueCapacity:     queueSize,
		WorkerCount:       workerCount,
		ContainerPoolSize: poolSize,
	}
}

func NewQueue(cfg *Config) *queue.Queue {
	return queue.NewQueue(cfg.QueueCapacity)
}

func NewStrategyRegistry() *executor.StrategyRegistry {
	return executor.NewStrategyRegistry()
}

func NewContainerPool(cfg *Config, lc fx.Lifecycle) *executor.ContainerPool {
	pool := executor.NewContainerPool(cfg.ContainerPoolSize)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Println(LogInitializingPool)
			if err := pool.Initialize(); err != nil {
				return err
			}
			log.Println(LogPoolInitialized)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Println(LogCleaningUpPool)
			pool.Cleanup()
			return nil
		},
	})

	return pool
}

func NewExecutor(registry *executor.StrategyRegistry, pool executor.PoolInterface) *executor.DockerExecutor {
	return executor.NewDockerExecutor(registry, pool)
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

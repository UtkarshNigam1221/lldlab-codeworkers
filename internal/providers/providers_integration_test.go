package providers

import (
	"context"
	"testing"
	"time"

	"codeworkers/internal/api"
	"codeworkers/internal/executor"
	"codeworkers/internal/runner"

	"go.uber.org/fx"
)

func TestNewRunnerLifecycle(t *testing.T) {
	t.Skip("Skipping Fx lifecycle test - causes container name conflicts in parallel execution")
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	var r *runner.Runner
	app := fx.New(
		fx.Provide(
			NewConfig,
			NewQueue,
			NewStrategyRegistry,
			NewContainerPool,
			NewExecutor,
			NewRunner,
		),
		fx.Populate(&r),
		fx.NopLogger,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.Start(ctx); err != nil {
		t.Fatalf("Failed to start app: %v", err)
	}

	if r == nil {
		t.Fatal("Runner was not populated")
	}

	if err := app.Stop(ctx); err != nil {
		t.Fatalf("Failed to stop app: %v", err)
	}
}

func TestNewAPILifecycle(t *testing.T) {
	t.Skip("Skipping Fx lifecycle test - causes container name conflicts in parallel execution")
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	var apiServer *api.API
	app := fx.New(
		fx.Provide(
			NewConfig,
			NewQueue,
			NewStrategyRegistry,
			NewContainerPool,
			NewExecutor,
			NewRunner,
			NewAPI,
		),
		fx.Populate(&apiServer),
		fx.NopLogger,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.Start(ctx); err != nil {
		t.Fatalf("Failed to start app: %v", err)
	}

	if apiServer == nil {
		t.Fatal("API server was not populated")
	}

	if err := app.Stop(ctx); err != nil {
		t.Fatalf("Failed to stop app: %v", err)
	}
}

func TestProvidersIntegration(t *testing.T) {
	cfg := NewConfig()
	q := NewQueue(cfg)
	registry := NewStrategyRegistry()
	pool := executor.NewMockContainerPool()
	exec := NewExecutor(registry, pool)

	if cfg == nil {
		t.Fatal("Config is nil")
	}
	if q == nil {
		t.Fatal("Queue is nil")
	}
	if registry == nil {
		t.Fatal("Registry is nil")
	}
	if exec == nil {
		t.Fatal("Executor is nil")
	}
}

func TestConfigValues(t *testing.T) {
	cfg := NewConfig()

	tests := []struct {
		name     string
		actual   interface{}
		expected interface{}
	}{
		{"Port", cfg.Port, ":8080"},
		{"QueueCapacity", cfg.QueueCapacity, 100},
		{"WorkerCount", cfg.WorkerCount, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.actual != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, tt.actual)
			}
		})
	}
}

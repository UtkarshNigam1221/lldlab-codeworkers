package providers

import (
	"testing"

	"codeworkers/internal/executor"
)

func TestNewConfig(t *testing.T) {
	cfg := NewConfig()

	if cfg == nil {
		t.Fatal("NewConfig returned nil")
	}

	if cfg.Port != ":8080" {
		t.Errorf("Expected port :8080, got %s", cfg.Port)
	}

	if cfg.QueueCapacity != 100 {
		t.Errorf("Expected queue capacity 100, got %d", cfg.QueueCapacity)
	}

	if cfg.WorkerCount != 5 {
		t.Errorf("Expected worker count 5, got %d", cfg.WorkerCount)
	}
}

func TestLogConstants(t *testing.T) {
	constants := []struct {
		name  string
		value string
	}{
		{"LogStartingWorkerPool", LogStartingWorkerPool},
		{"LogStoppingWorkerPool", LogStoppingWorkerPool},
		{"LogStartingAPIServer", LogStartingAPIServer},
		{"LogAPIServerFailed", LogAPIServerFailed},
		{"LogAPIServerStopped", LogAPIServerStopped},
	}

	for _, c := range constants {
		t.Run(c.name, func(t *testing.T) {
			if c.value == "" {
				t.Errorf("Constant %s should not be empty", c.name)
			}
		})
	}
}

func TestNewQueue(t *testing.T) {
	cfg := NewConfig()
	q := NewQueue(cfg)

	if q == nil {
		t.Fatal("NewQueue returned nil")
	}
}

func TestNewStrategyRegistry(t *testing.T) {
	registry := NewStrategyRegistry()

	if registry == nil {
		t.Fatal("NewStrategyRegistry returned nil")
	}
}

func TestNewExecutor(t *testing.T) {
	registry := NewStrategyRegistry()
	pool := executor.NewMockContainerPool()
	exec := NewExecutor(registry, pool)

	if exec == nil {
		t.Fatal("NewExecutor returned nil")
	}
}

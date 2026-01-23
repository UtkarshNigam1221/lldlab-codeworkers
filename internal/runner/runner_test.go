package runner

import (
	"testing"
	"time"

	"codeworkers/internal/executor"
	"codeworkers/internal/models"
	"codeworkers/internal/queue"
)

func TestNewRunner(t *testing.T) {
	q := queue.NewQueue(10)
	registry := executor.NewStrategyRegistry()
	exec := executor.NewDockerExecutor(registry)
	workers := 5

	r := NewRunner(q, exec, workers)

	if r == nil {
		t.Fatal("NewRunner returned nil")
	}
	if r.workers != workers {
		t.Errorf("Expected %d workers, got %d", workers, r.workers)
	}
	if r.queue == nil {
		t.Error("Queue not set")
	}
	if r.executor == nil {
		t.Error("Executor not set")
	}
	if r.stop == nil {
		t.Error("Stop channel not initialized")
	}
}

func TestRunnerStartStop(t *testing.T) {
	q := queue.NewQueue(10)
	registry := executor.NewStrategyRegistry()
	exec := executor.NewDockerExecutor(registry)
	r := NewRunner(q, exec, 2)

	r.Start()
	time.Sleep(100 * time.Millisecond)
	r.Stop()

	time.Sleep(100 * time.Millisecond)
}

func TestConstants(t *testing.T) {
	if StatusPassed != "PASSED" {
		t.Errorf("Expected StatusPassed to be PASSED, got %s", StatusPassed)
	}
	if StatusFailed != "FAILED" {
		t.Errorf("Expected StatusFailed to be FAILED, got %s", StatusFailed)
	}
}

func TestLogConstants(t *testing.T) {
	constants := []string{
		LogWorkerStarted,
		LogWorkerStopped,
		LogWorkerProcessing,
		LogWorkerCompleted,
		LogRunningTestCase,
	}

	for _, c := range constants {
		if c == "" {
			t.Error("Log constant should not be empty")
		}
	}
}

func TestExecuteSync(t *testing.T) {
	q := queue.NewQueue(10)
	registry := executor.NewStrategyRegistry()
	exec := executor.NewDockerExecutor(registry)
	r := NewRunner(q, exec, 1)

	sub := &models.Submission{
		ID:          "test-1",
		Language:    "python",
		Code:        "print(input())",
		TestCases:   []models.TestCase{{Input: "hello", Expected: "hello"}},
		TimeLimit:   5,
		MemoryLimit: 256,
	}

	result := r.ExecuteSync(sub)

	if result == nil {
		t.Fatal("ExecuteSync returned nil")
	}
	if result.SubmissionID != sub.ID {
		t.Errorf("Expected submission ID %s, got %s", sub.ID, result.SubmissionID)
	}
	if result.Status == "" {
		t.Error("Result status should not be empty")
	}
}

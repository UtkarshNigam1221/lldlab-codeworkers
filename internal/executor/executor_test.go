package executor

import (
	"context"
	"strings"
	"testing"

	"codeworkers/internal/models"
)

func TestNewDockerExecutor(t *testing.T) {
	registry := NewStrategyRegistry()
	executor := NewDockerExecutor(registry)

	if executor == nil {
		t.Fatal("NewDockerExecutor returned nil")
	}
	if executor.registry == nil {
		t.Error("Registry not set")
	}
}

func TestExecutorConstants(t *testing.T) {
	if DockerImage != "code-sandbox" {
		t.Errorf("Expected DockerImage to be code-sandbox, got %s", DockerImage)
	}
	if DockerCPULimit != "1.0" {
		t.Errorf("Expected DockerCPULimit to be 1.0, got %s", DockerCPULimit)
	}
	if DockerPidsLimit != "50" {
		t.Errorf("Expected DockerPidsLimit to be 50, got %s", DockerPidsLimit)
	}
	if DockerNetwork != "none" {
		t.Errorf("Expected DockerNetwork to be none, got %s", DockerNetwork)
	}
}

func TestExecuteUnsupportedLanguage(t *testing.T) {
	registry := NewStrategyRegistry()
	executor := NewDockerExecutor(registry)

	submission := &models.Submission{
		ID:          "test-1",
		Language:    "rust",
		Code:        "fn main() {}",
		TimeLimit:   5,
		MemoryLimit: 256,
	}

	ctx := context.Background()
	_, err := executor.Execute(ctx, submission, "")

	if err == nil {
		t.Error("Expected error for unsupported language")
	}
	if !strings.Contains(err.Error(), "unsupported language") {
		t.Errorf("Expected unsupported language error, got: %v", err)
	}
}

func TestExecuteContextCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker test in short mode")
	}

	registry := NewStrategyRegistry()
	executor := NewDockerExecutor(registry)

	submission := &models.Submission{
		ID:          "test-timeout",
		Language:    "python",
		Code:        "import time; time.sleep(10)",
		TimeLimit:   1,
		MemoryLimit: 256,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := executor.Execute(ctx, submission, "")

	if err == nil {
		t.Error("Expected error for canceled context")
	}
}

func TestExecuteCommandBuilding(t *testing.T) {
	registry := NewStrategyRegistry()
	executor := NewDockerExecutor(registry)

	tests := []struct {
		language    string
		code        string
		expectError bool
	}{
		{"python", "print('test')", false},
		{"javascript", "console.log('test')", false},
		{"cpp", "int main() { return 0; }", false},
		{"java", "public class Solution {}", false},
		{"typescript", "console.log('test')", false},
		{"go", "package main", false},
		{"invalid", "code", true},
	}

	for _, tt := range tests {
		t.Run(tt.language, func(t *testing.T) {
			submission := &models.Submission{
				Language:    tt.language,
				Code:        tt.code,
				TimeLimit:   5,
				MemoryLimit: 256,
			}

			strategy, err := executor.registry.GetStrategy(submission.Language)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error for invalid language")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			cmd := strategy.BuildRunCommand(submission.Code, "")
			if cmd == "" {
				t.Error("Expected non-empty command")
			}
		})
	}
}

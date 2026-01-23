package executor

import (
	"bytes"
	"codeworkers/internal/models"
	"context"
	"errors"
	"fmt"
	"os/exec"
)

const (
	DockerImage     = "code-sandbox"
	DockerCPULimit  = "1.0"
	DockerPidsLimit = "50"
	DockerNetwork   = "none"
)

type DockerExecutor struct {
	registry *StrategyRegistry
}

func NewDockerExecutor(registry *StrategyRegistry) *DockerExecutor {
	return &DockerExecutor{registry: registry}
}

func (e *DockerExecutor) Execute(ctx context.Context, sub *models.Submission, input string) (string, error) {
	strategy, err := e.registry.GetStrategy(sub.Language)
	if err != nil {
		return "", err
	}

	runCmd := strategy.BuildRunCommand(sub.Code, input)
	memLimit := fmt.Sprintf("%dm", sub.MemoryLimit)

	cmd := exec.CommandContext(ctx, "docker", "run", "--rm",
		"--network", DockerNetwork,
		"--memory", memLimit,
		"--memory-swap", memLimit,
		"--cpus", DockerCPULimit,
		"--pids-limit", DockerPidsLimit,
		DockerImage,
		"sh", "-c", runCmd,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return "", fmt.Errorf("time limit exceeded")
		}
		if stderr.Len() > 0 {
			return "", fmt.Errorf("runtime error: %s", stderr.String())
		}
		return "", err
	}

	return stdout.String(), nil
}

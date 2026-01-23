package executor

import (
	"bytes"
	"codeworkers/internal/models"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"
)

const (
	DockerImage     = "code-sandbox"
	DockerCPULimit  = "1.0"
	DockerPidsLimit = "50"
	DockerNetwork   = "none"
	AcquireTimeout  = 10 * time.Second
)

type DockerExecutor struct {
	registry *StrategyRegistry
	pool     PoolInterface
}

func NewDockerExecutor(registry *StrategyRegistry, pool PoolInterface) *DockerExecutor {
	return &DockerExecutor{
		registry: registry,
		pool:     pool,
	}
}

func (e *DockerExecutor) Execute(ctx context.Context, sub *models.Submission, input string) (string, error) {
	strategy, err := e.registry.GetStrategy(sub.Language)
	if err != nil {
		return "", err
	}

	runCmd := strategy.BuildRunCommand(sub.Code, input)

	acquireCtx, cancel := context.WithTimeout(ctx, AcquireTimeout)
	defer cancel()

	container, err := e.pool.Acquire(acquireCtx)
	if err != nil {
		return "", fmt.Errorf("failed to acquire container: %w", err)
	}
	defer e.pool.Release(container)

	cmd := exec.CommandContext(ctx, "docker", "exec", "-i",
		container.Name,
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

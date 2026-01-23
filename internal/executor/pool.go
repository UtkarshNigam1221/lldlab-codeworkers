package executor

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"sync"
	"time"
)

const (
	DefaultPoolSize      = 10
	ContainerIdleTimeout = 5 * time.Minute
	ContainerNamePrefix  = "code-sandbox-pool-"
	HealthCheckInterval  = 30 * time.Second
)

// PoolInterface defines the interface for container pools.
type PoolInterface interface {
	Initialize() error
	Acquire(ctx context.Context) (*Container, error)
	Release(container *Container)
	Cleanup()
	Stats() map[string]interface{}
}

type Container struct {
	ID        string
	Name      string
	CreatedAt time.Time
	LastUsed  time.Time
	InUse     bool
}

type ContainerPool struct {
	containers []*Container
	available  chan *Container
	mu         sync.RWMutex
	size       int
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewContainerPool(size int) *ContainerPool {
	if size <= 0 {
		size = DefaultPoolSize
	}

	ctx, cancel := context.WithCancel(context.Background())
	pool := &ContainerPool{
		containers: make([]*Container, 0, size),
		available:  make(chan *Container, size),
		size:       size,
		ctx:        ctx,
		cancel:     cancel,
	}

	return pool
}

func (p *ContainerPool) Initialize() error {
	log.Printf("Initializing container pool with %d containers...", p.size)

	for i := 0; i < p.size; i++ {
		container, err := p.createContainer(i)
		if err != nil {
			log.Printf("Failed to create container %d: %v", i, err)
			p.Cleanup()
			return fmt.Errorf("failed to initialize pool: %w", err)
		}

		p.mu.Lock()
		p.containers = append(p.containers, container)
		p.mu.Unlock()

		p.available <- container
	}

	log.Printf("Container pool initialized with %d containers", len(p.containers))

	go p.healthCheckLoop()

	return nil
}

func (p *ContainerPool) createContainer(index int) (*Container, error) {
	name := fmt.Sprintf("%s%d", ContainerNamePrefix, index)

	cmd := exec.Command("docker", "run", "-d",
		"--name", name,
		"--network", DockerNetwork,
		"--memory", "256m",
		"--memory-swap", "256m",
		"--cpus", DockerCPULimit,
		"--pids-limit", DockerPidsLimit,
		"--rm",
		DockerImage,
		"sleep", "infinity",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w, output: %s", err, string(output))
	}

	containerID := string(output[:12])

	container := &Container{
		ID:        containerID,
		Name:      name,
		CreatedAt: time.Now(),
		LastUsed:  time.Now(),
		InUse:     false,
	}

	return container, nil
}

func (p *ContainerPool) Acquire(ctx context.Context) (*Container, error) {
	select {
	case container := <-p.available:
		p.mu.Lock()
		container.InUse = true
		container.LastUsed = time.Now()
		p.mu.Unlock()
		return container, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("timeout acquiring container: %w", ctx.Err())
	case <-p.ctx.Done():
		return nil, fmt.Errorf("pool is shutting down")
	}
}

func (p *ContainerPool) Release(container *Container) {
	if container == nil {
		return
	}

	p.mu.Lock()
	container.InUse = false
	container.LastUsed = time.Now()
	p.mu.Unlock()

	select {
	case p.available <- container:
	default:
		log.Printf("Warning: container %s could not be returned to pool", container.Name)
	}
}

func (p *ContainerPool) healthCheckLoop() {
	ticker := time.NewTicker(HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.performHealthCheck()
		case <-p.ctx.Done():
			return
		}
	}
}

func (p *ContainerPool) performHealthCheck() {
	p.mu.RLock()
	containers := make([]*Container, len(p.containers))
	copy(containers, p.containers)
	p.mu.RUnlock()

	for _, container := range containers {
		if container.InUse {
			continue
		}

		cmd := exec.Command("docker", "inspect", "--format={{.State.Running}}", container.Name)
		output, err := cmd.Output()
		if err != nil || string(output) != "true\n" {
			log.Printf("Container %s is unhealthy, recreating...", container.Name)
			p.recreateContainer(container)
		}
	}
}

func (p *ContainerPool) recreateContainer(old *Container) {
	// Best effort cleanup - container may already be removed
	exec.Command("docker", "rm", "-f", old.Name).Run() //nolint:errcheck

	index := 0
	p.mu.RLock()
	for i, c := range p.containers {
		if c.Name == old.Name {
			index = i
			break
		}
	}
	p.mu.RUnlock()

	newContainer, err := p.createContainer(index)
	if err != nil {
		log.Printf("Failed to recreate container: %v", err)
		return
	}

	p.mu.Lock()
	for i, c := range p.containers {
		if c.Name == old.Name {
			p.containers[i] = newContainer
			break
		}
	}
	p.mu.Unlock()

	p.available <- newContainer
}

func (p *ContainerPool) Cleanup() {
	log.Println("Cleaning up container pool...")
	p.cancel()

	p.mu.RLock()
	containers := make([]*Container, len(p.containers))
	copy(containers, p.containers)
	p.mu.RUnlock()

	for _, container := range containers {
		cmd := exec.Command("docker", "rm", "-f", container.Name)
		if err := cmd.Run(); err != nil {
			log.Printf("Failed to remove container %s: %v", container.Name, err)
		}
	}

	log.Println("Container pool cleanup complete")
}

func (p *ContainerPool) Stats() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	inUse := 0
	for _, c := range p.containers {
		if c.InUse {
			inUse++
		}
	}

	return map[string]interface{}{
		"total":     len(p.containers),
		"in_use":    inUse,
		"available": len(p.available),
	}
}

package executor

import (
	"context"
)

// MockContainerPool is a mock implementation for testing.
type MockContainerPool struct {
	container *Container
}

func NewMockContainerPool() *MockContainerPool {
	return &MockContainerPool{
		container: &Container{
			ID:   "mock-container",
			Name: "mock-container",
		},
	}
}

func (m *MockContainerPool) Initialize() error {
	return nil
}

func (m *MockContainerPool) Acquire(ctx context.Context) (*Container, error) {
	return m.container, nil
}

func (m *MockContainerPool) Release(container *Container) {
	// No-op for mock
}

func (m *MockContainerPool) Cleanup() {
	// No-op for mock
}

func (m *MockContainerPool) Stats() map[string]interface{} {
	return map[string]interface{}{
		"total":     1,
		"in_use":    0,
		"available": 1,
	}
}

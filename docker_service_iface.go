package truenas

import "context"

// DockerServiceAPI defines the interface for Docker operations.
type DockerServiceAPI interface {
	GetStatus(ctx context.Context) (*DockerStatus, error)
	GetConfig(ctx context.Context) (*DockerConfig, error)
}

// Compile-time checks.
var _ DockerServiceAPI = (*DockerService)(nil)
var _ DockerServiceAPI = (*MockDockerService)(nil)

// MockDockerService is a test double for DockerServiceAPI.
type MockDockerService struct {
	GetStatusFunc func(ctx context.Context) (*DockerStatus, error)
	GetConfigFunc func(ctx context.Context) (*DockerConfig, error)
}

func (m *MockDockerService) GetStatus(ctx context.Context) (*DockerStatus, error) {
	if m.GetStatusFunc != nil {
		return m.GetStatusFunc(ctx)
	}
	return nil, nil
}

func (m *MockDockerService) GetConfig(ctx context.Context) (*DockerConfig, error) {
	if m.GetConfigFunc != nil {
		return m.GetConfigFunc(ctx)
	}
	return nil, nil
}

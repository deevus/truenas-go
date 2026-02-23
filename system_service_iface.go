package truenas

import "context"

// SystemServiceAPI defines the interface for system operations.
type SystemServiceAPI interface {
	GetInfo(ctx context.Context) (*SystemInfo, error)
	GetVersion(ctx context.Context) (string, error)
}

// Compile-time checks.
var _ SystemServiceAPI = (*SystemService)(nil)
var _ SystemServiceAPI = (*MockSystemService)(nil)

// MockSystemService is a test double for SystemServiceAPI.
type MockSystemService struct {
	GetInfoFunc    func(ctx context.Context) (*SystemInfo, error)
	GetVersionFunc func(ctx context.Context) (string, error)
}

func (m *MockSystemService) GetInfo(ctx context.Context) (*SystemInfo, error) {
	if m.GetInfoFunc != nil {
		return m.GetInfoFunc(ctx)
	}
	return nil, nil
}

func (m *MockSystemService) GetVersion(ctx context.Context) (string, error) {
	if m.GetVersionFunc != nil {
		return m.GetVersionFunc(ctx)
	}
	return "", nil
}

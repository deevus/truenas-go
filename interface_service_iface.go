package truenas

import "context"

// InterfaceServiceAPI defines the interface for network interface operations.
type InterfaceServiceAPI interface {
	List(ctx context.Context) ([]NetworkInterface, error)
	Get(ctx context.Context, id string) (*NetworkInterface, error)
}

// Compile-time checks.
var _ InterfaceServiceAPI = (*InterfaceService)(nil)
var _ InterfaceServiceAPI = (*MockInterfaceService)(nil)

// MockInterfaceService is a test double for InterfaceServiceAPI.
type MockInterfaceService struct {
	ListFunc func(ctx context.Context) ([]NetworkInterface, error)
	GetFunc  func(ctx context.Context, id string) (*NetworkInterface, error)
}

func (m *MockInterfaceService) List(ctx context.Context) ([]NetworkInterface, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx)
	}
	return nil, nil
}

func (m *MockInterfaceService) Get(ctx context.Context, id string) (*NetworkInterface, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, id)
	}
	return nil, nil
}

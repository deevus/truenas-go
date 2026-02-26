package truenas

import "context"

// NetworkServiceAPI defines the interface for network operations.
type NetworkServiceAPI interface {
	GetSummary(ctx context.Context) (*NetworkSummary, error)
}

// Compile-time checks.
var _ NetworkServiceAPI = (*NetworkService)(nil)
var _ NetworkServiceAPI = (*MockNetworkService)(nil)

// MockNetworkService is a test double for NetworkServiceAPI.
type MockNetworkService struct {
	GetSummaryFunc func(ctx context.Context) (*NetworkSummary, error)
}

func (m *MockNetworkService) GetSummary(ctx context.Context) (*NetworkSummary, error) {
	if m.GetSummaryFunc != nil {
		return m.GetSummaryFunc(ctx)
	}
	return nil, nil
}

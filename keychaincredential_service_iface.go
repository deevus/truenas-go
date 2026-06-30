package truenas

import "context"

// SSHServiceAPI defines the interface for cloud sync credential and task operations.
type SSHServiceAPI interface {
	CreateSSHKeyPair(ctx context.Context, opts CreateSSHKeyPairOpts) (*SSHKeyPair, error)
	GetSSHKeyPair(ctx context.Context, id int64) (*SSHKeyPair, error)
	ListSSHKeyPairs(ctx context.Context) ([]SSHKeyPair, error)
	UpdateSSHKeyPair(ctx context.Context, id int64, opts UpdateSSHKeyPairOpts) (*SSHKeyPair, error)
	DeleteSSHKeyPair(ctx context.Context, id int64) error

	CreateSSHConnection(ctx context.Context, opts CreateSSHConnectionOpts) (*SSHConnection, error)
	GetSSHConnection(ctx context.Context, id int64) (*SSHConnection, error)
	ListSSHConnections(ctx context.Context) ([]SSHConnection, error)
	UpdateSSHConnection(ctx context.Context, id int64, opts UpdateSSHConnectionOpts) (*SSHConnection, error)
	DeleteSSHConnection(ctx context.Context, id int64) error
}

// Compile-time checks.
var _ SSHServiceAPI = (*SSHService)(nil)
var _ SSHServiceAPI = (*MockSSHService)(nil)

// MockSSHService is a test double for SSHServiceAPI.
type MockSSHService struct {
	CreateSSHKeyPairFunc func(ctx context.Context, opts CreateSSHKeyPairOpts) (*SSHKeyPair, error)
	GetSSHKeyPairFunc    func(ctx context.Context, id int64) (*SSHKeyPair, error)
	ListSSHKeyPairsFunc  func(ctx context.Context) ([]SSHKeyPair, error)
	UpdateSSHKeyPairFunc func(ctx context.Context, id int64, opts UpdateSSHKeyPairOpts) (*SSHKeyPair, error)
	DeleteSSHKeyPairFunc func(ctx context.Context, id int64) error

	CreateSSHConnectionFunc func(ctx context.Context, opts CreateSSHConnectionOpts) (*SSHConnection, error)
	GetSSHConnectionFunc    func(ctx context.Context, id int64) (*SSHConnection, error)
	ListSSHConnectionsFunc  func(ctx context.Context) ([]SSHConnection, error)
	UpdateSSHConnectionFunc func(ctx context.Context, id int64, opts UpdateSSHConnectionOpts) (*SSHConnection, error)
	DeleteSSHConnectionFunc func(ctx context.Context, id int64) error
}

func (m *MockSSHService) CreateSSHKeyPair(ctx context.Context, opts CreateSSHKeyPairOpts) (*SSHKeyPair, error) {
	if m.CreateSSHKeyPairFunc != nil {
		return m.CreateSSHKeyPairFunc(ctx, opts)
	}
	return nil, nil
}

func (m *MockSSHService) GetSSHKeyPair(ctx context.Context, id int64) (*SSHKeyPair, error) {
	if m.GetSSHKeyPairFunc != nil {
		return m.GetSSHKeyPairFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockSSHService) ListSSHKeyPairs(ctx context.Context) ([]SSHKeyPair, error) {
	if m.ListSSHKeyPairsFunc != nil {
		return m.ListSSHKeyPairsFunc(ctx)
	}
	return nil, nil
}

func (m *MockSSHService) UpdateSSHKeyPair(ctx context.Context, id int64, opts UpdateSSHKeyPairOpts) (*SSHKeyPair, error) {
	if m.UpdateSSHKeyPairFunc != nil {
		return m.UpdateSSHKeyPairFunc(ctx, id, opts)
	}
	return nil, nil
}

func (m *MockSSHService) DeleteSSHKeyPair(ctx context.Context, id int64) error {
	if m.DeleteSSHKeyPairFunc != nil {
		return m.DeleteSSHKeyPairFunc(ctx, id)
	}
	return nil
}

func (m *MockSSHService) CreateSSHConnection(ctx context.Context, opts CreateSSHConnectionOpts) (*SSHConnection, error) {
	if m.CreateSSHConnectionFunc != nil {
		return m.CreateSSHConnectionFunc(ctx, opts)
	}
	return nil, nil
}

func (m *MockSSHService) GetSSHConnection(ctx context.Context, id int64) (*SSHConnection, error) {
	if m.GetSSHConnectionFunc != nil {
		return m.GetSSHConnectionFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockSSHService) ListSSHConnections(ctx context.Context) ([]SSHConnection, error) {
	if m.ListSSHConnectionsFunc != nil {
		return m.ListSSHConnectionsFunc(ctx)
	}
	return nil, nil
}

func (m *MockSSHService) UpdateSSHConnection(ctx context.Context, id int64, opts UpdateSSHConnectionOpts) (*SSHConnection, error) {
	if m.UpdateSSHConnectionFunc != nil {
		return m.UpdateSSHConnectionFunc(ctx, id, opts)
	}
	return nil, nil
}

func (m *MockSSHService) DeleteSSHConnection(ctx context.Context, id int64) error {
	if m.DeleteSSHConnectionFunc != nil {
		return m.DeleteSSHConnectionFunc(ctx, id)
	}
	return nil
}

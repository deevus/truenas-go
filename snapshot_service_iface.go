package truenas

import "context"

// SnapshotServiceAPI defines the interface for snapshot operations.
type SnapshotServiceAPI interface {
	Create(ctx context.Context, opts CreateSnapshotOpts) (*Snapshot, error)
	Get(ctx context.Context, id string) (*Snapshot, error)
	List(ctx context.Context) ([]Snapshot, error)
	Delete(ctx context.Context, id string) error
	Hold(ctx context.Context, id string) error
	Release(ctx context.Context, id string) error
	Query(ctx context.Context, filters [][]any) ([]Snapshot, error)
	Rollback(ctx context.Context, id string) error
	Clone(ctx context.Context, snapshot, datasetDst string) error
}

// Compile-time checks.
var _ SnapshotServiceAPI = (*SnapshotService)(nil)
var _ SnapshotServiceAPI = (*MockSnapshotService)(nil)

// MockSnapshotService is a test double for SnapshotServiceAPI.
type MockSnapshotService struct {
	CreateFunc  func(ctx context.Context, opts CreateSnapshotOpts) (*Snapshot, error)
	GetFunc     func(ctx context.Context, id string) (*Snapshot, error)
	ListFunc    func(ctx context.Context) ([]Snapshot, error)
	DeleteFunc  func(ctx context.Context, id string) error
	HoldFunc    func(ctx context.Context, id string) error
	ReleaseFunc func(ctx context.Context, id string) error
	QueryFunc    func(ctx context.Context, filters [][]any) ([]Snapshot, error)
	RollbackFunc func(ctx context.Context, id string) error
	CloneFunc    func(ctx context.Context, snapshot, datasetDst string) error
}

func (m *MockSnapshotService) Create(ctx context.Context, opts CreateSnapshotOpts) (*Snapshot, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, opts)
	}
	return nil, nil
}

func (m *MockSnapshotService) Get(ctx context.Context, id string) (*Snapshot, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockSnapshotService) List(ctx context.Context) ([]Snapshot, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx)
	}
	return nil, nil
}

func (m *MockSnapshotService) Delete(ctx context.Context, id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockSnapshotService) Hold(ctx context.Context, id string) error {
	if m.HoldFunc != nil {
		return m.HoldFunc(ctx, id)
	}
	return nil
}

func (m *MockSnapshotService) Release(ctx context.Context, id string) error {
	if m.ReleaseFunc != nil {
		return m.ReleaseFunc(ctx, id)
	}
	return nil
}

func (m *MockSnapshotService) Query(ctx context.Context, filters [][]any) ([]Snapshot, error) {
	if m.QueryFunc != nil {
		return m.QueryFunc(ctx, filters)
	}
	return nil, nil
}

func (m *MockSnapshotService) Rollback(ctx context.Context, id string) error {
	if m.RollbackFunc != nil {
		return m.RollbackFunc(ctx, id)
	}
	return nil
}

func (m *MockSnapshotService) Clone(ctx context.Context, snapshot, datasetDst string) error {
	if m.CloneFunc != nil {
		return m.CloneFunc(ctx, snapshot, datasetDst)
	}
	return nil
}

package truenas

import "context"

// DatasetServiceAPI defines the interface for dataset, zvol, and pool operations.
type DatasetServiceAPI interface {
	CreateDataset(ctx context.Context, opts CreateDatasetOpts) (*Dataset, error)
	GetDataset(ctx context.Context, id string) (*Dataset, error)
	ListDatasets(ctx context.Context) ([]Dataset, error)
	UpdateDataset(ctx context.Context, id string, opts UpdateDatasetOpts) (*Dataset, error)
	DeleteDataset(ctx context.Context, id string, recursive bool) error
	CreateZvol(ctx context.Context, opts CreateZvolOpts) (*Zvol, error)
	GetZvol(ctx context.Context, id string) (*Zvol, error)
	UpdateZvol(ctx context.Context, id string, opts UpdateZvolOpts) (*Zvol, error)
	DeleteZvol(ctx context.Context, id string) error
	ListPools(ctx context.Context) ([]Pool, error)
}

// Compile-time checks.
var _ DatasetServiceAPI = (*DatasetService)(nil)
var _ DatasetServiceAPI = (*MockDatasetService)(nil)

// MockDatasetService is a test double for DatasetServiceAPI.
type MockDatasetService struct {
	CreateDatasetFunc func(ctx context.Context, opts CreateDatasetOpts) (*Dataset, error)
	GetDatasetFunc    func(ctx context.Context, id string) (*Dataset, error)
	ListDatasetsFunc  func(ctx context.Context) ([]Dataset, error)
	UpdateDatasetFunc func(ctx context.Context, id string, opts UpdateDatasetOpts) (*Dataset, error)
	DeleteDatasetFunc func(ctx context.Context, id string, recursive bool) error
	CreateZvolFunc    func(ctx context.Context, opts CreateZvolOpts) (*Zvol, error)
	GetZvolFunc       func(ctx context.Context, id string) (*Zvol, error)
	UpdateZvolFunc    func(ctx context.Context, id string, opts UpdateZvolOpts) (*Zvol, error)
	DeleteZvolFunc    func(ctx context.Context, id string) error
	ListPoolsFunc     func(ctx context.Context) ([]Pool, error)
}

func (m *MockDatasetService) CreateDataset(ctx context.Context, opts CreateDatasetOpts) (*Dataset, error) {
	if m.CreateDatasetFunc != nil {
		return m.CreateDatasetFunc(ctx, opts)
	}
	return nil, nil
}

func (m *MockDatasetService) GetDataset(ctx context.Context, id string) (*Dataset, error) {
	if m.GetDatasetFunc != nil {
		return m.GetDatasetFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockDatasetService) ListDatasets(ctx context.Context) ([]Dataset, error) {
	if m.ListDatasetsFunc != nil {
		return m.ListDatasetsFunc(ctx)
	}
	return nil, nil
}

func (m *MockDatasetService) UpdateDataset(ctx context.Context, id string, opts UpdateDatasetOpts) (*Dataset, error) {
	if m.UpdateDatasetFunc != nil {
		return m.UpdateDatasetFunc(ctx, id, opts)
	}
	return nil, nil
}

func (m *MockDatasetService) DeleteDataset(ctx context.Context, id string, recursive bool) error {
	if m.DeleteDatasetFunc != nil {
		return m.DeleteDatasetFunc(ctx, id, recursive)
	}
	return nil
}

func (m *MockDatasetService) CreateZvol(ctx context.Context, opts CreateZvolOpts) (*Zvol, error) {
	if m.CreateZvolFunc != nil {
		return m.CreateZvolFunc(ctx, opts)
	}
	return nil, nil
}

func (m *MockDatasetService) GetZvol(ctx context.Context, id string) (*Zvol, error) {
	if m.GetZvolFunc != nil {
		return m.GetZvolFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockDatasetService) UpdateZvol(ctx context.Context, id string, opts UpdateZvolOpts) (*Zvol, error) {
	if m.UpdateZvolFunc != nil {
		return m.UpdateZvolFunc(ctx, id, opts)
	}
	return nil, nil
}

func (m *MockDatasetService) DeleteZvol(ctx context.Context, id string) error {
	if m.DeleteZvolFunc != nil {
		return m.DeleteZvolFunc(ctx, id)
	}
	return nil
}

func (m *MockDatasetService) ListPools(ctx context.Context) ([]Pool, error) {
	if m.ListPoolsFunc != nil {
		return m.ListPoolsFunc(ctx)
	}
	return nil, nil
}

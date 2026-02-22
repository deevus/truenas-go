package truenas

import "context"

// CloudSyncServiceAPI defines the interface for cloud sync credential and task operations.
type CloudSyncServiceAPI interface {
	CreateCredential(ctx context.Context, opts CreateCredentialOpts) (*CloudSyncCredential, error)
	GetCredential(ctx context.Context, id int64) (*CloudSyncCredential, error)
	ListCredentials(ctx context.Context) ([]CloudSyncCredential, error)
	UpdateCredential(ctx context.Context, id int64, opts UpdateCredentialOpts) (*CloudSyncCredential, error)
	DeleteCredential(ctx context.Context, id int64) error
	CreateTask(ctx context.Context, opts CreateCloudSyncTaskOpts) (*CloudSyncTask, error)
	GetTask(ctx context.Context, id int64) (*CloudSyncTask, error)
	ListTasks(ctx context.Context) ([]CloudSyncTask, error)
	UpdateTask(ctx context.Context, id int64, opts UpdateCloudSyncTaskOpts) (*CloudSyncTask, error)
	DeleteTask(ctx context.Context, id int64) error
	Sync(ctx context.Context, id int64) error
}

// Compile-time checks.
var _ CloudSyncServiceAPI = (*CloudSyncService)(nil)
var _ CloudSyncServiceAPI = (*MockCloudSyncService)(nil)

// MockCloudSyncService is a test double for CloudSyncServiceAPI.
type MockCloudSyncService struct {
	CreateCredentialFunc func(ctx context.Context, opts CreateCredentialOpts) (*CloudSyncCredential, error)
	GetCredentialFunc    func(ctx context.Context, id int64) (*CloudSyncCredential, error)
	ListCredentialsFunc  func(ctx context.Context) ([]CloudSyncCredential, error)
	UpdateCredentialFunc func(ctx context.Context, id int64, opts UpdateCredentialOpts) (*CloudSyncCredential, error)
	DeleteCredentialFunc func(ctx context.Context, id int64) error
	CreateTaskFunc       func(ctx context.Context, opts CreateCloudSyncTaskOpts) (*CloudSyncTask, error)
	GetTaskFunc          func(ctx context.Context, id int64) (*CloudSyncTask, error)
	ListTasksFunc        func(ctx context.Context) ([]CloudSyncTask, error)
	UpdateTaskFunc       func(ctx context.Context, id int64, opts UpdateCloudSyncTaskOpts) (*CloudSyncTask, error)
	DeleteTaskFunc       func(ctx context.Context, id int64) error
	SyncFunc             func(ctx context.Context, id int64) error
}

func (m *MockCloudSyncService) CreateCredential(ctx context.Context, opts CreateCredentialOpts) (*CloudSyncCredential, error) {
	if m.CreateCredentialFunc != nil {
		return m.CreateCredentialFunc(ctx, opts)
	}
	return nil, nil
}

func (m *MockCloudSyncService) GetCredential(ctx context.Context, id int64) (*CloudSyncCredential, error) {
	if m.GetCredentialFunc != nil {
		return m.GetCredentialFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockCloudSyncService) ListCredentials(ctx context.Context) ([]CloudSyncCredential, error) {
	if m.ListCredentialsFunc != nil {
		return m.ListCredentialsFunc(ctx)
	}
	return nil, nil
}

func (m *MockCloudSyncService) UpdateCredential(ctx context.Context, id int64, opts UpdateCredentialOpts) (*CloudSyncCredential, error) {
	if m.UpdateCredentialFunc != nil {
		return m.UpdateCredentialFunc(ctx, id, opts)
	}
	return nil, nil
}

func (m *MockCloudSyncService) DeleteCredential(ctx context.Context, id int64) error {
	if m.DeleteCredentialFunc != nil {
		return m.DeleteCredentialFunc(ctx, id)
	}
	return nil
}

func (m *MockCloudSyncService) CreateTask(ctx context.Context, opts CreateCloudSyncTaskOpts) (*CloudSyncTask, error) {
	if m.CreateTaskFunc != nil {
		return m.CreateTaskFunc(ctx, opts)
	}
	return nil, nil
}

func (m *MockCloudSyncService) GetTask(ctx context.Context, id int64) (*CloudSyncTask, error) {
	if m.GetTaskFunc != nil {
		return m.GetTaskFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockCloudSyncService) ListTasks(ctx context.Context) ([]CloudSyncTask, error) {
	if m.ListTasksFunc != nil {
		return m.ListTasksFunc(ctx)
	}
	return nil, nil
}

func (m *MockCloudSyncService) UpdateTask(ctx context.Context, id int64, opts UpdateCloudSyncTaskOpts) (*CloudSyncTask, error) {
	if m.UpdateTaskFunc != nil {
		return m.UpdateTaskFunc(ctx, id, opts)
	}
	return nil, nil
}

func (m *MockCloudSyncService) DeleteTask(ctx context.Context, id int64) error {
	if m.DeleteTaskFunc != nil {
		return m.DeleteTaskFunc(ctx, id)
	}
	return nil
}

func (m *MockCloudSyncService) Sync(ctx context.Context, id int64) error {
	if m.SyncFunc != nil {
		return m.SyncFunc(ctx, id)
	}
	return nil
}

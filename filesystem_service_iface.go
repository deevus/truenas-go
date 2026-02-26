package truenas

import "context"

// FilesystemServiceAPI defines the interface for filesystem operations.
type FilesystemServiceAPI interface {
	Client() FileCaller
	WriteFile(ctx context.Context, path string, params WriteFileParams) error
	Stat(ctx context.Context, path string) (*StatResult, error)
	SetPermissions(ctx context.Context, opts SetPermOpts) error
}

// Compile-time checks.
var _ FilesystemServiceAPI = (*FilesystemService)(nil)
var _ FilesystemServiceAPI = (*MockFilesystemService)(nil)

// MockFilesystemService is a test double for FilesystemServiceAPI.
type MockFilesystemService struct {
	ClientFunc         func() FileCaller
	WriteFileFunc      func(ctx context.Context, path string, params WriteFileParams) error
	StatFunc           func(ctx context.Context, path string) (*StatResult, error)
	SetPermissionsFunc func(ctx context.Context, opts SetPermOpts) error
}

func (m *MockFilesystemService) Client() FileCaller {
	if m.ClientFunc != nil {
		return m.ClientFunc()
	}
	return nil
}

func (m *MockFilesystemService) WriteFile(ctx context.Context, path string, params WriteFileParams) error {
	if m.WriteFileFunc != nil {
		return m.WriteFileFunc(ctx, path, params)
	}
	return nil
}

func (m *MockFilesystemService) Stat(ctx context.Context, path string) (*StatResult, error) {
	if m.StatFunc != nil {
		return m.StatFunc(ctx, path)
	}
	return nil, nil
}

func (m *MockFilesystemService) SetPermissions(ctx context.Context, opts SetPermOpts) error {
	if m.SetPermissionsFunc != nil {
		return m.SetPermissionsFunc(ctx, opts)
	}
	return nil
}

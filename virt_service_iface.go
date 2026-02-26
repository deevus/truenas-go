package truenas

import "context"

// VirtServiceAPI defines the interface for virtualization instance and device operations.
type VirtServiceAPI interface {
	GetGlobalConfig(ctx context.Context) (*VirtGlobalConfig, error)
	UpdateGlobalConfig(ctx context.Context, opts UpdateVirtGlobalConfigOpts) (*VirtGlobalConfig, error)
	CreateInstance(ctx context.Context, opts CreateVirtInstanceOpts) (*VirtInstance, error)
	GetInstance(ctx context.Context, name string) (*VirtInstance, error)
	UpdateInstance(ctx context.Context, name string, opts UpdateVirtInstanceOpts) (*VirtInstance, error)
	DeleteInstance(ctx context.Context, name string) error
	StartInstance(ctx context.Context, name string) error
	StopInstance(ctx context.Context, name string, opts StopVirtInstanceOpts) error
	ListInstances(ctx context.Context, filters [][]any) ([]VirtInstance, error)
	ListDevices(ctx context.Context, instanceID string) ([]VirtDevice, error)
	AddDevice(ctx context.Context, instanceID string, opts VirtDeviceOpts) error
	DeleteDevice(ctx context.Context, instanceID string, deviceName string) error
}

// Compile-time checks.
var _ VirtServiceAPI = (*VirtService)(nil)
var _ VirtServiceAPI = (*MockVirtService)(nil)

// MockVirtService is a test double for VirtServiceAPI.
type MockVirtService struct {
	GetGlobalConfigFunc    func(ctx context.Context) (*VirtGlobalConfig, error)
	UpdateGlobalConfigFunc func(ctx context.Context, opts UpdateVirtGlobalConfigOpts) (*VirtGlobalConfig, error)
	CreateInstanceFunc     func(ctx context.Context, opts CreateVirtInstanceOpts) (*VirtInstance, error)
	GetInstanceFunc        func(ctx context.Context, name string) (*VirtInstance, error)
	UpdateInstanceFunc     func(ctx context.Context, name string, opts UpdateVirtInstanceOpts) (*VirtInstance, error)
	DeleteInstanceFunc     func(ctx context.Context, name string) error
	StartInstanceFunc      func(ctx context.Context, name string) error
	StopInstanceFunc       func(ctx context.Context, name string, opts StopVirtInstanceOpts) error
	ListInstancesFunc      func(ctx context.Context, filters [][]any) ([]VirtInstance, error)
	ListDevicesFunc        func(ctx context.Context, instanceID string) ([]VirtDevice, error)
	AddDeviceFunc          func(ctx context.Context, instanceID string, opts VirtDeviceOpts) error
	DeleteDeviceFunc       func(ctx context.Context, instanceID string, deviceName string) error
}

func (m *MockVirtService) GetGlobalConfig(ctx context.Context) (*VirtGlobalConfig, error) {
	if m.GetGlobalConfigFunc != nil {
		return m.GetGlobalConfigFunc(ctx)
	}
	return nil, nil
}

func (m *MockVirtService) UpdateGlobalConfig(ctx context.Context, opts UpdateVirtGlobalConfigOpts) (*VirtGlobalConfig, error) {
	if m.UpdateGlobalConfigFunc != nil {
		return m.UpdateGlobalConfigFunc(ctx, opts)
	}
	return nil, nil
}

func (m *MockVirtService) CreateInstance(ctx context.Context, opts CreateVirtInstanceOpts) (*VirtInstance, error) {
	if m.CreateInstanceFunc != nil {
		return m.CreateInstanceFunc(ctx, opts)
	}
	return nil, nil
}

func (m *MockVirtService) GetInstance(ctx context.Context, name string) (*VirtInstance, error) {
	if m.GetInstanceFunc != nil {
		return m.GetInstanceFunc(ctx, name)
	}
	return nil, nil
}

func (m *MockVirtService) UpdateInstance(ctx context.Context, name string, opts UpdateVirtInstanceOpts) (*VirtInstance, error) {
	if m.UpdateInstanceFunc != nil {
		return m.UpdateInstanceFunc(ctx, name, opts)
	}
	return nil, nil
}

func (m *MockVirtService) DeleteInstance(ctx context.Context, name string) error {
	if m.DeleteInstanceFunc != nil {
		return m.DeleteInstanceFunc(ctx, name)
	}
	return nil
}

func (m *MockVirtService) StartInstance(ctx context.Context, name string) error {
	if m.StartInstanceFunc != nil {
		return m.StartInstanceFunc(ctx, name)
	}
	return nil
}

func (m *MockVirtService) StopInstance(ctx context.Context, name string, opts StopVirtInstanceOpts) error {
	if m.StopInstanceFunc != nil {
		return m.StopInstanceFunc(ctx, name, opts)
	}
	return nil
}

func (m *MockVirtService) ListInstances(ctx context.Context, filters [][]any) ([]VirtInstance, error) {
	if m.ListInstancesFunc != nil {
		return m.ListInstancesFunc(ctx, filters)
	}
	return nil, nil
}

func (m *MockVirtService) ListDevices(ctx context.Context, instanceID string) ([]VirtDevice, error) {
	if m.ListDevicesFunc != nil {
		return m.ListDevicesFunc(ctx, instanceID)
	}
	return nil, nil
}

func (m *MockVirtService) AddDevice(ctx context.Context, instanceID string, opts VirtDeviceOpts) error {
	if m.AddDeviceFunc != nil {
		return m.AddDeviceFunc(ctx, instanceID, opts)
	}
	return nil
}

func (m *MockVirtService) DeleteDevice(ctx context.Context, instanceID string, deviceName string) error {
	if m.DeleteDeviceFunc != nil {
		return m.DeleteDeviceFunc(ctx, instanceID, deviceName)
	}
	return nil
}

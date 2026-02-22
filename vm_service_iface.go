package truenas

import "context"

// VMServiceAPI defines the interface for VM and VM device operations.
type VMServiceAPI interface {
	CreateVM(ctx context.Context, opts CreateVMOpts) (*VM, error)
	GetVM(ctx context.Context, id int64) (*VM, error)
	UpdateVM(ctx context.Context, id int64, opts UpdateVMOpts) (*VM, error)
	DeleteVM(ctx context.Context, id int64) error
	StartVM(ctx context.Context, id int64) error
	StopVM(ctx context.Context, id int64, opts StopVMOpts) error
	ListDevices(ctx context.Context, vmID int64) ([]VMDevice, error)
	GetDevice(ctx context.Context, id int64) (*VMDevice, error)
	CreateDevice(ctx context.Context, opts CreateVMDeviceOpts) (*VMDevice, error)
	UpdateDevice(ctx context.Context, id int64, opts UpdateVMDeviceOpts) (*VMDevice, error)
	DeleteDevice(ctx context.Context, id int64) error
}

// Compile-time checks.
var _ VMServiceAPI = (*VMService)(nil)
var _ VMServiceAPI = (*MockVMService)(nil)

// MockVMService is a test double for VMServiceAPI.
type MockVMService struct {
	CreateVMFunc     func(ctx context.Context, opts CreateVMOpts) (*VM, error)
	GetVMFunc        func(ctx context.Context, id int64) (*VM, error)
	UpdateVMFunc     func(ctx context.Context, id int64, opts UpdateVMOpts) (*VM, error)
	DeleteVMFunc     func(ctx context.Context, id int64) error
	StartVMFunc      func(ctx context.Context, id int64) error
	StopVMFunc       func(ctx context.Context, id int64, opts StopVMOpts) error
	ListDevicesFunc  func(ctx context.Context, vmID int64) ([]VMDevice, error)
	GetDeviceFunc    func(ctx context.Context, id int64) (*VMDevice, error)
	CreateDeviceFunc func(ctx context.Context, opts CreateVMDeviceOpts) (*VMDevice, error)
	UpdateDeviceFunc func(ctx context.Context, id int64, opts UpdateVMDeviceOpts) (*VMDevice, error)
	DeleteDeviceFunc func(ctx context.Context, id int64) error
}

func (m *MockVMService) CreateVM(ctx context.Context, opts CreateVMOpts) (*VM, error) {
	if m.CreateVMFunc != nil {
		return m.CreateVMFunc(ctx, opts)
	}
	return nil, nil
}

func (m *MockVMService) GetVM(ctx context.Context, id int64) (*VM, error) {
	if m.GetVMFunc != nil {
		return m.GetVMFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockVMService) UpdateVM(ctx context.Context, id int64, opts UpdateVMOpts) (*VM, error) {
	if m.UpdateVMFunc != nil {
		return m.UpdateVMFunc(ctx, id, opts)
	}
	return nil, nil
}

func (m *MockVMService) DeleteVM(ctx context.Context, id int64) error {
	if m.DeleteVMFunc != nil {
		return m.DeleteVMFunc(ctx, id)
	}
	return nil
}

func (m *MockVMService) StartVM(ctx context.Context, id int64) error {
	if m.StartVMFunc != nil {
		return m.StartVMFunc(ctx, id)
	}
	return nil
}

func (m *MockVMService) StopVM(ctx context.Context, id int64, opts StopVMOpts) error {
	if m.StopVMFunc != nil {
		return m.StopVMFunc(ctx, id, opts)
	}
	return nil
}

func (m *MockVMService) ListDevices(ctx context.Context, vmID int64) ([]VMDevice, error) {
	if m.ListDevicesFunc != nil {
		return m.ListDevicesFunc(ctx, vmID)
	}
	return nil, nil
}

func (m *MockVMService) GetDevice(ctx context.Context, id int64) (*VMDevice, error) {
	if m.GetDeviceFunc != nil {
		return m.GetDeviceFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockVMService) CreateDevice(ctx context.Context, opts CreateVMDeviceOpts) (*VMDevice, error) {
	if m.CreateDeviceFunc != nil {
		return m.CreateDeviceFunc(ctx, opts)
	}
	return nil, nil
}

func (m *MockVMService) UpdateDevice(ctx context.Context, id int64, opts UpdateVMDeviceOpts) (*VMDevice, error) {
	if m.UpdateDeviceFunc != nil {
		return m.UpdateDeviceFunc(ctx, id, opts)
	}
	return nil, nil
}

func (m *MockVMService) DeleteDevice(ctx context.Context, id int64) error {
	if m.DeleteDeviceFunc != nil {
		return m.DeleteDeviceFunc(ctx, id)
	}
	return nil
}

package truenas

import "context"

// AppServiceAPI defines the interface for app and registry operations.
type AppServiceAPI interface {
	CreateApp(ctx context.Context, opts CreateAppOpts) (*App, error)
	GetApp(ctx context.Context, name string) (*App, error)
	GetAppWithConfig(ctx context.Context, name string) (*App, error)
	UpdateApp(ctx context.Context, name string, opts UpdateAppOpts) (*App, error)
	ListApps(ctx context.Context) ([]App, error)
	StartApp(ctx context.Context, name string) error
	StopApp(ctx context.Context, name string) error
	DeleteApp(ctx context.Context, name string) error
	UpgradeSummary(ctx context.Context, name string) (*AppUpgradeSummary, error)
	ListImages(ctx context.Context) ([]AppImage, error)
	AvailableSpace(ctx context.Context) (int64, error)
	UpgradeApp(ctx context.Context, name string) error
	RedeployApp(ctx context.Context, name string) error
	CreateRegistry(ctx context.Context, opts CreateRegistryOpts) (*Registry, error)
	GetRegistry(ctx context.Context, id int64) (*Registry, error)
	ListRegistries(ctx context.Context) ([]Registry, error)
	UpdateRegistry(ctx context.Context, id int64, opts UpdateRegistryOpts) (*Registry, error)
	DeleteRegistry(ctx context.Context, id int64) error
	SubscribeStats(ctx context.Context) (*Subscription[[]AppStats], error)
	SubscribeContainerLogs(ctx context.Context, opts ContainerLogOpts) (*Subscription[AppContainerLogEntry], error)
}

// Compile-time checks.
var _ AppServiceAPI = (*AppService)(nil)
var _ AppServiceAPI = (*MockAppService)(nil)

// MockAppService is a test double for AppServiceAPI.
type MockAppService struct {
	CreateAppFunc        func(ctx context.Context, opts CreateAppOpts) (*App, error)
	GetAppFunc           func(ctx context.Context, name string) (*App, error)
	GetAppWithConfigFunc func(ctx context.Context, name string) (*App, error)
	UpdateAppFunc        func(ctx context.Context, name string, opts UpdateAppOpts) (*App, error)
	ListAppsFunc         func(ctx context.Context) ([]App, error)
	StartAppFunc         func(ctx context.Context, name string) error
	StopAppFunc          func(ctx context.Context, name string) error
	DeleteAppFunc        func(ctx context.Context, name string) error
	UpgradeSummaryFunc   func(ctx context.Context, name string) (*AppUpgradeSummary, error)
	ListImagesFunc       func(ctx context.Context) ([]AppImage, error)
	AvailableSpaceFunc   func(ctx context.Context) (int64, error)
	UpgradeAppFunc       func(ctx context.Context, name string) error
	RedeployAppFunc      func(ctx context.Context, name string) error
	CreateRegistryFunc   func(ctx context.Context, opts CreateRegistryOpts) (*Registry, error)
	GetRegistryFunc      func(ctx context.Context, id int64) (*Registry, error)
	ListRegistriesFunc   func(ctx context.Context) ([]Registry, error)
	UpdateRegistryFunc   func(ctx context.Context, id int64, opts UpdateRegistryOpts) (*Registry, error)
	DeleteRegistryFunc             func(ctx context.Context, id int64) error
	SubscribeStatsFunc             func(ctx context.Context) (*Subscription[[]AppStats], error)
	SubscribeContainerLogsFunc     func(ctx context.Context, opts ContainerLogOpts) (*Subscription[AppContainerLogEntry], error)
}

func (m *MockAppService) CreateApp(ctx context.Context, opts CreateAppOpts) (*App, error) {
	if m.CreateAppFunc != nil {
		return m.CreateAppFunc(ctx, opts)
	}
	return nil, nil
}

func (m *MockAppService) GetApp(ctx context.Context, name string) (*App, error) {
	if m.GetAppFunc != nil {
		return m.GetAppFunc(ctx, name)
	}
	return nil, nil
}

func (m *MockAppService) GetAppWithConfig(ctx context.Context, name string) (*App, error) {
	if m.GetAppWithConfigFunc != nil {
		return m.GetAppWithConfigFunc(ctx, name)
	}
	return nil, nil
}

func (m *MockAppService) UpdateApp(ctx context.Context, name string, opts UpdateAppOpts) (*App, error) {
	if m.UpdateAppFunc != nil {
		return m.UpdateAppFunc(ctx, name, opts)
	}
	return nil, nil
}

func (m *MockAppService) ListApps(ctx context.Context) ([]App, error) {
	if m.ListAppsFunc != nil {
		return m.ListAppsFunc(ctx)
	}
	return nil, nil
}

func (m *MockAppService) StartApp(ctx context.Context, name string) error {
	if m.StartAppFunc != nil {
		return m.StartAppFunc(ctx, name)
	}
	return nil
}

func (m *MockAppService) StopApp(ctx context.Context, name string) error {
	if m.StopAppFunc != nil {
		return m.StopAppFunc(ctx, name)
	}
	return nil
}

func (m *MockAppService) DeleteApp(ctx context.Context, name string) error {
	if m.DeleteAppFunc != nil {
		return m.DeleteAppFunc(ctx, name)
	}
	return nil
}

func (m *MockAppService) UpgradeSummary(ctx context.Context, name string) (*AppUpgradeSummary, error) {
	if m.UpgradeSummaryFunc != nil {
		return m.UpgradeSummaryFunc(ctx, name)
	}
	return nil, nil
}

func (m *MockAppService) ListImages(ctx context.Context) ([]AppImage, error) {
	if m.ListImagesFunc != nil {
		return m.ListImagesFunc(ctx)
	}
	return nil, nil
}

func (m *MockAppService) AvailableSpace(ctx context.Context) (int64, error) {
	if m.AvailableSpaceFunc != nil {
		return m.AvailableSpaceFunc(ctx)
	}
	return 0, nil
}

func (m *MockAppService) UpgradeApp(ctx context.Context, name string) error {
	if m.UpgradeAppFunc != nil {
		return m.UpgradeAppFunc(ctx, name)
	}
	return nil
}

func (m *MockAppService) RedeployApp(ctx context.Context, name string) error {
	if m.RedeployAppFunc != nil {
		return m.RedeployAppFunc(ctx, name)
	}
	return nil
}

func (m *MockAppService) CreateRegistry(ctx context.Context, opts CreateRegistryOpts) (*Registry, error) {
	if m.CreateRegistryFunc != nil {
		return m.CreateRegistryFunc(ctx, opts)
	}
	return nil, nil
}

func (m *MockAppService) GetRegistry(ctx context.Context, id int64) (*Registry, error) {
	if m.GetRegistryFunc != nil {
		return m.GetRegistryFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockAppService) ListRegistries(ctx context.Context) ([]Registry, error) {
	if m.ListRegistriesFunc != nil {
		return m.ListRegistriesFunc(ctx)
	}
	return nil, nil
}

func (m *MockAppService) UpdateRegistry(ctx context.Context, id int64, opts UpdateRegistryOpts) (*Registry, error) {
	if m.UpdateRegistryFunc != nil {
		return m.UpdateRegistryFunc(ctx, id, opts)
	}
	return nil, nil
}

func (m *MockAppService) DeleteRegistry(ctx context.Context, id int64) error {
	if m.DeleteRegistryFunc != nil {
		return m.DeleteRegistryFunc(ctx, id)
	}
	return nil
}

func (m *MockAppService) SubscribeStats(ctx context.Context) (*Subscription[[]AppStats], error) {
	if m.SubscribeStatsFunc != nil {
		return m.SubscribeStatsFunc(ctx)
	}
	return nil, nil
}

func (m *MockAppService) SubscribeContainerLogs(ctx context.Context, opts ContainerLogOpts) (*Subscription[AppContainerLogEntry], error) {
	if m.SubscribeContainerLogsFunc != nil {
		return m.SubscribeContainerLogsFunc(ctx, opts)
	}
	return nil, nil
}

package truenas

import "context"

// CronServiceAPI defines the interface for cron job operations.
type CronServiceAPI interface {
	Create(ctx context.Context, opts CreateCronJobOpts) (*CronJob, error)
	Get(ctx context.Context, id int64) (*CronJob, error)
	List(ctx context.Context) ([]CronJob, error)
	Update(ctx context.Context, id int64, opts UpdateCronJobOpts) (*CronJob, error)
	Delete(ctx context.Context, id int64) error
	Run(ctx context.Context, id int64, skipDisabled bool) error
}

// Compile-time checks.
var _ CronServiceAPI = (*CronService)(nil)
var _ CronServiceAPI = (*MockCronService)(nil)

// MockCronService is a test double for CronServiceAPI.
type MockCronService struct {
	CreateFunc func(ctx context.Context, opts CreateCronJobOpts) (*CronJob, error)
	GetFunc    func(ctx context.Context, id int64) (*CronJob, error)
	ListFunc   func(ctx context.Context) ([]CronJob, error)
	UpdateFunc func(ctx context.Context, id int64, opts UpdateCronJobOpts) (*CronJob, error)
	DeleteFunc func(ctx context.Context, id int64) error
	RunFunc    func(ctx context.Context, id int64, skipDisabled bool) error
}

func (m *MockCronService) Create(ctx context.Context, opts CreateCronJobOpts) (*CronJob, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, opts)
	}
	return nil, nil
}

func (m *MockCronService) Get(ctx context.Context, id int64) (*CronJob, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockCronService) List(ctx context.Context) ([]CronJob, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx)
	}
	return nil, nil
}

func (m *MockCronService) Update(ctx context.Context, id int64, opts UpdateCronJobOpts) (*CronJob, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, opts)
	}
	return nil, nil
}

func (m *MockCronService) Delete(ctx context.Context, id int64) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockCronService) Run(ctx context.Context, id int64, skipDisabled bool) error {
	if m.RunFunc != nil {
		return m.RunFunc(ctx, id, skipDisabled)
	}
	return nil
}

package truenas

import "context"

// UserServiceAPI defines the interface for user operations.
type UserServiceAPI interface {
	Create(ctx context.Context, opts CreateUserOpts) (*User, error)
	Get(ctx context.Context, id int64) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByUID(ctx context.Context, uid int64) (*User, error)
	List(ctx context.Context) ([]User, error)
	Update(ctx context.Context, id int64, opts UpdateUserOpts) (*User, error)
	Delete(ctx context.Context, id int64) error
}

// Compile-time checks.
var _ UserServiceAPI = (*UserService)(nil)
var _ UserServiceAPI = (*MockUserService)(nil)

// MockUserService is a test double for UserServiceAPI.
type MockUserService struct {
	CreateFunc        func(ctx context.Context, opts CreateUserOpts) (*User, error)
	GetFunc           func(ctx context.Context, id int64) (*User, error)
	GetByUsernameFunc func(ctx context.Context, username string) (*User, error)
	GetByUIDFunc      func(ctx context.Context, uid int64) (*User, error)
	ListFunc          func(ctx context.Context) ([]User, error)
	UpdateFunc        func(ctx context.Context, id int64, opts UpdateUserOpts) (*User, error)
	DeleteFunc        func(ctx context.Context, id int64) error
}

func (m *MockUserService) Create(ctx context.Context, opts CreateUserOpts) (*User, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, opts)
	}
	return nil, nil
}

func (m *MockUserService) Get(ctx context.Context, id int64) (*User, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockUserService) GetByUsername(ctx context.Context, username string) (*User, error) {
	if m.GetByUsernameFunc != nil {
		return m.GetByUsernameFunc(ctx, username)
	}
	return nil, nil
}

func (m *MockUserService) GetByUID(ctx context.Context, uid int64) (*User, error) {
	if m.GetByUIDFunc != nil {
		return m.GetByUIDFunc(ctx, uid)
	}
	return nil, nil
}

func (m *MockUserService) List(ctx context.Context) ([]User, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx)
	}
	return nil, nil
}

func (m *MockUserService) Update(ctx context.Context, id int64, opts UpdateUserOpts) (*User, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, opts)
	}
	return nil, nil
}

func (m *MockUserService) Delete(ctx context.Context, id int64) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

package truenas

import (
	"context"
	"testing"
)

func TestMockUserService_ImplementsInterface(t *testing.T) {
	var _ UserServiceAPI = (*UserService)(nil)
	var _ UserServiceAPI = (*MockUserService)(nil)
}

func TestMockUserService_DefaultsToNil(t *testing.T) {
	mock := &MockUserService{}
	ctx := context.Background()

	user, err := mock.Get(ctx, 1)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if user != nil {
		t.Fatalf("expected nil result, got: %v", user)
	}
}

func TestMockUserService_CallsFunc(t *testing.T) {
	called := false
	mock := &MockUserService{
		GetFunc: func(ctx context.Context, id int64) (*User, error) {
			called = true
			return &User{ID: id}, nil
		},
	}

	user, err := mock.Get(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected GetFunc to be called")
	}
	if user.ID != 42 {
		t.Fatalf("expected ID 42, got %d", user.ID)
	}
}

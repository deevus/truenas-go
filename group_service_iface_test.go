package truenas

import (
	"context"
	"testing"
)

func TestMockGroupService_ImplementsInterface(t *testing.T) {
	var _ GroupServiceAPI = (*GroupService)(nil)
	var _ GroupServiceAPI = (*MockGroupService)(nil)
}

func TestMockGroupService_DefaultsToNil(t *testing.T) {
	mock := &MockGroupService{}
	ctx := context.Background()

	group, err := mock.Get(ctx, 1)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if group != nil {
		t.Fatalf("expected nil result, got: %v", group)
	}
}

func TestMockGroupService_CallsFunc(t *testing.T) {
	called := false
	mock := &MockGroupService{
		GetFunc: func(ctx context.Context, id int64) (*Group, error) {
			called = true
			return &Group{ID: id}, nil
		},
	}

	group, err := mock.Get(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected GetFunc to be called")
	}
	if group.ID != 42 {
		t.Fatalf("expected ID 42, got %d", group.ID)
	}
}

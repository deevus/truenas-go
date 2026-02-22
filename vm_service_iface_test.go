package truenas

import (
	"context"
	"testing"
)

func TestMockVMService_ImplementsInterface(t *testing.T) {
	var _ VMServiceAPI = (*VMService)(nil)
	var _ VMServiceAPI = (*MockVMService)(nil)
}

func TestMockVMService_DefaultsToNil(t *testing.T) {
	mock := &MockVMService{}
	ctx := context.Background()

	vm, err := mock.GetVM(ctx, 1)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if vm != nil {
		t.Fatalf("expected nil result, got: %v", vm)
	}
}

func TestMockVMService_CallsFunc(t *testing.T) {
	called := false
	mock := &MockVMService{
		GetVMFunc: func(ctx context.Context, id int64) (*VM, error) {
			called = true
			return &VM{ID: id}, nil
		},
	}

	vm, err := mock.GetVM(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected GetVMFunc to be called")
	}
	if vm.ID != 42 {
		t.Fatalf("expected ID 42, got %d", vm.ID)
	}
}

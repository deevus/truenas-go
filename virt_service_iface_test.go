package truenas

import (
	"context"
	"testing"
)

func TestMockVirtService_ImplementsInterface(t *testing.T) {
	var _ VirtServiceAPI = (*VirtService)(nil)
	var _ VirtServiceAPI = (*MockVirtService)(nil)
}

func TestMockVirtService_DefaultsToNil(t *testing.T) {
	mock := &MockVirtService{}
	ctx := context.Background()

	inst, err := mock.GetInstance(ctx, "test-instance")
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if inst != nil {
		t.Fatalf("expected nil result, got: %v", inst)
	}
}

func TestMockVirtService_CallsFunc(t *testing.T) {
	called := false
	mock := &MockVirtService{
		GetInstanceFunc: func(ctx context.Context, name string) (*VirtInstance, error) {
			called = true
			return &VirtInstance{Name: name}, nil
		},
	}

	inst, err := mock.GetInstance(context.Background(), "test-instance")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected GetInstanceFunc to be called")
	}
	if inst.Name != "test-instance" {
		t.Fatalf("expected name test-instance, got %s", inst.Name)
	}
}

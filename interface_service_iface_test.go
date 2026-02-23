package truenas

import (
	"context"
	"testing"
)

func TestMockInterfaceService_ImplementsInterface(t *testing.T) {
	// Compile-time check
	var _ InterfaceServiceAPI = (*InterfaceService)(nil)
	var _ InterfaceServiceAPI = (*MockInterfaceService)(nil)
}

func TestMockInterfaceService_DefaultsToNil(t *testing.T) {
	mock := &MockInterfaceService{}
	ctx := context.Background()

	ifaces, err := mock.List(ctx)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if ifaces != nil {
		t.Fatalf("expected nil result, got: %v", ifaces)
	}

	iface, err := mock.Get(ctx, "eno1")
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if iface != nil {
		t.Fatalf("expected nil result, got: %v", iface)
	}
}

func TestMockInterfaceService_CallsListFunc(t *testing.T) {
	called := false
	mock := &MockInterfaceService{
		ListFunc: func(ctx context.Context) ([]NetworkInterface, error) {
			called = true
			return []NetworkInterface{{ID: "eno1", Name: "eno1"}}, nil
		},
	}

	ifaces, err := mock.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected ListFunc to be called")
	}
	if len(ifaces) != 1 {
		t.Fatalf("expected 1 interface, got %d", len(ifaces))
	}
	if ifaces[0].ID != "eno1" {
		t.Fatalf("expected ID eno1, got %s", ifaces[0].ID)
	}
}

func TestMockInterfaceService_CallsGetFunc(t *testing.T) {
	called := false
	mock := &MockInterfaceService{
		GetFunc: func(ctx context.Context, id string) (*NetworkInterface, error) {
			called = true
			return &NetworkInterface{ID: id, Name: id}, nil
		},
	}

	iface, err := mock.Get(context.Background(), "eno1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected GetFunc to be called")
	}
	if iface.ID != "eno1" {
		t.Fatalf("expected ID eno1, got %s", iface.ID)
	}
}

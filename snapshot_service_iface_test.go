package truenas

import (
	"context"
	"testing"
)

func TestMockSnapshotService_ImplementsInterface(t *testing.T) {
	// Compile-time check
	var _ SnapshotServiceAPI = (*SnapshotService)(nil)
	var _ SnapshotServiceAPI = (*MockSnapshotService)(nil)
}

func TestMockSnapshotService_DefaultsToNil(t *testing.T) {
	mock := &MockSnapshotService{}
	ctx := context.Background()

	snap, err := mock.Get(ctx, "pool/test@snap1")
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if snap != nil {
		t.Fatalf("expected nil result, got: %v", snap)
	}
}

func TestMockSnapshotService_CallsFunc(t *testing.T) {
	called := false
	mock := &MockSnapshotService{
		GetFunc: func(ctx context.Context, id string) (*Snapshot, error) {
			called = true
			return &Snapshot{ID: id}, nil
		},
	}

	snap, err := mock.Get(context.Background(), "pool/test@snap1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected GetFunc to be called")
	}
	if snap.ID != "pool/test@snap1" {
		t.Fatalf("expected ID pool/test@snap1, got %s", snap.ID)
	}
}

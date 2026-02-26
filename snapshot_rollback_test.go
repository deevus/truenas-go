package truenas

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestSnapshotService_Rollback(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "zfs.snapshot.rollback" {
				t.Errorf("expected method zfs.snapshot.rollback, got %s", method)
			}
			id, ok := params.(string)
			if !ok || id != "pool/dataset@snap1" {
				t.Errorf("expected id pool/dataset@snap1, got %v", params)
			}
			return nil, nil
		},
	}

	svc := NewSnapshotService(mock, Version{Major: 24, Minor: 10})
	err := svc.Rollback(context.Background(), "pool/dataset@snap1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSnapshotService_Rollback_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("rollback failed")
		},
	}

	svc := NewSnapshotService(mock, Version{Major: 24, Minor: 10})
	err := svc.Rollback(context.Background(), "pool/dataset@snap1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSnapshotService_Rollback_VersionPrefix(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "pool.snapshot.rollback" {
				t.Errorf("expected method pool.snapshot.rollback, got %s", method)
			}
			return nil, nil
		},
	}

	svc := NewSnapshotService(mock, Version{Major: 25, Minor: 10})
	err := svc.Rollback(context.Background(), "pool/dataset@snap1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

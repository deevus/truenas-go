package truenas

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestSnapshotService_Query(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "zfs.snapshot.query" {
				t.Errorf("expected method zfs.snapshot.query, got %s", method)
			}
			filters := params.([][]any)
			if len(filters) != 1 {
				t.Fatalf("expected 1 filter, got %d", len(filters))
			}
			if filters[0][0] != "dataset" || filters[0][1] != "=" || filters[0][2] != "pool/dataset" {
				t.Errorf("unexpected filter: %v", filters[0])
			}
			return sampleSnapshotJSON(), nil
		},
	}

	svc := NewSnapshotService(mock, Version{Major: 24, Minor: 10})
	snaps, err := svc.Query(context.Background(), [][]any{{"dataset", "=", "pool/dataset"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(snaps) != 1 {
		t.Fatalf("expected 1 snapshot, got %d", len(snaps))
	}
	if snaps[0].ID != "pool/dataset@snap1" {
		t.Errorf("expected ID pool/dataset@snap1, got %s", snaps[0].ID)
	}
	if snaps[0].Dataset != "pool/dataset" {
		t.Errorf("expected dataset pool/dataset, got %s", snaps[0].Dataset)
	}
}

func TestSnapshotService_Query_NoFilter(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if params != nil {
				t.Errorf("expected nil params for no filter, got %v", params)
			}
			return sampleSnapshotJSON(), nil
		},
	}

	svc := NewSnapshotService(mock, Version{Major: 24, Minor: 10})
	snaps, err := svc.Query(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(snaps) != 1 {
		t.Fatalf("expected 1 snapshot, got %d", len(snaps))
	}
}

func TestSnapshotService_Query_Empty(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return json.RawMessage(`[]`), nil
		},
	}

	svc := NewSnapshotService(mock, Version{Major: 24, Minor: 10})
	snaps, err := svc.Query(context.Background(), [][]any{{"dataset", "=", "pool/nonexistent"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snaps == nil {
		t.Fatal("expected non-nil empty slice")
	}
	if len(snaps) != 0 {
		t.Errorf("expected 0 snapshots, got %d", len(snaps))
	}
}

func TestSnapshotService_Query_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("network error")
		},
	}

	svc := NewSnapshotService(mock, Version{Major: 24, Minor: 10})
	_, err := svc.Query(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSnapshotService_Query_ParseError(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return json.RawMessage(`not json`), nil
		},
	}

	svc := NewSnapshotService(mock, Version{Major: 24, Minor: 10})
	_, err := svc.Query(context.Background(), nil)
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestSnapshotService_Query_VersionPrefix(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "pool.snapshot.query" {
				t.Errorf("expected method pool.snapshot.query, got %s", method)
			}
			return json.RawMessage(`[]`), nil
		},
	}

	svc := NewSnapshotService(mock, Version{Major: 25, Minor: 10})
	_, err := svc.Query(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

package truenas

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestDatasetService_ListPools(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "pool.query" {
				t.Errorf("expected method pool.query, got %s", method)
			}
			if params != nil {
				t.Error("expected nil params for ListPools")
			}
			return samplePoolQueryJSON(), nil
		},
	}

	svc := NewDatasetService(mock, Version{})
	pools, err := svc.ListPools(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pools) != 2 {
		t.Fatalf("expected 2 pools, got %d", len(pools))
	}
	if pools[0].ID != 1 {
		t.Errorf("expected first pool ID 1, got %d", pools[0].ID)
	}
	if pools[0].Name != "pool1" {
		t.Errorf("expected first pool name pool1, got %s", pools[0].Name)
	}
	if pools[0].Path != "/mnt/pool1" {
		t.Errorf("expected first pool path /mnt/pool1, got %s", pools[0].Path)
	}
	if pools[1].ID != 2 {
		t.Errorf("expected second pool ID 2, got %d", pools[1].ID)
	}
	if pools[1].Name != "pool2" {
		t.Errorf("expected second pool name pool2, got %s", pools[1].Name)
	}
}

func TestDatasetService_ListPools_Empty(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return json.RawMessage(`[]`), nil
		},
	}

	svc := NewDatasetService(mock, Version{})
	pools, err := svc.ListPools(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pools) != 0 {
		t.Errorf("expected 0 pools, got %d", len(pools))
	}
}

func TestDatasetService_ListPools_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("network error")
		},
	}

	svc := NewDatasetService(mock, Version{})
	_, err := svc.ListPools(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDatasetService_ListPools_AllFields(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return json.RawMessage(`[{
				"id": 1,
				"name": "tank",
				"path": "/mnt/tank",
				"status": "ONLINE",
				"size": 1099511627776,
				"allocated": 549755813888,
				"free": 549755813888
			}]`), nil
		},
	}

	svc := NewDatasetService(mock, Version{})
	pools, err := svc.ListPools(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pools) != 1 {
		t.Fatalf("expected 1 pool, got %d", len(pools))
	}

	pool := pools[0]
	if pool.Status != "ONLINE" {
		t.Errorf("expected Status ONLINE, got %s", pool.Status)
	}
	if pool.Size != 1099511627776 {
		t.Errorf("expected Size 1099511627776, got %d", pool.Size)
	}
	if pool.Allocated != 549755813888 {
		t.Errorf("expected Allocated 549755813888, got %d", pool.Allocated)
	}
	if pool.Free != 549755813888 {
		t.Errorf("expected Free 549755813888, got %d", pool.Free)
	}
}

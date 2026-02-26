package truenas

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func sampleVirtInstanceListJSON() json.RawMessage {
	return json.RawMessage(`[
		{
			"id": "abc123",
			"name": "px-test1",
			"type": "CONTAINER",
			"status": "RUNNING",
			"cpu": "2",
			"memory": 1073741824,
			"autostart": true,
			"environment": {},
			"aliases": [{"type": "INET", "address": "10.0.0.5", "netmask": 24}],
			"image": {"architecture": "x86_64", "description": "Ubuntu 24.04", "os": "ubuntu", "release": "24.04", "variant": "default"},
			"storage_pool": "tank"
		},
		{
			"id": "def456",
			"name": "px-test2",
			"type": "CONTAINER",
			"status": "STOPPED",
			"cpu": "1",
			"memory": 536870912,
			"autostart": false,
			"environment": {},
			"aliases": [],
			"image": {"architecture": "x86_64", "description": "Alpine 3.19", "os": "alpine", "release": "3.19", "variant": "default"},
			"storage_pool": "tank"
		}
	]`)
}

func TestVirtService_ListInstances(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				if method != "virt.instance.query" {
					t.Errorf("expected method virt.instance.query, got %s", method)
				}
				// Verify filter is passed correctly
				slice := params.([]any)
				filters := slice[0].([][]any)
				if len(filters) != 1 {
					t.Fatalf("expected 1 filter, got %d", len(filters))
				}
				if filters[0][0] != "name" || filters[0][1] != "^" || filters[0][2] != "px-" {
					t.Errorf("unexpected filter: %v", filters[0])
				}
				return sampleVirtInstanceListJSON(), nil
			},
		},
	}

	svc := NewVirtService(mock, Version{})
	instances, err := svc.ListInstances(context.Background(), [][]any{{"name", "^", "px-"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(instances) != 2 {
		t.Fatalf("expected 2 instances, got %d", len(instances))
	}
	if instances[0].Name != "px-test1" {
		t.Errorf("expected name px-test1, got %s", instances[0].Name)
	}
	if instances[0].Status != "RUNNING" {
		t.Errorf("expected status RUNNING, got %s", instances[0].Status)
	}
	if instances[1].Name != "px-test2" {
		t.Errorf("expected name px-test2, got %s", instances[1].Name)
	}
	if instances[1].Status != "STOPPED" {
		t.Errorf("expected status STOPPED, got %s", instances[1].Status)
	}
}

func TestVirtService_ListInstances_NoFilter(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				if params != nil {
					t.Errorf("expected nil params for no filter, got %v", params)
				}
				return sampleVirtInstanceListJSON(), nil
			},
		},
	}

	svc := NewVirtService(mock, Version{})
	instances, err := svc.ListInstances(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(instances) != 2 {
		t.Fatalf("expected 2 instances, got %d", len(instances))
	}
}

func TestVirtService_ListInstances_Empty(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return json.RawMessage(`[]`), nil
			},
		},
	}

	svc := NewVirtService(mock, Version{})
	instances, err := svc.ListInstances(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if instances == nil {
		t.Fatal("expected non-nil empty slice")
	}
	if len(instances) != 0 {
		t.Errorf("expected 0 instances, got %d", len(instances))
	}
}

func TestVirtService_ListInstances_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, errors.New("network error")
			},
		},
	}

	svc := NewVirtService(mock, Version{})
	_, err := svc.ListInstances(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestVirtService_ListInstances_ParseError(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return json.RawMessage(`not json`), nil
			},
		},
	}

	svc := NewVirtService(mock, Version{})
	_, err := svc.ListInstances(context.Background(), nil)
	if err == nil {
		t.Fatal("expected parse error")
	}
}

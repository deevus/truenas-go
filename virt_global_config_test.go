package truenas

import (
	"context"
	"encoding/json"
	"testing"
)

func TestVirtGlobalConfigFromResponse_WithNewFields(t *testing.T) {
	dataset := "tank/ix-virt"
	state := "INITIALIZED"
	pool := "tank"
	resp := VirtGlobalConfigResponse{
		Pool:         &pool,
		Dataset:      &dataset,
		StoragePools: []string{"tank", "ssd"},
		State:        &state,
	}

	cfg := virtGlobalConfigFromResponse(resp)
	if cfg.Dataset != "tank/ix-virt" {
		t.Errorf("expected dataset tank/ix-virt, got %s", cfg.Dataset)
	}
	if len(cfg.StoragePools) != 2 {
		t.Fatalf("expected 2 storage pools, got %d", len(cfg.StoragePools))
	}
	if cfg.StoragePools[0] != "tank" {
		t.Errorf("expected first storage pool tank, got %s", cfg.StoragePools[0])
	}
	if cfg.StoragePools[1] != "ssd" {
		t.Errorf("expected second storage pool ssd, got %s", cfg.StoragePools[1])
	}
	if cfg.State != "INITIALIZED" {
		t.Errorf("expected state INITIALIZED, got %s", cfg.State)
	}
}

func TestVirtGlobalConfigFromResponse_NilNewFields(t *testing.T) {
	resp := VirtGlobalConfigResponse{}

	cfg := virtGlobalConfigFromResponse(resp)
	if cfg.Dataset != "" {
		t.Errorf("expected empty dataset, got %s", cfg.Dataset)
	}
	if cfg.StoragePools == nil {
		t.Error("expected non-nil storage pools slice")
	}
	if len(cfg.StoragePools) != 0 {
		t.Errorf("expected 0 storage pools, got %d", len(cfg.StoragePools))
	}
	if cfg.State != "" {
		t.Errorf("expected empty state, got %s", cfg.State)
	}
}

func TestVirtService_GetGlobalConfig_AllFields(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				if method != "virt.global.config" {
					t.Errorf("expected method virt.global.config, got %s", method)
				}
				return json.RawMessage(`{
					"bridge": "br0",
					"v4_network": "10.0.0.0/24",
					"v6_network": "fd00::/64",
					"pool": "tank",
					"dataset": "tank/ix-virt",
					"storage_pools": ["tank"],
					"state": "INITIALIZED"
				}`), nil
			},
		},
	}

	svc := NewVirtService(mock, Version{})
	cfg, err := svc.GetGlobalConfig(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Dataset != "tank/ix-virt" {
		t.Errorf("expected dataset tank/ix-virt, got %s", cfg.Dataset)
	}
	if len(cfg.StoragePools) != 1 || cfg.StoragePools[0] != "tank" {
		t.Errorf("expected storage_pools [tank], got %v", cfg.StoragePools)
	}
	if cfg.State != "INITIALIZED" {
		t.Errorf("expected state INITIALIZED, got %s", cfg.State)
	}
}

func TestVirtService_GetGlobalConfig_NullNewFields(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return json.RawMessage(`{
					"bridge": null,
					"v4_network": null,
					"v6_network": null,
					"pool": null,
					"dataset": null,
					"storage_pools": null,
					"state": null
				}`), nil
			},
		},
	}

	svc := NewVirtService(mock, Version{})
	cfg, err := svc.GetGlobalConfig(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Dataset != "" {
		t.Errorf("expected empty dataset, got %s", cfg.Dataset)
	}
	if cfg.StoragePools == nil {
		t.Error("expected non-nil storage pools")
	}
	if len(cfg.StoragePools) != 0 {
		t.Errorf("expected 0 storage pools, got %d", len(cfg.StoragePools))
	}
	if cfg.State != "" {
		t.Errorf("expected empty state, got %s", cfg.State)
	}
}

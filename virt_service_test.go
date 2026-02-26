package truenas

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestVirtService_GetGlobalConfig(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				if method != "virt.global.config" {
					t.Errorf("expected method virt.global.config, got %s", method)
				}
				if params != nil {
					t.Error("expected nil params")
				}
				return json.RawMessage(`{
					"bridge": "br0",
					"v4_network": "10.0.0.0/24",
					"v6_network": "fd00::/64",
					"pool": "tank"
				}`), nil
			},
		},
	}

	svc := NewVirtService(mock, Version{})
	cfg, err := svc.GetGlobalConfig(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.Bridge != "br0" {
		t.Errorf("expected bridge br0, got %s", cfg.Bridge)
	}
	if cfg.V4Network != "10.0.0.0/24" {
		t.Errorf("expected v4_network 10.0.0.0/24, got %s", cfg.V4Network)
	}
	if cfg.V6Network != "fd00::/64" {
		t.Errorf("expected v6_network fd00::/64, got %s", cfg.V6Network)
	}
	if cfg.Pool != "tank" {
		t.Errorf("expected pool tank, got %s", cfg.Pool)
	}
}

func TestVirtService_GetGlobalConfig_NilFields(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return json.RawMessage(`{
					"bridge": null,
					"v4_network": null,
					"v6_network": null,
					"pool": null
				}`), nil
			},
		},
	}

	svc := NewVirtService(mock, Version{})
	cfg, err := svc.GetGlobalConfig(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Bridge != "" {
		t.Errorf("expected empty bridge, got %s", cfg.Bridge)
	}
	if cfg.V4Network != "" {
		t.Errorf("expected empty v4_network, got %s", cfg.V4Network)
	}
	if cfg.V6Network != "" {
		t.Errorf("expected empty v6_network, got %s", cfg.V6Network)
	}
	if cfg.Pool != "" {
		t.Errorf("expected empty pool, got %s", cfg.Pool)
	}
}

func TestVirtService_GetGlobalConfig_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, errors.New("connection refused")
			},
		},
	}

	svc := NewVirtService(mock, Version{})
	cfg, err := svc.GetGlobalConfig(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if cfg != nil {
		t.Error("expected nil config on error")
	}
}

func TestVirtService_GetGlobalConfig_ParseError(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return json.RawMessage(`not json`), nil
			},
		},
	}

	svc := NewVirtService(mock, Version{})
	_, err := svc.GetGlobalConfig(context.Background())
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestVirtService_UpdateGlobalConfig(t *testing.T) {
	callCount := 0
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				callCount++
				if callCount == 1 {
					if method != "virt.global.update" {
						t.Errorf("expected method virt.global.update, got %s", method)
					}
					p := params.(map[string]any)
					if p["bridge"] != "br1" {
						t.Errorf("expected bridge br1, got %v", p["bridge"])
					}
					if p["pool"] != "newpool" {
						t.Errorf("expected pool newpool, got %v", p["pool"])
					}
					if p["v4_network"] != "192.168.1.0/24" {
						t.Errorf("expected v4_network 192.168.1.0/24, got %v", p["v4_network"])
					}
					if p["v6_network"] != "fd01::/64" {
						t.Errorf("expected v6_network fd01::/64, got %v", p["v6_network"])
					}
					return json.RawMessage(`{}`), nil
				}
				// Re-read via GetGlobalConfig
				return json.RawMessage(`{
					"bridge": "br1",
					"v4_network": "192.168.1.0/24",
					"v6_network": "fd01::/64",
					"pool": "newpool"
				}`), nil
			},
		},
	}

	svc := NewVirtService(mock, Version{})
	bridge := "br1"
	v4 := "192.168.1.0/24"
	v6 := "fd01::/64"
	pool := "newpool"
	cfg, err := svc.UpdateGlobalConfig(context.Background(), UpdateVirtGlobalConfigOpts{
		Bridge:    &bridge,
		V4Network: &v4,
		V6Network: &v6,
		Pool:      &pool,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.Bridge != "br1" {
		t.Errorf("expected bridge br1, got %s", cfg.Bridge)
	}
	if cfg.Pool != "newpool" {
		t.Errorf("expected pool newpool, got %s", cfg.Pool)
	}
}

func TestVirtService_UpdateGlobalConfig_Partial(t *testing.T) {
	callCount := 0
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				callCount++
				if callCount == 1 {
					p := params.(map[string]any)
					// Only pool should be present
					if _, ok := p["bridge"]; ok {
						t.Error("bridge should not be in params")
					}
					if _, ok := p["v4_network"]; ok {
						t.Error("v4_network should not be in params")
					}
					if _, ok := p["v6_network"]; ok {
						t.Error("v6_network should not be in params")
					}
					if p["pool"] != "tank2" {
						t.Errorf("expected pool tank2, got %v", p["pool"])
					}
					return json.RawMessage(`{}`), nil
				}
				return json.RawMessage(`{
					"bridge": "br0",
					"v4_network": "10.0.0.0/24",
					"v6_network": null,
					"pool": "tank2"
				}`), nil
			},
		},
	}

	svc := NewVirtService(mock, Version{})
	pool := "tank2"
	cfg, err := svc.UpdateGlobalConfig(context.Background(), UpdateVirtGlobalConfigOpts{
		Pool: &pool,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Pool != "tank2" {
		t.Errorf("expected pool tank2, got %s", cfg.Pool)
	}
}

func TestVirtService_UpdateGlobalConfig_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, errors.New("permission denied")
			},
		},
	}

	svc := NewVirtService(mock, Version{})
	pool := "tank"
	_, err := svc.UpdateGlobalConfig(context.Background(), UpdateVirtGlobalConfigOpts{
		Pool: &pool,
	})
	if err == nil {
		t.Fatal("expected error")
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

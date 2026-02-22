package truenas

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

// --- Global Config Tests ---

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

// --- Instance Tests ---

func sampleVirtInstanceJSON() json.RawMessage {
	return json.RawMessage(`{
		"id": "abc123",
		"name": "myvm",
		"type": "VM",
		"status": "RUNNING",
		"cpu": "2",
		"memory": 2147483648,
		"autostart": true,
		"environment": {"FOO": "bar", "BAZ": "qux"},
		"aliases": [
			{"type": "INET", "address": "10.0.0.5", "netmask": 24},
			{"type": "INET6", "address": "fd00::5", "netmask": 64}
		],
		"image": {
			"architecture": "x86_64",
			"description": "Ubuntu 22.04",
			"os": "ubuntu",
			"release": "22.04",
			"variant": "default"
		},
		"storage_pool": "tank"
	}`)
}

func TestVirtService_CreateInstance(t *testing.T) {
	callCount := 0
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				callCount++
				// Re-read via GetInstance
				if method != "virt.instance.get_instance" {
					t.Errorf("expected method virt.instance.get_instance, got %s", method)
				}
				return sampleVirtInstanceJSON(), nil
			},
		},
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "virt.instance.create" {
				t.Errorf("expected method virt.instance.create, got %s", method)
			}
			p := params.(map[string]any)
			if p["name"] != "myvm" {
				t.Errorf("expected name myvm, got %v", p["name"])
			}
			if p["instance_type"] != "VM" {
				t.Errorf("expected instance_type VM, got %v", p["instance_type"])
			}
			return json.RawMessage(`{}`), nil
		},
	}

	svc := NewVirtService(mock, Version{})
	inst, err := svc.CreateInstance(context.Background(), CreateVirtInstanceOpts{
		Name:      "myvm",
		Type:      "VM",
		Image:     "ubuntu/22.04",
		CPU:       "2",
		Memory:    2147483648,
		Autostart: true,
		Environment: map[string]string{"FOO": "bar", "BAZ": "qux"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inst == nil {
		t.Fatal("expected non-nil instance")
	}
	if inst.Name != "myvm" {
		t.Errorf("expected name myvm, got %s", inst.Name)
	}
	if inst.Type != "VM" {
		t.Errorf("expected type VM, got %s", inst.Type)
	}
	if inst.Status != "RUNNING" {
		t.Errorf("expected status RUNNING, got %s", inst.Status)
	}
	if inst.CPU != "2" {
		t.Errorf("expected cpu 2, got %s", inst.CPU)
	}
	if inst.Memory != 2147483648 {
		t.Errorf("expected memory 2147483648, got %d", inst.Memory)
	}
	if len(inst.Aliases) != 2 {
		t.Fatalf("expected 2 aliases, got %d", len(inst.Aliases))
	}
	if inst.Aliases[0].Address != "10.0.0.5" {
		t.Errorf("expected alias address 10.0.0.5, got %s", inst.Aliases[0].Address)
	}
}

func TestVirtService_CreateInstance_WithDevices(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return sampleVirtInstanceJSON(), nil
			},
		},
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			p := params.(map[string]any)
			devices, ok := p["devices"].([]map[string]any)
			if !ok {
				t.Fatal("expected devices in params")
			}
			if len(devices) != 1 {
				t.Fatalf("expected 1 device, got %d", len(devices))
			}
			if devices[0]["dev_type"] != "DISK" {
				t.Errorf("expected dev_type DISK, got %v", devices[0]["dev_type"])
			}
			return json.RawMessage(`{}`), nil
		},
	}

	svc := NewVirtService(mock, Version{})
	_, err := svc.CreateInstance(context.Background(), CreateVirtInstanceOpts{
		Name:  "myvm",
		Type:  "VM",
		Image: "ubuntu/22.04",
		Devices: []VirtDeviceOpts{
			{DevType: "DISK", Source: "/mnt/tank/data", Destination: "/data"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVirtService_CreateInstance_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("creation failed")
		},
	}

	svc := NewVirtService(mock, Version{})
	inst, err := svc.CreateInstance(context.Background(), CreateVirtInstanceOpts{
		Name: "myvm",
		Type: "VM",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if inst != nil {
		t.Error("expected nil instance on error")
	}
}

func TestVirtService_GetInstance(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				if method != "virt.instance.get_instance" {
					t.Errorf("expected method virt.instance.get_instance, got %s", method)
				}
				if params != "myvm" {
					t.Errorf("expected params myvm, got %v", params)
				}
				return sampleVirtInstanceJSON(), nil
			},
		},
	}

	svc := NewVirtService(mock, Version{})
	inst, err := svc.GetInstance(context.Background(), "myvm")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inst == nil {
		t.Fatal("expected non-nil instance")
	}
	if inst.Name != "myvm" {
		t.Errorf("expected name myvm, got %s", inst.Name)
	}
	if inst.Autostart != true {
		t.Error("expected autostart true")
	}
	if inst.Environment["FOO"] != "bar" {
		t.Errorf("expected FOO=bar, got %s", inst.Environment["FOO"])
	}
	if inst.StoragePool != "tank" {
		t.Errorf("expected storage_pool tank, got %s", inst.StoragePool)
	}
	if inst.Image.OS != "ubuntu" {
		t.Errorf("expected image os ubuntu, got %s", inst.Image.OS)
	}
}

func TestVirtService_GetInstance_NotFound(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, errors.New("instance does not exist")
			},
		},
	}

	svc := NewVirtService(mock, Version{})
	inst, err := svc.GetInstance(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inst != nil {
		t.Error("expected nil instance for not found")
	}
}

func TestVirtService_GetInstance_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, errors.New("connection timeout")
			},
		},
	}

	svc := NewVirtService(mock, Version{})
	_, err := svc.GetInstance(context.Background(), "myvm")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestVirtService_GetInstance_ParseError(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return json.RawMessage(`not json`), nil
			},
		},
	}

	svc := NewVirtService(mock, Version{})
	_, err := svc.GetInstance(context.Background(), "myvm")
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestVirtService_UpdateInstance(t *testing.T) {
	callCount := 0
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				callCount++
				return sampleVirtInstanceJSON(), nil
			},
		},
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "virt.instance.update" {
				t.Errorf("expected method virt.instance.update, got %s", method)
			}
			slice := params.([]any)
			if len(slice) != 2 {
				t.Fatalf("expected 2 elements, got %d", len(slice))
			}
			if slice[0] != "myvm" {
				t.Errorf("expected name myvm, got %v", slice[0])
			}
			p := slice[1].(map[string]any)
			if p["autostart"] != false {
				t.Errorf("expected autostart false, got %v", p["autostart"])
			}
			env := p["environment"].(map[string]string)
			if env["KEY"] != "val" {
				t.Errorf("expected KEY=val, got %v", env["KEY"])
			}
			return json.RawMessage(`{}`), nil
		},
	}

	svc := NewVirtService(mock, Version{})
	autostart := false
	inst, err := svc.UpdateInstance(context.Background(), "myvm", UpdateVirtInstanceOpts{
		Autostart:   &autostart,
		Environment: map[string]string{"KEY": "val"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inst == nil {
		t.Fatal("expected non-nil instance")
	}
}

func TestVirtService_UpdateInstance_AutostartOnly(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return sampleVirtInstanceJSON(), nil
			},
		},
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			slice := params.([]any)
			p := slice[1].(map[string]any)
			if _, ok := p["environment"]; ok {
				t.Error("environment should not be in params")
			}
			if p["autostart"] != true {
				t.Errorf("expected autostart true, got %v", p["autostart"])
			}
			return json.RawMessage(`{}`), nil
		},
	}

	svc := NewVirtService(mock, Version{})
	autostart := true
	_, err := svc.UpdateInstance(context.Background(), "myvm", UpdateVirtInstanceOpts{
		Autostart: &autostart,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVirtService_UpdateInstance_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("update failed")
		},
	}

	svc := NewVirtService(mock, Version{})
	autostart := true
	_, err := svc.UpdateInstance(context.Background(), "myvm", UpdateVirtInstanceOpts{
		Autostart: &autostart,
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestVirtService_DeleteInstance(t *testing.T) {
	mock := &mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "virt.instance.delete" {
				t.Errorf("expected method virt.instance.delete, got %s", method)
			}
			if params != "myvm" {
				t.Errorf("expected params myvm, got %v", params)
			}
			return nil, nil
		},
	}

	svc := NewVirtService(mock, Version{})
	err := svc.DeleteInstance(context.Background(), "myvm")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVirtService_DeleteInstance_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("delete failed")
		},
	}

	svc := NewVirtService(mock, Version{})
	err := svc.DeleteInstance(context.Background(), "myvm")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestVirtService_StartInstance(t *testing.T) {
	mock := &mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "virt.instance.start" {
				t.Errorf("expected method virt.instance.start, got %s", method)
			}
			if params != "myvm" {
				t.Errorf("expected params myvm, got %v", params)
			}
			return nil, nil
		},
	}

	svc := NewVirtService(mock, Version{})
	err := svc.StartInstance(context.Background(), "myvm")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVirtService_StartInstance_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("start failed")
		},
	}

	svc := NewVirtService(mock, Version{})
	err := svc.StartInstance(context.Background(), "myvm")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestVirtService_StopInstance(t *testing.T) {
	mock := &mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "virt.instance.stop" {
				t.Errorf("expected method virt.instance.stop, got %s", method)
			}
			slice := params.([]any)
			if len(slice) != 2 {
				t.Fatalf("expected 2 elements, got %d", len(slice))
			}
			if slice[0] != "myvm" {
				t.Errorf("expected name myvm, got %v", slice[0])
			}
			stopArgs := slice[1].(map[string]any)
			if stopArgs["timeout"] != int64(30) {
				t.Errorf("expected timeout 30, got %v", stopArgs["timeout"])
			}
			return nil, nil
		},
	}

	svc := NewVirtService(mock, Version{})
	err := svc.StopInstance(context.Background(), "myvm", StopVirtInstanceOpts{Timeout: 30})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVirtService_StopInstance_DefaultTimeout(t *testing.T) {
	mock := &mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			slice := params.([]any)
			stopArgs := slice[1].(map[string]any)
			if _, ok := stopArgs["timeout"]; ok {
				t.Error("timeout should not be set for default (0)")
			}
			return nil, nil
		},
	}

	svc := NewVirtService(mock, Version{})
	err := svc.StopInstance(context.Background(), "myvm", StopVirtInstanceOpts{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVirtService_StopInstance_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("stop failed")
		},
	}

	svc := NewVirtService(mock, Version{})
	err := svc.StopInstance(context.Background(), "myvm", StopVirtInstanceOpts{Timeout: 10})
	if err == nil {
		t.Fatal("expected error")
	}
}

// --- Device Tests ---

func TestVirtService_ListDevices(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				if method != "virt.instance.device_list" {
					t.Errorf("expected method virt.instance.device_list, got %s", method)
				}
				if params != "myvm" {
					t.Errorf("expected params myvm, got %v", params)
				}
				return json.RawMessage(`[
					{
						"dev_type": "DISK",
						"name": "data",
						"description": "Data disk",
						"readonly": false,
						"source": "/mnt/tank/data",
						"destination": "/data"
					},
					{
						"dev_type": "NIC",
						"name": "eth0",
						"description": "Primary NIC",
						"readonly": false,
						"network": "br0",
						"nic_type": "BRIDGED",
						"parent": "enp0s3"
					}
				]`), nil
			},
		},
	}

	svc := NewVirtService(mock, Version{})
	devices, err := svc.ListDevices(context.Background(), "myvm")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(devices) != 2 {
		t.Fatalf("expected 2 devices, got %d", len(devices))
	}
	if devices[0].DevType != "DISK" {
		t.Errorf("expected dev_type DISK, got %s", devices[0].DevType)
	}
	if devices[0].Name != "data" {
		t.Errorf("expected name data, got %s", devices[0].Name)
	}
	if devices[0].Source != "/mnt/tank/data" {
		t.Errorf("expected source /mnt/tank/data, got %s", devices[0].Source)
	}
	if devices[1].DevType != "NIC" {
		t.Errorf("expected dev_type NIC, got %s", devices[1].DevType)
	}
	if devices[1].Network != "br0" {
		t.Errorf("expected network br0, got %s", devices[1].Network)
	}
}

func TestVirtService_ListDevices_Empty(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return json.RawMessage(`[]`), nil
			},
		},
	}

	svc := NewVirtService(mock, Version{})
	devices, err := svc.ListDevices(context.Background(), "myvm")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(devices) != 0 {
		t.Errorf("expected 0 devices, got %d", len(devices))
	}
}

func TestVirtService_ListDevices_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, errors.New("network error")
			},
		},
	}

	svc := NewVirtService(mock, Version{})
	_, err := svc.ListDevices(context.Background(), "myvm")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestVirtService_ListDevices_ParseError(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return json.RawMessage(`not json`), nil
			},
		},
	}

	svc := NewVirtService(mock, Version{})
	_, err := svc.ListDevices(context.Background(), "myvm")
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestVirtService_AddDevice_Disk(t *testing.T) {
	mock := &mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "virt.instance.device_add" {
				t.Errorf("expected method virt.instance.device_add, got %s", method)
			}
			slice := params.([]any)
			if slice[0] != "myvm" {
				t.Errorf("expected instanceID myvm, got %v", slice[0])
			}
			dev := slice[1].(map[string]any)
			if dev["dev_type"] != "DISK" {
				t.Errorf("expected dev_type DISK, got %v", dev["dev_type"])
			}
			if dev["source"] != "/mnt/tank/data" {
				t.Errorf("expected source /mnt/tank/data, got %v", dev["source"])
			}
			if dev["destination"] != "/data" {
				t.Errorf("expected destination /data, got %v", dev["destination"])
			}
			return nil, nil
		},
	}

	svc := NewVirtService(mock, Version{})
	err := svc.AddDevice(context.Background(), "myvm", VirtDeviceOpts{
		DevType:     "DISK",
		Source:      "/mnt/tank/data",
		Destination: "/data",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVirtService_AddDevice_NIC(t *testing.T) {
	mock := &mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			slice := params.([]any)
			dev := slice[1].(map[string]any)
			if dev["dev_type"] != "NIC" {
				t.Errorf("expected dev_type NIC, got %v", dev["dev_type"])
			}
			if dev["network"] != "br0" {
				t.Errorf("expected network br0, got %v", dev["network"])
			}
			if dev["nic_type"] != "BRIDGED" {
				t.Errorf("expected nic_type BRIDGED, got %v", dev["nic_type"])
			}
			if dev["parent"] != "enp0s3" {
				t.Errorf("expected parent enp0s3, got %v", dev["parent"])
			}
			return nil, nil
		},
	}

	svc := NewVirtService(mock, Version{})
	err := svc.AddDevice(context.Background(), "myvm", VirtDeviceOpts{
		DevType: "NIC",
		Name:    "eth0",
		Network: "br0",
		NICType: "BRIDGED",
		Parent:  "enp0s3",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVirtService_AddDevice_Proxy(t *testing.T) {
	mock := &mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			slice := params.([]any)
			dev := slice[1].(map[string]any)
			if dev["dev_type"] != "PROXY" {
				t.Errorf("expected dev_type PROXY, got %v", dev["dev_type"])
			}
			if dev["source_proto"] != "TCP" {
				t.Errorf("expected source_proto TCP, got %v", dev["source_proto"])
			}
			if dev["source_port"] != int64(8080) {
				t.Errorf("expected source_port 8080, got %v", dev["source_port"])
			}
			if dev["dest_proto"] != "TCP" {
				t.Errorf("expected dest_proto TCP, got %v", dev["dest_proto"])
			}
			if dev["dest_port"] != int64(80) {
				t.Errorf("expected dest_port 80, got %v", dev["dest_port"])
			}
			return nil, nil
		},
	}

	svc := NewVirtService(mock, Version{})
	err := svc.AddDevice(context.Background(), "myvm", VirtDeviceOpts{
		DevType:     "PROXY",
		SourceProto: "TCP",
		SourcePort:  8080,
		DestProto:   "TCP",
		DestPort:    80,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVirtService_AddDevice_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("device add failed")
		},
	}

	svc := NewVirtService(mock, Version{})
	err := svc.AddDevice(context.Background(), "myvm", VirtDeviceOpts{DevType: "DISK"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestVirtService_DeleteDevice(t *testing.T) {
	mock := &mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "virt.instance.device_delete" {
				t.Errorf("expected method virt.instance.device_delete, got %s", method)
			}
			slice := params.([]any)
			if slice[0] != "myvm" {
				t.Errorf("expected instanceID myvm, got %v", slice[0])
			}
			if slice[1] != "data" {
				t.Errorf("expected deviceName data, got %v", slice[1])
			}
			return nil, nil
		},
	}

	svc := NewVirtService(mock, Version{})
	err := svc.DeleteDevice(context.Background(), "myvm", "data")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVirtService_DeleteDevice_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("device delete failed")
		},
	}

	svc := NewVirtService(mock, Version{})
	err := svc.DeleteDevice(context.Background(), "myvm", "data")
	if err == nil {
		t.Fatal("expected error")
	}
}

// --- Conversion Function Tests ---

func TestVirtGlobalConfigFromResponse(t *testing.T) {
	bridge := "br0"
	v4 := "10.0.0.0/24"
	v6 := "fd00::/64"
	pool := "tank"
	resp := VirtGlobalConfigResponse{
		Bridge:    &bridge,
		V4Network: &v4,
		V6Network: &v6,
		Pool:      &pool,
	}

	cfg := virtGlobalConfigFromResponse(resp)
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

func TestVirtGlobalConfigFromResponse_NilFields(t *testing.T) {
	resp := VirtGlobalConfigResponse{}
	cfg := virtGlobalConfigFromResponse(resp)
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

func TestVirtGlobalConfigOptsToParams(t *testing.T) {
	bridge := "br1"
	pool := "tank2"
	params := virtGlobalConfigOptsToParams(UpdateVirtGlobalConfigOpts{
		Bridge: &bridge,
		Pool:   &pool,
	})
	if params["bridge"] != "br1" {
		t.Errorf("expected bridge br1, got %v", params["bridge"])
	}
	if params["pool"] != "tank2" {
		t.Errorf("expected pool tank2, got %v", params["pool"])
	}
	if _, ok := params["v4_network"]; ok {
		t.Error("v4_network should not be in params")
	}
	if _, ok := params["v6_network"]; ok {
		t.Error("v6_network should not be in params")
	}
}

func TestVirtInstanceFromResponse(t *testing.T) {
	cpu := "4"
	mem := int64(4294967296)
	netmask := int64(24)
	resp := VirtInstanceResponse{
		ID:        "inst1",
		Name:      "testvm",
		Type:      "CONTAINER",
		Status:    "STOPPED",
		CPU:       &cpu,
		Memory:    &mem,
		Autostart: false,
		Environment: map[string]string{"A": "1"},
		Aliases: []VirtInstanceAliasResponse{
			{Type: "INET", Address: "192.168.1.10", Netmask: &netmask},
		},
		Image: VirtInstanceImageResponse{
			Architecture: "amd64",
			Description:  "Alpine 3.18",
			OS:           "alpine",
			Release:      "3.18",
			Variant:      "default",
		},
		StoragePool: "pool1",
	}

	inst := virtInstanceFromResponse(resp)
	if inst.ID != "inst1" {
		t.Errorf("expected ID inst1, got %s", inst.ID)
	}
	if inst.Name != "testvm" {
		t.Errorf("expected name testvm, got %s", inst.Name)
	}
	if inst.Type != "CONTAINER" {
		t.Errorf("expected type CONTAINER, got %s", inst.Type)
	}
	if inst.CPU != "4" {
		t.Errorf("expected cpu 4, got %s", inst.CPU)
	}
	if inst.Memory != 4294967296 {
		t.Errorf("expected memory 4294967296, got %d", inst.Memory)
	}
	if len(inst.Aliases) != 1 {
		t.Fatalf("expected 1 alias, got %d", len(inst.Aliases))
	}
	if inst.Aliases[0].Netmask != 24 {
		t.Errorf("expected netmask 24, got %d", inst.Aliases[0].Netmask)
	}
	if inst.Image.OS != "alpine" {
		t.Errorf("expected image os alpine, got %s", inst.Image.OS)
	}
}

func TestVirtInstanceFromResponse_NoAliases(t *testing.T) {
	resp := VirtInstanceResponse{
		ID:     "inst2",
		Name:   "emptyvm",
		Type:   "VM",
		Status: "STOPPED",
	}

	inst := virtInstanceFromResponse(resp)
	if inst.CPU != "" {
		t.Errorf("expected empty cpu, got %s", inst.CPU)
	}
	if inst.Memory != 0 {
		t.Errorf("expected memory 0, got %d", inst.Memory)
	}
	if len(inst.Aliases) != 0 {
		t.Errorf("expected 0 aliases, got %d", len(inst.Aliases))
	}
	if inst.Environment == nil {
		t.Error("expected non-nil environment map")
	}
}

func TestVirtDeviceFromResponse_Disk(t *testing.T) {
	name := "data"
	desc := "Data volume"
	src := "/mnt/tank/data"
	dst := "/data"
	resp := VirtDeviceResponse{
		DevType:     "DISK",
		Name:        &name,
		Description: &desc,
		Readonly:    false,
		Source:      &src,
		Destination: &dst,
	}

	dev := virtDeviceFromResponse(resp)
	if dev.DevType != "DISK" {
		t.Errorf("expected dev_type DISK, got %s", dev.DevType)
	}
	if dev.Name != "data" {
		t.Errorf("expected name data, got %s", dev.Name)
	}
	if dev.Description != "Data volume" {
		t.Errorf("expected description Data volume, got %s", dev.Description)
	}
	if dev.Source != "/mnt/tank/data" {
		t.Errorf("expected source /mnt/tank/data, got %s", dev.Source)
	}
	if dev.Destination != "/data" {
		t.Errorf("expected destination /data, got %s", dev.Destination)
	}
}

func TestVirtDeviceFromResponse_NIC(t *testing.T) {
	name := "eth0"
	network := "br0"
	nicType := "BRIDGED"
	parent := "enp0s3"
	resp := VirtDeviceResponse{
		DevType: "NIC",
		Name:    &name,
		Network: &network,
		NICType: &nicType,
		Parent:  &parent,
	}

	dev := virtDeviceFromResponse(resp)
	if dev.DevType != "NIC" {
		t.Errorf("expected dev_type NIC, got %s", dev.DevType)
	}
	if dev.Network != "br0" {
		t.Errorf("expected network br0, got %s", dev.Network)
	}
	if dev.NICType != "BRIDGED" {
		t.Errorf("expected nic_type BRIDGED, got %s", dev.NICType)
	}
	if dev.Parent != "enp0s3" {
		t.Errorf("expected parent enp0s3, got %s", dev.Parent)
	}
}

func TestVirtDeviceFromResponse_Proxy(t *testing.T) {
	srcProto := "TCP"
	srcPort := int64(8080)
	dstProto := "TCP"
	dstPort := int64(80)
	resp := VirtDeviceResponse{
		DevType:     "PROXY",
		SourceProto: &srcProto,
		SourcePort:  &srcPort,
		DestProto:   &dstProto,
		DestPort:    &dstPort,
	}

	dev := virtDeviceFromResponse(resp)
	if dev.DevType != "PROXY" {
		t.Errorf("expected dev_type PROXY, got %s", dev.DevType)
	}
	if dev.SourceProto != "TCP" {
		t.Errorf("expected source_proto TCP, got %s", dev.SourceProto)
	}
	if dev.SourcePort != 8080 {
		t.Errorf("expected source_port 8080, got %d", dev.SourcePort)
	}
	if dev.DestProto != "TCP" {
		t.Errorf("expected dest_proto TCP, got %s", dev.DestProto)
	}
	if dev.DestPort != 80 {
		t.Errorf("expected dest_port 80, got %d", dev.DestPort)
	}
}

func TestVirtDeviceFromResponse_NilFields(t *testing.T) {
	resp := VirtDeviceResponse{
		DevType: "DISK",
	}

	dev := virtDeviceFromResponse(resp)
	if dev.Name != "" {
		t.Errorf("expected empty name, got %s", dev.Name)
	}
	if dev.Description != "" {
		t.Errorf("expected empty description, got %s", dev.Description)
	}
	if dev.Source != "" {
		t.Errorf("expected empty source, got %s", dev.Source)
	}
	if dev.Destination != "" {
		t.Errorf("expected empty destination, got %s", dev.Destination)
	}
	if dev.Network != "" {
		t.Errorf("expected empty network, got %s", dev.Network)
	}
	if dev.NICType != "" {
		t.Errorf("expected empty nic_type, got %s", dev.NICType)
	}
	if dev.Parent != "" {
		t.Errorf("expected empty parent, got %s", dev.Parent)
	}
	if dev.SourceProto != "" {
		t.Errorf("expected empty source_proto, got %s", dev.SourceProto)
	}
	if dev.SourcePort != 0 {
		t.Errorf("expected source_port 0, got %d", dev.SourcePort)
	}
	if dev.DestProto != "" {
		t.Errorf("expected empty dest_proto, got %s", dev.DestProto)
	}
	if dev.DestPort != 0 {
		t.Errorf("expected dest_port 0, got %d", dev.DestPort)
	}
}

func TestVirtDeviceOptToParam_Disk(t *testing.T) {
	m := virtDeviceOptToParam(VirtDeviceOpts{
		DevType:     "DISK",
		Name:        "data",
		Readonly:    true,
		Source:      "/mnt/tank/data",
		Destination: "/data",
	})
	if m["dev_type"] != "DISK" {
		t.Errorf("expected dev_type DISK, got %v", m["dev_type"])
	}
	if m["name"] != "data" {
		t.Errorf("expected name data, got %v", m["name"])
	}
	if m["readonly"] != true {
		t.Errorf("expected readonly true, got %v", m["readonly"])
	}
	if m["source"] != "/mnt/tank/data" {
		t.Errorf("expected source /mnt/tank/data, got %v", m["source"])
	}
	if m["destination"] != "/data" {
		t.Errorf("expected destination /data, got %v", m["destination"])
	}
	// Should not have NIC or PROXY fields
	if _, ok := m["network"]; ok {
		t.Error("network should not be in DISK params")
	}
	if _, ok := m["source_proto"]; ok {
		t.Error("source_proto should not be in DISK params")
	}
}

func TestVirtDeviceOptToParam_NIC(t *testing.T) {
	m := virtDeviceOptToParam(VirtDeviceOpts{
		DevType: "NIC",
		Network: "br0",
		NICType: "BRIDGED",
		Parent:  "enp0s3",
	})
	if m["dev_type"] != "NIC" {
		t.Errorf("expected dev_type NIC, got %v", m["dev_type"])
	}
	if m["network"] != "br0" {
		t.Errorf("expected network br0, got %v", m["network"])
	}
	if m["nic_type"] != "BRIDGED" {
		t.Errorf("expected nic_type BRIDGED, got %v", m["nic_type"])
	}
	if m["parent"] != "enp0s3" {
		t.Errorf("expected parent enp0s3, got %v", m["parent"])
	}
	// Should not have DISK or PROXY fields
	if _, ok := m["source"]; ok {
		t.Error("source should not be in NIC params")
	}
	if _, ok := m["source_proto"]; ok {
		t.Error("source_proto should not be in NIC params")
	}
}

func TestVirtDeviceOptToParam_Proxy(t *testing.T) {
	m := virtDeviceOptToParam(VirtDeviceOpts{
		DevType:     "PROXY",
		SourceProto: "TCP",
		SourcePort:  8080,
		DestProto:   "TCP",
		DestPort:    80,
	})
	if m["dev_type"] != "PROXY" {
		t.Errorf("expected dev_type PROXY, got %v", m["dev_type"])
	}
	if m["source_proto"] != "TCP" {
		t.Errorf("expected source_proto TCP, got %v", m["source_proto"])
	}
	if m["source_port"] != int64(8080) {
		t.Errorf("expected source_port 8080, got %v", m["source_port"])
	}
	if m["dest_proto"] != "TCP" {
		t.Errorf("expected dest_proto TCP, got %v", m["dest_proto"])
	}
	if m["dest_port"] != int64(80) {
		t.Errorf("expected dest_port 80, got %v", m["dest_port"])
	}
	// Should not have DISK or NIC fields
	if _, ok := m["source"]; ok {
		t.Error("source should not be in PROXY params")
	}
	if _, ok := m["network"]; ok {
		t.Error("network should not be in PROXY params")
	}
}

func TestVirtDeviceOptToParam_WithoutName(t *testing.T) {
	m := virtDeviceOptToParam(VirtDeviceOpts{
		DevType: "DISK",
		Source:  "/mnt/tank/data",
		Destination: "/data",
	})
	if _, ok := m["name"]; ok {
		t.Error("name should not be in params when empty")
	}
}

func TestVirtDeviceOptToParam_WithName(t *testing.T) {
	m := virtDeviceOptToParam(VirtDeviceOpts{
		DevType: "DISK",
		Name:    "mydevice",
		Source:  "/mnt/tank/data",
		Destination: "/data",
	})
	if m["name"] != "mydevice" {
		t.Errorf("expected name mydevice, got %v", m["name"])
	}
}

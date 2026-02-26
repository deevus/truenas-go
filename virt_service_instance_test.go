package truenas

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

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
		Name:         "myvm",
		InstanceType: "VM",
		Image:        "ubuntu/22.04",
		CPU:          "2",
		Memory:       2147483648,
		Autostart:    true,
		Environment:  map[string]string{"FOO": "bar", "BAZ": "qux"},
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
		Name:         "myvm",
		InstanceType: "VM",
		Image:        "ubuntu/22.04",
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
		Name:         "myvm",
		InstanceType: "VM",
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

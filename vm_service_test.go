package truenas

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

// --- Sample fixtures ---

func sampleVMJSON() json.RawMessage {
	return json.RawMessage(`{
		"id": 1,
		"name": "test-vm",
		"description": "A test VM",
		"vcpus": 1,
		"cores": 2,
		"threads": 1,
		"memory": 2048,
		"min_memory": null,
		"autostart": true,
		"time": "LOCAL",
		"bootloader": "UEFI",
		"bootloader_ovmf": "OVMF_CODE.fd",
		"cpu_mode": "HOST-MODEL",
		"cpu_model": null,
		"shutdown_timeout": 90,
		"command_line_args": "",
		"status": {
			"state": "RUNNING",
			"pid": 12345,
			"domain_state": "RUNNING"
		}
	}`)
}

func sampleVMWithCPUModelJSON() json.RawMessage {
	return json.RawMessage(`{
		"id": 2,
		"name": "model-vm",
		"description": "VM with CPU model",
		"vcpus": 2,
		"cores": 4,
		"threads": 2,
		"memory": 4096,
		"min_memory": 2048,
		"autostart": false,
		"time": "UTC",
		"bootloader": "UEFI",
		"bootloader_ovmf": "OVMF_CODE.fd",
		"cpu_mode": "CUSTOM",
		"cpu_model": "Haswell",
		"shutdown_timeout": 60,
		"command_line_args": "-cpu host",
		"status": {
			"state": "STOPPED",
			"pid": null,
			"domain_state": "SHUTOFF"
		}
	}`)
}

func sampleDiskDeviceJSON() json.RawMessage {
	return json.RawMessage(`{
		"id": 10,
		"vm": 1,
		"order": 1001,
		"attributes": {
			"dtype": "DISK",
			"path": "/dev/zvol/tank/vm-disk",
			"type": "VIRTIO",
			"physical_sectorsize": null,
			"logical_sectorsize": null
		}
	}`)
}

func sampleRawDeviceJSON() json.RawMessage {
	return json.RawMessage(`{
		"id": 11,
		"vm": 1,
		"order": 1002,
		"attributes": {
			"dtype": "RAW",
			"path": "/mnt/tank/vm/raw.img",
			"type": "VIRTIO",
			"boot": true,
			"size": 10737418240,
			"physical_sectorsize": null,
			"logical_sectorsize": null
		}
	}`)
}

func sampleCDROMDeviceJSON() json.RawMessage {
	return json.RawMessage(`{
		"id": 12,
		"vm": 1,
		"order": 1003,
		"attributes": {
			"dtype": "CDROM",
			"path": "/mnt/tank/iso/ubuntu.iso"
		}
	}`)
}

func sampleNICDeviceJSON() json.RawMessage {
	return json.RawMessage(`{
		"id": 13,
		"vm": 1,
		"order": 1004,
		"attributes": {
			"dtype": "NIC",
			"type": "VIRTIO",
			"nic_attach": "br0",
			"mac": "00:a0:98:6b:0c:01",
			"trust_guest_rx_filters": false
		}
	}`)
}

func sampleDisplayDeviceJSON() json.RawMessage {
	return json.RawMessage(`{
		"id": 14,
		"vm": 1,
		"order": 1005,
		"attributes": {
			"dtype": "DISPLAY",
			"type": "SPICE",
			"port": 5900,
			"bind": "0.0.0.0",
			"password": "secret",
			"web": true,
			"resolution": "1024x768",
			"wait": false
		}
	}`)
}

func samplePCIDeviceJSON() json.RawMessage {
	return json.RawMessage(`{
		"id": 15,
		"vm": 1,
		"order": 1006,
		"attributes": {
			"dtype": "PCI",
			"pptdev": "pci_0000_01_00_0"
		}
	}`)
}

func sampleUSBDeviceJSON() json.RawMessage {
	return json.RawMessage(`{
		"id": 16,
		"vm": 1,
		"order": 1007,
		"attributes": {
			"dtype": "USB",
			"controller_type": "nec-xhci",
			"device": "usb_0_1_2",
			"usb_speed": "HIGH"
		}
	}`)
}

// --- VM CRUD Tests ---

func TestVMService_CreateVM(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				if method != "vm.create" {
					t.Errorf("expected method vm.create, got %s", method)
				}
				p := params.(map[string]any)
				if p["name"] != "test-vm" {
					t.Errorf("expected name test-vm, got %v", p["name"])
				}
				if p["memory"] != int64(2048) {
					t.Errorf("expected memory 2048, got %v", p["memory"])
				}
				return sampleVMJSON(), nil
			},
		},
	}

	svc := NewVMService(mock, Version{})
	vm, err := svc.CreateVM(context.Background(), CreateVMOpts{
		Name:            "test-vm",
		Description:     "A test VM",
		VCPUs:           1,
		Cores:           2,
		Threads:         1,
		Memory:          2048,
		Autostart:       true,
		Time:            "LOCAL",
		Bootloader:      "UEFI",
		BootloaderOVMF:  "OVMF_CODE.fd",
		CPUMode:         "HOST-MODEL",
		ShutdownTimeout: 90,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vm == nil {
		t.Fatal("expected non-nil vm")
	}
	if vm.ID != 1 {
		t.Errorf("expected ID 1, got %d", vm.ID)
	}
	if vm.Name != "test-vm" {
		t.Errorf("expected name test-vm, got %s", vm.Name)
	}
	if vm.State != "RUNNING" {
		t.Errorf("expected state RUNNING, got %s", vm.State)
	}
	if vm.CPUModel != "" {
		t.Errorf("expected empty cpu_model, got %s", vm.CPUModel)
	}
	// Verify only one call was made (no re-read)
	if len(mock.calls) != 1 {
		t.Errorf("expected 1 call (no re-read), got %d", len(mock.calls))
	}
}

func TestVMService_CreateVM_WithMinMemory(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				p := params.(map[string]any)
				if _, ok := p["min_memory"]; !ok {
					t.Error("expected min_memory to be set in params")
				}
				if p["min_memory"] != int64(1024) {
					t.Errorf("expected min_memory 1024, got %v", p["min_memory"])
				}
				return sampleVMJSON(), nil
			},
		},
	}

	minMem := int64(1024)
	svc := NewVMService(mock, Version{})
	_, err := svc.CreateVM(context.Background(), CreateVMOpts{
		Name:      "test-vm",
		Memory:    2048,
		MinMemory: &minMem,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVMService_CreateVM_WithCPUModel(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				p := params.(map[string]any)
				if p["cpu_model"] != "Haswell" {
					t.Errorf("expected cpu_model Haswell, got %v", p["cpu_model"])
				}
				return sampleVMWithCPUModelJSON(), nil
			},
		},
	}

	svc := NewVMService(mock, Version{})
	vm, err := svc.CreateVM(context.Background(), CreateVMOpts{
		Name:     "model-vm",
		CPUMode:  "CUSTOM",
		CPUModel: "Haswell",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vm.CPUModel != "Haswell" {
		t.Errorf("expected CPUModel Haswell, got %s", vm.CPUModel)
	}
}

func TestVMService_CreateVM_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, errors.New("connection refused")
			},
		},
	}

	svc := NewVMService(mock, Version{})
	vm, err := svc.CreateVM(context.Background(), CreateVMOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
	if vm != nil {
		t.Error("expected nil vm on error")
	}
}

func TestVMService_CreateVM_ParseError(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return json.RawMessage(`not json`), nil
			},
		},
	}

	svc := NewVMService(mock, Version{})
	_, err := svc.CreateVM(context.Background(), CreateVMOpts{})
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestVMService_GetVM(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				if method != "vm.get_instance" {
					t.Errorf("expected method vm.get_instance, got %s", method)
				}
				id, ok := params.(int64)
				if !ok || id != 1 {
					t.Errorf("expected id 1, got %v", params)
				}
				return sampleVMJSON(), nil
			},
		},
	}

	svc := NewVMService(mock, Version{})
	vm, err := svc.GetVM(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vm == nil {
		t.Fatal("expected non-nil vm")
	}
	if vm.ID != 1 {
		t.Errorf("expected ID 1, got %d", vm.ID)
	}
	if vm.Name != "test-vm" {
		t.Errorf("expected name test-vm, got %s", vm.Name)
	}
	if vm.Description != "A test VM" {
		t.Errorf("expected description 'A test VM', got %q", vm.Description)
	}
	if vm.Memory != 2048 {
		t.Errorf("expected memory 2048, got %d", vm.Memory)
	}
	if vm.Autostart != true {
		t.Error("expected autostart true")
	}
	if vm.State != "RUNNING" {
		t.Errorf("expected state RUNNING, got %s", vm.State)
	}
}

func TestVMService_GetVM_NullCPUModel(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return sampleVMJSON(), nil // cpu_model is null
			},
		},
	}

	svc := NewVMService(mock, Version{})
	vm, err := svc.GetVM(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vm.CPUModel != "" {
		t.Errorf("expected empty CPUModel for null, got %q", vm.CPUModel)
	}
}

func TestVMService_GetVM_NullMinMemory(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return sampleVMJSON(), nil // min_memory is null
			},
		},
	}

	svc := NewVMService(mock, Version{})
	vm, err := svc.GetVM(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vm.MinMemory != nil {
		t.Errorf("expected nil MinMemory, got %v", vm.MinMemory)
	}
}

func TestVMService_GetVM_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, errors.New("timeout")
			},
		},
	}

	svc := NewVMService(mock, Version{})
	_, err := svc.GetVM(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestVMService_GetVM_ParseError(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return json.RawMessage(`not json`), nil
			},
		},
	}

	svc := NewVMService(mock, Version{})
	_, err := svc.GetVM(context.Background(), 1)
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestVMService_UpdateVM(t *testing.T) {
	callCount := 0
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				callCount++
				if callCount == 1 {
					if method != "vm.update" {
						t.Errorf("expected method vm.update, got %s", method)
					}
					slice, ok := params.([]any)
					if !ok {
						t.Fatal("expected []any params for update")
					}
					if len(slice) != 2 {
						t.Fatalf("expected 2 elements, got %d", len(slice))
					}
					id, ok := slice[0].(int64)
					if !ok || id != 1 {
						t.Errorf("expected id 1, got %v", slice[0])
					}
					return json.RawMessage(`{"id": 1}`), nil
				}
				// Re-read via GetVM
				if method != "vm.get_instance" {
					t.Errorf("expected method vm.get_instance for re-read, got %s", method)
				}
				return sampleVMJSON(), nil
			},
		},
	}

	svc := NewVMService(mock, Version{})
	vm, err := svc.UpdateVM(context.Background(), 1, UpdateVMOpts{
		Name:   "test-vm",
		Memory: 2048,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vm == nil {
		t.Fatal("expected non-nil vm")
	}
	if vm.ID != 1 {
		t.Errorf("expected ID 1, got %d", vm.ID)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls (update + re-read), got %d", callCount)
	}
}

func TestVMService_UpdateVM_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, errors.New("not found")
			},
		},
	}

	svc := NewVMService(mock, Version{})
	_, err := svc.UpdateVM(context.Background(), 999, UpdateVMOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestVMService_UpdateVM_ReReadError(t *testing.T) {
	callCount := 0
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				callCount++
				if callCount == 1 {
					return json.RawMessage(`{"id": 1}`), nil
				}
				return nil, errors.New("re-read failed")
			},
		},
	}

	svc := NewVMService(mock, Version{})
	_, err := svc.UpdateVM(context.Background(), 1, UpdateVMOpts{})
	if err == nil {
		t.Fatal("expected error on re-read")
	}
	if err.Error() != "re-read failed" {
		t.Errorf("expected 're-read failed', got %q", err.Error())
	}
}

func TestVMService_DeleteVM(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				if method != "vm.delete" {
					t.Errorf("expected method vm.delete, got %s", method)
				}
				id, ok := params.(int64)
				if !ok || id != 5 {
					t.Errorf("expected id 5, got %v", params)
				}
				return nil, nil
			},
		},
	}

	svc := NewVMService(mock, Version{})
	err := svc.DeleteVM(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVMService_DeleteVM_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, errors.New("permission denied")
			},
		},
	}

	svc := NewVMService(mock, Version{})
	err := svc.DeleteVM(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestVMService_StartVM(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				if method != "vm.start" {
					t.Errorf("expected method vm.start, got %s", method)
				}
				id, ok := params.(int64)
				if !ok || id != 1 {
					t.Errorf("expected id 1, got %v", params)
				}
				return nil, nil
			},
		},
	}

	svc := NewVMService(mock, Version{})
	err := svc.StartVM(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Verify it used Call, not CallAndWait
	if len(mock.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(mock.calls))
	}
	if mock.calls[0].Method != "vm.start" {
		t.Errorf("expected method vm.start, got %s", mock.calls[0].Method)
	}
}

func TestVMService_StartVM_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, errors.New("vm already running")
			},
		},
	}

	svc := NewVMService(mock, Version{})
	err := svc.StartVM(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestVMService_StopVM(t *testing.T) {
	mock := &mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "vm.stop" {
				t.Errorf("expected method vm.stop, got %s", method)
			}
			slice, ok := params.([]any)
			if !ok {
				t.Fatal("expected []any params")
			}
			if len(slice) != 2 {
				t.Fatalf("expected 2 elements, got %d", len(slice))
			}
			id, ok := slice[0].(int64)
			if !ok || id != 1 {
				t.Errorf("expected id 1, got %v", slice[0])
			}
			stopParams, ok := slice[1].(map[string]any)
			if !ok {
				t.Fatal("expected map[string]any for stop params")
			}
			if stopParams["force"] != false {
				t.Errorf("expected force=false, got %v", stopParams["force"])
			}
			if stopParams["force_after_timeout"] != false {
				t.Errorf("expected force_after_timeout=false, got %v", stopParams["force_after_timeout"])
			}
			return nil, nil
		},
	}

	svc := NewVMService(mock, Version{})
	err := svc.StopVM(context.Background(), 1, StopVMOpts{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Verify it used CallAndWait
	if len(mock.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(mock.calls))
	}
	if mock.calls[0].Method != "vm.stop" {
		t.Errorf("expected method vm.stop, got %s", mock.calls[0].Method)
	}
}

func TestVMService_StopVM_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("vm not running")
		},
	}

	svc := NewVMService(mock, Version{})
	err := svc.StopVM(context.Background(), 1, StopVMOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestVMService_StopVM_ForceOptions(t *testing.T) {
	mock := &mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			slice := params.([]any)
			stopParams := slice[1].(map[string]any)
			if stopParams["force"] != true {
				t.Errorf("expected force=true, got %v", stopParams["force"])
			}
			if stopParams["force_after_timeout"] != true {
				t.Errorf("expected force_after_timeout=true, got %v", stopParams["force_after_timeout"])
			}
			return nil, nil
		},
	}

	svc := NewVMService(mock, Version{})
	err := svc.StopVM(context.Background(), 1, StopVMOpts{
		Force:             true,
		ForceAfterTimeout: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- Device CRUD Tests ---

func TestVMService_ListDevices(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				if method != "vm.device.query" {
					t.Errorf("expected method vm.device.query, got %s", method)
				}
				// Verify filter shape
				filter, ok := params.([]any)
				if !ok {
					t.Fatal("expected []any params")
				}
				if len(filter) != 1 {
					t.Fatalf("expected 1 outer element, got %d", len(filter))
				}
				return json.RawMessage(`[` +
					string(sampleDiskDeviceJSON()) + `,` +
					string(sampleNICDeviceJSON()) +
					`]`), nil
			},
		},
	}

	svc := NewVMService(mock, Version{})
	devices, err := svc.ListDevices(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(devices) != 2 {
		t.Fatalf("expected 2 devices, got %d", len(devices))
	}
	if devices[0].DeviceType != DeviceTypeDisk {
		t.Errorf("expected first device type DISK, got %s", devices[0].DeviceType)
	}
	if devices[1].DeviceType != DeviceTypeNIC {
		t.Errorf("expected second device type NIC, got %s", devices[1].DeviceType)
	}
}

func TestVMService_ListDevices_Empty(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return json.RawMessage(`[]`), nil
			},
		},
	}

	svc := NewVMService(mock, Version{})
	devices, err := svc.ListDevices(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(devices) != 0 {
		t.Errorf("expected 0 devices, got %d", len(devices))
	}
}

func TestVMService_ListDevices_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, errors.New("network error")
			},
		},
	}

	svc := NewVMService(mock, Version{})
	_, err := svc.ListDevices(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestVMService_ListDevices_ParseError(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return json.RawMessage(`not json`), nil
			},
		},
	}

	svc := NewVMService(mock, Version{})
	_, err := svc.ListDevices(context.Background(), 1)
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestVMService_GetDevice(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				if method != "vm.device.query" {
					t.Errorf("expected method vm.device.query, got %s", method)
				}
				return json.RawMessage(`[` + string(sampleDiskDeviceJSON()) + `]`), nil
			},
		},
	}

	svc := NewVMService(mock, Version{})
	device, err := svc.GetDevice(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if device == nil {
		t.Fatal("expected non-nil device")
	}
	if device.ID != 10 {
		t.Errorf("expected ID 10, got %d", device.ID)
	}
	if device.DeviceType != DeviceTypeDisk {
		t.Errorf("expected type DISK, got %s", device.DeviceType)
	}
	if device.Disk == nil {
		t.Fatal("expected non-nil Disk")
	}
	if device.Disk.Path != "/dev/zvol/tank/vm-disk" {
		t.Errorf("expected path /dev/zvol/tank/vm-disk, got %s", device.Disk.Path)
	}
}

func TestVMService_GetDevice_NotFound(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return json.RawMessage(`[]`), nil
			},
		},
	}

	svc := NewVMService(mock, Version{})
	device, err := svc.GetDevice(context.Background(), 999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if device != nil {
		t.Error("expected nil device for not found")
	}
}

func TestVMService_GetDevice_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, errors.New("timeout")
			},
		},
	}

	svc := NewVMService(mock, Version{})
	_, err := svc.GetDevice(context.Background(), 10)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestVMService_GetDevice_ParseError(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return json.RawMessage(`not json`), nil
			},
		},
	}

	svc := NewVMService(mock, Version{})
	_, err := svc.GetDevice(context.Background(), 10)
	if err == nil {
		t.Fatal("expected parse error")
	}
}

// --- Create Device for all 7 types ---

func TestVMService_CreateDevice_Disk(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				if method != "vm.device.create" {
					t.Errorf("expected method vm.device.create, got %s", method)
				}
				p := params.(map[string]any)
				attrs := p["attributes"].(map[string]any)
				if attrs["dtype"] != "DISK" {
					t.Errorf("expected dtype DISK, got %v", attrs["dtype"])
				}
				if attrs["path"] != "/dev/zvol/tank/vm-disk" {
					t.Errorf("expected path, got %v", attrs["path"])
				}
				return sampleDiskDeviceJSON(), nil
			},
		},
	}

	svc := NewVMService(mock, Version{})
	device, err := svc.CreateDevice(context.Background(), CreateVMDeviceOpts{
		VM:         1,
		DeviceType: DeviceTypeDisk,
		Disk: &DiskDevice{
			Path: "/dev/zvol/tank/vm-disk",
			Type: "VIRTIO",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if device.DeviceType != DeviceTypeDisk {
		t.Errorf("expected DISK, got %s", device.DeviceType)
	}
	if device.Disk == nil {
		t.Fatal("expected non-nil Disk")
	}
}

func TestVMService_CreateDevice_Raw(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				p := params.(map[string]any)
				attrs := p["attributes"].(map[string]any)
				if attrs["dtype"] != "RAW" {
					t.Errorf("expected dtype RAW, got %v", attrs["dtype"])
				}
				if attrs["boot"] != true {
					t.Errorf("expected boot=true, got %v", attrs["boot"])
				}
				return sampleRawDeviceJSON(), nil
			},
		},
	}

	size := int64(10737418240)
	svc := NewVMService(mock, Version{})
	device, err := svc.CreateDevice(context.Background(), CreateVMDeviceOpts{
		VM:         1,
		DeviceType: DeviceTypeRaw,
		Raw: &RawDevice{
			Path: "/mnt/tank/vm/raw.img",
			Type: "VIRTIO",
			Boot: true,
			Size: &size,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if device.DeviceType != DeviceTypeRaw {
		t.Errorf("expected RAW, got %s", device.DeviceType)
	}
	if device.Raw == nil {
		t.Fatal("expected non-nil Raw")
	}
	if !device.Raw.Boot {
		t.Error("expected boot=true")
	}
}

func TestVMService_CreateDevice_CDROM(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				p := params.(map[string]any)
				attrs := p["attributes"].(map[string]any)
				if attrs["dtype"] != "CDROM" {
					t.Errorf("expected dtype CDROM, got %v", attrs["dtype"])
				}
				return sampleCDROMDeviceJSON(), nil
			},
		},
	}

	svc := NewVMService(mock, Version{})
	device, err := svc.CreateDevice(context.Background(), CreateVMDeviceOpts{
		VM:         1,
		DeviceType: DeviceTypeCDROM,
		CDROM: &CDROMDevice{
			Path: "/mnt/tank/iso/ubuntu.iso",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if device.DeviceType != DeviceTypeCDROM {
		t.Errorf("expected CDROM, got %s", device.DeviceType)
	}
	if device.CDROM == nil {
		t.Fatal("expected non-nil CDROM")
	}
	if device.CDROM.Path != "/mnt/tank/iso/ubuntu.iso" {
		t.Errorf("expected path, got %s", device.CDROM.Path)
	}
}

func TestVMService_CreateDevice_NIC(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				p := params.(map[string]any)
				attrs := p["attributes"].(map[string]any)
				if attrs["dtype"] != "NIC" {
					t.Errorf("expected dtype NIC, got %v", attrs["dtype"])
				}
				if attrs["trust_guest_rx_filters"] != false {
					t.Errorf("expected trust_guest_rx_filters=false, got %v", attrs["trust_guest_rx_filters"])
				}
				return sampleNICDeviceJSON(), nil
			},
		},
	}

	svc := NewVMService(mock, Version{})
	device, err := svc.CreateDevice(context.Background(), CreateVMDeviceOpts{
		VM:         1,
		DeviceType: DeviceTypeNIC,
		NIC: &NICDevice{
			Type:      "VIRTIO",
			NICAttach: "br0",
			MAC:       "00:a0:98:6b:0c:01",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if device.DeviceType != DeviceTypeNIC {
		t.Errorf("expected NIC, got %s", device.DeviceType)
	}
	if device.NIC == nil {
		t.Fatal("expected non-nil NIC")
	}
	if device.NIC.MAC != "00:a0:98:6b:0c:01" {
		t.Errorf("expected MAC, got %s", device.NIC.MAC)
	}
}

func TestVMService_CreateDevice_Display(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				p := params.(map[string]any)
				attrs := p["attributes"].(map[string]any)
				if attrs["dtype"] != "DISPLAY" {
					t.Errorf("expected dtype DISPLAY, got %v", attrs["dtype"])
				}
				if attrs["web"] != true {
					t.Errorf("expected web=true, got %v", attrs["web"])
				}
				if attrs["wait"] != false {
					t.Errorf("expected wait=false, got %v", attrs["wait"])
				}
				return sampleDisplayDeviceJSON(), nil
			},
		},
	}

	svc := NewVMService(mock, Version{})
	device, err := svc.CreateDevice(context.Background(), CreateVMDeviceOpts{
		VM:         1,
		DeviceType: DeviceTypeDisplay,
		Display: &DisplayDevice{
			Type:       "SPICE",
			Port:       5900,
			Bind:       "0.0.0.0",
			Password:   "secret",
			Web:        true,
			Resolution: "1024x768",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if device.DeviceType != DeviceTypeDisplay {
		t.Errorf("expected DISPLAY, got %s", device.DeviceType)
	}
	if device.Display == nil {
		t.Fatal("expected non-nil Display")
	}
	if device.Display.Password != "secret" {
		t.Errorf("expected password 'secret', got %s", device.Display.Password)
	}
}

func TestVMService_CreateDevice_PCI(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				p := params.(map[string]any)
				attrs := p["attributes"].(map[string]any)
				if attrs["dtype"] != "PCI" {
					t.Errorf("expected dtype PCI, got %v", attrs["dtype"])
				}
				return samplePCIDeviceJSON(), nil
			},
		},
	}

	svc := NewVMService(mock, Version{})
	device, err := svc.CreateDevice(context.Background(), CreateVMDeviceOpts{
		VM:         1,
		DeviceType: DeviceTypePCI,
		PCI: &PCIDevice{
			PPTDev: "pci_0000_01_00_0",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if device.DeviceType != DeviceTypePCI {
		t.Errorf("expected PCI, got %s", device.DeviceType)
	}
	if device.PCI == nil {
		t.Fatal("expected non-nil PCI")
	}
	if device.PCI.PPTDev != "pci_0000_01_00_0" {
		t.Errorf("expected pptdev, got %s", device.PCI.PPTDev)
	}
}

func TestVMService_CreateDevice_USB(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				p := params.(map[string]any)
				attrs := p["attributes"].(map[string]any)
				if attrs["dtype"] != "USB" {
					t.Errorf("expected dtype USB, got %v", attrs["dtype"])
				}
				return sampleUSBDeviceJSON(), nil
			},
		},
	}

	svc := NewVMService(mock, Version{})
	device, err := svc.CreateDevice(context.Background(), CreateVMDeviceOpts{
		VM:         1,
		DeviceType: DeviceTypeUSB,
		USB: &USBDevice{
			ControllerType: "nec-xhci",
			Device:         "usb_0_1_2",
			USBSpeed:       "HIGH",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if device.DeviceType != DeviceTypeUSB {
		t.Errorf("expected USB, got %s", device.DeviceType)
	}
	if device.USB == nil {
		t.Fatal("expected non-nil USB")
	}
	if device.USB.ControllerType != "nec-xhci" {
		t.Errorf("expected controller_type nec-xhci, got %s", device.USB.ControllerType)
	}
	if device.USB.Device != "usb_0_1_2" {
		t.Errorf("expected device usb_0_1_2, got %v", device.USB.Device)
	}
}

func TestVMService_CreateDevice_WithOrder(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				p := params.(map[string]any)
				if p["order"] != int64(1001) {
					t.Errorf("expected order 1001, got %v", p["order"])
				}
				return sampleDiskDeviceJSON(), nil
			},
		},
	}

	order := int64(1001)
	svc := NewVMService(mock, Version{})
	_, err := svc.CreateDevice(context.Background(), CreateVMDeviceOpts{
		VM:         1,
		Order:      &order,
		DeviceType: DeviceTypeDisk,
		Disk: &DiskDevice{
			Path: "/dev/zvol/tank/vm-disk",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVMService_CreateDevice_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, errors.New("validation error")
			},
		},
	}

	svc := NewVMService(mock, Version{})
	device, err := svc.CreateDevice(context.Background(), CreateVMDeviceOpts{
		VM:         1,
		DeviceType: DeviceTypeDisk,
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if device != nil {
		t.Error("expected nil device on error")
	}
}

func TestVMService_CreateDevice_ParseError(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return json.RawMessage(`not json`), nil
			},
		},
	}

	svc := NewVMService(mock, Version{})
	_, err := svc.CreateDevice(context.Background(), CreateVMDeviceOpts{
		VM:         1,
		DeviceType: DeviceTypeDisk,
	})
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestVMService_UpdateDevice(t *testing.T) {
	callCount := 0
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				callCount++
				if callCount == 1 {
					if method != "vm.device.update" {
						t.Errorf("expected method vm.device.update, got %s", method)
					}
					slice, ok := params.([]any)
					if !ok {
						t.Fatal("expected []any params")
					}
					if len(slice) != 2 {
						t.Fatalf("expected 2 elements, got %d", len(slice))
					}
					id, ok := slice[0].(int64)
					if !ok || id != 10 {
						t.Errorf("expected id 10, got %v", slice[0])
					}
					return json.RawMessage(`{"id": 10}`), nil
				}
				// Re-read via GetDevice
				if method != "vm.device.query" {
					t.Errorf("expected method vm.device.query for re-read, got %s", method)
				}
				return json.RawMessage(`[` + string(sampleDiskDeviceJSON()) + `]`), nil
			},
		},
	}

	svc := NewVMService(mock, Version{})
	device, err := svc.UpdateDevice(context.Background(), 10, UpdateVMDeviceOpts{
		VM:         1,
		DeviceType: DeviceTypeDisk,
		Disk: &DiskDevice{
			Path: "/dev/zvol/tank/vm-disk",
			Type: "VIRTIO",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if device == nil {
		t.Fatal("expected non-nil device")
	}
	if device.ID != 10 {
		t.Errorf("expected ID 10, got %d", device.ID)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls (update + re-read), got %d", callCount)
	}
}

func TestVMService_UpdateDevice_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, errors.New("not found")
			},
		},
	}

	svc := NewVMService(mock, Version{})
	_, err := svc.UpdateDevice(context.Background(), 999, UpdateVMDeviceOpts{
		VM:         1,
		DeviceType: DeviceTypeDisk,
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestVMService_UpdateDevice_ReReadError(t *testing.T) {
	callCount := 0
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				callCount++
				if callCount == 1 {
					return json.RawMessage(`{"id": 10}`), nil
				}
				return nil, errors.New("re-read failed")
			},
		},
	}

	svc := NewVMService(mock, Version{})
	_, err := svc.UpdateDevice(context.Background(), 10, UpdateVMDeviceOpts{
		VM:         1,
		DeviceType: DeviceTypeDisk,
	})
	if err == nil {
		t.Fatal("expected error on re-read")
	}
}

func TestVMService_DeleteDevice(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				if method != "vm.device.delete" {
					t.Errorf("expected method vm.device.delete, got %s", method)
				}
				id, ok := params.(int64)
				if !ok || id != 10 {
					t.Errorf("expected id 10, got %v", params)
				}
				return nil, nil
			},
		},
	}

	svc := NewVMService(mock, Version{})
	err := svc.DeleteDevice(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVMService_DeleteDevice_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, errors.New("permission denied")
			},
		},
	}

	svc := NewVMService(mock, Version{})
	err := svc.DeleteDevice(context.Background(), 10)
	if err == nil {
		t.Fatal("expected error")
	}
}

// --- Conversion Tests ---

func TestVMFromResponse(t *testing.T) {
	minMem := int64(2048)
	cpuModel := "Haswell"
	pid := int64(12345)

	resp := VMResponse{
		ID:              2,
		Name:            "model-vm",
		Description:     "VM with CPU model",
		VCPUs:           2,
		Cores:           4,
		Threads:         2,
		Memory:          4096,
		MinMemory:       &minMem,
		Autostart:       false,
		Time:            "UTC",
		Bootloader:      "UEFI",
		BootloaderOVMF:  "OVMF_CODE.fd",
		CPUMode:         "CUSTOM",
		CPUModel:        &cpuModel,
		ShutdownTimeout: 60,
		CommandLineArgs: "-cpu host",
		Status: VMStatusField{
			State:       "RUNNING",
			PID:         &pid,
			DomainState: "RUNNING",
		},
	}

	vm := vmFromResponse(resp)

	if vm.ID != 2 {
		t.Errorf("expected ID 2, got %d", vm.ID)
	}
	if vm.Name != "model-vm" {
		t.Errorf("expected name model-vm, got %s", vm.Name)
	}
	if vm.CPUModel != "Haswell" {
		t.Errorf("expected CPUModel Haswell, got %s", vm.CPUModel)
	}
	if vm.MinMemory == nil || *vm.MinMemory != 2048 {
		t.Errorf("expected MinMemory 2048, got %v", vm.MinMemory)
	}
	if vm.State != "RUNNING" {
		t.Errorf("expected state RUNNING, got %s", vm.State)
	}
	if vm.CommandLineArgs != "-cpu host" {
		t.Errorf("expected command_line_args '-cpu host', got %s", vm.CommandLineArgs)
	}
}

func TestVMFromResponse_NullCPUModel(t *testing.T) {
	resp := VMResponse{
		ID:       1,
		Name:     "test",
		CPUModel: nil,
		Status:   VMStatusField{State: "STOPPED"},
	}

	vm := vmFromResponse(resp)
	if vm.CPUModel != "" {
		t.Errorf("expected empty CPUModel for nil, got %q", vm.CPUModel)
	}
}

func TestVMDeviceFromResponse_Disk(t *testing.T) {
	resp := VMDeviceResponse{
		ID:    10,
		VM:    1,
		Order: 1001,
		Attributes: map[string]any{
			"dtype":                "DISK",
			"path":                 "/dev/zvol/tank/vm-disk",
			"type":                 "VIRTIO",
			"physical_sectorsize":  nil,
			"logical_sectorsize":   nil,
		},
	}

	device := vmDeviceFromResponse(resp)
	if device.DeviceType != DeviceTypeDisk {
		t.Errorf("expected DISK, got %s", device.DeviceType)
	}
	if device.Disk == nil {
		t.Fatal("expected non-nil Disk")
	}
	if device.Disk.Path != "/dev/zvol/tank/vm-disk" {
		t.Errorf("expected path, got %s", device.Disk.Path)
	}
	if device.Disk.Type != "VIRTIO" {
		t.Errorf("expected type VIRTIO, got %s", device.Disk.Type)
	}
	if device.Disk.PhysicalSectorSize != nil {
		t.Errorf("expected nil PhysicalSectorSize, got %v", device.Disk.PhysicalSectorSize)
	}
}

func TestVMDeviceFromResponse_Raw(t *testing.T) {
	resp := VMDeviceResponse{
		ID:    11,
		VM:    1,
		Order: 1002,
		Attributes: map[string]any{
			"dtype": "RAW",
			"path":  "/mnt/tank/vm/raw.img",
			"type":  "VIRTIO",
			"boot":  true,
			"size":  float64(10737418240),
		},
	}

	device := vmDeviceFromResponse(resp)
	if device.DeviceType != DeviceTypeRaw {
		t.Errorf("expected RAW, got %s", device.DeviceType)
	}
	if device.Raw == nil {
		t.Fatal("expected non-nil Raw")
	}
	if !device.Raw.Boot {
		t.Error("expected boot=true")
	}
	if device.Raw.Size == nil || *device.Raw.Size != 10737418240 {
		t.Errorf("expected size 10737418240, got %v", device.Raw.Size)
	}
}

func TestVMDeviceFromResponse_CDROM(t *testing.T) {
	resp := VMDeviceResponse{
		ID:    12,
		VM:    1,
		Order: 1003,
		Attributes: map[string]any{
			"dtype": "CDROM",
			"path":  "/mnt/tank/iso/ubuntu.iso",
		},
	}

	device := vmDeviceFromResponse(resp)
	if device.DeviceType != DeviceTypeCDROM {
		t.Errorf("expected CDROM, got %s", device.DeviceType)
	}
	if device.CDROM == nil {
		t.Fatal("expected non-nil CDROM")
	}
	if device.CDROM.Path != "/mnt/tank/iso/ubuntu.iso" {
		t.Errorf("expected path, got %s", device.CDROM.Path)
	}
}

func TestVMDeviceFromResponse_NIC(t *testing.T) {
	resp := VMDeviceResponse{
		ID:    13,
		VM:    1,
		Order: 1004,
		Attributes: map[string]any{
			"dtype":                    "NIC",
			"type":                     "VIRTIO",
			"nic_attach":              "br0",
			"mac":                      "00:a0:98:6b:0c:01",
			"trust_guest_rx_filters":  false,
		},
	}

	device := vmDeviceFromResponse(resp)
	if device.DeviceType != DeviceTypeNIC {
		t.Errorf("expected NIC, got %s", device.DeviceType)
	}
	if device.NIC == nil {
		t.Fatal("expected non-nil NIC")
	}
	if device.NIC.NICAttach != "br0" {
		t.Errorf("expected nic_attach br0, got %s", device.NIC.NICAttach)
	}
	if device.NIC.TrustGuestRxFilters {
		t.Error("expected trust_guest_rx_filters=false")
	}
}

func TestVMDeviceFromResponse_Display(t *testing.T) {
	resp := VMDeviceResponse{
		ID:    14,
		VM:    1,
		Order: 1005,
		Attributes: map[string]any{
			"dtype":      "DISPLAY",
			"type":       "SPICE",
			"port":       float64(5900),
			"bind":       "0.0.0.0",
			"password":   "secret",
			"web":        true,
			"resolution": "1024x768",
			"wait":       false,
		},
	}

	device := vmDeviceFromResponse(resp)
	if device.DeviceType != DeviceTypeDisplay {
		t.Errorf("expected DISPLAY, got %s", device.DeviceType)
	}
	if device.Display == nil {
		t.Fatal("expected non-nil Display")
	}
	if device.Display.Port != 5900 {
		t.Errorf("expected port 5900, got %v", device.Display.Port)
	}
	if !device.Display.Web {
		t.Error("expected web=true")
	}
	if device.Display.Wait {
		t.Error("expected wait=false")
	}
}

func TestVMDeviceFromResponse_PCI(t *testing.T) {
	resp := VMDeviceResponse{
		ID:    15,
		VM:    1,
		Order: 1006,
		Attributes: map[string]any{
			"dtype":  "PCI",
			"pptdev": "pci_0000_01_00_0",
		},
	}

	device := vmDeviceFromResponse(resp)
	if device.DeviceType != DeviceTypePCI {
		t.Errorf("expected PCI, got %s", device.DeviceType)
	}
	if device.PCI == nil {
		t.Fatal("expected non-nil PCI")
	}
	if device.PCI.PPTDev != "pci_0000_01_00_0" {
		t.Errorf("expected pptdev, got %s", device.PCI.PPTDev)
	}
}

func TestVMDeviceFromResponse_USB(t *testing.T) {
	resp := VMDeviceResponse{
		ID:    16,
		VM:    1,
		Order: 1007,
		Attributes: map[string]any{
			"dtype":           "USB",
			"controller_type": "nec-xhci",
			"device":          "usb_0_1_2",
			"usb_speed":       "HIGH",
		},
	}

	device := vmDeviceFromResponse(resp)
	if device.DeviceType != DeviceTypeUSB {
		t.Errorf("expected USB, got %s", device.DeviceType)
	}
	if device.USB == nil {
		t.Fatal("expected non-nil USB")
	}
	if device.USB.ControllerType != "nec-xhci" {
		t.Errorf("expected controller_type nec-xhci, got %s", device.USB.ControllerType)
	}
	if device.USB.Device != "usb_0_1_2" {
		t.Errorf("expected device usb_0_1_2, got %v", device.USB.Device)
	}
	if device.USB.USBSpeed != "HIGH" {
		t.Errorf("expected usb_speed HIGH, got %s", device.USB.USBSpeed)
	}
}

func TestVMDeviceFromResponse_USB_EmptyDevice(t *testing.T) {
	resp := VMDeviceResponse{
		ID:    17,
		VM:    1,
		Order: 1008,
		Attributes: map[string]any{
			"dtype":           "USB",
			"controller_type": "nec-xhci",
			"device":          "",
			"usb_speed":       "HIGH",
		},
	}

	device := vmDeviceFromResponse(resp)
	if device.USB == nil {
		t.Fatal("expected non-nil USB")
	}
	// Empty string device should remain empty string
	if device.USB.Device != "" {
		t.Errorf("expected empty Device for empty string, got %v", device.USB.Device)
	}
}

// --- Map Helper Tests ---

func TestStringFromMap(t *testing.T) {
	tests := []struct {
		name string
		m    map[string]any
		key  string
		want string
	}{
		{"existing string", map[string]any{"key": "value"}, "key", "value"},
		{"missing key", map[string]any{}, "key", ""},
		{"nil value", map[string]any{"key": nil}, "key", ""},
		{"non-string value", map[string]any{"key": 123}, "key", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stringFromMap(tt.m, tt.key)
			if got != tt.want {
				t.Errorf("stringFromMap() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBoolFromMap(t *testing.T) {
	tests := []struct {
		name string
		m    map[string]any
		key  string
		want bool
	}{
		{"true value", map[string]any{"key": true}, "key", true},
		{"false value", map[string]any{"key": false}, "key", false},
		{"missing key", map[string]any{}, "key", false},
		{"nil value", map[string]any{"key": nil}, "key", false},
		{"non-bool value", map[string]any{"key": "true"}, "key", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := boolFromMap(tt.m, tt.key)
			if got != tt.want {
				t.Errorf("boolFromMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIntFromMap(t *testing.T) {
	tests := []struct {
		name string
		m    map[string]any
		key  string
		want int64
	}{
		{"float64 value", map[string]any{"key": float64(42)}, "key", 42},
		{"int64 value", map[string]any{"key": int64(42)}, "key", 42},
		{"json.Number value", map[string]any{"key": json.Number("42")}, "key", 42},
		{"json.Number invalid", map[string]any{"key": json.Number("notanum")}, "key", 0},
		{"missing key", map[string]any{}, "key", 0},
		{"nil value", map[string]any{"key": nil}, "key", 0},
		{"non-numeric value", map[string]any{"key": "42"}, "key", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := intFromMap(tt.m, tt.key)
			if got != tt.want {
				t.Errorf("intFromMap() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestIntPtrFromMap(t *testing.T) {
	tests := []struct {
		name    string
		m       map[string]any
		key     string
		wantNil bool
		want    int64
	}{
		{"float64 value", map[string]any{"key": float64(42)}, "key", false, 42},
		{"int64 value", map[string]any{"key": int64(42)}, "key", false, 42},
		{"json.Number value", map[string]any{"key": json.Number("42")}, "key", false, 42},
		{"json.Number invalid", map[string]any{"key": json.Number("notanum")}, "key", true, 0},
		{"missing key", map[string]any{}, "key", true, 0},
		{"nil value", map[string]any{"key": nil}, "key", true, 0},
		{"non-numeric value", map[string]any{"key": "42"}, "key", true, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := intPtrFromMap(tt.m, tt.key)
			if tt.wantNil {
				if got != nil {
					t.Errorf("intPtrFromMap() = %v, want nil", *got)
				}
			} else {
				if got == nil {
					t.Fatal("intPtrFromMap() = nil, want non-nil")
				}
				if *got != tt.want {
					t.Errorf("intPtrFromMap() = %d, want %d", *got, tt.want)
				}
			}
		})
	}
}

// --- Param Builder Tests ---

func TestVMOptsToParams(t *testing.T) {
	opts := CreateVMOpts{
		Name:            "test-vm",
		Description:     "A test VM",
		VCPUs:           1,
		Cores:           2,
		Threads:         1,
		Memory:          2048,
		Autostart:       true,
		Time:            "LOCAL",
		Bootloader:      "UEFI",
		BootloaderOVMF:  "OVMF_CODE.fd",
		CPUMode:         "HOST-MODEL",
		ShutdownTimeout: 90,
	}

	params := vmOptsToParams(opts)

	if params["name"] != "test-vm" {
		t.Errorf("expected name test-vm, got %v", params["name"])
	}
	if params["memory"] != int64(2048) {
		t.Errorf("expected memory 2048, got %v", params["memory"])
	}
	// min_memory should NOT be set when nil
	if _, ok := params["min_memory"]; ok {
		t.Error("expected min_memory to be absent when nil")
	}
	// cpu_model should NOT be set when empty
	if _, ok := params["cpu_model"]; ok {
		t.Error("expected cpu_model to be absent when empty")
	}
}

func TestVMOptsToParams_WithMinMemory(t *testing.T) {
	minMem := int64(1024)
	opts := CreateVMOpts{
		MinMemory: &minMem,
	}

	params := vmOptsToParams(opts)

	if params["min_memory"] != int64(1024) {
		t.Errorf("expected min_memory 1024, got %v", params["min_memory"])
	}
}

func TestVMOptsToParams_WithCPUModel(t *testing.T) {
	opts := CreateVMOpts{
		CPUModel: "Haswell",
	}

	params := vmOptsToParams(opts)

	if params["cpu_model"] != "Haswell" {
		t.Errorf("expected cpu_model Haswell, got %v", params["cpu_model"])
	}
}

func TestStopVMOptsToParams(t *testing.T) {
	params := stopVMOptsToParams(StopVMOpts{
		Force:             true,
		ForceAfterTimeout: false,
	})

	if params["force"] != true {
		t.Errorf("expected force=true, got %v", params["force"])
	}
	if params["force_after_timeout"] != false {
		t.Errorf("expected force_after_timeout=false, got %v", params["force_after_timeout"])
	}
}

func TestDeviceOptsToParams_NoOrder(t *testing.T) {
	opts := CreateVMDeviceOpts{
		VM:         1,
		DeviceType: DeviceTypeDisk,
		Disk: &DiskDevice{
			Path: "/dev/zvol/tank/vm-disk",
		},
	}

	params := deviceOptsToParams(opts)

	if _, ok := params["order"]; ok {
		t.Error("expected order to be absent when nil")
	}
	if params["vm"] != int64(1) {
		t.Errorf("expected vm 1, got %v", params["vm"])
	}
	attrs := params["attributes"].(map[string]any)
	if attrs["dtype"] != "DISK" {
		t.Errorf("expected dtype DISK, got %v", attrs["dtype"])
	}
}

func TestSetNonEmpty(t *testing.T) {
	m := map[string]any{}
	setNonEmpty(m, "key1", "value1")
	setNonEmpty(m, "key2", "")

	if m["key1"] != "value1" {
		t.Errorf("expected key1=value1, got %v", m["key1"])
	}
	if _, ok := m["key2"]; ok {
		t.Error("expected key2 to be absent for empty string")
	}
}

func TestSetNonNilInt(t *testing.T) {
	m := map[string]any{}
	val := int64(42)
	setNonNilInt(m, "key1", &val)
	setNonNilInt(m, "key2", nil)

	if m["key1"] != int64(42) {
		t.Errorf("expected key1=42, got %v", m["key1"])
	}
	if _, ok := m["key2"]; ok {
		t.Error("expected key2 to be absent for nil")
	}
}

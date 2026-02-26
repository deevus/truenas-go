package truenas

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

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

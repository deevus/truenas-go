package truenas

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

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

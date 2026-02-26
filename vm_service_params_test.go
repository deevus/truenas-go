package truenas

import (
	"encoding/json"
	"testing"
)

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

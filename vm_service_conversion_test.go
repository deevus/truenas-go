package truenas

import "testing"

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
			"dtype":               "DISK",
			"path":                "/dev/zvol/tank/vm-disk",
			"type":                "VIRTIO",
			"physical_sectorsize": nil,
			"logical_sectorsize":  nil,
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
			"dtype":                   "NIC",
			"type":                    "VIRTIO",
			"nic_attach":             "br0",
			"mac":                     "00:a0:98:6b:0c:01",
			"trust_guest_rx_filters": false,
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

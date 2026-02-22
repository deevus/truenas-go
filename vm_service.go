package truenas

import (
	"context"
	"encoding/json"
	"fmt"
)

// DeviceType represents the type of a VM device.
type DeviceType string

const (
	DeviceTypeDisk    DeviceType = "DISK"
	DeviceTypeRaw     DeviceType = "RAW"
	DeviceTypeCDROM   DeviceType = "CDROM"
	DeviceTypeNIC     DeviceType = "NIC"
	DeviceTypeDisplay DeviceType = "DISPLAY"
	DeviceTypePCI     DeviceType = "PCI"
	DeviceTypeUSB     DeviceType = "USB"
)

// VM is the user-facing representation of a TrueNAS virtual machine.
type VM struct {
	ID              int64
	Name            string
	Description     string
	VCPUs           int64
	Cores           int64
	Threads         int64
	Memory          int64
	MinMemory       *int64
	Autostart       bool
	Time            string
	Bootloader      string
	BootloaderOVMF  string
	CPUMode         string
	CPUModel        string
	ShutdownTimeout int64
	CommandLineArgs string
	State           string
}

// CreateVMOpts contains options for creating a VM.
type CreateVMOpts struct {
	Name            string
	Description     string
	VCPUs           int64
	Cores           int64
	Threads         int64
	Memory          int64
	MinMemory       *int64
	Autostart       bool
	Time            string
	Bootloader      string
	BootloaderOVMF  string
	CPUMode         string
	CPUModel        string
	ShutdownTimeout int64
	CommandLineArgs string
}

// UpdateVMOpts contains options for updating a VM.
// All fields are always sent on update.
type UpdateVMOpts = CreateVMOpts

// StopVMOpts contains options for stopping a VM.
type StopVMOpts struct {
	Force             bool
	ForceAfterTimeout bool
}

// DiskDevice contains attributes for a DISK device.
type DiskDevice struct {
	Path         string
	Type         string
	PhysicalSectorSize *int64
	Logical_Sector_Size *int64
}

// RawDevice contains attributes for a RAW device.
type RawDevice struct {
	Path         string
	Type         string
	Boot         bool
	Size         *int64
	PhysicalSectorSize *int64
	Logical_Sector_Size *int64
}

// CDROMDevice contains attributes for a CDROM device.
type CDROMDevice struct {
	Path string
}

// NICDevice contains attributes for a NIC device.
type NICDevice struct {
	Type    string
	NICAttach string
	MAC     string
	TrustGuestRXFilters bool
}

// DisplayDevice contains attributes for a DISPLAY device.
type DisplayDevice struct {
	Type       string
	Port       *int64
	Bind       string
	Password   string
	Web        bool
	Resolution string
	Wait       bool
}

// PCIDevice contains attributes for a PCI device.
type PCIDevice struct {
	Pptdev string
}

// USBDevice contains attributes for a USB device.
type USBDevice struct {
	ControllerType string
	Device         *string
	USBSpeed       string
}

// VMDevice is the user-facing representation of a VM device.
type VMDevice struct {
	ID         int64
	VM         int64
	Order      int64
	DeviceType DeviceType
	Disk       *DiskDevice
	Raw        *RawDevice
	CDROM      *CDROMDevice
	NIC        *NICDevice
	Display    *DisplayDevice
	PCI        *PCIDevice
	USB        *USBDevice
}

// CreateVMDeviceOpts contains options for creating a VM device.
type CreateVMDeviceOpts struct {
	VM         int64
	Order      *int64
	DeviceType DeviceType
	Disk       *DiskDevice
	Raw        *RawDevice
	CDROM      *CDROMDevice
	NIC        *NICDevice
	Display    *DisplayDevice
	PCI        *PCIDevice
	USB        *USBDevice
}

// UpdateVMDeviceOpts contains options for updating a VM device.
type UpdateVMDeviceOpts = CreateVMDeviceOpts

// VMService provides typed methods for the vm.* API namespace.
type VMService struct {
	client  AsyncCaller
	version Version
}

// NewVMService creates a new VMService.
func NewVMService(c AsyncCaller, v Version) *VMService {
	return &VMService{client: c, version: v}
}

// CreateVM creates a VM and returns the full object.
// The create response includes the full VM, so no re-read is needed.
func (s *VMService) CreateVM(ctx context.Context, opts CreateVMOpts) (*VM, error) {
	params := vmOptsToParams(opts)
	result, err := s.client.Call(ctx, "vm.create", params)
	if err != nil {
		return nil, err
	}

	var resp VMResponse
	if err := json.Unmarshal(result, &resp); err != nil {
		return nil, fmt.Errorf("parse create response: %w", err)
	}

	vm := vmFromResponse(resp)
	return &vm, nil
}

// GetVM returns a VM by ID.
func (s *VMService) GetVM(ctx context.Context, id int64) (*VM, error) {
	result, err := s.client.Call(ctx, "vm.get_instance", id)
	if err != nil {
		return nil, err
	}

	var resp VMResponse
	if err := json.Unmarshal(result, &resp); err != nil {
		return nil, fmt.Errorf("parse get response: %w", err)
	}

	vm := vmFromResponse(resp)
	return &vm, nil
}

// UpdateVM updates a VM and returns the full object via re-read.
func (s *VMService) UpdateVM(ctx context.Context, id int64, opts UpdateVMOpts) (*VM, error) {
	params := vmOptsToParams(opts)
	_, err := s.client.Call(ctx, "vm.update", []any{id, params})
	if err != nil {
		return nil, err
	}

	return s.GetVM(ctx, id)
}

// DeleteVM deletes a VM by ID.
func (s *VMService) DeleteVM(ctx context.Context, id int64) error {
	_, err := s.client.Call(ctx, "vm.delete", id)
	return err
}

// StartVM starts a VM by ID.
func (s *VMService) StartVM(ctx context.Context, id int64) error {
	_, err := s.client.Call(ctx, "vm.start", id)
	return err
}

// StopVM stops a VM by ID using CallAndWait since it is a long-running operation.
func (s *VMService) StopVM(ctx context.Context, id int64, opts StopVMOpts) error {
	params := stopVMOptsToParams(opts)
	_, err := s.client.CallAndWait(ctx, "vm.stop", []any{id, params})
	return err
}

// ListDevices returns all devices for a VM.
func (s *VMService) ListDevices(ctx context.Context, vmID int64) ([]VMDevice, error) {
	filter := []any{[]any{[]any{"vm", "=", vmID}}}
	result, err := s.client.Call(ctx, "vm.device.query", filter)
	if err != nil {
		return nil, err
	}

	var responses []VMDeviceResponse
	if err := json.Unmarshal(result, &responses); err != nil {
		return nil, fmt.Errorf("parse device query response: %w", err)
	}

	devices := make([]VMDevice, len(responses))
	for i, resp := range responses {
		devices[i] = vmDeviceFromResponse(resp)
	}
	return devices, nil
}

// GetDevice returns a VM device by ID, or nil if not found.
func (s *VMService) GetDevice(ctx context.Context, id int64) (*VMDevice, error) {
	filter := []any{[]any{[]any{"id", "=", id}}}
	result, err := s.client.Call(ctx, "vm.device.query", filter)
	if err != nil {
		return nil, err
	}

	var responses []VMDeviceResponse
	if err := json.Unmarshal(result, &responses); err != nil {
		return nil, fmt.Errorf("parse device query response: %w", err)
	}

	if len(responses) == 0 {
		return nil, nil
	}

	device := vmDeviceFromResponse(responses[0])
	return &device, nil
}

// CreateDevice creates a VM device and returns the full object.
func (s *VMService) CreateDevice(ctx context.Context, opts CreateVMDeviceOpts) (*VMDevice, error) {
	params := deviceOptsToParams(opts)
	result, err := s.client.Call(ctx, "vm.device.create", params)
	if err != nil {
		return nil, err
	}

	var resp VMDeviceResponse
	if err := json.Unmarshal(result, &resp); err != nil {
		return nil, fmt.Errorf("parse device create response: %w", err)
	}

	device := vmDeviceFromResponse(resp)
	return &device, nil
}

// UpdateDevice updates a VM device and returns the full object via re-read.
func (s *VMService) UpdateDevice(ctx context.Context, id int64, opts UpdateVMDeviceOpts) (*VMDevice, error) {
	params := deviceOptsToParams(opts)
	_, err := s.client.Call(ctx, "vm.device.update", []any{id, params})
	if err != nil {
		return nil, err
	}

	return s.GetDevice(ctx, id)
}

// DeleteDevice deletes a VM device by ID.
func (s *VMService) DeleteDevice(ctx context.Context, id int64) error {
	_, err := s.client.Call(ctx, "vm.device.delete", id)
	return err
}

// vmOptsToParams converts CreateVMOpts to API parameters.
func vmOptsToParams(opts CreateVMOpts) map[string]any {
	params := map[string]any{
		"name":              opts.Name,
		"description":       opts.Description,
		"vcpus":             opts.VCPUs,
		"cores":             opts.Cores,
		"threads":           opts.Threads,
		"memory":            opts.Memory,
		"autostart":         opts.Autostart,
		"time":              opts.Time,
		"bootloader":        opts.Bootloader,
		"bootloader_ovmf":   opts.BootloaderOVMF,
		"cpu_mode":          opts.CPUMode,
		"shutdown_timeout":  opts.ShutdownTimeout,
		"command_line_args": opts.CommandLineArgs,
	}

	if opts.MinMemory != nil {
		params["min_memory"] = *opts.MinMemory
	}

	if opts.CPUModel != "" {
		params["cpu_model"] = opts.CPUModel
	}

	return params
}

// stopVMOptsToParams converts StopVMOpts to API parameters.
func stopVMOptsToParams(opts StopVMOpts) map[string]any {
	return map[string]any{
		"force":               opts.Force,
		"force_after_timeout": opts.ForceAfterTimeout,
	}
}

// deviceOptsToParams converts CreateVMDeviceOpts to API parameters.
func deviceOptsToParams(opts CreateVMDeviceOpts) map[string]any {
	attrs := map[string]any{
		"dtype": string(opts.DeviceType),
	}

	switch opts.DeviceType {
	case DeviceTypeDisk:
		if opts.Disk != nil {
			setNonEmpty(attrs, "path", opts.Disk.Path)
			setNonEmpty(attrs, "type", opts.Disk.Type)
			setNonNilInt(attrs, "physical_sectorsize", opts.Disk.PhysicalSectorSize)
			setNonNilInt(attrs, "logical_sectorsize", opts.Disk.Logical_Sector_Size)
		}
	case DeviceTypeRaw:
		if opts.Raw != nil {
			setNonEmpty(attrs, "path", opts.Raw.Path)
			setNonEmpty(attrs, "type", opts.Raw.Type)
			attrs["boot"] = opts.Raw.Boot
			setNonNilInt(attrs, "size", opts.Raw.Size)
			setNonNilInt(attrs, "physical_sectorsize", opts.Raw.PhysicalSectorSize)
			setNonNilInt(attrs, "logical_sectorsize", opts.Raw.Logical_Sector_Size)
		}
	case DeviceTypeCDROM:
		if opts.CDROM != nil {
			setNonEmpty(attrs, "path", opts.CDROM.Path)
		}
	case DeviceTypeNIC:
		if opts.NIC != nil {
			setNonEmpty(attrs, "type", opts.NIC.Type)
			setNonEmpty(attrs, "nic_attach", opts.NIC.NICAttach)
			setNonEmpty(attrs, "mac", opts.NIC.MAC)
			attrs["trust_guest_rx_filters"] = opts.NIC.TrustGuestRXFilters
		}
	case DeviceTypeDisplay:
		if opts.Display != nil {
			setNonEmpty(attrs, "type", opts.Display.Type)
			setNonNilInt(attrs, "port", opts.Display.Port)
			setNonEmpty(attrs, "bind", opts.Display.Bind)
			setNonEmpty(attrs, "password", opts.Display.Password)
			attrs["web"] = opts.Display.Web
			setNonEmpty(attrs, "resolution", opts.Display.Resolution)
			attrs["wait"] = opts.Display.Wait
		}
	case DeviceTypePCI:
		if opts.PCI != nil {
			setNonEmpty(attrs, "pptdev", opts.PCI.Pptdev)
		}
	case DeviceTypeUSB:
		if opts.USB != nil {
			setNonEmpty(attrs, "controller_type", opts.USB.ControllerType)
			if opts.USB.Device != nil {
				attrs["device"] = *opts.USB.Device
			}
			setNonEmpty(attrs, "usb_speed", opts.USB.USBSpeed)
		}
	}

	params := map[string]any{
		"vm":         opts.VM,
		"attributes": attrs,
	}

	if opts.Order != nil {
		params["order"] = *opts.Order
	}

	return params
}

// setNonEmpty sets a map key only if the value is non-empty.
func setNonEmpty(m map[string]any, key, value string) {
	if value != "" {
		m[key] = value
	}
}

// setNonNilInt sets a map key only if the value is non-nil.
func setNonNilInt(m map[string]any, key string, value *int64) {
	if value != nil {
		m[key] = *value
	}
}

// vmFromResponse converts a VMResponse to a user-facing VM.
func vmFromResponse(resp VMResponse) VM {
	cpuModel := ""
	if resp.CPUModel != nil {
		cpuModel = *resp.CPUModel
	}

	return VM{
		ID:              resp.ID,
		Name:            resp.Name,
		Description:     resp.Description,
		VCPUs:           resp.VCPUs,
		Cores:           resp.Cores,
		Threads:         resp.Threads,
		Memory:          resp.Memory,
		MinMemory:       resp.MinMemory,
		Autostart:       resp.Autostart,
		Time:            resp.Time,
		Bootloader:      resp.Bootloader,
		BootloaderOVMF:  resp.BootloaderOVMF,
		CPUMode:         resp.CPUMode,
		CPUModel:        cpuModel,
		ShutdownTimeout: resp.ShutdownTimeout,
		CommandLineArgs: resp.CommandLineArgs,
		State:           resp.Status.State,
	}
}

// vmDeviceFromResponse converts a VMDeviceResponse to a user-facing VMDevice.
func vmDeviceFromResponse(resp VMDeviceResponse) VMDevice {
	dtype := DeviceType(stringFromMap(resp.Attributes, "dtype"))

	device := VMDevice{
		ID:         resp.ID,
		VM:         resp.VM,
		Order:      resp.Order,
		DeviceType: dtype,
	}

	switch dtype {
	case DeviceTypeDisk:
		device.Disk = &DiskDevice{
			Path:                stringFromMap(resp.Attributes, "path"),
			Type:                stringFromMap(resp.Attributes, "type"),
			PhysicalSectorSize:  intPtrFromMap(resp.Attributes, "physical_sectorsize"),
			Logical_Sector_Size: intPtrFromMap(resp.Attributes, "logical_sectorsize"),
		}
	case DeviceTypeRaw:
		device.Raw = &RawDevice{
			Path:                stringFromMap(resp.Attributes, "path"),
			Type:                stringFromMap(resp.Attributes, "type"),
			Boot:                boolFromMap(resp.Attributes, "boot"),
			Size:                intPtrFromMap(resp.Attributes, "size"),
			PhysicalSectorSize:  intPtrFromMap(resp.Attributes, "physical_sectorsize"),
			Logical_Sector_Size: intPtrFromMap(resp.Attributes, "logical_sectorsize"),
		}
	case DeviceTypeCDROM:
		device.CDROM = &CDROMDevice{
			Path: stringFromMap(resp.Attributes, "path"),
		}
	case DeviceTypeNIC:
		device.NIC = &NICDevice{
			Type:                stringFromMap(resp.Attributes, "type"),
			NICAttach:           stringFromMap(resp.Attributes, "nic_attach"),
			MAC:                 stringFromMap(resp.Attributes, "mac"),
			TrustGuestRXFilters: boolFromMap(resp.Attributes, "trust_guest_rx_filters"),
		}
	case DeviceTypeDisplay:
		device.Display = &DisplayDevice{
			Type:       stringFromMap(resp.Attributes, "type"),
			Port:       intPtrFromMap(resp.Attributes, "port"),
			Bind:       stringFromMap(resp.Attributes, "bind"),
			Password:   stringFromMap(resp.Attributes, "password"),
			Web:        boolFromMap(resp.Attributes, "web"),
			Resolution: stringFromMap(resp.Attributes, "resolution"),
			Wait:       boolFromMap(resp.Attributes, "wait"),
		}
	case DeviceTypePCI:
		device.PCI = &PCIDevice{
			Pptdev: stringFromMap(resp.Attributes, "pptdev"),
		}
	case DeviceTypeUSB:
		var devicePtr *string
		if v := stringFromMap(resp.Attributes, "device"); v != "" {
			devicePtr = &v
		}
		device.USB = &USBDevice{
			ControllerType: stringFromMap(resp.Attributes, "controller_type"),
			Device:         devicePtr,
			USBSpeed:       stringFromMap(resp.Attributes, "usb_speed"),
		}
	}

	return device
}

// stringFromMap extracts a string value from a map.
func stringFromMap(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}

// boolFromMap extracts a bool value from a map.
func boolFromMap(m map[string]any, key string) bool {
	v, ok := m[key]
	if !ok || v == nil {
		return false
	}
	b, ok := v.(bool)
	if !ok {
		return false
	}
	return b
}

// intFromMap extracts an int64 value from a map.
// Handles float64 (standard JSON unmarshalling), int64, and json.Number.
func intFromMap(m map[string]any, key string) int64 {
	v, ok := m[key]
	if !ok || v == nil {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return int64(n)
	case int64:
		return n
	case json.Number:
		i, err := n.Int64()
		if err != nil {
			return 0
		}
		return i
	}
	return 0
}

// intPtrFromMap extracts an *int64 value from a map. Returns nil if the key is absent or null.
func intPtrFromMap(m map[string]any, key string) *int64 {
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	var result int64
	switch n := v.(type) {
	case float64:
		result = int64(n)
	case int64:
		result = n
	case json.Number:
		i, err := n.Int64()
		if err != nil {
			return nil
		}
		result = i
	default:
		return nil
	}
	return &result
}

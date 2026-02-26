package truenas

import (
	"context"
	"encoding/json"
	"fmt"
)

// VirtGlobalConfig is the user-facing representation of the global virt configuration.
type VirtGlobalConfig struct {
	Bridge       string
	V4Network    string
	V6Network    string
	Pool         string
	Dataset      string
	StoragePools []string
	State        string
}

// UpdateVirtGlobalConfigOpts contains options for updating global virt configuration.
// Only non-nil fields are sent in the update request.
type UpdateVirtGlobalConfigOpts struct {
	Bridge    *string
	V4Network *string
	V6Network *string
	Pool      *string
}

// VirtInstance is the user-facing representation of a virt instance.
type VirtInstance struct {
	ID          string
	Name        string
	Type        string
	Status      string
	CPU         string
	Memory      int64
	Autostart   bool
	Environment map[string]string
	Aliases     []VirtAlias
	Image       VirtInstanceImageResponse
	StoragePool string
}

// VirtAlias represents a network alias for a virt instance.
type VirtAlias struct {
	Type    string
	Address string
	Netmask *int64
}

// CreateVirtInstanceOpts contains options for creating a virt instance.
type CreateVirtInstanceOpts struct {
	Name         string
	InstanceType string
	Image        string
	CPU          string
	Memory       int64
	Autostart    bool
	Environment  map[string]string
	Devices      []VirtDeviceOpts
	StoragePool  string
}

// UpdateVirtInstanceOpts contains options for updating a virt instance.
// Only non-nil fields are sent in the update request.
type UpdateVirtInstanceOpts struct {
	Autostart   *bool
	Environment map[string]string
}

// StopVirtInstanceOpts contains options for stopping a virt instance.
type StopVirtInstanceOpts struct {
	Timeout int64
}

// VirtDeviceOpts contains options for adding a device to a virt instance.
// Fields are used based on DevType: DISK uses Source/Destination,
// NIC uses Network/NICType/Parent, PROXY uses SourceProto/SourcePort/DestProto/DestPort.
type VirtDeviceOpts struct {
	DevType     string
	Name        string
	Readonly    bool
	// DISK fields
	Source      string
	Destination string
	// NIC fields
	Network string
	NICType string
	Parent  string
	// PROXY fields
	SourceProto string
	SourcePort  int64
	DestProto   string
	DestPort    int64
}

// VirtDevice is the user-facing representation of a device attached to a virt instance.
type VirtDevice struct {
	DevType     string
	Name        string
	Description string
	Readonly    bool
	// DISK fields
	Source      string
	Destination string
	// NIC fields
	Network string
	NICType string
	Parent  string
	// PROXY fields
	SourceProto string
	SourcePort  int64
	DestProto   string
	DestPort    int64
}

// VirtService provides typed methods for the virt.* API namespace.
type VirtService struct {
	client  AsyncCaller
	version Version
}

// NewVirtService creates a new VirtService.
func NewVirtService(c AsyncCaller, v Version) *VirtService {
	return &VirtService{client: c, version: v}
}

// GetGlobalConfig returns the global virt configuration.
func (s *VirtService) GetGlobalConfig(ctx context.Context) (*VirtGlobalConfig, error) {
	result, err := s.client.Call(ctx, "virt.global.config", nil)
	if err != nil {
		return nil, err
	}

	var resp VirtGlobalConfigResponse
	if err := json.Unmarshal(result, &resp); err != nil {
		return nil, fmt.Errorf("parse global config response: %w", err)
	}

	cfg := virtGlobalConfigFromResponse(resp)
	return &cfg, nil
}

// UpdateGlobalConfig updates the global virt configuration and returns the updated config.
// Only non-nil fields in opts are sent.
func (s *VirtService) UpdateGlobalConfig(ctx context.Context, opts UpdateVirtGlobalConfigOpts) (*VirtGlobalConfig, error) {
	params := virtGlobalConfigOptsToParams(opts)
	_, err := s.client.Call(ctx, "virt.global.update", params)
	if err != nil {
		return nil, err
	}

	return s.GetGlobalConfig(ctx)
}

// CreateInstance creates a virt instance and returns the full object.
func (s *VirtService) CreateInstance(ctx context.Context, opts CreateVirtInstanceOpts) (*VirtInstance, error) {
	params := virtInstanceCreateOptsToParams(opts)
	_, err := s.client.CallAndWait(ctx, "virt.instance.create", params)
	if err != nil {
		return nil, err
	}

	return s.GetInstance(ctx, opts.Name)
}

// GetInstance returns a virt instance by name, or nil if not found.
func (s *VirtService) GetInstance(ctx context.Context, name string) (*VirtInstance, error) {
	result, err := s.client.Call(ctx, "virt.instance.get_instance", name)
	if err != nil {
		if isNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}

	var resp VirtInstanceResponse
	if err := json.Unmarshal(result, &resp); err != nil {
		return nil, fmt.Errorf("parse instance response: %w", err)
	}

	inst := virtInstanceFromResponse(resp)
	return &inst, nil
}

// UpdateInstance updates a virt instance and returns the full object.
func (s *VirtService) UpdateInstance(ctx context.Context, name string, opts UpdateVirtInstanceOpts) (*VirtInstance, error) {
	params := virtInstanceUpdateOptsToParams(opts)
	_, err := s.client.CallAndWait(ctx, "virt.instance.update", []any{name, params})
	if err != nil {
		return nil, err
	}

	return s.GetInstance(ctx, name)
}

// DeleteInstance deletes a virt instance by name.
func (s *VirtService) DeleteInstance(ctx context.Context, name string) error {
	_, err := s.client.CallAndWait(ctx, "virt.instance.delete", name)
	return err
}

// StartInstance starts a virt instance by name.
func (s *VirtService) StartInstance(ctx context.Context, name string) error {
	_, err := s.client.CallAndWait(ctx, "virt.instance.start", name)
	return err
}

// StopInstance stops a virt instance by name with optional timeout.
func (s *VirtService) StopInstance(ctx context.Context, name string, opts StopVirtInstanceOpts) error {
	stopArgs := map[string]any{}
	if opts.Timeout > 0 {
		stopArgs["timeout"] = opts.Timeout
	}
	_, err := s.client.CallAndWait(ctx, "virt.instance.stop", []any{name, stopArgs})
	return err
}

// ListInstances queries virt instances with optional filters.
// Filters use TrueNAS query format: [][]any{{"field", "op", "value"}}.
// Pass nil for no filtering.
func (s *VirtService) ListInstances(ctx context.Context, filters [][]any) ([]VirtInstance, error) {
	var params any
	if len(filters) > 0 {
		params = []any{filters}
	}

	result, err := s.client.Call(ctx, "virt.instance.query", params)
	if err != nil {
		return nil, err
	}

	var responses []VirtInstanceResponse
	if err := json.Unmarshal(result, &responses); err != nil {
		return nil, fmt.Errorf("parse instance list response: %w", err)
	}

	instances := make([]VirtInstance, len(responses))
	for i, resp := range responses {
		instances[i] = virtInstanceFromResponse(resp)
	}
	return instances, nil
}

// ListDevices returns all devices attached to a virt instance.
func (s *VirtService) ListDevices(ctx context.Context, instanceID string) ([]VirtDevice, error) {
	result, err := s.client.Call(ctx, "virt.instance.device_list", instanceID)
	if err != nil {
		return nil, err
	}

	var responses []VirtDeviceResponse
	if err := json.Unmarshal(result, &responses); err != nil {
		return nil, fmt.Errorf("parse device list response: %w", err)
	}

	devices := make([]VirtDevice, len(responses))
	for i, resp := range responses {
		devices[i] = virtDeviceFromResponse(resp)
	}
	return devices, nil
}

// AddDevice adds a device to a virt instance.
func (s *VirtService) AddDevice(ctx context.Context, instanceID string, opts VirtDeviceOpts) error {
	devMap := virtDeviceOptToParam(opts)
	_, err := s.client.CallAndWait(ctx, "virt.instance.device_add", []any{instanceID, devMap})
	return err
}

// DeleteDevice removes a device from a virt instance by device name.
func (s *VirtService) DeleteDevice(ctx context.Context, instanceID string, deviceName string) error {
	_, err := s.client.CallAndWait(ctx, "virt.instance.device_delete", []any{instanceID, deviceName})
	return err
}

// virtGlobalConfigFromResponse converts a wire-format response to a user-facing config.
// Nil pointer fields are converted to empty strings.
func virtGlobalConfigFromResponse(resp VirtGlobalConfigResponse) VirtGlobalConfig {
	storagePools := resp.StoragePools
	if storagePools == nil {
		storagePools = []string{}
	}
	return VirtGlobalConfig{
		Bridge:       derefString(resp.Bridge),
		V4Network:    derefString(resp.V4Network),
		V6Network:    derefString(resp.V6Network),
		Pool:         derefString(resp.Pool),
		Dataset:      derefString(resp.Dataset),
		StoragePools: storagePools,
		State:        derefString(resp.State),
	}
}

// virtGlobalConfigOptsToParams converts update options to API parameters.
// Only non-nil fields are included.
func virtGlobalConfigOptsToParams(opts UpdateVirtGlobalConfigOpts) map[string]any {
	params := map[string]any{}
	if opts.Bridge != nil {
		params["bridge"] = *opts.Bridge
	}
	if opts.V4Network != nil {
		params["v4_network"] = *opts.V4Network
	}
	if opts.V6Network != nil {
		params["v6_network"] = *opts.V6Network
	}
	if opts.Pool != nil {
		params["pool"] = *opts.Pool
	}
	return params
}

// virtInstanceFromResponse converts a wire-format response to a user-facing instance.
func virtInstanceFromResponse(resp VirtInstanceResponse) VirtInstance {
	aliases := make([]VirtAlias, len(resp.Aliases))
	for i, a := range resp.Aliases {
		aliases[i] = virtAliasFromResponse(a)
	}

	env := resp.Environment
	if env == nil {
		env = map[string]string{}
	}

	return VirtInstance{
		ID:          resp.ID,
		Name:        resp.Name,
		Type:        resp.Type,
		Status:      resp.Status,
		CPU:         derefString(resp.CPU),
		Memory:      derefInt64(resp.Memory),
		Autostart:   resp.Autostart,
		Environment: env,
		Aliases:     aliases,
		Image:       resp.Image,
		StoragePool: resp.StoragePool,
	}
}

// virtAliasFromResponse converts a wire-format alias response to a user-facing alias.
func virtAliasFromResponse(resp VirtInstanceAliasResponse) VirtAlias {
	return VirtAlias{
		Type:    resp.Type,
		Address: resp.Address,
		Netmask: resp.Netmask,
	}
}

// virtInstanceCreateOptsToParams converts create options to API parameters.
func virtInstanceCreateOptsToParams(opts CreateVirtInstanceOpts) map[string]any {
	params := map[string]any{
		"name":          opts.Name,
		"instance_type": opts.InstanceType,
		"image":         opts.Image,
		"cpu":           opts.CPU,
		"memory":        opts.Memory,
		"autostart":     opts.Autostart,
		"environment":   opts.Environment,
	}
	if len(opts.Devices) > 0 {
		params["devices"] = virtDeviceOptsToParams(opts.Devices)
	}
	if opts.StoragePool != "" {
		params["storage_pool"] = opts.StoragePool
	}
	return params
}

// virtInstanceUpdateOptsToParams converts update options to API parameters.
// Only non-nil fields are included.
func virtInstanceUpdateOptsToParams(opts UpdateVirtInstanceOpts) map[string]any {
	params := map[string]any{}
	if opts.Autostart != nil {
		params["autostart"] = *opts.Autostart
	}
	if opts.Environment != nil {
		params["environment"] = opts.Environment
	}
	return params
}

// virtDeviceOptsToParams converts a slice of device options to API parameters.
func virtDeviceOptsToParams(opts []VirtDeviceOpts) []map[string]any {
	result := make([]map[string]any, len(opts))
	for i, opt := range opts {
		result[i] = virtDeviceOptToParam(opt)
	}
	return result
}

// virtDeviceOptToParam converts a single device option to an API parameter map.
// Fields included depend on DevType.
func virtDeviceOptToParam(opt VirtDeviceOpts) map[string]any {
	m := map[string]any{
		"dev_type": opt.DevType,
		"readonly": opt.Readonly,
	}
	if opt.Name != "" {
		m["name"] = opt.Name
	}

	switch opt.DevType {
	case "DISK":
		m["source"] = opt.Source
		m["destination"] = opt.Destination
	case "NIC":
		if opt.Network != "" {
			m["network"] = opt.Network
		}
		if opt.NICType != "" {
			m["nic_type"] = opt.NICType
		}
		if opt.Parent != "" {
			m["parent"] = opt.Parent
		}
	case "PROXY":
		m["source_proto"] = opt.SourceProto
		m["source_port"] = opt.SourcePort
		m["dest_proto"] = opt.DestProto
		m["dest_port"] = opt.DestPort
	}

	return m
}

// virtDeviceFromResponse converts a wire-format device response to a user-facing device.
// Nil pointer fields are converted to zero values.
func virtDeviceFromResponse(resp VirtDeviceResponse) VirtDevice {
	return VirtDevice{
		DevType:     resp.DevType,
		Name:        derefString(resp.Name),
		Description: derefString(resp.Description),
		Readonly:    resp.Readonly,
		Source:      derefString(resp.Source),
		Destination: derefString(resp.Destination),
		Network:     derefString(resp.Network),
		NICType:     derefString(resp.NICType),
		Parent:      derefString(resp.Parent),
		SourceProto: derefString(resp.SourceProto),
		SourcePort:  derefInt64(resp.SourcePort),
		DestProto:   derefString(resp.DestProto),
		DestPort:    derefInt64(resp.DestPort),
	}
}

// derefString returns the string value of a pointer, or "" if nil.
func derefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// derefInt64 returns the int64 value of a pointer, or 0 if nil.
func derefInt64(p *int64) int64 {
	if p == nil {
		return 0
	}
	return *p
}

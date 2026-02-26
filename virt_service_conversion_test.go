package truenas

import (
	"testing"
)

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
	if inst.Aliases[0].Netmask == nil || *inst.Aliases[0].Netmask != 24 {
		t.Errorf("expected netmask 24, got %v", inst.Aliases[0].Netmask)
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
		DevType:     "DISK",
		Source:      "/mnt/tank/data",
		Destination: "/data",
	})
	if _, ok := m["name"]; ok {
		t.Error("name should not be in params when empty")
	}
}

func TestVirtDeviceOptToParam_WithName(t *testing.T) {
	m := virtDeviceOptToParam(VirtDeviceOpts{
		DevType:     "DISK",
		Name:        "mydevice",
		Source:      "/mnt/tank/data",
		Destination: "/data",
	})
	if m["name"] != "mydevice" {
		t.Errorf("expected name mydevice, got %v", m["name"])
	}
}

func TestVirtDeviceOptToParam_NIC_PartialFields(t *testing.T) {
	t.Run("only nic_type and parent", func(t *testing.T) {
		m := virtDeviceOptToParam(VirtDeviceOpts{
			DevType: "NIC",
			NICType: "MACVLAN",
			Parent:  "eno1",
		})
		if _, ok := m["network"]; ok {
			t.Error("network should not be in params when empty")
		}
		if m["nic_type"] != "MACVLAN" {
			t.Errorf("expected nic_type MACVLAN, got %v", m["nic_type"])
		}
		if m["parent"] != "eno1" {
			t.Errorf("expected parent eno1, got %v", m["parent"])
		}
	})

	t.Run("only network", func(t *testing.T) {
		m := virtDeviceOptToParam(VirtDeviceOpts{
			DevType: "NIC",
			Network: "br0",
		})
		if m["network"] != "br0" {
			t.Errorf("expected network br0, got %v", m["network"])
		}
		if _, ok := m["nic_type"]; ok {
			t.Error("nic_type should not be in params when empty")
		}
		if _, ok := m["parent"]; ok {
			t.Error("parent should not be in params when empty")
		}
	})

	t.Run("no NIC fields set", func(t *testing.T) {
		m := virtDeviceOptToParam(VirtDeviceOpts{
			DevType: "NIC",
		})
		if _, ok := m["network"]; ok {
			t.Error("network should not be in params when empty")
		}
		if _, ok := m["nic_type"]; ok {
			t.Error("nic_type should not be in params when empty")
		}
		if _, ok := m["parent"]; ok {
			t.Error("parent should not be in params when empty")
		}
		if m["dev_type"] != "NIC" {
			t.Errorf("expected dev_type NIC, got %v", m["dev_type"])
		}
	})
}

func TestVirtGlobalConfigFromResponse_WithNewFields(t *testing.T) {
	dataset := "tank/ix-virt"
	state := "INITIALIZED"
	pool := "tank"
	resp := VirtGlobalConfigResponse{
		Pool:         &pool,
		Dataset:      &dataset,
		StoragePools: []string{"tank", "ssd"},
		State:        &state,
	}

	cfg := virtGlobalConfigFromResponse(resp)
	if cfg.Dataset != "tank/ix-virt" {
		t.Errorf("expected dataset tank/ix-virt, got %s", cfg.Dataset)
	}
	if len(cfg.StoragePools) != 2 {
		t.Fatalf("expected 2 storage pools, got %d", len(cfg.StoragePools))
	}
	if cfg.StoragePools[0] != "tank" {
		t.Errorf("expected first storage pool tank, got %s", cfg.StoragePools[0])
	}
	if cfg.StoragePools[1] != "ssd" {
		t.Errorf("expected second storage pool ssd, got %s", cfg.StoragePools[1])
	}
	if cfg.State != "INITIALIZED" {
		t.Errorf("expected state INITIALIZED, got %s", cfg.State)
	}
}

func TestVirtGlobalConfigFromResponse_NilNewFields(t *testing.T) {
	resp := VirtGlobalConfigResponse{}

	cfg := virtGlobalConfigFromResponse(resp)
	if cfg.Dataset != "" {
		t.Errorf("expected empty dataset, got %s", cfg.Dataset)
	}
	if cfg.StoragePools == nil {
		t.Error("expected non-nil storage pools slice")
	}
	if len(cfg.StoragePools) != 0 {
		t.Errorf("expected 0 storage pools, got %d", len(cfg.StoragePools))
	}
	if cfg.State != "" {
		t.Errorf("expected empty state, got %s", cfg.State)
	}
}

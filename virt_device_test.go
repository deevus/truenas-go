package truenas

import "testing"

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

package truenas

import "encoding/json"

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

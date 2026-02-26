package truenas

// VirtGlobalConfigResponse represents the global virt configuration from the API.
type VirtGlobalConfigResponse struct {
	Bridge       *string  `json:"bridge"`
	V4Network    *string  `json:"v4_network"`
	V6Network    *string  `json:"v6_network"`
	Pool         *string  `json:"pool"`
	Dataset      *string  `json:"dataset"`
	StoragePools []string `json:"storage_pools"`
	State        *string  `json:"state"`
}

// VirtInstanceResponse represents a virt instance from the API.
type VirtInstanceResponse struct {
	ID          string                      `json:"id"`
	Name        string                      `json:"name"`
	Type        string                      `json:"type"`
	Status      string                      `json:"status"`
	CPU         *string                     `json:"cpu"`
	Memory      *int64                      `json:"memory"`
	Autostart   bool                        `json:"autostart"`
	Environment map[string]string           `json:"environment"`
	Aliases     []VirtInstanceAliasResponse `json:"aliases"`
	Image       VirtInstanceImageResponse   `json:"image"`
	StoragePool string                      `json:"storage_pool"`
}

// VirtInstanceAliasResponse represents a network alias from the API.
type VirtInstanceAliasResponse struct {
	Type    string `json:"type"`
	Address string `json:"address"`
	Netmask *int64 `json:"netmask"`
}

// VirtInstanceImageResponse represents instance image metadata from the API.
type VirtInstanceImageResponse struct {
	Architecture string `json:"architecture"`
	Description  string `json:"description"`
	OS           string `json:"os"`
	Release      string `json:"release"`
	Variant      string `json:"variant"`
}

// VirtDeviceResponse represents a device attached to a virt instance from the API.
type VirtDeviceResponse struct {
	DevType     string  `json:"dev_type"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Readonly    bool    `json:"readonly"`
	// DISK fields
	Source      *string `json:"source"`
	Destination *string `json:"destination"`
	// NIC fields
	Network *string `json:"network"`
	NICType *string `json:"nic_type"`
	Parent  *string `json:"parent"`
	// PROXY fields
	SourceProto *string `json:"source_proto"`
	SourcePort  *int64  `json:"source_port"`
	DestProto   *string `json:"dest_proto"`
	DestPort    *int64  `json:"dest_port"`
}

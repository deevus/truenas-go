package truenas

// InterfaceType represents the type of a network interface.
type InterfaceType string

const (
	InterfaceTypePhysical InterfaceType = "PHYSICAL"
	InterfaceTypeBridge   InterfaceType = "BRIDGE"
	InterfaceTypeLAGG     InterfaceType = "LINK_AGGREGATION"
	InterfaceTypeVLAN     InterfaceType = "VLAN"
)

// LinkState represents the link state of a network interface.
type LinkState string

const (
	LinkStateUp   LinkState = "LINK_STATE_UP"
	LinkStateDown LinkState = "LINK_STATE_DOWN"
)

// AliasType represents the type of an interface alias (IPv4 or IPv6).
type AliasType string

const (
	AliasTypeINET  AliasType = "INET"
	AliasTypeINET6 AliasType = "INET6"
)

// InterfaceResponse represents a network interface from the query API.
type InterfaceResponse struct {
	ID          string                   `json:"id"`
	Name        string                   `json:"name"`
	Type        string                   `json:"type"`
	Description string                   `json:"description"`
	MTU         int                      `json:"mtu"`
	State       InterfaceStateResponse   `json:"state"`
	Aliases     []InterfaceAliasResponse `json:"aliases"`
}

// InterfaceStateResponse represents the state sub-object of a network interface.
type InterfaceStateResponse struct {
	Name               string `json:"name"`
	LinkState          string `json:"link_state"`
	ActiveMediaType    string `json:"active_media_type"`
	ActiveMediaSubtype string `json:"active_media_subtype"`
}

// InterfaceAliasResponse represents an IP alias on a network interface.
type InterfaceAliasResponse struct {
	Type    string `json:"type"`
	Address string `json:"address"`
	Netmask int    `json:"netmask"`
}

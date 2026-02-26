package truenas

// NetworkSummaryResponse represents the network.general.summary API response.
type NetworkSummaryResponse struct {
	IPs           map[string]NetworkInterfaceIPsResponse `json:"ips"`
	DefaultRoutes []string                               `json:"default_routes"`
	Nameservers   []string                               `json:"nameservers"`
}

// NetworkInterfaceIPsResponse represents per-interface IPs from the API.
type NetworkInterfaceIPsResponse struct {
	IPV4 []string `json:"IPV4"`
	IPV6 []string `json:"IPV6"`
}

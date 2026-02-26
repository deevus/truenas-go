package truenas

import (
	"context"
	"encoding/json"
	"fmt"
)

// NetworkSummary is the user-facing representation of network.general.summary.
type NetworkSummary struct {
	IPs           map[string]NetworkInterfaceIPs
	DefaultRoutes []string
	Nameservers   []string
}

// NetworkInterfaceIPs contains the IPv4 and IPv6 addresses for an interface.
type NetworkInterfaceIPs struct {
	IPV4 []string
	IPV6 []string
}

// NetworkService provides typed methods for the network.* API namespace.
type NetworkService struct {
	client  Caller
	version Version
}

// NewNetworkService creates a new NetworkService.
func NewNetworkService(c Caller, v Version) *NetworkService {
	return &NetworkService{client: c, version: v}
}

// GetSummary returns general network information including default routes,
// nameservers, and per-interface IP addresses.
func (s *NetworkService) GetSummary(ctx context.Context) (*NetworkSummary, error) {
	result, err := s.client.Call(ctx, "network.general.summary", nil)
	if err != nil {
		return nil, err
	}

	var resp NetworkSummaryResponse
	if err := json.Unmarshal(result, &resp); err != nil {
		return nil, fmt.Errorf("parse network summary response: %w", err)
	}

	summary := networkSummaryFromResponse(resp)
	return &summary, nil
}

// networkSummaryFromResponse converts a wire-format response to a user-facing summary.
func networkSummaryFromResponse(resp NetworkSummaryResponse) NetworkSummary {
	ips := make(map[string]NetworkInterfaceIPs, len(resp.IPs))
	for iface, ipResp := range resp.IPs {
		ipv4 := ipResp.IPV4
		if ipv4 == nil {
			ipv4 = []string{}
		}
		ipv6 := ipResp.IPV6
		if ipv6 == nil {
			ipv6 = []string{}
		}
		ips[iface] = NetworkInterfaceIPs{
			IPV4: ipv4,
			IPV6: ipv6,
		}
	}

	defaultRoutes := resp.DefaultRoutes
	if defaultRoutes == nil {
		defaultRoutes = []string{}
	}

	nameservers := resp.Nameservers
	if nameservers == nil {
		nameservers = []string{}
	}

	return NetworkSummary{
		IPs:           ips,
		DefaultRoutes: defaultRoutes,
		Nameservers:   nameservers,
	}
}

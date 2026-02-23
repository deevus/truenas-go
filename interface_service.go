package truenas

import (
	"context"
	"encoding/json"
	"fmt"
)

// NetworkInterface is the user-facing representation of a TrueNAS network interface.
type NetworkInterface struct {
	ID          string
	Name        string
	Type        InterfaceType
	Description string
	MTU         int
	State       InterfaceState
	Aliases     []InterfaceAlias
}

// InterfaceState is the user-facing representation of a network interface's link state.
type InterfaceState struct {
	Name               string
	LinkState          LinkState
	ActiveMediaType    string
	ActiveMediaSubtype string
}

// InterfaceAlias is the user-facing representation of an IP alias on a network interface.
type InterfaceAlias struct {
	Type    AliasType
	Address string
	Netmask int
}

// InterfaceService provides typed methods for the network interface API namespace.
type InterfaceService struct {
	client  Caller
	version Version
}

// NewInterfaceService creates a new InterfaceService.
func NewInterfaceService(c Caller, v Version) *InterfaceService {
	return &InterfaceService{client: c, version: v}
}

// List returns all network interfaces.
func (s *InterfaceService) List(ctx context.Context) ([]NetworkInterface, error) {
	result, err := s.client.Call(ctx, "interface.query", nil)
	if err != nil {
		return nil, err
	}

	var responses []InterfaceResponse
	if err := json.Unmarshal(result, &responses); err != nil {
		return nil, fmt.Errorf("parse interface.query response: %w", err)
	}

	ifaces := make([]NetworkInterface, len(responses))
	for i, resp := range responses {
		ifaces[i] = networkInterfaceFromResponse(resp)
	}
	return ifaces, nil
}

// Get returns a network interface by ID, or nil if not found.
func (s *InterfaceService) Get(ctx context.Context, id string) (*NetworkInterface, error) {
	filter := [][]any{{"id", "=", id}}
	result, err := s.client.Call(ctx, "interface.query", filter)
	if err != nil {
		return nil, err
	}

	var responses []InterfaceResponse
	if err := json.Unmarshal(result, &responses); err != nil {
		return nil, fmt.Errorf("parse interface.query response: %w", err)
	}

	if len(responses) == 0 {
		return nil, nil
	}

	iface := networkInterfaceFromResponse(responses[0])
	return &iface, nil
}

// networkInterfaceFromResponse converts a wire-format InterfaceResponse to a user-facing NetworkInterface.
func networkInterfaceFromResponse(resp InterfaceResponse) NetworkInterface {
	aliases := make([]InterfaceAlias, len(resp.Aliases))
	for i, a := range resp.Aliases {
		aliases[i] = InterfaceAlias{
			Type:    AliasType(a.Type),
			Address: a.Address,
			Netmask: a.Netmask,
		}
	}
	return NetworkInterface{
		ID:          resp.ID,
		Name:        resp.Name,
		Type:        InterfaceType(resp.Type),
		Description: resp.Description,
		MTU:         resp.MTU,
		State: InterfaceState{
			Name:               resp.State.Name,
			LinkState:          LinkState(resp.State.LinkState),
			ActiveMediaType:    resp.State.ActiveMediaType,
			ActiveMediaSubtype: resp.State.ActiveMediaSubtype,
		},
		Aliases: aliases,
	}
}

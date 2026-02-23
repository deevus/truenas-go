package truenas

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func sampleInterfaceJSON() json.RawMessage {
	return json.RawMessage(`[{
		"id": "eno1",
		"name": "eno1",
		"type": "PHYSICAL",
		"state": {
			"name": "eno1",
			"link_state": "LINK_STATE_UP",
			"active_media_type": "Ethernet",
			"active_media_subtype": "1000baseT"
		},
		"aliases": [
			{"type": "INET", "address": "192.168.1.100", "netmask": 24},
			{"type": "INET6", "address": "fe80::1", "netmask": 64}
		],
		"description": "Primary NIC",
		"mtu": 1500
	}]`)
}

func TestInterfaceService_List(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "interface.query" {
				t.Errorf("expected method interface.query, got %s", method)
			}
			return sampleInterfaceJSON(), nil
		},
	}

	svc := NewInterfaceService(mock, Version{Major: 25, Minor: 4})
	ifaces, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ifaces) != 1 {
		t.Fatalf("expected 1 interface, got %d", len(ifaces))
	}
	iface := ifaces[0]
	if iface.ID != "eno1" {
		t.Errorf("expected ID eno1, got %s", iface.ID)
	}
	if iface.Type != InterfaceTypePhysical {
		t.Errorf("expected type PHYSICAL, got %s", iface.Type)
	}
	if iface.State.LinkState != LinkStateUp {
		t.Errorf("expected LINK_STATE_UP, got %s", iface.State.LinkState)
	}
	if len(iface.Aliases) != 2 {
		t.Fatalf("expected 2 aliases, got %d", len(iface.Aliases))
	}
	if iface.Aliases[0].Address != "192.168.1.100" {
		t.Errorf("expected address 192.168.1.100, got %s", iface.Aliases[0].Address)
	}
	if iface.MTU != 1500 {
		t.Errorf("expected MTU 1500, got %d", iface.MTU)
	}
}

func TestInterfaceService_List_Empty(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return json.RawMessage(`[]`), nil
		},
	}

	svc := NewInterfaceService(mock, Version{Major: 25, Minor: 4})
	ifaces, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ifaces) != 0 {
		t.Errorf("expected 0 interfaces, got %d", len(ifaces))
	}
}

func TestInterfaceService_List_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("network error")
		},
	}

	svc := NewInterfaceService(mock, Version{Major: 25, Minor: 4})
	_, err := svc.List(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestInterfaceService_List_ParseError(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return json.RawMessage(`not json`), nil
		},
	}

	svc := NewInterfaceService(mock, Version{Major: 25, Minor: 4})
	_, err := svc.List(context.Background())
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestInterfaceService_Get(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			filter := params.([][]any)
			if filter[0][2] != "eno1" {
				t.Errorf("expected filter for eno1, got %v", filter)
			}
			return sampleInterfaceJSON(), nil
		},
	}

	svc := NewInterfaceService(mock, Version{Major: 25, Minor: 4})
	iface, err := svc.Get(context.Background(), "eno1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if iface == nil {
		t.Fatal("expected non-nil interface")
	}
	if iface.ID != "eno1" {
		t.Errorf("expected ID eno1, got %s", iface.ID)
	}
}

func TestInterfaceService_Get_NotFound(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return json.RawMessage(`[]`), nil
		},
	}

	svc := NewInterfaceService(mock, Version{Major: 25, Minor: 4})
	iface, err := svc.Get(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if iface != nil {
		t.Error("expected nil interface for not found")
	}
}

func TestInterfaceService_Get_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("timeout")
		},
	}

	svc := NewInterfaceService(mock, Version{Major: 25, Minor: 4})
	_, err := svc.Get(context.Background(), "eno1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestInterfaceService_Get_ParseError(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return json.RawMessage(`not json`), nil
		},
	}

	svc := NewInterfaceService(mock, Version{Major: 25, Minor: 4})
	_, err := svc.Get(context.Background(), "eno1")
	if err == nil {
		t.Fatal("expected parse error")
	}
}

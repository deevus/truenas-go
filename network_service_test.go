package truenas

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestNetworkService_GetSummary(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "network.general.summary" {
				t.Errorf("expected method network.general.summary, got %s", method)
			}
			if params != nil {
				t.Error("expected nil params")
			}
			return json.RawMessage(`{
				"ips": {
					"eno1": {
						"IPV4": ["192.168.1.10/24"],
						"IPV6": ["fd00::10/64"]
					},
					"lo": {
						"IPV4": ["127.0.0.1/8"],
						"IPV6": ["::1/128"]
					}
				},
				"default_routes": ["192.168.1.1"],
				"nameservers": ["1.1.1.1", "8.8.8.8"]
			}`), nil
		},
	}

	svc := NewNetworkService(mock, Version{})
	summary, err := svc.GetSummary(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary == nil {
		t.Fatal("expected non-nil summary")
	}

	// Check IPs
	if len(summary.IPs) != 2 {
		t.Fatalf("expected 2 interfaces, got %d", len(summary.IPs))
	}
	eno1 := summary.IPs["eno1"]
	if len(eno1.IPV4) != 1 || eno1.IPV4[0] != "192.168.1.10/24" {
		t.Errorf("expected eno1 IPV4 [192.168.1.10/24], got %v", eno1.IPV4)
	}
	if len(eno1.IPV6) != 1 || eno1.IPV6[0] != "fd00::10/64" {
		t.Errorf("expected eno1 IPV6 [fd00::10/64], got %v", eno1.IPV6)
	}

	// Check default routes
	if len(summary.DefaultRoutes) != 1 || summary.DefaultRoutes[0] != "192.168.1.1" {
		t.Errorf("expected default routes [192.168.1.1], got %v", summary.DefaultRoutes)
	}

	// Check nameservers
	if len(summary.Nameservers) != 2 {
		t.Fatalf("expected 2 nameservers, got %d", len(summary.Nameservers))
	}
	if summary.Nameservers[0] != "1.1.1.1" {
		t.Errorf("expected first nameserver 1.1.1.1, got %s", summary.Nameservers[0])
	}
}

func TestNetworkService_GetSummary_Empty(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return json.RawMessage(`{
				"ips": {},
				"default_routes": [],
				"nameservers": []
			}`), nil
		},
	}

	svc := NewNetworkService(mock, Version{})
	summary, err := svc.GetSummary(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(summary.IPs) != 0 {
		t.Errorf("expected 0 interfaces, got %d", len(summary.IPs))
	}
	if len(summary.DefaultRoutes) != 0 {
		t.Errorf("expected 0 default routes, got %d", len(summary.DefaultRoutes))
	}
	if len(summary.Nameservers) != 0 {
		t.Errorf("expected 0 nameservers, got %d", len(summary.Nameservers))
	}
}

func TestNetworkService_GetSummary_NullFields(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return json.RawMessage(`{
				"ips": {"eno1": {"IPV4": null, "IPV6": null}},
				"default_routes": null,
				"nameservers": null
			}`), nil
		},
	}

	svc := NewNetworkService(mock, Version{})
	summary, err := svc.GetSummary(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.DefaultRoutes == nil {
		t.Error("expected non-nil default routes")
	}
	if len(summary.DefaultRoutes) != 0 {
		t.Errorf("expected 0 default routes, got %d", len(summary.DefaultRoutes))
	}
	if summary.Nameservers == nil {
		t.Error("expected non-nil nameservers")
	}
	eno1 := summary.IPs["eno1"]
	if eno1.IPV4 == nil {
		t.Error("expected non-nil IPV4")
	}
	if eno1.IPV6 == nil {
		t.Error("expected non-nil IPV6")
	}
}

func TestNetworkService_GetSummary_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("connection refused")
		},
	}

	svc := NewNetworkService(mock, Version{})
	summary, err := svc.GetSummary(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if summary != nil {
		t.Error("expected nil summary on error")
	}
}

func TestNetworkService_GetSummary_ParseError(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return json.RawMessage(`not json`), nil
		},
	}

	svc := NewNetworkService(mock, Version{})
	_, err := svc.GetSummary(context.Background())
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestNetworkSummaryFromResponse(t *testing.T) {
	resp := NetworkSummaryResponse{
		IPs: map[string]NetworkInterfaceIPsResponse{
			"eth0": {IPV4: []string{"10.0.0.1/24"}, IPV6: []string{"fd00::1/64"}},
		},
		DefaultRoutes: []string{"10.0.0.1"},
		Nameservers:   []string{"8.8.8.8"},
	}

	summary := networkSummaryFromResponse(resp)
	if len(summary.IPs) != 1 {
		t.Fatalf("expected 1 interface, got %d", len(summary.IPs))
	}
	eth0 := summary.IPs["eth0"]
	if len(eth0.IPV4) != 1 || eth0.IPV4[0] != "10.0.0.1/24" {
		t.Errorf("expected IPV4 [10.0.0.1/24], got %v", eth0.IPV4)
	}
	if len(summary.DefaultRoutes) != 1 || summary.DefaultRoutes[0] != "10.0.0.1" {
		t.Errorf("expected default routes [10.0.0.1], got %v", summary.DefaultRoutes)
	}
}

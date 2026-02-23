package truenas

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

// sampleSystemInfoJSON returns a JSON response for system.info.
func sampleSystemInfoJSON() json.RawMessage {
	return json.RawMessage(`{
		"model": "Intel(R) Xeon(R) E-2278G",
		"cores": 16,
		"physical_cores": 8,
		"hostname": "truenas.local",
		"uptime": "7 days, 3:42:15.123456",
		"uptime_seconds": 617535.123456,
		"loadavg": [0.5, 1.2, 0.8],
		"ecc_memory": true
	}`)
}

// --- GetInfo tests ---

func TestSystemService_GetInfo(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "system.info" {
				t.Errorf("expected method system.info, got %s", method)
			}
			if params != nil {
				t.Errorf("expected nil params, got %v", params)
			}
			return sampleSystemInfoJSON(), nil
		},
	}

	svc := NewSystemService(mock, Version{})
	info, err := svc.GetInfo(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info == nil {
		t.Fatal("expected non-nil info")
	}
	if info.Model != "Intel(R) Xeon(R) E-2278G" {
		t.Errorf("expected model Intel(R) Xeon(R) E-2278G, got %s", info.Model)
	}
	if info.Cores != 16 {
		t.Errorf("expected 16 cores, got %d", info.Cores)
	}
	if info.PhysicalCores != 8 {
		t.Errorf("expected 8 physical cores, got %d", info.PhysicalCores)
	}
	if info.Hostname != "truenas.local" {
		t.Errorf("expected hostname truenas.local, got %s", info.Hostname)
	}
	if info.Uptime != "7 days, 3:42:15.123456" {
		t.Errorf("expected uptime '7 days, 3:42:15.123456', got %s", info.Uptime)
	}
	if info.UptimeSeconds != 617535.123456 {
		t.Errorf("expected uptime_seconds 617535.123456, got %f", info.UptimeSeconds)
	}
	if info.LoadAvg != [3]float64{0.5, 1.2, 0.8} {
		t.Errorf("expected loadavg [0.5 1.2 0.8], got %v", info.LoadAvg)
	}
	if !info.EccMemory {
		t.Error("expected ecc_memory true")
	}
}

func TestSystemService_GetInfo_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("connection refused")
		},
	}

	svc := NewSystemService(mock, Version{})
	info, err := svc.GetInfo(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if info != nil {
		t.Error("expected nil info on error")
	}
	if err.Error() != "connection refused" {
		t.Errorf("expected 'connection refused', got %q", err.Error())
	}
}

func TestSystemService_GetInfo_ParseError(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return json.RawMessage(`not json`), nil
		},
	}

	svc := NewSystemService(mock, Version{})
	info, err := svc.GetInfo(context.Background())
	if err == nil {
		t.Fatal("expected parse error")
	}
	if info != nil {
		t.Error("expected nil info on parse error")
	}
}

// --- GetVersion tests ---

func TestSystemService_GetVersion(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "system.version" {
				t.Errorf("expected method system.version, got %s", method)
			}
			if params != nil {
				t.Errorf("expected nil params, got %v", params)
			}
			return json.RawMessage(`"TrueNAS-SCALE-24.10.0"`), nil
		},
	}

	svc := NewSystemService(mock, Version{})
	version, err := svc.GetVersion(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if version != "TrueNAS-SCALE-24.10.0" {
		t.Errorf("expected TrueNAS-SCALE-24.10.0, got %s", version)
	}
}

func TestSystemService_GetVersion_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("timeout")
		},
	}

	svc := NewSystemService(mock, Version{})
	version, err := svc.GetVersion(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if version != "" {
		t.Errorf("expected empty version on error, got %q", version)
	}
}

func TestSystemService_GetVersion_ParseError(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return json.RawMessage(`not json`), nil
		},
	}

	svc := NewSystemService(mock, Version{})
	version, err := svc.GetVersion(context.Background())
	if err == nil {
		t.Fatal("expected parse error")
	}
	if version != "" {
		t.Errorf("expected empty version on parse error, got %q", version)
	}
}

// --- Conversion tests ---

func TestSystemInfoFromResponse(t *testing.T) {
	resp := SystemInfoResponse{
		Model:         "AMD EPYC 7302P",
		Cores:         32,
		PhysicalCores: 16,
		Hostname:      "nas.example.com",
		Uptime:        "1 day, 2:30:00.000000",
		UptimeSeconds: 95400.0,
		LoadAvg:       [3]float64{1.0, 2.0, 3.0},
		EccMemory:     false,
	}

	info := systemInfoFromResponse(resp)

	if info.Model != "AMD EPYC 7302P" {
		t.Errorf("expected model AMD EPYC 7302P, got %s", info.Model)
	}
	if info.Cores != 32 {
		t.Errorf("expected 32 cores, got %d", info.Cores)
	}
	if info.PhysicalCores != 16 {
		t.Errorf("expected 16 physical cores, got %d", info.PhysicalCores)
	}
	if info.Hostname != "nas.example.com" {
		t.Errorf("expected hostname nas.example.com, got %s", info.Hostname)
	}
	if info.Uptime != "1 day, 2:30:00.000000" {
		t.Errorf("expected uptime '1 day, 2:30:00.000000', got %s", info.Uptime)
	}
	if info.UptimeSeconds != 95400.0 {
		t.Errorf("expected uptime_seconds 95400.0, got %f", info.UptimeSeconds)
	}
	if info.LoadAvg != [3]float64{1.0, 2.0, 3.0} {
		t.Errorf("expected loadavg [1.0 2.0 3.0], got %v", info.LoadAvg)
	}
	if info.EccMemory {
		t.Error("expected ecc_memory false")
	}
}

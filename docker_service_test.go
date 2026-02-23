package truenas

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func sampleDockerStatusJSON() json.RawMessage {
	return json.RawMessage(`{
		"status": "RUNNING",
		"description": "Docker is running"
	}`)
}

func sampleDockerConfigJSON() json.RawMessage {
	return json.RawMessage(`{
		"pool": "tank",
		"enable_image_updates": true,
		"nvidia": false,
		"address_pools": [
			{"base": "172.17.0.0/12", "size": 24}
		]
	}`)
}

// --- GetStatus tests ---

func TestDockerService_GetStatus(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "docker.status" {
				t.Errorf("expected method docker.status, got %s", method)
			}
			if params != nil {
				t.Errorf("expected nil params, got %v", params)
			}
			return sampleDockerStatusJSON(), nil
		},
	}

	svc := NewDockerService(mock, Version{Major: 25, Minor: 4})
	status, err := svc.GetStatus(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status == nil {
		t.Fatal("expected non-nil status")
	}
	if status.Status != DockerStateRunning {
		t.Errorf("expected RUNNING, got %s", status.Status)
	}
	if status.Description != "Docker is running" {
		t.Errorf("expected 'Docker is running', got %s", status.Description)
	}
}

func TestDockerService_GetStatus_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("connection refused")
		},
	}

	svc := NewDockerService(mock, Version{Major: 25, Minor: 4})
	status, err := svc.GetStatus(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if status != nil {
		t.Error("expected nil status on error")
	}
}

func TestDockerService_GetStatus_ParseError(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return json.RawMessage(`not json`), nil
		},
	}

	svc := NewDockerService(mock, Version{Major: 25, Minor: 4})
	status, err := svc.GetStatus(context.Background())
	if err == nil {
		t.Fatal("expected parse error")
	}
	if status != nil {
		t.Error("expected nil status on parse error")
	}
}

// --- GetConfig tests ---

func TestDockerService_GetConfig(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "docker.config" {
				t.Errorf("expected method docker.config, got %s", method)
			}
			if params != nil {
				t.Errorf("expected nil params, got %v", params)
			}
			return sampleDockerConfigJSON(), nil
		},
	}

	svc := NewDockerService(mock, Version{Major: 25, Minor: 4})
	config, err := svc.GetConfig(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config == nil {
		t.Fatal("expected non-nil config")
	}
	if config.Pool != "tank" {
		t.Errorf("expected pool tank, got %s", config.Pool)
	}
	if !config.EnableImageUpdates {
		t.Error("expected EnableImageUpdates=true")
	}
	if config.NvidiaEnabled {
		t.Error("expected NvidiaEnabled=false")
	}
	if len(config.AddressPools) != 1 {
		t.Fatalf("expected 1 address pool, got %d", len(config.AddressPools))
	}
	if config.AddressPools[0].Base != "172.17.0.0/12" {
		t.Errorf("expected base 172.17.0.0/12, got %s", config.AddressPools[0].Base)
	}
	if config.AddressPools[0].Size != 24 {
		t.Errorf("expected size 24, got %d", config.AddressPools[0].Size)
	}
}

func TestDockerService_GetConfig_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("timeout")
		},
	}

	svc := NewDockerService(mock, Version{Major: 25, Minor: 4})
	config, err := svc.GetConfig(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if config != nil {
		t.Error("expected nil config on error")
	}
}

func TestDockerService_GetConfig_ParseError(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return json.RawMessage(`not json`), nil
		},
	}

	svc := NewDockerService(mock, Version{Major: 25, Minor: 4})
	config, err := svc.GetConfig(context.Background())
	if err == nil {
		t.Fatal("expected parse error")
	}
	if config != nil {
		t.Error("expected nil config on parse error")
	}
}

// --- Conversion tests ---

func TestDockerStatusFromResponse(t *testing.T) {
	resp := DockerStatusResponse{
		Status:      "STOPPED",
		Description: "Docker is stopped",
	}
	status := dockerStatusFromResponse(resp)
	if status.Status != DockerStateStopped {
		t.Errorf("expected STOPPED, got %s", status.Status)
	}
	if status.Description != "Docker is stopped" {
		t.Errorf("expected 'Docker is stopped', got %s", status.Description)
	}
}

func TestDockerConfigFromResponse(t *testing.T) {
	resp := DockerConfigResponse{
		Pool:               "data",
		EnableImageUpdates: false,
		NvidiaEnabled:      true,
		AddressPoolsV4: []DockerAddressPoolResponse{
			{Base: "10.0.0.0/8", Size: 16},
			{Base: "192.168.0.0/16", Size: 24},
		},
	}
	config := dockerConfigFromResponse(resp)
	if config.Pool != "data" {
		t.Errorf("expected pool data, got %s", config.Pool)
	}
	if config.EnableImageUpdates {
		t.Error("expected EnableImageUpdates=false")
	}
	if !config.NvidiaEnabled {
		t.Error("expected NvidiaEnabled=true")
	}
	if len(config.AddressPools) != 2 {
		t.Fatalf("expected 2 address pools, got %d", len(config.AddressPools))
	}
}

package truenas

import (
	"context"
	"testing"
)

func TestMockDockerService_ImplementsInterface(t *testing.T) {
	var _ DockerServiceAPI = (*DockerService)(nil)
	var _ DockerServiceAPI = (*MockDockerService)(nil)
}

func TestMockDockerService_DefaultsToNil(t *testing.T) {
	mock := &MockDockerService{}
	ctx := context.Background()

	status, err := mock.GetStatus(ctx)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if status != nil {
		t.Fatalf("expected nil result, got: %v", status)
	}

	config, err := mock.GetConfig(ctx)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if config != nil {
		t.Fatalf("expected nil result, got: %v", config)
	}
}

func TestMockDockerService_CallsGetStatusFunc(t *testing.T) {
	called := false
	mock := &MockDockerService{
		GetStatusFunc: func(ctx context.Context) (*DockerStatus, error) {
			called = true
			return &DockerStatus{Status: DockerStateRunning, Description: "running"}, nil
		},
	}

	status, err := mock.GetStatus(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected GetStatusFunc to be called")
	}
	if status.Status != DockerStateRunning {
		t.Fatalf("expected RUNNING, got %s", status.Status)
	}
}

func TestMockDockerService_CallsGetConfigFunc(t *testing.T) {
	called := false
	mock := &MockDockerService{
		GetConfigFunc: func(ctx context.Context) (*DockerConfig, error) {
			called = true
			return &DockerConfig{Pool: "tank"}, nil
		},
	}

	config, err := mock.GetConfig(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected GetConfigFunc to be called")
	}
	if config.Pool != "tank" {
		t.Fatalf("expected pool tank, got %s", config.Pool)
	}
}

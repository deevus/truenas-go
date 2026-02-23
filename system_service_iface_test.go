package truenas

import (
	"context"
	"testing"
)

func TestMockSystemService_ImplementsInterface(t *testing.T) {
	// Compile-time check
	var _ SystemServiceAPI = (*SystemService)(nil)
	var _ SystemServiceAPI = (*MockSystemService)(nil)
}

func TestMockSystemService_DefaultsToNil(t *testing.T) {
	mock := &MockSystemService{}
	ctx := context.Background()

	info, err := mock.GetInfo(ctx)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if info != nil {
		t.Fatalf("expected nil result, got: %v", info)
	}

	version, err := mock.GetVersion(ctx)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if version != "" {
		t.Fatalf("expected empty string, got: %q", version)
	}
}

func TestMockSystemService_CallsFunc(t *testing.T) {
	called := false
	mock := &MockSystemService{
		GetInfoFunc: func(ctx context.Context) (*SystemInfo, error) {
			called = true
			return &SystemInfo{Hostname: "truenas.local"}, nil
		},
	}

	info, err := mock.GetInfo(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected GetInfoFunc to be called")
	}
	if info.Hostname != "truenas.local" {
		t.Fatalf("expected hostname truenas.local, got %s", info.Hostname)
	}
}

func TestMockSystemService_GetVersionCallsFunc(t *testing.T) {
	called := false
	mock := &MockSystemService{
		GetVersionFunc: func(ctx context.Context) (string, error) {
			called = true
			return "TrueNAS-SCALE-24.10.0", nil
		},
	}

	version, err := mock.GetVersion(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected GetVersionFunc to be called")
	}
	if version != "TrueNAS-SCALE-24.10.0" {
		t.Fatalf("expected TrueNAS-SCALE-24.10.0, got %s", version)
	}
}

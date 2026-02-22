package truenas

import (
	"context"
	"testing"
)

func TestMockFilesystemService_ImplementsInterface(t *testing.T) {
	var _ FilesystemServiceAPI = (*FilesystemService)(nil)
	var _ FilesystemServiceAPI = (*MockFilesystemService)(nil)
}

func TestMockFilesystemService_DefaultsToNil(t *testing.T) {
	mock := &MockFilesystemService{}
	ctx := context.Background()

	result, err := mock.Stat(ctx, "/mnt/pool/test")
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil result, got: %v", result)
	}

	c := mock.Client()
	if c != nil {
		t.Fatalf("expected nil client, got: %v", c)
	}
}

func TestMockFilesystemService_CallsFunc(t *testing.T) {
	called := false
	mock := &MockFilesystemService{
		StatFunc: func(ctx context.Context, path string) (*StatResult, error) {
			called = true
			return &StatResult{}, nil
		},
	}

	_, err := mock.Stat(context.Background(), "/mnt/pool/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected StatFunc to be called")
	}
}

package truenas

import (
	"context"
	"testing"
)

func TestMockSSHService_ImplementsInterface(t *testing.T) {
	var _ SSHServiceAPI = (*SSHService)(nil)
	var _ SSHServiceAPI = (*MockSSHService)(nil)
}

func TestMockSSHService_DefaultsToNil(t *testing.T) {
	mock := &MockSSHService{}
	ctx := context.Background()

	cred, err := mock.GetSSHKeyPair(ctx, 1)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if cred != nil {
		t.Fatalf("expected nil result, got: %v", cred)
	}
}

func TestMockSSHService_CallsFunc(t *testing.T) {
	called := false
	mock := &MockSSHService{
		GetSSHKeyPairFunc: func(ctx context.Context, id int64) (*SSHKeyPair, error) {
			called = true
			return &SSHKeyPair{ID: id}, nil
		},
	}

	cred, err := mock.GetSSHKeyPair(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected GetSSHKeyPairFunc to be called")
	}
	if cred.ID != 42 {
		t.Fatalf("expected ID 42, got %d", cred.ID)
	}
}

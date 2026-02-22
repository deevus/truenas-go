package truenas

import (
	"context"
	"testing"
)

func TestMockCloudSyncService_ImplementsInterface(t *testing.T) {
	var _ CloudSyncServiceAPI = (*CloudSyncService)(nil)
	var _ CloudSyncServiceAPI = (*MockCloudSyncService)(nil)
}

func TestMockCloudSyncService_DefaultsToNil(t *testing.T) {
	mock := &MockCloudSyncService{}
	ctx := context.Background()

	cred, err := mock.GetCredential(ctx, 1)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if cred != nil {
		t.Fatalf("expected nil result, got: %v", cred)
	}
}

func TestMockCloudSyncService_CallsFunc(t *testing.T) {
	called := false
	mock := &MockCloudSyncService{
		GetCredentialFunc: func(ctx context.Context, id int64) (*CloudSyncCredential, error) {
			called = true
			return &CloudSyncCredential{ID: id}, nil
		},
	}

	cred, err := mock.GetCredential(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected GetCredentialFunc to be called")
	}
	if cred.ID != 42 {
		t.Fatalf("expected ID 42, got %d", cred.ID)
	}
}

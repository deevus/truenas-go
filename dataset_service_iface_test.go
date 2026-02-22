package truenas

import (
	"context"
	"testing"
)

func TestMockDatasetService_ImplementsInterface(t *testing.T) {
	var _ DatasetServiceAPI = (*DatasetService)(nil)
	var _ DatasetServiceAPI = (*MockDatasetService)(nil)
}

func TestMockDatasetService_DefaultsToNil(t *testing.T) {
	mock := &MockDatasetService{}
	ctx := context.Background()

	ds, err := mock.GetDataset(ctx, "pool/test")
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if ds != nil {
		t.Fatalf("expected nil result, got: %v", ds)
	}
}

func TestMockDatasetService_CallsFunc(t *testing.T) {
	called := false
	mock := &MockDatasetService{
		GetDatasetFunc: func(ctx context.Context, id string) (*Dataset, error) {
			called = true
			return &Dataset{ID: id}, nil
		},
	}

	ds, err := mock.GetDataset(context.Background(), "pool/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected GetDatasetFunc to be called")
	}
	if ds.ID != "pool/test" {
		t.Fatalf("expected ID pool/test, got %s", ds.ID)
	}
}

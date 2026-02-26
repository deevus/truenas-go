package truenas

import (
	"context"
	"testing"
)

func TestMockCronService_ImplementsInterface(t *testing.T) {
	var _ CronServiceAPI = (*CronService)(nil)
	var _ CronServiceAPI = (*MockCronService)(nil)
}

func TestMockCronService_DefaultsToNil(t *testing.T) {
	mock := &MockCronService{}
	ctx := context.Background()

	job, err := mock.Get(ctx, 1)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if job != nil {
		t.Fatalf("expected nil result, got: %v", job)
	}

	err = mock.Run(ctx, 1, false)
	if err != nil {
		t.Fatalf("expected nil error from Run, got: %v", err)
	}
}

func TestMockCronService_CallsFunc(t *testing.T) {
	called := false
	mock := &MockCronService{
		GetFunc: func(ctx context.Context, id int64) (*CronJob, error) {
			called = true
			return &CronJob{ID: id}, nil
		},
	}

	job, err := mock.Get(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected GetFunc to be called")
	}
	if job.ID != 42 {
		t.Fatalf("expected ID 42, got %d", job.ID)
	}
}

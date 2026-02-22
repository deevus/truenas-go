package truenas

import (
	"context"
	"testing"
)

func TestMockAppService_ImplementsInterface(t *testing.T) {
	var _ AppServiceAPI = (*AppService)(nil)
	var _ AppServiceAPI = (*MockAppService)(nil)
}

func TestMockAppService_DefaultsToNil(t *testing.T) {
	mock := &MockAppService{}
	ctx := context.Background()

	app, err := mock.GetApp(ctx, "test-app")
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if app != nil {
		t.Fatalf("expected nil result, got: %v", app)
	}
}

func TestMockAppService_CallsFunc(t *testing.T) {
	called := false
	mock := &MockAppService{
		GetAppFunc: func(ctx context.Context, name string) (*App, error) {
			called = true
			return &App{Name: name}, nil
		},
	}

	app, err := mock.GetApp(context.Background(), "test-app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected GetAppFunc to be called")
	}
	if app.Name != "test-app" {
		t.Fatalf("expected name test-app, got %s", app.Name)
	}
}

package truenas

import (
	"context"
	"testing"
)

func TestMockReportingService_ImplementsInterface(t *testing.T) {
	var _ ReportingServiceAPI = (*ReportingService)(nil)
	var _ ReportingServiceAPI = (*MockReportingService)(nil)
}

func TestMockReportingService_DefaultsToNil(t *testing.T) {
	mock := &MockReportingService{}
	ctx := context.Background()

	graphs, err := mock.ListGraphs(ctx)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if graphs != nil {
		t.Fatalf("expected nil result, got: %v", graphs)
	}

	data, err := mock.GetData(ctx, ReportingGetDataParams{})
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if data != nil {
		t.Fatalf("expected nil result, got: %v", data)
	}

	sub, err := mock.SubscribeRealtime(ctx)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if sub != nil {
		t.Fatalf("expected nil result, got: %v", sub)
	}
}

func TestMockReportingService_CallsListGraphsFunc(t *testing.T) {
	called := false
	mock := &MockReportingService{
		ListGraphsFunc: func(ctx context.Context) ([]ReportingGraph, error) {
			called = true
			return []ReportingGraph{{Name: "cpu", Title: "CPU Usage"}}, nil
		},
	}

	graphs, err := mock.ListGraphs(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected ListGraphsFunc to be called")
	}
	if len(graphs) != 1 {
		t.Fatalf("expected 1 graph, got %d", len(graphs))
	}
	if graphs[0].Name != "cpu" {
		t.Fatalf("expected name cpu, got %s", graphs[0].Name)
	}
}

func TestMockReportingService_CallsGetDataFunc(t *testing.T) {
	called := false
	mock := &MockReportingService{
		GetDataFunc: func(ctx context.Context, params ReportingGetDataParams) ([]ReportingData, error) {
			called = true
			return []ReportingData{{Name: "cpu"}}, nil
		},
	}

	data, err := mock.GetData(context.Background(), ReportingGetDataParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected GetDataFunc to be called")
	}
	if len(data) != 1 {
		t.Fatalf("expected 1 data, got %d", len(data))
	}
}

func TestMockReportingService_CallsSubscribeRealtimeFunc(t *testing.T) {
	called := false
	ch := make(chan RealtimeUpdate)
	mock := &MockReportingService{
		SubscribeRealtimeFunc: func(ctx context.Context) (*Subscription[RealtimeUpdate], error) {
			called = true
			return &Subscription[RealtimeUpdate]{C: ch, cancel: func() { close(ch) }}, nil
		},
	}

	sub, err := mock.SubscribeRealtime(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected SubscribeRealtimeFunc to be called")
	}
	sub.Close()
}

func TestMockReportingService_SubscribeRealtimeDefaultsToNil(t *testing.T) {
	mock := &MockReportingService{}
	sub, err := mock.SubscribeRealtime(context.Background())
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if sub != nil {
		t.Fatal("expected nil subscription")
	}
}

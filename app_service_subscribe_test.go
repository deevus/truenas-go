package truenas

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestAppService_SubscribeStats(t *testing.T) {
	eventData := json.RawMessage(`[{
		"app_name": "plex",
		"memory": 536870912,
		"cpu_usage": 25.5,
		"blkio": {"read": 1024, "write": 2048},
		"networks": [{"interface_name": "eth0", "rx_bytes": 1024, "tx_bytes": 2048}]
	}]`)

	ch := make(chan json.RawMessage, 1)
	ch <- eventData

	mock := &mockSubscribeCaller{
		subscribeFunc: func(ctx context.Context, collection string, params any) (*Subscription[json.RawMessage], error) {
			if collection != "app.stats" {
				t.Errorf("expected collection app.stats, got %s", collection)
			}
			return &Subscription[json.RawMessage]{
				C:      ch,
				cancel: func() { close(ch) },
			}, nil
		},
	}

	svc := NewAppService(mock, Version{Major: 25, Minor: 4})
	sub, err := svc.SubscribeStats(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer sub.Close()

	stats := <-sub.C
	if len(stats) != 1 {
		t.Fatalf("expected 1 app stats, got %d", len(stats))
	}
	if stats[0].AppName != "plex" {
		t.Errorf("expected app name plex, got %s", stats[0].AppName)
	}
	if stats[0].CPUUsage != 25.5 {
		t.Errorf("expected CPU usage 25.5, got %f", stats[0].CPUUsage)
	}
	if stats[0].Memory != 536870912 {
		t.Errorf("expected memory 536870912, got %d", stats[0].Memory)
	}
	if len(stats[0].Networks) != 1 || stats[0].Networks[0].RxBytes != 1024 {
		t.Errorf("expected rx_bytes 1024, got %v", stats[0].Networks)
	}
}

func TestAppService_SubscribeStats_Error(t *testing.T) {
	mock := &mockSubscribeCaller{
		subscribeFunc: func(ctx context.Context, collection string, params any) (*Subscription[json.RawMessage], error) {
			return nil, errors.New("subscribe failed")
		},
	}

	svc := NewAppService(mock, Version{Major: 25, Minor: 4})
	_, err := svc.SubscribeStats(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAppService_SubscribeStats_MalformedEvent(t *testing.T) {
	ch := make(chan json.RawMessage, 2)
	ch <- json.RawMessage(`not json`)
	ch <- json.RawMessage(`[{"app_name": "valid", "memory": 0, "cpu_usage": 0, "blkio": {"read": 0, "write": 0}, "networks": []}]`)

	mock := &mockSubscribeCaller{
		subscribeFunc: func(ctx context.Context, collection string, params any) (*Subscription[json.RawMessage], error) {
			return &Subscription[json.RawMessage]{
				C:      ch,
				cancel: func() { close(ch) },
			}, nil
		},
	}

	svc := NewAppService(mock, Version{Major: 25, Minor: 4})
	sub, err := svc.SubscribeStats(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer sub.Close()

	// Should skip malformed and deliver valid
	stats := <-sub.C
	if len(stats) != 1 {
		t.Fatalf("expected 1 app stats, got %d", len(stats))
	}
	if stats[0].AppName != "valid" {
		t.Errorf("expected app name valid, got %s", stats[0].AppName)
	}
}

func TestAppService_SubscribeContainerLogs(t *testing.T) {
	eventData := json.RawMessage(`{
		"timestamp": "2025-01-15T10:30:00Z",
		"data": "Server started on port 8080"
	}`)

	ch := make(chan json.RawMessage, 1)
	ch <- eventData

	mock := &mockSubscribeCaller{
		subscribeFunc: func(ctx context.Context, collection string, params any) (*Subscription[json.RawMessage], error) {
			if collection != "app.container_log_follow" {
				t.Errorf("expected collection app.container_log_follow, got %s", collection)
			}
			p := params.(map[string]any)
			if p["app_name"] != "plex" {
				t.Errorf("expected app_name plex, got %v", p["app_name"])
			}
			if p["container_id"] != "abc123" {
				t.Errorf("expected container_id abc123, got %v", p["container_id"])
			}
			if p["tail_lines"] != 100 {
				t.Errorf("expected tail_lines 100, got %v", p["tail_lines"])
			}
			return &Subscription[json.RawMessage]{
				C:      ch,
				cancel: func() { close(ch) },
			}, nil
		},
	}

	svc := NewAppService(mock, Version{Major: 25, Minor: 4})
	sub, err := svc.SubscribeContainerLogs(context.Background(), ContainerLogOpts{
		AppName:     "plex",
		ContainerID: "abc123",
		TailLines:   100,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer sub.Close()

	entry := <-sub.C
	if entry.Timestamp != "2025-01-15T10:30:00Z" {
		t.Errorf("expected timestamp 2025-01-15T10:30:00Z, got %s", entry.Timestamp)
	}
	if entry.Message != "Server started on port 8080" {
		t.Errorf("expected message 'Server started on port 8080', got %s", entry.Message)
	}
}

func TestAppService_SubscribeContainerLogs_Error(t *testing.T) {
	mock := &mockSubscribeCaller{
		subscribeFunc: func(ctx context.Context, collection string, params any) (*Subscription[json.RawMessage], error) {
			return nil, errors.New("subscribe failed")
		},
	}

	svc := NewAppService(mock, Version{Major: 25, Minor: 4})
	_, err := svc.SubscribeContainerLogs(context.Background(), ContainerLogOpts{
		AppName:     "plex",
		ContainerID: "abc123",
		TailLines:   50,
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

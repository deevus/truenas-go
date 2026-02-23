package truenas

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestReportingService_SubscribeRealtime(t *testing.T) {
	eventData := json.RawMessage(`{
		"cpu": {"cpu": {"usage": 42.5, "temperature": 65.0}},
		"memory": {
			"physical_memory_total": 17179869184,
			"physical_memory_available": 8589934592,
			"arc_size": 4294967296
		},
		"disks": {"read_ops": 10, "read_bytes": 1024, "write_ops": 20, "write_bytes": 2048, "busy": 15.5},
		"interfaces": {
			"eno1": {"received_bytes_rate": 50000.0, "sent_bytes_rate": 25000.0, "link_state": "LINK_STATE_UP", "speed": 1000}
		}
	}`)

	rawCh := make(chan json.RawMessage, 1)
	rawCh <- eventData

	mock := &mockSubscribeCaller{
		subscribeFunc: func(ctx context.Context, collection string, params any) (*Subscription[json.RawMessage], error) {
			if collection != "reporting.realtime" {
				t.Errorf("expected collection reporting.realtime, got %s", collection)
			}
			return &Subscription[json.RawMessage]{
				C:      rawCh,
				cancel: func() { close(rawCh) },
			}, nil
		},
	}

	svc := NewReportingService(mock, Version{Major: 25, Minor: 4})
	sub, err := svc.SubscribeRealtime(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer sub.Close()

	select {
	case update := <-sub.C:
		// Check CPU
		if cpu, ok := update.CPU["cpu"]; !ok {
			t.Error("expected cpu key in CPU map")
		} else {
			if cpu.Usage != 42.5 {
				t.Errorf("expected CPU usage 42.5, got %f", cpu.Usage)
			}
			if cpu.Temperature != 65.0 {
				t.Errorf("expected CPU temp 65.0, got %f", cpu.Temperature)
			}
		}
		// Check memory
		if update.Memory.PhysicalTotal != 17179869184 {
			t.Errorf("unexpected physical memory total: %d", update.Memory.PhysicalTotal)
		}
		if update.Memory.PhysicalAvailable != 8589934592 {
			t.Errorf("unexpected physical memory available: %d", update.Memory.PhysicalAvailable)
		}
		if update.Memory.ArcSize != 4294967296 {
			t.Errorf("unexpected arc size: %d", update.Memory.ArcSize)
		}
		// Check disks
		if update.Disks.ReadBytes != 1024.0 {
			t.Errorf("expected read bytes 1024.0, got %f", update.Disks.ReadBytes)
		}
		if update.Disks.BusyPercent != 15.5 {
			t.Errorf("expected busy percent 15.5, got %f", update.Disks.BusyPercent)
		}
		// Check interfaces
		if iface, ok := update.Interfaces["eno1"]; !ok {
			t.Error("expected eno1 key in Interfaces map")
		} else {
			if iface.ReceivedBytesRate != 50000.0 {
				t.Errorf("expected received rate 50000.0, got %f", iface.ReceivedBytesRate)
			}
			if iface.Speed != 1000 {
				t.Errorf("expected speed 1000, got %d", iface.Speed)
			}
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for realtime update")
	}
}

func TestReportingService_SubscribeRealtime_Error(t *testing.T) {
	mock := &mockSubscribeCaller{
		subscribeFunc: func(ctx context.Context, collection string, params any) (*Subscription[json.RawMessage], error) {
			return nil, errors.New("subscription failed")
		},
	}

	svc := NewReportingService(mock, Version{Major: 25, Minor: 4})
	sub, err := svc.SubscribeRealtime(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if sub != nil {
		t.Error("expected nil subscription on error")
	}
}

func TestReportingService_SubscribeRealtime_MalformedEvent(t *testing.T) {
	rawCh := make(chan json.RawMessage, 2)
	rawCh <- json.RawMessage(`not json`)
	rawCh <- json.RawMessage(`{"cpu": {"cpu": {"usage": 10.0, "temperature": 50.0}}, "memory": {"physical_memory_total": 1024, "physical_memory_available": 512, "arc_size": 256}, "disks": {"read_ops": 0, "read_bytes": 0, "write_ops": 0, "write_bytes": 0, "busy": 0}, "interfaces": {}}`)
	close(rawCh)

	mock := &mockSubscribeCaller{
		subscribeFunc: func(ctx context.Context, collection string, params any) (*Subscription[json.RawMessage], error) {
			return &Subscription[json.RawMessage]{
				C:      rawCh,
				cancel: func() {},
			}, nil
		},
	}

	svc := NewReportingService(mock, Version{Major: 25, Minor: 4})
	sub, err := svc.SubscribeRealtime(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should skip the malformed event and receive the valid one
	select {
	case update := <-sub.C:
		if update.CPU["cpu"].Usage != 10.0 {
			t.Errorf("expected CPU usage 10.0, got %f", update.CPU["cpu"].Usage)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for valid update after malformed event")
	}
}

package truenas

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

// sampleSnapshotJSON returns a JSON response for a single snapshot with no hold.
func sampleSnapshotJSON() json.RawMessage {
	return json.RawMessage(`[{
		"id": "pool/dataset@snap1",
		"name": "pool/dataset@snap1",
		"snapshot_name": "snap1",
		"dataset": "pool/dataset",
		"properties": {
			"createtxg": {"value": "12345"},
			"used": {"parsed": 1024},
			"referenced": {"parsed": 2048},
			"userrefs": {"parsed": "0"}
		}
	}]`)
}

// sampleSnapshotWithHoldJSON returns a JSON response for a single snapshot with a hold.
func sampleSnapshotWithHoldJSON() json.RawMessage {
	return json.RawMessage(`[{
		"id": "pool/dataset@snap1",
		"name": "pool/dataset@snap1",
		"snapshot_name": "snap1",
		"dataset": "pool/dataset",
		"properties": {
			"createtxg": {"value": "12345"},
			"used": {"parsed": 1024},
			"referenced": {"parsed": 2048},
			"userrefs": {"parsed": "1"}
		}
	}]`)
}

func TestSnapshotService_Create(t *testing.T) {
	callCount := 0
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			callCount++
			if callCount == 1 {
				if method != "zfs.snapshot.create" {
					t.Errorf("expected method zfs.snapshot.create, got %s", method)
				}
				p := params.(map[string]any)
				if p["dataset"] != "pool/dataset" {
					t.Errorf("expected dataset pool/dataset, got %v", p["dataset"])
				}
				if p["name"] != "snap1" {
					t.Errorf("expected name snap1, got %v", p["name"])
				}
				if _, ok := p["recursive"]; ok {
					t.Error("expected no recursive key when Recursive=false")
				}
				return json.RawMessage(`null`), nil
			}
			// query for re-read
			if method != "zfs.snapshot.query" {
				t.Errorf("expected method zfs.snapshot.query, got %s", method)
			}
			return sampleSnapshotJSON(), nil
		},
	}

	svc := NewSnapshotService(mock, Version{Major: 24, Minor: 10})
	snap, err := svc.Create(context.Background(), CreateSnapshotOpts{
		Dataset: "pool/dataset",
		Name:    "snap1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snap == nil {
		t.Fatal("expected non-nil snapshot")
	}
	if snap.ID != "pool/dataset@snap1" {
		t.Errorf("expected ID pool/dataset@snap1, got %s", snap.ID)
	}
	if snap.Dataset != "pool/dataset" {
		t.Errorf("expected dataset pool/dataset, got %s", snap.Dataset)
	}
	if snap.SnapshotName != "snap1" {
		t.Errorf("expected snapshot name snap1, got %s", snap.SnapshotName)
	}
	if snap.CreateTXG != "12345" {
		t.Errorf("expected createtxg 12345, got %s", snap.CreateTXG)
	}
	if snap.Used != 1024 {
		t.Errorf("expected used 1024, got %d", snap.Used)
	}
	if snap.Referenced != 2048 {
		t.Errorf("expected referenced 2048, got %d", snap.Referenced)
	}
	if snap.HasHold {
		t.Error("expected HasHold=false")
	}
}

func TestSnapshotService_Create_Recursive(t *testing.T) {
	callCount := 0
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			callCount++
			if callCount == 1 {
				p := params.(map[string]any)
				if p["recursive"] != true {
					t.Errorf("expected recursive=true, got %v", p["recursive"])
				}
				return json.RawMessage(`null`), nil
			}
			return sampleSnapshotJSON(), nil
		},
	}

	svc := NewSnapshotService(mock, Version{Major: 24, Minor: 10})
	_, err := svc.Create(context.Background(), CreateSnapshotOpts{
		Dataset:   "pool/dataset",
		Name:      "snap1",
		Recursive: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSnapshotService_Create_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("connection refused")
		},
	}

	svc := NewSnapshotService(mock, Version{Major: 24, Minor: 10})
	snap, err := svc.Create(context.Background(), CreateSnapshotOpts{
		Dataset: "pool/dataset",
		Name:    "snap1",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if snap != nil {
		t.Error("expected nil snapshot on error")
	}
	if err.Error() != "connection refused" {
		t.Errorf("expected 'connection refused', got %q", err.Error())
	}
}

func TestSnapshotService_Create_GetError(t *testing.T) {
	callCount := 0
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			callCount++
			if callCount == 1 {
				return json.RawMessage(`null`), nil
			}
			return nil, errors.New("query failed")
		},
	}

	svc := NewSnapshotService(mock, Version{Major: 24, Minor: 10})
	snap, err := svc.Create(context.Background(), CreateSnapshotOpts{
		Dataset: "pool/dataset",
		Name:    "snap1",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if snap != nil {
		t.Error("expected nil snapshot on error")
	}
}

func TestSnapshotService_Create_ParseError(t *testing.T) {
	callCount := 0
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			callCount++
			if callCount == 1 {
				return json.RawMessage(`null`), nil
			}
			return json.RawMessage(`not json`), nil
		},
	}

	svc := NewSnapshotService(mock, Version{Major: 24, Minor: 10})
	_, err := svc.Create(context.Background(), CreateSnapshotOpts{
		Dataset: "pool/dataset",
		Name:    "snap1",
	})
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestSnapshotService_Create_VersionPrefix(t *testing.T) {
	callCount := 0
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			callCount++
			if callCount == 1 {
				if method != "pool.snapshot.create" {
					t.Errorf("expected method pool.snapshot.create, got %s", method)
				}
				return json.RawMessage(`null`), nil
			}
			return sampleSnapshotJSON(), nil
		},
	}

	svc := NewSnapshotService(mock, Version{Major: 25, Minor: 10})
	_, err := svc.Create(context.Background(), CreateSnapshotOpts{
		Dataset: "pool/dataset",
		Name:    "snap1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSnapshotService_Get(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "zfs.snapshot.query" {
				t.Errorf("expected method zfs.snapshot.query, got %s", method)
			}
			filter := params.([][]any)
			if len(filter) != 1 || filter[0][0] != "id" || filter[0][1] != "=" || filter[0][2] != "pool/dataset@snap1" {
				t.Errorf("unexpected filter: %v", filter)
			}
			return sampleSnapshotJSON(), nil
		},
	}

	svc := NewSnapshotService(mock, Version{Major: 24, Minor: 10})
	snap, err := svc.Get(context.Background(), "pool/dataset@snap1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snap == nil {
		t.Fatal("expected non-nil snapshot")
	}
	if snap.ID != "pool/dataset@snap1" {
		t.Errorf("expected ID pool/dataset@snap1, got %s", snap.ID)
	}
	if snap.Dataset != "pool/dataset" {
		t.Errorf("expected dataset pool/dataset, got %s", snap.Dataset)
	}
	if snap.SnapshotName != "snap1" {
		t.Errorf("expected snapshot name snap1, got %s", snap.SnapshotName)
	}
	if snap.CreateTXG != "12345" {
		t.Errorf("expected createtxg 12345, got %s", snap.CreateTXG)
	}
	if snap.Used != 1024 {
		t.Errorf("expected used 1024, got %d", snap.Used)
	}
	if snap.Referenced != 2048 {
		t.Errorf("expected referenced 2048, got %d", snap.Referenced)
	}
	if snap.HasHold {
		t.Error("expected HasHold=false")
	}
}

func TestSnapshotService_Get_WithHold(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return sampleSnapshotWithHoldJSON(), nil
		},
	}

	svc := NewSnapshotService(mock, Version{Major: 24, Minor: 10})
	snap, err := svc.Get(context.Background(), "pool/dataset@snap1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snap == nil {
		t.Fatal("expected non-nil snapshot")
	}
	if !snap.HasHold {
		t.Error("expected HasHold=true")
	}
}

func TestSnapshotService_Get_NotFound(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return json.RawMessage(`[]`), nil
		},
	}

	svc := NewSnapshotService(mock, Version{Major: 24, Minor: 10})
	snap, err := svc.Get(context.Background(), "pool/dataset@nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snap != nil {
		t.Error("expected nil snapshot for not found")
	}
}

func TestSnapshotService_Get_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("timeout")
		},
	}

	svc := NewSnapshotService(mock, Version{Major: 24, Minor: 10})
	_, err := svc.Get(context.Background(), "pool/dataset@snap1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSnapshotService_Get_ParseError(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return json.RawMessage(`not json`), nil
		},
	}

	svc := NewSnapshotService(mock, Version{Major: 24, Minor: 10})
	_, err := svc.Get(context.Background(), "pool/dataset@snap1")
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestSnapshotService_List(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "zfs.snapshot.query" {
				t.Errorf("expected method zfs.snapshot.query, got %s", method)
			}
			if params != nil {
				t.Error("expected nil params for List")
			}
			return json.RawMessage(`[
				{"id": "pool/ds@snap1", "name": "pool/ds@snap1", "snapshot_name": "snap1", "dataset": "pool/ds", "properties": {"createtxg": {"value": "100"}, "used": {"parsed": 512}, "referenced": {"parsed": 1024}, "userrefs": {"parsed": "0"}}},
				{"id": "pool/ds@snap2", "name": "pool/ds@snap2", "snapshot_name": "snap2", "dataset": "pool/ds", "properties": {"createtxg": {"value": "200"}, "used": {"parsed": 256}, "referenced": {"parsed": 4096}, "userrefs": {"parsed": "1"}}}
			]`), nil
		},
	}

	svc := NewSnapshotService(mock, Version{Major: 24, Minor: 10})
	snaps, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(snaps) != 2 {
		t.Fatalf("expected 2 snapshots, got %d", len(snaps))
	}
	if snaps[0].ID != "pool/ds@snap1" {
		t.Errorf("expected first snapshot ID pool/ds@snap1, got %s", snaps[0].ID)
	}
	if snaps[1].SnapshotName != "snap2" {
		t.Errorf("expected second snapshot name snap2, got %s", snaps[1].SnapshotName)
	}
	if snaps[1].HasHold != true {
		t.Error("expected second snapshot HasHold=true")
	}
}

func TestSnapshotService_List_Empty(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return json.RawMessage(`[]`), nil
		},
	}

	svc := NewSnapshotService(mock, Version{Major: 24, Minor: 10})
	snaps, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snaps == nil {
		t.Fatal("expected non-nil empty slice")
	}
	if len(snaps) != 0 {
		t.Errorf("expected 0 snapshots, got %d", len(snaps))
	}
}

func TestSnapshotService_List_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("network error")
		},
	}

	svc := NewSnapshotService(mock, Version{Major: 24, Minor: 10})
	_, err := svc.List(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSnapshotService_List_ParseError(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return json.RawMessage(`not json`), nil
		},
	}

	svc := NewSnapshotService(mock, Version{Major: 24, Minor: 10})
	_, err := svc.List(context.Background())
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestSnapshotService_Delete(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "zfs.snapshot.delete" {
				t.Errorf("expected method zfs.snapshot.delete, got %s", method)
			}
			id, ok := params.(string)
			if !ok || id != "pool/dataset@snap1" {
				t.Errorf("expected id pool/dataset@snap1, got %v", params)
			}
			return nil, nil
		},
	}

	svc := NewSnapshotService(mock, Version{Major: 24, Minor: 10})
	err := svc.Delete(context.Background(), "pool/dataset@snap1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSnapshotService_Delete_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("permission denied")
		},
	}

	svc := NewSnapshotService(mock, Version{Major: 24, Minor: 10})
	err := svc.Delete(context.Background(), "pool/dataset@snap1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSnapshotService_Hold(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "zfs.snapshot.hold" {
				t.Errorf("expected method zfs.snapshot.hold, got %s", method)
			}
			id, ok := params.(string)
			if !ok || id != "pool/dataset@snap1" {
				t.Errorf("expected id pool/dataset@snap1, got %v", params)
			}
			return nil, nil
		},
	}

	svc := NewSnapshotService(mock, Version{Major: 24, Minor: 10})
	err := svc.Hold(context.Background(), "pool/dataset@snap1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSnapshotService_Hold_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("hold failed")
		},
	}

	svc := NewSnapshotService(mock, Version{Major: 24, Minor: 10})
	err := svc.Hold(context.Background(), "pool/dataset@snap1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSnapshotService_Release(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "zfs.snapshot.release" {
				t.Errorf("expected method zfs.snapshot.release, got %s", method)
			}
			id, ok := params.(string)
			if !ok || id != "pool/dataset@snap1" {
				t.Errorf("expected id pool/dataset@snap1, got %v", params)
			}
			return nil, nil
		},
	}

	svc := NewSnapshotService(mock, Version{Major: 24, Minor: 10})
	err := svc.Release(context.Background(), "pool/dataset@snap1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSnapshotService_Release_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("release failed")
		},
	}

	svc := NewSnapshotService(mock, Version{Major: 24, Minor: 10})
	err := svc.Release(context.Background(), "pool/dataset@snap1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSnapshotService_Clone(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "zfs.snapshot.clone" {
				t.Errorf("expected method zfs.snapshot.clone, got %s", method)
			}
			p := params.(map[string]any)
			if p["snapshot"] != "pool/dataset@snap1" {
				t.Errorf("expected snapshot pool/dataset@snap1, got %v", p["snapshot"])
			}
			if p["dataset_dst"] != "pool/clone1" {
				t.Errorf("expected dataset_dst pool/clone1, got %v", p["dataset_dst"])
			}
			return nil, nil
		},
	}

	svc := NewSnapshotService(mock, Version{Major: 24, Minor: 10})
	err := svc.Clone(context.Background(), "pool/dataset@snap1", "pool/clone1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSnapshotService_Clone_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("clone failed")
		},
	}

	svc := NewSnapshotService(mock, Version{Major: 24, Minor: 10})
	err := svc.Clone(context.Background(), "pool/dataset@snap1", "pool/clone1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSnapshotService_Clone_VersionPrefix(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "pool.snapshot.clone" {
				t.Errorf("expected method pool.snapshot.clone, got %s", method)
			}
			return nil, nil
		},
	}

	svc := NewSnapshotService(mock, Version{Major: 25, Minor: 10})
	err := svc.Clone(context.Background(), "pool/dataset@snap1", "pool/clone1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSnapshotFromResponse(t *testing.T) {
	t.Run("without hold", func(t *testing.T) {
		resp := SnapshotResponse{
			ID:           "pool/dataset@snap1",
			Name:         "pool/dataset@snap1",
			SnapshotName: "snap1",
			Dataset:      "pool/dataset",
			Properties: SnapshotProperties{
				CreateTXG:  PropertyValue{Value: "12345"},
				Used:       ParsedValue{Parsed: 1024},
				Referenced: ParsedValue{Parsed: 2048},
				UserRefs:   UserRefsProperty{Parsed: "0"},
			},
		}

		snap := snapshotFromResponse(resp)

		if snap.ID != "pool/dataset@snap1" {
			t.Errorf("expected ID pool/dataset@snap1, got %s", snap.ID)
		}
		if snap.Dataset != "pool/dataset" {
			t.Errorf("expected dataset pool/dataset, got %s", snap.Dataset)
		}
		if snap.SnapshotName != "snap1" {
			t.Errorf("expected snapshot name snap1, got %s", snap.SnapshotName)
		}
		if snap.CreateTXG != "12345" {
			t.Errorf("expected createtxg 12345, got %s", snap.CreateTXG)
		}
		if snap.Used != 1024 {
			t.Errorf("expected used 1024, got %d", snap.Used)
		}
		if snap.Referenced != 2048 {
			t.Errorf("expected referenced 2048, got %d", snap.Referenced)
		}
		if snap.HasHold {
			t.Error("expected HasHold=false")
		}
	})

	t.Run("with hold", func(t *testing.T) {
		resp := SnapshotResponse{
			ID:           "pool/dataset@snap1",
			Name:         "pool/dataset@snap1",
			SnapshotName: "snap1",
			Dataset:      "pool/dataset",
			Properties: SnapshotProperties{
				CreateTXG:  PropertyValue{Value: "12345"},
				Used:       ParsedValue{Parsed: 512},
				Referenced: ParsedValue{Parsed: 4096},
				UserRefs:   UserRefsProperty{Parsed: "2"},
			},
		}

		snap := snapshotFromResponse(resp)

		if !snap.HasHold {
			t.Error("expected HasHold=true")
		}
	})
}

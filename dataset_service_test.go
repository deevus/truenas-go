package truenas

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestDatasetService_CreateDataset(t *testing.T) {
	callCount := 0
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			callCount++
			if callCount == 1 {
				if method != "pool.dataset.create" {
					t.Errorf("expected method pool.dataset.create, got %s", method)
				}
				p := params.(map[string]any)
				if p["name"] != "pool1/ds1" {
					t.Errorf("expected name pool1/ds1, got %v", p["name"])
				}
				if p["type"] != "FILESYSTEM" {
					t.Errorf("expected type FILESYSTEM, got %v", p["type"])
				}
				if p["comments"] != "test dataset" {
					t.Errorf("expected comments 'test dataset', got %v", p["comments"])
				}
				if p["compression"] != "lz4" {
					t.Errorf("expected compression lz4, got %v", p["compression"])
				}
				if p["quota"] != int64(1073741824) {
					t.Errorf("expected quota 1073741824, got %v", p["quota"])
				}
				if p["refquota"] != int64(536870912) {
					t.Errorf("expected refquota 536870912, got %v", p["refquota"])
				}
				if p["atime"] != "on" {
					t.Errorf("expected atime on, got %v", p["atime"])
				}
				return json.RawMessage(`{"id": "pool1/ds1", "name": "pool1/ds1", "mountpoint": "/mnt/pool1/ds1"}`), nil
			}
			// pool.dataset.query for re-read
			return sampleDatasetQueryJSON(), nil
		},
	}

	svc := NewDatasetService(mock, Version{})
	ds, err := svc.CreateDataset(context.Background(), CreateDatasetOpts{
		Name:        "pool1/ds1",
		Comments:    "test dataset",
		Compression: "lz4",
		Quota:       1073741824,
		RefQuota:    536870912,
		Atime:       "on",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ds == nil {
		t.Fatal("expected non-nil dataset")
	}
	if ds.ID != "pool1/ds1" {
		t.Errorf("expected ID pool1/ds1, got %s", ds.ID)
	}
	if ds.Name != "pool1/ds1" {
		t.Errorf("expected Name pool1/ds1, got %s", ds.Name)
	}
	if ds.Pool != "pool1" {
		t.Errorf("expected Pool pool1, got %s", ds.Pool)
	}
	if ds.Mountpoint != "/mnt/pool1/ds1" {
		t.Errorf("expected Mountpoint /mnt/pool1/ds1, got %s", ds.Mountpoint)
	}
	if ds.Comments != "test dataset" {
		t.Errorf("expected Comments 'test dataset', got %s", ds.Comments)
	}
	if ds.Compression != "lz4" {
		t.Errorf("expected Compression lz4, got %s", ds.Compression)
	}
	if ds.Quota != 1073741824 {
		t.Errorf("expected Quota 1073741824, got %d", ds.Quota)
	}
	if ds.RefQuota != 536870912 {
		t.Errorf("expected RefQuota 536870912, got %d", ds.RefQuota)
	}
	if ds.Atime != "on" {
		t.Errorf("expected Atime on, got %s", ds.Atime)
	}
}

func TestDatasetService_CreateDataset_MinimalOpts(t *testing.T) {
	callCount := 0
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			callCount++
			if callCount == 1 {
				p := params.(map[string]any)
				if _, ok := p["comments"]; ok {
					t.Error("expected no comments key in minimal opts")
				}
				if _, ok := p["compression"]; ok {
					t.Error("expected no compression key in minimal opts")
				}
				if _, ok := p["quota"]; ok {
					t.Error("expected no quota key in minimal opts")
				}
				if _, ok := p["refquota"]; ok {
					t.Error("expected no refquota key in minimal opts")
				}
				if _, ok := p["atime"]; ok {
					t.Error("expected no atime key in minimal opts")
				}
				return json.RawMessage(`{"id": "pool1/ds1", "name": "pool1/ds1", "mountpoint": "/mnt/pool1/ds1"}`), nil
			}
			return sampleDatasetQueryJSON(), nil
		},
	}

	svc := NewDatasetService(mock, Version{})
	ds, err := svc.CreateDataset(context.Background(), CreateDatasetOpts{
		Name: "pool1/ds1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ds == nil {
		t.Fatal("expected non-nil dataset")
	}
}

func TestDatasetService_CreateDataset_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("connection refused")
		},
	}

	svc := NewDatasetService(mock, Version{})
	ds, err := svc.CreateDataset(context.Background(), CreateDatasetOpts{Name: "pool1/ds1"})
	if err == nil {
		t.Fatal("expected error")
	}
	if ds != nil {
		t.Error("expected nil dataset on error")
	}
	if err.Error() != "connection refused" {
		t.Errorf("expected 'connection refused', got %q", err.Error())
	}
}

func TestDatasetService_CreateDataset_ParseError(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return json.RawMessage(`not json`), nil
		},
	}

	svc := NewDatasetService(mock, Version{})
	_, err := svc.CreateDataset(context.Background(), CreateDatasetOpts{Name: "pool1/ds1"})
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestDatasetService_GetDataset(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "pool.dataset.query" {
				t.Errorf("expected method pool.dataset.query, got %s", method)
			}
			return sampleDatasetQueryJSON(), nil
		},
	}

	svc := NewDatasetService(mock, Version{})
	ds, err := svc.GetDataset(context.Background(), "pool1/ds1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ds == nil {
		t.Fatal("expected non-nil dataset")
	}
	if ds.ID != "pool1/ds1" {
		t.Errorf("expected ID pool1/ds1, got %s", ds.ID)
	}
	if ds.Comments != "test dataset" {
		t.Errorf("expected Comments 'test dataset', got %q", ds.Comments)
	}
}

func TestDatasetService_GetDataset_NotFound(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return json.RawMessage(`[]`), nil
		},
	}

	svc := NewDatasetService(mock, Version{})
	ds, err := svc.GetDataset(context.Background(), "pool1/nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ds != nil {
		t.Error("expected nil dataset for not found")
	}
}

func TestDatasetService_GetDataset_NotFoundError(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("dataset does not exist")
		},
	}

	svc := NewDatasetService(mock, Version{})
	ds, err := svc.GetDataset(context.Background(), "pool1/nonexistent")
	if err != nil {
		t.Fatalf("expected nil error for not-found, got %v", err)
	}
	if ds != nil {
		t.Error("expected nil dataset for not found")
	}
}

func TestDatasetService_GetDataset_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("timeout")
		},
	}

	svc := NewDatasetService(mock, Version{})
	_, err := svc.GetDataset(context.Background(), "pool1/ds1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDatasetService_GetDataset_ParseError(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return json.RawMessage(`not json`), nil
		},
	}

	svc := NewDatasetService(mock, Version{})
	_, err := svc.GetDataset(context.Background(), "pool1/ds1")
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestDatasetService_ListDatasets(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "pool.dataset.query" {
				t.Errorf("expected method pool.dataset.query, got %s", method)
			}
			if params != nil {
				t.Error("expected nil params for ListDatasets")
			}
			// Return a mix of FILESYSTEM and VOLUME types
			return json.RawMessage(`[
				{"id": "pool1/ds1", "name": "pool1/ds1", "pool": "pool1", "type": "FILESYSTEM", "mountpoint": "/mnt/pool1/ds1", "comments": {"value": ""}, "compression": {"value": "lz4"}, "quota": {"parsed": 0, "value": ""}, "refquota": {"parsed": 0, "value": ""}, "atime": {"value": "on"}, "volsize": {"parsed": 0, "value": ""}, "volblocksize": {"value": ""}, "sparse": {"value": ""}},
				{"id": "pool1/zvol1", "name": "pool1/zvol1", "pool": "pool1", "type": "VOLUME", "mountpoint": "", "comments": {"value": ""}, "compression": {"value": ""}, "quota": {"parsed": 0, "value": ""}, "refquota": {"parsed": 0, "value": ""}, "atime": {"value": ""}, "volsize": {"parsed": 10737418240, "value": "10G"}, "volblocksize": {"value": "16K"}, "sparse": {"value": "true"}},
				{"id": "pool1/ds2", "name": "pool1/ds2", "pool": "pool1", "type": "FILESYSTEM", "mountpoint": "/mnt/pool1/ds2", "comments": {"value": ""}, "compression": {"value": "off"}, "quota": {"parsed": 0, "value": ""}, "refquota": {"parsed": 0, "value": ""}, "atime": {"value": "off"}, "volsize": {"parsed": 0, "value": ""}, "volblocksize": {"value": ""}, "sparse": {"value": ""}}
			]`), nil
		},
	}

	svc := NewDatasetService(mock, Version{})
	datasets, err := svc.ListDatasets(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(datasets) != 2 {
		t.Fatalf("expected 2 datasets (FILESYSTEM only), got %d", len(datasets))
	}
	if datasets[0].ID != "pool1/ds1" {
		t.Errorf("expected first dataset ID pool1/ds1, got %s", datasets[0].ID)
	}
	if datasets[1].ID != "pool1/ds2" {
		t.Errorf("expected second dataset ID pool1/ds2, got %s", datasets[1].ID)
	}
}

func TestDatasetService_ListDatasets_Empty(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return json.RawMessage(`[]`), nil
		},
	}

	svc := NewDatasetService(mock, Version{})
	datasets, err := svc.ListDatasets(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(datasets) != 0 {
		t.Errorf("expected 0 datasets, got %d", len(datasets))
	}
	if datasets == nil {
		t.Error("expected non-nil empty slice")
	}
}

func TestDatasetService_ListDatasets_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("network error")
		},
	}

	svc := NewDatasetService(mock, Version{})
	_, err := svc.ListDatasets(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDatasetService_UpdateDataset(t *testing.T) {
	callCount := 0
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			callCount++
			if callCount == 1 {
				if method != "pool.dataset.update" {
					t.Errorf("expected method pool.dataset.update, got %s", method)
				}
				slice, ok := params.([]any)
				if !ok {
					t.Fatal("expected []any params for update")
				}
				if len(slice) != 2 {
					t.Fatalf("expected 2 elements, got %d", len(slice))
				}
				id, ok := slice[0].(string)
				if !ok || id != "pool1/ds1" {
					t.Errorf("expected id pool1/ds1, got %v", slice[0])
				}
				p := slice[1].(map[string]any)
				if p["comments"] != "updated" {
					t.Errorf("expected comments 'updated', got %v", p["comments"])
				}
				return json.RawMessage(`{"id": "pool1/ds1"}`), nil
			}
			// Re-query
			return sampleDatasetQueryJSON(), nil
		},
	}

	svc := NewDatasetService(mock, Version{})
	ds, err := svc.UpdateDataset(context.Background(), "pool1/ds1", UpdateDatasetOpts{
		Comments: StringPtr("updated"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ds == nil {
		t.Fatal("expected non-nil dataset")
	}
	if ds.ID != "pool1/ds1" {
		t.Errorf("expected ID pool1/ds1, got %s", ds.ID)
	}
}

func TestDatasetService_UpdateDataset_QuotaToZero(t *testing.T) {
	callCount := 0
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			callCount++
			if callCount == 1 {
				slice := params.([]any)
				p := slice[1].(map[string]any)
				if p["quota"] != int64(0) {
					t.Errorf("expected quota 0, got %v", p["quota"])
				}
				return json.RawMessage(`{"id": "pool1/ds1"}`), nil
			}
			return sampleDatasetQueryJSON(), nil
		},
	}

	svc := NewDatasetService(mock, Version{})
	ds, err := svc.UpdateDataset(context.Background(), "pool1/ds1", UpdateDatasetOpts{
		Quota: Int64Ptr(0),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ds == nil {
		t.Fatal("expected non-nil dataset")
	}
}

func TestDatasetService_UpdateDataset_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("not found")
		},
	}

	svc := NewDatasetService(mock, Version{})
	_, err := svc.UpdateDataset(context.Background(), "pool1/ds1", UpdateDatasetOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDatasetService_DeleteDataset(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "pool.dataset.delete" {
				t.Errorf("expected method pool.dataset.delete, got %s", method)
			}
			id, ok := params.(string)
			if !ok || id != "pool1/ds1" {
				t.Errorf("expected id pool1/ds1, got %v", params)
			}
			return nil, nil
		},
	}

	svc := NewDatasetService(mock, Version{})
	err := svc.DeleteDataset(context.Background(), "pool1/ds1", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDatasetService_DeleteDataset_Recursive(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "pool.dataset.delete" {
				t.Errorf("expected method pool.dataset.delete, got %s", method)
			}
			slice, ok := params.([]any)
			if !ok {
				t.Fatal("expected []any params for recursive delete")
			}
			if len(slice) != 2 {
				t.Fatalf("expected 2 elements, got %d", len(slice))
			}
			id, ok := slice[0].(string)
			if !ok || id != "pool1/ds1" {
				t.Errorf("expected id pool1/ds1, got %v", slice[0])
			}
			opts, ok := slice[1].(map[string]any)
			if !ok {
				t.Fatal("expected map[string]any as second element")
			}
			if opts["recursive"] != true {
				t.Error("expected recursive=true")
			}
			return nil, nil
		},
	}

	svc := NewDatasetService(mock, Version{})
	err := svc.DeleteDataset(context.Background(), "pool1/ds1", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDatasetService_DeleteDataset_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("permission denied")
		},
	}

	svc := NewDatasetService(mock, Version{})
	err := svc.DeleteDataset(context.Background(), "pool1/ds1", false)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDatasetService_GetDataset_UsageFields(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return json.RawMessage(`[{
				"id": "pool1/ds1",
				"name": "pool1/ds1",
				"pool": "pool1",
				"type": "FILESYSTEM",
				"mountpoint": "/mnt/pool1/ds1",
				"comments": {"value": ""},
				"compression": {"value": "lz4"},
				"quota": {"parsed": 0, "value": "0"},
				"refquota": {"parsed": 0, "value": "0"},
				"atime": {"value": "on"},
				"volsize": {"parsed": 0, "value": ""},
				"volblocksize": {"value": ""},
				"sparse": {"value": ""},
				"used": {"parsed": 1073741824, "value": "1G"},
				"available": {"parsed": 5368709120, "value": "5G"}
			}]`), nil
		},
	}

	svc := NewDatasetService(mock, Version{})
	ds, err := svc.GetDataset(context.Background(), "pool1/ds1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ds == nil {
		t.Fatal("expected non-nil dataset")
	}
	if ds.Used != 1073741824 {
		t.Errorf("expected Used 1073741824, got %d", ds.Used)
	}
	if ds.Available != 5368709120 {
		t.Errorf("expected Available 5368709120, got %d", ds.Available)
	}
}

func TestDatasetService_UpdateDataset_CompressionAndAtime(t *testing.T) {
	callCount := 0
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			callCount++
			if callCount == 1 {
				if method != "pool.dataset.update" {
					t.Errorf("expected method pool.dataset.update, got %s", method)
				}
				slice := params.([]any)
				p := slice[1].(map[string]any)
				if p["compression"] != "zstd" {
					t.Errorf("expected compression zstd, got %v", p["compression"])
				}
				if p["atime"] != "off" {
					t.Errorf("expected atime off, got %v", p["atime"])
				}
				if _, ok := p["comments"]; ok {
					t.Error("expected no comments key when nil")
				}
				return json.RawMessage(`{"id": "pool1/ds1"}`), nil
			}
			return sampleDatasetQueryJSON(), nil
		},
	}

	svc := NewDatasetService(mock, Version{})
	ds, err := svc.UpdateDataset(context.Background(), "pool1/ds1", UpdateDatasetOpts{
		Compression: "zstd",
		Atime:       "off",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ds == nil {
		t.Fatal("expected non-nil dataset")
	}
	if ds.ID != "pool1/ds1" {
		t.Errorf("expected ID pool1/ds1, got %s", ds.ID)
	}
}

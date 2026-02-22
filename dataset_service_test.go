package truenas

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

// sampleDatasetQueryJSON returns a JSON response for a FILESYSTEM dataset with all property fields.
func sampleDatasetQueryJSON() json.RawMessage {
	return json.RawMessage(`[{
		"id": "pool1/ds1",
		"name": "pool1/ds1",
		"pool": "pool1",
		"type": "FILESYSTEM",
		"mountpoint": "/mnt/pool1/ds1",
		"comments": {"value": "test dataset"},
		"compression": {"value": "lz4"},
		"quota": {"parsed": 1073741824, "value": "1G"},
		"refquota": {"parsed": 536870912, "value": "512M"},
		"atime": {"value": "on"},
		"volsize": {"parsed": 0, "value": ""},
		"volblocksize": {"value": ""},
		"sparse": {"value": ""}
	}]`)
}

// sampleZvolQueryJSON returns a JSON response for a VOLUME (zvol).
func sampleZvolQueryJSON() json.RawMessage {
	return json.RawMessage(`[{
		"id": "pool1/zvol1",
		"name": "pool1/zvol1",
		"pool": "pool1",
		"type": "VOLUME",
		"mountpoint": "",
		"comments": {"value": "test zvol"},
		"compression": {"value": "lz4"},
		"quota": {"parsed": 0, "value": ""},
		"refquota": {"parsed": 0, "value": ""},
		"atime": {"value": ""},
		"volsize": {"parsed": 10737418240, "value": "10G"},
		"volblocksize": {"value": "16K"},
		"sparse": {"value": "true"}
	}]`)
}

// samplePoolQueryJSON returns a JSON response for two pools.
func samplePoolQueryJSON() json.RawMessage {
	return json.RawMessage(`[
		{"id": 1, "name": "pool1", "path": "/mnt/pool1"},
		{"id": 2, "name": "pool2", "path": "/mnt/pool2"}
	]`)
}

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

func TestDatasetService_CreateZvol(t *testing.T) {
	callCount := 0
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			callCount++
			if callCount == 1 {
				if method != "pool.dataset.create" {
					t.Errorf("expected method pool.dataset.create, got %s", method)
				}
				p := params.(map[string]any)
				if p["name"] != "pool1/zvol1" {
					t.Errorf("expected name pool1/zvol1, got %v", p["name"])
				}
				if p["type"] != "VOLUME" {
					t.Errorf("expected type VOLUME, got %v", p["type"])
				}
				if p["volsize"] != int64(10737418240) {
					t.Errorf("expected volsize 10737418240, got %v", p["volsize"])
				}
				if p["volblocksize"] != "16K" {
					t.Errorf("expected volblocksize 16K, got %v", p["volblocksize"])
				}
				if p["sparse"] != true {
					t.Errorf("expected sparse true, got %v", p["sparse"])
				}
				if p["comments"] != "test zvol" {
					t.Errorf("expected comments 'test zvol', got %v", p["comments"])
				}
				return json.RawMessage(`{"id": "pool1/zvol1", "name": "pool1/zvol1", "mountpoint": ""}`), nil
			}
			return sampleZvolQueryJSON(), nil
		},
	}

	svc := NewDatasetService(mock, Version{})
	zvol, err := svc.CreateZvol(context.Background(), CreateZvolOpts{
		Name:         "pool1/zvol1",
		Volsize:      10737418240,
		Volblocksize: "16K",
		Sparse:       true,
		Comments:     "test zvol",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if zvol == nil {
		t.Fatal("expected non-nil zvol")
	}
	if zvol.ID != "pool1/zvol1" {
		t.Errorf("expected ID pool1/zvol1, got %s", zvol.ID)
	}
	if zvol.Volsize != 10737418240 {
		t.Errorf("expected Volsize 10737418240, got %d", zvol.Volsize)
	}
	if zvol.Volblocksize != "16K" {
		t.Errorf("expected Volblocksize 16K, got %s", zvol.Volblocksize)
	}
	if !zvol.Sparse {
		t.Error("expected Sparse=true")
	}
	if zvol.Comments != "test zvol" {
		t.Errorf("expected Comments 'test zvol', got %s", zvol.Comments)
	}
	if zvol.Compression != "lz4" {
		t.Errorf("expected Compression lz4, got %s", zvol.Compression)
	}
}

func TestDatasetService_CreateZvol_MinimalOpts(t *testing.T) {
	callCount := 0
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			callCount++
			if callCount == 1 {
				p := params.(map[string]any)
				if _, ok := p["volblocksize"]; ok {
					t.Error("expected no volblocksize key in minimal opts")
				}
				if _, ok := p["sparse"]; ok {
					t.Error("expected no sparse key in minimal opts")
				}
				if _, ok := p["comments"]; ok {
					t.Error("expected no comments key in minimal opts")
				}
				return json.RawMessage(`{"id": "pool1/zvol1", "name": "pool1/zvol1", "mountpoint": ""}`), nil
			}
			return sampleZvolQueryJSON(), nil
		},
	}

	svc := NewDatasetService(mock, Version{})
	zvol, err := svc.CreateZvol(context.Background(), CreateZvolOpts{
		Name:    "pool1/zvol1",
		Volsize: 10737418240,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if zvol == nil {
		t.Fatal("expected non-nil zvol")
	}
}

func TestDatasetService_CreateZvol_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("connection refused")
		},
	}

	svc := NewDatasetService(mock, Version{})
	zvol, err := svc.CreateZvol(context.Background(), CreateZvolOpts{Name: "pool1/zvol1", Volsize: 1024})
	if err == nil {
		t.Fatal("expected error")
	}
	if zvol != nil {
		t.Error("expected nil zvol on error")
	}
}

func TestDatasetService_GetZvol(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "pool.dataset.query" {
				t.Errorf("expected method pool.dataset.query, got %s", method)
			}
			return sampleZvolQueryJSON(), nil
		},
	}

	svc := NewDatasetService(mock, Version{})
	zvol, err := svc.GetZvol(context.Background(), "pool1/zvol1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if zvol == nil {
		t.Fatal("expected non-nil zvol")
	}
	if zvol.ID != "pool1/zvol1" {
		t.Errorf("expected ID pool1/zvol1, got %s", zvol.ID)
	}
	if zvol.Volsize != 10737418240 {
		t.Errorf("expected Volsize 10737418240, got %d", zvol.Volsize)
	}
	if zvol.Volblocksize != "16K" {
		t.Errorf("expected Volblocksize 16K, got %s", zvol.Volblocksize)
	}
	if !zvol.Sparse {
		t.Error("expected Sparse=true")
	}
	if zvol.Compression != "lz4" {
		t.Errorf("expected Compression lz4, got %s", zvol.Compression)
	}
}

func TestDatasetService_GetZvol_NotFound(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return json.RawMessage(`[]`), nil
		},
	}

	svc := NewDatasetService(mock, Version{})
	zvol, err := svc.GetZvol(context.Background(), "pool1/nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if zvol != nil {
		t.Error("expected nil zvol for not found")
	}
}

func TestDatasetService_GetZvol_NotFoundError(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("dataset does not exist")
		},
	}

	svc := NewDatasetService(mock, Version{})
	zvol, err := svc.GetZvol(context.Background(), "pool1/nonexistent")
	if err != nil {
		t.Fatalf("expected nil error for not-found, got %v", err)
	}
	if zvol != nil {
		t.Error("expected nil zvol for not found")
	}
}

func TestDatasetService_GetZvol_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("timeout")
		},
	}

	svc := NewDatasetService(mock, Version{})
	_, err := svc.GetZvol(context.Background(), "pool1/zvol1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDatasetService_UpdateZvol(t *testing.T) {
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
				if !ok || id != "pool1/zvol1" {
					t.Errorf("expected id pool1/zvol1, got %v", slice[0])
				}
				p := slice[1].(map[string]any)
				if p["volsize"] != int64(21474836480) {
					t.Errorf("expected volsize 21474836480, got %v", p["volsize"])
				}
				if p["comments"] != "resized" {
					t.Errorf("expected comments 'resized', got %v", p["comments"])
				}
				return json.RawMessage(`{"id": "pool1/zvol1"}`), nil
			}
			return sampleZvolQueryJSON(), nil
		},
	}

	svc := NewDatasetService(mock, Version{})
	zvol, err := svc.UpdateZvol(context.Background(), "pool1/zvol1", UpdateZvolOpts{
		Volsize:  Int64Ptr(21474836480),
		Comments: StringPtr("resized"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if zvol == nil {
		t.Fatal("expected non-nil zvol")
	}
	if zvol.ID != "pool1/zvol1" {
		t.Errorf("expected ID pool1/zvol1, got %s", zvol.ID)
	}
}

func TestDatasetService_UpdateZvol_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("not found")
		},
	}

	svc := NewDatasetService(mock, Version{})
	_, err := svc.UpdateZvol(context.Background(), "pool1/zvol1", UpdateZvolOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDatasetService_DeleteZvol(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "pool.dataset.delete" {
				t.Errorf("expected method pool.dataset.delete, got %s", method)
			}
			id, ok := params.(string)
			if !ok || id != "pool1/zvol1" {
				t.Errorf("expected id pool1/zvol1, got %v", params)
			}
			return nil, nil
		},
	}

	svc := NewDatasetService(mock, Version{})
	err := svc.DeleteZvol(context.Background(), "pool1/zvol1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDatasetService_DeleteZvol_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("permission denied")
		},
	}

	svc := NewDatasetService(mock, Version{})
	err := svc.DeleteZvol(context.Background(), "pool1/zvol1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDatasetService_ListPools(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "pool.query" {
				t.Errorf("expected method pool.query, got %s", method)
			}
			if params != nil {
				t.Error("expected nil params for ListPools")
			}
			return samplePoolQueryJSON(), nil
		},
	}

	svc := NewDatasetService(mock, Version{})
	pools, err := svc.ListPools(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pools) != 2 {
		t.Fatalf("expected 2 pools, got %d", len(pools))
	}
	if pools[0].ID != 1 {
		t.Errorf("expected first pool ID 1, got %d", pools[0].ID)
	}
	if pools[0].Name != "pool1" {
		t.Errorf("expected first pool name pool1, got %s", pools[0].Name)
	}
	if pools[0].Path != "/mnt/pool1" {
		t.Errorf("expected first pool path /mnt/pool1, got %s", pools[0].Path)
	}
	if pools[1].ID != 2 {
		t.Errorf("expected second pool ID 2, got %d", pools[1].ID)
	}
	if pools[1].Name != "pool2" {
		t.Errorf("expected second pool name pool2, got %s", pools[1].Name)
	}
}

func TestDatasetService_ListPools_Empty(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return json.RawMessage(`[]`), nil
		},
	}

	svc := NewDatasetService(mock, Version{})
	pools, err := svc.ListPools(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pools) != 0 {
		t.Errorf("expected 0 pools, got %d", len(pools))
	}
}

func TestDatasetService_ListPools_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("network error")
		},
	}

	svc := NewDatasetService(mock, Version{})
	_, err := svc.ListPools(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDatasetFromResponse(t *testing.T) {
	resp := DatasetResponse{
		ID:          "pool1/ds1",
		Name:        "pool1/ds1",
		Pool:        "pool1",
		Type:        "FILESYSTEM",
		Mountpoint:  "/mnt/pool1/ds1",
		Comments:    PropertyValue{Value: "my dataset"},
		Compression: PropertyValue{Value: "zstd"},
		Quota:       SizePropertyField{Parsed: 2147483648, Value: "2G"},
		RefQuota:    SizePropertyField{Parsed: 1073741824, Value: "1G"},
		Atime:       PropertyValue{Value: "off"},
	}

	ds := datasetFromResponse(resp)

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
	if ds.Comments != "my dataset" {
		t.Errorf("expected Comments 'my dataset', got %s", ds.Comments)
	}
	if ds.Compression != "zstd" {
		t.Errorf("expected Compression zstd, got %s", ds.Compression)
	}
	if ds.Quota != 2147483648 {
		t.Errorf("expected Quota 2147483648, got %d", ds.Quota)
	}
	if ds.RefQuota != 1073741824 {
		t.Errorf("expected RefQuota 1073741824, got %d", ds.RefQuota)
	}
	if ds.Atime != "off" {
		t.Errorf("expected Atime off, got %s", ds.Atime)
	}
}

func TestZvolFromResponse(t *testing.T) {
	resp := DatasetResponse{
		ID:           "pool1/zvol1",
		Name:         "pool1/zvol1",
		Pool:         "pool1",
		Type:         "VOLUME",
		Comments:     PropertyValue{Value: "my zvol"},
		Compression:  PropertyValue{Value: "lz4"},
		Volsize:      SizePropertyField{Parsed: 10737418240, Value: "10G"},
		Volblocksize: PropertyValue{Value: "16K"},
		Sparse:       PropertyValue{Value: "true"},
	}

	zvol := zvolFromResponse(resp)

	if zvol.ID != "pool1/zvol1" {
		t.Errorf("expected ID pool1/zvol1, got %s", zvol.ID)
	}
	if zvol.Name != "pool1/zvol1" {
		t.Errorf("expected Name pool1/zvol1, got %s", zvol.Name)
	}
	if zvol.Pool != "pool1" {
		t.Errorf("expected Pool pool1, got %s", zvol.Pool)
	}
	if zvol.Comments != "my zvol" {
		t.Errorf("expected Comments 'my zvol', got %s", zvol.Comments)
	}
	if zvol.Compression != "lz4" {
		t.Errorf("expected Compression lz4, got %s", zvol.Compression)
	}
	if zvol.Volsize != 10737418240 {
		t.Errorf("expected Volsize 10737418240, got %d", zvol.Volsize)
	}
	if zvol.Volblocksize != "16K" {
		t.Errorf("expected Volblocksize 16K, got %s", zvol.Volblocksize)
	}
	if !zvol.Sparse {
		t.Error("expected Sparse=true")
	}
}

func TestZvolFromResponse_SparseTrue(t *testing.T) {
	resp := DatasetResponse{
		Sparse: PropertyValue{Value: "true"},
	}
	zvol := zvolFromResponse(resp)
	if !zvol.Sparse {
		t.Error("expected Sparse=true for value 'true'")
	}
}

func TestZvolFromResponse_SparseFalse(t *testing.T) {
	resp := DatasetResponse{
		Sparse: PropertyValue{Value: "false"},
	}
	zvol := zvolFromResponse(resp)
	if zvol.Sparse {
		t.Error("expected Sparse=false for value 'false'")
	}
}

func TestPoolFromResponse(t *testing.T) {
	resp := PoolResponse{
		ID:   1,
		Name: "tank",
		Path: "/mnt/tank",
	}

	pool := poolFromResponse(resp)

	if pool.ID != 1 {
		t.Errorf("expected ID 1, got %d", pool.ID)
	}
	if pool.Name != "tank" {
		t.Errorf("expected Name tank, got %s", pool.Name)
	}
	if pool.Path != "/mnt/tank" {
		t.Errorf("expected Path /mnt/tank, got %s", pool.Path)
	}
}

func TestDatasetCreateParams(t *testing.T) {
	opts := CreateDatasetOpts{
		Name:        "pool1/ds1",
		Comments:    "test",
		Compression: "lz4",
		Quota:       1073741824,
		RefQuota:    536870912,
		Atime:       "on",
	}

	params := datasetCreateParams(opts)

	if params["name"] != "pool1/ds1" {
		t.Errorf("expected name pool1/ds1, got %v", params["name"])
	}
	if params["type"] != "FILESYSTEM" {
		t.Errorf("expected type FILESYSTEM, got %v", params["type"])
	}
	if params["comments"] != "test" {
		t.Errorf("expected comments test, got %v", params["comments"])
	}
	if params["compression"] != "lz4" {
		t.Errorf("expected compression lz4, got %v", params["compression"])
	}
	if params["quota"] != int64(1073741824) {
		t.Errorf("expected quota 1073741824, got %v", params["quota"])
	}
	if params["refquota"] != int64(536870912) {
		t.Errorf("expected refquota 536870912, got %v", params["refquota"])
	}
	if params["atime"] != "on" {
		t.Errorf("expected atime on, got %v", params["atime"])
	}
}

func TestDatasetCreateParams_Minimal(t *testing.T) {
	opts := CreateDatasetOpts{
		Name: "pool1/ds1",
	}

	params := datasetCreateParams(opts)

	if params["name"] != "pool1/ds1" {
		t.Errorf("expected name pool1/ds1, got %v", params["name"])
	}
	if params["type"] != "FILESYSTEM" {
		t.Errorf("expected type FILESYSTEM, got %v", params["type"])
	}
	if _, ok := params["comments"]; ok {
		t.Error("expected no comments key")
	}
	if _, ok := params["compression"]; ok {
		t.Error("expected no compression key")
	}
	if _, ok := params["quota"]; ok {
		t.Error("expected no quota key")
	}
	if _, ok := params["refquota"]; ok {
		t.Error("expected no refquota key")
	}
	if _, ok := params["atime"]; ok {
		t.Error("expected no atime key")
	}
}

func TestDatasetUpdateParams(t *testing.T) {
	opts := UpdateDatasetOpts{
		Compression: "zstd",
		Quota:       Int64Ptr(2147483648),
		RefQuota:    Int64Ptr(1073741824),
		Atime:       "off",
		Comments:    StringPtr("updated"),
	}

	params := datasetUpdateParams(opts)

	if params["compression"] != "zstd" {
		t.Errorf("expected compression zstd, got %v", params["compression"])
	}
	if params["quota"] != int64(2147483648) {
		t.Errorf("expected quota 2147483648, got %v", params["quota"])
	}
	if params["refquota"] != int64(1073741824) {
		t.Errorf("expected refquota 1073741824, got %v", params["refquota"])
	}
	if params["atime"] != "off" {
		t.Errorf("expected atime off, got %v", params["atime"])
	}
	if params["comments"] != "updated" {
		t.Errorf("expected comments updated, got %v", params["comments"])
	}
}

func TestDatasetUpdateParams_Empty(t *testing.T) {
	opts := UpdateDatasetOpts{}

	params := datasetUpdateParams(opts)

	if len(params) != 0 {
		t.Errorf("expected empty params, got %v", params)
	}
}

func TestDatasetUpdateParams_CompressionAndAtime(t *testing.T) {
	opts := UpdateDatasetOpts{
		Compression: "gzip",
		Atime:       "on",
	}

	params := datasetUpdateParams(opts)

	if params["compression"] != "gzip" {
		t.Errorf("expected compression gzip, got %v", params["compression"])
	}
	if params["atime"] != "on" {
		t.Errorf("expected atime on, got %v", params["atime"])
	}
	if _, ok := params["comments"]; ok {
		t.Error("expected no comments key when nil")
	}
	if _, ok := params["quota"]; ok {
		t.Error("expected no quota key when nil")
	}
	if _, ok := params["refquota"]; ok {
		t.Error("expected no refquota key when nil")
	}
	if len(params) != 2 {
		t.Errorf("expected 2 params, got %d", len(params))
	}
}

func TestZvolCreateParams(t *testing.T) {
	opts := CreateZvolOpts{
		Name:         "pool1/zvol1",
		Volsize:      10737418240,
		Volblocksize: "16K",
		Sparse:       true,
		ForceSize:    true,
		Compression:  "lz4",
		Comments:     "my zvol",
	}

	params := zvolCreateParams(opts)

	if params["name"] != "pool1/zvol1" {
		t.Errorf("expected name pool1/zvol1, got %v", params["name"])
	}
	if params["type"] != "VOLUME" {
		t.Errorf("expected type VOLUME, got %v", params["type"])
	}
	if params["volsize"] != int64(10737418240) {
		t.Errorf("expected volsize 10737418240, got %v", params["volsize"])
	}
	if params["volblocksize"] != "16K" {
		t.Errorf("expected volblocksize 16K, got %v", params["volblocksize"])
	}
	if params["sparse"] != true {
		t.Errorf("expected sparse true, got %v", params["sparse"])
	}
	if params["force_size"] != true {
		t.Errorf("expected force_size true, got %v", params["force_size"])
	}
	if params["compression"] != "lz4" {
		t.Errorf("expected compression lz4, got %v", params["compression"])
	}
	if params["comments"] != "my zvol" {
		t.Errorf("expected comments 'my zvol', got %v", params["comments"])
	}
}

func TestZvolCreateParams_Minimal(t *testing.T) {
	opts := CreateZvolOpts{
		Name:    "pool1/zvol1",
		Volsize: 10737418240,
	}

	params := zvolCreateParams(opts)

	if params["name"] != "pool1/zvol1" {
		t.Errorf("expected name pool1/zvol1, got %v", params["name"])
	}
	if params["type"] != "VOLUME" {
		t.Errorf("expected type VOLUME, got %v", params["type"])
	}
	if params["volsize"] != int64(10737418240) {
		t.Errorf("expected volsize 10737418240, got %v", params["volsize"])
	}
	if _, ok := params["volblocksize"]; ok {
		t.Error("expected no volblocksize key")
	}
	if _, ok := params["sparse"]; ok {
		t.Error("expected no sparse key")
	}
	if _, ok := params["force_size"]; ok {
		t.Error("expected no force_size key")
	}
	if _, ok := params["compression"]; ok {
		t.Error("expected no compression key")
	}
	if _, ok := params["comments"]; ok {
		t.Error("expected no comments key")
	}
}

func TestZvolUpdateParams(t *testing.T) {
	opts := UpdateZvolOpts{
		Volsize:     Int64Ptr(21474836480),
		ForceSize:   true,
		Compression: "zstd",
		Comments:    StringPtr("resized"),
	}

	params := zvolUpdateParams(opts)

	if params["volsize"] != int64(21474836480) {
		t.Errorf("expected volsize 21474836480, got %v", params["volsize"])
	}
	if params["force_size"] != true {
		t.Errorf("expected force_size true, got %v", params["force_size"])
	}
	if params["compression"] != "zstd" {
		t.Errorf("expected compression zstd, got %v", params["compression"])
	}
	if params["comments"] != "resized" {
		t.Errorf("expected comments resized, got %v", params["comments"])
	}
}

func TestZvolUpdateParams_Empty(t *testing.T) {
	opts := UpdateZvolOpts{}

	params := zvolUpdateParams(opts)

	if len(params) != 0 {
		t.Errorf("expected empty params, got %v", params)
	}
}

func TestZvolUpdateParams_ForceSizeOnly(t *testing.T) {
	opts := UpdateZvolOpts{
		ForceSize: true,
	}

	params := zvolUpdateParams(opts)

	if params["force_size"] != true {
		t.Errorf("expected force_size true, got %v", params["force_size"])
	}
	if len(params) != 1 {
		t.Errorf("expected 1 param, got %d", len(params))
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

func TestDatasetService_CreateZvol_WithForceSize(t *testing.T) {
	callCount := 0
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			callCount++
			if callCount == 1 {
				if method != "pool.dataset.create" {
					t.Errorf("expected method pool.dataset.create, got %s", method)
				}
				p := params.(map[string]any)
				if p["force_size"] != true {
					t.Errorf("expected force_size true, got %v", p["force_size"])
				}
				if p["compression"] != "gzip" {
					t.Errorf("expected compression gzip, got %v", p["compression"])
				}
				return json.RawMessage(`{"id": "pool1/zvol1", "name": "pool1/zvol1", "mountpoint": ""}`), nil
			}
			return sampleZvolQueryJSON(), nil
		},
	}

	svc := NewDatasetService(mock, Version{})
	zvol, err := svc.CreateZvol(context.Background(), CreateZvolOpts{
		Name:        "pool1/zvol1",
		Volsize:     10737418240,
		ForceSize:   true,
		Compression: "gzip",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if zvol == nil {
		t.Fatal("expected non-nil zvol")
	}
	if zvol.Compression != "lz4" {
		t.Errorf("expected Compression lz4 from re-read, got %s", zvol.Compression)
	}
}

func TestDatasetService_UpdateZvol_WithCompression(t *testing.T) {
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
				if p["force_size"] != true {
					t.Errorf("expected force_size true, got %v", p["force_size"])
				}
				return json.RawMessage(`{"id": "pool1/zvol1"}`), nil
			}
			return sampleZvolQueryJSON(), nil
		},
	}

	svc := NewDatasetService(mock, Version{})
	zvol, err := svc.UpdateZvol(context.Background(), "pool1/zvol1", UpdateZvolOpts{
		Compression: "zstd",
		ForceSize:   true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if zvol == nil {
		t.Fatal("expected non-nil zvol")
	}
	if zvol.ID != "pool1/zvol1" {
		t.Errorf("expected ID pool1/zvol1, got %s", zvol.ID)
	}
}

func TestInt64Ptr(t *testing.T) {
	p := Int64Ptr(42)
	if *p != 42 {
		t.Errorf("expected 42, got %d", *p)
	}

	z := Int64Ptr(0)
	if *z != 0 {
		t.Errorf("expected 0, got %d", *z)
	}
}

func TestStringPtr(t *testing.T) {
	p := StringPtr("hello")
	if *p != "hello" {
		t.Errorf("expected hello, got %s", *p)
	}

	e := StringPtr("")
	if *e != "" {
		t.Errorf("expected empty string, got %s", *e)
	}
}

package truenas

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

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

package truenas

import (
	"context"
	"encoding/json"
	"fmt"
)

// Dataset is the user-facing representation of a TrueNAS filesystem dataset.
type Dataset struct {
	ID          string
	Name        string
	Pool        string
	Mountpoint  string
	Comments    string
	Compression string
	Quota       int64
	RefQuota    int64
	Atime       string
}

// CreateDatasetOpts contains options for creating a filesystem dataset.
type CreateDatasetOpts struct {
	Name        string
	Comments    string
	Compression string
	Quota       int64
	RefQuota    int64
	Atime       string
}

// UpdateDatasetOpts contains options for updating a filesystem dataset.
// Pointer fields distinguish "don't change" (nil) from "set to zero/empty".
// String fields use empty string to mean "don't change".
type UpdateDatasetOpts struct {
	Compression string  // Empty = don't change
	Quota       *int64
	RefQuota    *int64
	Atime       string  // Empty = don't change
	Comments    *string
}

// Zvol is the user-facing representation of a TrueNAS zvol.
type Zvol struct {
	ID           string
	Name         string
	Pool         string
	Comments     string
	Compression  string
	Volsize      int64
	Volblocksize string
	Sparse       bool
}

// CreateZvolOpts contains options for creating a zvol.
type CreateZvolOpts struct {
	Name         string
	Volsize      int64
	Volblocksize string
	Sparse       bool
	ForceSize    bool
	Compression  string
	Comments     string
}

// UpdateZvolOpts contains options for updating a zvol.
// Pointer fields distinguish "don't change" (nil) from "set to zero/empty".
// String fields use empty string to mean "don't change".
type UpdateZvolOpts struct {
	Volsize     *int64
	ForceSize   bool   // Only sent when true
	Compression string // Empty = don't change
	Comments    *string
}

// Pool is the user-facing representation of a TrueNAS pool.
type Pool struct {
	ID        int64
	Name      string
	Path      string
	Status    string
	Size      int64
	Allocated int64
	Free      int64
}

// Int64Ptr returns a pointer to an int64. Helper for setting optional fields.
func Int64Ptr(v int64) *int64 { return &v }

// StringPtr returns a pointer to a string. Helper for setting optional fields.
func StringPtr(v string) *string { return &v }

// DatasetService provides typed methods for the pool.dataset.* and pool.query API namespaces.
type DatasetService struct {
	client  Caller
	version Version
}

// NewDatasetService creates a new DatasetService.
func NewDatasetService(c Caller, v Version) *DatasetService {
	return &DatasetService{client: c, version: v}
}

// CreateDataset creates a filesystem dataset and returns the full object.
func (s *DatasetService) CreateDataset(ctx context.Context, opts CreateDatasetOpts) (*Dataset, error) {
	params := datasetCreateParams(opts)
	result, err := s.client.Call(ctx, "pool.dataset.create", params)
	if err != nil {
		return nil, err
	}

	var createResp DatasetCreateResponse
	if err := json.Unmarshal(result, &createResp); err != nil {
		return nil, fmt.Errorf("parse create response: %w", err)
	}

	return s.GetDataset(ctx, createResp.ID)
}

// GetDataset returns a dataset by ID, or nil if not found.
func (s *DatasetService) GetDataset(ctx context.Context, id string) (*Dataset, error) {
	filter := [][]any{{"id", "=", id}}
	result, err := s.client.Call(ctx, "pool.dataset.query", filter)
	if err != nil {
		if isNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}

	var responses []DatasetResponse
	if err := json.Unmarshal(result, &responses); err != nil {
		return nil, fmt.Errorf("parse query response: %w", err)
	}

	if len(responses) == 0 {
		return nil, nil
	}

	ds := datasetFromResponse(responses[0])
	return &ds, nil
}

// ListDatasets returns all filesystem datasets (type FILESYSTEM only).
func (s *DatasetService) ListDatasets(ctx context.Context) ([]Dataset, error) {
	result, err := s.client.Call(ctx, "pool.dataset.query", nil)
	if err != nil {
		return nil, err
	}

	var responses []DatasetResponse
	if err := json.Unmarshal(result, &responses); err != nil {
		return nil, fmt.Errorf("parse query response: %w", err)
	}

	var datasets []Dataset
	for _, resp := range responses {
		if resp.Type == "FILESYSTEM" {
			datasets = append(datasets, datasetFromResponse(resp))
		}
	}
	if datasets == nil {
		datasets = []Dataset{}
	}
	return datasets, nil
}

// UpdateDataset updates a dataset and returns the full object.
func (s *DatasetService) UpdateDataset(ctx context.Context, id string, opts UpdateDatasetOpts) (*Dataset, error) {
	params := datasetUpdateParams(opts)
	_, err := s.client.Call(ctx, "pool.dataset.update", []any{id, params})
	if err != nil {
		return nil, err
	}

	return s.GetDataset(ctx, id)
}

// DeleteDataset deletes a dataset by ID. If recursive is true, child datasets are also deleted.
func (s *DatasetService) DeleteDataset(ctx context.Context, id string, recursive bool) error {
	var params any
	if recursive {
		params = []any{id, map[string]any{"recursive": true}}
	} else {
		params = id
	}
	_, err := s.client.Call(ctx, "pool.dataset.delete", params)
	return err
}

// CreateZvol creates a zvol and returns the full object.
func (s *DatasetService) CreateZvol(ctx context.Context, opts CreateZvolOpts) (*Zvol, error) {
	params := zvolCreateParams(opts)
	result, err := s.client.Call(ctx, "pool.dataset.create", params)
	if err != nil {
		return nil, err
	}

	var createResp DatasetCreateResponse
	if err := json.Unmarshal(result, &createResp); err != nil {
		return nil, fmt.Errorf("parse create response: %w", err)
	}

	return s.GetZvol(ctx, createResp.ID)
}

// GetZvol returns a zvol by ID, or nil if not found.
func (s *DatasetService) GetZvol(ctx context.Context, id string) (*Zvol, error) {
	filter := [][]any{{"id", "=", id}}
	result, err := s.client.Call(ctx, "pool.dataset.query", filter)
	if err != nil {
		if isNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}

	var responses []DatasetResponse
	if err := json.Unmarshal(result, &responses); err != nil {
		return nil, fmt.Errorf("parse query response: %w", err)
	}

	if len(responses) == 0 {
		return nil, nil
	}

	zvol := zvolFromResponse(responses[0])
	return &zvol, nil
}

// UpdateZvol updates a zvol and returns the full object.
func (s *DatasetService) UpdateZvol(ctx context.Context, id string, opts UpdateZvolOpts) (*Zvol, error) {
	params := zvolUpdateParams(opts)
	_, err := s.client.Call(ctx, "pool.dataset.update", []any{id, params})
	if err != nil {
		return nil, err
	}

	return s.GetZvol(ctx, id)
}

// DeleteZvol deletes a zvol by ID.
func (s *DatasetService) DeleteZvol(ctx context.Context, id string) error {
	_, err := s.client.Call(ctx, "pool.dataset.delete", id)
	return err
}

// ListPools returns all pools.
func (s *DatasetService) ListPools(ctx context.Context) ([]Pool, error) {
	result, err := s.client.Call(ctx, "pool.query", nil)
	if err != nil {
		return nil, err
	}

	var responses []PoolResponse
	if err := json.Unmarshal(result, &responses); err != nil {
		return nil, fmt.Errorf("parse query response: %w", err)
	}

	pools := make([]Pool, len(responses))
	for i, resp := range responses {
		pools[i] = poolFromResponse(resp)
	}
	return pools, nil
}

// datasetFromResponse converts a wire-format DatasetResponse to a user-facing Dataset.
func datasetFromResponse(resp DatasetResponse) Dataset {
	return Dataset{
		ID:          resp.ID,
		Name:        resp.Name,
		Pool:        resp.Pool,
		Mountpoint:  resp.Mountpoint,
		Comments:    resp.Comments.Value,
		Compression: resp.Compression.Value,
		Quota:       resp.Quota.Parsed,
		RefQuota:    resp.RefQuota.Parsed,
		Atime:       resp.Atime.Value,
	}
}

// zvolFromResponse converts a wire-format DatasetResponse to a user-facing Zvol.
func zvolFromResponse(resp DatasetResponse) Zvol {
	return Zvol{
		ID:           resp.ID,
		Name:         resp.Name,
		Pool:         resp.Pool,
		Comments:     resp.Comments.Value,
		Compression:  resp.Compression.Value,
		Volsize:      resp.Volsize.Parsed,
		Volblocksize: resp.Volblocksize.Value,
		Sparse:       resp.Sparse.Value == "true",
	}
}

// poolFromResponse converts a wire-format PoolResponse to a user-facing Pool.
func poolFromResponse(resp PoolResponse) Pool {
	return Pool{
		ID:        resp.ID,
		Name:      resp.Name,
		Path:      resp.Path,
		Status:    resp.Status,
		Size:      resp.Size,
		Allocated: resp.Allocated,
		Free:      resp.Free,
	}
}

// datasetCreateParams builds API parameters for pool.dataset.create (filesystem).
func datasetCreateParams(opts CreateDatasetOpts) map[string]any {
	params := map[string]any{
		"name": opts.Name,
		"type": "FILESYSTEM",
	}
	if opts.Comments != "" {
		params["comments"] = opts.Comments
	}
	if opts.Compression != "" {
		params["compression"] = opts.Compression
	}
	if opts.Quota != 0 {
		params["quota"] = opts.Quota
	}
	if opts.RefQuota != 0 {
		params["refquota"] = opts.RefQuota
	}
	if opts.Atime != "" {
		params["atime"] = opts.Atime
	}
	return params
}

// datasetUpdateParams builds API parameters for pool.dataset.update (filesystem).
func datasetUpdateParams(opts UpdateDatasetOpts) map[string]any {
	params := map[string]any{}
	if opts.Compression != "" {
		params["compression"] = opts.Compression
	}
	if opts.Quota != nil {
		params["quota"] = *opts.Quota
	}
	if opts.RefQuota != nil {
		params["refquota"] = *opts.RefQuota
	}
	if opts.Atime != "" {
		params["atime"] = opts.Atime
	}
	if opts.Comments != nil {
		params["comments"] = *opts.Comments
	}
	return params
}

// zvolCreateParams builds API parameters for pool.dataset.create (zvol).
func zvolCreateParams(opts CreateZvolOpts) map[string]any {
	params := map[string]any{
		"name":    opts.Name,
		"type":    "VOLUME",
		"volsize": opts.Volsize,
	}
	if opts.Volblocksize != "" {
		params["volblocksize"] = opts.Volblocksize
	}
	if opts.Sparse {
		params["sparse"] = true
	}
	if opts.ForceSize {
		params["force_size"] = true
	}
	if opts.Compression != "" {
		params["compression"] = opts.Compression
	}
	if opts.Comments != "" {
		params["comments"] = opts.Comments
	}
	return params
}

// zvolUpdateParams builds API parameters for pool.dataset.update (zvol).
func zvolUpdateParams(opts UpdateZvolOpts) map[string]any {
	params := map[string]any{}
	if opts.Volsize != nil {
		params["volsize"] = *opts.Volsize
	}
	if opts.ForceSize {
		params["force_size"] = true
	}
	if opts.Compression != "" {
		params["compression"] = opts.Compression
	}
	if opts.Comments != nil {
		params["comments"] = *opts.Comments
	}
	return params
}

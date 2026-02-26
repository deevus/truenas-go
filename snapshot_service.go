package truenas

import (
	"context"
	"encoding/json"
	"fmt"
)

// Snapshot method names (without prefix).
const (
	methodSnapshotCreate  = "create"
	methodSnapshotQuery   = "query"
	methodSnapshotDelete  = "delete"
	methodSnapshotHold    = "hold"
	methodSnapshotRelease = "release"
	methodSnapshotClone    = "clone"
	methodSnapshotRollback = "rollback"
)

// resolveSnapshotMethod returns the full API method name for the given version.
// Pre-25.10 uses "zfs.snapshot.*", 25.10+ uses "pool.snapshot.*".
func resolveSnapshotMethod(v Version, method string) string {
	prefix := "zfs.snapshot"
	if v.AtLeast(25, 10) {
		prefix = "pool.snapshot"
	}
	return prefix + "." + method
}

// Snapshot is the user-facing representation of a TrueNAS ZFS snapshot.
type Snapshot struct {
	ID           string
	Dataset      string
	SnapshotName string
	CreateTXG    string
	Used         int64
	Referenced   int64
	HasHold      bool
}

// CreateSnapshotOpts contains options for creating a snapshot.
type CreateSnapshotOpts struct {
	Dataset   string
	Name      string
	Recursive bool
}

// SnapshotService provides typed methods for the snapshot API namespace.
type SnapshotService struct {
	client  Caller
	version Version
}

// NewSnapshotService creates a new SnapshotService.
func NewSnapshotService(c Caller, v Version) *SnapshotService {
	return &SnapshotService{client: c, version: v}
}

// Create creates a snapshot and returns the full object.
func (s *SnapshotService) Create(ctx context.Context, opts CreateSnapshotOpts) (*Snapshot, error) {
	params := map[string]any{
		"dataset": opts.Dataset,
		"name":    opts.Name,
	}
	if opts.Recursive {
		params["recursive"] = true
	}

	method := resolveSnapshotMethod(s.version, methodSnapshotCreate)
	_, err := s.client.Call(ctx, method, params)
	if err != nil {
		return nil, err
	}

	id := opts.Dataset + "@" + opts.Name
	return s.Get(ctx, id)
}

// Get returns a snapshot by ID, or nil if not found.
func (s *SnapshotService) Get(ctx context.Context, id string) (*Snapshot, error) {
	filter := [][]any{{"id", "=", id}}
	method := resolveSnapshotMethod(s.version, methodSnapshotQuery)
	result, err := s.client.Call(ctx, method, filter)
	if err != nil {
		return nil, err
	}

	var snapshots []SnapshotResponse
	if err := json.Unmarshal(result, &snapshots); err != nil {
		return nil, fmt.Errorf("parse query response: %w", err)
	}

	if len(snapshots) == 0 {
		return nil, nil
	}

	snap := snapshotFromResponse(snapshots[0])
	return &snap, nil
}

// List returns all snapshots.
func (s *SnapshotService) List(ctx context.Context) ([]Snapshot, error) {
	method := resolveSnapshotMethod(s.version, methodSnapshotQuery)
	result, err := s.client.Call(ctx, method, nil)
	if err != nil {
		return nil, err
	}

	var responses []SnapshotResponse
	if err := json.Unmarshal(result, &responses); err != nil {
		return nil, fmt.Errorf("parse query response: %w", err)
	}

	snapshots := make([]Snapshot, len(responses))
	for i, resp := range responses {
		snapshots[i] = snapshotFromResponse(resp)
	}
	return snapshots, nil
}

// Delete deletes a snapshot by ID.
func (s *SnapshotService) Delete(ctx context.Context, id string) error {
	method := resolveSnapshotMethod(s.version, methodSnapshotDelete)
	_, err := s.client.Call(ctx, method, id)
	return err
}

// Hold places a hold on a snapshot.
func (s *SnapshotService) Hold(ctx context.Context, id string) error {
	method := resolveSnapshotMethod(s.version, methodSnapshotHold)
	_, err := s.client.Call(ctx, method, id)
	return err
}

// Release releases a hold on a snapshot.
func (s *SnapshotService) Release(ctx context.Context, id string) error {
	method := resolveSnapshotMethod(s.version, methodSnapshotRelease)
	_, err := s.client.Call(ctx, method, id)
	return err
}

// Query returns snapshots matching the given filters.
// Filters use TrueNAS query format: [][]any{{"field", "op", "value"}}.
// Pass nil for no filtering (equivalent to List).
func (s *SnapshotService) Query(ctx context.Context, filters [][]any) ([]Snapshot, error) {
	var params any
	if len(filters) > 0 {
		params = filters
	}

	method := resolveSnapshotMethod(s.version, methodSnapshotQuery)
	result, err := s.client.Call(ctx, method, params)
	if err != nil {
		return nil, err
	}

	var responses []SnapshotResponse
	if err := json.Unmarshal(result, &responses); err != nil {
		return nil, fmt.Errorf("parse query response: %w", err)
	}

	snapshots := make([]Snapshot, len(responses))
	for i, resp := range responses {
		snapshots[i] = snapshotFromResponse(resp)
	}
	return snapshots, nil
}

// Rollback rolls back to the given snapshot by ID (dataset@name).
func (s *SnapshotService) Rollback(ctx context.Context, id string) error {
	method := resolveSnapshotMethod(s.version, methodSnapshotRollback)
	_, err := s.client.Call(ctx, method, id)
	return err
}

// Clone clones a snapshot to a new dataset.
func (s *SnapshotService) Clone(ctx context.Context, snapshot, datasetDst string) error {
	params := map[string]any{
		"snapshot":    snapshot,
		"dataset_dst": datasetDst,
	}
	method := resolveSnapshotMethod(s.version, methodSnapshotClone)
	_, err := s.client.Call(ctx, method, params)
	return err
}

// snapshotFromResponse converts a wire-format SnapshotResponse to a user-facing Snapshot.
func snapshotFromResponse(resp SnapshotResponse) Snapshot {
	return Snapshot{
		ID:           resp.ID,
		Dataset:      resp.Dataset,
		SnapshotName: resp.SnapshotName,
		CreateTXG:    resp.Properties.CreateTXG.Value,
		Used:         resp.Properties.Used.Parsed,
		Referenced:   resp.Properties.Referenced.Parsed,
		HasHold:      resp.HasHold(),
	}
}

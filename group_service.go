package truenas

import (
	"context"
	"encoding/json"
	"fmt"
)

// Group is the user-facing representation of a TrueNAS group.
type Group struct {
	ID                   int64
	GID                  int64
	Name                 string
	Builtin              bool
	SMB                  bool
	SudoCommands         []string
	SudoCommandsNopasswd []string
	Users                []int64
	Local                bool
	Immutable            bool
}

// CreateGroupOpts contains options for creating a group.
type CreateGroupOpts struct {
	Name                 string
	GID                  int64 // 0 = auto-assign
	SMB                  bool
	SudoCommands         []string
	SudoCommandsNopasswd []string
}

// UpdateGroupOpts contains options for updating a group.
// GID is immutable and cannot be changed after creation.
type UpdateGroupOpts struct {
	Name                 string
	SMB                  bool
	SudoCommands         []string
	SudoCommandsNopasswd []string
}

// GroupService provides typed methods for the group.* API namespace.
type GroupService struct {
	client  Caller
	version Version
}

// NewGroupService creates a new GroupService.
func NewGroupService(c Caller, v Version) *GroupService {
	return &GroupService{client: c, version: v}
}

// Create creates a group and returns the full object.
func (s *GroupService) Create(ctx context.Context, opts CreateGroupOpts) (*Group, error) {
	params := groupCreateOptsToParams(opts)
	result, err := s.client.Call(ctx, "group.create", params)
	if err != nil {
		return nil, err
	}

	var id int64
	if err := json.Unmarshal(result, &id); err != nil {
		return nil, fmt.Errorf("parse create response: %w", err)
	}

	return s.Get(ctx, id)
}

// Get returns a group by ID, or nil if not found.
func (s *GroupService) Get(ctx context.Context, id int64) (*Group, error) {
	result, err := s.client.Call(ctx, "group.get_instance", id)
	if err != nil {
		if isNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}

	var resp GroupResponse
	if err := json.Unmarshal(result, &resp); err != nil {
		return nil, fmt.Errorf("parse get_instance response: %w", err)
	}

	group := groupFromResponse(resp)
	return &group, nil
}

// GetByName returns a group by name, or nil if not found.
func (s *GroupService) GetByName(ctx context.Context, name string) (*Group, error) {
	return s.queryOne(ctx, "group", name)
}

// GetByGID returns a group by GID, or nil if not found.
func (s *GroupService) GetByGID(ctx context.Context, gid int64) (*Group, error) {
	return s.queryOne(ctx, "gid", gid)
}

// List returns all groups.
func (s *GroupService) List(ctx context.Context) ([]Group, error) {
	result, err := s.client.Call(ctx, "group.query", nil)
	if err != nil {
		return nil, err
	}

	var responses []GroupResponse
	if err := json.Unmarshal(result, &responses); err != nil {
		return nil, fmt.Errorf("parse query response: %w", err)
	}

	groups := make([]Group, len(responses))
	for i, resp := range responses {
		groups[i] = groupFromResponse(resp)
	}
	return groups, nil
}

// Update updates a group and returns the full object.
func (s *GroupService) Update(ctx context.Context, id int64, opts UpdateGroupOpts) (*Group, error) {
	params := groupUpdateOptsToParams(opts)
	_, err := s.client.Call(ctx, "group.update", []any{id, params})
	if err != nil {
		return nil, err
	}

	return s.Get(ctx, id)
}

// Delete deletes a group by ID. Does not delete member users.
func (s *GroupService) Delete(ctx context.Context, id int64) error {
	_, err := s.client.Call(ctx, "group.delete", []any{id, map[string]any{"delete_users": false}})
	return err
}

// queryOne queries for a single group by field and value.
func (s *GroupService) queryOne(ctx context.Context, field string, value any) (*Group, error) {
	filter := [][]any{{field, "=", value}}
	result, err := s.client.Call(ctx, "group.query", filter)
	if err != nil {
		return nil, err
	}

	var responses []GroupResponse
	if err := json.Unmarshal(result, &responses); err != nil {
		return nil, fmt.Errorf("parse query response: %w", err)
	}

	if len(responses) == 0 {
		return nil, nil
	}

	group := groupFromResponse(responses[0])
	return &group, nil
}

// groupCreateOptsToParams converts CreateGroupOpts to API parameters.
func groupCreateOptsToParams(opts CreateGroupOpts) map[string]any {
	params := map[string]any{
		"name": opts.Name,
		"smb":  opts.SMB,
	}
	if opts.GID != 0 {
		params["gid"] = opts.GID
	}
	if opts.SudoCommands != nil {
		params["sudo_commands"] = opts.SudoCommands
	}
	if opts.SudoCommandsNopasswd != nil {
		params["sudo_commands_nopasswd"] = opts.SudoCommandsNopasswd
	}
	return params
}

// groupUpdateOptsToParams converts UpdateGroupOpts to API parameters.
func groupUpdateOptsToParams(opts UpdateGroupOpts) map[string]any {
	params := map[string]any{
		"name": opts.Name,
		"smb":  opts.SMB,
	}
	if opts.SudoCommands != nil {
		params["sudo_commands"] = opts.SudoCommands
	}
	if opts.SudoCommandsNopasswd != nil {
		params["sudo_commands_nopasswd"] = opts.SudoCommandsNopasswd
	}
	return params
}

// groupFromResponse converts a wire-format GroupResponse to a user-facing Group.
func groupFromResponse(resp GroupResponse) Group {
	return Group{
		ID:                   resp.ID,
		GID:                  resp.GID,
		Name:                 resp.Name,
		Builtin:              resp.Builtin,
		SMB:                  resp.SMB,
		SudoCommands:         resp.SudoCommands,
		SudoCommandsNopasswd: resp.SudoCommandsNopasswd,
		Users:                resp.Users,
		Local:                resp.Local,
		Immutable:            resp.Immutable,
	}
}

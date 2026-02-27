package truenas

import (
	"context"
	"encoding/json"
	"fmt"
)

// User is the user-facing representation of a TrueNAS user.
type User struct {
	ID                   int64
	UID                  int64
	Username             string
	FullName             string
	Email                string
	Home                 string
	Shell                string
	HomeMode             string
	GroupID              int64
	Groups               []int64
	SMB                  bool
	PasswordDisabled     bool
	SSHPasswordEnabled   bool
	SSHPubKey            string
	Locked               bool
	SudoCommands         []string
	SudoCommandsNopasswd []string
	Builtin              bool
	Local                bool
	Immutable            bool
}

// CreateUserOpts contains options for creating a user.
type CreateUserOpts struct {
	Username             string
	FullName             string
	Email                string
	UID                  int64 // 0 = auto-assign
	Password             string
	PasswordDisabled     bool
	Group                int64 // primary group ID; 0 = omit
	GroupCreate          bool
	Groups               []int64
	Home                 string
	HomeCreate           bool
	HomeMode             string
	Shell                string
	SMB                  bool
	SSHPasswordEnabled   bool
	SSHPubKey            string
	Locked               bool
	SudoCommands         []string
	SudoCommandsNopasswd []string
}

// UpdateUserOpts contains options for updating a user.
// UID, GroupCreate, and HomeCreate are immutable after creation.
type UpdateUserOpts struct {
	Username             string
	FullName             string
	Email                string
	Password             string
	PasswordDisabled     bool
	Group                int64
	Groups               []int64
	Home                 string
	HomeMode             string
	Shell                string
	SMB                  bool
	SSHPasswordEnabled   bool
	SSHPubKey            string
	Locked               bool
	SudoCommands         []string
	SudoCommandsNopasswd []string
}

// UserService provides typed methods for the user.* API namespace.
type UserService struct {
	client  Caller
	version Version
}

// NewUserService creates a new UserService.
func NewUserService(c Caller, v Version) *UserService {
	return &UserService{client: c, version: v}
}

// Create creates a user and returns the full object.
func (s *UserService) Create(ctx context.Context, opts CreateUserOpts) (*User, error) {
	params := userCreateOptsToParams(opts)
	result, err := s.client.Call(ctx, "user.create", params)
	if err != nil {
		return nil, err
	}

	var createResp struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(result, &createResp); err != nil {
		return nil, fmt.Errorf("parse create response: %w", err)
	}

	return s.Get(ctx, createResp.ID)
}

// Get returns a user by ID, or nil if not found.
func (s *UserService) Get(ctx context.Context, id int64) (*User, error) {
	result, err := s.client.Call(ctx, "user.get_instance", id)
	if err != nil {
		if isNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}

	var resp UserResponse
	if err := json.Unmarshal(result, &resp); err != nil {
		return nil, fmt.Errorf("parse get_instance response: %w", err)
	}

	user := userFromResponse(resp)
	return &user, nil
}

// GetByUsername returns a user by username, or nil if not found.
func (s *UserService) GetByUsername(ctx context.Context, username string) (*User, error) {
	return s.queryOne(ctx, "username", username)
}

// GetByUID returns a user by UID, or nil if not found.
func (s *UserService) GetByUID(ctx context.Context, uid int64) (*User, error) {
	return s.queryOne(ctx, "uid", uid)
}

// List returns all users.
func (s *UserService) List(ctx context.Context) ([]User, error) {
	result, err := s.client.Call(ctx, "user.query", nil)
	if err != nil {
		return nil, err
	}

	var responses []UserResponse
	if err := json.Unmarshal(result, &responses); err != nil {
		return nil, fmt.Errorf("parse query response: %w", err)
	}

	users := make([]User, len(responses))
	for i, resp := range responses {
		users[i] = userFromResponse(resp)
	}
	return users, nil
}

// Update updates a user and returns the full object.
func (s *UserService) Update(ctx context.Context, id int64, opts UpdateUserOpts) (*User, error) {
	params := userUpdateOptsToParams(opts)
	_, err := s.client.Call(ctx, "user.update", []any{id, params})
	if err != nil {
		return nil, err
	}

	return s.Get(ctx, id)
}

// Delete deletes a user by ID. Also deletes the user's auto-created primary group.
func (s *UserService) Delete(ctx context.Context, id int64) error {
	_, err := s.client.Call(ctx, "user.delete", []any{id, map[string]any{"delete_group": true}})
	return err
}

// queryOne queries for a single user by field and value.
func (s *UserService) queryOne(ctx context.Context, field string, value any) (*User, error) {
	filter := [][]any{{field, "=", value}}
	result, err := s.client.Call(ctx, "user.query", filter)
	if err != nil {
		return nil, err
	}

	var responses []UserResponse
	if err := json.Unmarshal(result, &responses); err != nil {
		return nil, fmt.Errorf("parse query response: %w", err)
	}

	if len(responses) == 0 {
		return nil, nil
	}

	user := userFromResponse(responses[0])
	return &user, nil
}

// userCreateOptsToParams converts CreateUserOpts to API parameters.
func userCreateOptsToParams(opts CreateUserOpts) map[string]any {
	params := map[string]any{
		"username":             opts.Username,
		"full_name":            opts.FullName,
		"email":                opts.Email,
		"password_disabled":    opts.PasswordDisabled,
		"home":                 opts.Home,
		"home_mode":            opts.HomeMode,
		"shell":                opts.Shell,
		"smb":                  opts.SMB,
		"ssh_password_enabled": opts.SSHPasswordEnabled,
		"locked":               opts.Locked,
	}
	if opts.UID != 0 {
		params["uid"] = opts.UID
	}
	if opts.GroupCreate {
		params["group_create"] = true
	}
	if opts.HomeCreate {
		params["home_create"] = true
	}
	if opts.Password != "" {
		params["password"] = opts.Password
	}
	if opts.Group != 0 {
		params["group"] = opts.Group
	}
	if opts.Groups != nil {
		params["groups"] = opts.Groups
	}
	if opts.SSHPubKey != "" {
		params["sshpubkey"] = opts.SSHPubKey
	}
	if opts.SudoCommands != nil {
		params["sudo_commands"] = opts.SudoCommands
	}
	if opts.SudoCommandsNopasswd != nil {
		params["sudo_commands_nopasswd"] = opts.SudoCommandsNopasswd
	}
	return params
}

// userUpdateOptsToParams converts UpdateUserOpts to API parameters.
// Excludes UID, GroupCreate, and HomeCreate (immutable after creation).
func userUpdateOptsToParams(opts UpdateUserOpts) map[string]any {
	params := map[string]any{
		"username":             opts.Username,
		"full_name":            opts.FullName,
		"email":                opts.Email,
		"password_disabled":    opts.PasswordDisabled,
		"home":                 opts.Home,
		"home_mode":            opts.HomeMode,
		"shell":                opts.Shell,
		"smb":                  opts.SMB,
		"ssh_password_enabled": opts.SSHPasswordEnabled,
		"locked":               opts.Locked,
	}
	if opts.Password != "" {
		params["password"] = opts.Password
	}
	if opts.Group != 0 {
		params["group"] = opts.Group
	}
	if opts.Groups != nil {
		params["groups"] = opts.Groups
	}
	if opts.SSHPubKey != "" {
		params["sshpubkey"] = opts.SSHPubKey
	}
	if opts.SudoCommands != nil {
		params["sudo_commands"] = opts.SudoCommands
	}
	if opts.SudoCommandsNopasswd != nil {
		params["sudo_commands_nopasswd"] = opts.SudoCommandsNopasswd
	}
	return params
}

// userFromResponse converts a wire-format UserResponse to a user-facing User.
func userFromResponse(resp UserResponse) User {
	var email string
	if resp.Email != nil {
		email = *resp.Email
	}
	var sshPubKey string
	if resp.SSHPubKey != nil {
		sshPubKey = *resp.SSHPubKey
	}

	return User{
		ID:                   resp.ID,
		UID:                  resp.UID,
		Username:             resp.Username,
		FullName:             resp.FullName,
		Email:                email,
		Home:                 resp.Home,
		Shell:                resp.Shell,
		HomeMode:             resp.HomeMode,
		GroupID:              resp.Group.ID,
		Groups:               resp.Groups,
		SMB:                  resp.SMB,
		PasswordDisabled:     resp.PasswordDisabled,
		SSHPasswordEnabled:   resp.SSHPasswordEnabled,
		SSHPubKey:            sshPubKey,
		Locked:               resp.Locked,
		SudoCommands:         resp.SudoCommands,
		SudoCommandsNopasswd: resp.SudoCommandsNopasswd,
		Builtin:              resp.Builtin,
		Local:                resp.Local,
		Immutable:            resp.Immutable,
	}
}

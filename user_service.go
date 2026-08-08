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
	HomeMode             string // write-only; the API never returns it
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
	PasswordDisabled     *bool
	Group                int64
	Groups               []int64
	Home                 string
	HomeMode             string // write-only; the API never returns it
	Shell                string
	SMB                  *bool
	SSHPasswordEnabled   *bool
	SSHPubKey            string
	Locked               *bool
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

	id, err := parseCreatedUserID(result)
	if err != nil {
		return nil, err
	}

	user, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("user %d not found after create", id)
	}
	return user, nil
}

// parseCreatedUserID extracts the new user's ID from a user.create response.
// TrueNAS 24.x returns the bare primary key; 25.04+ returns the full user
// object. Both shapes are accepted so the caller does not have to know which
// version it is talking to.
func parseCreatedUserID(result json.RawMessage) (int64, error) {
	var id int64
	if err := json.Unmarshal(result, &id); err == nil {
		if id < 1 {
			return 0, fmt.Errorf("parse create response: invalid user id %d", id)
		}
		return id, nil
	}

	var createResp struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(result, &createResp); err != nil {
		return 0, fmt.Errorf("parse create response: %w", err)
	}
	if createResp.ID < 1 {
		return 0, fmt.Errorf("parse create response: invalid user id %d", createResp.ID)
	}

	return createResp.ID, nil
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

	user, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("user %d not found after update", id)
	}
	return user, nil
}

// Delete deletes a user by ID. Pass deleteGroup=true when the user was created
// with group_create=true so the auto-created primary group is cleaned up;
// pass false when the primary group is managed separately.
func (s *UserService) Delete(ctx context.Context, id int64, deleteGroup bool) error {
	_, err := s.client.Call(ctx, "user.delete", []any{id, map[string]any{"delete_group": deleteGroup}})
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

// emailParam returns the value to send for the API's email field. The API types
// it as an email address or null, so an empty string is rejected by validation;
// null is how an address is cleared. Always sent so it can be cleared on update.
func emailParam(email string) any {
	if email == "" {
		return nil
	}
	return email
}

// userCreateOptsToParams converts CreateUserOpts to API parameters.
// Home, HomeMode, and Shell are omitted when empty so the API applies its own
// defaults — sending an empty string for them fails validation.
func userCreateOptsToParams(opts CreateUserOpts) map[string]any {
	params := map[string]any{
		"username":             opts.Username,
		"full_name":            opts.FullName,
		"email":                emailParam(opts.Email),
		"password_disabled":    opts.PasswordDisabled,
		"smb":                  opts.SMB,
		"ssh_password_enabled": opts.SSHPasswordEnabled,
		"locked":               opts.Locked,
	}
	if opts.UID != 0 {
		params["uid"] = opts.UID
	}
	if opts.Home != "" {
		params["home"] = opts.Home
	}
	if opts.HomeMode != "" {
		params["home_mode"] = opts.HomeMode
	}
	if opts.Shell != "" {
		params["shell"] = opts.Shell
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
// Home, HomeMode, and Shell are omitted when empty, leaving them unchanged —
// sending an empty string for them fails validation.
func userUpdateOptsToParams(opts UpdateUserOpts) map[string]any {
	params := map[string]any{
		"email": emailParam(opts.Email),
	}
	if opts.Username != "" {
		params["username"] = opts.Username
	}
	if opts.FullName != "" {
		params["full_name"] = opts.FullName
	}
	if opts.PasswordDisabled != nil {
		params["password_disabled"] = *opts.PasswordDisabled
	}
	if opts.SMB != nil {
		params["smb"] = *opts.SMB
	}
	if opts.SSHPasswordEnabled != nil {
		params["ssh_password_enabled"] = *opts.SSHPasswordEnabled
	}
	if opts.Locked != nil {
		params["locked"] = *opts.Locked
	}
	if opts.Home != "" {
		params["home"] = opts.Home
	}
	if opts.HomeMode != "" {
		params["home_mode"] = opts.HomeMode
	}
	if opts.Shell != "" {
		params["shell"] = opts.Shell
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

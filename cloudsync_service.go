package truenas

import (
	"context"
	"encoding/json"
	"fmt"
)

// CloudSyncCredential is the user-facing representation of a cloud sync credential.
type CloudSyncCredential struct {
	ID           int64
	Name         string
	ProviderType string
	Attributes   map[string]string
}

// CreateCredentialOpts contains options for creating a cloud sync credential.
type CreateCredentialOpts struct {
	Name         string
	ProviderType string
	Attributes   map[string]string
}

// UpdateCredentialOpts contains options for updating a cloud sync credential.
type UpdateCredentialOpts = CreateCredentialOpts

// CloudSyncTask is the user-facing representation of a cloud sync task.
type CloudSyncTask struct {
	ID                 int64
	Description        string
	Path               string
	CredentialID       int64
	Direction          string
	TransferMode       string
	Snapshot           bool
	Transfers          int64
	BWLimit            []BwLimit
	FollowSymlinks     bool
	CreateEmptySrcDirs bool
	FastList           bool
	Enabled            bool
	Encryption         bool
	EncryptionPassword string
	EncryptionSalt     string
	Schedule           Schedule
	Attributes         map[string]any
	Exclude            []string
	Include            []string
}

// CreateCloudSyncTaskOpts contains options for creating a cloud sync task.
type CreateCloudSyncTaskOpts struct {
	Description        string
	Path               string
	CredentialID       int64
	Direction          string
	TransferMode       string
	Snapshot           bool
	Transfers          int64
	BWLimit            []BwLimit
	FollowSymlinks     bool
	CreateEmptySrcDirs bool
	FastList           bool
	Enabled            bool
	Encryption         bool
	EncryptionPassword string
	EncryptionSalt     string
	Schedule           Schedule
	Attributes         map[string]any
	Exclude            []string
	Include            []string
}

// UpdateCloudSyncTaskOpts contains options for updating a cloud sync task.
type UpdateCloudSyncTaskOpts = CreateCloudSyncTaskOpts

// CloudSyncService provides typed methods for the cloudsync.* API namespace.
type CloudSyncService struct {
	client  AsyncCaller
	version Version
}

// NewCloudSyncService creates a new CloudSyncService.
func NewCloudSyncService(c AsyncCaller, v Version) *CloudSyncService {
	return &CloudSyncService{client: c, version: v}
}

// CreateCredential creates a cloud sync credential and returns the full object.
func (s *CloudSyncService) CreateCredential(ctx context.Context, opts CreateCredentialOpts) (*CloudSyncCredential, error) {
	attrs := credentialOptsToAttrsAny(opts.Attributes)
	params := BuildCredentialsParams(s.version, opts.Name, opts.ProviderType, attrs)

	result, err := s.client.Call(ctx, "cloudsync.credentials.create", params)
	if err != nil {
		return nil, err
	}

	var createResp struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(result, &createResp); err != nil {
		return nil, fmt.Errorf("parse create response: %w", err)
	}

	return s.GetCredential(ctx, createResp.ID)
}

// GetCredential returns a cloud sync credential by ID, or nil if not found.
func (s *CloudSyncService) GetCredential(ctx context.Context, id int64) (*CloudSyncCredential, error) {
	filter := [][]any{{"id", "=", id}}
	result, err := s.client.Call(ctx, "cloudsync.credentials.query", filter)
	if err != nil {
		return nil, err
	}

	creds, err := ParseCredentials(result, s.version)
	if err != nil {
		return nil, err
	}

	if len(creds) == 0 {
		return nil, nil
	}

	cred := credentialFromResponse(creds[0])
	return &cred, nil
}

// ListCredentials returns all cloud sync credentials.
func (s *CloudSyncService) ListCredentials(ctx context.Context) ([]CloudSyncCredential, error) {
	result, err := s.client.Call(ctx, "cloudsync.credentials.query", nil)
	if err != nil {
		return nil, err
	}

	responses, err := ParseCredentials(result, s.version)
	if err != nil {
		return nil, err
	}

	creds := make([]CloudSyncCredential, len(responses))
	for i, resp := range responses {
		creds[i] = credentialFromResponse(resp)
	}
	return creds, nil
}

// UpdateCredential updates a cloud sync credential and returns the full object.
func (s *CloudSyncService) UpdateCredential(ctx context.Context, id int64, opts UpdateCredentialOpts) (*CloudSyncCredential, error) {
	attrs := credentialOptsToAttrsAny(opts.Attributes)
	params := BuildCredentialsParams(s.version, opts.Name, opts.ProviderType, attrs)

	_, err := s.client.Call(ctx, "cloudsync.credentials.update", []any{id, params})
	if err != nil {
		return nil, err
	}

	return s.GetCredential(ctx, id)
}

// DeleteCredential deletes a cloud sync credential by ID.
func (s *CloudSyncService) DeleteCredential(ctx context.Context, id int64) error {
	_, err := s.client.Call(ctx, "cloudsync.credentials.delete", id)
	return err
}

// CreateTask creates a cloud sync task and returns the full object.
func (s *CloudSyncService) CreateTask(ctx context.Context, opts CreateCloudSyncTaskOpts) (*CloudSyncTask, error) {
	params := taskOptsToParams(opts)

	result, err := s.client.Call(ctx, "cloudsync.create", params)
	if err != nil {
		return nil, err
	}

	var createResp struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(result, &createResp); err != nil {
		return nil, fmt.Errorf("parse create response: %w", err)
	}

	return s.GetTask(ctx, createResp.ID)
}

// GetTask returns a cloud sync task by ID, or nil if not found.
func (s *CloudSyncService) GetTask(ctx context.Context, id int64) (*CloudSyncTask, error) {
	filter := [][]any{{"id", "=", id}}
	result, err := s.client.Call(ctx, "cloudsync.query", filter)
	if err != nil {
		return nil, err
	}

	var tasks []CloudSyncTaskResponse
	if err := json.Unmarshal(result, &tasks); err != nil {
		return nil, fmt.Errorf("parse query response: %w", err)
	}

	if len(tasks) == 0 {
		return nil, nil
	}

	task := taskFromResponse(tasks[0])
	return &task, nil
}

// ListTasks returns all cloud sync tasks.
func (s *CloudSyncService) ListTasks(ctx context.Context) ([]CloudSyncTask, error) {
	result, err := s.client.Call(ctx, "cloudsync.query", nil)
	if err != nil {
		return nil, err
	}

	var responses []CloudSyncTaskResponse
	if err := json.Unmarshal(result, &responses); err != nil {
		return nil, fmt.Errorf("parse query response: %w", err)
	}

	tasks := make([]CloudSyncTask, len(responses))
	for i, resp := range responses {
		tasks[i] = taskFromResponse(resp)
	}
	return tasks, nil
}

// UpdateTask updates a cloud sync task and returns the full object.
func (s *CloudSyncService) UpdateTask(ctx context.Context, id int64, opts UpdateCloudSyncTaskOpts) (*CloudSyncTask, error) {
	params := taskOptsToParams(opts)

	_, err := s.client.Call(ctx, "cloudsync.update", []any{id, params})
	if err != nil {
		return nil, err
	}

	return s.GetTask(ctx, id)
}

// DeleteTask deletes a cloud sync task by ID.
func (s *CloudSyncService) DeleteTask(ctx context.Context, id int64) error {
	_, err := s.client.Call(ctx, "cloudsync.delete", id)
	return err
}

// Sync triggers a cloud sync task and waits for it to complete.
func (s *CloudSyncService) Sync(ctx context.Context, id int64) error {
	_, err := s.client.CallAndWait(ctx, "cloudsync.sync", id)
	return err
}

// credentialFromResponse converts a wire-format CloudSyncCredentialResponse to a user-facing CloudSyncCredential.
func credentialFromResponse(resp CloudSyncCredentialResponse) CloudSyncCredential {
	attrs := make(map[string]string)
	addIfNonEmpty(attrs, "access_key_id", resp.Provider.AccessKeyID)
	addIfNonEmpty(attrs, "secret_access_key", resp.Provider.SecretAccessKey)
	addIfNonEmpty(attrs, "endpoint", resp.Provider.Endpoint)
	addIfNonEmpty(attrs, "region", resp.Provider.Region)
	addIfNonEmpty(attrs, "account", resp.Provider.Account)
	addIfNonEmpty(attrs, "key", resp.Provider.Key)
	addIfNonEmpty(attrs, "service_account_credentials", resp.Provider.ServiceAccountCredentials)
	return CloudSyncCredential{
		ID:           resp.ID,
		Name:         resp.Name,
		ProviderType: resp.Provider.Type,
		Attributes:   attrs,
	}
}

// addIfNonEmpty adds a key-value pair to the map if the value is non-empty.
func addIfNonEmpty(m map[string]string, key, value string) {
	if value != "" {
		m[key] = value
	}
}

// credentialOptsToAttrsAny converts a map[string]string to map[string]any.
func credentialOptsToAttrsAny(attrs map[string]string) map[string]any {
	if attrs == nil {
		return nil
	}
	result := make(map[string]any, len(attrs))
	for k, v := range attrs {
		result[k] = v
	}
	return result
}

// taskOptsToParams converts CreateCloudSyncTaskOpts to API parameters.
func taskOptsToParams(opts CreateCloudSyncTaskOpts) map[string]any {
	params := map[string]any{
		"description":          opts.Description,
		"path":                 opts.Path,
		"credentials":         opts.CredentialID,
		"direction":           opts.Direction,
		"transfer_mode":       opts.TransferMode,
		"snapshot":            opts.Snapshot,
		"transfers":           opts.Transfers,
		"follow_symlinks":     opts.FollowSymlinks,
		"create_empty_src_dirs": opts.CreateEmptySrcDirs,
		"enabled":             opts.Enabled,
		"encryption":          opts.Encryption,
		"schedule": map[string]any{
			"minute": opts.Schedule.Minute,
			"hour":   opts.Schedule.Hour,
			"dom":    opts.Schedule.Dom,
			"month":  opts.Schedule.Month,
			"dow":    opts.Schedule.Dow,
		},
	}

	attrs := opts.Attributes
	if attrs == nil {
		attrs = map[string]any{}
	}
	attrs["fast_list"] = opts.FastList
	params["attributes"] = attrs

	if opts.Encryption {
		params["encryption_password"] = opts.EncryptionPassword
		params["encryption_salt"] = opts.EncryptionSalt
	}

	if len(opts.BWLimit) > 0 {
		params["bwlimit"] = opts.BWLimit
	}

	if len(opts.Exclude) > 0 {
		params["exclude"] = opts.Exclude
	}

	if len(opts.Include) > 0 {
		params["include"] = opts.Include
	}

	return params
}

// taskFromResponse converts a wire-format CloudSyncTaskResponse to a user-facing CloudSyncTask.
func taskFromResponse(resp CloudSyncTaskResponse) CloudSyncTask {
	task := CloudSyncTask{
		ID:                 resp.ID,
		Description:        resp.Description,
		Path:               resp.Path,
		CredentialID:       resp.Credentials.ID,
		Direction:          resp.Direction,
		TransferMode:       resp.TransferMode,
		Snapshot:           resp.Snapshot,
		Transfers:          resp.Transfers,
		BWLimit:            resp.BWLimit,
		FollowSymlinks:     resp.FollowSymlinks,
		CreateEmptySrcDirs: resp.CreateEmptySrcDirs,
		Enabled:            resp.Enabled,
		Encryption:         resp.Encryption,
		EncryptionPassword: resp.EncryptionPassword,
		EncryptionSalt:     resp.EncryptionSalt,
		Schedule: Schedule{
			Minute: resp.Schedule.Minute,
			Hour:   resp.Schedule.Hour,
			Dom:    resp.Schedule.Dom,
			Month:  resp.Schedule.Month,
			Dow:    resp.Schedule.Dow,
		},
		Exclude: resp.Exclude,
		Include: resp.Include,
	}

	// Handle attributes - can be false in API response, so ignore unmarshal errors
	if len(resp.Attributes) > 0 {
		var attrs map[string]any
		if err := json.Unmarshal(resp.Attributes, &attrs); err == nil {
			if fl, ok := attrs["fast_list"].(bool); ok {
				task.FastList = fl
				delete(attrs, "fast_list")
			}
			task.Attributes = attrs
		}
	}

	return task
}

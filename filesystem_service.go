package truenas

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// StatResult is the user-facing representation of a filesystem stat.
// Mode contains only permission bits (masked with 0o777).
type StatResult struct {
	Mode int64
	UID  int64
	GID  int64
}

// SetPermOpts contains options for setting filesystem permissions.
type SetPermOpts struct {
	Path      string
	UID       *int64
	GID       *int64
	Mode      string // Octal string e.g. "755", empty omits
	Recursive bool
	StripACL  bool
	Traverse  bool
}

// FilesystemService provides typed methods for the filesystem.* API namespace.
type FilesystemService struct {
	client  FileCaller
	version Version
}

// NewFilesystemService creates a new FilesystemService.
func NewFilesystemService(c FileCaller, v Version) *FilesystemService {
	return &FilesystemService{client: c, version: v}
}

// Client returns the underlying FileCaller.
func (s *FilesystemService) Client() FileCaller {
	return s.client
}

// WriteFile writes content to a file on the remote system via filesystem.file_receive.
func (s *FilesystemService) WriteFile(ctx context.Context, path string, params WriteFileParams) error {
	b64Content := base64.StdEncoding.EncodeToString(params.Content)

	uid := -1
	if params.UID != nil {
		uid = *params.UID
	}
	gid := -1
	if params.GID != nil {
		gid = *params.GID
	}

	apiParams := []any{
		path,
		b64Content,
		map[string]any{
			"mode": int(params.Mode),
			"uid":  uid,
			"gid":  gid,
		},
	}

	_, err := s.client.Call(ctx, "filesystem.file_receive", apiParams)
	if err != nil {
		return fmt.Errorf("write file %q: %w", path, err)
	}
	return nil
}

// Stat returns filesystem stat information for the given path.
// Mode is masked with 0o777 to strip file type bits.
func (s *FilesystemService) Stat(ctx context.Context, path string) (*StatResult, error) {
	result, err := s.client.Call(ctx, "filesystem.stat", path)
	if err != nil {
		return nil, err
	}

	var resp StatResponse
	if err := json.Unmarshal(result, &resp); err != nil {
		return nil, fmt.Errorf("parse stat response: %w", err)
	}

	return &StatResult{
		Mode: resp.Mode & 0o777,
		UID:  resp.UID,
		GID:  resp.GID,
	}, nil
}

// SetPermissions sets filesystem permissions via the filesystem.setperm API.
// This is a job-based operation that blocks until complete.
func (s *FilesystemService) SetPermissions(ctx context.Context, opts SetPermOpts) error {
	params := buildSetPermParams(opts)
	_, err := s.client.CallAndWait(ctx, "filesystem.setperm", params)
	return err
}

// buildSetPermParams converts SetPermOpts to API parameters.
// Only includes fields that are set (non-nil/non-empty/non-false).
func buildSetPermParams(opts SetPermOpts) map[string]any {
	params := map[string]any{
		"path": opts.Path,
	}

	if opts.UID != nil {
		params["uid"] = *opts.UID
	}
	if opts.GID != nil {
		params["gid"] = *opts.GID
	}
	if opts.Mode != "" {
		params["mode"] = opts.Mode
	}

	options := map[string]any{}
	if opts.Recursive {
		options["recursive"] = true
	}
	if opts.StripACL {
		options["stripacl"] = true
	}
	if opts.Traverse {
		options["traverse"] = true
	}
	if len(options) > 0 {
		params["options"] = options
	}

	return params
}

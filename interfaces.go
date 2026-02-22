package truenas

import (
	"context"
	"encoding/json"
	"io/fs"
)

// Caller is the base interface for making API calls.
// It is satisfied by client.Client and client.MockClient.
type Caller interface {
	Call(ctx context.Context, method string, params any) (json.RawMessage, error)
}

// AsyncCaller adds support for long-running job-based operations.
type AsyncCaller interface {
	Caller
	CallAndWait(ctx context.Context, method string, params any) (json.RawMessage, error)
}

// FileCaller adds file system operations over SSH.
// Embeds AsyncCaller because filesystem.setperm is job-based.
type FileCaller interface {
	AsyncCaller
	WriteFile(ctx context.Context, path string, params WriteFileParams) error
	ReadFile(ctx context.Context, path string) ([]byte, error)
	DeleteFile(ctx context.Context, path string) error
	RemoveDir(ctx context.Context, path string) error
	RemoveAll(ctx context.Context, path string) error
	FileExists(ctx context.Context, path string) (bool, error)
	Chown(ctx context.Context, path string, uid, gid int) error
	ChmodRecursive(ctx context.Context, path string, mode fs.FileMode) error
	MkdirAll(ctx context.Context, path string, mode fs.FileMode) error
}

// WriteFileParams contains parameters for writing a file.
type WriteFileParams struct {
	Content []byte      // Required - file data to write
	Mode    fs.FileMode // Default: 0644
	UID     *int        // nil = unchanged, pointer allows explicit 0 (root)
	GID     *int        // nil = unchanged, pointer allows explicit 0 (root)
}

// DefaultWriteFileParams returns params with sensible defaults.
// Mode defaults to 0644. UID/GID are nil (unchanged).
func DefaultWriteFileParams(content []byte) WriteFileParams {
	return WriteFileParams{
		Content: content,
		Mode:    0644,
	}
}

// IntPtr returns a pointer to an int. Helper for setting UID/GID.
func IntPtr(i int) *int { return &i }

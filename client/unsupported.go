package client

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"

	truenas "github.com/deevus/truenas-go"
)

// ErrUnsupportedOperation is returned when an operation is not supported by the
// client implementation.
var ErrUnsupportedOperation = errors.New("operation not supported")

// UnsupportedClient implements Client and returns ErrUnsupportedOperation for
// operations that require SSH. It is used as the default fallback when no SSH
// client is configured on WebSocketConfig.
type UnsupportedClient struct{}

var _ Client = (*UnsupportedClient)(nil)

func (u *UnsupportedClient) Connect(ctx context.Context) error { return nil }
func (u *UnsupportedClient) Version() truenas.Version          { return truenas.Version{} }
func (u *UnsupportedClient) Close() error                      { return nil }

func (u *UnsupportedClient) Call(ctx context.Context, method string, params any) (json.RawMessage, error) {
	return nil, ErrUnsupportedOperation
}

func (u *UnsupportedClient) CallAndWait(ctx context.Context, method string, params any) (json.RawMessage, error) {
	return nil, ErrUnsupportedOperation
}

func (u *UnsupportedClient) WriteFile(ctx context.Context, path string, params truenas.WriteFileParams) error {
	return ErrUnsupportedOperation
}

func (u *UnsupportedClient) ReadFile(ctx context.Context, path string) ([]byte, error) {
	return nil, ErrUnsupportedOperation
}

func (u *UnsupportedClient) DeleteFile(ctx context.Context, path string) error {
	return ErrUnsupportedOperation
}

func (u *UnsupportedClient) RemoveDir(ctx context.Context, path string) error {
	return ErrUnsupportedOperation
}

func (u *UnsupportedClient) RemoveAll(ctx context.Context, path string) error {
	return ErrUnsupportedOperation
}

func (u *UnsupportedClient) FileExists(ctx context.Context, path string) (bool, error) {
	return false, ErrUnsupportedOperation
}

func (u *UnsupportedClient) Chown(ctx context.Context, path string, uid, gid int) error {
	return ErrUnsupportedOperation
}

func (u *UnsupportedClient) ChmodRecursive(ctx context.Context, path string, mode fs.FileMode) error {
	return ErrUnsupportedOperation
}

func (u *UnsupportedClient) MkdirAll(ctx context.Context, path string, mode fs.FileMode) error {
	return ErrUnsupportedOperation
}

package truenas

import (
	"context"
	"encoding/json"
	"io/fs"
)

// mockCaller is a test double for the Caller interface.
type mockCaller struct {
	callFunc func(ctx context.Context, method string, params any) (json.RawMessage, error)
	calls    []mockCall
}

type mockCall struct {
	Method string
	Params any
}

func (m *mockCaller) Call(ctx context.Context, method string, params any) (json.RawMessage, error) {
	m.calls = append(m.calls, mockCall{Method: method, Params: params})
	if m.callFunc != nil {
		return m.callFunc(ctx, method, params)
	}
	return nil, nil
}

// mockAsyncCaller is a test double for the AsyncCaller interface.
type mockAsyncCaller struct {
	mockCaller
	callAndWaitFunc func(ctx context.Context, method string, params any) (json.RawMessage, error)
}

func (m *mockAsyncCaller) CallAndWait(ctx context.Context, method string, params any) (json.RawMessage, error) {
	m.calls = append(m.calls, mockCall{Method: method, Params: params})
	if m.callAndWaitFunc != nil {
		return m.callAndWaitFunc(ctx, method, params)
	}
	return nil, nil
}

// mockFileCaller is a test double for the FileCaller interface.
type mockFileCaller struct {
	mockAsyncCaller
	writeFileFunc      func(ctx context.Context, path string, params WriteFileParams) error
	readFileFunc       func(ctx context.Context, path string) ([]byte, error)
	deleteFileFunc     func(ctx context.Context, path string) error
	removeDirFunc      func(ctx context.Context, path string) error
	removeAllFunc      func(ctx context.Context, path string) error
	fileExistsFunc     func(ctx context.Context, path string) (bool, error)
	chownFunc          func(ctx context.Context, path string, uid, gid int) error
	chmodRecursiveFunc func(ctx context.Context, path string, mode fs.FileMode) error
	mkdirAllFunc       func(ctx context.Context, path string, mode fs.FileMode) error
}

func (m *mockFileCaller) WriteFile(ctx context.Context, path string, params WriteFileParams) error {
	m.calls = append(m.calls, mockCall{Method: "WriteFile", Params: path})
	if m.writeFileFunc != nil {
		return m.writeFileFunc(ctx, path, params)
	}
	return nil
}

func (m *mockFileCaller) ReadFile(ctx context.Context, path string) ([]byte, error) {
	m.calls = append(m.calls, mockCall{Method: "ReadFile", Params: path})
	if m.readFileFunc != nil {
		return m.readFileFunc(ctx, path)
	}
	return nil, nil
}

func (m *mockFileCaller) DeleteFile(ctx context.Context, path string) error {
	m.calls = append(m.calls, mockCall{Method: "DeleteFile", Params: path})
	if m.deleteFileFunc != nil {
		return m.deleteFileFunc(ctx, path)
	}
	return nil
}

func (m *mockFileCaller) RemoveDir(ctx context.Context, path string) error {
	m.calls = append(m.calls, mockCall{Method: "RemoveDir", Params: path})
	if m.removeDirFunc != nil {
		return m.removeDirFunc(ctx, path)
	}
	return nil
}

func (m *mockFileCaller) RemoveAll(ctx context.Context, path string) error {
	m.calls = append(m.calls, mockCall{Method: "RemoveAll", Params: path})
	if m.removeAllFunc != nil {
		return m.removeAllFunc(ctx, path)
	}
	return nil
}

func (m *mockFileCaller) FileExists(ctx context.Context, path string) (bool, error) {
	m.calls = append(m.calls, mockCall{Method: "FileExists", Params: path})
	if m.fileExistsFunc != nil {
		return m.fileExistsFunc(ctx, path)
	}
	return false, nil
}

func (m *mockFileCaller) Chown(ctx context.Context, path string, uid, gid int) error {
	m.calls = append(m.calls, mockCall{Method: "Chown", Params: path})
	if m.chownFunc != nil {
		return m.chownFunc(ctx, path, uid, gid)
	}
	return nil
}

func (m *mockFileCaller) ChmodRecursive(ctx context.Context, path string, mode fs.FileMode) error {
	m.calls = append(m.calls, mockCall{Method: "ChmodRecursive", Params: path})
	if m.chmodRecursiveFunc != nil {
		return m.chmodRecursiveFunc(ctx, path, mode)
	}
	return nil
}

func (m *mockFileCaller) MkdirAll(ctx context.Context, path string, mode fs.FileMode) error {
	m.calls = append(m.calls, mockCall{Method: "MkdirAll", Params: path})
	if m.mkdirAllFunc != nil {
		return m.mkdirAllFunc(ctx, path, mode)
	}
	return nil
}

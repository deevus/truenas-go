package truenas

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func sampleStatJSON() json.RawMessage {
	return json.RawMessage(`{"mode": 16877, "uid": 1000, "gid": 1000}`)
}

func sampleFileStatJSON() json.RawMessage {
	return json.RawMessage(`{"mode": 33188, "uid": 0, "gid": 0}`)
}

func TestFilesystemService_WriteFile(t *testing.T) {
	mock := &mockFileCaller{
		mockAsyncCaller: mockAsyncCaller{
			mockCaller: mockCaller{
				callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
					if method != "filesystem.file_receive" {
						t.Errorf("expected method filesystem.file_receive, got %s", method)
					}
					args := params.([]any)
					if args[0] != "/mnt/pool/test.txt" {
						t.Errorf("expected path /mnt/pool/test.txt, got %v", args[0])
					}
					// Content should be base64-encoded "hello"
					if args[1] != "aGVsbG8=" {
						t.Errorf("expected base64 'aGVsbG8=', got %v", args[1])
					}
					opts := args[2].(map[string]any)
					if opts["mode"] != int(0o644) {
						t.Errorf("expected mode %d, got %v", int(0o644), opts["mode"])
					}
					if opts["uid"] != -1 {
						t.Errorf("expected uid -1 (unset), got %v", opts["uid"])
					}
					if opts["gid"] != -1 {
						t.Errorf("expected gid -1 (unset), got %v", opts["gid"])
					}
					return nil, nil
				},
			},
		},
	}

	svc := NewFilesystemService(mock, Version{})
	err := svc.WriteFile(context.Background(), "/mnt/pool/test.txt", WriteFileParams{
		Content: []byte("hello"),
		Mode:    0o644,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFilesystemService_WriteFile_WithUID(t *testing.T) {
	mock := &mockFileCaller{
		mockAsyncCaller: mockAsyncCaller{
			mockCaller: mockCaller{
				callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
					args := params.([]any)
					opts := args[2].(map[string]any)
					if opts["uid"] != 1000 {
						t.Errorf("expected uid 1000, got %v", opts["uid"])
					}
					if opts["gid"] != 1000 {
						t.Errorf("expected gid 1000, got %v", opts["gid"])
					}
					return nil, nil
				},
			},
		},
	}

	svc := NewFilesystemService(mock, Version{})
	uid, gid := 1000, 1000
	err := svc.WriteFile(context.Background(), "/mnt/pool/test.txt", WriteFileParams{
		Content: []byte("hello"),
		Mode:    0o644,
		UID:     &uid,
		GID:     &gid,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFilesystemService_WriteFile_Error(t *testing.T) {
	mock := &mockFileCaller{
		mockAsyncCaller: mockAsyncCaller{
			mockCaller: mockCaller{
				callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
					return nil, errors.New("permission denied")
				},
			},
		},
	}

	svc := NewFilesystemService(mock, Version{})
	err := svc.WriteFile(context.Background(), "/mnt/pool/test.txt", WriteFileParams{
		Content: []byte("hello"),
		Mode:    0o644,
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFilesystemService_Stat(t *testing.T) {
	mock := &mockFileCaller{
		mockAsyncCaller: mockAsyncCaller{
			mockCaller: mockCaller{
				callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
					return sampleStatJSON(), nil
				},
			},
		},
	}

	svc := NewFilesystemService(mock, Version{})
	result, err := svc.Stat(context.Background(), "/mnt/pool/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Verify method and params
	if len(mock.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(mock.calls))
	}
	if mock.calls[0].Method != "filesystem.stat" {
		t.Errorf("expected method filesystem.stat, got %s", mock.calls[0].Method)
	}
	if mock.calls[0].Params != "/mnt/pool/test" {
		t.Errorf("expected params /mnt/pool/test, got %v", mock.calls[0].Params)
	}

	// Directory mode 16877 (0o40755) → masked to 0o755 = 493
	if result.Mode != 493 {
		t.Errorf("expected Mode 493 (0o755), got %d", result.Mode)
	}
	if result.UID != 1000 {
		t.Errorf("expected UID 1000, got %d", result.UID)
	}
	if result.GID != 1000 {
		t.Errorf("expected GID 1000, got %d", result.GID)
	}
}

func TestFilesystemService_Stat_File(t *testing.T) {
	mock := &mockFileCaller{
		mockAsyncCaller: mockAsyncCaller{
			mockCaller: mockCaller{
				callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
					return sampleFileStatJSON(), nil
				},
			},
		},
	}

	svc := NewFilesystemService(mock, Version{})
	result, err := svc.Stat(context.Background(), "/mnt/pool/file.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// File mode 33188 (0o100644) → masked to 0o644 = 420
	if result.Mode != 420 {
		t.Errorf("expected Mode 420 (0o644), got %d", result.Mode)
	}
	if result.UID != 0 {
		t.Errorf("expected UID 0, got %d", result.UID)
	}
	if result.GID != 0 {
		t.Errorf("expected GID 0, got %d", result.GID)
	}
}

func TestFilesystemService_Stat_Error(t *testing.T) {
	mock := &mockFileCaller{
		mockAsyncCaller: mockAsyncCaller{
			mockCaller: mockCaller{
				callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
					return nil, errors.New("connection refused")
				},
			},
		},
	}

	svc := NewFilesystemService(mock, Version{})
	result, err := svc.Stat(context.Background(), "/mnt/pool/test")
	if err == nil {
		t.Fatal("expected error")
	}
	if result != nil {
		t.Error("expected nil result on error")
	}
	if err.Error() != "connection refused" {
		t.Errorf("expected 'connection refused', got %q", err.Error())
	}
}

func TestFilesystemService_Stat_ParseError(t *testing.T) {
	mock := &mockFileCaller{
		mockAsyncCaller: mockAsyncCaller{
			mockCaller: mockCaller{
				callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
					return json.RawMessage(`not json`), nil
				},
			},
		},
	}

	svc := NewFilesystemService(mock, Version{})
	_, err := svc.Stat(context.Background(), "/mnt/pool/test")
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestFilesystemService_SetPermissions(t *testing.T) {
	uid := int64(1000)
	gid := int64(1000)
	mock := &mockFileCaller{
		mockAsyncCaller: mockAsyncCaller{
			callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, nil
			},
		},
	}

	svc := NewFilesystemService(mock, Version{})
	err := svc.SetPermissions(context.Background(), SetPermOpts{
		Path:      "/mnt/pool/test",
		UID:       &uid,
		GID:       &gid,
		Mode:      "755",
		Recursive: true,
		StripACL:  true,
		Traverse:  true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify method and params
	if len(mock.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(mock.calls))
	}
	if mock.calls[0].Method != "filesystem.setperm" {
		t.Errorf("expected method filesystem.setperm, got %s", mock.calls[0].Method)
	}

	p := mock.calls[0].Params.(map[string]any)
	if p["path"] != "/mnt/pool/test" {
		t.Errorf("expected path /mnt/pool/test, got %v", p["path"])
	}
	if p["uid"] != int64(1000) {
		t.Errorf("expected uid 1000, got %v", p["uid"])
	}
	if p["gid"] != int64(1000) {
		t.Errorf("expected gid 1000, got %v", p["gid"])
	}
	if p["mode"] != "755" {
		t.Errorf("expected mode 755, got %v", p["mode"])
	}

	opts, ok := p["options"].(map[string]any)
	if !ok {
		t.Fatal("expected options sub-map")
	}
	if opts["recursive"] != true {
		t.Errorf("expected recursive=true, got %v", opts["recursive"])
	}
	if opts["stripacl"] != true {
		t.Errorf("expected stripacl=true, got %v", opts["stripacl"])
	}
	if opts["traverse"] != true {
		t.Errorf("expected traverse=true, got %v", opts["traverse"])
	}
}

func TestFilesystemService_SetPermissions_MinimalOpts(t *testing.T) {
	mock := &mockFileCaller{
		mockAsyncCaller: mockAsyncCaller{
			callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, nil
			},
		},
	}

	svc := NewFilesystemService(mock, Version{})
	err := svc.SetPermissions(context.Background(), SetPermOpts{
		Path: "/mnt/pool/test",
		Mode: "644",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p := mock.calls[0].Params.(map[string]any)
	if p["path"] != "/mnt/pool/test" {
		t.Errorf("expected path /mnt/pool/test, got %v", p["path"])
	}
	if p["mode"] != "644" {
		t.Errorf("expected mode 644, got %v", p["mode"])
	}
	// uid and gid should not be present
	if _, ok := p["uid"]; ok {
		t.Error("expected no uid key")
	}
	if _, ok := p["gid"]; ok {
		t.Error("expected no gid key")
	}
	// options should not be present
	if _, ok := p["options"]; ok {
		t.Error("expected no options key")
	}
}

func TestFilesystemService_SetPermissions_UIDOnly(t *testing.T) {
	uid := int64(0)
	mock := &mockFileCaller{
		mockAsyncCaller: mockAsyncCaller{
			callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, nil
			},
		},
	}

	svc := NewFilesystemService(mock, Version{})
	err := svc.SetPermissions(context.Background(), SetPermOpts{
		Path: "/mnt/pool/test",
		UID:  &uid,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p := mock.calls[0].Params.(map[string]any)
	if p["path"] != "/mnt/pool/test" {
		t.Errorf("expected path /mnt/pool/test, got %v", p["path"])
	}
	if p["uid"] != int64(0) {
		t.Errorf("expected uid 0, got %v", p["uid"])
	}
	// mode should not be present
	if _, ok := p["mode"]; ok {
		t.Error("expected no mode key")
	}
	// gid should not be present
	if _, ok := p["gid"]; ok {
		t.Error("expected no gid key")
	}
}

func TestFilesystemService_SetPermissions_Error(t *testing.T) {
	mock := &mockFileCaller{
		mockAsyncCaller: mockAsyncCaller{
			callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, errors.New("permission denied")
			},
		},
	}

	svc := NewFilesystemService(mock, Version{})
	err := svc.SetPermissions(context.Background(), SetPermOpts{
		Path: "/mnt/pool/test",
		Mode: "755",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "permission denied" {
		t.Errorf("expected 'permission denied', got %q", err.Error())
	}
}

func TestBuildSetPermParams_AllFields(t *testing.T) {
	uid := int64(1000)
	gid := int64(1000)
	params := buildSetPermParams(SetPermOpts{
		Path:      "/mnt/pool/test",
		UID:       &uid,
		GID:       &gid,
		Mode:      "755",
		Recursive: true,
		StripACL:  true,
		Traverse:  true,
	})

	if params["path"] != "/mnt/pool/test" {
		t.Errorf("expected path /mnt/pool/test, got %v", params["path"])
	}
	if params["uid"] != int64(1000) {
		t.Errorf("expected uid 1000, got %v", params["uid"])
	}
	if params["gid"] != int64(1000) {
		t.Errorf("expected gid 1000, got %v", params["gid"])
	}
	if params["mode"] != "755" {
		t.Errorf("expected mode 755, got %v", params["mode"])
	}

	opts, ok := params["options"].(map[string]any)
	if !ok {
		t.Fatal("expected options sub-map")
	}
	if opts["recursive"] != true {
		t.Errorf("expected recursive=true, got %v", opts["recursive"])
	}
	if opts["stripacl"] != true {
		t.Errorf("expected stripacl=true, got %v", opts["stripacl"])
	}
	if opts["traverse"] != true {
		t.Errorf("expected traverse=true, got %v", opts["traverse"])
	}
}

func TestBuildSetPermParams_PathOnly(t *testing.T) {
	params := buildSetPermParams(SetPermOpts{
		Path: "/mnt/pool/test",
	})

	if params["path"] != "/mnt/pool/test" {
		t.Errorf("expected path /mnt/pool/test, got %v", params["path"])
	}
	if _, ok := params["uid"]; ok {
		t.Error("expected no uid key")
	}
	if _, ok := params["gid"]; ok {
		t.Error("expected no gid key")
	}
	if _, ok := params["mode"]; ok {
		t.Error("expected no mode key")
	}
	if _, ok := params["options"]; ok {
		t.Error("expected no options key")
	}
}

func TestBuildSetPermParams_NoOptions(t *testing.T) {
	uid := int64(500)
	gid := int64(500)
	params := buildSetPermParams(SetPermOpts{
		Path: "/mnt/pool/data",
		UID:  &uid,
		GID:  &gid,
		Mode: "700",
	})

	if params["path"] != "/mnt/pool/data" {
		t.Errorf("expected path /mnt/pool/data, got %v", params["path"])
	}
	if params["uid"] != int64(500) {
		t.Errorf("expected uid 500, got %v", params["uid"])
	}
	if params["gid"] != int64(500) {
		t.Errorf("expected gid 500, got %v", params["gid"])
	}
	if params["mode"] != "700" {
		t.Errorf("expected mode 700, got %v", params["mode"])
	}
	// No recursive/stripacl/traverse → no options key
	if _, ok := params["options"]; ok {
		t.Error("expected no options key when no boolean options set")
	}
}

func TestNewFilesystemService(t *testing.T) {
	mock := &mockFileCaller{}
	v := Version{Major: 24, Minor: 10, Patch: 2, Build: 4}
	svc := NewFilesystemService(mock, v)

	if svc.client != mock {
		t.Error("expected client to be the provided mock")
	}
	if svc.version != v {
		t.Errorf("expected version %v, got %v", v, svc.version)
	}
}

func TestFilesystemService_Client(t *testing.T) {
	mock := &mockFileCaller{}
	svc := NewFilesystemService(mock, Version{})

	if svc.Client() != mock {
		t.Error("expected Client() to return the same FileCaller")
	}
}

package truenas

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

// sampleCredentialV25JSON returns a V25 format credential JSON response.
func sampleCredentialV25JSON() json.RawMessage {
	return json.RawMessage(`[{
		"id": 1,
		"name": "My S3 Cred",
		"provider": {
			"type": "S3",
			"access_key_id": "AKIATEST",
			"secret_access_key": "secret123",
			"endpoint": "s3.example.com",
			"region": "us-east-1"
		}
	}]`)
}

// sampleCredentialV24JSON returns a V24 format credential JSON response.
func sampleCredentialV24JSON() json.RawMessage {
	return json.RawMessage(`[{
		"id": 1,
		"name": "My S3 Cred",
		"provider": "S3",
		"attributes": {
			"access_key_id": "AKIATEST",
			"secret_access_key": "secret123",
			"endpoint": "s3.example.com",
			"region": "us-east-1"
		}
	}]`)
}

// sampleTaskJSON returns a cloud sync task JSON response.
func sampleTaskJSON() json.RawMessage {
	return json.RawMessage(`[{
		"id": 1,
		"description": "Backup to S3",
		"path": "/mnt/tank/data",
		"credentials": {"id": 5, "name": "My S3 Cred"},
		"attributes": {"bucket": "my-bucket", "folder": "/backups"},
		"schedule": {"minute": "0", "hour": "3", "dom": "*", "month": "*", "dow": "*"},
		"direction": "PUSH",
		"transfer_mode": "SYNC",
		"encryption": false,
		"snapshot": true,
		"transfers": 4,
		"bwlimit": [{"time": "08:00", "bandwidth": 1048576}],
		"exclude": ["*.tmp"],
		"include": [],
		"follow_symlinks": false,
		"create_empty_src_dirs": true,
		"enabled": true
	}]`)
}

// sampleTaskFalseAttrsJSON returns a task JSON response where attributes is false.
func sampleTaskFalseAttrsJSON() json.RawMessage {
	return json.RawMessage(`[{
		"id": 2,
		"description": "Task with false attrs",
		"path": "/mnt/pool/data",
		"credentials": {"id": 3, "name": "Cred"},
		"attributes": false,
		"schedule": {"minute": "30", "hour": "*/2", "dom": "*", "month": "*", "dow": "*"},
		"direction": "PULL",
		"transfer_mode": "COPY",
		"encryption": true,
		"encryption_password": "mypass",
		"encryption_salt": "mysalt",
		"snapshot": false,
		"transfers": 2,
		"bwlimit": [],
		"exclude": [],
		"include": ["*.dat"],
		"follow_symlinks": true,
		"create_empty_src_dirs": false,
		"enabled": false
	}]`)
}

// --- Credential tests ---

func TestCloudSyncService_CreateCredential_V25(t *testing.T) {
	callCount := 0
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				callCount++
				if callCount == 1 {
					if method != "cloudsync.credentials.create" {
						t.Errorf("expected method cloudsync.credentials.create, got %s", method)
					}
					// Verify V25 format: provider is object
					p := params.(map[string]any)
					provider, ok := p["provider"].(map[string]any)
					if !ok {
						t.Fatal("expected provider to be map[string]any for V25")
					}
					if provider["type"] != "S3" {
						t.Errorf("expected provider.type S3, got %v", provider["type"])
					}
					if provider["access_key_id"] != "AKIATEST" {
						t.Errorf("expected access_key_id AKIATEST, got %v", provider["access_key_id"])
					}
					return json.RawMessage(`{"id": 1}`), nil
				}
				// Re-read via GetCredential
				return sampleCredentialV25JSON(), nil
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 25, Minor: 4})
	cred, err := svc.CreateCredential(context.Background(), CreateCredentialOpts{
		Name:         "My S3 Cred",
		ProviderType: "S3",
		Attributes: map[string]string{
			"access_key_id":     "AKIATEST",
			"secret_access_key": "secret123",
			"endpoint":          "s3.example.com",
			"region":            "us-east-1",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cred == nil {
		t.Fatal("expected non-nil credential")
	}
	if cred.ID != 1 {
		t.Errorf("expected ID 1, got %d", cred.ID)
	}
	if cred.Name != "My S3 Cred" {
		t.Errorf("expected name 'My S3 Cred', got %q", cred.Name)
	}
	if cred.ProviderType != "S3" {
		t.Errorf("expected provider type S3, got %s", cred.ProviderType)
	}
	if cred.Attributes["access_key_id"] != "AKIATEST" {
		t.Errorf("expected access_key_id AKIATEST, got %s", cred.Attributes["access_key_id"])
	}
	if cred.Attributes["region"] != "us-east-1" {
		t.Errorf("expected region us-east-1, got %s", cred.Attributes["region"])
	}
}

func TestCloudSyncService_CreateCredential_V24(t *testing.T) {
	callCount := 0
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				callCount++
				if callCount == 1 {
					if method != "cloudsync.credentials.create" {
						t.Errorf("expected method cloudsync.credentials.create, got %s", method)
					}
					// Verify V24 format: provider is string, attributes separate
					p := params.(map[string]any)
					provider, ok := p["provider"].(string)
					if !ok {
						t.Fatal("expected provider to be string for V24")
					}
					if provider != "S3" {
						t.Errorf("expected provider S3, got %s", provider)
					}
					attrs, ok := p["attributes"].(map[string]any)
					if !ok {
						t.Fatal("expected attributes to be map[string]any for V24")
					}
					if attrs["access_key_id"] != "AKIATEST" {
						t.Errorf("expected attributes.access_key_id AKIATEST, got %v", attrs["access_key_id"])
					}
					return json.RawMessage(`{"id": 1}`), nil
				}
				return sampleCredentialV24JSON(), nil
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 24, Minor: 10})
	cred, err := svc.CreateCredential(context.Background(), CreateCredentialOpts{
		Name:         "My S3 Cred",
		ProviderType: "S3",
		Attributes: map[string]string{
			"access_key_id":     "AKIATEST",
			"secret_access_key": "secret123",
			"endpoint":          "s3.example.com",
			"region":            "us-east-1",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cred == nil {
		t.Fatal("expected non-nil credential")
	}
	if cred.ID != 1 {
		t.Errorf("expected ID 1, got %d", cred.ID)
	}
	if cred.ProviderType != "S3" {
		t.Errorf("expected provider type S3, got %s", cred.ProviderType)
	}
}

func TestCloudSyncService_CreateCredential_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, errors.New("connection refused")
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 25, Minor: 4})
	cred, err := svc.CreateCredential(context.Background(), CreateCredentialOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
	if cred != nil {
		t.Error("expected nil credential on error")
	}
	if err.Error() != "connection refused" {
		t.Errorf("expected 'connection refused', got %q", err.Error())
	}
}

func TestCloudSyncService_CreateCredential_ParseError(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return json.RawMessage(`not json`), nil
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 25, Minor: 4})
	_, err := svc.CreateCredential(context.Background(), CreateCredentialOpts{})
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestCloudSyncService_GetCredential_V25(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				if method != "cloudsync.credentials.query" {
					t.Errorf("expected method cloudsync.credentials.query, got %s", method)
				}
				return sampleCredentialV25JSON(), nil
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 25, Minor: 4})
	cred, err := svc.GetCredential(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cred == nil {
		t.Fatal("expected non-nil credential")
	}
	if cred.ID != 1 {
		t.Errorf("expected ID 1, got %d", cred.ID)
	}
	if cred.Name != "My S3 Cred" {
		t.Errorf("expected name 'My S3 Cred', got %q", cred.Name)
	}
	if cred.ProviderType != "S3" {
		t.Errorf("expected provider type S3, got %s", cred.ProviderType)
	}
	if cred.Attributes["secret_access_key"] != "secret123" {
		t.Errorf("expected secret_access_key secret123, got %s", cred.Attributes["secret_access_key"])
	}
	if cred.Attributes["endpoint"] != "s3.example.com" {
		t.Errorf("expected endpoint s3.example.com, got %s", cred.Attributes["endpoint"])
	}
}

func TestCloudSyncService_GetCredential_V24(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return sampleCredentialV24JSON(), nil
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 24, Minor: 10})
	cred, err := svc.GetCredential(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cred == nil {
		t.Fatal("expected non-nil credential")
	}
	if cred.ID != 1 {
		t.Errorf("expected ID 1, got %d", cred.ID)
	}
	if cred.ProviderType != "S3" {
		t.Errorf("expected provider type S3, got %s", cred.ProviderType)
	}
	if cred.Attributes["access_key_id"] != "AKIATEST" {
		t.Errorf("expected access_key_id AKIATEST, got %s", cred.Attributes["access_key_id"])
	}
}

func TestCloudSyncService_GetCredential_NotFound(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return json.RawMessage(`[]`), nil
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 25, Minor: 4})
	cred, err := svc.GetCredential(context.Background(), 999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cred != nil {
		t.Error("expected nil credential for not found")
	}
}

func TestCloudSyncService_GetCredential_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, errors.New("timeout")
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 25, Minor: 4})
	_, err := svc.GetCredential(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCloudSyncService_GetCredential_ParseError(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return json.RawMessage(`not json`), nil
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 25, Minor: 4})
	_, err := svc.GetCredential(context.Background(), 1)
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestCloudSyncService_ListCredentials_V25(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				if method != "cloudsync.credentials.query" {
					t.Errorf("expected method cloudsync.credentials.query, got %s", method)
				}
				if params != nil {
					t.Error("expected nil params for ListCredentials")
				}
				return json.RawMessage(`[
					{"id": 1, "name": "S3 Cred", "provider": {"type": "S3", "access_key_id": "key1", "secret_access_key": "sec1"}},
					{"id": 2, "name": "B2 Cred", "provider": {"type": "B2", "account": "acct", "key": "b2key"}}
				]`), nil
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 25, Minor: 4})
	creds, err := svc.ListCredentials(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(creds) != 2 {
		t.Fatalf("expected 2 credentials, got %d", len(creds))
	}
	if creds[0].ProviderType != "S3" {
		t.Errorf("expected first cred provider type S3, got %s", creds[0].ProviderType)
	}
	if creds[1].ProviderType != "B2" {
		t.Errorf("expected second cred provider type B2, got %s", creds[1].ProviderType)
	}
	if creds[1].Attributes["account"] != "acct" {
		t.Errorf("expected second cred account 'acct', got %s", creds[1].Attributes["account"])
	}
}

func TestCloudSyncService_ListCredentials_V24(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return json.RawMessage(`[
					{"id": 1, "name": "S3 Cred", "provider": "S3", "attributes": {"access_key_id": "key1"}},
					{"id": 2, "name": "B2 Cred", "provider": "B2", "attributes": {"account": "acct"}}
				]`), nil
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 24, Minor: 10})
	creds, err := svc.ListCredentials(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(creds) != 2 {
		t.Fatalf("expected 2 credentials, got %d", len(creds))
	}
	if creds[0].Attributes["access_key_id"] != "key1" {
		t.Errorf("expected first cred access_key_id 'key1', got %s", creds[0].Attributes["access_key_id"])
	}
}

func TestCloudSyncService_ListCredentials_Empty(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return json.RawMessage(`[]`), nil
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 25, Minor: 4})
	creds, err := svc.ListCredentials(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(creds) != 0 {
		t.Errorf("expected 0 credentials, got %d", len(creds))
	}
}

func TestCloudSyncService_ListCredentials_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, errors.New("network error")
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 25, Minor: 4})
	_, err := svc.ListCredentials(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCloudSyncService_ListCredentials_ParseError(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return json.RawMessage(`not json`), nil
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 25, Minor: 4})
	_, err := svc.ListCredentials(context.Background())
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestCloudSyncService_UpdateCredential_V25(t *testing.T) {
	callCount := 0
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				callCount++
				if callCount == 1 {
					if method != "cloudsync.credentials.update" {
						t.Errorf("expected method cloudsync.credentials.update, got %s", method)
					}
					// Verify params shape: []any{id, map}
					slice, ok := params.([]any)
					if !ok {
						t.Fatal("expected []any params for update")
					}
					if len(slice) != 2 {
						t.Fatalf("expected 2 elements, got %d", len(slice))
					}
					id, ok := slice[0].(int64)
					if !ok || id != 1 {
						t.Errorf("expected id 1, got %v", slice[0])
					}
					// Verify V25 format
					p := slice[1].(map[string]any)
					provider, ok := p["provider"].(map[string]any)
					if !ok {
						t.Fatal("expected provider to be map for V25")
					}
					if provider["type"] != "S3" {
						t.Errorf("expected provider.type S3, got %v", provider["type"])
					}
					return json.RawMessage(`{"id": 1}`), nil
				}
				return sampleCredentialV25JSON(), nil
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 25, Minor: 4})
	cred, err := svc.UpdateCredential(context.Background(), 1, UpdateCredentialOpts{
		Name:         "My S3 Cred",
		ProviderType: "S3",
		Attributes: map[string]string{
			"access_key_id":     "AKIATEST",
			"secret_access_key": "secret123",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cred == nil {
		t.Fatal("expected non-nil credential")
	}
	if cred.ID != 1 {
		t.Errorf("expected ID 1, got %d", cred.ID)
	}
}

func TestCloudSyncService_UpdateCredential_V24(t *testing.T) {
	callCount := 0
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				callCount++
				if callCount == 1 {
					// Verify V24 format
					slice := params.([]any)
					p := slice[1].(map[string]any)
					if _, ok := p["provider"].(string); !ok {
						t.Fatal("expected provider to be string for V24")
					}
					if _, ok := p["attributes"].(map[string]any); !ok {
						t.Fatal("expected attributes to be map for V24")
					}
					return json.RawMessage(`{"id": 1}`), nil
				}
				return sampleCredentialV24JSON(), nil
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 24, Minor: 10})
	cred, err := svc.UpdateCredential(context.Background(), 1, UpdateCredentialOpts{
		Name:         "My S3 Cred",
		ProviderType: "S3",
		Attributes: map[string]string{
			"access_key_id": "AKIATEST",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cred == nil {
		t.Fatal("expected non-nil credential")
	}
}

func TestCloudSyncService_UpdateCredential_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, errors.New("not found")
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 25, Minor: 4})
	_, err := svc.UpdateCredential(context.Background(), 999, UpdateCredentialOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCloudSyncService_DeleteCredential(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				if method != "cloudsync.credentials.delete" {
					t.Errorf("expected method cloudsync.credentials.delete, got %s", method)
				}
				id, ok := params.(int64)
				if !ok || id != 5 {
					t.Errorf("expected id 5, got %v", params)
				}
				return nil, nil
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 25, Minor: 4})
	err := svc.DeleteCredential(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCloudSyncService_DeleteCredential_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, errors.New("permission denied")
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 25, Minor: 4})
	err := svc.DeleteCredential(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

// --- Task tests ---

func TestCloudSyncService_CreateTask(t *testing.T) {
	callCount := 0
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				callCount++
				if callCount == 1 {
					if method != "cloudsync.create" {
						t.Errorf("expected method cloudsync.create, got %s", method)
					}
					p := params.(map[string]any)
					if p["description"] != "Backup to S3" {
						t.Errorf("expected description 'Backup to S3', got %v", p["description"])
					}
					if p["path"] != "/mnt/tank/data" {
						t.Errorf("expected path '/mnt/tank/data', got %v", p["path"])
					}
					if p["direction"] != "PUSH" {
						t.Errorf("expected direction PUSH, got %v", p["direction"])
					}
					if p["snapshot"] != true {
						t.Error("expected snapshot=true")
					}
					// Verify schedule
					sched := p["schedule"].(map[string]any)
					if sched["hour"] != "3" {
						t.Errorf("expected schedule hour 3, got %v", sched["hour"])
					}
					// Verify attributes present
					if _, ok := p["attributes"]; !ok {
						t.Error("expected attributes in params")
					}
					// Verify bwlimit present
					if _, ok := p["bwlimit"]; !ok {
						t.Error("expected bwlimit in params")
					}
					// Verify exclude present
					if _, ok := p["exclude"]; !ok {
						t.Error("expected exclude in params")
					}
					return json.RawMessage(`{"id": 1}`), nil
				}
				return sampleTaskJSON(), nil
			},
		},
	}

	bw := int64(1048576)
	svc := NewCloudSyncService(mock, Version{Major: 25, Minor: 4})
	task, err := svc.CreateTask(context.Background(), CreateCloudSyncTaskOpts{
		Description:        "Backup to S3",
		Path:               "/mnt/tank/data",
		CredentialID:       5,
		Direction:          "PUSH",
		TransferMode:       "SYNC",
		Snapshot:           true,
		Transfers:          4,
		BWLimit:            []BwLimit{{Time: "08:00", Bandwidth: &bw}},
		FollowSymlinks:     false,
		CreateEmptySrcDirs: true,
		Enabled:            true,
		Schedule: Schedule{
			Minute: "0",
			Hour:   "3",
			Dom:    "*",
			Month:  "*",
			Dow:    "*",
		},
		Attributes: map[string]any{"bucket": "my-bucket", "folder": "/backups"},
		Exclude:    []string{"*.tmp"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task == nil {
		t.Fatal("expected non-nil task")
	}
	if task.ID != 1 {
		t.Errorf("expected ID 1, got %d", task.ID)
	}
	if task.Description != "Backup to S3" {
		t.Errorf("expected description 'Backup to S3', got %q", task.Description)
	}
	if task.Path != "/mnt/tank/data" {
		t.Errorf("expected path '/mnt/tank/data', got %q", task.Path)
	}
	if task.CredentialID != 5 {
		t.Errorf("expected credential ID 5, got %d", task.CredentialID)
	}
	if task.Direction != "PUSH" {
		t.Errorf("expected direction PUSH, got %s", task.Direction)
	}
	if !task.Snapshot {
		t.Error("expected snapshot=true")
	}
	if task.Schedule.Hour != "3" {
		t.Errorf("expected schedule hour 3, got %s", task.Schedule.Hour)
	}
	if len(task.BWLimit) != 1 {
		t.Fatalf("expected 1 bwlimit entry, got %d", len(task.BWLimit))
	}
	if task.BWLimit[0].Time != "08:00" {
		t.Errorf("expected bwlimit time 08:00, got %s", task.BWLimit[0].Time)
	}
	if len(task.Exclude) != 1 || task.Exclude[0] != "*.tmp" {
		t.Errorf("expected exclude [*.tmp], got %v", task.Exclude)
	}
	if !task.CreateEmptySrcDirs {
		t.Error("expected create_empty_src_dirs=true")
	}
}

func TestCloudSyncService_CreateTask_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, errors.New("connection refused")
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 25, Minor: 4})
	task, err := svc.CreateTask(context.Background(), CreateCloudSyncTaskOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
	if task != nil {
		t.Error("expected nil task on error")
	}
}

func TestCloudSyncService_CreateTask_ParseError(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return json.RawMessage(`not json`), nil
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 25, Minor: 4})
	_, err := svc.CreateTask(context.Background(), CreateCloudSyncTaskOpts{})
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestCloudSyncService_GetTask(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				if method != "cloudsync.query" {
					t.Errorf("expected method cloudsync.query, got %s", method)
				}
				return sampleTaskJSON(), nil
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 25, Minor: 4})
	task, err := svc.GetTask(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task == nil {
		t.Fatal("expected non-nil task")
	}
	if task.ID != 1 {
		t.Errorf("expected ID 1, got %d", task.ID)
	}
	if task.Description != "Backup to S3" {
		t.Errorf("expected description 'Backup to S3', got %q", task.Description)
	}
	if task.CredentialID != 5 {
		t.Errorf("expected credential ID 5, got %d", task.CredentialID)
	}
	if task.TransferMode != "SYNC" {
		t.Errorf("expected transfer mode SYNC, got %s", task.TransferMode)
	}
	if task.Transfers != 4 {
		t.Errorf("expected transfers 4, got %d", task.Transfers)
	}
	if !task.Enabled {
		t.Error("expected enabled=true")
	}
	if task.Attributes["bucket"] != "my-bucket" {
		t.Errorf("expected attributes.bucket 'my-bucket', got %v", task.Attributes["bucket"])
	}
	if task.Attributes["folder"] != "/backups" {
		t.Errorf("expected attributes.folder '/backups', got %v", task.Attributes["folder"])
	}
}

func TestCloudSyncService_GetTask_NotFound(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return json.RawMessage(`[]`), nil
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 25, Minor: 4})
	task, err := svc.GetTask(context.Background(), 999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task != nil {
		t.Error("expected nil task for not found")
	}
}

func TestCloudSyncService_GetTask_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, errors.New("timeout")
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 25, Minor: 4})
	_, err := svc.GetTask(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCloudSyncService_GetTask_ParseError(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return json.RawMessage(`not json`), nil
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 25, Minor: 4})
	_, err := svc.GetTask(context.Background(), 1)
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestCloudSyncService_GetTask_FalseAttributes(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return sampleTaskFalseAttrsJSON(), nil
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 25, Minor: 4})
	task, err := svc.GetTask(context.Background(), 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task == nil {
		t.Fatal("expected non-nil task")
	}
	if task.ID != 2 {
		t.Errorf("expected ID 2, got %d", task.ID)
	}
	// Attributes should be nil when API returns false
	if task.Attributes != nil {
		t.Errorf("expected nil attributes for false, got %v", task.Attributes)
	}
	if task.Direction != "PULL" {
		t.Errorf("expected direction PULL, got %s", task.Direction)
	}
	if task.TransferMode != "COPY" {
		t.Errorf("expected transfer mode COPY, got %s", task.TransferMode)
	}
	if !task.Encryption {
		t.Error("expected encryption=true")
	}
	if task.EncryptionPassword != "mypass" {
		t.Errorf("expected encryption password 'mypass', got %q", task.EncryptionPassword)
	}
	if task.EncryptionSalt != "mysalt" {
		t.Errorf("expected encryption salt 'mysalt', got %q", task.EncryptionSalt)
	}
	if !task.FollowSymlinks {
		t.Error("expected follow_symlinks=true")
	}
	if task.Enabled {
		t.Error("expected enabled=false")
	}
	if len(task.Include) != 1 || task.Include[0] != "*.dat" {
		t.Errorf("expected include [*.dat], got %v", task.Include)
	}
}

func TestCloudSyncService_ListTasks(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				if method != "cloudsync.query" {
					t.Errorf("expected method cloudsync.query, got %s", method)
				}
				if params != nil {
					t.Error("expected nil params for ListTasks")
				}
				return json.RawMessage(`[
					{
						"id": 1, "description": "Task 1", "path": "/mnt/a",
						"credentials": {"id": 1, "name": "Cred1"},
						"attributes": {"bucket": "b1"},
						"schedule": {"minute": "0", "hour": "1", "dom": "*", "month": "*", "dow": "*"},
						"direction": "PUSH", "transfer_mode": "SYNC",
						"encryption": false, "snapshot": false, "transfers": 4,
						"bwlimit": [], "exclude": [], "include": [],
						"follow_symlinks": false, "create_empty_src_dirs": false, "enabled": true
					},
					{
						"id": 2, "description": "Task 2", "path": "/mnt/b",
						"credentials": {"id": 2, "name": "Cred2"},
						"attributes": {"bucket": "b2"},
						"schedule": {"minute": "30", "hour": "*/2", "dom": "1", "month": "1-6", "dow": "1-5"},
						"direction": "PULL", "transfer_mode": "COPY",
						"encryption": false, "snapshot": false, "transfers": 2,
						"bwlimit": [], "exclude": [], "include": [],
						"follow_symlinks": false, "create_empty_src_dirs": false, "enabled": false
					}
				]`), nil
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 25, Minor: 4})
	tasks, err := svc.ListTasks(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
	if tasks[0].ID != 1 {
		t.Errorf("expected first task ID 1, got %d", tasks[0].ID)
	}
	if tasks[1].Direction != "PULL" {
		t.Errorf("expected second task direction PULL, got %s", tasks[1].Direction)
	}
	if tasks[1].Schedule.Dom != "1" {
		t.Errorf("expected second task dom '1', got %s", tasks[1].Schedule.Dom)
	}
}

func TestCloudSyncService_ListTasks_Empty(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return json.RawMessage(`[]`), nil
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 25, Minor: 4})
	tasks, err := svc.ListTasks(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("expected 0 tasks, got %d", len(tasks))
	}
}

func TestCloudSyncService_ListTasks_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, errors.New("network error")
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 25, Minor: 4})
	_, err := svc.ListTasks(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCloudSyncService_ListTasks_ParseError(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return json.RawMessage(`not json`), nil
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 25, Minor: 4})
	_, err := svc.ListTasks(context.Background())
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestCloudSyncService_UpdateTask(t *testing.T) {
	callCount := 0
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				callCount++
				if callCount == 1 {
					if method != "cloudsync.update" {
						t.Errorf("expected method cloudsync.update, got %s", method)
					}
					slice, ok := params.([]any)
					if !ok {
						t.Fatal("expected []any params for update")
					}
					if len(slice) != 2 {
						t.Fatalf("expected 2 elements, got %d", len(slice))
					}
					id, ok := slice[0].(int64)
					if !ok || id != 1 {
						t.Errorf("expected id 1, got %v", slice[0])
					}
					p := slice[1].(map[string]any)
					if p["description"] != "Updated Backup" {
						t.Errorf("expected description 'Updated Backup', got %v", p["description"])
					}
					return json.RawMessage(`{"id": 1}`), nil
				}
				return sampleTaskJSON(), nil
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 25, Minor: 4})
	task, err := svc.UpdateTask(context.Background(), 1, UpdateCloudSyncTaskOpts{
		Description:  "Updated Backup",
		Path:         "/mnt/tank/data",
		CredentialID: 5,
		Direction:    "PUSH",
		TransferMode: "SYNC",
		Transfers:    4,
		Enabled:      true,
		Schedule: Schedule{
			Minute: "0",
			Hour:   "3",
			Dom:    "*",
			Month:  "*",
			Dow:    "*",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task == nil {
		t.Fatal("expected non-nil task")
	}
	if task.ID != 1 {
		t.Errorf("expected ID 1, got %d", task.ID)
	}
}

func TestCloudSyncService_UpdateTask_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, errors.New("not found")
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 25, Minor: 4})
	_, err := svc.UpdateTask(context.Background(), 999, UpdateCloudSyncTaskOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCloudSyncService_DeleteTask(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				if method != "cloudsync.delete" {
					t.Errorf("expected method cloudsync.delete, got %s", method)
				}
				id, ok := params.(int64)
				if !ok || id != 5 {
					t.Errorf("expected id 5, got %v", params)
				}
				return nil, nil
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 25, Minor: 4})
	err := svc.DeleteTask(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCloudSyncService_DeleteTask_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, errors.New("permission denied")
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 25, Minor: 4})
	err := svc.DeleteTask(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCloudSyncService_Sync(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{},
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "cloudsync.sync" {
				t.Errorf("expected method cloudsync.sync, got %s", method)
			}
			id, ok := params.(int64)
			if !ok || id != 3 {
				t.Errorf("expected id 3, got %v", params)
			}
			return nil, nil
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 25, Minor: 4})
	err := svc.Sync(context.Background(), 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify CallAndWait was used (not Call)
	if len(mock.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(mock.calls))
	}
	if mock.calls[0].Method != "cloudsync.sync" {
		t.Errorf("expected method cloudsync.sync, got %s", mock.calls[0].Method)
	}
}

func TestCloudSyncService_Sync_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{},
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("sync failed: timeout")
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 25, Minor: 4})
	err := svc.Sync(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "sync failed: timeout" {
		t.Errorf("expected 'sync failed: timeout', got %q", err.Error())
	}
}

// --- Conversion function tests ---

func TestCredentialFromResponse_S3(t *testing.T) {
	resp := CloudSyncCredentialResponse{
		ID:   1,
		Name: "S3 Cred",
		Provider: CloudSyncCredentialProvider{
			Type:            "S3",
			AccessKeyID:     "AKIATEST",
			SecretAccessKey: "secret",
			Endpoint:        "s3.example.com",
			Region:          "us-east-1",
		},
	}

	cred := credentialFromResponse(resp)
	if cred.ID != 1 {
		t.Errorf("expected ID 1, got %d", cred.ID)
	}
	if cred.ProviderType != "S3" {
		t.Errorf("expected provider type S3, got %s", cred.ProviderType)
	}
	if cred.Attributes["access_key_id"] != "AKIATEST" {
		t.Errorf("expected access_key_id AKIATEST, got %s", cred.Attributes["access_key_id"])
	}
	if cred.Attributes["secret_access_key"] != "secret" {
		t.Errorf("expected secret_access_key secret, got %s", cred.Attributes["secret_access_key"])
	}
	if cred.Attributes["endpoint"] != "s3.example.com" {
		t.Errorf("expected endpoint s3.example.com, got %s", cred.Attributes["endpoint"])
	}
	if cred.Attributes["region"] != "us-east-1" {
		t.Errorf("expected region us-east-1, got %s", cred.Attributes["region"])
	}
	// Should not have B2/GCS keys
	if _, ok := cred.Attributes["account"]; ok {
		t.Error("unexpected account attribute for S3")
	}
}

func TestCredentialFromResponse_B2(t *testing.T) {
	resp := CloudSyncCredentialResponse{
		ID:   2,
		Name: "B2 Cred",
		Provider: CloudSyncCredentialProvider{
			Type:    "B2",
			Account: "b2account",
			Key:     "b2key",
		},
	}

	cred := credentialFromResponse(resp)
	if cred.ProviderType != "B2" {
		t.Errorf("expected provider type B2, got %s", cred.ProviderType)
	}
	if cred.Attributes["account"] != "b2account" {
		t.Errorf("expected account b2account, got %s", cred.Attributes["account"])
	}
	if cred.Attributes["key"] != "b2key" {
		t.Errorf("expected key b2key, got %s", cred.Attributes["key"])
	}
	// Should not have S3 keys
	if _, ok := cred.Attributes["access_key_id"]; ok {
		t.Error("unexpected access_key_id attribute for B2")
	}
}

func TestCredentialFromResponse_GCS(t *testing.T) {
	resp := CloudSyncCredentialResponse{
		ID:   3,
		Name: "GCS Cred",
		Provider: CloudSyncCredentialProvider{
			Type:                      "GOOGLE_CLOUD_STORAGE",
			ServiceAccountCredentials: `{"type": "service_account"}`,
		},
	}

	cred := credentialFromResponse(resp)
	if cred.ProviderType != "GOOGLE_CLOUD_STORAGE" {
		t.Errorf("expected provider type GOOGLE_CLOUD_STORAGE, got %s", cred.ProviderType)
	}
	if cred.Attributes["service_account_credentials"] != `{"type": "service_account"}` {
		t.Errorf("unexpected service_account_credentials: %s", cred.Attributes["service_account_credentials"])
	}
}

func TestCredentialFromResponse_EmptyProvider(t *testing.T) {
	resp := CloudSyncCredentialResponse{
		ID:   4,
		Name: "Empty",
		Provider: CloudSyncCredentialProvider{
			Type: "UNKNOWN",
		},
	}

	cred := credentialFromResponse(resp)
	if len(cred.Attributes) != 0 {
		t.Errorf("expected empty attributes, got %v", cred.Attributes)
	}
}

func TestCredentialOptsToAttrsAny(t *testing.T) {
	attrs := map[string]string{
		"access_key_id":     "AKIATEST",
		"secret_access_key": "secret",
	}
	result := credentialOptsToAttrsAny(attrs)
	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}
	if result["access_key_id"] != "AKIATEST" {
		t.Errorf("expected access_key_id AKIATEST, got %v", result["access_key_id"])
	}
}

func TestCredentialOptsToAttrsAny_Nil(t *testing.T) {
	result := credentialOptsToAttrsAny(nil)
	if result != nil {
		t.Errorf("expected nil result for nil input, got %v", result)
	}
}

func TestTaskOptsToParams(t *testing.T) {
	bw := int64(1048576)
	opts := CreateCloudSyncTaskOpts{
		Description:        "Test Task",
		Path:               "/mnt/tank/data",
		CredentialID:       5,
		Direction:          "PUSH",
		TransferMode:       "SYNC",
		Snapshot:           true,
		Transfers:          4,
		BWLimit:            []BwLimit{{Time: "08:00", Bandwidth: &bw}},
		FollowSymlinks:     false,
		CreateEmptySrcDirs: true,
		Enabled:            true,
		Encryption:         true,
		EncryptionPassword: "pass",
		EncryptionSalt:     "salt",
		Schedule: Schedule{
			Minute: "0",
			Hour:   "3",
			Dom:    "*",
			Month:  "*",
			Dow:    "*",
		},
		Attributes: map[string]any{"bucket": "my-bucket"},
		Exclude:    []string{"*.tmp"},
		Include:    []string{"*.dat"},
	}

	params := taskOptsToParams(opts)

	if params["description"] != "Test Task" {
		t.Errorf("expected description 'Test Task', got %v", params["description"])
	}
	if params["path"] != "/mnt/tank/data" {
		t.Errorf("expected path '/mnt/tank/data', got %v", params["path"])
	}
	if params["credentials"] != int64(5) {
		t.Errorf("expected credentials 5, got %v", params["credentials"])
	}
	if params["direction"] != "PUSH" {
		t.Errorf("expected direction PUSH, got %v", params["direction"])
	}
	if params["encryption"] != true {
		t.Error("expected encryption=true")
	}
	if params["encryption_password"] != "pass" {
		t.Errorf("expected encryption_password 'pass', got %v", params["encryption_password"])
	}
	if params["encryption_salt"] != "salt" {
		t.Errorf("expected encryption_salt 'salt', got %v", params["encryption_salt"])
	}

	sched := params["schedule"].(map[string]any)
	if sched["hour"] != "3" {
		t.Errorf("expected schedule hour 3, got %v", sched["hour"])
	}

	if _, ok := params["bwlimit"]; !ok {
		t.Error("expected bwlimit in params")
	}
	if _, ok := params["exclude"]; !ok {
		t.Error("expected exclude in params")
	}
	if _, ok := params["include"]; !ok {
		t.Error("expected include in params")
	}
	if _, ok := params["attributes"]; !ok {
		t.Error("expected attributes in params")
	}
}

func TestTaskOptsToParams_NoOptionalFields(t *testing.T) {
	opts := CreateCloudSyncTaskOpts{
		Description:  "Minimal Task",
		Path:         "/mnt/tank",
		CredentialID: 1,
		Direction:    "PUSH",
		TransferMode: "SYNC",
		Schedule: Schedule{
			Minute: "*",
			Hour:   "*",
			Dom:    "*",
			Month:  "*",
			Dow:    "*",
		},
	}

	params := taskOptsToParams(opts)

	// Encryption is false, so no encryption_password/encryption_salt
	if _, ok := params["encryption_password"]; ok {
		t.Error("unexpected encryption_password when encryption=false")
	}
	if _, ok := params["encryption_salt"]; ok {
		t.Error("unexpected encryption_salt when encryption=false")
	}
	// No bwlimit, exclude, include, attributes
	if _, ok := params["bwlimit"]; ok {
		t.Error("unexpected bwlimit when empty")
	}
	if _, ok := params["exclude"]; ok {
		t.Error("unexpected exclude when empty")
	}
	if _, ok := params["include"]; ok {
		t.Error("unexpected include when empty")
	}
	if _, ok := params["attributes"]; ok {
		t.Error("unexpected attributes when nil")
	}
}

func TestTaskFromResponse_NilBWLimit(t *testing.T) {
	resp := CloudSyncTaskResponse{
		ID:          1,
		Description: "Test",
		Path:        "/mnt/tank",
		Credentials: CloudSyncTaskCredentialRef{ID: 1, Name: "Cred"},
		Attributes:  json.RawMessage(`{"bucket": "b"}`),
		Schedule:    ScheduleResponse{Minute: "*", Hour: "*", Dom: "*", Month: "*", Dow: "*"},
		Direction:   "PUSH",
		TransferMode: "SYNC",
		Enabled:     true,
	}

	task := taskFromResponse(resp)
	if task.BWLimit != nil {
		t.Errorf("expected nil bwlimit, got %v", task.BWLimit)
	}
}

func TestTaskFromResponse_EmptyAttributes(t *testing.T) {
	resp := CloudSyncTaskResponse{
		ID:          1,
		Description: "Test",
		Credentials: CloudSyncTaskCredentialRef{ID: 1},
		Attributes:  json.RawMessage(`{}`),
		Schedule:    ScheduleResponse{Minute: "*", Hour: "*", Dom: "*", Month: "*", Dow: "*"},
	}

	task := taskFromResponse(resp)
	if task.Attributes == nil {
		t.Error("expected non-nil attributes for empty object")
	}
	if len(task.Attributes) != 0 {
		t.Errorf("expected 0 attributes, got %d", len(task.Attributes))
	}
}

func TestTaskFromResponse_NullAttributes(t *testing.T) {
	resp := CloudSyncTaskResponse{
		ID:          1,
		Description: "Test",
		Credentials: CloudSyncTaskCredentialRef{ID: 1},
		Attributes:  nil,
		Schedule:    ScheduleResponse{Minute: "*", Hour: "*", Dom: "*", Month: "*", Dow: "*"},
	}

	task := taskFromResponse(resp)
	if task.Attributes != nil {
		t.Errorf("expected nil attributes, got %v", task.Attributes)
	}
}

func TestNewCloudSyncService(t *testing.T) {
	mock := &mockAsyncCaller{}
	v := Version{Major: 25, Minor: 4}
	svc := NewCloudSyncService(mock, v)
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.client != mock {
		t.Error("expected client to be set")
	}
	if svc.version != v {
		t.Error("expected version to be set")
	}
}

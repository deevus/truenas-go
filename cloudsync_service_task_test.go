package truenas

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

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

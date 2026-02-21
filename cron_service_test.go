package truenas

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
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

// sampleCronJobJSON returns a JSON response for a cron job with inverted stdout/stderr.
func sampleCronJobJSON() json.RawMessage {
	return json.RawMessage(`[{
		"id": 1,
		"user": "root",
		"command": "/usr/local/bin/backup.sh",
		"description": "Daily backup",
		"enabled": true,
		"stdout": false,
		"stderr": true,
		"schedule": {
			"minute": "0",
			"hour": "3",
			"dom": "*",
			"month": "*",
			"dow": "*"
		}
	}]`)
}

func TestCronService_Create(t *testing.T) {
	callCount := 0
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			callCount++
			if callCount == 1 {
				// cronjob.create returns object with ID
				if method != "cronjob.create" {
					t.Errorf("expected method cronjob.create, got %s", method)
				}
				// Verify stdout/stderr inversion in params
				p := params.(map[string]any)
				if p["stdout"] != true {
					t.Error("expected stdout=true (CaptureStdout=false inverted)")
				}
				if p["stderr"] != false {
					t.Error("expected stderr=false (CaptureStderr=true inverted)")
				}
				return json.RawMessage(`{"id": 1}`), nil
			}
			// cronjob.query for re-read
			return sampleCronJobJSON(), nil
		},
	}

	svc := NewCronService(mock)
	job, err := svc.Create(context.Background(), CreateCronJobOpts{
		User:          "root",
		Command:       "/usr/local/bin/backup.sh",
		Description:   "Daily backup",
		Enabled:       true,
		CaptureStdout: false,
		CaptureStderr: true,
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
	if job == nil {
		t.Fatal("expected non-nil job")
	}
	if job.ID != 1 {
		t.Errorf("expected ID 1, got %d", job.ID)
	}
	if job.User != "root" {
		t.Errorf("expected user root, got %s", job.User)
	}
	if job.Command != "/usr/local/bin/backup.sh" {
		t.Errorf("expected command /usr/local/bin/backup.sh, got %s", job.Command)
	}
	// Verify inversion: API stdout=false → CaptureStdout=true
	if !job.CaptureStdout {
		t.Error("expected CaptureStdout=true (API stdout=false)")
	}
	// Verify inversion: API stderr=true → CaptureStderr=false
	if job.CaptureStderr {
		t.Error("expected CaptureStderr=false (API stderr=true)")
	}
	if job.Schedule.Hour != "3" {
		t.Errorf("expected hour 3, got %s", job.Schedule.Hour)
	}
}

func TestCronService_Create_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("connection refused")
		},
	}

	svc := NewCronService(mock)
	job, err := svc.Create(context.Background(), CreateCronJobOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
	if job != nil {
		t.Error("expected nil job on error")
	}
	if err.Error() != "connection refused" {
		t.Errorf("expected 'connection refused', got %q", err.Error())
	}
}

func TestCronService_Create_ParseError(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return json.RawMessage(`not json`), nil
		},
	}

	svc := NewCronService(mock)
	_, err := svc.Create(context.Background(), CreateCronJobOpts{})
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestCronService_Get(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "cronjob.query" {
				t.Errorf("expected method cronjob.query, got %s", method)
			}
			return sampleCronJobJSON(), nil
		},
	}

	svc := NewCronService(mock)
	job, err := svc.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job == nil {
		t.Fatal("expected non-nil job")
	}
	if job.ID != 1 {
		t.Errorf("expected ID 1, got %d", job.ID)
	}
	if job.Description != "Daily backup" {
		t.Errorf("expected description 'Daily backup', got %q", job.Description)
	}
	if !job.Enabled {
		t.Error("expected enabled=true")
	}
}

func TestCronService_Get_NotFound(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return json.RawMessage(`[]`), nil
		},
	}

	svc := NewCronService(mock)
	job, err := svc.Get(context.Background(), 999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job != nil {
		t.Error("expected nil job for not found")
	}
}

func TestCronService_Get_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("timeout")
		},
	}

	svc := NewCronService(mock)
	_, err := svc.Get(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCronService_List(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "cronjob.query" {
				t.Errorf("expected method cronjob.query, got %s", method)
			}
			if params != nil {
				t.Error("expected nil params for List")
			}
			return json.RawMessage(`[
				{"id": 1, "user": "root", "command": "cmd1", "description": "", "enabled": true, "stdout": true, "stderr": true, "schedule": {"minute": "0", "hour": "1", "dom": "*", "month": "*", "dow": "*"}},
				{"id": 2, "user": "admin", "command": "cmd2", "description": "job2", "enabled": false, "stdout": false, "stderr": false, "schedule": {"minute": "30", "hour": "*/2", "dom": "1", "month": "1-6", "dow": "1-5"}}
			]`), nil
		},
	}

	svc := NewCronService(mock)
	jobs, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(jobs))
	}
	if jobs[0].ID != 1 {
		t.Errorf("expected first job ID 1, got %d", jobs[0].ID)
	}
	if jobs[1].User != "admin" {
		t.Errorf("expected second job user admin, got %s", jobs[1].User)
	}
	if jobs[1].Schedule.Dom != "1" {
		t.Errorf("expected second job dom '1', got %s", jobs[1].Schedule.Dom)
	}
}

func TestCronService_List_Empty(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return json.RawMessage(`[]`), nil
		},
	}

	svc := NewCronService(mock)
	jobs, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(jobs) != 0 {
		t.Errorf("expected 0 jobs, got %d", len(jobs))
	}
}

func TestCronService_List_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("network error")
		},
	}

	svc := NewCronService(mock)
	_, err := svc.List(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCronService_Update(t *testing.T) {
	callCount := 0
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			callCount++
			if callCount == 1 {
				if method != "cronjob.update" {
					t.Errorf("expected method cronjob.update, got %s", method)
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
				return json.RawMessage(`{"id": 1}`), nil
			}
			// Re-query
			return sampleCronJobJSON(), nil
		},
	}

	svc := NewCronService(mock)
	job, err := svc.Update(context.Background(), 1, UpdateCronJobOpts{
		User:    "root",
		Command: "/usr/local/bin/backup.sh",
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
	if job == nil {
		t.Fatal("expected non-nil job")
	}
	if job.ID != 1 {
		t.Errorf("expected ID 1, got %d", job.ID)
	}
}

func TestCronService_Update_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("not found")
		},
	}

	svc := NewCronService(mock)
	_, err := svc.Update(context.Background(), 999, UpdateCronJobOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCronService_Delete(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "cronjob.delete" {
				t.Errorf("expected method cronjob.delete, got %s", method)
			}
			id, ok := params.(int64)
			if !ok || id != 5 {
				t.Errorf("expected id 5, got %v", params)
			}
			return nil, nil
		},
	}

	svc := NewCronService(mock)
	err := svc.Delete(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCronService_Delete_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("permission denied")
		},
	}

	svc := NewCronService(mock)
	err := svc.Delete(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestOptsToParams_Inversion(t *testing.T) {
	opts := CreateCronJobOpts{
		User:          "root",
		Command:       "echo hello",
		CaptureStdout: true,
		CaptureStderr: false,
		Schedule: Schedule{
			Minute: "*/5",
			Hour:   "*",
			Dom:    "*",
			Month:  "*",
			Dow:    "*",
		},
	}

	params := optsToParams(opts)

	// CaptureStdout=true should become stdout=false
	if params["stdout"] != false {
		t.Errorf("expected stdout=false, got %v", params["stdout"])
	}
	// CaptureStderr=false should become stderr=true
	if params["stderr"] != true {
		t.Errorf("expected stderr=true, got %v", params["stderr"])
	}
}

func TestCronJobFromResponse_Inversion(t *testing.T) {
	resp := CronJobResponse{
		ID:      1,
		User:    "root",
		Command: "test",
		Stdout:  true,  // API: true = don't capture
		Stderr:  false, // API: false = capture
		Schedule: ScheduleResponse{
			Minute: "0",
			Hour:   "0",
			Dom:    "*",
			Month:  "*",
			Dow:    "*",
		},
	}

	job := cronJobFromResponse(resp)

	// stdout=true → CaptureStdout=false
	if job.CaptureStdout {
		t.Error("expected CaptureStdout=false (API stdout=true)")
	}
	// stderr=false → CaptureStderr=true
	if !job.CaptureStderr {
		t.Error("expected CaptureStderr=true (API stderr=false)")
	}
}

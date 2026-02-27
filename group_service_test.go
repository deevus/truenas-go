package truenas

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

// --- Conversion tests ---

func TestGroupFromResponse(t *testing.T) {
	resp := GroupResponse{
		ID:                   42,
		GID:                  5000,
		Name:                 "developers",
		Builtin:              false,
		SMB:                  true,
		SudoCommands:         []string{"/usr/bin/apt"},
		SudoCommandsNopasswd: []string{"/usr/bin/systemctl"},
		Users:                []int64{1, 2, 3},
		Local:                true,
		Immutable:            false,
	}

	group := groupFromResponse(resp)

	if group.ID != 42 {
		t.Errorf("expected ID 42, got %d", group.ID)
	}
	if group.GID != 5000 {
		t.Errorf("expected GID 5000, got %d", group.GID)
	}
	if group.Name != "developers" {
		t.Errorf("expected Name developers, got %s", group.Name)
	}
	if group.Builtin {
		t.Error("expected Builtin false")
	}
	if !group.SMB {
		t.Error("expected SMB true")
	}
	if len(group.SudoCommands) != 1 || group.SudoCommands[0] != "/usr/bin/apt" {
		t.Errorf("unexpected SudoCommands: %v", group.SudoCommands)
	}
	if len(group.SudoCommandsNopasswd) != 1 || group.SudoCommandsNopasswd[0] != "/usr/bin/systemctl" {
		t.Errorf("unexpected SudoCommandsNopasswd: %v", group.SudoCommandsNopasswd)
	}
	if len(group.Users) != 3 {
		t.Errorf("expected 3 users, got %d", len(group.Users))
	}
	if !group.Local {
		t.Error("expected Local true")
	}
	if group.Immutable {
		t.Error("expected Immutable false")
	}
}

func TestGroupFromResponse_NilSlices(t *testing.T) {
	resp := GroupResponse{
		ID:   1,
		GID:  1000,
		Name: "empty",
	}

	group := groupFromResponse(resp)

	if group.SudoCommands != nil {
		t.Errorf("expected nil SudoCommands, got %v", group.SudoCommands)
	}
	if group.SudoCommandsNopasswd != nil {
		t.Errorf("expected nil SudoCommandsNopasswd, got %v", group.SudoCommandsNopasswd)
	}
	if group.Users != nil {
		t.Errorf("expected nil Users, got %v", group.Users)
	}
}

func TestGroupCreateOptsToParams(t *testing.T) {
	opts := CreateGroupOpts{
		Name:                 "devs",
		GID:                  5000,
		SMB:                  true,
		SudoCommands:         []string{"/usr/bin/apt"},
		SudoCommandsNopasswd: []string{"/usr/bin/systemctl"},
	}

	params := groupCreateOptsToParams(opts)

	if params["name"] != "devs" {
		t.Errorf("expected name=devs, got %v", params["name"])
	}
	if params["gid"] != int64(5000) {
		t.Errorf("expected gid=5000, got %v", params["gid"])
	}
	if params["smb"] != true {
		t.Errorf("expected smb=true, got %v", params["smb"])
	}
	sudoCmds, ok := params["sudo_commands"].([]string)
	if !ok || len(sudoCmds) != 1 || sudoCmds[0] != "/usr/bin/apt" {
		t.Errorf("unexpected sudo_commands: %v", params["sudo_commands"])
	}
}

func TestGroupCreateOptsToParams_ZeroGID(t *testing.T) {
	opts := CreateGroupOpts{
		Name: "auto-gid",
	}

	params := groupCreateOptsToParams(opts)

	if _, ok := params["gid"]; ok {
		t.Error("expected gid to be omitted when zero")
	}
}

func TestGroupUpdateOptsToParams(t *testing.T) {
	opts := UpdateGroupOpts{
		Name:                 "new-name",
		SMB:                  false,
		SudoCommands:         []string{},
		SudoCommandsNopasswd: []string{},
	}

	params := groupUpdateOptsToParams(opts)

	if params["name"] != "new-name" {
		t.Errorf("expected name=new-name, got %v", params["name"])
	}
	if params["smb"] != false {
		t.Errorf("expected smb=false, got %v", params["smb"])
	}
	// No gid field in update
	if _, ok := params["gid"]; ok {
		t.Error("gid should not be in update params")
	}
}

// --- Service CRUD tests ---

func TestNewGroupService(t *testing.T) {
	mock := &mockCaller{}
	v := Version{Major: 25, Minor: 4}
	svc := NewGroupService(mock, v)
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

func TestGroupService_Create(t *testing.T) {
	groupJSON := `{"id": 42, "gid": 5000, "name": "devs", "builtin": false, "smb": true, "sudo_commands": [], "sudo_commands_nopasswd": [], "users": [], "local": true, "immutable": false}`

	callCount := 0
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			callCount++
			switch callCount {
			case 1:
				// create call
				if method != "group.create" {
					t.Errorf("expected group.create, got %s", method)
				}
				return json.RawMessage(`42`), nil
			case 2:
				// get_instance call
				if method != "group.get_instance" {
					t.Errorf("expected group.get_instance, got %s", method)
				}
				return json.RawMessage(groupJSON), nil
			default:
				t.Fatalf("unexpected call %d: %s", callCount, method)
				return nil, nil
			}
		},
	}

	svc := NewGroupService(mock, Version{})
	group, err := svc.Create(context.Background(), CreateGroupOpts{
		Name: "devs",
		GID:  5000,
		SMB:  true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if group.ID != 42 {
		t.Errorf("expected ID 42, got %d", group.ID)
	}
	if group.Name != "devs" {
		t.Errorf("expected Name devs, got %s", group.Name)
	}
}

func TestGroupService_Get(t *testing.T) {
	groupJSON := `{"id": 42, "gid": 5000, "name": "devs", "builtin": false, "smb": true, "sudo_commands": [], "sudo_commands_nopasswd": [], "users": [], "local": true, "immutable": false}`

	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "group.get_instance" {
				t.Errorf("expected group.get_instance, got %s", method)
			}
			return json.RawMessage(groupJSON), nil
		},
	}

	svc := NewGroupService(mock, Version{})
	group, err := svc.Get(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if group == nil {
		t.Fatal("expected non-nil group")
	}
	if group.ID != 42 {
		t.Errorf("expected ID 42, got %d", group.ID)
	}
}

func TestGroupService_Get_NotFound(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("group 999 does not exist")
		},
	}

	svc := NewGroupService(mock, Version{})
	group, err := svc.Get(context.Background(), 999)
	if err != nil {
		t.Fatalf("expected nil error for not-found, got: %v", err)
	}
	if group != nil {
		t.Fatalf("expected nil group for not-found, got: %v", group)
	}
}

func TestGroupService_GetByName(t *testing.T) {
	groupJSON := `[{"id": 42, "gid": 5000, "name": "devs", "builtin": false, "smb": true, "sudo_commands": [], "sudo_commands_nopasswd": [], "users": [], "local": true, "immutable": false}]`

	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "group.query" {
				t.Errorf("expected group.query, got %s", method)
			}
			// Verify filter
			filters, ok := params.([][]any)
			if !ok || len(filters) != 1 {
				t.Fatalf("expected filter array, got %v", params)
			}
			if filters[0][0] != "group" || filters[0][1] != "=" || filters[0][2] != "devs" {
				t.Errorf("unexpected filter: %v", filters[0])
			}
			return json.RawMessage(groupJSON), nil
		},
	}

	svc := NewGroupService(mock, Version{})
	group, err := svc.GetByName(context.Background(), "devs")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if group == nil {
		t.Fatal("expected non-nil group")
	}
	if group.Name != "devs" {
		t.Errorf("expected Name devs, got %s", group.Name)
	}
}

func TestGroupService_GetByName_NotFound(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return json.RawMessage(`[]`), nil
		},
	}

	svc := NewGroupService(mock, Version{})
	group, err := svc.GetByName(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if group != nil {
		t.Fatalf("expected nil group, got: %v", group)
	}
}

func TestGroupService_GetByGID(t *testing.T) {
	groupJSON := `[{"id": 42, "gid": 5000, "name": "devs", "builtin": false, "smb": true, "sudo_commands": [], "sudo_commands_nopasswd": [], "users": [], "local": true, "immutable": false}]`

	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "group.query" {
				t.Errorf("expected group.query, got %s", method)
			}
			filters, ok := params.([][]any)
			if !ok || len(filters) != 1 {
				t.Fatalf("expected filter array, got %v", params)
			}
			if filters[0][0] != "gid" || filters[0][1] != "=" || filters[0][2] != int64(5000) {
				t.Errorf("unexpected filter: %v", filters[0])
			}
			return json.RawMessage(groupJSON), nil
		},
	}

	svc := NewGroupService(mock, Version{})
	group, err := svc.GetByGID(context.Background(), 5000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if group == nil {
		t.Fatal("expected non-nil group")
	}
	if group.GID != 5000 {
		t.Errorf("expected GID 5000, got %d", group.GID)
	}
}

func TestGroupService_List(t *testing.T) {
	listJSON := `[
		{"id": 1, "gid": 1000, "name": "wheel", "builtin": true, "smb": false, "sudo_commands": [], "sudo_commands_nopasswd": [], "users": [1], "local": false, "immutable": true},
		{"id": 2, "gid": 5000, "name": "devs", "builtin": false, "smb": true, "sudo_commands": [], "sudo_commands_nopasswd": [], "users": [], "local": true, "immutable": false}
	]`

	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "group.query" {
				t.Errorf("expected group.query, got %s", method)
			}
			if params != nil {
				t.Errorf("expected nil params for List, got %v", params)
			}
			return json.RawMessage(listJSON), nil
		},
	}

	svc := NewGroupService(mock, Version{})
	groups, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
	if groups[0].Name != "wheel" {
		t.Errorf("expected first group wheel, got %s", groups[0].Name)
	}
	if groups[1].Name != "devs" {
		t.Errorf("expected second group devs, got %s", groups[1].Name)
	}
}

func TestGroupService_Update(t *testing.T) {
	groupJSON := `{"id": 42, "gid": 5000, "name": "new-devs", "builtin": false, "smb": false, "sudo_commands": [], "sudo_commands_nopasswd": [], "users": [], "local": true, "immutable": false}`

	callCount := 0
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			callCount++
			switch callCount {
			case 1:
				if method != "group.update" {
					t.Errorf("expected group.update, got %s", method)
				}
				// Verify params are [id, updateMap]
				arr, ok := params.([]any)
				if !ok || len(arr) != 2 {
					t.Fatalf("expected [id, params] array, got %v", params)
				}
				if arr[0] != int64(42) {
					t.Errorf("expected id=42, got %v", arr[0])
				}
				return json.RawMessage(`42`), nil
			case 2:
				if method != "group.get_instance" {
					t.Errorf("expected group.get_instance, got %s", method)
				}
				return json.RawMessage(groupJSON), nil
			default:
				t.Fatalf("unexpected call %d", callCount)
				return nil, nil
			}
		},
	}

	svc := NewGroupService(mock, Version{})
	group, err := svc.Update(context.Background(), 42, UpdateGroupOpts{
		Name: "new-devs",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if group.Name != "new-devs" {
		t.Errorf("expected Name new-devs, got %s", group.Name)
	}
}

func TestGroupService_Delete(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "group.delete" {
				t.Errorf("expected group.delete, got %s", method)
			}
			// Verify params are [id, {delete_users: false}]
			arr, ok := params.([]any)
			if !ok || len(arr) != 2 {
				t.Fatalf("expected [id, opts] array, got %v", params)
			}
			if arr[0] != int64(42) {
				t.Errorf("expected id=42, got %v", arr[0])
			}
			opts, ok := arr[1].(map[string]any)
			if !ok {
				t.Fatalf("expected opts map, got %T", arr[1])
			}
			if opts["delete_users"] != false {
				t.Errorf("expected delete_users=false, got %v", opts["delete_users"])
			}
			return json.RawMessage(`true`), nil
		},
	}

	svc := NewGroupService(mock, Version{})
	err := svc.Delete(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- Error path tests ---

func TestGroupService_Create_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("create failed")
		},
	}

	svc := NewGroupService(mock, Version{})
	_, err := svc.Create(context.Background(), CreateGroupOpts{Name: "fail"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGroupService_Create_BadResponse(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return json.RawMessage(`"not a number"`), nil
		},
	}

	svc := NewGroupService(mock, Version{})
	_, err := svc.Create(context.Background(), CreateGroupOpts{Name: "fail"})
	if err == nil {
		t.Fatal("expected error for bad create response")
	}
}

func TestGroupService_Get_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("connection refused")
		},
	}

	svc := NewGroupService(mock, Version{})
	_, err := svc.Get(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGroupService_Get_BadJSON(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return json.RawMessage(`{bad json`), nil
		},
	}

	svc := NewGroupService(mock, Version{})
	_, err := svc.Get(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error for bad JSON")
	}
}

func TestGroupService_List_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("connection refused")
		},
	}

	svc := NewGroupService(mock, Version{})
	_, err := svc.List(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGroupService_List_BadJSON(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return json.RawMessage(`[{bad`), nil
		},
	}

	svc := NewGroupService(mock, Version{})
	_, err := svc.List(context.Background())
	if err == nil {
		t.Fatal("expected error for bad JSON")
	}
}

func TestGroupService_Update_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("update failed")
		},
	}

	svc := NewGroupService(mock, Version{})
	_, err := svc.Update(context.Background(), 1, UpdateGroupOpts{Name: "fail"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGroupService_Delete_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("delete failed")
		},
	}

	svc := NewGroupService(mock, Version{})
	err := svc.Delete(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGroupService_QueryOne_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("query failed")
		},
	}

	svc := NewGroupService(mock, Version{})
	_, err := svc.GetByName(context.Background(), "fail")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGroupService_QueryOne_BadJSON(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return json.RawMessage(`[{bad`), nil
		},
	}

	svc := NewGroupService(mock, Version{})
	_, err := svc.GetByName(context.Background(), "fail")
	if err == nil {
		t.Fatal("expected error for bad JSON")
	}
}

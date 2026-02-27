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
		ID:                   10,
		GID:                  1000,
		Name:                 "developers",
		Builtin:              false,
		SMB:                  true,
		SudoCommands:         []string{"ALL"},
		SudoCommandsNopasswd: []string{"/usr/bin/apt"},
		Users:                []int64{1, 2, 3},
		Local:                true,
		Immutable:            false,
	}

	group := groupFromResponse(resp)

	if group.ID != 10 {
		t.Errorf("expected ID 10, got %d", group.ID)
	}
	if group.GID != 1000 {
		t.Errorf("expected GID 1000, got %d", group.GID)
	}
	if group.Name != "developers" {
		t.Errorf("expected name developers, got %s", group.Name)
	}
	if group.Builtin {
		t.Error("expected Builtin=false")
	}
	if !group.SMB {
		t.Error("expected SMB=true")
	}
	if len(group.SudoCommands) != 1 || group.SudoCommands[0] != "ALL" {
		t.Errorf("expected SudoCommands=[ALL], got %v", group.SudoCommands)
	}
	if len(group.SudoCommandsNopasswd) != 1 || group.SudoCommandsNopasswd[0] != "/usr/bin/apt" {
		t.Errorf("expected SudoCommandsNopasswd=[/usr/bin/apt], got %v", group.SudoCommandsNopasswd)
	}
	if len(group.Users) != 3 {
		t.Errorf("expected 3 users, got %d", len(group.Users))
	}
	if !group.Local {
		t.Error("expected Local=true")
	}
	if group.Immutable {
		t.Error("expected Immutable=false")
	}
}

func TestGroupFromResponse_NilSlices(t *testing.T) {
	resp := GroupResponse{
		ID:   1,
		GID:  500,
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

func TestCreateGroupParams(t *testing.T) {
	opts := CreateGroupOpts{
		Name:                 "devs",
		GID:                  5000,
		SMB:                  true,
		SudoCommands:         []string{"ALL"},
		SudoCommandsNopasswd: []string{"/usr/bin/apt"},
	}

	params := createGroupParams(opts)

	if params["name"] != "devs" {
		t.Errorf("expected name devs, got %v", params["name"])
	}
	if params["gid"] != int64(5000) {
		t.Errorf("expected gid 5000, got %v", params["gid"])
	}
	if params["smb"] != true {
		t.Errorf("expected smb true, got %v", params["smb"])
	}
	if sc, ok := params["sudo_commands"].([]string); !ok || len(sc) != 1 {
		t.Errorf("expected sudo_commands [ALL], got %v", params["sudo_commands"])
	}
	if sc, ok := params["sudo_commands_nopasswd"].([]string); !ok || len(sc) != 1 {
		t.Errorf("expected sudo_commands_nopasswd [/usr/bin/apt], got %v", params["sudo_commands_nopasswd"])
	}
}

func TestCreateGroupParams_NoGID(t *testing.T) {
	opts := CreateGroupOpts{
		Name: "auto-gid",
		SMB:  false,
	}

	params := createGroupParams(opts)

	if _, ok := params["gid"]; ok {
		t.Error("expected no gid when GID is 0")
	}
	if params["smb"] != false {
		t.Errorf("expected smb false, got %v", params["smb"])
	}
}

func TestCreateGroupParams_NoSudo(t *testing.T) {
	opts := CreateGroupOpts{
		Name: "no-sudo",
	}

	params := createGroupParams(opts)

	if _, ok := params["sudo_commands"]; ok {
		t.Error("expected no sudo_commands when nil")
	}
	if _, ok := params["sudo_commands_nopasswd"]; ok {
		t.Error("expected no sudo_commands_nopasswd when nil")
	}
}

func TestUpdateGroupParams(t *testing.T) {
	opts := UpdateGroupOpts{
		Name: "renamed",
		SMB:  true,
	}

	params := updateGroupParams(opts)

	if params["name"] != "renamed" {
		t.Errorf("expected name renamed, got %v", params["name"])
	}
	if params["smb"] != true {
		t.Errorf("expected smb true, got %v", params["smb"])
	}
	// GID must never be included in update
	if _, ok := params["gid"]; ok {
		t.Error("expected no gid in update params")
	}
}

// --- CRUD tests ---

func TestGroupService_Create(t *testing.T) {
	mock := &mockCaller{}
	callCount := 0
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		callCount++
		switch method {
		case "group.create":
			p := params.(map[string]any)
			if p["name"] != "devs" {
				t.Errorf("expected name devs, got %v", p["name"])
			}
			return json.RawMessage(`10`), nil
		case "group.get_instance":
			return json.RawMessage(`{"id":10,"gid":5000,"group":"devs","smb":true,"builtin":false,"local":true,"immutable":false}`), nil
		default:
			t.Errorf("unexpected method: %s", method)
			return nil, nil
		}
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
	if group == nil {
		t.Fatal("expected non-nil group")
	}
	if group.ID != 10 {
		t.Errorf("expected ID 10, got %d", group.ID)
	}
	if group.Name != "devs" {
		t.Errorf("expected name devs, got %s", group.Name)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls (create + get_instance), got %d", callCount)
	}
}

func TestGroupService_Create_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("permission denied")
		},
	}

	svc := NewGroupService(mock, Version{})
	group, err := svc.Create(context.Background(), CreateGroupOpts{Name: "fail"})
	if err == nil {
		t.Fatal("expected error")
	}
	if group != nil {
		t.Error("expected nil group on error")
	}
}

func TestGroupService_Get(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "group.get_instance" {
				t.Errorf("expected method group.get_instance, got %s", method)
			}
			id, ok := params.(int64)
			if !ok || id != 10 {
				t.Errorf("expected id 10, got %v", params)
			}
			return json.RawMessage(`{"id":10,"gid":1000,"group":"wheel","smb":false,"builtin":true,"local":true,"immutable":true}`), nil
		},
	}

	svc := NewGroupService(mock, Version{})
	group, err := svc.Get(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if group == nil {
		t.Fatal("expected non-nil group")
	}
	if group.Name != "wheel" {
		t.Errorf("expected name wheel, got %s", group.Name)
	}
	if !group.Builtin {
		t.Error("expected Builtin=true")
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
		t.Error("expected nil group for not-found")
	}
}

func TestGroupService_GetByName(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "group.query" {
				t.Errorf("expected method group.query, got %s", method)
			}
			filter, ok := params.([][]any)
			if !ok {
				t.Fatal("expected [][]any params")
			}
			if len(filter) != 1 || filter[0][0] != "group" || filter[0][1] != "=" || filter[0][2] != "wheel" {
				t.Errorf("unexpected filter: %v", filter)
			}
			return json.RawMessage(`[{"id":10,"gid":0,"group":"wheel","smb":false,"builtin":true,"local":true,"immutable":true}]`), nil
		},
	}

	svc := NewGroupService(mock, Version{})
	group, err := svc.GetByName(context.Background(), "wheel")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if group == nil {
		t.Fatal("expected non-nil group")
	}
	if group.Name != "wheel" {
		t.Errorf("expected name wheel, got %s", group.Name)
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
		t.Error("expected nil group for not-found")
	}
}

func TestGroupService_GetByGID(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "group.query" {
				t.Errorf("expected method group.query, got %s", method)
			}
			filter := params.([][]any)
			if filter[0][0] != "gid" || filter[0][2] != int64(1000) {
				t.Errorf("unexpected filter: %v", filter)
			}
			return json.RawMessage(`[{"id":5,"gid":1000,"group":"staff","smb":true,"builtin":false,"local":true,"immutable":false}]`), nil
		},
	}

	svc := NewGroupService(mock, Version{})
	group, err := svc.GetByGID(context.Background(), 1000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if group == nil {
		t.Fatal("expected non-nil group")
	}
	if group.GID != 1000 {
		t.Errorf("expected GID 1000, got %d", group.GID)
	}
}

func TestGroupService_List(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "group.query" {
				t.Errorf("expected method group.query, got %s", method)
			}
			if params != nil {
				t.Error("expected nil params for List")
			}
			return json.RawMessage(`[
				{"id":1,"gid":0,"group":"wheel","builtin":true,"smb":false,"local":true,"immutable":true},
				{"id":2,"gid":1000,"group":"devs","builtin":false,"smb":true,"local":true,"immutable":false}
			]`), nil
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

func TestGroupService_List_Empty(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return json.RawMessage(`[]`), nil
		},
	}

	svc := NewGroupService(mock, Version{})
	groups, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 0 {
		t.Errorf("expected 0 groups, got %d", len(groups))
	}
}

func TestGroupService_Update(t *testing.T) {
	mock := &mockCaller{}
	callCount := 0
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		callCount++
		switch method {
		case "group.update":
			slice := params.([]any)
			if slice[0] != int64(10) {
				t.Errorf("expected id 10, got %v", slice[0])
			}
			p := slice[1].(map[string]any)
			if p["name"] != "renamed" {
				t.Errorf("expected name renamed, got %v", p["name"])
			}
			return nil, nil
		case "group.get_instance":
			return json.RawMessage(`{"id":10,"gid":1000,"group":"renamed","smb":true,"builtin":false,"local":true,"immutable":false}`), nil
		default:
			t.Errorf("unexpected method: %s", method)
			return nil, nil
		}
	}

	svc := NewGroupService(mock, Version{})
	group, err := svc.Update(context.Background(), 10, UpdateGroupOpts{
		Name: "renamed",
		SMB:  true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if group == nil {
		t.Fatal("expected non-nil group")
	}
	if group.Name != "renamed" {
		t.Errorf("expected name renamed, got %s", group.Name)
	}
}

func TestGroupService_Update_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("not found")
		},
	}

	svc := NewGroupService(mock, Version{})
	_, err := svc.Update(context.Background(), 999, UpdateGroupOpts{Name: "fail"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGroupService_Delete(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "group.delete" {
				t.Errorf("expected method group.delete, got %s", method)
			}
			slice := params.([]any)
			if slice[0] != int64(10) {
				t.Errorf("expected id 10, got %v", slice[0])
			}
			opts := slice[1].(map[string]any)
			if opts["delete_users"] != false {
				t.Errorf("expected delete_users=false, got %v", opts["delete_users"])
			}
			return nil, nil
		},
	}

	svc := NewGroupService(mock, Version{})
	err := svc.Delete(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGroupService_Delete_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("permission denied")
		},
	}

	svc := NewGroupService(mock, Version{})
	err := svc.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error")
	}
}

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

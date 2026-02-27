package truenas

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

// --- Conversion tests ---

func TestUserFromResponse(t *testing.T) {
	email := "jdoe@example.com"
	sshkey := "ssh-rsa AAAA..."
	resp := UserResponse{
		ID:       1,
		UID:      1000,
		Username: "jdoe",
		FullName: "John Doe",
		Email:    &email,
		Home:     "/home/jdoe",
		Shell:    "/usr/bin/bash",
		HomeMode: "755",
		Group: UserGroupResponse{
			ID:        10,
			GID:       1000,
			GroupName: "jdoe",
		},
		Groups:               []int64{20, 30},
		SMB:                  true,
		PasswordDisabled:     false,
		SSHPasswordEnabled:   true,
		SSHPubKey:            &sshkey,
		Locked:               false,
		SudoCommands:         []string{"ALL"},
		SudoCommandsNopasswd: []string{"/usr/bin/apt"},
		Builtin:              false,
		Local:                true,
		Immutable:            false,
	}

	user := userFromResponse(resp)

	if user.ID != 1 {
		t.Errorf("expected ID 1, got %d", user.ID)
	}
	if user.UID != 1000 {
		t.Errorf("expected UID 1000, got %d", user.UID)
	}
	if user.Username != "jdoe" {
		t.Errorf("expected username jdoe, got %s", user.Username)
	}
	if user.FullName != "John Doe" {
		t.Errorf("expected full name John Doe, got %s", user.FullName)
	}
	if user.Email != "jdoe@example.com" {
		t.Errorf("expected email jdoe@example.com, got %s", user.Email)
	}
	if user.Home != "/home/jdoe" {
		t.Errorf("expected home /home/jdoe, got %s", user.Home)
	}
	if user.Shell != "/usr/bin/bash" {
		t.Errorf("expected shell /usr/bin/bash, got %s", user.Shell)
	}
	if user.HomeMode != "755" {
		t.Errorf("expected home mode 755, got %s", user.HomeMode)
	}
	if user.GroupID != 10 {
		t.Errorf("expected GroupID 10, got %d", user.GroupID)
	}
	if len(user.Groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(user.Groups))
	}
	if !user.SMB {
		t.Error("expected SMB=true")
	}
	if user.PasswordDisabled {
		t.Error("expected PasswordDisabled=false")
	}
	if !user.SSHPasswordEnabled {
		t.Error("expected SSHPasswordEnabled=true")
	}
	if user.SSHPubKey != "ssh-rsa AAAA..." {
		t.Errorf("expected SSHPubKey, got %s", user.SSHPubKey)
	}
	if user.Locked {
		t.Error("expected Locked=false")
	}
	if len(user.SudoCommands) != 1 || user.SudoCommands[0] != "ALL" {
		t.Errorf("expected SudoCommands=[ALL], got %v", user.SudoCommands)
	}
	if !user.Local {
		t.Error("expected Local=true")
	}
	if user.Immutable {
		t.Error("expected Immutable=false")
	}
}

func TestUserFromResponse_NullableFields(t *testing.T) {
	resp := UserResponse{
		ID:       2,
		UID:      1001,
		Username: "nomail",
		FullName: "No Mail",
		Email:    nil,
		SSHPubKey: nil,
		Group:    UserGroupResponse{ID: 10},
	}

	user := userFromResponse(resp)

	if user.Email != "" {
		t.Errorf("expected empty email for nil, got %q", user.Email)
	}
	if user.SSHPubKey != "" {
		t.Errorf("expected empty SSHPubKey for nil, got %q", user.SSHPubKey)
	}
}

func TestCreateUserParams(t *testing.T) {
	opts := CreateUserOpts{
		Username:         "jdoe",
		FullName:         "John Doe",
		Email:            "jdoe@example.com",
		UID:              1000,
		Password:         "secret",
		PasswordDisabled: false,
		Group:            10,
		GroupCreate:      false,
		Groups:           []int64{20, 30},
		Home:             "/home/jdoe",
		HomeCreate:       true,
		HomeMode:         "755",
		Shell:            "/usr/bin/bash",
		SMB:              true,
		SSHPasswordEnabled: true,
		SSHPubKey:        "ssh-rsa AAAA...",
		Locked:           false,
		SudoCommands:     []string{"ALL"},
		SudoCommandsNopasswd: []string{"/usr/bin/apt"},
	}

	params := createUserParams(opts)

	if params["username"] != "jdoe" {
		t.Errorf("expected username jdoe, got %v", params["username"])
	}
	if params["full_name"] != "John Doe" {
		t.Errorf("expected full_name John Doe, got %v", params["full_name"])
	}
	if params["email"] != "jdoe@example.com" {
		t.Errorf("expected email, got %v", params["email"])
	}
	if params["uid"] != int64(1000) {
		t.Errorf("expected uid 1000, got %v", params["uid"])
	}
	if params["password"] != "secret" {
		t.Errorf("expected password, got %v", params["password"])
	}
	if params["group"] != int64(10) {
		t.Errorf("expected group 10, got %v", params["group"])
	}
	if params["group_create"] != false {
		t.Errorf("expected group_create false, got %v", params["group_create"])
	}
	if groups, ok := params["groups"].([]int64); !ok || len(groups) != 2 {
		t.Errorf("expected groups [20 30], got %v", params["groups"])
	}
	if params["home_create"] != true {
		t.Errorf("expected home_create true, got %v", params["home_create"])
	}
	if params["smb"] != true {
		t.Errorf("expected smb true, got %v", params["smb"])
	}
	if params["sshpubkey"] != "ssh-rsa AAAA..." {
		t.Errorf("expected sshpubkey, got %v", params["sshpubkey"])
	}
}

func TestCreateUserParams_Minimal(t *testing.T) {
	opts := CreateUserOpts{
		Username: "minimal",
		FullName: "Minimal User",
	}

	params := createUserParams(opts)

	if params["username"] != "minimal" {
		t.Errorf("expected username minimal, got %v", params["username"])
	}
	// UID should be omitted when 0
	if _, ok := params["uid"]; ok {
		t.Error("expected no uid when 0")
	}
	// Password should be omitted when empty
	if _, ok := params["password"]; ok {
		t.Error("expected no password when empty")
	}
	// Group should be omitted when 0
	if _, ok := params["group"]; ok {
		t.Error("expected no group when 0")
	}
	// Groups should be omitted when nil
	if _, ok := params["groups"]; ok {
		t.Error("expected no groups when nil")
	}
	// SSHPubKey should be omitted when empty
	if _, ok := params["sshpubkey"]; ok {
		t.Error("expected no sshpubkey when empty")
	}
	// SudoCommands should be omitted when nil
	if _, ok := params["sudo_commands"]; ok {
		t.Error("expected no sudo_commands when nil")
	}
}

func TestUpdateUserParams(t *testing.T) {
	opts := UpdateUserOpts{
		Username: "renamed",
		FullName: "Renamed User",
		SMB:      true,
	}

	params := updateUserParams(opts)

	if params["username"] != "renamed" {
		t.Errorf("expected username renamed, got %v", params["username"])
	}
	if params["full_name"] != "Renamed User" {
		t.Errorf("expected full_name Renamed User, got %v", params["full_name"])
	}
	// uid must never be in update params
	if _, ok := params["uid"]; ok {
		t.Error("expected no uid in update params")
	}
	// group_create must never be in update params
	if _, ok := params["group_create"]; ok {
		t.Error("expected no group_create in update params")
	}
	// home_create must never be in update params
	if _, ok := params["home_create"]; ok {
		t.Error("expected no home_create in update params")
	}
}

// --- CRUD tests ---

func TestUserService_Create(t *testing.T) {
	mock := &mockCaller{}
	callCount := 0
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		callCount++
		switch method {
		case "user.create":
			p := params.(map[string]any)
			if p["username"] != "jdoe" {
				t.Errorf("expected username jdoe, got %v", p["username"])
			}
			return json.RawMessage(`{"id": 42}`), nil
		case "user.get_instance":
			return json.RawMessage(`{
				"id":42,"uid":1000,"username":"jdoe","full_name":"John Doe",
				"email":"jdoe@example.com","home":"/home/jdoe","shell":"/usr/bin/bash",
				"home_mode":"755","group":{"id":10,"bsdgrp_gid":1000,"bsdgrp_group":"jdoe"},
				"groups":[],"smb":true,"password_disabled":false,"ssh_password_enabled":false,
				"locked":false,"builtin":false,"local":true,"immutable":false
			}`), nil
		default:
			t.Errorf("unexpected method: %s", method)
			return nil, nil
		}
	}

	svc := NewUserService(mock, Version{})
	user, err := svc.Create(context.Background(), CreateUserOpts{
		Username: "jdoe",
		FullName: "John Doe",
		Email:    "jdoe@example.com",
		Password: "secret",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("expected non-nil user")
	}
	if user.ID != 42 {
		t.Errorf("expected ID 42, got %d", user.ID)
	}
	if user.Username != "jdoe" {
		t.Errorf("expected username jdoe, got %s", user.Username)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls (create + get_instance), got %d", callCount)
	}
}

func TestUserService_Create_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("permission denied")
		},
	}

	svc := NewUserService(mock, Version{})
	user, err := svc.Create(context.Background(), CreateUserOpts{Username: "fail"})
	if err == nil {
		t.Fatal("expected error")
	}
	if user != nil {
		t.Error("expected nil user on error")
	}
}

func TestUserService_Get(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "user.get_instance" {
				t.Errorf("expected method user.get_instance, got %s", method)
			}
			id, ok := params.(int64)
			if !ok || id != 42 {
				t.Errorf("expected id 42, got %v", params)
			}
			return json.RawMessage(`{
				"id":42,"uid":1000,"username":"jdoe","full_name":"John Doe",
				"email":null,"home":"/home/jdoe","shell":"/usr/bin/bash",
				"home_mode":"755","group":{"id":10,"bsdgrp_gid":1000,"bsdgrp_group":"jdoe"},
				"groups":[20],"smb":false,"password_disabled":true,"ssh_password_enabled":false,
				"sshpubkey":null,"locked":false,"builtin":false,"local":true,"immutable":false
			}`), nil
		},
	}

	svc := NewUserService(mock, Version{})
	user, err := svc.Get(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("expected non-nil user")
	}
	if user.Username != "jdoe" {
		t.Errorf("expected username jdoe, got %s", user.Username)
	}
	if user.Email != "" {
		t.Errorf("expected empty email for null, got %q", user.Email)
	}
	if user.GroupID != 10 {
		t.Errorf("expected GroupID 10, got %d", user.GroupID)
	}
	if len(user.Groups) != 1 || user.Groups[0] != 20 {
		t.Errorf("expected groups [20], got %v", user.Groups)
	}
	if !user.PasswordDisabled {
		t.Error("expected PasswordDisabled=true")
	}
}

func TestUserService_Get_NotFound(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("user 999 does not exist")
		},
	}

	svc := NewUserService(mock, Version{})
	user, err := svc.Get(context.Background(), 999)
	if err != nil {
		t.Fatalf("expected nil error for not-found, got: %v", err)
	}
	if user != nil {
		t.Error("expected nil user for not-found")
	}
}

func TestUserService_GetByUsername(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "user.query" {
				t.Errorf("expected method user.query, got %s", method)
			}
			filter, ok := params.([][]any)
			if !ok {
				t.Fatal("expected [][]any params")
			}
			if len(filter) != 1 || filter[0][0] != "username" || filter[0][1] != "=" || filter[0][2] != "jdoe" {
				t.Errorf("unexpected filter: %v", filter)
			}
			return json.RawMessage(`[{
				"id":42,"uid":1000,"username":"jdoe","full_name":"John Doe",
				"email":null,"home":"/home/jdoe","shell":"/usr/bin/bash",
				"home_mode":"755","group":{"id":10,"bsdgrp_gid":1000,"bsdgrp_group":"jdoe"},
				"groups":[],"smb":false,"password_disabled":false,"ssh_password_enabled":false,
				"locked":false,"builtin":false,"local":true,"immutable":false
			}]`), nil
		},
	}

	svc := NewUserService(mock, Version{})
	user, err := svc.GetByUsername(context.Background(), "jdoe")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("expected non-nil user")
	}
	if user.Username != "jdoe" {
		t.Errorf("expected username jdoe, got %s", user.Username)
	}
}

func TestUserService_GetByUsername_NotFound(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return json.RawMessage(`[]`), nil
		},
	}

	svc := NewUserService(mock, Version{})
	user, err := svc.GetByUsername(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if user != nil {
		t.Error("expected nil user for not-found")
	}
}

func TestUserService_GetByUID(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			filter := params.([][]any)
			if filter[0][0] != "uid" || filter[0][2] != int64(1000) {
				t.Errorf("unexpected filter: %v", filter)
			}
			return json.RawMessage(`[{
				"id":42,"uid":1000,"username":"jdoe","full_name":"John Doe",
				"email":null,"home":"/home/jdoe","shell":"/usr/bin/bash",
				"home_mode":"755","group":{"id":10,"bsdgrp_gid":1000,"bsdgrp_group":"jdoe"},
				"groups":[],"smb":false,"password_disabled":false,"ssh_password_enabled":false,
				"locked":false,"builtin":false,"local":true,"immutable":false
			}]`), nil
		},
	}

	svc := NewUserService(mock, Version{})
	user, err := svc.GetByUID(context.Background(), 1000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("expected non-nil user")
	}
	if user.UID != 1000 {
		t.Errorf("expected UID 1000, got %d", user.UID)
	}
}

func TestUserService_List(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "user.query" {
				t.Errorf("expected method user.query, got %s", method)
			}
			if params != nil {
				t.Error("expected nil params for List")
			}
			return json.RawMessage(`[
				{"id":1,"uid":0,"username":"root","full_name":"root","home":"/root","shell":"/usr/bin/zsh","home_mode":"755","group":{"id":1,"bsdgrp_gid":0,"bsdgrp_group":"wheel"},"groups":[],"smb":false,"password_disabled":false,"ssh_password_enabled":false,"locked":false,"builtin":true,"local":true,"immutable":true},
				{"id":42,"uid":1000,"username":"jdoe","full_name":"John Doe","home":"/home/jdoe","shell":"/usr/bin/bash","home_mode":"755","group":{"id":10,"bsdgrp_gid":1000,"bsdgrp_group":"jdoe"},"groups":[],"smb":true,"password_disabled":false,"ssh_password_enabled":false,"locked":false,"builtin":false,"local":true,"immutable":false}
			]`), nil
		},
	}

	svc := NewUserService(mock, Version{})
	users, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}
	if users[0].Username != "root" {
		t.Errorf("expected first user root, got %s", users[0].Username)
	}
	if users[1].Username != "jdoe" {
		t.Errorf("expected second user jdoe, got %s", users[1].Username)
	}
}

func TestUserService_List_Empty(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return json.RawMessage(`[]`), nil
		},
	}

	svc := NewUserService(mock, Version{})
	users, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 0 {
		t.Errorf("expected 0 users, got %d", len(users))
	}
}

func TestUserService_Update(t *testing.T) {
	mock := &mockCaller{}
	callCount := 0
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		callCount++
		switch method {
		case "user.update":
			slice := params.([]any)
			if slice[0] != int64(42) {
				t.Errorf("expected id 42, got %v", slice[0])
			}
			p := slice[1].(map[string]any)
			if p["full_name"] != "Jane Doe" {
				t.Errorf("expected full_name Jane Doe, got %v", p["full_name"])
			}
			// Verify immutable fields are not included
			if _, ok := p["uid"]; ok {
				t.Error("uid must not be in update params")
			}
			if _, ok := p["group_create"]; ok {
				t.Error("group_create must not be in update params")
			}
			if _, ok := p["home_create"]; ok {
				t.Error("home_create must not be in update params")
			}
			return nil, nil
		case "user.get_instance":
			return json.RawMessage(`{
				"id":42,"uid":1000,"username":"jdoe","full_name":"Jane Doe",
				"email":null,"home":"/home/jdoe","shell":"/usr/bin/bash",
				"home_mode":"755","group":{"id":10,"bsdgrp_gid":1000,"bsdgrp_group":"jdoe"},
				"groups":[],"smb":false,"password_disabled":false,"ssh_password_enabled":false,
				"locked":false,"builtin":false,"local":true,"immutable":false
			}`), nil
		default:
			t.Errorf("unexpected method: %s", method)
			return nil, nil
		}
	}

	svc := NewUserService(mock, Version{})
	user, err := svc.Update(context.Background(), 42, UpdateUserOpts{
		Username: "jdoe",
		FullName: "Jane Doe",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("expected non-nil user")
	}
	if user.FullName != "Jane Doe" {
		t.Errorf("expected full_name Jane Doe, got %s", user.FullName)
	}
}

func TestUserService_Update_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("not found")
		},
	}

	svc := NewUserService(mock, Version{})
	_, err := svc.Update(context.Background(), 999, UpdateUserOpts{Username: "fail"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUserService_Delete(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "user.delete" {
				t.Errorf("expected method user.delete, got %s", method)
			}
			slice := params.([]any)
			if slice[0] != int64(42) {
				t.Errorf("expected id 42, got %v", slice[0])
			}
			opts := slice[1].(map[string]any)
			if opts["delete_group"] != true {
				t.Errorf("expected delete_group=true, got %v", opts["delete_group"])
			}
			return nil, nil
		},
	}

	svc := NewUserService(mock, Version{})
	err := svc.Delete(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUserService_Delete_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("permission denied")
		},
	}

	svc := NewUserService(mock, Version{})
	err := svc.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewUserService(t *testing.T) {
	mock := &mockCaller{}
	v := Version{Major: 25, Minor: 4}
	svc := NewUserService(mock, v)
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

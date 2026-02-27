package truenas

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

// --- Conversion tests ---

func TestUserFromResponse(t *testing.T) {
	resp := UserResponse{
		ID:       10,
		UID:      1001,
		Username: "jdoe",
		FullName: "John Doe",
		Email:    strPtr("john@example.com"),
		Home:     "/home/jdoe",
		Shell:    "/usr/bin/zsh",
		HomeMode: "755",
		Group: UserGroupRef{
			ID:   42,
			GID:  5000,
			Name: "devs",
		},
		Groups:               []int64{100, 200},
		SMB:                  true,
		PasswordDisabled:     false,
		SSHPasswordEnabled:   true,
		SSHPubKey:            strPtr("ssh-ed25519 AAAA..."),
		Locked:               false,
		SudoCommands:         []string{"/usr/bin/apt"},
		SudoCommandsNopasswd: []string{"/usr/bin/systemctl"},
		Builtin:              false,
		Local:                true,
		Immutable:            false,
	}

	user := userFromResponse(resp)

	if user.ID != 10 {
		t.Errorf("expected ID 10, got %d", user.ID)
	}
	if user.UID != 1001 {
		t.Errorf("expected UID 1001, got %d", user.UID)
	}
	if user.Username != "jdoe" {
		t.Errorf("expected Username jdoe, got %s", user.Username)
	}
	if user.FullName != "John Doe" {
		t.Errorf("expected FullName John Doe, got %s", user.FullName)
	}
	if user.Email != "john@example.com" {
		t.Errorf("expected Email john@example.com, got %s", user.Email)
	}
	if user.Home != "/home/jdoe" {
		t.Errorf("expected Home /home/jdoe, got %s", user.Home)
	}
	if user.Shell != "/usr/bin/zsh" {
		t.Errorf("expected Shell /usr/bin/zsh, got %s", user.Shell)
	}
	if user.HomeMode != "755" {
		t.Errorf("expected HomeMode 755, got %s", user.HomeMode)
	}
	if user.GroupID != 42 {
		t.Errorf("expected GroupID 42, got %d", user.GroupID)
	}
	if len(user.Groups) != 2 || user.Groups[0] != 100 || user.Groups[1] != 200 {
		t.Errorf("unexpected Groups: %v", user.Groups)
	}
	if !user.SMB {
		t.Error("expected SMB true")
	}
	if user.PasswordDisabled {
		t.Error("expected PasswordDisabled false")
	}
	if !user.SSHPasswordEnabled {
		t.Error("expected SSHPasswordEnabled true")
	}
	if user.SSHPubKey != "ssh-ed25519 AAAA..." {
		t.Errorf("expected SSHPubKey, got %s", user.SSHPubKey)
	}
	if user.Locked {
		t.Error("expected Locked false")
	}
	if len(user.SudoCommands) != 1 || user.SudoCommands[0] != "/usr/bin/apt" {
		t.Errorf("unexpected SudoCommands: %v", user.SudoCommands)
	}
	if len(user.SudoCommandsNopasswd) != 1 || user.SudoCommandsNopasswd[0] != "/usr/bin/systemctl" {
		t.Errorf("unexpected SudoCommandsNopasswd: %v", user.SudoCommandsNopasswd)
	}
	if user.Builtin {
		t.Error("expected Builtin false")
	}
	if !user.Local {
		t.Error("expected Local true")
	}
	if user.Immutable {
		t.Error("expected Immutable false")
	}
}

func TestUserFromResponse_NullableFields(t *testing.T) {
	resp := UserResponse{
		ID:       1,
		UID:      1000,
		Username: "test",
		Email:    nil,
		SSHPubKey: nil,
		Group:    UserGroupRef{ID: 1},
	}

	user := userFromResponse(resp)

	if user.Email != "" {
		t.Errorf("expected empty Email for nil, got %s", user.Email)
	}
	if user.SSHPubKey != "" {
		t.Errorf("expected empty SSHPubKey for nil, got %s", user.SSHPubKey)
	}
}

func TestUserFromResponse_NilSlices(t *testing.T) {
	resp := UserResponse{
		ID:       1,
		UID:      1000,
		Username: "test",
		Group:    UserGroupRef{ID: 1},
	}

	user := userFromResponse(resp)

	if user.Groups != nil {
		t.Errorf("expected nil Groups, got %v", user.Groups)
	}
	if user.SudoCommands != nil {
		t.Errorf("expected nil SudoCommands, got %v", user.SudoCommands)
	}
	if user.SudoCommandsNopasswd != nil {
		t.Errorf("expected nil SudoCommandsNopasswd, got %v", user.SudoCommandsNopasswd)
	}
}

func TestUserCreateOptsToParams(t *testing.T) {
	opts := CreateUserOpts{
		Username:             "jdoe",
		FullName:             "John Doe",
		Email:                "john@example.com",
		UID:                  1001,
		Password:             "secret123",
		PasswordDisabled:     false,
		Group:                42,
		GroupCreate:          true,
		Groups:               []int64{100, 200},
		Home:                 "/home/jdoe",
		HomeCreate:           true,
		HomeMode:             "755",
		Shell:                "/usr/bin/zsh",
		SMB:                  true,
		SSHPasswordEnabled:   true,
		SSHPubKey:            "ssh-ed25519 AAAA...",
		Locked:               false,
		SudoCommands:         []string{"/usr/bin/apt"},
		SudoCommandsNopasswd: []string{"/usr/bin/systemctl"},
	}

	params := userCreateOptsToParams(opts)

	if params["username"] != "jdoe" {
		t.Errorf("expected username=jdoe, got %v", params["username"])
	}
	if params["full_name"] != "John Doe" {
		t.Errorf("expected full_name=John Doe, got %v", params["full_name"])
	}
	if params["email"] != "john@example.com" {
		t.Errorf("expected email, got %v", params["email"])
	}
	if params["uid"] != int64(1001) {
		t.Errorf("expected uid=1001, got %v", params["uid"])
	}
	if params["password"] != "secret123" {
		t.Errorf("expected password=secret123, got %v", params["password"])
	}
	if params["password_disabled"] != false {
		t.Errorf("expected password_disabled=false, got %v", params["password_disabled"])
	}
	if params["group"] != int64(42) {
		t.Errorf("expected group=42, got %v", params["group"])
	}
	if params["group_create"] != true {
		t.Errorf("expected group_create=true, got %v", params["group_create"])
	}
	groups, ok := params["groups"].([]int64)
	if !ok || len(groups) != 2 {
		t.Errorf("expected groups [100,200], got %v", params["groups"])
	}
	if params["home"] != "/home/jdoe" {
		t.Errorf("expected home=/home/jdoe, got %v", params["home"])
	}
	if params["home_create"] != true {
		t.Errorf("expected home_create=true, got %v", params["home_create"])
	}
	if params["home_mode"] != "755" {
		t.Errorf("expected home_mode=755, got %v", params["home_mode"])
	}
	if params["shell"] != "/usr/bin/zsh" {
		t.Errorf("expected shell=/usr/bin/zsh, got %v", params["shell"])
	}
	if params["smb"] != true {
		t.Errorf("expected smb=true, got %v", params["smb"])
	}
	if params["ssh_password_enabled"] != true {
		t.Errorf("expected ssh_password_enabled=true, got %v", params["ssh_password_enabled"])
	}
	if params["sshpubkey"] != "ssh-ed25519 AAAA..." {
		t.Errorf("expected sshpubkey, got %v", params["sshpubkey"])
	}
	if params["locked"] != false {
		t.Errorf("expected locked=false, got %v", params["locked"])
	}
}

func TestUserCreateOptsToParams_ZeroUID(t *testing.T) {
	opts := CreateUserOpts{
		Username: "auto-uid",
		FullName: "Auto",
	}

	params := userCreateOptsToParams(opts)

	if _, ok := params["uid"]; ok {
		t.Error("expected uid to be omitted when zero")
	}
}

func TestUserCreateOptsToParams_OptionalFields(t *testing.T) {
	opts := CreateUserOpts{
		Username: "minimal",
		FullName: "Minimal User",
	}

	params := userCreateOptsToParams(opts)

	// email is always sent (even empty, so it can be cleared)
	if _, ok := params["email"]; !ok {
		t.Error("expected email to always be present")
	}
	if _, ok := params["password"]; ok {
		t.Error("expected password to be omitted when empty")
	}
	if _, ok := params["group"]; ok {
		t.Error("expected group to be omitted when zero")
	}
	if _, ok := params["group_create"]; ok {
		t.Error("expected group_create to be omitted when false")
	}
	if _, ok := params["home_create"]; ok {
		t.Error("expected home_create to be omitted when false")
	}
	if _, ok := params["sshpubkey"]; ok {
		t.Error("expected sshpubkey to be omitted when empty")
	}
}

func TestUserUpdateOptsToParams(t *testing.T) {
	opts := UpdateUserOpts{
		Username:         "newname",
		FullName:         "New Name",
		Email:            "new@example.com",
		PasswordDisabled: true,
		SMB:              false,
	}

	params := userUpdateOptsToParams(opts)

	if params["username"] != "newname" {
		t.Errorf("expected username=newname, got %v", params["username"])
	}
	if params["full_name"] != "New Name" {
		t.Errorf("expected full_name=New Name, got %v", params["full_name"])
	}
	if params["email"] != "new@example.com" {
		t.Errorf("expected email=new@example.com, got %v", params["email"])
	}
	// No uid, group_create, home_create in update
	if _, ok := params["uid"]; ok {
		t.Error("uid should not be in update params")
	}
	if _, ok := params["group_create"]; ok {
		t.Error("group_create should not be in update params")
	}
	if _, ok := params["home_create"]; ok {
		t.Error("home_create should not be in update params")
	}
}

// --- Service CRUD tests ---

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

func sampleUserJSON() string {
	return `{"id": 10, "uid": 1001, "username": "jdoe", "full_name": "John Doe", "email": "john@example.com", "home": "/home/jdoe", "shell": "/usr/bin/zsh", "home_mode": "755", "group": {"id": 42, "bsdgrp_gid": 5000, "bsdgrp_group": "devs"}, "groups": [100], "smb": true, "password_disabled": false, "ssh_password_enabled": false, "sshpubkey": null, "locked": false, "sudo_commands": [], "sudo_commands_nopasswd": [], "builtin": false, "local": true, "immutable": false}`
}

func TestUserService_Create(t *testing.T) {
	callCount := 0
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			callCount++
			switch callCount {
			case 1:
				if method != "user.create" {
					t.Errorf("expected user.create, got %s", method)
				}
				return json.RawMessage(`{"id": 10}`), nil
			case 2:
				if method != "user.get_instance" {
					t.Errorf("expected user.get_instance, got %s", method)
				}
				return json.RawMessage(sampleUserJSON()), nil
			default:
				t.Fatalf("unexpected call %d: %s", callCount, method)
				return nil, nil
			}
		},
	}

	svc := NewUserService(mock, Version{})
	user, err := svc.Create(context.Background(), CreateUserOpts{
		Username: "jdoe",
		FullName: "John Doe",
		UID:      1001,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != 10 {
		t.Errorf("expected ID 10, got %d", user.ID)
	}
	if user.Username != "jdoe" {
		t.Errorf("expected Username jdoe, got %s", user.Username)
	}
}

func TestUserService_Get(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "user.get_instance" {
				t.Errorf("expected user.get_instance, got %s", method)
			}
			return json.RawMessage(sampleUserJSON()), nil
		},
	}

	svc := NewUserService(mock, Version{})
	user, err := svc.Get(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("expected non-nil user")
	}
	if user.ID != 10 {
		t.Errorf("expected ID 10, got %d", user.ID)
	}
	if user.GroupID != 42 {
		t.Errorf("expected GroupID 42, got %d", user.GroupID)
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
		t.Fatalf("expected nil user for not-found, got: %v", user)
	}
}

func TestUserService_GetByUsername(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "user.query" {
				t.Errorf("expected user.query, got %s", method)
			}
			filters, ok := params.([][]any)
			if !ok || len(filters) != 1 {
				t.Fatalf("expected filter array, got %v", params)
			}
			if filters[0][0] != "username" || filters[0][1] != "=" || filters[0][2] != "jdoe" {
				t.Errorf("unexpected filter: %v", filters[0])
			}
			return json.RawMessage("[" + sampleUserJSON() + "]"), nil
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
		t.Errorf("expected Username jdoe, got %s", user.Username)
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
		t.Fatalf("expected nil user, got: %v", user)
	}
}

func TestUserService_GetByUID(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "user.query" {
				t.Errorf("expected user.query, got %s", method)
			}
			filters, ok := params.([][]any)
			if !ok || len(filters) != 1 {
				t.Fatalf("expected filter array, got %v", params)
			}
			if filters[0][0] != "uid" || filters[0][1] != "=" || filters[0][2] != int64(1001) {
				t.Errorf("unexpected filter: %v", filters[0])
			}
			return json.RawMessage("[" + sampleUserJSON() + "]"), nil
		},
	}

	svc := NewUserService(mock, Version{})
	user, err := svc.GetByUID(context.Background(), 1001)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("expected non-nil user")
	}
	if user.UID != 1001 {
		t.Errorf("expected UID 1001, got %d", user.UID)
	}
}

func TestUserService_List(t *testing.T) {
	listJSON := `[` + sampleUserJSON() + `]`

	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "user.query" {
				t.Errorf("expected user.query, got %s", method)
			}
			if params != nil {
				t.Errorf("expected nil params for List, got %v", params)
			}
			return json.RawMessage(listJSON), nil
		},
	}

	svc := NewUserService(mock, Version{})
	users, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(users))
	}
	if users[0].Username != "jdoe" {
		t.Errorf("expected jdoe, got %s", users[0].Username)
	}
}

func TestUserService_Update(t *testing.T) {
	callCount := 0
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			callCount++
			switch callCount {
			case 1:
				if method != "user.update" {
					t.Errorf("expected user.update, got %s", method)
				}
				arr, ok := params.([]any)
				if !ok || len(arr) != 2 {
					t.Fatalf("expected [id, params] array, got %v", params)
				}
				if arr[0] != int64(10) {
					t.Errorf("expected id=10, got %v", arr[0])
				}
				return json.RawMessage(`10`), nil
			case 2:
				if method != "user.get_instance" {
					t.Errorf("expected user.get_instance, got %s", method)
				}
				return json.RawMessage(sampleUserJSON()), nil
			default:
				t.Fatalf("unexpected call %d", callCount)
				return nil, nil
			}
		},
	}

	svc := NewUserService(mock, Version{})
	user, err := svc.Update(context.Background(), 10, UpdateUserOpts{
		Username: "jdoe",
		FullName: "John Doe Updated",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != 10 {
		t.Errorf("expected ID 10, got %d", user.ID)
	}
}

func TestUserService_Delete(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "user.delete" {
				t.Errorf("expected user.delete, got %s", method)
			}
			arr, ok := params.([]any)
			if !ok || len(arr) != 2 {
				t.Fatalf("expected [id, opts] array, got %v", params)
			}
			if arr[0] != int64(10) {
				t.Errorf("expected id=10, got %v", arr[0])
			}
			opts, ok := arr[1].(map[string]any)
			if !ok {
				t.Fatalf("expected opts map, got %T", arr[1])
			}
			if opts["delete_group"] != true {
				t.Errorf("expected delete_group=true, got %v", opts["delete_group"])
			}
			return json.RawMessage(`true`), nil
		},
	}

	svc := NewUserService(mock, Version{})
	err := svc.Delete(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- Error path tests ---

func TestUserService_Create_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("create failed")
		},
	}

	svc := NewUserService(mock, Version{})
	_, err := svc.Create(context.Background(), CreateUserOpts{Username: "fail"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUserService_Create_BadResponse(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return json.RawMessage(`"not a number"`), nil
		},
	}

	svc := NewUserService(mock, Version{})
	_, err := svc.Create(context.Background(), CreateUserOpts{Username: "fail"})
	if err == nil {
		t.Fatal("expected error for bad create response")
	}
}

func TestUserService_Get_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("connection refused")
		},
	}

	svc := NewUserService(mock, Version{})
	_, err := svc.Get(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUserService_Get_BadJSON(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return json.RawMessage(`{bad json`), nil
		},
	}

	svc := NewUserService(mock, Version{})
	_, err := svc.Get(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error for bad JSON")
	}
}

func TestUserService_List_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("connection refused")
		},
	}

	svc := NewUserService(mock, Version{})
	_, err := svc.List(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUserService_List_BadJSON(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return json.RawMessage(`[{bad`), nil
		},
	}

	svc := NewUserService(mock, Version{})
	_, err := svc.List(context.Background())
	if err == nil {
		t.Fatal("expected error for bad JSON")
	}
}

func TestUserService_Update_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("update failed")
		},
	}

	svc := NewUserService(mock, Version{})
	_, err := svc.Update(context.Background(), 1, UpdateUserOpts{Username: "fail"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUserService_QueryOne_Error(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("query failed")
		},
	}

	svc := NewUserService(mock, Version{})
	_, err := svc.GetByUsername(context.Background(), "fail")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUserService_QueryOne_BadJSON(t *testing.T) {
	mock := &mockCaller{
		callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return json.RawMessage(`[{bad`), nil
		},
	}

	svc := NewUserService(mock, Version{})
	_, err := svc.GetByUsername(context.Background(), "fail")
	if err == nil {
		t.Fatal("expected error for bad JSON")
	}
}

func TestUserUpdateOptsToParams_AllOptionalFields(t *testing.T) {
	opts := UpdateUserOpts{
		Username:             "jdoe",
		FullName:             "John",
		Email:                "j@example.com",
		Password:             "newpass",
		Group:                42,
		Groups:               []int64{100},
		SSHPubKey:            "ssh-ed25519 AAAA...",
		SudoCommands:         []string{"/usr/bin/apt"},
		SudoCommandsNopasswd: []string{"/usr/bin/systemctl"},
	}

	params := userUpdateOptsToParams(opts)

	if params["email"] != "j@example.com" {
		t.Errorf("expected email, got %v", params["email"])
	}
	if params["password"] != "newpass" {
		t.Errorf("expected password, got %v", params["password"])
	}
	if params["group"] != int64(42) {
		t.Errorf("expected group=42, got %v", params["group"])
	}
	if params["sshpubkey"] != "ssh-ed25519 AAAA..." {
		t.Errorf("expected sshpubkey, got %v", params["sshpubkey"])
	}
	groups, ok := params["groups"].([]int64)
	if !ok || len(groups) != 1 {
		t.Errorf("expected groups, got %v", params["groups"])
	}
	sudoCmds, ok := params["sudo_commands"].([]string)
	if !ok || len(sudoCmds) != 1 {
		t.Errorf("expected sudo_commands, got %v", params["sudo_commands"])
	}
	sudoNp, ok := params["sudo_commands_nopasswd"].([]string)
	if !ok || len(sudoNp) != 1 {
		t.Errorf("expected sudo_commands_nopasswd, got %v", params["sudo_commands_nopasswd"])
	}
}

// helper
func strPtr(s string) *string { return &s }

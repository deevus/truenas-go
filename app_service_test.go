package truenas

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// sampleAppJSON returns a JSON response for a single app (no config).
func sampleAppJSON() json.RawMessage {
	return json.RawMessage(`[{
		"name": "my-app",
		"state": "RUNNING",
		"custom_app": true
	}]`)
}

// sampleAppWithConfigJSON returns a JSON response for a single app with config.
func sampleAppWithConfigJSON() json.RawMessage {
	return json.RawMessage(`[{
		"name": "my-app",
		"state": "RUNNING",
		"custom_app": true,
		"config": {"version": "1.0", "port": 8080}
	}]`)
}

// sampleRegistryJSON returns a JSON response for a single registry.
func sampleRegistryJSON() json.RawMessage {
	return json.RawMessage(`[{
		"id": 1,
		"name": "my-registry",
		"description": "A test registry",
		"username": "admin",
		"password": "secret",
		"uri": "https://registry.example.com"
	}]`)
}

// sampleRegistryNullDescJSON returns a JSON response for a registry with null description.
func sampleRegistryNullDescJSON() json.RawMessage {
	return json.RawMessage(`[{
		"id": 2,
		"name": "null-desc-registry",
		"description": null,
		"username": "user",
		"password": "pass",
		"uri": "https://registry2.example.com"
	}]`)
}

// --- App CRUD tests ---

func TestAppService_CreateApp(t *testing.T) {
	callCount := 0
	mock := &mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "app.create" {
				t.Errorf("expected method app.create, got %s", method)
			}
			p := params.(map[string]any)
			if p["app_name"] != "my-app" {
				t.Errorf("expected app_name my-app, got %v", p["app_name"])
			}
			if p["custom_app"] != true {
				t.Errorf("expected custom_app true, got %v", p["custom_app"])
			}
			if p["custom_compose_config_string"] != "version: '3'" {
				t.Errorf("expected custom_compose_config_string, got %v", p["custom_compose_config_string"])
			}
			return nil, nil
		},
	}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		callCount++
		if method != "app.query" {
			t.Errorf("expected method app.query, got %s", method)
		}
		return sampleAppJSON(), nil
	}

	svc := NewAppService(mock, Version{})
	app, err := svc.CreateApp(context.Background(), CreateAppOpts{
		Name:                "my-app",
		CustomApp:           true,
		CustomComposeConfig: "version: '3'",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if app == nil {
		t.Fatal("expected non-nil app")
	}
	if app.Name != "my-app" {
		t.Errorf("expected name my-app, got %s", app.Name)
	}
	if app.State != "RUNNING" {
		t.Errorf("expected state RUNNING, got %s", app.State)
	}
	if !app.CustomApp {
		t.Error("expected custom_app true")
	}
}

func TestAppService_CreateApp_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("connection refused")
		},
	}

	svc := NewAppService(mock, Version{})
	app, err := svc.CreateApp(context.Background(), CreateAppOpts{Name: "fail-app"})
	if err == nil {
		t.Fatal("expected error")
	}
	if app != nil {
		t.Error("expected nil app on error")
	}
	if err.Error() != "connection refused" {
		t.Errorf("expected 'connection refused', got %q", err.Error())
	}
}

func TestAppService_CreateApp_NotFoundAfterCreate(t *testing.T) {
	mock := &mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, nil
		},
	}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		return json.RawMessage(`[]`), nil
	}

	svc := NewAppService(mock, Version{})
	app, err := svc.CreateApp(context.Background(), CreateAppOpts{Name: "ghost-app"})
	if err == nil {
		t.Fatal("expected error for not found after create")
	}
	if app != nil {
		t.Error("expected nil app")
	}
}

func TestAppService_CreateApp_ParseError(t *testing.T) {
	mock := &mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, nil
		},
	}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		return json.RawMessage(`not json`), nil
	}

	svc := NewAppService(mock, Version{})
	_, err := svc.CreateApp(context.Background(), CreateAppOpts{Name: "parse-fail"})
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestAppService_GetApp(t *testing.T) {
	mock := &mockAsyncCaller{}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		if method != "app.query" {
			t.Errorf("expected method app.query, got %s", method)
		}
		// Verify filter params
		filter, ok := params.([][]any)
		if !ok {
			t.Fatal("expected [][]any params for GetApp")
		}
		if len(filter) != 1 || filter[0][0] != "name" || filter[0][1] != "=" || filter[0][2] != "my-app" {
			t.Errorf("unexpected filter: %v", filter)
		}
		return sampleAppJSON(), nil
	}

	svc := NewAppService(mock, Version{})
	app, err := svc.GetApp(context.Background(), "my-app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if app == nil {
		t.Fatal("expected non-nil app")
	}
	if app.Name != "my-app" {
		t.Errorf("expected name my-app, got %s", app.Name)
	}
	if app.State != "RUNNING" {
		t.Errorf("expected state RUNNING, got %s", app.State)
	}
	if !app.CustomApp {
		t.Error("expected custom_app true")
	}
}

func TestAppService_GetApp_NotFound(t *testing.T) {
	mock := &mockAsyncCaller{}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		return json.RawMessage(`[]`), nil
	}

	svc := NewAppService(mock, Version{})
	app, err := svc.GetApp(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if app != nil {
		t.Error("expected nil app for not found")
	}
}

func TestAppService_GetApp_Error(t *testing.T) {
	mock := &mockAsyncCaller{}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		return nil, errors.New("timeout")
	}

	svc := NewAppService(mock, Version{})
	_, err := svc.GetApp(context.Background(), "my-app")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAppService_GetApp_ParseError(t *testing.T) {
	mock := &mockAsyncCaller{}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		return json.RawMessage(`not json`), nil
	}

	svc := NewAppService(mock, Version{})
	_, err := svc.GetApp(context.Background(), "my-app")
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestAppService_GetAppWithConfig(t *testing.T) {
	mock := &mockAsyncCaller{}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		if method != "app.query" {
			t.Errorf("expected method app.query, got %s", method)
		}
		// Verify params shape: []any{filter, extra}
		slice, ok := params.([]any)
		if !ok {
			t.Fatal("expected []any params for GetAppWithConfig")
		}
		if len(slice) != 2 {
			t.Fatalf("expected 2 elements, got %d", len(slice))
		}
		extra, ok := slice[1].(map[string]any)
		if !ok {
			t.Fatal("expected map[string]any for extra")
		}
		extraInner, ok := extra["extra"].(map[string]any)
		if !ok {
			t.Fatal("expected extra.extra to be map[string]any")
		}
		if extraInner["retrieve_config"] != true {
			t.Error("expected retrieve_config=true")
		}
		return sampleAppWithConfigJSON(), nil
	}

	svc := NewAppService(mock, Version{})
	app, err := svc.GetAppWithConfig(context.Background(), "my-app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if app == nil {
		t.Fatal("expected non-nil app")
	}
	if app.Name != "my-app" {
		t.Errorf("expected name my-app, got %s", app.Name)
	}
	if app.Config == nil {
		t.Fatal("expected non-nil config")
	}
	if app.Config["version"] != "1.0" {
		t.Errorf("expected config version 1.0, got %v", app.Config["version"])
	}
	// JSON numbers unmarshal as float64
	if app.Config["port"] != float64(8080) {
		t.Errorf("expected config port 8080, got %v", app.Config["port"])
	}
}

func TestAppService_GetAppWithConfig_NotFound(t *testing.T) {
	mock := &mockAsyncCaller{}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		return json.RawMessage(`[]`), nil
	}

	svc := NewAppService(mock, Version{})
	app, err := svc.GetAppWithConfig(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if app != nil {
		t.Error("expected nil app for not found")
	}
}

func TestAppService_GetAppWithConfig_Error(t *testing.T) {
	mock := &mockAsyncCaller{}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		return nil, errors.New("timeout")
	}

	svc := NewAppService(mock, Version{})
	_, err := svc.GetAppWithConfig(context.Background(), "my-app")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAppService_GetAppWithConfig_ParseError(t *testing.T) {
	mock := &mockAsyncCaller{}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		return json.RawMessage(`not json`), nil
	}

	svc := NewAppService(mock, Version{})
	_, err := svc.GetAppWithConfig(context.Background(), "my-app")
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestAppService_UpdateApp(t *testing.T) {
	mock := &mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "app.update" {
				t.Errorf("expected method app.update, got %s", method)
			}
			slice, ok := params.([]any)
			if !ok {
				t.Fatal("expected []any params for update")
			}
			if len(slice) != 2 {
				t.Fatalf("expected 2 elements, got %d", len(slice))
			}
			name, ok := slice[0].(string)
			if !ok || name != "my-app" {
				t.Errorf("expected name my-app, got %v", slice[0])
			}
			p, ok := slice[1].(map[string]any)
			if !ok {
				t.Fatal("expected map[string]any for update params")
			}
			if p["custom_compose_config_string"] != "version: '3.1'" {
				t.Errorf("expected custom_compose_config_string, got %v", p["custom_compose_config_string"])
			}
			return nil, nil
		},
	}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		return sampleAppJSON(), nil
	}

	svc := NewAppService(mock, Version{})
	app, err := svc.UpdateApp(context.Background(), "my-app", UpdateAppOpts{
		CustomComposeConfig: "version: '3.1'",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if app == nil {
		t.Fatal("expected non-nil app")
	}
	if app.Name != "my-app" {
		t.Errorf("expected name my-app, got %s", app.Name)
	}
}

func TestAppService_UpdateApp_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("not found")
		},
	}

	svc := NewAppService(mock, Version{})
	_, err := svc.UpdateApp(context.Background(), "my-app", UpdateAppOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAppService_UpdateApp_NotFoundAfterUpdate(t *testing.T) {
	mock := &mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, nil
		},
	}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		return json.RawMessage(`[]`), nil
	}

	svc := NewAppService(mock, Version{})
	app, err := svc.UpdateApp(context.Background(), "ghost-app", UpdateAppOpts{})
	if err == nil {
		t.Fatal("expected error for not found after update")
	}
	if app != nil {
		t.Error("expected nil app")
	}
}

func TestAppService_UpdateApp_EmptyOpts(t *testing.T) {
	mock := &mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			slice := params.([]any)
			p := slice[1].(map[string]any)
			if _, ok := p["custom_compose_config_string"]; ok {
				t.Error("expected no custom_compose_config_string for empty opts")
			}
			return nil, nil
		},
	}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		return sampleAppJSON(), nil
	}

	svc := NewAppService(mock, Version{})
	app, err := svc.UpdateApp(context.Background(), "my-app", UpdateAppOpts{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if app == nil {
		t.Fatal("expected non-nil app")
	}
}

func TestAppService_ListApps(t *testing.T) {
	mock := &mockAsyncCaller{}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		if method != "app.query" {
			t.Errorf("expected method app.query, got %s", method)
		}
		if params != nil {
			t.Error("expected nil params for ListApps")
		}
		return json.RawMessage(`[
			{"name": "app1", "state": "RUNNING", "custom_app": false},
			{"name": "app2", "state": "STOPPED", "custom_app": true}
		]`), nil
	}

	svc := NewAppService(mock, Version{})
	apps, err := svc.ListApps(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(apps) != 2 {
		t.Fatalf("expected 2 apps, got %d", len(apps))
	}
	if apps[0].Name != "app1" {
		t.Errorf("expected first app name app1, got %s", apps[0].Name)
	}
	if apps[0].State != "RUNNING" {
		t.Errorf("expected first app state RUNNING, got %s", apps[0].State)
	}
	if apps[1].Name != "app2" {
		t.Errorf("expected second app name app2, got %s", apps[1].Name)
	}
	if apps[1].State != "STOPPED" {
		t.Errorf("expected second app state STOPPED, got %s", apps[1].State)
	}
	if !apps[1].CustomApp {
		t.Error("expected second app custom_app true")
	}
}

func TestAppService_ListApps_Empty(t *testing.T) {
	mock := &mockAsyncCaller{}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		return json.RawMessage(`[]`), nil
	}

	svc := NewAppService(mock, Version{})
	apps, err := svc.ListApps(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(apps) != 0 {
		t.Errorf("expected 0 apps, got %d", len(apps))
	}
}

func TestAppService_ListApps_Error(t *testing.T) {
	mock := &mockAsyncCaller{}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		return nil, errors.New("network error")
	}

	svc := NewAppService(mock, Version{})
	_, err := svc.ListApps(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAppService_StartApp(t *testing.T) {
	mock := &mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "app.start" {
				t.Errorf("expected method app.start, got %s", method)
			}
			name, ok := params.(string)
			if !ok || name != "my-app" {
				t.Errorf("expected name my-app, got %v", params)
			}
			return nil, nil
		},
	}

	svc := NewAppService(mock, Version{})
	err := svc.StartApp(context.Background(), "my-app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAppService_StartApp_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("app does not exist")
		},
	}

	svc := NewAppService(mock, Version{})
	err := svc.StartApp(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAppService_StopApp(t *testing.T) {
	mock := &mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "app.stop" {
				t.Errorf("expected method app.stop, got %s", method)
			}
			name, ok := params.(string)
			if !ok || name != "my-app" {
				t.Errorf("expected name my-app, got %v", params)
			}
			return nil, nil
		},
	}

	svc := NewAppService(mock, Version{})
	err := svc.StopApp(context.Background(), "my-app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAppService_StopApp_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("app does not exist")
		},
	}

	svc := NewAppService(mock, Version{})
	err := svc.StopApp(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAppService_DeleteApp(t *testing.T) {
	mock := &mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "app.delete" {
				t.Errorf("expected method app.delete, got %s", method)
			}
			name, ok := params.(string)
			if !ok || name != "my-app" {
				t.Errorf("expected name my-app, got %v", params)
			}
			return nil, nil
		},
	}

	svc := NewAppService(mock, Version{})
	err := svc.DeleteApp(context.Background(), "my-app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAppService_DeleteApp_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("permission denied")
		},
	}

	svc := NewAppService(mock, Version{})
	err := svc.DeleteApp(context.Background(), "my-app")
	if err == nil {
		t.Fatal("expected error")
	}
}

// --- Registry CRUD tests ---

func TestAppService_CreateRegistry(t *testing.T) {
	callCount := 0
	mock := &mockAsyncCaller{}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		callCount++
		if callCount == 1 {
			if method != "app.registry.create" {
				t.Errorf("expected method app.registry.create, got %s", method)
			}
			p := params.(map[string]any)
			if p["name"] != "my-registry" {
				t.Errorf("expected name my-registry, got %v", p["name"])
			}
			if p["description"] != "A test registry" {
				t.Errorf("expected description 'A test registry', got %v", p["description"])
			}
			if p["username"] != "admin" {
				t.Errorf("expected username admin, got %v", p["username"])
			}
			if p["password"] != "secret" {
				t.Errorf("expected password secret, got %v", p["password"])
			}
			if p["uri"] != "https://registry.example.com" {
				t.Errorf("expected uri https://registry.example.com, got %v", p["uri"])
			}
			return json.RawMessage(`{"id": 1}`), nil
		}
		// Re-query
		return sampleRegistryJSON(), nil
	}

	svc := NewAppService(mock, Version{})
	reg, err := svc.CreateRegistry(context.Background(), CreateRegistryOpts{
		Name:        "my-registry",
		Description: "A test registry",
		Username:    "admin",
		Password:    "secret",
		URI:         "https://registry.example.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reg == nil {
		t.Fatal("expected non-nil registry")
	}
	if reg.ID != 1 {
		t.Errorf("expected ID 1, got %d", reg.ID)
	}
	if reg.Name != "my-registry" {
		t.Errorf("expected name my-registry, got %s", reg.Name)
	}
	if reg.Description != "A test registry" {
		t.Errorf("expected description 'A test registry', got %s", reg.Description)
	}
	if reg.Username != "admin" {
		t.Errorf("expected username admin, got %s", reg.Username)
	}
	if reg.Password != "secret" {
		t.Errorf("expected password secret, got %s", reg.Password)
	}
	if reg.URI != "https://registry.example.com" {
		t.Errorf("expected uri https://registry.example.com, got %s", reg.URI)
	}
}

func TestAppService_CreateRegistry_Error(t *testing.T) {
	mock := &mockAsyncCaller{}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		return nil, errors.New("connection refused")
	}

	svc := NewAppService(mock, Version{})
	reg, err := svc.CreateRegistry(context.Background(), CreateRegistryOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
	if reg != nil {
		t.Error("expected nil registry on error")
	}
}

func TestAppService_CreateRegistry_ParseError(t *testing.T) {
	mock := &mockAsyncCaller{}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		return json.RawMessage(`not json`), nil
	}

	svc := NewAppService(mock, Version{})
	_, err := svc.CreateRegistry(context.Background(), CreateRegistryOpts{})
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestAppService_CreateRegistry_NotFoundAfterCreate(t *testing.T) {
	callCount := 0
	mock := &mockAsyncCaller{}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		callCount++
		if callCount == 1 {
			return json.RawMessage(`{"id": 99}`), nil
		}
		// Re-query returns empty
		return json.RawMessage(`[]`), nil
	}

	svc := NewAppService(mock, Version{})
	reg, err := svc.CreateRegistry(context.Background(), CreateRegistryOpts{})
	if err == nil {
		t.Fatal("expected error for not found after create")
	}
	if !strings.Contains(err.Error(), "not found after create") {
		t.Errorf("expected error to contain 'not found after create', got %q", err.Error())
	}
	if reg != nil {
		t.Error("expected nil registry for not found after create")
	}
}

func TestAppService_CreateRegistry_NullDescription(t *testing.T) {
	callCount := 0
	mock := &mockAsyncCaller{}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		callCount++
		if callCount == 1 {
			return json.RawMessage(`{"id": 2}`), nil
		}
		return sampleRegistryNullDescJSON(), nil
	}

	svc := NewAppService(mock, Version{})
	reg, err := svc.CreateRegistry(context.Background(), CreateRegistryOpts{
		Name:     "null-desc-registry",
		Username: "user",
		Password: "pass",
		URI:      "https://registry2.example.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reg == nil {
		t.Fatal("expected non-nil registry")
	}
	if reg.Description != "" {
		t.Errorf("expected empty description for null, got %q", reg.Description)
	}
}

func TestAppService_GetRegistry(t *testing.T) {
	mock := &mockAsyncCaller{}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		if method != "app.registry.query" {
			t.Errorf("expected method app.registry.query, got %s", method)
		}
		filter, ok := params.([][]any)
		if !ok {
			t.Fatal("expected [][]any params for GetRegistry")
		}
		if len(filter) != 1 || filter[0][0] != "id" || filter[0][1] != "=" {
			t.Errorf("unexpected filter: %v", filter)
		}
		return sampleRegistryJSON(), nil
	}

	svc := NewAppService(mock, Version{})
	reg, err := svc.GetRegistry(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reg == nil {
		t.Fatal("expected non-nil registry")
	}
	if reg.ID != 1 {
		t.Errorf("expected ID 1, got %d", reg.ID)
	}
	if reg.Name != "my-registry" {
		t.Errorf("expected name my-registry, got %s", reg.Name)
	}
	if reg.Description != "A test registry" {
		t.Errorf("expected description 'A test registry', got %s", reg.Description)
	}
}

func TestAppService_GetRegistry_NotFound(t *testing.T) {
	mock := &mockAsyncCaller{}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		return json.RawMessage(`[]`), nil
	}

	svc := NewAppService(mock, Version{})
	reg, err := svc.GetRegistry(context.Background(), 999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reg != nil {
		t.Error("expected nil registry for not found")
	}
}

func TestAppService_GetRegistry_Error(t *testing.T) {
	mock := &mockAsyncCaller{}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		return nil, errors.New("timeout")
	}

	svc := NewAppService(mock, Version{})
	_, err := svc.GetRegistry(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAppService_GetRegistry_NullDescription(t *testing.T) {
	mock := &mockAsyncCaller{}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		return sampleRegistryNullDescJSON(), nil
	}

	svc := NewAppService(mock, Version{})
	reg, err := svc.GetRegistry(context.Background(), 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reg == nil {
		t.Fatal("expected non-nil registry")
	}
	if reg.Description != "" {
		t.Errorf("expected empty description for null, got %q", reg.Description)
	}
}

func TestAppService_ListRegistries(t *testing.T) {
	mock := &mockAsyncCaller{}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		if method != "app.registry.query" {
			t.Errorf("expected method app.registry.query, got %s", method)
		}
		if params != nil {
			t.Error("expected nil params for ListRegistries")
		}
		return json.RawMessage(`[
			{"id": 1, "name": "reg1", "description": "First", "username": "u1", "password": "p1", "uri": "https://r1.example.com"},
			{"id": 2, "name": "reg2", "description": null, "username": "u2", "password": "p2", "uri": "https://r2.example.com"}
		]`), nil
	}

	svc := NewAppService(mock, Version{})
	registries, err := svc.ListRegistries(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(registries) != 2 {
		t.Fatalf("expected 2 registries, got %d", len(registries))
	}
	if registries[0].ID != 1 {
		t.Errorf("expected first registry ID 1, got %d", registries[0].ID)
	}
	if registries[0].Name != "reg1" {
		t.Errorf("expected first registry name reg1, got %s", registries[0].Name)
	}
	if registries[0].Description != "First" {
		t.Errorf("expected first registry description 'First', got %s", registries[0].Description)
	}
	if registries[1].Description != "" {
		t.Errorf("expected second registry description empty for null, got %s", registries[1].Description)
	}
}

func TestAppService_ListRegistries_Empty(t *testing.T) {
	mock := &mockAsyncCaller{}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		return json.RawMessage(`[]`), nil
	}

	svc := NewAppService(mock, Version{})
	registries, err := svc.ListRegistries(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(registries) != 0 {
		t.Errorf("expected 0 registries, got %d", len(registries))
	}
}

func TestAppService_ListRegistries_Error(t *testing.T) {
	mock := &mockAsyncCaller{}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		return nil, errors.New("network error")
	}

	svc := NewAppService(mock, Version{})
	_, err := svc.ListRegistries(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAppService_UpdateRegistry(t *testing.T) {
	callCount := 0
	mock := &mockAsyncCaller{}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		callCount++
		if callCount == 1 {
			if method != "app.registry.update" {
				t.Errorf("expected method app.registry.update, got %s", method)
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
			p, ok := slice[1].(map[string]any)
			if !ok {
				t.Fatal("expected map[string]any for update params")
			}
			if p["name"] != "updated-registry" {
				t.Errorf("expected name updated-registry, got %v", p["name"])
			}
			return json.RawMessage(`{"id": 1}`), nil
		}
		// Re-query
		return sampleRegistryJSON(), nil
	}

	svc := NewAppService(mock, Version{})
	reg, err := svc.UpdateRegistry(context.Background(), 1, UpdateRegistryOpts{
		Name:        "updated-registry",
		Description: "Updated description",
		Username:    "newuser",
		Password:    "newpass",
		URI:         "https://new.example.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reg == nil {
		t.Fatal("expected non-nil registry")
	}
	if reg.ID != 1 {
		t.Errorf("expected ID 1, got %d", reg.ID)
	}
}

func TestAppService_UpdateRegistry_Error(t *testing.T) {
	mock := &mockAsyncCaller{}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		return nil, errors.New("not found")
	}

	svc := NewAppService(mock, Version{})
	_, err := svc.UpdateRegistry(context.Background(), 999, UpdateRegistryOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAppService_UpdateRegistry_NotFoundAfterUpdate(t *testing.T) {
	callCount := 0
	mock := &mockAsyncCaller{}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		callCount++
		if callCount == 1 {
			return json.RawMessage(`{"id": 1}`), nil
		}
		// Re-query returns empty
		return json.RawMessage(`[]`), nil
	}

	svc := NewAppService(mock, Version{})
	reg, err := svc.UpdateRegistry(context.Background(), 1, UpdateRegistryOpts{})
	if err == nil {
		t.Fatal("expected error for not found after update")
	}
	if !strings.Contains(err.Error(), "not found after update") {
		t.Errorf("expected error to contain 'not found after update', got %q", err.Error())
	}
	if reg != nil {
		t.Error("expected nil registry for not found after update")
	}
}

func TestAppService_DeleteRegistry(t *testing.T) {
	mock := &mockAsyncCaller{}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		if method != "app.registry.delete" {
			t.Errorf("expected method app.registry.delete, got %s", method)
		}
		id, ok := params.(int64)
		if !ok || id != 5 {
			t.Errorf("expected id 5, got %v", params)
		}
		return nil, nil
	}

	svc := NewAppService(mock, Version{})
	err := svc.DeleteRegistry(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAppService_DeleteRegistry_Error(t *testing.T) {
	mock := &mockAsyncCaller{}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		return nil, errors.New("permission denied")
	}

	svc := NewAppService(mock, Version{})
	err := svc.DeleteRegistry(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

// --- Conversion tests ---

func TestAppFromResponse(t *testing.T) {
	resp := AppResponse{
		Name:      "test-app",
		State:     "STOPPED",
		CustomApp: true,
		Config:    map[string]any{"key": "value"},
	}

	app := appFromResponse(resp)

	if app.Name != "test-app" {
		t.Errorf("expected name test-app, got %s", app.Name)
	}
	if app.State != "STOPPED" {
		t.Errorf("expected state STOPPED, got %s", app.State)
	}
	if !app.CustomApp {
		t.Error("expected custom_app true")
	}
	if app.Config["key"] != "value" {
		t.Errorf("expected config key=value, got %v", app.Config["key"])
	}
}

func TestAppFromResponse_NilConfig(t *testing.T) {
	resp := AppResponse{
		Name:      "no-config",
		State:     "RUNNING",
		CustomApp: false,
	}

	app := appFromResponse(resp)

	if app.Name != "no-config" {
		t.Errorf("expected name no-config, got %s", app.Name)
	}
	if app.Config != nil {
		t.Errorf("expected nil config, got %v", app.Config)
	}
}

func TestRegistryFromResponse(t *testing.T) {
	desc := "A registry"
	resp := AppRegistryResponse{
		ID:          1,
		Name:        "test-reg",
		Description: &desc,
		Username:    "user",
		Password:    "pass",
		URI:         "https://example.com",
	}

	reg := registryFromResponse(resp)

	if reg.ID != 1 {
		t.Errorf("expected ID 1, got %d", reg.ID)
	}
	if reg.Name != "test-reg" {
		t.Errorf("expected name test-reg, got %s", reg.Name)
	}
	if reg.Description != "A registry" {
		t.Errorf("expected description 'A registry', got %s", reg.Description)
	}
	if reg.Username != "user" {
		t.Errorf("expected username user, got %s", reg.Username)
	}
	if reg.Password != "pass" {
		t.Errorf("expected password pass, got %s", reg.Password)
	}
	if reg.URI != "https://example.com" {
		t.Errorf("expected uri https://example.com, got %s", reg.URI)
	}
}

func TestRegistryFromResponse_NilDescription(t *testing.T) {
	resp := AppRegistryResponse{
		ID:          2,
		Name:        "null-desc",
		Description: nil,
		Username:    "user",
		Password:    "pass",
		URI:         "https://example.com",
	}

	reg := registryFromResponse(resp)

	if reg.Description != "" {
		t.Errorf("expected empty description for nil, got %q", reg.Description)
	}
}

// --- Param builder tests ---

func TestCreateAppParams(t *testing.T) {
	opts := CreateAppOpts{
		Name:                "my-app",
		CustomApp:           true,
		CustomComposeConfig: "version: '3'",
	}

	params := createAppParams(opts)

	if params["app_name"] != "my-app" {
		t.Errorf("expected app_name my-app, got %v", params["app_name"])
	}
	if params["custom_app"] != true {
		t.Errorf("expected custom_app true, got %v", params["custom_app"])
	}
	if params["custom_compose_config_string"] != "version: '3'" {
		t.Errorf("expected custom_compose_config_string, got %v", params["custom_compose_config_string"])
	}
}

func TestCreateAppParams_NoCompose(t *testing.T) {
	opts := CreateAppOpts{
		Name:      "simple-app",
		CustomApp: false,
	}

	params := createAppParams(opts)

	if params["app_name"] != "simple-app" {
		t.Errorf("expected app_name simple-app, got %v", params["app_name"])
	}
	if params["custom_app"] != false {
		t.Errorf("expected custom_app false, got %v", params["custom_app"])
	}
	if _, ok := params["custom_compose_config_string"]; ok {
		t.Error("expected no custom_compose_config_string for empty config")
	}
}

func TestUpdateAppParams(t *testing.T) {
	opts := UpdateAppOpts{
		CustomComposeConfig: "version: '3.1'",
	}

	params := updateAppParams(opts)

	if params["custom_compose_config_string"] != "version: '3.1'" {
		t.Errorf("expected custom_compose_config_string, got %v", params["custom_compose_config_string"])
	}
}

func TestUpdateAppParams_Empty(t *testing.T) {
	opts := UpdateAppOpts{}

	params := updateAppParams(opts)

	if _, ok := params["custom_compose_config_string"]; ok {
		t.Error("expected no custom_compose_config_string for empty opts")
	}
	if len(params) != 0 {
		t.Errorf("expected empty params map, got %v", params)
	}
}

func TestRegistryParams(t *testing.T) {
	opts := CreateRegistryOpts{
		Name:        "my-reg",
		Description: "A description",
		Username:    "admin",
		Password:    "secret",
		URI:         "https://example.com",
	}

	params := registryParams(opts)

	if params["name"] != "my-reg" {
		t.Errorf("expected name my-reg, got %v", params["name"])
	}
	if params["description"] != "A description" {
		t.Errorf("expected description 'A description', got %v", params["description"])
	}
	if params["username"] != "admin" {
		t.Errorf("expected username admin, got %v", params["username"])
	}
	if params["password"] != "secret" {
		t.Errorf("expected password secret, got %v", params["password"])
	}
	if params["uri"] != "https://example.com" {
		t.Errorf("expected uri https://example.com, got %v", params["uri"])
	}
}

func TestRegistryParams_EmptyDescription(t *testing.T) {
	opts := CreateRegistryOpts{
		Name:     "no-desc",
		Username: "user",
		Password: "pass",
		URI:      "https://example.com",
	}

	params := registryParams(opts)

	if params["description"] != nil {
		t.Errorf("expected nil description for empty string, got %v", params["description"])
	}
}

func TestAppService_ListApps_ParseError(t *testing.T) {
	mock := &mockAsyncCaller{}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		return json.RawMessage(`not json`), nil
	}

	svc := NewAppService(mock, Version{})
	_, err := svc.ListApps(context.Background())
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestAppService_ListRegistries_ParseError(t *testing.T) {
	mock := &mockAsyncCaller{}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		return json.RawMessage(`not json`), nil
	}

	svc := NewAppService(mock, Version{})
	_, err := svc.ListRegistries(context.Background())
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestAppService_GetRegistry_ParseError(t *testing.T) {
	mock := &mockAsyncCaller{}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		return json.RawMessage(`not json`), nil
	}

	svc := NewAppService(mock, Version{})
	_, err := svc.GetRegistry(context.Background(), 1)
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestAppService_CreateApp_ReReadError(t *testing.T) {
	mock := &mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, nil
		},
	}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		return nil, errors.New("re-read failed")
	}

	svc := NewAppService(mock, Version{})
	app, err := svc.CreateApp(context.Background(), CreateAppOpts{Name: "fail-reread"})
	if err == nil {
		t.Fatal("expected error")
	}
	if app != nil {
		t.Error("expected nil app on re-read error")
	}
}

func TestAppService_UpdateApp_ReReadError(t *testing.T) {
	mock := &mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, nil
		},
	}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		return nil, errors.New("re-read failed")
	}

	svc := NewAppService(mock, Version{})
	app, err := svc.UpdateApp(context.Background(), "my-app", UpdateAppOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
	if app != nil {
		t.Error("expected nil app on re-read error")
	}
}

func TestAppService_CreateRegistry_ReReadError(t *testing.T) {
	callCount := 0
	mock := &mockAsyncCaller{}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		callCount++
		if callCount == 1 {
			return json.RawMessage(`{"id": 1}`), nil
		}
		return nil, errors.New("re-read failed")
	}

	svc := NewAppService(mock, Version{})
	reg, err := svc.CreateRegistry(context.Background(), CreateRegistryOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
	if reg != nil {
		t.Error("expected nil registry on re-read error")
	}
}

func TestAppService_UpdateRegistry_ReReadError(t *testing.T) {
	callCount := 0
	mock := &mockAsyncCaller{}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		callCount++
		if callCount == 1 {
			return json.RawMessage(`{"id": 1}`), nil
		}
		return nil, errors.New("re-read failed")
	}

	svc := NewAppService(mock, Version{})
	reg, err := svc.UpdateRegistry(context.Background(), 1, UpdateRegistryOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
	if reg != nil {
		t.Error("expected nil registry on re-read error")
	}
}

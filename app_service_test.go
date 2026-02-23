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
	mock := &mockSubscribeCaller{mockAsyncCaller: mockAsyncCaller{
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
	}}
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
	mock := &mockSubscribeCaller{mockAsyncCaller: mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("connection refused")
		},
	}}

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
	mock := &mockSubscribeCaller{mockAsyncCaller: mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, nil
		},
	}}
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
	mock := &mockSubscribeCaller{mockAsyncCaller: mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, nil
		},
	}}
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
	mock := &mockSubscribeCaller{}
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
	mock := &mockSubscribeCaller{}
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
	mock := &mockSubscribeCaller{}
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
	mock := &mockSubscribeCaller{}
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
	mock := &mockSubscribeCaller{}
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
	mock := &mockSubscribeCaller{}
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
	mock := &mockSubscribeCaller{}
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
	mock := &mockSubscribeCaller{}
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
	mock := &mockSubscribeCaller{mockAsyncCaller: mockAsyncCaller{
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
	}}
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
	mock := &mockSubscribeCaller{mockAsyncCaller: mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("not found")
		},
	}}

	svc := NewAppService(mock, Version{})
	_, err := svc.UpdateApp(context.Background(), "my-app", UpdateAppOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAppService_UpdateApp_NotFoundAfterUpdate(t *testing.T) {
	mock := &mockSubscribeCaller{mockAsyncCaller: mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, nil
		},
	}}
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
	mock := &mockSubscribeCaller{mockAsyncCaller: mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			slice := params.([]any)
			p := slice[1].(map[string]any)
			if _, ok := p["custom_compose_config_string"]; ok {
				t.Error("expected no custom_compose_config_string for empty opts")
			}
			return nil, nil
		},
	}}
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
	mock := &mockSubscribeCaller{}
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
	mock := &mockSubscribeCaller{}
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
	mock := &mockSubscribeCaller{}
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
	mock := &mockSubscribeCaller{mockAsyncCaller: mockAsyncCaller{
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
	}}

	svc := NewAppService(mock, Version{})
	err := svc.StartApp(context.Background(), "my-app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAppService_StartApp_Error(t *testing.T) {
	mock := &mockSubscribeCaller{mockAsyncCaller: mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("app does not exist")
		},
	}}

	svc := NewAppService(mock, Version{})
	err := svc.StartApp(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAppService_StopApp(t *testing.T) {
	mock := &mockSubscribeCaller{mockAsyncCaller: mockAsyncCaller{
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
	}}

	svc := NewAppService(mock, Version{})
	err := svc.StopApp(context.Background(), "my-app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAppService_StopApp_Error(t *testing.T) {
	mock := &mockSubscribeCaller{mockAsyncCaller: mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("app does not exist")
		},
	}}

	svc := NewAppService(mock, Version{})
	err := svc.StopApp(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAppService_DeleteApp(t *testing.T) {
	mock := &mockSubscribeCaller{mockAsyncCaller: mockAsyncCaller{
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
	}}

	svc := NewAppService(mock, Version{})
	err := svc.DeleteApp(context.Background(), "my-app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAppService_DeleteApp_Error(t *testing.T) {
	mock := &mockSubscribeCaller{mockAsyncCaller: mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("permission denied")
		},
	}}

	svc := NewAppService(mock, Version{})
	err := svc.DeleteApp(context.Background(), "my-app")
	if err == nil {
		t.Fatal("expected error")
	}
}

// --- Registry CRUD tests ---

func TestAppService_CreateRegistry(t *testing.T) {
	callCount := 0
	mock := &mockSubscribeCaller{}
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
	mock := &mockSubscribeCaller{}
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
	mock := &mockSubscribeCaller{}
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
	mock := &mockSubscribeCaller{}
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
	mock := &mockSubscribeCaller{}
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
	mock := &mockSubscribeCaller{}
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
	mock := &mockSubscribeCaller{}
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
	mock := &mockSubscribeCaller{}
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
	mock := &mockSubscribeCaller{}
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
	mock := &mockSubscribeCaller{}
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
	mock := &mockSubscribeCaller{}
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
	mock := &mockSubscribeCaller{}
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
	mock := &mockSubscribeCaller{}
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
	mock := &mockSubscribeCaller{}
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
	mock := &mockSubscribeCaller{}
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
	mock := &mockSubscribeCaller{}
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
	mock := &mockSubscribeCaller{}
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

func TestAppFromResponse_WithWorkloads(t *testing.T) {
	resp := AppResponse{
		Name:             "plex",
		State:            "RUNNING",
		Version:          "1.0.0",
		HumanVersion:     "Plex 1.0.0",
		LatestVersion:    "1.1.0",
		UpgradeAvailable: true,
		ActiveWorkloads: AppActiveWorkloadsResponse{
			Containers: 2,
			UsedPorts: []AppUsedPortResponse{
				{ContainerPort: 32400, HostPort: 32400, Protocol: "tcp"},
			},
			ContainerDetails: []AppContainerDetailsResponse{
				{
					ID:          "abc123",
					ServiceName: "plex",
					Image:       "plexinc/pms-docker:latest",
					State:       "running",
				},
			},
		},
	}

	app := appFromResponse(resp)
	if app.Version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", app.Version)
	}
	if app.HumanVersion != "Plex 1.0.0" {
		t.Errorf("expected human version 'Plex 1.0.0', got %s", app.HumanVersion)
	}
	if app.LatestVersion != "1.1.0" {
		t.Errorf("expected latest version 1.1.0, got %s", app.LatestVersion)
	}
	if !app.UpgradeAvailable {
		t.Error("expected UpgradeAvailable=true")
	}
	if app.ActiveWorkloads.Containers != 2 {
		t.Errorf("expected 2 containers, got %d", app.ActiveWorkloads.Containers)
	}
	if len(app.ActiveWorkloads.UsedPorts) != 1 {
		t.Fatalf("expected 1 used port, got %d", len(app.ActiveWorkloads.UsedPorts))
	}
	if app.ActiveWorkloads.UsedPorts[0].ContainerPort != 32400 {
		t.Errorf("expected container port 32400, got %d", app.ActiveWorkloads.UsedPorts[0].ContainerPort)
	}
	if len(app.ActiveWorkloads.ContainerDetails) != 1 {
		t.Fatalf("expected 1 container detail, got %d", len(app.ActiveWorkloads.ContainerDetails))
	}
	if app.ActiveWorkloads.ContainerDetails[0].State != ContainerStateRunning {
		t.Errorf("expected running state, got %s", app.ActiveWorkloads.ContainerDetails[0].State)
	}
	if app.ActiveWorkloads.ContainerDetails[0].Image != "plexinc/pms-docker:latest" {
		t.Errorf("expected image plexinc/pms-docker:latest, got %s", app.ActiveWorkloads.ContainerDetails[0].Image)
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
	mock := &mockSubscribeCaller{}
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
	mock := &mockSubscribeCaller{}
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
	mock := &mockSubscribeCaller{}
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
	mock := &mockSubscribeCaller{mockAsyncCaller: mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, nil
		},
	}}
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
	mock := &mockSubscribeCaller{mockAsyncCaller: mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, nil
		},
	}}
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
	mock := &mockSubscribeCaller{}
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
	mock := &mockSubscribeCaller{}
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

// --- UpgradeSummary tests ---

func TestAppService_UpgradeSummary(t *testing.T) {
	mock := &mockSubscribeCaller{}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		if method != "app.upgrade_summary" {
			t.Errorf("expected method app.upgrade_summary, got %s", method)
		}
		slice, ok := params.([]any)
		if !ok {
			t.Fatal("expected []any params")
		}
		if len(slice) != 1 || slice[0] != "plex" {
			t.Errorf("expected [plex], got %v", slice)
		}
		return json.RawMessage(`{
			"latest_version": "1.2.0",
			"latest_human_version": "Plex 1.2.0",
			"upgrade_version": "1.1.0",
			"upgrade_human_version": "Plex 1.1.0",
			"changelog": "Bug fixes and improvements",
			"available_versions_for_upgrade": [
				{"version": "1.1.0", "human_version": "Plex 1.1.0"},
				{"version": "1.2.0", "human_version": "Plex 1.2.0"}
			]
		}`), nil
	}

	svc := NewAppService(mock, Version{})
	summary, err := svc.UpgradeSummary(context.Background(), "plex")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary == nil {
		t.Fatal("expected non-nil summary")
	}
	if summary.LatestVersion != "1.2.0" {
		t.Errorf("expected latest version 1.2.0, got %s", summary.LatestVersion)
	}
	if summary.LatestHumanVersion != "Plex 1.2.0" {
		t.Errorf("expected latest human version 'Plex 1.2.0', got %s", summary.LatestHumanVersion)
	}
	if summary.UpgradeVersion != "1.1.0" {
		t.Errorf("expected upgrade version 1.1.0, got %s", summary.UpgradeVersion)
	}
	if summary.UpgradeHumanVersion != "Plex 1.1.0" {
		t.Errorf("expected upgrade human version 'Plex 1.1.0', got %s", summary.UpgradeHumanVersion)
	}
	if summary.Changelog != "Bug fixes and improvements" {
		t.Errorf("expected changelog 'Bug fixes and improvements', got %s", summary.Changelog)
	}
	if len(summary.AvailableVersions) != 2 {
		t.Fatalf("expected 2 available versions, got %d", len(summary.AvailableVersions))
	}
	if summary.AvailableVersions[0].Version != "1.1.0" {
		t.Errorf("expected first version 1.1.0, got %s", summary.AvailableVersions[0].Version)
	}
	if summary.AvailableVersions[1].Version != "1.2.0" {
		t.Errorf("expected second version 1.2.0, got %s", summary.AvailableVersions[1].Version)
	}
}

func TestAppService_UpgradeSummary_Error(t *testing.T) {
	mock := &mockSubscribeCaller{}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		return nil, errors.New("app not found")
	}

	svc := NewAppService(mock, Version{})
	summary, err := svc.UpgradeSummary(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
	if summary != nil {
		t.Error("expected nil summary on error")
	}
}

// --- ListImages tests ---

func TestAppService_ListImages(t *testing.T) {
	mock := &mockSubscribeCaller{}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		if method != "app.image.query" {
			t.Errorf("expected method app.image.query, got %s", method)
		}
		if params != nil {
			t.Error("expected nil params for ListImages")
		}
		return json.RawMessage(`[
			{
				"id": "sha256:abc123",
				"repo_tags": ["nginx:latest", "nginx:1.25"],
				"size": 187654321,
				"created": "2024-01-15T10:30:00Z",
				"dangling": false
			},
			{
				"id": "sha256:def456",
				"repo_tags": [],
				"size": 52345678,
				"created": "2024-01-10T08:00:00Z",
				"dangling": true
			}
		]`), nil
	}

	svc := NewAppService(mock, Version{})
	images, err := svc.ListImages(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(images) != 2 {
		t.Fatalf("expected 2 images, got %d", len(images))
	}
	if images[0].ID != "sha256:abc123" {
		t.Errorf("expected first image ID sha256:abc123, got %s", images[0].ID)
	}
	if len(images[0].RepoTags) != 2 {
		t.Fatalf("expected 2 repo tags, got %d", len(images[0].RepoTags))
	}
	if images[0].RepoTags[0] != "nginx:latest" {
		t.Errorf("expected first tag nginx:latest, got %s", images[0].RepoTags[0])
	}
	if images[0].Size != 187654321 {
		t.Errorf("expected size 187654321, got %d", images[0].Size)
	}
	if images[0].Created != "2024-01-15T10:30:00Z" {
		t.Errorf("expected created 2024-01-15T10:30:00Z, got %s", images[0].Created)
	}
	if images[0].Dangling {
		t.Error("expected first image not dangling")
	}
	if !images[1].Dangling {
		t.Error("expected second image dangling")
	}
	if len(images[1].RepoTags) != 0 {
		t.Errorf("expected empty repo tags for dangling image, got %v", images[1].RepoTags)
	}
}

func TestAppService_ListImages_Empty(t *testing.T) {
	mock := &mockSubscribeCaller{}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		return json.RawMessage(`[]`), nil
	}

	svc := NewAppService(mock, Version{})
	images, err := svc.ListImages(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(images) != 0 {
		t.Errorf("expected 0 images, got %d", len(images))
	}
}

func TestAppService_ListImages_Error(t *testing.T) {
	mock := &mockSubscribeCaller{}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		return nil, errors.New("network error")
	}

	svc := NewAppService(mock, Version{})
	_, err := svc.ListImages(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

// --- AvailableSpace tests ---

func TestAppService_AvailableSpace(t *testing.T) {
	mock := &mockSubscribeCaller{}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		if method != "app.available_space" {
			t.Errorf("expected method app.available_space, got %s", method)
		}
		if params != nil {
			t.Error("expected nil params for AvailableSpace")
		}
		return json.RawMessage(`1099511627776`), nil
	}

	svc := NewAppService(mock, Version{})
	space, err := svc.AvailableSpace(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if space != 1099511627776 {
		t.Errorf("expected 1099511627776 bytes, got %d", space)
	}
}

func TestAppService_AvailableSpace_Error(t *testing.T) {
	mock := &mockSubscribeCaller{}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		return nil, errors.New("pool not found")
	}

	svc := NewAppService(mock, Version{})
	space, err := svc.AvailableSpace(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if space != 0 {
		t.Errorf("expected 0 on error, got %d", space)
	}
}

// --- UpgradeApp tests ---

func TestAppService_UpgradeApp(t *testing.T) {
	mock := &mockSubscribeCaller{mockAsyncCaller: mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "app.upgrade" {
				t.Errorf("expected method app.upgrade, got %s", method)
			}
			slice, ok := params.([]any)
			if !ok {
				t.Fatal("expected []any params")
			}
			if len(slice) != 1 || slice[0] != "plex" {
				t.Errorf("expected [plex], got %v", slice)
			}
			return nil, nil
		},
	}}

	svc := NewAppService(mock, Version{})
	err := svc.UpgradeApp(context.Background(), "plex")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAppService_UpgradeApp_Error(t *testing.T) {
	mock := &mockSubscribeCaller{mockAsyncCaller: mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("upgrade failed")
		},
	}}

	svc := NewAppService(mock, Version{})
	err := svc.UpgradeApp(context.Background(), "plex")
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "upgrade failed" {
		t.Errorf("expected 'upgrade failed', got %q", err.Error())
	}
}

// --- RedeployApp tests ---

func TestAppService_RedeployApp(t *testing.T) {
	mock := &mockSubscribeCaller{mockAsyncCaller: mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "app.redeploy" {
				t.Errorf("expected method app.redeploy, got %s", method)
			}
			name, ok := params.(string)
			if !ok || name != "plex" {
				t.Errorf("expected name plex, got %v", params)
			}
			return nil, nil
		},
	}}

	svc := NewAppService(mock, Version{})
	err := svc.RedeployApp(context.Background(), "plex")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAppService_RedeployApp_Error(t *testing.T) {
	mock := &mockSubscribeCaller{mockAsyncCaller: mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("redeploy failed")
		},
	}}

	svc := NewAppService(mock, Version{})
	err := svc.RedeployApp(context.Background(), "plex")
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "redeploy failed" {
		t.Errorf("expected 'redeploy failed', got %q", err.Error())
	}
}

// --- Conversion function tests ---

func TestAppUpgradeSummaryFromResponse(t *testing.T) {
	changelog := "Major update"
	resp := AppUpgradeSummaryResponse{
		LatestVersion:       "2.0.0",
		LatestHumanVersion:  "App 2.0.0",
		UpgradeVersion:      "1.5.0",
		UpgradeHumanVersion: "App 1.5.0",
		Changelog:           &changelog,
		AvailableVersions: []AppAvailableVersionResponse{
			{Version: "1.5.0", HumanVersion: "App 1.5.0"},
			{Version: "2.0.0", HumanVersion: "App 2.0.0"},
		},
	}

	summary := appUpgradeSummaryFromResponse(resp)

	if summary.LatestVersion != "2.0.0" {
		t.Errorf("expected latest version 2.0.0, got %s", summary.LatestVersion)
	}
	if summary.LatestHumanVersion != "App 2.0.0" {
		t.Errorf("expected latest human version 'App 2.0.0', got %s", summary.LatestHumanVersion)
	}
	if summary.UpgradeVersion != "1.5.0" {
		t.Errorf("expected upgrade version 1.5.0, got %s", summary.UpgradeVersion)
	}
	if summary.UpgradeHumanVersion != "App 1.5.0" {
		t.Errorf("expected upgrade human version 'App 1.5.0', got %s", summary.UpgradeHumanVersion)
	}
	if summary.Changelog != "Major update" {
		t.Errorf("expected changelog 'Major update', got %s", summary.Changelog)
	}
	if len(summary.AvailableVersions) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(summary.AvailableVersions))
	}
	if summary.AvailableVersions[0].Version != "1.5.0" {
		t.Errorf("expected first version 1.5.0, got %s", summary.AvailableVersions[0].Version)
	}
}

func TestAppUpgradeSummaryFromResponse_NilChangelog(t *testing.T) {
	resp := AppUpgradeSummaryResponse{
		LatestVersion:       "2.0.0",
		LatestHumanVersion:  "App 2.0.0",
		UpgradeVersion:      "1.5.0",
		UpgradeHumanVersion: "App 1.5.0",
		Changelog:           nil,
		AvailableVersions:   []AppAvailableVersionResponse{},
	}

	summary := appUpgradeSummaryFromResponse(resp)
	if summary.Changelog != "" {
		t.Errorf("expected empty changelog for nil, got %q", summary.Changelog)
	}
}

func TestAppImageFromResponse(t *testing.T) {
	resp := AppImageResponse{
		ID:       "sha256:abc",
		RepoTags: []string{"test:latest"},
		Size:     12345,
		Created:  "2024-06-01T00:00:00Z",
		Dangling: true,
	}

	image := appImageFromResponse(resp)

	if image.ID != "sha256:abc" {
		t.Errorf("expected ID sha256:abc, got %s", image.ID)
	}
	if len(image.RepoTags) != 1 || image.RepoTags[0] != "test:latest" {
		t.Errorf("expected repo tags [test:latest], got %v", image.RepoTags)
	}
	if image.Size != 12345 {
		t.Errorf("expected size 12345, got %d", image.Size)
	}
	if image.Created != "2024-06-01T00:00:00Z" {
		t.Errorf("expected created 2024-06-01T00:00:00Z, got %s", image.Created)
	}
	if !image.Dangling {
		t.Error("expected dangling true")
	}
}

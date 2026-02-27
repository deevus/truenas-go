package truenas

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

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

func TestAppService_CreateApp_CatalogApp(t *testing.T) {
	var capturedMethod string
	var capturedParams any

	mock := &mockSubscribeCaller{mockAsyncCaller: mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			capturedMethod = method
			capturedParams = params
			return nil, nil
		},
	}}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		return sampleCatalogAppJSON(), nil
	}

	svc := NewAppService(mock, Version{})
	app, err := svc.CreateApp(context.Background(), CreateAppOpts{
		Name:       "tailscale",
		CatalogApp: "tailscale",
		Train:      "community",
		Values:     map[string]any{"tailscale": map[string]any{"auth_key": "tskey-xxx"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if app == nil {
		t.Fatal("expected non-nil app")
	}

	// Verify create method
	if capturedMethod != "app.create" {
		t.Errorf("expected method app.create, got %s", capturedMethod)
	}

	// Verify create params include catalog fields
	p, ok := capturedParams.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any params, got %T", capturedParams)
	}
	if p["catalog_app"] != "tailscale" {
		t.Errorf("expected catalog_app 'tailscale', got %v", p["catalog_app"])
	}
	if p["train"] != "community" {
		t.Errorf("expected train 'community', got %v", p["train"])
	}
	if p["custom_app"] != false {
		t.Errorf("expected custom_app false, got %v", p["custom_app"])
	}
	values, ok := p["values"].(map[string]any)
	if !ok {
		t.Fatalf("expected values map, got %T", p["values"])
	}
	ts, ok := values["tailscale"].(map[string]any)
	if !ok {
		t.Fatalf("expected tailscale values map, got %T", values["tailscale"])
	}
	if ts["auth_key"] != "tskey-xxx" {
		t.Errorf("expected auth_key 'tskey-xxx', got %v", ts["auth_key"])
	}

	// Verify re-read populated catalog metadata
	if app.CatalogApp != "tailscale" {
		t.Errorf("expected CatalogApp 'tailscale', got %q", app.CatalogApp)
	}
	if app.Train != "community" {
		t.Errorf("expected Train 'community', got %q", app.Train)
	}
	if app.CustomApp {
		t.Error("expected CustomApp false")
	}
	if app.Version != "1.3.32" {
		t.Errorf("expected Version '1.3.32', got %q", app.Version)
	}
}

func TestAppService_UpdateApp_CatalogAppValues(t *testing.T) {
	var capturedParams any

	mock := &mockSubscribeCaller{mockAsyncCaller: mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "app.update" {
				t.Errorf("expected method app.update, got %s", method)
			}
			capturedParams = params
			return nil, nil
		},
	}}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		return sampleCatalogAppJSON(), nil
	}

	svc := NewAppService(mock, Version{})
	app, err := svc.UpdateApp(context.Background(), "tailscale", UpdateAppOpts{
		Values: map[string]any{"TZ": "America/New_York"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if app == nil {
		t.Fatal("expected non-nil app")
	}

	// Verify update params contain values
	slice, ok := capturedParams.([]any)
	if !ok || len(slice) != 2 {
		t.Fatalf("expected [name, params] slice, got %T", capturedParams)
	}
	if slice[0] != "tailscale" {
		t.Errorf("expected name 'tailscale', got %v", slice[0])
	}
	updateMap, ok := slice[1].(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any update params, got %T", slice[1])
	}
	values, ok := updateMap["values"].(map[string]any)
	if !ok {
		t.Fatalf("expected values map in update params, got %v", updateMap)
	}
	if values["TZ"] != "America/New_York" {
		t.Errorf("expected TZ 'America/New_York', got %v", values["TZ"])
	}
}

package truenas

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

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

package truenas

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestCloudSyncService_CreateCredential_V25(t *testing.T) {
	callCount := 0
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				callCount++
				if callCount == 1 {
					if method != "cloudsync.credentials.create" {
						t.Errorf("expected method cloudsync.credentials.create, got %s", method)
					}
					// Verify V25 format: provider is object
					p := params.(map[string]any)
					provider, ok := p["provider"].(map[string]any)
					if !ok {
						t.Fatal("expected provider to be map[string]any for V25")
					}
					if provider["type"] != "S3" {
						t.Errorf("expected provider.type S3, got %v", provider["type"])
					}
					if provider["access_key_id"] != "AKIATEST" {
						t.Errorf("expected access_key_id AKIATEST, got %v", provider["access_key_id"])
					}
					return json.RawMessage(`{"id": 1}`), nil
				}
				// Re-read via GetCredential
				return sampleCredentialV25JSON(), nil
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 25, Minor: 4})
	cred, err := svc.CreateCredential(context.Background(), CreateCredentialOpts{
		Name:         "My S3 Cred",
		ProviderType: "S3",
		Attributes: map[string]string{
			"access_key_id":     "AKIATEST",
			"secret_access_key": "secret123",
			"endpoint":          "s3.example.com",
			"region":            "us-east-1",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cred == nil {
		t.Fatal("expected non-nil credential")
	}
	if cred.ID != 1 {
		t.Errorf("expected ID 1, got %d", cred.ID)
	}
	if cred.Name != "My S3 Cred" {
		t.Errorf("expected name 'My S3 Cred', got %q", cred.Name)
	}
	if cred.ProviderType != "S3" {
		t.Errorf("expected provider type S3, got %s", cred.ProviderType)
	}
	if cred.Attributes["access_key_id"] != "AKIATEST" {
		t.Errorf("expected access_key_id AKIATEST, got %s", cred.Attributes["access_key_id"])
	}
	if cred.Attributes["region"] != "us-east-1" {
		t.Errorf("expected region us-east-1, got %s", cred.Attributes["region"])
	}
}

func TestCloudSyncService_CreateCredential_V24(t *testing.T) {
	callCount := 0
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				callCount++
				if callCount == 1 {
					if method != "cloudsync.credentials.create" {
						t.Errorf("expected method cloudsync.credentials.create, got %s", method)
					}
					// Verify V24 format: provider is string, attributes separate
					p := params.(map[string]any)
					provider, ok := p["provider"].(string)
					if !ok {
						t.Fatal("expected provider to be string for V24")
					}
					if provider != "S3" {
						t.Errorf("expected provider S3, got %s", provider)
					}
					attrs, ok := p["attributes"].(map[string]any)
					if !ok {
						t.Fatal("expected attributes to be map[string]any for V24")
					}
					if attrs["access_key_id"] != "AKIATEST" {
						t.Errorf("expected attributes.access_key_id AKIATEST, got %v", attrs["access_key_id"])
					}
					return json.RawMessage(`{"id": 1}`), nil
				}
				return sampleCredentialV24JSON(), nil
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 24, Minor: 10})
	cred, err := svc.CreateCredential(context.Background(), CreateCredentialOpts{
		Name:         "My S3 Cred",
		ProviderType: "S3",
		Attributes: map[string]string{
			"access_key_id":     "AKIATEST",
			"secret_access_key": "secret123",
			"endpoint":          "s3.example.com",
			"region":            "us-east-1",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cred == nil {
		t.Fatal("expected non-nil credential")
	}
	if cred.ID != 1 {
		t.Errorf("expected ID 1, got %d", cred.ID)
	}
	if cred.ProviderType != "S3" {
		t.Errorf("expected provider type S3, got %s", cred.ProviderType)
	}
}

func TestCloudSyncService_CreateCredential_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, errors.New("connection refused")
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 25, Minor: 4})
	cred, err := svc.CreateCredential(context.Background(), CreateCredentialOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
	if cred != nil {
		t.Error("expected nil credential on error")
	}
	if err.Error() != "connection refused" {
		t.Errorf("expected 'connection refused', got %q", err.Error())
	}
}

func TestCloudSyncService_CreateCredential_ParseError(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return json.RawMessage(`not json`), nil
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 25, Minor: 4})
	_, err := svc.CreateCredential(context.Background(), CreateCredentialOpts{})
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestCloudSyncService_GetCredential_V25(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				if method != "cloudsync.credentials.query" {
					t.Errorf("expected method cloudsync.credentials.query, got %s", method)
				}
				return sampleCredentialV25JSON(), nil
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 25, Minor: 4})
	cred, err := svc.GetCredential(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cred == nil {
		t.Fatal("expected non-nil credential")
	}
	if cred.ID != 1 {
		t.Errorf("expected ID 1, got %d", cred.ID)
	}
	if cred.Name != "My S3 Cred" {
		t.Errorf("expected name 'My S3 Cred', got %q", cred.Name)
	}
	if cred.ProviderType != "S3" {
		t.Errorf("expected provider type S3, got %s", cred.ProviderType)
	}
	if cred.Attributes["secret_access_key"] != "secret123" {
		t.Errorf("expected secret_access_key secret123, got %s", cred.Attributes["secret_access_key"])
	}
	if cred.Attributes["endpoint"] != "s3.example.com" {
		t.Errorf("expected endpoint s3.example.com, got %s", cred.Attributes["endpoint"])
	}
}

func TestCloudSyncService_GetCredential_V24(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return sampleCredentialV24JSON(), nil
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 24, Minor: 10})
	cred, err := svc.GetCredential(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cred == nil {
		t.Fatal("expected non-nil credential")
	}
	if cred.ID != 1 {
		t.Errorf("expected ID 1, got %d", cred.ID)
	}
	if cred.ProviderType != "S3" {
		t.Errorf("expected provider type S3, got %s", cred.ProviderType)
	}
	if cred.Attributes["access_key_id"] != "AKIATEST" {
		t.Errorf("expected access_key_id AKIATEST, got %s", cred.Attributes["access_key_id"])
	}
}

func TestCloudSyncService_GetCredential_NotFound(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return json.RawMessage(`[]`), nil
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 25, Minor: 4})
	cred, err := svc.GetCredential(context.Background(), 999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cred != nil {
		t.Error("expected nil credential for not found")
	}
}

func TestCloudSyncService_GetCredential_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, errors.New("timeout")
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 25, Minor: 4})
	_, err := svc.GetCredential(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCloudSyncService_GetCredential_ParseError(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return json.RawMessage(`not json`), nil
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 25, Minor: 4})
	_, err := svc.GetCredential(context.Background(), 1)
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestCloudSyncService_ListCredentials_V25(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				if method != "cloudsync.credentials.query" {
					t.Errorf("expected method cloudsync.credentials.query, got %s", method)
				}
				if params != nil {
					t.Error("expected nil params for ListCredentials")
				}
				return json.RawMessage(`[
					{"id": 1, "name": "S3 Cred", "provider": {"type": "S3", "access_key_id": "key1", "secret_access_key": "sec1"}},
					{"id": 2, "name": "B2 Cred", "provider": {"type": "B2", "account": "acct", "key": "b2key"}}
				]`), nil
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 25, Minor: 4})
	creds, err := svc.ListCredentials(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(creds) != 2 {
		t.Fatalf("expected 2 credentials, got %d", len(creds))
	}
	if creds[0].ProviderType != "S3" {
		t.Errorf("expected first cred provider type S3, got %s", creds[0].ProviderType)
	}
	if creds[1].ProviderType != "B2" {
		t.Errorf("expected second cred provider type B2, got %s", creds[1].ProviderType)
	}
	if creds[1].Attributes["account"] != "acct" {
		t.Errorf("expected second cred account 'acct', got %s", creds[1].Attributes["account"])
	}
}

func TestCloudSyncService_ListCredentials_V24(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return json.RawMessage(`[
					{"id": 1, "name": "S3 Cred", "provider": "S3", "attributes": {"access_key_id": "key1"}},
					{"id": 2, "name": "B2 Cred", "provider": "B2", "attributes": {"account": "acct"}}
				]`), nil
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 24, Minor: 10})
	creds, err := svc.ListCredentials(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(creds) != 2 {
		t.Fatalf("expected 2 credentials, got %d", len(creds))
	}
	if creds[0].Attributes["access_key_id"] != "key1" {
		t.Errorf("expected first cred access_key_id 'key1', got %s", creds[0].Attributes["access_key_id"])
	}
}

func TestCloudSyncService_ListCredentials_Empty(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return json.RawMessage(`[]`), nil
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 25, Minor: 4})
	creds, err := svc.ListCredentials(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(creds) != 0 {
		t.Errorf("expected 0 credentials, got %d", len(creds))
	}
}

func TestCloudSyncService_ListCredentials_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, errors.New("network error")
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 25, Minor: 4})
	_, err := svc.ListCredentials(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCloudSyncService_ListCredentials_ParseError(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return json.RawMessage(`not json`), nil
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 25, Minor: 4})
	_, err := svc.ListCredentials(context.Background())
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestCloudSyncService_UpdateCredential_V25(t *testing.T) {
	callCount := 0
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				callCount++
				if callCount == 1 {
					if method != "cloudsync.credentials.update" {
						t.Errorf("expected method cloudsync.credentials.update, got %s", method)
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
					// Verify V25 format
					p := slice[1].(map[string]any)
					provider, ok := p["provider"].(map[string]any)
					if !ok {
						t.Fatal("expected provider to be map for V25")
					}
					if provider["type"] != "S3" {
						t.Errorf("expected provider.type S3, got %v", provider["type"])
					}
					return json.RawMessage(`{"id": 1}`), nil
				}
				return sampleCredentialV25JSON(), nil
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 25, Minor: 4})
	cred, err := svc.UpdateCredential(context.Background(), 1, UpdateCredentialOpts{
		Name:         "My S3 Cred",
		ProviderType: "S3",
		Attributes: map[string]string{
			"access_key_id":     "AKIATEST",
			"secret_access_key": "secret123",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cred == nil {
		t.Fatal("expected non-nil credential")
	}
	if cred.ID != 1 {
		t.Errorf("expected ID 1, got %d", cred.ID)
	}
}

func TestCloudSyncService_UpdateCredential_V24(t *testing.T) {
	callCount := 0
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				callCount++
				if callCount == 1 {
					// Verify V24 format
					slice := params.([]any)
					p := slice[1].(map[string]any)
					if _, ok := p["provider"].(string); !ok {
						t.Fatal("expected provider to be string for V24")
					}
					if _, ok := p["attributes"].(map[string]any); !ok {
						t.Fatal("expected attributes to be map for V24")
					}
					return json.RawMessage(`{"id": 1}`), nil
				}
				return sampleCredentialV24JSON(), nil
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 24, Minor: 10})
	cred, err := svc.UpdateCredential(context.Background(), 1, UpdateCredentialOpts{
		Name:         "My S3 Cred",
		ProviderType: "S3",
		Attributes: map[string]string{
			"access_key_id": "AKIATEST",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cred == nil {
		t.Fatal("expected non-nil credential")
	}
}

func TestCloudSyncService_UpdateCredential_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, errors.New("not found")
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 25, Minor: 4})
	_, err := svc.UpdateCredential(context.Background(), 999, UpdateCredentialOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCloudSyncService_DeleteCredential(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				if method != "cloudsync.credentials.delete" {
					t.Errorf("expected method cloudsync.credentials.delete, got %s", method)
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
	err := svc.DeleteCredential(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCloudSyncService_DeleteCredential_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, errors.New("permission denied")
			},
		},
	}

	svc := NewCloudSyncService(mock, Version{Major: 25, Minor: 4})
	err := svc.DeleteCredential(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

package truenas

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestSSHService_CreateSSHKeyPair(t *testing.T) {
	callCount := 0
	testData := sampleSSHKeyPair()
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				callCount++
				if callCount == 1 {
					if method != "keychaincredential.create" {
						t.Errorf("expected method keychaincredential.create, got %s", method)
					}
					p := params.(map[string]any)
					if p["name"] != testData.Name {
						t.Errorf("expected name '%s', got '%s'", testData.Name, p["name"])
					}
					if p["type"] != KeychainCredentialTypeSSHKeyPair {
						t.Errorf("expected type '%s', got '%s'", KeychainCredentialTypeSSHKeyPair, p["type"])
					}
					if _, ok := p["attributes"]; !ok {
						t.Error("expected attributes in params")
					} else {
						attrs := p["attributes"].(map[string]interface{})
						if attrs["public_key"] != testData.PublicKey {
							t.Errorf("expected attrs.public_key '%s', got '%s'", testData.PublicKey, p["public_key"])
						}
						if attrs["private_key"] != testData.PrivateKey {
							t.Errorf("expected attrs.private_key '%s', got '%s'", testData.PrivateKey, p["private_key"])
						}
					}

					return json.RawMessage(`{"id": 1}`), nil
				}
				return sampleSSHKeyPairJSON(), nil
			},
		},
	}

	svc := NewSSHService(mock, Version{Major: 25, Minor: 4})
	sshKeyPair, err := svc.CreateSSHKeyPair(context.Background(), CreateSSHKeyPairOpts{
		Name:       testData.Name,
		PublicKey:  testData.PublicKey,
		PrivateKey: testData.PrivateKey,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sshKeyPair == nil {
		t.Fatal("expected non-nil sshkeypair")
	}
	if sshKeyPair.ID != 1 {
		t.Errorf("expected ID %d, got %d", testData.ID, sshKeyPair.ID)
	}
	if sshKeyPair.Name != "ssh keypair name" {
		t.Errorf("expected name '%s', got '%s'", testData.Name, sshKeyPair.Name)
	}
	if sshKeyPair.PublicKey != "public key" {
		t.Errorf("expected publickey '%s', got '%s'", testData.PublicKey, sshKeyPair.PublicKey)
	}
	if sshKeyPair.PrivateKey != "private key" {
		t.Errorf("expected privatekey '%s', got '%s'", testData.PrivateKey, sshKeyPair.PrivateKey)
	}
}

func TestSSHService_CreateSSHKeyPair_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, errors.New("connection refused")
			},
		},
	}

	svc := NewSSHService(mock, Version{Major: 25, Minor: 4})
	sshKeyPair, err := svc.CreateSSHKeyPair(context.Background(), CreateSSHKeyPairOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
	if sshKeyPair != nil {
		t.Error("expected nil sshkeypair on error")
	}
}

func TestSSHService_CreateSSHKeyPair_ParseError(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return json.RawMessage(`not json`), nil
			},
		},
	}

	svc := NewSSHService(mock, Version{Major: 25, Minor: 4})
	_, err := svc.CreateSSHKeyPair(context.Background(), CreateSSHKeyPairOpts{})
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestSSHService_GetSSHKeyPair(t *testing.T) {
	testData := sampleSSHKeyPair()
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				if method != "keychaincredential.query" {
					t.Errorf("expected method keychaincredential.query, got %s", method)
				}
				return sampleSSHKeyPairJSON(), nil
			},
		},
	}

	svc := NewSSHService(mock, Version{Major: 25, Minor: 4})
	sshKeyPair, err := svc.GetSSHKeyPair(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sshKeyPair == nil {
		t.Fatal("expected non-nil sshKeyPair")
	}
	if sshKeyPair.ID != testData.ID {
		t.Errorf("expected ID %d, got %d", testData.ID, sshKeyPair.ID)
	}
	if sshKeyPair.Name != testData.Name {
		t.Errorf("expected name '%s', got '%s'", testData.Name, sshKeyPair.Name)
	}
	if sshKeyPair.PublicKey != testData.PublicKey {
		t.Errorf("expected publickey '%s', got '%s'", testData.PublicKey, sshKeyPair.PublicKey)
	}
	if sshKeyPair.PrivateKey != testData.PrivateKey {
		t.Errorf("expected privatekey '%s', got '%s'", testData.PrivateKey, sshKeyPair.PrivateKey)
	}
}

func TestSSHService_GetSSHKeyPair_NotFound(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return json.RawMessage(`[]`), nil
			},
		},
	}

	svc := NewSSHService(mock, Version{Major: 25, Minor: 4})
	sshKeyPair, err := svc.GetSSHKeyPair(context.Background(), 999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sshKeyPair != nil {
		t.Error("expected nil sshkeypair for not found")
	}
}

func TestSSHService_GetSSHKeyPair_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, errors.New("timeout")
			},
		},
	}

	svc := NewSSHService(mock, Version{Major: 25, Minor: 4})
	_, err := svc.GetSSHKeyPair(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSSHService_GetSSHKeyPair_ParseError(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return json.RawMessage(`not json`), nil
			},
		},
	}

	svc := NewSSHService(mock, Version{Major: 25, Minor: 4})
	_, err := svc.GetSSHKeyPair(context.Background(), 1)
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestSSHService_ListSSHKeyPairs(t *testing.T) {
	testData := sampleSSHKeyPairs()
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				if method != "keychaincredential.query" {
					t.Errorf("expected method keychaincredential.query, got %s", method)
				}
				if params == nil {
					t.Error("expected filter params for ListSSHKeyPairs")
				}
				return json.RawMessage(sampleSSHKeyPairsJSON()), nil
			},
		},
	}

	svc := NewSSHService(mock, Version{Major: 25, Minor: 4})
	sshKeyPairs, err := svc.ListSSHKeyPairs(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sshKeyPairs) != 2 {
		t.Fatalf("expected 2 sshKeyPairs, got %d", len(sshKeyPairs))
	}

	if sshKeyPairs[0].ID != testData[0].ID {
		t.Errorf("expected ID[0] %d, got %d", testData[0].ID, sshKeyPairs[0].ID)
	}
	if sshKeyPairs[0].Name != testData[0].Name {
		t.Errorf("expected name[0] '%s', got '%s'", testData[0].Name, sshKeyPairs[0].Name)
	}
	if sshKeyPairs[0].PublicKey != testData[0].PublicKey {
		t.Errorf("expected publickey[0] '%s', got '%s'", testData[0].PublicKey, sshKeyPairs[0].PublicKey)
	}
	if sshKeyPairs[0].PrivateKey != testData[0].PrivateKey {
		t.Errorf("expected privatekey[0] '%s', got '%s'", testData[0].PrivateKey, sshKeyPairs[0].PrivateKey)
	}

	if sshKeyPairs[1].ID != testData[1].ID {
		t.Errorf("expected ID[1] %d, got %d", testData[1].ID, sshKeyPairs[1].ID)
	}
	if sshKeyPairs[1].Name != testData[1].Name {
		t.Errorf("expected name[1] '%s', got '%s'", testData[1].Name, sshKeyPairs[1].Name)
	}
	if sshKeyPairs[1].PublicKey != testData[1].PublicKey {
		t.Errorf("expected publickey[1] '%s', got '%s'", testData[1].PublicKey, sshKeyPairs[1].PublicKey)
	}
	if sshKeyPairs[1].PrivateKey != testData[1].PrivateKey {
		t.Errorf("expected privatekey[1] '%s', got '%s'", testData[1].PrivateKey, sshKeyPairs[1].PrivateKey)
	}
}

func TestSSHService_ListSSHKeyPairs_Empty(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return json.RawMessage(`[]`), nil
			},
		},
	}

	svc := NewSSHService(mock, Version{Major: 25, Minor: 4})
	sshKeyPairs, err := svc.ListSSHKeyPairs(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sshKeyPairs) != 0 {
		t.Errorf("expected 0 sshkeypairs, got %d", len(sshKeyPairs))
	}
}

func TestSSHService_ListSSHKeyPairs_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, errors.New("network error")
			},
		},
	}

	svc := NewSSHService(mock, Version{Major: 25, Minor: 4})
	_, err := svc.ListSSHKeyPairs(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSSHService_ListSSHKeyPairs_ParseError(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return json.RawMessage(`not json`), nil
			},
		},
	}

	svc := NewSSHService(mock, Version{Major: 25, Minor: 4})
	_, err := svc.ListSSHKeyPairs(context.Background())
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestSSHService_UpdateSSHKeyPair(t *testing.T) {
	callCount := 0
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				callCount++
				if callCount == 1 {
					if method != "keychaincredential.update" {
						t.Errorf("expected method keychaincredential.update, got %s", method)
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
					p := slice[1].(map[string]any)
					if p["name"] != "Updated keypair" {
						t.Errorf("expected name 'Updated keypair', got %v", p["name"])
					}
					return json.RawMessage(`{"id": 1}`), nil
				}
				return sampleSSHKeyPairJSON(), nil
			},
		},
	}

	svc := NewSSHService(mock, Version{Major: 25, Minor: 4})
	sshKeyPair, err := svc.UpdateSSHKeyPair(context.Background(), 1, UpdateSSHKeyPairOpts{
		Name: "Updated keypair",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sshKeyPair == nil {
		t.Fatal("expected non-nil sshkeypair")
	}
	if sshKeyPair.ID != 1 {
		t.Errorf("expected ID 1, got %d", sshKeyPair.ID)
	}
}

func TestSSHService_UpdateSSHKeyPair_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, errors.New("not found")
			},
		},
	}

	svc := NewSSHService(mock, Version{Major: 25, Minor: 4})
	_, err := svc.UpdateSSHKeyPair(context.Background(), 999, UpdateSSHKeyPairOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSSHService_DeleteSSHKeyPair(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				if method != "keychaincredential.delete" {
					t.Errorf("expected method keychaincredential.delete, got %s", method)
				}
				id, ok := params.(int64)
				if !ok || id != 5 {
					t.Errorf("expected id 5, got %v", params)
				}
				return nil, nil
			},
		},
	}

	svc := NewSSHService(mock, Version{Major: 25, Minor: 4})
	err := svc.DeleteSSHKeyPair(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSSHService_DeleteSSHKeyPair_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, errors.New("permission denied")
			},
		},
	}

	svc := NewSSHService(mock, Version{Major: 25, Minor: 4})
	err := svc.DeleteSSHKeyPair(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSSHService_CreateSSHConnection(t *testing.T) {
	callCount := 0
	testData := sampleSSHConnection()
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				callCount++
				if callCount == 1 {
					if method != "keychaincredential.create" {
						t.Errorf("expected method keychaincredential.create, got %s", method)
					}
					p := params.(map[string]any)
					if p["name"] != testData.Name {
						t.Errorf("expected name '%s', got '%s'", testData.Name, p["name"])
					}
					if p["type"] != KeychainCredentialTypeSSHConnection {
						t.Errorf("expected type '%s', got '%s'", KeychainCredentialTypeSSHConnection, p["type"])
					}

					if _, ok := p["attributes"]; !ok {
						t.Error("expected attributes in params")
					} else {
						attrs := p["attributes"].(map[string]interface{})
						if attrs["host"] != testData.Host {
							t.Errorf("expected host '%s', got '%s'", testData.Host, p["host"])
						}
						if attrs["port"] != testData.Port {
							t.Errorf("expected port '%d', got '%d'", testData.Port, p["port"])
						}
						if attrs["username"] != testData.Username {
							t.Errorf("expected username '%s', got '%s'", testData.Username, p["username"])
						}
						if attrs["private_key"] != testData.PrivateKeyID {
							t.Errorf("expected private_key '%d', got '%d'", testData.PrivateKeyID, p["private_key"])
						}
						if attrs["remote_host_key"] != testData.RemoteHostKey {
							t.Errorf("expected remote_host_key '%s', got '%s'", testData.RemoteHostKey, p["remote_host_key"])
						}
						if attrs["connect_timeout"] != testData.ConnectTimeout {
							t.Errorf("expected connect_timeout '%d', got '%d'", testData.ConnectTimeout, p["connect_timeout"])
						}
					}

					return json.RawMessage(`{"id": 1}`), nil
				}
				return sampleSSHConnectionJSON(), nil
			},
		},
	}

	svc := NewSSHService(mock, Version{Major: 25, Minor: 4})
	sshConnection, err := svc.CreateSSHConnection(context.Background(), CreateSSHConnectionOpts{
		Name:           testData.Name,
		Host:           testData.Host,
		Port:           testData.Port,
		Username:       testData.Username,
		PrivateKeyID:   testData.PrivateKeyID,
		RemoteHostKey:  testData.RemoteHostKey,
		ConnectTimeout: testData.ConnectTimeout,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sshConnection == nil {
		t.Fatal("expected non-nil sshconnection")
	}
	if sshConnection.ID != 1 {
		t.Errorf("expected ID %d, got %d", testData.ID, sshConnection.ID)
	}
	if sshConnection.Name != testData.Name {
		t.Errorf("expected name '%s', got '%s'", testData.Name, sshConnection.Name)
	}
	if sshConnection.Host != testData.Host {
		t.Errorf("expected host '%s', got %s", testData.Host, sshConnection.Host)
	}
	if sshConnection.Port != testData.Port {
		t.Errorf("expected port '%d', got %d", testData.Port, sshConnection.Port)
	}
	if sshConnection.Username != testData.Username {
		t.Errorf("expected username '%s', got %s", testData.Username, sshConnection.Username)
	}
	if sshConnection.PrivateKeyID != testData.PrivateKeyID {
		t.Errorf("expected private_key '%d', got %d", testData.PrivateKeyID, sshConnection.PrivateKeyID)
	}
	if sshConnection.RemoteHostKey != testData.RemoteHostKey {
		t.Errorf("expected remote_host_key '%s', got %s", testData.RemoteHostKey, sshConnection.RemoteHostKey)
	}
	if sshConnection.ConnectTimeout != testData.ConnectTimeout {
		t.Errorf("expected connect_timeout '%d', got %d", testData.ConnectTimeout, sshConnection.ConnectTimeout)
	}
}

func TestSSHService_CreateSSHConnection_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, errors.New("connection refused")
			},
		},
	}

	svc := NewSSHService(mock, Version{Major: 25, Minor: 4})
	sshConnection, err := svc.CreateSSHConnection(context.Background(), CreateSSHConnectionOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
	if sshConnection != nil {
		t.Error("expected nil sshkeypair on error")
	}
}

func TestSSHService_CreateSSHConnection_ParseError(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return json.RawMessage(`not json`), nil
			},
		},
	}

	svc := NewSSHService(mock, Version{Major: 25, Minor: 4})
	_, err := svc.CreateSSHConnection(context.Background(), CreateSSHConnectionOpts{})
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestSSHService_GetSSHConnection(t *testing.T) {
	testData := sampleSSHConnection()
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				if method != "keychaincredential.query" {
					t.Errorf("expected method keychaincredential.query, got %s", method)
				}
				return sampleSSHConnectionJSON(), nil
			},
		},
	}

	svc := NewSSHService(mock, Version{Major: 25, Minor: 4})
	sshConnection, err := svc.GetSSHConnection(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sshConnection == nil {
		t.Fatal("expected non-nil sshConnection")
	}
	if sshConnection.ID != testData.ID {
		t.Errorf("expected ID %d, got %d", testData.ID, sshConnection.ID)
	}
	if sshConnection.Name != testData.Name {
		t.Errorf("expected name '%s', got '%s'", testData.Name, sshConnection.Name)
	}
	if sshConnection.Host != testData.Host {
		t.Errorf("expected host '%s', got %s", testData.Host, sshConnection.Host)
	}
	if sshConnection.Port != testData.Port {
		t.Errorf("expected port '%d', got %d", testData.Port, sshConnection.Port)
	}
	if sshConnection.Username != testData.Username {
		t.Errorf("expected username '%s', got %s", testData.Username, sshConnection.Username)
	}
	if sshConnection.PrivateKeyID != testData.PrivateKeyID {
		t.Errorf("expected private_key '%d', got %d", testData.PrivateKeyID, sshConnection.PrivateKeyID)
	}
	if sshConnection.RemoteHostKey != testData.RemoteHostKey {
		t.Errorf("expected remote_host_key '%s', got %s", testData.RemoteHostKey, sshConnection.RemoteHostKey)
	}
	if sshConnection.ConnectTimeout != testData.ConnectTimeout {
		t.Errorf("expected connect_timeout '%d', got %d", testData.ConnectTimeout, sshConnection.ConnectTimeout)
	}
}

func TestSSHService_GetSSHConnection_NotFound(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return json.RawMessage(`[]`), nil
			},
		},
	}

	svc := NewSSHService(mock, Version{Major: 25, Minor: 4})
	sshConnection, err := svc.GetSSHConnection(context.Background(), 999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sshConnection != nil {
		t.Error("expected nil sshkeypair for not found")
	}
}

func TestSSHService_GetSSHConnection_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, errors.New("timeout")
			},
		},
	}

	svc := NewSSHService(mock, Version{Major: 25, Minor: 4})
	_, err := svc.GetSSHConnection(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSSHService_GetSSHConnection_ParseError(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return json.RawMessage(`not json`), nil
			},
		},
	}

	svc := NewSSHService(mock, Version{Major: 25, Minor: 4})
	_, err := svc.GetSSHConnection(context.Background(), 1)
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestSSHService_ListSSHConnections(t *testing.T) {
	testData := sampleSSHConnections()
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				if method != "keychaincredential.query" {
					t.Errorf("expected method keychaincredential.query, got %s", method)
				}
				if params == nil {
					t.Error("expected filter params for ListSSHConnections")
				}
				return json.RawMessage(sampleSSHConnectionsJSON()), nil
			},
		},
	}

	svc := NewSSHService(mock, Version{Major: 25, Minor: 4})
	sshConnections, err := svc.ListSSHConnections(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sshConnections) != 2 {
		t.Fatalf("expected 2 sshConnections, got %d", len(sshConnections))
	}

	if sshConnections[0].ID != testData[0].ID {
		t.Errorf("expected ID[0] %d, got %d", testData[0].ID, sshConnections[0].ID)
	}
	if sshConnections[0].Name != testData[0].Name {
		t.Errorf("expected name[0] '%s', got '%s'", testData[0].Name, sshConnections[0].Name)
	}
	if sshConnections[0].Host != testData[0].Host {
		t.Errorf("expected host '%s', got %s", testData[0].Host, sshConnections[0].Host)
	}
	if sshConnections[0].Port != testData[0].Port {
		t.Errorf("expected port '%d', got %d", testData[0].Port, sshConnections[0].Port)
	}
	if sshConnections[0].Username != testData[0].Username {
		t.Errorf("expected username '%s', got %s", testData[0].Username, sshConnections[0].Username)
	}
	if sshConnections[0].PrivateKeyID != testData[0].PrivateKeyID {
		t.Errorf("expected private_key '%d', got %d", testData[0].PrivateKeyID, sshConnections[0].PrivateKeyID)
	}
	if sshConnections[0].RemoteHostKey != testData[0].RemoteHostKey {
		t.Errorf("expected remote_host_key '%s', got %s", testData[0].RemoteHostKey, sshConnections[0].RemoteHostKey)
	}
	if sshConnections[0].ConnectTimeout != testData[0].ConnectTimeout {
		t.Errorf("expected connect_timeout '%d', got %d", testData[0].ConnectTimeout, sshConnections[0].ConnectTimeout)
	}

	if sshConnections[1].ID != testData[1].ID {
		t.Errorf("expected ID[1] %d, got %d", testData[1].ID, sshConnections[1].ID)
	}
	if sshConnections[1].Name != testData[1].Name {
		t.Errorf("expected name[1] '%s', got '%s'", testData[1].Name, sshConnections[1].Name)
	}
	if sshConnections[1].Host != testData[1].Host {
		t.Errorf("expected host '%s', got %s", testData[1].Host, sshConnections[1].Host)
	}
	if sshConnections[1].Port != testData[1].Port {
		t.Errorf("expected port '%d', got %d", testData[1].Port, sshConnections[1].Port)
	}
	if sshConnections[1].Username != testData[1].Username {
		t.Errorf("expected username '%s', got %s", testData[1].Username, sshConnections[1].Username)
	}
	if sshConnections[1].PrivateKeyID != testData[1].PrivateKeyID {
		t.Errorf("expected private_key '%d', got %d", testData[1].PrivateKeyID, sshConnections[1].PrivateKeyID)
	}
	if sshConnections[1].RemoteHostKey != testData[1].RemoteHostKey {
		t.Errorf("expected remote_host_key '%s', got %s", testData[1].RemoteHostKey, sshConnections[1].RemoteHostKey)
	}
	if sshConnections[1].ConnectTimeout != testData[1].ConnectTimeout {
		t.Errorf("expected connect_timeout '%d', got %d", testData[1].ConnectTimeout, sshConnections[1].ConnectTimeout)
	}
}

func TestSSHService_ListSSHConnections_Empty(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return json.RawMessage(`[]`), nil
			},
		},
	}

	svc := NewSSHService(mock, Version{Major: 25, Minor: 4})
	sshConnections, err := svc.ListSSHConnections(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sshConnections) != 0 {
		t.Errorf("expected 0 sshkeypairs, got %d", len(sshConnections))
	}
}

func TestSSHService_ListSSHConnections_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, errors.New("network error")
			},
		},
	}

	svc := NewSSHService(mock, Version{Major: 25, Minor: 4})
	_, err := svc.ListSSHConnections(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSSHService_ListSSHConnections_ParseError(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return json.RawMessage(`not json`), nil
			},
		},
	}

	svc := NewSSHService(mock, Version{Major: 25, Minor: 4})
	_, err := svc.ListSSHConnections(context.Background())
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestSSHService_UpdateSSHConnection(t *testing.T) {
	callCount := 0
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				callCount++
				if callCount == 1 {
					if method != "keychaincredential.update" {
						t.Errorf("expected method keychaincredential.update, got %s", method)
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
					p := slice[1].(map[string]any)
					if p["name"] != "Updated keypair" {
						t.Errorf("expected name 'Updated keypair', got %v", p["name"])
					}
					return json.RawMessage(`{"id": 1}`), nil
				}
				return sampleSSHConnectionJSON(), nil
			},
		},
	}

	svc := NewSSHService(mock, Version{Major: 25, Minor: 4})
	sshConnection, err := svc.UpdateSSHConnection(context.Background(), 1, UpdateSSHConnectionOpts{
		Name: "Updated keypair",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sshConnection == nil {
		t.Fatal("expected non-nil sshkeypair")
	}
	if sshConnection.ID != 1 {
		t.Errorf("expected ID 1, got %d", sshConnection.ID)
	}
}

func TestSSHService_UpdateSSHConnection_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, errors.New("not found")
			},
		},
	}

	svc := NewSSHService(mock, Version{Major: 25, Minor: 4})
	_, err := svc.UpdateSSHConnection(context.Background(), 999, UpdateSSHConnectionOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSSHService_DeleteSSHConnection(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				if method != "keychaincredential.delete" {
					t.Errorf("expected method keychaincredential.delete, got %s", method)
				}
				id, ok := params.(int64)
				if !ok || id != 5 {
					t.Errorf("expected id 5, got %v", params)
				}
				return nil, nil
			},
		},
	}

	svc := NewSSHService(mock, Version{Major: 25, Minor: 4})
	err := svc.DeleteSSHConnection(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSSHService_DeleteSSHConnection_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, errors.New("permission denied")
			},
		},
	}

	svc := NewSSHService(mock, Version{Major: 25, Minor: 4})
	err := svc.DeleteSSHConnection(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

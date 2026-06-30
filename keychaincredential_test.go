package truenas

import (
	"encoding/json"
	"testing"
)

func Test_ParseSSHKeyPair(t *testing.T) {
	testData := sampleSSHKeyPair()
	tests := []struct {
		name    string
		raw     keychainCredentialRaw
		want    SSHKeyPairResponse
		wantErr bool
	}{
		{
			name: "parses valid credential json",
			raw: keychainCredentialRaw{
				ID:         testData.ID,
				Name:       testData.Name,
				Type:       KeychainCredentialTypeSSHKeyPair,
				Attributes: json.RawMessage(sampleSSHKeyPairAttributesJSON()),
			},
			want: SSHKeyPairResponse{
				ID:         testData.ID,
				Name:       testData.Name,
				PublicKey:  testData.PublicKey,
				PrivateKey: testData.PrivateKey,
			},
		},
		{
			name: "invalid JSON in provider",
			raw: keychainCredentialRaw{
				ID:         2,
				Name:       "Invalid keypair",
				Type:       KeychainCredentialTypeSSHKeyPair,
				Attributes: json.RawMessage(`{not valid json`),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSSHKeyPair(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSSHKeyPair() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got.ID != tt.want.ID {
				t.Errorf("ID = %v, want %v", got.ID, tt.want.ID)
			}
			if got.Name != tt.want.Name {
				t.Errorf("Name = %v, want %v", got.Name, tt.want.Name)
			}
			if got.PrivateKey != tt.want.PrivateKey {
				t.Errorf("PrivateKey = %v, want %v", got.PrivateKey, tt.want.PrivateKey)
			}
			if got.PublicKey != tt.want.PublicKey {
				t.Errorf("PublicKey = %v, want %v", got.PublicKey, tt.want.PublicKey)
			}
		})
	}
}

func Test_ParseSSHKeyPairs(t *testing.T) {
	testData := sampleSSHKeyPair()
	testDatas := sampleSSHKeyPairs()
	tests := []struct {
		name    string
		data    []byte
		want    []SSHKeyPairResponse
		wantErr bool
	}{
		{
			name: "Single keypair",
			data: []byte(sampleSSHKeyPairJSON()),
			want: []SSHKeyPairResponse{
				{
					ID:         testData.ID,
					Name:       testData.Name,
					PublicKey:  testData.PublicKey,
					PrivateKey: testData.PrivateKey,
				},
			},
		},
		{
			name: "Multiple keypairs",
			data: []byte(sampleSSHKeyPairsJSON()),
			want: []SSHKeyPairResponse{
				{
					ID:         testDatas[0].ID,
					Name:       testDatas[0].Name,
					PublicKey:  testDatas[0].PublicKey,
					PrivateKey: testDatas[0].PrivateKey,
				},
				{
					ID:         testDatas[1].ID,
					Name:       testDatas[1].Name,
					PublicKey:  testDatas[1].PublicKey,
					PrivateKey: testDatas[1].PrivateKey,
				},
			},
		},
		{
			name:    "invalid JSON array",
			data:    []byte(`{not valid json`),
			wantErr: true,
		},
		{
			name: "invalid attributes JSON",
			data: []byte(`[{
				"id": 1,
				"name": "Invalid",
				"attributes": {invalid json}
			}]`),
			wantErr: true,
		},
		{
			name: "empty array",
			data: []byte(`[]`),
			want: []SSHKeyPairResponse{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSSHKeyPairs(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSSHKeyPairs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("ParseSSHKeyPairs() returned %d SSHKeyPairs, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if got[i].ID != tt.want[i].ID {
					t.Errorf("SSHKeyPair[%d].ID = %v, want %v", i, got[i].ID, tt.want[i].ID)
				}
				if got[i].Name != tt.want[i].Name {
					t.Errorf("SSHKeyPair[%d].Name = %v, want %v", i, got[i].Name, tt.want[i].Name)
				}
				if got[i].PublicKey != tt.want[i].PublicKey {
					t.Errorf("SSHKeyPair[%d].PublicKey = %v, want %v", i, got[i].PublicKey, tt.want[i].PublicKey)
				}
				if got[i].PrivateKey != tt.want[i].PrivateKey {
					t.Errorf("SSHKeyPair[%d].PrivateKey = %v, want %v", i, got[i].PrivateKey, tt.want[i].PrivateKey)
				}
			}
		})
	}
}

func Test_ParseSSHConnection(t *testing.T) {
	testData := sampleSSHConnection()
	tests := []struct {
		name    string
		raw     keychainCredentialRaw
		want    SSHConnectionResponse
		wantErr bool
	}{
		{
			name: "parses valid credential json",
			raw: keychainCredentialRaw{
				ID:         testData.ID,
				Name:       testData.Name,
				Type:       KeychainCredentialTypeSSHConnection,
				Attributes: json.RawMessage(sampleSSHConnectionAttributesJSON()),
			},
			want: SSHConnectionResponse{
				ID:             testData.ID,
				Name:           testData.Name,
				Host:           testData.Host,
				Port:           testData.Port,
				Username:       testData.Username,
				PrivateKeyID:   testData.PrivateKeyID,
				RemoteHostKey:  testData.RemoteHostKey,
				ConnectTimeout: testData.ConnectTimeout,
			},
		},
		{
			name: "invalid JSON in provider",
			raw: keychainCredentialRaw{
				ID:         2,
				Name:       "Invalid credential",
				Type:       KeychainCredentialTypeSSHConnection,
				Attributes: json.RawMessage(`{not valid json`),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSSHConnection(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSSHConnection() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got.ID != tt.want.ID {
				t.Errorf("ID = %v, want %v", got.ID, tt.want.ID)
			}
			if got.Name != tt.want.Name {
				t.Errorf("Name = %v, want %v", got.Name, tt.want.Name)
			}
			if got.Host != tt.want.Host {
				t.Errorf("Host = %v, want %v", got.Host, tt.want.Host)
			}
			if got.Port != tt.want.Port {
				t.Errorf("Port = %v, want %v", got.Port, tt.want.Port)
			}
			if got.Username != tt.want.Username {
				t.Errorf("Username = %v, want %v", got.Username, tt.want.Username)
			}
			if got.PrivateKeyID != tt.want.PrivateKeyID {
				t.Errorf("PrivateKeyID = %v, want %v", got.PrivateKeyID, tt.want.PrivateKeyID)
			}
			if got.RemoteHostKey != tt.want.RemoteHostKey {
				t.Errorf("RemoteHostKey = %v, want %v", got.RemoteHostKey, tt.want.RemoteHostKey)
			}
			if got.ConnectTimeout != tt.want.ConnectTimeout {
				t.Errorf("ConnectTimeout = %v, want %v", got.ConnectTimeout, tt.want.ConnectTimeout)
			}
		})
	}
}

func Test_ParseSSHConnections(t *testing.T) {
	testData := sampleSSHConnection()
	testDatas := sampleSSHConnections()
	tests := []struct {
		name    string
		data    []byte
		want    []SSHConnectionResponse
		wantErr bool
	}{
		{
			name: "Single credential",
			data: []byte(sampleSSHConnectionJSON()),
			want: []SSHConnectionResponse{
				{
					ID:             testData.ID,
					Name:           testData.Name,
					Host:           testData.Host,
					Port:           testData.Port,
					Username:       testData.Username,
					PrivateKeyID:   testData.PrivateKeyID,
					RemoteHostKey:  testData.RemoteHostKey,
					ConnectTimeout: testData.ConnectTimeout,
				},
			},
		},
		{
			name: "Multiple credentials",
			data: []byte(sampleSSHConnectionsJSON()),
			want: []SSHConnectionResponse{
				{
					ID:             testDatas[0].ID,
					Name:           testDatas[0].Name,
					Host:           testDatas[0].Host,
					Port:           testDatas[0].Port,
					Username:       testDatas[0].Username,
					PrivateKeyID:   testDatas[0].PrivateKeyID,
					RemoteHostKey:  testDatas[0].RemoteHostKey,
					ConnectTimeout: testDatas[0].ConnectTimeout,
				},
				{
					ID:             testDatas[1].ID,
					Name:           testDatas[1].Name,
					Host:           testDatas[1].Host,
					Port:           testDatas[1].Port,
					Username:       testDatas[1].Username,
					PrivateKeyID:   testDatas[1].PrivateKeyID,
					RemoteHostKey:  testDatas[1].RemoteHostKey,
					ConnectTimeout: testDatas[1].ConnectTimeout,
				},
			},
		},
		{
			name:    "invalid JSON array",
			data:    []byte(`{not valid json`),
			wantErr: true,
		},
		{
			name: "invalid attributes JSON",
			data: []byte(`[{
				"id": 1,
				"name": "Invalid",
				"attributes": {invalid json}
			}]`),
			wantErr: true,
		},
		{
			name: "empty array",
			data: []byte(`[]`),
			want: []SSHConnectionResponse{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSSHConnections(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSSHConnections() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("ParseSSHConnections() returned %d SSHConnections, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if got[i].ID != tt.want[i].ID {
					t.Errorf("SSHConnection[%d].ID = %v, want %v", i, got[i].ID, tt.want[i].ID)
				}
				if got[i].Name != tt.want[i].Name {
					t.Errorf("SSHConnection[%d].Name = %v, want %v", i, got[i].Name, tt.want[i].Name)
				}
				if got[i].Host != tt.want[i].Host {
					t.Errorf("SSHConnection[%d].Host = %v, want %v", i, got[i].Host, tt.want[i].Host)
				}
				if got[i].Port != tt.want[i].Port {
					t.Errorf("SSHConnection[%d].Port = %v, want %v", i, got[i].Port, tt.want[i].Port)
				}
				if got[i].Username != tt.want[i].Username {
					t.Errorf("SSHConnection[%d].Username = %v, want %v", i, got[i].Username, tt.want[i].Username)
				}
				if got[i].PrivateKeyID != tt.want[i].PrivateKeyID {
					t.Errorf("SSHConnection[%d].PrivateKeyID = %v, want %v", i, got[i].PrivateKeyID, tt.want[i].PrivateKeyID)
				}
				if got[i].RemoteHostKey != tt.want[i].RemoteHostKey {
					t.Errorf("SSHConnection[%d].RemoteHostKey = %v, want %v", i, got[i].RemoteHostKey, tt.want[i].RemoteHostKey)
				}
				if got[i].ConnectTimeout != tt.want[i].ConnectTimeout {
					t.Errorf("SSHConnection[%d].ConnectTimeout = %v, want %v", i, got[i].ConnectTimeout, tt.want[i].ConnectTimeout)
				}
			}
		})
	}
}

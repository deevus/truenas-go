package truenas

import (
	"encoding/json"
	"fmt"
)

const KeychainCredentialTypeSSHKeyPair = "SSH_KEY_PAIR"
const KeychainCredentialTypeSSHConnection = "SSH_CREDENTIALS"

// keychainCredentialRaw is the intermediate struct for parsing API responses.
type keychainCredentialRaw struct {
	ID         int64           `json:"id"`
	Name       string          `json:"name"`
	Type       string          `json:"type"`
	Attributes json.RawMessage `json:"attributes"`
}

// SSHKeyPairResponse represents a SSH keypair from the API.
type SSHKeyPairResponse struct {
	ID         int64
	Name       string
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
}

// SSHConnectionResponse represents a SSH credential from the API.
type SSHConnectionResponse struct {
	ID             int64
	Name           string
	Host           string `json:"host"`
	Port           int32  `json:"port"`
	Username       string `json:"username"`
	PrivateKeyID   int64  `json:"private_key"`
	RemoteHostKey  string `json:"remote_host_key"`
	ConnectTimeout int32  `json:"connect_timeout,omitempty"`
}

func ParseSSHKeyPairs(data []byte) ([]SSHKeyPairResponse, error) {
	var raws []keychainCredentialRaw
	if err := json.Unmarshal(data, &raws); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	results := make([]SSHKeyPairResponse, 0, len(raws))
	for _, raw := range raws {
		cred, err := parseSSHKeyPair(raw)
		if err != nil {
			return nil, err
		}
		results = append(results, cred)
	}
	return results, nil
}

func parseSSHKeyPair(raw keychainCredentialRaw) (SSHKeyPairResponse, error) {
	var c SSHKeyPairResponse
	c.ID = raw.ID
	c.Name = raw.Name

	if len(raw.Attributes) > 0 {
		if err := json.Unmarshal(raw.Attributes, &c); err != nil {
			return c, fmt.Errorf("parse attributes: %w", err)
		}
	}
	return c, nil
}

func ParseSSHConnections(data []byte) ([]SSHConnectionResponse, error) {
	var raws []keychainCredentialRaw
	if err := json.Unmarshal(data, &raws); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	results := make([]SSHConnectionResponse, 0, len(raws))
	for _, raw := range raws {
		cred, err := parseSSHConnection(raw)
		if err != nil {
			return nil, err
		}
		results = append(results, cred)
	}
	return results, nil
}

func parseSSHConnection(raw keychainCredentialRaw) (SSHConnectionResponse, error) {
	var c SSHConnectionResponse
	c.ID = raw.ID
	c.Name = raw.Name

	if len(raw.Attributes) > 0 {
		if err := json.Unmarshal(raw.Attributes, &c); err != nil {
			return c, fmt.Errorf("parse attributes: %w", err)
		}
	}
	return c, nil
}

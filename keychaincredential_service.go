package truenas

import (
	"context"
	"encoding/json"
	"fmt"
)

// SSHKeyPair is the user-facing representation of a TrueNAS ssh keypair.
type SSHKeyPair struct {
	ID         int64
	Name       string
	PrivateKey string
	PublicKey  string
}

// CreateSSHKeyPairOpts contains options for creating a ssh keypair.
type CreateSSHKeyPairOpts struct {
	Name       string
	PublicKey  string
	PrivateKey string
}

// UpdateSSHKeyPairOpts contains options for updating a ssh keypair.
type UpdateSSHKeyPairOpts = CreateSSHKeyPairOpts

// SSHConnection is the user-facing representation of a TrueNAS ssh credential.
type SSHConnection struct {
	ID             int64
	Name           string
	Host           string
	Port           int32
	Username       string
	PrivateKeyID   int64
	RemoteHostKey  string
	ConnectTimeout int32
}

// CreateSSHConnectionOpts contains options for creating an ssh keypair.
type CreateSSHConnectionOpts struct {
	Name           string
	Host           string
	Port           int32
	Username       string
	PrivateKeyID   int64
	RemoteHostKey  string
	ConnectTimeout int32
}

// UpdateSSHConnectionOpts contains options for updating an ssh credential.
type UpdateSSHConnectionOpts = CreateSSHConnectionOpts

// SSHService provides typed methods for the keychaincredential API namespaces.
type SSHService struct {
	client  Caller
	version Version
}

// NewSSHService creates a new SSHService.
func NewSSHService(c Caller, v Version) *SSHService {
	return &SSHService{client: c, version: v}
}

// CreateSSHKeyPair creates an ssh keypair and returns the full object.
func (s *SSHService) CreateSSHKeyPair(ctx context.Context, opts CreateSSHKeyPairOpts) (*SSHKeyPair, error) {
	params := sshKeyPairOptsToParams(opts, true)

	result, err := s.client.Call(ctx, "keychaincredential.create", params)
	if err != nil {
		return nil, err
	}

	var createResp struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(result, &createResp); err != nil {
		return nil, fmt.Errorf("parse create response: %w", err)
	}

	return s.GetSSHKeyPair(ctx, createResp.ID)
}

// GetSSHKeyPair returns a SSH keypair by ID, or nil if not found.
func (s *SSHService) GetSSHKeyPair(ctx context.Context, id int64) (*SSHKeyPair, error) {
	filter := [][]any{{"id", "=", id}}
	result, err := s.client.Call(ctx, "keychaincredential.query", filter)
	if err != nil {
		return nil, err
	}

	keyPairs, err := ParseSSHKeyPairs(result)
	if err != nil {
		return nil, err
	}

	if len(keyPairs) == 0 {
		return nil, nil
	}

	keyPair := sshKeyPairFromResponse(keyPairs[0])
	return &keyPair, nil
}

// ListSSHKeyPairs returns all keychain credentials.
func (s *SSHService) ListSSHKeyPairs(ctx context.Context) ([]SSHKeyPair, error) {
	filter := [][]any{{"type", "=", KeychainCredentialTypeSSHKeyPair}}
	result, err := s.client.Call(ctx, "keychaincredential.query", filter)
	if err != nil {
		return nil, err
	}

	keyPairs, err := ParseSSHKeyPairs(result)
	if err != nil {
		return nil, err
	}

	sshKeyPairs := make([]SSHKeyPair, len(keyPairs))
	for i, resp := range keyPairs {
		sshKeyPairs[i] = sshKeyPairFromResponse(resp)
	}
	return sshKeyPairs, nil
}

// UpdateSSHKeyPair updates a SSH keypair and returns the full object.
func (s *SSHService) UpdateSSHKeyPair(ctx context.Context, id int64, opts UpdateSSHKeyPairOpts) (*SSHKeyPair, error) {
	params := sshKeyPairOptsToParams(opts, false)

	_, err := s.client.Call(ctx, "keychaincredential.update", []any{id, params})
	if err != nil {
		return nil, err
	}

	return s.GetSSHKeyPair(ctx, id)
}

// DeleteSSHKeyPair deletes a SSH keypair by ID.
func (s *SSHService) DeleteSSHKeyPair(ctx context.Context, id int64) error {
	_, err := s.client.Call(ctx, "keychaincredential.delete", id)
	return err
}

// CreateSSHConnection creates an ssh credential and returns the full object.
func (s *SSHService) CreateSSHConnection(ctx context.Context, opts CreateSSHConnectionOpts) (*SSHConnection, error) {
	params := sshConnectionOptsToParams(opts, true)

	result, err := s.client.Call(ctx, "keychaincredential.create", params)
	if err != nil {
		return nil, err
	}

	var createResp struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(result, &createResp); err != nil {
		return nil, fmt.Errorf("parse create response: %w", err)
	}

	return s.GetSSHConnection(ctx, createResp.ID)
}

// GetSSHConnection returns a SSH credential by ID, or nil if not found.
func (s *SSHService) GetSSHConnection(ctx context.Context, id int64) (*SSHConnection, error) {
	filter := [][]any{{"id", "=", id}}
	result, err := s.client.Call(ctx, "keychaincredential.query", filter)
	if err != nil {
		return nil, err
	}

	credentials, err := ParseSSHConnections(result)
	if err != nil {
		return nil, err
	}

	if len(credentials) == 0 {
		return nil, nil
	}

	credential := sshConnectionFromResponse(credentials[0])
	return &credential, nil
}

// ListSSHConnections returns all keychain credentials.
func (s *SSHService) ListSSHConnections(ctx context.Context) ([]SSHConnection, error) {
	filter := [][]any{{"type", "=", KeychainCredentialTypeSSHConnection}}
	result, err := s.client.Call(ctx, "keychaincredential.query", filter)
	if err != nil {
		return nil, err
	}

	credentials, err := ParseSSHConnections(result)
	if err != nil {
		return nil, err
	}

	sshConnections := make([]SSHConnection, len(credentials))
	for i, resp := range credentials {
		sshConnections[i] = sshConnectionFromResponse(resp)
	}
	return sshConnections, nil
}

// UpdateSSHConnection updates a SSH credential and returns the full object.
func (s *SSHService) UpdateSSHConnection(ctx context.Context, id int64, opts UpdateSSHConnectionOpts) (*SSHConnection, error) {
	params := sshConnectionOptsToParams(opts, false)

	_, err := s.client.Call(ctx, "keychaincredential.update", []any{id, params})
	if err != nil {
		return nil, err
	}

	return s.GetSSHConnection(ctx, id)
}

// DeleteSSHConnection deletes a SSH credential by ID.
func (s *SSHService) DeleteSSHConnection(ctx context.Context, id int64) error {
	_, err := s.client.Call(ctx, "keychaincredential.delete", id)
	return err
}

// sshKeyPairOptsToParams converts CreateSSHKeyPairOpts to API parameters.
func sshKeyPairOptsToParams(opts CreateSSHKeyPairOpts, includeType bool) map[string]any {
	attrs := map[string]any{
		"public_key":  opts.PublicKey,
		"private_key": opts.PrivateKey,
	}
	params := map[string]any{
		"name":       opts.Name,
		"attributes": attrs,
	}

	if includeType {
		params["type"] = KeychainCredentialTypeSSHKeyPair
	}

	return params
}

// sshKeyPairFromResponse converts a wire-format SSHKeyPairResponse to a user-facing SSHKeyPair.
func sshKeyPairFromResponse(resp SSHKeyPairResponse) SSHKeyPair {
	return SSHKeyPair{
		ID:         resp.ID,
		Name:       resp.Name,
		PrivateKey: resp.PrivateKey,
		PublicKey:  resp.PublicKey,
	}
}

// sshConnectionOptsToParams converts CreateSSHConnectionOpts to API parameters.
func sshConnectionOptsToParams(opts CreateSSHConnectionOpts, includeType bool) map[string]any {
	attrs := map[string]any{
		"host":            opts.Host,
		"port":            opts.Port,
		"username":        opts.Username,
		"private_key":     opts.PrivateKeyID,
		"remote_host_key": opts.RemoteHostKey,
		"connect_timeout": opts.ConnectTimeout,
	}
	params := map[string]any{
		"name":       opts.Name,
		"attributes": attrs,
	}

	if includeType {
		params["type"] = KeychainCredentialTypeSSHConnection
	}

	return params
}

// sshConnectionFromResponse converts a wire-format SSHConnectionResponse to a user-facing SSHConnection.
func sshConnectionFromResponse(resp SSHConnectionResponse) SSHConnection {
	return SSHConnection{
		ID:             resp.ID,
		Name:           resp.Name,
		Host:           resp.Host,
		Port:           resp.Port,
		Username:       resp.Username,
		PrivateKeyID:   resp.PrivateKeyID,
		RemoteHostKey:  resp.RemoteHostKey,
		ConnectTimeout: resp.ConnectTimeout,
	}
}

package truenas

import (
	"context"
	"encoding/json"
	"fmt"
)

// DockerStatus is the user-facing representation of Docker runtime status.
type DockerStatus struct {
	Status      DockerState
	Description string
}

// DockerConfig is the user-facing representation of Docker configuration.
type DockerConfig struct {
	Pool               string
	EnableImageUpdates bool
	NvidiaEnabled      bool
	AddressPools       []DockerAddressPool
}

// DockerAddressPool represents an address pool in Docker config.
type DockerAddressPool struct {
	Base string
	Size int
}

// DockerService provides typed methods for the docker.* API namespace.
type DockerService struct {
	client  Caller
	version Version
}

// NewDockerService creates a new DockerService.
func NewDockerService(c Caller, v Version) *DockerService {
	return &DockerService{client: c, version: v}
}

// GetStatus returns the current Docker runtime status.
func (s *DockerService) GetStatus(ctx context.Context) (*DockerStatus, error) {
	result, err := s.client.Call(ctx, "docker.status", nil)
	if err != nil {
		return nil, err
	}

	var resp DockerStatusResponse
	if err := json.Unmarshal(result, &resp); err != nil {
		return nil, fmt.Errorf("parse docker.status response: %w", err)
	}

	status := dockerStatusFromResponse(resp)
	return &status, nil
}

// GetConfig returns the current Docker configuration.
func (s *DockerService) GetConfig(ctx context.Context) (*DockerConfig, error) {
	result, err := s.client.Call(ctx, "docker.config", nil)
	if err != nil {
		return nil, err
	}

	var resp DockerConfigResponse
	if err := json.Unmarshal(result, &resp); err != nil {
		return nil, fmt.Errorf("parse docker.config response: %w", err)
	}

	config := dockerConfigFromResponse(resp)
	return &config, nil
}

func dockerStatusFromResponse(resp DockerStatusResponse) DockerStatus {
	return DockerStatus{
		Status:      DockerState(resp.Status),
		Description: resp.Description,
	}
}

func dockerConfigFromResponse(resp DockerConfigResponse) DockerConfig {
	pools := make([]DockerAddressPool, len(resp.AddressPoolsV4))
	for i, p := range resp.AddressPoolsV4 {
		pools[i] = DockerAddressPool{
			Base: p.Base,
			Size: p.Size,
		}
	}
	return DockerConfig{
		Pool:               resp.Pool,
		EnableImageUpdates: resp.EnableImageUpdates,
		NvidiaEnabled:      resp.NvidiaEnabled,
		AddressPools:       pools,
	}
}

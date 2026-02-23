package truenas

import (
	"context"
	"encoding/json"
	"fmt"
)

// SystemInfo is the user-facing representation of TrueNAS system information.
type SystemInfo struct {
	Model         string
	Cores         int
	PhysicalCores int
	Hostname      string
	Uptime        string
	UptimeSeconds float64
	LoadAvg       [3]float64
	EccMemory     bool
}

// SystemService provides typed methods for the system.* API namespace.
type SystemService struct {
	client  Caller
	version Version
}

// NewSystemService creates a new SystemService.
func NewSystemService(c Caller, v Version) *SystemService {
	return &SystemService{client: c, version: v}
}

// GetInfo returns system information.
func (s *SystemService) GetInfo(ctx context.Context) (*SystemInfo, error) {
	result, err := s.client.Call(ctx, "system.info", nil)
	if err != nil {
		return nil, err
	}

	var resp SystemInfoResponse
	if err := json.Unmarshal(result, &resp); err != nil {
		return nil, fmt.Errorf("parse system.info response: %w", err)
	}

	info := systemInfoFromResponse(resp)
	return &info, nil
}

// GetVersion returns the TrueNAS version string.
func (s *SystemService) GetVersion(ctx context.Context) (string, error) {
	result, err := s.client.Call(ctx, "system.version", nil)
	if err != nil {
		return "", err
	}

	var version string
	if err := json.Unmarshal(result, &version); err != nil {
		return "", fmt.Errorf("parse system.version response: %w", err)
	}

	return version, nil
}

// systemInfoFromResponse converts a wire-format SystemInfoResponse to a user-facing SystemInfo.
func systemInfoFromResponse(resp SystemInfoResponse) SystemInfo {
	return SystemInfo{
		Model:         resp.Model,
		Cores:         resp.Cores,
		PhysicalCores: resp.PhysicalCores,
		Hostname:      resp.Hostname,
		Uptime:        resp.Uptime,
		UptimeSeconds: resp.UptimeSeconds,
		LoadAvg:       resp.LoadAvg,
		EccMemory:     resp.EccMemory,
	}
}

package truenas

// DockerState represents the state of the Docker/container runtime.
type DockerState string

const (
	DockerStateRunning      DockerState = "RUNNING"
	DockerStateStopped      DockerState = "STOPPED"
	DockerStateInitializing DockerState = "INITIALIZING"
	DockerStateError        DockerState = "ERROR"
	DockerStateUnconfigured DockerState = "UNCONFIGURED"
)

// DockerStatusResponse represents the wire-format response from docker.status.
type DockerStatusResponse struct {
	Status      string `json:"status"`
	Description string `json:"description"`
}

// DockerConfigResponse represents the wire-format response from docker.config.
type DockerConfigResponse struct {
	Pool               string                      `json:"pool"`
	EnableImageUpdates bool                        `json:"enable_image_updates"`
	NvidiaEnabled      bool                        `json:"nvidia"`
	AddressPoolsV4    []DockerAddressPoolResponse `json:"address_pools"`
}

// DockerAddressPoolResponse represents an address pool in Docker config.
type DockerAddressPoolResponse struct {
	Base string `json:"base"`
	Size int    `json:"size"`
}

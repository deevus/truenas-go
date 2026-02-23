package truenas

// ContainerState represents the state of a container.
type ContainerState string

const (
	ContainerStateRunning  ContainerState = "running"
	ContainerStateExited   ContainerState = "exited"
	ContainerStateStarting ContainerState = "starting"
	ContainerStateStopped  ContainerState = "stopped"
)

// AppResponse represents an app from the TrueNAS API.
type AppResponse struct {
	Name             string                     `json:"name"`
	State            string                     `json:"state"`
	CustomApp        bool                       `json:"custom_app"`
	Config           map[string]any             `json:"config"`
	Version          string                     `json:"version"`
	HumanVersion     string                     `json:"human_version"`
	LatestVersion    string                     `json:"latest_version"`
	UpgradeAvailable bool                       `json:"upgrade_available"`
	ActiveWorkloads  AppActiveWorkloadsResponse `json:"active_workloads"`
}

// AppActiveWorkloadsResponse is the wire-format for active workload data.
type AppActiveWorkloadsResponse struct {
	Containers       int                           `json:"containers"`
	UsedPorts        []AppUsedPortResponse         `json:"used_ports"`
	ContainerDetails []AppContainerDetailsResponse `json:"container_details"`
}

// AppUsedPortResponse represents a port mapping.
type AppUsedPortResponse struct {
	ContainerPort int    `json:"container_port"`
	HostPort      int    `json:"host_port"`
	Protocol      string `json:"protocol"`
}

// AppContainerDetailsResponse represents a container detail.
type AppContainerDetailsResponse struct {
	ID          string `json:"id"`
	ServiceName string `json:"service_name"`
	Image       string `json:"image"`
	State       string `json:"state"`
}

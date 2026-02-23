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

// AppUpgradeSummaryResponse is the wire-format for app.upgrade_summary.
type AppUpgradeSummaryResponse struct {
	LatestVersion       string                          `json:"latest_version"`
	LatestHumanVersion  string                          `json:"latest_human_version"`
	UpgradeVersion      string                          `json:"upgrade_version"`
	UpgradeHumanVersion string                          `json:"upgrade_human_version"`
	Changelog           *string                         `json:"changelog"`
	AvailableVersions   []AppAvailableVersionResponse   `json:"available_versions_for_upgrade"`
}

// AppAvailableVersionResponse is the wire-format for an available upgrade version.
type AppAvailableVersionResponse struct {
	Version      string `json:"version"`
	HumanVersion string `json:"human_version"`
}

// AppImageResponse is the wire-format for app.image.query results.
type AppImageResponse struct {
	ID       string   `json:"id"`
	RepoTags []string `json:"repo_tags"`
	Size     int64    `json:"size"`
	Created  string   `json:"created"`
	Dangling bool     `json:"dangling"`
}

// AppStatsResponse is the wire-format for app.stats events.
type AppStatsResponse struct {
	AppName  string                         `json:"app_name"`
	Memory   int64                          `json:"memory"`
	CPUUsage float64                        `json:"cpu_usage"`
	Blkio    AppStatsBlkioResponse          `json:"blkio"`
	Networks []AppStatsNetworkResponse      `json:"networks"`
}

// AppStatsBlkioResponse is the wire-format for block I/O stats.
type AppStatsBlkioResponse struct {
	Read  int64 `json:"read"`
	Write int64 `json:"write"`
}

// AppStatsNetworkResponse is the wire-format for per-interface network stats.
type AppStatsNetworkResponse struct {
	InterfaceName string `json:"interface_name"`
	RxBytes       int64  `json:"rx_bytes"`
	TxBytes       int64  `json:"tx_bytes"`
}

// AppContainerLogEntryResponse is the wire-format for container log events.
type AppContainerLogEntryResponse struct {
	Timestamp string `json:"timestamp"`
	Message   string `json:"data"`
}

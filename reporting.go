package truenas

import "encoding/json"

// ReportingGraphName represents standard graph identifiers.
type ReportingGraphName string

const (
	ReportingGraphCPU       ReportingGraphName = "cpu"
	ReportingGraphCPUTemp   ReportingGraphName = "cputemp"
	ReportingGraphMemory    ReportingGraphName = "memory"
	ReportingGraphDisk      ReportingGraphName = "disk"
	ReportingGraphDiskTemp  ReportingGraphName = "disktemp"
	ReportingGraphInterface ReportingGraphName = "interface"
	ReportingGraphArcSize   ReportingGraphName = "arcsize"
	ReportingGraphArcRate   ReportingGraphName = "arcrate"
	ReportingGraphUptime    ReportingGraphName = "uptime"
)

// ReportingGraphResponse represents a graph definition from reporting.netdata_graphs.
type ReportingGraphResponse struct {
	Name             string   `json:"name"`
	Title            string   `json:"title"`
	VerticalLabel    string   `json:"vertical_label"`
	Identifiers      []string `json:"identifiers"`
	Stacked          bool     `json:"stacked"`
	StackedShowTotal bool     `json:"stacked_show_total"`
}

// ReportingDataResponse represents data returned from reporting.netdata_get_data.
type ReportingDataResponse struct {
	Name         string          `json:"name"`
	Identifier   string          `json:"identifier"`
	Data         [][]json.Number `json:"data"`
	Start        int64           `json:"start"`
	End          int64           `json:"end"`
	Legend       []string        `json:"legend"`
	Aggregations struct {
		Min  map[string]json.Number `json:"min"`
		Max  map[string]json.Number `json:"max"`
		Mean map[string]json.Number `json:"mean"`
	} `json:"aggregations"`
}

// RealtimeUpdateResponse represents the wire-format for reporting.realtime events.
type RealtimeUpdateResponse struct {
	CPU        map[string]RealtimeCPUResponse       `json:"cpu"`
	Memory     RealtimeMemoryResponse               `json:"memory"`
	Disks      map[string]RealtimeDiskResponse      `json:"disks"`
	Interfaces map[string]RealtimeInterfaceResponse `json:"interfaces"`
}

// RealtimeCPUResponse is the wire-format for per-CPU metrics.
type RealtimeCPUResponse struct {
	Usage       float64 `json:"usage"`
	Temperature float64 `json:"temperature"`
}

// RealtimeMemoryResponse is the wire-format for memory metrics.
type RealtimeMemoryResponse struct {
	PhysicalMemoryTotal     int64 `json:"physical_memory_total"`
	PhysicalMemoryAvailable int64 `json:"physical_memory_available"`
	ArcSize                 int64 `json:"arc_size"`
}

// RealtimeDiskResponse is the wire-format for per-disk I/O metrics.
type RealtimeDiskResponse struct {
	ReadBytesPerSec  float64 `json:"read_bytes_per_sec"`
	WriteBytesPerSec float64 `json:"write_bytes_per_sec"`
	BusyPercent      float64 `json:"busy_percent"`
}

// RealtimeInterfaceResponse is the wire-format for per-interface network metrics.
type RealtimeInterfaceResponse struct {
	ReceivedBytesRate float64 `json:"received_bytes_rate"`
	SentBytesRate     float64 `json:"sent_bytes_rate"`
	LinkState         string  `json:"link_state"`
	Speed             int     `json:"speed"`
}

// ReportingGetDataParams contains parameters for the GetData call.
type ReportingGetDataParams struct {
	Graphs []ReportingGraphQuery
	Unit   string // "HOUR", "DAY", "WEEK", "MONTH", "YEAR"
	Page   int
}

// ReportingGraphQuery specifies a graph to query.
type ReportingGraphQuery struct {
	Name       ReportingGraphName
	Identifier string // e.g. disk name, interface name
}

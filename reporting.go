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

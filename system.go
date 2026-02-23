package truenas

// SystemInfoResponse represents the wire-format response from system.info.
type SystemInfoResponse struct {
	Model         string     `json:"model"`
	Cores         int        `json:"cores"`
	PhysicalCores int        `json:"physical_cores"`
	Hostname      string     `json:"hostname"`
	Uptime        string     `json:"uptime"`
	UptimeSeconds float64    `json:"uptime_seconds"`
	LoadAvg       [3]float64 `json:"loadavg"`
	EccMemory     bool       `json:"ecc_memory"`
}

package truenas

// VMResponse represents a VM from the TrueNAS API.
type VMResponse struct {
	ID              int64         `json:"id"`
	Name            string        `json:"name"`
	Description     string        `json:"description"`
	VCPUs           int64         `json:"vcpus"`
	Cores           int64         `json:"cores"`
	Threads         int64         `json:"threads"`
	Memory          int64         `json:"memory"`
	MinMemory       *int64        `json:"min_memory"`
	Autostart       bool          `json:"autostart"`
	Time            string        `json:"time"`
	Bootloader      string        `json:"bootloader"`
	BootloaderOVMF  string        `json:"bootloader_ovmf"`
	CPUMode         string        `json:"cpu_mode"`
	CPUModel        *string       `json:"cpu_model"`
	ShutdownTimeout int64         `json:"shutdown_timeout"`
	CommandLineArgs string        `json:"command_line_args"`
	Status          VMStatusField `json:"status"`
}

// VMStatusField represents the status of a VM.
type VMStatusField struct {
	State       string `json:"state"`
	PID         *int64 `json:"pid"`
	DomainState string `json:"domain_state"`
}

// VMDeviceResponse represents a VM device from the TrueNAS API.
type VMDeviceResponse struct {
	ID         int64          `json:"id"`
	VM         int64          `json:"vm"`
	Order      int64          `json:"order"`
	Attributes map[string]any `json:"attributes"`
}

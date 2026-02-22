package truenas

// AppResponse represents an app from the TrueNAS API.
type AppResponse struct {
	Name      string         `json:"name"`
	State     string         `json:"state"`
	CustomApp bool           `json:"custom_app"`
	Config    map[string]any `json:"config"`
}

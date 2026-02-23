package truenas

import (
	"context"
	"encoding/json"
	"fmt"
)

// App is the user-facing representation of a TrueNAS app.
type App struct {
	Name             string
	State            string
	CustomApp        bool
	Config           map[string]any
	Version          string
	HumanVersion     string
	LatestVersion    string
	UpgradeAvailable bool
	ActiveWorkloads  AppActiveWorkloads
}

// AppActiveWorkloads contains active workload information for an app.
type AppActiveWorkloads struct {
	Containers       int
	UsedPorts        []AppUsedPort
	ContainerDetails []AppContainerDetails
}

// AppUsedPort represents a port mapping for an app.
type AppUsedPort struct {
	ContainerPort int
	HostPort      int
	Protocol      string
}

// AppContainerDetails represents details of a container within an app.
type AppContainerDetails struct {
	ID          string
	ServiceName string
	Image       string
	State       ContainerState
}

// CreateAppOpts contains options for creating an app.
type CreateAppOpts struct {
	Name                string
	CustomApp           bool
	CustomComposeConfig string
}

// UpdateAppOpts contains options for updating an app.
type UpdateAppOpts struct {
	CustomComposeConfig string
}

// Registry is the user-facing representation of a TrueNAS app registry.
type Registry struct {
	ID          int64
	Name        string
	Description string
	Username    string
	Password    string
	URI         string
}

// CreateRegistryOpts contains options for creating a registry.
type CreateRegistryOpts struct {
	Name        string
	Description string
	Username    string
	Password    string
	URI         string
}

// UpdateRegistryOpts contains options for updating a registry.
type UpdateRegistryOpts = CreateRegistryOpts

// AppService provides typed methods for the app.* API namespace.
type AppService struct {
	client  AsyncCaller
	version Version
}

// NewAppService creates a new AppService.
func NewAppService(c AsyncCaller, v Version) *AppService {
	return &AppService{client: c, version: v}
}

// CreateApp creates an app and returns the full object.
func (s *AppService) CreateApp(ctx context.Context, opts CreateAppOpts) (*App, error) {
	params := createAppParams(opts)
	_, err := s.client.CallAndWait(ctx, "app.create", params)
	if err != nil {
		return nil, err
	}

	app, err := s.GetApp(ctx, opts.Name)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, fmt.Errorf("app %q not found after create", opts.Name)
	}
	return app, nil
}

// GetApp returns an app by name, or nil if not found.
func (s *AppService) GetApp(ctx context.Context, name string) (*App, error) {
	filter := [][]any{{"name", "=", name}}
	result, err := s.client.Call(ctx, "app.query", filter)
	if err != nil {
		return nil, err
	}

	var apps []AppResponse
	if err := json.Unmarshal(result, &apps); err != nil {
		return nil, fmt.Errorf("parse query response: %w", err)
	}

	if len(apps) == 0 {
		return nil, nil
	}

	app := appFromResponse(apps[0])
	return &app, nil
}

// GetAppWithConfig returns an app by name with its config populated, or nil if not found.
func (s *AppService) GetAppWithConfig(ctx context.Context, name string) (*App, error) {
	filter := [][]any{{"name", "=", name}}
	params := []any{filter, map[string]any{"extra": map[string]any{"retrieve_config": true}}}
	result, err := s.client.Call(ctx, "app.query", params)
	if err != nil {
		return nil, err
	}

	var apps []AppResponse
	if err := json.Unmarshal(result, &apps); err != nil {
		return nil, fmt.Errorf("parse query response: %w", err)
	}

	if len(apps) == 0 {
		return nil, nil
	}

	app := appFromResponse(apps[0])
	return &app, nil
}

// UpdateApp updates an app and returns the full object.
func (s *AppService) UpdateApp(ctx context.Context, name string, opts UpdateAppOpts) (*App, error) {
	params := []any{name, updateAppParams(opts)}
	_, err := s.client.CallAndWait(ctx, "app.update", params)
	if err != nil {
		return nil, err
	}

	app, err := s.GetApp(ctx, name)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, fmt.Errorf("app %q not found after update", name)
	}
	return app, nil
}

// ListApps returns all apps.
func (s *AppService) ListApps(ctx context.Context) ([]App, error) {
	result, err := s.client.Call(ctx, "app.query", nil)
	if err != nil {
		return nil, err
	}

	var responses []AppResponse
	if err := json.Unmarshal(result, &responses); err != nil {
		return nil, fmt.Errorf("parse query response: %w", err)
	}

	apps := make([]App, len(responses))
	for i, resp := range responses {
		apps[i] = appFromResponse(resp)
	}
	return apps, nil
}

// StartApp starts an app by name.
func (s *AppService) StartApp(ctx context.Context, name string) error {
	_, err := s.client.CallAndWait(ctx, "app.start", name)
	return err
}

// StopApp stops an app by name.
func (s *AppService) StopApp(ctx context.Context, name string) error {
	_, err := s.client.CallAndWait(ctx, "app.stop", name)
	return err
}

// DeleteApp deletes an app by name.
func (s *AppService) DeleteApp(ctx context.Context, name string) error {
	_, err := s.client.CallAndWait(ctx, "app.delete", name)
	return err
}

// CreateRegistry creates a registry and returns the full object.
func (s *AppService) CreateRegistry(ctx context.Context, opts CreateRegistryOpts) (*Registry, error) {
	params := registryParams(opts)
	result, err := s.client.Call(ctx, "app.registry.create", params)
	if err != nil {
		return nil, err
	}

	var createResp struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(result, &createResp); err != nil {
		return nil, fmt.Errorf("parse create response: %w", err)
	}

	reg, err := s.GetRegistry(ctx, createResp.ID)
	if err != nil {
		return nil, err
	}
	if reg == nil {
		return nil, fmt.Errorf("registry %d not found after create", createResp.ID)
	}
	return reg, nil
}

// GetRegistry returns a registry by ID, or nil if not found.
func (s *AppService) GetRegistry(ctx context.Context, id int64) (*Registry, error) {
	filter := [][]any{{"id", "=", id}}
	result, err := s.client.Call(ctx, "app.registry.query", filter)
	if err != nil {
		return nil, err
	}

	var registries []AppRegistryResponse
	if err := json.Unmarshal(result, &registries); err != nil {
		return nil, fmt.Errorf("parse query response: %w", err)
	}

	if len(registries) == 0 {
		return nil, nil
	}

	reg := registryFromResponse(registries[0])
	return &reg, nil
}

// ListRegistries returns all registries.
func (s *AppService) ListRegistries(ctx context.Context) ([]Registry, error) {
	result, err := s.client.Call(ctx, "app.registry.query", nil)
	if err != nil {
		return nil, err
	}

	var responses []AppRegistryResponse
	if err := json.Unmarshal(result, &responses); err != nil {
		return nil, fmt.Errorf("parse query response: %w", err)
	}

	registries := make([]Registry, len(responses))
	for i, resp := range responses {
		registries[i] = registryFromResponse(resp)
	}
	return registries, nil
}

// UpdateRegistry updates a registry and returns the full object.
func (s *AppService) UpdateRegistry(ctx context.Context, id int64, opts UpdateRegistryOpts) (*Registry, error) {
	params := registryParams(opts)
	_, err := s.client.Call(ctx, "app.registry.update", []any{id, params})
	if err != nil {
		return nil, err
	}

	reg, err := s.GetRegistry(ctx, id)
	if err != nil {
		return nil, err
	}
	if reg == nil {
		return nil, fmt.Errorf("registry %d not found after update", id)
	}
	return reg, nil
}

// DeleteRegistry deletes a registry by ID.
func (s *AppService) DeleteRegistry(ctx context.Context, id int64) error {
	_, err := s.client.Call(ctx, "app.registry.delete", id)
	return err
}

// createAppParams converts CreateAppOpts to API parameters.
func createAppParams(opts CreateAppOpts) map[string]any {
	params := map[string]any{
		"app_name":   opts.Name,
		"custom_app": opts.CustomApp,
	}
	if opts.CustomComposeConfig != "" {
		params["custom_compose_config_string"] = opts.CustomComposeConfig
	}
	return params
}

// updateAppParams converts UpdateAppOpts to API parameters.
func updateAppParams(opts UpdateAppOpts) map[string]any {
	params := map[string]any{}
	if opts.CustomComposeConfig != "" {
		params["custom_compose_config_string"] = opts.CustomComposeConfig
	}
	return params
}

// registryParams converts CreateRegistryOpts to API parameters.
func registryParams(opts CreateRegistryOpts) map[string]any {
	params := map[string]any{
		"name":     opts.Name,
		"username": opts.Username,
		"password": opts.Password,
		"uri":      opts.URI,
	}
	if opts.Description != "" {
		params["description"] = opts.Description
	} else {
		params["description"] = nil
	}
	return params
}

// appFromResponse converts a wire-format AppResponse to a user-facing App.
func appFromResponse(resp AppResponse) App {
	ports := make([]AppUsedPort, len(resp.ActiveWorkloads.UsedPorts))
	for i, p := range resp.ActiveWorkloads.UsedPorts {
		ports[i] = AppUsedPort{
			ContainerPort: p.ContainerPort,
			HostPort:      p.HostPort,
			Protocol:      p.Protocol,
		}
	}

	containers := make([]AppContainerDetails, len(resp.ActiveWorkloads.ContainerDetails))
	for i, c := range resp.ActiveWorkloads.ContainerDetails {
		containers[i] = AppContainerDetails{
			ID:          c.ID,
			ServiceName: c.ServiceName,
			Image:       c.Image,
			State:       ContainerState(c.State),
		}
	}

	return App{
		Name:             resp.Name,
		State:            resp.State,
		CustomApp:        resp.CustomApp,
		Config:           resp.Config,
		Version:          resp.Version,
		HumanVersion:     resp.HumanVersion,
		LatestVersion:    resp.LatestVersion,
		UpgradeAvailable: resp.UpgradeAvailable,
		ActiveWorkloads: AppActiveWorkloads{
			Containers:       resp.ActiveWorkloads.Containers,
			UsedPorts:        ports,
			ContainerDetails: containers,
		},
	}
}

// registryFromResponse converts a wire-format AppRegistryResponse to a user-facing Registry.
func registryFromResponse(resp AppRegistryResponse) Registry {
	desc := ""
	if resp.Description != nil {
		desc = *resp.Description
	}
	return Registry{
		ID:          resp.ID,
		Name:        resp.Name,
		Description: desc,
		Username:    resp.Username,
		Password:    resp.Password,
		URI:         resp.URI,
	}
}

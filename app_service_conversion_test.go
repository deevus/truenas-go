package truenas

import "testing"

func TestAppFromResponse(t *testing.T) {
	resp := AppResponse{
		Name:      "test-app",
		State:     "STOPPED",
		CustomApp: true,
		Config:    map[string]any{"key": "value"},
	}

	app := appFromResponse(resp)

	if app.Name != "test-app" {
		t.Errorf("expected name test-app, got %s", app.Name)
	}
	if app.State != "STOPPED" {
		t.Errorf("expected state STOPPED, got %s", app.State)
	}
	if !app.CustomApp {
		t.Error("expected custom_app true")
	}
	if app.Config["key"] != "value" {
		t.Errorf("expected config key=value, got %v", app.Config["key"])
	}
}

func TestAppFromResponse_NilConfig(t *testing.T) {
	resp := AppResponse{
		Name:      "no-config",
		State:     "RUNNING",
		CustomApp: false,
	}

	app := appFromResponse(resp)

	if app.Name != "no-config" {
		t.Errorf("expected name no-config, got %s", app.Name)
	}
	if app.Config != nil {
		t.Errorf("expected nil config, got %v", app.Config)
	}
}

func TestAppFromResponse_WithWorkloads(t *testing.T) {
	resp := AppResponse{
		Name:             "plex",
		State:            "RUNNING",
		Version:          "1.0.0",
		HumanVersion:     "Plex 1.0.0",
		LatestVersion:    "1.1.0",
		UpgradeAvailable: true,
		ActiveWorkloads: AppActiveWorkloadsResponse{
			Containers: 2,
			UsedPorts: []AppUsedPortResponse{
				{ContainerPort: 32400, HostPort: 32400, Protocol: "tcp"},
			},
			ContainerDetails: []AppContainerDetailsResponse{
				{
					ID:          "abc123",
					ServiceName: "plex",
					Image:       "plexinc/pms-docker:latest",
					State:       "running",
				},
			},
		},
	}

	app := appFromResponse(resp)
	if app.Version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", app.Version)
	}
	if app.HumanVersion != "Plex 1.0.0" {
		t.Errorf("expected human version 'Plex 1.0.0', got %s", app.HumanVersion)
	}
	if app.LatestVersion != "1.1.0" {
		t.Errorf("expected latest version 1.1.0, got %s", app.LatestVersion)
	}
	if !app.UpgradeAvailable {
		t.Error("expected UpgradeAvailable=true")
	}
	if app.ActiveWorkloads.Containers != 2 {
		t.Errorf("expected 2 containers, got %d", app.ActiveWorkloads.Containers)
	}
	if len(app.ActiveWorkloads.UsedPorts) != 1 {
		t.Fatalf("expected 1 used port, got %d", len(app.ActiveWorkloads.UsedPorts))
	}
	if app.ActiveWorkloads.UsedPorts[0].ContainerPort != 32400 {
		t.Errorf("expected container port 32400, got %d", app.ActiveWorkloads.UsedPorts[0].ContainerPort)
	}
	if len(app.ActiveWorkloads.ContainerDetails) != 1 {
		t.Fatalf("expected 1 container detail, got %d", len(app.ActiveWorkloads.ContainerDetails))
	}
	if app.ActiveWorkloads.ContainerDetails[0].State != ContainerStateRunning {
		t.Errorf("expected running state, got %s", app.ActiveWorkloads.ContainerDetails[0].State)
	}
	if app.ActiveWorkloads.ContainerDetails[0].Image != "plexinc/pms-docker:latest" {
		t.Errorf("expected image plexinc/pms-docker:latest, got %s", app.ActiveWorkloads.ContainerDetails[0].Image)
	}
}

func TestAppFromResponse_CatalogApp(t *testing.T) {
	resp := AppResponse{
		Name:      "tailscale",
		State:     "RUNNING",
		CustomApp: false,
		Version:   "1.3.32",
		Metadata: AppMetadataResponse{
			Name:  "tailscale",
			Train: "community",
		},
	}

	app := appFromResponse(resp)

	if app.CatalogApp != "tailscale" {
		t.Errorf("expected CatalogApp 'tailscale', got %q", app.CatalogApp)
	}
	if app.Train != "community" {
		t.Errorf("expected Train 'community', got %q", app.Train)
	}
	if app.CustomApp {
		t.Error("expected CustomApp false")
	}
	if app.Version != "1.3.32" {
		t.Errorf("expected Version '1.3.32', got %q", app.Version)
	}
}

func TestCreateAppParams_CatalogApp(t *testing.T) {
	opts := CreateAppOpts{
		Name:       "tailscale",
		CatalogApp: "tailscale",
		Train:      "community",
	}

	params, err := createAppParams(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if params["app_name"] != "tailscale" {
		t.Errorf("expected app_name 'tailscale', got %v", params["app_name"])
	}
	if params["custom_app"] != false {
		t.Errorf("expected custom_app false, got %v", params["custom_app"])
	}
	if params["catalog_app"] != "tailscale" {
		t.Errorf("expected catalog_app 'tailscale', got %v", params["catalog_app"])
	}
	if params["train"] != "community" {
		t.Errorf("expected train 'community', got %v", params["train"])
	}
	if _, ok := params["custom_compose_config_string"]; ok {
		t.Error("expected no custom_compose_config_string for catalog app")
	}
}

func TestCreateAppParams_CatalogAppWithValues(t *testing.T) {
	opts := CreateAppOpts{
		Name:       "tailscale",
		CatalogApp: "tailscale",
		Train:      "community",
		Version:    "1.3.32",
		Values:     map[string]any{"tailscale": map[string]any{"auth_key": "tskey-xxx"}},
	}

	params, err := createAppParams(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if params["catalog_app"] != "tailscale" {
		t.Errorf("expected catalog_app 'tailscale', got %v", params["catalog_app"])
	}
	if params["version"] != "1.3.32" {
		t.Errorf("expected version '1.3.32', got %v", params["version"])
	}
	values, ok := params["values"].(map[string]any)
	if !ok {
		t.Fatalf("expected values map, got %T", params["values"])
	}
	ts, ok := values["tailscale"].(map[string]any)
	if !ok {
		t.Fatalf("expected tailscale values map, got %T", values["tailscale"])
	}
	if ts["auth_key"] != "tskey-xxx" {
		t.Errorf("expected auth_key 'tskey-xxx', got %v", ts["auth_key"])
	}
}

func TestUpdateAppParams_WithValues(t *testing.T) {
	opts := UpdateAppOpts{
		Values: map[string]any{"TZ": "America/New_York"},
	}

	params := updateAppParams(opts)

	values, ok := params["values"].(map[string]any)
	if !ok {
		t.Fatalf("expected values map, got %T", params["values"])
	}
	if values["TZ"] != "America/New_York" {
		t.Errorf("expected TZ 'America/New_York', got %v", values["TZ"])
	}
}

func TestCreateAppParams_CatalogAndCustomAppError(t *testing.T) {
	_, err := createAppParams(CreateAppOpts{
		Name:       "confused",
		CustomApp:  true,
		CatalogApp: "plex",
	})
	if err == nil {
		t.Fatal("expected error when both CatalogApp and CustomApp are set")
	}
}

func TestCreateAppParams_CatalogAndComposeConfigError(t *testing.T) {
	_, err := createAppParams(CreateAppOpts{
		Name:                "confused",
		CatalogApp:          "plex",
		CustomComposeConfig: "services:\n  web:\n    image: nginx",
	})
	if err == nil {
		t.Fatal("expected error when both CatalogApp and CustomComposeConfig are set")
	}
}

func TestCreateAppParams_NeitherCatalogNorCustomError(t *testing.T) {
	_, err := createAppParams(CreateAppOpts{
		Name: "nothing",
	})
	if err == nil {
		t.Fatal("expected error when neither CatalogApp nor CustomApp is set")
	}
}

func TestCreateAppParams_ValuesWithCustomAppError(t *testing.T) {
	_, err := createAppParams(CreateAppOpts{
		Name:      "bad",
		CustomApp: true,
		Values:    map[string]any{"TZ": "UTC"},
	})
	if err == nil {
		t.Fatal("expected error when Values is used with CustomApp")
	}
}

func TestRegistryFromResponse(t *testing.T) {
	desc := "A registry"
	resp := AppRegistryResponse{
		ID:          1,
		Name:        "test-reg",
		Description: &desc,
		Username:    "user",
		Password:    "pass",
		URI:         "https://example.com",
	}

	reg := registryFromResponse(resp)

	if reg.ID != 1 {
		t.Errorf("expected ID 1, got %d", reg.ID)
	}
	if reg.Name != "test-reg" {
		t.Errorf("expected name test-reg, got %s", reg.Name)
	}
	if reg.Description != "A registry" {
		t.Errorf("expected description 'A registry', got %s", reg.Description)
	}
	if reg.Username != "user" {
		t.Errorf("expected username user, got %s", reg.Username)
	}
	if reg.Password != "pass" {
		t.Errorf("expected password pass, got %s", reg.Password)
	}
	if reg.URI != "https://example.com" {
		t.Errorf("expected uri https://example.com, got %s", reg.URI)
	}
}

func TestRegistryFromResponse_NilDescription(t *testing.T) {
	resp := AppRegistryResponse{
		ID:          2,
		Name:        "null-desc",
		Description: nil,
		Username:    "user",
		Password:    "pass",
		URI:         "https://example.com",
	}

	reg := registryFromResponse(resp)

	if reg.Description != "" {
		t.Errorf("expected empty description for nil, got %q", reg.Description)
	}
}

func TestCreateAppParams(t *testing.T) {
	opts := CreateAppOpts{
		Name:                "my-app",
		CustomApp:           true,
		CustomComposeConfig: "version: '3'",
	}

	params, err := createAppParams(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if params["app_name"] != "my-app" {
		t.Errorf("expected app_name my-app, got %v", params["app_name"])
	}
	if params["custom_app"] != true {
		t.Errorf("expected custom_app true, got %v", params["custom_app"])
	}
	if params["custom_compose_config_string"] != "version: '3'" {
		t.Errorf("expected custom_compose_config_string, got %v", params["custom_compose_config_string"])
	}
}

func TestCreateAppParams_CustomAppNoCompose(t *testing.T) {
	opts := CreateAppOpts{
		Name:      "simple-app",
		CustomApp: true,
	}

	params, err := createAppParams(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if params["app_name"] != "simple-app" {
		t.Errorf("expected app_name simple-app, got %v", params["app_name"])
	}
	if params["custom_app"] != true {
		t.Errorf("expected custom_app true, got %v", params["custom_app"])
	}
	if _, ok := params["custom_compose_config_string"]; ok {
		t.Error("expected no custom_compose_config_string for empty config")
	}
}

func TestUpdateAppParams(t *testing.T) {
	opts := UpdateAppOpts{
		CustomComposeConfig: "version: '3.1'",
	}

	params := updateAppParams(opts)

	if params["custom_compose_config_string"] != "version: '3.1'" {
		t.Errorf("expected custom_compose_config_string, got %v", params["custom_compose_config_string"])
	}
}

func TestUpdateAppParams_Empty(t *testing.T) {
	opts := UpdateAppOpts{}

	params := updateAppParams(opts)

	if _, ok := params["custom_compose_config_string"]; ok {
		t.Error("expected no custom_compose_config_string for empty opts")
	}
	if len(params) != 0 {
		t.Errorf("expected empty params map, got %v", params)
	}
}

func TestRegistryParams(t *testing.T) {
	opts := CreateRegistryOpts{
		Name:        "my-reg",
		Description: "A description",
		Username:    "admin",
		Password:    "secret",
		URI:         "https://example.com",
	}

	params := registryParams(opts)

	if params["name"] != "my-reg" {
		t.Errorf("expected name my-reg, got %v", params["name"])
	}
	if params["description"] != "A description" {
		t.Errorf("expected description 'A description', got %v", params["description"])
	}
	if params["username"] != "admin" {
		t.Errorf("expected username admin, got %v", params["username"])
	}
	if params["password"] != "secret" {
		t.Errorf("expected password secret, got %v", params["password"])
	}
	if params["uri"] != "https://example.com" {
		t.Errorf("expected uri https://example.com, got %v", params["uri"])
	}
}

func TestRegistryParams_EmptyDescription(t *testing.T) {
	opts := CreateRegistryOpts{
		Name:     "no-desc",
		Username: "user",
		Password: "pass",
		URI:      "https://example.com",
	}

	params := registryParams(opts)

	if params["description"] != nil {
		t.Errorf("expected nil description for empty string, got %v", params["description"])
	}
}

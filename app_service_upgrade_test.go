package truenas

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestAppService_UpgradeSummary(t *testing.T) {
	mock := &mockSubscribeCaller{}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		if method != "app.upgrade_summary" {
			t.Errorf("expected method app.upgrade_summary, got %s", method)
		}
		slice, ok := params.([]any)
		if !ok {
			t.Fatal("expected []any params")
		}
		if len(slice) != 1 || slice[0] != "plex" {
			t.Errorf("expected [plex], got %v", slice)
		}
		return json.RawMessage(`{
			"latest_version": "1.2.0",
			"latest_human_version": "Plex 1.2.0",
			"upgrade_version": "1.1.0",
			"upgrade_human_version": "Plex 1.1.0",
			"changelog": "Bug fixes and improvements",
			"available_versions_for_upgrade": [
				{"version": "1.1.0", "human_version": "Plex 1.1.0"},
				{"version": "1.2.0", "human_version": "Plex 1.2.0"}
			]
		}`), nil
	}

	svc := NewAppService(mock, Version{})
	summary, err := svc.UpgradeSummary(context.Background(), "plex")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary == nil {
		t.Fatal("expected non-nil summary")
	}
	if summary.LatestVersion != "1.2.0" {
		t.Errorf("expected latest version 1.2.0, got %s", summary.LatestVersion)
	}
	if summary.LatestHumanVersion != "Plex 1.2.0" {
		t.Errorf("expected latest human version 'Plex 1.2.0', got %s", summary.LatestHumanVersion)
	}
	if summary.UpgradeVersion != "1.1.0" {
		t.Errorf("expected upgrade version 1.1.0, got %s", summary.UpgradeVersion)
	}
	if summary.UpgradeHumanVersion != "Plex 1.1.0" {
		t.Errorf("expected upgrade human version 'Plex 1.1.0', got %s", summary.UpgradeHumanVersion)
	}
	if summary.Changelog != "Bug fixes and improvements" {
		t.Errorf("expected changelog 'Bug fixes and improvements', got %s", summary.Changelog)
	}
	if len(summary.AvailableVersions) != 2 {
		t.Fatalf("expected 2 available versions, got %d", len(summary.AvailableVersions))
	}
	if summary.AvailableVersions[0].Version != "1.1.0" {
		t.Errorf("expected first version 1.1.0, got %s", summary.AvailableVersions[0].Version)
	}
	if summary.AvailableVersions[1].Version != "1.2.0" {
		t.Errorf("expected second version 1.2.0, got %s", summary.AvailableVersions[1].Version)
	}
}

func TestAppService_UpgradeSummary_Error(t *testing.T) {
	mock := &mockSubscribeCaller{}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		return nil, errors.New("app not found")
	}

	svc := NewAppService(mock, Version{})
	summary, err := svc.UpgradeSummary(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
	if summary != nil {
		t.Error("expected nil summary on error")
	}
}

func TestAppService_ListImages(t *testing.T) {
	mock := &mockSubscribeCaller{}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		if method != "app.image.query" {
			t.Errorf("expected method app.image.query, got %s", method)
		}
		if params != nil {
			t.Error("expected nil params for ListImages")
		}
		return json.RawMessage(`[
			{
				"id": "sha256:abc123",
				"repo_tags": ["nginx:latest", "nginx:1.25"],
				"size": 187654321,
				"created": "2024-01-15T10:30:00Z",
				"dangling": false
			},
			{
				"id": "sha256:def456",
				"repo_tags": [],
				"size": 52345678,
				"created": "2024-01-10T08:00:00Z",
				"dangling": true
			}
		]`), nil
	}

	svc := NewAppService(mock, Version{})
	images, err := svc.ListImages(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(images) != 2 {
		t.Fatalf("expected 2 images, got %d", len(images))
	}
	if images[0].ID != "sha256:abc123" {
		t.Errorf("expected first image ID sha256:abc123, got %s", images[0].ID)
	}
	if len(images[0].RepoTags) != 2 {
		t.Fatalf("expected 2 repo tags, got %d", len(images[0].RepoTags))
	}
	if images[0].RepoTags[0] != "nginx:latest" {
		t.Errorf("expected first tag nginx:latest, got %s", images[0].RepoTags[0])
	}
	if images[0].Size != 187654321 {
		t.Errorf("expected size 187654321, got %d", images[0].Size)
	}
	if images[0].Created != "2024-01-15T10:30:00Z" {
		t.Errorf("expected created 2024-01-15T10:30:00Z, got %s", images[0].Created)
	}
	if images[0].Dangling {
		t.Error("expected first image not dangling")
	}
	if !images[1].Dangling {
		t.Error("expected second image dangling")
	}
	if len(images[1].RepoTags) != 0 {
		t.Errorf("expected empty repo tags for dangling image, got %v", images[1].RepoTags)
	}
}

func TestAppService_ListImages_Empty(t *testing.T) {
	mock := &mockSubscribeCaller{}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		return json.RawMessage(`[]`), nil
	}

	svc := NewAppService(mock, Version{})
	images, err := svc.ListImages(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(images) != 0 {
		t.Errorf("expected 0 images, got %d", len(images))
	}
}

func TestAppService_ListImages_Error(t *testing.T) {
	mock := &mockSubscribeCaller{}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		return nil, errors.New("network error")
	}

	svc := NewAppService(mock, Version{})
	_, err := svc.ListImages(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAppService_AvailableSpace(t *testing.T) {
	mock := &mockSubscribeCaller{}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		if method != "app.available_space" {
			t.Errorf("expected method app.available_space, got %s", method)
		}
		if params != nil {
			t.Error("expected nil params for AvailableSpace")
		}
		return json.RawMessage(`1099511627776`), nil
	}

	svc := NewAppService(mock, Version{})
	space, err := svc.AvailableSpace(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if space != 1099511627776 {
		t.Errorf("expected 1099511627776 bytes, got %d", space)
	}
}

func TestAppService_AvailableSpace_Error(t *testing.T) {
	mock := &mockSubscribeCaller{}
	mock.callFunc = func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		return nil, errors.New("pool not found")
	}

	svc := NewAppService(mock, Version{})
	space, err := svc.AvailableSpace(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if space != 0 {
		t.Errorf("expected 0 on error, got %d", space)
	}
}

func TestAppService_UpgradeApp(t *testing.T) {
	mock := &mockSubscribeCaller{mockAsyncCaller: mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "app.upgrade" {
				t.Errorf("expected method app.upgrade, got %s", method)
			}
			slice, ok := params.([]any)
			if !ok {
				t.Fatal("expected []any params")
			}
			if len(slice) != 1 || slice[0] != "plex" {
				t.Errorf("expected [plex], got %v", slice)
			}
			return nil, nil
		},
	}}

	svc := NewAppService(mock, Version{})
	err := svc.UpgradeApp(context.Background(), "plex")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAppService_UpgradeApp_Error(t *testing.T) {
	mock := &mockSubscribeCaller{mockAsyncCaller: mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("upgrade failed")
		},
	}}

	svc := NewAppService(mock, Version{})
	err := svc.UpgradeApp(context.Background(), "plex")
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "upgrade failed" {
		t.Errorf("expected 'upgrade failed', got %q", err.Error())
	}
}

func TestAppService_RedeployApp(t *testing.T) {
	mock := &mockSubscribeCaller{mockAsyncCaller: mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "app.redeploy" {
				t.Errorf("expected method app.redeploy, got %s", method)
			}
			name, ok := params.(string)
			if !ok || name != "plex" {
				t.Errorf("expected name plex, got %v", params)
			}
			return nil, nil
		},
	}}

	svc := NewAppService(mock, Version{})
	err := svc.RedeployApp(context.Background(), "plex")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAppService_RedeployApp_Error(t *testing.T) {
	mock := &mockSubscribeCaller{mockAsyncCaller: mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("redeploy failed")
		},
	}}

	svc := NewAppService(mock, Version{})
	err := svc.RedeployApp(context.Background(), "plex")
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "redeploy failed" {
		t.Errorf("expected 'redeploy failed', got %q", err.Error())
	}
}

func TestAppUpgradeSummaryFromResponse(t *testing.T) {
	changelog := "Major update"
	resp := AppUpgradeSummaryResponse{
		LatestVersion:       "2.0.0",
		LatestHumanVersion:  "App 2.0.0",
		UpgradeVersion:      "1.5.0",
		UpgradeHumanVersion: "App 1.5.0",
		Changelog:           &changelog,
		AvailableVersions: []AppAvailableVersionResponse{
			{Version: "1.5.0", HumanVersion: "App 1.5.0"},
			{Version: "2.0.0", HumanVersion: "App 2.0.0"},
		},
	}

	summary := appUpgradeSummaryFromResponse(resp)

	if summary.LatestVersion != "2.0.0" {
		t.Errorf("expected latest version 2.0.0, got %s", summary.LatestVersion)
	}
	if summary.LatestHumanVersion != "App 2.0.0" {
		t.Errorf("expected latest human version 'App 2.0.0', got %s", summary.LatestHumanVersion)
	}
	if summary.UpgradeVersion != "1.5.0" {
		t.Errorf("expected upgrade version 1.5.0, got %s", summary.UpgradeVersion)
	}
	if summary.UpgradeHumanVersion != "App 1.5.0" {
		t.Errorf("expected upgrade human version 'App 1.5.0', got %s", summary.UpgradeHumanVersion)
	}
	if summary.Changelog != "Major update" {
		t.Errorf("expected changelog 'Major update', got %s", summary.Changelog)
	}
	if len(summary.AvailableVersions) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(summary.AvailableVersions))
	}
	if summary.AvailableVersions[0].Version != "1.5.0" {
		t.Errorf("expected first version 1.5.0, got %s", summary.AvailableVersions[0].Version)
	}
}

func TestAppUpgradeSummaryFromResponse_NilChangelog(t *testing.T) {
	resp := AppUpgradeSummaryResponse{
		LatestVersion:       "2.0.0",
		LatestHumanVersion:  "App 2.0.0",
		UpgradeVersion:      "1.5.0",
		UpgradeHumanVersion: "App 1.5.0",
		Changelog:           nil,
		AvailableVersions:   []AppAvailableVersionResponse{},
	}

	summary := appUpgradeSummaryFromResponse(resp)
	if summary.Changelog != "" {
		t.Errorf("expected empty changelog for nil, got %q", summary.Changelog)
	}
}

func TestAppImageFromResponse(t *testing.T) {
	resp := AppImageResponse{
		ID:       "sha256:abc",
		RepoTags: []string{"test:latest"},
		Size:     12345,
		Created:  "2024-06-01T00:00:00Z",
		Dangling: true,
	}

	image := appImageFromResponse(resp)

	if image.ID != "sha256:abc" {
		t.Errorf("expected ID sha256:abc, got %s", image.ID)
	}
	if len(image.RepoTags) != 1 || image.RepoTags[0] != "test:latest" {
		t.Errorf("expected repo tags [test:latest], got %v", image.RepoTags)
	}
	if image.Size != 12345 {
		t.Errorf("expected size 12345, got %d", image.Size)
	}
	if image.Created != "2024-06-01T00:00:00Z" {
		t.Errorf("expected created 2024-06-01T00:00:00Z, got %s", image.Created)
	}
	if !image.Dangling {
		t.Error("expected dangling true")
	}
}

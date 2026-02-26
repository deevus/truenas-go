package api

import (
	"testing"
)

func TestVersions(t *testing.T) {
	vs := Versions()
	if len(vs) == 0 {
		t.Fatal("expected at least one embedded version")
	}
	if vs[0] != "25.04" {
		t.Errorf("expected first version 25.04, got %s", vs[0])
	}
}

func TestLatestVersion(t *testing.T) {
	v := LatestVersion()
	if v == "" {
		t.Fatal("expected non-empty latest version")
	}
	if v != "25.04" {
		t.Errorf("expected latest version 25.04, got %s", v)
	}
}

func TestMethods(t *testing.T) {
	methods, err := Methods("25.04")
	if err != nil {
		t.Fatalf("Methods(25.04) error: %v", err)
	}
	if len(methods) == 0 {
		t.Fatal("expected non-empty methods map")
	}

	// Spot-check a known method
	m, ok := methods["system.info"]
	if !ok {
		t.Fatal("expected system.info to be present")
	}
	if m.Description == nil {
		t.Error("expected system.info to have a description")
	}
}

func TestMethods_InvalidVersion(t *testing.T) {
	_, err := Methods("99.99")
	if err == nil {
		t.Fatal("expected error for invalid version")
	}
}

func TestNamespace(t *testing.T) {
	tests := []struct {
		method string
		want   string
	}{
		{"app.create", "app"},
		{"app.registry.create", "app.registry"},
		{"system.info", "system"},
		{"pool.dataset.query", "pool.dataset"},
		{"standalone", "standalone"},
	}
	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			got := Namespace(tt.method)
			if got != tt.want {
				t.Errorf("Namespace(%q) = %q, want %q", tt.method, got, tt.want)
			}
		})
	}
}

func TestMethods_KnownFields(t *testing.T) {
	methods, err := Methods("25.04")
	if err != nil {
		t.Fatalf("Methods(25.04) error: %v", err)
	}

	// Check a job method
	if m, ok := methods["app.create"]; ok {
		if !m.Job {
			t.Error("expected app.create to be a job method")
		}
	}

	// Check a filterable method
	if m, ok := methods["app.query"]; ok {
		if !m.Filterable {
			t.Error("expected app.query to be filterable")
		}
	}
}

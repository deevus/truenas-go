package client

import (
	"strings"
	"testing"
)

func TestBuildCommand_NoParams(t *testing.T) {
	got := BuildCommand("system.info", nil)
	want := "midclt call system.info"

	if got != want {
		t.Errorf("BuildCommand() = %q, want %q", got, want)
	}
}

func TestBuildCommand_WithArray(t *testing.T) {
	// Array filter params like [["id", "=", 123]]
	params := [][]any{{"id", "=", 123}}

	got := BuildCommand("pool.dataset.query", params)
	// shellescape.Quote wraps in single quotes
	want := `midclt call pool.dataset.query '[["id","=",123]]'`

	if got != want {
		t.Errorf("BuildCommand() = %q, want %q", got, want)
	}
}

func TestBuildCommand_WithObject(t *testing.T) {
	// Object/map params
	params := map[string]any{
		"name": "tank/test",
		"type": "FILESYSTEM",
	}

	got := BuildCommand("pool.dataset.create", params)

	// The order of map keys is not guaranteed, so we need to check both possibilities
	want1 := `midclt call pool.dataset.create '{"name":"tank/test","type":"FILESYSTEM"}'`
	want2 := `midclt call pool.dataset.create '{"type":"FILESYSTEM","name":"tank/test"}'`

	if got != want1 && got != want2 {
		t.Errorf("BuildCommand() = %q, want %q or %q", got, want1, want2)
	}
}

func TestBuildCommand_StringParam(t *testing.T) {
	// Single string param
	params := "tank/dataset"

	got := BuildCommand("pool.dataset.get_instance", params)
	want := `midclt call pool.dataset.get_instance '"tank/dataset"'`

	if got != want {
		t.Errorf("BuildCommand() = %q, want %q", got, want)
	}
}

func TestBuildCommand_AppCreateParams(t *testing.T) {
	params := AppCreateParams{
		AppName:                   "myapp",
		CustomApp:                 true,
		CustomComposeConfigString: "services:\n  web:\n    image: nginx",
	}

	got := BuildCommand("app.create", params)

	// Check that it contains the expected JSON structure
	if got == "" {
		t.Error("BuildCommand() returned empty string")
	}

	// The command should start with the midclt prefix
	if len(got) < 20 || got[:20] != "midclt call app.crea" {
		t.Errorf("BuildCommand() = %q, expected to start with 'midclt call app.crea'", got)
	}
}

func TestBuildCommand_DatasetCreateParams(t *testing.T) {
	params := DatasetCreateParams{
		Name:        "tank/test",
		Type:        "FILESYSTEM",
		Compression: "lz4",
		Quota:       1073741824,
	}

	got := BuildCommand("pool.dataset.create", params)

	// The command should start with the midclt prefix
	if len(got) < 30 || got[:30] != "midclt call pool.dataset.creat" {
		t.Errorf("BuildCommand() = %q, expected to start with 'midclt call pool.dataset.creat'", got)
	}
}

func TestBuildCommand_SpecialCharacters(t *testing.T) {
	// Test that special characters are properly escaped
	params := map[string]string{
		"path": "/mnt/tank/test's data",
	}

	got := BuildCommand("test.method", params)

	// shellescape wraps in single quotes and escapes internal single quotes
	// The internal single quote becomes '"'"' (end single quote, double-quoted single quote, start single quote)
	want := `midclt call test.method '{"path":"/mnt/tank/test'"'"'s data"}'`

	if got != want {
		t.Errorf("BuildCommand() = %q, want %q", got, want)
	}
}

func TestBuildCommand_AppCreateParamsWithCompose(t *testing.T) {
	// Test that compose config is properly serialized
	params := AppCreateParams{
		AppName:                   "myapp",
		CustomApp:                 true,
		CustomComposeConfigString: "services:\n  nginx:\n    image: nginx:alpine",
	}

	got := BuildCommand("app.create", params)

	if got == "" {
		t.Error("BuildCommand() returned empty string")
	}

	// Should contain the compose config in JSON
	if !strings.Contains(got, "custom_compose_config_string") {
		t.Errorf("BuildCommand() = %q, expected to contain 'custom_compose_config_string'", got)
	}
}

func TestBuildCommand_MarshalError(t *testing.T) {
	// Create an unmarshallable type (channel) to trigger json.Marshal error
	params := make(chan int)

	got := BuildCommand("test.method", params)
	want := "midclt call test.method"

	if got != want {
		t.Errorf("BuildCommand() = %q, want %q", got, want)
	}
}

func TestBuildCommand_PositionalArgs(t *testing.T) {
	// Test []any slices get passed as separate positional arguments
	// This is needed for CRUD update methods: app.update name {data}
	params := []any{"myapp", map[string]any{"custom_compose_config_string": "services:\n  web:\n    image: nginx"}}

	got := BuildCommand("app.update", params)
	want := `midclt call app.update '"myapp"' '{"custom_compose_config_string":"services:\n  web:\n    image: nginx"}'`

	if got != want {
		t.Errorf("BuildCommand() = %q, want %q", got, want)
	}
}

func TestBuildCommand_PositionalArgsWithUnmarshallable(t *testing.T) {
	// Test that unmarshallable elements in []any are skipped
	params := []any{"myapp", make(chan int), map[string]any{"key": "value"}}

	got := BuildCommand("app.update", params)
	// The channel is skipped, so we get: "myapp" {"key": "value"}
	want := `midclt call app.update '"myapp"' '{"key":"value"}'`

	if got != want {
		t.Errorf("BuildCommand() = %q, want %q", got, want)
	}
}

func TestBuildCommand_InvalidMethod(t *testing.T) {
	tests := []struct {
		name   string
		method string
		want   string
	}{
		{
			name:   "command injection attempt",
			method: "test; rm -rf /",
			want:   `midclt call 'test; rm -rf /'`,
		},
		{
			name:   "starts with uppercase",
			method: "Test.method",
			want:   `midclt call Test.method`,
		},
		{
			name:   "starts with number",
			method: "1test.method",
			want:   `midclt call 1test.method`,
		},
		{
			name:   "contains hyphen",
			method: "test-method",
			want:   `midclt call test-method`,
		},
		{
			name:   "empty string",
			method: "",
			want:   `midclt call ''`,
		},
		{
			name:   "single character",
			method: "a",
			want:   `midclt call a`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildCommand(tt.method, nil)
			if got != tt.want {
				t.Errorf("BuildCommand(%q, nil) = %q, want %q", tt.method, got, tt.want)
			}
		})
	}
}

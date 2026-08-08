package truenas

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/deevus/truenas-go/api"
)

// These tests validate the parameter maps the service layer builds against the
// TrueNAS JSON Schema embedded in the api package. They catch a class of bug
// that mock-based tests cannot: params that are well-formed Go but rejected by
// the middleware, such as sending "" for a field the API declares non-empty or
// formats as an email address.

// compileArgSchema compiles the schema for one positional argument of a method.
// The embedded schemas use draft-7 (notably array-form "items"), and formats
// are asserted rather than annotated so "format": "email" is enforced.
func compileArgSchema(t *testing.T, method string, argIdx int) *jsonschema.Schema {
	t.Helper()

	version := api.LatestVersion()
	raw, err := api.ArgSchema(version, method, argIdx)
	if err != nil {
		t.Fatalf("loading schema for %s arg %d: %v", method, argIdx, err)
	}

	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("parsing schema for %s: %v", method, err)
	}

	url := "truenas://" + version + "/" + method
	c := jsonschema.NewCompiler()
	c.DefaultDraft(jsonschema.Draft7)
	c.AssertFormat()
	if err := c.AddResource(url, doc); err != nil {
		t.Fatalf("adding schema resource for %s: %v", method, err)
	}
	schema, err := c.Compile(url)
	if err != nil {
		t.Fatalf("compiling schema for %s: %v", method, err)
	}
	return schema
}

// asJSONValue round-trips a params map through JSON so the validator sees the
// same types the wire does (int64 becomes a number, nil becomes null).
func asJSONValue(t *testing.T, params map[string]any) any {
	t.Helper()

	encoded, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("marshaling params: %v", err)
	}
	value, err := jsonschema.UnmarshalJSON(bytes.NewReader(encoded))
	if err != nil {
		t.Fatalf("decoding params: %v", err)
	}
	return value
}

func assertValid(t *testing.T, method string, argIdx int, params map[string]any) {
	t.Helper()

	if err := compileArgSchema(t, method, argIdx).Validate(asJSONValue(t, params)); err != nil {
		t.Errorf("params rejected by the %s schema:\n%v", method, err)
	}
}

func TestUserCreateParamsMatchSchema(t *testing.T) {
	tests := []struct {
		name string
		opts CreateUserOpts
	}{
		{
			name: "minimal",
			opts: CreateUserOpts{Username: "jdoe", FullName: "John Doe"},
		},
		{
			name: "existing primary group",
			opts: CreateUserOpts{Username: "jdoe", FullName: "John Doe", Password: "hunter2xyz", Group: 42},
		},
		{
			name: "all fields",
			opts: CreateUserOpts{
				Username: "jdoe", FullName: "John Doe", Email: "jdoe@example.com", UID: 3001,
				Password: "hunter2xyz", GroupCreate: true, Groups: []int64{100},
				Home: "/mnt/tank/home/jdoe", HomeCreate: true, HomeMode: "700",
				Shell: "/usr/bin/bash", SMB: true, SSHPasswordEnabled: true,
				SSHPubKey: "ssh-ed25519 AAAAC3Nz", SudoCommands: []string{"/usr/bin/apt"},
				SudoCommandsNopasswd: []string{"/usr/bin/systemctl"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertValid(t, "user.create", 0, userCreateOptsToParams(tt.opts))
		})
	}
}

func TestUserUpdateParamsMatchSchema(t *testing.T) {
	tests := []struct {
		name string
		opts UpdateUserOpts
	}{
		{
			name: "minimal",
			opts: UpdateUserOpts{Username: "jdoe", FullName: "John Doe"},
		},
		{
			name: "all fields",
			opts: UpdateUserOpts{
				Username: "jdoe", FullName: "John Doe", Email: StringPtr("jdoe@example.com"),
				Password: "hunter2xyz", Group: 42, Groups: []int64{100},
				Home: "/mnt/tank/home/jdoe", HomeMode: "700", Shell: "/usr/bin/bash",
				SMB: BoolPtr(true), SSHPasswordEnabled: BoolPtr(true), SSHPubKey: StringPtr("ssh-ed25519 AAAAC3Nz"),
				Locked: BoolPtr(true), SudoCommands: []string{}, SudoCommandsNopasswd: []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertValid(t, "user.update", 1, userUpdateOptsToParams(tt.opts))
		})
	}
}

func TestGroupCreateParamsMatchSchema(t *testing.T) {
	tests := []struct {
		name string
		opts CreateGroupOpts
	}{
		{
			name: "minimal",
			opts: CreateGroupOpts{Name: "devs"},
		},
		{
			name: "all fields",
			opts: CreateGroupOpts{
				Name: "devs", GID: 5000, SMB: true,
				SudoCommands: []string{"/usr/bin/apt"}, SudoCommandsNopasswd: []string{"/usr/bin/systemctl"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertValid(t, "group.create", 0, groupCreateOptsToParams(tt.opts))
		})
	}
}

func TestGroupUpdateParamsMatchSchema(t *testing.T) {
	assertValid(t, "group.update", 1, groupUpdateOptsToParams(UpdateGroupOpts{Name: "devs"}))
}

// TestEmptyStringParamsRejectedBySchema is a guard on the guard: it pins the
// reason optional string fields are omitted rather than sent empty. The API
// declares home and shell non-empty and email as an address or null, so the
// obvious "always send everything" shape is rejected.
func TestEmptyStringParamsRejectedBySchema(t *testing.T) {
	alwaysSendEverything := map[string]any{
		"username":             "jdoe",
		"full_name":            "John Doe",
		"email":                "",
		"password_disabled":    false,
		"home":                 "",
		"home_mode":            "",
		"shell":                "",
		"smb":                  false,
		"ssh_password_enabled": false,
		"locked":               false,
	}

	err := compileArgSchema(t, "user.create", 0).Validate(asJSONValue(t, alwaysSendEverything))
	if err == nil {
		t.Fatal("expected the schema to reject empty strings for email, home and shell")
	}

	for _, field := range []string{"email", "home", "shell"} {
		if !bytes.Contains([]byte(err.Error()), []byte(field)) {
			t.Errorf("expected %s to be reported as invalid, got:\n%v", field, err)
		}
	}
}

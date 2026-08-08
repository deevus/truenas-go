// Package api provides embedded TrueNAS API method definitions keyed by version.
package api

import (
	"embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

//go:embed */methods.json
var methodsFS embed.FS

// MethodDef describes a single TrueNAS API method.
type MethodDef struct {
	Description  *string `json:"description"`
	Job          bool    `json:"job"`
	Filterable   bool    `json:"filterable"`
	Downloadable bool    `json:"downloadable"`
	Uploadable   bool    `json:"uploadable"`
	ItemMethod   bool    `json:"item_method"`
	RequireWS    bool    `json:"require_websocket"`
	// Accepts holds the JSON Schema for each positional argument, in order.
	// e.g. user.create takes one argument, user.update takes (id, params).
	Accepts []json.RawMessage `json:"accepts"`
}

// ArgSchema returns the JSON Schema for the argIdx'th positional argument of a
// method, for the given TrueNAS version.
func ArgSchema(version, method string, argIdx int) (json.RawMessage, error) {
	methods, err := Methods(version)
	if err != nil {
		return nil, err
	}
	def, ok := methods[method]
	if !ok {
		return nil, fmt.Errorf("method %s not found in version %s", method, version)
	}
	if argIdx < 0 || argIdx >= len(def.Accepts) {
		return nil, fmt.Errorf("method %s takes %d args, want index %d", method, len(def.Accepts), argIdx)
	}
	return def.Accepts[argIdx], nil
}

// Methods returns all API methods for a given TrueNAS version (e.g. "25.04").
func Methods(version string) (map[string]MethodDef, error) {
	data, err := methodsFS.ReadFile(version + "/methods.json")
	if err != nil {
		return nil, fmt.Errorf("no methods for version %s: %w", version, err)
	}
	var methods map[string]MethodDef
	if err := json.Unmarshal(data, &methods); err != nil {
		return nil, fmt.Errorf("parsing methods for %s: %w", version, err)
	}
	return methods, nil
}

// LatestVersion returns the highest embedded version string.
func LatestVersion() string {
	vs := Versions()
	if len(vs) == 0 {
		return ""
	}
	return vs[len(vs)-1]
}

// Versions returns all embedded TrueNAS versions, sorted.
func Versions() []string {
	entries, err := methodsFS.ReadDir(".")
	if err != nil {
		return nil
	}
	var versions []string
	for _, e := range entries {
		if e.IsDir() {
			versions = append(versions, e.Name())
		}
	}
	sort.Strings(versions)
	return versions
}

// Namespace extracts the service namespace from an API method name.
// e.g. "app.registry.create" → "app.registry", "system.info" → "system"
func Namespace(method string) string {
	i := strings.LastIndex(method, ".")
	if i < 0 {
		return method
	}
	return method[:i]
}

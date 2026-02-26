package truenas

import "encoding/json"

// sampleAppJSON returns a JSON response for a single app (no config).
func sampleAppJSON() json.RawMessage {
	return json.RawMessage(`[{
		"name": "my-app",
		"state": "RUNNING",
		"custom_app": true
	}]`)
}

// sampleAppWithConfigJSON returns a JSON response for a single app with config.
func sampleAppWithConfigJSON() json.RawMessage {
	return json.RawMessage(`[{
		"name": "my-app",
		"state": "RUNNING",
		"custom_app": true,
		"config": {"version": "1.0", "port": 8080}
	}]`)
}

// sampleRegistryJSON returns a JSON response for a single registry.
func sampleRegistryJSON() json.RawMessage {
	return json.RawMessage(`[{
		"id": 1,
		"name": "my-registry",
		"description": "A test registry",
		"username": "admin",
		"password": "secret",
		"uri": "https://registry.example.com"
	}]`)
}

// sampleRegistryNullDescJSON returns a JSON response for a registry with null description.
func sampleRegistryNullDescJSON() json.RawMessage {
	return json.RawMessage(`[{
		"id": 2,
		"name": "null-desc-registry",
		"description": null,
		"username": "user",
		"password": "pass",
		"uri": "https://registry2.example.com"
	}]`)
}

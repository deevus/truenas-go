package truenas

import "encoding/json"

// sampleDatasetQueryJSON returns a JSON response for a FILESYSTEM dataset with all property fields.
func sampleDatasetQueryJSON() json.RawMessage {
	return json.RawMessage(`[{
		"id": "pool1/ds1",
		"name": "pool1/ds1",
		"pool": "pool1",
		"type": "FILESYSTEM",
		"mountpoint": "/mnt/pool1/ds1",
		"comments": {"value": "test dataset"},
		"compression": {"value": "lz4"},
		"quota": {"parsed": 1073741824, "value": "1G"},
		"refquota": {"parsed": 536870912, "value": "512M"},
		"atime": {"value": "on"},
		"volsize": {"parsed": 0, "value": ""},
		"volblocksize": {"value": ""},
		"sparse": {"value": ""}
	}]`)
}

// sampleZvolQueryJSON returns a JSON response for a VOLUME (zvol).
func sampleZvolQueryJSON() json.RawMessage {
	return json.RawMessage(`[{
		"id": "pool1/zvol1",
		"name": "pool1/zvol1",
		"pool": "pool1",
		"type": "VOLUME",
		"mountpoint": "",
		"comments": {"value": "test zvol"},
		"compression": {"value": "lz4"},
		"quota": {"parsed": 0, "value": ""},
		"refquota": {"parsed": 0, "value": ""},
		"atime": {"value": ""},
		"volsize": {"parsed": 10737418240, "value": "10G"},
		"volblocksize": {"value": "16K"},
		"sparse": {"value": "true"}
	}]`)
}

// samplePoolQueryJSON returns a JSON response for two pools.
func samplePoolQueryJSON() json.RawMessage {
	return json.RawMessage(`[
		{"id": 1, "name": "pool1", "path": "/mnt/pool1"},
		{"id": 2, "name": "pool2", "path": "/mnt/pool2"}
	]`)
}

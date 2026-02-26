package truenas

import (
	"encoding/json"
)

// sampleCredentialV25JSON returns a V25 format credential JSON response.
func sampleCredentialV25JSON() json.RawMessage {
	return json.RawMessage(`[{
		"id": 1,
		"name": "My S3 Cred",
		"provider": {
			"type": "S3",
			"access_key_id": "AKIATEST",
			"secret_access_key": "secret123",
			"endpoint": "s3.example.com",
			"region": "us-east-1"
		}
	}]`)
}

// sampleCredentialV24JSON returns a V24 format credential JSON response.
func sampleCredentialV24JSON() json.RawMessage {
	return json.RawMessage(`[{
		"id": 1,
		"name": "My S3 Cred",
		"provider": "S3",
		"attributes": {
			"access_key_id": "AKIATEST",
			"secret_access_key": "secret123",
			"endpoint": "s3.example.com",
			"region": "us-east-1"
		}
	}]`)
}

// sampleTaskJSON returns a cloud sync task JSON response.
func sampleTaskJSON() json.RawMessage {
	return json.RawMessage(`[{
		"id": 1,
		"description": "Backup to S3",
		"path": "/mnt/tank/data",
		"credentials": {"id": 5, "name": "My S3 Cred"},
		"attributes": {"bucket": "my-bucket", "folder": "/backups"},
		"schedule": {"minute": "0", "hour": "3", "dom": "*", "month": "*", "dow": "*"},
		"direction": "PUSH",
		"transfer_mode": "SYNC",
		"encryption": false,
		"snapshot": true,
		"transfers": 4,
		"bwlimit": [{"time": "08:00", "bandwidth": 1048576}],
		"exclude": ["*.tmp"],
		"include": [],
		"follow_symlinks": false,
		"create_empty_src_dirs": true,
		"enabled": true
	}]`)
}

// sampleTaskFalseAttrsJSON returns a task JSON response where attributes is false.
func sampleTaskFalseAttrsJSON() json.RawMessage {
	return json.RawMessage(`[{
		"id": 2,
		"description": "Task with false attrs",
		"path": "/mnt/pool/data",
		"credentials": {"id": 3, "name": "Cred"},
		"attributes": false,
		"schedule": {"minute": "30", "hour": "*/2", "dom": "*", "month": "*", "dow": "*"},
		"direction": "PULL",
		"transfer_mode": "COPY",
		"encryption": true,
		"encryption_password": "mypass",
		"encryption_salt": "mysalt",
		"snapshot": false,
		"transfers": 2,
		"bwlimit": [],
		"exclude": [],
		"include": ["*.dat"],
		"follow_symlinks": true,
		"create_empty_src_dirs": false,
		"enabled": false
	}]`)
}

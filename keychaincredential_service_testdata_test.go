package truenas

import "encoding/json"

// sampleSSHKeyPairAttributesJSON returns an ssh keypair attributes structure as json
func sampleSSHKeyPairAttributesJSON() json.RawMessage {
	return json.RawMessage(`{
		"public_key": "public key",
		"private_key": "private key"
	}`)
}

// sampleSSHKeyPairJSON returns a single ssh keypair JSON response in an array.
func sampleSSHKeyPairJSON() json.RawMessage {
	return json.RawMessage(`[{
		"id": 1,
		"name": "ssh keypair name",
		"attributes" : {
			"public_key": "public key",
			"private_key": "private key"
        }
	}]`)
}

// sampleSSHKeyPairJSON returns two ssh keypair JSON responses in an array.
func sampleSSHKeyPairsJSON() json.RawMessage {
	return json.RawMessage(`[
		{
			"id": 1,
			"name": "ssh keypair name 1",
			"type" : "SSH_KEY_PAIR",
			"attributes" : {
				"public_key": "public key 1",
				"private_key": "private key 1"
			}
		},
		{
			"id": 2,
			"name": "ssh keypair name 2",
			"type" : "SSH_KEY_PAIR",
			"attributes" : {
				"public_key": "public key 2",
				"private_key": "private key 2"
			}
		}
	]`)
}

// sampleSSHKeyPair returns a single SSHKeyPair.
func sampleSSHKeyPair() SSHKeyPair {
	return SSHKeyPair{
		ID:         1,
		Name:       "ssh keypair name",
		PublicKey:  "public key",
		PrivateKey: "private key",
	}
}

// sampleSSHKeyPair returns an array with two single SSHKeyPairs.
func sampleSSHKeyPairs() []SSHKeyPair {
	return []SSHKeyPair{
		{
			ID:         1,
			Name:       "ssh keypair name 1",
			PublicKey:  "public key 1",
			PrivateKey: "private key 1",
		},
		{
			ID:         2,
			Name:       "ssh keypair name 2",
			PublicKey:  "public key 2",
			PrivateKey: "private key 2",
		},
	}
}

// sampleSSHConnectionAttributesJSON returns an ssh credential attributes structure as json
func sampleSSHConnectionAttributesJSON() json.RawMessage {
	return json.RawMessage(`{
		"host": "some host", 
		"port": 123, 
		"username": "some username", 
		"remote_host_key": "some remote host key", 
		"connect_timeout": 42, 
		"private_key": 1
	}`)
}

// sampleSSHConnectionJSON returns a single ssh credential JSON response in an array.
func sampleSSHConnectionJSON() json.RawMessage {
	return json.RawMessage(`[{
		"id": 1,
		"name": "ssh credential name",
		"attributes" : {
			"host": "some host",
			"port": 123,
			"username": "some username",
			"remote_host_key": "some remote host key",
			"connect_timeout": 42,
			"private_key": 1
        }
	}]`)
}

// sampleSSHConnectionJSON returns two ssh credential JSON responses in an array.
func sampleSSHConnectionsJSON() json.RawMessage {
	return json.RawMessage(`[
		{
			"id": 1,
			"name": "ssh credential name 1",
			"attributes" : {
				"host": "some host 1",
				"port": 123,
				"username": "some username 1",
				"remote_host_key": "some remote host key 1",
				"connect_timeout": 42,
				"private_key": 1
			}
		},
		{
			"id": 2,
			"name": "ssh credential name 2",
			"attributes" : {
				"host": "some host 2",
				"port": 456,
				"username": "some username 2",
				"remote_host_key": "some remote host key 2",
				"connect_timeout": 37,
				"private_key": 2
			}
		}
	]`)
}

// sampleSSHConnection returns a single SSHConnection.
func sampleSSHConnection() SSHConnection {
	return SSHConnection{
		ID:             1,
		Name:           "ssh credential name",
		Host:           "some host",
		Port:           123,
		Username:       "some username",
		PrivateKeyID:   1,
		RemoteHostKey:  "some remote host key",
		ConnectTimeout: 42,
	}
}

// sampleSSHConnections returns an array with two SSHConnections.
func sampleSSHConnections() []SSHConnection {
	return []SSHConnection{
		{
			ID:             1,
			Name:           "ssh credential name 1",
			Host:           "some host 1",
			Port:           123,
			Username:       "some username 1",
			PrivateKeyID:   1,
			RemoteHostKey:  "some remote host key 1",
			ConnectTimeout: 42,
		},
		{
			ID:             2,
			Name:           "ssh credential name 2",
			Host:           "some host 2",
			Port:           456,
			Username:       "some username 2",
			PrivateKeyID:   2,
			RemoteHostKey:  "some remote host key 2",
			ConnectTimeout: 37,
		},
	}
}

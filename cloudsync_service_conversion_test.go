package truenas

import (
	"encoding/json"
	"testing"
)

func TestCredentialFromResponse_S3(t *testing.T) {
	resp := CloudSyncCredentialResponse{
		ID:   1,
		Name: "S3 Cred",
		Provider: CloudSyncCredentialProvider{
			Type:            "S3",
			AccessKeyID:     "AKIATEST",
			SecretAccessKey: "secret",
			Endpoint:        "s3.example.com",
			Region:          "us-east-1",
		},
	}

	cred := credentialFromResponse(resp)
	if cred.ID != 1 {
		t.Errorf("expected ID 1, got %d", cred.ID)
	}
	if cred.ProviderType != "S3" {
		t.Errorf("expected provider type S3, got %s", cred.ProviderType)
	}
	if cred.Attributes["access_key_id"] != "AKIATEST" {
		t.Errorf("expected access_key_id AKIATEST, got %s", cred.Attributes["access_key_id"])
	}
	if cred.Attributes["secret_access_key"] != "secret" {
		t.Errorf("expected secret_access_key secret, got %s", cred.Attributes["secret_access_key"])
	}
	if cred.Attributes["endpoint"] != "s3.example.com" {
		t.Errorf("expected endpoint s3.example.com, got %s", cred.Attributes["endpoint"])
	}
	if cred.Attributes["region"] != "us-east-1" {
		t.Errorf("expected region us-east-1, got %s", cred.Attributes["region"])
	}
	// Should not have B2/GCS keys
	if _, ok := cred.Attributes["account"]; ok {
		t.Error("unexpected account attribute for S3")
	}
}

func TestCredentialFromResponse_B2(t *testing.T) {
	resp := CloudSyncCredentialResponse{
		ID:   2,
		Name: "B2 Cred",
		Provider: CloudSyncCredentialProvider{
			Type:    "B2",
			Account: "b2account",
			Key:     "b2key",
		},
	}

	cred := credentialFromResponse(resp)
	if cred.ProviderType != "B2" {
		t.Errorf("expected provider type B2, got %s", cred.ProviderType)
	}
	if cred.Attributes["account"] != "b2account" {
		t.Errorf("expected account b2account, got %s", cred.Attributes["account"])
	}
	if cred.Attributes["key"] != "b2key" {
		t.Errorf("expected key b2key, got %s", cred.Attributes["key"])
	}
	// Should not have S3 keys
	if _, ok := cred.Attributes["access_key_id"]; ok {
		t.Error("unexpected access_key_id attribute for B2")
	}
}

func TestCredentialFromResponse_GCS(t *testing.T) {
	resp := CloudSyncCredentialResponse{
		ID:   3,
		Name: "GCS Cred",
		Provider: CloudSyncCredentialProvider{
			Type:                      "GOOGLE_CLOUD_STORAGE",
			ServiceAccountCredentials: `{"type": "service_account"}`,
		},
	}

	cred := credentialFromResponse(resp)
	if cred.ProviderType != "GOOGLE_CLOUD_STORAGE" {
		t.Errorf("expected provider type GOOGLE_CLOUD_STORAGE, got %s", cred.ProviderType)
	}
	if cred.Attributes["service_account_credentials"] != `{"type": "service_account"}` {
		t.Errorf("unexpected service_account_credentials: %s", cred.Attributes["service_account_credentials"])
	}
}

func TestCredentialFromResponse_EmptyProvider(t *testing.T) {
	resp := CloudSyncCredentialResponse{
		ID:   4,
		Name: "Empty",
		Provider: CloudSyncCredentialProvider{
			Type: "UNKNOWN",
		},
	}

	cred := credentialFromResponse(resp)
	if len(cred.Attributes) != 0 {
		t.Errorf("expected empty attributes, got %v", cred.Attributes)
	}
}

func TestCredentialOptsToAttrsAny(t *testing.T) {
	attrs := map[string]string{
		"access_key_id":     "AKIATEST",
		"secret_access_key": "secret",
	}
	result := credentialOptsToAttrsAny(attrs)
	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}
	if result["access_key_id"] != "AKIATEST" {
		t.Errorf("expected access_key_id AKIATEST, got %v", result["access_key_id"])
	}
}

func TestCredentialOptsToAttrsAny_Nil(t *testing.T) {
	result := credentialOptsToAttrsAny(nil)
	if result != nil {
		t.Errorf("expected nil result for nil input, got %v", result)
	}
}

func TestTaskOptsToParams(t *testing.T) {
	bw := int64(1048576)
	opts := CreateCloudSyncTaskOpts{
		Description:        "Test Task",
		Path:               "/mnt/tank/data",
		CredentialID:       5,
		Direction:          "PUSH",
		TransferMode:       "SYNC",
		Snapshot:           true,
		Transfers:          4,
		BWLimit:            []BwLimit{{Time: "08:00", Bandwidth: &bw}},
		FollowSymlinks:     false,
		CreateEmptySrcDirs: true,
		Enabled:            true,
		Encryption:         true,
		EncryptionPassword: "pass",
		EncryptionSalt:     "salt",
		Schedule: Schedule{
			Minute: "0",
			Hour:   "3",
			Dom:    "*",
			Month:  "*",
			Dow:    "*",
		},
		Attributes: map[string]any{"bucket": "my-bucket"},
		Exclude:    []string{"*.tmp"},
		Include:    []string{"*.dat"},
	}

	params := taskOptsToParams(opts)

	if params["description"] != "Test Task" {
		t.Errorf("expected description 'Test Task', got %v", params["description"])
	}
	if params["path"] != "/mnt/tank/data" {
		t.Errorf("expected path '/mnt/tank/data', got %v", params["path"])
	}
	if params["credentials"] != int64(5) {
		t.Errorf("expected credentials 5, got %v", params["credentials"])
	}
	if params["direction"] != "PUSH" {
		t.Errorf("expected direction PUSH, got %v", params["direction"])
	}
	if params["encryption"] != true {
		t.Error("expected encryption=true")
	}
	if params["encryption_password"] != "pass" {
		t.Errorf("expected encryption_password 'pass', got %v", params["encryption_password"])
	}
	if params["encryption_salt"] != "salt" {
		t.Errorf("expected encryption_salt 'salt', got %v", params["encryption_salt"])
	}

	sched := params["schedule"].(map[string]any)
	if sched["hour"] != "3" {
		t.Errorf("expected schedule hour 3, got %v", sched["hour"])
	}

	if _, ok := params["bwlimit"]; !ok {
		t.Error("expected bwlimit in params")
	}
	if _, ok := params["exclude"]; !ok {
		t.Error("expected exclude in params")
	}
	if _, ok := params["include"]; !ok {
		t.Error("expected include in params")
	}
	if _, ok := params["attributes"]; !ok {
		t.Error("expected attributes in params")
	}
}

func TestTaskOptsToParams_NoOptionalFields(t *testing.T) {
	opts := CreateCloudSyncTaskOpts{
		Description:  "Minimal Task",
		Path:         "/mnt/tank",
		CredentialID: 1,
		Direction:    "PUSH",
		TransferMode: "SYNC",
		Schedule: Schedule{
			Minute: "*",
			Hour:   "*",
			Dom:    "*",
			Month:  "*",
			Dow:    "*",
		},
	}

	params := taskOptsToParams(opts)

	// Encryption is false, so no encryption_password/encryption_salt
	if _, ok := params["encryption_password"]; ok {
		t.Error("unexpected encryption_password when encryption=false")
	}
	if _, ok := params["encryption_salt"]; ok {
		t.Error("unexpected encryption_salt when encryption=false")
	}
	// No bwlimit, exclude, include
	if _, ok := params["bwlimit"]; ok {
		t.Error("unexpected bwlimit when empty")
	}
	if _, ok := params["exclude"]; ok {
		t.Error("unexpected exclude when empty")
	}
	if _, ok := params["include"]; ok {
		t.Error("unexpected include when empty")
	}
	// Attributes should still be present with fast_list even when Attributes is nil
	attrs, ok := params["attributes"].(map[string]any)
	if !ok {
		t.Fatal("expected attributes map in params")
	}
	if attrs["fast_list"] != false {
		t.Errorf("expected fast_list=false in attributes, got %v", attrs["fast_list"])
	}
}

func TestTaskOptsToParams_FastList(t *testing.T) {
	opts := CreateCloudSyncTaskOpts{
		Description:  "FastList Task",
		Path:         "/mnt/tank",
		CredentialID: 1,
		Direction:    "PUSH",
		TransferMode: "SYNC",
		FastList:     true,
		Schedule: Schedule{
			Minute: "*",
			Hour:   "*",
			Dom:    "*",
			Month:  "*",
			Dow:    "*",
		},
	}

	params := taskOptsToParams(opts)

	attrs, ok := params["attributes"].(map[string]any)
	if !ok {
		t.Fatal("expected attributes map in params")
	}
	if attrs["fast_list"] != true {
		t.Errorf("expected fast_list=true in attributes, got %v", attrs["fast_list"])
	}
}

func TestTaskOptsToParams_FastList_MergesWithExistingAttributes(t *testing.T) {
	opts := CreateCloudSyncTaskOpts{
		Description:  "Merge Task",
		Path:         "/mnt/tank",
		CredentialID: 1,
		Direction:    "PUSH",
		TransferMode: "SYNC",
		FastList:     true,
		Attributes:   map[string]any{"bucket": "my-bucket"},
		Schedule: Schedule{
			Minute: "*",
			Hour:   "*",
			Dom:    "*",
			Month:  "*",
			Dow:    "*",
		},
	}

	params := taskOptsToParams(opts)

	attrs, ok := params["attributes"].(map[string]any)
	if !ok {
		t.Fatal("expected attributes map in params")
	}
	if attrs["bucket"] != "my-bucket" {
		t.Errorf("expected bucket=my-bucket, got %v", attrs["bucket"])
	}
	if attrs["fast_list"] != true {
		t.Errorf("expected fast_list=true in attributes, got %v", attrs["fast_list"])
	}
}

func TestTaskFromResponse_NilBWLimit(t *testing.T) {
	resp := CloudSyncTaskResponse{
		ID:           1,
		Description:  "Test",
		Path:         "/mnt/tank",
		Credentials:  CloudSyncTaskCredentialRef{ID: 1, Name: "Cred"},
		Attributes:   json.RawMessage(`{"bucket": "b"}`),
		Schedule:     ScheduleResponse{Minute: "*", Hour: "*", Dom: "*", Month: "*", Dow: "*"},
		Direction:    "PUSH",
		TransferMode: "SYNC",
		Enabled:      true,
	}

	task := taskFromResponse(resp)
	if task.BWLimit != nil {
		t.Errorf("expected nil bwlimit, got %v", task.BWLimit)
	}
}

func TestTaskFromResponse_EmptyAttributes(t *testing.T) {
	resp := CloudSyncTaskResponse{
		ID:          1,
		Description: "Test",
		Credentials: CloudSyncTaskCredentialRef{ID: 1},
		Attributes:  json.RawMessage(`{}`),
		Schedule:    ScheduleResponse{Minute: "*", Hour: "*", Dom: "*", Month: "*", Dow: "*"},
	}

	task := taskFromResponse(resp)
	if task.Attributes == nil {
		t.Error("expected non-nil attributes for empty object")
	}
	if len(task.Attributes) != 0 {
		t.Errorf("expected 0 attributes, got %d", len(task.Attributes))
	}
}

func TestTaskFromResponse_NullAttributes(t *testing.T) {
	resp := CloudSyncTaskResponse{
		ID:          1,
		Description: "Test",
		Credentials: CloudSyncTaskCredentialRef{ID: 1},
		Attributes:  nil,
		Schedule:    ScheduleResponse{Minute: "*", Hour: "*", Dom: "*", Month: "*", Dow: "*"},
	}

	task := taskFromResponse(resp)
	if task.Attributes != nil {
		t.Errorf("expected nil attributes, got %v", task.Attributes)
	}
}

func TestTaskFromResponse_FastList(t *testing.T) {
	resp := CloudSyncTaskResponse{
		ID:          1,
		Description: "Test",
		Credentials: CloudSyncTaskCredentialRef{ID: 1},
		Attributes:  json.RawMessage(`{"bucket": "my-bucket", "fast_list": true}`),
		Schedule:    ScheduleResponse{Minute: "*", Hour: "*", Dom: "*", Month: "*", Dow: "*"},
	}

	task := taskFromResponse(resp)
	if !task.FastList {
		t.Error("expected FastList=true")
	}
	// fast_list should be removed from the Attributes map
	if _, ok := task.Attributes["fast_list"]; ok {
		t.Error("fast_list should be removed from Attributes map")
	}
	// Other attributes should be preserved
	if task.Attributes["bucket"] != "my-bucket" {
		t.Errorf("expected bucket=my-bucket, got %v", task.Attributes["bucket"])
	}
}

func TestTaskFromResponse_FastListFalse(t *testing.T) {
	resp := CloudSyncTaskResponse{
		ID:          1,
		Description: "Test",
		Credentials: CloudSyncTaskCredentialRef{ID: 1},
		Attributes:  json.RawMessage(`{"bucket": "b"}`),
		Schedule:    ScheduleResponse{Minute: "*", Hour: "*", Dom: "*", Month: "*", Dow: "*"},
	}

	task := taskFromResponse(resp)
	if task.FastList {
		t.Error("expected FastList=false when not in attributes")
	}
}

func TestNewCloudSyncService(t *testing.T) {
	mock := &mockAsyncCaller{}
	v := Version{Major: 25, Minor: 4}
	svc := NewCloudSyncService(mock, v)
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.client != mock {
		t.Error("expected client to be set")
	}
	if svc.version != v {
		t.Error("expected version to be set")
	}
}

package client

import (
	"encoding/json"
	"testing"
)

func TestJSONRPCRequest_Marshal(t *testing.T) {
	tests := []struct {
		name     string
		request  JSONRPCRequest
		expected string
	}{
		{
			name: "request_with_params",
			request: JSONRPCRequest{
				JSONRPC: "2.0",
				Method:  "pool.dataset.query",
				Params:  []any{[][]any{{"id", "=", "tank"}}},
				ID:      "1",
			},
			expected: `{"jsonrpc":"2.0","method":"pool.dataset.query","params":[[["id","=","tank"]]],"id":"1"}`,
		},
		{
			name: "request_without_params",
			request: JSONRPCRequest{
				JSONRPC: "2.0",
				Method:  "system.info",
				ID:      "2",
			},
			expected: `{"jsonrpc":"2.0","method":"system.info","id":"2"}`,
		},
		{
			name: "request_with_map_params",
			request: JSONRPCRequest{
				JSONRPC: "2.0",
				Method:  "pool.dataset.create",
				Params:  map[string]any{"name": "tank/test"},
				ID:      "3",
			},
			expected: `{"jsonrpc":"2.0","method":"pool.dataset.create","params":{"name":"tank/test"},"id":"3"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.request)
			if err != nil {
				t.Fatalf("failed to marshal request: %v", err)
			}

			// Parse both to compare as objects (handles key ordering)
			var got, want any
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("failed to unmarshal result: %v", err)
			}
			if err := json.Unmarshal([]byte(tt.expected), &want); err != nil {
				t.Fatalf("failed to unmarshal expected: %v", err)
			}

			gotJSON, _ := json.Marshal(got)
			wantJSON, _ := json.Marshal(want)
			if string(gotJSON) != string(wantJSON) {
				t.Errorf("marshal mismatch:\ngot:  %s\nwant: %s", gotJSON, wantJSON)
			}
		})
	}
}

func TestJSONRPCResponse_Unmarshal_Success(t *testing.T) {
	raw := `{"jsonrpc":"2.0","result":{"hostname":"truenas","version":"TrueNAS-SCALE-24.04"},"id":"1"}`

	var resp JSONRPCResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.JSONRPC != "2.0" {
		t.Errorf("expected jsonrpc 2.0, got %s", resp.JSONRPC)
	}
	if resp.ID != "1" {
		t.Errorf("expected id 1, got %s", resp.ID)
	}
	if resp.Error != nil {
		t.Errorf("expected no error, got %v", resp.Error)
	}
	if resp.Result == nil {
		t.Fatal("expected result, got nil")
	}

	// Verify we can decode the result
	var result map[string]string
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	if result["hostname"] != "truenas" {
		t.Errorf("expected hostname truenas, got %s", result["hostname"])
	}
}

func TestJSONRPCResponse_Unmarshal_Error(t *testing.T) {
	raw := `{
		"jsonrpc": "2.0",
		"error": {
			"code": -32001,
			"message": "TrueNAS call error",
			"data": {
				"reason": "[EINVAL] storage.config.host_path_config.path: Field was not expected",
				"error": 22,
				"extra": ["storage.config", "host_path_config"]
			}
		},
		"id": "5"
	}`

	var resp JSONRPCResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.JSONRPC != "2.0" {
		t.Errorf("expected jsonrpc 2.0, got %s", resp.JSONRPC)
	}
	if resp.ID != "5" {
		t.Errorf("expected id 5, got %s", resp.ID)
	}
	if resp.Result != nil {
		t.Errorf("expected no result, got %s", string(resp.Result))
	}
	if resp.Error == nil {
		t.Fatal("expected error, got nil")
	}
	if resp.Error.Code != ErrCodeTrueNASCall {
		t.Errorf("expected code %d, got %d", ErrCodeTrueNASCall, resp.Error.Code)
	}
	if resp.Error.Message != "TrueNAS call error" {
		t.Errorf("expected message 'TrueNAS call error', got %s", resp.Error.Message)
	}
	if resp.Error.Data == nil {
		t.Fatal("expected error data, got nil")
	}
	if resp.Error.Data.Reason != "[EINVAL] storage.config.host_path_config.path: Field was not expected" {
		t.Errorf("unexpected reason: %s", resp.Error.Data.Reason)
	}
	if resp.Error.Data.Error != 22 {
		t.Errorf("expected errno 22, got %d", resp.Error.Data.Error)
	}
	if len(resp.Error.Data.Extra) != 2 {
		t.Errorf("expected 2 extra items, got %d", len(resp.Error.Data.Extra))
	}
}

func TestJSONRPCResponse_Unmarshal_ErrorWithoutData(t *testing.T) {
	raw := `{
		"jsonrpc": "2.0",
		"error": {
			"code": -32000,
			"message": "Too many concurrent calls"
		},
		"id": "10"
	}`

	var resp JSONRPCResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Error == nil {
		t.Fatal("expected error, got nil")
	}
	if resp.Error.Code != ErrCodeTooManyConcurrent {
		t.Errorf("expected code %d, got %d", ErrCodeTooManyConcurrent, resp.Error.Code)
	}
	if resp.Error.Data != nil {
		t.Errorf("expected no data, got %v", resp.Error.Data)
	}
}

func TestJSONRPCError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *JSONRPCError
		expected string
	}{
		{
			name: "with_data_reason",
			err: &JSONRPCError{
				Code:    -32001,
				Message: "TrueNAS call error",
				Data: &JSONRPCData{
					Reason: "[EINVAL] Field was not expected",
					Error:  22,
				},
			},
			expected: "[EINVAL] Field was not expected",
		},
		{
			name: "without_data",
			err: &JSONRPCError{
				Code:    -32000,
				Message: "Too many concurrent calls",
			},
			expected: "Too many concurrent calls",
		},
		{
			name: "with_data_empty_reason",
			err: &JSONRPCError{
				Code:    -32001,
				Message: "Some error",
				Data: &JSONRPCData{
					Reason: "",
					Error:  1,
				},
			},
			expected: "Some error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.expected {
				t.Errorf("Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestJSONRPCErrorCodes(t *testing.T) {
	// Verify error code constants are correct
	if ErrCodeTooManyConcurrent != -32000 {
		t.Errorf("ErrCodeTooManyConcurrent = %d, want -32000", ErrCodeTooManyConcurrent)
	}
	if ErrCodeTrueNASCall != -32001 {
		t.Errorf("ErrCodeTrueNASCall = %d, want -32001", ErrCodeTrueNASCall)
	}
}

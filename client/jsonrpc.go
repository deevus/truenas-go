package client

import "encoding/json"

// JSONRPCRequest represents a JSON-RPC 2.0 request.
type JSONRPCRequest struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
	ID      string `json:"id"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response.
type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
	ID      string          `json:"id"`
}

// JSONRPCError represents a JSON-RPC 2.0 error.
type JSONRPCError struct {
	Code    int          `json:"code"`
	Message string       `json:"message"`
	Data    *JSONRPCData `json:"data,omitempty"`
}

// JSONRPCData contains additional error details from TrueNAS.
type JSONRPCData struct {
	Reason string `json:"reason"`
	Error  int    `json:"error"` // errno value
	Extra  []any  `json:"extra,omitempty"`
}

// Error implements the error interface.
func (e *JSONRPCError) Error() string {
	if e.Data != nil && e.Data.Reason != "" {
		return e.Data.Reason
	}
	return e.Message
}

// JSON-RPC error codes from TrueNAS.
const (
	ErrCodeTooManyConcurrent = -32000 // TOO_MANY_CONCURRENT_CALLS
	ErrCodeTrueNASCall       = -32001 // TRUENAS_CALL_ERROR
)

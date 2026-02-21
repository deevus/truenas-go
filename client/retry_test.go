package client

import (
	"errors"
	"io"
	"testing"
	"time"
)

func TestSSHRetryClassifier_IsRetriable(t *testing.T) {
	classifier := &SSHRetryClassifier{}

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "connection handshake failure",
			err:      errors.New("Failed connection handshake"),
			expected: true,
		},
		{
			name:     "unexpected closure",
			err:      errors.New("Unexpected closure of remote connection"),
			expected: true,
		},
		{
			name:     "connection refused",
			err:      errors.New("dial tcp: connection refused"),
			expected: true,
		},
		{
			name:     "connection reset",
			err:      errors.New("read: connection reset by peer"),
			expected: true,
		},
		{
			name:     "io timeout",
			err:      errors.New("i/o timeout"),
			expected: true,
		},
		{
			name:     "validation error not retriable",
			err:      errors.New("[EINVAL] name: invalid character"),
			expected: false,
		},
		{
			name:     "not found error not retriable",
			err:      errors.New("[ENOENT] resource not found"),
			expected: false,
		},
		{
			name:     "generic error not retriable",
			err:      errors.New("something went wrong"),
			expected: false,
		},
		{
			name:     "case insensitive matching",
			err:      errors.New("CONNECTION REFUSED"),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifier.IsRetriable(tt.err)
			if result != tt.expected {
				t.Errorf("IsRetriable(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestCalculateBackoff(t *testing.T) {
	tests := []struct {
		attempt int
		minWant time.Duration
		maxWant time.Duration
	}{
		{attempt: 0, minWant: 1500 * time.Millisecond, maxWant: 2500 * time.Millisecond},
		{attempt: 1, minWant: 3 * time.Second, maxWant: 5 * time.Second},
		{attempt: 2, minWant: 6 * time.Second, maxWant: 10 * time.Second},
		{attempt: 3, minWant: 12 * time.Second, maxWant: 20 * time.Second},
		{attempt: 10, minWant: 22500 * time.Millisecond, maxWant: 37500 * time.Millisecond}, // capped at 30s base
	}

	for _, tt := range tests {
		t.Run("attempt_"+string(rune('0'+tt.attempt)), func(t *testing.T) {
			// Run multiple times to account for jitter
			for i := 0; i < 10; i++ {
				result := CalculateBackoff(tt.attempt)
				if result < tt.minWant || result > tt.maxWant {
					t.Errorf("CalculateBackoff(%d) = %v, want between %v and %v",
						tt.attempt, result, tt.minWant, tt.maxWant)
				}
			}
		})
	}
}

func TestCalculateBackoff_Deterministic(t *testing.T) {
	// With same seed, should get consistent behavior within jitter range
	result1 := CalculateBackoff(0)
	result2 := CalculateBackoff(0)

	// Both should be in valid range even if different due to jitter
	minWant := 1500 * time.Millisecond
	maxWant := 2500 * time.Millisecond

	if result1 < minWant || result1 > maxWant {
		t.Errorf("first call: got %v, want between %v and %v", result1, minWant, maxWant)
	}
	if result2 < minWant || result2 > maxWant {
		t.Errorf("second call: got %v, want between %v and %v", result2, minWant, maxWant)
	}
}

func TestWebSocketRetryClassifier_IsRetriable(t *testing.T) {
	classifier := &WebSocketRetryClassifier{}

	tests := []struct {
		name      string
		err       error
		retriable bool
	}{
		{
			name:      "nil error",
			err:       nil,
			retriable: false,
		},
		{
			name:      "internal connection error",
			err:       &JSONRPCError{Code: ErrCodeInternal, Message: "connection closed"},
			retriable: true,
		},
		{
			name:      "too many concurrent calls",
			err:       &JSONRPCError{Code: ErrCodeTooManyConcurrent, Message: "Too many concurrent calls"},
			retriable: true,
		},
		{
			name:      "truenas call error - not retriable",
			err:       &JSONRPCError{Code: ErrCodeTrueNASCall, Message: "Validation error"},
			retriable: false,
		},
		{
			name:      "truenas call error - EAGAIN retriable",
			err:       &JSONRPCError{Code: ErrCodeTrueNASCall, Message: "Resource busy", Data: &JSONRPCData{Error: 11}},
			retriable: true,
		},
		{
			name:      "truenas call error - EBUSY retriable",
			err:       &JSONRPCError{Code: ErrCodeTrueNASCall, Message: "Device busy", Data: &JSONRPCData{Error: 16}},
			retriable: true,
		},
		{
			name:      "truenas call error - EINVAL not retriable",
			err:       &JSONRPCError{Code: ErrCodeTrueNASCall, Message: "Invalid argument", Data: &JSONRPCData{Error: 22}},
			retriable: false,
		},
		{
			name:      "truenas call error - ENOTAUTHENTICATED retriable",
			err:       &JSONRPCError{Code: ErrCodeTrueNASCall, Message: "Not authenticated", Data: &JSONRPCData{Reason: "[ENOTAUTHENTICATED] Not authenticated"}},
			retriable: true,
		},
		{
			name:      "connection closed error",
			err:       errors.New("websocket: close sent"),
			retriable: true,
		},
		{
			name:      "EOF error",
			err:       io.EOF,
			retriable: true,
		},
		{
			name:      "random error",
			err:       errors.New("something went wrong"),
			retriable: false,
		},
		// Network errors - transient connection failures
		{
			name:      "connection reset by peer",
			err:       errors.New("read tcp 100.66.219.70:63734->192.168.1.127:8443: read: connection reset by peer"),
			retriable: true,
		},
		{
			name:      "broken pipe",
			err:       errors.New("write tcp: broken pipe"),
			retriable: true,
		},
		{
			name:      "connection refused",
			err:       errors.New("dial tcp: connection refused"),
			retriable: true,
		},
		{
			name:      "no route to host",
			err:       errors.New("dial tcp: no route to host"),
			retriable: true,
		},
		{
			name:      "network is unreachable",
			err:       errors.New("dial tcp: network is unreachable"),
			retriable: true,
		},
		{
			name:      "i/o timeout",
			err:       errors.New("read tcp: i/o timeout"),
			retriable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifier.IsRetriable(tt.err)
			if got != tt.retriable {
				t.Errorf("IsRetriable() = %v, want %v", got, tt.retriable)
			}
		})
	}
}

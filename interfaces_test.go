package truenas

import (
	"encoding/json"
	"testing"
)

func TestSubscription_Close(t *testing.T) {
	ch := make(chan json.RawMessage, 1)
	closed := false
	sub := &Subscription[json.RawMessage]{
		C:      ch,
		cancel: func() { closed = true },
	}
	sub.Close()
	if !closed {
		t.Error("expected cancel to be called")
	}
}

func TestSubscribeCaller_Interface(t *testing.T) {
	// Compile-time check that mockSubscribeCaller satisfies SubscribeCaller
	var _ SubscribeCaller = (*mockSubscribeCaller)(nil)
}

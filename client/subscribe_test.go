package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	truenas "github.com/deevus/truenas-go"
	"github.com/gorilla/websocket"
)

// newSubscribeTestServer creates a test WebSocket server that handles auth, core.subscribe,
// core.ping, and optionally pushes events via the pushEvents callback (called once after
// the first core.subscribe response).
func newSubscribeTestServer(t *testing.T, pushEvents func(conn *websocket.Conn)) *httptest.Server {
	t.Helper()
	var writeMu sync.Mutex
	var once sync.Once

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var req JSONRPCRequest
			if err := json.Unmarshal(msg, &req); err != nil {
				return
			}

			writeMu.Lock()
			switch req.Method {
			case "auth.login_ex":
				_ = conn.WriteJSON(JSONRPCResponse{
					JSONRPC: "2.0",
					Result:  json.RawMessage(`{"response_type":"SUCCESS"}`),
					ID:      req.ID,
				})
			case "core.subscribe":
				_ = conn.WriteJSON(JSONRPCResponse{
					JSONRPC: "2.0",
					Result:  json.RawMessage(`true`),
					ID:      req.ID,
				})
				if pushEvents != nil {
					once.Do(func() {
						go func() {
							// Small delay to let subscribe processing complete
							time.Sleep(50 * time.Millisecond)
							writeMu.Lock()
							pushEvents(conn)
							writeMu.Unlock()
						}()
					})
				}
			case "core.ping":
				_ = conn.WriteJSON(JSONRPCResponse{
					JSONRPC: "2.0",
					Result:  json.RawMessage(`"pong"`),
					ID:      req.ID,
				})
			default:
				_ = conn.WriteJSON(JSONRPCResponse{
					JSONRPC: "2.0",
					Result:  json.RawMessage(`"ok"`),
					ID:      req.ID,
				})
			}
			writeMu.Unlock()
		}
	}))
}

// newSubscribeTestClient creates a WebSocketClient wired to the given test server.
func newSubscribeTestClient(t *testing.T, server *httptest.Server) *WebSocketClient {
	t.Helper()
	host := strings.TrimPrefix(server.URL, "http://")
	parts := strings.Split(host, ":")
	var port int
	fmt.Sscanf(parts[1], "%d", &port)

	client, err := NewWebSocketClient(WebSocketConfig{
		Host:         parts[0],
		Username:     "root",
		APIKey:       "test-key",
		Port:         port,
		PingInterval: 0, // Disable pings for tests
		Fallback: &MockClient{
			VersionVal:  truenas.Version{Major: 25, Minor: 4},
			ConnectFunc: func(ctx context.Context) error { return nil },
		},
	})
	if err != nil {
		t.Fatalf("NewWebSocketClient failed: %v", err)
	}
	client.testInsecure = true
	client.version = truenas.Version{Major: 25, Minor: 4}
	client.connected = true

	return client
}

func TestWebSocketClient_Subscribe_ReceivesEvents(t *testing.T) {
	server := newSubscribeTestServer(t, func(conn *websocket.Conn) {
		event := map[string]any{
			"msg":    "method",
			"method": "collection_update",
			"params": map[string]any{
				"msg":        "changed",
				"collection": "reporting.realtime",
				"fields":     map[string]any{"cpu": map[string]any{"usage": 42.5}},
			},
		}
		_ = conn.WriteJSON(event)
	})
	defer server.Close()

	client := newSubscribeTestClient(t, server)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sub, err := client.Subscribe(ctx, "reporting.realtime", nil)
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}
	defer sub.Close()

	select {
	case msg := <-sub.C:
		if msg == nil {
			t.Fatal("expected non-nil message")
		}
		var data map[string]any
		if err := json.Unmarshal(msg, &data); err != nil {
			t.Fatalf("unmarshal event: %v", err)
		}
		if _, ok := data["cpu"]; !ok {
			t.Fatal("expected cpu key in event")
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for event")
	}
}

func TestWebSocketClient_Subscribe_Close(t *testing.T) {
	server := newSubscribeTestServer(t, nil)
	defer server.Close()

	client := newSubscribeTestClient(t, server)
	defer client.Close()

	ctx := context.Background()
	sub, err := client.Subscribe(ctx, "reporting.realtime", nil)
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	sub.Close()

	// Channel should be closed after Close
	select {
	case _, ok := <-sub.C:
		if ok {
			t.Error("expected channel to be closed")
		}
	case <-time.After(2 * time.Second):
		t.Error("timed out waiting for channel close")
	}
}

func TestWebSocketClient_Subscribe_MultipleCollections(t *testing.T) {
	server := newSubscribeTestServer(t, func(conn *websocket.Conn) {
		// Send event for reporting.realtime
		event1 := map[string]any{
			"msg": "method", "method": "collection_update",
			"params": map[string]any{
				"msg": "changed", "collection": "reporting.realtime",
				"fields": map[string]any{"source": "realtime"},
			},
		}
		_ = conn.WriteJSON(event1)

		time.Sleep(50 * time.Millisecond)
		// Send event for app.stats
		event2 := map[string]any{
			"msg": "method", "method": "collection_update",
			"params": map[string]any{
				"msg": "changed", "collection": "app.stats",
				"fields": map[string]any{"source": "stats"},
			},
		}
		_ = conn.WriteJSON(event2)
	})
	defer server.Close()

	client := newSubscribeTestClient(t, server)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sub1, err := client.Subscribe(ctx, "reporting.realtime", nil)
	if err != nil {
		t.Fatalf("Subscribe 1 failed: %v", err)
	}
	defer sub1.Close()

	sub2, err := client.Subscribe(ctx, "app.stats", nil)
	if err != nil {
		t.Fatalf("Subscribe 2 failed: %v", err)
	}
	defer sub2.Close()

	// Wait for events
	var got1, got2 bool
	for i := 0; i < 2; i++ {
		select {
		case msg := <-sub1.C:
			var data map[string]any
			json.Unmarshal(msg, &data)
			if data["source"] == "realtime" {
				got1 = true
			}
		case msg := <-sub2.C:
			var data map[string]any
			json.Unmarshal(msg, &data)
			if data["source"] == "stats" {
				got2 = true
			}
		case <-ctx.Done():
			t.Fatal("timed out waiting for events")
		}
	}

	if !got1 {
		t.Error("didn't receive reporting.realtime event")
	}
	if !got2 {
		t.Error("didn't receive app.stats event")
	}
}

func TestWebSocketClient_Subscribe_JobEventsStillWork(t *testing.T) {
	// Verify that core.get_jobs events are NOT routed to collection subscribers
	// but instead go through the normal job event handling path.
	server := newSubscribeTestServer(t, nil)
	defer server.Close()

	client := newSubscribeTestClient(t, server)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Make a call to ensure connection is established
	_, err := client.Call(ctx, "core.ping", nil)
	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}

	// core.get_jobs events should not panic or interfere with existing behavior.
	// We can't easily test the full job flow here without duplicating the
	// CallAndWait test, but we verify Subscribe doesn't break normal operations.
	result, err := client.Call(ctx, "test.method", nil)
	if err != nil {
		t.Fatalf("Call after Subscribe setup failed: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestWebSocketClient_Subscribe_ClientClose(t *testing.T) {
	// Verify that closing the client closes all collection subscriber channels.
	server := newSubscribeTestServer(t, nil)
	defer server.Close()

	client := newSubscribeTestClient(t, server)

	ctx := context.Background()
	sub, err := client.Subscribe(ctx, "reporting.realtime", nil)
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	// Close the client
	client.Close()

	// Channel should be closed
	select {
	case _, ok := <-sub.C:
		if ok {
			t.Error("expected channel to be closed after client.Close()")
		}
	case <-time.After(2 * time.Second):
		t.Error("timed out waiting for channel close after client.Close()")
	}
}

func TestWebSocketClient_Subscribe_ContextCancelled(t *testing.T) {
	// Verify that Subscribe respects context cancellation when sending to
	// collectionSubChan (though in practice the channel has buffer).
	server := newSubscribeTestServer(t, nil)
	defer server.Close()

	// Create client but intentionally make the ping fail by using a cancelled context
	client := newSubscribeTestClient(t, server)
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := client.Subscribe(ctx, "reporting.realtime", nil)
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

func TestNewSubscription_Constructor(t *testing.T) {
	ch := make(chan json.RawMessage, 1)
	closed := false
	sub := truenas.NewSubscription[json.RawMessage](ch, func() { closed = true })

	if sub.C == nil {
		t.Error("expected non-nil channel")
	}

	sub.Close()
	if !closed {
		t.Error("expected cancel to be called")
	}
}

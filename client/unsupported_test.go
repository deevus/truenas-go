package client

import (
	"context"
	"errors"
	"testing"

	truenas "github.com/deevus/truenas-go"
)

func TestUnsupportedClient_NoOps(t *testing.T) {
	c := &UnsupportedClient{}
	ctx := context.Background()

	if err := c.Connect(ctx); err != nil {
		t.Fatalf("Connect() = %v, want nil", err)
	}
	if err := c.Close(); err != nil {
		t.Fatalf("Close() = %v, want nil", err)
	}
	if v := c.Version(); v != (truenas.Version{}) {
		t.Fatalf("Version() = %v, want zero-value", v)
	}
}

func TestUnsupportedClient_OperationsReturnError(t *testing.T) {
	c := &UnsupportedClient{}
	ctx := context.Background()

	tests := []struct {
		name string
		fn   func() error
	}{
		{"Call", func() error { _, err := c.Call(ctx, "test", nil); return err }},
		{"CallAndWait", func() error { _, err := c.CallAndWait(ctx, "test", nil); return err }},
		{"WriteFile", func() error { return c.WriteFile(ctx, "/test", truenas.WriteFileParams{}) }},
		{"ReadFile", func() error { _, err := c.ReadFile(ctx, "/test"); return err }},
		{"DeleteFile", func() error { return c.DeleteFile(ctx, "/test") }},
		{"RemoveDir", func() error { return c.RemoveDir(ctx, "/test") }},
		{"RemoveAll", func() error { return c.RemoveAll(ctx, "/test") }},
		{"FileExists", func() error { _, err := c.FileExists(ctx, "/test"); return err }},
		{"Chown", func() error { return c.Chown(ctx, "/test", 0, 0) }},
		{"ChmodRecursive", func() error { return c.ChmodRecursive(ctx, "/test", 0644) }},
		{"MkdirAll", func() error { return c.MkdirAll(ctx, "/test", 0755) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			if !errors.Is(err, ErrUnsupportedOperation) {
				t.Errorf("got %v, want ErrUnsupportedOperation", err)
			}
		})
	}
}

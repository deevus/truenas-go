# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

Go client library for the TrueNAS middleware API. Provides typed service methods over SSH and WebSocket transports, with version-aware API resolution across TrueNAS 24.x and 25.x releases.

Module: `github.com/deevus/truenas-go`

## Commands

```bash
# Build
go build ./...

# Test (CI runs with race detector)
go test ./... -race -v
go test ./client/... -v              # client package only
go test ./... -run TestSnapshotService -v  # single test

# Vet
go vet ./...
```

## Architecture

### Three-layer design

1. **Client layer** (`client/` package) — transport abstraction. Two implementations: `SSHClient` (calls `midclt` over SSH) and `WebSocketClient` (JSON-RPC 2.0 over WebSocket). Both implement `client.Client`. The WebSocket client optionally delegates file operations (ReadFile, DeleteFile, RemoveDir, RemoveAll) to an SSH fallback; without one, an `UnsupportedClient` returns `ErrUnsupportedOperation`.

2. **Service layer** (root package) — typed business logic. Each service (`SnapshotService`, `DatasetService`, `AppService`, etc.) takes a caller interface and a `Version`, then maps Go methods to TrueNAS API calls with version-aware method resolution.

3. **Domain models** (root package) — `Snapshot`, `Dataset`, `App`, `VM`, `CloudSyncTask`, etc. Parsed from JSON-RPC responses.

### Interface hierarchy

Services depend on the narrowest caller interface they need:

```
Caller                    → Call()
  └─ AsyncCaller          → Call() + CallAndWait()
       └─ FileCaller      → Call() + CallAndWait() + file operations
```

Services declare which they require: `NewSnapshotService(c Caller, v Version)` vs `NewAppService(c AsyncCaller, v Version)` vs `NewFilesystemService(c FileCaller, v Version)`.

### File conventions

Each service follows a three-file pattern:

| File | Contents |
|------|----------|
| `xxx_service.go` | Service struct, constructor, domain models, API methods |
| `xxx_service_iface.go` | Exported `XxxServiceAPI` interface + `MockXxxService` struct |
| `xxx_service_iface_test.go` | Compile-time interface checks, mock default/func tests |

### Version-aware method resolution

API namespaces change across TrueNAS versions. Services resolve this at call time:

```go
// Pre-25.10: "zfs.snapshot.create", 25.10+: "pool.snapshot.create"
func resolveSnapshotMethod(v Version, method string) string
```

`Version.AtLeast(major, minor)` is the standard check. `Version.IsZero()` indicates undetected.

### WebSocket client internals

- Single `writerLoop` goroutine owns connection state (connect, auth, write, reconnect)
- `readerLoop` goroutine forwards responses and events
- Job subscriptions via `core.subscribe` with local filtering by job ID
- Reconnect handling: notifies job subscribers of disconnect/reconnect, polls to catch missed events
- Version detection: prefers fallback version if available, otherwise calls `system.version` natively

### Mock patterns

Function-pointer mocks throughout (no reflection):

```go
// client package — exported
type MockClient struct {
    ReadFileFunc func(ctx context.Context, path string) ([]byte, error)
}

// root package — exported (for consumers)
type MockSnapshotService struct {
    CreateFunc func(ctx context.Context, opts CreateSnapshotOpts) (*Snapshot, error)
}

// root package — unexported (internal test helpers in mock_test.go)
type mockCaller struct { callFunc func(...) }
type mockAsyncCaller struct { mockCaller; callAndWaitFunc func(...) }
type mockFileCaller struct { mockAsyncCaller; writeFileFunc func(...); ... }
```

When the function field is nil, mock methods return zero values (nil, nil).

### Error handling

- Services return `(nil, nil)` for not-found (checked via `isNotFoundError` matching "does not exist", "[ENOENT]", "not found")
- `client.TrueNASError` is a structured error with Code, Message, Field, Suggestion, JobID
- `ParseTrueNASError` extracts error codes like EINVAL, ENOENT, EFAULT from middleware output
- `EnrichAppLifecycleError` enriches app deployment errors by reading `/var/log/app_lifecycle.log` (best-effort, silently fails)

### SSH command safety

`midclt.BuildCommand` validates method names against `^[a-z][a-z0-9_.]+$` to prevent injection, and shell-escapes all parameters.

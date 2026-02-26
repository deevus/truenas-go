# truenas-go

Go client library for the [TrueNAS](https://www.truenas.com/) middleware API.

```go
import (
    truenas "github.com/deevus/truenas-go"
    "github.com/deevus/truenas-go/client"
)
```

## Install

```
go get github.com/deevus/truenas-go
```

Requires Go 1.25+.

For more complete usage, see [terraform-provider-truenas](https://github.com/deevus/terraform-provider-truenas) which uses this library extensively.

## Usage

### WebSocket client (TrueNAS 25.0+)

```go
c, err := client.NewWebSocketClient(client.WebSocketConfig{
    Host:     "truenas.local",
    Username: "root",
    APIKey:   "1-aBcDeFg...",
})
if err != nil {
    log.Fatal(err)
}
defer c.Close()

ctx := context.Background()
if err := c.Connect(ctx); err != nil {
    log.Fatal(err)
}

snapshots := truenas.NewSnapshotService(c, c.Version())
snap, err := snapshots.Create(ctx, truenas.CreateSnapshotOpts{
    Dataset: "tank/data",
    Name:    "backup",
})
```

### SSH client

```go
c, err := client.NewSSHClient(client.SSHConfig{
    Host:               "truenas.local",
    PrivateKey:         string(key),
    HostKeyFingerprint: "SHA256:...",
})
if err != nil {
    log.Fatal(err)
}
defer c.Close()

ctx := context.Background()
if err := c.Connect(ctx); err != nil {
    log.Fatal(err)
}

datasets := truenas.NewDatasetService(c, c.Version())
ds, err := datasets.GetDataset(ctx, "tank/data")
```

### WebSocket with SSH fallback

Some operations (ReadFile, DeleteFile, RemoveDir, RemoveAll) require SSH. Pass an SSH client as the fallback:

```go
ssh, _ := client.NewSSHClient(client.SSHConfig{...})
ws, _ := client.NewWebSocketClient(client.WebSocketConfig{
    Host:     "truenas.local",
    Username: "root",
    APIKey:   "1-aBcDeFg...",
    Fallback: ssh,
})
```

Without a fallback, these operations return `client.ErrUnsupportedOperation`.

## Services

| Service | Interface | Constructor |
|---------|-----------|-------------|
| Snapshots | `SnapshotServiceAPI` | `NewSnapshotService(Caller, Version)` |
| Datasets & Pools | `DatasetServiceAPI` | `NewDatasetService(Caller, Version)` |
| Apps & Registries | `AppServiceAPI` | `NewAppService(AsyncCaller, Version)` |
| Cloud Sync | `CloudSyncServiceAPI` | `NewCloudSyncService(AsyncCaller, Version)` |
| Cron Jobs | `CronServiceAPI` | `NewCronService(Caller, Version)` |
| Filesystem | `FilesystemServiceAPI` | `NewFilesystemService(FileCaller, Version)` |
| VMs | `VMServiceAPI` | `NewVMService(AsyncCaller, Version)` |
| Virt (Containers) | `VirtServiceAPI` | `NewVirtService(AsyncCaller, Version)` |

For the full per-method breakdown of which API endpoints are implemented and tested, see the [Feature Matrix](FEATURES.md). The library currently targets the latest stable release, **TrueNAS 25.04**.

To regenerate the feature matrix: `go run ./cmd/featurematrix -o FEATURES.md`

## Testing

Every service interface has a corresponding mock:

```go
mock := &truenas.MockSnapshotService{
    GetFunc: func(ctx context.Context, id string) (*truenas.Snapshot, error) {
        return &truenas.Snapshot{ID: "tank/data@backup"}, nil
    },
}

// Use mock wherever SnapshotServiceAPI is accepted
var svc truenas.SnapshotServiceAPI = mock
```

## Version support

The library handles API differences between TrueNAS versions automatically. Services resolve the correct API method at call time based on the detected version (e.g., `zfs.snapshot.*` on 24.x vs `pool.snapshot.*` on 25.10+).

- **WebSocket transport**: TrueNAS 25.0+ (JSON-RPC 2.0 over `/api/current`)
- **SSH transport**: TrueNAS 24.x and 25.x (calls `midclt` over SSH)

## License

[MIT](LICENSE)

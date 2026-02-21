# CronService Design

**Issue:** #45
**Date:** 2026-02-21

## Summary

Create a typed CronService in the truenas-go root package wrapping the cronjob.* API namespace. This is the simplest service and establishes the pattern for all others (#46-#52).

## Design Decisions

- **Root package** — service lives alongside existing types (`cron_service.go`), not in a subpackage. Consumers import just `truenas "github.com/deevus/truenas-go"`.
- **Takes `client.Client` directly** — no narrower interface abstraction. MockClient already exists for testing.
- **User-friendly stdout/stderr semantics** — `CaptureStdout`/`CaptureStderr` booleans (true = capture). Service handles the API's inverted `stdout`/`stderr` fields internally.

## Package Structure

```
truenas-go/
  cron.go              # existing CronJobResponse (wire type)
  cron_service.go      # NEW: CronService, CronJob, opts, helpers
  cron_service_test.go # NEW: tests with MockClient
```

## Types

### User-facing (new)

```go
type CronJob struct {
    ID            int64
    User          string
    Command       string
    Description   string
    Enabled       bool
    CaptureStdout bool   // inverted from API's stdout field
    CaptureStderr bool   // inverted from API's stderr field
    Schedule      Schedule
}

type Schedule struct {
    Minute string
    Hour   string
    Dom    string
    Month  string
    Dow    string
}

type CreateCronJobOpts struct {
    User          string
    Command       string
    Description   string
    Enabled       bool
    CaptureStdout bool
    CaptureStderr bool
    Schedule      Schedule
}

type UpdateCronJobOpts = CreateCronJobOpts
```

### Wire format (existing, unchanged)

`CronJobResponse` and `ScheduleResponse` — used internally by the service for JSON unmarshalling.

## Service API

```go
type CronService struct {
    client client.Client
}

func NewCronService(c client.Client) *CronService

func (s *CronService) Create(ctx context.Context, opts CreateCronJobOpts) (*CronJob, error)
func (s *CronService) Get(ctx context.Context, id int64) (*CronJob, error)
func (s *CronService) List(ctx context.Context) ([]CronJob, error)
func (s *CronService) Update(ctx context.Context, id int64, opts UpdateCronJobOpts) (*CronJob, error)
func (s *CronService) Delete(ctx context.Context, id int64) error
```

- `Get` returns `nil, nil` when not found (convention for detecting out-of-band deletions).
- `Create` and `Update` re-query after the write to return full state.
- Errors pass through from `client.Call()` — no additional wrapping.

## Internal Helpers

- `optsToParams(opts CreateCronJobOpts) map[string]any` — builds API params, inverts stdout/stderr.
- `cronJobFromResponse(resp CronJobResponse) CronJob` — converts wire to user-facing, inverts stdout/stderr.

## Testing

Tests use `client.MockClient`:

- Create: mock returns `{"id": 1}`, then re-query returns full response. Verify inversion.
- Get: mock returns `[{...}]`, verify mapping. Empty array returns `nil, nil`.
- List: mock returns `[{...}, {...}]`, verify all mapped.
- Update: mock verifies `[]any{id, params}` shape. Verify re-query.
- Delete: mock verifies method and ID.
- Error propagation for each method.

## Pattern Established

This service sets the convention for #46-#52:
- Constructor taking `client.Client`
- Typed opts/response structs (user-friendly, not raw API)
- Re-query after mutation
- `nil, nil` for not-found
- Tests with MockClient

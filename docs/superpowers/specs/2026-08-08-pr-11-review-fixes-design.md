# PR 11 review fixes design

**Status:** Approved
**Date:** 2026-08-08

## Context

PR 11 adds typed user and group modules. CodeRabbit reported correctness, test, and dependency issues after the initial implementation. This design covers the accepted review fixes.

TrueNAS implements `user.update` with PATCH semantics. Its versioned models accept any subset of update fields. The middleware loads the current user and overlays only the submitted keys. See [the research brief](../../research/truenas-user-update-partial-semantics.md).

## Goals

- Make `UpdateUserOpts` represent omission separately from explicit zero values.
- Prevent create and update methods from returning `(nil, nil)` after a successful write.
- Reject malformed create responses before a lookup uses ID zero.
- Prevent `ArgSchema` from panicking on a negative index.
- Verify that `MockUserService.Delete` forwards `deleteGroup` unchanged.
- Upgrade `golang.org/x/crypto` to a release that fixes the reported SSH vulnerabilities.

## Non-goals

- Do not add version-based resolution for `group.*` method names. No supported version changes those names.
- Do not add `home_create` or `random_password` to `UpdateUserOpts` in this review-fix change.
- Do not replace explicit wire-to-domain conversion with direct struct conversion.
- Do not extract a shared create/update parameter mapper before the update interface is stable.
- Do not change create option semantics.

## Public interface

`UpdateUserOpts` will keep strings for values that cannot be cleared with an empty string. An empty value omits those fields. Slices keep their existing nil-versus-empty behavior.

The following fields will become pointers:

```go
Email              *string
PasswordDisabled   *bool
SMB                *bool
SSHPasswordEnabled *bool
SSHPubKey           *string
Locked              *bool
```

The package will add `BoolPtr`, next to the existing pointer helpers. Callers can use `StringPtr` for `Email` and `SSHPubKey`.

### Update conversion rules

| Go value | Wire behavior |
|---|---|
| `Email == nil` | Omit `email`; preserve the current address. |
| `Email == StringPtr("")` | Send `email: null`; clear the current address. |
| Non-empty `Email` | Send the address. |
| `SSHPubKey == nil` | Omit `sshpubkey`; preserve authorized keys. |
| `SSHPubKey == StringPtr("")` | Send an empty key; remove authorized keys. |
| Boolean pointer is `nil` | Omit the field; preserve the current value. |
| Boolean pointer is non-nil | Send the pointed value, including `false`. |
| Slice is `nil` | Omit the field. |
| Slice is empty | Send an empty array to clear the collection. |
| Other string is empty | Omit the field. |
| `Group == 0` | Omit `group`. |

These rules match TrueNAS 24.10, 25.04, and 25.10.

## Write-result errors

`UserService.Create`, `UserService.Update`, `GroupService.Create`, and `GroupService.Update` fetch the object after a successful write. If the fetch returns no object and no error, the write method will return a descriptive error instead of `(nil, nil)`.

Examples:

- `user 42 not found after create`
- `user 42 not found after update`
- `group 42 not found after create`
- `group 42 not found after update`

A transport or parse error from the fetch will pass through unchanged.

## Create-response validation

`UserService.Create` supports both known response shapes:

- a numeric ID on TrueNAS 24.x;
- an object containing `id` on TrueNAS 25.04 and later.

Both shapes must contain an ID greater than zero. JSON `null`, `{}`, numeric zero, and an object with ID zero will return `parse create response` errors. Invalid JSON will keep its existing parse error behavior.

## Schema bounds

`api.ArgSchema` will reject `argIdx < 0` and `argIdx >= len(def.Accepts)`. Both cases will return an error. Neither case may panic.

## Dependency update

Upgrade `golang.org/x/crypto` from `v0.48.0` to `v0.54.0`. Run `go mod tidy` so compatible indirect versions of `x/sys`, `x/term`, and `x/text` are recorded. This change remains limited to module metadata.

## Test strategy

Use vertical red-green-refactor cycles. Each cycle starts with one failing behavior test through a public interface or transport seam.

1. Call `api.ArgSchema` with a negative index and verify an error instead of a panic.
2. Call `UserService.Create` with malformed zero-ID response shapes and verify parse errors.
3. Call each new write method with a post-write not-found response and verify its descriptive error.
4. Call `UserService.Update` with one field at a time and inspect the request sent through `Caller`:
   - omission preserves fields;
   - explicit `false` is sent;
   - empty email clears with `null`;
   - empty SSH key is sent to clear keys;
   - empty slices are sent while nil slices are omitted.
5. Call `MockUserService.Delete` with `true` and verify that `DeleteFunc` receives `true`.
6. Update module metadata and run the full verification suite.

Mocks are allowed only at the `Caller` transport seam or through the exported mock module interface.

## Verification

Run these checks after the focused cycles:

```bash
go mod tidy
gofmt -w api/schema.go api/schema_test.go dataset_service.go user_service.go user_service_test.go user_service_iface_test.go group_service.go group_service_test.go
go test ./... -race -v
go vet ./...
git diff --check
```

The final diff must contain no version resolver, no unrelated user fields, and no generated feature-matrix changes unless test counts require regeneration.

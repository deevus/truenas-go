# TrueNAS Go ecosystem repository design

**Status:** Accepted

## Summary

Move the Terraform provider source into `truenas-go`. Keep `truenas-go` as the canonical repository and retain its existing Go module path. Use one Go module so library and provider changes can share one pull request and test run.

Keep `terraform-provider-truenas` as a publication repository. It preserves the existing Terraform Registry identity and receives generated documentation, plain version tags, and signed release artifacts. It does not contain canonical source.

See [CONTEXT.md](../../../CONTEXT.md) for the agreed terminology and [ADR-0001](../../adr/0001-use-truenas-go-as-the-canonical-ecosystem-repository.md) for the decision record.

## Goals

- Change the client library and Terraform provider atomically.
- Test provider consumers against library changes before merge.
- Remove library tagging and dependency bumps from active provider development.
- Preserve `github.com/deevus/truenas-go` for library consumers.
- Preserve `registry.terraform.io/deevus/truenas` for Terraform users.
- Keep the client library conceptually independent from Terraform.
- Preserve both repositories' histories where practical.

## Non-goals

- Combine library and provider release versions.
- Make the Terraform provider an importable Go module.
- Mirror canonical source into the publication repository.
- Move `pixels` or `truenas-tui` into the canonical repository.
- Refactor unrelated client or provider behavior during migration.

## Repository layout

```text
truenas-go/
├── go.mod
├── go.sum
├── *.go
├── api/
├── client/
├── cmd/
│   ├── featurematrix/
│   └── terraform-provider-truenas/
│       └── main.go
├── internal/
│   └── terraformprovider/
│       ├── provider/
│       ├── resources/
│       ├── datasources/
│       ├── services/
│       └── types/
├── terraform-provider-truenas/
│   ├── docs/
│   ├── examples/
│   ├── templates/
│   ├── terraform-registry-manifest.json
│   └── .goreleaser.yml
└── .github/workflows/
    ├── test.yml
    ├── release-library.yml
    └── release-terraform-provider-truenas.yml
```

Public library packages remain at the module root, `api/`, and `client/`. The provider command is the composition root. All Terraform implementation packages remain private under `internal/terraformprovider`.

The `terraform-provider-truenas/` directory contains non-Go assets for that integration product. Its full name avoids overloading the term “provider” and leaves room for future products.

## Dependency direction

```text
cmd/terraform-provider-truenas
              │
              ▼
internal/terraformprovider
              │
              ▼
github.com/deevus/truenas-go
├── service interfaces
├── domain models
└── client transports
```

The provider can import the public library and transport packages. Public library packages must not import `internal/terraformprovider` or any Terraform dependency. Terraform types and dependencies must not appear in public library interfaces.

The existing service interfaces and mocks remain the main seam between the products. Shared behavior belongs in the client library only when it is useful outside the Terraform provider.

## Development and CI

Every pull request runs the full module checks:

```bash
go test ./... -race
go vet ./...
```

Do not add path filters initially. The combined test surface is small enough to run in full, and full testing is the main benefit of the migration.

Add an import check for the dependency rules. It must reject imports from public library packages into `internal/terraformprovider`. It must also reject Terraform dependencies outside `cmd/terraform-provider-truenas`, `internal/terraformprovider`, and approved provider asset-generation tooling. Provider tests continue to use the exported library mocks and provider-local test helpers.

## Versioning

Library and provider releases use separate tag namespaces in the canonical repository:

```text
v0.6.0
terraform-provider-truenas/v0.16.0
```

A `vX.Y.Z` tag releases the Go module. A `terraform-provider-truenas/vX.Y.Z` tag identifies the canonical source commit for a provider binary release. The provider tag is not a Go module version.

The publication repository receives a plain tag for Terraform Registry compatibility:

```text
v0.16.0
```

Existing library tags remain in the canonical repository. Historical plain provider tags remain only in the publication repository and its legacy source branch. Do not import those plain tags into the canonical tag namespace. The first provider release after migration continues from the provider's current version sequence and uses the new canonical prefix.

## Provider publication

A canonical provider tag starts this workflow:

1. Run the complete test and vet checks.
2. Generate provider documentation under `terraform-provider-truenas/docs/`.
3. Build provider archives from `cmd/terraform-provider-truenas`.
4. Generate checksums, signatures, the Registry manifest, and provenance metadata.
5. Update `README.md` and root `docs/` in the publication repository.
6. Commit the documentation with the canonical tag and commit SHA in the message.
7. Create the plain provider tag on that documentation commit.
8. Create a draft GitHub Release and upload every artifact.
9. Verify the expected artifact set, checksums, signature, and canonical source reference.
10. Publish the release.

The publication repository contains only:

```text
terraform-provider-truenas/
├── README.md
└── docs/
```

The README states that the repository is a read-only publication target and links to the canonical source. Each documentation commit and release links to the exact canonical tag and commit SHA.

Publication must be idempotent. A failed run leaves a draft release and can resume from the immutable canonical tag. Automation must stop if an existing publication tag records a different canonical SHA.

## Dependency impact

A throwaway prototype compared the current library module with a combined module containing the provider's current Terraform dependencies.

The prototype ran on 2026-08-08 with Go `1.26.5-X:nodwarf5`. It used `truenas-go` commit `1b52bb4` and `terraform-provider-truenas` commit `88a2346`. The combined module added the provider's five direct requirements at their recorded versions. Test consumers imported the root package and `client/` through local replacements. Each comparison used isolated module and build caches.

The measured commands were `go list -m all`, `go mod graph`, `go test`, `go mod download all`, and `govulncheck ./...`. The results are directional because dependency versions and Go tooling can change before migration.

| Measure | Current library | Combined module |
|---|---:|---:|
| Module graph nodes | 8 | 32 |
| Consumer cold-test downloads | 5 archives / 8.1 MB | 5 archives / 8.1 MB |
| Consumer reachable modules | 7 | 7 |
| `go mod download all` | 18.7 MB | 113.2 MB |
| Repository `go.mod` size | 274 bytes | 1,530 bytes |
| Repository `go.sum` size | 1,370 bytes | 10,632 bytes |

Ordinary consumers importing both the root package and `client/` had identical downloads, build-cache size, reachable packages, module metadata, and `govulncheck` results. Go's lazy loading excluded unused Terraform packages.

Module-wide tools still see the larger graph. Dependency bots, full-download commands, repository-wide license checks, and some SBOM tools will include Terraform dependencies. This maintenance cost is accepted in exchange for atomic development.


## Migration

Use `truenas-go` as the migration target. Import the provider history under a temporary prefix before moving files into the approved layout. Preserve meaningful history and blame where practical, but prefer the clean final tree over exact historical paths.

Migration work includes:

1. Import the provider Git history into a migration branch.
2. Move the provider entrypoint to `cmd/terraform-provider-truenas`.
3. Move provider implementation packages under `internal/terraformprovider`.
4. Move provider documentation and release assets under `terraform-provider-truenas/`.
5. Rewrite provider imports to the canonical module path.
6. Merge provider dependencies into the root `go.mod` and `go.sum`.
7. Consolidate CI while preserving the provider's signing and Registry requirements.
8. Run all tests, vet checks, documentation generation, and a GoReleaser dry run.
9. Keep historical plain provider tags in the publication repository and retain its former source history on a legacy branch. Do not copy those tags into the canonical repository.
10. Replace its default branch contents with the publication README and generated documentation.

Do not change provider behavior during this migration. Behavioral changes require separate commits after the combined baseline passes.

## Failure handling

The migration must remain reversible until the first provider release succeeds. Keep backup refs for both repositories and do not remove existing release assets or tags.

Publication credentials must have access only to the publication repository. Build and sign artifacts before changing the publication tag. Publish the GitHub Release only after all verification checks pass.

If documentation publication succeeds but artifact publication fails, keep the release as a draft. A rerun must verify the canonical SHA before reusing the documentation commit and tag.

## Future integration products

A future integration product gets its own product-specific command, private implementation, asset directory, tag namespace, and release workflow. For example:

```text
cmd/pulumi-resource-truenas/
internal/pulumiprovider/
pulumi-provider-truenas/
```

Integration products depend on the client library and do not import each other. Reconsider nested Go modules only when measured dependency, ownership, or release pressure justifies the extra release choreography.

## Alternatives considered

### Make terraform-provider-truenas canonical

Rejected because it makes the reusable client library appear subordinate to Terraform and changes the library's established identity.

### Use nested Go modules and go.work

Rejected for now because active development would be fast, but cross-cutting releases would still require a library release and provider dependency bump.

### Use one release version for the library and provider

Rejected because the products have independent release needs and existing version sequences.

### Mirror the full canonical source repository

Rejected because the canonical repository and Terraform Registry require different root documentation, tags, and repository presentation.

### Publish a generated provider-only source projection

Rejected because source rewriting and generated module metadata add more release machinery than a documentation-and-artifact publication target.

## Validation

The migration is complete when:

- all existing library and provider tests pass together with the race detector;
- `go vet ./...` passes;
- the provider builds from `cmd/terraform-provider-truenas`;
- generated documentation matches the provider schemas;
- a GoReleaser dry run produces the expected archives, checksums, signature inputs, and manifest;
- a test publication creates correct documentation, tag mapping, provenance, and draft release in a disposable repository;
- existing library imports remain unchanged;
- the Terraform provider address remains `registry.terraform.io/deevus/truenas`;
- no Terraform package appears in a public library interface.

## Supporting research

See [Go monorepo exemplars](../../research/2026-08-08-go-monorepo-exemplars.md) for primary-source examples from containerd, Go tools, HashiCorp, gRPC-Go, OpenTelemetry Go, and Kubernetes.

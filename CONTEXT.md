# TrueNAS Go Ecosystem

This context defines the products and repositories that form the TrueNAS Go ecosystem.

## Language

**Client library**:
The public Go packages that provide typed access to the TrueNAS middleware interface and its transports.
_Avoid_: SDK, provider

**Integration product**:
A distributable product built on the client library for another tool or ecosystem.
_Avoid_: Provider when the target ecosystem is not clear

**Terraform provider**:
The integration product distributed as `registry.terraform.io/deevus/truenas`.
_Avoid_: Provider when more than one integration product is in scope

**Canonical repository**:
The `truenas-go` repository, which is the source of truth for the client library and integration products.
_Avoid_: Source repo. Use monorepo only when referring to the repository layout.

**Publication repository**:
A read-only repository used to publish versioned documentation, tags, and release artifacts to an external registry.
_Avoid_: Mirror, canonical repository

**Library release**:
A versioned release of the client library from the canonical repository.
_Avoid_: Ecosystem release

**Integration release**:
A versioned release of one integration product, independent of library releases.
_Avoid_: Provider release when the integration product is not Terraform

---
status: accepted
---

# Use truenas-go as the canonical ecosystem repository

Use `truenas-go` as the canonical repository for the client library and the Terraform provider. Keep one Go module so a pull request can change and test both products without an intermediate library release. Keep `terraform-provider-truenas` as a publication repository because the Terraform Registry requires that repository identity and plain semantic-version tags.

## Consequences

- The client library keeps the `github.com/deevus/truenas-go` module path.
- The Terraform provider has its own command, private implementation, assets, tag namespace, and release workflow.
- Terraform dependencies enter the root module graph, but package-level lazy loading prevents ordinary client consumers from downloading or building them.
- The publication repository contains generated documentation and release metadata, but not canonical source.

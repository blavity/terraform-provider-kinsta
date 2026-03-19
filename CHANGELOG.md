# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- `golangci-lint` pre-commit hook and `.golangci.yml` config (errcheck, staticcheck, gocritic, gofumpt, goimports, and more)
- Dependabot config for weekly Go module and GitHub Actions updates
- Terraform Registry-quality docs generated via `tfplugindocs` with custom templates, Import sections, and write-only field callouts
- `examples/provider/`, `examples/resources/kinsta_wordpress_site/`, `examples/resources/kinsta_wordpress_environment/` in tfplugindocs convention
- `templates/` directory for customizing generated docs across future runs
- `release-please` workflow for automated versioning and changelog generation from conventional commits
- `AGENTS.md` agent guide with toolchain table, hard stops, and commit scope list
- `.specify/memory/constitution.md` — 12-principle provider constitution

### Fixed
- `kinsta_wordpress_site` schema: removed invalid `Required + Computed` combination that Terraform rejects; write-only fields are now `Required` only with descriptions on all fields
- All `d.Set()` return values now checked (previously silently dropped errors)
- `resourceWordPressSiteRead` calls `d.SetId("")` on 404 via typed `IsNotFound` helper
- Both Delete operations now poll to completion before clearing state
- `api_key` provider field now has `Sensitive: true`
- All CI actions SHA-pinned with explicit `permissions` blocks

### Changed
- Release pipeline replaced: custom KMS/GCP script removed in favour of release-please + standard goreleaser GPG signing
- GoReleaser archive naming corrected to Registry convention (no `v` prefix in filenames)
- Example HCL uses generic placeholders; internal Vault/platform references removed

### Removed
- `scripts/build-registry.sh` — superseded by Terraform Registry's native GitHub Release integration
- `specs/` — internal AI planning artifacts not appropriate for a public repo
- `cmd/list-envs` — throwaway debug utility
- Old `examples/` structure replaced by tfplugindocs convention

## [0.0.2] - 2025-12-30

### Added
- GitHub Actions release pipeline with GoReleaser, GPG signing, and GitHub Pages registry publishing for `blavity.com/platform/kinsta`.

### Fixed
- Operation polling field names now match Kinsta API responses (`idSite`, `idEnv`), allowing site and environment creation to succeed.

### Removed
- `kinsta_application` resource (unsupported by current Kinsta API).

### Changed
- Module path updated to `github.com/blavity/terraform-provider-kinsta`.
- Test fixtures and docs updated for correct Kinsta API field names.

## [0.0.1] - 2025-12-11

### Added
- Initial provider implementation
- WordPress site resource (create, read, delete)
- WordPress environment resource
- Database resource
- Async operation polling
- Import support for WordPress sites

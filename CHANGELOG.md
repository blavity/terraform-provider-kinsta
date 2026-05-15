# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0](https://github.com/blavity/terraform-provider-kinsta/compare/v0.0.2...v0.1.0) (2026-05-15)


### Features

* Create LLM controller specification for Terraform provider split ([08155cd](https://github.com/blavity/terraform-provider-kinsta/commit/08155cd14c3457c9b27d8c05e27340fd5abcaf0c))
* Establish operations polling contract for Kinsta provider ([08155cd](https://github.com/blavity/terraform-provider-kinsta/commit/08155cd14c3457c9b27d8c05e27340fd5abcaf0c))
* **governance:** adopt local ADR governance pattern + baseline ADR ([#27](https://github.com/blavity/terraform-provider-kinsta/issues/27)) ([8ae22ec](https://github.com/blavity/terraform-provider-kinsta/commit/8ae22ec33982bc2403643dea29f710cad531bba5))
* initial implementation of Kinsta Terraform provider ([fd8bd15](https://github.com/blavity/terraform-provider-kinsta/commit/fd8bd15551271491c0a26ed7d6b402decf62cb49))
* Kinsta WordPress provider — Phases 0-5 complete ([#5](https://github.com/blavity/terraform-provider-kinsta/issues/5)) ([08155cd](https://github.com/blavity/terraform-provider-kinsta/commit/08155cd14c3457c9b27d8c05e27340fd5abcaf0c))
* **release:** adopt canonical plain-GPG signing pattern [closes [#1022](https://github.com/blavity/terraform-provider-kinsta/issues/1022)] ([#32](https://github.com/blavity/terraform-provider-kinsta/issues/32)) ([3890687](https://github.com/blavity/terraform-provider-kinsta/commit/38906871551d4b29e0df4661bb30ca1605773ac7))
* **release:** wire KMS-via-PKCS11 signing path for provider releases [closes blavity/platform[#1022](https://github.com/blavity/terraform-provider-kinsta/issues/1022)] ([#25](https://github.com/blavity/terraform-provider-kinsta/issues/25)) ([30f44a0](https://github.com/blavity/terraform-provider-kinsta/commit/30f44a09550050b2b0cd3ca57017f899cb34c4e4))


### Bug Fixes

* **wordpress-site:** make write-only fields Optional+Computed to fix import ([9de5f1e](https://github.com/blavity/terraform-provider-kinsta/commit/9de5f1ea5bf8f63d695717ea6f73666c0e19a47f))


### Miscellaneous

* Add phase 0 prompt for doc hygiene and architecture lock ([08155cd](https://github.com/blavity/terraform-provider-kinsta/commit/08155cd14c3457c9b27d8c05e27340fd5abcaf0c))
* Create phase prompts for Sevalla provider development ([08155cd](https://github.com/blavity/terraform-provider-kinsta/commit/08155cd14c3457c9b27d8c05e27340fd5abcaf0c))
* **deps:** bump actions/checkout from 4.2.2 to 6.0.2 ([#9](https://github.com/blavity/terraform-provider-kinsta/issues/9)) ([faada3c](https://github.com/blavity/terraform-provider-kinsta/commit/faada3c56a905eaf716d9af7136ee8b12d61b2fa))
* **deps:** bump actions/setup-go from 5.1.0 to 6.3.0 ([#12](https://github.com/blavity/terraform-provider-kinsta/issues/12)) ([4a695b0](https://github.com/blavity/terraform-provider-kinsta/commit/4a695b038e076affa804b01fb490885bc0c71151))
* **deps:** bump crazy-max/ghaction-import-gpg from 6.3.0 to 7.0.0 ([#10](https://github.com/blavity/terraform-provider-kinsta/issues/10)) ([058f363](https://github.com/blavity/terraform-provider-kinsta/commit/058f3630c27e187c9cc537e083fdb3899fd33658))
* **deps:** bump goreleaser/goreleaser-action from 6.4.0 to 7.0.0 ([#11](https://github.com/blavity/terraform-provider-kinsta/issues/11)) ([c53046d](https://github.com/blavity/terraform-provider-kinsta/commit/c53046d377e3aad6a1a6015657785869b8b2d759))
* **deps:** bump the go-modules group with 3 updates ([#13](https://github.com/blavity/terraform-provider-kinsta/issues/13)) ([447ce14](https://github.com/blavity/terraform-provider-kinsta/commit/447ce14785db3e895543794601e3a162a6f1cb57))
* disable automatic release workflow ([d50a876](https://github.com/blavity/terraform-provider-kinsta/commit/d50a876e54c694bd156b330272017d370a281533))
* Implement verification script for documentation hygiene ([08155cd](https://github.com/blavity/terraform-provider-kinsta/commit/08155cd14c3457c9b27d8c05e27340fd5abcaf0c))
* **renovate:** add shared preset config ([#6](https://github.com/blavity/terraform-provider-kinsta/issues/6)) ([01bcb7f](https://github.com/blavity/terraform-provider-kinsta/commit/01bcb7f3f09e8962e6ec3bafa8293a300144b1d3))

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

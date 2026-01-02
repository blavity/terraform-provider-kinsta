# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

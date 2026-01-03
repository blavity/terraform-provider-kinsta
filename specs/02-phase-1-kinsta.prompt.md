# Phase 1 (Kinsta): Cleanup & Documentation

## Context
- Repository: terraform-provider-kinsta (api.kinsta.com)
- Phase 0 complete: ADR and operations polling contract exist
- Existing resources need refinement and documentation before adding new features

## Goal
Stabilize existing WordPress resources and deprecate kinsta_database before implementing new features

## Tasks

### 1. Deprecate kinsta_database
- Add deprecation warning in resource description
- Update docs/resources/database.md with:
  - Deprecation notice at top
  - Migration guide to sevalla_database
  - Link to Sevalla provider
  - Removal timeline (next major version)
- NO bug fixes (NO users confirmed by stakeholder)

### 2. Refine kinsta_wordpress_site
- Create specs/20-kinsta-wordpress-site-resource.md with:
  - Complete schema mapping from swagger.json
  - Field validation rules
  - ForceNew rules
  - Async operation handling (polls operation_id)
  - Test plan
- Add missing schema fields to internal/provider/wordpress_site_resource.go:
  - is_multisite (bool, Optional, ForceNew)
  - is_subdomain_multisite (bool, Optional, ForceNew)
  - woocommerce (bool, Optional, ForceNew)
  - wordpressseo (bool, Optional, ForceNew)
- Add corresponding fields to internal/client/wordpress.go structs
- Update unit tests in wordpress_site_resource_unit_test.go
- Update acceptance tests in wordpress_site_resource_test.go
- Update docs/resources/wordpress_site.md with new fields

### 3. Document kinsta_wordpress_environment
- Create specs/21-kinsta-wordpress-environment-resource.md documenting:
  - Environment ID discovery pattern (before/after comparison)
  - Eventual consistency handling
  - Write-only fields with DiffSuppressFunc
  - Import format: site_id:env_id
  - Parent-child relationship with wordpress_site
  - Test plan
- Create docs/resources/wordpress_environment.md with:
  - Description and usage
  - Argument reference
  - Attribute reference
  - Import instructions
  - Example configuration
- Create examples/wordpress_environment/ directory with main.tf
- Create acceptance tests: internal/provider/wordpress_environment_resource_test.go

## Validation
- All unit tests pass: `go test ./internal/provider -run "_unit_test" -v`
- All acceptance tests pass: `TF_ACC=1 go test ./internal/provider -v`
- Lint passes: `go vet ./...`
- Build succeeds: `go build`

## Success Criteria
- [ ] kinsta_database has deprecation warning in description and docs
- [ ] kinsta_wordpress_site has all 4 missing schema fields
- [ ] specs/20-kinsta-wordpress-site-resource.md exists
- [ ] specs/21-kinsta-wordpress-environment-resource.md exists
- [ ] docs/resources/wordpress_environment.md exists
- [ ] examples/wordpress_environment/main.tf exists
- [ ] All tests pass

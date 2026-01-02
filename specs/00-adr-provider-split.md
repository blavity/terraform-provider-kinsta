# ADR-001: Split Terraform Provider into Kinsta and Sevalla

**Status:** Accepted  
**Date:** 2026-01-01  
**Deciders:** Engineering Team  
**Evidence:** MyKinsta API swagger.json v1.87.0, Sevalla API sevalla.openapi.json v1.80.0

---

## Context

The MyKinsta API has deprecated all Platform-as-a-Service (PaaS) endpoints including applications, databases, static sites, and pipelines, in favor of the new Sevalla API. The deprecation is effective January 31, 2026.

**Evidence:**
- `swagger.json#/paths/~1applications/get/deprecated` = `true` (all 19 application endpoints)
- `swagger.json#/paths/~1databases/post/deprecated` = `true` (all 5 database endpoints)
- `swagger.json#/paths/~1static-sites/get/deprecated` = `true` (all 5 static-site endpoints)
- `swagger.json#/paths/~1pipelines/get/deprecated` = `true` (all 2 pipeline endpoints)

The current `terraform-provider-kinsta` implements:
- `kinsta_database` resource (using deprecated MyKinsta endpoint)
- `kinsta_wordpress_site` resource (using active MyKinsta endpoint)
- `kinsta_wordpress_environment` resource (using active MyKinsta endpoint)

Users are blocked from managing applications and static sites via Terraform. The database resource will cease functioning after the MyKinsta deprecation deadline.

---

## Decision

We will split the current provider into two separate providers:

### terraform-provider-kinsta (MyKinsta API)
**Base URL:** `https://api.kinsta.com/v2`

**MUST implement:**
- WordPress sites (`/sites`)
- WordPress environments (`/sites/{site_id}/environments`)
- WordPress domains (`/sites/environments/{env_id}/domains`)
- WordPress tools (`/sites/tools/*`)
- WordPress backups (`/sites/environments/{env_id}/backups`)
- WordPress SFTP (`/sites/environments/{env_id}/additional-sftp-accounts`)
- Company resources (`/company/{id}/*`)
- Domains and DNS (`/domains`)
- Operations polling (`/operations/{operation_id}`)

**MUST NOT implement:**
- Applications (deprecated: `swagger.json#/paths/~1applications/get/deprecated=true`)
- Databases (deprecated: `swagger.json#/paths/~1databases/post/deprecated=true`)
- Static sites (deprecated: `swagger.json#/paths/~1static-sites/get/deprecated=true`)
- Pipelines (deprecated: `swagger.json#/paths/~1pipelines/get/deprecated=true`)

**Total scope:** 56 active endpoints

### terraform-provider-sevalla (Sevalla API)
**Base URL:** `https://api.sevalla.com/v2`

**MUST implement:**
- Databases (`/databases`)
- Applications (`/applications`) - **Note:** See limitations below
- Static sites (`/static-sites`)
- Pipelines (`/pipelines`)
- Application deployments (`/applications/deployments`)
- Static site deployments (`/static-sites/deployments`)

**MUST NOT implement:**
- WordPress sites (`/sites/*`) - These endpoints exist in Sevalla spec but belong to Kinsta provider
- **Rationale:** Sevalla spec contains 29 WordPress endpoints for backward compatibility, but implementing them would create provider overlap and confusion. WordPress management must remain exclusively in `terraform-provider-kinsta`.

**Total scope:** 31 core endpoints (excluding 29 WordPress endpoints from total of 60)

---

## Consequences

### Positive

1. **Clear separation of concerns:**
   - WordPress hosting → Kinsta provider
   - PaaS (apps, databases, static sites) → Sevalla provider

2. **Future-proof:**
   - Kinsta provider survives MyKinsta deprecation
   - Sevalla provider uses non-deprecated API

3. **Better user experience:**
   - Users can manage WordPress and PaaS resources independently
   - Clear migration path for database users

### Negative

1. **Breaking change for database users:**
   - Must migrate from `kinsta_database` to `sevalla_database`
   - Requires manual state migration (import + state rm)
   - Field name changes: `region`→`location`, `db_type`→`type`, `size`→`resource_type`

2. **Two providers to maintain:**
   - Separate release cycles
   - Separate documentation
   - Separate issue tracking

3. **State migration complexity:**
   - Terraform cannot automatically move state between providers
   - Users must manually run: `terraform import sevalla_database.x <id>` then `terraform state rm kinsta_database.x`

### Known Limitations

#### Applications Resource Blocked
**Evidence:** `sevalla.openapi.json#/paths/~1applications` contains only GET method (no POST)

**Decision:** Implement `sevalla_applications` data source (P0), defer `sevalla_application` resource to P2 pending API availability.

**Impact:** Users cannot create applications via Terraform initially. Read-only access via data source.

#### Static Sites Priority Downgrade
**Evidence:** `swagger.json#/paths/~1static-sites/get/deprecated=true` confirms deprecation

**Decision:** Static sites remain P0 for Sevalla (non-deprecated API exists), but implementation priority adjusted based on team capacity.

---

## Migration Strategy

### Phase 0: Deprecation (Week 1-2)
1. Add deprecation warning to `kinsta_database` resource
2. Update documentation pointing to Sevalla provider
3. Publish migration guide

### Phase 1: Sevalla Bootstrap (Week 3-4)
1. Create `terraform-provider-sevalla` repository
2. Implement `sevalla_database` resource (synchronous, 200 response)
3. Implement `sevalla_databases` data source
4. Implement `sevalla_applications` data source (read-only)
5. Beta release

### Phase 2: User Migration (Week 5-8)
1. Support users migrating database resources
2. Monitor feedback and fix issues
3. Update examples and tutorials

### Phase 3: Kinsta Completion (Week 9-12)
1. Complete WordPress domain management
2. Add backup management
3. Add SFTP user management

### Phase 4: Sevalla Completion (Week 13+)
1. Static site resources
2. Application resource (when POST available)
3. Deployment resources
4. Advanced features

---

## State Migration Reality

**Critical:** Terraform does not support automatic state migration between providers.

**User migration steps:**
```bash
# 1. Add sevalla provider to configuration
terraform {
  required_providers {
    sevalla = {
      source = "your-org/sevalla"
      version = "~> 1.0"
    }
  }
}

# 2. Import existing database to sevalla provider
terraform import sevalla_database.main <database-id>

# 3. Remove from kinsta provider state
terraform state rm kinsta_database.main

# 4. Update configuration with new field names
# region → location
# db_type → type  
# size → resource_type

# 5. Plan should show no changes
terraform plan
```

**Breaking changes:**
- Field renames require configuration updates
- `db_password` and `db_user` now required (were auto-generated)
- New computed outputs available

---

## Authentication Strategy

Both providers use **identical authentication**:
- Bearer token (API key from MyKinsta)
- Same environment variable: `KINSTA_API_KEY`
- Same company ID context

**Provider configuration:**
```hcl
# Kinsta provider
provider "kinsta" {
  api_key    = var.kinsta_api_key
  company_id = var.company_id
}

# Sevalla provider
provider "sevalla" {
  api_key    = var.kinsta_api_key  # Same token
  company_id = var.company_id      # Same company
}
```

---

## Operations Polling Scope

**Kinsta provider only:**
- WordPress operations are asynchronous (202 + `operation_id`)
- Must poll `/operations/{operation_id}` until 200 or 500
- `operation.data` is OPAQUE (`swagger.json#/components/schemas/OperationResponse/properties/data={}`)
- Cannot rely on `data.idSite` or `data.idEnv` keys (observed but not guaranteed)
- Must use lookup-after-poll strategy for environment creation

**Sevalla provider:**
- Databases are synchronous (200 immediate response)
- No polling required for database operations
- Applications/static sites: TBD based on deployment patterns

See `specs/02-operations-polling-contract.md` for detailed Kinsta polling implementation.

---

## Exclusion Lists (Formal)

### Kinsta Provider MUST NOT Implement

Based on `swagger.json` deprecation flags:

- `/applications` (GET, deprecated)
- `/applications/{id}` (DELETE, GET, PUT, deprecated)
- `/applications/{name}` (GET, deprecated)
- `/applications/{id}/internal-connections` (POST, deprecated)
- `/applications/{id}/cdn/toggle-status` (POST, deprecated)
- `/applications/{id}/edge-cache/toggle-status` (POST, deprecated)
- `/applications/{id}/clear-cache` (POST, deprecated)
- `/applications/{id}/metrics/*` (8 GET endpoints, deprecated)
- `/applications/deployments` (POST, deprecated)
- `/applications/deployments/{deployment_id}` (GET, deprecated)
- `/applications/processes/{id}` (GET, PUT, deprecated)
- `/applications/promote` (POST, deprecated)
- `/databases` (GET, POST, deprecated)
- `/databases/{id}` (DELETE, GET, PUT, deprecated)
- `/databases/{name}` (GET, deprecated)
- `/static-sites` (GET, deprecated)
- `/static-sites/{id}` (DELETE, GET, PUT, deprecated)
- `/static-sites/deployments` (POST, deprecated)
- `/static-sites/deployments/{deployment_id}` (GET, deprecated)
- `/static-sites/deployments/redeploy` (POST, deprecated)
- `/pipelines` (GET, deprecated)
- `/pipelines/{id}/create-preview-app` (POST, deprecated)

**Total:** 29 deprecated endpoints

### Sevalla Provider MUST NOT Implement

Based on overlap with Kinsta provider scope:

- All `/sites/*` endpoints (29 endpoints present in Sevalla spec)
- **Rationale:** WordPress sites must be managed exclusively via Kinsta provider to avoid confusion and maintain clear separation of concerns.

---

## Success Criteria

**Phase 0 complete when:**
- [ ] Deprecation warnings added to kinsta_database
- [ ] Migration guide published
- [ ] Users notified via changelog

**Phase 1 complete when:**
- [ ] sevalla_database implements full CRUD
- [ ] sevalla_databases data source lists databases
- [ ] sevalla_applications data source provides read-only access
- [ ] Beta release published to Terraform Registry

**Phase 2 complete when:**
- [ ] 50+ users successfully migrated databases
- [ ] No critical migration bugs reported
- [ ] Documentation validated by users

**Final success when:**
- [ ] kinsta_database removed from Kinsta provider
- [ ] Sevalla provider covers 80% of PaaS use cases
- [ ] Both providers have 80%+ test coverage
- [ ] Community adoption > 100 users per provider

---

## References

- MyKinsta API Spec: `./swagger.json` (v1.87.0)
- Sevalla API Spec: `./_spec_cache/sevalla.openapi.json` (v1.80.0)
- Deprecation announcement: [Sevalla Overview](https://kinsta.com/docs/service-information/sevalla-overview/)
- Migration guide: `docs/migration/kinsta-to-sevalla.md` (to be created)

---

## Revision History

- 2026-01-01: Initial ADR based on spec analysis
- Status: Accepted, implementation pending

---

## Approval

This ADR documents the architectural decision. Implementation requires:
- Technical review (spec accuracy validation)
- Product review (user impact assessment)
- Engineering approval (implementation feasibility)

**Next steps:**
1. Review this ADR with team
2. Validate spec citations
3. Begin Phase 0 implementation

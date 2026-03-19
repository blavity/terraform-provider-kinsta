# Kinsta Provider Remaining Work

**Date:** 2026-01-03  
**Status:** Phases 0-4 Complete for Kinsta provider  
**Current Repository:** terraform-provider-kinsta (api.kinsta.com)

---

## Executive Summary

### What's Complete

✅ **Phase 0** - Documentation hygiene and architecture lock (BOTH providers)
- Canonical specs created (ADR, operations polling contract)
- Doc patches applied with OpenAPI evidence
- Architecture invariants locked and documented

✅ **Sevalla Provider** - Phases 1-4 complete (separate repository)
- Phase 1: Repository bootstrap
- Phase 2: sevalla_database resource
- Phase 3: Data sources (applications, databases, static_sites, pipelines)
- Phase 4: Deployment/action resources

✅ **Kinsta Provider** - Phases 0-4 complete (this repository)
- Phase 0: Documentation hygiene and architecture lock
- Phase 1: Repository setup and client foundation
- Phase 2: kinsta_database deprecation handling
- Phase 3: WordPress site and environment resources
- Phase 4: Acceptance tests for core resources

### What Remains for Kinsta Provider

The Kinsta provider repository currently has:
- ✅ `kinsta_wordpress_site` - Implemented with unit + acceptance tests
- ✅ `kinsta_wordpress_environment` - Implemented with unit + acceptance tests
- ⏳ `kinsta_database` - **Deprecated immediately** (NO users), remove after migration window
- ❌ WordPress domains, backups, SFTP, tools - Not yet implemented
- ❌ Data sources - Not yet implemented

---

## Current Implementation Status

### Existing Resources (in provider.go)

```go
ResourcesMap: map[string]*schema.Resource{
    "kinsta_database":              resourceDatabase(),        // DEPRECATED
    "kinsta_wordpress_site":        resourceWordPressSite(),   // NEEDS REFINEMENT
    "kinsta_wordpress_environment": resourceWordPressEnvironment(), // NEEDS DOCS
}
```

### Resource Health Assessment

#### 1. kinsta_database ❌ DEPRECATED
**Status:** Implemented but deprecated  
**Action:** Remove after Sevalla migration window  
**Evidence:** `swagger.json#/paths/~1databases/post/deprecated=true`  
**Strategy:** Deprecate immediately; no bugfix work (NO users confirmed); remove in future release

**Known Issues (not being fixed):**
- Missing ForceNew on immutable fields
- Auto-generates passwords instead of requiring input
- No 404 handling
- Field name mismatches (size vs resource_type, etc.)
- Missing computed fields

**Next Steps:**
1. Add deprecation warning message to resource
2. Update documentation with migration guide to sevalla_database
3. Remove resource after 60-90 day migration window

#### 2. kinsta_wordpress_site ✅ COMPLETE
**Status:** Implemented, tested, documented  
**Files:**
- `internal/provider/wordpress_site_resource.go`
- `internal/provider/wordpress_site_resource_test.go` (acceptance)
- `internal/provider/wordpress_site_resource_unit_test.go`
- `specs/20-kinsta-wordpress-site-resource.md` (spec)

**Current Implementation:**
- ✅ POST /sites (async with operation_id)
- ✅ GET /sites/{id}
- ✅ DELETE /sites/{id} (async)
- ✅ Full schema with all fields
- ✅ Unit tests with mock client
- ✅ Acceptance tests (basic, custom language, migrate mode)
- ✅ Operations polling integration
- ✅ Lookup-after-poll implementation
- ✅ Documentation with examples

**Priority:** ✅ P0 Complete

#### 3. kinsta_wordpress_environment ✅ COMPLETE
**Status:** Implemented, tested, documented  
**Files:**
- `internal/provider/wordpress_environment_resource.go`
- `internal/provider/wordpress_environment_resource_test.go` (acceptance)
- `internal/provider/wordpress_environment_resource_unit_test.go`

**Current Implementation:**
- ✅ POST /sites/{site_id}/environments
- ✅ DELETE /sites/environments/{env_id}
- ✅ GET /sites/{site_id} (read via environments list)
- ✅ Environment ID discovery (before/after comparison)
- ✅ Eventual consistency handling
- ✅ Write-only fields with DiffSuppressFunc
- ✅ Import support (site_id:env_id format)
- ✅ Unit tests
- ✅ Acceptance tests (basic, premium, custom settings)
- ✅ Documentation

**Priority:** ✅ P0 Complete

---

## Remaining Work Breakdown

### Phase 5: WordPress Domain Management (Week 1-2)
**Goal:** Implement domain resources for WordPress sites

#### 2.1 Implement kinsta_wordpress_domain
- [ ] Create `specs/22-kinsta-wordpress-domain-resource.md`
- [ ] Add client methods:
  - [ ] POST /sites/environments/{env_id}/domains
  - [ ] DELETE /sites/environments/{env_id}/domains
  - [ ] GET /sites/{site_id} (read via environments[].domains[])
  - [ ] PUT /sites/environments/{env_id}/change-primary-domain
- [ ] Implement resource with schema
- [ ] Handle domain verification status
- [ ] Handle SSL certificate status
- [ ] Support primary domain changes
- [ ] Support multiple domains per environment
- [ ] Write unit tests
- [ ] Write acceptance tests
- [ ] Write documentation
- [ ] Create examples

**Estimated Effort:** 2-3 days

### Phase 6: WordPress Operational Resources (Week 3-4)
**Goal:** Implement backup and SFTP management

#### 3.1 Implement kinsta_wordpress_backup
- [ ] Create `specs/23-kinsta-wordpress-backup-resource.md`
- [ ] Add client methods:
  - [ ] POST /sites/environments/{env_id}/manual-backups
  - [ ] POST /sites/environments/{target_env_id}/backups/restore
  - [ ] DELETE /sites/environments/backups/{backup_id}
  - [ ] GET /sites/environments/{env_id}/backups
  - [ ] GET /sites/environments/{env_id}/downloadable-backups
- [ ] Implement backup resource (manual backups)
- [ ] Handle async operations
- [ ] Implement restore functionality
- [ ] Write unit tests
- [ ] Write acceptance tests
- [ ] Write documentation
- [ ] Create examples

**Estimated Effort:** 2-3 days

#### 3.2 Implement kinsta_wordpress_sftp
- [ ] Create `specs/24-kinsta-wordpress-sftp-resource.md`
- [ ] Add client methods:
  - [ ] POST /sites/environments/{env_id}/additional-sftp-accounts
  - [ ] DELETE /sites/environments/additional-sftp-accounts/{sftp_account_id}
  - [ ] GET /sites/environments/{env_id}/additional-sftp-accounts
  - [ ] PUT /sites/environments/{env_id}/additional-sftp-accounts/toggle-status
- [ ] Implement SFTP resource
- [ ] Handle password management (sensitive)
- [ ] Support enable/disable toggle
- [ ] Write unit tests
- [ ] Write acceptance tests
- [ ] Write documentation
- [ ] Create examples

**Estimated Effort:** 1-2 days

### Phase 7: Data Sources (Week 5-6)
**Goal:** Implement read-only data sources for discovery

#### 4.1 Implement Core Data Sources
- [ ] Create `specs/26-kinsta-data-sources.md`
- [ ] Implement data sources:
  - [ ] `kinsta_wordpress_sites` - List all sites with filtering
  - [ ] `kinsta_wordpress_backups` - List backups for environment
  - [ ] `kinsta_company_regions` - Available regions
  - [ ] `kinsta_company_users` - List company users
- [ ] Support pagination where applicable
- [ ] Write unit tests
- [ ] Write documentation
- [ ] Create examples

**Estimated Effort:** 2-3 days

### Phase 8: WordPress Tools (Week 7-8)
**Goal:** Implement operational trigger resources

#### 5.1 Implement Tool Resources
- [ ] Create `specs/25-kinsta-wordpress-tools.md`
- [ ] Decide on implementation pattern:
  - Option A: Trigger resources (recommended)
  - Option B: terraform_data triggers
  - Option C: Provider functions (TF 1.8+)
- [ ] Implement resources:
  - [ ] `kinsta_wordpress_cache_clear` - POST /sites/tools/clear-cache
  - [ ] `kinsta_wordpress_php_restart` - POST /sites/tools/restart-php
  - [ ] `kinsta_wordpress_php_version` - PUT /sites/tools/modify-php-version
  - [ ] `kinsta_wordpress_denied_ips` - GET/PUT /sites/tools/denied-ips
- [ ] Document lifecycle semantics (create triggers action, delete no-op)
- [ ] Write unit tests
- [ ] Write acceptance tests
- [ ] Write documentation
- [ ] Create examples

**Estimated Effort:** 2-3 days

### Phase 9: Testing & Quality (Ongoing)
**Goal:** Establish comprehensive testing patterns

#### 6.1 Testing Standards
- [ ] Create `specs/30-testing-standards.md`
- [ ] Document unit test patterns
- [ ] Document acceptance test patterns
- [ ] Create test fixtures and helpers
- [ ] Implement random name generation utility
- [ ] Implement resource sweeping (cleanup)
- [ ] Set up CI/CD integration

**Estimated Effort:** 2-3 days

#### 6.2 Documentation Standards
- [ ] Create `specs/31-documentation-standards.md`
- [ ] Create resource documentation template
- [ ] Document argument reference format
- [ ] Document attributes reference format
- [ ] Document example usage patterns
- [ ] Document import documentation format
- [ ] Document timeout configuration
- [ ] Document common errors and solutions

**Estimated Effort:** 1-2 days

---

## Specification Files to Create

### Priority 0 (Complete)
- [x] `specs/00-adr-provider-split.md` - Provider split decision
- [x] `specs/02-operations-polling-contract.md` - Async operations spec
- [x] `specs/20-kinsta-wordpress-site-resource.md` - Site resource spec
- [x] `specs/03-phase-4-kinsta-complete.md` - Phase 4 completion report

### Priority 2 (Near-term - Week 1-2)
- [ ] `specs/22-kinsta-wordpress-domain-resource.md` - Domain management

### Priority 3 (Medium-term - Week 3-4)
- [ ] `specs/23-kinsta-wordpress-backup-resource.md` - Backup management
- [ ] `specs/24-kinsta-wordpress-sftp-resource.md` - SFTP management

### Priority 4 (Longer-term - Week 5+)
- [ ] `specs/25-kinsta-wordpress-tools.md` - Operational tools
- [ ] `specs/26-kinsta-data-sources.md` - Data sources
- [ ] `specs/30-testing-standards.md` - Testing patterns
- [ ] `specs/31-documentation-standards.md` - Documentation standards

---

## Key Architectural Decisions (Locked)

### ✅ kinsta_database Strategy
**Decision:** Deprecate immediately, remove after migration window  
**Rationale:** NO users confirmed, deprecated API, Sevalla provider exists  
**Action:** Do NOT fix bugs, add deprecation warning only

### ✅ Operations Polling
**Decision:** Treat operation.data as opaque  
**Evidence:** `swagger.json#/components/schemas/OperationResponse/properties/data={}`  
**Strategy:** Implement lookup-after-poll for all async operations  
**Spec:** `specs/02-operations-polling-contract.md`

### ✅ WordPress-Only Scope
**Decision:** Kinsta provider = WordPress only  
**Evidence:** All PaaS endpoints deprecated in MyKinsta  
**Rationale:** Clean separation (WordPress → Kinsta, PaaS → Sevalla)

### ✅ Base URL
**Current:** `https://api.kinsta.com/v2`  
**Correct:** Remains unchanged for Kinsta provider

---

## Timeline Estimate

### Optimistic (Single developer, full-time)
- **Phases 0-4 (Complete):** 2 weeks ✅
- **Phase 5:** 1 week (domains)
- **Phase 6:** 1.5 weeks (backups + SFTP)
- **Phase 7:** 1 week (data sources)
- **Phase 8:** 1 week (tools)
- **Phase 9:** Ongoing (testing/quality)

**Total Remaining:** 4.5 weeks

### Realistic (Part-time or with interruptions)
- **Phases 0-4 (Complete):** 3 weeks ✅
- **Phase 5:** 2 weeks
- **Phase 6:** 2 weeks
- **Phase 7:** 1.5 weeks
- **Phase 8:** 1.5 weeks
- **Phase 9:** Ongoing

**Total Remaining:** 7 weeks

---

## Next Steps (Immediate)

### Phase 5: WordPress Domain Management

#### Step 1: Create domain resource spec (4 hours)
```bash
# 1. Create spec file
# Create: specs/22-kinsta-wordpress-domain-resource.md

# 2. Document API endpoints
# - POST /sites/environments/{env_id}/domains
# - DELETE /sites/environments/{env_id}/domains
# - GET /sites/{site_id} (read via environments[].domains[])
# - PUT /sites/environments/{env_id}/change-primary-domain
```

#### Step 2: Implement domain resource (1-2 days)
```bash
# 1. Add client methods
# Edit: internal/client/client.go
# Add: GetDomains, AddDomain, RemoveDomain, SetPrimaryDomain

# 2. Implement resource
# Create: internal/provider/wordpress_domain_resource.go
# Create: internal/provider/wordpress_domain_resource_unit_test.go

# 3. Add to provider
# Edit: internal/provider/provider.go

# 4. Create acceptance tests
# Create: internal/provider/wordpress_domain_resource_test.go

# 5. Create documentation
# Create: docs/resources/wordpress_domain.md
# Create: examples/wordpress_domain/
```

---

## Success Criteria

### Phases 0-4 Complete ✅
- [x] Architecture decisions locked (ADR, operations polling contract)
- [x] kinsta_database has deprecation warning
- [x] kinsta_wordpress_site fully implemented with tests
- [x] kinsta_wordpress_environment fully implemented with tests
- [x] All existing tests pass
- [x] Documentation updated
- [x] Acceptance tests for core resources

### Provider Complete When:
- [ ] All WordPress endpoints covered (sites, environments, domains, backups, SFTP)
- [ ] All tool operations available
- [ ] Data sources for discovery implemented
- [ ] All resources have unit + acceptance tests
- [ ] All resources have docs + examples
- [ ] Import support documented
- [ ] kinsta_database removed
- [ ] Provider published to registry

---

## References

### Key Documents
- `specs/00-adr-provider-split.md` - Provider split architecture
- `specs/02-operations-polling-contract.md` - Async operations pattern
- `specs/20-kinsta-wordpress-site-resource.md` - Site resource spec
- `specs/03-phase-4-kinsta-complete.md` - Phase 4 completion report
- `ANALYSIS_SUMMARY.md` - Implementation analysis
- `SEVALLA_SPEC_FINDINGS.md` - Sevalla API analysis
- `SPECS_ROADMAP.md` - Detailed spec file roadmap
- `PHASE_0_COMPLETE.md` - Phase 0 completion report

### OpenAPI Specs
- MyKinsta: `swagger.json` (v1.87.0)
- Sevalla: `./_spec_cache/sevalla.openapi.json` (v1.80.0)

### MyKinsta API Endpoints (WordPress Only)
Total: 56 endpoints covering:
- Sites (13 endpoints)
- Environments (12 endpoints)
- Domains (5 endpoints)
- Tools (8 endpoints)
- Backups (6 endpoints)
- SFTP (4 endpoints)
- Company/Users (5 endpoints)
- Operations (3 endpoints)

---

**Last Updated:** 2026-01-03  
**Next Review:** After Phase 5 completion  
**Current Phase:** Phase 5 - WordPress Domain Management

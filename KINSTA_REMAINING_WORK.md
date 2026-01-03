# Kinsta Provider Remaining Work

**Date:** 2026-01-02  
**Status:** Phase 0 Complete for both providers  
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

### What Remains for Kinsta Provider

The Kinsta provider repository currently has:
- ✅ `kinsta_wordpress_site` - Implemented, needs refinement
- ✅ `kinsta_wordpress_environment` - Implemented, needs documentation
- ⏳ `kinsta_database` - **Deprecated immediately** (NO users), remove after migration window
- ❌ WordPress domains, backups, SFTP, tools - Not yet implemented
- ❌ Data sources - Not yet implemented
- ❌ Comprehensive specs and documentation

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

#### 2. kinsta_wordpress_site ⚠️ NEEDS REFINEMENT
**Status:** Implemented, functional, incomplete  
**Files:**
- `internal/provider/wordpress_site_resource.go`
- `internal/provider/wordpress_site_resource_test.go`
- `internal/provider/wordpress_site_resource_unit_test.go`

**Current Implementation:**
- ✅ POST /sites (async with operation_id)
- ✅ GET /sites/{id}
- ✅ DELETE /sites/{id} (async)
- ✅ Basic schema fields
- ✅ Unit tests
- ✅ Acceptance tests

**Missing Features:**
- Missing schema fields: `is_multisite`, `is_subdomain_multisite`, `woocommerce`, `wordpressseo`
- Missing computed output: `site_id`
- Limited install_mode support (needs "plain" and "clone")
- Import strategy needs documentation
- Operations polling integration may need review per `specs/02-operations-polling-contract.md`

**Priority:** P0 (Core WordPress functionality)

#### 3. kinsta_wordpress_environment ⚠️ NEEDS DOCUMENTATION
**Status:** Implemented and working  
**Files:**
- `internal/provider/wordpress_environment_resource.go`
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

**Missing:**
- Comprehensive documentation of patterns
- Acceptance tests
- Example usage
- Import documentation

**Priority:** P0 (Core WordPress functionality)

---

## Remaining Work Breakdown

### Phase 1: Cleanup & Documentation (Week 1-2)
**Goal:** Stabilize existing resources and remove deprecated code

#### 1.1 Deprecate kinsta_database
- [ ] Add deprecation warning to resource
- [ ] Update docs/resources/database.md with migration guide
- [ ] Point users to sevalla_database
- [ ] Add removal timeline (e.g., v2.0.0)

**Estimated Effort:** 2-4 hours

#### 1.2 Refine kinsta_wordpress_site
- [ ] Create `specs/20-kinsta-wordpress-site-resource.md`
- [ ] Add missing schema fields:
  - [ ] `is_multisite` (bool)
  - [ ] `is_subdomain_multisite` (bool)
  - [ ] `woocommerce` (bool)
  - [ ] `wordpressseo` (bool)
- [ ] Add computed `site_id` output
- [ ] Expand install_mode support ("plain", "clone")
- [ ] Add missing field to client structs
- [ ] Update unit tests
- [ ] Update acceptance tests
- [ ] Review operations polling implementation against `specs/02-operations-polling-contract.md`
- [ ] Update documentation and examples

**Estimated Effort:** 1-2 days

#### 1.3 Document kinsta_wordpress_environment
- [ ] Create `specs/21-kinsta-wordpress-environment-resource.md`
- [ ] Document environment ID discovery pattern
- [ ] Document eventual consistency handling
- [ ] Document write-only fields pattern
- [ ] Document import format
- [ ] Create acceptance tests
- [ ] Create comprehensive docs/resources/wordpress_environment.md
- [ ] Create examples/wordpress_environment/

**Estimated Effort:** 1 day

### Phase 2: WordPress Domain Management (Week 3-4)
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

### Phase 3: WordPress Operational Resources (Week 5-6)
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

### Phase 4: Data Sources (Week 7-8)
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

### Phase 5: WordPress Tools (Week 9-10)
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

### Phase 6: Testing & Quality (Ongoing)
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

### Priority 0 (Already Exists)
- [x] `specs/00-adr-provider-split.md` - Provider split decision
- [x] `specs/02-operations-polling-contract.md` - Async operations spec

### Priority 1 (Immediate - Week 1-2)
- [ ] `specs/20-kinsta-wordpress-site-resource.md` - Site resource refinement
- [ ] `specs/21-kinsta-wordpress-environment-resource.md` - Environment documentation

### Priority 2 (Near-term - Week 3-4)
- [ ] `specs/22-kinsta-wordpress-domain-resource.md` - Domain management

### Priority 3 (Medium-term - Week 5-6)
- [ ] `specs/23-kinsta-wordpress-backup-resource.md` - Backup management
- [ ] `specs/24-kinsta-wordpress-sftp-resource.md` - SFTP management

### Priority 4 (Longer-term - Week 7+)
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
- **Phase 1:** 1 week (cleanup & docs)
- **Phase 2:** 1 week (domains)
- **Phase 3:** 1.5 weeks (backups + SFTP)
- **Phase 4:** 1 week (data sources)
- **Phase 5:** 1 week (tools)
- **Phase 6:** Ongoing (testing/quality)

**Total:** 5.5 weeks

### Realistic (Part-time or with interruptions)
- **Phase 1:** 2 weeks
- **Phase 2:** 2 weeks
- **Phase 3:** 2 weeks
- **Phase 4:** 1.5 weeks
- **Phase 5:** 1.5 weeks
- **Phase 6:** Ongoing

**Total:** 9 weeks

---

## Next Steps (Immediate)

### Step 1: Deprecate kinsta_database (2-4 hours)
```bash
# 1. Add deprecation warning to resource
# Edit: internal/provider/database_resource.go

# 2. Update documentation
# Edit: docs/resources/database.md
# Add migration guide to sevalla_database

# 3. Update README with deprecation notice

# 4. Test deprecation warning displays correctly
go test ./internal/provider -run TestDatabaseResource
```

### Step 2: Refine kinsta_wordpress_site (1-2 days)
```bash
# 1. Create spec file
# Create: specs/20-kinsta-wordpress-site-resource.md

# 2. Add missing schema fields
# Edit: internal/provider/wordpress_site_resource.go

# 3. Add client struct fields
# Edit: internal/client/wordpress.go

# 4. Update tests
# Edit: internal/provider/wordpress_site_resource_unit_test.go
# Edit: internal/provider/wordpress_site_resource_test.go

# 5. Update documentation
# Edit: docs/resources/wordpress_site.md
# Edit: examples/wordpress_site/
```

### Step 3: Document kinsta_wordpress_environment (1 day)
```bash
# 1. Create spec file
# Create: specs/21-kinsta-wordpress-environment-resource.md

# 2. Create acceptance tests
# Create: internal/provider/wordpress_environment_resource_test.go

# 3. Create documentation
# Create: docs/resources/wordpress_environment.md
# Create: examples/wordpress_environment/

# 4. Test everything
TF_ACC=1 go test ./internal/provider -run TestAccWordPressEnvironment -v
```

---

## Success Criteria

### Phase 1 Complete When:
- [x] kinsta_database has deprecation warning
- [x] kinsta_wordpress_site has all schema fields from API
- [x] kinsta_wordpress_environment has comprehensive docs
- [x] All existing tests pass
- [x] Documentation updated

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

**Last Updated:** 2026-01-02  
**Next Review:** After Phase 1 completion

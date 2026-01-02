# Terraform Provider Split Analysis - Summary

**Date:** 2026-01-01  
**Analysis Completed With:** Sevalla API Spec v1.80.0

---

## Documents Overview

Three analysis documents have been created:

1. **`PROVIDER_SPLIT_ANALYSIS.md`** (46KB) - Complete baseline analysis before Sevalla spec
2. **`SEVALLA_SPEC_FINDINGS.md`** (13KB) - Sevalla API analysis and database migration guide  
3. **`ANALYSIS_SUMMARY.md`** (this file) - Executive summary and action plan

---

## Key Findings

### ✅ Database Migration is Straightforward

The Sevalla database API is **identical** to the deprecated MyKinsta endpoint:
- Same request/response schemas
- Same field names and types
- Same synchronous operation (200 immediate ID)
- Same update semantics (only resource_type and display_name)

**Key changes:**
- Base URL: `api.kinsta.com` → `api.sevalla.com`
- Synchronous operations: Databases return 200 immediately (no polling)
- Evidence: `sevalla.openapi.json#/paths/~1databases/post/responses=["200","401","404","500"]`

**Additional benefits in Sevalla:**
- More computed fields (connection strings, internal hostname, credentials)
- Better for application-to-database internal connections
- No async polling required (faster resource operations)

### ⚠️ Critical Bugs in Current Implementation

**kinsta_database resource has 5 critical bugs:**

1. **Missing ForceNew** - Can't change location, type, version, db_name (immutable) but Terraform doesn't enforce
2. **Random passwords** - Generates db_password/db_user randomly instead of requiring user input
3. **No 404 handling** - Won't detect when database deleted outside Terraform
4. **Field name mismatches** - Uses size/db_type/region instead of API's resource_type/type/location
5. **Missing computed fields** - Doesn't expose status, created_at, limits, cluster info

**Impact:** Resource works but is brittle, not production-ready

### ✅ Provider Split Boundary Clear

**terraform-provider-kinsta:**
- WordPress sites, environments, domains, tools, backups
- 56 WordPress-related endpoints
- Base URL: https://api.kinsta.com/v2

**terraform-provider-sevalla:**
- **Scope:** Applications, databases, static sites, pipelines, and related deployment/action endpoints
- **Excludes:** All `/sites/*` endpoints even if present in Sevalla OpenAPI spec (overlap with Kinsta provider)
- **Base URL:** https://api.sevalla.com/v2

**Evidence:** Sevalla spec contains WordPress endpoints but provider MUST NOT implement them to maintain clean separation.

**Note:** Both providers use same Bearer token authentication

---

## Migration Impact

### For Database Users

**Breaking Changes Required:**
```hcl
# OLD (kinsta_database)
resource "kinsta_database" "main" {
  name         = "mydb"        # auto-generates password
  display_name = "My Database"
  region       = "us-central1" 
  db_type      = "postgresql"
  version      = "15"
  size         = "db1"
}

# NEW (sevalla_database)
resource "sevalla_database" "main" {
  db_name       = "mydb"
  display_name  = "My Database"
  location      = "us-central1"  # renamed from region
  type          = "postgresql"   # renamed from db_type
  version       = "15"
  resource_type = "db1"          # renamed from size
  
  # NOW REQUIRED (security best practice)
  db_password = var.db_password
  db_user     = "myuser"
}

# NEW OUTPUTS AVAILABLE
output "connection_string" {
  value     = sevalla_database.main.external_connection_string
  sensitive = true
}
```

**State Migration Command:**
```bash
# CRITICAL: Terraform cannot automatically migrate state between providers
# Manual steps required:

# 1. Add sevalla provider to configuration
terraform {
  required_providers {
    sevalla = {
      source = "REPLACE_ME/sevalla"  # Update when registry name decided
      version = "~> 1.0"
    }
  }
}

# 2. Import existing database to sevalla provider
terraform import sevalla_database.main <existing-database-id>

# 3. Remove from kinsta provider state
terraform state rm kinsta_database.main

# 4. Update configuration with new field names (region→location, etc.)

# 5. Verify no changes required
terraform plan  # Should show "No changes"
```

### For Sevalla Provider Users

**Critical exclusion:**
Sevalla provider MUST NOT implement WordPress sites (`/sites/*` endpoints).

**Evidence:** WordPress endpoints exist in Sevalla spec (`sevalla.openapi.json`) but belong exclusively to Kinsta provider to avoid overlap and confusion.

**What to use instead:**
- WordPress sites → `terraform-provider-kinsta`
- Applications/databases/static-sites → `terraform-provider-sevalla`

**Rationale:** Clear separation of concerns prevents users from accidentally managing WordPress via wrong provider.

### For WordPress Users

**No changes required** - continue using kinsta provider

---

## Ordered Action Plan

### Phase 0: Foundation Fixes (1-2 weeks)

**Priority: P0 - Must fix before split**

1. **Deprecate kinsta_database immediately** (1 day)
   - Add deprecation warning to kinsta_database resource
   - Update documentation pointing to Sevalla provider
   - Publish migration guide
   - **NO users exist:** Skip all bugfix work; deprecate and remove after migration window

**Evidence:** `swagger.json#/paths/~1databases/post/deprecated=true`

**Rationale:** The kinsta_database resource uses deprecated MyKinsta endpoints. Since there are no users, do not waste effort on fixes. Users creating new databases should use terraform-provider-sevalla from the start.

3. **Centralize error handling** (2 days)
   - Parse JSON error responses
   - Add structured error types
   - Add 404 detection helpers
   - Update all resources

4. **Improve polling** (1 day)
   - Add exponential backoff
   - Make timeout configurable
   - Add progress logging

### Phase 1: Sevalla Provider Bootstrap (1 week)

**Priority: P0 - Foundation for new provider**

5. **Create sevalla provider repo** (1 day)
   - Fork/copy provider scaffold
   - Configure Sevalla base URL
   - Set up authentication
   - Configure CI/CD

6. **Implement sevalla_database** (3 days)
   - Port fixed kinsta_database code
   - Change base URL
   - Add new computed fields (connection strings, etc.)
   - Update schema with correct field names
   - Write unit tests
   - Write acceptance tests
   - Write documentation + examples

7. **Test database migration** (1 day)
   - Test import existing databases
   - Test state migration
   - Validate no unintended diffs
   - Document migration process

### Phase 2: Core Sevalla Resources (2-3 weeks)

**Priority: P0 - Essential resources**

8. **Analyze application schema** (1 day)
   - Review full MKApplicationSchema
   - Document create/update/delete patterns
   - Identify ForceNew fields
   - Plan deployment resource

9. **Implement sevalla_application** (4-5 days)
   - Full CRUD implementation
   - Handle deployment triggers
   - Unit + acceptance tests
   - Documentation + examples

10. **Implement sevalla_static_site** (3-4 days)
    - Similar to application pattern
    - Deployment management
    - Tests + documentation

11. **Implement data sources** (2 days)
    - sevalla_databases (list)
    - sevalla_applications (list)
    - sevalla_static_sites (list)

### Phase 3: Kinsta Provider Completion (2-3 weeks)

**Priority: P1 - Complete WordPress coverage**

12. **Core WordPress resources** (1 week)
    - kinsta_wordpress_site_domain
    - kinsta_wordpress_backup
    - kinsta_wordpress_sftp_user

13. **WordPress data sources** (3 days)
    - kinsta_wordpress_sites
    - kinsta_company_regions
    - kinsta_wordpress_logs

14. **Enhanced testing** (3 days)
    - Add acceptance tests for environment
    - Import tests for all resources
    - Integration tests

### Phase 4: Advanced Features (P2)

**Priority: P2 - Nice to have**

15. **Kinsta advanced resources**
    - Tool operations (cache clear, PHP restart)
    - Plugin/theme management
    - Redirect rules
    - CDN/edge cache management

16. **Sevalla advanced resources**
    - Application processes
    - Application metrics
    - Pipeline preview apps
    - Internal connections

---

## Timeline Estimate

**Phase 0 (Foundation):** 2 weeks  
**Phase 1 (Sevalla Bootstrap):** 1 week  
**Phase 2 (Core Sevalla):** 3 weeks  
**Phase 3 (Kinsta Complete):** 3 weeks  

**Total for P0/P1:** ~9 weeks (2.25 months)

**Phase 4 (Advanced):** 4-6 weeks (as needed)

---

## Risk Assessment

### High Risk ⚠️
- **Database migration** - Breaking changes for existing users
  - Mitigation: Deprecation period, migration guide, state migration tooling
  
### Medium Risk ⚠️
- **Application resource** - No create endpoint in spec (UI/CLI only?)
  - Mitigation: Need clarification from API team
  - Alternative: Read-only data source until create endpoint available

### Low Risk ✅
- **WordPress resources** - Well understood, current implementation works
- **Sevalla database** - Identical to MyKinsta, straightforward migration
- **Authentication** - Same token works for both providers

---

## Success Criteria

### Phase 0 Complete
- ✅ kinsta_database has no critical bugs
- ✅ All resources handle 404 correctly
- ✅ Errors are structured and user-friendly
- ✅ Deprecation warnings in place

### Phase 1 Complete  
- ✅ sevalla_database resource works identical to fixed kinsta_database
- ✅ Migration guide tested and documented
- ✅ Users can import existing databases

### Phase 2 Complete
- ✅ sevalla_application fully functional (or documented limitation)
- ✅ sevalla_static_site fully functional
- ✅ Core data sources available

### Phase 3 Complete
- ✅ kinsta provider covers essential WordPress operations
- ✅ Domain management working
- ✅ Backup management working

---

## Recommended Immediate Actions

### This Week

1. **Review findings** with team
2. **Validate application create endpoint** - Confirm with API team if POST /applications exists or planned
3. **Create GitHub issues** for Phase 0 bugs
4. **Set up project board** with phases
5. **Draft user communication** for database deprecation

### Next Week

1. **Begin Phase 0 fixes** - Database resource bugs
2. **Set up sevalla provider repo**
3. **Start writing detailed spec files** for Phase 1 resources

---

## Open Questions for API Team

### High Priority
1. **Application Create:** POST /applications does NOT exist in Sevalla spec (`sevalla.openapi.json#/paths/~1applications` has only GET). Is application creation via API planned? Blocks `sevalla_application` resource (only data source possible).
2. **Application Update:** Which fields are updatable via PUT /applications/{id}?
3. **Deployment Lifecycle:** Are deployments immediate (200) or async (202 + operation_id)?
4. **Operations data contract (NEW):** MyKinsta polling returns `operation.data={}` (opaque per `swagger.json#/components/schemas/OperationResponse`). Can we rely on `data.idSite`/`data.idEnv` keys, or should all resources use lookup-after-poll strategy? Current implementation attempts extraction but falls back.

### Medium Priority
4. **Static Site Create:** Same questions as application
5. **Internal Connections:** How are application-database connections managed?
6. **Metrics Access:** Any rate limiting on metrics endpoints?

---

## Resource Coverage Matrix

### Sevalla Provider

| Resource | Priority | Status | Endpoints | Blockers |
|----------|----------|--------|-----------|----------|
| sevalla_database | P0 | Ready | Analyzed | None |
| sevalla_application | P2 (blocked) | No POST endpoint | Analyzed | `sevalla.openapi.json#/paths/~1applications` has only GET |
| sevalla_applications (data source) | P0 | Ready | Analyzed | Read-only list |
| sevalla_static_site | P1 | Ready to spec | Analyzed | Deprecated in MyKinsta (`swagger.json#/paths/~1static-sites/get/deprecated=true`), active in Sevalla |
| sevalla_application_deployment | P1 | Ready | Analyzed | None |
| sevalla_static_site_deployment | P1 | Ready | Analyzed | None |
| Data sources | P1 | Ready | Analyzed | None |

### Kinsta Provider

| Resource | Priority | Status | Endpoints | Blockers |
|----------|----------|--------|-----------|----------|
| kinsta_wordpress_site | P0 | Implemented | Working | Needs enhancements |
| kinsta_wordpress_environment | P0 | Implemented | Working | Needs tests |
| kinsta_database | P0 | Deprecated | Working | Has bugs |
| kinsta_wordpress_site_domain | P0 | Not started | Analyzed | None |
| kinsta_wordpress_backup | P1 | Not started | Analyzed | None |
| kinsta_wordpress_sftp_user | P1 | Not started | Analyzed | None |
| Data sources | P1 | Not started | Analyzed | None |

---

## Conclusion

The provider split is **well-defined and achievable**. The Sevalla spec confirms that migration is straightforward, especially for databases. The main risks are around application resource patterns and user migration communication.

**Recommended approach:**
1. Fix current bugs (Phase 0)
2. Launch Sevalla with database resource first (Phase 1)
3. Add applications/static sites after clarifying patterns (Phase 2)
4. Complete Kinsta WordPress coverage in parallel (Phase 3)

**Timeline:** 9 weeks for P0/P1 functionality, with Phase 0 being critical path for any Sevalla work.

---

For detailed analysis:
- Technical depth: See `PROVIDER_SPLIT_ANALYSIS.md`
- Sevalla specifics: See `SEVALLA_SPEC_FINDINGS.md`
- This summary: `ANALYSIS_SUMMARY.md`

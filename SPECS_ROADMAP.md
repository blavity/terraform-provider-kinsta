# Terraform Provider Specs Roadmap

**Purpose:** Specification files to be created in `./specs/` directory  
**Format:** Markdown with detailed resource/data source specifications

---

## Recommended Spec File Structure

Each spec file should contain:
1. **Scope** - Which endpoints/resources covered
2. **Schema Mapping** - Terraform schema → API fields with types, validation
3. **Lifecycle Rules** - Create/Read/Update/Delete behavior, ForceNew fields
4. **Async/Sync Patterns** - Operation polling requirements
5. **Validation Rules** - Enums, constraints, conditionals
6. **Error Handling** - Expected error codes, handling approach
7. **Test Plan** - Unit test cases, acceptance test scenarios
8. **Documentation Plan** - Argument reference, attributes, examples
9. **Import Support** - Import ID format, lookup strategy

---

## Spec Files to Create

### Foundation (Phase 0)

#### `specs/00-provider-split-strategy.md`
**Status:** ✅ Content in ANALYSIS_SUMMARY.md, needs extraction  
**Contents:**
- Split rationale (Kinsta vs Sevalla APIs)
- Provider naming and registry publishing
- Shared authentication approach
- User migration timeline
- Cross-provider usage patterns

#### `specs/01-error-handling-patterns.md`
**Status:** ⏳ Not started  
**Contents:**
- APIError type structure
- Status code interpretation (400, 401, 404, 429, 500)
- Retry strategy (exponential backoff)
- Rate limit detection and handling
- User-friendly diagnostic messages
- Centralized error parsing implementation

#### `specs/02-operations-polling-contract.md`
**Status:** ⏳ Not started  
**Contents:**
- Operation lifecycle (202 → 404 grace → 202 → 200/500)
- Polling intervals (2s, 4s, 8s, 15s, 30s cap)
- Timeout configuration (default 10 minutes)
- Resource ID extraction patterns
- Context cancellation handling
- Progress logging approach

---

### Sevalla Provider Specs (Phase 1-2)

#### `specs/10-sevalla-database-resource.md`
**Status:** ✅ Core analysis in SEVALLA_SPEC_FINDINGS.md, needs spec format  
**Priority:** P0  
**API Endpoints:**
- POST /databases (create)
- GET /databases/{id} (read)
- PUT /databases/{id} (update)
- DELETE /databases/{id} (delete)
- GET /databases (list for import)

**Key Decisions:**
- ✅ Synchronous operations (200 immediate)
- ✅ ForceNew: location, type, version, db_name
- ✅ Updatable: resource_type, display_name only
- ✅ Required inputs: db_password, db_user (breaking change from kinsta)
- ✅ Computed outputs: connection strings, hostnames, credentials

**Contents Needed:**
- Complete schema with all 20+ fields
- Validation rules (type enum, resource_type enum, db_user conditional)
- Sensitive field handling
- Import by ID or name
- Migration guide from kinsta_database
- Test cases (create PostgreSQL, create Redis, update size, 404 handling)

#### `specs/11-sevalla-application-resource.md`
**Status:** ⏳ Partial analysis, needs completion  
**Priority:** P0  
**API Endpoints:**
- GET /applications (list)
- GET /applications/{id} (read)
- PUT /applications/{id} (update)
- DELETE /applications/{id} (delete)
- POST /applications - **NEEDS CLARIFICATION** (not in spec?)

**Blockers:**
- ❓ How are applications created via API?
- ❓ Which fields are updatable?
- ❓ Which fields are ForceNew?
- ❓ Deployment trigger mechanism?

**Contents Needed:**
- Complete schema analysis
- Relationship with deployments
- Process management approach
- Environment variables handling
- Git repository configuration
- Build/runtime settings

#### `specs/12-sevalla-application-deployment-resource.md`
**Status:** ⏳ Not started  
**Priority:** P1  
**API Endpoints:**
- POST /applications/deployments
- GET /applications/deployments/{deployment_id}

**Key Decisions:**
- Design: Separate resource or application trigger?
- Recommendation: Separate resource for explicit control

**Contents Needed:**
- Creation triggers (manual, on-push, on-PR)
- Status polling
- Relationship with application resource
- Deployment rollback strategy

#### `specs/13-sevalla-static-site-resource.md`
**Status:** ⏳ Not started  
**Priority:** P0  
**Contents:** Similar structure to application resource

#### `specs/14-sevalla-data-sources.md`
**Status:** ⏳ Not started  
**Priority:** P1  
**Data Sources:**
- sevalla_databases (list with filtering)
- sevalla_applications (list with filtering)
- sevalla_static_sites (list with filtering)
- sevalla_pipelines (list)

**Contents Needed:**
- Filter/search parameters
- Pagination handling
- Output schema
- Use cases vs resources

---

### Kinsta Provider Specs (Phase 3)

#### `specs/20-kinsta-wordpress-site-resource.md`
**Status:** ⏳ Needs spec format (implementation exists)  
**Priority:** P0 (refine existing)  
**API Endpoints:**
- POST /sites (create)
- GET /sites/{id} (read)
- DELETE /sites/{id} (delete)
- GET /sites (list for import)

**Key Decisions:**
- ✅ Async operations (202 + operation_id)
- ✅ All fields ForceNew (no PUT endpoint)
- ⏳ Add missing fields: is_multisite, is_subdomain_multisite, woocommerce, wordpressseo
- ⏳ Add site_id computed output
- ⏳ Support install_mode="plain" and "clone"

**Contents Needed:**
- Complete schema with all fields
- Operation polling integration
- 404 handling (deleted sites)
- Import strategies (by ID or display_name)
- Test cases covering all install modes

#### `specs/21-kinsta-wordpress-environment-resource.md`
**Status:** ✅ Implementation exists and works, needs documentation  
**Priority:** P0 (document existing)  
**API Endpoints:**
- POST /sites/{site_id}/environments
- DELETE /sites/environments/{env_id}
- GET /sites/{site_id} (read via environments list)

**Key Patterns to Document:**
- Environment ID discovery (before/after comparison)
- Eventual consistency handling (display_name retry)
- Write-only fields (DiffSuppressFunc pattern)
- Import format (site_id:env_id)
- Read via parent resource (no direct GET endpoint)

#### `specs/22-kinsta-wordpress-site-domain-resource.md`
**Status:** ⏳ Not started  
**Priority:** P0  
**API Endpoints:**
- POST /sites/environments/{env_id}/domains
- DELETE /sites/environments/{env_id}/domains
- GET /sites/{site_id} (read via environments[].domains[])
- PUT /sites/environments/{env_id}/change-primary-domain

**Contents Needed:**
- Domain creation/deletion
- Primary domain management
- Domain verification status
- SSL certificate handling
- Multiple domains per environment

#### `specs/23-kinsta-wordpress-backup-resource.md`
**Status:** ⏳ Not started  
**Priority:** P1  
**API Endpoints:**
- POST /sites/environments/{env_id}/manual-backups
- POST /sites/environments/{target_env_id}/backups/restore
- DELETE /sites/environments/backups/{backup_id}
- GET /sites/environments/{env_id}/backups (list)
- GET /sites/environments/{env_id}/downloadable-backups

**Contents Needed:**
- Manual backup creation
- Backup restoration (different environment)
- Backup deletion
- Downloadable backup data source
- Async operation handling

#### `specs/24-kinsta-wordpress-sftp-resource.md`
**Status:** ⏳ Not started  
**Priority:** P1  
**API Endpoints:**
- POST /sites/environments/{env_id}/additional-sftp-accounts
- DELETE /sites/environments/additional-sftp-accounts/{sftp_account_id}
- GET /sites/environments/{env_id}/additional-sftp-accounts
- PUT /sites/environments/{env_id}/additional-sftp-accounts/toggle-status

**Contents Needed:**
- SFTP user creation with password
- Enable/disable SFTP access
- List SFTP users
- Permission management

#### `specs/25-kinsta-wordpress-tools.md`
**Status:** ⏳ Not started  
**Priority:** P2  
**Resources:**
- kinsta_wordpress_cache_clear (POST /sites/tools/clear-cache)
- kinsta_wordpress_php_restart (POST /sites/tools/restart-php)
- kinsta_wordpress_php_version (PUT /sites/tools/modify-php-version)
- kinsta_wordpress_denied_ips (GET/PUT /sites/tools/denied-ips)

**Design Decisions:**
- Trigger resources vs data sources vs functions?
- Recommendation: Resources with terraform_data triggers

#### `specs/26-kinsta-data-sources.md`
**Status:** ⏳ Not started  
**Priority:** P1  
**Data Sources:**
- kinsta_wordpress_sites (list)
- kinsta_company_regions (available regions)
- kinsta_company_users (list)
- kinsta_wordpress_logs (environment logs)
- kinsta_wordpress_backups (list)
- kinsta_company_activity_logs
- kinsta_wordpress_analytics_* (metrics)

---

### Testing & Quality Specs

#### `specs/30-testing-standards.md`
**Status:** ⏳ Not started  
**Priority:** P1  
**Contents:**
- Unit test patterns (mock client, schema validation)
- Acceptance test patterns (TF_ACC, resource lifecycle)
- Test fixtures and helpers
- Random name generation
- Resource sweeping (cleanup)
- CI/CD integration

#### `specs/31-documentation-standards.md`
**Status:** ⏳ Not started  
**Priority:** P1  
**Contents:**
- Resource documentation template
- Argument reference format
- Attributes reference format
- Example usage patterns
- Import documentation
- Timeout configuration
- Common errors and solutions

---

## Spec File Priority Matrix

### Phase 0 (Foundation) - Start Immediately
- [ ] `specs/00-provider-split-strategy.md` (extract from analysis)
- [ ] `specs/01-error-handling-patterns.md`
- [ ] `specs/02-operations-polling-contract.md`

### Phase 1 (Sevalla Bootstrap) - Week 2
- [ ] `specs/10-sevalla-database-resource.md` (high detail)

### Phase 2 (Sevalla Core) - Week 3-5
- [ ] `specs/11-sevalla-application-resource.md` (after API clarification)
- [ ] `specs/13-sevalla-static-site-resource.md`
- [ ] `specs/12-sevalla-application-deployment-resource.md`
- [ ] `specs/14-sevalla-data-sources.md`

### Phase 3 (Kinsta Complete) - Week 6-8
- [ ] `specs/20-kinsta-wordpress-site-resource.md` (refine)
- [ ] `specs/21-kinsta-wordpress-environment-resource.md` (document)
- [ ] `specs/22-kinsta-wordpress-site-domain-resource.md`
- [ ] `specs/23-kinsta-wordpress-backup-resource.md`
- [ ] `specs/24-kinsta-wordpress-sftp-resource.md`
- [ ] `specs/26-kinsta-data-sources.md`

### Phase 4 (Quality) - Ongoing
- [ ] `specs/30-testing-standards.md`
- [ ] `specs/31-documentation-standards.md`
- [ ] `specs/25-kinsta-wordpress-tools.md` (P2)

---

## Spec Writing Guidelines

Each spec file should:

1. **Start with context**
   - API version
   - Affected resources
   - Dependencies

2. **Define the contract**
   - Exact API endpoints with methods
   - Request/response schemas
   - Error codes

3. **Terraform schema**
   - Field-by-field mapping
   - Type conversions
   - Default values
   - Validation rules

4. **Lifecycle behavior**
   - Create: What happens? Async? Validation?
   - Read: 404 handling? Nested reads?
   - Update: Which fields? ForceNew rules?
   - Delete: Async? Cascade rules?
   - Import: ID format? Lookup strategy?

5. **Test plan**
   - Unit test list (schema, CRUD, errors)
   - Acceptance test scenarios
   - Edge cases

6. **Examples**
   - Basic usage
   - Advanced usage
   - Common patterns
   - Import examples

7. **Migration guide** (if applicable)
   - From what resource?
   - State migration steps
   - Breaking changes
   - Codemods/scripts

---

## Next Steps

1. **This Week:**
   - Create `./specs/` directory
   - Write foundation specs (00-02)
   - Extract database spec detail

2. **Next Week:**
   - Complete sevalla_database spec
   - Begin implementation based on spec
   - Validate spec accuracy during implementation

3. **Ongoing:**
   - Write spec before implementing resource
   - Update spec if API behavior differs
   - Link from code comments to specs
   - Version specs with API version

---

## Spec Maintenance

**When to update specs:**
- API version changes
- New endpoints discovered
- Field behavior clarified
- Implementation finds edge cases
- User reports gaps

**Spec review process:**
- Technical review before implementation
- Implementation validation (does code match spec?)
- Post-implementation updates (actual vs expected)
- Quarterly spec refresh

---

For analysis details:
- Technical analysis: `PROVIDER_SPLIT_ANALYSIS.md`
- Sevalla specifics: `SEVALLA_SPEC_FINDINGS.md`
- Summary: `ANALYSIS_SUMMARY.md`
- This roadmap: `SPECS_ROADMAP.md`

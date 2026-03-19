# Terraform Provider Split Analysis: Kinsta → Kinsta + Sevalla

**Date:** 2026-01-01  
**Scope:** Split terraform-provider-kinsta into two providers based on API deprecation boundaries

---

## Executive Summary

### Current State
- **Implementation:** 3 resources (database, wordpress_site, wordpress_environment)
- **Framework:** terraform-plugin-sdk/v2
- **Test Coverage:** Unit tests (100% for implemented), Acceptance tests (partial)
- **API Coverage:** ~10% of MyKinsta API, 0% of Sevalla API

### Split Boundary
1. **terraform-provider-kinsta (MyKinsta API):** WordPress sites, environments, domains, tools, company resources
2. **terraform-provider-sevalla (Sevalla API):** Applications, databases, static sites, pipelines

### Key Finding
**⚠️ CRITICAL:** Database resource (kinsta_database) is implemented but uses DEPRECATED MyKinsta endpoint. Must migrate to Sevalla provider.

---

## 1. Implementation Inventory

### 1.1 Implemented Resources

#### kinsta_database (⚠️ DEPRECATED API)
- **Status:** Fully implemented but using deprecated /databases endpoint
- **Location:** `internal/provider/database_resource.go`
- **Client Methods:** CreateDatabase, GetDatabase, UpdateDatabase, DeleteDatabase
- **Tests:**
  - Unit: `database_resource_unit_test.go` (280 lines)
  - Acceptance: `database_resource_test.go` (58 lines, basic)
- **Lifecycle:** CRUD implemented
- **Issues:**
  - Using deprecated MyKinsta API (sunset: 2026-01-31)
  - No ForceNew on immutable fields (region, db_type, version, name)
  - Update logic allows changes to fields that should be ForceNew
  - Generates random passwords/usernames (not idempotent, not configurable)
  - Missing: created_at, memory_limit, cpu_limit, storage_size, cluster info
  - No 404 handling in Read (doesn't set d.SetId(""))

#### kinsta_wordpress_site
- **Status:** Partially implemented
- **Location:** `internal/provider/wordpress_site_resource.go`
- **Client Methods:** CreateWordPressSite, GetWordPressSite, GetWordPressSites, DeleteWordPressSite
- **Tests:**
  - Unit: `wordpress_site_resource_unit_test.go` (487 lines, comprehensive)
  - Acceptance: `wordpress_site_resource_test.go` (150 lines, 3 test cases)
- **Lifecycle:** CRD implemented (no Update)
- **Issues:**
  - All fields are ForceNew (correct - no PUT endpoint exists)
  - Missing fields from API: is_multisite, is_subdomain_multisite, woocommerce, wordpressseo
  - Missing computed: site_id (only has environment_id)
  - No import support
  - Operation polling works correctly
  - Missing: install_mode="plain", install_mode="clone" variants

#### kinsta_wordpress_environment
- **Status:** Well implemented with workarounds
- **Location:** `internal/provider/wordpress_environment_resource.go`
- **Client Methods:** CreateWordPressEnvironment, DeleteWordPressEnvironment
- **Tests:**
  - Unit: `wordpress_environment_resource_unit_test.go` (443 lines, comprehensive)
  - Acceptance: None
- **Lifecycle:** CR_D implemented (Update is no-op, all fields ForceNew)
- **Features:**
  - Sophisticated eventual consistency handling (display_name conflicts after delete)
  - DiffSuppressFunc for write-only fields
  - Import support (site_id:env_id format)
  - Environment ID discovery via before/after comparison
- **Issues:**
  - PollOperation doesn't return idEnv, requires workaround
  - No environment GET endpoint, must read via site's environments list
  - Write-only fields preserved in state (correct approach)

### 1.2 Client Implementation (internal/client/client.go)

#### HTTP Client
```go
func (c *Client) do(ctx context.Context, method, path string, body io.Reader, v interface{}) error
```
- **Error Handling:** Basic string formatting `fmt.Errorf("API error: %s", resp.Status)`
- **Issues:**
  - No structured error response parsing
  - No retry logic
  - No rate limit handling
  - No status code differentiation (404 vs 500 vs 400)
  - Error messages not user-friendly

#### Operation Polling
```go
func (c *Client) PollOperation(ctx context.Context, operationID string) (string, error)
```
- **Strategy:** Fixed interval (5s), max attempts (60 = 5 minutes)
- **Features:**
  - Handles 404 grace period (first 25 seconds)
  - Extracts idSite or idEnv from response data
  - Returns empty string if no ID in response
- **Issues:**
  - Treats operation.Data as typed map[string]interface{}
  - No exponential backoff
  - Hardcoded timeout
  - No configurable poll interval

#### Interface
```go
type KinstaClient interface {
    CompanyID() string
    // Database methods (DEPRECATED)
    CreateDatabase(...)
    GetDatabase(...)
    UpdateDatabase(...)
    DeleteDatabase(...)
    // WordPress methods (ACTIVE)
    CreateWordPressSite(...)
    GetWordPressSite(...)
    GetWordPressSites(...)
    DeleteWordPressSite(...)
    CreateWordPressEnvironment(...)
    DeleteWordPressEnvironment(...)
    // Polling
    PollOperation(...)
}
```

### 1.3 Test Infrastructure

#### Provider Test Setup (`provider_test.go`)
```go
var testAccProviderFactories map[string]func() (*schema.Provider, error)
```
- Minimal - only provider factory
- No shared precheck utilities
- No random name generators
- No cleanup helpers

#### Unit Test Pattern
- Uses `schema.TestResourceDataRaw()`
- Mock client interfaces with function fields
- Testify assert/require
- Good coverage of edge cases
- Tests schema validation (ForceNew, Sensitive, etc.)

#### Acceptance Test Pattern
- Uses `TF_ACC` environment variable
- Basic checks (exists, attributes)
- Limited test cases (1-3 per resource)
- No update tests
- No import tests
- No error scenario tests

---

## 2. OpenAPI Specification Analysis

### 2.1 MyKinsta API (swagger.json)

**Stats:**
- **Version:** 1.87.0
- **Base URL:** https://api.kinsta.com/v2
- **Total Endpoints:** 85
- **Total Schemas:** 275
- **WordPress-related Endpoints:** 56
- **Deprecated Endpoints:** 29 (applications, databases, static-sites, pipelines)

**Tags (Active for Kinsta Provider):**
- WordPress Sites (7 endpoints)
- WordPress Site Environments (12 endpoints)
- WordPress Site Tools (4 endpoints)
- WordPress Site Themes & Plugins (4 endpoints)
- WordPress Site Domains (4 endpoints)
- WordPress Edge Caching (2 endpoints)
- WordPress CDN (2 endpoints)
- Backups (5 endpoints)
- Logs (1 endpoint)
- Additional SFTP Users (4 endpoints)
- Analytics (4 endpoints)
- Company Users (1 endpoint)
- API Keys (1 endpoint)
- Available Regions (1 endpoint)
- Activity Logs (1 endpoint)
- Domains (4 endpoints)
- Operations (1 endpoint)
- Authentication (1 endpoint)

**Tags (Deprecated → Sevalla):**
- Applications (9 endpoints)
- Application Deployments (2 endpoints)
- Application Processes (2 endpoints)
- Application Metrics (8 endpoints)
- Application CDN (1 endpoint)
- Application Clear Cache (1 endpoint)
- Application Edge Caching (1 endpoint)
- Internal Connections (1 endpoint)
- Databases (4 endpoints)
- Static Sites (3 endpoints)
- Static Site Deployments (2 endpoints)

### 2.2 Sevalla API

**Status:** No spec file provided at `./_spec_cache/sevalla.openapi.json`

**Known Information:**
- **Base URL:** https://api.sevalla.com/v2
- **Auth:** Bearer token (same MyKinsta API key works)
- **Documentation:** https://api-docs.sevalla.com/
- **Resources:** Applications, Databases, Static Sites, Pipelines

**Uncertainties Requiring Evidence:**
1. Complete request/response schemas for all resources
2. Async operation patterns (operation_id or immediate ID?)
3. Update semantics (PUT/PATCH availability, field immutability)
4. Error response format
5. Rate limiting headers/policies
6. Pagination patterns
7. Computed vs input field boundaries
8. Import ID formats

---

## 3. Spec vs Code Gap Analysis

### 3.1 Database Resource (kinsta_database)

| Aspect | Spec | Implementation | Gap |
|--------|------|----------------|-----|
| **API Status** | DEPRECATED (sunset 2026-01-31) | Using deprecated endpoint | ⚠️ MUST MIGRATE TO SEVALLA |
| **Create Fields** | company_id, location, resource_type, display_name, db_name, db_password, db_user, type, version | Most present | ❌ db_user optional (Redis), generates random values |
| **Resource Type** | Enum: db1-db9 | size field (string) | ⚠️ Field name mismatch |
| **Type** | Enum: postgresql, redis, mariadb, mysql | db_type field (string) | ⚠️ Field name mismatch, no validation |
| **Read Fields** | Database schema with cluster, created_at, memory_limit, cpu_limit, storage_size, resource_type_name | Only: name, display_name, region, type, version, size | ❌ Missing: created_at, limits, cluster info |
| **Update** | PUT /databases/{id} - resource_type, display_name | Implements update | ❌ BUG: Should ForceNew most fields |
| **Update Semantics** | Only resource_type and display_name updatable | Allows updating any field | ❌ BUG: Missing ForceNew on immutable fields |
| **Delete** | DELETE /databases/{id} returns 200 | Implemented | ✅ |
| **Response** | 200 sync response with immediate ID | Uses immediate ID | ✅ |
| **404 Handling** | Should clear state | Not implemented | ❌ BUG |
| **ForceNew** | region, type, version, db_name should be ForceNew | None marked ForceNew | ❌ CRITICAL BUG |

### 3.2 WordPress Site Resource (kinsta_wordpress_site)

| Aspect | Spec | Implementation | Gap |
|--------|------|----------------|-----|
| **Create Fields** | company, display_name, region, install_mode, admin_email, admin_password, admin_user, site_title, wp_language, is_multisite, is_subdomain_multisite, woocommerce, wordpressseo | Only: company (auto), display_name, region, install_mode, admin_email, admin_password, admin_user, site_title, wp_language | ❌ Missing: is_multisite, is_subdomain_multisite, woocommerce, wordpressseo |
| **install_mode** | Enum: new, plain, clone (deprecated field) | Default: "new", no plain/clone support | ⚠️ Partial - missing plain/clone variants |
| **Response** | 202 with operation_id | Polls operation correctly | ✅ |
| **Read Fields** | id, name (auto-generated), display_name, status, siteLabels, environments[] | Gets site | ❌ Missing computed: site_id output |
| **Update** | No PUT endpoint | All fields ForceNew | ✅ Correctly immutable |
| **Delete** | DELETE /sites/{site_id} returns 202 with operation_id | Polls operation | ✅ |
| **Import** | N/A | Not implemented | ❌ Missing |
| **404 Handling** | Should clear state | Not implemented | ❌ BUG |
| **Operation Polling** | /operations/{op_id} can return 404 initially | Handles 404 grace period (5 attempts * 5s) | ✅ Documented workaround |

### 3.3 WordPress Environment Resource (kinsta_wordpress_environment)

| Aspect | Spec | Implementation | Gap |
|--------|------|----------------|-----|
| **Create Fields** | display_name, site_title, is_premium, admin_email, admin_password, admin_user, wp_language | All present | ✅ |
| **Create Endpoint** | POST /sites/{site_id}/environments | Correct | ✅ |
| **Response** | 202 with operation_id (doesn't return idEnv in data) | Workaround: before/after environment list comparison | ✅ Creative solution |
| **Read Endpoint** | No GET /sites/environments/{env_id} - must use site's environments list | Reads via GetWordPressSite, filters by env_id | ✅ Correct workaround |
| **Update** | No PUT endpoint | All fields ForceNew, Update is no-op | ✅ |
| **Delete** | DELETE /sites/environments/{env_id} | Implemented | ✅ |
| **Import** | N/A | Implemented (site_id:env_id format) | ✅ |
| **404 Handling** | Environment not in site's list | Returns if foundEnv == nil, sets d.SetId("") | ✅ |
| **Eventual Consistency** | display_name reserved ~30s after delete | Exponential backoff retry (up to 64s) | ✅ Sophisticated handling |
| **Write-only Fields** | site_title, is_premium, admin_* fields not returned | DiffSuppressFunc prevents drift, preserves in state | ✅ Best practice |

### 3.4 Missing Resources (from Spec)

#### WordPress Resources (Priority for Kinsta Provider)

**P0 - Core Site Management:**
- kinsta_wordpress_site_domain (POST/DELETE /sites/environments/{env_id}/domains)
- kinsta_wordpress_site_primary_domain (PUT /sites/environments/{env_id}/change-primary-domain)

**P1 - Site Operations:**
- kinsta_wordpress_tool_clear_cache (POST /sites/tools/clear-cache) - could be resource or data source trigger
- kinsta_wordpress_tool_restart_php (POST /sites/tools/restart-php)
- kinsta_wordpress_tool_php_version (PUT /sites/tools/modify-php-version)
- kinsta_wordpress_tool_denied_ips (GET/PUT /sites/tools/denied-ips)

**P1 - Backups:**
- kinsta_wordpress_backup (POST /sites/environments/{env_id}/manual-backups)
- kinsta_wordpress_backup_restore (POST /sites/environments/{env_id}/backups/restore)
- kinsta_wordpress_backup_download (data source - GET /sites/environments/{env_id}/downloadable-backups)

**P1 - SFTP:**
- kinsta_wordpress_sftp_user (POST/DELETE /sites/environments/{env_id}/additional-sftp-accounts)
- kinsta_wordpress_sftp_toggle (PUT /sites/environments/{env_id}/additional-sftp-accounts/toggle-status)

**P2 - Advanced:**
- kinsta_wordpress_plugin (PUT /sites/environments/{env_id}/plugins, /bulk-update)
- kinsta_wordpress_theme (PUT /sites/environments/{env_id}/themes, /bulk-update)
- kinsta_wordpress_php_allocation (POST /sites/{site_id}/change-site-php-allocation)
- kinsta_wordpress_redirect_rule (POST/GET /sites/environments/{env_id}/redirect-rules)
- kinsta_wordpress_edge_cache (POST/PUT /sites/edge-caching/*)
- kinsta_wordpress_cdn (POST/PUT /sites/cdn/*)
- kinsta_wordpress_ssh_config (POST/GET /sites/environments/{env_id}/ssh/*)

#### Data Sources (Priority for Kinsta Provider)

**P0:**
- kinsta_wordpress_sites (GET /sites?company={id})
- kinsta_company_regions (GET /company/{id}/available-regions)

**P1:**
- kinsta_company_users (GET /company/{id}/users)
- kinsta_company_api_keys (GET /company/{id}/api-keys)
- kinsta_wordpress_logs (GET /sites/environments/{env_id}/logs)
- kinsta_wordpress_backups (GET /sites/environments/{env_id}/backups)

**P2:**
- kinsta_company_activity_logs (GET /company/{id}/activity-logs)
- kinsta_wordpress_analytics_* (GET /sites/environments/{env_id}/analytics/*)
- kinsta_domains (GET /domains?company={id})
- kinsta_dns_records (GET /domains/{domain_id}/dns-records)

#### Sevalla Resources (New Provider)

**P0 - Core Resources:**
- sevalla_application
- sevalla_database
- sevalla_static_site

**P1:**
- sevalla_application_deployment
- sevalla_application_process
- sevalla_static_site_deployment
- sevalla_application_internal_connection
- sevalla_pipeline_preview_app

**Data Sources:**
- sevalla_applications
- sevalla_databases
- sevalla_static_sites
- sevalla_pipelines
- sevalla_application_metrics_*

---

## 4. Correctness Bugs in Current Code

### 4.1 CRITICAL: Database Resource Immutability Violations

**Bug:** Update function allows changing immutable fields

**Location:** `internal/provider/database_resource.go:104-126`

```go
func resourceDatabaseUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    // BUG: This allows updating display_name and size
    // But spec shows only these are updatable
    // PROBLEM: region, db_type, version, name should be ForceNew but aren't
}
```

**Impact:** User can change `region`, `db_type`, `version`, `name` in config, Terraform will try to update (fails or worse: silently no-ops), state diverges

**Fix:** Add ForceNew to schema for: name, region, db_type, version

### 4.2 CRITICAL: No 404 Handling in Read Operations

**Bug:** Read functions don't detect resource deletion outside Terraform

**Locations:**
- `database_resource.go:85-102` - GetDatabase error not checked for 404
- `wordpress_site_resource.go:102-118` - GetWordPressSite error not checked for 404

**Impact:** If resource deleted outside Terraform, next apply fails instead of detecting drift

**Fix:** Parse status code from error, if 404 then `d.SetId("")` and return nil

### 4.3 BUG: PollOperation Relies on Opaque operation.Data

**Bug:** Assumes operation.Data contains idSite/idEnv with specific keys

**Location:** `internal/client/client.go:395-408`

```go
if siteID, ok := opResp.Data["idSite"].(string); ok {
    return siteID, nil
}
```

**Problem:** Spec says operation.Data is {} (opaque). Key names not guaranteed.

**Impact:** If API changes data structure, polling breaks silently

**Fix:** Document assumption, or return operation status separately from resource ID extraction

### 4.4 BUG: Random Password Generation Not Idempotent

**Bug:** Database passwords/usernames generated randomly on each plan

**Location:** `database_resource.go:69-70`

```go
DBPassword:   generateRandomString(16),
DBUser:       generateRandomString(12),
```

**Impact:**
- Not user-controllable
- Not recoverable if state lost
- Security risk (passwords not managed properly)

**Fix:** Make db_password and db_user required sensitive input fields

### 4.5 BUG: Client Error Messages Not Structured

**Bug:** Errors returned as generic strings

**Location:** `client.go:50-52`

```go
if resp.StatusCode >= 400 {
    return fmt.Errorf("API error: %s", resp.Status)
}
```

**Impact:** No structured error data for retry logic, user guidance, or specific error handling

**Fix:** Parse JSON error response body, return structured error type

---

## 5. Proposed Spec Files (./specs/)

### 5.1 Roadmap & Foundations

#### specs/00-roadmap-split-kinsta-vs-sevalla.md
**Scope:** Provider split strategy, deprecation timeline, migration guide

**Contents:**
- Split boundary rationale (MyKinsta vs Sevalla APIs)
- Timeline: Database resource migration, new provider creation
- Provider naming: `terraform-provider-kinsta` vs `terraform-provider-sevalla`
- Registry publishing plan
- User migration guide (existing kinsta_database users)
- Deprecation warnings in kinsta provider for database resource
- Cross-provider references (if any)

#### specs/10-foundations-http-errors-and-retries.md
**Scope:** Centralized error handling, retry logic, rate limiting

**Contents:**
- Error response schema parsing
  ```go
  type KinstaError struct {
      StatusCode int
      Message    string
      Status     int    // From JSON body
      Data       interface{}
  }
  ```
- Error types: 400 (validation), 401 (auth), 404 (not found), 429 (rate limit), 500 (server error)
- Retry strategy: Exponential backoff for 429, 500, 502, 503, 504
- Rate limit detection (check response headers)
- User-friendly diagnostic messages (diag.Diagnostic with detail)
- Centralized `doRequest()` helper with error parsing

#### specs/11-foundations-operations-polling.md
**Scope:** Async operation polling contract

**Contents:**
- Operation lifecycle: 202 → 404 grace (0-30s) → 202 (in-progress) → 200 (success) or 500 (failure)
- Polling strategy:
  - Initial interval: 2s
  - Exponential backoff: 2s, 4s, 8s, 15s, 30s (cap at 30s)
  - Timeout: Configurable (default 10 minutes)
  - 404 grace window: Up to 30s (6 attempts * 5s)
- Resource ID extraction:
  - Document known keys: idSite, idEnv
  - Fall back to listing resources if key not present (environment pattern)
- Context cancellation handling
- Progress logging (using terraform-plugin-log)

### 5.2 Kinsta Provider Resources

#### specs/20-kinsta-wordpress-site.md
**Scope:** kinsta_wordpress_site resource spec-driven design

**Contents:**
- **Endpoints:** POST /sites, GET /sites/{id}, DELETE /sites/{id}, GET /sites (list for import)
- **Schema Mapping:**
  - **Inputs:** company (auto from provider), display_name, region, install_mode, admin_email*, admin_password*, admin_user, site_title, wp_language, is_multisite, is_subdomain_multisite, woocommerce, wordpressseo
  - **Computed:** id, name (auto-generated by API), site_id, status, environments[] (nested block or separate resource?)
- **Lifecycle:**
  - Create: POST /sites → 202 + operation_id → poll → read
  - Read: GET /sites/{id} → handle 404 → d.SetId("")
  - Update: None (all ForceNew)
  - Delete: DELETE /sites/{id} → 202 + operation_id → poll
- **ForceNew Rules:** All fields (no PUT endpoint)
- **Async:** Yes (create, delete)
- **Validation:**
  - region: Validate against available regions (data source?)
  - install_mode: Enum [new, plain, clone] (deprecated but still accepted)
  - wp_language: String (validate against common locales?)
- **Sensitive:** admin_email, admin_password
- **Import:** GET /sites?company={id}, match by display_name or prompt user with list
- **Test Plan:**
  - Unit: Schema validation, create/read/delete flows, operation polling, 404 handling
  - Acceptance: Basic site creation, custom language, multisite mode, deletion

#### specs/21-kinsta-wordpress-environment.md
**Scope:** kinsta_wordpress_environment (refine existing)

**Contents:**
- **Endpoints:** POST /sites/{site_id}/environments, DELETE /sites/environments/{env_id}, GET /sites/{site_id} (for read)
- **Schema Mapping:** (existing correct, document rationale)
- **Lifecycle:** (existing correct, document workarounds)
- **ForceNew Rules:** All fields (documented)
- **Async:** Yes (create, delete)
- **Environment ID Discovery:** before/after comparison (document pattern)
- **Eventual Consistency:** Retry display_name conflicts (document exponential backoff)
- **Write-only Fields:** site_title, is_premium, admin_* (DiffSuppressFunc pattern)
- **Import:** site_id:env_id format
- **Test Plan:**
  - Unit: (existing comprehensive)
  - Acceptance: Add basic create/delete, import test

#### specs/22-kinsta-wordpress-site-domain.md
**Scope:** New resource for site domains

**Contents:**
- **Endpoints:**
  - POST /sites/environments/{env_id}/domains
  - DELETE /sites/environments/{env_id}/domains (requires domain_id)
  - GET /sites/{site_id} → environments[].domains[]
  - PUT /sites/environments/{env_id}/change-primary-domain
- **Schema Mapping:**
  - **Inputs:** environment_id, domain (FQDN)
  - **Computed:** id (domain_id), type, is_primary, verification_status
- **Lifecycle:**
  - Create: POST → async operation? (need to verify from spec)
  - Read: GET site → find environment → find domain in list
  - Update: If is_primary changes → PUT change-primary-domain
  - Delete: DELETE
- **ForceNew:** environment_id, domain
- **Validation:** domain FQDN format
- **Test Plan:** Unit + acceptance (add domain, change primary, delete)

#### specs/23-kinsta-wordpress-site-tools.md
**Scope:** Tool operations (clear cache, restart PHP, etc.)

**Design Decision Required:**
- **Option A:** Resources that trigger on creation (kinsta_wordpress_cache_clear)
  - Pro: Explicit in plan, can depend on other resources
  - Con: Must use null_resource or other triggers to re-run
- **Option B:** Data sources with triggers (data.kinsta_wordpress_cache_clear)
  - Pro: Can be used in depends_on chains
  - Con: Data sources shouldn't have side effects
- **Option C:** Functions (Terraform 1.8+)
  - Pro: Clean semantics
  - Con: Requires newer Terraform

**Recommendation:** Option A with terraform_data replacement triggers

**Tools:**
- kinsta_wordpress_cache_clear (POST /sites/tools/clear-cache)
- kinsta_wordpress_php_restart (POST /sites/tools/restart-php)
- kinsta_wordpress_php_version (PUT /sites/tools/modify-php-version) - could be resource with Update
- kinsta_wordpress_denied_ips (GET/PUT /sites/tools/denied-ips) - resource managing IP list

### 5.3 Sevalla Provider Resources

#### specs/50-sevalla-database.md
**Scope:** Migrate kinsta_database to Sevalla API

**Required Evidence:**
1. Sevalla OpenAPI spec: Request POST /databases
2. Response format: Sync or async? (MyKinsta is sync 200, likely same)
3. Update PUT /databases/{id}: Which fields updatable?
4. Read GET /databases/{id}: Full schema
5. Delete DELETE /databases/{id}: Sync or async?
6. Error response format

**Provisional Schema (based on MyKinsta, to be validated):**
- **Inputs:** location, resource_type (enum), display_name, db_name, db_password, db_user (optional for Redis), type (enum), version
- **Computed:** id, created_at, memory_limit, cpu_limit, storage_size, cluster{id, location, display_name}, status, resource_type_name
- **ForceNew:** location, db_name, type, version (CRITICAL FIX)
- **Updatable:** resource_type, display_name (validate from Sevalla spec)
- **Validation:**
  - type enum: postgresql, redis, mariadb, mysql
  - resource_type enum: db1-db9
  - db_user required unless type=redis
- **Sensitive:** db_password
- **Lifecycle:** CRUD (all sync based on MyKinsta pattern, verify for Sevalla)

#### specs/51-sevalla-application.md
**Scope:** New application resource

**Required Evidence:**
1. Sevalla OpenAPI spec: POST /applications (or different endpoint?)
2. Application lifecycle: Create, read, update, delete
3. Deployment semantics: Separate resource or nested?
4. Process management: Separate resource or nested?
5. Computed vs input fields

**Provisional Schema:**
- **Inputs:** name, display_name, build_path, default_branch, git_repository, runtime (Node.js, Python, etc.), start_command, environment_variables (map)
- **Computed:** id, status, created_at, deployment_id (latest), url (app URL)
- **ForceNew:** TBD (likely git_repository, runtime)
- **Updatable:** TBD (likely display_name, environment_variables)

#### specs/52-sevalla-static-site.md
**Scope:** New static site resource

**Required Evidence:** Similar to application

#### specs/53-sevalla-application-deployment.md
**Scope:** Separate deployment resource or sub-resource?

**Design Decision Required:**
- **Option A:** Separate resource (sevalla_application_deployment)
  - Pro: Explicit control over deployments
  - Con: Dependency management complex
- **Option B:** Computed attribute + trigger mechanism
  - Pro: Simpler
  - Con: Less control

### 5.4 Data Sources

#### specs/30-kinsta-data-sources.md
**Scope:** All Kinsta provider data sources

**Priority List:**
- kinsta_wordpress_sites (GET /sites?company={id})
- kinsta_company_regions (GET /company/{id}/available-regions)
- kinsta_company_users (GET /company/{id}/users)
- kinsta_wordpress_logs (GET /sites/environments/{env_id}/logs)
- kinsta_wordpress_backups (GET /sites/environments/{env_id}/backups)

**Each Data Source Needs:**
- Endpoint mapping
- Filter/search parameters
- Output schema
- Pagination handling (if applicable)
- Use cases (why data source vs resource?)

#### specs/54-sevalla-data-sources.md
**Scope:** All Sevalla provider data sources

---

## 6. Quality Gates & Refactoring Opportunities

### 6.1 Client Refactoring

#### Centralized Error Parsing
```go
// internal/client/error.go
type APIError struct {
    StatusCode int
    Message    string
    Status     int
    Data       map[string]interface{}
}

func (e *APIError) Error() string {
    return fmt.Sprintf("Kinsta API error (HTTP %d): %s", e.StatusCode, e.Message)
}

func (e *APIError) IsNotFound() bool {
    return e.StatusCode == 404
}

func (e *APIError) IsRateLimited() bool {
    return e.StatusCode == 429
}

func parseAPIError(resp *http.Response) error {
    // Parse JSON body, extract message/status/data
}
```

#### Request Helpers
```go
// internal/client/request.go
func (c *Client) doWithRetry(ctx context.Context, method, path string, body io.Reader, v interface{}) error {
    // Implements exponential backoff for retryable errors
}

func (c *Client) getJSON(ctx context.Context, path string, v interface{}) error {
    return c.do(ctx, http.MethodGet, path, nil, v)
}

func (c *Client) postJSON(ctx context.Context, path string, body interface{}, v interface{}) error {
    // Marshal body, call do
}

func (c *Client) putJSON(ctx context.Context, path string, body interface{}, v interface{}) error {
    // Marshal body, call do
}

func (c *Client) deleteJSON(ctx context.Context, path string, v interface{}) error {
    return c.do(ctx, http.MethodDelete, path, nil, v)
}
```

### 6.2 Provider Refactoring

#### Shared Resource Helpers
```go
// internal/provider/helpers.go

// FindResourceByDisplayName searches a list for a resource with matching display_name
// Useful for import operations where user provides display_name instead of ID
func findResourceByDisplayName(resources []Resource, displayName string) (*Resource, error) {
    var matches []Resource
    for _, r := range resources {
        if r.DisplayName == displayName {
            matches = append(matches, r)
        }
    }
    if len(matches) == 0 {
        return nil, fmt.Errorf("no resource found with display_name: %s", displayName)
    }
    if len(matches) > 1 {
        return nil, fmt.Errorf("multiple resources found with display_name: %s (use ID for import)", displayName)
    }
    return &matches[0], nil
}

// HandleResourceNotFound checks if error is 404, clears state, returns nil
func handleResourceNotFound(err error, d *schema.ResourceData) error {
    if apiErr, ok := err.(*client.APIError); ok && apiErr.IsNotFound() {
        d.SetId("")
        return nil
    }
    return err
}

// Random name generator for tests
func randomTestName(prefix string) string {
    return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}
```

#### Consistent Importer Pattern
```go
// internal/provider/importer.go

// ImportStatePassthroughWithDisplayName allows importing by ID or display_name
func importStatePassthroughWithDisplayName(
    ctx context.Context,
    d *schema.ResourceData,
    meta interface{},
    listFunc func(context.Context, interface{}) ([]Resource, error),
) ([]*schema.ResourceData, error) {
    id := d.Id()
    
    // Try as UUID first
    if isUUID(id) {
        return schema.ImportStatePassthroughContext(ctx, d, meta)
    }
    
    // Try as display_name
    resources, err := listFunc(ctx, meta)
    if err != nil {
        return nil, err
    }
    
    resource, err := findResourceByDisplayName(resources, id)
    if err != nil {
        return nil, err
    }
    
    d.SetId(resource.ID)
    return []*schema.ResourceData{d}, nil
}
```

### 6.3 Test Refactoring

#### Shared Test Utilities
```go
// internal/provider/testing.go

// AccTestPreCheck validates required environment variables for acceptance tests
func accTestPreCheck(t *testing.T) {
    if os.Getenv("KINSTA_API_KEY") == "" {
        t.Fatal("KINSTA_API_KEY must be set for acceptance tests")
    }
    if os.Getenv("KINSTA_COMPANY_ID") == "" {
        t.Fatal("KINSTA_COMPANY_ID must be set for acceptance tests")
    }
}

// RandomName generates a random test resource name with cleanup prefix
func randomName(prefix string) string {
    return fmt.Sprintf("tf-test-%s-%d", prefix, time.Now().UnixNano())
}

// SweepResources is a cleanup function for leftover test resources
func sweepResources(region string, resourceType string) error {
    // List resources with "tf-test-" prefix, delete them
}
```

#### Mock Client Generator
```go
// internal/provider/mock_client.go

type MockKinstaClient struct {
    client.KinstaClient
    CompanyIDValue string
    OnCreateDatabase func(context.Context, *client.CreateDatabaseRequest) (*client.CreateDatabaseResponse, error)
    OnGetDatabase func(context.Context, string) (*client.GetDatabaseResponse, error)
    // ... etc for all interface methods
}

func (m *MockKinstaClient) CompanyID() string {
    return m.CompanyIDValue
}

func (m *MockKinstaClient) CreateDatabase(ctx context.Context, req *client.CreateDatabaseRequest) (*client.CreateDatabaseResponse, error) {
    if m.OnCreateDatabase != nil {
        return m.OnCreateDatabase(ctx, req)
    }
    return nil, fmt.Errorf("CreateDatabase not mocked")
}
```

### 6.4 Documentation Standards

#### Resource Documentation Template
```markdown
# kinsta_wordpress_site

Manages a WordPress site in Kinsta hosting.

## Example Usage

### Basic Site
\`\`\`hcl
resource "kinsta_wordpress_site" "example" {
  display_name   = "My WordPress Site"
  region         = "us-central1"
  admin_email    = "admin@example.com"
  admin_password = var.admin_password
  admin_user     = "admin"
  site_title     = "My WordPress Site"
  wp_language    = "en_US"
}
\`\`\`

### Multisite
\`\`\`hcl
resource "kinsta_wordpress_site" "multisite" {
  display_name          = "My Network"
  region                = "us-central1"
  is_multisite          = true
  is_subdomain_multisite = true
  admin_email           = "admin@example.com"
  admin_password        = var.admin_password
  admin_user            = "admin"
  site_title            = "My Network"
}
\`\`\`

## Argument Reference

* `display_name` - (Required) Display name for the site.
* `region` - (Required, ForceNew) Data center region. See [Kinsta regions](https://kinsta.com/docs/data-center-locations/).
...

## Attributes Reference

* `id` - The site ID.
* `name` - The auto-generated internal site name.
* `status` - Current site status.
* `environment_id` - The ID of the site's live environment.

## Import

Sites can be imported using the site ID:

\`\`\`
terraform import kinsta_wordpress_site.example 54fb80af-576c-4fdc-ba4f-b596c83f15a1
\`\`\`

Or by display name (must be unique):

\`\`\`
terraform import kinsta_wordpress_site.example "My WordPress Site"
\`\`\`

## Timeouts

* `create` - (Default 15m)
* `delete` - (Default 15m)
```

---

## 7. Ordered Backlog

### Phase 0: Foundation (Pre-Split) - P0

**Goal:** Fix critical bugs, establish patterns, prepare for split

1. **Fix Database Resource Bugs** (1-2 days)
   - Add ForceNew to immutable fields
   - Make db_password, db_user required inputs (breaking change - document)
   - Add 404 handling to Read
   - Update documentation
   - Spec: Update specs/50-sevalla-database.md with fixes

2. **Centralize Error Handling** (2-3 days)
   - Implement specs/10-foundations-http-errors-and-retries.md
   - Create APIError type
   - Update client.do() to parse errors
   - Add handleResourceNotFound helper
   - Update all resources to use new error handling

3. **Improve Operation Polling** (1-2 days)
   - Implement specs/11-foundations-operations-polling.md
   - Add configurable timeout
   - Add exponential backoff
   - Add progress logging
   - Document assumptions about operation.Data

4. **Add Import Support** (1 day)
   - Add import to kinsta_wordpress_site
   - Document import formats
   - Test import by ID and display_name

5. **Create Shared Test Utilities** (1 day)
   - accTestPreCheck
   - randomName generator
   - Mock client builder

### Phase 1: Kinsta Provider Completion - P0/P1

**Goal:** Complete essential WordPress resources before split

6. **Add Site Domains Resource** (2-3 days)
   - Implement specs/22-kinsta-wordpress-site-domain.md
   - kinsta_wordpress_site_domain resource
   - Unit tests + acceptance test
   - Documentation + examples

7. **Add Essential Data Sources** (2-3 days)
   - Implement specs/30-kinsta-data-sources.md (subset)
   - kinsta_wordpress_sites
   - kinsta_company_regions (for validation)
   - Unit + acceptance tests
   - Documentation

8. **Enhance WordPress Site Resource** (1-2 days)
   - Add missing fields: is_multisite, is_subdomain_multisite, woocommerce, wordpressseo
   - Add site_id computed output
   - Update specs/20-kinsta-wordpress-site.md
   - Tests + documentation

9. **Add Acceptance Tests** (2 days)
   - kinsta_wordpress_environment acceptance tests
   - Import tests for all resources
   - Update test sweep utilities

### Phase 2: Provider Split Preparation - P0

**Goal:** Prepare for clean split, deprecate database resource

10. **Create Split Roadmap** (1 day)
    - Write specs/00-roadmap-split-kinsta-vs-sevalla.md
    - Document migration timeline
    - Draft user communication

11. **Add Deprecation Warnings** (1 day)
    - Add deprecation notice to kinsta_database resource
    - Point to Sevalla provider (when available)
    - Add to docs

12. **Prepare Sevalla Provider Scaffold** (2 days)
    - Create terraform-provider-sevalla repository
    - Copy SDK v2 provider scaffold
    - Set up authentication (Bearer token)
    - Configure base URL: https://api.sevalla.com/v2

### Phase 3: Sevalla Provider - P0

**Goal:** Migrate database resource, add essential Sevalla resources

**Blocker:** Requires Sevalla OpenAPI spec or API documentation analysis

13. **Obtain & Analyze Sevalla API Spec** (1-2 days)
    - Fetch/cache Sevalla OpenAPI spec
    - Document request/response schemas
    - Identify async vs sync patterns
    - Update all specs/5X-sevalla-*.md files

14. **Implement sevalla_database** (3-4 days)
    - Implement specs/50-sevalla-database.md
    - Migrate logic from kinsta_database
    - Fix all identified bugs
    - Unit + acceptance tests
    - Documentation + migration guide

15. **Implement sevalla_application** (3-5 days)
    - Implement specs/51-sevalla-application.md
    - Handle async operations (if any)
    - Unit + acceptance tests
    - Documentation + examples

16. **Implement sevalla_static_site** (3-5 days)
    - Implement specs/52-sevalla-static-site.md
    - Unit + acceptance tests
    - Documentation + examples

### Phase 4: Kinsta Provider Enhancement - P1/P2

**Goal:** Add remaining WordPress resources

17. **Add Site Tools Resources** (3-4 days)
    - Implement specs/23-kinsta-wordpress-site-tools.md
    - Tool trigger resources
    - PHP version management
    - Denied IPs resource

18. **Add Backup Resources** (2-3 days)
    - kinsta_wordpress_backup
    - kinsta_wordpress_backup_restore
    - Backup data sources

19. **Add SFTP Resources** (2 days)
    - kinsta_wordpress_sftp_user
    - kinsta_wordpress_sftp_toggle

20. **Add Advanced Resources** (P2, as needed)
    - Plugins, themes
    - PHP allocation
    - Redirect rules
    - Edge cache, CDN
    - SSH configuration

### Phase 5: Sevalla Provider Enhancement - P1/P2

21. **Add Deployment Resources** (2-3 days)
    - Implement specs/53-sevalla-application-deployment.md
    - Design decision on deployment management

22. **Add Data Sources** (2-3 days)
    - Implement specs/54-sevalla-data-sources.md
    - List resources, metrics

---

## 8. Critical Path Summary

### Immediate Actions (Before Split)
1. ✅ Fix kinsta_database ForceNew bugs
2. ✅ Implement centralized error handling
3. ✅ Add 404 handling to all Read operations
4. ✅ Document current state and gaps (this document)

### Split Prerequisites
1. ⏳ Obtain Sevalla OpenAPI spec or API documentation
2. ⏳ Validate Sevalla authentication (Bearer token compatibility)
3. ⏳ Confirm Sevalla async operation patterns
4. ⏳ Create provider split communication plan

### Migration Risks
1. **Database users on kinsta provider:** Need migration guide, deprecation period
2. **State migration:** If kinsta_database moves to sevalla provider, need state mv instructions
3. **API changes:** Sevalla API may differ from deprecated MyKinsta endpoints

---

## 9. Unknowns Requiring Evidence

### Sevalla API (HIGH PRIORITY)

**Without these, Sevalla provider cannot be implemented:**

1. **Complete OpenAPI Spec**
   - Request: Provide `./_spec_cache/sevalla.openapi.json`
   - Or: Manual API documentation analysis from https://api-docs.sevalla.com/
   - Need: All request/response schemas, error formats, operation IDs

2. **Database Resource**
   - Are field names identical to MyKinsta? (location vs region, resource_type vs size, type vs db_type?)
   - Which fields are updatable? (Same as MyKinsta: resource_type, display_name?)
   - Is create sync (200 immediate) or async (202 operation_id)?
   - Is delete sync or async?
   - Full read schema (are cluster, limits, created_at present?)

3. **Application Resource**
   - Complete create request schema
   - Update capabilities (PUT/PATCH?)
   - Deployment lifecycle (automatic vs manual?)
   - Process management (nested or separate resource?)

4. **Static Site Resource**
   - Create request schema
   - Deployment management
   - Build configuration

5. **Error Response Format**
   - Is it identical to MyKinsta?
   - Status code usage (404, 400, 401, 429, 500)
   - Error message structure

6. **Rate Limiting**
   - Response headers?
   - Retry-After header?
   - Limits per endpoint?

### MyKinsta API Clarifications

**Lower priority but would improve implementation:**

1. **Site Creation with plain/clone modes**
   - Different endpoints? (/sites/plain, /sites/clone)
   - Additional required fields?
   - Response format differences?

2. **Domain Verification**
   - Automatic or manual?
   - Verification records endpoint usage
   - How long for verification?

3. **Tool Operations**
   - Are they async (202 + operation_id) or sync (200)?
   - Spec says some provide 202, need to confirm per-endpoint

4. **Pagination**
   - Do list endpoints support pagination?
   - Query parameters (limit, offset)?
   - Response format (cursor-based or offset-based)?

---

## 10. Recommended Next Steps

### Immediate (This Week)
1. **Review this analysis** with team
2. **Request Sevalla API spec** from API team or scrape documentation
3. **Create GitHub issues** for Phase 0 bugs
4. **Set up project board** with backlog phases
5. **Draft user communication** about database resource deprecation

### Short Term (Next 2 Weeks)
1. **Fix database resource bugs** (Phase 0, items 1-3)
2. **Implement error handling improvements** (Phase 0, items 2-3)
3. **Begin Sevalla API analysis** (once spec obtained)
4. **Write detailed specs** for Phase 1 resources

### Medium Term (Next Month)
1. **Complete Phase 0 & 1** (foundation + essential Kinsta resources)
2. **Begin Sevalla provider** (Phase 2-3)
3. **Beta release** Sevalla provider
4. **Deprecation communication** for kinsta_database

### Long Term (Next Quarter)
1. **Complete both providers** (P0 + P1 resources)
2. **Sunset kinsta_database** after migration period
3. **Community feedback** and iteration
4. **P2 enhancements** based on user needs

---

## 11. Spec Files Manifest

The following spec files should be created under `./specs/`:

### Foundations
- ✅ `00-roadmap-split-kinsta-vs-sevalla.md` - Split strategy and timeline
- ✅ `10-foundations-http-errors-and-retries.md` - Error handling patterns
- ✅ `11-foundations-operations-polling.md` - Async operation contract

### Kinsta Provider - Resources
- ✅ `20-kinsta-wordpress-site.md` - Site resource (refine existing)
- ✅ `21-kinsta-wordpress-environment.md` - Environment resource (document existing)
- ✅ `22-kinsta-wordpress-site-domain.md` - Domain management
- ✅ `23-kinsta-wordpress-site-tools.md` - Tool operations (cache, PHP, etc.)
- 🔲 `24-kinsta-wordpress-backup.md` - Backup management
- 🔲 `25-kinsta-wordpress-sftp.md` - SFTP user management
- 🔲 `26-kinsta-wordpress-advanced.md` - Advanced resources (plugins, themes, etc.)

### Kinsta Provider - Data Sources
- ✅ `30-kinsta-data-sources.md` - All data sources (sites, regions, users, logs, etc.)

### Sevalla Provider - Resources
- ✅ `50-sevalla-database.md` - Database resource (migrate from Kinsta)
- ✅ `51-sevalla-application.md` - Application resource
- ✅ `52-sevalla-static-site.md` - Static site resource
- ✅ `53-sevalla-application-deployment.md` - Deployment management
- 🔲 `55-sevalla-pipeline.md` - Pipeline resources

### Sevalla Provider - Data Sources
- ✅ `54-sevalla-data-sources.md` - All data sources

### Testing & Quality
- 🔲 `90-testing-standards.md` - Unit and acceptance test patterns
- 🔲 `91-documentation-standards.md` - Resource documentation template
- 🔲 `92-examples-library.md` - Common usage patterns

---

## Appendix A: API Endpoint Coverage Matrix

### MyKinsta API (kinsta provider)

| Endpoint | Method | Implemented | Priority | Notes |
|----------|--------|-------------|----------|-------|
| `/sites` | GET | ❌ | P0 | Data source |
| `/sites` | POST | ✅ | P0 | kinsta_wordpress_site |
| `/sites/{site_id}` | GET | ✅ | P0 | kinsta_wordpress_site read |
| `/sites/{site_id}` | DELETE | ✅ | P0 | kinsta_wordpress_site delete |
| `/sites/{site_id}/environments` | GET | ✅ | P0 | Used by environment read |
| `/sites/{site_id}/environments` | POST | ✅ | P0 | kinsta_wordpress_environment |
| `/sites/environments/{env_id}` | DELETE | ✅ | P0 | kinsta_wordpress_environment delete |
| `/sites/environments/{env_id}/domains` | GET | ❌ | P0 | Read via site |
| `/sites/environments/{env_id}/domains` | POST | ❌ | P0 | kinsta_wordpress_site_domain |
| `/sites/environments/{env_id}/domains` | DELETE | ❌ | P0 | kinsta_wordpress_site_domain delete |
| `/operations/{operation_id}` | GET | ✅ | P0 | PollOperation |
| `/company/{id}/available-regions` | GET | ❌ | P1 | Data source |
| `/sites/tools/clear-cache` | POST | ❌ | P1 | Tool resource |
| `/sites/tools/restart-php` | POST | ❌ | P1 | Tool resource |
| `/sites/tools/modify-php-version` | PUT | ❌ | P1 | Resource with update |
| (other 40+ endpoints) | * | ❌ | P1/P2 | Backups, SFTP, plugins, etc. |

### Sevalla API (sevalla provider)

| Endpoint (Presumed) | Method | Implemented | Priority | Blocker |
|---------------------|--------|-------------|----------|---------|
| `/databases` | POST | ❌ | P0 | Need spec |
| `/databases/{id}` | GET | ❌ | P0 | Need spec |
| `/databases/{id}` | PUT | ❌ | P0 | Need spec |
| `/databases/{id}` | DELETE | ❌ | P0 | Need spec |
| `/applications` | POST | ❌ | P0 | Need spec |
| `/applications/{id}` | GET | ❌ | P0 | Need spec |
| `/applications/{id}` | PUT | ❌ | P0 | Need spec |
| `/applications/{id}` | DELETE | ❌ | P0 | Need spec |
| (other Sevalla endpoints) | * | ❌ | P0-P2 | Need spec |

---

## Appendix B: Schema Field Coverage

### Database Resource Field Coverage

| Field | Spec (MyKinsta) | Implemented | Spec (Sevalla) | Notes |
|-------|-----------------|-------------|----------------|-------|
| **Inputs (Create)** |
| company_id | ✅ required | ✅ (from provider) | ? | From auth context |
| location | ✅ required | ✅ (as "region") | ? | ⚠️ Field name differs |
| resource_type | ✅ required, enum | ✅ (as "size") | ? | ⚠️ Field name differs |
| display_name | ✅ required | ✅ | ? | |
| db_name | ✅ required | ✅ (as "name") | ? | ⚠️ Field name differs |
| db_password | ✅ required | ✅ (generated) | ? | ❌ Should be user input |
| db_user | optional (Redis) | ✅ (generated) | ? | ❌ Should be user input |
| type | ✅ required, enum | ✅ (as "db_type") | ? | ⚠️ Field name differs |
| version | ✅ required | ✅ | ? | |
| **Computed (Read)** |
| id | ✅ | ✅ | ? | |
| name | ✅ | ✅ | ? | |
| display_name | ✅ | ✅ | ? | |
| status | ✅ | ❌ | ? | ❌ Missing |
| created_at | ✅ | ❌ | ? | ❌ Missing |
| memory_limit | ✅ | ❌ | ? | ❌ Missing |
| cpu_limit | ✅ | ❌ | ? | ❌ Missing |
| storage_size | ✅ | ❌ | ? | ❌ Missing |
| cluster | ✅ nested | ❌ | ? | ❌ Missing |
| resource_type_name | ✅ | ❌ | ? | ❌ Missing |
| **Update** |
| resource_type | ✅ updatable | ✅ | ? | |
| display_name | ✅ updatable | ✅ | ? | |
| **ForceNew (Should Be)** |
| location | ❌ NOT in spec | ❌ NOT in code | ? | ❌ BUG |
| db_name | ❌ NOT in spec | ❌ NOT in code | ? | ❌ BUG |
| type | ❌ NOT in spec | ❌ NOT in code | ? | ❌ BUG |
| version | ❌ NOT in spec | ❌ NOT in code | ? | ❌ BUG |

---

**End of Analysis**

**Questions:** Contact before implementation of Sevalla provider.

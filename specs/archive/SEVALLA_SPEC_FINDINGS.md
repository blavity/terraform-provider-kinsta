# Sevalla API Specification Analysis

**Date:** 2026-01-01  
**Sevalla API Version:** 1.80.0  
**Source:** https://api-docs.sevalla.com/openapi.json

---

## Executive Summary

The Sevalla OpenAPI spec has been obtained and analyzed. **Key finding: Database resource migration is straightforward** - schemas are identical to MyKinsta deprecated endpoint, only base URL changes.

---

## 1. Sevalla API Overview

- **Base URL:** `https://api.sevalla.com/v2`
- **Authentication:** Bearer token (same API key as MyKinsta)
- **Total Endpoints:** 60
- **Total Schemas:** 183
- **Shared with MyKinsta:** Operations polling endpoint

---

## 2. Resource Categories

### Core Resources (Sevalla Provider)
1. **Applications** (18 endpoints) - Full CRUD + deployments, processes, metrics
2. **Databases** (4 endpoints) - Full CRUD
3. **Static Sites** (5 endpoints) - Full CRUD + deployments
4. **Pipelines** (2 endpoints) - List + create preview apps

### Critical Exclusion (MUST NOT Implement)
**WordPress Sites:** Present in Sevalla spec but Sevalla provider MUST NOT implement.

**Evidence:** `sevalla.openapi.json` contains `/sites/*` endpoints

**Requirement:** These endpoints exist for backward compatibility but implementing them in Sevalla provider would:
1. Create provider overlap (same resources in two providers)
2. Confuse users about which provider to use
3. Violate clean separation: WordPress → Kinsta, PaaS → Sevalla

**Action:** Explicitly exclude from Sevalla provider scope and implementation plans.

**Users managing WordPress MUST use:** `terraform-provider-kinsta`

---

## 3. Database Resource - Migration Analysis

### 3.1 Schema Comparison: MyKinsta vs Sevalla

#### CREATE Request (POST /databases)
**Status: ✅ IDENTICAL**

```json
{
  "company_id": "string (uuid)",
  "location": "string",
  "resource_type": "enum[db1-db9]",
  "display_name": "string",
  "db_name": "string",
  "db_password": "string",
  "db_user": "string (optional for Redis)",
  "type": "enum[postgresql, redis, mariadb, mysql]",
  "version": "string"
}
```

**Required fields:** All except `db_user` (optional only for Redis)

#### CREATE Response
**Status: ✅ IDENTICAL - Synchronous**

```json
HTTP 200
{
  "database": {
    "id": "uuid"
  }
}
```

No `operation_id` - immediate ID return.

#### READ Response (GET /databases/{id})
**Status: ✅ ENHANCED - More fields than MyKinsta**

**MyKinsta fields:**
- id, name, display_name, status, created_at
- memory_limit, cpu_limit, storage_size
- type, version, cluster{}, resource_type_name

**Sevalla ADDITIONAL fields:**
- ✨ `internal_hostname` - Internal DNS name for internal connections
- ✨ `internal_port` - Internal port
- ✨ `internal_connections[]` - Array of connected applications
- ✨ `data` object with credentials:
  - `db_name`
  - `db_password`
  - `db_root_password`
  - `db_user`
- ✨ `external_connection_string` - Full connection string
- ✨ `external_hostname` - External DNS name
- ✨ `external_port` - External port

**Impact:** Terraform schema should include these as computed fields

#### UPDATE Request (PUT /databases/{id})
**Status: ✅ IDENTICAL**

```json
{
  "resource_type": "enum[db1-db9]",  // optional
  "display_name": "string"           // optional
}
```

Only these two fields are updatable.

#### DELETE (DELETE /databases/{id})
**Status: ✅ Synchronous (assumed 200 response)**

### 3.2 Migration Checklist

**Current kinsta_database Issues:**
- ❌ Missing ForceNew on immutable fields (location, type, version, db_name)
- ❌ Generates random db_password/db_user instead of requiring user input
- ❌ Missing computed fields (status, created_at, limits, cluster, connection info)
- ❌ No 404 handling in Read
- ❌ Field name mismatches (size→resource_type, db_type→type, region→location)

**Required Changes for sevalla_database:**
1. ✅ Change base URL: `api.kinsta.com` → `api.sevalla.com`
2. ✅ Add ForceNew: location, type, version, db_name
3. ✅ Make db_password, db_user required inputs (breaking change)
4. ✅ Add computed fields:
   - status
   - created_at  
   - memory_limit, cpu_limit, storage_size
   - cluster (nested block)
   - internal_hostname, internal_port
   - internal_connections (list)
   - external_connection_string
   - external_hostname, external_port
5. ✅ Add sensitive flag to data.db_password, data.db_root_password
6. ✅ Fix field names to match API:
   - size → resource_type
   - db_type → type
   - region → location
7. ✅ Add 404 handling to Read
8. ✅ Update validation:
   - type enum validation
   - resource_type enum validation
   - db_user conditional requirement (optional for Redis only)

**Estimated Effort:** 2-3 days (straightforward, mostly adding computed fields)

---

## 4. Application Resource - Schema Analysis

### 4.1 Endpoints

**Evidence:** `sevalla.openapi.json#/paths/~1applications`

```
GET    /applications?company={id}&limit={n}&offset={n}  ✅ List
GET    /applications/{id}                                 ✅ Read
GET    /applications/{name}                               ✅ Read by name
PUT    /applications/{id}                                 ✅ Update
DELETE /applications/{id}                                 ✅ Delete
POST   /applications                                      ❌ NOT IN SPEC
```

**Critical finding:** No POST endpoint exists for application creation.

**Impact:** 
- `sevalla_application` resource cannot be fully implemented (no Create operation)
- Can implement read-only `sevalla_applications` data source
- Resource implementation blocked pending API endpoint availability

**Priority adjustment:** P0 = data source, P2 = resource (blocked on API)

### 4.2 Application Schema Analysis

**Status:** Blocked on POST endpoint availability. GET schema analyzed, but without create endpoint, full resource spec cannot be completed.

**Next step:** Confirm with API team if POST /applications is planned or if applications are UI/CLI-only.

**Schema notes from GET endpoints:**

**Key fields expected:**
- id, name, display_name
- status (deploying, deployed, failed, etc.)
- git_repository, default_branch
- build_path, start_command
- runtime/language (Node.js, Python, Go, etc.)
- environment_variables
- deployment_id (latest deployment)
- processes[] (web, worker, etc.)
- domains[]
- created_at

### 4.3 Application Lifecycle

**Create:** Not in spec - applications created via UI/CLI only?
**Read:** GET /applications/{id} - full application details
**Update:** PUT /applications/{id} - update settings
**Delete:** DELETE /applications/{id} - remove application

**Deployment:** Separate POST /applications/deployments
- Triggers new deployment
- Returns deployment.id immediately (synchronous)
- Poll deployment status via GET /applications/deployments/{deployment_id}

**Design Decision Required:**
- **Option A:** Application resource + separate deployment resource
- **Option B:** Application resource with deployment trigger attribute
- **Recommendation:** Option A for explicit control

### 4.4 Estimated Application Resource Fields

**Inputs:**
- display_name (required)
- ... (need full schema analysis)

**Computed:**
- id
- name (auto-generated)
- status
- deployment_id (latest)
- ... (need full schema)

**Updatable:**
- display_name
- ... (need PUT schema analysis)

**ForceNew:**
- TBD (need to determine from API behavior)

---

## 5. Static Site Resource - Schema Analysis

Similar pattern to applications. Need detailed schema review.

---

## 6. Comparison: MyKinsta vs Sevalla

### 6.1 Shared Patterns

✅ **Authentication:** Both use Bearer token  
✅ **Error Responses:** Same format (status, message, data)  
✅ **Operations Polling:** Same endpoint `/operations/{operation_id}`  
✅ **Company Context:** Both use company_id from auth  

### 6.2 Differences

| Aspect | MyKinsta | Sevalla |
|--------|----------|---------|
| **Async Operations** | WordPress sites use 202 + operation_id | Databases/apps use 200 immediate |
| **WordPress Sites** | Primary API | Also present (should NOT use) |
| **Databases** | Deprecated endpoint | Primary API (identical schema) |
| **Applications** | Deprecated endpoint | Primary API |
| **Static Sites** | Deprecated endpoint | Primary API |

### 6.3 Operations Endpoint Behavior

**MyKinsta WordPress operations:**
- Site creation: 202 → poll → 200 (data is OPAQUE per `swagger.json#/components/schemas/OperationResponse`)
- Site deletion: 202 → poll → 200
- Environment creation: 202 → poll → 200

**Critical:** `operation.data` field is defined as `{}` (empty object) in spec. Cannot rely on `data.idSite` or `data.idEnv` keys - these are observed behavior but not guaranteed by API contract.

**Implementation strategy:** Use lookup-after-poll (list resources and match by display_name/timestamp) instead of relying on data extraction. See `specs/02-operations-polling-contract.md` for details.

**Sevalla operations:**
- Databases: Immediate 200 (no polling)
- Applications: TBD (need to check deployment patterns)
- Static Sites: TBD (need to check deployment patterns)

---

## 7. Updated Provider Split Strategy

### terraform-provider-kinsta (MyKinsta API)

**Scope:** WordPress ONLY
- kinsta_wordpress_site
- kinsta_wordpress_environment  
- kinsta_wordpress_site_domain
- kinsta_wordpress_tool_* (cache, PHP, etc.)
- kinsta_wordpress_backup
- kinsta_wordpress_sftp_user
- Data sources: sites, regions, logs, backups

**Remove from scope:**
- ❌ kinsta_database (migrate to Sevalla)
- ❌ kinsta_application (never implement - use Sevalla)
- ❌ kinsta_static_site (never implement - use Sevalla)

### terraform-provider-sevalla (Sevalla API)

**Scope:** Applications, Databases, Static Sites
- sevalla_database (migrate from kinsta_database)
- sevalla_application (new)
- sevalla_application_deployment (new)
- sevalla_static_site (new)
- sevalla_static_site_deployment (new)
- sevalla_pipeline_preview_app (new)
- Data sources: applications, databases, static_sites, pipelines

**Exclude from implementation:**
- ❌ /sites/* endpoints (use terraform-provider-kinsta instead)

---

## 8. Migration Guide for Database Users

### Current State (kinsta_database)
```hcl
resource "kinsta_database" "main" {
  name         = "mydb"
  display_name = "My Database"
  region       = "us-central1"
  db_type      = "postgresql"
  version      = "15"
  size         = "db1"
}
```

### Future State (sevalla_database)
```hcl
resource "sevalla_database" "main" {
  db_name      = "mydb"
  display_name = "My Database"
  location     = "us-central1"
  type         = "postgresql"
  version      = "15"
  resource_type = "db1"
  
  # NEW: Required user inputs (previously generated)
  db_password  = var.db_password
  db_user      = "myuser"
}

output "connection_string" {
  value     = sevalla_database.main.external_connection_string
  sensitive = true
}

output "internal_hostname" {
  value = sevalla_database.main.internal_hostname
}
```

### Breaking Changes
1. **Field renames:**
   - `region` → `location`
   - `db_type` → `type`
   - `size` → `resource_type`
   - `name` → `db_name`

2. **Required inputs:**
   - `db_password` now required (was auto-generated)
   - `db_user` now required for non-Redis (was auto-generated)

3. **New computed outputs:**
   - `external_connection_string` - Full connection string
   - `internal_hostname` - For app-to-db connections
   - `status`, `created_at`, `memory_limit`, `cpu_limit`, `storage_size`

### State Migration
```bash
# 1. Import existing database to new provider
terraform import sevalla_database.main <database-id>

# 2. Remove from old provider state
terraform state rm kinsta_database.main

# 3. Update configuration with new field names
# 4. Apply (should show no changes if migration correct)
```

---

## 9. Next Steps

### Immediate (This Week)
1. ✅ Sevalla spec obtained and analyzed
2. 🔲 Complete application schema analysis
3. 🔲 Complete static site schema analysis
4. 🔲 Create detailed spec files for sevalla_database
5. 🔲 Create detailed spec files for sevalla_application

### Short Term (Next 2 Weeks)
1. 🔲 Fix kinsta_database bugs in current provider
2. 🔲 Add deprecation warnings to kinsta_database
3. 🔲 Create sevalla provider repository
4. 🔲 Implement sevalla_database (migrate from kinsta)

### Medium Term (Next Month)
1. 🔲 Implement sevalla_application
2. 🔲 Implement sevalla_static_site
3. 🔲 Beta release sevalla provider
4. 🔲 User migration communication

---

## 10. Unknowns Resolved

### ✅ Confirmed
1. **Database schemas:** Identical to MyKinsta deprecated endpoint
2. **Authentication:** Same Bearer token works for both APIs
3. **Base URLs:** Different (api.kinsta.com vs api.sevalla.com)
4. **Database operations:** Synchronous (200 immediate ID)
5. **Update semantics:** Only resource_type and display_name updatable
6. **Error format:** Same structure across both APIs

### ⏳ Still Need Evidence
1. **Application create endpoint:** Not found in spec - UI/CLI only?
2. **Application update fields:** Which fields are updatable via PUT?
3. **Application ForceNew fields:** Which require resource replacement?
4. **Static site patterns:** Similar to applications?
5. **Deployment lifecycle:** Immediate or polled?

---

## Appendix: Sevalla Endpoint List

### Applications (18)
- GET /applications
- GET /applications/{id}
- GET /applications/{name}
- PUT /applications/{id}
- DELETE /applications/{id}
- POST /applications/{id}/internal-connections
- POST /applications/{id}/cdn/toggle-status
- POST /applications/{id}/edge-cache/toggle-status
- POST /applications/{id}/clear-cache
- GET /applications/{id}/metrics/bandwidth
- GET /applications/{id}/metrics/build-time
- GET /applications/{id}/metrics/run-time
- GET /applications/{id}/metrics/http-requests
- GET /applications/{id}/metrics/response-time
- GET /applications/{id}/metrics/slowest-requests
- GET /applications/{id}/metrics/cpu-usage
- GET /applications/{id}/metrics/memory-usage
- GET /applications/deployments/{deployment_id}
- GET /applications/processes/{process_id}
- PUT /applications/processes/{id}
- POST /applications/deployments
- POST /applications/promote

### Databases (4)
- GET /databases
- POST /databases
- GET /databases/{id}
- GET /databases/{name}
- PUT /databases/{id}
- DELETE /databases/{id}

### Static Sites (5)
- GET /static-sites
- GET /static-sites/{id}
- PUT /static-sites/{id}
- DELETE /static-sites/{id}
- POST /static-sites/deployments
- GET /static-sites/deployments/{deployment_id}
- POST /static-sites/deployments/redeploy

### Pipelines (2)
- GET /pipelines
- POST /pipelines/{id}/create-preview-app

### Operations (1)
- GET /operations/{operation_id}

### Company (1)
- GET /company/{id}/users

---

**End of Sevalla Findings**

For full provider split analysis, see `PROVIDER_SPLIT_ANALYSIS.md`

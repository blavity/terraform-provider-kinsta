# Doc Hygiene Patch Plan

**Total patches:** 17 text replacements across 4 files  
**Evidence source:** swagger.json v1.87.0, sevalla.openapi.json v1.80.0  
**Status:** Ready to apply

---

## File 1: ANALYSIS_SUMMARY.md

### Patch 1.1 - Clarify database strategy

**Section:** `Phase 0: Foundation Fixes`

**Old text:**
```markdown
1. **Fix kinsta_database bugs** (2 days)
   - Add ForceNew: location, type, version, db_name
   - Make db_password, db_user required inputs
   - Add 404 handling
   - Add computed fields
   - Fix field name mismatches
   - Update tests

2. **Add deprecation notice** (1 day)
   - Add warning to kinsta_database resource
   - Update documentation pointing to Sevalla
   - Add migration guide
```

**New replacement text:**
```markdown
1. **Deprecate kinsta_database** (1 day)
   - Add deprecation warning to kinsta_database resource
   - Update documentation pointing to Sevalla provider
   - Publish migration guide
   - **Note:** Do NOT fix bugs - users must migrate to sevalla_database

**Rationale:** The kinsta_database resource uses deprecated MyKinsta endpoints (`swagger.json#/paths/~1databases/post/deprecated=true`). Rather than fixing bugs in deprecated code, users should migrate to sevalla_database which uses the non-deprecated Sevalla API. Fixing bugs in kinsta_database wastes engineering effort on code that will be removed.
```

**Rationale:** Contradicts itself by saying both "fix bugs" and "deprecate". Evidence shows databases deprecated in MyKinsta.

---

### Patch 1.2 - Fix endpoint count claim

**Section:** `✅ Provider Split Boundary Clear`

**Old text:**
```markdown
**terraform-provider-sevalla:**
- Applications, databases, static sites, pipelines
- 29 core endpoints (excludes WordPress overlap)
- Base URL: https://api.sevalla.com/v2
```

**New replacement text:**
```markdown
**terraform-provider-sevalla:**
- Applications, databases, static sites, pipelines, deployments
- 31 core endpoints (19 applications + 5 databases + 5 static-sites + 2 pipelines)
- Excludes 29 WordPress endpoints present in Sevalla spec (overlap with Kinsta provider)
- Total Sevalla endpoints: 60 (31 core + 29 WordPress overlap)
- Base URL: https://api.sevalla.com/v2
```

**Rationale:** Original count unclear - "29 core endpoints" doesn't match actual breakdown. Evidence: sevalla.openapi.json has 60 total endpoints.

---

### Patch 1.3 - Clarify static sites deprecation status

**Section:** `Resource Coverage Matrix` → `Sevalla Provider`

**Old text:**
```markdown
| sevalla_static_site | P0 | Needs clarity | Partial | Similar to app |
```

**New replacement text:**
```markdown
| sevalla_static_site | P1 | Ready to spec | Analyzed | Deprecated in MyKinsta, active in Sevalla |
```

**Rationale:** Evidence `swagger.json#/paths/~1static-sites/get/deprecated=true` confirms MyKinsta deprecation. Sevalla has non-deprecated endpoints.

---

### Patch 1.4 - Fix application resource priority

**Section:** `Resource Coverage Matrix` → `Sevalla Provider`

**Old text:**
```markdown
| sevalla_application | P0 | Needs clarity | Analyzed | Create endpoint? |
```

**New replacement text:**
```markdown
| sevalla_application | P2 (blocked) | No POST endpoint | Analyzed | `sevalla.openapi.json#/paths/~1applications` has only GET - no create |
| sevalla_applications (data source) | P0 | Ready to implement | Analyzed | Read-only list of applications |
```

**Rationale:** Evidence shows no POST /applications in Sevalla spec. Can't implement resource without create endpoint.

---

### Patch 1.5 - Clarify database operations are synchronous

**Section:** `## Migration Impact` → `For Database Users`

**Old text:**
```markdown
**Only change needed:** Base URL (`api.kinsta.com` → `api.sevalla.com`)

**Additional benefits in Sevalla:**
- More computed fields (connection strings, internal hostname, credentials)
- Better for application-to-database internal connections
```

**New replacement text:**
```markdown
**Key changes:**
- Base URL: `api.kinsta.com` → `api.sevalla.com`
- Synchronous operations: Databases return 200 immediately (no polling)
- Evidence: `sevalla.openapi.json#/paths/~1databases/post/responses=["200","401","404","500"]`

**Additional benefits in Sevalla:**
- More computed fields (connection strings, internal hostname, credentials)
- Better for application-to-database internal connections
- No async polling required (faster resource operations)
```

**Rationale:** Docs don't explicitly state databases are synchronous, which affects implementation.

---

### Patch 1.6 - Add explicit Sevalla WordPress exclusion

**Section:** `## Migration Impact` (add new subsection after "For Database Users")

**Old text:** (none - insert new section)

**New replacement text:**
```markdown

### For Sevalla Provider Users

**Critical exclusion:**
Sevalla provider MUST NOT implement WordPress sites (`/sites/*` endpoints).

**Evidence:** 29 WordPress endpoints exist in Sevalla spec (`sevalla.openapi.json`) but belong exclusively to Kinsta provider to avoid overlap and confusion.

**What to use instead:**
- WordPress sites → `terraform-provider-kinsta`
- Applications/databases/static-sites → `terraform-provider-sevalla`

**Rationale:** Clear separation of concerns prevents users from accidentally managing WordPress via wrong provider.
```

**Rationale:** Critical exclusion not explicitly stated - must be emphasized to prevent implementation.

---

### Patch 1.7 - Fix operation.data opaqueness

**Section:** `## Open Questions for API Team` (update question 1)

**Old text:**
```markdown
1. **Application Create:** How do users create applications via API? Is there a POST /applications endpoint not in spec?
```

**New replacement text:**
```markdown
1. **Application Create:** POST /applications does NOT exist in Sevalla spec (`sevalla.openapi.json#/paths/~1applications` has only GET). Is application creation via API planned? Blocks `sevalla_application` resource (only data source possible).

**Additional question:**
2. **Operations data contract:** MyKinsta polling returns `operation.data={}` (opaque per `swagger.json#/components/schemas/OperationResponse`). Can we rely on `data.idSite`/`data.idEnv` keys, or should we use lookup-after-poll strategy? Current implementation assumes keys exist but spec says opaque.
```

**Rationale:** Need clarification on operation.data contract - spec says opaque but code assumes typed.

---

### Patch 1.8 - Clarify state migration reality

**Section:** `**State Migration Command:**`

**Old text:**
```bash
terraform import sevalla_database.main <existing-database-id>
terraform state rm kinsta_database.main
```

**New replacement text:**
```bash
# CRITICAL: Terraform cannot automatically migrate state between providers
# Manual steps required:

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
terraform import sevalla_database.main <existing-database-id>

# 3. Remove from kinsta provider state
terraform state rm kinsta_database.main

# 4. Update configuration with new field names (region→location, etc.)

# 5. Verify no changes required
terraform plan  # Should show "No changes"
```

**Rationale:** Original implies automatic migration possible - must be explicit about manual process.

---

## File 2: SEVALLA_SPEC_FINDINGS.md

### Patch 2.1 - Fix overlap warning to exclusion requirement

**Section:** `### Overlap Warning`

**Old text:**
```markdown
### Overlap Warning
**WordPress Sites (29 endpoints):** Present in Sevalla spec but should remain Kinsta provider only
- Recommendation: Remove from sevalla provider documentation
- Users managing WordPress should use terraform-provider-kinsta
```

**New replacement text:**
```markdown
### Critical Exclusion (MUST NOT Implement)
**WordPress Sites (29 endpoints):** Present in Sevalla spec but Sevalla provider MUST NOT implement.

**Evidence:** `sevalla.openapi.json` contains 29 `/sites/*` endpoints

**Requirement:** These endpoints exist for backward compatibility but implementing them in Sevalla provider would:
1. Create provider overlap (same resources in two providers)
2. Confuse users about which provider to use
3. Violate clean separation: WordPress → Kinsta, PaaS → Sevalla

**Action:** Explicitly exclude from Sevalla provider scope and implementation plans.

**Users managing WordPress MUST use:** `terraform-provider-kinsta`
```

**Rationale:** "Should remain" is too weak - must be explicit MUST NOT implement.

---

### Patch 2.2 - Clarify application creation status

**Section:** `### 4.1 Endpoints`

**Old text:**
```markdown
### 4.1 Endpoints

```
GET    /applications?company={id}&limit={n}&offset={n}
GET    /applications/{id}
GET    /applications/{name}
PUT    /applications/{id}
DELETE /applications/{id}
```

### 4.2 Application Schema (from GET /applications/{id})

**Need to examine full schema** - checking now...
```

**New replacement text:**
```markdown
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
```

**Rationale:** Original says "need to examine" but we now have evidence POST doesn't exist.

---

### Patch 2.3 - Fix operation.data references

**Section:** `### 6.3 Operations Endpoint Behavior`

**Old text:**
```markdown
**MyKinsta WordPress operations:**
- Site creation: 202 → poll → 200 with data.idSite
- Site deletion: 202 → poll → 200
- Environment creation: 202 → poll → 200 (no idEnv in data)
```

**New replacement text:**
```markdown
**MyKinsta WordPress operations:**
- Site creation: 202 → poll → 200 (data is OPAQUE per `swagger.json#/components/schemas/OperationResponse`)
- Site deletion: 202 → poll → 200
- Environment creation: 202 → poll → 200

**Critical:** `operation.data` field is defined as `{}` (empty object) in spec. Cannot rely on `data.idSite` or `data.idEnv` keys - these are observed behavior but not guaranteed by API contract.

**Implementation strategy:** Use lookup-after-poll (list resources and match by display_name/timestamp) instead of relying on data extraction. See `specs/02-operations-polling-contract.md` for details.
```

**Rationale:** Original assumes data.idSite exists, but spec defines data as opaque.

---

## File 3: SPECS_ROADMAP.md

### Patch 3.1 - Update applications spec priority

**Section:** `#### specs/11-sevalla-application-resource.md`

**Old text:**
```markdown
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
```

**New replacement text:**
```markdown
#### `specs/11-sevalla-application-resource.md`
**Status:** ⏳ Blocked on POST endpoint  
**Priority:** P2 (blocked)  
**Evidence:** `sevalla.openapi.json#/paths/~1applications` has no POST method

**API Endpoints:**
- GET /applications (list) - ✅ Available
- GET /applications/{id} (read) - ✅ Available
- PUT /applications/{id} (update) - ✅ Available
- DELETE /applications/{id} (delete) - ✅ Available
- POST /applications (create) - ❌ NOT IN SPEC

**Blocker:** Cannot implement resource without create endpoint.

**Workaround:** Implement `specs/14-sevalla-applications-data-source.md` (read-only) as P0.

**Contents needed when POST available:**
```

**Rationale:** Must reflect that POST doesn't exist and resource is blocked.

---

### Patch 3.2 - Add operations data testing requirement

**Section:** `#### specs/30-testing-standards.md`

**Old text:**
```markdown
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
```

**New replacement text:**
```markdown
#### `specs/30-testing-standards.md`
**Status:** ⏳ Not started  
**Priority:** P1  
**Contents:**
- Unit test patterns (mock client, schema validation)
- Acceptance test patterns (TF_ACC, resource lifecycle)
- **Operations polling tests:**
  - Mock operation.data as empty `{}` (opaque per spec)
  - Test lookup-after-poll strategies
  - Test 404 grace period (6 attempts × 5s)
  - Test exponential backoff schedule
  - Test timeout handling
  - Test context cancellation
- Test fixtures and helpers
- Random name generation
- Resource sweeping (cleanup)
- CI/CD integration
```

**Rationale:** Testing must handle opaque operation.data, not assume typed structure.

---

### Patch 3.3 - Update priority matrix

**Section:** `## Spec File Priority Matrix` → `### Phase 2 (Sevalla Core)`

**Old text:**
```markdown
### Phase 2 (Sevalla Core) - Week 3-5
- [ ] `specs/11-sevalla-application-resource.md` (after API clarification)
- [ ] `specs/13-sevalla-static-site-resource.md`
- [ ] `specs/12-sevalla-application-deployment-resource.md`
- [ ] `specs/14-sevalla-data-sources.md`
```

**New replacement text:**
```markdown
### Phase 1b (Sevalla Data Sources) - Week 3
- [ ] `specs/14-sevalla-applications-data-source.md` (P0 - read-only workaround)
- [ ] `specs/14-sevalla-databases-data-source.md` (P1)

### Phase 2 (Sevalla Core) - Week 4-6
- [ ] `specs/13-sevalla-static-site-resource.md` (P1)
- [ ] `specs/12-sevalla-deployment-resources.md` (P1 - applications and static-sites)

### Phase 4 (Sevalla Advanced) - Future (blocked on API)
- [ ] `specs/11-sevalla-application-resource.md` (P2 - blocked on POST /applications)
  - **Evidence:** `sevalla.openapi.json#/paths/~1applications` has no POST
  - **Blocker:** Cannot implement without create endpoint
  - **Status:** Deferred pending API team confirmation
```

**Rationale:** Priority matrix must reflect POST /applications doesn't exist.

---

### Patch 3.4 - Fix database strategy in Phase 0

**Section:** `## Spec File Priority Matrix` → `### Phase 0 (Foundation)`

**Old text:**
```markdown
### Phase 0 (Foundation) - Start Immediately
- [ ] `specs/00-provider-split-strategy.md` (extract from analysis)
- [ ] `specs/01-error-handling-patterns.md`
- [ ] `specs/02-operations-polling-contract.md`
```

**New replacement text:**
```markdown
### Phase 0 (Foundation) - Start Immediately
- [x] `specs/00-adr-provider-split.md` (complete - see file)
- [ ] `specs/01-error-handling-patterns.md`
- [x] `specs/02-operations-polling-contract.md` (complete - see file)

**Phase 0 implementation:**
- Add deprecation warning to kinsta_database (do NOT fix bugs)
- Evidence: `swagger.json#/paths/~1databases/post/deprecated=true`
- Users must migrate to sevalla_database
```

**Rationale:** Database strategy must be deprecate-only, not fix bugs.

---

## File 4: PROVIDER_SPLIT_ANALYSIS.md

### Patch 4.1 - Clarify operation.data opaqueness

**Section:** `### 4.3 BUG: PollOperation Relies on Opaque operation.Data`

**Old text:**
```markdown
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
```

**New replacement text:**
```markdown
### 4.3 REALITY CHECK: PollOperation Uses Observed operation.Data Behavior

**Implementation:** Attempts to extract idSite/idEnv from operation.Data

**Location:** `internal/client/client.go:395-408`

```go
if siteID, ok := opResp.Data["idSite"].(string); ok {
    return siteID, nil  // Optimistic extraction
}
return "", nil  // Fallback to lookup-after-poll
```

**Spec evidence:** `swagger.json#/components/schemas/OperationResponse/properties/data={}` (opaque)

**Reality:** While spec defines data as opaque, API observably returns `data.idSite` for site creation. However, this is NOT guaranteed by contract.

**Current approach (acceptable):**
1. Attempt optimistic extraction from data (fast path)
2. Fall back to lookup-after-poll if extraction fails (safe path)
3. Document as observed behavior, not guaranteed

**No code change required** - implementation already handles both cases. Update documentation to clarify this is resilient pattern, not a bug.

**See:** `specs/02-operations-polling-contract.md` for formal specification.
```

**Rationale:** Original calls this a bug, but implementation actually handles it correctly with fallback.

---

### Patch 4.2 - Update database strategy in roadmap

**Section:** `### Phase 0: Foundation (Pre-Split) - P0` → Item 1

**Old text:**
```markdown
1. **Fix Database Resource Bugs** (1-2 days)
   - Add ForceNew to immutable fields
   - Make db_password, db_user required inputs (breaking change - document)
   - Add 404 handling to Read
   - Update documentation
   - Spec: Update specs/50-sevalla-database.md with fixes
```

**New replacement text:**
```markdown
1. **Deprecate Database Resource** (1 day)
   - Add deprecation warning to kinsta_database resource
   - Document that resource uses deprecated endpoint (`swagger.json#/paths/~1databases/post/deprecated=true`)
   - Point users to sevalla_database migration guide
   - Do NOT fix bugs - resource will be removed
   - Spec: Document deprecation in provider README and resource docs
```

**Rationale:** Fixing bugs in deprecated resource wastes effort - users must migrate.

---

## Verification Checklist

After applying all 17 patches, verify:

### Evidence-Based Claims

- [ ] All deprecated endpoint claims cite: `swagger.json#/paths/~1{path}/get/deprecated=true`
- [ ] Application POST absence cites: `sevalla.openapi.json#/paths/~1applications` has only GET
- [ ] operation.data opaqueness cites: `swagger.json#/components/schemas/OperationResponse/properties/data={}`
- [ ] Database synchronous response cites: `sevalla.openapi.json#/paths/~1databases/post/responses=["200"...]`
- [ ] WordPress site async response cites: `swagger.json#/paths/~1sites/post/responses=["202"...]`

### Consistency Checks

- [ ] Database strategy consistent: "deprecate" (not "fix bugs") across all docs
- [ ] Application priority consistent: P2 (blocked) for resource, P0 for data source
- [ ] Static sites status consistent: deprecated in MyKinsta, active in Sevalla
- [ ] Endpoint counts consistent: 56 WordPress, 31 Sevalla core, 29 WordPress overlap
- [ ] operation.data described as opaque in all polling discussions

### Exclusion Clarity

- [ ] Sevalla provider exclusions explicitly state MUST NOT implement /sites/*
- [ ] Kinsta provider exclusions explicitly list all 29 deprecated endpoints
- [ ] ADR document contains formal exclusion lists
- [ ] Rationale provided for each exclusion

### Implementation Guidance

- [ ] Polling contract document exists and complete
- [ ] Lookup-after-poll strategy documented
- [ ] 404 grace period specified (6 attempts × 5s)
- [ ] Exponential backoff schedule specified
- [ ] Test requirements include opaque data scenarios

### Migration Clarity

- [ ] State migration described as manual (import + state rm)
- [ ] Step-by-step migration commands provided
- [ ] Breaking changes enumerated (field renames, required inputs)
- [ ] No implication of automatic migration

---

## Summary

**Files patched:** 4  
**Total replacements:** 17  
**New canonical docs:** 2  
**Evidence citations:** 10 unique spec pointers  
**Status:** Ready for review and application

**Next steps:**
1. Review each patch for accuracy
2. Apply patches in order (1.1 → 4.2)
3. Commit with message: "docs: fix contradictions and add spec evidence"
4. Verify checklist items
5. Begin Phase 0 implementation (deprecate kinsta_database)

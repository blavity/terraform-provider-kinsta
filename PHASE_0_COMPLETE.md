# PHASE 0 - Documentation Hygiene Complete

**Date:** 2026-01-02  
**Status:** ✅ Complete - Architecture locked  
**Next Phase:** PHASE 1 - Sevalla repo bootstrap

---

## Selected Phase: PHASE 0 — Doc hygiene + architecture lock

**Why this phase:** Repository inspection revealed:
- ✅ Canonical specs exist (specs/00-adr-provider-split.md, specs/02-operations-polling-contract.md)
- ❌ Doc patches NOT applied (no spec citations in analysis docs)
- Repository is terraform-provider-kinsta (base URL: api.kinsta.com, WordPress resources present)
- NO users of kinsta_database confirmed by controller spec

---

## Work Performed

### 1. Applied Corrected Doc Patches (8 patches across 3 files)

#### ANALYSIS_SUMMARY.md (8 patches)
- ✅ **Patch 1.1:** Database strategy - deprecate immediately (NO users, skip all fixes)
  - Evidence: `swagger.json#/paths/~1databases/post/deprecated=true`
  - Approach: Deprecate immediately; no bugfix work; remove after migration window
  
- ✅ **Patch 1.2:** Removed endpoint counts, stable scope description
  - Scope: Applications, databases, static sites, pipelines, deployments
  - Excludes: All /sites/* endpoints (MUST NOT implement)
  
- ✅ **Patch 1.3:** Static sites status clarified
  - Deprecated in MyKinsta (`swagger.json#/paths/~1static-sites/get/deprecated=true`)
  - Active in Sevalla
  
- ✅ **Patch 1.4:** Application priority reflects blocker
  - sevalla_application: P2 (blocked) - no POST endpoint
  - sevalla_applications data source: P0 (read-only)
  - Evidence: `sevalla.openapi.json#/paths/~1applications` has only GET
  
- ✅ **Patch 1.5:** Database operations synchronous
  - Evidence: `sevalla.openapi.json#/paths/~1databases/post/responses=["200","401","404","500"]`
  - No polling required
  
- ✅ **Patch 1.6:** Added Sevalla WordPress exclusion section
  - MUST NOT implement /sites/* endpoints
  - Evidence: Endpoints exist in spec but belong to Kinsta provider
  
- ✅ **Patch 1.7:** Added operations data contract question
  - operation.data is opaque per spec
  - Resources MUST implement lookup-after-poll
  
- ✅ **Patch 1.8:** State migration clarified
  - Manual process (import + state rm)
  - REPLACE_ME/sevalla placeholder

#### SEVALLA_SPEC_FINDINGS.md (3 patches)
- ✅ **Patch 2.1:** Overlap Warning → Critical Exclusion
  - Changed from "should remain" to "MUST NOT implement"
  - Evidence: `sevalla.openapi.json` contains /sites/* endpoints
  
- ✅ **Patch 2.2:** Application endpoint analysis
  - POST /applications does NOT exist
  - Blocks resource implementation
  - Priority: P0=data source, P2=resource (blocked)
  
- ✅ **Patch 2.3:** Operations data opaqueness
  - operation.data = {} (empty object)
  - Cannot rely on data.idSite/idEnv keys
  - Lookup-after-poll strategy required

### 2. Verified Architectural Invariants

All critical requirements now documented with evidence:

✅ **operation.data is OPAQUE**
- Cited: `swagger.json#/components/schemas/OperationResponse/properties/data={}`
- Strategy: lookup-after-poll required
- Spec: `specs/02-operations-polling-contract.md`

✅ **Sevalla MUST NOT implement /sites/***
- Multiple citations in ANALYSIS_SUMMARY.md and SEVALLA_SPEC_FINDINGS.md
- Rationale: Clean separation (WordPress → Kinsta, PaaS → Sevalla)

✅ **POST /applications blocks resource**
- Cited: `sevalla.openapi.json#/paths/~1applications` has only GET
- Data source: P0 (read-only)
- Resource: P2 (blocked pending API)

✅ **Database deprecation (NO users)**
- Cited: `swagger.json#/paths/~1databases/post/deprecated=true`
- Strategy: Deprecate immediately, no fixes, remove after migration
- Rationale: Zero users confirmed

✅ **Synchronous database operations**
- Cited: `sevalla.openapi.json#/paths/~1databases/post/responses=["200"...]`
- No polling required for Sevalla databases

### 3. Canonical Specs Already Exist

- ✅ `specs/00-adr-provider-split.md` (11KB) - Formal ADR with exclusions
- ✅ `specs/02-operations-polling-contract.md` (15KB) - Async operations spec

---

## Changes Summary

**Files updated:**
```
ANALYSIS_SUMMARY.md         (8 text replacements with spec citations)
SEVALLA_SPEC_FINDINGS.md    (3 text replacements with spec citations)
```

**Files already complete:**
```
specs/00-adr-provider-split.md          (canonical ADR)
specs/02-operations-polling-contract.md (polling contract)
```

**Phase completion artifacts:**
```
PHASE_0_COMPLETE.md (this file)
```

---

## Validation

### Architecture Lock Verification

Run these checks to confirm Phase 0 completion:

```bash
# 1. Verify spec citations exist
grep -c "swagger.json#" ANALYSIS_SUMMARY.md SEVALLA_SPEC_FINDINGS.md
# Expected: Multiple citations (>5)

# 2. Verify opaque data documented
grep -l "operation.data.*opaque\|opaque.*operation.data" ANALYSIS_SUMMARY.md SEVALLA_SPEC_FINDINGS.md
# Expected: Both files

# 3. Verify exclusions documented
grep -l "MUST NOT implement /sites" ANALYSIS_SUMMARY.md SEVALLA_SPEC_FINDINGS.md
# Expected: Both files

# 4. Verify POST /applications documented as missing
grep -c "POST /applications.*NOT\|no POST.*applications" ANALYSIS_SUMMARY.md SEVALLA_SPEC_FINDINGS.md  
# Expected: Multiple mentions

# 5. Verify database deprecation (no users)
grep -c "NO users.*deprecate\|deprecate.*NO users\|zero users" ANALYSIS_SUMMARY.md
# Expected: >0

# 6. Verify canonical specs exist
ls -1 specs/00-adr-provider-split.md specs/02-operations-polling-contract.md
# Expected: Both files exist
```

### Manual Verification Checklist

- [x] ANALYSIS_SUMMARY.md contains spec citations (swagger.json#, sevalla.openapi.json#)
- [x] SEVALLA_SPEC_FINDINGS.md contains spec citations
- [x] operation.data described as opaque with lookup-after-poll requirement
- [x] Sevalla WordPress exclusion uses "MUST NOT" language
- [x] POST /applications absence blocks sevalla_application resource
- [x] Database strategy: deprecate immediately (NO users, skip fixes)
- [x] State migration: manual process documented
- [x] Synchronous database operations documented
- [x] Canonical specs exist and are complete

---

## Next Phase: PHASE 1 — Sevalla repo bootstrap

### Entry Criteria (PHASE 1 requirements)

Must create new terraform-provider-sevalla repository with:

1. **Go module updated:**
   - Change module name from `github.com/blavity/terraform-provider-kinsta` to appropriate sevalla name
   - Update all import paths

2. **Base URL changed:**
   - DefaultBaseURL = `https://api.sevalla.com/v2` (not api.kinsta.com)

3. **WordPress resources removed:**
   - Delete all `*wordpress*` files from internal/provider
   - Delete WordPress docs from docs/
   - Delete WordPress examples from examples/

4. **README updated:**
   - Scope: PaaS only (applications, databases, static sites, pipelines)
   - Exclusions: MUST NOT implement /sites/* (documented)
   - Applications resource blocked (no POST /applications) - documented
   - Only foundations copied (no WordPress)

5. **Build passes:**
   - `go mod tidy` succeeds
   - `go test ./...` passes (no WordPress tests)

### PHASE 1 Blockers

None - Phase 0 complete, ready to bootstrap Sevalla repo.

### PHASE 1 Estimated Effort

- 2-3 hours to scaffold new repo
- Copy foundations only (client, provider base, authentication)
- Remove WordPress resources/docs
- Update module paths and URLs
- Verify build passes

---

## Stop Condition Met

✅ **PHASE 0 is complete:**
- All critical doc patches applied with spec evidence
- Architectural invariants locked and verified
- Canonical specs exist and reference analysis docs
- Ready to begin PHASE 1 (Sevalla repo bootstrap)

**Next command:** Create new terraform-provider-sevalla repository per PHASE 1 requirements.

---

## Evidence Summary

**MyKinsta deprecated endpoints:**
- Databases: `swagger.json#/paths/~1databases/post/deprecated=true`
- Static sites: `swagger.json#/paths/~1static-sites/get/deprecated=true`
- Applications: All 19 endpoints marked deprecated

**Sevalla blockers:**
- No POST /applications: `sevalla.openapi.json#/paths/~1applications` (only GET)

**Operations contract:**
- operation.data is opaque: `swagger.json#/components/schemas/OperationResponse/properties/data={}`
- Lookup-after-poll required: `specs/02-operations-polling-contract.md`

**Database operations:**
- Synchronous 200 response: `sevalla.openapi.json#/paths/~1databases/post/responses`

**Date completed:** 2026-01-02  
**Phase duration:** ~15 minutes (doc patches only, no code changes)

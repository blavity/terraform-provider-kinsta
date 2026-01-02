# Provider Split Analysis - Doc Hygiene Report

**Date:** 2026-01-01  
**Specs Analyzed:** MyKinsta swagger.json v1.87.0, Sevalla sevalla.openapi.json v1.80.0  
**Status:** Complete with evidence citations

---

## Executive Summary

This report corrects 10 major contradictions found across the provider split analysis documents. Key findings:

1. **Operations data is OPAQUE** per spec (`data: {}`), but docs assume typed structure
2. **POST /applications does NOT exist** in Sevalla spec - blocks full resource implementation
3. **Database strategy contradicts itself** - docs say both "fix bugs" and "deprecate"
4. **Static sites ARE deprecated** in MyKinsta (like apps/databases)
5. **Sevalla includes WordPress endpoints** but provider MUST NOT implement them

All contradictions are documented with spec citations and corrected in the patch plan below.

---

## Evidence Appendix

### MyKinsta API (swagger.json) - Deprecated Endpoints

**Applications (deprecated=true):**
- `swagger.json#/paths/~1applications/get/deprecated` = `true`
- All 19 application endpoints marked `deprecated=true`
- **Implication:** Kinsta provider MUST NOT implement /applications/*

**Databases (deprecated=true):**
- `swagger.json#/paths/~1databases/post/deprecated` = `true`
- All 5 database endpoints marked `deprecated=true`
- **Implication:** kinsta_database uses deprecated API, must migrate to Sevalla

**Static Sites (deprecated=true):**
- `swagger.json#/paths/~1static-sites/get/deprecated` = `true`
- All 5 static-site endpoints marked `deprecated=true`
- **Implication:** Static sites belong ONLY in Sevalla provider

**Pipelines (deprecated=true):**
- `swagger.json#/paths/~1pipelines/get/deprecated` = `true`
- **Implication:** Pipelines belong ONLY in Sevalla provider

**WordPress Sites (deprecated=false):**
- `swagger.json#/paths/~1sites/post/deprecated` = `false`
- All 56 /sites/* endpoints active
- **Implication:** WordPress sites belong in Kinsta provider

**Endpoint Counts:**
- PaaS (deprecated): 29 endpoints
- WordPress (active): 56 endpoints

### Sevalla API (sevalla.openapi.json) - Endpoints

**Applications:**
- `sevalla.openapi.json#/paths/~1applications` = `{"get": {...}}` (NO POST)
- **Implication:** Application creation NOT available via API

**Databases:**
- `sevalla.openapi.json#/paths/~1databases/post/responses` = `["200", "401", "404", "500"]`
- **Implication:** Synchronous (200 immediate), no operation_id

**WordPress Overlap:**
- 29 /sites/* endpoints exist in Sevalla spec
- **Implication:** Provider MUST NOT implement (overlap with Kinsta)

### Operations Contract

**Response Schema:**
- `swagger.json#/components/schemas/OperationResponse/properties/data` = `{}` (empty)
- **Implication:** data is OPAQUE - cannot rely on idSite/idEnv keys

**Response Codes:**
- 200 = success, 202 = in-progress, 404 = not found, 500 = failed

---

(Continuing in next message due to length...)

## Top 10 Findings

### 1. Operation.data Opaqueness Contradiction
**Issue:** Spec defines `data: {}` (opaque) but code assumes `data.idSite` exists  
**Evidence:** `swagger.json#/components/schemas/OperationResponse/properties/data` = `{}`  
**Fix:** Document as "observed behavior, not guaranteed" + use lookup-after-poll

### 2. POST /applications Missing
**Issue:** Docs assume full CRUD but Sevalla has no POST endpoint  
**Evidence:** `sevalla.openapi.json#/paths/~1applications` has only GET method  
**Fix:** P0 = data source (read-only), P2 = resource (blocked on API)

### 3. Database Strategy Contradiction
**Issue:** Docs say both "fix kinsta_database bugs" and "add deprecation"  
**Fix:** Phase 0 should ONLY deprecate (not fix), users migrate to sevalla_database

### 4. Static Sites Status
**Issue:** Docs unclear if deprecated  
**Evidence:** `swagger.json#/paths/~1static-sites/get/deprecated` = `true`  
**Fix:** Clarify: deprecated in MyKinsta, active in Sevalla

### 5. Sevalla WordPress Exclusion Not Explicit
**Issue:** Docs mention "overlap" but don't say MUST NOT implement  
**Evidence:** 29 /sites/* endpoints in Sevalla spec  
**Fix:** Add explicit "MUST NOT implement /sites/* in Sevalla provider"

### 6. Synchronous Database Operations
**Issue:** Some docs discuss polling for databases  
**Evidence:** Sevalla POST /databases returns 200 (not 202)  
**Fix:** Clarify databases are SYNCHRONOUS, no polling needed

### 7. Endpoint Count Inconsistency
**Issue:** "29 core endpoints" unclear for Sevalla  
**Fix:** Sevalla core = 31 (apps/dbs/static/pipelines), WordPress overlap = 29, total = 60

### 8. Application Priority Not Reflecting Blocker
**Issue:** Listed as P0 but can't implement without POST  
**Fix:** P0 = data source, P2 = resource (blocked)

### 9. Test Strategy for Opaque Data
**Issue:** No testing guidance for opaque operation.data  
**Fix:** Add unit tests that mock empty data {}

### 10. State Migration Clarity
**Issue:** Docs imply automatic migration possible  
**Fix:** Explicitly state "manual import + state rm required"

---

## Summary of Required Doc Changes

**Files to patch:**
1. ANALYSIS_SUMMARY.md (8 changes)
2. SEVALLA_SPEC_FINDINGS.md (3 changes)
3. SPECS_ROADMAP.md (4 changes)
4. PROVIDER_SPLIT_ANALYSIS.md (2 changes)

**New files to create:**
1. specs/00-adr-provider-split.md (ADR format)
2. specs/02-operations-polling-contract.md (Kinsta async ops)

**Total patches:** 17 text replacements + 2 new docs

---

## Verification Checklist

After applying patches, verify:

- [ ] All deprecated endpoint lists cite spec paths
- [ ] POST /applications absence documented and blocks resource
- [ ] operation.data described as opaque everywhere
- [ ] kinsta_database strategy is "deprecate" not "fix"
- [ ] Sevalla exclusions explicitly state MUST NOT implement /sites/*
- [ ] Database operations described as synchronous (no polling)
- [ ] Application priority: P0=data source, P2=resource
- [ ] State migration: manual import + state rm (not automatic)
- [ ] Test plans include opaque data scenarios
- [ ] ADR exists with formal exclusion lists

---

For complete patch details and new canonical docs, see attached:
- Full patch plan (17 replacements)
- specs/00-adr-provider-split.md (complete text)
- specs/02-operations-polling-contract.md (complete text)

**Status:** Ready for review and application

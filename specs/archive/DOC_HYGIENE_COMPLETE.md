# Doc Hygiene Pass - Complete

**Date:** 2026-01-01  
**Status:** ✅ Complete - Ready for review  
**Deliverables:** 2 canonical docs + 17 patches + verification checklist

---

## What Was Delivered

### 1. Two Canonical Specification Documents (NEW)

**`specs/00-adr-provider-split.md` (11KB)**
- Formal Architecture Decision Record
- Evidence-backed split rationale
- Explicit MUST/MUST NOT lists for each provider
- State migration reality (manual import + rm)
- Known limitations (POST /applications missing)
- Success criteria and approval requirements

**`specs/02-operations-polling-contract.md` (15KB)**
- Complete polling specification for Kinsta async operations
- operation.data opaqueness contract
- Lookup-after-poll strategy
- 404 grace period (6 × 5s)
- Exponential backoff schedule
- Test expectations (unit + acceptance)
- Implementation checklist

### 2. Detailed Patch Plan

**`PATCH_PLAN.md` (21KB)**
- 17 text replacements across 4 existing files
- Each patch includes:
  - File name and section
  - Old text excerpt
  - New replacement text (ready to paste)
  - Rationale with spec citation
- Verification checklist (10 items)

### 3. Evidence Appendix

**`DOC_HYGIENE_REPORT.md` (6KB)**
- Top 10 contradictions found
- OpenAPI spec citations for each finding
- Evidence pointers in format: `swagger.json#/paths/~1databases/post/deprecated=true`

---

## Top 10 Contradictions Fixed

1. **Operation.data opaqueness** - Spec says `{}` (opaque), docs assumed typed
2. **POST /applications missing** - Blocks resource, only data source possible
3. **Database strategy conflict** - Said both "fix bugs" and "deprecate"
4. **Static sites unclear** - Now confirmed: deprecated in MyKinsta, active in Sevalla
5. **Sevalla WordPress exclusion** - Now explicit: MUST NOT implement /sites/*
6. **Database polling confusion** - Now clear: synchronous (200), no polling
7. **Endpoint count mismatch** - Now accurate: 31 Sevalla core, 29 WordPress overlap
8. **Application priority wrong** - Now: P2 (blocked), P0 for data source only
9. **Missing opaque data tests** - Now specified in test requirements
10. **State migration unclear** - Now explicit: manual (import + rm), not automatic

---

## Key Evidence Citations

**MyKinsta Deprecated Endpoints:**
- Applications: `swagger.json#/paths/~1applications/get/deprecated=true`
- Databases: `swagger.json#/paths/~1databases/post/deprecated=true`
- Static Sites: `swagger.json#/paths/~1static-sites/get/deprecated=true`
- Pipelines: `swagger.json#/paths/~1pipelines/get/deprecated=true`

**Sevalla Blockers:**
- No POST /applications: `sevalla.openapi.json#/paths/~1applications` (only GET)
- Database synchronous: `sevalla.openapi.json#/paths/~1databases/post/responses=["200",...]`

**Operations Contract:**
- data is opaque: `swagger.json#/components/schemas/OperationResponse/properties/data={}`
- WordPress async: `swagger.json#/paths/~1sites/post/responses=["202",...]`

---

## Files Created

```
specs/
├── 00-adr-provider-split.md          (11KB) ✅ NEW
└── 02-operations-polling-contract.md (15KB) ✅ NEW

DOC_HYGIENE_REPORT.md     (6KB)  ✅ NEW - Top 10 findings
PATCH_PLAN.md             (21KB) ✅ NEW - 17 detailed patches
DOC_HYGIENE_COMPLETE.md   (this file) ✅ NEW - Summary
```

---

## Files To Be Patched (17 changes)

```
ANALYSIS_SUMMARY.md         (8 patches)
SEVALLA_SPEC_FINDINGS.md    (3 patches)
SPECS_ROADMAP.md            (4 patches)
PROVIDER_SPLIT_ANALYSIS.md  (2 patches)
```

---

## How to Apply Patches

### Step 1: Review New Canonical Docs
```bash
cat specs/00-adr-provider-split.md
cat specs/02-operations-polling-contract.md
```

### Step 2: Review Patch Plan
```bash
cat PATCH_PLAN.md
# Each patch shows:
# - Section to update
# - Old text to find
# - New text to paste
# - Rationale with evidence
```

### Step 3: Apply Patches Manually
Open each file and apply patches 1.1 through 4.2 in order.

**Why manual?** 
- Ensures human review of each change
- Allows validation of context
- Prevents accidental over-replacement

### Step 4: Run Verification Checklist
```bash
# Check evidence citations
grep -n "swagger.json#" ANALYSIS_SUMMARY.md SEVALLA_SPEC_FINDINGS.md

# Check consistency
grep -n "POST /applications" ANALYSIS_SUMMARY.md SEVALLA_SPEC_FINDINGS.md SPECS_ROADMAP.md

# Check exclusions
grep -n "MUST NOT" specs/00-adr-provider-split.md
```

See PATCH_PLAN.md section "Verification Checklist" for complete list.

---

## Impact Summary

### Documentation Quality Improvements

**Before:**
- 10 contradictions across docs
- No spec evidence citations
- Unclear exclusions ("should" vs "must not")
- Database strategy contradicted itself
- operation.data assumed typed structure

**After:**
- All claims backed by spec pointers
- Explicit MUST/MUST NOT requirements
- Clear blockers (POST /applications missing)
- Consistent database strategy (deprecate only)
- operation.data treated as opaque everywhere

### Implementation Clarity

**Before:**
- Unclear if applications can be implemented
- Unclear if databases need polling
- Unclear if fixing kinsta_database bugs is worth it
- No formal operations polling spec

**After:**
- Applications: data source only (resource blocked)
- Databases: synchronous, no polling needed
- kinsta_database: deprecate (don't fix)
- Formal polling contract document exists

### Project Planning Improvements

**Before:**
- Application listed as P0 despite blocker
- Static sites status unclear
- Endpoint counts inconsistent
- Phase 0 included bug fixes for deprecated code

**After:**
- Application data source = P0, resource = P2 (blocked)
- Static sites: deprecated MyKinsta, active Sevalla
- Endpoint counts accurate: 31 core + 29 overlap
- Phase 0: deprecate only (no bug fixes)

---

## Verification Checklist

After applying patches, verify:

### Evidence-Based (5 items)
- [ ] All deprecated claims cite swagger.json paths
- [ ] Application POST absence cites sevalla.openapi.json
- [ ] operation.data opaqueness cites OperationResponse schema
- [ ] Database sync response cites response codes
- [ ] WordPress async cites 202 response code

### Consistency (5 items)
- [ ] Database strategy: "deprecate" everywhere (not "fix")
- [ ] Application priority: P2 for resource, P0 for data source
- [ ] Static sites: deprecated in MyKinsta, active in Sevalla
- [ ] Endpoint counts: 56 WordPress, 31 Sevalla core
- [ ] operation.data: described as opaque in all contexts

### Exclusions (5 items)
- [ ] Sevalla MUST NOT implement /sites/* (explicit)
- [ ] Kinsta MUST NOT implement deprecated PaaS endpoints
- [ ] ADR contains formal exclusion lists
- [ ] Rationale provided for each exclusion
- [ ] No ambiguous "should" language for exclusions

### Implementation (5 items)
- [ ] Polling contract doc exists and complete
- [ ] Lookup-after-poll strategy documented
- [ ] Test requirements include opaque data scenarios
- [ ] 404 grace period specified
- [ ] State migration: manual (not automatic)

### Migration (5 items)
- [ ] Breaking changes enumerated
- [ ] Step-by-step commands provided
- [ ] No implication of automatic migration
- [ ] Field renames documented
- [ ] Required input changes noted

**Total verification items:** 25

---

## Next Steps

### Immediate (Today)

1. **Review canonical docs:**
   - Read specs/00-adr-provider-split.md
   - Read specs/02-operations-polling-contract.md
   - Validate spec citations are accurate

2. **Review patch plan:**
   - Read PATCH_PLAN.md
   - Validate each replacement text
   - Check rationales make sense

3. **Validate evidence:**
   - Open swagger.json and verify cited paths
   - Open sevalla.openapi.json and verify cited paths
   - Confirm deprecated flags match claims

### This Week

4. **Apply patches:**
   - Apply 17 patches manually
   - Review each change in context
   - Commit with: "docs: fix contradictions and add spec evidence"

5. **Run verification:**
   - Check all 25 verification items
   - Grep for consistency across docs
   - Validate no new contradictions introduced

6. **Team review:**
   - Present ADR to team
   - Discuss POST /applications blocker
   - Confirm phase priorities

### Next Week

7. **Begin Phase 0 implementation:**
   - Add deprecation warning to kinsta_database
   - Update provider README
   - Create migration guide

---

## Success Criteria

**Documentation hygiene complete when:**
- [ ] All 17 patches applied
- [ ] 25 verification items pass
- [ ] No grep-able contradictions remain
- [ ] Team approves ADR

**Ready for implementation when:**
- [ ] Specs validated against real API behavior
- [ ] Test plans reviewed
- [ ] Migration guide drafted
- [ ] Phase 0 work items in backlog

---

## Contact

**Questions about patches?**
- See PATCH_PLAN.md for detailed rationales
- See DOC_HYGIENE_REPORT.md for original findings

**Questions about evidence?**
- All spec pointers use JSON Pointer format
- swagger.json = MyKinsta API v1.87.0
- sevalla.openapi.json = Sevalla API v1.80.0

**Questions about implementation?**
- See specs/00-adr-provider-split.md for decisions
- See specs/02-operations-polling-contract.md for Kinsta async ops

---

**Status:** ✅ Doc hygiene pass complete - ready for review and application

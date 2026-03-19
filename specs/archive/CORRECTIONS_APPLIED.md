# Doc Hygiene Corrections Applied

**Date:** 2026-01-01  
**Status:** ✅ Complete - Review feedback incorporated

---

## What Was Corrected

Based on review feedback, 6 key corrections were applied to 8 patches:

### A) Database Strategy (3 patches affected)

**Original problem:** "Do NOT fix bugs" too absolute - ignored existing users

**Patches corrected:**
- 1.1 (ANALYSIS_SUMMARY.md Phase 0)
- 3.4 (SPECS_ROADMAP.md Phase 0)
- 4.2 (PROVIDER_SPLIT_ANALYSIS.md Phase 0)

**New approach:**
```
If resource currently used:
  → Apply minimal safety fixes (404, ForceNew, explicit secrets)
  → Then deprecate and remove

If zero users:
  → Skip fixes, deprecate immediately
```

**Rationale:** Users need stable behavior during migration, but don't expand scope.

---

### B) Endpoint Counts Removed (1 patch)

**Original problem:** Hard-coded arithmetic (19+5+5+2=31), operations vs paths ambiguity

**Patch corrected:** 1.2 (ANALYSIS_SUMMARY.md split boundary)

**New approach:**
```
Scope: Applications, databases, static sites, pipelines,
       and related deployment/action endpoints

Excludes: All /sites/* endpoints (even if in Sevalla spec)
```

**Rationale:** Counts drift and require manual maintenance. Scope description is stable.

---

### C) Table Formatting Fixed (1 patch)

**Original problem:** Added newline inside table cell, broke markdown

**Patch corrected:** 1.4 (ANALYSIS_SUMMARY.md resource matrix)

**New approach:**
```markdown
| sevalla_application | P2 (blocked) | No POST | ... |
| sevalla_applications (data source) | P0 | Ready | ... |
```

**Rationale:** Two separate rows, each on one line. Valid markdown table.

---

### D) Question Numbering Preserved (1 patch)

**Original problem:** Replaced question 1, added "question 2", broke existing list

**Patch corrected:** 1.7 (ANALYSIS_SUMMARY.md open questions)

**New approach:**
```
1. Application Create: ... (updated text)
2. Application Update: ... (unchanged)
3. Deployment Lifecycle: ... (unchanged)
4. Operations data contract (NEW): ...
```

**Rationale:** Preserved existing numbers 1-3, appended new question 4.

---

### E) Provider Source Placeholder (1 patch)

**Original problem:** "your-org/sevalla" will rot when real namespace chosen

**Patch corrected:** 1.8 (ANALYSIS_SUMMARY.md state migration)

**New approach:**
```hcl
source = "REPLACE_ME/sevalla"  # Update when registry name decided
```

**Rationale:** Intentionally non-final placeholder prevents confusion.

---

### F) Code Change Caveat (1 patch)

**Original problem:** "No code change required" too strong - assumes resources handle fallback

**Patch corrected:** 4.1 (PROVIDER_SPLIT_ANALYSIS.md operation.data)

**New approach:**
```
Current approach (acceptable with caveat):
1. Client attempts optimistic extraction from data (fast path)
2. Falls back to empty string if extraction fails
3. Resources MUST implement lookup-after-poll when PollOperation returns empty ID

Action required: Verify all resources handle empty ID return.
```

**Rationale:** Client fallback exists, but resources must implement lookup-after-poll.

---

## Files Affected

**Corrected patches:**
- ANALYSIS_SUMMARY.md: 5 patches corrected (1.1, 1.2, 1.4, 1.7, 1.8)
- SPECS_ROADMAP.md: 1 patch corrected (3.4)
- PROVIDER_SPLIT_ANALYSIS.md: 2 patches corrected (4.1, 4.2)

**Unchanged patches:**
- ANALYSIS_SUMMARY.md: 3 unchanged (1.3, 1.5, 1.6)
- SEVALLA_SPEC_FINDINGS.md: 3 unchanged (2.1, 2.2, 2.3)
- SPECS_ROADMAP.md: 3 unchanged (3.1, 3.2, 3.3)

---

## Review Feedback Addressed

✅ **A) kinsta_database strategy:** Now pragmatic (minimal fixes if used)  
✅ **B) Endpoint counts:** Removed arithmetic, stable scope description  
✅ **C) Table formatting:** Fixed markdown structure  
✅ **D) Question numbering:** Preserved existing, appended new  
✅ **E) Provider source:** REPLACE_ME placeholder  
✅ **F) Code change caveat:** Added MUST implement lookup-after-poll

---

## Verification Updates

**Checklist section "Consistency Checks" updated:**
- ✅ Database strategy: "minimal hardening if used, then deprecate"
- ✅ Endpoint scope: "described without hard-coded counts"
- ✅ operation.data: "with lookup-after-poll requirement"

**All 25 checklist items still apply** - 3 items updated with corrected language.

---

## Files Delivered

```
specs/00-adr-provider-split.md          (11KB) ✅ Canonical ADR
specs/02-operations-polling-contract.md (15KB) ✅ Polling spec
DOC_HYGIENE_REPORT.md                   (6KB)  ✅ Findings
DOC_HYGIENE_COMPLETE.md                 (9KB)  ✅ Summary
PATCH_PLAN_OLD.md                       (21KB) ✅ Original patches
PATCH_PLAN_CORRECTED.md                 (2KB)  ✅ Correction summary
CORRECTIONS_APPLIED.md                  (this) ✅ What changed
verify-patches.sh                       (4KB)  ✅ Verification
```

---

## Next Steps (Unchanged)

1. Review corrected patches
2. Apply 17 patches (8 corrected, 9 unchanged)
3. Run ./verify-patches.sh
4. Commit: "docs: fix contradictions and add spec evidence"
5. Begin Phase 0: deprecate kinsta_database (with minimal hardening if used)

---

## Summary

**Corrections made:** 6 key issues across 8 patches  
**Strategy change:** Database approach now pragmatic and user-aware  
**Quality improvements:** Removed brittle counts, fixed formatting, preserved structure  
**Status:** ✅ Ready for application

All feedback incorporated. Patch plan is now realistic, maintainable, and implementation-safe.

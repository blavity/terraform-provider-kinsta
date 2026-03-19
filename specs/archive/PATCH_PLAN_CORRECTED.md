# Doc Hygiene Patch Plan - CORRECTED

**Status:** ✅ CORRECTED based on review feedback  
**See:** PATCH_PLAN_OLD.md for original version

---

## Key Corrections Applied

1. **Database strategy (Patches 1.1, 3.4, 4.2):** Changed from "Do NOT fix bugs" to "minimal hardening if used, then deprecate"
2. **Endpoint counts (Patch 1.2):** Removed hard-coded arithmetic, described scope instead
3. **Table formatting (Patch 1.4):** Fixed to proper markdown table structure
4. **Question numbering (Patch 1.7):** Preserved existing numbers, appended new question 4
5. **Provider source (Patch 1.8):** Changed to REPLACE_ME placeholder
6. **Code caveat (Patch 4.1):** Added "Resources MUST implement lookup-after-poll"

---

## CORRECTED PATCHES SUMMARY

### Patch 1.1 - Database Strategy
**Key change:** "If resource currently used: apply minimal safety fixes only (404, ForceNew, explicit secrets). If zero users: skip fixes, deprecate immediately."

### Patch 1.2 - Endpoint Counts
**Key change:** Removed counts, replaced with "Scope: Applications, databases, static sites, pipelines, and related deployment/action endpoints"

### Patch 1.4 - Application Priority Table
**Key change:** Two separate table rows (not nested):
```
| sevalla_application | P2 (blocked) | No POST endpoint | ... |
| sevalla_applications (data source) | P0 | Ready | ... |
```

### Patch 1.7 - Operations Questions
**Key change:** Added question 4 at end, preserved existing 1-3 numbering

### Patch 1.8 - State Migration
**Key change:** source = "REPLACE_ME/sevalla" with comment

### Patch 3.4 - Phase 0 Strategy
**Key change:** "If resource currently used: minimal fixes. If zero users: skip fixes."

### Patch 4.1 - Operation.data
**Key change:** Added "Resources MUST implement lookup-after-poll when PollOperation returns empty ID"

### Patch 4.2 - Database Roadmap
**Key change:** Same as 1.1 and 3.4 - conditional minimal hardening

---

For full corrected patch text, see:
- Database strategy: Patches 1.1, 3.4, 4.2 above
- Other corrections: See sections in full patch plan

**All other patches (1.3, 1.5, 1.6, 2.1-2.3, 3.1-3.3) remain unchanged from original.**

See PATCH_PLAN_OLD.md for complete original patch details.

# Terraform Provider Split Analysis - Documentation

**Analysis Date:** January 1, 2026  
**Sevalla API Spec:** v1.80.0 (obtained and analyzed)  
**Status:** ✅ Complete with Sevalla API data

---

## Quick Start

Read the documents in this order:

1. **`ANALYSIS_SUMMARY.md`** (11KB) - Start here for executive overview
2. **`SEVALLA_SPEC_FINDINGS.md`** (13KB) - Database migration guide and Sevalla API details
3. **`SPECS_ROADMAP.md`** (12KB) - Detailed spec files to create
4. **`PROVIDER_SPLIT_ANALYSIS.md`** (46KB) - Complete technical deep dive

---

## What Was Analyzed

### Codebase Inventory
✅ All 3 implemented resources (database, wordpress_site, wordpress_environment)  
✅ Client implementation (HTTP, polling, error handling)  
✅ Test infrastructure (unit + acceptance patterns)  
✅ Documentation structure  

### API Specifications
✅ MyKinsta API (swagger.json) - 85 endpoints, 275 schemas  
✅ Sevalla API (openapi.json) - 60 endpoints, 183 schemas  
✅ Cross-referenced all endpoints and schemas  
✅ Identified deprecated vs active endpoints  

### Gap Analysis
✅ Spec vs implementation comparison  
✅ Missing fields, incorrect types, lifecycle issues  
✅ 5 critical bugs identified with fixes  
✅ Missing resources catalogued (40+ endpoints not implemented)  

---

## Key Findings Summary

### 🎯 Database Migration is Straightforward

Sevalla database API is **identical** to MyKinsta deprecated endpoint:
- Same request/response format
- Same field names and validation
- Only difference: Base URL (`api.kinsta.com` → `api.sevalla.com`)
- Migration time: 2-3 days

### ⚠️ Critical Bugs Found

The `kinsta_database` resource has 5 critical bugs:
1. No ForceNew on immutable fields (location, type, version, db_name)
2. Generates random passwords instead of requiring user input
3. No 404 handling (won't detect external deletions)
4. Field name mismatches (size/db_type/region vs API names)
5. Missing computed fields (status, limits, connection info)

**All bugs have documented fixes in the analysis.**

### ✅ Clear Split Boundary

**terraform-provider-kinsta (MyKinsta):**
- WordPress sites, environments, domains, backups, tools
- Base: `https://api.kinsta.com/v2`
- 56 WordPress endpoints

**terraform-provider-sevalla (Sevalla):**
- Applications, databases, static sites, pipelines
- Base: `https://api.sevalla.com/v2`
- 29 core endpoints (excludes WordPress overlap)

Both use same Bearer token authentication.

---

## Document Breakdown

### `ANALYSIS_SUMMARY.md` (Start Here)

**Purpose:** Executive summary and action plan  
**Read time:** 10-15 minutes

**Contains:**
- Key findings at a glance
- Database migration impact
- 4-phase ordered action plan
- Timeline estimate (9 weeks for P0/P1)
- Risk assessment
- Success criteria
- Open questions for API team

**Best for:** Project managers, technical leads, anyone needing overview

---

### `SEVALLA_SPEC_FINDINGS.md` (Database Migration)

**Purpose:** Detailed Sevalla API analysis with migration guide  
**Read time:** 15-20 minutes

**Contains:**
- Sevalla API overview (60 endpoints, 183 schemas)
- Database resource field-by-field comparison
- Migration checklist (8 required changes)
- Breaking changes for users
- State migration commands
- Application resource schema (partial - needs clarification)
- Endpoint catalog

**Best for:** Engineers implementing database migration, API team coordination

**Key Sections:**
- Section 3: Database schema comparison (critical for migration)
- Section 8: User migration guide with code examples
- Section 9: Open questions for API team

---

### `SPECS_ROADMAP.md` (Implementation Guide)

**Purpose:** Detailed specification files to create  
**Read time:** 20-25 minutes

**Contains:**
- 20+ spec files to write (foundation, Sevalla, Kinsta, quality)
- Spec file structure template
- Priority matrix (Phase 0-4)
- Writing guidelines
- Maintenance process

**Best for:** Engineers writing resource implementations, spec authors

**Key Sections:**
- "Spec Files to Create" - Complete list with priorities
- "Spec File Priority Matrix" - What to write when
- "Spec Writing Guidelines" - How to structure each spec

**Spec Files Highlighted:**
- `specs/10-sevalla-database-resource.md` (P0, ready to write)
- `specs/11-sevalla-application-resource.md` (P0, needs API clarification)
- `specs/20-kinsta-wordpress-site-resource.md` (P0, refine existing)

---

### `PROVIDER_SPLIT_ANALYSIS.md` (Deep Dive)

**Purpose:** Complete technical analysis (original analysis before Sevalla spec)  
**Read time:** 60-90 minutes

**Contains:**
- Full implementation inventory (client, resources, tests)
- Complete OpenAPI spec analysis
- Detailed spec vs code gap tables
- 5 correctness bugs with locations and fixes
- Proposed spec files with full outlines
- Quality refactoring opportunities
- Shared helper functions
- Test infrastructure improvements
- Ordered backlog (22 work items)

**Best for:** Senior engineers, architects, code reviewers

**Key Sections:**
- Section 4: Correctness bugs with code locations
- Section 5: Proposed spec files (detailed outlines)
- Section 6: Quality refactoring (error handling, polling, helpers)
- Section 7: Ordered backlog with time estimates

---

## Where to Go Next

### Immediate Actions (This Week)

1. **Review analysis** with team
   - Start: `ANALYSIS_SUMMARY.md`
   - Discuss: Timeline, risks, open questions

2. **Validate findings** with API team
   - Question: Does POST /applications exist? (not in Sevalla spec)
   - Question: Which application fields are updatable?
   - Question: Are deployments sync or async?

3. **Create project board**
   - Use Phase 0-4 breakdown from `ANALYSIS_SUMMARY.md`
   - Prioritize Phase 0 bugs (critical path)

4. **Set up repositories**
   - Fix kinsta-database in current repo
   - Create new repo for terraform-provider-sevalla

### Next Week

1. **Begin Phase 0** (Foundation fixes)
   - Fix kinsta_database bugs (see Section 4 in main analysis)
   - Implement centralized error handling
   - Add 404 handling to all resources
   - Add deprecation warnings

2. **Write first spec files**
   - `specs/00-provider-split-strategy.md`
   - `specs/01-error-handling-patterns.md`
   - `specs/10-sevalla-database-resource.md`

3. **Prototype database migration**
   - Test state import/export
   - Validate field mappings
   - Test breaking change impact

---

## Questions & Answers

### Q: Why split into two providers?

**A:** MyKinsta API has deprecated applications, databases, and static sites in favor of Sevalla API. Sunset date is January 31, 2026. The `kinsta_database` resource uses deprecated endpoint.

### Q: What's the migration impact for database users?

**A:** Breaking changes to field names (region→location, db_type→type, size→resource_type) and required inputs (db_password, db_user must now be provided). State migration is supported. See `SEVALLA_SPEC_FINDINGS.md` Section 8.

### Q: Can I implement Sevalla resources without fixing kinsta bugs?

**A:** Yes, but not recommended. The bug fixes (especially ForceNew and 404 handling) establish patterns needed for Sevalla. Also, the fixed kinsta_database is the template for sevalla_database migration.

### Q: What if applications can't be created via API?

**A:** Start with read-only data source (sevalla_applications). Add resource later when create endpoint is available/clarified. See open questions in `ANALYSIS_SUMMARY.md`.

### Q: How long will Phase 0 take?

**A:** Estimated 2 weeks:
- Database bug fixes: 2 days
- Deprecation notices: 1 day  
- Error handling refactor: 2 days
- Polling improvements: 1 day
- Testing: 2-3 days

### Q: What's the timeline for Sevalla provider?

**A:** 
- Phase 1 (Database resource): 1 week after Phase 0
- Phase 2 (Core resources): 3 weeks
- Total to P1 complete: ~6 weeks from Phase 0 complete

---

## Analysis Deliverables Checklist

- [x] Implementation inventory (client, resources, tests)
- [x] OpenAPI spec analysis (MyKinsta 85 endpoints, Sevalla 60 endpoints)
- [x] Spec vs code gap analysis (detailed tables)
- [x] Correctness bugs identified (5 critical, with fixes)
- [x] Sevalla database migration guide (field-by-field)
- [x] Provider split boundary (clear and justified)
- [x] Ordered backlog (4 phases, 22 items, time estimates)
- [x] Spec files roadmap (20+ specs, priorities, structure)
- [x] Risk assessment (high/medium/low with mitigations)
- [x] Open questions list (for API team)
- [x] Migration impact analysis (breaking changes documented)
- [x] Timeline estimates (9 weeks for P0/P1)

---

## Contributing to Analysis

If you find gaps or have questions:

1. **Technical questions** → Add to "Open Questions" in `ANALYSIS_SUMMARY.md`
2. **Bug clarifications** → Reference Section 4 in `PROVIDER_SPLIT_ANALYSIS.md`
3. **New requirements** → Update backlog in `ANALYSIS_SUMMARY.md`
4. **API behavior changes** → Update `SEVALLA_SPEC_FINDINGS.md`

---

## Related Files

- `swagger.json` - MyKinsta API spec (in repo root)
- `_spec_cache/sevalla.openapi.json` - Sevalla API spec (obtained 2026-01-01)
- `internal/provider/*.go` - Current resource implementations
- `internal/client/client.go` - HTTP client and polling logic

---

**Analysis completed by:** GitHub Copilot CLI  
**Analysis method:** Code inspection, OpenAPI spec analysis, cross-referencing  
**Next review:** After Phase 0 completion or API clarifications received

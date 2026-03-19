# Phase 5 Planning Document

**Status:** READY TO EXECUTE  
**Date:** 2026-01-03  
**Provider:** terraform-provider-kinsta (MyKinsta API)

---

## Overview

This document outlines the cleanup work needed to finalize the Kinsta provider by removing all Sevalla-related code and archiving split-analysis documentation.

---

## Current State Assessment

### Database Resource Files (TO REMOVE)
- `internal/provider/database_resource.go` (deprecated, marked for removal)
- `internal/provider/database_resource_test.go` (acceptance tests)
- `internal/provider/database_resource_unit_test.go` (unit tests)
- `docs/resources/database.md` (documentation with deprecation notice)

**Note:** `internal/client/database.go` does NOT exist - database operations were likely inline or never fully implemented.

### Split Analysis Docs (TO ARCHIVE)
Root directory markdown files:
- `ANALYSIS_README.md`
- `ANALYSIS_SUMMARY.md`
- `CORRECTIONS_APPLIED.md`
- `DOC_HYGIENE_COMPLETE.md`
- `DOC_HYGIENE_REPORT.md`
- `KINSTA_REMAINING_WORK.md`
- `PATCH_PLAN_CORRECTED.md`
- `PATCH_PLAN_OLD.md`
- `PROVIDER_SPLIT_ANALYSIS.md`
- `PROVIDER_SPLIT_ANALYSIS_UPDATED.md`
- `SEVALLA_SPEC_FINDINGS.md`

### Phase Tracking Docs (KEEP IN specs/)
- `specs/00-adr-provider-split.md` (architectural decision)
- `specs/00-llm-controller.md` (phase controller)
- `specs/01-phase-0.prompt.md` through `specs/06-phase-5-kinsta.prompt.md`
- `specs/02-operations-polling-contract.md` (technical spec)
- `specs/03-phase-*-complete.md` (completion markers)
- `specs/20-kinsta-wordpress-site-resource.md` (resource spec)
- `specs/21-kinsta-wordpress-environment-resource.md` (resource spec)

---

## Execution Plan

### Step 1: Create Archive Directory
```bash
mkdir -p specs/archive
```

### Step 2: Archive Split Analysis Docs
```bash
mv ANALYSIS_README.md specs/archive/
mv ANALYSIS_SUMMARY.md specs/archive/
mv CORRECTIONS_APPLIED.md specs/archive/
mv DOC_HYGIENE_COMPLETE.md specs/archive/
mv DOC_HYGIENE_REPORT.md specs/archive/
mv KINSTA_REMAINING_WORK.md specs/archive/
mv PATCH_PLAN_CORRECTED.md specs/archive/
mv PATCH_PLAN_OLD.md specs/archive/
mv PROVIDER_SPLIT_ANALYSIS.md specs/archive/
mv PROVIDER_SPLIT_ANALYSIS_UPDATED.md specs/archive/
mv SEVALLA_SPEC_FINDINGS.md specs/archive/
```

### Step 3: Remove Database Resource
```bash
rm internal/provider/database_resource.go
rm internal/provider/database_resource_test.go
rm internal/provider/database_resource_unit_test.go
rm docs/resources/database.md
```

### Step 4: Update provider.go
- Remove "kinsta_database" from ResourcesMap
- Remove database imports if any

### Step 5: Update README.md
- Replace with WordPress-focused content
- Clarify scope (WordPress only)
- Note Sevalla separation

### Step 6: Validation
```bash
go vet ./...
go build
go test ./internal/provider -v
grep -r "kinsta_database" internal/provider/ docs/
```

### Step 7: Create Completion Document
Create `specs/03-phase-5-kinsta-complete.md` with:
- List of removed files
- List of archived files
- Validation results
- Updated provider scope

---

## Expected Outcomes

**After Phase 5:**
- Clean repository focused solely on WordPress resources
- No deprecated or unused code
- Historical analysis preserved in archive
- Provider ready for Terraform Registry
- Clear documentation of scope and capabilities

**Files Remaining:**
- 2 resources: wordpress_site, wordpress_environment
- 2 docs: docs/resources/wordpress_site.md, docs/resources/wordpress_environment.md
- Clean examples directory
- Updated README
- Organized specs directory with archive

---

## Validation Checklist

- [x] All database files removed
- [x] Database removed from provider.go
- [x] 11 analysis docs moved to specs/archive/
- [x] README.md updated
- [x] go vet ./... passes
- [x] go build passes
- [x] go test ./internal/provider passes
- [x] No "kinsta_database" in active code
- [x] No "sevalla" in user-facing docs
- [x] specs/03-phase-5-kinsta-complete.md created

---

## Next Steps After Phase 5

1. **Terraform Registry Publication**
   - Prepare registry metadata
   - Set up GPG signing
   - Configure release automation

2. **Documentation Enhancement**
   - Add more examples
   - Create tutorial content
   - Document best practices

3. **CI/CD Setup**
   - Configure automated testing
   - Set up release pipeline
   - Add status badges

---

**Ready for Execution:** YES  
**Estimated Time:** 15-20 minutes  
**Risk Level:** LOW (only removal/archival, no code changes)

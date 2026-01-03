# Phase 5: Repository Cleanup & Finalization - Kinsta Provider

**Objective:** Remove all Sevalla-related code and documentation from terraform-provider-kinsta, leaving a clean WordPress-only provider ready for publication.

**Context:** Phases 0, 1, and 4 are complete. The Kinsta provider has working WordPress resources (site + environment) with full documentation and tests. The deprecated database resource was meant for Sevalla and should be removed entirely.

---

## Tasks

### 1. Remove Database Resource Code

**Files to delete:**
- `internal/provider/database_resource.go`
- `internal/provider/database_resource_test.go`
- `internal/provider/database_resource_unit_test.go`
- `internal/client/database.go`
- `docs/resources/database.md`
- Any examples using database resource

**Update `internal/provider/provider.go`:**
- Remove database resource from `ResourcesMap`
- Remove any database-related imports

### 2. Archive Split-Related Documentation

**Create directory:** `specs/archive/`

**Move these files to archive:**
- `PROVIDER_SPLIT_ANALYSIS.md`
- `PROVIDER_SPLIT_ANALYSIS_UPDATED.md`
- `SEVALLA_SPEC_FINDINGS.md`
- `PATCH_PLAN_CORRECTED.md`
- `PATCH_PLAN_OLD.md`
- `CORRECTIONS_APPLIED.md`
- `KINSTA_REMAINING_WORK.md`
- `ANALYSIS_README.md`
- `ANALYSIS_SUMMARY.md`
- `DOC_HYGIENE_COMPLETE.md`
- `DOC_HYGIENE_REPORT.md`

**Keep active:**
- `README.md` (update)
- `CHANGELOG.md`
- `specs/` directory (keep all phase docs)

### 3. Update README.md

**New structure should include:**

```markdown
# Terraform Provider for Kinsta (MyKinsta API)

Manage your Kinsta WordPress hosting infrastructure with Terraform.

## Supported Resources

- `kinsta_wordpress_site` - WordPress site management
- `kinsta_wordpress_environment` - Environment management (staging, production)

## Scope

This provider manages WordPress resources via the MyKinsta API (api.kinsta.com/v2).

**Note:** PaaS resources (applications, databases, static sites) are managed via a separate provider for the Sevalla API.

## Installation

[Installation instructions...]

## Authentication

[Auth instructions...]

## Example Usage

[Examples...]

## Documentation

See `docs/resources/` for complete resource documentation.

## Development

[Development setup...]

## License

[License info...]
```

### 4. Clean Up Codebase References

**Search and verify removal of:**
- All references to "database" resource in Go code (except test fixtures if needed)
- All references to "sevalla" in user-facing documentation
- Ensure `_spec_cache/` directory is in `.gitignore` (transient cache only)

### 5. Final Validation

**Run these commands and ensure they pass:**

```bash
# Linting
go vet ./...

# Build
go build

# Unit tests
go test ./internal/provider -v

# Ensure no database references in provider code
grep -r "kinsta_database" internal/provider/*.go
# (Should only appear in removed files)

# Ensure no sevalla references in docs
grep -r "sevalla" docs/
# (Should return nothing)
```

---

## Success Criteria

- [ ] All database resource files deleted
- [ ] Database removed from provider.go registration
- [ ] Split analysis docs archived to specs/archive/
- [ ] README.md updated to reflect WordPress-only scope
- [ ] No references to database resource in active code
- [ ] No references to sevalla in user-facing docs
- [ ] All validation commands pass
- [ ] specs/03-phase-5-kinsta-complete.md created

---

## Completion Marker

Create `specs/03-phase-5-kinsta-complete.md` documenting:
- Files removed
- Files archived
- Validation results
- Provider scope statement
- Next steps (if any)

---

## Notes

- This phase is purely cleanup - no new features
- The provider should be ready for Terraform Registry publication after this phase
- Database resource was deprecated in Phase 1, this phase completes the removal
- Archive preserves history without cluttering the main workspace

You are a senior Terraform provider engineer and technical editor.

Goal
- Apply the corrected doc hygiene changes and lock the provider split architecture.
- Because there are NO users of kinsta_database, do not do any safety hardening work for it. Deprecate-only.

Inputs
- PATCH_PLAN_CORRECTED.md
- CORRECTIONS_APPLIED.md
- Existing docs:
  - ANALYSIS_SUMMARY.md
  - SEVALLA_SPEC_FINDINGS.md
  - SPECS_ROADMAP.md
  - PROVIDER_SPLIT_ANALYSIS.md

Tasks
1. Apply all 17 corrected text replacements exactly as specified in PATCH_PLAN_CORRECTED.md.
2. Add the two canonical spec files verbatim:
   - specs/00-adr-provider-split.md
   - specs/02-operations-polling-contract.md
3. Ensure all deprecated/blocked claims cite explicit OpenAPI pointers.
4. Ensure kinsta_database strategy is explicitly:
   - “Deprecate immediately; no bugfix work since there are no users; remove after migration window.”
5. Verify the following invariants are true everywhere:
   - operation.data treated as opaque (lookup-after-poll required)
   - Sevalla provider MUST NOT implement /sites/*
   - POST /applications absence blocks sevalla_application resource (data source only)
   - Static sites and pipelines are deprecated in MyKinsta and belong in Sevalla
6. Produce a short verification checklist confirming architectural lock.

Output
- Updated markdown files ready to commit
- The two new canonical specs in ./specs/
- A summary of applied changes + a verification checklist
- No Go code changes

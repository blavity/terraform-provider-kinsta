<!--
Thanks for the contribution. Fill in the sections below and confirm the checklist before requesting review.

For security issues do NOT open a PR — use GitHub Security Advisories instead (see SECURITY.md).
-->

## Summary

<!-- One or two sentences: what does this change and why? -->

## Related issue

<!-- Closes #N, refs #N, or "none" if this is small enough not to warrant one. -->

## Type of change

<!-- Check one. -->

- [ ] `feat` — new resource, data source, attribute, or behavior
- [ ] `fix` — bug fix
- [ ] `chore` — tooling, deps, CI, repo plumbing
- [ ] `docs` — README/docs/comments only
- [ ] `refactor` — internal change with no user-visible effect

## Checklist

- [ ] Conventional-commit subject (`type(scope): summary`)
- [ ] Tests added or updated for behavioral changes
- [ ] `tfplugindocs generate --provider-name kinsta` ran clean (or no schema change)
- [ ] `golangci-lint run ./...` passes
- [ ] CHANGELOG impact noted in the PR title (release notes are generated from commit subjects)

## Notes for reviewers

<!-- Anything subtle: trade-offs, alternatives considered, follow-ups deferred. -->

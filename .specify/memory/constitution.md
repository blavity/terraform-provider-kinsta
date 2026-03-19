# terraform-provider-kinsta Constitution

## Core Principles

### I. Single-Resource Responsibility

Each resource (`kinsta_wordpress_site`, `kinsta_wordpress_environment`, and
future additions) MUST handle exactly one API resource lifecycle. No resource
MAY read or mutate another resource's state. Shared API logic MUST live in
`internal/client/`; shared provider utilities MUST live in `internal/provider/`
as unexported helpers. No package inside `internal/` MAY import another
sub-package of `internal/`.

**Rationale**: Isolated resource packages make individual resources testable in
isolation, limit blast radius for bugs, and allow new resources to be added
without touching unrelated code paths.

### II. Resource Schema Stability

Fields declared in a resource schema are a public contract for every Terraform
user pinning to a provider version. The following are BREAKING CHANGES and MUST
trigger a semver MAJOR bump:

- Removing an existing schema field
- Renaming an existing schema field
- Changing a field's type
- Adding `ForceNew: true` to an existing non-ForceNew field (silently triggers
  resource replacement for all existing users)
- Changing `Optional` to `Required`

The following are backward-compatible (MINOR):

- Adding a new `Optional` field with a sensible `Default`
- Adding a new `Computed` field
- Expanding the set of valid values for a validated field

**Rationale**: Terraform users pin provider versions. A schema change that
triggers unexpected resource replacement in production is a reliability incident
for every caller.

### III. Idempotency & Safe Defaults

Every resource MUST be safe to re-apply without corrupting remote or local
state. Specifically:

- Async operations (site creation, deletion, environment creation) MUST be
  polled to completion before returning from Create or Delete. Partial state —
  where a remote resource exists but the provider did not capture its ID — is
  a data loss bug.
- Resources MUST handle `404` on Read by calling `d.SetId("")` to signal
  out-of-band deletion, never by returning an error.
- Import (`terraform import`) MUST work for all resources. Fields that cannot
  be recovered from the API after import MUST be marked `Computed: true` so
  the state remains consistent.
- Default values for optional fields MUST represent the least-surprising, most
  conservative option (e.g., `install_mode = "new"`, not an aggressive default
  that alters existing content).

**Rationale**: Terraform's apply loop assumes idempotency. A non-idempotent
resource causes cascading plan drift and erodes user trust in the provider.

### IV. Credential Handling — No Ambient Auth

The API key and company ID MUST only be sourced from declared provider schema
attributes (`api_key`, `company_id`) or their designated environment variable
equivalents (`KINSTA_API_KEY`, `KINSTA_COMPANY_ID`). No resource or client
function MAY read credentials from files, ambient environment variables beyond
the declared ones, or any undocumented source.

Credentials MUST NOT be:
- Logged via `tflog` at any level
- Written to Terraform state in plaintext (use `Sensitive: true` on all
  credential-adjacent schema fields)
- Included in error message strings

**Rationale**: Providers run in user environments where logs and state files
may be shared or committed. Accidental credential exposure in a public provider
is a critical security incident.

### V. Test Every Behavior Unit

All resource CRUD logic, client methods, error paths, and polling behavior MUST
have unit tests using `github.com/stretchr/testify`. Tests MUST be runnable
without network access (mock servers via `net/http/httptest`). Acceptance tests
(requiring live credentials and `TF_ACC=1`) are required for the primary
create/read/delete cycle of every resource but are NOT run in standard CI.

**Coverage threshold**: Minimum acceptable statement coverage is **80%** for
`internal/client` and `internal/provider` packages, measured by:

```
go test -race -coverprofile=coverage.out ./internal/...
go tool cover -func coverage.out
```

**Ratchet rule**: Coverage MUST NOT decrease between merges to `main`. PRs that
reduce coverage MUST include a justification; a net reduction without
justification is a CI failure once the gate is active.

**New code**: New resource logic added in a PR with no corresponding test MUST
be justified in the PR description.

**Rationale**: The provider is distributed as a compiled binary. Unit tests are
the primary quality gate. The race detector is mandatory because provider
operations may be called concurrently by Terraform's parallel resource graph.

### VI. Conventional Commits Drive Automated Releases

All commits to `main` MUST follow the Conventional Commits specification with
scopes matching the changed component. Valid scopes:

| Scope | Coverage |
|---|---|
| `provider` | `internal/provider/provider.go`, provider-level config |
| `client` | `internal/client/` |
| `wordpress-site` | `wordpress_site_resource.go` and its tests |
| `wordpress-environment` | `wordpress_environment_resource.go` and its tests |
| `wordpress-domain` | domain resource (future) |
| `wordpress-backup` | backup resource (future) |
| `wordpress-sftp` | SFTP resource (future) |
| `data-sources` | data source implementations (future) |
| `ci` | `.github/workflows/` |
| `docs` | `docs/`, `examples/`, `README.md`, `CHANGELOG.md` |

`feat:` bumps MINOR. `fix:` bumps PATCH. `feat!:` / `BREAKING CHANGE:` bumps
MAJOR. Release automation MUST NOT be bypassed. Manual version tagging is
prohibited.

**Rationale**: Terraform Registry consumers read the changelog and pin to
version constraints. Automated, accurate changelogs depend entirely on commit
discipline.

### VII. Minimal, Pinned Dependencies

Direct Go dependencies are restricted to what is strictly necessary:

- `github.com/hashicorp/terraform-plugin-sdk/v2` — Terraform provider SDK
- `github.com/hashicorp/terraform-plugin-log` — structured provider logging
- `github.com/stretchr/testify` — test assertions

Adding a new direct Go dependency requires explicit justification in the PR.
`go.sum` MUST be committed atomically with any `go.mod` change. Run `go mod
tidy` after any dependency change; never commit a stale `go.sum`.

All GitHub Actions workflow steps MUST be pinned to a full commit SHA alongside
the version tag comment (e.g., `uses: actions/checkout@<sha> # v4`).

**Rationale**: The provider binary is distributed to end users. A larger
dependency surface increases supply-chain attack area, slows builds, and
creates additional maintenance burden for a library with no runtime update
mechanism.

### VIII. No Org-Specific Secrets in CI

Standard CI (lint, build, unit tests) MUST require only the built-in
`GITHUB_TOKEN`. Acceptance tests require `KINSTA_API_KEY` and
`KINSTA_COMPANY_ID`, which are caller-supplied and MUST be documented as such.
No workflow MAY depend on org-level or repository secrets beyond these. Any
contributor with a valid Kinsta account MUST be able to run the full test suite.

`GITHUB_TOKEN` MUST be scoped to the minimum permissions required for the job,
declared explicitly in the workflow `permissions` block.

**Rationale**: The provider is a public repository. Workflows that require
org-specific secrets cannot be validated by external contributors or forks,
breaking the open-source contribution model.

### IX. Organizational Information Confidentiality

No artifact committed to this repository MAY reference, describe, or expose:

- Internal processes, tooling, or organizational policies of any host
  organization
- Internal hostnames, internal API endpoints, or internal project names
- Trade secrets or proprietary information belonging to any organization using
  this provider
- Personally identifiable information of employees beyond what is already
  public (e.g., a GitHub username on a commit)

Documentation, examples, and agent instructions MUST be written generically.
All examples MUST use placeholder values (e.g., `admin@example.com`,
`us-central1`) that any user can substitute.

**Rationale**: This repository is public and targeted for publication on the
Terraform Registry. Any org-specific content leaks internal information and
reduces the provider's utility to the broader community.

### X. Own the Codebase

Pre-existing issues encountered during work MUST be surfaced and tracked, never
silently ignored.

- CI failures, lint warnings, and bugs found during implementation work MUST be
  clearly distinguished as pre-existing or caused by the current change.
- Trivial fixes (stale comment, one-line lint fix) MUST be fixed in the current
  PR when low-risk.
- Non-trivial pre-existing issues MUST have a GitHub issue created before the
  session ends. "Deferred" without an issue is the same as forgotten.
- Sentry, Copilot, and Gitleaks PR findings MUST be triaged and responded to;
  silence is not resolution.

**Rationale**: A Terraform Registry provider accumulates users. Technical debt
that silently corrupts state or causes plan drift harms every user who pins to
an affected version.

### XI. No Stranded Work

Completed work MUST reach its destination.

- Every local commit MUST be pushed within the same work session.
- After every commit, `git status` MUST be clean. Orphaned files are never
  acceptable.
- PRs MUST be opened for all pushed feature branches.
- `go.sum` and `go.mod` MUST be consistent and committed together.
- Agents MUST verify `git status` is clean before ending a session.

**Rationale**: Provider releases blocked on unpushed commits delay every user
waiting for a bug fix or feature. The cost is invisible until a downstream
Terraform run silently uses stale code.

### XII. Documentation as Artifact

User-facing and behavioral changes MUST include documentation updates in the
same PR.

- New or changed schema fields MUST be reflected in `docs/resources/*.md` (the
  Terraform Registry doc surface).
- New resources MUST include an `examples/` directory with at least a `main.tf`
  and `variables.tf`.
- Breaking changes MUST include a migration note in the PR description and in
  `CHANGELOG.md`.
- Timeout configuration (create/delete) MUST be documented in the relevant
  `docs/resources/*.md` file.
- `README.md` MUST accurately reflect the current set of supported resources.

Documentation MUST NOT reference internal systems, internal runbooks, or
org-specific tooling (Principle IX).

**Rationale**: `docs/resources/` is surfaced verbatim on the Terraform Registry.
When it drifts from the schema, every user reading the Registry documentation
gets incorrect information — with no error message.

---

## Security & Supply Chain

The provider runs with access to caller Kinsta API keys. The following controls
are NON-NEGOTIABLE:

- All GitHub Actions steps MUST use SHA-pinned action references.
- CI workflow permissions MUST follow least-privilege; scopes MUST be declared
  explicitly in the workflow `permissions` block.
- No API key or credential value MAY appear in `tflog`, `log`, `fmt.Print*`, or
  any output stream.
- Dependabot MUST remain enabled for Go modules and GitHub Actions.
- Schema fields that accept credentials (`admin_password`, `api_key`) MUST be
  marked `Sensitive: true`.
- PRs from external contributors MUST pass CI before any maintainer reviews
  credential-adjacent code paths.
- Provider functions MUST surface failures through `diag.Diagnostics`, never
  through silent zero-value returns.

---

## Agentic Development Standards

This repository is designed for heavy agentic (AI agent) use. The following
rules govern agent-driven contributions:

- **Constitution first**: Read the full constitution before making any change.
  There are no exceptions for "small" or "quick" changes.
- **Scope discipline**: Confine changes to the resource(s) named in the task.
  Do not modify sibling resources or provider-level code unless explicitly
  instructed.
- **No speculative dependencies**: Do NOT add Go module dependencies without
  explicit user approval. Run `go mod tidy` after any dependency change and
  commit `go.sum` atomically with `go.mod`.
- **Conventional commits are mandatory**: Every commit MUST use a scoped
  conventional commit message. Unscoped commits break release automation.
- **Always validate before proposing a PR**: `go vet ./... && go build ./... &&
  go test ./internal/...` must all pass cleanly.
- **No credential logging**: Audit all written code for accidental credential
  exposure before committing.
- **Sensitive fields**: Any schema field that accepts a credential or secret
  MUST have `Sensitive: true`. This is non-negotiable.
- **Clean working tree**: After every commit, verify `git status` is clean.
  Orphaned or untracked files are never acceptable.
- **Public repo awareness**: This repository is public and will be listed on the
  Terraform Registry. Do NOT commit `.env` files, token values, internal
  hostnames, or any non-public information.
- **No org-specific context**: Do NOT embed references to any host
  organization's internal tooling, team names, processes, or policies in any
  committed artifact. All documentation MUST remain generic and portable
  (Principle IX).
- **Surface and track pre-existing issues**: When CI failures, lint warnings, or
  bugs unrelated to the current task are encountered, create a GitHub issue
  before proceeding (Principle X).
- **No stranded commits**: Push all commits and verify the remote is up to date
  before ending a session (Principle XI).
- **Agent directories are gitignored**: `.claude/`, `.codex/`, `.opencode/`, and
  `.specify/**` (except `memory/constitution.md`) MUST remain in `.gitignore`.
  Do NOT commit agent working directories.

---

## Responsible Agentic Use & Pull Request Policy

**Scope of autonomy**:

- Agents MUST operate only within the scope explicitly defined by the current
  task. Refactoring, restructuring, or adding features beyond the stated task
  is prohibited without explicit user instruction.
- Agents MUST NOT open speculative PRs.
- When requirements are ambiguous, agents MUST ask before implementing.

**Authorship transparency**:

- PRs authored by agents MUST include a statement identifying the changes as
  agent-generated and summarizing the session scope. The statement MUST be
  placed where a reviewer will see it before approving.
- Agent commits MUST NOT be attributed to a human identity. Co-authorship
  trailers are encouraged when the agent worked from a human-provided spec.

**Pull request policy**:

- Every agent-opened PR MUST receive at least one human maintainer approval
  before merge. Agents MUST NOT self-approve or dismiss reviews.
- Agent PRs MUST include a Compliance Statement in the PR description listing
  each principle materially affected and confirming compliance or documenting
  a justified exception.
- PRs that add, remove, or rename schema fields MUST explicitly state the
  Principle II impact and the semver bump type.
- PRs that introduce a new direct Go dependency MUST include a Principle VII
  justification.
- Agent PRs MUST NOT be merged while CI is failing.

**Reversibility preference**:

- Agents MUST prefer small, targeted changes over large-surface rewrites.
- Agents MUST NOT force-push to any branch.

---

## Governance

This constitution supersedes all other development guidance in this repository.
When a rule in `CONTRIBUTING.md`, a PR comment, or an agent instruction
conflicts with this document, this document wins.

**Amendment procedure**:
1. Open a PR with the proposed change to this file.
2. State the version bump type (MAJOR/MINOR/PATCH) and rationale in the PR.
3. At least one maintainer MUST approve before merge.
4. `CONSTITUTION_VERSION` and `Last Amended` MUST be updated in the same commit.

**Versioning**: Principle removals or redefinitions = MAJOR. New principles or
sections = MINOR. Clarifications, wording, typos = PATCH.

**Compliance review**: Every PR description MUST include a brief Compliance
Statement confirming compliance with affected principles or documenting a
justified exception.

**Version**: 1.0.0 | **Ratified**: 2026-03-19 | **Last Amended**: 2026-03-19

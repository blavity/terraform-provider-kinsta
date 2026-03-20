# Agent Guide — terraform-provider-kinsta

## Constitution check is mandatory

**Before making any change**, read the full constitution:

> `.specify/memory/constitution.md`

Every principle applies to every task. There are no exceptions scoped to
"small" or "quick" changes. If you have not read the constitution in this
session, read it now before proceeding.

---

## Repo at a glance

A Terraform provider for managing WordPress hosting via the MyKinsta API
(`api.kinsta.com/v2`). Single provider binary with two resources:

```
internal/
  client/         — HTTP client + types for MyKinsta API
    client.go
    client_test.go
  provider/       — Terraform resource implementations
    provider.go
    wordpress_site_resource.go
    wordpress_environment_resource.go
    (+ _test.go and _unit_test.go for each)

docs/resources/   — Terraform Registry documentation (one .md per resource)
examples/         — Example HCL configurations (one dir per resource)
specs/            — Architecture decisions and phase specs
```

Rules:
- `internal/client` and `internal/provider` MUST NOT import each other
  (provider imports client, never the reverse).
- Schema fields are a public API — see Principle II before touching any schema.
- `docs/resources/*.md` is the Terraform Registry doc surface — keep it in sync
  with the schema.

## Toolchain

| Command | What it does |
|---|---|
| `go vet ./...` | Static analysis |
| `go build ./...` | Compile |
| `go test ./internal/...` | Unit tests (no network required) |
| `golangci-lint run ./...` | Full lint suite (errcheck, staticcheck, gocritic, etc.) |
| `TF_ACC=1 go test ./internal/provider/ -v` | Acceptance tests (requires live credentials) |
| `go mod tidy` | Sync go.sum after dependency changes |

`golangci-lint` is enforced by a pre-commit hook (`.pre-commit-config.yaml`).
Zero issues is required — never add `//nolint` without a reason comment.

**Always run `go vet ./... && go build ./... && go test ./internal/... && golangci-lint run ./...`
before proposing a PR.** Never open a PR that fails CI.

## Commit format

```
type(scope): description
```

Valid scopes: `provider`, `client`, `wordpress-site`, `wordpress-environment`,
`wordpress-domain`, `wordpress-backup`, `wordpress-sftp`, `data-sources`,
`ci`, `docs`.

Commits without a valid scope break release automation.

## Every change requires a pull request

**Never commit directly to `main`.** All changes — including small fixes,
doc updates, and constitution-driven corrections — MUST go through a PR.

Workflow:
1. Create a feature branch: `git checkout -b fix/description` or `feat/description`.
2. Make changes, commit with a scoped conventional commit message.
3. Push and open a PR: `git push -u origin <branch> && gh pr create ...`.
4. Wait for human approval before merging.

There are no exceptions for "quick" or "trivial" changes. A direct push to
`main` is a process violation regardless of change size.

## Before opening a PR

1. `go vet ./... && go build ./... && go test ./internal/... && golangci-lint run ./...` passes cleanly.
2. Compliance Statement in the PR description — list each affected principle
   and confirm compliance or document a justified exception (required, not
   advisory).
3. If you changed schema fields: state the Principle II impact and semver bump.
4. If you added a Go dependency: include a Principle VII justification.
5. The PR description states that changes are agent-generated.
6. `git status` is clean and the branch is pushed.

A human maintainer must approve before merge. You may not self-approve or
dismiss reviews.

## Hard stops — ask before proceeding

Stop and ask the user if you encounter any of the following:

- A task that would remove, rename, or change the type of an existing schema
  field (Principle II — potential breaking change)
- A task that would add `ForceNew: true` to an existing field (Principle II —
  forces resource replacement for all users)
- A task that would add a new direct Go dependency (Principle VII)
- Ambiguous requirements where multiple interpretations are plausible
- A pre-existing CI failure or non-trivial bug
- Anything that would expand scope beyond what was explicitly requested

## What to do with pre-existing issues

- **Trivial** (stale comment, one-line lint fix): fix in the current PR.
- **Non-trivial**: open a GitHub issue with context and root cause, then
  continue. Do not silently work around or ignore (Principle X).

## Sensitive fields

Any schema field that accepts a credential or secret MUST have `Sensitive: true`.
Current sensitive fields: `api_key` (provider), `admin_password`, `admin_email`.
If you add a field that accepts a password, token, or key, `Sensitive: true` is
non-negotiable (Principle IV).

## Acceptance tests

Acceptance tests (`TF_ACC=1`) create real resources in a live Kinsta account
and will incur costs. They are not run in standard CI. When writing or updating
acceptance tests:

- Use `t.Skip` if `TF_ACC` is not set.
- Use `ImportStateVerifyIgnore` for write-only fields not returned by the API.
- Clean up resources in test cleanup functions.

## What this repo does not have

No MkDocs site, no integration test suite beyond acceptance tests, no
observability pipeline, no org-level CI secrets. If a task assumes otherwise,
stop and verify before implementing.

# Architecture Decision Records

An ADR (Architecture Decision Record) is a short document that captures a significant decision: what was decided, why,
and what tradeoffs were accepted. ADRs are the primary mechanism for preserving institutional knowledge in this repo.

## When to Write an ADR

Write an ADR when a decision is:

- **Hard to reverse** — changing it later would require migration work or consumer coordination (e.g. removing a
  resource schema field, changing a default behavior, accepting an inevitable disclosure surface).
- **Non-obvious** — a future contributor reading the code would reasonably ask "why is it done this way?"
- **Worth debating** — if reasonable people could disagree, the outcome deserves a record.

You do not need an ADR for every PR. Bug fixes, docs updates, and straightforward feature additions generally do not
warrant one.

## Directory and Naming

```text
docs/adr/
  template.md          # Copy this to start a new ADR
  0001-title.md        # Accepted ADRs are numbered sequentially
  0002-title.md
```

Files are named `NNNN-kebab-case-title.md`. Numbers are zero-padded to four digits and assigned sequentially — do not
skip or reuse numbers.

## Status Values

| Status       | Meaning                                          |
| ------------ | ------------------------------------------------ |
| `Proposed`   | Under discussion — not yet accepted              |
| `Accepted`   | In effect                                        |
| `Deprecated` | Still in effect but being phased out             |
| `Superseded` | Replaced by a later ADR (reference it by number) |

## Writing a Good ADR

Use `docs/adr/template.md` as your starting point. A few principles:

- **Context is the most important section.** State the situation clearly — what changed, what constraint exists, why a
  decision is needed now.
- **Decision is a statement, not a discussion.** "We do X" not "We considered X".
- **Consequences should be honest.** Include the downsides. An ADR with no negatives is incomplete.
- **Alternatives Considered should be real alternatives**, not strawmen. If you seriously evaluated something and
  rejected it, say why.

## Documenting Legacy Exceptions

When code predates a governance rule or ADR and cannot be immediately brought into compliance, document the exception
rather than ignoring it:

```markdown
<!-- In the relevant ADR or a standalone exception note -->

**Exception**: `<path/to/legacy>` predates rule X (ADR-NNNN). Tracked for removal in #NNNN. Owner: @maintainer.
```

This prevents the exception from becoming invisible debt. It should reference a tracking issue with a remediation plan.

## CI Lint Gate

The `Governance` workflow (`.github/workflows/governance.yaml`) validates ADR format on pull requests. It checks that
numbered ADR files (`docs/adr/NNNN-*.md`) have the required headings and a valid status value. The workflow ships in
warn-only mode by default; flip `FAIL_ON_VIOLATIONS` to `"true"` in the workflow to graduate to a blocking check once
the ADR backlog is established.

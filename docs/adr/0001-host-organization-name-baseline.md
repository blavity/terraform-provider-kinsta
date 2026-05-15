# 0001: Host Organization Name Baseline (Principle IX Inevitable Disclosures)

**Status**: Accepted **Date**: 2026-05-15

## Context

This repository is intended for publication on the public Terraform Registry. The repository's constitution, Principle
IX (Organizational Information Confidentiality), forbids exposure of internal organizational information in any
committed artifact.

However, a small set of disclosures are **inherent** to the act of publishing a Go-based Terraform provider from a
named GitHub organization. They cannot be remediated without renaming the GitHub organization or repository, replacing
the Go module path (a breaking change for every downstream consumer), or republishing under a different Terraform
Registry namespace.

The concrete name of the host organization unavoidably appears in surfaces such as `github.com/<org>/<repo>` and
`registry.terraform.io/<org>/<name>`. This ADR records the decision to accept those specific surfaces as a fixed
disclosure baseline so that future audits can distinguish "inevitable surface" from "remediable leak."

## Decision

The following disclosures are accepted as the **Principle IX baseline** for this repository. They are not violations
and do not require remediation:

1. **GitHub repository URL** — `github.com/<org>/<repo>`. Set at repo creation; renaming would break every existing
   clone, fork, link, and CI integration.
2. **Go module path** — declared in `go.mod`, must match the canonical GitHub URL per Go convention
   (`go mod` and `go get` resolve modules by their import path).
3. **Terraform Registry namespace slug** — `registry.terraform.io/<org>/<name>`. Derived from the GitHub organization at
   publish time; changing it invalidates every consumer's `required_providers` `source` line.
4. **`LICENSE` file copyright holder** — naming the copyright holder is open-source convention and is required by most
   permissive licenses to be enforceable.
5. **Commit author identity metadata** — Git records `user.name` and `user.email` at commit time; this is intrinsic to
   the commit graph and cannot be retroactively scrubbed without rewriting history.
6. **Commit signature identity** — GPG/SSH signing keys identify their owner. This is the point of signing.

These items are the inherent disclosure surface of "having a named GitHub-hosted, Go-based Terraform provider published
under a public namespace." Reducing them requires either renaming the host organization and republishing, or adopting
`noreply`-style author identities (see Consequences).

**Every other internal disclosure is NOT covered by this baseline** and MUST be remediated per Principle IX. This
includes, but is not limited to:

- Internal project names, internal product code names, or internal team names appearing in committed files, comments,
  PR bodies, or commit messages.
- Internal service account emails, internal Doppler paths, internal Workload Identity Federation pool/provider names,
  or other host-platform-specific identifiers in workflows or documentation.
- Cross-references to internal companion repositories ("see internal repo X for the IAM setup").
- Internal hostnames, internal API endpoints, or internal runbook URLs.

When such items are encountered, the remediation is generic placeholder language plus, where the value is necessary at
runtime, sourcing it from a repo-scoped secret with a generic name.

## Consequences

- Users installing the provider will see `<org>/<name>` in their `terraform { required_providers { … } }` block. This
  is acceptable and expected for any Terraform Registry provider.
- Contributors are encouraged (but not required) to use `noreply`-style email forms in their git config when
  contributing, to limit author identity leakage. The repository does not enforce this.
- Future remediation could include renaming the host organization and republishing the provider under a different
  namespace. This is out of scope for this ADR and would require coordinated consumer migration.
- Audits against Principle IX can use this ADR as a fixed allow-list of inherent surfaces. Any other internal
  disclosure remains a violation.

## Alternatives Considered

- **Rename the GitHub organization and republish.** Rejected: high coordination cost, breaks every existing consumer,
  invalidates the existing Terraform Registry namespace, and the constitution permits the inherent surface explicitly
  via this ADR mechanism.
- **Use a vanity Go module path (e.g. via a redirect domain).** Rejected: introduces a new infrastructure dependency
  (the redirect host) that itself becomes a confidentiality and reliability concern, and consumer tooling still
  resolves to the underlying GitHub URL.
- **Mark Principle IX as not applicable to "inherent" disclosures without a written record.** Rejected: leaves the
  baseline implicit and re-litigates the question on every future audit. ADRs are the mechanism the constitution
  expects for accepted exceptions (see "Documenting Legacy Exceptions" in `docs/governance/adr-guide.md`).

# Contributing to terraform-provider-kinsta

Thanks for your interest in contributing. This guide covers how to set up a local dev loop, what we expect from PRs, and how we coordinate releases.

## Code of Conduct

This project follows the [Contributor Covenant v2.1](https://www.contributor-covenant.org/version/2/1/code_of_conduct/). By participating you agree to abide by its terms. A local `CODE_OF_CONDUCT.md` will be added in a follow-up PR; until then, the canonical URL above is authoritative.

## Reporting bugs and requesting features

- **Bugs**: open a [Bug report](https://github.com/blavity/terraform-provider-kinsta/issues/new?template=bug_report.yml) issue. Include the provider version, Terraform/OpenTofu version, a minimal config that reproduces the issue, and the output of `TF_LOG=DEBUG terraform <command>` if relevant.
- **Features**: open a [Feature request](https://github.com/blavity/terraform-provider-kinsta/issues/new?template=feature_request.yml) issue describing the resource/data-source/behavior and a concrete use case.
- **Questions**: prefer [Discussions](https://github.com/blavity/terraform-provider-kinsta/discussions) (if enabled) or a [Question](https://github.com/blavity/terraform-provider-kinsta/issues/new?template=question.yml) issue.
- **Security vulnerabilities**: see [SECURITY.md](SECURITY.md) — do **not** file public issues for security findings.

## Dev loop

### Requirements

- Go (version matches `go.mod`)
- [golangci-lint](https://golangci-lint.run/welcome/install/) v2+
- `tfplugindocs` pinned to the version in `.github/workflows/ci.yml` (currently `v0.25.0`)

### Common commands

```bash
# Build
go build ./...

# Vet + unit tests
go vet ./...
go test -race ./internal/...

# Lint
golangci-lint run ./...

# Acceptance tests (live MyKinsta credentials required)
export KINSTA_API_KEY="..."
export KINSTA_COMPANY_ID="..."
TF_ACC=1 go test ./internal/provider/ -v

# Regenerate docs (must match the CI pin)
go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@v0.25.0
tfplugindocs generate --provider-name kinsta
```

### Local-binary testing

Use a `dev_overrides` block to test changes against your own Terraform configs without publishing:

```hcl
# ~/.terraformrc
provider_installation {
  dev_overrides {
    "blavity/kinsta" = "/path/to/terraform-provider-kinsta"
  }
  direct {}
}
```

Then `go build .` produces a binary at the repo root that Terraform will pick up.

## Pull requests

### Commit messages

Use [Conventional Commits](https://www.conventionalcommits.org/) with a scope:

```
feat(resource): add kinsta_wordpress_domain
fix(client): handle 429 rate-limit responses
chore(deps): bump terraform-plugin-sdk to 2.40.1
docs(readme): clarify auth env vars
```

Scopes route automated tooling (release-note grouping, Dependabot prefixes). Common scopes: `provider`, `client`, `resource`, `release`, `ci`, `deps`, `docs`.

### PR checklist

Before requesting review:

- [ ] Tests added or updated for behavioral changes.
- [ ] `tfplugindocs generate --provider-name kinsta` ran clean — committed any updates to `docs/`.
- [ ] `golangci-lint run ./...` passes.
- [ ] Commit subject follows conventional commits.
- [ ] Linked issue (if applicable).

The PR template in `.github/PULL_REQUEST_TEMPLATE.md` carries the same checklist.

### Bot reviewers

This repo runs Copilot, Gitleaks, Semgrep, Trivy, and Socket Security on every PR. Address bot findings (or reply with a justification and resolve the thread) before requesting human review.

## Releasing

Releases are cut manually by a maintainer pushing a semver tag — see the **Releasing** section in [README.md](README.md). The release pipeline:

1. Push tag `vX.Y.Z` (or pre-release `vX.Y.Z-rc.1`).
2. `.github/workflows/release.yml` runs GoReleaser, signs `SHA256SUMS` with GPG, generates the changelog, and uploads artifacts.
3. Terraform Registry and OpenTofu Registry ingest the release automatically.

Contributors don't need release access — open a PR, get it merged, and the next release will include it.

## AI-assisted contributions

We welcome contributions made with AI assistance (Copilot, Claude, Cursor, etc.). The bar is the same as for any other contribution: **you are the author of record and responsible for what you submit.**

Specifically:

- **You verified it.** Read every line. Run the tests. Don't open a PR with code you don't understand or haven't exercised locally.
- **Same quality gates.** AI-assisted code passes the same lint, tests, and `tfplugindocs generate --provider-name kinsta` docs-generation requirements as hand-written code. If you add or change Terraform Plugin SDK schema fields, include an appropriate `Description` so the schema is self-documenting and generated docs stay useful. Use of an AI tool is not a justification to skip a check or lower a bar.
- **Don't file auto-generated bug reports.** Issues that read as raw model output with no human verification (no repro tested, no debug log gathered, suggested fixes that don't compile) will be closed. Either verify the report yourself or don't file it.
- **Disclose substantial AI involvement.** A `Co-Authored-By:` trailer naming the model (e.g., `Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>`) on commits where AI did meaningful authoring is appreciated. Optional, not required.
- **License and provenance still apply.** Don't submit code from an AI tool unless you have verified its provenance and confirmed it is compatible with MPL-2.0.

Reviewers may push back harder on AI-assisted PRs that look unverified — that's a feature, not a slight against the tooling.

## License

By contributing, you agree your contributions are licensed under [MPL-2.0](LICENSE), matching the project license.

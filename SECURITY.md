# Security Policy

## Reporting a vulnerability

**Do not** open a public GitHub issue, discussion, or pull request for security findings.

Instead, please use GitHub's private vulnerability reporting:

1. Open <https://github.com/blavity/terraform-provider-kinsta/security/advisories/new>
2. Provide a clear description, reproduction steps, affected versions, and any proof-of-concept.

We will acknowledge receipt within 5 business days and aim to confirm a triage outcome within 10 business days. Coordinated disclosure timelines are negotiated case-by-case based on severity and complexity.

## Scope

In scope:

- Code in this repository that interacts with the MyKinsta API on a user's behalf
- The release-signing pipeline and its produced artifacts (zips, `SHA256SUMS`, `SHA256SUMS.sig`, `manifest.json`)
- Documentation that, if followed verbatim, would lead a user to expose credentials

Out of scope:

- Vulnerabilities in upstream dependencies (file those upstream; we track via Dependabot)
- Issues in the MyKinsta API itself (report to Kinsta directly)
- Findings that require physical access, social engineering, or compromise of a user's local workstation

## Supported versions

The latest minor release on the latest major receives security fixes. Older majors may receive backports for critical issues at our discretion until the next major has been generally available for at least 90 days.

| Version | Supported |
|---------|-----------|
| latest `0.x` | ✅ |
| older `0.x` | best effort |

## Release signing

All releases are GPG-signed. The signing key fingerprint is published on the [Terraform Registry](https://registry.terraform.io/providers/blavity/kinsta) signing-keys page and the [OpenTofu Registry](https://registry.opentofu.org/providers/blavity/kinsta). To verify a release locally see `KEYS.md` → "Local verification".

# 0002: Release Signing Pattern for Terraform Registry Publication

**Status**: Accepted **Date**: 2026-05-15

## Context

Releases of this provider must be signed for the public Terraform Registry to
accept them. The Terraform Registry currently accepts GPG signatures only
(RSA or DSA keys; ECC is not supported). Sigstore / cosign signing is on the
roadmap for OpenTofu's registry but is not yet supported for new community
provider claims on the HashiCorp Terraform Registry.

An initial implementation attempted to bind release signing to an HSM-backed
asymmetric key via PKCS#11 (`libkmsp11` + `gnupg-pkcs11-scd`) so that no
exportable private key material ever existed outside a managed HSM. After
substantial wiring and debugging effort, that approach surfaced a meaningful
operational cost (multi-component install on each release runner, cross-
project IAM that didn't align with the rest of the host platform's
architecture, no diagnostic telemetry on first-call failures, environment
inheritance complications between gpg-agent and its scdaemon child) that was
disproportionate to the threat being mitigated for an open-source provider
release pipeline whose signature verification model is determined entirely by
the Terraform Registry trust chain — not by where the signing happened.

Investigation of the canonical ecosystem pattern showed that the maintained
HashiCorp reference template
(`hashicorp/terraform-provider-scaffolding-framework`) uses a plain GPG
key stored as a GitHub Actions repository secret, imported at release time
via `crazy-max/ghaction-import-gpg`, and consumed by GoReleaser. This is the
de-facto industry pattern for the vast majority of public Terraform
providers.

A decision was needed on which signing pattern this provider should adopt
before its first public release.

## Decision

Adopt the canonical HashiCorp pattern: a dedicated GPG keypair generated
once by the maintainer, with the private half stored as a GitHub Actions
repository secret (encrypted at rest by GitHub) alongside its passphrase,
imported into the release workflow via `crazy-max/ghaction-import-gpg`,
fingerprint passed to GoReleaser via env var, public half registered with
the Terraform Registry.

HSM-backed signing via PKCS#11 to a cloud KMS is **not** used.

## Consequences

**Easier:**

- Operational surface area for the release pipeline is reduced to a handful
  of well-documented components; the release workflow matches the canonical
  HashiCorp scaffolding template line-for-line.
- Rotation is straightforward: generate a new keypair, register the new
  public key with the Terraform Registry, update the repository secret. No
  cloud-side state to coordinate.
- No cross-project cloud IAM dependencies, no PKCS#11 provider library to
  pin/verify/track, no daemon-environment-inheritance pitfalls.
- New maintainers can onboard against published HashiCorp documentation and
  the canonical template; nothing about this pipeline is bespoke.

**Harder / accepted risks:**

- The private signing key exists outside an organization-controlled HSM.
  Mitigations: the key is access-controlled by GitHub's secret storage
  (encrypted at rest, no API surface for retrieval), the key is dedicated
  to provider release signing only (low blast radius), and rotation is
  fast enough that compromise-response is acceptable.
- The Terraform Registry trust model does not distinguish between
  signatures produced by HSM-backed and software-backed keys — both are
  byte-identical from the Registry's perspective — so the operational cost
  of HSM-backing did not buy additional trust at the registry boundary.

**Forward-looking:**

- If / when the HashiCorp Terraform Registry adopts Sigstore / cosign
  keyless signing for community providers, this provider can revisit the
  signing pattern. Keyless signing would eliminate the persistent-private-
  key concern entirely. This is tracked separately and is not a near-term
  blocker.

## Alternatives Considered

1. **HSM-backed signing via PKCS#11 to a cloud KMS.** Rejected after
   implementation: high integration cost, cross-project authorization
   complications, opaque first-call failures, no proportional security
   benefit at the registry trust boundary. Constitutes a future option if
   the Terraform Registry adds first-class HSM-attestation support.

2. **Sigstore / cosign keyless signing.** The 2026-modern pattern with
   ephemeral keys, OIDC-bound identity, and a public transparency log.
   Adopted by OpenTofu but not yet by the HashiCorp Terraform Registry
   for community providers. Revisit when registry support lands.

3. **Hybrid: GPG key generated in HSM, exported once, stored in secret
   manager.** Worst of both worlds — the key still leaves the HSM (so the
   "HSM-backed" property is only meaningful at generation time, not at
   signing time), but operationally still requires standing up the HSM.

---

- References:
  - HashiCorp Terraform Registry — provider publishing documentation (GPG-only
    signature requirement; RSA or DSA accepted)
  - `hashicorp/terraform-provider-scaffolding-framework` — canonical release
    workflow used by the upstream-maintained reference template
  - `crazy-max/ghaction-import-gpg` — the GPG-import action used by the
    reference template
  - GoReleaser `signs:` documentation — the consumer of the imported key
  - OpenTofu issue #307 — open proposal for Sigstore support (not yet
    available on the HashiCorp registry)

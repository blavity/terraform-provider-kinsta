# Release Signing & Registry Checklist

This document is the one-time maintainer setup for publishing signed
releases to the public Terraform Registry. The architectural rationale for
the chosen pattern (plain GPG via GitHub Actions secret, not HSM-via-PKCS#11
to a cloud KMS) is captured in [ADR-0002](docs/adr/0002-release-signing-pattern.md).

The release flow is the canonical pattern from
[hashicorp/terraform-provider-scaffolding-framework](https://github.com/hashicorp/terraform-provider-scaffolding-framework):

1. Generate a dedicated GPG keypair (one time, maintainer's workstation)
2. Register the public half with the Terraform Registry
3. Store the private half + passphrase as repository secrets
4. `release.yml` imports the key on each tag-triggered run and signs via
   GoReleaser

## 1) Generate the signing keypair (one time)

On the maintainer's workstation:

```bash
# Set this to the address that will appear publicly on the GPG UID baked
# into your signed releases (typically a role mailbox, not a personal one).
# The same address is what shows up on the Terraform Registry's key page.
RELEASE_EMAIL="releases@your-org.example"

gpg --full-generate-key
# Choose: (1) RSA and RSA
# Keysize:  4096
# Validity: 0 (does not expire — rotate manually if compromised)
# Real name: terraform-provider-kinsta release signing
# Email:    use the same value as $RELEASE_EMAIL above
# Passphrase: pick a strong one; you'll store it as a repo secret

# Export both halves.
gpg --armor --export "$RELEASE_EMAIL" > public.asc
gpg --armor --export-secret-keys "$RELEASE_EMAIL" > private.asc
```

The key type must be RSA or DSA — the Terraform Registry does not accept
ECC. RSA-4096 is the recommended default.

## 2) Register the public key with the Terraform Registry

1. Sign in to https://registry.terraform.io with the GitHub account that
   will publish under the `blavity` namespace.
2. Open user settings → **Signing Keys** → **New GPG Key**.
3. Paste the contents of `public.asc` (full ASCII-armored block, including
   the `-----BEGIN PGP PUBLIC KEY BLOCK-----` lines).
4. Save. The fingerprint listed in the Registry must match
   `gpg --fingerprint $RELEASE_EMAIL` locally.

## 3) Provision the repository secrets

```bash
gh secret set GPG_PRIVATE_KEY \
  --repo blavity/terraform-provider-kinsta < private.asc

gh secret set PASSPHRASE \
  --repo blavity/terraform-provider-kinsta \
  --body '<the passphrase from step 1>'
```

After both secrets are set, **discard the local `private.asc`** — the only
authoritative copy lives in the repo secret. If you need to recover later,
generate a fresh keypair and re-register with the Registry (rotation is
cheap; see ADR-0002).

## 4) First public release

Once secrets are provisioned and `public.asc` is registered:

1. Land any pending changes on `main`.
2. From `main`, push a semver tag:
   ```bash
   git tag v0.3.0
   git push origin v0.3.0
   ```
3. The tag triggers `.github/workflows/release.yml`, which imports the GPG
   key, runs GoReleaser, signs the `SHA256SUMS` checksum file, generates
   the changelog into the GitHub Release body, and uploads the
   registry-shaped artifacts.
4. Once the workflow succeeds, the Terraform Registry's webhook (configured
   one-time via the Registry's "Publish Provider" flow) ingests the
   release. The new version becomes installable as
   `registry.terraform.io/blavity/kinsta` within minutes.

## Release recovery

If GoReleaser fails partway, the GitHub Release may exist with partial or
no artifacts. Recovery:

```bash
gh release delete vX.Y.Z --repo blavity/terraform-provider-kinsta --yes --cleanup-tag
# `--cleanup-tag` removes the remote tag but not the local one. Drop the
# stale local tag (or force-overwrite it below) before re-pushing.
git tag -d vX.Y.Z 2>/dev/null || true
# Fix the underlying cause on main, then re-push the tag:
git tag vX.Y.Z <sha> && git push origin vX.Y.Z
```

The `workflow_dispatch` trigger in `release.yml` is an alternative re-run
path against an already-cleaned-up tag.

## Rotation

To rotate the signing key:

1. Repeat step 1 to generate a fresh keypair.
2. Add the new `public.asc` to the Terraform Registry (the Registry allows
   multiple active keys per user; verifications against the old signature
   on already-published releases continue working).
3. Replace `GPG_PRIVATE_KEY` and `PASSPHRASE` repo secrets (step 3).
4. Cut a release. Verify it carries the new fingerprint.
5. Optionally remove the old public key from the Registry once you are
   confident no future verifications need it.

## Local verification (optional)

To verify a published release matches the registered key:

```bash
TAG=v0.1.0
gh release download "$TAG" --repo blavity/terraform-provider-kinsta \
  --pattern "terraform-provider-kinsta_${TAG#v}_SHA256SUMS*"
gpg --verify \
  "terraform-provider-kinsta_${TAG#v}_SHA256SUMS.sig" \
  "terraform-provider-kinsta_${TAG#v}_SHA256SUMS"
```

A `Good signature` line confirms the artifact was signed by the registered
key.

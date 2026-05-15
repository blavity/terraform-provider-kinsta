# Release Signing & Registry Checklist

Use this to prepare for publishing to the public Terraform Registry. Private registry publishing has been removed; artifacts stay private until you intentionally tag and publish.

## 1) Create an accepted signing key (Terraform Registry)

Preferred (no manual key creation): use KMS-backed signing key via PKCS#11
- Create or reference an RSA 4096 HSM key in your GCP project: `projects/YOUR_GCP_PROJECT/locations/global/keyRings/YOUR_KEY_RING/cryptoKeys/provider-signing`.
- Present it to GPG with libkmsp11 + `gnupg-pkcs11-scd`; no `gpg --full-generate-key` needed.
- Export public key (safe to publish):
  ```bash
  gpg --export --armor > public.asc
  ```

Fallback (only if KMS path unavailable)
- Generate RSA (not ECC): `gpg --full-generate-key` → RSA+RSA, 4096, non-expiring.
  - Real Name: `Your Org Terraform Registry Signing`
  - Email: `engineering@example.com`
  - Comment: optional (leave blank or `registry signing`)
- Export public key: `gpg --armor --export <KEY_ID> > public.asc`
- Export private key: `gpg --armor --export-secret-keys <KEY_ID> > private.asc`

## 2) Register key in Terraform Registry
- Sign in as org admin → User Settings → Signing Keys → New GPG Key → paste `public.asc`.
- Verify key appears under the org; this is what validates `SHA256SUMS.sig`.

## 3) Wire CI secrets (GitHub repo)
- `GPG_PUBLIC_KEY` = contents of `public.asc` (for Terraform Registry signing key)
- If using fallback local key: `GPG_PRIVATE_KEY`, `GPG_PASSPHRASE`
- If using KMS-backed key (preferred):
  - `KMS_KEY_RESOURCE` = e.g., `projects/YOUR_GCP_PROJECT/locations/global/keyRings/YOUR_KEY_RING/cryptoKeys/provider-signing/cryptoKeyVersions/1`
  - `WIF_PROVIDER` = workload identity provider resource for repo (`projects/YOUR_PROJECT_NUMBER/locations/global/workloadIdentityPools/YOUR_POOL/providers/YOUR_PROVIDER`)
  - `WIF_SERVICE_ACCOUNT` = `provider-signing@YOUR_GCP_PROJECT.iam.gserviceaccount.com`

## 4) Public Terraform Registry publish workflow (prep)
- Ensure repo name matches `terraform-provider-kinsta` and module path `github.com/blavity/terraform-provider-kinsta` (already done).
- Cut a tagged release `vX.Y.Z` with assets:
  - `terraform-provider-kinsta_vX.Y.Z_<os>_<arch>.zip`
  - `terraform-provider-kinsta_vX.Y.Z_SHA256SUMS`
  - `terraform-provider-kinsta_vX.Y.Z_SHA256SUMS.sig` (signed with above key)
- First publish: log into Terraform Registry, choose **Publish Provider**, point to GitHub repo, and follow OAuth prompts. The Registry will install a webhook and read the release artifacts.
- Subsequent publishes: push a tag; GoReleaser + signature is enough.

## 4a) CI signing with KMS (recipe)

This is the authoritative recipe used by `.github/workflows/release.yml`. It is validated against:

- `ubuntu-24.04` GitHub-hosted runners
- `libkmsp11` 1.9 (`pkcs11-v1.9` release of `GoogleCloudPlatform/kms-integrations`)
- `gnupg-pkcs11-scd` from the Ubuntu **noble** apt repository (binary at `/usr/bin/gnupg-pkcs11-scd`, per [packages.ubuntu.com](https://packages.ubuntu.com/noble/amd64/gnupg-pkcs11-scd/filelist))

### IAM prerequisites

The signing service account (`provider-signing@<project>.iam.gserviceaccount.com`) needs **both** roles on the keyring (not just the key):

- `roles/cloudkms.signerVerifier` — sign with the key version
- `roles/cloudkms.viewer` — `libkmsp11` calls `ListCryptoKeys` during token init; without `viewer` the token comes back empty and `SCD LEARN` returns nothing

These are bound by the host platform's IAM (Cloud KMS Workload Identity Federation).

### Install + configure the signing chain

```bash
sudo apt-get update
sudo apt-get install -y gnupg gnupg-pkcs11-scd wget ca-certificates tar

# libkmsp11 1.9 — note tag is `pkcs11-v1.9`, asset is `libkmsp11-1.9-linux-amd64.tar.gz`.
LIBKMSP11_VERSION=1.9
tarball="libkmsp11-${LIBKMSP11_VERSION}-linux-amd64.tar.gz"
wget -qO "/tmp/${tarball}" \
  "https://github.com/GoogleCloudPlatform/kms-integrations/releases/download/pkcs11-v${LIBKMSP11_VERSION}/${tarball}"
# (verify sha256 — see release.yml for pinned digest)

sudo mkdir -p /opt/libkmsp11
sudo tar -xzf "/tmp/${tarball}" -C /opt/libkmsp11 --strip-components=1

mkdir -p ~/.gnupg && chmod 700 ~/.gnupg

# Delegate scdaemon to gnupg-pkcs11-scd. Note `/usr/bin`, not `/usr/lib`.
cat > ~/.gnupg/gpg-agent.conf <<'EOF'
allow-loopback-pinentry
default-cache-ttl 600
scdaemon-program /usr/bin/gnupg-pkcs11-scd
EOF

# gnupg-pkcs11-scd uses whitespace-separated directives (NOT YAML).
# Per https://manpages.debian.org/testing/gnupg-pkcs11-scd/gnupg-pkcs11-scd.1.en.html
cat > ~/.gnupg/gnupg-pkcs11-scd.conf <<'EOF'
providers kmsp11
provider-kmsp11-library /opt/libkmsp11/libkmsp11.so
log-file /tmp/gnupg-pkcs11-scd.log
verbose
EOF

# libkmsp11 config. `generate_certs: true` is REQUIRED — SCD LEARN enumerates
# keys via cert objects, and libkmsp11 defaults to not generating them.
# Reference: https://github.com/GoogleCloudPlatform/kms-integrations/blob/master/kmsp11/docs/user_guide.md
key_ring="${KMS_KEY_RESOURCE%/cryptoKeys/*}"
cat > ~/.gnupg/kmsp11.yaml <<EOF
---
tokens:
  - key_ring: "${key_ring}"
    label: "kms-provider-signing"
generate_certs: true
log_directory: "/tmp"
EOF
export KMS_PKCS11_CONFIG="$HOME/.gnupg/kmsp11.yaml"

gpgconf --kill all || true
gpgconf --launch gpg-agent
```

### Wrap the PKCS#11 key as an OpenPGP identity

`gpg --card-status` alone does **not** make a PKCS#11-backed key visible to `gpg --detach-sign`. You must surface its keygrip via `SCD LEARN` and then create an OpenPGP wrapper identity around it using `--full-generate-key` option 13 (`Existing key from keygrip`):

> The email below (`release-signing@example.com`) is a placeholder per the project's confidentiality rules. Consumers MUST substitute their actual signing UID email; CI sources this value from the `SIGNING_KEY_EMAIL` repo secret documented in §4a (no default — the workflow fails fast if it is unset).

```bash
# 1. Surface the key + extract the keygrip.
keygrip=$(gpg-connect-agent "SCD LEARN --force" /bye 2>&1 \
  | awk '/KEY-FRIEDNLY|KEYPAIRINFO/ {print $2; exit}')

# 2. Wrap it in an OpenPGP identity. Option 13 == existing keygrip.
gpg --batch --expert --command-fd 0 --pinentry-mode loopback --full-generate-key <<EOF
13
${keygrip}
0
y
terraform-provider-kinsta release signing
release-signing@example.com

O
EOF

# 3. Pin as default-key so goreleaser's `gpg --detach-sign` finds it.
keyid=$(gpg --list-keys --with-colons --keyid-format LONG \
  | awk -F: '/^pub:/ {print $5; exit}')
echo "default-key ${keyid}" >> ~/.gnupg/gpg.conf
```

### Signing

GoReleaser invokes `gpg --detach-sign terraform-provider-kinsta_vX.Y.Z_SHA256SUMS`. GPG resolves the OpenPGP key → keygrip → gpg-agent → gnupg-pkcs11-scd → libkmsp11 → Cloud KMS `AsymmetricSign`. The private key never leaves the HSM.

### Test plan (manual validation before first release)

Run locally on an `ubuntu-24.04` container, with `gcloud auth application-default login` for an identity that has both roles on the keyring:

1. Run all of the install + configure block above.
2. `gpg-connect-agent "SCD LEARN --force" /bye` — confirm a `KEYPAIRINFO` line with a keygrip is returned. If empty, check `/tmp/gnupg-pkcs11-scd.log` and `/tmp/libkmsp11_*.log`.
3. `gpg --list-keys` after the wrap step — confirm the new `pub` entry with your configured signing UID email exists (the placeholder `release-signing@example.com` is used in this doc; the actual value is supplied at CI time via the `SIGNING_KEY_EMAIL` repo secret documented in §4a, and any user MUST substitute their own address).
4. `echo test > /tmp/x && gpg --detach-sign /tmp/x` — confirm `/tmp/x.sig` is produced with no errors.
5. `gpg --verify /tmp/x.sig /tmp/x` — confirm a GOOD signature.
6. Optionally cross-check against the exported public key (`gpg --export --armor > /tmp/pub.asc`) by importing into a fresh keyring and re-verifying.

If any step fails, do **not** push a release tag — the GitHub Release will publish an unsigned `SHA256SUMS.sig` and the Terraform Registry will reject the version.

> This recipe is validated against ubuntu-24.04 runners, libkmsp11 1.9, and gnupg-pkcs11-scd from the Ubuntu noble apt repo.

## 5) Roadmap to first public release
- Add GPG key to Terraform Registry org (Section 2).
- Confirm CI secrets for signing are present (Section 3).
- Tag a release when ready; artifacts will publish to GitHub Releases and be consumed by the public Terraform Registry.
- Update README/examples to use `registry.terraform.io/blavity/kinsta` as the `source`.

## 6) OpenTofu registry considerations
- OpenTofu can consume the public Terraform Registry once published.
- If a dedicated OpenTofu private registry emerges later, mirror the same signed artifacts/JSON with an additional publish step.

## 7) Local verification (optional but recommended)
- After a tagged release, download artifacts and verify:  
  `gpg --verify terraform-provider-kinsta_vX.Y.Z_SHA256SUMS.sig terraform-provider-kinsta_vX.Y.Z_SHA256SUMS`
- Check registry JSON for the version: `jq .versions registry/v1/providers/platform/kinsta/versions`

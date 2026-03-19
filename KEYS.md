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

## 4a) CI signing with KMS (outline)
- Auth: GitHub Actions uses Workload Identity Federation to impersonate `provider-signing@YOUR_GCP_PROJECT.iam.gserviceaccount.com`.
- Configure gpg to use KMS:
  ```bash
  apt-get update && apt-get install -y gnupg gnupg-pkcs11-scd wget ca-certificates tar
  wget -qO /tmp/libkmsp11.tar.gz https://github.com/GoogleCloudPlatform/kms-integrations/releases/download/v1.3.1/libkmsp11-1.3.1-linux-amd64.tar.gz
  mkdir -p /opt/libkmsp11 && tar -xzf /tmp/libkmsp11.tar.gz -C /opt/libkmsp11 --strip-components=1
  cat > ~/.gnupg/gpg-agent.conf <<'EOF'
  allow-loopback-pinentry
  default-cache-ttl 600
  scdaemon-program /usr/lib/gnupg-pkcs11-scd
  EOF
  cat > ~/.gnupg/gnupg-pkcs11-scd.conf <<'EOF'
  module: /opt/libkmsp11/libkmsp11.so
  provider: libkmsp11
  nss_init: none
  slot: 0
  pkcs11:token-label: kms-provider-signing
  provider-privkey: ${KMS_KEY_RESOURCE}
  EOF
  chmod 700 ~/.gnupg
  gpgconf --kill all || true
  gpg --card-status || true   # trigger discovery of the KMS key
  ```
- Signing: GoReleaser runs `gpg --detach-sign terraform-provider-kinsta_vX.Y.Z_SHA256SUMS`; gpg forwards to KMS.

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

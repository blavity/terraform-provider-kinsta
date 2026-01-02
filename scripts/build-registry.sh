#!/usr/bin/env bash
# Generate static Terraform provider registry metadata for GitHub Pages.
# Inputs:
#   - Version (positional) may be "v0.0.2" or "0.0.2"
#   - REGISTRY_HOSTNAME (default: blavity.com)
#   - REGISTRY_NAMESPACE (default: platform)
#   - REGISTRY_PROVIDER (default: kinsta)
#   - REGISTRY_OUTPUT_DIR (default: registry)
#   - DIST_DIR (default: dist)
#   - GITHUB_REPOSITORY (default: blavity/terraform-provider-kinsta)
#   - GPG_PUBLIC_KEY (ascii-armored public key, required)
#   - GPG_KEY_ID (hex key id, required)

set -euo pipefail

RAW_VERSION="${1:?version (e.g., v0.0.2 or 0.0.2) is required}"
VERSION="${RAW_VERSION#v}"
HOST="${TERRAFORM_REGISTRY_HOSTNAME:-${REGISTRY_HOSTNAME:-blavity.com}}"
NAMESPACE="${REGISTRY_NAMESPACE:-platform}"
PROVIDER="${REGISTRY_PROVIDER:-kinsta}"
DIST_DIR="${DIST_DIR:-dist}"
OUT_DIR="${REGISTRY_OUTPUT_DIR:-registry}"
REPO="${GITHUB_REPOSITORY:-blavity/terraform-provider-kinsta}"
GPG_PUBLIC_KEY="${GPG_PUBLIC_KEY:-}"
GPG_KEY_ID="${GPG_KEY_ID:-}"

if [[ -z "${GPG_PUBLIC_KEY}" || -z "${GPG_KEY_ID}" ]]; then
  echo "GPG_PUBLIC_KEY and GPG_KEY_ID environment variables are required" >&2
  exit 1
fi

CHECKSUM_FILE="${DIST_DIR}/terraform-provider-kinsta_v${VERSION}_SHA256SUMS"
if [[ ! -f "${CHECKSUM_FILE}" ]]; then
  echo "Checksum file not found: ${CHECKSUM_FILE}" >&2
  exit 1
fi

rm -rf "${OUT_DIR}"
mkdir -p "${OUT_DIR}/.well-known"

cat > "${OUT_DIR}/.well-known/terraform.json" <<EOF
{
  "providers.v1": "https://${HOST}/v1/providers/"
}
EOF

export VERSION
export OUT_DIR
export NAMESPACE
export PROVIDER
export CHECKSUM_FILE
export REPO
export GPG_PUBLIC_KEY
export GPG_KEY_ID
export VERSIONS_PATH="${OUT_DIR}/v1/providers/${NAMESPACE}/${PROVIDER}/versions"

python3 - <<'PY'
import json
import os
import pathlib
import re
import sys

version = os.environ["VERSION"]
checksum_file = pathlib.Path(os.environ["CHECKSUM_FILE"])
out_dir = pathlib.Path(os.environ["OUT_DIR"])
namespace = os.environ["NAMESPACE"]
provider = os.environ["PROVIDER"]
repo = os.environ["REPO"]
gpg_key = os.environ["GPG_PUBLIC_KEY"]
gpg_id = os.environ["GPG_KEY_ID"]
versions_path = pathlib.Path(os.environ["VERSIONS_PATH"])

pattern = re.compile(
    rf"^(?P<sha>[a-f0-9]{{64}})\\s+terraform-provider-kinsta_v{re.escape(version)}_(?P<os>[^_]+)_(?P<arch>[^.]+)\\.zip$"
)

entries = []
with checksum_file.open() as f:
    for line in f:
        line = line.strip()
        if not line:
            continue
        match = pattern.match(line)
        if not match:
            continue
        sha = match.group("sha")
        os_name = match.group("os")
        arch = match.group("arch")
        filename = line.split()[1]
        entries.append((os_name, arch, sha, filename))

if not entries:
    sys.stderr.write(f"No matching artifacts found in {checksum_file}\\n")
    sys.exit(1)

platforms = sorted({(os_name, arch) for os_name, arch, _, _ in entries})
versions_payload = {
    "versions": [
        {
            "version": version,
            "protocols": ["5.0"],
            "platforms": [{"os": os_name, "arch": arch} for os_name, arch in platforms],
        }
    ]
}

versions_path.parent.mkdir(parents=True, exist_ok=True)
with versions_path.open("w") as f:
    json.dump(versions_payload, f, indent=2)

download_base = f"https://github.com/{repo}/releases/download/v{version}"

for os_name, arch, sha, filename in entries:
    target_dir = out_dir / "v1" / "providers" / namespace / provider / version / "download" / os_name
    target_dir.mkdir(parents=True, exist_ok=True)
    payload = {
        "protocols": ["5.0"],
        "os": os_name,
        "arch": arch,
        "filename": filename,
        "download_url": f"{download_base}/{filename}",
        "shasum": sha,
        "signing_keys": {
            "gpg_public_keys": [
                {"key_id": gpg_id, "ascii_armor": gpg_key}
            ]
        },
    }
    with (target_dir / arch).open("w") as f:
        json.dump(payload, f, indent=2)

print(f"Wrote registry metadata to {out_dir}")
PY

# Terraform Provider for Kinsta (MyKinsta API)

Manage your Kinsta WordPress hosting infrastructure with Terraform.

**Status:** Pre-release (v0.1)

## Supported Resources

- `kinsta_wordpress_site` — WordPress site management (create, read, delete)
- `kinsta_wordpress_environment` — Environment management (staging, production, clone)

## Scope

This provider manages WordPress resources via the MyKinsta API (`api.kinsta.com/v2`).

**Note:** PaaS resources (applications, databases, static sites) are managed by the separate Sevalla provider.

## Installation

```hcl
terraform {
  required_providers {
    kinsta = {
      source  = "blavity/kinsta"
      version = "~> 0.1"
    }
  }
}

provider "kinsta" {
  api_key    = var.kinsta_api_key    # or env KINSTA_API_KEY
  company_id = var.kinsta_company_id # or env KINSTA_COMPANY_ID
}
```

## Authentication

Set credentials via environment variables (recommended):

```bash
export KINSTA_API_KEY="your-api-key"
export KINSTA_COMPANY_ID="your-company-id"
```

Or provide them directly in the provider block (not recommended for production).

## Example Usage

```hcl
resource "kinsta_wordpress_site" "example" {
  display_name   = "My WordPress Site"
  region         = "us-central1"
  admin_email    = "admin@example.com"
  admin_password = var.admin_password
  admin_user     = "admin"
  site_title     = "My WordPress Site"
  wp_language    = "en_US"
}

resource "kinsta_wordpress_environment" "staging" {
  site_id      = kinsta_wordpress_site.example.site_id
  display_name = "staging"
}

output "site_id" {
  value = kinsta_wordpress_site.example.site_id
}
```

## Documentation

See `docs/resources/` for complete resource documentation:

- [kinsta_wordpress_site](docs/resources/wordpress_site.md)
- [kinsta_wordpress_environment](docs/resources/wordpress_environment.md)

## Development

Requirements: Go 1.25+, [golangci-lint](https://golangci-lint.run/welcome/install/) v2+

```bash
# Build
go build ./...

# Unit tests
go test ./internal/...

# Lint
golangci-lint run ./...

# Acceptance tests (requires live Kinsta credentials)
TF_ACC=1 go test ./internal/provider/ -v

# Regenerate docs (requires tfplugindocs)
go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest
tfplugindocs generate --provider-name kinsta
```

Dev override (use a local build without the registry):

```hcl
# ~/.terraformrc
provider_installation {
  dev_overrides {
    "blavity/kinsta" = "/path/to/terraform-provider-kinsta"
  }
  direct {}
}
```

## Releasing

Releases use [GoReleaser](https://goreleaser.com/) triggered by a manual tag push — the canonical Terraform provider pattern (`terraform-provider-aws`, `-google`, the HashiCorp scaffolding-framework template):

1. From `main`, push a semver tag:
   ```bash
   git tag v0.3.0
   git push origin v0.3.0
   ```
2. The tag triggers the release workflow, which builds multi-platform binaries, signs them with GPG, generates a conventional-commit-grouped changelog into the GitHub Release body, and uploads the registry-shaped artifacts (zips, `SHA256SUMS`, `.sig`, manifest).
3. The [Terraform Registry](https://registry.terraform.io) (and [OpenTofu Registry](https://github.com/opentofu/registry)) pick up the release automatically.

Required repository secrets: `GPG_PRIVATE_KEY`, `PASSPHRASE`.

## Trademarks

Kinsta, MyKinsta, and WordPress are trademarks or registered trademarks of their respective owners. Blavity, Inc. is not affiliated with, endorsed by, or sponsored by Kinsta Ltd. or Automattic Inc. All other trademarks are the property of their respective owners.

## License

[Mozilla Public License 2.0](LICENSE) — Copyright (c) 2024–2026 Blavity, Inc.

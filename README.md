# Terraform Provider for Kinsta (MyKinsta API)

Manage your Kinsta WordPress hosting infrastructure with Terraform.

**Status:** Private registry-ready (v0.0.2)

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
      source  = "blavity.com/platform/kinsta"
      version = "~> 0.0.2"
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
  install_mode = "clone"
  source_env_id = kinsta_wordpress_site.example.environment_id
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

Requirements: Go 1.25+

```bash
# Build
go build ./...

# Unit tests
go test ./internal/...

# Acceptance tests (requires live credentials)
TF_ACC=1 go test ./internal/provider/ -v
```

Dev override (no registry):

```hcl
provider_installation {
  dev_overrides {
    "blavity.com/platform/kinsta" = "/path/to/terraform-provider-kinsta"
  }
  direct {}
}
```

See `KEYS.md` for Terraform Registry publication instructions.

## License

Internal use only — Blavity Inc.

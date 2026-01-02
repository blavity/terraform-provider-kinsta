# Terraform Provider for Kinsta

Terraform/OpenTofu provider for managing Kinsta WordPress hosting sites.

**Status**: Private registry-ready (v0.0.2)

## Using the provider (private registry)

Once the GitHub Pages registry is published, consume the provider without dev overrides:

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
  api_key    = var.kinsta_api_key      # or env KINSTA_API_KEY
  company_id = var.kinsta_company_id   # or env KINSTA_COMPANY_ID
}
```

## Local development

- Go 1.25+
- Kinsta API key (see [Kinsta API docs](https://kinsta.com/docs/kinsta-api/))

Dev override (no registry) remains available:

```hcl
provider_installation {
  dev_overrides {
    "blavity.com/platform/kinsta" = "/path/to/terraform-provider-kinsta"
  }
  direct {}
}
```

Export for local plans:

```bash
export TF_CLI_CONFIG_FILE=~/.terraform.d/provider_override.tfrc
go build ./...
tofu plan
```

## API Specification

The Kinsta API spec (`swagger.json`) is **not included** in this repository.  
Download the latest spec from [Kinsta API Documentation](https://api-docs.kinsta.com/) for reference.

## Preparing for the public Terraform Registry

- Follow `KEYS.md` to register a Terraform Registry–approved GPG key and wire CI secrets.
- When ready to switch from GitHub Pages to the public registry, disable the Pages publish step in `.github/workflows/release.yml` and update consumers to `source = "registry.terraform.io/blavity/kinsta"`.
- OpenTofu users can consume either the private registry (`blavity.com/platform/kinsta`) or the public Terraform Registry once published—the protocol is identical.

## v0.0.1 Features

- ✅ Create WordPress sites
- ✅ Read site details (site_id, environment_id)
- ✅ Delete sites
- ✅ Async operation polling

**Out of scope**: SFTP credentials retrieval (use Kinsta API directly), site updates

## Example Usage

```hcl
resource "kinsta_wordpress_site" "example" {
  display_name   = "My WordPress Site"
  region         = "us-central1"
  install_mode   = "new"
  admin_email    = "admin@example.com"
  admin_password = "secure-password"
  admin_user     = "admin"
  site_title     = "My Site"
  wp_language    = "en_US"
}

output "site_id" {
  value = kinsta_wordpress_site.example.site_id
}
```

## License

Internal use only - Blavity Inc.

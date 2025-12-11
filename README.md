# Terraform Provider for Kinsta

**Status**: Development (v0.0.1)

Terraform/OpenTofu provider for managing Kinsta WordPress hosting sites.

## Development

### Prerequisites

- Go 1.21+
- Kinsta API key (see [Kinsta API docs](https://kinsta.com/docs/kinsta-api/))

### Building

```bash
go build
```

### Local Testing

Set up dev overrides in `~/.terraform.d/provider_override.tfrc`:

```hcl
provider_installation {
  dev_overrides {
    "blavity.com/platform/kinsta" = "/path/to/terraform-provider-kinsta"
  }
  direct {}
}
```

Then use with `TF_CLI_CONFIG_FILE`:

```bash
export TF_CLI_CONFIG_FILE=~/.terraform.d/provider_override.tfrc
tofu plan
```

### API Specification

The Kinsta API spec (`swagger.json`) is **not included** in this repository.  
Download the latest spec from [Kinsta API Documentation](https://api-docs.kinsta.com/) for reference.

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

# Terraform/OpenTofu Provider for Kinsta

A Terraform/OpenTofu provider for managing Kinsta resources through the Kinsta API.

[![License: MPL 2.0](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)

> **⚠️ IMPORTANT NOTICE:** This provider is independently developed by Blavity, Inc. and is NOT officially supported by Kinsta. See [NOTICE](NOTICE) for full disclaimers. Kinsta and related trademarks are property of Kinsta Inc.

## Features

This provider allows you to manage the following Kinsta resources:

- **WordPress Sites** - Create and manage WordPress installations
- **Applications** - Deploy and manage applications on Kinsta's platform
- **Databases** - Create and configure database instances

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.5.7 or [OpenTofu](https://opentofu.org/) >= 1.6.0
- [Go](https://golang.org/doc/install) >= 1.21 (for development)
- A Kinsta account with API access
- Kinsta API credentials (API Key and Company ID)

## Installation

### Terraform Registry (Coming Soon)

Once published to the registry, you'll be able to use:

```hcl
terraform {
  required_providers {
    kinsta = {
      source  = "blavity/kinsta"
      version = "~> 1.0"
    }
  }
}
```

### OpenTofu Registry (Coming Soon)

For OpenTofu users:

```hcl
terraform {
  required_providers {
    kinsta = {
      source  = "blavity/kinsta"
      version = "~> 1.0"
    }
  }
}
```

### Manual Installation (Current)

Until published to registries, you can build from source:

```bash
git clone https://github.com/blavity/terraform-provider-kinsta
cd terraform-provider-kinsta
go build -o terraform-provider-kinsta
```

Then copy the binary to your Terraform plugins directory.

## Usage

### Provider Configuration

```hcl
provider "kinsta" {
  api_key    = var.kinsta_api_key     # or set KINSTA_API_KEY env var
  company_id = var.kinsta_company_id  # or set KINSTA_COMPANY_ID env var
}
```

### Environment Variables

The provider supports the following environment variables:

- `KINSTA_API_KEY` - Your Kinsta API key
- `KINSTA_COMPANY_ID` - Your Kinsta company ID

### Example: Creating a WordPress Site

```hcl
resource "kinsta_wordpress_site" "example" {
  display_name = "My WordPress Site"
  # Additional configuration...
}
```

### Example: Creating an Application

```hcl
resource "kinsta_application" "example" {
  display_name = "My Application"
  # Additional configuration...
}
```

### Example: Creating a Database

```hcl
resource "kinsta_database" "example" {
  display_name = "My Database"
  # Additional configuration...
}
```

For complete examples, see the [examples/](examples/) directory.

## Documentation

Full documentation for resources and data sources is available in the [docs/](docs/) directory:

- [Provider Configuration](docs/index.md)
- [kinsta_wordpress_site](docs/resources/wordpress_site.md)
- [kinsta_application](docs/resources/application.md)
- [kinsta_database](docs/resources/database.md)

## Development

### Building

```bash
go build -o terraform-provider-kinsta
```

### Testing

Run unit tests:

```bash
go test ./... -v
```

Run acceptance tests (requires Kinsta API credentials):

```bash
TF_ACC=true KINSTA_API_KEY=your_key KINSTA_COMPANY_ID=your_id go test ./... -v
```

### Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details on:

- How to submit issues
- How to submit pull requests
- Code style guidelines
- Testing requirements

## Compatibility

| Provider Version | Terraform Version | OpenTofu Version | Go Version |
|-----------------|-------------------|------------------|------------|
| 1.x             | >= 1.5.7          | >= 1.6.0         | >= 1.21    |

**Note:** We primarily test with OpenTofu and Terraform 1.5.7. Newer Terraform versions may work but are not officially tested.

## Support

This is a community-maintained provider with no official support from either Blavity or Kinsta.

- **Issues & Bug Reports:** [GitHub Issues](https://github.com/blavity/terraform-provider-kinsta/issues)
- **Feature Requests:** [GitHub Issues](https://github.com/blavity/terraform-provider-kinsta/issues)
- **Discussions:** [GitHub Discussions](https://github.com/blavity/terraform-provider-kinsta/discussions)

**Please note:** Support is provided on a best-effort basis by the community.

## Security

For security concerns, please see [SECURITY.md](SECURITY.md).

## License

This provider is licensed under the Mozilla Public License 2.0. See [LICENSE](LICENSE) for full details.

## Legal

This provider is an independent project by Blavity, Inc. and is not affiliated with, endorsed by, or sponsored by Kinsta Inc. All trademarks are property of their respective owners. See [NOTICE](NOTICE) for complete disclaimers.

## Acknowledgments

- Built using the [Terraform Plugin SDK](https://github.com/hashicorp/terraform-plugin-sdk)
- Compatible with [OpenTofu](https://opentofu.org/)
- Uses the Kinsta API (https://kinsta.com/api/)

---

**Made with ❤️ by Blavity, Inc.**

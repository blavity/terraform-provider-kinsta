# Terraform Provider Kinsta

Terraform provider for managing Kinsta resources.

## Example Usage

```hcl
terraform {
  required_providers {
    kinsta = {
      source = "blavity/kinsta"
      # It's a good practice to pin the version of the provider.
      # version = "x.y.z"
    }
  }
}

provider "kinsta" {
  api_key    = var.kinsta_api_key
  company_id = var.kinsta_company_id
}
```

## Argument Reference
The following arguments are supported in the provider block:

- `api_key` - (Required) The API key for the Kinsta API. This can also be provided via the `KINSTA_API_KEY` environment variable.
- `company_id` - (Required) The ID of your Kinsta company. This can also be provided via the `KINSTA_COMPANY_ID` environment variable.
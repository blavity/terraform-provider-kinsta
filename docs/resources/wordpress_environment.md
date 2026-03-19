# kinsta_wordpress_environment

Manages a WordPress environment (staging or premium staging) for an existing Kinsta WordPress site.

Environments are additional instances of your WordPress site used for development, testing, or preview purposes. You can create standard staging environments (free) or premium staging environments (paid feature with dedicated resources).

## Important Notes

- **All fields are immutable** - Any change to configuration requires replacing the environment
- **Write-only fields** - Several fields (`is_premium`, admin credentials, `site_title`) cannot be read back from the API after creation, which is normal behavior
- **Depends on parent site** - Must reference an existing `kinsta_wordpress_site` resource
- **Async operation** - Environment creation takes 1-3 minutes and uses polling to track progress

## Example Usage

### Basic Staging Environment

```hcl
resource "kinsta_wordpress_site" "main" {
  display_name   = "My Production Site"
  region         = "us-central1"
  admin_email    = "admin@example.com"
  admin_password = var.admin_password
  admin_user     = "admin"
  site_title     = "My WordPress Site"
}

resource "kinsta_wordpress_environment" "staging" {
  site_id      = kinsta_wordpress_site.main.site_id
  display_name = "staging"
}
```

### Premium Staging Environment

```hcl
resource "kinsta_wordpress_environment" "premium_staging" {
  site_id        = kinsta_wordpress_site.main.site_id
  display_name   = "premium-staging"
  is_premium     = true
  admin_email    = "staging@example.com"
  admin_user     = "stagingadmin"
  admin_password = var.staging_password
  site_title     = "Premium Staging"
}
```

## Argument Reference

### Required Arguments

- `site_id` (String, ForceNew) - The ID of the parent WordPress site. Get this from `kinsta_wordpress_site.example.site_id`.
- `display_name` (String, ForceNew) - Environment name as shown in MyKinsta (e.g., "staging", "premium-staging"). Should be unique within the site.

### Optional Arguments

- `is_premium` (Boolean, ForceNew) - Whether to create a premium staging environment with dedicated resources. Default: `false`. **Note:** This is a write-only field and cannot be read back after creation.
- `admin_email` (String, ForceNew, Sensitive) - WordPress admin email address
- `admin_user` (String, ForceNew, Sensitive) - WordPress admin username
- `admin_password` (String, ForceNew, Sensitive) - WordPress admin password
- `site_title` (String, ForceNew) - WordPress site title
- `wp_language` (String, ForceNew) - WordPress locale code (e.g., `"en_US"`, `"fr_FR"`). Default: `"en_US"`

**Note:** Admin credential fields are write-only and cannot be read back after creation.

## Attribute Reference

- `id` (String) - Environment ID, discovered after creation
- `environment_id` (String) - Environment ID (same as `id`)

## Timeouts

| Operation | Default |
|-----------|---------|
| Create    | 15 minutes |
| Delete    | 15 minutes |

## Import

Environments can be imported using the format `site_id:environment_id`:

```bash
terraform import kinsta_wordpress_environment.staging fbab4927-e354-4044-b226-29ac0fbd20ca:c84ce214-69b9-4a32-8e67-880672cf1d38
```

You can find both IDs in the MyKinsta URL when viewing an environment:
```
https://my.kinsta.com/sites/details/{site_id}/{environment_id}?idCompany=...
```

## Limitations and Behavior

### No Updates - ForceNew on All Fields

All configuration fields are immutable. Changing any field will trigger environment replacement (destroy + recreate). This matches Kinsta's platform behavior where environments cannot be modified after creation.

### Write-Only Fields Don't Cause Drift

The following fields are write-only and cannot be read back from Kinsta's API:
- `is_premium`
- `admin_email`, `admin_user`, `admin_password`
- `site_title`

Terraform will not show these as "changed" after initial creation, even though they're not visible in API responses. This is expected behavior.

### Environment ID Discovery

After creation, the environment ID is discovered by comparing the site's environment list before and after the operation. For this to work reliably:
- Use unique `display_name` values within each site
- Avoid creating multiple environments simultaneously with the same name

### Creation Time

Environment creation is asynchronous and typically takes 1-3 minutes.

## References

- [Kinsta Staging Environments Documentation](https://kinsta.com/docs/wordpress-hosting/staging-environment/)
- [Premium Staging Environments](https://kinsta.com/docs/wordpress-hosting/staging-environment#premium-staging-environments)
- [Kinsta API - WordPress Site Environments](https://api-docs.kinsta.com/tag/WordPress-Site-Environments)

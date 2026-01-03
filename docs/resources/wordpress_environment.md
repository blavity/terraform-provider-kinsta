# kinsta_wordpress_environment

Manages a WordPress environment (staging or premium staging) for an existing Kinsta WordPress site.

Environments are additional instances of your WordPress site used for development, testing, or preview purposes. You can create standard staging environments (free) or premium staging environments (paid feature with dedicated resources).

## Important Notes

- **All fields are immutable** - Any change to configuration requires replacing the environment
- **Write-only fields** - Several fields (`is_premium`, admin credentials, `site_title`) cannot be read back from the API after creation, which is normal behavior
- **Depends on parent site** - Must reference an existing `kinsta_wordpress_site` resource
- **Async operation** - Environment creation takes 1-3 minutes and uses polling to track progress

## Example Usage

### Basic Staging Environment (Clone from Live)

```hcl
resource "kinsta_wordpress_site" "main" {
  display_name   = "My Production Site"
  region         = "us-central1"
  admin_email    = "admin@example.com"
  admin_password = "SecurePassword123!"
  admin_user     = "admin"
  site_title     = "My WordPress Site"
}

resource "kinsta_wordpress_environment" "staging" {
  site_id       = kinsta_wordpress_site.main.site_id
  display_name  = "staging"
  install_mode  = "clone"
  source_env_id = kinsta_wordpress_site.main.environment_id
}
```

### Premium Staging Environment with Fresh WordPress Install

```hcl
resource "kinsta_wordpress_environment" "premium_staging" {
  site_id        = kinsta_wordpress_site.main.site_id
  display_name   = "premium-staging"
  is_premium     = true
  install_mode   = "new"
  admin_email    = "staging@example.com"
  admin_user     = "stagingadmin"
  admin_password = "StagingPass123!"
  site_title     = "Premium Staging"
  php_version    = "8.3"
}
```

### Staging Environment with Custom PHP and Debug Settings

```hcl
resource "kinsta_wordpress_environment" "dev_staging" {
  site_id            = kinsta_wordpress_site.main.site_id
  display_name       = "dev-staging"
  install_mode       = "clone"
  source_env_id      = kinsta_wordpress_site.main.environment_id
  php_version        = "8.2"
  wp_debug           = true
  wp_debug_display   = true
  wp_debug_log       = true
}
```

### Empty Environment (No WordPress)

```hcl
resource "kinsta_wordpress_environment" "empty" {
  site_id      = kinsta_wordpress_site.main.site_id
  display_name = "custom-app"
  install_mode = "none"
}
```

## Argument Reference

### Required Arguments

- `site_id` (String, ForceNew) - The ID of the parent WordPress site. Get this from `kinsta_wordpress_site.example.site_id`.
- `display_name` (String, ForceNew) - Environment name as shown in MyKinsta (e.g., "staging", "premium-staging", "dev"). Should be unique within the site for easier management.

### Optional Arguments

#### Installation Mode

- `install_mode` (String, ForceNew) - How to initialize the environment. Options:
  - `"clone"` (default) - Clone from another environment (requires `source_env_id`)
  - `"new"` - Fresh WordPress installation (requires `admin_email`, `admin_user`, `admin_password`)
  - `"none"` - Empty environment without WordPress

- `source_env_id` (String, ForceNew) - Source environment ID when `install_mode = "clone"`. Typically use `kinsta_wordpress_site.example.environment_id` to clone from live.

#### Premium Features

- `is_premium` (Boolean, ForceNew) - Whether to create a premium staging environment with dedicated resources. Default: `false`. **Note:** This is a write-only field and cannot be read back after creation.

#### WordPress Admin Credentials (Required for `install_mode = "new"`)

- `admin_email` (String, ForceNew, Sensitive) - WordPress admin email address
- `admin_user` (String, ForceNew, Sensitive) - WordPress admin username
- `admin_password` (String, ForceNew, Sensitive) - WordPress admin password
- `site_title` (String, ForceNew) - WordPress site title

**Note:** These fields are write-only and cannot be read back after creation.

#### PHP and Debug Settings

- `php_version` (String, ForceNew) - PHP version for the environment. Options: `"8.0"`, `"8.1"`, `"8.2"`, `"8.3"`. If not specified, uses site default.

- `wp_debug` (Boolean, ForceNew) - Enable WordPress debug mode (WP_DEBUG). Default: `false`
- `wp_debug_display` (Boolean, ForceNew) - Display errors on page (WP_DEBUG_DISPLAY). Default: `false`
- `wp_debug_log` (Boolean, ForceNew) - Log errors to file (WP_DEBUG_LOG). Default: `false`

## Attribute Reference

- `id` (String) - Environment ID, discovered after creation
- `is_blocked` (Boolean) - Whether the environment is currently blocked
- `container_id` (String) - Container ID for the environment
- `ssh_connection_string` (String) - SSH connection details

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

Environment creation is asynchronous and typically takes 1-3 minutes depending on the installation mode:
- `"none"` (empty): ~1 minute
- `"new"` (fresh WordPress): ~2 minutes  
- `"clone"`: ~2-3 minutes (depends on source environment size)

## Common Use Cases

### Development Workflow

```hcl
# Production site
resource "kinsta_wordpress_site" "prod" {
  display_name   = "production.example.com"
  region         = "us-central1"
  admin_email    = "admin@example.com"
  admin_password = var.wp_admin_password
  admin_user     = "admin"
  site_title     = "Example Production"
}

# Staging for testing
resource "kinsta_wordpress_environment" "staging" {
  site_id       = kinsta_wordpress_site.prod.site_id
  display_name  = "staging"
  install_mode  = "clone"
  source_env_id = kinsta_wordpress_site.prod.environment_id
}

# Development with debug enabled
resource "kinsta_wordpress_environment" "dev" {
  site_id          = kinsta_wordpress_site.prod.site_id
  display_name     = "dev"
  install_mode     = "clone"
  source_env_id    = kinsta_wordpress_site.prod.environment_id
  wp_debug         = true
  wp_debug_display = true
  wp_debug_log     = true
}
```

### Premium Staging for Client Reviews

```hcl
resource "kinsta_wordpress_environment" "client_preview" {
  site_id        = kinsta_wordpress_site.prod.site_id
  display_name   = "client-preview"
  is_premium     = true
  install_mode   = "clone"
  source_env_id  = kinsta_wordpress_site.prod.environment_id
}
```

## References

- [Kinsta Staging Environments Documentation](https://kinsta.com/docs/wordpress-hosting/staging-environment/)
- [Premium Staging Environments](https://kinsta.com/docs/wordpress-hosting/staging-environment#premium-staging-environments)
- [Kinsta API - WordPress Site Environments](https://api-docs.kinsta.com/tag/WordPress-Site-Environments)

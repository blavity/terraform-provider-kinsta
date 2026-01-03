terraform {
  required_providers {
    kinsta = {
      source  = "blavity/kinsta"
      version = "~> 0.1"
    }
  }
}

provider "kinsta" {
  # Authentication via KINSTA_API_KEY environment variable
  # api_key = var.kinsta_api_key  # Alternative: explicit configuration
}

# Variables for sensitive data
variable "wp_admin_password" {
  description = "WordPress admin password"
  type        = string
  sensitive   = true
}

variable "staging_admin_password" {
  description = "Staging environment admin password"
  type        = string
  sensitive   = true
}

# Main production WordPress site
resource "kinsta_wordpress_site" "production" {
  display_name   = "production.example.com"
  region         = "us-central1"
  admin_email    = "admin@example.com"
  admin_password = var.wp_admin_password
  admin_user     = "admin"
  site_title     = "Example Production Site"
  wp_language    = "en_US"
}

# Example 1: Basic staging environment (clone from live)
resource "kinsta_wordpress_environment" "staging" {
  site_id       = kinsta_wordpress_site.production.site_id
  display_name  = "staging"
  install_mode  = "clone"
  source_env_id = kinsta_wordpress_site.production.environment_id

  # Staging inherits all content and settings from live environment
}

# Example 2: Premium staging environment with fresh WordPress install
resource "kinsta_wordpress_environment" "premium_staging" {
  site_id        = kinsta_wordpress_site.production.site_id
  display_name   = "premium-staging"
  is_premium     = true
  install_mode   = "new"
  admin_email    = "staging@example.com"
  admin_user     = "stagingadmin"
  admin_password = var.staging_admin_password
  site_title     = "Premium Staging"
  php_version    = "8.3"
}

# Example 3: Development environment with debug enabled
resource "kinsta_wordpress_environment" "dev" {
  site_id          = kinsta_wordpress_site.production.site_id
  display_name     = "dev"
  install_mode     = "clone"
  source_env_id    = kinsta_wordpress_site.production.environment_id
  php_version      = "8.2"
  wp_debug         = true
  wp_debug_display = true
  wp_debug_log     = true
}

# Example 4: Empty environment for custom application
resource "kinsta_wordpress_environment" "custom_app" {
  site_id      = kinsta_wordpress_site.production.site_id
  display_name = "custom-app"
  install_mode = "none"
  php_version  = "8.3"
}

# Outputs
output "production_site_id" {
  description = "Production site ID"
  value       = kinsta_wordpress_site.production.site_id
}

output "production_environment_id" {
  description = "Production environment ID (live)"
  value       = kinsta_wordpress_site.production.environment_id
}

output "staging_environment_id" {
  description = "Staging environment ID"
  value       = kinsta_wordpress_environment.staging.id
}

output "premium_staging_environment_id" {
  description = "Premium staging environment ID"
  value       = kinsta_wordpress_environment.premium_staging.id
}

output "dev_environment_id" {
  description = "Dev environment ID"
  value       = kinsta_wordpress_environment.dev.id
}

output "staging_ssh_connection" {
  description = "SSH connection string for staging"
  value       = kinsta_wordpress_environment.staging.ssh_connection_string
  sensitive   = true
}

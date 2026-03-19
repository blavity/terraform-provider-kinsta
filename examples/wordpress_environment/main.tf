terraform {
  required_providers {
    kinsta = {
      source  = "blavity/kinsta"
      version = "~> 0.1"
    }
  }
}

provider "kinsta" {
  # Authentication via environment variables:
  # KINSTA_API_KEY and KINSTA_COMPANY_ID
}

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

# Example 1: Staging environment
resource "kinsta_wordpress_environment" "staging" {
  site_id      = kinsta_wordpress_site.production.site_id
  display_name = "staging"
}

# Example 2: Premium staging environment
resource "kinsta_wordpress_environment" "premium_staging" {
  site_id        = kinsta_wordpress_site.production.site_id
  display_name   = "premium-staging"
  is_premium     = true
  admin_email    = "staging@example.com"
  admin_user     = "stagingadmin"
  admin_password = var.staging_admin_password
  site_title     = "Premium Staging"
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

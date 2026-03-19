terraform {
  required_providers {
    kinsta = {
      source  = "blavity/kinsta"
      version = "~> 0.1"
    }
  }
}

provider "kinsta" {
  # Authentication via environment variables (recommended):
  # export KINSTA_API_KEY="your-api-key"
  # export KINSTA_COMPANY_ID="your-company-id"
}

variable "admin_password" {
  description = "WordPress admin password"
  type        = string
  sensitive   = true
}

resource "kinsta_wordpress_site" "staging" {
  display_name   = "My Staging Site"
  region         = "us-central1"
  admin_email    = "admin@example.com"
  admin_password = var.admin_password
  admin_user     = "admin"
  site_title     = "My Staging Site"
  wp_language    = "en_US"
}

resource "kinsta_wordpress_environment" "staging" {
  site_id      = kinsta_wordpress_site.staging.site_id
  display_name = "staging"
}

output "site_id" {
  value = kinsta_wordpress_site.staging.site_id
}

output "environment_id" {
  value = kinsta_wordpress_environment.staging.id
}

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

# Basic WordPress site
resource "kinsta_wordpress_site" "example" {
  display_name   = "My WordPress Site"
  region         = "us-central1"
  admin_email    = "admin@example.com"
  admin_password = var.admin_password
  admin_user     = "admin"
  site_title     = "My WordPress Site"
  wp_language    = "en_US"
}

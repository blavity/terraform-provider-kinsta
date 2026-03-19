terraform {
  required_version = ">= 1.10.0"
  required_providers {
    kinsta = {
      source  = "kinsta/kinsta"
      version = "0.1.0"
    }
  }
}

provider "kinsta" {
  api_key    = var.kinsta_api_key
  company_id = var.kinsta_company_id
}

resource "kinsta_wordpress_site" "test" {
  display_name   = "test-debug-site-v2"
  region         = "us-central1"
  admin_email    = "admin@example.com"
  admin_password = "TempPassword123!"
  admin_user     = "admin"
  site_title     = "Test Debug Site V2"
  wp_language    = "en_US"
}

output "site_id" {
  value = kinsta_wordpress_site.test.site_id
}

output "environment_id" {
  value = kinsta_wordpress_site.test.environment_id
}

resource "kinsta_wordpress_site" "example" {
  display_name   = "my-production-site"
  region         = "us-central1"
  admin_email    = "admin@example.com"
  admin_password = var.admin_password
  admin_user     = "admin"
  site_title     = "My WordPress Site"
  wp_language    = "en_US"
}

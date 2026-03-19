# WordPress Multisite with WooCommerce and Yoast SEO
resource "kinsta_wordpress_site" "multisite" {
  display_name          = "My Multisite Network"
  region                = "us-central1"
  admin_email           = "admin@example.com"
  admin_password        = var.admin_password
  admin_user            = "admin"
  site_title            = "My Multisite Network"
  wp_language           = "en_US"
  is_multisite          = true
  is_subdomain_multisite = false # subdirectory-based multisite
  woocommerce           = true
  wordpressseo          = true
}

resource "kinsta_wordpress_environment" "staging" {
  site_id      = kinsta_wordpress_site.example.site_id
  display_name = "staging"
}

output "site_id" {
  description = "The WordPress site ID"
  value       = kinsta_wordpress_site.example.site_id
}

output "environment_id" {
  description = "The live environment ID (auto-created with the site)"
  value       = kinsta_wordpress_site.example.environment_id
}
